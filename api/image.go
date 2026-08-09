package api

import (
	"Slink/applog"
	"Slink/middleware"
	"Slink/model"
	"Slink/utils"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// discord面板返回结构
type DiscordResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UploadImageResponse 图片上传响应结构
type UploadImageResponse struct {
	Status  bool       `json:"status"`
	Message string     `json:"message"`
	Data    *ImageData `json:"data"`
}

// ImageData 图片数据结构
type ImageData struct {
	ID           uint              `json:"id"`
	Pathname     string            `json:"pathname"`
	OriginName   string            `json:"origin_name"`
	Size         float64           `json:"size"`
	SizeBytes    int64             `json:"size_bytes"`
	Mimetype     string            `json:"mimetype"`
	MD5          string            `json:"md5"`
	SHA1         string            `json:"sha1"`
	Links        map[string]string `json:"links"`
	StrategyID   uint              `json:"strategy_id"`
	StrategyName string            `json:"strategy_name"`
	Permission   uint              `json:"permission"`
	CreatedAt    time.Time         `json:"created_at"`
}

// GetImagesResponse 获取图片列表响应结构
type GetImagesResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    []ImageData `json:"data"`
}

func UploadImage(c *gin.Context) {
	rid := middleware.RequestID(c)
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		applog.Logger.Warn("upload: unauthenticated", "request_id", rid, "handler", "UploadImage")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取用户信息
	user, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		applog.Logger.Error("upload: get user failed", "request_id", rid, "handler", "UploadImage", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	policy, ok := readUserUploadPolicy(c, userID.(uint))
	if !ok {
		return
	}

	// 获取文件
	file, err := c.FormFile("image")
	if err != nil {
		applog.Logger.Warn("upload: no file", "request_id", rid, "handler", "UploadImage", "user_id", userID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "未选择图片"})
		return
	}

	applog.Logger.Info("upload: received file", "request_id", rid, "handler", "UploadImage", "user_id", userID,
		"filename", file.Filename, "size", file.Size, "content_type", file.Header.Get("Content-Type"), "client_ip", c.ClientIP())

	// 检查文件大小限制
	maxSizeKB := policy.MaximumFileSize
	if file.Size > int64(maxSizeKB*1024) {
		applog.Logger.Warn("upload: file too large", "request_id", rid, "handler", "UploadImage", "user_id", userID, "size", file.Size, "max_kb", maxSizeKB)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("文件大小超过限制（%dKB）", maxSizeKB)})
		return
	}

	// 检查文件类型
	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	isValidType := false
	for _, suffix := range policy.AcceptedFileSuffixes {
		if "."+suffix == fileExt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		applog.Logger.Warn("upload: bad file type", "request_id", rid, "handler", "UploadImage", "user_id", userID, "ext", fileExt)
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型"})
		return
	}

	// 检查用户存储容量
	if !checkUserCapacity(c, user, file.Size) {
		applog.Logger.Warn("upload: capacity exceeded", "request_id", rid, "handler", "UploadImage", "user_id", userID, "size", file.Size)
		return
	}

	userConfig, err := user.GetUserConfig()
	if err != nil {
		userConfig = model.GetDefaultUserConfig()
	}

	// 落盘前按内容 MD5 去重：重复图片只建引用记录，不再写入存储
	if md5Str, herr := hashMultipartFileMD5(file); herr == nil {
		if attemptDedupUpload(c, md5Str, userID.(uint), file.Filename, userConfig) {
			applog.Logger.Info("upload: dedup hit (pre-save)", "request_id", rid, "handler", "UploadImage", "user_id", userID, "md5", md5Str)
			return
		}
	}

	strategy, strategyConfigData, err := resolveUploadStrategy(model.DB, userID.(uint), userConfig, c.PostForm("strategy_id"))
	if err != nil {
		applog.Logger.Error("upload: resolve strategy failed", "request_id", rid, "handler", "UploadImage", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	applog.Logger.Debug("upload: strategy chosen", "request_id", rid, "handler", "UploadImage", "strategy_id", strategy.ID, "strategy_name", strategy.Name)

	storageConfig := convertStrategyToStorageConfig(strategyConfigData)
	imageInfo, err := utils.SaveImageWithStorage(file, userID.(uint), policy.PathNamingRule, policy.FileNamingRule, storageConfig)
	if err != nil {
		applog.Logger.Error("upload: save storage failed", "request_id", rid, "handler", "UploadImage", "user_id", userID, "strategy_id", strategy.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片失败: " + err.Error()})
		return
	}

	utilsConfig := &utils.StrategyConfig{
		URL:      strategyConfigData.URL,
		Root:     strategyConfigData.Root,
		Queries:  strategyConfigData.Queries,
		AuthType: strategyConfigData.AuthType,
	}

	groupID := model.EffectiveImageGroupID(model.DB)

	// 创建图片记录
	imageRecord := &model.Images{
		UserID:      userID.(uint),
		GroupID:     groupID,
		StrategyID:  strategy.ID,
		ImageKey:    imageInfo.MD5,
		Path:        imageInfo.Path,
		Name:        filepath.Base(imageInfo.Path),
		OriginName:  imageInfo.OriginName,
		Size:        int64(imageInfo.Size * 1024 * 1024), // 转换为字节
		Mimetype:    imageInfo.Mimetype,
		Extension:   imageInfo.Extension,
		Md5:         imageInfo.MD5,
		SHA1:        imageInfo.SHA1,
		Width:       imageInfo.Width,
		Height:      imageInfo.Height,
		Permissions: uint(userConfig.DefaultPermission),
		IsUnhealthy: 0,
		UploadIp:    c.ClientIP(),
	}

	// 根据用户配置设置图片权限
	model.SetImagePermissionFromUserConfig(imageRecord, userConfig)

	// 保存到数据库（含MD5去重）
	imageRecord, isDup, err := createImageRecordWithDedup(imageRecord)
	if err != nil {
		applog.Logger.Error("upload: db insert failed", "request_id", rid, "handler", "UploadImage", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片记录失败"})
		return
	}

	// 更新用户图片数量
	if err := model.IncrementUserImageCount(model.DB, userID.(uint)); err != nil {
		applog.Logger.Warn("upload: increment image count failed", "request_id", rid, "user_id", userID, "error", err)
	}

	pathname := normalizePathnameForLinks(imageRecord.Path)
	links := utils.GenerateImageLinksWithConfig(pathname, imageInfo.OriginName, utilsConfig)

	if isDup {
		applog.Logger.Info("upload: dedup hit", "request_id", rid, "handler", "UploadImage", "user_id", userID,
			"image_id", imageRecord.ID, "ref_image_id", *imageRecord.RefImageID, "md5", imageInfo.MD5)
	} else {
		applog.Logger.Info("upload: success", "request_id", rid, "handler", "UploadImage", "user_id", userID,
			"image_id", imageRecord.ID, "pathname", pathname, "md5", imageInfo.MD5, "strategy_id", strategy.ID)
	}

	// 构建响应数据
	response := &UploadImageResponse{
		Status:  true,
		Message: "上传成功",
		Data: &ImageData{
			ID:         imageRecord.ID,
			Pathname:   pathname,
			OriginName: imageInfo.OriginName,
			Size:       imageInfo.Size,
			Mimetype:   imageInfo.Mimetype,
			MD5:        imageInfo.MD5,
			SHA1:       imageInfo.SHA1,
			Links:      links,
		},
	}

	c.JSON(http.StatusOK, response)
}

// UploadImageFromURL 从URL上传图片
func UploadImageFromURL(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取用户信息
	user, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	policy, ok := readUserUploadPolicy(c, userID.(uint))
	if !ok {
		return
	}

	// 解析请求体
	var requestData struct {
		URL        string `json:"url" binding:"required"`
		StrategyID int    `json:"strategy_id"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误"})
		return
	}

	// 下载图片（带协议校验、超时与大小限制）
	imageData, contentType, err := fetchRemoteImage(requestData.URL, policy.MaximumFileSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取文件扩展名
	urlPath := requestData.URL
	ext := filepath.Ext(urlPath)
	if ext == "" {
		// 根据Content-Type推断扩展名
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}

	// 检查文件类型
	isValidType := false
	for _, suffix := range policy.AcceptedFileSuffixes {
		if "."+suffix == strings.ToLower(ext) {
			isValidType = true
			break
		}
	}
	if !isValidType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型"})
		return
	}

	filename := filepath.Base(urlPath)
	if filename == "" || filename == "." {
		filename = "image" + ext
	}

	// 检查用户存储容量
	if !checkUserCapacity(c, user, int64(len(imageData))) {
		return
	}

	userConfig, err := user.GetUserConfig()
	if err != nil {
		userConfig = model.GetDefaultUserConfig()
	}

	// 落盘前按内容 MD5 去重：重复图片只建引用记录，不再写入存储
	if attemptDedupUpload(c, hashBytesMD5(imageData), userID.(uint), filename, userConfig) {
		return
	}

	strategyIDRaw := ""
	if requestData.StrategyID > 0 {
		strategyIDRaw = strconv.Itoa(requestData.StrategyID)
	}
	strategy, strategyConfigData, err := resolveUploadStrategy(model.DB, userID.(uint), userConfig, strategyIDRaw)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	storageConfig := convertStrategyToStorageConfig(strategyConfigData)
	imageInfo, err := utils.SaveImageBytesWithStorage(imageData, filename, userID.(uint), policy.PathNamingRule, policy.FileNamingRule, storageConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片失败: " + err.Error()})
		return
	}

	utilsConfig := &utils.StrategyConfig{
		URL:      strategyConfigData.URL,
		Root:     strategyConfigData.Root,
		Queries:  strategyConfigData.Queries,
		AuthType: strategyConfigData.AuthType,
	}

	groupID := model.EffectiveImageGroupID(model.DB)

	// 创建图片记录
	imageRecord := &model.Images{
		UserID:      userID.(uint),
		GroupID:     groupID,
		StrategyID:  strategy.ID,
		ImageKey:    imageInfo.MD5,
		Path:        imageInfo.Path,
		Name:        filepath.Base(imageInfo.Path),
		OriginName:  imageInfo.OriginName,
		Size:        int64(imageInfo.Size * 1024 * 1024),
		Mimetype:    imageInfo.Mimetype,
		Extension:   imageInfo.Extension,
		Md5:         imageInfo.MD5,
		SHA1:        imageInfo.SHA1,
		Width:       imageInfo.Width,
		Height:      imageInfo.Height,
		Permissions: uint(userConfig.DefaultPermission),
		IsUnhealthy: 0,
		UploadIp:    c.ClientIP(),
	}

	// 根据用户配置设置图片权限
	model.SetImagePermissionFromUserConfig(imageRecord, userConfig)

	// 保存到数据库（含MD5去重）
	imageRecord, _, err = createImageRecordWithDedup(imageRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片记录失败"})
		return
	}

	// 更新用户图片数量
	model.IncrementUserImageCount(model.DB, userID.(uint))

	pathname := normalizePathnameForLinks(imageRecord.Path)
	links := utils.GenerateImageLinksWithConfig(pathname, imageInfo.OriginName, utilsConfig)

	response := &UploadImageResponse{
		Status:  true,
		Message: "上传成功",
		Data: &ImageData{
			ID:         imageRecord.ID,
			Pathname:   pathname,
			OriginName: imageInfo.OriginName,
			Size:       imageInfo.Size,
			Mimetype:   imageInfo.Mimetype,
			MD5:        imageInfo.MD5,
			SHA1:       imageInfo.SHA1,
			Links:      links,
		},
	}

	c.JSON(http.StatusOK, response)
}

// 获取制作用户仪表盘信息
func Dashboard(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取用户信息（用于验证用户存在性）
	_, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 获取该用户的图片数量
	var imageCount int64
	model.DB.Model(&model.Images{}).Where("user_id = ?", userID).Count(&imageCount)

	// 获取该用户今日上传的图片数量
	todayImageCount, err := model.GetTodayImageCountByUserID(model.DB, userID.(uint))
	if err != nil {
		todayImageCount = 0
	}

	// 获取该用户的总存储大小
	var totalSize int64
	model.DB.Model(&model.Images{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)

	// 计算已使用空间（MB）
	usedSizeMB := float64(totalSize) / (1024 * 1024)

	// 构建返回体数据
	response := gin.H{
		"status":  "success",
		"message": "获取成功",
		"data": gin.H{
			"dashboard": gin.H{
				"used_size_mb":       usedSizeMB,
				"image_count":        imageCount,
				"today_upload_count": todayImageCount,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUserConfig 更新用户配置
func UpdateUserConfig(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取用户信息
	user, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 解析请求体
	var requestData struct {
		Name               *string `json:"name"`
		Password           *string `json:"password"`
		DefaultStrategy    *int    `json:"default_strategy"`
		DefaultPermission  *int    `json:"default_permission"`
		PastedAction       *int    `json:"pasted_action"`
		IsAutoClearPreview *int    `json:"is_auto_clear_preview"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误"})
		return
	}

	// 更新用户基本信息（只更新提供的字段）
	if requestData.Name != nil {
		user.Name = *requestData.Name
	}

	// 更新密码（只有当提供了密码字段时才更新）
	if requestData.Password != nil && *requestData.Password != "" {
		hash, err := utils.HashPassword(*requestData.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		user.Password = hash
	}

	// 获取当前用户配置
	userConfig, err := user.GetUserConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户配置失败"})
		return
	}

	// 更新配置（只更新提供的字段）
	if requestData.DefaultStrategy != nil {
		userConfig.DefaultStrategy = *requestData.DefaultStrategy
	}
	if requestData.DefaultPermission != nil {
		userConfig.DefaultPermission = *requestData.DefaultPermission
	}
	if requestData.PastedAction != nil {
		userConfig.PastedAction = *requestData.PastedAction
	}
	if requestData.IsAutoClearPreview != nil {
		userConfig.IsAutoClearPreview = *requestData.IsAutoClearPreview
	}

	// 验证后台配置
	if err := userConfig.ValidateBackendConfig(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "后台配置验证失败: " + err.Error()})
		return
	}

	// 保存用户配置
	if err := user.SetUserConfig(userConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存用户配置失败"})
		return
	}

	// 更新数据库中的用户记录
	if err := model.UpdateUser(model.DB, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
		return
	}

	// 构建响应数据
	response := gin.H{
		"status":  true,
		"message": "更新用户信息成功",
		"data": gin.H{
			"name":                  user.Name,
			"default_strategy":      userConfig.DefaultStrategy,
			"default_permission":    userConfig.DefaultPermission,
			"pasted_action":         userConfig.PastedAction,
			"is_auto_clear_preview": userConfig.IsAutoClearPreview,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetUserConfig 获取用户配置
func GetUserConfig(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取用户信息
	user, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 获取用户配置
	userConfig, err := user.GetUserConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户配置失败"})
		return
	}

	// 构建响应数据
	response := gin.H{
		"status":  true,
		"message": "获取用户配置成功",
		"data": gin.H{
			"default_strategy":      userConfig.DefaultStrategy,
			"default_permission":    userConfig.DefaultPermission,
			"pasted_action":         userConfig.PastedAction,
			"is_auto_clear_preview": userConfig.IsAutoClearPreview,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetImages 获取用户图片列表（分页）
func GetImages(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 500 {
		limit = 500
	}

	// 分页获取用户的图片列表
	images, total, err := model.GetImagesByUserIDPaginated(model.DB, userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取图片列表失败"})
		return
	}

	// 构建响应数据
	var imageDataList []ImageData
	for i := range images {
		img := &images[i]
		st, err := model.GetStrategyForStoredImage(model.DB, img.StrategyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取存储策略失败"})
			return
		}
		utilsConfig, err := strategyUtilsConfigFromModel(st)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析策略配置失败"})
			return
		}

		pathname := normalizePathnameForLinks(img.Path)
		links := utils.GenerateImageLinksWithConfig(pathname, img.OriginName, utilsConfig)

		imageData := ImageData{
			ID:           img.ID,
			Pathname:     pathname,
			OriginName:   img.OriginName,
			Size:         float64(img.Size) / (1024 * 1024),
			SizeBytes:    img.Size,
			Mimetype:     img.Mimetype,
			MD5:          img.Md5,
			SHA1:         img.SHA1,
			Links:        links,
			StrategyID:   img.StrategyID,
			StrategyName: st.Name,
			Permission:   img.Permissions,
			CreatedAt:    img.CreatedAt,
		}
		imageDataList = append(imageDataList, imageData)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data":    imageDataList,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// isAdminRequest 判断当前请求是否来自管理员（由 JWTOrBearerAuthMiddleware 写入上下文）
func isAdminRequest(c *gin.Context) bool {
	if v, ok := c.Get("isAdmin"); ok {
		if n, ok := v.(uint); ok {
			return n == 1
		}
	}
	return false
}

// DeleteImage 删除图片
func DeleteImage(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取图片ID
	imageIDStr := c.Param("id")
	imageID, err := strconv.ParseUint(imageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图片ID"})
		return
	}

	// 获取图片信息
	image, err := model.GetImageByID(model.DB, uint(imageID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	// 检查权限（只能删除自己的图片，管理员可删除任意图片）
	if image.UserID != userID.(uint) && !isAdminRequest(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除此图片"})
		return
	}

	// 删除图片记录
	if err := model.DeleteImage(model.DB, uint(imageID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除图片失败"})
		return
	}

	// 减少图片所有者的图片数量
	if err := model.DecrementUserImageCount(model.DB, image.UserID); err != nil {
		// 这里只是记录错误，不影响删除成功
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "删除成功",
	})
}

// BatchDeleteImages 批量删除图片
func BatchDeleteImages(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 解析请求体
	var requestData struct {
		ImageIDs []uint `json:"image_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "请求参数格式错误",
		})
		return
	}

	if len(requestData.ImageIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "图片ID列表不能为空",
		})
		return
	}

	// 批量删除图片
	successCount := 0
	failedIDs := []uint{}

	for _, imageID := range requestData.ImageIDs {
		// 获取图片信息
		image, err := model.GetImageByID(model.DB, imageID)
		if err != nil {
			failedIDs = append(failedIDs, imageID)
			continue
		}

		// 检查权限（只能删除自己的图片，管理员可删除任意图片）
		if image.UserID != userID.(uint) && !isAdminRequest(c) {
			failedIDs = append(failedIDs, imageID)
			continue
		}

		// 删除图片记录
		if err := model.DeleteImage(model.DB, imageID); err != nil {
			failedIDs = append(failedIDs, imageID)
			continue
		}

		// 减少图片所有者的图片数量
		model.DecrementUserImageCount(model.DB, image.UserID)
		successCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": fmt.Sprintf("成功删除%d张图片", successCount),
		"data": gin.H{
			"success_count": successCount,
			"failed_ids":    failedIDs,
		},
	})
}

// RenameImage 重命名图片
func RenameImage(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 获取图片ID
	imageIDStr := c.Param("id")
	imageID, err := strconv.ParseUint(imageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "无效的图片ID",
		})
		return
	}

	// 解析请求体
	var requestData struct {
		NewName string `json:"new_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "请求参数格式错误",
		})
		return
	}

	// 获取图片信息
	image, err := model.GetImageByID(model.DB, uint(imageID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "图片不存在",
		})
		return
	}

	// 检查权限（只能重命名自己的图片，管理员可重命名任意图片）
	if image.UserID != userID.(uint) && !isAdminRequest(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "无权限重命名此图片",
		})
		return
	}

	// 更新图片名称
	image.OriginName = requestData.NewName
	if err := model.DB.Save(image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "重命名失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "重命名成功",
		"data": gin.H{
			"id":          image.ID,
			"origin_name": image.OriginName,
		},
	})
}

// GetUserGroupConfig 获取全局上传策略（原用户组配置字段，现由系统设置维护）
func GetUserGroupConfig(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	if _, err := model.GetUserByID(model.DB, userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	policy, err := model.GetGlobalUploadPolicy(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取全局上传策略失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data":    policy,
	})
}

// GetRateLimitInfo 获取用户频率限制信息
func GetRateLimitInfo(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	if _, err := model.GetUserByID(model.DB, userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	policy, err := model.GetGlobalUploadPolicy(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取全局上传策略失败"})
		return
	}

	rateLimitConfig := &utils.GroupConfig{
		LimitPerMinute: policy.LimitPerMinute,
		LimitPerHour:   policy.LimitPerHour,
		LimitPerDay:    policy.LimitPerDay,
		LimitPerWeek:   policy.LimitPerWeek,
		LimitPerMonth:  policy.LimitPerMonth,
	}

	// 获取频率限制信息
	rateLimitInfo := utils.GetRateLimitInfo(userID.(uint), rateLimitConfig)

	// 构建响应数据
	response := gin.H{
		"status":  true,
		"message": "获取频率限制信息成功",
		"data":    rateLimitInfo,
	}

	c.JSON(http.StatusOK, response)
}

// ResetUserRateLimit 重置用户频率限制（管理员功能）
func ResetUserRateLimit(c *gin.Context) {
	// 管理员权限已由中间件验证

	// 获取要重置的用户ID
	targetUserIDStr := c.Param("userID")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 重置频率限制
	if err := utils.ResetRateLimit(uint(targetUserID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置频率限制失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "重置用户频率限制成功",
		"data": gin.H{
			"user_id": targetUserID,
		},
	})
}

// GetRateLimitStats 获取频率限制统计信息（管理员功能）
func GetRateLimitStats(c *gin.Context) {
	// 管理员权限已由中间件验证

	// 获取统计信息
	stats := utils.GetRateLimitStats()

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取频率限制统计信息成功",
		"data":    stats,
	})
}

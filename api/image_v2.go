package api

import (
	"Slink/applog"
	"Slink/middleware"
	"Slink/model"
	"Slink/storage"
	"Slink/utils"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// localStorageDiskBasePath 仅根据 base_path 决定本地磁盘根目录，默认 static（相对程序工作目录）。
// root 是访问 URL 中的路径段（如 / 或 /static），绝不能当作磁盘根，否则在 Linux 上会变成在 / 下建 2026/ 等目录。
func localStorageDiskBasePath(sc *model.StrategyConfigData) string {
	bp := strings.TrimSpace(sc.BasePath)
	if bp == "" {
		return "static"
	}
	c := filepath.Clean(bp)
	if c == string(filepath.Separator) || c == "/" {
		return "static"
	}
	if c == `\\` || c == `\` {
		return "static"
	}
	return c
}

// UploadImageV2 使用存储策略上传图片
func UploadImageV2(c *gin.Context) {
	rid := middleware.RequestID(c)
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		applog.Logger.Warn("upload v2: unauthenticated", "request_id", rid, "handler", "UploadImageV2")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取用户信息
	user, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		applog.Logger.Error("upload v2: get user failed", "request_id", rid, "user_id", userID, "error", err)
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
		applog.Logger.Warn("upload v2: no file", "request_id", rid, "user_id", userID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "未选择图片"})
		return
	}

	applog.Logger.Info("upload v2: received file", "request_id", rid, "user_id", userID,
		"filename", file.Filename, "size", file.Size, "client_ip", c.ClientIP())

	// 检查文件大小限制
	maxSizeKB := policy.MaximumFileSize
	if file.Size > int64(maxSizeKB*1024) {
		applog.Logger.Warn("upload v2: file too large", "request_id", rid, "user_id", userID, "size", file.Size, "max_kb", maxSizeKB)
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
		applog.Logger.Warn("upload v2: bad file type", "request_id", rid, "user_id", userID, "ext", fileExt)
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型"})
		return
	}

	// 检查用户存储容量
	if !checkUserCapacity(c, user, file.Size) {
		applog.Logger.Warn("upload v2: capacity exceeded", "request_id", rid, "user_id", userID, "size", file.Size)
		return
	}

	userConfig, err := user.GetUserConfig()
	if err != nil {
		userConfig = model.GetDefaultUserConfig()
	}

	// 落盘前按内容 MD5 去重：重复图片只建引用记录，不再写入存储
	if md5Str, herr := hashMultipartFileMD5(file); herr == nil {
		if attemptDedupUpload(c, md5Str, userID.(uint), file.Filename, userConfig) {
			applog.Logger.Info("upload v2: dedup hit (pre-save)", "request_id", rid, "user_id", userID, "md5", md5Str)
			return
		}
	}

	strategy, strategyConfigData, err := resolveUploadStrategy(model.DB, userID.(uint), userConfig, c.PostForm("strategy_id"))
	if err != nil {
		applog.Logger.Error("upload v2: resolve strategy failed", "request_id", rid, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	applog.Logger.Debug("upload v2: strategy", "request_id", rid, "strategy_id", strategy.ID, "strategy_name", strategy.Name)

	storageConfig := convertStrategyToStorageConfig(strategyConfigData)

	// 使用存储策略保存图片
	imageInfo, err := utils.SaveImageWithStorage(file, userID.(uint), policy.PathNamingRule, policy.FileNamingRule, storageConfig)
	if err != nil {
		applog.Logger.Error("upload v2: save storage failed", "request_id", rid, "user_id", userID, "strategy_id", strategy.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片失败: " + err.Error()})
		return
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
	imageRecord, isDup, err := createImageRecordWithDedup(imageRecord)
	if err != nil {
		applog.Logger.Error("upload v2: db insert failed", "request_id", rid, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片记录失败"})
		return
	}

	// 更新用户图片数量
	if err := model.IncrementUserImageCount(model.DB, userID.(uint)); err != nil {
		applog.Logger.Warn("upload v2: increment image count failed", "request_id", rid, "user_id", userID, "error", err)
	}

	// 生成图片链接
	utilsConfig := &utils.StrategyConfig{
		URL:      strategyConfigData.URL,
		Root:     strategyConfigData.Root,
		Queries:  strategyConfigData.Queries,
		AuthType: strategyConfigData.AuthType,
	}

	pathname := normalizePathnameForLinks(imageRecord.Path)
	links := utils.GenerateImageLinksWithConfig(pathname, imageInfo.OriginName, utilsConfig)

	if isDup {
		applog.Logger.Info("upload v2: dedup hit", "request_id", rid, "user_id", userID,
			"image_id", imageRecord.ID, "ref_image_id", *imageRecord.RefImageID, "md5", imageInfo.MD5)
	} else {
		applog.Logger.Info("upload v2: success", "request_id", rid, "user_id", userID,
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

// UploadImageFromURLV2 从URL上传图片（使用存储策略）
func UploadImageFromURLV2(c *gin.Context) {
	rid := middleware.RequestID(c)
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
		applog.Logger.Warn("upload url v2: bad json", "request_id", rid, "user_id", userID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误"})
		return
	}

	applog.Logger.Info("upload url v2: start", "request_id", rid, "user_id", userID, "remote_url", requestData.URL, "strategy_id", requestData.StrategyID)

	// 下载图片（带协议校验、超时与大小限制）
	imageData, contentType, err := fetchRemoteImage(requestData.URL, policy.MaximumFileSize)
	if err != nil {
		applog.Logger.Warn("upload url v2: fetch failed", "request_id", rid, "user_id", userID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取文件扩展名
	urlPath := requestData.URL
	ext := filepath.Ext(urlPath)
	if ext == "" {
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

	// 生成文件名
	filename := filepath.Base(urlPath)
	if filename == "" || filename == "." {
		filename = "image" + ext
	}

	// 检查用户存储容量
	if !checkUserCapacity(c, user, int64(len(imageData))) {
		applog.Logger.Warn("upload url v2: capacity exceeded", "request_id", rid, "user_id", userID)
		return
	}

	userConfig, err := user.GetUserConfig()
	if err != nil {
		userConfig = model.GetDefaultUserConfig()
	}

	// 落盘前按内容 MD5 去重：重复图片只建引用记录，不再写入存储
	if attemptDedupUpload(c, hashBytesMD5(imageData), userID.(uint), filename, userConfig) {
		applog.Logger.Info("upload url v2: dedup hit (pre-save)", "request_id", rid, "user_id", userID)
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

	// 使用存储策略保存图片
	imageInfo, err := utils.SaveImageBytesWithStorage(imageData, filename, userID.(uint), policy.PathNamingRule, policy.FileNamingRule, storageConfig)
	if err != nil {
		applog.Logger.Error("upload url v2: save storage failed", "request_id", rid, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片失败: " + err.Error()})
		return
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

	model.SetImagePermissionFromUserConfig(imageRecord, userConfig)

	// 保存到数据库（含MD5去重）
	imageRecord, _, err = createImageRecordWithDedup(imageRecord)
	if err != nil {
		applog.Logger.Error("upload url v2: db insert failed", "request_id", rid, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片记录失败"})
		return
	}

	if err := model.IncrementUserImageCount(model.DB, userID.(uint)); err != nil {
		applog.Logger.Warn("upload url v2: increment image count failed", "request_id", rid, "user_id", userID, "error", err)
	}

	// 生成图片链接
	utilsConfig := &utils.StrategyConfig{
		URL:      strategyConfigData.URL,
		Root:     strategyConfigData.Root,
		Queries:  strategyConfigData.Queries,
		AuthType: strategyConfigData.AuthType,
	}

	pathname := normalizePathnameForLinks(imageRecord.Path)
	links := utils.GenerateImageLinksWithConfig(pathname, imageInfo.OriginName, utilsConfig)

	applog.Logger.Info("upload url v2: success", "request_id", rid, "user_id", userID,
		"image_id", imageRecord.ID, "pathname", pathname, "md5", imageInfo.MD5, "strategy_id", strategy.ID)

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

// convertStrategyToStorageConfig 转换策略配置为存储配置
func convertStrategyToStorageConfig(strategyConfig *model.StrategyConfigData) *storage.Config {
	config := &storage.Config{
		Type:      strategyConfig.StorageType,
		Endpoint:  strategyConfig.Endpoint,
		AccessKey: strategyConfig.AccessKey,
		SecretKey: strategyConfig.SecretKey,
		Bucket:    strategyConfig.Bucket,
		Region:    strategyConfig.Region,
		Domain:    strategyConfig.Domain,
		BasePath:  strategyConfig.BasePath,
		UseSSL:    strategyConfig.UseSSL,
		Extra:     make(map[string]string),
	}

	// 如果没有指定存储类型，默认使用本地存储
	if config.Type == "" {
		config.Type = "local"
	}

	if config.Type == "local" {
		config.BasePath = localStorageDiskBasePath(strategyConfig)
	} else if config.BasePath == "" {
		if strategyConfig.Root != "" {
			config.BasePath = strategyConfig.Root
		} else {
			config.BasePath = "static"
		}
	}

	// 如果没有指定域名，使用URL
	if config.Domain == "" {
		config.Domain = strategyConfig.URL
	}

	return config
}

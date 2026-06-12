package api

import (
	"Slink/applog"
	"Slink/middleware"
	"Slink/model"
	"Slink/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// LoginRequest 登录：优先 username，兼容旧字段 email（均表示登录账号，非必须为邮箱）
type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required,max=128"`
}

func (r *LoginRequest) loginAccount() string {
	s := strings.TrimSpace(r.Username)
	if s != "" {
		return s
	}
	return strings.TrimSpace(r.Email)
}

// UserListResponse 用户列表响应结构（不包含敏感信息）
type UserListResponse struct {
	ID           uint    `json:"id"`
	Account      string  `json:"account"`
	Name         string  `json:"name"`
	IsAdmin      bool    `json:"is_admin"`
	ImageNums    uint    `json:"image_nums"`
	Capacity     uint    `json:"capacity"`
	UsedSize     float64 `json:"used_size"`
	RegisteredIP string  `json:"registered_ip"`
	CreatedAt    string  `json:"created_at"`
}

// StrategyInfo 存储策略信息结构
type StrategyInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Introduction string `json:"introduction"`
	Key          string `json:"key"`
}

// Register 已废弃：个人图床仅在首次初始化时创建唯一账号，不提供公开注册。
func Register(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"status":  false,
		"message": "个人图床不支持公开注册，请使用初始化时设置的账号登录",
	})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		applog.Logger.Warn("login bind failed", "request_id", middleware.RequestID(c), "client_ip", c.ClientIP(), "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}
	acc := req.loginAccount()
	if acc == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "请填写登录账号",
			"error":   "缺少 username",
		})
		return
	}
	applog.Logger.Info("login attempt", "request_id", middleware.RequestID(c), "account", acc, "client_ip", c.ClientIP())
	var user model.User
	if err := model.DB.Where("email = ?", acc).First(&user).Error; err != nil {
		applog.Logger.Warn("login failed: user not found", "request_id", middleware.RequestID(c), "account", acc)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户不存在",
			"error":   "用户不存在",
		})
		return
	}
	if !utils.CheckPassword(user.Password, req.Password) {
		applog.Logger.Warn("login failed: bad password", "request_id", middleware.RequestID(c), "user_id", user.ID, "account", acc)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "密码错误",
			"error":   "密码错误",
		})
		return
	}
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		applog.Logger.Error("login token error", "request_id", middleware.RequestID(c), "user_id", user.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Token生成失败",
			"error":   "Token生成失败",
		})
		return
	}

	applog.Logger.Info("login ok", "request_id", middleware.RequestID(c), "user_id", user.ID, "account", user.Email, "is_admin", user.IsAdmin == 1)

	// 返回登录成功信息，包含token和用户信息
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "登录成功",
		"data": gin.H{
			"token":    token,
			"user_id":  user.ID,
			"account":  user.Email,
			"name":     user.Name,
			"is_admin": user.IsAdmin == 1,
		},
	})
}

// GetUsers 获取所有用户（管理员功能）
func GetUsers(c *gin.Context) {
	// 管理员权限已由中间件验证

	users, err := model.GetAllUsers(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	// 构建不包含敏感信息的用户列表
	var userResponses []UserListResponse
	for _, user := range users {
		//获取用户的所有存储图片的大小
		// 获取该策略的总存储大小
		var totalSize int64
		model.DB.Model(&model.Images{}).Where("user_id = ?", user.ID).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)

		// 计算MB
		usedSizeMB := float64(totalSize) / (1024 * 1024)
		userResponse := UserListResponse{
			ID:           user.ID,
			Account:      user.Email,
			Name:         user.Name,
			IsAdmin:      user.IsAdmin == 1,
			ImageNums:    user.ImageNums,
			Capacity:     user.Capacity,
			UsedSize:     usedSizeMB,
			RegisteredIP: user.RegisteredIP,
			CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		userResponses = append(userResponses, userResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data":    userResponses,
	})
}

// GetUser 获取指定用户（管理员功能）
func GetUser(c *gin.Context) {
	// 管理员权限已由中间件验证

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	user, err := model.GetUserByID(model.DB, uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 构建不包含敏感信息的用户信息
	userResponse := UserListResponse{
		ID:           user.ID,
		Account:      user.Email,
		Name:         user.Name,
		IsAdmin:      user.IsAdmin == 1,
		ImageNums:    user.ImageNums,
		Capacity:     user.Capacity,
		RegisteredIP: user.RegisteredIP,
		CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data":    userResponse,
	})
}

// UpdateUser 更新用户信息（管理员功能）
func UpdateUser(c *gin.Context) {
	// 管理员权限已由中间件验证

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	existing, err := model.GetUserByID(model.DB, uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	var req struct {
		Name     string `json:"name"`
		Account  string `json:"account"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Capacity uint   `json:"capacity"`
		IsAdmin  int    `json:"is_admin"`
		GroupID  *uint  `json:"group_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if a := strings.TrimSpace(req.Account); a != "" {
		existing.Email = a
	} else if e := strings.TrimSpace(req.Email); e != "" {
		existing.Email = e
	}
	existing.Capacity = req.Capacity
	if req.IsAdmin == 0 || req.IsAdmin == 1 {
		existing.IsAdmin = uint(req.IsAdmin)
	}
	if req.GroupID != nil {
		existing.GroupID = *req.GroupID
	}
	if req.Password != "" {
		hash, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码处理失败"})
			return
		}
		existing.Password = hash
	}

	if err := model.UpdateUser(model.DB, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "更新成功",
		"data":    existing,
	})
}

// DeleteUser 删除用户（管理员功能）
func DeleteUser(c *gin.Context) {
	// 管理员权限已由中间件验证

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 检查是否为管理员用户
	user, err := model.GetUserByID(model.DB, uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if user.IsAdmin == 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能删除管理员用户"})
		return
	}

	if err := model.DeleteUser(model.DB, uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "删除成功",
	})
}

// GetUserImages 获取指定用户的图片列表（管理员功能）
func GetUserImages(c *gin.Context) {
	// 管理员权限已由中间件验证

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 获取用户的图片列表
	images, err := model.GetImagesByUserID(model.DB, uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取图片列表失败"})
		return
	}

	// 构建响应数据
	var imageDataList []map[string]interface{}
	for _, img := range images {
		utilsConfig, err := strategyUtilsConfigForStoredImage(model.DB, &img)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取存储策略失败"})
			return
		}
		pathname := normalizePathnameForLinks(img.Path)
		links := utils.GenerateImageLinksWithConfig(pathname, img.OriginName, utilsConfig)

		imageData := map[string]interface{}{
			"id":          img.ID,
			"pathname":    pathname,
			"origin_name": img.OriginName,
			"size":        img.Size,
			"mimetype":    img.Mimetype,
			"md5":         img.Md5,
			"sha1":        img.SHA1,
			"permission":  img.Permissions,
			"created_at":  img.CreatedAt,
			"links":       links,
		}
		imageDataList = append(imageDataList, imageData)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data":    imageDataList,
	})
}

// GetUserStats 获取用户统计信息（管理员功能）
func GetUserStats(c *gin.Context) {
	// 管理员权限已由中间件验证

	// 获取用户总数
	totalUsers, err := model.GetUserCount(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户总数失败"})
		return
	}

	// 获取管理员用户数
	adminUsers, err := model.GetAdminUsers(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取管理员用户数失败"})
		return
	}

	stats := gin.H{
		"total_users":   totalUsers,
		"admin_users":   len(adminUsers),
		"regular_users": totalUsers - int64(len(adminUsers)),
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取统计信息成功",
		"data":    stats,
	})
}

// GetCurrentUser 获取当前用户信息
func GetCurrentUser(c *gin.Context) {
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
		// 如果获取配置失败，使用默认配置
		userConfig = model.GetDefaultUserConfig()
	}

	// 个人图床：可使用系统中配置的全部存储策略
	allStrategies, err := model.GetAllStrategies(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取存储策略失败"})
		return
	}
	strategies := make([]StrategyInfo, 0, len(allStrategies))
	for _, strategy := range allStrategies {
		strategies = append(strategies, StrategyInfo{
			ID:           strategy.ID,
			Name:         strategy.Name,
			Introduction: strategy.Introduction,
			Key:          strategy.StrategyKey,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data": gin.H{
			"user": gin.H{
				"id":            user.ID,
				"account":       user.Email,
				"name":          user.Name,
				"is_admin":      user.IsAdmin == 1,
				"image_nums":    user.ImageNums,
				"created_at":    user.CreatedAt,
				"registered_ip": user.RegisteredIP,
			},
			"strategies": strategies,
			"config": gin.H{
				"default_strategy":      userConfig.DefaultStrategy,
				"default_permission":    userConfig.DefaultPermission,
				"pasted_action":         userConfig.PastedAction,
				"is_auto_clear_preview": userConfig.IsAutoClearPreview,
			},
		},
	})
}

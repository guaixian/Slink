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
	ID            uint    `json:"id"`
	Account       string  `json:"account"`
	Name          string  `json:"name"`
	IsAdmin       bool    `json:"is_admin"`
	ImageNums     uint    `json:"image_nums"`
	Capacity      uint    `json:"capacity"`
	UsedSize      float64 `json:"used_size"`
	PolicyGroupID uint    `json:"policy_group_id"`
	RegisteredIP  string  `json:"registered_ip"`
	CreatedAt     string  `json:"created_at"`
}

// StrategyInfo 存储策略信息结构
type StrategyInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Introduction string `json:"introduction"`
	Key          string `json:"key"`
}

// Register 用户注册（需要系统开启注册开关）
func Register(c *gin.Context) {
	// 检查是否开启注册
	enableRegister, _ := model.GetConfigValue(model.DB, "enable_register")
	if enableRegister != "true" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "系统未开启公开注册",
		})
		return
	}

	var req struct {
		Account  string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查账号是否已存在
	existing, _ := model.GetUserByEmail(model.DB, req.Account)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  false,
			"message": "该账号已被注册",
		})
		return
	}

	hash, _ := utils.HashPassword(req.Password)
	defaultConfig := model.GetDefaultUserConfig()
	configJSON, _ := defaultConfig.ToJSON()

	name := req.Name
	if name == "" {
		name = req.Account
	}

	// 获取默认上传策略组ID
	var defaultPolicyGroupID uint
	if pg, err := model.GetDefaultUploadPolicyGroup(model.DB); err == nil {
		defaultPolicyGroupID = pg.ID
	}

	user := model.User{
		Email:         req.Account,
		Name:          name,
		Password:      hash,
		GroupID:       1,
		PolicyGroupID: defaultPolicyGroupID,
		IsAdmin:       0,
		// 应用系统设置中的"新注册用户默认存储空间"(字节,0 表示无限制)
		Capacity:     uint(model.GetDefaultUserCapacityBytes(model.DB)),
		Configs:      configJSON,
		RegisteredIP: c.ClientIP(),
	}

	if err := model.CreateUser(model.DB, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "注册失败",
		})
		return
	}

	// 生成 token 并自动登录
	token, err := middleware.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "注册成功但登录失败，请手动登录",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "注册成功",
		"data": gin.H{
			"token":    token,
			"user_id":  user.ID,
			"account":  user.Email,
			"name":     user.Name,
			"is_admin": false,
		},
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

// CreateUser 创建用户（管理员功能）
func CreateUser(c *gin.Context) {
	var req struct {
		Account  string `json:"account" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Capacity uint   `json:"capacity"`
		IsAdmin  int    `json:"is_admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "参数错误: " + err.Error()})
		return
	}

	// 检查账号是否已存在
	existing, _ := model.GetUserByEmail(model.DB, req.Account)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"status": false, "message": "该账号已存在"})
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "密码处理失败"})
		return
	}

	defaultConfig := model.GetDefaultUserConfig()
	configJSON, _ := defaultConfig.ToJSON()

	// 获取默认上传策略组
	var defaultPolicyGroupID uint
	if pg, err := model.GetDefaultUploadPolicyGroup(model.DB); err == nil {
		defaultPolicyGroupID = pg.ID
	}

	user := model.User{
		Email:         req.Account,
		Name:          req.Name,
		Password:      hash,
		GroupID:       1,
		PolicyGroupID: defaultPolicyGroupID,
		IsAdmin:       uint(req.IsAdmin),
		Capacity:      req.Capacity,
		Configs:       configJSON,
		RegisteredIP:  c.ClientIP(),
	}

	if err := model.CreateUser(model.DB, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "创建用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "创建用户成功",
		"data": gin.H{
			"id":       user.ID,
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
			ID:            user.ID,
			Account:       user.Email,
			Name:          user.Name,
			IsAdmin:       user.IsAdmin == 1,
			ImageNums:     user.ImageNums,
			Capacity:      user.Capacity,
			UsedSize:      usedSizeMB,
			PolicyGroupID: user.PolicyGroupID,
			RegisteredIP:  user.RegisteredIP,
			CreatedAt:     user.CreatedAt.Format("2006-01-02 15:04:05"),
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
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data": gin.H{
			"id":              user.ID,
			"account":         user.Email,
			"name":            user.Name,
			"is_admin":        user.IsAdmin == 1,
			"image_nums":      user.ImageNums,
			"capacity":        user.Capacity,
			"policy_group_id": user.PolicyGroupID,
			"registered_ip":   user.RegisteredIP,
			"created_at":      user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
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
		Capacity *uint  `json:"capacity"`
		IsAdmin  *int   `json:"is_admin"`
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
	// 仅当请求显式携带 capacity 时才更新，避免缺省值 0 清空容量限制
	if req.Capacity != nil {
		existing.Capacity = *req.Capacity
	}
	// 仅当请求显式携带 is_admin 时才更新，避免缺省值 0 把管理员降级
	if req.IsAdmin != nil {
		// 不允许管理员取消自己的管理员权限（防止系统失去管理员）
		if currentID, ok := c.Get("userID"); ok && currentID.(uint) == existing.ID && *req.IsAdmin == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "不能取消自己的管理员权限"})
			return
		}
		if *req.IsAdmin == 0 || *req.IsAdmin == 1 {
			existing.IsAdmin = uint(*req.IsAdmin)
		}
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
		// 不回传完整 User 结构，避免泄露密码哈希
		"data": gin.H{
			"id":              existing.ID,
			"account":         existing.Email,
			"name":            existing.Name,
			"is_admin":        existing.IsAdmin == 1,
			"capacity":        existing.Capacity,
			"policy_group_id": existing.PolicyGroupID,
		},
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

	// 清理关联数据：访问令牌、分享、策略分配（图片与文件保留，可由图片管理另行处理）
	model.DeletePersonalAccessTokensByUsername(model.DB, user.Email)
	model.DB.Where("user_id = ?", user.ID).Delete(&model.Share{})
	model.DeleteUserStrategiesByUserID(model.DB, user.ID)

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

	// 获取用户可用的存储策略：管理员全部，普通用户只看已分配的
	var visibleStrategies []model.Strategies
	if user.IsAdmin == 1 {
		visibleStrategies, err = model.GetAllStrategies(model.DB)
	} else {
		visibleStrategies, err = model.GetUserStrategies(model.DB, user.ID)
		if err != nil || len(visibleStrategies) == 0 {
			// 如果用户没有分配任何策略，回退到全部策略
			visibleStrategies, err = model.GetAllStrategies(model.DB)
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取存储策略失败"})
		return
	}
	strategies := make([]StrategyInfo, 0, len(visibleStrategies))
	for _, strategy := range visibleStrategies {
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
				"id":              user.ID,
				"account":         user.Email,
				"name":            user.Name,
				"is_admin":        user.IsAdmin == 1,
				"image_nums":      user.ImageNums,
				"capacity":        user.Capacity,
				"policy_group_id": user.PolicyGroupID,
				"created_at":      user.CreatedAt,
				"registered_ip":   user.RegisteredIP,
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

package main

import (
	"Slink/api"
	"Slink/middleware"
	"Slink/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置所有路由
// 用于初始化Gin引擎并注册所有API端点
func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	// 配置静态文件服务
	configureStaticFiles(r)

	// 配置路由组
	apiGroup := r.Group("/api")
	{
		// 初始化相关接口（无需登录）
		initGroup := apiGroup.Group("/init")
		{
			initGroup.GET("/status", api.GetInitStatus)
			initGroup.POST("/setup", api.PerformSetup)
			initGroup.POST("/test-db", api.TestDatabaseConnection)
		}

		// 基础配置接口
		apiGroup.GET("/base_config", getBaseConfig)

		// 用户认证（个人图床仅允许初始化时创建的唯一账号登录；注册接口返回明确拒绝）
		apiGroup.POST("/register", api.Register)
		apiGroup.POST("/login", api.Login)

		// 图片管理接口（需要JWT认证）
		configureImageRoutes(apiGroup)

		// 用户相关接口（需要JWT认证）
		configureUserRoutes(apiGroup)

		// 访问令牌接口（需要JWT认证）
		configureTokenRoutes(apiGroup)

		// 分享接口（需要JWT认证）
		configureShareRoutes(apiGroup)

		// 公开分享接口
		apiGroup.POST("/share/:code", api.GetShare)

		// 管理员接口（需要JWT和管理员权限）
		configureAdminRoutes(apiGroup)

		// 存储策略接口（需要JWT和管理员权限）
		configureStorageRoutes(apiGroup)
	}

	// 配置前端路由
	configureFrontendRoutes(r)

	return r
}

// configureStaticFiles 配置静态文件服务
// 包括favicon和受保护的图片资源
func configureStaticFiles(r *gin.Engine) {
	r.GET("/static/*filepath", api.ServeProtectedImage)
}

// configureImageRoutes 配置图片相关路由
// 包括上传、列表、删除、修改等操作
func configureImageRoutes(apiGroup *gin.RouterGroup) {
	imgGroup := apiGroup.Group("/image")
	imgGroup.Use(middleware.JWTOrBearerAuthMiddleware())
	{
		imgGroup.POST("/upload", api.UploadImage)
		imgGroup.POST("/upload-v2", api.UploadImageV2)
		imgGroup.POST("/upload-url", api.UploadImageFromURL)
		imgGroup.POST("/upload-url-v2", api.UploadImageFromURLV2)
		imgGroup.GET("/list", api.GetImages)
		imgGroup.DELETE("/:id", api.DeleteImage)
		imgGroup.POST("/batch-delete", api.BatchDeleteImages)
		imgGroup.PUT("/:id/rename", api.RenameImage)
		imgGroup.GET("/:id/qrcode", api.GenerateQRCode)
		imgGroup.GET("/:id/qrcode-base64", api.GenerateQRCodeBase64)
		imgGroup.GET("/config", api.GetUserGroupConfig)
		imgGroup.GET("/rate-limit", api.GetRateLimitInfo)

		// 本地图片处理（纯Go，无外部依赖）：压缩/格式转换/高质量缩放放大/缩略图
		imgGroup.POST("/process", api.ProcessImage)
		imgGroup.GET("/process/capabilities", api.GetProcessCapabilities)
	}
}

// configureUserRoutes 配置用户相关路由
// 包括用户信息、配置、仪表盘等
func configureUserRoutes(apiGroup *gin.RouterGroup) {
	userGroup := apiGroup.Group("/user")
	userGroup.Use(middleware.JWTAuthMiddleware())
	{
		userGroup.GET("/info", api.GetCurrentUser)
		userGroup.GET("/config", api.GetUserConfig)
		userGroup.PUT("/info", api.UpdateUserConfig)
		userGroup.GET("/dashboard", api.Dashboard)
	}
}

// configureTokenRoutes 配置访问令牌相关路由
// 用于管理和操作用户的个人访问令牌
func configureTokenRoutes(apiGroup *gin.RouterGroup) {
	tokenGroup := apiGroup.Group("/tokens")
	tokenGroup.Use(middleware.JWTAuthMiddleware())
	{
		tokenGroup.POST("/", api.CreateToken)
		tokenGroup.GET("/", api.GetTokens)
		tokenGroup.DELETE("/:id", api.DeleteToken)
	}
}

// configureShareRoutes 配置分享相关路由
// 用于创建和管理图片分享链接
func configureShareRoutes(apiGroup *gin.RouterGroup) {
	shareGroup := apiGroup.Group("/shares")
	shareGroup.Use(middleware.JWTAuthMiddleware())
	{
		shareGroup.POST("/", api.CreateShare)
		shareGroup.GET("/my", api.GetMyShares)
		shareGroup.DELETE("/:id", api.DeleteShareByID)
		shareGroup.PUT("/:id/status", api.UpdateShareStatus)
	}
}

// configureAdminRoutes 配置管理员相关路由
// 用于系统管理和用户管理
func configureAdminRoutes(apiGroup *gin.RouterGroup) {
	adminGroup := apiGroup.Group("/admin")
	adminGroup.Use(middleware.JWTAuthMiddleware())
	adminGroup.Use(middleware.AdminAuthMiddleware())
	{
		// 系统配置管理
		adminGroup.GET("/configs", api.GetConfigs)
		adminGroup.PUT("/configs", api.UpdateConfigs)

		adminGroup.GET("/upload-policy", api.GetUploadPolicy)
		adminGroup.PUT("/upload-policy", api.UpdateUploadPolicy)

		// 全量备份（config + sqlite + static）
		adminGroup.GET("/backup/export", api.ExportFullBackup)
		adminGroup.POST("/backup/import", api.ImportFullBackup)

		// 系统统计信息
		adminGroup.GET("/stats", api.GetSystemStats)

		// 多用户/用户组相关接口已移除（个人图床）
	}
}

// configureStorageRoutes 配置存储策略相关路由
// 用于管理不同的存储后端策略
func configureStorageRoutes(apiGroup *gin.RouterGroup) {
	storage := apiGroup.Group("/storages")
	storage.Use(middleware.JWTAuthMiddleware())
	storage.Use(middleware.AdminAuthMiddleware())
	{
		storage.GET("/", api.Storage)
		storage.GET("/strategies", api.GetStrategies)
		storage.POST("/strategies", api.CreateStrategy)
		storage.PUT("/strategies/:id", api.UpdateStrategy)
		storage.GET("/strategies/:id", api.GetStrategyByID)
		storage.DELETE("/strategies/:id", api.DeleteStrategy)
	}
}

// configureFrontendRoutes 配置前端路由
// 处理前端路由和404页面
func configureFrontendRoutes(r *gin.Engine) {
	// 静态资源
	r.Static("/assets", "SlinkWeb/dist/assets")

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{
				"code":    404,
				"message": "接口不存在",
				"data":    nil,
			})
			return
		}
		c.File("SlinkWeb/dist/index.html")
	})
}

// getBaseConfig 获取基础配置
// 返回系统的所有配置项
func getBaseConfig(c *gin.Context) {
	res, err := model.GetAllConfigs(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"msg":     "读取配置失败",
			"data":    nil,
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": res,
	})
}

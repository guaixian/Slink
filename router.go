package main

import (
	"Slink/api"
	"Slink/middleware"
	"Slink/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thinkerou/favicon"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	// 以下中间件必须在注册任何路由之前 Use，否则对已注册路由不生效
	//（gin 的中间件链在路由注册时固化，事后 Use 不会回溯）
	r.Use(favicon.New("SlinkWeb/dist/favicon.ico"))
	configureCORSMiddleware(r)
	r.Use(InitGuard())
	configureStaticFiles(r)

	apiGroup := r.Group("/api")
	{
		initGroup := apiGroup.Group("/init")
		{
			initGroup.GET("/status", api.GetInitStatus)
			initGroup.POST("/setup", api.PerformSetup)
			initGroup.POST("/test-db", api.TestDatabaseConnection)
			initGroup.POST("/test-redis", api.TestRedisConnection)
		}
		apiGroup.GET("/base_config", getBaseConfig)
		apiGroup.POST("/register", api.Register)
		apiGroup.POST("/login", api.Login)

		configureImageRoutes(apiGroup)
		configureUserRoutes(apiGroup)
		configureTokenRoutes(apiGroup)
		configureShareRoutes(apiGroup)
		apiGroup.POST("/share/:code", api.GetShare)
		configureAdminRoutes(apiGroup)
		configureStorageRoutes(apiGroup)
	}
	configureFrontendRoutes(r)
	return r
}

func configureStaticFiles(r *gin.Engine) {
	r.GET("/static/*filepath", api.ServeProtectedImage)
}

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
	}
}

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

func configureTokenRoutes(apiGroup *gin.RouterGroup) {
	tokenGroup := apiGroup.Group("/tokens")
	tokenGroup.Use(middleware.JWTAuthMiddleware())
	{
		tokenGroup.POST("/", api.CreateToken)
		tokenGroup.GET("/", api.GetTokens)
		tokenGroup.DELETE("/:id", api.DeleteToken)
	}
}

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

func configureAdminRoutes(apiGroup *gin.RouterGroup) {
	adminGroup := apiGroup.Group("/admin")
	adminGroup.Use(middleware.JWTAuthMiddleware())
	adminGroup.Use(middleware.AdminAuthMiddleware())
	{
		adminGroup.GET("/configs", api.GetConfigs)
		adminGroup.PUT("/configs", api.UpdateConfigs)
		adminGroup.GET("/upload-policy", api.GetUploadPolicy)
		adminGroup.PUT("/upload-policy", api.UpdateUploadPolicy)

		adminGroup.GET("/policy-groups", api.GetUploadPolicyGroups)
		adminGroup.POST("/policy-groups", api.CreateUploadPolicyGroup)
		adminGroup.GET("/policy-groups/:id", api.GetUploadPolicyGroup)
		adminGroup.PUT("/policy-groups/:id", api.UpdateUploadPolicyGroup)
		adminGroup.DELETE("/policy-groups/:id", api.DeleteUploadPolicyGroup)
		adminGroup.PUT("/policy-groups/:id/set-default", api.SetDefaultPolicyGroup)

		adminGroup.GET("/backup/export", api.ExportFullBackup)
		adminGroup.POST("/backup/import", api.ImportFullBackup)
		adminGroup.GET("/stats", api.GetSystemStats)
		adminGroup.GET("/images", api.AdminGetImages)

		adminGroup.GET("/users", api.GetUsers)
		adminGroup.POST("/users", api.CreateUser)
		adminGroup.GET("/users/:id", api.GetUser)
		adminGroup.PUT("/users/:id", api.UpdateUser)
		adminGroup.DELETE("/users/:id", api.DeleteUser)
		adminGroup.GET("/users/:id/images", api.GetUserImages)
		adminGroup.PUT("/users/:id/policy-group", api.SetUserPolicyGroup)
		adminGroup.GET("/users/:id/strategies", api.GetUserStrategies)
		adminGroup.POST("/users/:id/strategies", api.AssignStrategiesToUser)
		adminGroup.DELETE("/users/:id/strategies/:strategyId", api.RemoveUserStrategy)
	}
}

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

func configureFrontendRoutes(r *gin.Engine) {
	r.Static("/assets", "SlinkWeb/dist/assets")
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{"code": 404, "message": "接口不存在", "data": nil})
			return
		}
		c.File("SlinkWeb/dist/index.html")
	})
}

func getBaseConfig(c *gin.Context) {
	res, err := model.GetAllConfigs(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "读取配置失败", "data": nil, "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": res})
}

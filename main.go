package main

import (
	"Slink/applog"
	"Slink/model"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/thinkerou/favicon"
)

// main 程序入口点
// 负责初始化数据库、配置中间件并启动HTTP服务
func main() {
	applog.Init()

	// 初始化数据库连接
	initializeDatabase()

	if strings.TrimSpace(os.Getenv("SLINK_JWT_SECRET")) == "" {
		applog.Logger.Warn("未设置环境变量 SLINK_JWT_SECRET，正使用内置开发密钥；生产或 Docker 部署请务必设置长随机串")
	}

	// 创建Gin引擎并配置路由
	r := SetupRouter()

	// 配置favicon
	r.Use(favicon.New("SlinkWeb/dist/favicon.ico"))

	// 配置CORS跨域中间件
	configureCORSMiddleware(r)

	// 应用初始化守卫中间件
	r.Use(InitGuard())

	// 启动HTTP服务器
	startServer(r)
}

// initializeDatabase 初始化数据库连接
// 加载数据库配置并建立连接
func initializeDatabase() {
	dbConfig, err := model.LoadDatabaseConfig()
	if err != nil {
		panic(fmt.Sprintf("加载数据库配置失败: %v", err))
	}

	if err := model.InitDBWithConfig(dbConfig); err != nil {
		panic(fmt.Sprintf("数据库初始化失败: %v", err))
	}
}

// configureCORSMiddleware 配置CORS中间件
// 允许跨域请求，支持各种HTTP方法和头部
func configureCORSMiddleware(r *gin.Engine) {
	config := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false, // 认证走 Authorization，与 AllowOrigins: * 组合时须为 false
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(config))
}

// startServer 启动HTTP服务器
// 在指定端口监听并处理请求（环境变量 SLINK_ADDR，默认 :8080）
func startServer(r *gin.Engine) {
	addr := strings.TrimSpace(os.Getenv("SLINK_ADDR"))
	if addr == "" {
		addr = ":8080"
	}
	applog.Logger.Info("HTTP 服务监听中", "addr", addr)
	if err := r.Run(addr); err != nil {
		applog.Logger.Error("HTTP 服务退出", "error", err)
		panic(err)
	}
}

package api

import (
	"Slink/cache"
	"Slink/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetInitStatus 获取初始化状态
func GetInitStatus(c *gin.Context) {
	status, err := model.CheckInitStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取初始化状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": status,
	})
}

// PerformSetup 执行初始化设置
func PerformSetup(c *gin.Context) {
	var req model.SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 执行初始化
	if err := model.PerformInitialSetup(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "初始化失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "初始化成功",
		"data": gin.H{
			"admin_account": req.ResolvedAdminAccount(),
		},
	})
}

// TestDatabaseConnection 测试数据库连接
func TestDatabaseConnection(c *gin.Context) {
	var config model.DBConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 仅测试连通性，不迁移表结构、不切换全局数据库连接
	if err := model.TestDBConnection(&config); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "数据库连接失败: " + err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "数据库连接成功",
		"data": gin.H{
			"success": true,
		},
	})
}

// TestRedisConnection 测试Redis连接
func TestRedisConnection(c *gin.Context) {
	var req struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	config := cache.RedisConfig{
		Host:     req.Host,
		Port:     req.Port,
		Password: req.Password,
		DB:       req.DB,
	}

	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 6379
	}

	if err := cache.TestRedisConnection(config); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "Redis连接失败: " + err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "Redis连接成功",
		"data": gin.H{
			"success": true,
		},
	})
}

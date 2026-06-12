package api

import (
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

	// 尝试连接数据库
	if err := model.InitDBWithConfig(&config); err != nil {
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

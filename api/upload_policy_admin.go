package api

import (
	"Slink/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUploadPolicy 获取全局上传策略（结构化，与 GroupConfig 一致）
func GetUploadPolicy(c *gin.Context) {
	pol, err := model.GetGlobalUploadPolicy(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "读取上传策略失败",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data":    pol,
	})
}

// UpdateUploadPolicy 更新全局上传策略（请求体为完整 GroupConfig JSON）
func UpdateUploadPolicy(c *gin.Context) {
	var body model.GroupConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}
	model.NormalizeWatermarkConfig(body.WatermarkConfigs)
	if err := model.SaveGlobalUploadPolicyFromGroupConfig(model.DB, &body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "保存失败",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "保存成功",
	})
}

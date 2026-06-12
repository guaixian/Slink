package api

import (
	"Slink/model"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// GetConfigs 获取所有全局配置
func GetConfigs(c *gin.Context) {
	db := model.DB
	configs, err := model.GetAllConfigs(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "获取配置失败",
			"error":   err.Error(),
		})
		return
	}

	// 转换为map格式，方便前端使用
	configMap := make(map[string]interface{})
	for _, config := range configs {
		if config.ConfigKey == model.ConfigKeyUploadPolicy {
			continue
		}
		// 根据key转换value类型
		switch config.ConfigKey {
		case "enable_register", "enable_gallery", "enable_api", "guest_upload",
			"email_verify", "enable_antihotlink", "antihotlink_allow_empty":
			configMap[config.ConfigKey] = config.Value == "true"
		case "default_storage_gb":
			var val int
			if _, err := fmt.Sscanf(config.Value, "%d", &val); err == nil {
				configMap[config.ConfigKey] = val
			} else {
				configMap[config.ConfigKey] = config.Value
			}
		default:
			configMap[config.ConfigKey] = config.Value
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data":    configMap,
	})
}

// UpdateConfigs 批量更新全局配置
func UpdateConfigs(c *gin.Context) {
	db := model.DB

	// 接收map格式的配置
	var configMap map[string]interface{}
	if err := c.ShouldBindJSON(&configMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	delete(configMap, model.ConfigKeyUploadPolicy)
	delete(configMap, "upload_policy")

	// 转换为Config数组
	var configs []model.Config
	for key, value := range configMap {
		var valueStr string
		switch v := value.(type) {
		case bool:
			if v {
				valueStr = "true"
			} else {
				valueStr = "false"
			}
		case float64:
			valueStr = fmt.Sprintf("%.0f", v)
		case string:
			valueStr = v
		case map[string]interface{}, []interface{}:
			b, err := json.Marshal(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  false,
					"message": "配置值无法序列化: " + key,
					"error":   err.Error(),
				})
				return
			}
			valueStr = string(b)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}

		configs = append(configs, model.Config{
			ConfigKey: key,
			Value:     valueStr,
		})
	}

	if err := model.UpdateConfigs(db, configs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "更新配置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "配置更新成功",
	})
}

// GetSystemStats 获取系统统计信息
func GetSystemStats(c *gin.Context) {
	db := model.DB

	// 获取用户总数
	var totalUsers int64
	db.Model(&model.User{}).Count(&totalUsers)

	// 获取图片总数
	var totalImages int64
	db.Model(&model.Images{}).Count(&totalImages)

	// 获取总存储大小
	var totalSize int64
	db.Model(&model.Images{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)
	totalSizeMB := float64(totalSize) / (1024 * 1024)

	// 获取今日上传数量
	today := time.Now().Format("2006-01-02")
	var todayUpload int64
	db.Model(&model.Images{}).Where("DATE(created_at) = ?", today).Count(&todayUpload)

	// 获取昨日上传数量
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var yesterdayUpload int64
	db.Model(&model.Images{}).Where("DATE(created_at) = ?", yesterday).Count(&yesterdayUpload)

	// 获取本周上传数量
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	var weekUpload int64
	db.Model(&model.Images{}).Where("created_at >= ?", weekStart).Count(&weekUpload)

	// 获取本月上传数量
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	var monthUpload int64
	db.Model(&model.Images{}).Where("created_at >= ?", monthStart).Count(&monthUpload)

	// 获取近30天的上传趋势数据
	var trendData []map[string]interface{}
	for i := 29; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var count int64
		db.Model(&model.Images{}).Where("DATE(created_at) = ?", date).Count(&count)
		trendData = append(trendData, map[string]interface{}{
			"date":  date,
			"count": count,
		})
	}

	// 获取系统信息
	osInfo := runtime.GOOS + "/" + runtime.GOARCH
	goVersion := runtime.Version()

	uploadLimitStr := "—"
	postLimitStr := "—"
	if pol, err := model.GetGlobalUploadPolicy(db); err == nil && pol.MaximumFileSize > 0 {
		uploadLimitStr = fmt.Sprintf("单文件最大 %d KB", pol.MaximumFileSize)
		postLimitStr = uploadLimitStr
	}

	stats := gin.H{
		"total_users":      totalUsers,
		"total_images":     totalImages,
		"total_albums":     0, // 如果有相册功能，这里需要实现
		"total_size_mb":    totalSizeMB,
		"today_upload":     todayUpload,
		"yesterday_upload": yesterdayUpload,
		"week_upload":      weekUpload,
		"month_upload":     monthUpload,
		"trend_data":       trendData,
		"system_info": gin.H{
			"os":           osInfo,
			"web_server":   "Gin Framework",
			"go_version":   goVersion,
			"upload_limit": uploadLimitStr,
			"post_limit":   postLimitStr,
		},
		"software_info": gin.H{
			"version":     "Slink 个人图床",
			"description": "单用户自托管图床",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data":    stats,
	})
}

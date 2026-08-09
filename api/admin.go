package api

import (
	"Slink/model"
	"Slink/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
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

// AdminGetImages 获取所有图片列表（管理员）
func AdminGetImages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }
	if limit > 500 { limit = 500 }

	images, total, err := model.GetAllImagesPaginated(model.DB, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "获取图片列表失败"})
		return
	}

	var list []map[string]interface{}
	for i := range images {
		img := &images[i]
		// 生成访问链接与策略名，供管理端预览/复制（与图片列表接口保持一致）
		strategyName := ""
		var links map[string]string
		if utilsConfig, err := strategyUtilsConfigForStoredImage(model.DB, img); err == nil {
			pathname := normalizePathnameForLinks(img.Path)
			links = utils.GenerateImageLinksWithConfig(pathname, img.OriginName, utilsConfig)
		}
		if st, err := model.GetStrategyForStoredImage(model.DB, img.StrategyID); err == nil {
			strategyName = st.Name
		}
		list = append(list, map[string]interface{}{
			"id":            img.ID,
			"user_id":       img.UserID,
			"pathname":      normalizePathnameForLinks(img.Path),
			"origin_name":   img.OriginName,
			"size_bytes":    img.Size,
			"size":          float64(img.Size) / (1024 * 1024),
			"mimetype":      img.Mimetype,
			"permission":    img.Permissions,
			"strategy_id":   img.StrategyID,
			"strategy_name": strategyName,
			"links":         links,
			"created_at":    img.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   list,
		"total":  total,
		"page":   page,
		"limit":  limit,
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

// GetUserStrategies 获取用户分配的策略列表
func GetUserStrategies(c *gin.Context) {
	userIDStr := c.Param("id")
	var uid uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "无效的用户ID"})
		return
	}

	strategies, err := model.GetUserStrategies(model.DB, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "获取用户策略失败", "error": err.Error()})
		return
	}

	strategyIDs, _ := model.GetStrategyIDsByUserID(model.DB, uid)

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data": gin.H{
			"strategies":   strategies,
			"strategy_ids": strategyIDs,
		},
	})
}

// AssignStrategiesToUser 为用户分配策略
func AssignStrategiesToUser(c *gin.Context) {
	userIDStr := c.Param("id")
	var uid uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "无效的用户ID"})
		return
	}

	var req struct {
		// 允许传空数组以清空用户的全部分配策略
		StrategyIDs []uint `json:"strategy_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "请求参数错误"})
		return
	}

	// 先移除用户所有策略，再重新分配
	if err := model.DeleteUserStrategiesByUserID(model.DB, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "清除用户策略失败"})
		return
	}

	for _, sid := range req.StrategyIDs {
		if err := model.AddStrategyToUser(model.DB, uid, sid); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "分配策略失败", "error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "策略分配成功",
	})
}

// RemoveUserStrategy 移除用户的某个策略
func RemoveUserStrategy(c *gin.Context) {
	userIDStr := c.Param("id")
	strategyIDStr := c.Param("strategyId")

	var uid, sid uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "无效的用户ID"})
		return
	}
	if _, err := fmt.Sscanf(strategyIDStr, "%d", &sid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "无效的策略ID"})
		return
	}

	if err := model.RemoveStrategyFromUser(model.DB, uid, sid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "移除策略失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "移除成功",
	})
}

package api

import (
	"Slink/model"
	"Slink/utils"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// generateShareCode 生成分享码
func generateShareCode() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateShare 创建图片分享
func CreateShare(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 解析请求体
	var requestData struct {
		ImageID   uint   `json:"image_id" binding:"required"`
		Password  string `json:"password"`
		ExpiresIn int    `json:"expires_in"` // 过期时间（小时）
		MaxViews  int    `json:"max_views"`  // 最大查看次数
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "请求参数格式错误",
		})
		return
	}

	// 获取图片信息
	image, err := model.GetImageByID(model.DB, requestData.ImageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "图片不存在",
		})
		return
	}

	// 检查权限（只能分享自己的图片）
	if image.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "无权限分享此图片",
		})
		return
	}

	// 生成分享码
	shareCode, err := generateShareCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "生成分享码失败",
		})
		return
	}

	// 处理密码
	var hashedPassword string
	if requestData.Password != "" {
		hash, err := utils.HashPassword(requestData.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "密码加密失败",
			})
			return
		}
		hashedPassword = hash
	}

	// 处理过期时间
	var expiresAt *time.Time
	if requestData.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(requestData.ExpiresIn) * time.Hour)
		expiresAt = &expiry
	}

	// 创建分享记录
	share := &model.Share{
		UserID:    userID.(uint),
		ImageID:   requestData.ImageID,
		ShareCode: shareCode,
		Password:  hashedPassword,
		ExpiresAt: expiresAt,
		MaxViews:  requestData.MaxViews,
		IsActive:  true,
	}

	if err := model.CreateShare(model.DB, share); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "创建分享失败",
		})
		return
	}

	// 构建分享链接
	shareURL := c.Request.Host + "/share/" + shareCode

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "创建分享成功",
		"data": gin.H{
			"id":         share.ID,
			"share_code": share.ShareCode,
			"share_url":  shareURL,
			"expires_at": share.ExpiresAt,
			"max_views":  share.MaxViews,
		},
	})
}

// GetShare 获取分享信息（需要密码验证）
func GetShare(c *gin.Context) {
	// 获取分享码
	shareCode := c.Param("code")

	// 获取分享信息
	share, err := model.GetShareByCode(model.DB, shareCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "分享不存在",
		})
		return
	}

	// 检查分享是否有效
	if !share.IsValid() {
		c.JSON(http.StatusGone, gin.H{
			"status":  false,
			"message": "分享已过期或已失效",
		})
		return
	}

	// 如果有密码保护，需要验证密码
	if share.Password != "" {
		var requestData struct {
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": "需要密码",
				"data": gin.H{
					"password_required": true,
				},
			})
			return
		}

		// 验证密码
		if !utils.CheckPassword(share.Password, requestData.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "密码错误",
			})
			return
		}
	}

	// 获取图片信息
	image, err := model.GetImageByID(model.DB, share.ImageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "图片不存在",
		})
		return
	}

	// 增加查看次数
	model.IncrementShareViewCount(model.DB, share.ID)

	strategy, err := model.GetStrategyForStoredImage(model.DB, image.StrategyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "获取存储策略失败",
		})
		return
	}

	utilsConfig, err := strategyUtilsConfigFromModel(strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "解析策略配置失败",
		})
		return
	}

	pathname := normalizePathnameForLinks(image.Path)
	links := utils.GenerateImageLinksWithConfig(pathname, image.OriginName, utilsConfig)

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取分享成功",
		"data": gin.H{
			"image": gin.H{
				"id":          image.ID,
				"origin_name": image.OriginName,
				"size":        float64(image.Size) / (1024 * 1024),
				"mimetype":    image.Mimetype,
				"width":       image.Width,
				"height":      image.Height,
				"links":       links,
			},
			"share": gin.H{
				"view_count": share.ViewCount + 1,
				"max_views":  share.MaxViews,
				"expires_at": share.ExpiresAt,
			},
		},
	})
}

// GetMyShares 获取当前用户的所有分享
func GetMyShares(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 获取用户的所有分享
	shares, err := model.GetSharesByUserID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "获取分享列表失败",
		})
		return
	}

	// 构建响应数据
	var shareList []gin.H
	for _, share := range shares {
		// 获取图片信息
		image, err := model.GetImageByID(model.DB, share.ImageID)
		if err != nil {
			continue
		}

		shareURL := c.Request.Host + "/share/" + share.ShareCode

		shareList = append(shareList, gin.H{
			"id":               share.ID,
			"share_code":       share.ShareCode,
			"share_url":        shareURL,
			"image_id":         share.ImageID,
			"image_name":       image.OriginName,
			"password_enabled": share.Password != "",
			"expires_at":       share.ExpiresAt,
			"view_count":       share.ViewCount,
			"max_views":        share.MaxViews,
			"is_active":        share.IsActive,
			"is_expired":       share.IsExpired(),
			"created_at":       share.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取分享列表成功",
		"data":    shareList,
	})
}

// DeleteShare 删除分享
func DeleteShareByID(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 获取分享ID
	shareIDStr := c.Param("id")
	shareID, err := strconv.ParseUint(shareIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "无效的分享ID",
		})
		return
	}

	// 获取分享信息
	share, err := model.GetShareByID(model.DB, uint(shareID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "分享不存在",
		})
		return
	}

	// 检查权限（只能删除自己的分享）
	if share.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "无权限删除此分享",
		})
		return
	}

	// 删除分享
	if err := model.DeleteShare(model.DB, uint(shareID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "删除分享失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "删除分享成功",
	})
}

// UpdateShareStatus 更新分享状态
func UpdateShareStatus(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 获取分享ID
	shareIDStr := c.Param("id")
	shareID, err := strconv.ParseUint(shareIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "无效的分享ID",
		})
		return
	}

	// 解析请求体
	var requestData struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "请求参数格式错误",
		})
		return
	}

	// 获取分享信息
	share, err := model.GetShareByID(model.DB, uint(shareID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "分享不存在",
		})
		return
	}

	// 检查权限
	if share.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "无权限修改此分享",
		})
		return
	}

	// 更新状态
	share.IsActive = requestData.IsActive
	if err := model.UpdateShare(model.DB, share); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "更新分享状态失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "更新分享状态成功",
		"data": gin.H{
			"id":        share.ID,
			"is_active": share.IsActive,
		},
	})
}

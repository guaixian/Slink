package api

import (
	"Slink/model"
	"Slink/utils"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// GenerateQRCode 生成图片链接的二维码
func GenerateQRCode(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 获取图片ID
	imageIDStr := c.Param("id")
	imageID, err := strconv.ParseUint(imageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "无效的图片ID",
		})
		return
	}

	// 获取图片信息
	image, err := model.GetImageByID(model.DB, uint(imageID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "图片不存在",
		})
		return
	}

	// 检查权限（只能为自己的图片生成二维码）
	if image.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "无权限访问此图片",
		})
		return
	}

	utilsConfig, err := strategyUtilsConfigForStoredImage(model.DB, image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "获取存储策略失败",
		})
		return
	}

	pathname := normalizePathnameForLinks(image.Path)
	links := utils.GenerateImageLinksWithConfig(pathname, image.OriginName, utilsConfig)

	// 生成二维码
	qr, err := qrcode.New(links["url"], qrcode.Medium)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "生成二维码失败",
		})
		return
	}

	// 设置二维码大小
	qr.DisableBorder = false

	// 生成PNG格式的二维码
	pngBytes, err := qr.PNG(256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "生成二维码图片失败",
		})
		return
	}

	// 返回二维码图片
	c.Header("Content-Type", "image/png")
	c.Header("Content-Length", strconv.Itoa(len(pngBytes)))
	c.Data(http.StatusOK, "image/png", pngBytes)
}

// GenerateQRCodeBase64 生成图片链接的二维码（Base64格式）
func GenerateQRCodeBase64(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 获取图片ID
	imageIDStr := c.Param("id")
	imageID, err := strconv.ParseUint(imageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "无效的图片ID",
		})
		return
	}

	// 获取图片信息
	image, err := model.GetImageByID(model.DB, uint(imageID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "图片不存在",
		})
		return
	}

	// 检查权限
	if image.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "无权限访问此图片",
		})
		return
	}

	utilsConfig, err := strategyUtilsConfigForStoredImage(model.DB, image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "获取存储策略失败",
		})
		return
	}

	pathname := normalizePathnameForLinks(image.Path)
	links := utils.GenerateImageLinksWithConfig(pathname, image.OriginName, utilsConfig)

	// 生成二维码
	qr, err := qrcode.New(links["url"], qrcode.Medium)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "生成二维码失败",
		})
		return
	}

	// 生成PNG格式的二维码
	pngBytes, err := qr.PNG(256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "生成二维码图片失败",
		})
		return
	}

	// 转换为Base64
	base64Str := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "生成二维码成功",
		"data": gin.H{
			"qrcode": base64Str,
			"url":    links["url"],
		},
	})
}

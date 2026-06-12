package api

import (
	"Slink/model"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GenerateToken 生成随机token
func generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateToken 创建个人访问令牌
func CreateToken(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 获取用户信息
	user, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "获取用户信息失败",
		})
		return
	}

	// 解析请求体
	var requestData struct {
		ApiName string `json:"api_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "请求参数格式错误",
		})
		return
	}

	// 生成token
	token, err := generateRandomToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "生成token失败",
		})
		return
	}

	// 创建token记录
	accessToken := &model.PersonalAccessToken{
		ApiName:  requestData.ApiName,
		Username: user.Email,
		Token:    token,
	}

	if err := model.CreatePersonalAccessToken(model.DB, accessToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "创建token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "创建token成功",
		"data": gin.H{
			"id":         accessToken.ID,
			"api_name":   accessToken.ApiName,
			"token":      accessToken.Token,
			"created_at": accessToken.CreatedAt,
		},
	})
}

// GetTokens 获取当前用户的所有token
func GetTokens(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 获取用户信息
	user, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "获取用户信息失败",
		})
		return
	}

	// 获取用户的所有token
	tokens, err := model.GetPersonalAccessTokensByUsername(model.DB, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "获取token列表失败",
		})
		return
	}

	// 构建响应数据（隐藏完整token，只显示前8位）
	var tokenList []gin.H
	for _, token := range tokens {
		maskedToken := token.Token
		if len(maskedToken) > 8 {
			maskedToken = maskedToken[:8] + "..." + maskedToken[len(maskedToken)-4:]
		}

		tokenList = append(tokenList, gin.H{
			"id":         token.ID,
			"api_name":   token.ApiName,
			"token":      maskedToken,
			"created_at": token.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取token列表成功",
		"data":    tokenList,
	})
}

// DeleteToken 删除token
func DeleteToken(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "用户未认证",
		})
		return
	}

	// 获取用户信息
	user, err := model.GetUserByID(model.DB, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "获取用户信息失败",
		})
		return
	}

	// 获取token ID
	tokenIDStr := c.Param("id")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "无效的token ID",
		})
		return
	}

	// 获取token信息
	token, err := model.GetPersonalAccessTokenByID(model.DB, uint(tokenID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "token不存在",
		})
		return
	}

	// 检查权限（只能删除自己的token）
	if token.Username != user.Email {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "无权限删除此token",
		})
		return
	}

	// 删除token
	if err := model.DeletePersonalAccessToken(model.DB, uint(tokenID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "删除token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "删除token成功",
	})
}

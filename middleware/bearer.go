package middleware

import (
	"Slink/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BearerAuthMiddleware Bearer Token认证中间件
// 支持 Authorization: Bearer <token> 格式
func BearerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "缺少Authorization头",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Authorization格式错误，应为: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证token
		accessToken, err := model.GetPersonalAccessTokenByToken(model.DB, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "无效的token",
			})
			c.Abort()
			return
		}

		// 根据username获取用户信息
		user, err := model.GetUserByEmail(model.DB, accessToken.Username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set("userID", user.ID)
		c.Set("username", user.Email)
		c.Set("isAdmin", user.IsAdmin)

		c.Next()
	}
}

// JWTOrBearerAuthMiddleware JWT或Bearer Token认证中间件
// 优先尝试JWT，如果失败则尝试Bearer Token
func JWTOrBearerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "缺少Authorization头",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Authorization格式错误，应为: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 先尝试JWT认证
		claims, err := ParseToken(token)
		if err == nil {
			// JWT认证成功
			c.Set("userID", claims.UserID)
			c.Set("email", claims.Email)

			// 获取用户信息以确定是否为管理员（带短缓存，避免每请求查库）
			user, err := GetCachedUser(claims.UserID)
			if err == nil {
				c.Set("isAdmin", user.IsAdmin)
			}

			c.Next()
			return
		}

		// JWT认证失败，尝试Personal Access Token
		accessToken, err := model.GetPersonalAccessTokenByToken(model.DB, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "无效的token",
			})
			c.Abort()
			return
		}

		// 根据username获取用户信息
		user, err := model.GetUserByEmail(model.DB, accessToken.Username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set("userID", user.ID)
		c.Set("username", user.Email)
		c.Set("isAdmin", user.IsAdmin)

		c.Next()
	}
}

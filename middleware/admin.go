package middleware

import (
	"Slink/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 校验当前登录用户是否为管理员（IsAdmin == 1）。
// 注意：开放注册后普通注册用户也能拿到合法 JWT，必须在这里拦截，
// 否则任意用户都能调用 /api/admin/*（改配置、删用户、导出备份等）。
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		user, err := model.GetUserByID(model.DB, userID.(uint))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
			c.Abort()
			return
		}

		if user.IsAdmin != 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			c.Abort()
			return
		}

		c.Set("adminUser", user)
		c.Next()
	}
}

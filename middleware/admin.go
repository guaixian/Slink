package middleware

import (
	"Slink/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 个人图床后台权限：已通过 JWT 登录即视为唯一站长，不再区分多用户管理员。
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

		c.Set("adminUser", user)
		c.Next()
	}
}

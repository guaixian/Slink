package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AntiHotlinkMiddleware 防盗链中间件
// allowedDomains: 允许的域名列表，为空则允许所有
// allowEmpty: 是否允许空Referer（直接访问）
func AntiHotlinkMiddleware(allowedDomains []string, allowEmpty bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		referer := c.GetHeader("Referer")

		// 如果允许空Referer且Referer为空，则放行
		if allowEmpty && referer == "" {
			c.Next()
			return
		}

		// 如果没有配置允许的域名，则放行所有
		if len(allowedDomains) == 0 {
			c.Next()
			return
		}

		// 检查Referer是否在允许列表中
		isAllowed := false
		for _, domain := range allowedDomains {
			if strings.Contains(referer, domain) {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  false,
				"message": "禁止盗链访问",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckReferer 检查Referer是否合法
func CheckReferer(c *gin.Context, allowedDomains []string, allowEmpty bool) bool {
	referer := c.GetHeader("Referer")

	// 如果允许空Referer且Referer为空，则返回true
	if allowEmpty && referer == "" {
		return true
	}

	// 如果没有配置允许的域名，则返回true
	if len(allowedDomains) == 0 {
		return true
	}

	// 检查Referer是否在允许列表中
	for _, domain := range allowedDomains {
		if strings.Contains(referer, domain) {
			return true
		}
	}

	return false
}

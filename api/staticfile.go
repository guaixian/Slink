package api

import (
	"Slink/middleware"
	"Slink/model"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func IndexHtml(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "index",
	})
}

// ServeProtectedImage 提供受防盗链保护的图片
func ServeProtectedImage(c *gin.Context) {
	// 获取防盗链配置
	enableAntihotlink, _ := model.GetConfigValue(model.DB, "enable_antihotlink")

	// 如果启用了防盗链
	if enableAntihotlink == "true" {
		// 获取允许的域名列表
		domainsStr, _ := model.GetConfigValue(model.DB, "antihotlink_domains")
		allowedDomains := []string{}
		if domainsStr != "" {
			allowedDomains = strings.Split(domainsStr, ",")
			// 去除空格
			for i := range allowedDomains {
				allowedDomains[i] = strings.TrimSpace(allowedDomains[i])
			}
		}

		// 获取是否允许空Referer
		allowEmptyStr, _ := model.GetConfigValue(model.DB, "antihotlink_allow_empty")
		allowEmpty := allowEmptyStr == "true"

		// 检查Referer
		if !middleware.CheckReferer(c, allowedDomains, allowEmpty) {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  false,
				"message": "禁止盗链访问",
			})
			return
		}
	}

	// 获取文件路径
	filePath := c.Param("filepath")

	// 安全检查：防止路径遍历攻击
	if strings.Contains(filePath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "非法的文件路径",
		})
		return
	}

	// 清理路径
	filePath = filepath.Clean(filePath)
	fullPath := filepath.Join("static", filePath)

	// 确保文件路径在static目录下
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "文件路径解析失败",
		})
		return
	}

	staticDir, err := filepath.Abs("static")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "静态目录解析失败",
		})
		return
	}

	// 检查文件是否在static目录下
	if !strings.HasPrefix(absPath, staticDir) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"message": "禁止访问该文件",
		})
		return
	}

	// 图片内容按 UUID 命名、永不变更，可安全长缓存，显著减少重复请求
	c.Header("Cache-Control", "public, max-age=604800, immutable")

	// 提供文件
	c.File(fullPath)
}

package main

import (
	"Slink/model"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// initStatusCache 用于缓存系统初始化状态的线程安全结构
// 避免频繁查询数据库，提高性能
var initStatusCache struct {
	status   *model.InitStatus
	updateAt time.Time
	mu       sync.RWMutex
}

// InitGuard 初始化守卫中间件
// 用于保护API接口，确保系统在初始化前不可用
// 排除 /api/init 路径，允许系统在未初始化时完成设置
func InitGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 非API请求直接通过
		if !strings.HasPrefix(path, "/api") {
			c.Next()
			return
		}

		// 初始化接口无需检查
		if strings.HasPrefix(path, "/api/init") {
			c.Next()
			return
		}

		// 获取缓存的初始化状态
		status := getCachedInitStatus()

		// 状态检查失败
		if status == nil {
			c.JSON(500, gin.H{"msg": "初始化状态检查失败"})
			c.Abort()
			return
		}

		// 系统未初始化，拒绝访问
		if !status.IsInitialized {
			c.JSON(403, gin.H{
				"code": 403,
				"msg":  "系统尚未初始化",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getCachedInitStatus 获取缓存的初始化状态
// 实现5秒缓存机制，减少数据库查询频率
// 使用读写锁确保线程安全
func getCachedInitStatus() *model.InitStatus {
	// 尝试从缓存读取（读锁）
	initStatusCache.mu.RLock()
	if time.Since(initStatusCache.updateAt) < 5*time.Second && initStatusCache.status != nil {
		defer initStatusCache.mu.RUnlock()
		return initStatusCache.status
	}
	initStatusCache.mu.RUnlock()

	// 获取写锁，尝试从缓存读取（双重检查）
	initStatusCache.mu.Lock()
	defer initStatusCache.mu.Unlock()

	// 再次检查缓存（避免并发更新）
	if time.Since(initStatusCache.updateAt) < 5*time.Second && initStatusCache.status != nil {
		return initStatusCache.status
	}

	// 查询数据库获取最新状态
	status, err := model.CheckInitStatus()
	if err != nil {
		return nil
	}

	// 更新缓存
	initStatusCache.status = status
	initStatusCache.updateAt = time.Now()
	return status
}

package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache 缓存接口
type Cache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) bool
	Clear() error
}

// FileCache 文件缓存
type FileCache struct {
	basePath string
	mu       sync.RWMutex
}

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{} `json:"value"`
	ExpiresAt int64       `json:"expires_at"` // Unix timestamp
}

// NewFileCache 创建文件缓存实例
func NewFileCache(basePath string) (*FileCache, error) {
	if basePath == "" {
		basePath = "cache"
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	cache := &FileCache{
		basePath: basePath,
	}

	// 启动清理过期缓存的协程
	go cache.cleanupExpired()

	return cache, nil
}

// Get 获取缓存
func (c *FileCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	filePath := c.getFilePath(key)

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("缓存不存在")
		}
		return nil, err
	}

	// 解析缓存项
	var item CacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("解析缓存失败: %w", err)
	}

	// 检查是否过期
	if item.ExpiresAt > 0 && time.Now().Unix() > item.ExpiresAt {
		// 删除过期缓存
		os.Remove(filePath)
		return nil, fmt.Errorf("缓存已过期")
	}

	return item.Value, nil
}

// Set 设置缓存
func (c *FileCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建缓存项
	item := CacheItem{
		Value: value,
	}

	// 设置过期时间
	if ttl > 0 {
		item.ExpiresAt = time.Now().Add(ttl).Unix()
	}

	// 序列化
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入缓存失败: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (c *FileCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除缓存失败: %w", err)
	}

	return nil
}

// Exists 检查缓存是否存在
func (c *FileCache) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	filePath := c.getFilePath(key)
	_, err := os.Stat(filePath)
	return err == nil
}

// Clear 清空所有缓存
func (c *FileCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 删除缓存目录
	if err := os.RemoveAll(c.basePath); err != nil {
		return fmt.Errorf("清空缓存失败: %w", err)
	}

	// 重新创建缓存目录
	if err := os.MkdirAll(c.basePath, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	return nil
}

// getFilePath 获取缓存文件路径
func (c *FileCache) getFilePath(key string) string {
	// 使用 key 的前两个字符作为子目录，避免单个目录文件过多
	if len(key) >= 2 {
		subDir := key[:2]
		return filepath.Join(c.basePath, subDir, key+".cache")
	}
	return filepath.Join(c.basePath, key+".cache")
}

// cleanupExpired 清理过期缓存
func (c *FileCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		c.cleanupExpiredFiles(c.basePath)
		c.mu.Unlock()
	}
}

// cleanupExpiredFiles 递归清理过期文件
func (c *FileCache) cleanupExpiredFiles(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			c.cleanupExpiredFiles(path)
			continue
		}

		// 读取缓存文件
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// 解析缓存项
		var item CacheItem
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}

		// 检查是否过期
		if item.ExpiresAt > 0 && time.Now().Unix() > item.ExpiresAt {
			os.Remove(path)
		}
	}
}

// MemoryCache 内存缓存（简单实现）
type MemoryCache struct {
	data map[string]*CacheItem
	mu   sync.RWMutex
}

// NewMemoryCache 创建内存缓存实例
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]*CacheItem),
	}

	// 启动清理过期缓存的协程
	go cache.cleanupExpired()

	return cache
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, fmt.Errorf("缓存不存在")
	}

	// 检查是否过期
	if item.ExpiresAt > 0 && time.Now().Unix() > item.ExpiresAt {
		delete(c.data, key)
		return nil, fmt.Errorf("缓存已过期")
	}

	return item.Value, nil
}

// Set 设置缓存
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := &CacheItem{
		Value: value,
	}

	if ttl > 0 {
		item.ExpiresAt = time.Now().Add(ttl).Unix()
	}

	c.data[key] = item
	return nil
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// Exists 检查缓存是否存在
func (c *MemoryCache) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.data[key]
	return exists
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheItem)
	return nil
}

// cleanupExpired 清理过期缓存
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now().Unix()
		for key, item := range c.data {
			if item.ExpiresAt > 0 && now > item.ExpiresAt {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

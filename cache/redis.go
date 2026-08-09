package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	mu     sync.RWMutex
}

// RedisConfig Redis连接配置
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// DefaultRedisConfig 默认Redis配置
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}
}

// NewRedisCache 创建Redis缓存实例
func NewRedisCache(config RedisConfig) (*RedisCache, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 连接失败 (%s): %w", addr, err)
	}

	cache := &RedisCache{
		client: client,
	}

	// 启动清理过期缓存的协程（Redis自带TTL，此协程作为补充）
	go cache.cleanupExpired()

	return cache, nil
}

// TestRedisConnection 测试Redis连接
func TestRedisConnection(config RedisConfig) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Ping(ctx).Err()
}

// Get 获取缓存
func (c *RedisCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("缓存不存在")
		}
		return nil, err
	}

	// 尝试解析JSON
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// 如果不是JSON，直接返回字符串
		return val, nil
	}

	return result, nil
}

// Set 设置缓存
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 序列化值
	var data []byte
	var err error

	switch v := value.(type) {
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("序列化缓存失败: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.client.Set(ctx, key, string(data), ttl).Err()
}

// Delete 删除缓存
func (c *RedisCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("删除缓存失败: %w", err)
	}
	return nil
}

// Exists 检查缓存是否存在
func (c *RedisCache) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	val, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false
	}
	return val > 0
}

// Clear 清空当前DB的所有缓存
func (c *RedisCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.client.FlushDB(ctx).Err()
}

// Close 关闭Redis连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// cleanupExpired 清理过期缓存（Redis自带TTL，此处为兼容接口预留）
func (c *RedisCache) cleanupExpired() {
	// Redis 自带 TTL 过期机制，无需手动清理
	// 此函数保留为兼容接口
}

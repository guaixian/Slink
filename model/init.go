package model

import (
	"Slink/utils"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// InitConfig 初始化配置表
// 存储系统初始化配置信息
type InitConfig struct {
	ID        uint   `gorm:"primaryKey;comment:配置ID"`
	Key       string `gorm:"type:varchar(191);unique;not null;comment:配置键名"`
	Value     string `gorm:"not null;comment:配置值"`
	CreatedAt int64  `gorm:"autoCreateTime;comment:创建时间"`
}

// InitStatus 初始化状态
type InitStatus struct {
	IsInitialized bool   `json:"is_initialized"`
	AdminAccount  string `json:"admin_account,omitempty"`
	// AdminEmail 仅用于兼容旧版 init.json，新保存请使用 AdminAccount
	AdminEmail string `json:"admin_email,omitempty"`
}

// SetupRequest 初始化设置请求
type SetupRequest struct {
	AdminAccount   string    `json:"admin_account"`
	AdminEmail     string    `json:"admin_email"` // 兼容旧前端；与 AdminAccount 二选一
	AdminPassword  string    `json:"admin_password" binding:"required,min=6"`
	DatabaseConfig *DBConfig `json:"database_config" binding:"required"`
	CacheType      string    `json:"cache_type"` // "memory", "file", "redis"
	CachePath      string    `json:"cache_path,omitempty"`
	// Redis 配置
	RedisHost     string `json:"redis_host,omitempty"`
	RedisPort     int    `json:"redis_port,omitempty"`
	RedisPassword string `json:"redis_password,omitempty"`
	RedisDB       int    `json:"redis_db,omitempty"`
}

// ResolvedAdminAccount 返回非空的站长登录名（admin_account 或兼容 admin_email）
func (r *SetupRequest) ResolvedAdminAccount() string {
	a := strings.TrimSpace(r.AdminAccount)
	if a != "" {
		return a
	}
	return strings.TrimSpace(r.AdminEmail)
}

// CheckInitStatus 检查系统是否已初始化
func CheckInitStatus() (*InitStatus, error) {
	// 检查配置文件是否存在
	configFile := "config/init.json"
	if _, err := os.Stat(configFile); err == nil {
		// 配置文件存在，读取初始化状态
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, err
		}

		var status InitStatus
		if err := json.Unmarshal(data, &status); err != nil {
			return nil, err
		}
		if status.AdminAccount == "" && status.AdminEmail != "" {
			status.AdminAccount = status.AdminEmail
		}

		return &status, nil
	}

	// 配置文件不存在，返回未初始化状态
	return &InitStatus{
		IsInitialized: false,
	}, nil
}

// SaveInitStatus 保存初始化状态
func SaveInitStatus(status *InitStatus) error {
	// 确保配置目录存在
	if err := os.MkdirAll("config", 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 序列化状态
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化初始化状态失败: %w", err)
	}

	// 写入配置文件
	configFile := "config/init.json"
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("写入初始化状态失败: %w", err)
	}

	return nil
}

// PerformInitialSetup 执行初始化设置
func PerformInitialSetup(req *SetupRequest) error {
	// 检查是否已经初始化
	status, err := CheckInitStatus()
	if err != nil {
		return fmt.Errorf("检查初始化状态失败: %w", err)
	}

	if status.IsInitialized {
		return fmt.Errorf("系统已经初始化，无法重复初始化")
	}

	// 初始化数据库
	if err := InitDBWithConfig(req.DatabaseConfig); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 创建管理员账号
	adminAccount := req.ResolvedAdminAccount()
	if adminAccount == "" {
		return fmt.Errorf("请提供站长登录账号 admin_account")
	}

	hashedPassword, err := utils.HashPassword(req.AdminPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建默认配置
	defaultConfig := GetDefaultUserConfig()
	configJSON, err := defaultConfig.ToJSON()
	if err != nil {
		return fmt.Errorf("创建默认配置失败: %w", err)
	}

	admin := User{
		Email:        adminAccount,
		Password:     hashedPassword,
		Name:         "站长",
		GroupID:      1,
		IsAdmin:      1,
		Configs:      configJSON,
		RegisteredIP: "127.0.0.1",
	}

	if err := CreateUser(DB, &admin); err != nil {
		return fmt.Errorf("创建管理员账号失败: %w", err)
	}

	// 保存数据库配置到环境文件
	if err := saveDatabaseConfig(req.DatabaseConfig); err != nil {
		return fmt.Errorf("保存数据库配置失败: %w", err)
	}

	// 保存缓存配置
	if err := saveCacheConfig(req.CacheType, req.CachePath, req.RedisHost, req.RedisPort, req.RedisPassword, req.RedisDB); err != nil {
		return fmt.Errorf("保存缓存配置失败: %w", err)
	}

	// 标记为已初始化
	newStatus := &InitStatus{
		IsInitialized: true,
		AdminAccount:  adminAccount,
	}

	if err := SaveInitStatus(newStatus); err != nil {
		return fmt.Errorf("保存初始化状态失败: %w", err)
	}

	return nil
}

// saveDatabaseConfig 保存数据库配置到文件
func saveDatabaseConfig(config *DBConfig) error {
	// 确保配置目录存在
	if err := os.MkdirAll("config", 0755); err != nil {
		return err
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入配置文件
	configFile := "config/database.json"
	return os.WriteFile(configFile, data, 0644)
}

// LoadDatabaseConfigFromFile 从指定路径解析数据库配置（用于备份包校验，路径不受 cd 外约束）
func LoadDatabaseConfigFromFile(path string) (*DBConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config DBConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadDatabaseConfig 从文件加载数据库配置
func LoadDatabaseConfig() (*DBConfig, error) {
	configFile := "config/database.json"

	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return GetDBConfigFromEnv(), nil
	}

	return LoadDatabaseConfigFromFile(configFile)
}

// saveCacheConfig 保存缓存配置（支持Redis参数）
func saveCacheConfig(cacheType, cachePath string, redisHost string, redisPort int, redisPassword string, redisDB int) error {
	// 确保配置目录存在
	if err := os.MkdirAll("config", 0755); err != nil {
		return err
	}

	config := map[string]interface{}{
		"type": cacheType,
		"path": cachePath,
	}

	// 如果是Redis类型，保存Redis连接参数
	if cacheType == "redis" {
		config["redis_host"] = redisHost
		config["redis_port"] = redisPort
		config["redis_password"] = redisPassword
		config["redis_db"] = redisDB
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入配置文件
	configFile := "config/cache.json"
	return os.WriteFile(configFile, data, 0644)
}

// LoadCacheConfig 加载缓存配置（支持Redis参数）
func LoadCacheConfig() (map[string]string, error) {
	configFile := "config/cache.json"

	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return map[string]string{
			"type": "memory",
			"path": "",
		}, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// 解析配置
	var configRaw map[string]interface{}
	if err := json.Unmarshal(data, &configRaw); err != nil {
		return nil, err
	}

	// 转换为 map[string]string
	config := make(map[string]string)
	for k, v := range configRaw {
		switch val := v.(type) {
		case string:
			config[k] = val
		case float64:
			config[k] = fmt.Sprintf("%.0f", val)
		default:
			config[k] = fmt.Sprint(val)
		}
	}

	return config, nil
}

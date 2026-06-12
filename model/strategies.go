package model

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Strategies 存储策略表
// 定义图片存储的后端配置，支持多种存储方式
type Strategies struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;comment:策略ID"`
	Name         string    `gorm:"type:varchar(191);unique;not null;comment:策略名称"`
	Introduction string    `gorm:"type:varchar(255);not null;comment:策略简介"`
	StrategyKey  string    `gorm:"type:varchar(191);unique;not null;column:strategy_key;comment:策略标识符"`
	Configs      string    `gorm:"type:text;not null;comment:策略配置(JSON格式)"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

// 指定表名
func (Strategies) TableName() string {
	return "strategies"
}

// StrategyConfig 策略配置结构（向后兼容）
type StrategyConfig struct {
	URL      string            `json:"url"`
	Domain   string            `json:"domain"` // 对象存储等场景的自定义访问域名；外链生成时若 url 为空则回退到此字段
	Root     string            `json:"root"`
	Queries  map[string]string `json:"queries"`
	AuthType string            `json:"auth_type"`
}

// StrategyConfigData 完整的策略配置结构
type StrategyConfigData struct {
	// 基本配置（向后兼容）
	URL      string            `json:"url"`
	Root     string            `json:"root"`
	Queries  map[string]string `json:"queries"`
	AuthType string            `json:"auth_type"`

	// 存储配置
	StorageType string `json:"storage_type"` // local, s3, oss, cos, qiniu, upyun, minio
	Endpoint    string `json:"endpoint"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	Bucket      string `json:"bucket"`
	Region      string `json:"region"`
	Domain      string `json:"domain"`
	BasePath    string `json:"base_path"`
	UseSSL      bool   `json:"use_ssl"`
}

// CreateStrategy 创建策略
func CreateStrategy(db *gorm.DB, strategy *Strategies) error {
	return db.Create(strategy).Error
}

// GetStrategyByID 根据ID获取策略
func GetStrategyByID(db *gorm.DB, id uint) (*Strategies, error) {
	var strategy Strategies
	err := db.First(&strategy, id).Error
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

// GetStrategyByName 根据名称获取策略
func GetStrategyByName(db *gorm.DB, name string) (*Strategies, error) {
	var strategy Strategies
	err := db.Where("name = ?", name).First(&strategy).Error
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

// GetStrategyByKey 根据Key获取策略
func GetStrategyByKey(db *gorm.DB, key string) (*Strategies, error) {
	var strategy Strategies
	err := db.Where("strategy_key = ?", key).First(&strategy).Error
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

// GetAllStrategies 获取所有策略
func GetAllStrategies(db *gorm.DB) ([]Strategies, error) {
	var strategies []Strategies
	err := db.Find(&strategies).Error
	return strategies, err
}

// UpdateStrategy 更新策略
func UpdateStrategy(db *gorm.DB, strategy *Strategies) error {
	return db.Save(strategy).Error
}

// DeleteStrategy 删除策略
func DeleteStrategy(db *gorm.DB, id uint) error {
	return db.Delete(&Strategies{}, id).Error
}

// GetStrategyConfig 获取策略配置
func GetStrategyConfig(strategy *Strategies) (*StrategyConfig, error) {
	var config StrategyConfig
	err := json.Unmarshal([]byte(strategy.Configs), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetStrategyConfigData 获取完整的策略配置
func GetStrategyConfigData(strategy *Strategies) (*StrategyConfigData, error) {
	var config StrategyConfigData
	err := json.Unmarshal([]byte(strategy.Configs), &config)
	if err != nil {
		return nil, err
	}

	// 根据策略 key 映射存储类型
	keyToStorageTypeMap := map[string]string{
		"1":  "local",
		"2":  "ftp",
		"3":  "sftp",
		"4":  "oss",
		"5":  "cos",
		"6":  "qiniu",
		"7":  "webdav",
		"8":  "s3",
		"9":  "minio",
		"10": "upyun",
	}

	// 如果没有设置 storage_type，根据 key 设置
	if config.StorageType == "" {
		if storageType, ok := keyToStorageTypeMap[strategy.StrategyKey]; ok {
			config.StorageType = storageType
		}
	}

	// 处理前端保存的配置字段映射到后端期望的字段
	// 解析原始配置以获取前端保存的字段
	var rawConfig map[string]interface{}
	if err := json.Unmarshal([]byte(strategy.Configs), &rawConfig); err == nil {
		// WebDAV 字段映射
		if config.StorageType == "webdav" {
			// webdav_url -> endpoint
			if webdavURL, ok := rawConfig["webdav_url"].(string); ok && config.Endpoint == "" {
				config.Endpoint = webdavURL
			}
			// username -> access_key
			if username, ok := rawConfig["username"].(string); ok && config.AccessKey == "" {
				config.AccessKey = username
			}
			// password -> secret_key
			if password, ok := rawConfig["password"].(string); ok && config.SecretKey == "" {
				config.SecretKey = password
			}
			// path -> base_path
			if path, ok := rawConfig["path"].(string); ok && config.BasePath == "" {
				config.BasePath = path
			}
		}

		// FTP/SFTP 字段映射
		if config.StorageType == "ftp" || config.StorageType == "sftp" {
			if host, ok := rawConfig["host"].(string); ok && config.Endpoint == "" {
				config.Endpoint = host
			}
			if username, ok := rawConfig["username"].(string); ok && config.AccessKey == "" {
				config.AccessKey = username
			}
			if password, ok := rawConfig["password"].(string); ok && config.SecretKey == "" {
				config.SecretKey = password
			}
			if path, ok := rawConfig["path"].(string); ok && config.BasePath == "" {
				config.BasePath = path
			}
		}

		// 阿里云 OSS 字段映射
		if config.StorageType == "oss" {
			if accessKeyID, ok := rawConfig["access_key_id"].(string); ok && config.AccessKey == "" {
				config.AccessKey = accessKeyID
			}
			if accessKeySecret, ok := rawConfig["access_key_secret"].(string); ok && config.SecretKey == "" {
				config.SecretKey = accessKeySecret
			}
			if bucket, ok := rawConfig["bucket"].(string); ok && config.Bucket == "" {
				config.Bucket = bucket
			}
			if endpoint, ok := rawConfig["endpoint"].(string); ok && config.Endpoint == "" {
				config.Endpoint = endpoint
			}
			if path, ok := rawConfig["path"].(string); ok && config.BasePath == "" {
				config.BasePath = path
			}
		}

		// 腾讯云 COS 字段映射
		if config.StorageType == "cos" {
			if secretID, ok := rawConfig["secret_id"].(string); ok && config.AccessKey == "" {
				config.AccessKey = secretID
			}
			if secretKey, ok := rawConfig["secret_key"].(string); ok && config.SecretKey == "" {
				config.SecretKey = secretKey
			}
			if bucket, ok := rawConfig["bucket"].(string); ok && config.Bucket == "" {
				config.Bucket = bucket
			}
			if region, ok := rawConfig["region"].(string); ok && config.Region == "" {
				config.Region = region
			}
			if path, ok := rawConfig["path"].(string); ok && config.BasePath == "" {
				config.BasePath = path
			}
		}

		// 七牛云字段映射
		if config.StorageType == "qiniu" {
			if accessKey, ok := rawConfig["access_key"].(string); ok && config.AccessKey == "" {
				config.AccessKey = accessKey
			}
			if secretKey, ok := rawConfig["secret_key"].(string); ok && config.SecretKey == "" {
				config.SecretKey = secretKey
			}
			if bucket, ok := rawConfig["bucket"].(string); ok && config.Bucket == "" {
				config.Bucket = bucket
			}
			if region, ok := rawConfig["region"].(string); ok && config.Region == "" {
				config.Region = region
			}
			if path, ok := rawConfig["path"].(string); ok && config.BasePath == "" {
				config.BasePath = path
			}
		}

		// AWS S3 字段映射
		if config.StorageType == "s3" {
			if accessKeyID, ok := rawConfig["access_key_id"].(string); ok && config.AccessKey == "" {
				config.AccessKey = accessKeyID
			}
			if secretAccessKey, ok := rawConfig["secret_access_key"].(string); ok && config.SecretKey == "" {
				config.SecretKey = secretAccessKey
			}
			if bucket, ok := rawConfig["bucket"].(string); ok && config.Bucket == "" {
				config.Bucket = bucket
			}
			if region, ok := rawConfig["region"].(string); ok && config.Region == "" {
				config.Region = region
			}
			if endpoint, ok := rawConfig["endpoint"].(string); ok && config.Endpoint == "" {
				config.Endpoint = endpoint
			}
			if path, ok := rawConfig["path"].(string); ok && config.BasePath == "" {
				config.BasePath = path
			}
		}

		// MinIO 字段映射
		if config.StorageType == "minio" {
			if endpoint, ok := rawConfig["endpoint"].(string); ok && config.Endpoint == "" {
				config.Endpoint = endpoint
			}
			if accessKey, ok := rawConfig["access_key"].(string); ok && config.AccessKey == "" {
				config.AccessKey = accessKey
			}
			if secretKey, ok := rawConfig["secret_key"].(string); ok && config.SecretKey == "" {
				config.SecretKey = secretKey
			}
			if bucket, ok := rawConfig["bucket"].(string); ok && config.Bucket == "" {
				config.Bucket = bucket
			}
			if useSSL, ok := rawConfig["use_ssl"].(bool); ok {
				config.UseSSL = useSSL
			}
			if path, ok := rawConfig["path"].(string); ok && config.BasePath == "" {
				config.BasePath = path
			}
		}

		// 又拍云字段映射
		if config.StorageType == "upyun" {
			if bucket, ok := rawConfig["bucket"].(string); ok && config.Bucket == "" {
				config.Bucket = bucket
			}
			if operator, ok := rawConfig["operator"].(string); ok && config.AccessKey == "" {
				config.AccessKey = operator
			}
			if password, ok := rawConfig["password"].(string); ok && config.SecretKey == "" {
				config.SecretKey = password
			}
			if path, ok := rawConfig["path"].(string); ok && config.BasePath == "" {
				config.BasePath = path
			}
		}
	}

	// 外链根地址：与后台「访问网址」(url) 一致；若仅配置了 domain（常见于 OSS/CDN），则用于生成图片链接
	config.URL = strings.TrimSpace(config.URL)
	config.Domain = strings.TrimSpace(config.Domain)
	if config.URL == "" && config.Domain != "" {
		config.URL = config.Domain
	}

	return &config, nil
}

// InitDefaultStrategies 初始化默认策略
func InitDefaultStrategies(db *gorm.DB) error {
	var count int64
	db.Model(&Strategies{}).Count(&count)
	if count == 0 {
		// 默认策略配置；访问根须为「域名/static」，与路由 GET /static/* 及 static 目录一致
		defaultConfig := StrategyConfig{
			URL:      "http://127.0.0.1:8080/static",
			Root:     "/static",
			Queries:  nil,
			AuthType: "1",
		}

		configJSON, err := json.Marshal(defaultConfig)
		if err != nil {
			return err
		}

		defaultStrategy := Strategies{
			Name:         "默认存储",
			Introduction: "本地默认存储",
			StrategyKey:  "1",
			Configs:      string(configJSON),
		}

		return db.Create(&defaultStrategy).Error
	}
	return nil
}

// StrategyStats 策略统计信息结构体（不包含时间字段）
type StrategyStats struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Introduction string  `json:"introduction"`
	StrategyKey  string  `json:"key"`
	Configs      string  `json:"configs"`
	ImageCount   int64   `json:"image_count"`
	TotalSize    int64   `json:"total_size"`
	UsedSizeMB   float64 `json:"used_size_mb"`
	Roles        []uint  `json:"roles"`
}

func GetstrategyConfig(db *gorm.DB) (*StrategyConfig, error) {
	var config StrategyConfig
	err := db.First(&config).Error
	return &config, err
}

// GetStrategyStats 获取策略统计信息
func GetStrategyStats(db *gorm.DB) ([]StrategyStats, error) {
	var strategies []Strategies
	err := db.Find(&strategies).Error
	if err != nil {
		return nil, err
	}

	var stats []StrategyStats
	for _, strategy := range strategies {
		// 获取该策略的图片数量
		var imageCount int64
		db.Model(&Images{}).Where("strategy_id = ?", strategy.ID).Count(&imageCount)

		// 获取该策略的总存储大小
		var totalSize int64
		db.Model(&Images{}).Where("strategy_id = ?", strategy.ID).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)

		// 计算MB
		usedSizeMB := float64(totalSize) / (1024 * 1024)

		stat := StrategyStats{
			ID:           strategy.ID,
			Name:         strategy.Name,
			Introduction: strategy.Introduction,
			StrategyKey:  strategy.StrategyKey,
			Configs:      strategy.Configs,
			ImageCount:   imageCount,
			TotalSize:    totalSize,
			UsedSizeMB:   usedSizeMB,
			Roles:        []uint{},
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

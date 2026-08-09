package model

import (
	"Slink/cache"
	"strconv"
	"sync"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Config 系统配置表
// 存储系统运行所需的配置项
type Config struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;comment:配置ID"`
	ConfigKey   string    `gorm:"type:varchar(191);unique;not null;column:config_key;comment:配置键名"`
	Value       string    `gorm:"not null;comment:配置值"`
	Description string    `gorm:"comment:配置描述"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

// 指定表名
func (Config) TableName() string {
	return "configs"
}

// CreateConfig 创建配置
func CreateConfig(db *gorm.DB, config *Config) error {
	return db.Create(config).Error
}

// GetConfigByID 根据ID获取配置
func GetConfigByID(db *gorm.DB, id uint) (*Config, error) {
	var config Config
	err := db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetConfigByKey 根据Key获取配置
func GetConfigByKey(db *gorm.DB, key string) (*Config, error) {
	var config Config
	res := db.Where("config_key = ?", key).Limit(1).Find(&config)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &config, nil
}

// GetAllConfigs 获取所有配置
func GetAllConfigs(db *gorm.DB) ([]Config, error) {
	var configs []Config
	err := db.Find(&configs).Error
	return configs, err
}

// UpdateConfig 更新配置
func UpdateConfig(db *gorm.DB, config *Config) error {
	return db.Save(config).Error
}

// UpdateConfigByKey 根据Key更新配置
func UpdateConfigByKey(db *gorm.DB, key string, value string) error {
	return db.Model(&Config{}).Where("config_key = ?", key).Update("value", value).Error
}

// DeleteConfigByKey 根据Key删除配置
func DeleteConfigByKey(db *gorm.DB, key string) error {
	return db.Where("config_key = ?", key).Delete(&Config{}).Error
}

// 配置读取缓存（两级）：
//   L1 进程内 sync.Map(5s TTL)——热路径零网络开销;
//   L2 全局 cache.GlobalCache——配置 Redis 时数据落 Redis,多实例共享。
// 图片访问等热路径每请求会读多次配置（如防盗链开关）,
// 直连 Postgres/MySQL 时每次往返都要毫秒级,缓存可消除。
const (
	configCacheKeyPrefix = "slink:cfg:"
	configCacheTTL       = 5 * time.Second
)

type configCacheEntry struct {
	value string
	at    time.Time
}

var configL1 sync.Map // key: string(configKey)

// InvalidateConfigCache 使配置缓存失效（配置更新时调用；key 为空表示全部失效）
func InvalidateConfigCache(key string) {
	if key == "" {
		configL1.Range(func(k, _ any) bool { configL1.Delete(k); return true })
		if cache.GlobalCache != nil {
			_ = cache.GlobalCache.Clear()
		}
		return
	}
	configL1.Delete(key)
	if cache.GlobalCache != nil {
		_ = cache.GlobalCache.Delete(configCacheKeyPrefix + key)
	}
}

// GetConfigValue 通过 key 获取配置值（用 Find 而非 First，避免键不存在时 GORM 打印 record not found）
func GetConfigValue(db *gorm.DB, key string) (string, error) {
	// L1:进程内
	if v, ok := configL1.Load(key); ok {
		e := v.(configCacheEntry)
		if time.Since(e.at) < configCacheTTL {
			return e.value, nil
		}
		configL1.Delete(key)
	}

	// L2:全局缓存（Redis/文件/内存）
	cacheKey := configCacheKeyPrefix + key
	if cache.GlobalCache != nil {
		if v, err := cache.GlobalCache.Get(cacheKey); err == nil && v != nil {
			if s, ok := v.(string); ok {
				configL1.Store(key, configCacheEntry{value: s, at: time.Now()})
				return s, nil
			}
		}
	}

	var cfg Config
	res := db.Where("config_key = ?", key).Limit(1).Find(&cfg)
	if res.Error != nil {
		return "", res.Error
	}
	if res.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}
	if cache.GlobalCache != nil {
		_ = cache.GlobalCache.Set(cacheKey, cfg.Value, configCacheTTL)
	}
	configL1.Store(key, configCacheEntry{value: cfg.Value, at: time.Now()})
	return cfg.Value, nil
}

// SetConfigValue 设置配置值
func SetConfigValue(db *gorm.DB, key string, value string) error {
	InvalidateConfigCache(key)
	var cfg Config
	res := db.Where("config_key = ?", key).Limit(1).Find(&cfg)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		cfg = Config{
			ConfigKey: key,
			Value:     value,
		}
		return db.Create(&cfg).Error
	}
	return db.Model(&cfg).Update("value", value).Error
}

// DeleteConfig 删除配置
func DeleteConfig(db *gorm.DB, id uint) error {
	return db.Delete(&Config{}, id).Error
}

// UpdateConfigs 批量更新配置
func UpdateConfigs(db *gorm.DB, configs []Config) error {
	InvalidateConfigCache("") // 全部失效
	for _, cfg := range configs {
		err := db.Model(&Config{}).Where("config_key = ?", cfg.ConfigKey).Updates(Config{Value: cfg.Value, Description: cfg.Description}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func InitConfigTable(db *gorm.DB) error {
	return db.AutoMigrate(&Config{})
}

// 默认配置项
var defaultConfigs = []Config{
	{ConfigKey: "default_storage_gb", Value: "5", Description: "默认存储空间(GB)"},
	{ConfigKey: "enable_register", Value: "false", Description: "是否开启注册（个人图床固定关闭）"},
	{ConfigKey: "enable_gallery", Value: "true", Description: "是否启用画廊"},
	{ConfigKey: "enable_api", Value: "true", Description: "是否启用接口"},
	{ConfigKey: "guest_upload", Value: "false", Description: "是否允许游客上传"},
	{ConfigKey: "email_verify", Value: "false", Description: "是否开启邮箱验证"},
	{ConfigKey: "enable_antihotlink", Value: "false", Description: "是否启用防盗链"},
	{ConfigKey: "antihotlink_domains", Value: "", Description: "防盗链允许的域名列表（逗号分隔）"},
	{ConfigKey: "antihotlink_allow_empty", Value: "true", Description: "是否允许空Referer"},
	{ConfigKey: "user_initial_capacity", Value: "5120000.00", Description: "新注册用户默认存储空间(字节,0表示无限制)"},
}

// GetDefaultUserCapacityBytes 读取"新注册用户默认存储空间"配置(单位:字节)
// 配置缺失或无法解析时返回 0(无限制);配置值兼容 "5120000.00" 这类小数字符串
func GetDefaultUserCapacityBytes(db *gorm.DB) uint64 {
	val, err := GetConfigValue(db, "user_initial_capacity")
	if err != nil {
		return 0
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if err != nil || f <= 0 {
		return 0
	}
	return uint64(f)
}

// InitDefaultConfigs 检查并初始化默认配置
func InitDefaultConfigs(db *gorm.DB) error {
	var count int64
	db.Model(&Config{}).Count(&count)
	if count == 0 {
		for _, cfg := range defaultConfigs {
			if err := db.Create(&cfg).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

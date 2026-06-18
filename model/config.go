package model

import (
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

// GetConfigValue 通过 key 获取配置值（用 Find 而非 First，避免键不存在时 GORM 打印 record not found）
func GetConfigValue(db *gorm.DB, key string) (string, error) {
	var cfg Config
	res := db.Where("config_key = ?", key).Limit(1).Find(&cfg)
	if res.Error != nil {
		return "", res.Error
	}
	if res.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}
	return cfg.Value, nil
}

// SetConfigValue 设置配置值
func SetConfigValue(db *gorm.DB, key string, value string) error {
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
}

// InitDefaultConfigs 检查并初始化默认配置。
// 首次安装（表为空）写入全部默认项；随后幂等补齐任何缺失的键，
// 使已有安装在版本升级后也能获得新增配置项（如 AI 设置）。
func InitDefaultConfigs(db *gorm.DB) error {
	var count int64
	db.Model(&Config{}).Count(&count)
	if count == 0 {
		for _, cfg := range defaultConfigs {
			c := cfg
			if err := db.Create(&c).Error; err != nil {
				return err
			}
		}
	} else {
		// 已有数据：仅补齐缺失键，不覆盖管理员已改的值
		for _, cfg := range defaultConfigs {
			var n int64
			if err := db.Model(&Config{}).Where("config_key = ?", cfg.ConfigKey).Count(&n).Error; err != nil {
				return err
			}
			if n == 0 {
				c := cfg
				if err := db.Create(&c).Error; err != nil {
					return err
				}
			}
		}
	}
	// 幂等补齐 AI 配置项
	return EnsureAIConfigDefaults(db)
}

package model

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UploadPolicyGroup 上传策略组
// 每个策略组定义独立的限流、命名规则、水印等配置，可分配给多个用户
type UploadPolicyGroup struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"size:191;not null" json:"name"`
	Description string `gorm:"size:512" json:"description"`

	// 限流
	LimitPerMinute      uint `gorm:"not null;default:0" json:"limit_per_minute"`
	LimitPerHour        uint `gorm:"not null;default:0" json:"limit_per_hour"`
	LimitPerDay         uint `gorm:"not null;default:0" json:"limit_per_day"`
	LimitPerWeek        uint `gorm:"not null;default:0" json:"limit_per_week"`
	LimitPerMonth       uint `gorm:"not null;default:0" json:"limit_per_month"`
	MaximumFileSize     uint `gorm:"not null;default:0" json:"maximum_file_size"`
	ConcurrentUploadNum uint `gorm:"not null;default:0" json:"concurrent_upload_num"`

	// 命名
	FileNamingRule string `gorm:"size:512;not null;default:''" json:"file_naming_rule"`
	PathNamingRule string `gorm:"size:512;not null;default:''" json:"path_naming_rule"`

	// 图片
	ImageSaveFormat  *string `gorm:"size:64" json:"image_save_format"`
	ImageSaveQuality uint    `gorm:"not null;default:0" json:"image_save_quality"`

	// 后缀
	AcceptedSuffixes string `gorm:"type:text;not null;default:''" json:"accepted_suffixes"`

	// 保护
	IsEnableOriginalProtection uint `gorm:"not null;default:0" json:"is_enable_original_protection"`
	ImageCacheTTL              uint `gorm:"not null;default:0" json:"image_cache_ttl"`

	// 水印
	IsEnableWatermark    uint   `gorm:"not null;default:0" json:"is_enable_watermark"`
	WatermarkConfigsJSON string `gorm:"type:text;not null;default:''" json:"-"`

	IsDefault uint `gorm:"not null;default:0;comment:是否默认策略组" json:"is_default"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UploadPolicyGroup) TableName() string {
	return "upload_policy_groups"
}

// ToGroupConfig 转换为 GroupConfig
func (g *UploadPolicyGroup) ToGroupConfig() *GroupConfig {
	c := &GroupConfig{
		LimitPerMinute:             g.LimitPerMinute,
		LimitPerHour:               g.LimitPerHour,
		LimitPerDay:                g.LimitPerDay,
		LimitPerWeek:               g.LimitPerWeek,
		LimitPerMonth:              g.LimitPerMonth,
		MaximumFileSize:            g.MaximumFileSize,
		ConcurrentUploadNum:        g.ConcurrentUploadNum,
		FileNamingRule:             g.FileNamingRule,
		PathNamingRule:             g.PathNamingRule,
		ImageSaveFormat:            g.ImageSaveFormat,
		ImageSaveQuality:           g.ImageSaveQuality,
		AcceptedFileSuffixes:       parseAcceptedSuffixesCSV(g.AcceptedSuffixes),
		IsEnableOriginalProtection: g.IsEnableOriginalProtection,
		ImageCacheTTL:              g.ImageCacheTTL,
		IsEnableWatermark:          g.IsEnableWatermark,
	}
	if strings.TrimSpace(g.WatermarkConfigsJSON) != "" {
		var wc WatermarkConfig
		if json.Unmarshal([]byte(g.WatermarkConfigsJSON), &wc) == nil {
			c.WatermarkConfigs = &wc
		}
	}
	return c
}

// SetGroupConfig 从 GroupConfig 写入
func (g *UploadPolicyGroup) SetGroupConfig(c *GroupConfig) {
	g.LimitPerMinute = c.LimitPerMinute
	g.LimitPerHour = c.LimitPerHour
	g.LimitPerDay = c.LimitPerDay
	g.LimitPerWeek = c.LimitPerWeek
	g.LimitPerMonth = c.LimitPerMonth
	g.MaximumFileSize = c.MaximumFileSize
	g.ConcurrentUploadNum = c.ConcurrentUploadNum
	g.FileNamingRule = c.FileNamingRule
	g.PathNamingRule = c.PathNamingRule
	g.ImageSaveFormat = c.ImageSaveFormat
	g.ImageSaveQuality = c.ImageSaveQuality
	g.AcceptedSuffixes = joinAcceptedSuffixesCSV(c.AcceptedFileSuffixes)
	g.IsEnableOriginalProtection = c.IsEnableOriginalProtection
	g.ImageCacheTTL = c.ImageCacheTTL
	g.IsEnableWatermark = c.IsEnableWatermark
	if c.WatermarkConfigs != nil {
		b, _ := json.Marshal(c.WatermarkConfigs)
		g.WatermarkConfigsJSON = string(b)
	} else {
		g.WatermarkConfigsJSON = ""
	}
}

// ---- DB operations ----

func GetUploadPolicyGroupByID(db *gorm.DB, id uint) (*UploadPolicyGroup, error) {
	var g UploadPolicyGroup
	err := db.First(&g, id).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func GetAllUploadPolicyGroups(db *gorm.DB) ([]UploadPolicyGroup, error) {
	var groups []UploadPolicyGroup
	err := db.Order("id ASC").Find(&groups).Error
	return groups, err
}

func CreateUploadPolicyGroup(db *gorm.DB, g *UploadPolicyGroup) error {
	return db.Create(g).Error
}

func UpdateUploadPolicyGroup(db *gorm.DB, g *UploadPolicyGroup) error {
	return db.Save(g).Error
}

func DeleteUploadPolicyGroup(db *gorm.DB, id uint) error {
	return db.Delete(&UploadPolicyGroup{}, id).Error
}

func GetDefaultUploadPolicyGroup(db *gorm.DB) (*UploadPolicyGroup, error) {
	var g UploadPolicyGroup
	err := db.Where("is_default = ?", 1).First(&g).Error
	return &g, err
}

// GetUserUploadPolicyGroup 获取用户所属的上传策略组
// 如果用户未分配策略组，返回默认策略组
func GetUserUploadPolicyGroup(db *gorm.DB, userID uint) (*UploadPolicyGroup, error) {
	user, err := GetUserByID(db, userID)
	if err != nil {
		return nil, err
	}
	return GetUploadPolicyGroupOfUser(db, user)
}

// GetUploadPolicyGroupOfUser 同 GetUserUploadPolicyGroup，但直接接受已查出的用户，
// 避免在已持有 user 对象的热路径（上传）上重复查询 users 表
func GetUploadPolicyGroupOfUser(db *gorm.DB, user *User) (*UploadPolicyGroup, error) {
	// 如果用户有分配策略组，使用用户的策略组
	if user.PolicyGroupID > 0 {
		pg, err := GetUploadPolicyGroupByID(db, user.PolicyGroupID)
		if err == nil {
			return pg, nil
		}
	}
	// 回退到默认策略组
	return GetDefaultUploadPolicyGroup(db)
}

// InitDefaultUploadPolicyGroup 初始化默认上传策略组
func InitDefaultUploadPolicyGroup(db *gorm.DB) error {
	var count int64
	db.Model(&UploadPolicyGroup{}).Count(&count)
	if count > 0 {
		return nil
	}

	defaultCfg := NewDefaultGroupConfig()
	g := &UploadPolicyGroup{
		Name:        "默认上传策略组",
		Description: "系统默认上传策略，新用户自动分配到此组",
		IsDefault:   1,
	}
	g.SetGroupConfig(&defaultCfg)
	return db.Create(g).Error
}

// MigrateLegacyUploadPolicy 将旧的 global_upload_policies 单例数据迁移到默认策略组
func MigrateLegacyUploadPolicy(db *gorm.DB) error {
	// 检查旧表是否存在
	if !db.Migrator().HasTable(&GlobalUploadPolicy{}) {
		return nil
	}

	var legacy GlobalUploadPolicy
	res := db.Where("id = ?", 1).Limit(1).Find(&legacy)
	if res.Error != nil || res.RowsAffected == 0 {
		return nil
	}

	gc, err := legacy.toGroupConfig()
	if err != nil {
		return nil
	}

	// 找默认策略组（第一个）
	var defaultGroup UploadPolicyGroup
	if err := db.Where("is_default = ?", 1).First(&defaultGroup).Error; err != nil {
		return nil
	}

	defaultGroup.SetGroupConfig(gc)
	return db.Save(&defaultGroup).Error
}

package model

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

const globalUploadPolicySingletonID uint = 1

// GlobalUploadPolicy 全局上传策略（单例记录，主键固定为 1）
type GlobalUploadPolicy struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	LimitPerMinute      uint `gorm:"not null;default:0"`
	LimitPerHour        uint `gorm:"not null;default:0"`
	LimitPerDay         uint `gorm:"not null;default:0"`
	LimitPerWeek        uint `gorm:"not null;default:0"`
	LimitPerMonth       uint `gorm:"not null;default:0"`
	MaximumFileSize     uint `gorm:"not null;default:0"`
	ConcurrentUploadNum uint `gorm:"not null;default:0"`

	FileNamingRule string `gorm:"size:512;not null;default:''"`
	PathNamingRule string `gorm:"size:512;not null;default:''"`

	ImageSaveFormat  *string `gorm:"size:64"`
	ImageSaveQuality uint    `gorm:"not null;default:0"`

	// AcceptedSuffixes 允许的后缀，逗号分隔、不含点，如 jpeg,jpg,png
	AcceptedSuffixes string `gorm:"type:text;not null;default:''"`

	IsEnableOriginalProtection uint `gorm:"not null;default:0"`
	ImageCacheTTL              uint `gorm:"not null;default:0"`

	IsEnableWatermark uint `gorm:"not null;default:0"`
	// WatermarkConfigsJSON 水印嵌套配置
	WatermarkConfigsJSON string `gorm:"type:text;not null;default:''"`
}

func (GlobalUploadPolicy) TableName() string {
	return "global_upload_policies"
}

func parseAcceptedSuffixesCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		p = strings.TrimPrefix(p, ".")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func joinAcceptedSuffixesCSV(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	b := strings.Builder{}
	for i, s := range ss {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strings.TrimSpace(strings.ToLower(strings.TrimPrefix(s, "."))))
	}
	return b.String()
}

// globalUploadPolicyFromGroupConfig 将 GroupConfig 转为表行（不含 ID / 时间戳）
func globalUploadPolicyFromGroupConfig(c *GroupConfig) (*GlobalUploadPolicy, error) {
	r := &GlobalUploadPolicy{
		LimitPerMinute:             c.LimitPerMinute,
		LimitPerHour:               c.LimitPerHour,
		LimitPerDay:                c.LimitPerDay,
		LimitPerWeek:               c.LimitPerWeek,
		LimitPerMonth:              c.LimitPerMonth,
		MaximumFileSize:            c.MaximumFileSize,
		ConcurrentUploadNum:        c.ConcurrentUploadNum,
		FileNamingRule:             c.FileNamingRule,
		PathNamingRule:             c.PathNamingRule,
		ImageSaveFormat:            c.ImageSaveFormat,
		ImageSaveQuality:           c.ImageSaveQuality,
		AcceptedSuffixes:           joinAcceptedSuffixesCSV(c.AcceptedFileSuffixes),
		IsEnableOriginalProtection: c.IsEnableOriginalProtection,
		ImageCacheTTL:              c.ImageCacheTTL,
		IsEnableWatermark:          c.IsEnableWatermark,
	}
	if c.WatermarkConfigs != nil {
		b, err := json.Marshal(c.WatermarkConfigs)
		if err != nil {
			return nil, err
		}
		r.WatermarkConfigsJSON = string(b)
	}
	return r, nil
}

func (r *GlobalUploadPolicy) toGroupConfig() (*GroupConfig, error) {
	c := &GroupConfig{
		LimitPerMinute:             r.LimitPerMinute,
		LimitPerHour:               r.LimitPerHour,
		LimitPerDay:                r.LimitPerDay,
		LimitPerWeek:               r.LimitPerWeek,
		LimitPerMonth:              r.LimitPerMonth,
		MaximumFileSize:            r.MaximumFileSize,
		ConcurrentUploadNum:        r.ConcurrentUploadNum,
		FileNamingRule:             r.FileNamingRule,
		PathNamingRule:             r.PathNamingRule,
		ImageSaveFormat:            r.ImageSaveFormat,
		ImageSaveQuality:           r.ImageSaveQuality,
		AcceptedFileSuffixes:       parseAcceptedSuffixesCSV(r.AcceptedSuffixes),
		IsEnableOriginalProtection: r.IsEnableOriginalProtection,
		ImageCacheTTL:              r.ImageCacheTTL,
		IsEnableWatermark:          r.IsEnableWatermark,
	}
	if strings.TrimSpace(r.WatermarkConfigsJSON) != "" {
		var wc WatermarkConfig
		if err := json.Unmarshal([]byte(r.WatermarkConfigsJSON), &wc); err != nil {
			return nil, err
		}
		c.WatermarkConfigs = &wc
	}
	return c, nil
}

// GetGlobalUploadPolicyRow 读取单例策略行
func GetGlobalUploadPolicyRow(db *gorm.DB) (*GlobalUploadPolicy, error) {
	var row GlobalUploadPolicy
	res := db.Where("id = ?", globalUploadPolicySingletonID).Limit(1).Find(&row)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// SaveGlobalUploadPolicyFromGroupConfig 用完整 GroupConfig 覆盖保存单例
func SaveGlobalUploadPolicyFromGroupConfig(db *gorm.DB, c *GroupConfig) error {
	row, err := globalUploadPolicyFromGroupConfig(c)
	if err != nil {
		return err
	}
	row.ID = globalUploadPolicySingletonID

	var existing GlobalUploadPolicy
	q := db.Where("id = ?", globalUploadPolicySingletonID).Limit(1).Find(&existing)
	if q.Error != nil {
		return q.Error
	}
	if q.RowsAffected == 0 {
		return db.Create(row).Error
	}
	row.CreatedAt = existing.CreatedAt
	return db.Save(row).Error
}

// EnsureGlobalUploadPolicy 创建默认行并从 configs.upload_policy / 默认组迁移一次
func EnsureGlobalUploadPolicy(db *gorm.DB) error {
	var legacy *GroupConfig
	if raw, err := GetConfigValue(db, ConfigKeyUploadPolicy); err == nil && strings.TrimSpace(raw) != "" {
		var c GroupConfig
		if json.Unmarshal([]byte(raw), &c) == nil {
			legacy = &c
		}
	}

	var row GlobalUploadPolicy
	res := db.Where("id = ?", globalUploadPolicySingletonID).Limit(1).Find(&row)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		base := NewDefaultGroupConfig()
		if legacy != nil {
			base = *legacy
		} else if g, err := GetDefaultGroup(db); err == nil {
			if gc, err := GetGroupConfig(g); err == nil {
				base = *gc
			}
		}
		newRow, err := globalUploadPolicyFromGroupConfig(&base)
		if err != nil {
			return err
		}
		newRow.ID = globalUploadPolicySingletonID
		if err := db.Create(newRow).Error; err != nil {
			return err
		}
		if legacy != nil {
			_ = DeleteConfigByKey(db, ConfigKeyUploadPolicy)
		}
		return nil
	}

	if legacy != nil {
		newRow, err := globalUploadPolicyFromGroupConfig(legacy)
		if err != nil {
			return err
		}
		newRow.ID = globalUploadPolicySingletonID
		newRow.CreatedAt = row.CreatedAt
		if err := db.Save(newRow).Error; err != nil {
			return err
		}
		_ = DeleteConfigByKey(db, ConfigKeyUploadPolicy)
	}
	return nil
}

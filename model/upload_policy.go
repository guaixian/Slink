package model

import (
	"errors"

	"gorm.io/gorm"
)

// ConfigKeyUploadPolicy 历史遗留：曾将策略存于 configs 表；迁移完成后该键会被删除
const ConfigKeyUploadPolicy = "upload_policy"

// GetGlobalUploadPolicy 读取全局上传策略（来自 global_upload_policies 单例表）
func GetGlobalUploadPolicy(db *gorm.DB) (*GroupConfig, error) {
	row, err := GetGlobalUploadPolicyRow(db)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			def := NewDefaultGroupConfig()
			return &def, nil
		}
		return nil, err
	}
	return row.toGroupConfig()
}

// EnsureUploadPolicySeeded 兼容旧调用名：确保单例表存在并完成自 configs / 默认组 的迁移
func EnsureUploadPolicySeeded(db *gorm.DB) error {
	return EnsureGlobalUploadPolicy(db)
}

// EffectiveImageGroupID 写入 images.group_id（仅兼容表结构；上传策略以 global_upload_policies 为准）
func EffectiveImageGroupID(db *gorm.DB) uint {
	if g, err := GetDefaultGroup(db); err == nil {
		return g.ID
	}
	return 1
}

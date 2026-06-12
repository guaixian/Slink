package model

import (
	"time"

	"gorm.io/gorm"
)

// Share 图片分享记录表
// 存储用户分享图片的记录，支持密码保护、过期时间等功能
type Share struct {
	ID        uint       `gorm:"primaryKey;autoIncrement;comment:分享ID" json:"id"`
	UserID    uint       `gorm:"not null;index;comment:用户ID" json:"user_id"`
	ImageID   uint       `gorm:"not null;index;comment:图片ID" json:"image_id"`
	ShareCode string     `gorm:"type:varchar(32);uniqueIndex;not null;comment:分享码" json:"share_code"`
	Password  string     `gorm:"type:varchar(255);comment:分享密码(加密存储)" json:"-"` // 密码（加密存储）
	ExpiresAt *time.Time `gorm:"comment:过期时间" json:"expires_at"`
	ViewCount int        `gorm:"default:0;comment:查看次数" json:"view_count"`         // 查看次数
	MaxViews  int        `gorm:"default:0;comment:最大查看次数 0表示无限制" json:"max_views"` // 最大查看次数（0表示无限制）
	IsActive  bool       `gorm:"default:true;comment:是否激活" json:"is_active"`       // 是否激活
	CreatedAt time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// CreateShare 创建分享
func CreateShare(db *gorm.DB, share *Share) error {
	return db.Create(share).Error
}

// GetShareByCode 根据分享码获取分享
func GetShareByCode(db *gorm.DB, code string) (*Share, error) {
	var share Share
	err := db.Where("share_code = ?", code).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// GetShareByID 根据ID获取分享
func GetShareByID(db *gorm.DB, id uint) (*Share, error) {
	var share Share
	err := db.First(&share, id).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// GetSharesByUserID 根据用户ID获取所有分享
func GetSharesByUserID(db *gorm.DB, userID uint) ([]Share, error) {
	var shares []Share
	err := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&shares).Error
	return shares, err
}

// GetSharesByImageID 根据图片ID获取所有分享
func GetSharesByImageID(db *gorm.DB, imageID uint) ([]Share, error) {
	var shares []Share
	err := db.Where("image_id = ?", imageID).Order("created_at DESC").Find(&shares).Error
	return shares, err
}

// UpdateShare 更新分享
func UpdateShare(db *gorm.DB, share *Share) error {
	return db.Save(share).Error
}

// DeleteShare 删除分享
func DeleteShare(db *gorm.DB, id uint) error {
	return db.Delete(&Share{}, id).Error
}

// IncrementShareViewCount 增加分享查看次数
func IncrementShareViewCount(db *gorm.DB, id uint) error {
	return db.Model(&Share{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// IsShareExpired 检查分享是否过期
func (s *Share) IsExpired() bool {
	if s.ExpiresAt != nil && s.ExpiresAt.Before(time.Now()) {
		return true
	}
	if s.MaxViews > 0 && s.ViewCount >= s.MaxViews {
		return true
	}
	return false
}

// IsShareValid 检查分享是否有效
func (s *Share) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

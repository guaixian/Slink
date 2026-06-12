package model

import (
	"time"

	"gorm.io/gorm"
)

// Images 图片信息表
// 存储所有上传图片的元数据信息
type Images struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;comment:图片ID"`
	UserID      uint      `gorm:"not null;comment:上传用户ID"`
	GroupID     uint      `gorm:"not null;comment:所属用户组ID"`
	StrategyID  uint      `gorm:"not null;comment:使用的存储策略ID"`
	ImageKey    string    `gorm:"not null;column:image_key;comment:图片唯一标识符"`
	Path        string    `gorm:"not null;comment:图片存储路径"`
	Name        string    `gorm:"not null;comment:图片文件名"`
	OriginName  string    `gorm:"not null;comment:原始文件名"`
	Size        int64     `gorm:"not null;comment:文件大小(字节)"`
	Mimetype    string    `gorm:"not null;comment:MIME类型"`
	Extension   string    `gorm:"not null;comment:文件扩展名"`
	Md5         string    `gorm:"not null;comment:文件MD5值"`
	SHA1        string    `gorm:"not null;comment:文件SHA1值"`
	Width       int64     `gorm:"not null;comment:图片宽度"`
	Height      int64     `gorm:"not null;comment:图片高度"`
	Permissions uint      `gorm:"not null;default:0;comment:访问权限 0-公开 1-私有"`
	IsUnhealthy uint      `gorm:"not null;default:0;comment:是否不健康 0-否 1-是"`
	UploadIp    string    `gorm:"not null;comment:上传IP地址"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

// 指定表名
func (Images) TableName() string {
	return "images"
}

// CreateImage 创建图片记录
func CreateImage(db *gorm.DB, image *Images) error {
	return db.Create(image).Error
}

// GetImageByID 根据ID获取图片
func GetImageByID(db *gorm.DB, id uint) (*Images, error) {
	var image Images
	err := db.First(&image, id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImageByKey 根据Key获取图片
func GetImageByKey(db *gorm.DB, key string) (*Images, error) {
	var image Images
	err := db.Where("image_key = ?", key).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImagesByUserID 根据用户ID获取图片列表
func GetImagesByUserID(db *gorm.DB, userID uint) ([]Images, error) {
	var images []Images
	err := db.Where("user_id = ?", userID).Find(&images).Error
	return images, err
}

// GetImagesByGroupID 根据组ID获取图片列表
func GetImagesByGroupID(db *gorm.DB, groupID uint) ([]Images, error) {
	var images []Images
	err := db.Where("group_id = ?", groupID).Find(&images).Error
	return images, err
}

// GetImagesByStrategyID 根据策略ID获取图片列表
func GetImagesByStrategyID(db *gorm.DB, strategyID uint) ([]Images, error) {
	var images []Images
	err := db.Where("strategy_id = ?", strategyID).Find(&images).Error
	return images, err
}

// GetAllImages 获取所有图片
func GetAllImages(db *gorm.DB) ([]Images, error) {
	var images []Images
	err := db.Find(&images).Error
	return images, err
}

// UpdateImage 更新图片记录
func UpdateImage(db *gorm.DB, image *Images) error {
	return db.Save(image).Error
}

// DeleteImage 删除图片记录
func DeleteImage(db *gorm.DB, id uint) error {
	return db.Delete(&Images{}, id).Error
}

// DeleteImagesByUserID 删除用户的所有图片
func DeleteImagesByUserID(db *gorm.DB, userID uint) error {
	return db.Where("user_id = ?", userID).Delete(&Images{}).Error
}

// GetImageCountByUserID 获取用户的图片数量
func GetImageCountByUserID(db *gorm.DB, userID uint) (int64, error) {
	var count int64
	err := db.Model(&Images{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetUnhealthyImages 获取不健康的图片
func GetUnhealthyImages(db *gorm.DB) ([]Images, error) {
	var images []Images
	err := db.Where("is_unhealthy = ?", 1).Find(&images).Error
	return images, err
}

// SetImagePermissionFromUserConfig 根据用户配置设置图片权限
func SetImagePermissionFromUserConfig(image *Images, userConfig UserConfig) {
	// 根据用户的默认权限设置图片权限
	// 0: 公开 (Public) - 图片将出现在画廊中
	// 1: 私有 (Private) - 图片不公开
	image.Permissions = uint(userConfig.DefaultPermission)
}

// CreateImageWithUserConfig 根据用户配置创建图片记录
func CreateImageWithUserConfig(db *gorm.DB, image *Images, userConfig UserConfig) error {
	// 根据用户配置设置图片权限
	SetImagePermissionFromUserConfig(image, userConfig)

	// 创建图片记录
	return CreateImage(db, image)
}

// GetPublicImages 获取公开的图片列表
func GetPublicImages(db *gorm.DB) ([]Images, error) {
	var images []Images
	err := db.Where("permissions = ?", PERMISSION_PUBLIC).Find(&images).Error
	return images, err
}

// GetPrivateImages 获取私有的图片列表
func GetPrivateImages(db *gorm.DB) ([]Images, error) {
	var images []Images
	err := db.Where("permissions = ?", PERMISSION_PRIVATE).Find(&images).Error
	return images, err
}

// GetPublicImagesByUserID 获取用户公开的图片列表
func GetPublicImagesByUserID(db *gorm.DB, userID uint) ([]Images, error) {
	var images []Images
	err := db.Where("user_id = ? AND permissions = ?", userID, PERMISSION_PUBLIC).Find(&images).Error
	return images, err
}

// GetPrivateImagesByUserID 获取用户私有的图片列表
func GetPrivateImagesByUserID(db *gorm.DB, userID uint) ([]Images, error) {
	var images []Images
	err := db.Where("user_id = ? AND permissions = ?", userID, PERMISSION_PRIVATE).Find(&images).Error
	return images, err
}

// GetTodayImageCountByUserID 获取用户今日上传的图片数量
func GetTodayImageCountByUserID(db *gorm.DB, userID uint) (int64, error) {
	var count int64
	now := time.Now()

	// 获取今天的开始时间（00:00:00）
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取明天的开始时间（00:00:00）
	tomorrowStart := todayStart.Add(24 * time.Hour)

	err := db.Model(&Images{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, todayStart, tomorrowStart).
		Count(&count).Error
	return count, err
}

// GetImageCountByUserIDInTimeRange 获取用户在指定时间段内上传的图片数量
func GetImageCountByUserIDInTimeRange(db *gorm.DB, userID uint, startTime, endTime time.Time) (int64, error) {
	var count int64
	err := db.Model(&Images{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startTime, endTime).
		Count(&count).Error
	return count, err
}

// GetTodayImageCountByUserIDWithTimezone 获取用户今日上传的图片数量（考虑时区）
func GetTodayImageCountByUserIDWithTimezone(db *gorm.DB, userID uint, timezone *time.Location) (int64, error) {
	var count int64

	// 如果未指定时区，使用系统本地时区
	if timezone == nil {
		timezone = time.Local
	}

	now := time.Now().In(timezone)

	// 获取今天的开始时间（00:00:00）
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, timezone)
	// 获取明天的开始时间（00:00:00）
	tomorrowStart := todayStart.Add(24 * time.Hour)

	err := db.Model(&Images{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, todayStart, tomorrowStart).
		Count(&count).Error
	return count, err
}

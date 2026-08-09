package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Groups 用户组表
// 定义用户分组，用于批量管理和权限控制
type Groups struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;comment:组ID"`
	Name      string    `gorm:"not null;comment:组名称"`
	IsDefault uint      `gorm:"not null;default:0;comment:是否默认组 0-否 1-是"`
	ISGuest   uint      `gorm:"not null;default:0;comment:是否游客组 0-否 1-是"`
	Configs   string    `gorm:"type:text;not null;comment:组配置(JSON格式)"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

// GroupConfig 全局上传策略的 API / 业务结构（持久化在 global_upload_policies 表；历史上曾存于 configs 与用户组）
type GroupConfig struct {
	// 上传限制
	LimitPerMinute uint `json:"limit_per_minute"`
	LimitPerHour   uint `json:"limit_per_hour"`
	LimitPerDay    uint `json:"limit_per_day"`
	LimitPerWeek   uint `json:"limit_per_week"`
	LimitPerMonth  uint `json:"limit_per_month"`

	// 文件大小和并发
	MaximumFileSize     uint `json:"maximum_file_size"`
	ConcurrentUploadNum uint `json:"concurrent_upload_num"`

	// 文件命名规则
	FileNamingRule string `json:"file_naming_rule"`
	PathNamingRule string `json:"path_naming_rule"`

	// 图片保存设置
	ImageSaveFormat  *string `json:"image_save_format"`
	ImageSaveQuality uint    `json:"image_save_quality"`

	// 接受的文件类型
	AcceptedFileSuffixes []string `json:"accepted_file_suffixes"`

	// 原图保护
	IsEnableOriginalProtection uint `json:"is_enable_original_protection"`
	ImageCacheTTL              uint `json:"image_cache_ttl"`

	// 水印配置
	IsEnableWatermark uint             `json:"is_enable_watermark"`
	WatermarkConfigs  *WatermarkConfig `json:"watermark_configs"`
}

// WatermarkFontDriver 文字水印（字体文件置于 static/，填相对文件名如 2.ttf）
type WatermarkFontDriver struct {
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Font     string  `json:"font"`
	Size     float64 `json:"size"`
	Text     string  `json:"text"`
	Angle    int     `json:"angle"`
	Color    string  `json:"color"`
	Position string  `json:"position"`
}

// WatermarkImageDriver 图片水印（水印图置于 static/，填相对路径）
type WatermarkImageDriver struct {
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Image    string  `json:"image"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Rotate   int     `json:"rotate"`
	Opacity  float64 `json:"opacity"`
	Position string  `json:"position"`
}

// WatermarkDrivers 各驱动参数（与前端表单、旧版 JSON 的 drivers 对象结构一致）
type WatermarkDrivers struct {
	Font  *WatermarkFontDriver  `json:"font,omitempty"`
	Image *WatermarkImageDriver `json:"image,omitempty"`
}

// WatermarkConfig 水印配置
type WatermarkConfig struct {
	Mode    uint              `json:"mode"`
	Driver  string            `json:"driver"`
	Drivers *WatermarkDrivers `json:"drivers,omitempty"`
}

// NormalizeWatermarkConfig 补全水印嵌套结构，保证写入 global_upload_policies 时 JSON 完整
func NormalizeWatermarkConfig(wc *WatermarkConfig) {
	if wc == nil {
		return
	}
	if wc.Driver == "" {
		wc.Driver = "font"
	}
	if wc.Drivers == nil {
		wc.Drivers = &WatermarkDrivers{}
	}
	if wc.Drivers.Font == nil {
		wc.Drivers.Font = &WatermarkFontDriver{Size: 24, Color: "#ffffff", Position: "bottom-right"}
	}
	if wc.Drivers.Image == nil {
		wc.Drivers.Image = &WatermarkImageDriver{Opacity: 100, Position: "bottom-right"}
	}
}

// GetGroupConfig 获取用户组配置
func GetGroupConfig(group *Groups) (*GroupConfig, error) {
	var config GroupConfig
	err := json.Unmarshal([]byte(group.Configs), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateGroup 创建用户组
func CreateGroup(db *gorm.DB, group *Groups) error {
	return db.Create(group).Error
}

// GetGroupByID 根据ID获取用户组
func GetGroupByID(db *gorm.DB, id uint) (*Groups, error) {
	var group Groups
	err := db.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetGroupByName 根据名称获取用户组
func GetGroupByName(db *gorm.DB, name string) (*Groups, error) {
	var group Groups
	err := db.Where("name = ?", name).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetAllGroups 获取所有用户组
func GetAllGroups(db *gorm.DB) ([]Groups, error) {
	var groups []Groups
	err := db.Find(&groups).Error
	return groups, err
}

// UpdateGroup 更新用户组
func UpdateGroup(db *gorm.DB, group *Groups) error {
	return db.Save(group).Error
}

// DeleteGroup 删除用户组
func DeleteGroup(db *gorm.DB, id uint) error {
	return db.Delete(&Groups{}, id).Error
}

// GetDefaultGroup 获取默认用户组
func GetDefaultGroup(db *gorm.DB) (*Groups, error) {
	var group Groups
	err := db.Where("is_default = ?", 1).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetGuestGroup 获取游客用户组
func GetGuestGroup(db *gorm.DB) (*Groups, error) {
	var group Groups
	err := db.Where("is_guest = ?", 1).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// NewDefaultGroupConfig 系统默认上传/命名/限流策略（原「默认用户组」配置，现亦用于全局 upload_policy 默认值）
func NewDefaultGroupConfig() GroupConfig {
	return GroupConfig{
		LimitPerMinute:             99999,
		LimitPerHour:               99997,
		LimitPerDay:                99999,
		LimitPerWeek:               99999,
		LimitPerMonth:              99999,
		MaximumFileSize:            15120,
		ConcurrentUploadNum:        10,
		FileNamingRule:             "{uniqid}",
		PathNamingRule:             "{Y}/{m}/{d}",
		ImageSaveFormat:            nil,
		ImageSaveQuality:           0, // 0=不处理,原样存储(与原图 MD5 一致);1~99 仅对 JPEG/格式转换生效
		AcceptedFileSuffixes:       []string{"jpeg", "jpg", "png", "gif", "tif", "bmp", "ico", "psd", "webp"},
		IsEnableOriginalProtection: 0,
		ImageCacheTTL:              2626560,
		IsEnableWatermark:          0,
		WatermarkConfigs: &WatermarkConfig{
			Mode:   1,
			Driver: "font",
			Drivers: &WatermarkDrivers{
				Font: &WatermarkFontDriver{
					X: 10, Y: 10,
					Font: "2.ttf", Size: 50,
					Text: "image.darkmagic.site", Angle: 0,
					Color: "#ffffff", Position: "bottom-right",
				},
				Image: &WatermarkImageDriver{
					X: 10, Y: 10,
					Image: "", Width: 0, Height: 0, Rotate: 0,
					Opacity: 100, Position: "bottom-right",
				},
			},
		},
	}
}

// InitDefaultGroups 初始化默认用户组（保留表结构兼容；上传策略以 global_upload_policies 表为准）
func InitDefaultGroups(db *gorm.DB) error {
	var count int64
	db.Model(&Groups{}).Count(&count)
	if count == 0 {
		defaultConfig := NewDefaultGroupConfig()

		configJSON, err := json.Marshal(defaultConfig)
		if err != nil {
			return err
		}

		defaultGroup := Groups{
			Name:      "系统默认组&游客组",
			IsDefault: 1,
			ISGuest:   1,
			Configs:   string(configJSON),
		}

		return db.Create(&defaultGroup).Error
	}
	return nil
}

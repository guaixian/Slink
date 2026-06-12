package model

import (
	"Slink/utils"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// User 用户表
// 存储系统用户信息，包括管理员和普通用户
type User struct {
	ID      uint   `gorm:"primaryKey;autoIncrement;comment:用户ID"`
	GroupID uint   `gorm:"not null;default:1;comment:用户组ID"`
	Name    string `gorm:"not null;default:'集帅';comment:用户名称"`
	// Email 列名历史原因保留；存的是登录账号（任意非空字符串），不再要求邮箱格式
	Email        string    `gorm:"uniqueIndex:idx_email;size:191;not null;comment:登录账号"`
	Password     string    `gorm:"not null;comment:加密后的密码"`
	IsAdmin      uint      `gorm:"not null;default:0;comment:是否管理员 0-否 1-是"`
	Capacity     uint      `gorm:"not null;default:0;comment:存储容量限制(字节)"`
	Configs      string    `gorm:"type:text;not null;comment:用户配置(JSON格式)"`
	ImageNums    uint      `gorm:"not null;default:0;comment:图片数量"`
	RegisteredIP string    `gorm:"not null;default:'';comment:注册IP地址"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

// CreateUser 创建用户
func CreateUser(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}

// GetUserByID 根据ID获取用户
func GetUserByID(db *gorm.DB, id uint) (*User, error) {
	var user User
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据登录账号获取用户（列名仍为 email）
func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByLoginAccount 同 GetUserByEmail，语义为登录名
func GetUserByLoginAccount(db *gorm.DB, account string) (*User, error) {
	return GetUserByEmail(db, account)
}

// GetUsersByGroupID 根据组ID获取用户列表
func GetUsersByGroupID(db *gorm.DB, groupID uint) ([]User, error) {
	var users []User
	err := db.Where("group_id = ?", groupID).Find(&users).Error
	return users, err
}

// GetAllUsers 获取所有用户
func GetAllUsers(db *gorm.DB) ([]User, error) {
	var users []User
	err := db.Find(&users).Error
	return users, err
}

// UpdateUser 更新用户
func UpdateUser(db *gorm.DB, user *User) error {
	return db.Save(user).Error
}

// DeleteUser 删除用户
func DeleteUser(db *gorm.DB, id uint) error {
	return db.Delete(&User{}, id).Error
}

// GetAdminUsers 获取所有管理员用户
func GetAdminUsers(db *gorm.DB) ([]User, error) {
	var users []User
	err := db.Where("is_admin = ?", 1).Find(&users).Error
	return users, err
}

// UpdateUserImageCount 更新用户图片数量
func UpdateUserImageCount(db *gorm.DB, userID uint, count uint) error {
	return db.Model(&User{}).Where("id = ?", userID).Update("image_nums", count).Error
}

// IncrementUserImageCount 增加用户图片数量
func IncrementUserImageCount(db *gorm.DB, userID uint) error {
	return db.Model(&User{}).Where("id = ?", userID).UpdateColumn("image_nums", gorm.Expr("image_nums + ?", 1)).Error
}

// DecrementUserImageCount 减少用户图片数量
func DecrementUserImageCount(db *gorm.DB, userID uint) error {
	return db.Model(&User{}).Where("id = ?", userID).UpdateColumn("image_nums", gorm.Expr("image_nums - ?", 1)).Error
}

// GetUserConfig 获取用户配置
func (u *User) GetUserConfig() (UserConfig, error) {
	return ParseUserConfig(u.Configs)
}

// SetUserConfig 设置用户配置
func (u *User) SetUserConfig(config UserConfig) error {
	configJSON, err := config.ToJSON()
	if err != nil {
		return err
	}
	u.Configs = configJSON
	return nil
}

// GetDefaultUserConfig 获取默认用户配置
func (u *User) GetDefaultUserConfig() UserConfig {
	return GetDefaultUserConfig()
}

// InitializeUserConfig 初始化用户配置为默认值
func (u *User) InitializeUserConfig() error {
	defaultConfig := GetDefaultUserConfig()
	return u.SetUserConfig(defaultConfig)
}

// GetBackendConfig 获取用户的后台配置
func (u *User) GetBackendConfig() (map[string]interface{}, error) {
	config, err := u.GetUserConfig()
	if err != nil {
		return nil, err
	}
	return config.GetBackendConfig(), nil
}

// GetFrontendConfig 获取用户的前端配置
func (u *User) GetFrontendConfig() (map[string]interface{}, error) {
	config, err := u.GetUserConfig()
	if err != nil {
		return nil, err
	}
	return config.GetFrontendConfig(), nil
}

// ValidateUserBackendConfig 验证用户后台配置
func (u *User) ValidateUserBackendConfig() error {
	config, err := u.GetUserConfig()
	if err != nil {
		return err
	}
	return config.ValidateBackendConfig()
}

// GetUserCount 获取用户总数
func GetUserCount(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&User{}).Count(&count).Error
	return count, err
}

// migrateExistingUsers 迁移现有用户到默认用户组
func migrateExistingUsers(db *gorm.DB) error {
	// 更新所有GroupID为0的用户到GroupID为1
	return db.Model(&User{}).Where("group_id = ?", 0).Update("group_id", 1).Error
}

func InitDB() error {
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db
	if err := db.AutoMigrate(&User{}); err != nil {
		return err
	}
	// 自动迁移 Config 表
	if err := db.AutoMigrate(&Config{}); err != nil {
		return err
	}
	// 自动迁移其他表
	if err := db.AutoMigrate(&Groups{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&PersonalAccessToken{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&GroupStrategy{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&Strategies{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&Images{}); err != nil {
		return err
	}
	// 自动迁移 Share 表
	if err := db.AutoMigrate(&Share{}); err != nil {
		return err
	}
	// 初始化默认配置
	if err := InitDefaultConfigs(db); err != nil {
		return err
	}
	// 初始化默认用户组
	if err := InitDefaultGroups(db); err != nil {
		return err
	}
	// 初始化默认策略
	if err := InitDefaultStrategies(db); err != nil {
		return err
	}
	// 迁移现有用户
	if err := migrateExistingUsers(db); err != nil {
		return err
	}
	// 创建默认管理员账号
	if err := createDefaultAdmin(db); err != nil {
		return err
	}
	return nil
}

// createDefaultAdmin 创建默认管理员账号
func createDefaultAdmin(db *gorm.DB) error {
	var count int64
	db.Model(&User{}).Where("email = ?", "admin@slink.org").Count(&count)
	if count == 0 {
		hash, err := utils.HashPassword("123456")
		if err != nil {
			return err
		}

		// 创建默认配置
		defaultConfig := GetDefaultUserConfig()
		configJSON, err := defaultConfig.ToJSON()
		if err != nil {
			return err
		}

		admin := User{
			Email:        "admin@slink.org",
			Password:     hash,
			GroupID:      1, // 直接设置为默认用户组ID
			IsAdmin:      1,
			Configs:      configJSON,  // 设置默认配置
			RegisteredIP: "127.0.0.1", // 默认管理员IP
		}
		return db.Create(&admin).Error
	}
	return nil
}

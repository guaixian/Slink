package model

import (
	"time"

	"gorm.io/gorm"
)

// PersonalAccessToken 个人访问令牌表
// 存储用户的API访问令牌，用于程序化访问
type PersonalAccessToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;comment:令牌ID"`
	ApiName   string    `gorm:"not null;comment:API名称"`
	Username  string    `gorm:"not null;comment:所属用户名"`
	Token     string    `gorm:"type:varchar(255);not null;comment:访问令牌"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

// CreatePersonalAccessToken 创建个人访问令牌
func CreatePersonalAccessToken(db *gorm.DB, token *PersonalAccessToken) error {
	return db.Create(token).Error
}

// GetPersonalAccessTokenByID 根据ID获取个人访问令牌
func GetPersonalAccessTokenByID(db *gorm.DB, id uint) (*PersonalAccessToken, error) {
	var token PersonalAccessToken
	err := db.First(&token, id).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetPersonalAccessTokenByToken 根据令牌获取个人访问令牌
func GetPersonalAccessTokenByToken(db *gorm.DB, tokenStr string) (*PersonalAccessToken, error) {
	var token PersonalAccessToken
	err := db.Where("token = ?", tokenStr).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetPersonalAccessTokensByUsername 根据用户名获取所有个人访问令牌
func GetPersonalAccessTokensByUsername(db *gorm.DB, username string) ([]PersonalAccessToken, error) {
	var tokens []PersonalAccessToken
	err := db.Where("username = ?", username).Find(&tokens).Error
	return tokens, err
}

// GetAllPersonalAccessTokens 获取所有个人访问令牌
func GetAllPersonalAccessTokens(db *gorm.DB) ([]PersonalAccessToken, error) {
	var tokens []PersonalAccessToken
	err := db.Find(&tokens).Error
	return tokens, err
}

// UpdatePersonalAccessToken 更新个人访问令牌
func UpdatePersonalAccessToken(db *gorm.DB, token *PersonalAccessToken) error {
	return db.Save(token).Error
}

// DeletePersonalAccessToken 删除个人访问令牌
func DeletePersonalAccessToken(db *gorm.DB, id uint) error {
	return db.Delete(&PersonalAccessToken{}, id).Error
}

// DeletePersonalAccessTokensByUsername 删除用户的所有个人访问令牌
func DeletePersonalAccessTokensByUsername(db *gorm.DB, username string) error {
	return db.Where("username = ?", username).Delete(&PersonalAccessToken{}).Error
}

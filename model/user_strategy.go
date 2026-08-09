package model

import (
	"time"

	"gorm.io/gorm"
)

// UserStrategy 用户与存储策略关联表
// 实现用户和存储策略的1:N关系（一个策略可分配给多个用户）
type UserStrategy struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;comment:关联ID"`
	UserID     uint      `gorm:"not null;index:idx_user_strategy;comment:用户ID"`
	StrategyID uint      `gorm:"not null;index:idx_user_strategy;uniqueIndex:idx_user_strategy_unique;comment:存储策略ID"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间"`
}

func (UserStrategy) TableName() string {
	return "user_strategies"
}

// GetStrategyIDsByUserID 获取用户关联的所有策略ID
func GetStrategyIDsByUserID(db *gorm.DB, userID uint) ([]uint, error) {
	var strategyIDs []uint
	err := db.Model(&UserStrategy{}).Where("user_id = ?", userID).Pluck("strategy_id", &strategyIDs).Error
	return strategyIDs, err
}

// GetUserIDsByStrategyID 获取策略关联的所有用户ID
func GetUserIDsByStrategyID(db *gorm.DB, strategyID uint) ([]uint, error) {
	var userIDs []uint
	err := db.Model(&UserStrategy{}).Where("strategy_id = ?", strategyID).Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// AddStrategyToUser 为用户分配策略
func AddStrategyToUser(db *gorm.DB, userID, strategyID uint) error {
	us := UserStrategy{
		UserID:     userID,
		StrategyID: strategyID,
	}
	return db.Where(UserStrategy{UserID: userID, StrategyID: strategyID}).FirstOrCreate(&us).Error
}

// RemoveStrategyFromUser 移除用户的策略
func RemoveStrategyFromUser(db *gorm.DB, userID, strategyID uint) error {
	return db.Where("user_id = ? AND strategy_id = ?", userID, strategyID).Delete(&UserStrategy{}).Error
}

// DeleteUserStrategiesByUserID 删除用户的所有策略关联
func DeleteUserStrategiesByUserID(db *gorm.DB, userID uint) error {
	return db.Where("user_id = ?", userID).Delete(&UserStrategy{}).Error
}

// DeleteUserStrategiesByStrategyID 删除策略的所有用户关联
func DeleteUserStrategiesByStrategyID(db *gorm.DB, strategyID uint) error {
	return db.Where("strategy_id = ?", strategyID).Delete(&UserStrategy{}).Error
}

// GetUserStrategies 获取用户关联的所有策略详情
func GetUserStrategies(db *gorm.DB, userID uint) ([]Strategies, error) {
	var strategyIDs []uint
	if err := db.Model(&UserStrategy{}).Where("user_id = ?", userID).Pluck("strategy_id", &strategyIDs).Error; err != nil {
		return nil, err
	}
	if len(strategyIDs) == 0 {
		return nil, nil
	}
	var strategies []Strategies
	err := db.Where("id IN ?", strategyIDs).Find(&strategies).Error
	return strategies, err
}

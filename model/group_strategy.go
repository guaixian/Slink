package model

import "gorm.io/gorm"

// GroupStrategy 用户组与存储策略关联表
// 实现用户组和存储策略的多对多关系
type GroupStrategy struct {
	GroupID    uint  `gorm:"primaryKey;comment:用户组ID"`
	StrategyID uint  `gorm:"primaryKey;comment:存储策略ID"`
	CreatedAt  int64 `gorm:"comment:创建时间"`
	UpdatedAt  int64 `gorm:"comment:更新时间"`
}

func (GroupStrategy) TableName() string {
	return "group_strategies"
}

func (GroupStrategy) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func GetGroupIDsByStrategyID(db *gorm.DB, strategyID uint) ([]uint, error) {
	var groupIDs []uint
	err := db.Table("group_strategies").Where("strategy_id = ?", strategyID).Pluck("group_id", &groupIDs).Error
	return groupIDs, err
}

func GetStrategyIDsByGroupID(db *gorm.DB, groupID uint) ([]uint, error) {
	var strategyIDs []uint
	err := db.Table("group_strategies").Where("group_id = ?", groupID).Pluck("strategy_id", &strategyIDs).Error
	return strategyIDs, err
}

func AddStrategyToGroup(db *gorm.DB, groupID, strategyID uint) error {
	gs := GroupStrategy{
		GroupID:    groupID,
		StrategyID: strategyID,
	}
	return db.Table("group_strategies").FirstOrCreate(&gs, gs).Error
}

func RemoveStrategyFromGroup(db *gorm.DB, groupID, strategyID uint) error {
	return db.Table("group_strategies").Where("group_id = ? AND strategy_id = ?", groupID, strategyID).Delete(&GroupStrategy{}).Error
}

func DeleteGroupStrategiesByStrategyID(db *gorm.DB, strategyID uint) error {
	return db.Table("group_strategies").Where("strategy_id = ?", strategyID).Delete(&GroupStrategy{}).Error
}

func InitGroupStrategies(db *gorm.DB) error {
	return db.AutoMigrate(&GroupStrategy{})
}

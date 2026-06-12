package model

import (
	"fmt"

	"gorm.io/gorm"
)

// GetStrategyForStoredImage 按图片记录上的 strategy_id 解析策略；无效时回退到 ID 最小的已配置策略
func GetStrategyForStoredImage(db *gorm.DB, strategyID uint) (*Strategies, error) {
	if strategyID > 0 {
		if s, err := GetStrategyByID(db, strategyID); err == nil {
			return s, nil
		}
	}
	list, err := GetAllStrategies(db)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("未配置存储策略")
	}
	return GetStrategyByID(db, list[0].ID)
}

package utils

// GroupConfig 用户组配置结构（简化版本，避免循环导入）
type GroupConfig struct {
	LimitPerMinute uint `json:"limit_per_minute"`
	LimitPerHour   uint `json:"limit_per_hour"`
	LimitPerDay    uint `json:"limit_per_day"`
	LimitPerWeek   uint `json:"limit_per_week"`
	LimitPerMonth  uint `json:"limit_per_month"`
}

// CheckRateLimit 已不再启用上传频率限制，始终通过。
func CheckRateLimit(userID uint, groupConfig *GroupConfig) error {
	_, _ = userID, groupConfig
	return nil
}

// GetRateLimitInfo 返回策略中的限额配置（仅供参考）；实际上传不做限流。
func GetRateLimitInfo(userID uint, groupConfig *GroupConfig) map[string]interface{} {
	_ = userID
	if groupConfig == nil {
		groupConfig = &GroupConfig{}
	}
	return map[string]interface{}{
		"enabled": false,
		"message": "上传频率限制已关闭",
		"per_minute": map[string]interface{}{
			"current": uint(0),
			"limit":   groupConfig.LimitPerMinute,
		},
		"per_hour": map[string]interface{}{
			"current": uint(0),
			"limit":   groupConfig.LimitPerHour,
		},
		"per_day": map[string]interface{}{
			"current": uint(0),
			"limit":   groupConfig.LimitPerDay,
		},
		"per_week": map[string]interface{}{
			"current": uint(0),
			"limit":   groupConfig.LimitPerWeek,
		},
		"per_month": map[string]interface{}{
			"current": uint(0),
			"limit":   groupConfig.LimitPerMonth,
		},
	}
}

// ResetRateLimit 无操作（兼容旧管理接口）。
func ResetRateLimit(userID uint) error {
	_ = userID
	return nil
}

// GetRateLimitStats 无 Redis 时的占位响应。
func GetRateLimitStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":    false,
		"total_keys": 0,
		"keys":       []string{},
	}
}

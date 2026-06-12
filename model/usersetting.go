package model

import (
	"encoding/json"
	"fmt"
)

// UserConfig 用户配置结构
type UserConfig struct {
	// 默认上传策略 (Default upload strategy) - 后台处理
	// 1: 本地存储 (Local Storage)
	// 其他值: 其他存储策略
	DefaultStrategy int `json:"default_strategy"`

	// 默认图片权限 (Default image permission) - 后台处理
	// 0: 公开 (Public) - 图片将出现在画廊中
	// 1: 私有 (Private) - 图片不公开
	DefaultPermission int `json:"default_permission"`

	// 粘贴图片后的动作 (Action after pasting an image) - 前端处理
	// 0: 直接上传 (Direct upload)
	// 1: 等待上传 (Wait for upload)
	PastedAction int `json:"pasted_action"`

	// 是否自动清除预览图片 (Whether to automatically clear preview images) - 前端处理
	// 0: 否 (No)
	// 1: 是 (Yes)
	IsAutoClearPreview int `json:"is_auto_clear_preview"`
}

// 配置常量定义
const (
	// 粘贴动作常量
	PASTED_ACTION_DIRECT_UPLOAD = 0 // 直接上传
	PASTED_ACTION_WAIT_UPLOAD   = 1 // 等待上传

	// 默认策略常量
	DEFAULT_STRATEGY_LOCAL = 1 // 本地存储

	// 权限常量
	PERMISSION_PUBLIC  = 0 // 公开
	PERMISSION_PRIVATE = 1 // 私有

	// 自动清除预览常量
	AUTO_CLEAR_PREVIEW_NO  = 0 // 否
	AUTO_CLEAR_PREVIEW_YES = 1 // 是
)

// GetDefaultUserConfig 获取默认用户配置
func GetDefaultUserConfig() UserConfig {
	return UserConfig{
		PastedAction:       1, // 1: 等待上传
		DefaultStrategy:    1, // 1: 本地存储
		DefaultPermission:  0, // 0: 公开
		IsAutoClearPreview: 0, // 0: 否
	}
}

// ParseUserConfig 解析用户配置字符串
func ParseUserConfig(configStr string) (UserConfig, error) {
	var config UserConfig
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		// 如果解析失败，返回默认配置
		return GetDefaultUserConfig(), fmt.Errorf("failed to parse user config: %v", err)
	}
	return config, nil
}

// ToJSON 将配置转换为JSON字符串
func (c *UserConfig) ToJSON() (string, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user config: %v", err)
	}
	return string(bytes), nil
}

// GetPastedActionText 获取粘贴动作的文本描述
func (c *UserConfig) GetPastedActionText() string {
	switch c.PastedAction {
	case PASTED_ACTION_DIRECT_UPLOAD:
		return "直接上传"
	case PASTED_ACTION_WAIT_UPLOAD:
		return "等待上传"
	default:
		return "未知"
	}
}

// GetDefaultStrategyText 获取默认策略的文本描述
func (c *UserConfig) GetDefaultStrategyText() string {
	switch c.DefaultStrategy {
	case DEFAULT_STRATEGY_LOCAL:
		return "本地存储"
	default:
		return "未知策略"
	}
}

// GetDefaultPermissionText 获取默认权限的文本描述
func (c *UserConfig) GetDefaultPermissionText() string {
	switch c.DefaultPermission {
	case PERMISSION_PUBLIC:
		return "公开"
	case PERMISSION_PRIVATE:
		return "私有"
	default:
		return "未知"
	}
}

// GetAutoClearPreviewText 获取自动清除预览的文本描述
func (c *UserConfig) GetAutoClearPreviewText() string {
	switch c.IsAutoClearPreview {
	case AUTO_CLEAR_PREVIEW_YES:
		return "是"
	case AUTO_CLEAR_PREVIEW_NO:
		return "否"
	default:
		return "未知"
	}
}

// GetBackendConfig 获取后台处理的配置
func (c *UserConfig) GetBackendConfig() map[string]interface{} {
	return map[string]interface{}{
		"default_strategy":   c.DefaultStrategy,
		"default_permission": c.DefaultPermission,
	}
}

// GetFrontendConfig 获取前端处理的配置
func (c *UserConfig) GetFrontendConfig() map[string]interface{} {
	return map[string]interface{}{
		"pasted_action":         c.PastedAction,
		"is_auto_clear_preview": c.IsAutoClearPreview,
	}
}

// ValidateBackendConfig 验证后台配置是否有效
func (c *UserConfig) ValidateBackendConfig() error {
	// 验证默认策略 (后台处理)
	if c.DefaultStrategy < 1 {
		return fmt.Errorf("invalid default_strategy: %d", c.DefaultStrategy)
	}

	// 验证默认权限 (后台处理)
	if c.DefaultPermission != PERMISSION_PUBLIC && c.DefaultPermission != PERMISSION_PRIVATE {
		return fmt.Errorf("invalid default_permission: %d", c.DefaultPermission)
	}

	return nil
}

// ValidateConfig 验证配置是否有效
func (c *UserConfig) ValidateConfig() error {
	// 验证默认策略 (后台处理)
	if c.DefaultStrategy < 1 {
		return fmt.Errorf("invalid default_strategy: %d", c.DefaultStrategy)
	}

	// 验证默认权限 (后台处理)
	if c.DefaultPermission != PERMISSION_PUBLIC && c.DefaultPermission != PERMISSION_PRIVATE {
		return fmt.Errorf("invalid default_permission: %d", c.DefaultPermission)
	}

	// 验证粘贴动作 (前端处理，仅记录)
	if c.PastedAction != PASTED_ACTION_DIRECT_UPLOAD && c.PastedAction != PASTED_ACTION_WAIT_UPLOAD {
		return fmt.Errorf("invalid pasted_action: %d", c.PastedAction)
	}

	// 验证自动清除预览 (前端处理，仅记录)
	if c.IsAutoClearPreview != AUTO_CLEAR_PREVIEW_YES && c.IsAutoClearPreview != AUTO_CLEAR_PREVIEW_NO {
		return fmt.Errorf("invalid is_auto_clear_preview: %d", c.IsAutoClearPreview)
	}

	return nil
}

// ExampleUsage 示例用法
// 默认配置对应UI中的设置：
// {
//     "pasted_action": 1,           // 等待上传 (Wait for upload) - 前端处理
//     "default_strategy": 1,        // 本地存储 (Local Storage) - 后台处理
//     "default_permission": 0,      // 公开 (Public) - 后台处理
//     "is_auto_clear_preview": 0    // 否 (No) - 前端处理
// }
//
// 配置分类：
// 后台处理 (Backend Processing):
// 1. 默认图片权限 (default_permission: 0) - 影响图片存储和访问权限
// 2. 默认上传策略 (default_strategy: 1) - 影响图片存储位置
//
// 前端处理 (Frontend Processing):
// 1. 粘贴图片后的动作 (pasted_action: 1) - 前端UI行为控制
// 2. 自动清除预览图片 (is_auto_clear_preview: 0) - 前端UI行为控制
//
// 使用示例:
// user := &User{}
// config, err := user.GetUserConfig()
// if err != nil {
//     // 使用默认配置
//     config = GetDefaultUserConfig()
// }
//
// // 获取后台处理的配置
// backendConfig := config.GetBackendConfig()
// // backendConfig = {"default_strategy": 1, "default_permission": 0}
//
// // 获取前端处理的配置
// frontendConfig := config.GetFrontendConfig()
// // frontendConfig = {"pasted_action": 1, "is_auto_clear_preview": 0}
//
// // 验证后台配置
// if err := config.ValidateBackendConfig(); err != nil {
//     // 处理后台配置错误
// }
//
// // 检查后台设置
// if config.DefaultPermission == PERMISSION_PUBLIC {
//     // 用户选择了"公开"权限 - 后台需要处理
// }
//
// if config.DefaultStrategy == DEFAULT_STRATEGY_LOCAL {
//     // 用户选择了"本地存储" - 后台需要处理
// }

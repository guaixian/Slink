package model

import (
	"os"
	"strings"

	"gorm.io/gorm"
)

// AI 配置项的 config_key 常量。
// 这些键存于 configs 表，可在后台「AI 设置」中编辑；若设置了对应的 SLINK_AI_* 环境变量，
// 则环境变量优先（适合 Docker/K8s 注入密钥，避免明文落库）。
const (
	// 图像处理外置扩展（去水印/超分/打标）
	ConfigKeyAIImageEndpoint = "ai_image_endpoint"
	ConfigKeyAIImageToken    = "ai_image_token"
	ConfigKeyAIImageTimeout  = "ai_image_timeout"

	// LLM 平台（OpenAI 兼容），当前用于配置与连通性测试，能力预留
	ConfigKeyAILLMBaseURL = "ai_llm_base_url"
	ConfigKeyAILLMAPIKey  = "ai_llm_api_key"
	ConfigKeyAILLMModel   = "ai_llm_model"

	// 向量 Embedding 平台（OpenAI 兼容），当前用于配置与连通性测试，能力预留
	ConfigKeyAIEmbeddingBaseURL    = "ai_embedding_base_url"
	ConfigKeyAIEmbeddingAPIKey     = "ai_embedding_api_key"
	ConfigKeyAIEmbeddingModel      = "ai_embedding_model"
	ConfigKeyAIEmbeddingDimensions = "ai_embedding_dimensions"
)

// aiConfigDefaults 是需要幂等补齐到 configs 表的 AI 配置项（值留空，由管理员填写）。
var aiConfigDefaults = []Config{
	{ConfigKey: ConfigKeyAIImageEndpoint, Value: "", Description: "AI图像扩展服务地址(去水印/超分/打标)"},
	{ConfigKey: ConfigKeyAIImageToken, Value: "", Description: "AI图像扩展服务令牌"},
	{ConfigKey: ConfigKeyAIImageTimeout, Value: "120", Description: "AI图像扩展请求超时(秒)"},
	{ConfigKey: ConfigKeyAILLMBaseURL, Value: "", Description: "LLM平台BaseURL(OpenAI兼容)"},
	{ConfigKey: ConfigKeyAILLMAPIKey, Value: "", Description: "LLM平台APIKey"},
	{ConfigKey: ConfigKeyAILLMModel, Value: "", Description: "LLM默认模型"},
	{ConfigKey: ConfigKeyAIEmbeddingBaseURL, Value: "", Description: "Embedding平台BaseURL(OpenAI兼容)"},
	{ConfigKey: ConfigKeyAIEmbeddingAPIKey, Value: "", Description: "Embedding平台APIKey"},
	{ConfigKey: ConfigKeyAIEmbeddingModel, Value: "", Description: "Embedding默认模型"},
	{ConfigKey: ConfigKeyAIEmbeddingDimensions, Value: "", Description: "Embedding向量维度"},
}

// aiEnvOverride 将 config_key 映射到优先级更高的环境变量名。
var aiEnvOverride = map[string]string{
	ConfigKeyAIImageEndpoint:       "SLINK_AI_ENDPOINT",
	ConfigKeyAIImageToken:          "SLINK_AI_TOKEN",
	ConfigKeyAIImageTimeout:        "SLINK_AI_TIMEOUT",
	ConfigKeyAILLMBaseURL:          "SLINK_AI_LLM_BASE_URL",
	ConfigKeyAILLMAPIKey:           "SLINK_AI_LLM_API_KEY",
	ConfigKeyAILLMModel:            "SLINK_AI_LLM_MODEL",
	ConfigKeyAIEmbeddingBaseURL:    "SLINK_AI_EMBEDDING_BASE_URL",
	ConfigKeyAIEmbeddingAPIKey:     "SLINK_AI_EMBEDDING_API_KEY",
	ConfigKeyAIEmbeddingModel:      "SLINK_AI_EMBEDDING_MODEL",
	ConfigKeyAIEmbeddingDimensions: "SLINK_AI_EMBEDDING_DIMENSIONS",
}

// AISecretKeys 标记需要脱敏展示的密钥类配置键。
var AISecretKeys = map[string]bool{
	ConfigKeyAIImageToken:      true,
	ConfigKeyAILLMAPIKey:       true,
	ConfigKeyAIEmbeddingAPIKey: true,
}

// AIFieldValue 单个配置项的解析结果
type AIFieldValue struct {
	Value    string // 生效值（env 优先，否则 DB）
	FromEnv  bool   // 是否来自环境变量（来自 env 时后台不可编辑）
	HasValue bool   // 是否已配置（非空）
	IsSecret bool   // 是否为密钥类
}

// AISettings 所有 AI 配置项的解析结果，key 为 config_key
type AISettings map[string]AIFieldValue

// EnsureAIConfigDefaults 幂等补齐 AI 配置项（已有安装也能获得新键）。
func EnsureAIConfigDefaults(db *gorm.DB) error {
	for _, def := range aiConfigDefaults {
		var count int64
		if err := db.Model(&Config{}).Where("config_key = ?", def.ConfigKey).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			c := def // 拷贝，避免取循环变量地址
			if err := db.Create(&c).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadAISettings 读取并解析全部 AI 配置项（env 覆盖 DB）。
func LoadAISettings(db *gorm.DB) (AISettings, error) {
	out := make(AISettings, len(aiConfigDefaults))
	for _, def := range aiConfigDefaults {
		key := def.ConfigKey
		dbVal, err := GetConfigValue(db, key)
		if err == gorm.ErrRecordNotFound {
			dbVal = def.Value
		} else if err != nil {
			return nil, err
		}

		fv := AIFieldValue{IsSecret: AISecretKeys[key]}
		if envName, ok := aiEnvOverride[key]; ok {
			if envVal := strings.TrimSpace(os.Getenv(envName)); envVal != "" {
				fv.Value = envVal
				fv.FromEnv = true
			}
		}
		if !fv.FromEnv {
			fv.Value = dbVal
		}
		fv.HasValue = strings.TrimSpace(fv.Value) != ""
		out[key] = fv
	}
	return out, nil
}

// Get 返回某键的生效值
func (s AISettings) Get(key string) string {
	if v, ok := s[key]; ok {
		return v.Value
	}
	return ""
}

// SaveAISettings 将允许编辑的 AI 配置写回 DB。
// 由环境变量锁定（FromEnv）的键会被跳过，避免后台写入被环境变量覆盖造成困惑。
func SaveAISettings(db *gorm.DB, values map[string]string) error {
	current, err := LoadAISettings(db)
	if err != nil {
		return err
	}
	for key, val := range values {
		if _, ok := aiEnvOverride[key]; !ok {
			continue // 非 AI 配置键，忽略
		}
		if fv, ok := current[key]; ok && fv.FromEnv {
			continue // 环境变量锁定，跳过
		}
		if err := SetConfigValue(db, key, val); err != nil {
			return err
		}
	}
	return nil
}

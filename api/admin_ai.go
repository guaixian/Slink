package api

import (
	"Slink/aiext"
	"Slink/model"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// maskSecret 对密钥做脱敏展示：保留末 4 位，其余以 • 代替；过短则全部遮蔽。
func maskSecret(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if len(v) <= 4 {
		return strings.Repeat("•", len(v))
	}
	return strings.Repeat("•", 8) + v[len(v)-4:]
}

// secretMaskSentinel 前端回传该哨兵值表示「未修改密钥」，后端据此保留原值。
const secretMaskSentinel = "__UNCHANGED__"

// aiSettingField 单字段的对外表示
type aiSettingField struct {
	Key      string `json:"key"`
	Value    string `json:"value"`     // 非密钥：明文；密钥：脱敏值
	HasValue bool   `json:"has_value"` // 是否已配置
	FromEnv  bool   `json:"from_env"`  // 是否由环境变量锁定（只读）
	IsSecret bool   `json:"is_secret"`
}

func toField(key string, fv model.AIFieldValue) aiSettingField {
	value := fv.Value
	if fv.IsSecret {
		value = maskSecret(fv.Value)
	}
	return aiSettingField{
		Key:      key,
		Value:    value,
		HasValue: fv.HasValue,
		FromEnv:  fv.FromEnv,
		IsSecret: fv.IsSecret,
	}
}

// GetAISettings 返回 AI 设置（密钥脱敏、标注 env 锁定）
func GetAISettings(c *gin.Context) {
	s, err := model.LoadAISettings(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "读取AI设置失败", "error": err.Error()})
		return
	}

	data := gin.H{
		"image": gin.H{
			"endpoint": toField(model.ConfigKeyAIImageEndpoint, s[model.ConfigKeyAIImageEndpoint]),
			"token":    toField(model.ConfigKeyAIImageToken, s[model.ConfigKeyAIImageToken]),
			"timeout":  toField(model.ConfigKeyAIImageTimeout, s[model.ConfigKeyAIImageTimeout]),
		},
		"llm": gin.H{
			"base_url": toField(model.ConfigKeyAILLMBaseURL, s[model.ConfigKeyAILLMBaseURL]),
			"api_key":  toField(model.ConfigKeyAILLMAPIKey, s[model.ConfigKeyAILLMAPIKey]),
			"model":    toField(model.ConfigKeyAILLMModel, s[model.ConfigKeyAILLMModel]),
		},
		"embedding": gin.H{
			"base_url":   toField(model.ConfigKeyAIEmbeddingBaseURL, s[model.ConfigKeyAIEmbeddingBaseURL]),
			"api_key":    toField(model.ConfigKeyAIEmbeddingAPIKey, s[model.ConfigKeyAIEmbeddingAPIKey]),
			"model":      toField(model.ConfigKeyAIEmbeddingModel, s[model.ConfigKeyAIEmbeddingModel]),
			"dimensions": toField(model.ConfigKeyAIEmbeddingDimensions, s[model.ConfigKeyAIEmbeddingDimensions]),
		},
		"note": "密钥已脱敏显示；留空或不修改将保留原值。设置了对应 SLINK_AI_* 环境变量的字段由环境变量锁定，后台不可改。",
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "获取成功", "data": data})
}

// aiSettingsUpdateReq 接收的可编辑字段（与前端表单对应）
type aiSettingsUpdateReq map[string]string

// UpdateAISettings 保存 AI 设置。
// 密钥字段若为空或等于脱敏哨兵，则保留原值（不会因脱敏展示而清空）。
func UpdateAISettings(c *gin.Context) {
	var req aiSettingsUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "请求参数错误", "error": err.Error()})
		return
	}

	current, err := model.LoadAISettings(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "读取AI设置失败", "error": err.Error()})
		return
	}

	values := make(map[string]string, len(req))
	for key, val := range req {
		fv, known := current[key]
		if !known {
			continue // 忽略未知键
		}
		if fv.IsSecret {
			// 密钥：空串或哨兵 => 保留原值；前端展示的是脱敏值，避免误回写
			trimmed := strings.TrimSpace(val)
			if trimmed == "" || trimmed == secretMaskSentinel || isMaskedValue(trimmed) {
				continue
			}
		}
		if key == model.ConfigKeyAIImageTimeout && strings.TrimSpace(val) != "" {
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err != nil || n <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "超时须为正整数(秒)"})
				return
			}
		}
		values[key] = val
	}

	if err := model.SaveAISettings(model.DB, values); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "保存AI设置失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "保存成功"})
}

// isMaskedValue 判断字符串是否为脱敏展示值（含 • 字符），避免把脱敏串当真实密钥回写。
func isMaskedValue(v string) bool {
	return strings.Contains(v, "•")
}

// TestAISettings 测试 AI 连接。
// body: {"target":"image|llm|embedding"}，使用当前已保存（含 env 覆盖）的配置进行测试。
func TestAISettings(c *gin.Context) {
	var body struct {
		Target string `json:"target" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "请求参数错误"})
		return
	}

	s, err := model.LoadAISettings(model.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "读取AI设置失败"})
		return
	}

	var result aiext.TestResult
	switch body.Target {
	case "image":
		result = aiext.PingImageExtension(
			s.Get(model.ConfigKeyAIImageEndpoint),
			s.Get(model.ConfigKeyAIImageToken),
		)
	case "llm":
		result = aiext.TestOpenAICompatible(
			s.Get(model.ConfigKeyAILLMBaseURL),
			s.Get(model.ConfigKeyAILLMAPIKey),
		)
	case "embedding":
		result = aiext.TestOpenAICompatible(
			s.Get(model.ConfigKeyAIEmbeddingBaseURL),
			s.Get(model.ConfigKeyAIEmbeddingAPIKey),
		)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "未知的测试目标"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "测试完成", "data": result})
}

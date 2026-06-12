package api

import (
	"Slink/aiext"
	"Slink/applog"
	"Slink/middleware"
	"Slink/model"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getAIClient 依据当前 AI 设置（DB 配置，env 覆盖）构造图像扩展客户端。
// 每次请求解析一次，使后台修改即时生效；读 DB 成本极低。
func getAIClient() *aiext.Client {
	s, err := model.LoadAISettings(model.DB)
	if err != nil {
		return aiext.NewClient("", "", 0) // 解析失败按未配置处理
	}
	timeout := 0
	if v := s.Get(model.ConfigKeyAIImageTimeout); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			timeout = n
		}
	}
	return aiext.NewClient(
		s.Get(model.ConfigKeyAIImageEndpoint),
		s.Get(model.ConfigKeyAIImageToken),
		timeout,
	)
}

// AIStatus 返回外置 AI 扩展的启用状态，便于前端按需展示入口
func AIStatus(c *gin.Context) {
	client := getAIClient()
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "获取成功",
		"data": gin.H{
			"configured":   client.Configured(),
			"endpoint":     client.Endpoint(),
			"capabilities": []string{aiext.CapabilityDewatermark, aiext.CapabilityUpscale, aiext.CapabilityTag},
			"note":         "AI 能力由外置服务提供；未配置时相关接口返回 501。配置见 SLINK_AI_ENDPOINT，文档见 docs/ai-extension.md",
		},
	})
}

// readAIInput 读取上传文件并收集通用参数，供各 AI 能力复用
func readAIInput(c *gin.Context) (data []byte, mime string, params map[string]any, ok bool) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未选择图片"})
		return nil, "", nil, false
	}

	policy, pok := readGlobalUploadPolicy(c)
	if !pok {
		return nil, "", nil, false
	}
	if file.Size > int64(policy.MaximumFileSize*1024) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小超过限制"})
		return nil, "", nil, false
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return nil, "", nil, false
	}
	defer src.Close()

	data, err = io.ReadAll(io.LimitReader(src, int64(policy.MaximumFileSize*1024)+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return nil, "", nil, false
	}

	mime = file.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	// 收集可选参数转发给外部服务
	params = map[string]any{}
	if v := c.PostForm("scale"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			params["scale"] = f
		}
	}
	if v := c.PostForm("prompt"); v != "" {
		params["prompt"] = v
	}
	if v := c.PostForm("model"); v != "" {
		params["model"] = v
	}
	// regions 以原始 JSON 字符串传入（如水印框 [[x,y,w,h],...]），透传给外部服务
	if v := c.PostForm("regions"); v != "" {
		var regions any
		if err := json.Unmarshal([]byte(v), &regions); err == nil {
			params["regions"] = regions
		}
	}
	if len(params) == 0 {
		params = nil
	}
	return data, mime, params, true
}

// handleAIImageCapability 处理返回图像的 AI 能力（去水印 / 超分）
func handleAIImageCapability(c *gin.Context, capability string) {
	rid := middleware.RequestID(c)
	userID, _ := c.Get("userID")

	client := getAIClient()
	if !client.Configured() {
		respondAINotConfigured(c, capability)
		return
	}

	data, mime, params, ok := readAIInput(c)
	if !ok {
		return
	}

	result, err := client.Process(c.Request.Context(), capability, data, mime, params)
	if err != nil {
		writeAIError(c, rid, userID, capability, err)
		return
	}
	if len(result.Image) == 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI 扩展未返回处理后的图片"})
		return
	}

	outMime := result.MimeType
	if outMime == "" {
		outMime = "image/png"
	}

	applog.Logger.Info("ai capability ok", "request_id", rid, "user_id", userID,
		"capability", capability, "out_bytes", len(result.Image))

	if c.PostForm("response") == "json" {
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "处理成功",
			"data": gin.H{
				"mimetype":   outMime,
				"size_bytes": len(result.Image),
				"base64":     base64.StdEncoding.EncodeToString(result.Image),
				"data_uri":   "data:" + outMime + ";base64," + base64.StdEncoding.EncodeToString(result.Image),
				"meta":       result.Meta,
			},
		})
		return
	}

	c.Header("X-Image-Bytes", strconv.Itoa(len(result.Image)))
	c.Data(http.StatusOK, outMime, result.Image)
}

// AIDewatermark 智能去水印（外置扩展）
func AIDewatermark(c *gin.Context) {
	handleAIImageCapability(c, aiext.CapabilityDewatermark)
}

// AIUpscale AI 超分辨率 / 高清放大（外置扩展）
func AIUpscale(c *gin.Context) {
	handleAIImageCapability(c, aiext.CapabilityUpscale)
}

// AITag 智能识别 / 自动打标（外置扩展）
func AITag(c *gin.Context) {
	rid := middleware.RequestID(c)
	userID, _ := c.Get("userID")

	client := getAIClient()
	if !client.Configured() {
		respondAINotConfigured(c, aiext.CapabilityTag)
		return
	}

	data, mime, params, ok := readAIInput(c)
	if !ok {
		return
	}

	result, err := client.Process(c.Request.Context(), aiext.CapabilityTag, data, mime, params)
	if err != nil {
		writeAIError(c, rid, userID, aiext.CapabilityTag, err)
		return
	}

	applog.Logger.Info("ai tag ok", "request_id", rid, "user_id", userID, "tag_count", len(result.Tags))
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "识别成功",
		"data": gin.H{
			"tags": result.Tags,
			"meta": result.Meta,
		},
	})
}

// respondAINotConfigured 统一返回未配置外置扩展的提示
func respondAINotConfigured(c *gin.Context, capability string) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"status":              false,
		"message":             "该功能为外置扩展，依赖外部 AI 服务，当前未配置",
		"capability":          capability,
		"hint":                "设置环境变量 SLINK_AI_ENDPOINT 指向外部服务后启用；对接说明见 docs/ai-extension.md",
		"output_formats_note": "若仅需本地压缩/格式转换/高质量缩放，请改用 /api/image/process",
	})
}

// writeAIError 区分「未配置」与「调用失败」，给出合适的状态码
func writeAIError(c *gin.Context, rid string, userID any, capability string, err error) {
	if errors.Is(err, aiext.ErrNotConfigured) {
		respondAINotConfigured(c, capability)
		return
	}
	applog.Logger.Warn("ai capability failed", "request_id", rid, "user_id", userID,
		"capability", capability, "error", err)
	c.JSON(http.StatusBadGateway, gin.H{
		"status":  false,
		"message": "AI 扩展处理失败: " + err.Error(),
	})
}

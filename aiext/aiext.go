// Package aiext 定义 Slink 的「AI 图片处理外置扩展」客户端。
//
// 智能去水印、AI 超分辨率、智能识别打标等能力依赖真实的机器学习模型，
// 无法在纯 Go 进程内离线完成。Slink 通过一个 provider 无关的 HTTP 契约，
// 将这些能力委托给一个可自行部署/替换的外部服务（如基于 lama-cleaner、
// Real-ESRGAN、自建推理服务或第三方 SaaS 封装而成）。
//
// 启用方式（环境变量）：
//   - SLINK_AI_ENDPOINT  外部服务基础地址，例如 http://127.0.0.1:9000 （必填，未设置即视为未启用）
//   - SLINK_AI_TOKEN     调用令牌，作为 Authorization: Bearer 头发送（可选）
//   - SLINK_AI_TIMEOUT   单次请求超时秒数（可选，默认 120）
//
// 调用契约：
//
//	POST  {endpoint}/v1/{capability}
//	Header: Content-Type: application/json[, Authorization: Bearer <token>]
//	Body  : {"image":"<base64>","mime_type":"image/png","params":{...}}
//
//	200 OK（图像类能力 dewatermark/upscale）:
//	      {"image":"<base64>","mime_type":"image/png","meta":{...}}
//	200 OK（识别类能力 tag）:
//	      {"tags":[{"name":"cat","score":0.97}],"meta":{...}}
//	非 2xx: {"error":"<原因>"}
//
// 详见 docs/ai-extension.md。
package aiext

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// 能力常量
const (
	CapabilityDewatermark = "dewatermark" // 智能去水印
	CapabilityUpscale     = "upscale"     // AI 超分辨率/高清放大
	CapabilityTag         = "tag"         // 智能识别/自动打标
)

const defaultTimeoutSeconds = 120

// ErrNotConfigured 表示尚未配置外置 AI 扩展
var ErrNotConfigured = fmt.Errorf("AI 外置扩展未配置：请设置环境变量 SLINK_AI_ENDPOINT 指向外部服务")

// Client 外置 AI 服务客户端
type Client struct {
	endpoint string
	token    string
	http     *http.Client
}

// Tag 识别类能力返回的单个标签
type Tag struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// Response 外部服务的统一响应
type Response struct {
	// 图像类能力：处理后图片
	Image    []byte
	MimeType string
	// 识别类能力：标签
	Tags []Tag
	// 透传的附加信息
	Meta map[string]any
}

// wireResponse 与外部服务 JSON 对应的中间结构
type wireResponse struct {
	Image    string         `json:"image"`
	MimeType string         `json:"mime_type"`
	Tags     []Tag          `json:"tags"`
	Meta     map[string]any `json:"meta"`
	Error    string         `json:"error"`
}

// LoadFromEnv 依据环境变量构造客户端。
// 始终返回非 nil 的 *Client；是否真正可用由 Configured() 判断。
func LoadFromEnv() *Client {
	endpoint := strings.TrimSpace(os.Getenv("SLINK_AI_ENDPOINT"))
	token := strings.TrimSpace(os.Getenv("SLINK_AI_TOKEN"))
	timeout := defaultTimeoutSeconds
	if v := strings.TrimSpace(os.Getenv("SLINK_AI_TIMEOUT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = n
		}
	}
	return NewClient(endpoint, token, timeout)
}

// NewClient 用显式参数构造客户端（供从 DB 配置解析后调用）。
// endpoint 为空时 Configured() 返回 false；timeout<=0 时使用默认值。
func NewClient(endpoint, token string, timeoutSeconds int) *Client {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultTimeoutSeconds
	}
	return &Client{
		endpoint: endpoint,
		token:    strings.TrimSpace(token),
		http:     &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
	}
}

// Configured 是否已配置外部服务地址
func (c *Client) Configured() bool {
	return c != nil && c.endpoint != ""
}

// Endpoint 返回已配置的基础地址（用于状态展示）
func (c *Client) Endpoint() string {
	if c == nil {
		return ""
	}
	return c.endpoint
}

// Process 调用外部服务执行指定能力。
// image 为原始图片字节，mimeType 为其 MIME 类型，params 为能力相关参数（可为 nil）。
func (c *Client) Process(ctx context.Context, capability string, image []byte, mimeType string, params map[string]any) (*Response, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}

	reqBody := map[string]any{
		"image":     base64.StdEncoding.EncodeToString(image),
		"mime_type": mimeType,
	}
	if params != nil {
		reqBody["params"] = params
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.endpoint + "/v1/" + capability
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 AI 扩展失败: %w", err)
	}
	defer resp.Body.Close()

	// 限制响应体大小，避免异常的超大响应耗尽内存（64MiB 足够覆盖高清图）
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, fmt.Errorf("读取 AI 扩展响应失败: %w", err)
	}

	var wire wireResponse
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, fmt.Errorf("解析 AI 扩展响应失败(状态码 %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := wire.Error
		if msg == "" {
			msg = fmt.Sprintf("AI 扩展返回状态码 %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s", msg)
	}

	out := &Response{
		MimeType: wire.MimeType,
		Tags:     wire.Tags,
		Meta:     wire.Meta,
	}
	if wire.Image != "" {
		decoded, err := base64.StdEncoding.DecodeString(wire.Image)
		if err != nil {
			return nil, fmt.Errorf("解码 AI 扩展返回图片失败: %w", err)
		}
		out.Image = decoded
	}
	return out, nil
}

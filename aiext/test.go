package aiext

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// testTimeout 连接测试的短超时，避免后台「测试连接」长时间挂起
const testTimeout = 12 * time.Second

// TestResult 连接测试结果
type TestResult struct {
	OK      bool   `json:"ok"`
	Status  int    `json:"status,omitempty"`
	Message string `json:"message"`
}

// PingImageExtension 探测图像扩展服务可达性。
// 契约未定义健康端点，因此只做基础可达性判断：能建立 HTTP 连接并拿到任意响应即视为可达
// （即便返回 404/405，也说明地址正确、服务在线）。
func PingImageExtension(endpoint, token string) TestResult {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return TestResult{OK: false, Message: "未填写服务地址"}
	}
	if !validHTTPURL(endpoint) {
		return TestResult{OK: false, Message: "地址须为 http/https"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return TestResult{OK: false, Message: "构造请求失败: " + err.Error()}
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TestResult{OK: false, Message: "无法连接: " + err.Error()}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	return TestResult{OK: true, Status: resp.StatusCode, Message: fmt.Sprintf("可达（HTTP %d）", resp.StatusCode)}
}

// TestOpenAICompatible 对 OpenAI 兼容平台（LLM/Embedding）做连通性测试：
// GET {baseURL}/models，带 Authorization。2xx 视为成功，401/403 视为「可达但鉴权失败」。
func TestOpenAICompatible(baseURL, apiKey string) TestResult {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return TestResult{OK: false, Message: "未填写 BaseURL"}
	}
	if !validHTTPURL(baseURL) {
		return TestResult{OK: false, Message: "BaseURL 须为 http/https"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	endpoint := baseURL + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return TestResult{OK: false, Message: "构造请求失败: " + err.Error()}
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TestResult{OK: false, Message: "无法连接: " + err.Error()}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return TestResult{OK: true, Status: resp.StatusCode, Message: "连接成功，鉴权通过"}
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return TestResult{OK: false, Status: resp.StatusCode, Message: "服务可达，但 APIKey 鉴权失败"}
	case resp.StatusCode == http.StatusNotFound:
		return TestResult{OK: false, Status: resp.StatusCode, Message: "服务可达，但 /models 不存在，请确认 BaseURL 是否含正确路径（如 /v1）"}
	default:
		return TestResult{OK: false, Status: resp.StatusCode, Message: fmt.Sprintf("服务返回 HTTP %d", resp.StatusCode)}
	}
}

func validHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

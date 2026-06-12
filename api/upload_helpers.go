package api

import (
	"Slink/model"
	"Slink/utils"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// remoteImageDownloadTimeout 远程图片下载的整体超时，避免慢速/挂起的远端拖垮服务
const remoteImageDownloadTimeout = 30 * time.Second

// downloadRemoteImage 安全地下载远程图片。
// 相比直接使用 http.Get，这里：
//  1. 校验 URL 必须为 http/https（基础 SSRF 加固，拒绝 file://、gopher:// 等）；
//  2. 为请求设置整体超时，避免无限期挂起；
//  3. 在 Content-Length 已知时提前拒绝超限文件；
//  4. 使用 io.LimitReader 限制实际读取字节数，防止 Content-Length 缺失(-1)时的内存耗尽。
//
// 返回图片字节、响应的 Content-Type。任何校验失败都返回可直接展示给调用方的错误。
func downloadRemoteImage(rawURL string, maxBytes int64) ([]byte, string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, "", fmt.Errorf("无效的图片URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, "", fmt.Errorf("仅支持 http/https 协议的图片URL")
	}

	client := &http.Client{Timeout: remoteImageDownloadTimeout}
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, "", fmt.Errorf("下载图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", fmt.Errorf("URL不是有效的图片")
	}

	// Content-Length 已知且超限时提前拒绝，省去无谓的下载
	if maxBytes > 0 && resp.ContentLength > maxBytes {
		return nil, "", fmt.Errorf("文件大小超过限制")
	}

	// 即便 Content-Length 缺失或撒谎，也用 LimitReader 兜底，多读 1 字节用于判断是否超限
	limited := io.LimitReader(resp.Body, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", fmt.Errorf("读取图片数据失败: %w", err)
	}
	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, "", fmt.Errorf("文件大小超过限制")
	}

	return data, contentType, nil
}

// resolveUploadStrategy 解析本次上传使用的存储策略：表单/JSON 的 strategy_id 优先，否则用户偏好中的 default_strategy，否则第一个已配置策略
func resolveUploadStrategy(db *gorm.DB, userConfig model.UserConfig, strategyIDRaw string) (*model.Strategies, *model.StrategyConfigData, error) {
	sid := userConfig.DefaultStrategy
	if strategyIDRaw != "" {
		if v, err := strconv.ParseUint(strategyIDRaw, 10, 32); err == nil && v > 0 {
			sid = int(v)
		}
	}
	all, err := model.GetAllStrategies(db)
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, fmt.Errorf("未配置存储策略")
	}
	var st *model.Strategies
	for i := range all {
		if int(all[i].ID) == sid {
			st = &all[i]
			break
		}
	}
	if st == nil {
		st = &all[0]
	}
	scd, err := model.GetStrategyConfigData(st)
	if err != nil {
		return nil, nil, err
	}
	return st, scd, nil
}

func strategyUtilsConfigFromModel(strategy *model.Strategies) (*utils.StrategyConfig, error) {
	sc, err := model.GetStrategyConfig(strategy)
	if err != nil {
		return nil, err
	}
	baseURL := strings.TrimSpace(sc.URL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(sc.Domain)
	}
	return &utils.StrategyConfig{
		URL:      baseURL,
		Root:     sc.Root,
		Queries:  sc.Queries,
		AuthType: sc.AuthType,
	}, nil
}

func strategyUtilsConfigForStoredImage(db *gorm.DB, img *model.Images) (*utils.StrategyConfig, error) {
	st, err := model.GetStrategyForStoredImage(db, img.StrategyID)
	if err != nil {
		return nil, err
	}
	return strategyUtilsConfigFromModel(st)
}

func normalizePathnameForLinks(path string) string {
	pathname := path
	if len(pathname) > 7 && pathname[:7] == "static/" {
		pathname = pathname[7:]
	} else if len(pathname) > 8 && pathname[:8] == `static\` {
		pathname = pathname[8:]
	}
	return pathname
}

func readGlobalUploadPolicy(c *gin.Context) (*model.GroupConfig, bool) {
	pol, err := model.GetGlobalUploadPolicy(model.DB)
	if err != nil {
		c.JSON(500, gin.H{"error": "读取全局上传策略失败"})
		return nil, false
	}
	return pol, true
}

package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ImageInfo 图片信息结构
type ImageInfo struct {
	Path       string
	OriginName string
	Size       float64
	Mimetype   string
	Extension  string
	MD5        string
	SHA1       string
	Width      int64
	Height     int64
}

// SaveImageToStatic 保存图片到static目录，根据用户组配置生成路径
func SaveImageToStatic(file *multipart.FileHeader, userID uint, pathRule, fileRule string) (*ImageInfo, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isValidImageExt(ext) {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 根据用户组配置生成路径和文件名
	pathDir := GeneratePathFromRule(pathRule, userID, file.Filename)
	filename := GenerateFileNameFromRule(fileRule, userID, file.Filename, strings.TrimPrefix(ext, "."))

	// 构建完整路径
	fullPath := filepath.Join("static", pathDir, filename)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}

	// 获取文件信息
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	// 计算MD5和SHA1
	md5Hash, sha1Hash, err := calculateHashes(fullPath)
	if err != nil {
		return nil, err
	}

	// 获取图片尺寸
	width, height, err := getImageDimensions(fullPath)
	if err != nil {
		return nil, err
	}

	// 获取MIME类型
	mimetype := getMimeType(ext)

	// 计算文件大小（MB）
	sizeMB := float64(fileInfo.Size()) / (1024 * 1024)

	return &ImageInfo{
		Path:       fullPath,
		OriginName: file.Filename,
		Size:       sizeMB,
		Mimetype:   mimetype,
		Extension:  strings.TrimPrefix(ext, "."),
		MD5:        md5Hash,
		SHA1:       sha1Hash,
		Width:      width,
		Height:     height,
	}, nil
}

// SaveImageFromBytes 从字节数据保存图片
func SaveImageFromBytes(data []byte, filename string, userID uint, pathRule, fileRule string) (*ImageInfo, error) {
	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(filename))
	if !isValidImageExt(ext) {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 根据用户组配置生成路径和文件名
	pathDir := GeneratePathFromRule(pathRule, userID, filename)
	newFilename := GenerateFileNameFromRule(fileRule, userID, filename, strings.TrimPrefix(ext, "."))

	// 构建完整路径
	fullPath := filepath.Join("static", pathDir, newFilename)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 写入文件
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, err
	}

	// 获取文件信息
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	// 计算MD5和SHA1
	md5Hash := md5.Sum(data)
	sha1Hash := sha1.Sum(data)
	md5Str := hex.EncodeToString(md5Hash[:])
	sha1Str := hex.EncodeToString(sha1Hash[:])

	// 获取图片尺寸
	width, height, err := getImageDimensions(fullPath)
	if err != nil {
		return nil, err
	}

	// 获取MIME类型
	mimetype := getMimeType(ext)

	// 计算文件大小（MB）
	sizeMB := float64(fileInfo.Size()) / (1024 * 1024)

	return &ImageInfo{
		Path:       fullPath,
		OriginName: filename,
		Size:       sizeMB,
		Mimetype:   mimetype,
		Extension:  strings.TrimPrefix(ext, "."),
		MD5:        md5Str,
		SHA1:       sha1Str,
		Width:      width,
		Height:     height,
	}, nil
}

// isValidImageExt 检查是否为有效的图片扩展名
func isValidImageExt(ext string) bool {
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// getMimeType 根据扩展名获取MIME类型
func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
	}
	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	return "application/octet-stream"
}

// calculateHashes 计算文件的MD5和SHA1值
func calculateHashes(filePath string) (string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}

	md5Hash := md5.Sum(data)
	sha1Hash := sha1.Sum(data)

	md5Str := hex.EncodeToString(md5Hash[:])
	sha1Str := hex.EncodeToString(sha1Hash[:])

	return md5Str, sha1Str, nil
}

// getImageDimensions 获取图片尺寸
func getImageDimensions(filePath string) (int64, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return int64(img.Width), int64(img.Height), nil
}

// GenerateImageLinks 生成图片链接
func GenerateImageLinks(pathname string, originName string) map[string]string {
	baseURL := "https://img.intplg.com/img"
	url := fmt.Sprintf("%s/%s", baseURL, pathname)

	return map[string]string{
		"url":                url,
		"html":               fmt.Sprintf(`<img src="%s" alt="%s" title="%s" />`, url, originName, originName),
		"bbcode":             fmt.Sprintf("[img]%s[/img]", url),
		"markdown":           fmt.Sprintf("![%s](%s)", originName, url),
		"markdown_with_link": fmt.Sprintf("[![%s](%s)](%s)", originName, url, url),
		"thumbnail_url":      fmt.Sprintf("https://img.intplg.com/thumbnails/%s.png", strings.Split(pathname, "/")[len(strings.Split(pathname, "/"))-1]),
	}
}

// StrategyConfig 策略配置结构
type StrategyConfig struct {
	URL      string            `json:"url"`
	Root     string            `json:"root"`
	Queries  map[string]string `json:"queries"`
	AuthType string            `json:"auth_type"`
}

// GenerateImageLinksWithConfig 根据策略配置生成图片链接
func GenerateImageLinksWithConfig(pathname string, originName string, config *StrategyConfig) map[string]string {
	// 规范化 URL 和路径，避免双斜杠
	baseURL := strings.TrimSuffix(config.URL, "/")
	path := pathname
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 构建完整的URL
	url := baseURL + path

	// 如果有查询参数，添加到URL中
	if config.Queries != nil && len(config.Queries) > 0 {
		queryParams := make([]string, 0)
		for key, value := range config.Queries {
			queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
		}
		if len(queryParams) > 0 {
			url = fmt.Sprintf("%s?%s", url, strings.Join(queryParams, "&"))
		}
	}

	return map[string]string{
		"url":                url,
		"html":               fmt.Sprintf(`<img src="%s" alt="%s" title="%s" />`, url, originName, originName),
		"bbcode":             fmt.Sprintf("[img]%s[/img]", url),
		"markdown":           fmt.Sprintf("![%s](%s)", originName, url),
		"markdown_with_link": fmt.Sprintf("[![%s](%s)](%s)", originName, url, url),
		"thumbnail_url":      fmt.Sprintf("%s/thumbnails/%s.png", baseURL, strings.Split(pathname, "/")[len(strings.Split(pathname, "/"))-1]),
	}
}

// GeneratePathFromRule 根据规则生成路径
func GeneratePathFromRule(rule string, userID uint, originalName string) string {
	now := time.Now()

	// 替换规则中的占位符
	path := rule
	path = strings.ReplaceAll(path, "{Y}", strconv.Itoa(now.Year()))
	path = strings.ReplaceAll(path, "{y}", strconv.Itoa(now.Year())[2:])
	path = strings.ReplaceAll(path, "{m}", fmt.Sprintf("%02d", now.Month()))
	path = strings.ReplaceAll(path, "{d}", fmt.Sprintf("%02d", now.Day()))
	path = strings.ReplaceAll(path, "{timestamp}", strconv.FormatInt(now.Unix(), 10))
	path = strings.ReplaceAll(path, "{uniqid}", generateUniqID())
	path = strings.ReplaceAll(path, "{md5}", generateMD5())
	path = strings.ReplaceAll(path, "{md5-16}", generateMD5()[:16])
	path = strings.ReplaceAll(path, "{str-random-16}", generateRandomString(16))
	path = strings.ReplaceAll(path, "{str-random-10}", generateRandomString(10))
	path = strings.ReplaceAll(path, "{uid}", strconv.FormatUint(uint64(userID), 10))

	return path
}

// GenerateFileNameFromRule 根据规则生成文件名
func GenerateFileNameFromRule(rule string, userID uint, originalName string, extension string) string {
	// 统一使用UUID生成文件名
	uuid := generateUniqID()
	return uuid + "." + extension
}

// generateUniqID 生成唯一ID
func generateUniqID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// generateMD5 生成MD5值
func generateMD5() string {
	data := fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().Unix())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// generateRandomString 生成随机字符串
// 使用 crypto/rand 提供真正的随机性；旧实现以 time.Now().UnixNano()%len 取字符，
// 在同一纳秒内的循环中会得到几乎完全相同的字符，导致 {str-random-*} 路径规则高概率碰撞。
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		// 退化为基于纳秒种子的弱随机，仅作为 crypto/rand 不可用时的兜底
		for i := range buf {
			buf[i] = charset[(time.Now().UnixNano()+int64(i))%int64(len(charset))]
		}
		return string(buf)
	}
	for i := range buf {
		buf[i] = charset[int(buf[i])%len(charset)]
	}
	return string(buf)
}

// GetRealClientIP 获取真实的客户端IP地址
func GetRealClientIP(c *gin.Context) string {
	// 优先从X-Real-IP获取（Nginx等反向代理设置）
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 其次从X-Forwarded-For获取（代理链中的IP）
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For可能包含多个IP，取第一个
		if commaIndex := strings.Index(forwardedFor, ","); commaIndex != -1 {
			return strings.TrimSpace(forwardedFor[:commaIndex])
		}
		return strings.TrimSpace(forwardedFor)
	}

	// 最后使用gin的ClientIP()方法
	return c.ClientIP()
}

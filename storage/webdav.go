package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/studio-b12/gowebdav"
)

// WebDAVStorage WebDAV 存储
type WebDAVStorage struct {
	config *Config
	client *gowebdav.Client
	domain string
}

// NewWebDAVStorage 创建 WebDAV 存储实例
func NewWebDAVStorage(cfg *Config) (*WebDAVStorage, error) {
	// 构建 WebDAV URL
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost/webdav"
	}

	// 确保 endpoint 以 http:// 或 https:// 开头
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		if cfg.UseSSL {
			endpoint = "https://" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	}

	// 创建 WebDAV 客户端
	client := gowebdav.NewClient(endpoint, cfg.AccessKey, cfg.SecretKey)

	// 测试连接
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("WebDAV 连接失败: %w", err)
	}

	domain := cfg.Domain
	if domain == "" {
		domain = endpoint
	}

	return &WebDAVStorage{
		config: cfg,
		client: client,
		domain: domain,
	}, nil
}

// Upload 上传文件
func (s *WebDAVStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 读取文件内容
	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 构建完整路径
	remotePath := filepath.Join(s.config.BasePath, path)
	remotePath = filepath.ToSlash(remotePath) // WebDAV 使用正斜杠

	// 确保远程目录存在
	remoteDir := filepath.Dir(remotePath)
	if err := s.createDirRecursive(remoteDir); err != nil {
		return "", fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 上传文件
	if err := s.client.Write(remotePath, data, 0644); err != nil {
		return "", fmt.Errorf("上传 WebDAV 文件失败: %w", err)
	}

	return remotePath, nil
}

// UploadBytes 上传字节数据
func (s *WebDAVStorage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	remotePath := filepath.Join(s.config.BasePath, path)
	remotePath = filepath.ToSlash(remotePath)

	// 确保远程目录存在
	remoteDir := filepath.Dir(remotePath)
	if err := s.createDirRecursive(remoteDir); err != nil {
		return "", fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 上传文件
	if err := s.client.Write(remotePath, data, 0644); err != nil {
		return "", fmt.Errorf("上传 WebDAV 文件失败: %w", err)
	}

	return remotePath, nil
}

// Delete 删除文件
func (s *WebDAVStorage) Delete(path string) error {
	if err := s.client.Remove(path); err != nil {
		return fmt.Errorf("删除 WebDAV 文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *WebDAVStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, strings.TrimPrefix(path, "/"))
}

// Exists 检查文件是否存在
func (s *WebDAVStorage) Exists(path string) (bool, error) {
	info, err := s.client.Stat(path)
	if err != nil {
		return false, nil
	}
	return info != nil, nil
}

// GetSize 获取文件大小
func (s *WebDAVStorage) GetSize(path string) (int64, error) {
	info, err := s.client.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// createDirRecursive 递归创建目录
func (s *WebDAVStorage) createDirRecursive(dir string) error {
	dir = filepath.ToSlash(dir)
	if dir == "" || dir == "." || dir == "/" {
		return nil
	}

	// 检查目录是否存在
	info, err := s.client.Stat(dir)
	if err == nil && info.IsDir() {
		return nil // 目录已存在
	}

	// 递归创建父目录
	parent := filepath.Dir(dir)
	if parent != "" && parent != "." && parent != "/" {
		if err := s.createDirRecursive(parent); err != nil {
			return err
		}
	}

	// 创建当前目录
	return s.client.Mkdir(dir, 0755)
}

// ReadStream 读取文件流
func (s *WebDAVStorage) ReadStream(path string) (io.ReadCloser, error) {
	return s.client.ReadStream(path)
}

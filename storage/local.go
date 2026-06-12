package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// LocalStorage 本地存储
type LocalStorage struct {
	config   *Config
	basePath string
	baseURL  string
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(config *Config) (*LocalStorage, error) {
	basePath := config.BasePath
	if basePath == "" {
		basePath = "static"
	}
	// 防止误用 URL 根 / 为磁盘路径（在 Unix 上 filepath.Join("/", "2026") 会落在系统根下）
	if basePath == string(filepath.Separator) {
		basePath = "static"
	}
	if c := filepath.Clean(basePath); c == string(filepath.Separator) {
		basePath = "static"
	}

	// 确保目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	baseURL := config.Domain
	if baseURL == "" {
		baseURL = "/static"
	}

	return &LocalStorage{
		config:   config,
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// Upload 上传文件
func (s *LocalStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 构建完整路径
	fullPath := filepath.Join(s.basePath, path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("复制文件失败: %w", err)
	}

	return path, nil
}

// UploadBytes 上传字节数据
func (s *LocalStorage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	fullPath := filepath.Join(s.basePath, path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return path, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(s.basePath, path)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *LocalStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, path)
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetSize 获取文件大小
func (s *LocalStorage) GetSize(path string) (int64, error) {
	fullPath := filepath.Join(s.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

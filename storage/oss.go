package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSStorage 阿里云 OSS 存储
type OSSStorage struct {
	config *Config
	client *oss.Client
	bucket *oss.Bucket
	domain string
}

// NewOSSStorage 创建 OSS 存储实例
func NewOSSStorage(cfg *Config) (*OSSStorage, error) {
	// 创建 OSS 客户端
	client, err := oss.New(cfg.Endpoint, cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("创建 OSS 客户端失败: %w", err)
	}

	// 获取存储桶
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("获取 OSS 存储桶失败: %w", err)
	}

	domain := cfg.Domain
	if domain == "" {
		domain = fmt.Sprintf("https://%s.%s", cfg.Bucket, cfg.Endpoint)
	}

	return &OSSStorage{
		config: cfg,
		client: client,
		bucket: bucket,
		domain: domain,
	}, nil
}

// Upload 上传文件
func (s *OSSStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
	// 打开文件
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
	key := filepath.Join(s.config.BasePath, path)

	// 上传到 OSS
	err = s.bucket.PutObject(key, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("上传到 OSS 失败: %w", err)
	}

	return key, nil
}

// UploadBytes 上传字节数据
func (s *OSSStorage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	key := filepath.Join(s.config.BasePath, path)

	// 上传到 OSS
	err := s.bucket.PutObject(key, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("上传到 OSS 失败: %w", err)
	}

	return key, nil
}

// Delete 删除文件
func (s *OSSStorage) Delete(path string) error {
	err := s.bucket.DeleteObject(path)
	if err != nil {
		return fmt.Errorf("删除 OSS 文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *OSSStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, path)
}

// Exists 检查文件是否存在
func (s *OSSStorage) Exists(path string) (bool, error) {
	exists, err := s.bucket.IsObjectExist(path)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetSize 获取文件大小
func (s *OSSStorage) GetSize(path string) (int64, error) {
	meta, err := s.bucket.GetObjectMeta(path)
	if err != nil {
		return 0, err
	}

	contentLength := meta.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("无法获取文件大小")
	}

	var size int64
	fmt.Sscanf(contentLength, "%d", &size)
	return size, nil
}

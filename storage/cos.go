package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSStorage 腾讯云 COS 存储
type COSStorage struct {
	config *Config
	client *cos.Client
	domain string
}

// NewCOSStorage 创建 COS 存储实例
func NewCOSStorage(cfg *Config) (*COSStorage, error) {
	// 构建 Bucket URL
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("解析 Bucket URL 失败: %w", err)
	}

	// 创建 COS 客户端
	baseURL := &cos.BaseURL{BucketURL: bucketURL}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.AccessKey,
			SecretKey: cfg.SecretKey,
		},
	})

	domain := cfg.Domain
	if domain == "" {
		domain = bucketURL.String()
	}

	return &COSStorage{
		config: cfg,
		client: client,
		domain: domain,
	}, nil
}

// Upload 上传文件
func (s *COSStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
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

	// 上传到 COS
	_, err = s.client.Object.Put(context.Background(), key, bytes.NewReader(data), nil)
	if err != nil {
		return "", fmt.Errorf("上传到 COS 失败: %w", err)
	}

	return key, nil
}

// UploadBytes 上传字节数据
func (s *COSStorage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	key := filepath.Join(s.config.BasePath, path)

	// 上传到 COS
	_, err := s.client.Object.Put(context.Background(), key, bytes.NewReader(data), nil)
	if err != nil {
		return "", fmt.Errorf("上传到 COS 失败: %w", err)
	}

	return key, nil
}

// Delete 删除文件
func (s *COSStorage) Delete(path string) error {
	_, err := s.client.Object.Delete(context.Background(), path)
	if err != nil {
		return fmt.Errorf("删除 COS 文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *COSStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, path)
}

// Exists 检查文件是否存在
func (s *COSStorage) Exists(path string) (bool, error) {
	_, err := s.client.Object.Head(context.Background(), path, nil)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *COSStorage) GetSize(path string) (int64, error) {
	resp, err := s.client.Object.Head(context.Background(), path, nil)
	if err != nil {
		return 0, err
	}
	return resp.ContentLength, nil
}

package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioStorage Minio 存储
type MinioStorage struct {
	config *Config
	client *minio.Client
	bucket string
	domain string
}

// NewMinioStorage 创建 Minio 存储实例
func NewMinioStorage(cfg *Config) (*MinioStorage, error) {
	// 创建 Minio 客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Minio 客户端失败: %w", err)
	}

	// 检查存储桶是否存在
	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("检查存储桶失败: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("存储桶 %s 不存在", cfg.Bucket)
	}

	domain := cfg.Domain
	if domain == "" {
		protocol := "http"
		if cfg.UseSSL {
			protocol = "https"
		}
		domain = fmt.Sprintf("%s://%s/%s", protocol, cfg.Endpoint, cfg.Bucket)
	}

	return &MinioStorage{
		config: cfg,
		client: client,
		bucket: cfg.Bucket,
		domain: domain,
	}, nil
}

// Upload 上传文件
func (s *MinioStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
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

	// 上传到 Minio
	_, err = s.client.PutObject(context.Background(), s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("上传到 Minio 失败: %w", err)
	}

	return key, nil
}

// UploadBytes 上传字节数据
func (s *MinioStorage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	key := filepath.Join(s.config.BasePath, path)

	// 上传到 Minio
	_, err := s.client.PutObject(context.Background(), s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("上传到 Minio 失败: %w", err)
	}

	return key, nil
}

// Delete 删除文件
func (s *MinioStorage) Delete(path string) error {
	err := s.client.RemoveObject(context.Background(), s.bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除 Minio 文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *MinioStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, path)
}

// Exists 检查文件是否存在
func (s *MinioStorage) Exists(path string) (bool, error) {
	_, err := s.client.StatObject(context.Background(), s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *MinioStorage) GetSize(path string) (int64, error) {
	info, err := s.client.StatObject(context.Background(), s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/upyun/go-sdk/v3/upyun"
)

// UpyunStorage 又拍云存储
type UpyunStorage struct {
	config *Config
	client *upyun.UpYun
	domain string
}

// NewUpyunStorage 创建又拍云存储实例
func NewUpyunStorage(cfg *Config) (*UpyunStorage, error) {
	// 创建又拍云客户端
	client := upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   cfg.Bucket,
		Operator: cfg.AccessKey,
		Password: cfg.SecretKey,
	})

	domain := cfg.Domain
	if domain == "" {
		domain = fmt.Sprintf("http://%s.b0.upaiyun.com", cfg.Bucket)
	}

	return &UpyunStorage{
		config: cfg,
		client: client,
		domain: domain,
	}, nil
}

// Upload 上传文件
func (s *UpyunStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
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

	// 上传到又拍云
	err = s.client.Put(&upyun.PutObjectConfig{
		Path:   key,
		Reader: bytes.NewReader(data),
	})
	if err != nil {
		return "", fmt.Errorf("上传到又拍云失败: %w", err)
	}

	return key, nil
}

// UploadBytes 上传字节数据
func (s *UpyunStorage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	key := filepath.Join(s.config.BasePath, path)

	// 上传到又拍云
	err := s.client.Put(&upyun.PutObjectConfig{
		Path:   key,
		Reader: bytes.NewReader(data),
	})
	if err != nil {
		return "", fmt.Errorf("上传到又拍云失败: %w", err)
	}

	return key, nil
}

// Delete 删除文件
func (s *UpyunStorage) Delete(path string) error {
	err := s.client.Delete(&upyun.DeleteObjectConfig{
		Path: path,
	})
	if err != nil {
		return fmt.Errorf("删除又拍云文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *UpyunStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, path)
}

// Exists 检查文件是否存在
func (s *UpyunStorage) Exists(path string) (bool, error) {
	_, err := s.client.GetInfo(path)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *UpyunStorage) GetSize(path string) (int64, error) {
	info, err := s.client.GetInfo(path)
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

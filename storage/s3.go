package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage AWS S3 存储
type S3Storage struct {
	config *Config
	client *s3.Client
	bucket string
	domain string
}

// NewS3Storage 创建 S3 存储实例
func NewS3Storage(cfg *Config) (*S3Storage, error) {
	// 创建 AWS 配置
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("加载 AWS 配置失败: %w", err)
	}

	// 创建 S3 客户端
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = true // 使用路径风格访问
	})

	domain := cfg.Domain
	if domain == "" {
		if cfg.Endpoint != "" {
			domain = fmt.Sprintf("%s/%s", cfg.Endpoint, cfg.Bucket)
		} else {
			domain = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Bucket, cfg.Region)
		}
	}

	return &S3Storage{
		config: cfg,
		client: client,
		bucket: cfg.Bucket,
		domain: domain,
	}, nil
}

// Upload 上传文件
func (s *S3Storage) Upload(file *multipart.FileHeader, path string) (string, error) {
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

	// 上传到 S3
	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("上传到 S3 失败: %w", err)
	}

	return key, nil
}

// UploadBytes 上传字节数据
func (s *S3Storage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	key := filepath.Join(s.config.BasePath, path)

	// 上传到 S3
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", fmt.Errorf("上传到 S3 失败: %w", err)
	}

	return key, nil
}

// Delete 删除文件
func (s *S3Storage) Delete(path string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("删除 S3 文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *S3Storage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, path)
}

// Exists 检查文件是否存在
func (s *S3Storage) Exists(path string) (bool, error) {
	_, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *S3Storage) GetSize(path string) (int64, error) {
	output, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return 0, err
	}
	return *output.ContentLength, nil
}

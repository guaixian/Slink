package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// QiniuStorage 七牛云存储
type QiniuStorage struct {
	config    *Config
	mac       *qbox.Mac
	bucket    string
	domain    string
	putPolicy storage.PutPolicy
	uploader  *storage.FormUploader
	bucketMgr *storage.BucketManager
}

// NewQiniuStorage 创建七牛云存储实例
func NewQiniuStorage(cfg *Config) (*QiniuStorage, error) {
	mac := qbox.NewMac(cfg.AccessKey, cfg.SecretKey)

	// 配置上传策略
	putPolicy := storage.PutPolicy{
		Scope: cfg.Bucket,
	}

	// 创建上传器
	upCfg := storage.Config{
		Zone:          &storage.ZoneHuadong, // 根据实际情况选择区域
		UseCdnDomains: false,
		UseHTTPS:      cfg.UseSSL,
	}
	uploader := storage.NewFormUploader(&upCfg)

	// 创建存储桶管理器
	bucketMgr := storage.NewBucketManager(mac, &upCfg)

	domain := cfg.Domain
	if domain == "" {
		domain = fmt.Sprintf("http://%s.qiniudn.com", cfg.Bucket)
	}

	return &QiniuStorage{
		config:    cfg,
		mac:       mac,
		bucket:    cfg.Bucket,
		domain:    domain,
		putPolicy: putPolicy,
		uploader:  uploader,
		bucketMgr: bucketMgr,
	}, nil
}

// Upload 上传文件
func (s *QiniuStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
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

	// 生成上传凭证
	upToken := s.putPolicy.UploadToken(s.mac)

	// 上传到七牛云
	ret := storage.PutRet{}
	err = s.uploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(data), int64(len(data)), nil)
	if err != nil {
		return "", fmt.Errorf("上传到七牛云失败: %w", err)
	}

	return key, nil
}

// UploadBytes 上传字节数据
func (s *QiniuStorage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	key := filepath.Join(s.config.BasePath, path)

	// 生成上传凭证
	upToken := s.putPolicy.UploadToken(s.mac)

	// 上传到七牛云
	ret := storage.PutRet{}
	err := s.uploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(data), int64(len(data)), nil)
	if err != nil {
		return "", fmt.Errorf("上传到七牛云失败: %w", err)
	}

	return key, nil
}

// Delete 删除文件
func (s *QiniuStorage) Delete(path string) error {
	err := s.bucketMgr.Delete(s.bucket, path)
	if err != nil {
		return fmt.Errorf("删除七牛云文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *QiniuStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, path)
}

// Exists 检查文件是否存在
func (s *QiniuStorage) Exists(path string) (bool, error) {
	_, err := s.bucketMgr.Stat(s.bucket, path)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *QiniuStorage) GetSize(path string) (int64, error) {
	fileInfo, err := s.bucketMgr.Stat(s.bucket, path)
	if err != nil {
		return 0, err
	}
	return fileInfo.Fsize, nil
}

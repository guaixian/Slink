package storage

import (
	"mime/multipart"
)

// Storage 存储接口
type Storage interface {
	// Upload 上传文件
	Upload(file *multipart.FileHeader, path string) (string, error)

	// UploadBytes 上传字节数据
	UploadBytes(data []byte, filename string, path string) (string, error)

	// Delete 删除文件
	Delete(path string) error

	// GetURL 获取文件访问URL
	GetURL(path string) string

	// Exists 检查文件是否存在
	Exists(path string) (bool, error)

	// GetSize 获取文件大小
	GetSize(path string) (int64, error)
}

// Config 存储配置
type Config struct {
	Type      string            `json:"type"`       // local, s3, oss, cos, qiniu, upyun, sftp, ftp, webdav, minio
	Endpoint  string            `json:"endpoint"`   // 端点地址
	AccessKey string            `json:"access_key"` // 访问密钥
	SecretKey string            `json:"secret_key"` // 密钥
	Bucket    string            `json:"bucket"`     // 存储桶名称
	Region    string            `json:"region"`     // 区域
	Domain    string            `json:"domain"`     // 自定义域名
	BasePath  string            `json:"base_path"`  // 基础路径
	UseSSL    bool              `json:"use_ssl"`    // 是否使用SSL
	Extra     map[string]string `json:"extra"`      // 额外配置
}

// NewStorage 创建存储实例
func NewStorage(config *Config) (Storage, error) {
	switch config.Type {
	case "local":
		return NewLocalStorage(config)
	case "s3":
		return NewS3Storage(config)
	case "oss":
		return NewOSSStorage(config)
	case "cos":
		return NewCOSStorage(config)
	case "qiniu":
		return NewQiniuStorage(config)
	case "upyun":
		return NewUpyunStorage(config)
	case "minio":
		return NewMinioStorage(config)
	case "sftp":
		return NewSFTPStorage(config)
	case "ftp":
		return NewFTPStorage(config)
	case "webdav":
		return NewWebDAVStorage(config)
	default:
		return NewLocalStorage(config)
	}
}

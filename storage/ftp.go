package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
)

// FTPStorage FTP 存储
type FTPStorage struct {
	config *Config
	conn   *ftp.ServerConn
	domain string
}

// NewFTPStorage 创建 FTP 存储实例
func NewFTPStorage(cfg *Config) (*FTPStorage, error) {
	// 连接 FTP 服务器
	addr := cfg.Endpoint
	if addr == "" {
		addr = "localhost:21"
	}

	conn, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("FTP 连接失败: %w", err)
	}

	// 登录
	if err := conn.Login(cfg.AccessKey, cfg.SecretKey); err != nil {
		conn.Quit()
		return nil, fmt.Errorf("FTP 登录失败: %w", err)
	}

	domain := cfg.Domain
	if domain == "" {
		domain = fmt.Sprintf("ftp://%s", addr)
	}

	return &FTPStorage{
		config: cfg,
		conn:   conn,
		domain: domain,
	}, nil
}

// Upload 上传文件
func (s *FTPStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
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
	remotePath = filepath.ToSlash(remotePath) // FTP 使用正斜杠

	// 确保远程目录存在
	remoteDir := filepath.Dir(remotePath)
	if err := s.createDirRecursive(remoteDir); err != nil {
		return "", fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 上传文件
	if err := s.conn.Stor(remotePath, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("上传 FTP 文件失败: %w", err)
	}

	return remotePath, nil
}

// UploadBytes 上传字节数据
func (s *FTPStorage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	remotePath := filepath.Join(s.config.BasePath, path)
	remotePath = filepath.ToSlash(remotePath)

	// 确保远程目录存在
	remoteDir := filepath.Dir(remotePath)
	if err := s.createDirRecursive(remoteDir); err != nil {
		return "", fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 上传文件
	if err := s.conn.Stor(remotePath, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("上传 FTP 文件失败: %w", err)
	}

	return remotePath, nil
}

// Delete 删除文件
func (s *FTPStorage) Delete(path string) error {
	if err := s.conn.Delete(path); err != nil {
		return fmt.Errorf("删除 FTP 文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *FTPStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, path)
}

// Exists 检查文件是否存在
func (s *FTPStorage) Exists(path string) (bool, error) {
	_, err := s.conn.GetEntry(path)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *FTPStorage) GetSize(path string) (int64, error) {
	entry, err := s.conn.GetEntry(path)
	if err != nil {
		return 0, err
	}
	return int64(entry.Size), nil
}

// createDirRecursive 递归创建目录
func (s *FTPStorage) createDirRecursive(dir string) error {
	dir = filepath.ToSlash(dir)
	if dir == "" || dir == "." || dir == "/" {
		return nil
	}

	// 检查目录是否存在
	_, err := s.conn.GetEntry(dir)
	if err == nil {
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
	return s.conn.MakeDir(dir)
}

// Close 关闭连接
func (s *FTPStorage) Close() error {
	if s.conn != nil {
		return s.conn.Quit()
	}
	return nil
}

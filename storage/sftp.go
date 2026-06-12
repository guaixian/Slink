package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPStorage SFTP 存储
type SFTPStorage struct {
	config     *Config
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	domain     string
}

// NewSFTPStorage 创建 SFTP 存储实例
func NewSFTPStorage(cfg *Config) (*SFTPStorage, error) {
	// 配置 SSH 客户端
	config := &ssh.ClientConfig{
		User: cfg.AccessKey,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.SecretKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应该验证主机密钥
	}

	// 连接 SSH
	addr := cfg.Endpoint
	if addr == "" {
		addr = "localhost:22"
	}

	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败: %w", err)
	}

	// 创建 SFTP 客户端
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("创建 SFTP 客户端失败: %w", err)
	}

	domain := cfg.Domain
	if domain == "" {
		domain = fmt.Sprintf("sftp://%s", addr)
	}

	return &SFTPStorage{
		config:     cfg,
		sshClient:  sshClient,
		sftpClient: sftpClient,
		domain:     domain,
	}, nil
}

// Upload 上传文件
func (s *SFTPStorage) Upload(file *multipart.FileHeader, path string) (string, error) {
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

	// 确保远程目录存在
	remoteDir := filepath.Dir(remotePath)
	if err := s.sftpClient.MkdirAll(remoteDir); err != nil {
		return "", fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 创建远程文件
	dstFile, err := s.sftpClient.Create(remotePath)
	if err != nil {
		return "", fmt.Errorf("创建远程文件失败: %w", err)
	}
	defer dstFile.Close()

	// 写入数据
	if _, err := dstFile.Write(data); err != nil {
		return "", fmt.Errorf("写入远程文件失败: %w", err)
	}

	return remotePath, nil
}

// UploadBytes 上传字节数据
func (s *SFTPStorage) UploadBytes(data []byte, filename string, path string) (string, error) {
	// 构建完整路径
	remotePath := filepath.Join(s.config.BasePath, path)

	// 确保远程目录存在
	remoteDir := filepath.Dir(remotePath)
	if err := s.sftpClient.MkdirAll(remoteDir); err != nil {
		return "", fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 创建远程文件
	dstFile, err := s.sftpClient.Create(remotePath)
	if err != nil {
		return "", fmt.Errorf("创建远程文件失败: %w", err)
	}
	defer dstFile.Close()

	// 写入数据
	if _, err := io.Copy(dstFile, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("写入远程文件失败: %w", err)
	}

	return remotePath, nil
}

// Delete 删除文件
func (s *SFTPStorage) Delete(path string) error {
	if err := s.sftpClient.Remove(path); err != nil {
		return fmt.Errorf("删除 SFTP 文件失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *SFTPStorage) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, path)
}

// Exists 检查文件是否存在
func (s *SFTPStorage) Exists(path string) (bool, error) {
	_, err := s.sftpClient.Stat(path)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *SFTPStorage) GetSize(path string) (int64, error) {
	info, err := s.sftpClient.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// Close 关闭连接
func (s *SFTPStorage) Close() error {
	if s.sftpClient != nil {
		s.sftpClient.Close()
	}
	if s.sshClient != nil {
		s.sshClient.Close()
	}
	return nil
}

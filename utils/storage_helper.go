package utils

import (
	"Slink/storage"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

// SaveImageWithStorage 使用存储策略保存图片
func SaveImageWithStorage(file *multipart.FileHeader, userID uint, pathRule, fileRule string, storageConfig *storage.Config) (*ImageInfo, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isValidImageExt(ext) {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 读取文件内容
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 计算哈希值
	md5Hash := md5.Sum(data)
	sha1Hash := sha1.Sum(data)
	md5Str := hex.EncodeToString(md5Hash[:])
	sha1Str := hex.EncodeToString(sha1Hash[:])

	// 获取图片尺寸
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	var width, height int64
	if err == nil {
		width = int64(img.Width)
		height = int64(img.Height)
	}

	// 根据用户组配置生成路径和文件名
	pathDir := GeneratePathFromRule(pathRule, userID, file.Filename)
	filename := GenerateFileNameFromRule(fileRule, userID, file.Filename, strings.TrimPrefix(ext, "."))

	// 构建完整路径
	fullPath := filepath.Join(pathDir, filename)

	// 创建存储实例
	store, err := storage.NewStorage(storageConfig)
	if err != nil {
		return nil, fmt.Errorf("创建存储实例失败: %w", err)
	}

	// 上传文件
	uploadedPath, err := store.UploadBytes(data, filename, fullPath)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 获取MIME类型
	mimetype := getMimeType(ext)

	// 计算文件大小（MB）
	sizeMB := float64(len(data)) / (1024 * 1024)

	return &ImageInfo{
		Path:       uploadedPath,
		OriginName: file.Filename,
		Size:       sizeMB,
		Mimetype:   mimetype,
		Extension:  strings.TrimPrefix(ext, "."),
		MD5:        md5Str,
		SHA1:       sha1Str,
		Width:      width,
		Height:     height,
	}, nil
}

// SaveImageBytesWithStorage 使用存储策略保存字节数据
func SaveImageBytesWithStorage(data []byte, filename string, userID uint, pathRule, fileRule string, storageConfig *storage.Config) (*ImageInfo, error) {
	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(filename))
	if !isValidImageExt(ext) {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 计算哈希值
	md5Hash := md5.Sum(data)
	sha1Hash := sha1.Sum(data)
	md5Str := hex.EncodeToString(md5Hash[:])
	sha1Str := hex.EncodeToString(sha1Hash[:])

	// 获取图片尺寸
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	var width, height int64
	if err == nil {
		width = int64(img.Width)
		height = int64(img.Height)
	}

	// 根据用户组配置生成路径和文件名
	pathDir := GeneratePathFromRule(pathRule, userID, filename)
	newFilename := GenerateFileNameFromRule(fileRule, userID, filename, strings.TrimPrefix(ext, "."))

	// 构建完整路径
	fullPath := filepath.Join(pathDir, newFilename)

	// 创建存储实例
	store, err := storage.NewStorage(storageConfig)
	if err != nil {
		return nil, fmt.Errorf("创建存储实例失败: %w", err)
	}

	// 上传文件
	uploadedPath, err := store.UploadBytes(data, newFilename, fullPath)
	if err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 获取MIME类型
	mimetype := getMimeType(ext)

	// 计算文件大小（MB）
	sizeMB := float64(len(data)) / (1024 * 1024)

	return &ImageInfo{
		Path:       uploadedPath,
		OriginName: filename,
		Size:       sizeMB,
		Mimetype:   mimetype,
		Extension:  strings.TrimPrefix(ext, "."),
		MD5:        md5Str,
		SHA1:       sha1Str,
		Width:      width,
		Height:     height,
	}, nil
}

// GetStorageConfigFromStrategy 从策略配置转换为存储配置
func GetStorageConfigFromStrategy(strategyConfig map[string]interface{}) *storage.Config {
	config := &storage.Config{
		Extra: make(map[string]string),
	}

	// 提取基本配置
	if v, ok := strategyConfig["type"].(string); ok {
		config.Type = v
	}
	if v, ok := strategyConfig["endpoint"].(string); ok {
		config.Endpoint = v
	}
	if v, ok := strategyConfig["access_key"].(string); ok {
		config.AccessKey = v
	}
	if v, ok := strategyConfig["secret_key"].(string); ok {
		config.SecretKey = v
	}
	if v, ok := strategyConfig["bucket"].(string); ok {
		config.Bucket = v
	}
	if v, ok := strategyConfig["region"].(string); ok {
		config.Region = v
	}
	if v, ok := strategyConfig["domain"].(string); ok {
		config.Domain = v
	}
	if v, ok := strategyConfig["base_path"].(string); ok {
		config.BasePath = v
	}
	if v, ok := strategyConfig["use_ssl"].(bool); ok {
		config.UseSSL = v
	}

	// 提取额外配置
	for key, value := range strategyConfig {
		if key != "type" && key != "endpoint" && key != "access_key" &&
			key != "secret_key" && key != "bucket" && key != "region" &&
			key != "domain" && key != "base_path" && key != "use_ssl" {
			if strValue, ok := value.(string); ok {
				config.Extra[key] = strValue
			}
		}
	}

	return config
}

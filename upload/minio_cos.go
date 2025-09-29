package upload

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"novel-api/config"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioUploader Minio上传器
type MinioUploader struct {
	client *minio.Client
	cfg    *config.Config
}

// NewMinioUploader 创建新的Minio上传器
func NewMinioUploader(cfg *config.Config) (*MinioUploader, error) {
	// 处理endpoint，移除协议前缀
	endpoint := cfg.Minio.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	// 创建Minio客户端
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建Minio客户端失败: %v", err)
	}

	log.Printf("Minio客户端创建成功，endpoint: %s, useSSL: %v", endpoint, cfg.Minio.UseSSL)
	return &MinioUploader{
		client: client,
		cfg:    cfg,
	}, nil
}

// UploadFromBytes 从字节数组上传文件（实现Uploader接口）
func (m *MinioUploader) UploadFromBytes(data []byte, fileName, folder string) (*UploadResponse, error) {
	// 构建文件key
	key := m.buildFileKey(fileName, folder)

	// 创建读取器
	reader := bytes.NewReader(data)

	// 检查存储桶是否存在，如果不存在则创建
	ctx := context.Background()
	exists, err := m.client.BucketExists(ctx, m.cfg.Minio.BucketName)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("检查存储桶失败: %v", err),
		}, err
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.cfg.Minio.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return &UploadResponse{
				Success: false,
				Message: fmt.Sprintf("创建存储桶失败: %v", err),
			}, err
		}
		log.Printf("存储桶 %s 创建成功", m.cfg.Minio.BucketName)
	}

	// 上传文件
	_, err = m.client.PutObject(ctx, m.cfg.Minio.BucketName, key, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: getContentType(fileName),
	})
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("文件上传失败: %v", err),
		}, err
	}

	// 构建访问URL
	fileURL := m.buildFileURL(key)

	return &UploadResponse{
		Success: true,
		Message: "文件上传成功",
		Data: struct {
			URL      string `json:"url"`
			Key      string `json:"key"`
			Size     int64  `json:"size"`
			FileName string `json:"filename"`
		}{
			URL:      fileURL,
			Key:      key,
			Size:     int64(len(data)),
			FileName: fileName,
		},
	}, nil
}

// buildFileKey 构建文件在Minio中的key
func (m *MinioUploader) buildFileKey(fileName, folder string) string {
	// 生成时间戳前缀避免文件名冲突
	timestamp := time.Now().Format("20060102150405")

	// 清理文件名
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = fmt.Sprintf("file_%s", timestamp)
	}

	// 构建完整路径
	var key string
	if folder != "" {
		folder = strings.Trim(folder, "/")
		key = fmt.Sprintf("%s/%s_%s", folder, timestamp, fileName)
	} else {
		key = fmt.Sprintf("uploads/%s_%s", timestamp, fileName)
	}

	return key
}

// buildFileURL 构建文件访问URL
func (m *MinioUploader) buildFileURL(key string) string {
	// 如果配置了BaseURL，使用BaseURL
	if m.cfg.Minio.BaseURL != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(m.cfg.Minio.BaseURL, "/"), m.cfg.Minio.BucketName, key)
	}

	// 否则使用默认的endpoint构建URL
	protocol := "http"
	if m.cfg.Minio.UseSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, m.cfg.Minio.Endpoint, m.cfg.Minio.BucketName, key)
}

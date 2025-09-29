package upload

import (
	"fmt"
	"log"
	"novel-api/config"
	"strings"
)

// Uploader 通用上传接口
type Uploader interface {
	UploadFromBytes(data []byte, fileName, folder string) (*UploadResponse, error)
}

// CreateUploader 根据配置创建对应的上传器
func CreateUploader(cfg *config.Config) (Uploader, error) {
	bucketType := strings.ToLower(strings.TrimSpace(cfg.COS.Bucket))

	switch bucketType {
	case "tengxun", "tencent":
		log.Printf("使用腾讯云COS上传器")
		return NewTencentCOSUploader(cfg)
	case "minio":
		log.Printf("使用Minio上传器")
		return NewMinioUploader(cfg)
	case "alist":
		log.Printf("使用Alist上传器")
		return NewAlistUploader(cfg)
	default:
		return nil, fmt.Errorf("不支持的存储桶类型: %s，支持的类型: Tengxun, Minio, Alist", cfg.COS.Bucket)
	}
}

// UploadFile 通用上传函数，只需要传入文件数据和文件名
func UploadFile(data []byte, fileName string, cfg *config.Config) (*UploadResponse, error) {
	// 创建上传器
	uploader, err := CreateUploader(cfg)
	if err != nil {
		log.Printf("创建上传器失败: %v", err)
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("创建上传器失败: %v", err),
		}, err
	}

	// 上传文件到指定文件夹
	response, err := uploader.UploadFromBytes(data, fileName, "nai-images")
	if err != nil {
		log.Printf("文件上传失败: %v", err)
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("文件上传失败: %v", err),
		}, err
	}

	log.Printf("文件上传成功: %s", response.Data.URL)
	return response, nil
}

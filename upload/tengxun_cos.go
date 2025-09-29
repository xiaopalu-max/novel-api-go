package upload

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"novel-api/config"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// UploadRequest 上传请求结构体
type UploadRequest struct {
	Base64Data string `json:"base64_data,omitempty"` // Base64编码的文件数据
	FileName   string `json:"filename"`              // 文件名
	FileType   string `json:"file_type,omitempty"`   // 文件类型 (如: image/png)
	Folder     string `json:"folder,omitempty"`      // 上传到的文件夹路径
}

// UploadResponse 上传响应结构体
type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		URL      string `json:"url"`      // 文件访问URL
		Key      string `json:"key"`      // 文件在COS中的key
		Size     int64  `json:"size"`     // 文件大小
		FileName string `json:"filename"` // 文件名
	} `json:"data,omitempty"`
}

// TencentCOSUploader 腾讯云COS上传器
type TencentCOSUploader struct {
	client *cos.Client
	cfg    *config.Config
}

// NewTencentCOSUploader 创建新的腾讯云COS上传器
func NewTencentCOSUploader(cfg *config.Config) (*TencentCOSUploader, error) {
	// 构建COS URL
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com",
		cfg.TencentCOS.Bucket, cfg.TencentCOS.Region))
	if err != nil {
		return nil, fmt.Errorf("invalid COS URL: %v", err)
	}

	// 创建COS客户端
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.TencentCOS.SecretID,
			SecretKey: cfg.TencentCOS.SecretKey,
		},
	})

	return &TencentCOSUploader{
		client: client,
		cfg:    cfg,
	}, nil
}

// UploadFromBase64 从Base64数据上传文件
func (u *TencentCOSUploader) UploadFromBase64(base64Data, fileName, folder string) (*UploadResponse, error) {
	// 解码Base64数据
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("Base64解码失败: %v", err),
		}, err
	}

	return u.uploadBytes(data, fileName, folder)
}

// UploadFromBytes 从字节数组上传文件（公开方法）
func (u *TencentCOSUploader) UploadFromBytes(data []byte, fileName, folder string) (*UploadResponse, error) {
	return u.uploadBytes(data, fileName, folder)
}

// UploadFromBytes 从字节数组上传文件
func (u *TencentCOSUploader) uploadBytes(data []byte, fileName, folder string) (*UploadResponse, error) {
	// 构建文件key
	key := u.buildFileKey(fileName, folder)

	// 创建读取器
	reader := bytes.NewReader(data)

	// 上传文件
	_, err := u.client.Object.Put(context.Background(), key, reader, nil)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("文件上传失败: %v", err),
		}, err
	}

	// 构建访问URL
	fileURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(u.cfg.TencentCOS.BaseURL, "/"), key)

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

// buildFileKey 构建文件在COS中的key
func (u *TencentCOSUploader) buildFileKey(fileName, folder string) string {
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

// getFileExtension 从文件名获取扩展名
func getFileExtension(fileName string) string {
	return filepath.Ext(fileName)
}

// getContentType 根据文件扩展名获取Content-Type
func getContentType(fileName string) string {
	ext := strings.ToLower(getFileExtension(fileName))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

// HandleUpload 处理上传请求的HTTP处理函数
func HandleUpload(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// 处理OPTIONS请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 只允许POST请求
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Message: "只支持POST请求",
		})
		return
	}

	// 创建上传器
	uploader, err := NewTencentCOSUploader(cfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Message: fmt.Sprintf("上传器初始化失败: %v", err),
		})
		return
	}

	// 检查Content-Type决定处理方式
	contentType := r.Header.Get("Content-Type")

	var response *UploadResponse

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// 处理multipart/form-data上传
		response = handleMultipartUpload(r, uploader)
	} else if strings.HasPrefix(contentType, "application/json") {
		// 处理JSON格式上传
		response = handleJSONUpload(r, uploader)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Success: false,
			Message: "不支持的Content-Type，请使用multipart/form-data或application/json",
		})
		return
	}

	// 返回响应
	if response.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(response)
}

// handleMultipartUpload 处理multipart/form-data上传
func handleMultipartUpload(r *http.Request, uploader *TencentCOSUploader) *UploadResponse {
	// 解析multipart form，最大32MB
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("解析表单失败: %v", err),
		}
	}

	// 获取文件
	file, header, err := r.FormFile("file")
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("获取文件失败: %v", err),
		}
	}
	defer file.Close()

	// 读取文件数据
	data, err := io.ReadAll(file)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("读取文件失败: %v", err),
		}
	}

	// 获取文件夹参数
	folder := r.FormValue("folder")

	// 上传文件
	response, err := uploader.uploadBytes(data, header.Filename, folder)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("上传失败: %v", err),
		}
	}

	return response
}

// handleJSONUpload 处理JSON格式上传
func handleJSONUpload(r *http.Request, uploader *TencentCOSUploader) *UploadResponse {
	var req UploadRequest

	// 解析JSON请求
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("JSON解析失败: %v", err),
		}
	}

	// 验证必要字段
	if req.Base64Data == "" {
		return &UploadResponse{
			Success: false,
			Message: "base64_data字段不能为空",
		}
	}

	if req.FileName == "" {
		return &UploadResponse{
			Success: false,
			Message: "filename字段不能为空",
		}
	}

	// 上传文件
	response, err := uploader.UploadFromBase64(req.Base64Data, req.FileName, req.Folder)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("上传失败: %v", err),
		}
	}

	return response
}

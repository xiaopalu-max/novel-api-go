package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"novel-api/config"
	"strings"
	"time"
)

// LskyUploader 兰空图床上传器
type LskyUploader struct {
	cfg *config.Config
}

// LskyUploadResponse 兰空图床上传响应结构
type LskyUploadResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID         int     `json:"id"` // 图片ID
		Key        string  `json:"key"`
		Name       string  `json:"name"`
		Pathname   string  `json:"pathname"`
		OriginName string  `json:"origin_name"`
		Size       float64 `json:"size"` // 单位KB，使用float64
		Mimetype   string  `json:"mimetype"`
		Extension  string  `json:"extension"`
		MD5        string  `json:"md5"`
		SHA1       string  `json:"sha1"`
		Links      struct {
			URL              string `json:"url"`
			HTML             string `json:"html"`
			BBCode           string `json:"bbcode"`
			Markdown         string `json:"markdown"`
			MarkdownWithLink string `json:"markdown_with_link"`
			ThumbnailURL     string `json:"thumbnail_url"`
		} `json:"links"`
	} `json:"data"`
}

// NewLskyUploader 创建新的兰空图床上传器
func NewLskyUploader(cfg *config.Config) (*LskyUploader, error) {
	uploader := &LskyUploader{
		cfg: cfg,
	}

	// 验证配置
	if cfg.Lsky.BaseURL == "" {
		return nil, fmt.Errorf("Lsky配置不完整：BaseURL不能为空")
	}

	if cfg.Lsky.Token == "" {
		return nil, fmt.Errorf("Lsky配置不完整：Token不能为空")
	}

	log.Printf("兰空图床上传器初始化成功")
	return uploader, nil
}

// UploadFromBytes 从字节数组上传文件（实现Uploader接口）
func (l *LskyUploader) UploadFromBytes(data []byte, fileName, folder string) (*UploadResponse, error) {
	// 生成时间戳前缀避免文件名冲突
	timestamp := time.Now().Format("20060102150405")

	// 清理文件名
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = fmt.Sprintf("file_%s.png", timestamp)
	}

	// 添加时间戳前缀
	finalFileName := fmt.Sprintf("%s_%s", timestamp, fileName)

	// 创建multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件
	part, err := writer.CreateFormFile("file", finalFileName)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("创建form文件失败: %v", err),
		}, err
	}

	if _, err := part.Write(data); err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("写入文件数据失败: %v", err),
		}, err
	}

	// 如果配置了存储策略ID，添加到表单
	if l.cfg.Lsky.StrategyID > 0 {
		if err := writer.WriteField("strategy_id", fmt.Sprintf("%d", l.cfg.Lsky.StrategyID)); err != nil {
			log.Printf("警告：添加存储策略ID失败: %v", err)
		}
	}

	writer.Close()

	// 构建上传URL - 兰空图床的上传接口是 /upload
	uploadURL := fmt.Sprintf("%s/api/v1/upload", strings.TrimSuffix(l.cfg.Lsky.BaseURL, "/"))

	// 创建请求
	req, err := http.NewRequest("POST", uploadURL, &buf)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("创建上传请求失败: %v", err),
		}, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", l.cfg.Lsky.Token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Novel-API-Go/1.0")

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("发送上传请求失败: %v", err),
		}, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("读取上传响应失败: %v", err),
		}, err
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("上传失败，HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("上传失败，HTTP状态码: %d", resp.StatusCode)
	}

	var lskyResp LskyUploadResponse
	if err := json.Unmarshal(body, &lskyResp); err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("解析上传响应失败: %v, 响应内容: %s", err, string(body)),
		}, err
	}

	if !lskyResp.Status {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("上传失败: %s", lskyResp.Message),
		}, fmt.Errorf("上传失败: %s", lskyResp.Message)
	}

	log.Printf("文件上传成功: %s", lskyResp.Data.Links.URL)

	// 将KB转换为字节
	sizeInBytes := int64(lskyResp.Data.Size * 1024)

	return &UploadResponse{
		Success: true,
		Message: "文件上传成功",
		Data: struct {
			URL      string `json:"url"`
			Key      string `json:"key"`
			Size     int64  `json:"size"`
			FileName string `json:"filename"`
		}{
			URL:      lskyResp.Data.Links.URL,
			Key:      lskyResp.Data.Key,
			Size:     sizeInBytes,
			FileName: fileName,
		},
	}, nil
}

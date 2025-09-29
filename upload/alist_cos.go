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
	"path/filepath"
	"strings"
	"time"
)

// AlistUploader Alist上传器
type AlistUploader struct {
	cfg   *config.Config
	token string
}

// AlistLoginResponse Alist登录响应结构
type AlistLoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

// AlistUploadResponse Alist上传响应结构
type AlistUploadResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// AlistFileInfo Alist文件信息响应结构
type AlistFileInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Name     string        `json:"name"`
		Size     int64         `json:"size"`
		IsDir    bool          `json:"is_dir"`
		Modified string        `json:"modified"`
		Created  string        `json:"created"`
		Sign     string        `json:"sign"`
		Thumb    string        `json:"thumb"`
		Type     int           `json:"type"`
		RawURL   string        `json:"raw_url"`
		Readme   string        `json:"readme"`
		Header   string        `json:"header"`
		Provider string        `json:"provider"`
		Related  []interface{} `json:"related"`
	} `json:"data"`
}

// NewAlistUploader 创建新的Alist上传器
func NewAlistUploader(cfg *config.Config) (*AlistUploader, error) {
	uploader := &AlistUploader{
		cfg: cfg,
	}

	// 如果配置中有token，直接使用
	if cfg.Alist.Token != "" {
		uploader.token = cfg.Alist.Token
		log.Printf("使用配置中的Alist Token")
		return uploader, nil
	}

	// 否则使用用户名密码登录获取token
	if cfg.Alist.Username != "" && cfg.Alist.Password != "" {
		token, err := uploader.login()
		if err != nil {
			return nil, fmt.Errorf("Alist登录失败: %v", err)
		}
		uploader.token = token
		log.Printf("Alist登录成功，获取到Token")
		return uploader, nil
	}

	return nil, fmt.Errorf("Alist配置不完整：需要提供token或用户名密码")
}

// login 登录获取token
func (a *AlistUploader) login() (string, error) {
	loginURL := fmt.Sprintf("%s/api/auth/login", strings.TrimSuffix(a.cfg.Alist.BaseURL, "/"))

	loginData := map[string]string{
		"username": a.cfg.Alist.Username,
		"password": a.cfg.Alist.Password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return "", fmt.Errorf("序列化登录数据失败: %v", err)
	}

	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("发送登录请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取登录响应失败: %v", err)
	}

	var loginResp AlistLoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("解析登录响应失败: %v", err)
	}

	if loginResp.Code != 200 {
		return "", fmt.Errorf("登录失败: %s", loginResp.Message)
	}

	return loginResp.Data.Token, nil
}

// getRawURL 获取文件的真实访问链接
func (a *AlistUploader) getRawURL(filePath string) (string, error) {
	getURL := fmt.Sprintf("%s/api/fs/get", strings.TrimSuffix(a.cfg.Alist.BaseURL, "/"))

	requestData := map[string]interface{}{
		"path":     filePath,
		"password": "",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("序列化请求数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", getURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", a.token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var fileInfo AlistFileInfo
	if err := json.Unmarshal(body, &fileInfo); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if fileInfo.Code != 200 {
		return "", fmt.Errorf("获取文件信息失败: %s", fileInfo.Message)
	}

	if fileInfo.Data.RawURL == "" {
		return "", fmt.Errorf("文件未找到或无法获取访问链接")
	}

	return fileInfo.Data.RawURL, nil
}

// UploadFromBytes 从字节数组上传文件（实现Uploader接口）
func (a *AlistUploader) UploadFromBytes(data []byte, fileName, folder string) (*UploadResponse, error) {
	// 构建文件key和上传路径
	key := a.buildFileKey(fileName, folder)
	uploadPath := filepath.Dir(key)
	finalFileName := filepath.Base(key)

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

	writer.Close()

	// 构建上传URL
	uploadURL := fmt.Sprintf("%s/api/fs/form", strings.TrimSuffix(a.cfg.Alist.BaseURL, "/"))

	// 创建请求
	req, err := http.NewRequest("PUT", uploadURL, &buf)
	if err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("创建上传请求失败: %v", err),
		}, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", a.token)
	req.Header.Set("File-Path", fmt.Sprintf("%s/%s", strings.TrimSuffix(uploadPath, "/"), finalFileName))

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

	var uploadResp AlistUploadResponse
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("解析上传响应失败: %v", err),
		}, err
	}

	if uploadResp.Code != 200 {
		return &UploadResponse{
			Success: false,
			Message: fmt.Sprintf("上传失败: %s", uploadResp.Message),
		}, fmt.Errorf("上传失败: %s", uploadResp.Message)
	}

	log.Printf("文件上传成功，正在获取真实访问链接...")

	// 获取文件的真实访问链接
	rawURL, err := a.getRawURL(key)
	if err != nil {
		log.Printf("警告：获取真实链接失败，使用备用链接: %v", err)
		// 如果获取真实链接失败，使用备用的直链格式
		rawURL = a.buildFileURL(key)
	} else {
		log.Printf("获取到真实访问链接: %s", rawURL)
	}

	return &UploadResponse{
		Success: true,
		Message: "文件上传成功",
		Data: struct {
			URL      string `json:"url"`
			Key      string `json:"key"`
			Size     int64  `json:"size"`
			FileName string `json:"filename"`
		}{
			URL:      rawURL,
			Key:      key,
			Size:     int64(len(data)),
			FileName: fileName,
		},
	}, nil
}

// buildFileKey 构建文件在Alist中的路径
func (a *AlistUploader) buildFileKey(fileName, folder string) string {
	// 生成时间戳前缀避免文件名冲突
	timestamp := time.Now().Format("20060102150405")

	// 清理文件名
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = fmt.Sprintf("file_%s.png", timestamp)
	}

	// 构建完整路径
	basePath := strings.TrimSuffix(a.cfg.Alist.Path, "/")
	if basePath == "" {
		basePath = "/uploads"
	}

	var key string
	if folder != "" {
		folder = strings.Trim(folder, "/")
		key = fmt.Sprintf("%s/%s/%s_%s", basePath, folder, timestamp, fileName)
	} else {
		key = fmt.Sprintf("%s/%s_%s", basePath, timestamp, fileName)
	}

	return key
}

// buildFileURL 构建文件访问URL
func (a *AlistUploader) buildFileURL(key string) string {
	// 使用Alist的直链访问格式
	return fmt.Sprintf("%s/d%s", strings.TrimSuffix(a.cfg.Alist.BaseURL, "/"), key)
}

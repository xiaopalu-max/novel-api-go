package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"novel-api/config"
	"novel-api/logs"
	"strconv"
	"strings"
	"time"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Password string `json:"password"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

// LogsResponse 日志查询响应结构
type LogsResponse struct {
	Success  bool            `json:"success"`
	Data     []logs.ImageLog `json:"data"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Message  string          `json:"message,omitempty"`
}

// 简单的token存储（生产环境应使用更安全的方式）
var validTokens = make(map[string]time.Time)

// Login 处理登录请求
func Login(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := LoginResponse{
			Success: false,
			Message: "无效的请求格式",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证密码
	if req.Password != cfg.LogsAdmin.Password {
		response := LoginResponse{
			Success: false,
			Message: "密码错误",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 生成token
	token := generateToken()
	validTokens[token] = time.Now().Add(24 * time.Hour) // token有效期24小时

	response := LoginResponse{
		Success: true,
		Token:   token,
		Message: "登录成功",
	}
	json.NewEncoder(w).Encode(response)
}

// QueryLogs 查询日志
func QueryLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 验证token
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if !isValidToken(token) {
		response := LogsResponse{
			Success: false,
			Message: "未授权访问",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	keyword := r.URL.Query().Get("keyword")

	// 查询日志
	logsList, total, err := logs.GetLogs(page, pageSize, keyword)
	if err != nil {
		log.Printf("查询日志失败: %v", err)
		response := LogsResponse{
			Success: false,
			Message: "查询失败",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := LogsResponse{
		Success:  true,
		Data:     logsList,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	json.NewEncoder(w).Encode(response)
}

// GetLogDetail 获取日志详情
func GetLogDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 验证token
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if !isValidToken(token) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未授权访问",
		})
		return
	}

	// 获取ID参数
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "缺少ID参数",
		})
		return
	}

	// 查询日志
	logDetail, err := logs.GetLogByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "查询失败",
		})
		return
	}

	if logDetail == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "日志不存在",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    logDetail,
	})
}

// isValidToken 验证token是否有效
func isValidToken(token string) bool {
	if token == "" {
		return false
	}

	expireTime, exists := validTokens[token]
	if !exists {
		return false
	}

	if time.Now().After(expireTime) {
		delete(validTokens, token)
		return false
	}

	return true
}

// generateToken 生成随机token
func generateToken() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(32))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

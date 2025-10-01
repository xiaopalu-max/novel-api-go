package logs

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// ImageLog 图片生成日志结构
type ImageLog struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Model     string    `json:"model"`
	Prompt    string    `json:"prompt"`
	ImageURL  string    `json:"image_url"`
	UserIP    string    `json:"user_ip"`
	Status    string    `json:"status"` // success, failed
	Error     string    `json:"error,omitempty"`
}

var (
	logFile  *os.File
	logMutex sync.Mutex
	logPath  = "logs/image_logs.json"
)

// InitLogger 初始化日志系统
func InitLogger() error {
	// 创建logs目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return err
	}

	// 打开或创建日志文件
	var err error
	logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	return nil
}

// LogImage 记录图片生成日志
func LogImage(log ImageLog) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	log.Timestamp = time.Now()
	if log.ID == "" {
		log.ID = generateID()
	}

	data, err := json.Marshal(log)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = logFile.Write(data)
	if err != nil {
		return err
	}

	return logFile.Sync()
}

// GetLogs 获取日志列表（支持分页和筛选）
func GetLogs(page, pageSize int, keyword string) ([]ImageLog, int, error) {
	logMutex.Lock()
	defer logMutex.Unlock()

	// 读取所有日志
	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ImageLog{}, 0, nil
		}
		return nil, 0, err
	}
	defer file.Close()

	var allLogs []ImageLog
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var log ImageLog
		if err := decoder.Decode(&log); err != nil {
			continue
		}
		allLogs = append(allLogs, log)
	}

	// 筛选日志
	var filteredLogs []ImageLog
	for i := len(allLogs) - 1; i >= 0; i-- {
		log := allLogs[i]
		if keyword == "" || containsKeyword(log, keyword) {
			filteredLogs = append(filteredLogs, log)
		}
	}

	total := len(filteredLogs)

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []ImageLog{}, total, nil
	}
	if end > total {
		end = total
	}

	return filteredLogs[start:end], total, nil
}

// GetLogByID 根据ID获取单条日志
func GetLogByID(id string) (*ImageLog, error) {
	logMutex.Lock()
	defer logMutex.Unlock()

	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var log ImageLog
		if err := decoder.Decode(&log); err != nil {
			continue
		}
		if log.ID == id {
			return &log, nil
		}
	}

	return nil, nil
}

// containsKeyword 检查日志是否包含关键词
func containsKeyword(log ImageLog, keyword string) bool {
	return contains(log.Model, keyword) ||
		contains(log.Prompt, keyword) ||
		contains(log.UserIP, keyword) ||
		contains(log.Status, keyword)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// generateID 生成唯一ID
func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// Close 关闭日志文件
func Close() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

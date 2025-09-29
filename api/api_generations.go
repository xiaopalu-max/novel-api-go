package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"novel-api/config"
	"novel-api/models"
	"regexp"
	"strings"
	"time"
)

// GenerationRequest 定义 OpenAI DALL-E 格式的请求结构体
type GenerationRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	N       int    `json:"n,omitempty"`       // 生成图片数量，默认为1
	Size    string `json:"size,omitempty"`    // 图片尺寸，如 "1024x1024"
	Quality string `json:"quality,omitempty"` // 图片质量，如 "standard" 或 "hd"
}

// GenerationResponse 定义 OpenAI DALL-E 格式的响应结构体
type GenerationResponse struct {
	Created int64                 `json:"created"`
	Data    []GenerationImageData `json:"data"`
}

// GenerationImageData 定义生成的图片数据结构
type GenerationImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// 提取链接的函数 (复用自 completions)
func extractLinksFromPrompt(prompt string) []string {
	re := regexp.MustCompile(`https?://[^\s]+`)
	matches := re.FindAllString(prompt, -1)
	return matches
}

// Generations 处理 OpenAI DALL-E 格式的画图请求
func Generations(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// 设置 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 如果是 OPTIONS 请求，直接返回 200 OK
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. 获取 Authorization 请求头的值
	authHeader := r.Header.Get("Authorization")
	authHeader = strings.TrimPrefix(authHeader, "Bearer ")

	fmt.Println("Generations API 秘钥打印：", authHeader)

	// 2. 解析请求体
	var req GenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode generation request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Generation request: Model=%s, Prompt=%s", req.Model, req.Prompt)

	// 3. 处理默认值
	if req.N == 0 {
		req.N = 1 // 默认生成1张图片
	}

	// 4. 获取用户输入的提示词
	userInput := req.Prompt

	// 5. 如果启用翻译，则翻译用户输入
	if cfg.Translation.Enable {
		translatedInput, err := TranslateText(userInput, cfg)
		if err != nil {
			log.Printf("Translation failed, using original text: %v", err)
		} else {
			log.Printf("Translation enabled, translated: %s -> %s", userInput, translatedInput)
			userInput = translatedInput
		}
	}

	// 6. 提取用户输入中的链接 (用于参考图像)
	imageURL := extractLinksFromPrompt(userInput)
	var base64String string
	if len(imageURL) > 0 {
		// 选择第一个提取到的链接
		imageURLS := imageURL[0]
		// 解析图片为base64
		base64String, _ = ImageURLToBase64(imageURLS)
	}

	// 7. 生成一个随机种子
	rand.Seed(time.Now().UnixNano())
	randomSeed := rand.Intn(1000000)

	// 8. 构建兼容的 ChatRequest 结构 (复用现有模型)
	compatibleReq := config.ChatRequest{
		Authorization: authHeader,
		Model:         req.Model,
		Messages: []config.Message{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
	}

	// 9. 根据模型来调用相应的生成函数 (DALL-E 格式)
	// 标识这是 DALL-E 格式请求
	isDallRequest := true

	switch req.Model {
	case "nai-diffusion-3":
		models.Nai3WithFormat(w, r, compatibleReq, randomSeed, base64String, authHeader, cfg, userInput, isDallRequest)
	case "nai-diffusion-furry-3":
		models.Nai3WithFormat(w, r, compatibleReq, randomSeed, base64String, authHeader, cfg, userInput, isDallRequest)
	case "nai-diffusion-4-full":
		models.Nai4WithFormat(w, r, compatibleReq, randomSeed, base64String, authHeader, cfg, userInput, nil, isDallRequest)
	case "nai-diffusion-4-curated-preview":
		models.Nai4WithFormat(w, r, compatibleReq, randomSeed, base64String, authHeader, cfg, userInput, nil, isDallRequest)
	case "nai-diffusion-4-5-curated":
		models.Nai4WithFormat(w, r, compatibleReq, randomSeed, base64String, authHeader, cfg, userInput, nil, isDallRequest)
	case "nai-diffusion-4-5-full":
		models.Nai4WithFormat(w, r, compatibleReq, randomSeed, base64String, authHeader, cfg, userInput, nil, isDallRequest)
	default:
		// 对于不识别的模型，尝试使用默认的 NAI-3 模型
		log.Printf("Unknown model '%s', falling back to nai-diffusion-3", req.Model)
		compatibleReq.Model = "nai-diffusion-3"
		models.Nai3WithFormat(w, r, compatibleReq, randomSeed, base64String, authHeader, cfg, userInput, isDallRequest)
	}
}

// GenerationsJSON 处理 OpenAI DALL-E 格式的画图请求并返回 JSON 响应 (非流式)
func GenerationsJSON(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// 设置 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 如果是 OPTIONS 请求，直接返回 200 OK
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. 获取 Authorization 请求头的值
	authHeader := r.Header.Get("Authorization")
	authHeader = strings.TrimPrefix(authHeader, "Bearer ")

	fmt.Println("GenerationsJSON API 秘钥打印：", authHeader)

	// 2. 解析请求体
	var req GenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode generation JSON request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Generation JSON request: Model=%s, Prompt=%s", req.Model, req.Prompt)

	// 3. 处理默认值
	if req.N == 0 {
		req.N = 1 // 默认生成1张图片
	}

	// 4. 构建响应结构
	response := GenerationResponse{
		Created: time.Now().Unix(),
		Data: []GenerationImageData{
			{
				URL:           "", // 这里会在实际生成后填充
				RevisedPrompt: req.Prompt,
			},
		},
	}

	// 5. 设置响应头并返回 JSON
	w.Header().Set("Content-Type", "application/json")

	// 注意：这里是一个简化版本，实际应该调用相应的模型生成图片
	// 然后获取生成的图片URL填充到响应中
	// 为了保持一致性，建议使用流式响应版本 Generations 函数

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

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

// ChatRequest 定义请求结构体
type ChatRequest struct {
	Authorization string    `json:"Authorization"`
	Messages      []Message `json:"messages"`
	Model         string    `json:"model"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// 提取链接的函数
func extractLinks(userInput string) []string {
	re := regexp.MustCompile(`https?://[^\s]+`)
	matches := re.FindAllString(userInput, -1)
	return matches
}

func Completions(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// 1. 获取 Authorization 请求头的值
	authHeader := r.Header.Get("Authorization")
	authHeader = strings.TrimPrefix(authHeader, "Bearer ")

	fmt.Println("秘钥打印：", authHeader)

	// 如果是 OPTIONS 请求,直接返回 200 OK
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	// 解析请求体
	var req config.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// 获取最后一条用户输入
	var userInput string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userInput = req.Messages[i].Content
			log.Printf("User input found: %s", userInput)
			break
		}
	}

	// 如果启用翻译，则翻译用户输入
	if cfg.Translation.Enable {
		translatedInput, err := TranslateText(userInput, cfg)
		if err != nil {
			log.Printf("Translation failed, using original text: %v", err)
		} else {
			log.Printf("Translation enabled, translated: %s -> %s", userInput, translatedInput)
			userInput = translatedInput
		}
	}

	// 提取用户输入中的链接
	imageURL := extractLinks(userInput)
	var base64String string
	if len(imageURL) > 0 {
		// 选择第一个提取到的链接
		imageURLS := imageURL[0]
		// 解析图片为bash
		base64String, _ = ImageURLToBase64(imageURLS)
		//if err != nil {
		//	log.Fatalf("Error: %v", err)
		//}
	}

	// 调用翻译,翻译为专用提示词

	// 生成一个随机种子
	rand.Seed(time.Now().UnixNano()) // 使用当前时间的纳秒数作为随机数生成器的种子
	randomSeed := rand.Intn(1000000) // 生成一个0到999999之间的随机数

	// 根据模型来请求 url
	if req.Model == "nai-diffusion-3" {
		models.Nai3(w, r, req, randomSeed, base64String, authHeader, cfg, userInput)
	}
	if req.Model == "nai-diffusion-furry-3" {
		models.Nai3(w, r, req, randomSeed, base64String, authHeader, cfg, userInput)
	}
	if req.Model == "nai-diffusion-4-full" {
		models.Nai4(w, r, req, randomSeed, base64String, authHeader, cfg, userInput, nil)
	}
	if req.Model == "nai-diffusion-4-curated-preview" {
		models.Nai4(w, r, req, randomSeed, base64String, authHeader, cfg, userInput, nil)
	}
	if req.Model == "nai-diffusion-4-5-curated" {
		models.Nai4(w, r, req, randomSeed, base64String, authHeader, cfg, userInput, nil)
	}
	if req.Model == "nai-diffusion-4-5-full" {
		models.Nai4(w, r, req, randomSeed, base64String, authHeader, cfg, userInput, nil)
	}
}

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"novel-api/config"
)

// TranslationRequest 定义翻译请求结构体
type TranslationRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// TranslationResponse 定义翻译响应结构体
type TranslationResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// TranslateText 调用 AI 翻译文本
func TranslateText(text string, cfg *config.Config) (string, error) {
	// 检查是否启用翻译
	if !cfg.Translation.Enable {
		log.Println("Translation is disabled, returning original text")
		return text, nil
	}

	// 构建翻译请求
	translateReq := TranslationRequest{
		Model: cfg.Translation.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: cfg.Translation.Role,
			},
			{
				Role:    "user",
				Content: text,
			},
		},
	}

	// 转换为 JSON
	payloadBytes, err := json.Marshal(translateReq)
	if err != nil {
		log.Printf("Failed to marshal translation request: %v", err)
		return text, err
	}

	// 创建 HTTP 请求
	client := &http.Client{}
	req, err := http.NewRequest("POST", cfg.Translation.URL+"/v1/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Failed to create translation request: %v", err)
		return text, err
	}

	// 设置请求头
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Translation.Key)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Translating text: %s", text)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send translation request: %v", err)
		return text, err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Translation API error response: %s", string(bodyBytes))
		return text, fmt.Errorf("translation API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read translation response body: %v", err)
		return text, err
	}

	// 解析响应
	var translateResp TranslationResponse
	err = json.Unmarshal(bodyBytes, &translateResp)
	if err != nil {
		log.Printf("Failed to unmarshal translation response: %v", err)
		return text, err
	}

	// 检查是否有响应内容
	if len(translateResp.Choices) == 0 {
		log.Printf("No translation choices returned")
		return text, fmt.Errorf("no translation choices returned")
	}

	translatedText := translateResp.Choices[0].Message.Content
	log.Printf("Translation successful: %s -> %s", text, translatedText)

	return translatedText, nil
}

package models

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"novel-api/config"
	"novel-api/upload"
	"time"
)

// CharacterPrompt 定义角色提示词结构
type CharacterPrompt struct {
	Prompt  string `json:"prompt"`
	UC      string `json:"uc"`
	Center  Center `json:"center"`
	Enabled bool   `json:"enabled"`
}

// Center 定义中心点坐标
type Center struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// V4Prompt 定义 v4 提示词结构
type V4Prompt struct {
	Caption   V4Caption `json:"caption"`
	UseCoords bool      `json:"use_coords"`
	UseOrder  bool      `json:"use_order"`
}

// V4Caption 定义 v4 标题结构
type V4Caption struct {
	BaseCaption  string        `json:"base_caption"`
	CharCaptions []CharCaption `json:"char_captions"`
}

// CharCaption 定义角色标题结构
type CharCaption struct {
	CharCaption string   `json:"char_caption"`
	Centers     []Center `json:"centers"`
}

// V4NegativePrompt 定义 v4 负面提示词结构
type V4NegativePrompt struct {
	Caption  V4Caption `json:"caption"`
	LegacyUC bool      `json:"legacy_uc"`
}

// NAI4Response 定义 NAI-4 MessagePack 响应结构
type NAI4Response struct {
	EventType string  `msgp:"event_type"`
	SampIx    int     `msgp:"samp_ix"`
	StepIx    int     `msgp:"step_ix"`
	GenID     string  `msgp:"gen_id"`
	Sigma     float64 `msgp:"sigma"`
	Image     []byte  `msgp:"image"`
}

func Nai4(w http.ResponseWriter, r *http.Request, req config.ChatRequest, randomSeed int, base64String string, authHeader string, cfg *config.Config, userInput string, characterPrompts []CharacterPrompt) {
	Nai4WithFormat(w, r, req, randomSeed, base64String, authHeader, cfg, userInput, characterPrompts, false)
}

func Nai4WithFormat(w http.ResponseWriter, r *http.Request, req config.ChatRequest, randomSeed int, base64String string, authHeader string, cfg *config.Config, userInput string, characterPrompts []CharacterPrompt, isDallRequest bool) {
	// 请求连接
	apiURL := "https://image.novelai.net/ai/generate-image"
	log.Println("Preparing payload for NAI-4 API request.")

	// 构建 characterPrompts
	if len(characterPrompts) == 0 {
		// 默认角色提示词，使用配置文件中的反词
		characterPrompts = []CharacterPrompt{
			{
				Prompt:  userInput + ", best quality, very aesthetic, absurdres",
				UC:      cfg.Parameters.CustomAntiWords,
				Center:  Center{X: 0, Y: 0},
				Enabled: true,
			},
		}
	}

	// 构建 v4_prompt 结构
	charCaptions := make([]CharCaption, 0)
	for _, cp := range characterPrompts {
		if cp.Enabled {
			charCaptions = append(charCaptions, CharCaption{
				CharCaption: cp.Prompt,
				Centers:     []Center{cp.Center},
			})
		}
	}

	v4Prompt := V4Prompt{
		Caption: V4Caption{
			BaseCaption:  userInput + ", best quality, very aesthetic, absurdres",
			CharCaptions: charCaptions,
		},
		UseCoords: cfg.Parameters.UseCoords,
		UseOrder:  true,
	}

	// 构建 v4_negative_prompt 结构
	negativeCharCaptions := make([]CharCaption, 0)
	for _, cp := range characterPrompts {
		if cp.Enabled {
			negativeCharCaptions = append(negativeCharCaptions, CharCaption{
				CharCaption: cp.UC,
				Centers:     []Center{cp.Center},
			})
		}
	}

	v4NegativePrompt := V4NegativePrompt{
		Caption: V4Caption{
			BaseCaption:  cfg.Parameters.CustomAntiWords,
			CharCaptions: negativeCharCaptions,
		},
		LegacyUC: cfg.Parameters.LegacyUC,
	}

	// 支持自定义 payload
	payload := map[string]interface{}{
		"input":  userInput + ", best quality, very aesthetic, absurdres",
		"model":  req.Model,
		"action": "generate",
		"parameters": map[string]interface{}{
			"params_version":                        cfg.Parameters.ParamsVersion,
			"width":                                 cfg.Parameters.Width,
			"height":                                cfg.Parameters.Height,
			"scale":                                 cfg.Parameters.Scale,
			"sampler":                               cfg.Parameters.Sampler,
			"steps":                                 cfg.Parameters.Steps,
			"seed":                                  randomSeed,
			"n_samples":                             cfg.Parameters.NSamples,
			"ucPreset":                              cfg.Parameters.UCPreset,
			"qualityToggle":                         cfg.Parameters.QualityToggle,
			"autoSmea":                              cfg.Parameters.AutoSmea,
			"dynamic_thresholding":                  cfg.Parameters.DynamicThresholding,
			"controlnet_strength":                   cfg.Parameters.ControlnetStrength,
			"legacy":                                cfg.Parameters.Legacy,
			"add_original_image":                    cfg.Parameters.AddOriginalImage,
			"cfg_rescale":                           cfg.Parameters.CFGRescale,
			"noise_schedule":                        cfg.Parameters.NoiseSchedule,
			"legacy_v3_extend":                      cfg.Parameters.LegacyV3Extend,
			"skip_cfg_above_sigma":                  cfg.Parameters.SkipCFGAboveSigma,
			"use_coords":                            cfg.Parameters.UseCoords,
			"legacy_uc":                             cfg.Parameters.LegacyUC,
			"normalize_reference_strength_multiple": cfg.Parameters.NormalizeReferenceStrengthMultiple,
			"inpaintImg2ImgStrength":                cfg.Parameters.InpaintImg2ImgStrength,
			"characterPrompts":                      characterPrompts,
			"v4_prompt":                             v4Prompt,
			"v4_negative_prompt":                    v4NegativePrompt,
			"negative_prompt":                       cfg.Parameters.CustomAntiWords,
			"deliberate_euler_ancestral_bug":        cfg.Parameters.DeliberateEulerAncestralBug,
			"prefer_brownian":                       cfg.Parameters.PreferBrownian,
		},
		"use_new_shared_trial": cfg.Parameters.UseNewSharedTrial,
		"recaptcha_token":      " ",
	}

	// 根据是否有有效的 base64String 来决定是否添加参考图像字段
	if base64String != "" {
		payload["parameters"].(map[string]interface{})["reference_image_multiple"] = []interface{}{base64String}
		payload["parameters"].(map[string]interface{})["reference_information_extracted_multiple"] = []interface{}{1}
		payload["parameters"].(map[string]interface{})["reference_strength_multiple"] = []interface{}{0.6}
	}

	// 将 payload 转换为 JSON
	payloadBytes, _ := json.Marshal(payload)
	log.Println("NAI-4 payload marshaled to JSON")

	// 创建新的请求
	client := &http.Client{}
	request, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Failed to create new request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("NAI-4 API request created successfully")

	// 设置请求头
	request.Header.Set("Authorization", "Bearer "+authHeader)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Origin", "https://novelai.net")
	request.Header.Set("Pragma", "no-cache")
	request.Header.Set("Referer", "https://novelai.net/")
	log.Println("NAI-4 request headers set.")

	// 发送请求
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("(NAI-4 发送请求失败)Failed to send request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 检查HTTP响应状态
	log.Printf("NAI-4 HTTP Response Status: %d %s", resp.StatusCode, resp.Status)
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("NAI-4 API Error Response: %s", string(bodyBytes))
		http.Error(w, fmt.Sprintf("NAI-4 API request failed with status %d: %s", resp.StatusCode, string(bodyBytes)), resp.StatusCode)
		return
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Failed to read response body: %v", err)
		return
	}
	log.Printf("NAI-4 response body read successfully. Content-Type: %s, Body length: %d bytes", resp.Header.Get("Content-Type"), len(bodyBytes))

	// 检查响应是否为ZIP格式
	if len(bodyBytes) < 4 {
		log.Printf("Response too short to be a ZIP file: %d bytes", len(bodyBytes))
		http.Error(w, "Invalid response from API", http.StatusInternalServerError)
		return
	}

	// 检查ZIP文件头
	if bodyBytes[0] != 0x50 || bodyBytes[1] != 0x4B {
		log.Printf("Response is not a ZIP file. First 100 bytes: %s", string(bodyBytes[:min(100, len(bodyBytes))]))
		http.Error(w, "API response is not a ZIP file", http.StatusInternalServerError)
		return
	}

	// 创建 ZIP 读取器
	zipReader, err := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes)))
	if err != nil {
		log.Printf("Failed to create zip reader: %v", err)
		log.Printf("Response body (first 200 bytes): %s", string(bodyBytes[:min(200, len(bodyBytes))]))
		http.Error(w, "Failed to read ZIP file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("NAI-4 ZIP file read successfully.")

	// 获取当前时间戳
	timestamp := time.Now().Unix()
	imageName := fmt.Sprintf("nai4_%d.png", timestamp)
	log.Printf("NAI-4 image will be processed as: %s", imageName)

	// 提取指定的图像文件并直接上传到存储服务
	for _, file := range zipReader.File {
		if file.Name == "image_0.png" { // 根据实际文件名进行匹配
			// 打开 ZIP 中的文件
			srcFile, err := file.Open()
			if err != nil {
				http.Error(w, "打开 ZIP 中的文件失败: "+err.Error(), http.StatusInternalServerError)
				log.Printf("打开 ZIP 中的文件失败: %v", err)
				return
			}
			defer srcFile.Close()

			// 将图像数据读取到内存中
			imageData, err := io.ReadAll(srcFile)
			if err != nil {
				http.Error(w, "读取图像数据失败: "+err.Error(), http.StatusInternalServerError)
				log.Printf("读取图像数据失败: %v", err)
				return
			}
			log.Printf("NAI-4 图像数据读取成功，大小: %d bytes", len(imageData))

			var outputs string

			// 使用通用上传函数上传图片
			log.Printf("开始上传 NAI-4 图片: %s", imageName)

			// 调用通用上传函数
			response, err := upload.UploadFile(imageData, imageName, cfg)
			if err != nil {
				log.Printf("NAI-4 图片上传失败: %v", err)
				outputs = fmt.Sprintf("error: 上传失败 - %s", imageName)
			} else {
				log.Printf("NAI-4 图片上传成功: %s", response.Data.URL)
				outputs = response.Data.URL
			}

			publicLink := fmt.Sprintf("![%s](%s)", imageName, outputs)
			fmt.Println(publicLink)

			// 根据请求类型决定响应格式
			if isDallRequest {
				// DALL-E 格式响应
				dallResponse := map[string]interface{}{
					"data": []map[string]interface{}{
						{
							"url": outputs,
						},
					},
					"created": timestamp,
					"usage": map[string]interface{}{
						"prompt_tokens":     0,
						"completion_tokens": 0,
						// "total_tokens":      16384,
						"prompt_tokens_details": map[string]interface{}{
							"cached_tokens_details": map[string]interface{}{},
						},
						"completion_tokens_details": map[string]interface{}{},
						// "output_tokens":             16384,
					},
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(dallResponse)
			} else {
				// 原有的流式聊天响应格式
				sseResponse := fmt.Sprintf(
					"data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":\"%s\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"%s\"},\"logprobs\":null,\"finish_reason\":null}]}\n\n",
					"chatcmpl-"+fmt.Sprintf("%d", timestamp), // 生成一个唯一的 id
					timestamp,
					req.Model,
					publicLink,
				)

				w.Header().Set("Content-Type", "text/event-stream")
				w.Write([]byte(sseResponse))
				w.(http.Flusher).Flush() // 刷新响应缓冲区到客户端
			}
			break
		}
	}

	// 如果不是 DALL-E 请求，则结束流式输出
	if !isDallRequest {
		w.Write([]byte("event: end\n\n"))
		w.(http.Flusher).Flush() // 刷新最后一条消息
	}
}

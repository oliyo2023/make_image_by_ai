package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"

	"make_image_by_ai/config"
	"make_image_by_ai/models"
	"make_image_by_ai/utils"
)

// ImageService 图像服务
type ImageService struct {
	config    *config.Config
	r2Service *R2Service
	d1Service *D1Service
}

// NewImageService 创建图像服务实例
func NewImageService(cfg *config.Config) (*ImageService, error) {
	// 创建 R2 服务
	r2Service, err := NewR2Service(cfg)
	if err != nil {
		log.Printf("警告: R2 服务初始化失败: %v，将使用本地存储", err)
		r2Service = nil
	}

	// 创建 D1 服务
	d1Service, err := NewD1Service(cfg)
	if err != nil {
		log.Printf("警告: D1 服务初始化失败: %v，将无法保存图片记录", err)
		d1Service = nil
	}

	return &ImageService{
		config:    cfg,
		r2Service: r2Service,
		d1Service: d1Service,
	}, nil
}

// GenerateImage 生成图像
func (s *ImageService) GenerateImage(req *models.ImageGenerationRequest) (*models.ImageGenerationResponse, error) {
	// 翻译提示词
	translatedPrompt, err := s.translatePrompt(req.Prompt)
	if err != nil {
		return nil, fmt.Errorf("翻译失败: %v", err)
	}

	log.Printf("原始提示词: %s", req.Prompt)
	log.Printf("翻译后提示词: %s", translatedPrompt)

	// 选择模型
	model := req.Model
	if model == "" {
		model = s.config.DefaultOpenRouterModel()
	}

	// 生成图像
	imageURL, err := s.generateImageWithOpenRouter(translatedPrompt, model)
	if err != nil {
		return nil, fmt.Errorf("图像生成失败: %v", err)
	}

	// 保存图像（优先使用 R2，失败时回退到本地存储）
	var finalURL string
	var imageMetadata *utils.ImageMetadata
	if s.r2Service != nil {
		// 先保存到本地获取元数据，然后尝试上传到 R2
		compressionConfig := &utils.ImageCompressionConfig{
			MaxWidth:     s.config.ImageMaxWidth(),
			MaxHeight:    s.config.ImageMaxHeight(),
			Quality:      s.config.ImageQuality(),
			Format:       s.config.ImageFormat(),
			EnableResize: s.config.ImageEnableResize(),
		}
		var err error
		imageMetadata, err = utils.DownloadAndSaveImage(imageURL, req.Prompt, translatedPrompt, s.config.ImagesDir(), compressionConfig)
		if err != nil {
			return nil, fmt.Errorf("保存图像失败: %v", err)
		}

		// 尝试上传到 R2
		r2URL, err := s.r2Service.UploadFromURL(imageURL, req.Prompt, translatedPrompt)
		if err != nil {
			log.Printf("R2 上传失败: %v，使用本地存储", err)
			finalURL = imageMetadata.LocalURL
		} else {
			finalURL = r2URL
			log.Printf("图像已上传到 R2: %s", r2URL)
		}

		// 保存记录到D1数据库（使用真实的图片元数据）
		if s.d1Service != nil {
			record := &models.ImageRecord{
				OriginalPrompt: req.Prompt,
				EnglishPrompt:  translatedPrompt,
				LocalPath:      imageMetadata.LocalPath,
				R2URL:          r2URL,
				FileSize:       imageMetadata.FileSize,
				Width:          imageMetadata.Width,
				Height:         imageMetadata.Height,
				Format:         imageMetadata.Format,
			}

			if err := s.d1Service.SaveImageRecord(record); err != nil {
				log.Printf("警告: 保存图片记录到D1失败: %v", err)
			}
		}
	} else {
		// 使用本地存储
		compressionConfig := &utils.ImageCompressionConfig{
			MaxWidth:     s.config.ImageMaxWidth(),
			MaxHeight:    s.config.ImageMaxHeight(),
			Quality:      s.config.ImageQuality(),
			Format:       s.config.ImageFormat(),
			EnableResize: s.config.ImageEnableResize(),
		}
		imageMetadata, err := utils.DownloadAndSaveImage(imageURL, req.Prompt, translatedPrompt, s.config.ImagesDir(), compressionConfig)
		if err != nil {
			return nil, fmt.Errorf("保存图像失败: %v", err)
		}
		finalURL = imageMetadata.LocalURL
		log.Printf("图像已保存到本地: %s", finalURL)

		// 保存记录到D1数据库（使用真实的图片元数据）
		if s.d1Service != nil {
			record := &models.ImageRecord{
				OriginalPrompt: req.Prompt,
				EnglishPrompt:  translatedPrompt,
				LocalPath:      imageMetadata.LocalPath,
				R2URL:          "",
				FileSize:       imageMetadata.FileSize,
				Width:          imageMetadata.Width,
				Height:         imageMetadata.Height,
				Format:         imageMetadata.Format,
			}

			if err := s.d1Service.SaveImageRecord(record); err != nil {
				log.Printf("警告: 保存图片记录到D1失败: %v", err)
			}
		}
	}

	// 返回响应
	response := &models.ImageGenerationResponse{
		Created: time.Now().Unix(),
		Data: []struct {
			URL     string `json:"url,omitempty"`
			B64JSON string `json:"b64_json,omitempty"`
		}{
			{
				URL: finalURL,
			},
		},
	}

	return response, nil
}

// TranslateText 翻译文本
func (s *ImageService) TranslateText(text string) (*models.TranslationResponse, error) {
	translatedText, err := s.translatePrompt(text)
	if err != nil {
		return &models.TranslationResponse{
			TranslatedText: "",
			Success:        false,
			Error:          err.Error(),
		}, nil
	}

	return &models.TranslationResponse{
		TranslatedText: translatedText,
		Success:        true,
	}, nil
}

// GetImageList 获取图片列表
func (s *ImageService) GetImageList() (*models.ImageListResponse, error) {
	images, err := utils.GetImageList(s.config.ImagesDir())
	if err != nil {
		return nil, fmt.Errorf("获取图片列表失败: %v", err)
	}

	return &models.ImageListResponse{
		Images: images,
	}, nil
}

// GetImageRecords 获取图片记录列表
func (s *ImageService) GetImageRecords(req *models.ImageRecordRequest) (*models.ImageRecordResponse, error) {
	if s.d1Service == nil {
		return nil, fmt.Errorf("D1服务未初始化")
	}

	return s.d1Service.GetImageRecords(req)
}

// GetImageRecordByID 根据ID获取图片记录
func (s *ImageService) GetImageRecordByID(id int) (*models.ImageRecord, error) {
	if s.d1Service == nil {
		return nil, fmt.Errorf("D1服务未初始化")
	}

	return s.d1Service.GetImageRecordByID(id)
}

// translatePrompt 翻译提示词
func (s *ImageService) translatePrompt(prompt string) (string, error) {
	// 检查是否为英文
	if isEnglish(prompt) {
		return prompt, nil
	}

	// 使用ModelScope翻译
	translatedPrompt, err := s.translateWithModelScope(prompt)
	if err != nil {
		return "", fmt.Errorf("ModelScope翻译失败: %v", err)
	}

	return translatedPrompt, nil
}

// translateWithModelScope 使用ModelScope翻译
func (s *ImageService) translateWithModelScope(text string) (string, error) {
	// 构建翻译提示
	translationPrompt := fmt.Sprintf(`请将以下中文文本翻译成英文，只返回翻译结果，不要添加任何解释：

%s`, text)

	// 创建OpenAI客户端，配置ModelScope的base URL
	config := openai.DefaultConfig(s.config.ModelScopeToken())
	config.BaseURL = "https://api-inference.modelscope.cn/v1"
	client := openai.NewClientWithConfig(config)

	// 创建聊天完成请求
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: s.config.ModelScopeModel(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: translationPrompt,
				},
			},
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	)
	if err != nil {
		return "", fmt.Errorf("ModelScope API调用失败: %v", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("翻译结果为空")
	}

	// 清理翻译结果
	translatedText := strings.TrimSpace(resp.Choices[0].Message.Content)
	translatedText = strings.Trim(translatedText, "\"")
	translatedText = strings.Trim(translatedText, "'")

	return translatedText, nil
}

// generateImageWithOpenRouter 使用OpenRouter生成图像
func (s *ImageService) generateImageWithOpenRouter(prompt, model string) (string, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
		"max_tokens": 1000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送请求
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.OpenRouterAPIKey())
	req.Header.Set("HTTP-Referer", "http://localhost:8000")
	req.Header.Set("X-Title", "AI Image Generator")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// 限制响应体日志长度，避免打印大量图片数据
		responsePreview := string(body)
		if len(responsePreview) > 500 {
			responsePreview = responsePreview[:500] + "... (响应内容被截断)"
		}
		return "", fmt.Errorf("OpenRouter API请求失败，状态码: %d, 响应: %s", resp.StatusCode, responsePreview)
	}

	// 解析响应
	var imageResp models.OpenRouterImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&imageResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(imageResp.Choices) == 0 {
		return "", fmt.Errorf("OpenRouter响应中没有choices")
	}

	// 提取图像URL
	var imageURL string

	// 先从 images 字段尝试提取 URL
	if len(imageResp.Choices[0].Message.Images) > 0 {
		imageURL = imageResp.Choices[0].Message.Images[0].ImageURL.URL
		log.Printf("从 images 字段提取到 URL: %s", imageURL)
	}

	// 若未取到，再从 content 数组中尝试提取 image_url 或 base64 data
	if imageURL == "" {
		var contentArray []models.OpenRouterContentItem
		if err := json.Unmarshal(imageResp.Choices[0].Message.Content, &contentArray); err == nil {
			for _, item := range contentArray {
				if item.Type == "output_image" {
					if item.ImageURL != nil && item.ImageURL.URL != "" {
						imageURL = item.ImageURL.URL
						log.Printf("从 content 数组提取到 image_url: %s", imageURL)
						break
					}
					if item.Data != nil && *item.Data != "" {
						mimeType := item.MimeType
						if mimeType == "" {
							mimeType = "png"
						}
						// 构建 data URL
						imageURL = fmt.Sprintf("data:image/%s;base64,%s", mimeType, *item.Data)
						log.Printf("从 content 数组提取到 base64 data, 大小: %d 字节", len(*item.Data))
						break
					}
				}
			}
		} else {
			// 尝试作为字符串解析
			var contentStr string
			if err := json.Unmarshal(imageResp.Choices[0].Message.Content, &contentStr); err == nil {
				// 在文本中查找 image_url 或 base64
				if strings.Contains(contentStr, "image_url") {
					// 提取 image_url
					urlPattern := regexp.MustCompile(`"image_url":\s*"([^"]+)"`)
					matches := urlPattern.FindStringSubmatch(contentStr)
					if len(matches) >= 2 {
						imageURL = matches[1]
						log.Printf("从 content 字符串提取到 image_url: %s", imageURL)
					}
				} else if strings.Contains(contentStr, "data:image/") {
					// 提取 base64 data URL
					dataPattern := regexp.MustCompile(`data:image/[^;]+;base64,[A-Za-z0-9+/=]+`)
					matches := dataPattern.FindString(contentStr)
					if matches != "" {
						imageURL = matches
						log.Printf("从 content 字符串提取到 base64 data URL, 大小: %d 字节", len(matches))
					}
				}
			}
		}
	}

	if imageURL == "" {
		return "", fmt.Errorf("未能从OpenRouter响应中提取到图像URL")
	}

	return imageURL, nil
}

// isEnglish 检查文本是否为英文
func isEnglish(text string) bool {
	// 简单的英文检测：检查是否包含中文字符
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return false
		}
	}
	return true
}

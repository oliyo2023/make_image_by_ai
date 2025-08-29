package models

import (
	"encoding/json"
)

// ImageGenerationRequest 图像生成请求
type ImageGenerationRequest struct {
	Prompt         string `json:"prompt" binding:"required"`
	Model          string `json:"model,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Style          string `json:"style,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

// ImageGenerationResponse 图像生成响应
type ImageGenerationResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL     string `json:"url,omitempty"`
		B64JSON string `json:"b64_json,omitempty"`
	} `json:"data"`
}

// TranslationRequest 翻译请求
type TranslationRequest struct {
	Text string `json:"text" binding:"required"`
}

// TranslationResponse 翻译响应
type TranslationResponse struct {
	TranslatedText string `json:"translated_text"`
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
}

// ImageListResponse 图片列表响应
type ImageListResponse struct {
	Images []ImageInfo `json:"images"`
}

// ImageInfo 图片信息
type ImageInfo struct {
	Filename    string `json:"filename"`
	URL         string `json:"url"`
	CreatedTime string `json:"created_time"`
	Size        int64  `json:"size"`
}

// OpenRouterImageResponse OpenRouter图像响应
type OpenRouterImageResponse struct {
	Choices []struct {
		Message struct {
			Content json.RawMessage `json:"content"`
			Images  []struct {
				ImageURL struct {
					URL string `json:"url"`
				} `json:"image_url"`
			} `json:"images,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenRouterContentItem OpenRouter内容项
type OpenRouterContentItem struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
	Data     *string `json:"data,omitempty"`
	MimeType string  `json:"mime_type,omitempty"`
}

// ModelScopeResponse ModelScope响应
type ModelScopeResponse struct {
	Response string `json:"response"`
	History  []struct {
		Human string `json:"human"`
		Ass   string `json:"ass"`
	} `json:"history"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// ImageRecord 图片记录模型
type ImageRecord struct {
	ID             int    `json:"id" db:"id"`
	OriginalPrompt string `json:"original_prompt" db:"original_prompt"`
	EnglishPrompt  string `json:"english_prompt" db:"english_prompt"`
	LocalPath      string `json:"local_path" db:"local_path"`
	R2URL          string `json:"r2_url" db:"r2_url"`
	FileSize       int64  `json:"file_size" db:"file_size"`
	Width          int    `json:"width" db:"width"`
	Height         int    `json:"height" db:"height"`
	Format         string `json:"format" db:"format"`
	CreatedAt      string `json:"created_at" db:"created_at"`
	UpdatedAt      string `json:"updated_at" db:"updated_at"`
}

// ImageRecordRequest 图片记录查询请求
type ImageRecordRequest struct {
	Page     int    `json:"page" form:"page"`
	Limit    int    `json:"limit" form:"limit"`
	Keyword  string `json:"keyword" form:"keyword"`
	DateFrom string `json:"date_from" form:"date_from"`
	DateTo   string `json:"date_to" form:"date_to"`
}

// ImageRecordResponse 图片记录查询响应
type ImageRecordResponse struct {
	Records []ImageRecord `json:"records"`
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	Limit   int           `json:"limit"`
	Pages   int           `json:"pages"`
}

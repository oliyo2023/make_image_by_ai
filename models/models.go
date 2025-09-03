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

// ===== 图片编辑相关模型 =====

// EditType 编辑类型枚举
type EditType string

const (
	EditTypeEdit    EditType = "edit"    // 图片编辑
	EditTypeCompose EditType = "compose" // 图片合成
	EditTypeStyle   EditType = "style"   // 风格转换
	EditTypeFusion  EditType = "fusion"  // 图片融合
)

// EditStatus 编辑状态枚举
type EditStatus string

const (
	EditStatusPending    EditStatus = "pending"    // 等待处理
	EditStatusProcessing EditStatus = "processing" // 处理中
	EditStatusCompleted  EditStatus = "completed"  // 已完成
	EditStatusFailed     EditStatus = "failed"     // 失败
)

// ImageEditRecord 图片编辑记录
type ImageEditRecord struct {
	ID              int        `json:"id" db:"id"`
	OriginalImageID *int       `json:"original_image_id" db:"original_image_id"`
	EditType        EditType   `json:"edit_type" db:"edit_type"`
	EditPrompt      string     `json:"edit_prompt" db:"edit_prompt"`
	InputImageURLs  string     `json:"input_image_urls" db:"input_image_urls"` // JSON array
	ResultImageURL  string     `json:"result_image_url" db:"result_image_url"`
	LocalPath       string     `json:"local_path" db:"local_path"`
	R2URL           string     `json:"r2_url" db:"r2_url"`
	Status          EditStatus `json:"status" db:"status"`
	ErrorMessage    string     `json:"error_message" db:"error_message"`
	TaskID          string     `json:"task_id" db:"task_id"` // ModelScope 任务ID
	FileSize        int64      `json:"file_size" db:"file_size"`
	Width           int        `json:"width" db:"width"`
	Height          int        `json:"height" db:"height"`
	Format          string     `json:"format" db:"format"`
	CreatedAt       string     `json:"created_at" db:"created_at"`
	CompletedAt     string     `json:"completed_at" db:"completed_at"`
}

// ImageEditRequest 图片编辑请求
type ImageEditRequest struct {
	ImageURL   string `json:"image_url" binding:"required"`
	EditPrompt string `json:"edit_prompt" binding:"required"`
	ImageID    *int   `json:"image_id,omitempty"` // 可选，使用历史图片ID
}

// ImageComposeRequest 图片合成请求
type ImageComposeRequest struct {
	ImageURLs     []string `json:"image_urls" binding:"required,min=2"`
	ComposePrompt string   `json:"compose_prompt" binding:"required"`
	ImageIDs      []int    `json:"image_ids,omitempty"` // 可选，使用历史图片ID
}

// StyleTransferRequest 风格转换请求
type StyleTransferRequest struct {
	ImageURL    string `json:"image_url" binding:"required"`
	StylePrompt string `json:"style_prompt" binding:"required"`
	StyleType   string `json:"style_type,omitempty"` // 预定义风格类型
	ImageID     *int   `json:"image_id,omitempty"`   // 可选，使用历史图片ID
}

// ImageFusionRequest 图片融合请求
type ImageFusionRequest struct {
	Image1URL    string  `json:"image1_url" binding:"required"`
	Image2URL    string  `json:"image2_url" binding:"required"`
	FusionPrompt string  `json:"fusion_prompt" binding:"required"`
	FusionRatio  float32 `json:"fusion_ratio,omitempty"` // 融合比例 0.0-1.0
	Image1ID     *int    `json:"image1_id,omitempty"`    // 可选，使用历史图片ID
	Image2ID     *int    `json:"image2_id,omitempty"`    // 可选，使用历史图片ID
}

// EditTaskResponse 编辑任务响应
type EditTaskResponse struct {
	TaskID    string     `json:"task_id"`
	Status    EditStatus `json:"status"`
	EditType  EditType   `json:"edit_type"`
	CreatedAt string     `json:"created_at"`
	Message   string     `json:"message,omitempty"`
}

// EditTaskStatusResponse 编辑任务状态查询响应
type EditTaskStatusResponse struct {
	ID             int        `json:"id"`
	TaskID         string     `json:"task_id"`
	Status         EditStatus `json:"status"`
	EditType       EditType   `json:"edit_type"`
	EditPrompt     string     `json:"edit_prompt"`
	ResultImageURL string     `json:"result_image_url,omitempty"`
	R2URL          string     `json:"r2_url,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	Progress       int        `json:"progress"` // 进度百分比
	CreatedAt      string     `json:"created_at"`
	CompletedAt    string     `json:"completed_at,omitempty"`
}

// ImageUploadRequest 图片上传请求
type ImageUploadRequest struct {
	Base64Data string `json:"base64_data" binding:"required"`
	FileName   string `json:"file_name" binding:"required"`
	MimeType   string `json:"mime_type" binding:"required"`
}

// ImageUploadResponse 图片上传响应
type ImageUploadResponse struct {
	ImageURL  string `json:"image_url"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	CreatedAt string `json:"created_at"`
}

// EditRecordsRequest 编辑记录查询请求
type EditRecordsRequest struct {
	Page     int        `json:"page" form:"page"`
	Limit    int        `json:"limit" form:"limit"`
	EditType EditType   `json:"edit_type" form:"edit_type"`
	Status   EditStatus `json:"status" form:"status"`
	Keyword  string     `json:"keyword" form:"keyword"`
	DateFrom string     `json:"date_from" form:"date_from"`
	DateTo   string     `json:"date_to" form:"date_to"`
}

// EditRecordsResponse 编辑记录查询响应
type EditRecordsResponse struct {
	Records []ImageEditRecord `json:"records"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	Limit   int               `json:"limit"`
	Pages   int               `json:"pages"`
}

// ModelScopeEditResponse ModelScope图片编辑响应
type ModelScopeEditResponse struct {
	TaskID string `json:"task_id"`
}

// ModelScopeTaskStatusResponse ModelScope任务状态响应
type ModelScopeTaskStatusResponse struct {
	TaskStatus   string   `json:"task_status"` // PENDING, RUNNING, SUCCEED, FAILED
	OutputImages []string `json:"output_images,omitempty"`
	Message      string   `json:"message,omitempty"`
}

// PresetStyle 预设风格
type PresetStyle struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
	Example     string `json:"example,omitempty"` // 示例图片URL
}

// StyleCatalogResponse 风格目录响应
type StyleCatalogResponse struct {
	Styles []PresetStyle `json:"styles"`
}

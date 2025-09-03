package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"make_image_by_ai/config"
	"make_image_by_ai/models"
	"make_image_by_ai/utils"
)

// ImageEditService 图片编辑服务
type ImageEditService struct {
	config    *config.Config
	r2Service *R2Service
	d1Service *D1Service
	client    *http.Client
}

// NewImageEditService 创建图片编辑服务实例
func NewImageEditService(cfg *config.Config, r2Service *R2Service, d1Service *D1Service) *ImageEditService {
	return &ImageEditService{
		config:    cfg,
		r2Service: r2Service,
		d1Service: d1Service,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// EditImage 编辑图片
func (s *ImageEditService) EditImage(req *models.ImageEditRequest) (*models.EditTaskResponse, error) {
	// 创建编辑任务记录
	editRecord := &models.ImageEditRecord{
		OriginalImageID: req.ImageID,
		EditType:        models.EditTypeEdit,
		EditPrompt:      req.EditPrompt,
		InputImageURLs:  fmt.Sprintf(`["%s"]`, req.ImageURL),
		Status:          models.EditStatusPending,
		CreatedAt:       time.Now().Format("2006-01-02 15:04:05"),
	}

	// 保存任务记录到数据库
	if s.d1Service != nil {
		if err := s.d1Service.SaveEditRecord(editRecord); err != nil {
			log.Printf("保存编辑记录失败: %v", err)
		}
	}

	// 异步处理编辑任务
	go s.processEditTask(editRecord, req.ImageURL, req.EditPrompt)

	return &models.EditTaskResponse{
		TaskID:    editRecord.TaskID,
		Status:    models.EditStatusPending,
		EditType:  models.EditTypeEdit,
		CreatedAt: editRecord.CreatedAt,
		Message:   "图片编辑任务已提交，正在处理中...",
	}, nil
}

// ComposeImages 合成图片
func (s *ImageEditService) ComposeImages(req *models.ImageComposeRequest) (*models.EditTaskResponse, error) {
	// 将图片URLs转为JSON字符串
	imageURLsJSON, _ := json.Marshal(req.ImageURLs)

	editRecord := &models.ImageEditRecord{
		EditType:       models.EditTypeCompose,
		EditPrompt:     req.ComposePrompt,
		InputImageURLs: string(imageURLsJSON),
		Status:         models.EditStatusPending,
		CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
	}

	if s.d1Service != nil {
		if err := s.d1Service.SaveEditRecord(editRecord); err != nil {
			log.Printf("保存合成记录失败: %v", err)
		}
	}

	// 异步处理合成任务
	go s.processComposeTask(editRecord, req.ImageURLs, req.ComposePrompt)

	return &models.EditTaskResponse{
		TaskID:    editRecord.TaskID,
		Status:    models.EditStatusPending,
		EditType:  models.EditTypeCompose,
		CreatedAt: editRecord.CreatedAt,
		Message:   "图片合成任务已提交，正在处理中...",
	}, nil
}

// StyleTransfer 风格转换
func (s *ImageEditService) StyleTransfer(req *models.StyleTransferRequest) (*models.EditTaskResponse, error) {
	editRecord := &models.ImageEditRecord{
		OriginalImageID: req.ImageID,
		EditType:        models.EditTypeStyle,
		EditPrompt:      req.StylePrompt,
		InputImageURLs:  fmt.Sprintf(`["%s"]`, req.ImageURL),
		Status:          models.EditStatusPending,
		CreatedAt:       time.Now().Format("2006-01-02 15:04:05"),
	}

	if s.d1Service != nil {
		if err := s.d1Service.SaveEditRecord(editRecord); err != nil {
			log.Printf("保存风格转换记录失败: %v", err)
		}
	}

	go s.processStyleTransferTask(editRecord, req.ImageURL, req.StylePrompt)

	return &models.EditTaskResponse{
		TaskID:    editRecord.TaskID,
		Status:    models.EditStatusPending,
		EditType:  models.EditTypeStyle,
		CreatedAt: editRecord.CreatedAt,
		Message:   "风格转换任务已提交，正在处理中...",
	}, nil
}

// FusionImages 图片融合
func (s *ImageEditService) FusionImages(req *models.ImageFusionRequest) (*models.EditTaskResponse, error) {
	imageURLs := []string{req.Image1URL, req.Image2URL}
	imageURLsJSON, _ := json.Marshal(imageURLs)

	editRecord := &models.ImageEditRecord{
		EditType:       models.EditTypeFusion,
		EditPrompt:     req.FusionPrompt,
		InputImageURLs: string(imageURLsJSON),
		Status:         models.EditStatusPending,
		CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
	}

	if s.d1Service != nil {
		if err := s.d1Service.SaveEditRecord(editRecord); err != nil {
			log.Printf("保存融合记录失败: %v", err)
		}
	}

	go s.processFusionTask(editRecord, req.Image1URL, req.Image2URL, req.FusionPrompt)

	return &models.EditTaskResponse{
		TaskID:    editRecord.TaskID,
		Status:    models.EditStatusPending,
		EditType:  models.EditTypeFusion,
		CreatedAt: editRecord.CreatedAt,
		Message:   "图片融合任务已提交，正在处理中...",
	}, nil
}

// processEditTask 处理图片编辑任务
func (s *ImageEditService) processEditTask(record *models.ImageEditRecord, imageURL, editPrompt string) {
	s.updateTaskStatus(record, models.EditStatusProcessing, "")

	// 调用 ModelScope 图片编辑API
	taskID, err := s.callModelScopeImageEdit(imageURL, editPrompt)
	if err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("启动编辑任务失败: %v", err))
		return
	}

	record.TaskID = taskID
	s.updateEditRecord(record)

	// 轮询任务状态
	resultURL, err := s.pollTaskResult(taskID)
	if err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("获取编辑结果失败: %v", err))
		return
	}

	// 保存结果图片
	if err := s.saveResultImage(record, resultURL); err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("保存结果图片失败: %v", err))
		return
	}

	s.updateTaskStatus(record, models.EditStatusCompleted, "")
}

// processComposeTask 处理图片合成任务
func (s *ImageEditService) processComposeTask(record *models.ImageEditRecord, imageURLs []string, composePrompt string) {
	s.updateTaskStatus(record, models.EditStatusProcessing, "")

	// 使用第一张图片作为基础，将合成描述作为编辑指令
	baseImageURL := imageURLs[0]
	fullPrompt := fmt.Sprintf("Compose this image with other elements: %s", composePrompt)

	taskID, err := s.callModelScopeImageEdit(baseImageURL, fullPrompt)
	if err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("启动合成任务失败: %v", err))
		return
	}

	record.TaskID = taskID
	s.updateEditRecord(record)

	resultURL, err := s.pollTaskResult(taskID)
	if err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("获取合成结果失败: %v", err))
		return
	}

	if err := s.saveResultImage(record, resultURL); err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("保存合成图片失败: %v", err))
		return
	}

	s.updateTaskStatus(record, models.EditStatusCompleted, "")
}

// processStyleTransferTask 处理风格转换任务
func (s *ImageEditService) processStyleTransferTask(record *models.ImageEditRecord, imageURL, stylePrompt string) {
	s.updateTaskStatus(record, models.EditStatusProcessing, "")

	// 构建风格转换提示词
	fullPrompt := fmt.Sprintf("Apply style: %s", stylePrompt)

	taskID, err := s.callModelScopeImageEdit(imageURL, fullPrompt)
	if err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("启动风格转换任务失败: %v", err))
		return
	}

	record.TaskID = taskID
	s.updateEditRecord(record)

	resultURL, err := s.pollTaskResult(taskID)
	if err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("获取风格转换结果失败: %v", err))
		return
	}

	if err := s.saveResultImage(record, resultURL); err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("保存风格转换图片失败: %v", err))
		return
	}

	s.updateTaskStatus(record, models.EditStatusCompleted, "")
}

// processFusionTask 处理图片融合任务
func (s *ImageEditService) processFusionTask(record *models.ImageEditRecord, image1URL, image2URL, fusionPrompt string) {
	s.updateTaskStatus(record, models.EditStatusProcessing, "")

	// 使用第一张图片作为基础，融合描述作为编辑指令
	fullPrompt := fmt.Sprintf("Fuse with another image: %s", fusionPrompt)

	taskID, err := s.callModelScopeImageEdit(image1URL, fullPrompt)
	if err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("启动融合任务失败: %v", err))
		return
	}

	record.TaskID = taskID
	s.updateEditRecord(record)

	resultURL, err := s.pollTaskResult(taskID)
	if err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("获取融合结果失败: %v", err))
		return
	}

	if err := s.saveResultImage(record, resultURL); err != nil {
		s.updateTaskStatus(record, models.EditStatusFailed, fmt.Sprintf("保存融合图片失败: %v", err))
		return
	}

	s.updateTaskStatus(record, models.EditStatusCompleted, "")
}

// callModelScopeImageEdit 调用 ModelScope 图片编辑 API
func (s *ImageEditService) callModelScopeImageEdit(imageURL, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model":     "Qwen/Qwen-Image-Edit",
		"prompt":    prompt,
		"image_url": imageURL,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api-inference.modelscope.cn/v1/images/generations", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.ModelScopeToken())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ModelScope-Async-Mode", "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ModelScope API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response models.ModelScopeEditResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	return response.TaskID, nil
}

// pollTaskResult 轮询任务结果
func (s *ImageEditService) pollTaskResult(taskID string) (string, error) {
	maxAttempts := 30 // 最多轮询30次，每次5秒，总计150秒

	for attempt := 0; attempt < maxAttempts; attempt++ {
		status, resultURL, err := s.getTaskStatus(taskID)
		if err != nil {
			log.Printf("查询任务状态失败: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		switch status {
		case "SUCCEED":
			if resultURL != "" {
				return resultURL, nil
			}
			return "", fmt.Errorf("任务完成但未获取到结果图片")
		case "FAILED":
			return "", fmt.Errorf("ModelScope任务处理失败")
		case "PENDING", "RUNNING":
			log.Printf("任务 %s 状态: %s, 继续等待...", taskID, status)
			time.Sleep(5 * time.Second)
		default:
			log.Printf("未知任务状态: %s", status)
			time.Sleep(5 * time.Second)
		}
	}

	return "", fmt.Errorf("任务超时，处理时间过长")
}

// getTaskStatus 获取任务状态
func (s *ImageEditService) getTaskStatus(taskID string) (string, string, error) {
	url := fmt.Sprintf("https://api-inference.modelscope.cn/v1/tasks/%s", taskID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.ModelScopeToken())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ModelScope-Task-Type", "image_generation")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("ModelScope API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response models.ModelScopeTaskStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", fmt.Errorf("解析响应失败: %v", err)
	}

	var resultURL string
	if len(response.OutputImages) > 0 {
		resultURL = response.OutputImages[0]
	}

	return response.TaskStatus, resultURL, nil
}

// saveResultImage 保存结果图片
func (s *ImageEditService) saveResultImage(record *models.ImageEditRecord, imageURL string) error {
	var finalURL string
	var imageMetadata *utils.ImageMetadata

	// 优先保存到 R2，失败时回退到本地
	if s.r2Service != nil {
		r2URL, err := s.r2Service.UploadFromURL(imageURL, record.EditPrompt, record.EditPrompt)
		if err != nil {
			log.Printf("R2 上传失败: %v，尝试本地保存", err)
			// 回退到本地保存
			compressionConfig := &utils.ImageCompressionConfig{
				MaxWidth:     s.config.ImageMaxWidth(),
				MaxHeight:    s.config.ImageMaxHeight(),
				Quality:      s.config.ImageQuality(),
				Format:       s.config.ImageFormat(),
				EnableResize: s.config.ImageEnableResize(),
			}
			imageMetadata, err = utils.DownloadAndSaveImage(imageURL, record.EditPrompt, record.EditPrompt, s.config.ImagesDir(), compressionConfig)
			if err != nil {
				return fmt.Errorf("本地保存也失败: %v", err)
			}
			finalURL = imageMetadata.LocalURL
		} else {
			finalURL = r2URL
			record.R2URL = r2URL
		}
	} else {
		// 只使用本地保存
		compressionConfig := &utils.ImageCompressionConfig{
			MaxWidth:     s.config.ImageMaxWidth(),
			MaxHeight:    s.config.ImageMaxHeight(),
			Quality:      s.config.ImageQuality(),
			Format:       s.config.ImageFormat(),
			EnableResize: s.config.ImageEnableResize(),
		}
		imageMetadata, err := utils.DownloadAndSaveImage(imageURL, record.EditPrompt, record.EditPrompt, s.config.ImagesDir(), compressionConfig)
		if err != nil {
			return fmt.Errorf("本地保存失败: %v", err)
		}
		finalURL = imageMetadata.LocalURL
	}

	// 更新记录
	record.ResultImageURL = finalURL
	if imageMetadata != nil {
		record.LocalPath = imageMetadata.LocalPath
		record.FileSize = imageMetadata.FileSize
		record.Width = imageMetadata.Width
		record.Height = imageMetadata.Height
		record.Format = imageMetadata.Format
	}
	record.CompletedAt = time.Now().Format("2006-01-02 15:04:05")

	return nil
}

// updateTaskStatus 更新任务状态
func (s *ImageEditService) updateTaskStatus(record *models.ImageEditRecord, status models.EditStatus, errorMessage string) {
	record.Status = status
	if errorMessage != "" {
		record.ErrorMessage = errorMessage
	}
	if status == models.EditStatusCompleted {
		record.CompletedAt = time.Now().Format("2006-01-02 15:04:05")
	}

	s.updateEditRecord(record)
}

// updateEditRecord 更新编辑记录
func (s *ImageEditService) updateEditRecord(record *models.ImageEditRecord) {
	if s.d1Service != nil {
		if err := s.d1Service.UpdateEditRecord(record); err != nil {
			log.Printf("更新编辑记录失败: %v", err)
		}
	}
}

// GetTaskStatus 获取编辑任务状态
func (s *ImageEditService) GetTaskStatus(taskID string) (*models.EditTaskStatusResponse, error) {
	if s.d1Service == nil {
		return nil, fmt.Errorf("D1服务未初始化")
	}

	return s.d1Service.GetEditRecordByTaskID(taskID)
}

// GetEditRecords 获取编辑记录列表
func (s *ImageEditService) GetEditRecords(req *models.EditRecordsRequest) (*models.EditRecordsResponse, error) {
	if s.d1Service == nil {
		return nil, fmt.Errorf("D1服务未初始化")
	}

	return s.d1Service.GetEditRecords(req)
}

// GetPresetStyles 获取预设风格列表
func (s *ImageEditService) GetPresetStyles() (*models.StyleCatalogResponse, error) {
	// 返回预定义的风格列表
	styles := []models.PresetStyle{
		{
			ID:          "anime",
			Name:        "动漫风格",
			Description: "将图片转换为动漫/卡通风格",
			Prompt:      "convert to anime style, cartoon style, vibrant colors",
		},
		{
			ID:          "oil_painting",
			Name:        "油画风格",
			Description: "将图片转换为古典油画风格",
			Prompt:      "convert to oil painting style, classical painting, artistic brushstrokes",
		},
		{
			ID:          "watercolor",
			Name:        "水彩画风格",
			Description: "将图片转换为水彩画风格",
			Prompt:      "convert to watercolor painting style, soft colors, artistic",
		},
		{
			ID:          "sketch",
			Name:        "素描风格",
			Description: "将图片转换为铅笔素描风格",
			Prompt:      "convert to pencil sketch style, black and white drawing",
		},
		{
			ID:          "cyberpunk",
			Name:        "赛博朋克",
			Description: "将图片转换为赛博朋克未来科技风格",
			Prompt:      "convert to cyberpunk style, neon lights, futuristic, high tech",
		},
		{
			ID:          "vintage",
			Name:        "复古风格",
			Description: "将图片转换为复古怀旧风格",
			Prompt:      "convert to vintage style, retro, old-fashioned, sepia tones",
		},
	}

	return &models.StyleCatalogResponse{
		Styles: styles,
	}, nil
}

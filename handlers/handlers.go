package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"huiying/models"
	"huiying/services"
)

// Handler HTTP处理器
type Handler struct {
	imageService     *services.ImageService
	imageEditService *services.ImageEditService
}

// NewHandler 创建处理器实例
func NewHandler(imageService *services.ImageService, imageEditService *services.ImageEditService) *Handler {
	return &Handler{
		imageService:     imageService,
		imageEditService: imageEditService,
	}
}

// SetupRoutes 设置路由
func (h *Handler) SetupRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", h.HealthCheck)

	// 静态文件服务
	r.Static("/static", "./public/static")

	// API路由
	api := r.Group("/api")
	{
		api.POST("/generate-image", h.GenerateImage)
		api.POST("/translate", h.TranslateText)
		api.GET("/images", h.GetImageList)
		api.GET("/records", h.GetImageRecords)
		api.GET("/records/:id", h.GetImageRecordByID)

		// 图片编辑相关API
		api.POST("/edit-image", h.EditImage)
		api.POST("/compose-images", h.ComposeImages)
		api.POST("/style-transfer", h.StyleTransfer)
		api.POST("/fusion-images", h.FusionImages)
		api.POST("/upload-image", h.UploadImage)
		api.GET("/edit-tasks/:taskId", h.GetEditTaskStatus)
		api.GET("/edit-records", h.GetEditRecords)
		api.GET("/preset-styles", h.GetPresetStyles)
	}

	// 兼容性路由（与app.py保持一致）
	r.POST("/generate-image", h.GenerateImage)
	r.POST("/translate", h.TranslateText)
	r.GET("/images", h.GetImageList)
	r.GET("/records", h.GetImageRecords)
	r.GET("/records/:id", h.GetImageRecordByID)

	// 图片编辑兼容性路由
	r.POST("/edit-image", h.EditImage)
	r.POST("/compose-images", h.ComposeImages)
	r.POST("/style-transfer", h.StyleTransfer)
	r.POST("/fusion-images", h.FusionImages)
	r.POST("/upload-image", h.UploadImage)
	r.GET("/edit-tasks/:taskId", h.GetEditTaskStatus)
	r.GET("/edit-records", h.GetEditRecords)
	r.GET("/preset-styles", h.GetPresetStyles)
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	response := models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Version:   "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

// GenerateImage 生成图像
func (h *Handler) GenerateImage(c *gin.Context) {
	var req models.ImageGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 简化日志输出，避免打印大量内容
	promptPreview := req.Prompt
	if len(promptPreview) > 100 {
		promptPreview = promptPreview[:100] + "..."
	}
	log.Printf("收到图像生成请求: Prompt=%s, Model=%s", promptPreview, req.Model)

	response, err := h.imageService.GenerateImage(&req)
	if err != nil {
		log.Printf("图像生成失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// TranslateText 翻译文本
func (h *Handler) TranslateText(c *gin.Context) {
	var req models.TranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 简化日志输出，避免打印大量内容
	textPreview := req.Text
	if len(textPreview) > 100 {
		textPreview = textPreview[:100] + "..."
	}
	log.Printf("收到翻译请求: %s", textPreview)

	response, err := h.imageService.TranslateText(req.Text)
	if err != nil {
		log.Printf("翻译失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetImageList 获取图片列表
func (h *Handler) GetImageList(c *gin.Context) {
	log.Printf("收到图片列表请求")

	response, err := h.imageService.GetImageList()
	if err != nil {
		log.Printf("获取图片列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetImageRecords 获取图片记录列表
func (h *Handler) GetImageRecords(c *gin.Context) {
	log.Printf("收到图片记录查询请求")

	var req models.ImageRecordRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的查询参数: " + err.Error(),
		})
		return
	}

	response, err := h.imageService.GetImageRecords(&req)
	if err != nil {
		log.Printf("获取图片记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetImageRecordByID 根据ID获取图片记录
func (h *Handler) GetImageRecordByID(c *gin.Context) {
	idParam := c.Param("id")
	log.Printf("收到获取图片记录请求: ID=%s", idParam)

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的ID参数",
		})
		return
	}

	record, err := h.imageService.GetImageRecordByID(id)
	if err != nil {
		log.Printf("获取图片记录失败: %v", err)
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "图片记录不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, record)
}

// ===== 图片编辑相关处理器 =====

// EditImage 编辑图片
func (h *Handler) EditImage(c *gin.Context) {
	if h.imageEditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "图片编辑服务不可用",
		})
		return
	}

	var req models.ImageEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}

	log.Printf("收到图片编辑请求: ImageURL=%s, Prompt=%s", req.ImageURL, req.EditPrompt)

	response, err := h.imageEditService.EditImage(&req)
	if err != nil {
		log.Printf("图片编辑失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ComposeImages 合成图片
func (h *Handler) ComposeImages(c *gin.Context) {
	if h.imageEditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "图片编辑服务不可用",
		})
		return
	}

	var req models.ImageComposeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}

	log.Printf("收到图片合成请求: ImageURLs=%d张, Prompt=%s", len(req.ImageURLs), req.ComposePrompt)

	response, err := h.imageEditService.ComposeImages(&req)
	if err != nil {
		log.Printf("图片合成失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// StyleTransfer 风格转换
func (h *Handler) StyleTransfer(c *gin.Context) {
	if h.imageEditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "图片编辑服务不可用",
		})
		return
	}

	var req models.StyleTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}

	log.Printf("收到风格转换请求: ImageURL=%s, Style=%s", req.ImageURL, req.StylePrompt)

	response, err := h.imageEditService.StyleTransfer(&req)
	if err != nil {
		log.Printf("风格转换失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// FusionImages 图片融合
func (h *Handler) FusionImages(c *gin.Context) {
	if h.imageEditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "图片编辑服务不可用",
		})
		return
	}

	var req models.ImageFusionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}

	log.Printf("收到图片融合请求: Image1=%s, Image2=%s, Prompt=%s", req.Image1URL, req.Image2URL, req.FusionPrompt)

	response, err := h.imageEditService.FusionImages(&req)
	if err != nil {
		log.Printf("图片融合失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UploadImage 上传图片
func (h *Handler) UploadImage(c *gin.Context) {
	var req models.ImageUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}

	log.Printf("收到图片上传请求: FileName=%s, MimeType=%s", req.FileName, req.MimeType)

	// TODO: 实现图片上传逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "图片上传功能尚未实现",
	})
}

// GetEditTaskStatus 获取编辑任务状态
func (h *Handler) GetEditTaskStatus(c *gin.Context) {
	if h.imageEditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "图片编辑服务不可用",
		})
		return
	}

	taskID := c.Param("taskId")
	log.Printf("收到任务状态查询请求: TaskID=%s", taskID)

	response, err := h.imageEditService.GetTaskStatus(taskID)
	if err != nil {
		log.Printf("获取任务状态失败: %v", err)
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "任务不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetEditRecords 获取编辑记录列表
func (h *Handler) GetEditRecords(c *gin.Context) {
	if h.imageEditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "图片编辑服务不可用",
		})
		return
	}

	log.Printf("收到编辑记录查询请求")

	var req models.EditRecordsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的查询参数: " + err.Error(),
		})
		return
	}

	response, err := h.imageEditService.GetEditRecords(&req)
	if err != nil {
		log.Printf("获取编辑记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPresetStyles 获取预设风格列表
func (h *Handler) GetPresetStyles(c *gin.Context) {
	if h.imageEditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "图片编辑服务不可用",
		})
		return
	}

	log.Printf("收到预设风格查询请求")

	response, err := h.imageEditService.GetPresetStyles()
	if err != nil {
		log.Printf("获取预设风格失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

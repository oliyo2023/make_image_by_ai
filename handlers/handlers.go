package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"make_image_by_ai/models"
	"make_image_by_ai/services"
)

// Handler HTTP处理器
type Handler struct {
	imageService *services.ImageService
}

// NewHandler 创建处理器实例
func NewHandler(imageService *services.ImageService) *Handler {
	return &Handler{
		imageService: imageService,
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
	}

	// 兼容性路由（与app.py保持一致）
	r.POST("/generate-image", h.GenerateImage)
	r.POST("/translate", h.TranslateText)
	r.GET("/images", h.GetImageList)
	r.GET("/records", h.GetImageRecords)
	r.GET("/records/:id", h.GetImageRecordByID)
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

	id := 0
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil {
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

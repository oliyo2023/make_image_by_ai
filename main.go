package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"make_image_by_ai/config"
	"make_image_by_ai/handlers"
	"make_image_by_ai/services"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 创建服务实例
	imageService, err := services.NewImageService(cfg)
	if err != nil {
		log.Printf("警告: 图像服务初始化失败: %v", err)
	}

	// 创建图片编辑服务
	var imageEditService *services.ImageEditService
	if imageService != nil {
		// 获取R2和D1服务实例
		r2Service, _ := services.NewR2Service(cfg)
		d1Service, _ := services.NewD1Service(cfg)
		imageEditService = services.NewImageEditService(cfg, r2Service, d1Service)
	}

	// 创建处理器
	handler := handlers.NewHandler(imageService, imageEditService)

	// 设置Gin模式
	gin.SetMode(gin.DebugMode)

	// 创建路由
	r := gin.Default()

	// 设置CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 设置路由
	handler.SetupRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Port())
	log.Printf("服务器启动在端口 %d", cfg.Port())
	log.Printf("健康检查: http://localhost%s/health", addr)
	log.Printf("图像生成: http://localhost%s/generate-image", addr)
	log.Printf("翻译服务: http://localhost%s/translate", addr)
	log.Printf("图片列表: http://localhost%s/images", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

//go:build configtest
// +build configtest

package main

import (
	"fmt"
	"make_image_by_ai/config"
)

func main() {
	fmt.Println("=== 配置测试 ===")

	// 加载配置
	cfg := config.LoadConfig()

	fmt.Printf("服务器端口: %d\n", cfg.Port())
	fmt.Printf("图片目录: %s\n", cfg.ImagesDir())
	fmt.Printf("ModelScope模型: %s\n", cfg.ModelScopeModel())
	fmt.Printf("OpenRouter模型: %s\n", cfg.DefaultOpenRouterModel())
	fmt.Printf("图片最大宽度: %d\n", cfg.ImageMaxWidth())
	fmt.Printf("图片最大高度: %d\n", cfg.ImageMaxHeight())
	fmt.Printf("图片质量: %d\n", cfg.ImageQuality())
	fmt.Printf("图片格式: %s\n", cfg.ImageFormat())
	fmt.Printf("启用图片缩放: %t\n", cfg.ImageEnableResize())

	// 检查R2配置
	if cfg.R2AccountID() != "" {
		fmt.Printf("R2账户ID: %s\n", cfg.R2AccountID())
		fmt.Printf("R2存储桶: %s\n", cfg.R2Bucket())
	} else {
		fmt.Println("R2配置未设置，将使用本地存储")
	}

	fmt.Println("=== 配置测试完成 ===")
}

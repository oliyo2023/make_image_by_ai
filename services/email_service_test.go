package services

import (
	"huiying/config"
	"huiying/models"
	"testing"
)

func TestEmailService(t *testing.T) {
	// 加载配置
	cfg := config.LoadConfig()

	// 创建邮件服务
	emailService := NewEmailService(cfg)

	// 创建测试图像记录
	record := &models.ImageRecord{
		OriginalPrompt: "一只可爱的猫咪",
		EnglishPrompt:  "a cute cat",
		LocalPath:      "test_image.jpg",
		R2URL:          "https://example.com/test_image.jpg",
		FileSize:       102400,
		Width:          1024,
		Height:         768,
		Format:         "jpeg",
	}

	// 测试发送通知邮件（只有在启用SMTP时才会真正发送）
	err := emailService.SendNotification(record)
	if err != nil {
		t.Logf("发送通知邮件时出错: %v", err)
	}

	// 测试发送错误通知邮件（只有在启用SMTP时才会真正发送）
	err = emailService.SendErrorNotification("测试提示词", &EmailServiceError{"测试错误"})
	if err != nil {
		t.Logf("发送错误通知邮件时出错: %v", err)
	}
}

// EmailServiceError 自定义错误类型用于测试
type EmailServiceError struct {
	msg string
}

func (e *EmailServiceError) Error() string {
	return e.msg
}

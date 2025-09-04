package services

import (
	"fmt"
	"log"
	"net/smtp"

	"huiying/config"
	"huiying/models"
)

// EmailService 邮件服务
type EmailService struct {
	config *config.Config
}

// NewEmailService 创建邮件服务实例
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

// SendNotification 发送通知邮件
func (s *EmailService) SendNotification(record *models.ImageRecord) error {
	// 检查是否启用邮件通知
	if !s.config.SMTPEnable() {
		return nil
	}

	// 构建邮件内容
	subject := "AI图像生成完成通知"
	body := fmt.Sprintf(`
<html>
<head>
    <meta charset="UTF-8">
    <title>AI图像生成完成通知</title>
</head>
<body>
    <h2>AI图像生成完成通知</h2>
    <p>您的图像已生成完成：</p>
    <ul>
        <li><strong>原始提示词：</strong>%s</li>
        <li><strong>英文提示词：</strong>%s</li>
        <li><strong>图像尺寸：</strong>%dx%d</li>
        <li><strong>文件大小：</strong>%d 字节</li>
        <li><strong>图像格式：</strong>%s</li>
    </ul>
    %s
    <p>感谢使用我们的AI图像生成服务！</p>
</body>
</html>
`, record.OriginalPrompt, record.EnglishPrompt, record.Width, record.Height, record.FileSize, record.Format, s.getImageLink(record))

	// 发送邮件
	return s.sendEmail(subject, body)
}

// SendErrorNotification 发送错误通知邮件
func (s *EmailService) SendErrorNotification(prompt string, err error) error {
	// 检查是否启用邮件通知
	if !s.config.SMTPEnable() {
		return nil
	}

	// 构建邮件内容
	subject := "AI图像生成错误通知"
	body := fmt.Sprintf(`
<html>
<head>
    <meta charset="UTF-8">
    <title>AI图像生成错误通知</title>
</head>
<body>
    <h2>AI图像生成错误通知</h2>
    <p>图像生成过程中出现错误：</p>
    <ul>
        <li><strong>提示词：</strong>%s</li>
        <li><strong>错误信息：</strong>%s</li>
    </ul>
    <p>请检查服务状态或稍后重试。</p>
</body>
</html>
`, prompt, err.Error())

	// 发送邮件
	return s.sendEmail(subject, body)
}

// getImageLink 获取图像链接
func (s *EmailService) getImageLink(record *models.ImageRecord) string {
	if record.R2URL != "" {
		return fmt.Sprintf(`<p><strong>图像链接：</strong><a href="%s">点击查看图像</a></p>`, record.R2URL)
	} else if record.LocalPath != "" {
		// 如果是本地路径，需要根据实际情况构建访问URL
		// 这里假设有一个固定的访问路径前缀
		return fmt.Sprintf(`<p><strong>图像链接：</strong><a href="/static/images/%s">点击查看图像</a></p>`, record.LocalPath)
	}
	return ""
}

// sendEmail 发送邮件
func (s *EmailService) sendEmail(subject, body string) error {
	// 设置邮件服务器信息
	host := s.config.SMTPHost()
	port := s.config.SMTPPort()
	username := s.config.SMTPUsername()
	password := s.config.SMTPPassword()
	from := s.config.SMTPFrom()
	to := s.config.SMTPTo()

	// 构建邮件内容
	message := fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", to, from, subject, body)

	// 设置认证信息
	auth := smtp.PlainAuth("", username, password, host)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", host, port)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
	if err != nil {
		log.Printf("发送邮件失败: %v", err)
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	log.Printf("邮件已发送至: %s", to)
	return nil
}

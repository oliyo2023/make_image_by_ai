package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"

	"make_image_by_ai/config"
)

// ImageCompressionConfig 图片压缩配置
type ImageCompressionConfig struct {
	MaxWidth     int
	MaxHeight    int
	Quality      int
	Format       string // "jpeg", "png"
	EnableResize bool
}

// R2UploadResult R2上传结果
type R2UploadResult struct {
	URL      string
	FileSize int64
	Width    int
	Height   int
	Format   string
}

// R2Service Cloudflare R2 存储服务
type R2Service struct {
	config            *config.Config
	s3Client          *s3.S3
	compressionConfig ImageCompressionConfig
}

// NewR2Service 创建 R2 服务实例
func NewR2Service(cfg *config.Config) (*R2Service, error) {
	// 检查 R2 配置是否完整
	if cfg.R2AccountID() == "" || cfg.R2AccessKeyID() == "" || cfg.R2AccessKeySecret() == "" ||
		cfg.R2Endpoint() == "" || cfg.R2Bucket() == "" {
		return nil, fmt.Errorf("R2 配置不完整，请检查环境变量")
	}

	// 创建 AWS 会话
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("auto"),
		Credentials: credentials.NewStaticCredentials(
			cfg.R2AccessKeyID(),
			cfg.R2AccessKeySecret(),
			"",
		),
		Endpoint:         aws.String(cfg.R2Endpoint()),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("创建 R2 会话失败: %v", err)
	}

	// 创建 S3 客户端
	s3Client := s3.New(sess)

	// 从配置中读取压缩设置
	compressionConfig := ImageCompressionConfig{
		MaxWidth:     cfg.ImageMaxWidth(),
		MaxHeight:    cfg.ImageMaxHeight(),
		Quality:      cfg.ImageQuality(),
		Format:       cfg.ImageFormat(),
		EnableResize: cfg.ImageEnableResize(),
	}

	return &R2Service{
		config:            cfg,
		s3Client:          s3Client,
		compressionConfig: compressionConfig,
	}, nil
}

// SetCompressionConfig 设置压缩配置
func (r *R2Service) SetCompressionConfig(config ImageCompressionConfig) {
	r.compressionConfig = config
}

// UploadImage 上传图片到 R2
func (r *R2Service) UploadImage(imageData []byte, filename, contentType string) (string, error) {
	// 压缩图片
	compressedData, compressedContentType, err := r.compressImage(imageData, contentType)
	if err != nil {
		log.Printf("图片压缩失败: %v，使用原始图片", err)
		compressedData = imageData
		compressedContentType = contentType
	} else {
		log.Printf("图片压缩成功: 原始大小 %d bytes, 压缩后 %d bytes, 压缩率 %.1f%%",
			len(imageData), len(compressedData),
			float64(len(compressedData))/float64(len(imageData))*100)
	}

	// 生成唯一的文件名
	timestamp := time.Now().Format("20060102_150405")
	uniqueID := strings.ReplaceAll(filename, " ", "_")
	key := fmt.Sprintf("images/%s_%s", timestamp, uniqueID)

	// 创建上传参数
	input := &s3.PutObjectInput{
		Bucket:      aws.String(r.config.R2Bucket()),
		Key:         aws.String(key),
		Body:        bytes.NewReader(compressedData),
		ContentType: aws.String(compressedContentType),
		ACL:         aws.String("public-read"), // 设置为公开读取
	}

	// 上传到 R2
	_, err = r.s3Client.PutObject(input)
	if err != nil {
		return "", fmt.Errorf("上传到 R2 失败: %v", err)
	}

	// 构建公开访问的 URL
	publicURL := fmt.Sprintf("https://%s.%s/%s", r.config.R2Bucket(), strings.TrimPrefix(r.config.R2Endpoint(), "https://"), key)

	log.Printf("图片已上传到 R2: %s", publicURL)
	return publicURL, nil
}

// compressImage 压缩图片
func (r *R2Service) compressImage(imageData []byte, contentType string) ([]byte, string, error) {
	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, "", fmt.Errorf("解码图片失败: %v", err)
	}

	// 调整图片大小
	if r.compressionConfig.EnableResize {
		img = r.resizeImage(img)
	}

	// 编码压缩后的图片
	var buf bytes.Buffer
	var compressedContentType string

	switch strings.ToLower(r.compressionConfig.Format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: r.compressionConfig.Quality})
		compressedContentType = "image/jpeg"
	case "png":
		err = png.Encode(&buf, img)
		compressedContentType = "image/png"
	default:
		// 根据原始格式选择压缩格式
		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: r.compressionConfig.Quality})
			compressedContentType = "image/jpeg"
		case "png":
			err = png.Encode(&buf, img)
			compressedContentType = "image/png"
		default:
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: r.compressionConfig.Quality})
			compressedContentType = "image/jpeg"
		}
	}

	if err != nil {
		return nil, "", fmt.Errorf("编码压缩图片失败: %v", err)
	}

	return buf.Bytes(), compressedContentType, nil
}

// resizeImage 调整图片大小
func (r *R2Service) resizeImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 如果图片尺寸已经小于最大尺寸，不需要调整
	if width <= r.compressionConfig.MaxWidth && height <= r.compressionConfig.MaxHeight {
		return img
	}

	// 计算新的尺寸，保持宽高比
	var newWidth, newHeight uint
	if width > height {
		// 横向图片
		newWidth = uint(r.compressionConfig.MaxWidth)
		newHeight = uint(float64(height) * float64(r.compressionConfig.MaxWidth) / float64(width))
	} else {
		// 纵向图片
		newHeight = uint(r.compressionConfig.MaxHeight)
		newWidth = uint(float64(width) * float64(r.compressionConfig.MaxHeight) / float64(height))
	}

	// 调整大小
	resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	return resizedImg
}

// UploadFromURL 从 URL 下载并上传到 R2
func (r *R2Service) UploadFromURL(imageURL, originalPrompt, englishPrompt string) (string, error) {
	// 下载图片
	imageData, contentType, err := r.downloadImage(imageURL)
	if err != nil {
		return "", fmt.Errorf("下载图片失败: %v", err)
	}

	// 生成文件名，优先使用英文提示词
	var filename string
	if englishPrompt != "" && englishPrompt != originalPrompt {
		filename = r.generateEnglishFilename(englishPrompt, contentType)
	} else {
		filename = r.generateFilename(originalPrompt, contentType)
	}

	// 上传到 R2
	return r.UploadImage(imageData, filename, contentType)
}

// UploadBase64Image 上传 base64 编码的图片到 R2
func (r *R2Service) UploadBase64Image(base64Data, imageFormat, prompt string) (string, error) {
	// 解码 base64 数据
	imageData, err := r.decodeBase64(base64Data)
	if err != nil {
		return "", fmt.Errorf("解码 base64 失败: %v", err)
	}

	// 确定 content type
	contentType := r.getContentType(imageFormat)

	// 生成文件名
	filename := r.generateFilename(prompt, contentType)

	// 上传到 R2
	return r.UploadImage(imageData, filename, contentType)
}

// downloadImage 下载图片
func (r *R2Service) downloadImage(imageURL string) ([]byte, string, error) {
	// 检查是否为 data URL
	if strings.HasPrefix(imageURL, "data:") {
		return r.handleDataURL(imageURL)
	}

	// 下载远程图片
	resp, err := r.httpGet(imageURL)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// 读取响应体
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取响应体失败: %v", err)
	}

	// 获取 content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png" // 默认类型
	}

	return imageData, contentType, nil
}

// handleDataURL 处理 data URL
func (r *R2Service) handleDataURL(dataURL string) ([]byte, string, error) {
	// 解析 data URL
	parts := strings.Split(dataURL, ",")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("无效的 data URL 格式")
	}

	// 解析 content type
	header := parts[0]
	contentType := strings.TrimPrefix(header, "data:")
	contentType = strings.Split(contentType, ";")[0]

	// 解码 base64 数据
	imageData, err := r.decodeBase64(parts[1])
	if err != nil {
		return nil, "", err
	}

	return imageData, contentType, nil
}

// decodeBase64 解码 base64 数据
func (r *R2Service) decodeBase64(base64Data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Data)
}

// httpGet 发送 HTTP GET 请求
func (r *R2Service) httpGet(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	return client.Get(url)
}

// generateFilename 生成文件名
func (r *R2Service) generateFilename(prompt, contentType string) string {
	// 从 prompt 中提取关键词
	keywords := r.extractKeywords(prompt)

	// 从 content type 中提取扩展名
	ext := r.getExtensionFromContentType(contentType)

	return fmt.Sprintf("%s%s", keywords, ext)
}

// generateEnglishFilename 使用英文提示词生成文件名
func (r *R2Service) generateEnglishFilename(englishPrompt, contentType string) string {
	// 从英文提示词中提取关键词
	keywords := r.extractEnglishKeywords(englishPrompt)

	// 从 content type 中提取扩展名
	ext := r.getExtensionFromContentType(contentType)

	return fmt.Sprintf("%s%s", keywords, ext)
}

// extractKeywords 从提示词中提取关键词
func (r *R2Service) extractKeywords(prompt string) string {
	// 移除特殊字符，保留字母、数字和中文
	reg := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}]+`)
	cleanPrompt := reg.ReplaceAllString(prompt, "_")

	// 限制长度
	if len(cleanPrompt) > 50 {
		cleanPrompt = cleanPrompt[:50]
	}

	// 移除首尾下划线
	cleanPrompt = strings.Trim(cleanPrompt, "_")

	if cleanPrompt == "" {
		cleanPrompt = "image"
	}

	return cleanPrompt
}

// extractEnglishKeywords 从英文提示词中提取关键词
func (r *R2Service) extractEnglishKeywords(englishPrompt string) string {
	// 移除标点符号，只保留字母、数字和空格
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	cleanPrompt := reg.ReplaceAllString(englishPrompt, " ")

	// 分割为单词
	words := strings.Fields(cleanPrompt)

	// 过滤掉常见的停用词和短词
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "by": true,
		"for": true, "from": true, "has": true, "he": true, "in": true, "is": true, "it": true,
		"its": true, "of": true, "on": true, "that": true, "the": true, "to": true, "was": true,
		"will": true, "with": true, "or": true, "but": true, "not": true, "this": true, "then": true,
	}

	var keywords []string
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		// 只保留长度大于1且不是停用词的单词
		if len(word) > 1 && !stopWords[word] {
			keywords = append(keywords, word)
			// 最多取前4个有意义的单词
			if len(keywords) >= 4 {
				break
			}
		}
	}

	// 如果没有找到合适的关键词，使用默认值
	if len(keywords) == 0 {
		keywords = []string{"image"}
	}

	// 连接关键词，用下划线分隔
	result := strings.Join(keywords, "_")

	// 限制总长度
	if len(result) > 30 {
		result = result[:30]
	}

	return result
}

// getContentType 根据格式获取 content type
func (r *R2Service) getContentType(format string) string {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "image/png"
	}
}

// getExtensionFromContentType 从 content type 获取扩展名
func (r *R2Service) getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

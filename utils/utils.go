package utils

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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nfnt/resize"

	"huiying/models"
)

// ImageCompressionConfig 图片压缩配置
type ImageCompressionConfig struct {
	MaxWidth     int
	MaxHeight    int
	Quality      int
	Format       string // "jpeg", "png"
	EnableResize bool
}

// ImageMetadata 图片元数据信息
type ImageMetadata struct {
	LocalPath string
	LocalURL  string
	FileSize  int64
	Width     int
	Height    int
	Format    string
}

// DownloadAndSaveImage 下载并保存图片到本地
func DownloadAndSaveImage(imageURL, originalPrompt, englishPrompt, imagesDir string, compressionConfig *ImageCompressionConfig) (*ImageMetadata, error) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic occurred: %v", r)
			log.Printf("错误: %v", err)
		}
	}()

	// 检查是否为 data URL
	if strings.HasPrefix(imageURL, "data:") {
		// 解析 data URL
		base64Pattern := regexp.MustCompile(`data:image/(jpeg|png|webp);base64,([A-Za-z0-9+/=]+)`)
		matches := base64Pattern.FindStringSubmatch(imageURL)

		if len(matches) >= 3 {
			imageFormat := matches[1]
			base64Data := matches[2]
			return SaveBase64Image(base64Data, imageFormat, originalPrompt, englishPrompt, imagesDir, compressionConfig)
		}
		return nil, fmt.Errorf("invalid data URL format")
	}

	// 下载图片
	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image, status: %d", resp.StatusCode)
	}

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %v", err)
	}

	// 获取原始图片尺寸
	originalImg, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image for metadata: %v", err)
	}
	originalBounds := originalImg.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	// 压缩图片
	compressedData, compressedFormat, finalWidth, finalHeight, err := compressImageWithMetadata(imageData, compressionConfig)
	if err != nil {
		log.Printf("图片压缩失败: %v，使用原始图片", err)
		compressedData = imageData
		compressedFormat = "png" // 默认格式
		finalWidth = originalWidth
		finalHeight = originalHeight
	} else {
		log.Printf("图片压缩成功: 原始大小 %d bytes, 压缩后 %d bytes, 压缩率 %.1f%%",
			len(imageData), len(compressedData),
			float64(len(compressedData))/float64(len(imageData))*100)
	}

	// 生成唯一文件名
	timestamp := time.Now().Format("20060102_150405")
	uniqueID := uuid.New().String()[:8]

	// 优先使用英文提示词生成关键词，如果没有则使用原始提示词
	var keywords string
	if englishPrompt != "" && englishPrompt != originalPrompt {
		keywords = ExtractEnglishKeywords(englishPrompt)
	} else {
		keywords = extractKeywords(originalPrompt)
	}
	filename := fmt.Sprintf("ai_image_%s_%s_%s.%s", timestamp, uniqueID, keywords, compressedFormat)

	// 确保目录存在
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	localPath := filepath.Join(imagesDir, filename)
	localURL := fmt.Sprintf("/static/images/%s", filename)

	// 保存压缩后的图片
	err = os.WriteFile(localPath, compressedData, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %v", err)
	}

	log.Printf("图片已保存: %s", localPath)

	// 返回元数据
	return &ImageMetadata{
		LocalPath: localPath,
		LocalURL:  localURL,
		FileSize:  int64(len(compressedData)),
		Width:     finalWidth,
		Height:    finalHeight,
		Format:    compressedFormat,
	}, nil
}

// SaveBase64Image 保存base64编码的图片
func SaveBase64Image(base64Data, imageFormat, originalPrompt, englishPrompt, imagesDir string, compressionConfig *ImageCompressionConfig) (*ImageMetadata, error) {
	// 解码base64数据
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	// 获取原始图片尺寸
	originalImg, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image for metadata: %v", err)
	}
	originalBounds := originalImg.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	// 压缩图片
	compressedData, compressedFormat, finalWidth, finalHeight, err := compressImageWithMetadata(imageData, compressionConfig)
	if err != nil {
		log.Printf("图片压缩失败: %v，使用原始图片", err)
		compressedData = imageData
		compressedFormat = imageFormat
		finalWidth = originalWidth
		finalHeight = originalHeight
	} else {
		log.Printf("图片压缩成功: 原始大小 %d bytes, 压缩后 %d bytes, 压缩率 %.1f%%",
			len(imageData), len(compressedData),
			float64(len(compressedData))/float64(len(imageData))*100)
	}

	// 生成文件名
	timestamp := time.Now().Format("20060102_150405")
	uniqueID := uuid.New().String()[:8]

	// 优先使用英文提示词生成关键词，如果没有则使用原始提示词
	var keywords string
	if englishPrompt != "" && englishPrompt != originalPrompt {
		keywords = ExtractEnglishKeywords(englishPrompt)
	} else {
		keywords = extractKeywords(originalPrompt)
	}
	filename := fmt.Sprintf("ai_image_%s_%s_%s.%s", timestamp, uniqueID, keywords, compressedFormat)

	// 确保目录存在
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	localPath := filepath.Join(imagesDir, filename)
	localURL := fmt.Sprintf("/static/images/%s", filename)

	// 保存压缩后的文件
	err = os.WriteFile(localPath, compressedData, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %v", err)
	}

	log.Printf("Base64图片已保存: %s", localPath)

	// 返回元数据
	return &ImageMetadata{
		LocalPath: localPath,
		LocalURL:  localURL,
		FileSize:  int64(len(compressedData)),
		Width:     finalWidth,
		Height:    finalHeight,
		Format:    compressedFormat,
	}, nil
}

// compressImageWithMetadata 压缩图片并返回元数据
func compressImageWithMetadata(imageData []byte, compressionConfig *ImageCompressionConfig) ([]byte, string, int, int, error) {
	if compressionConfig == nil {
		// 如果没有压缩配置，获取原始尺寸
		img, _, err := image.Decode(bytes.NewReader(imageData))
		if err != nil {
			return imageData, "png", 0, 0, nil
		}
		bounds := img.Bounds()
		return imageData, "png", bounds.Dx(), bounds.Dy(), nil
	}

	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, "", 0, 0, fmt.Errorf("解码图片失败: %v", err)
	}

	// 调整图片大小
	var finalImg image.Image
	if compressionConfig.EnableResize {
		finalImg = resizeImage(img, compressionConfig.MaxWidth, compressionConfig.MaxHeight)
	} else {
		finalImg = img
	}

	// 获取最终尺寸
	finalBounds := finalImg.Bounds()
	finalWidth := finalBounds.Dx()
	finalHeight := finalBounds.Dy()

	// 编码压缩后的图片
	var buf bytes.Buffer
	var compressedFormat string

	switch strings.ToLower(compressionConfig.Format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, finalImg, &jpeg.Options{Quality: compressionConfig.Quality})
		compressedFormat = "jpg"
	case "png":
		err = png.Encode(&buf, finalImg)
		compressedFormat = "png"
	default:
		// 根据原始格式选择压缩格式
		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, finalImg, &jpeg.Options{Quality: compressionConfig.Quality})
			compressedFormat = "jpg"
		case "png":
			err = png.Encode(&buf, finalImg)
			compressedFormat = "png"
		default:
			err = jpeg.Encode(&buf, finalImg, &jpeg.Options{Quality: compressionConfig.Quality})
			compressedFormat = "jpg"
		}
	}

	if err != nil {
		return nil, "", 0, 0, fmt.Errorf("编码压缩图片失败: %v", err)
	}

	return buf.Bytes(), compressedFormat, finalWidth, finalHeight, nil
}

// compressImage 压缩图片（保留原有函数以保证兼容性）
func compressImage(imageData []byte, compressionConfig *ImageCompressionConfig) ([]byte, string, error) {
	if compressionConfig == nil {
		return imageData, "png", nil
	}

	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, "", fmt.Errorf("解码图片失败: %v", err)
	}

	// 调整图片大小
	if compressionConfig.EnableResize {
		img = resizeImage(img, compressionConfig.MaxWidth, compressionConfig.MaxHeight)
	}

	// 编码压缩后的图片
	var buf bytes.Buffer
	var compressedFormat string

	switch strings.ToLower(compressionConfig.Format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: compressionConfig.Quality})
		compressedFormat = "jpg"
	case "png":
		err = png.Encode(&buf, img)
		compressedFormat = "png"
	default:
		// 根据原始格式选择压缩格式
		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: compressionConfig.Quality})
			compressedFormat = "jpg"
		case "png":
			err = png.Encode(&buf, img)
			compressedFormat = "png"
		default:
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: compressionConfig.Quality})
			compressedFormat = "jpg"
		}
	}

	if err != nil {
		return nil, "", fmt.Errorf("编码压缩图片失败: %v", err)
	}

	return buf.Bytes(), compressedFormat, nil
}

// resizeImage 调整图片大小
func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 如果图片尺寸已经小于最大尺寸，不需要调整
	if width <= maxWidth && height <= maxHeight {
		return img
	}

	// 计算新的尺寸，保持宽高比
	var newWidth, newHeight uint
	if width > height {
		// 横向图片
		newWidth = uint(maxWidth)
		newHeight = uint(float64(height) * float64(maxWidth) / float64(width))
	} else {
		// 纵向图片
		newHeight = uint(maxHeight)
		newWidth = uint(float64(width) * float64(maxHeight) / float64(height))
	}

	// 调整大小
	resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	return resizedImg
}

// ExtractKeywords 从提示词中提取关键词
func extractKeywords(prompt string) string {
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

// ExtractEnglishKeywords 从英文提示词中提取前几个关键词作为文件名
func ExtractEnglishKeywords(englishPrompt string) string {
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

// GetImageList 获取图片列表
func GetImageList(imagesDir string) ([]models.ImageInfo, error) {
	var images []models.ImageInfo

	// 确保目录存在
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// 读取目录
	files, err := os.ReadDir(imagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 检查是否为图片文件
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
			continue
		}

		filePath := filepath.Join(imagesDir, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		images = append(images, models.ImageInfo{
			Filename:    file.Name(),
			URL:         fmt.Sprintf("/static/images/%s", file.Name()),
			CreatedTime: fileInfo.ModTime().Format("2006-01-02 15:04:05"),
			Size:        fileInfo.Size(),
		})
	}

	return images, nil
}

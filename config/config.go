package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Config 应用配置结构
type Config struct {
	// 服务器配置
	Server ServerConfig `toml:"server"`

	// API Keys
	APIKeys APIKeysConfig `toml:"api_keys"`

	// 模型配置
	Models ModelsConfig `toml:"models"`

	// Cloudflare R2 配置
	CloudflareR2 CloudflareR2Config `toml:"cloudflare_r2"`

	// Cloudflare D1 配置
	CloudflareD1 CloudflareD1Config `toml:"cloudflare_d1"`

	// 图片处理配置
	ImageProcessing ImageProcessingConfig `toml:"image_processing"`

	// 日志配置
	Logging LoggingConfig `toml:"logging"`

	// SMTP 邮件配置
	SMTP SMTPConfig `toml:"smtp"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port       int    `toml:"port"`
	ImagesDir  string `toml:"images_dir"`
	MaxRetries int    `toml:"max_retries"`
	Timeout    int    `toml:"timeout"`
}

// APIKeysConfig API密钥配置
type APIKeysConfig struct {
	ModelScopeToken  string `toml:"model_scope_token"`
	OpenRouterAPIKey string `toml:"openrouter_api_key"`
}

// ModelsConfig 模型配置
type ModelsConfig struct {
	ModelScopeModel        string `toml:"model_scope_model"`
	DefaultOpenRouterModel string `toml:"default_openrouter_model"`
}

// CloudflareR2Config Cloudflare R2配置
type CloudflareR2Config struct {
	AccountID       string `toml:"account_id"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`
	Endpoint        string `toml:"endpoint"`
	Bucket          string `toml:"bucket"`
}

// CloudflareD1Config Cloudflare D1配置
type CloudflareD1Config struct {
	AccountID    string `toml:"account_id"`
	APIToken     string `toml:"api_token"`
	DatabaseID   string `toml:"database_id"`
	DatabaseName string `toml:"database_name"`
}

// ImageProcessingConfig 图片处理配置
type ImageProcessingConfig struct {
	MaxWidth     int    `toml:"max_width"`
	MaxHeight    int    `toml:"max_height"`
	Quality      int    `toml:"quality"`
	Format       string `toml:"format"`
	EnableResize bool   `toml:"enable_resize"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
	File   string `toml:"file"`
}

// SMTPConfig SMTP邮件配置
type SMTPConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	From     string `toml:"from"`
	To       string `toml:"to"`
	Enable   bool   `toml:"enable"`
}

// LoadConfig 加载配置，优先从TOML文件，其次从环境变量
func LoadConfig() *Config {
	// 设置默认配置
	config := getDefaultConfig()

	// 尝试从TOML文件加载配置
	if err := loadFromTOML(config); err != nil {
		log.Printf("警告: 加载TOML配置文件失败: %v，将使用默认配置和环境变量", err)
	}

	// 环境变量覆盖配置 (优先级最高)
	loadFromEnv(config)

	// 验证配置
	if err := validateConfig(config); err != nil {
		log.Printf("警告: 配置验证失败: %v", err)
	}

	return config
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:       8000,
			ImagesDir:  "public/static/images",
			MaxRetries: 3,
			Timeout:    30,
		},
		APIKeys: APIKeysConfig{
			ModelScopeToken:  "", // 从环境变量 MODEL_SCOPE_TOKEN 获取
			OpenRouterAPIKey: "", // 从环境变量 OPENROUTER_API_KEY 获取
		},
		Models: ModelsConfig{
			ModelScopeModel:        "deepseek-ai/DeepSeek-V3.1",
			DefaultOpenRouterModel: "google/gemini-2.5-flash-image-preview:free",
		},
		CloudflareR2: CloudflareR2Config{},
		CloudflareD1: CloudflareD1Config{
			DatabaseName: "ai_images",
		},
		ImageProcessing: ImageProcessingConfig{
			MaxWidth:     1920,
			MaxHeight:    1080,
			Quality:      85,
			Format:       "jpeg",
			EnableResize: true,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			File:   "",
		},
		SMTP: SMTPConfig{
			Host:     "smtp.qq.com",
			Port:     587,
			Username: "",
			Password: "",
			From:     "",
			To:       "",
			Enable:   false,
		},
	}
}

// loadFromTOML 从TOML文件加载配置
func loadFromTOML(config *Config) error {
	configFiles := []string{"config.toml", "./config.toml", "../config.toml"}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); err == nil {
			log.Printf("正在加载配置文件: %s", configFile)
			if _, err := toml.DecodeFile(configFile, config); err != nil {
				return fmt.Errorf("解析配置文件 %s 失败: %v", configFile, err)
			}
			log.Printf("配置文件加载成功: %s", configFile)
			return nil
		}
	}

	return fmt.Errorf("未找到配置文件 (查找路径: %v)", configFiles)
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(config *Config) {
	// 服务器配置
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if dir := os.Getenv("IMAGES_DIR"); dir != "" {
		config.Server.ImagesDir = dir
	}
	if retries := os.Getenv("MAX_RETRIES"); retries != "" {
		if r, err := strconv.Atoi(retries); err == nil {
			config.Server.MaxRetries = r
		}
	}
	if timeout := os.Getenv("TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			config.Server.Timeout = t
		}
	}

	// API Keys
	if token := os.Getenv("MODEL_SCOPE_TOKEN"); token != "" {
		config.APIKeys.ModelScopeToken = token
	}
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		config.APIKeys.OpenRouterAPIKey = key
	}

	// 模型配置
	if model := os.Getenv("MODEL_SCOPE_MODEL"); model != "" {
		config.Models.ModelScopeModel = model
	}
	if model := os.Getenv("DEFAULT_OPENROUTER_MODEL"); model != "" {
		config.Models.DefaultOpenRouterModel = model
	}

	// R2 配置
	if accountID := os.Getenv("CLOUDFLARE_R2_ACCOUNT_ID"); accountID != "" {
		config.CloudflareR2.AccountID = accountID
	}
	if keyID := os.Getenv("CLOUDFLARE_R2_ACCOUNT_KEY_ID"); keyID != "" {
		config.CloudflareR2.AccessKeyID = keyID
	}
	if keySecret := os.Getenv("CLOUDFLARE_R2_ACCOUNT_KEY_SECRET"); keySecret != "" {
		config.CloudflareR2.AccessKeySecret = keySecret
	}
	if endpoint := os.Getenv("CLOUDFLARE_R2_URL"); endpoint != "" {
		config.CloudflareR2.Endpoint = endpoint
	}
	if bucket := os.Getenv("CLOUDFLARE_R2_BUCKET"); bucket != "" {
		config.CloudflareR2.Bucket = bucket
	}

	// D1 配置
	if accountID := os.Getenv("CLOUDFLARE_D1_ACCOUNT_ID"); accountID != "" {
		config.CloudflareD1.AccountID = accountID
	}
	if apiToken := os.Getenv("CLOUDFLARE_D1_API_TOKEN"); apiToken != "" {
		config.CloudflareD1.APIToken = apiToken
	}
	if databaseID := os.Getenv("CLOUDFLARE_D1_DATABASE_ID"); databaseID != "" {
		config.CloudflareD1.DatabaseID = databaseID
	}
	if databaseName := os.Getenv("CLOUDFLARE_D1_DATABASE_NAME"); databaseName != "" {
		config.CloudflareD1.DatabaseName = databaseName
	}

	// 图片处理配置
	if maxWidth := os.Getenv("IMAGE_MAX_WIDTH"); maxWidth != "" {
		if w, err := strconv.Atoi(maxWidth); err == nil {
			config.ImageProcessing.MaxWidth = w
		}
	}
	if maxHeight := os.Getenv("IMAGE_MAX_HEIGHT"); maxHeight != "" {
		if h, err := strconv.Atoi(maxHeight); err == nil {
			config.ImageProcessing.MaxHeight = h
		}
	}
	if quality := os.Getenv("IMAGE_QUALITY"); quality != "" {
		if q, err := strconv.Atoi(quality); err == nil {
			config.ImageProcessing.Quality = q
		}
	}
	if format := os.Getenv("IMAGE_FORMAT"); format != "" {
		config.ImageProcessing.Format = format
	}
	if enableResize := os.Getenv("IMAGE_ENABLE_RESIZE"); enableResize != "" {
		config.ImageProcessing.EnableResize = enableResize == "true"
	}

	// 日志配置
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Logging.Format = format
	}
	if file := os.Getenv("LOG_FILE"); file != "" {
		config.Logging.File = file
	}

	// SMTP 邮件配置
	if host := os.Getenv("SMTP_HOST"); host != "" {
		config.SMTP.Host = host
	}
	if port := os.Getenv("SMTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.SMTP.Port = p
		}
	}
	if username := os.Getenv("SMTP_USERNAME"); username != "" {
		config.SMTP.Username = username
	}
	if password := os.Getenv("SMTP_PASSWORD"); password != "" {
		config.SMTP.Password = password
	}
	if from := os.Getenv("SMTP_FROM"); from != "" {
		config.SMTP.From = from
	}
	if to := os.Getenv("SMTP_TO"); to != "" {
		config.SMTP.To = to
	}
	if enable := os.Getenv("SMTP_ENABLE"); enable != "" {
		config.SMTP.Enable = enable == "true"
	}
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d", config.Server.Port)
	}

	if config.APIKeys.ModelScopeToken == "" {
		return fmt.Errorf("ModelScope Token 未设置")
	}

	if config.APIKeys.OpenRouterAPIKey == "" {
		return fmt.Errorf("OpenRouter API Key 未设置")
	}

	if config.ImageProcessing.Quality < 1 || config.ImageProcessing.Quality > 100 {
		return fmt.Errorf("无效的图片质量设置: %d (应在1-100之间)", config.ImageProcessing.Quality)
	}

	// 验证SMTP配置
	if config.SMTP.Enable {
		if config.SMTP.Host == "" {
			return fmt.Errorf("SMTP主机未设置")
		}
		if config.SMTP.Port <= 0 {
			return fmt.Errorf("SMTP端口无效: %d", config.SMTP.Port)
		}
		if config.SMTP.Username == "" {
			return fmt.Errorf("SMTP用户名未设置")
		}
		if config.SMTP.Password == "" {
			return fmt.Errorf("SMTP密码未设置")
		}
		if config.SMTP.From == "" {
			return fmt.Errorf("SMTP发件人未设置")
		}
		if config.SMTP.To == "" {
			return fmt.Errorf("SMTP收件人未设置")
		}
	}

	return nil
}

// 兼容性方法 - 保持与旧配置接口的兼容性

// Port 获取服务器端口
func (c *Config) Port() int {
	return c.Server.Port
}

// ImagesDir 获取图片目录
func (c *Config) ImagesDir() string {
	return c.Server.ImagesDir
}

// ModelScopeToken 获取ModelScope令牌
func (c *Config) ModelScopeToken() string {
	return c.APIKeys.ModelScopeToken
}

// OpenRouterAPIKey 获取OpenRouter API密钥
func (c *Config) OpenRouterAPIKey() string {
	return c.APIKeys.OpenRouterAPIKey
}

// ModelScopeModel 获取ModelScope模型
func (c *Config) ModelScopeModel() string {
	return c.Models.ModelScopeModel
}

// DefaultOpenRouterModel 获取默认OpenRouter模型
func (c *Config) DefaultOpenRouterModel() string {
	return c.Models.DefaultOpenRouterModel
}

// R2AccountID 获取R2账户ID
func (c *Config) R2AccountID() string {
	return c.CloudflareR2.AccountID
}

// R2AccessKeyID 获取R2访问密钥ID
func (c *Config) R2AccessKeyID() string {
	return c.CloudflareR2.AccessKeyID
}

// R2AccessKeySecret 获取R2访问密钥密码
func (c *Config) R2AccessKeySecret() string {
	return c.CloudflareR2.AccessKeySecret
}

// R2Endpoint 获取R2端点
func (c *Config) R2Endpoint() string {
	return c.CloudflareR2.Endpoint
}

// R2Bucket 获取R2存储桶
func (c *Config) R2Bucket() string {
	return c.CloudflareR2.Bucket
}

// ImageMaxWidth 获取图片最大宽度
func (c *Config) ImageMaxWidth() int {
	return c.ImageProcessing.MaxWidth
}

// ImageMaxHeight 获取图片最大高度
func (c *Config) ImageMaxHeight() int {
	return c.ImageProcessing.MaxHeight
}

// ImageQuality 获取图片质量
func (c *Config) ImageQuality() int {
	return c.ImageProcessing.Quality
}

// ImageFormat 获取图片格式
func (c *Config) ImageFormat() string {
	return c.ImageProcessing.Format
}

// ImageEnableResize 获取是否启用图片缩放
func (c *Config) ImageEnableResize() bool {
	return c.ImageProcessing.EnableResize
}

// MaxRetries 获取最大重试次数
func (c *Config) MaxRetries() int {
	return c.Server.MaxRetries
}

// Timeout 获取超时时间
func (c *Config) Timeout() int {
	return c.Server.Timeout
}

// D1AccountID 获取D1账户ID
func (c *Config) D1AccountID() string {
	return c.CloudflareD1.AccountID
}

// D1APIToken 获取D1 API令牌
func (c *Config) D1APIToken() string {
	return c.CloudflareD1.APIToken
}

// D1DatabaseID 获取D1数据库ID
func (c *Config) D1DatabaseID() string {
	return c.CloudflareD1.DatabaseID
}

// D1DatabaseName 获取D1数据库名称
func (c *Config) D1DatabaseName() string {
	return c.CloudflareD1.DatabaseName
}

// SMTPHost 获取SMTP主机
func (c *Config) SMTPHost() string {
	return c.SMTP.Host
}

// SMTPPort 获取SMTP端口
func (c *Config) SMTPPort() int {
	return c.SMTP.Port
}

// SMTPUsername 获取SMTP用户名
func (c *Config) SMTPUsername() string {
	return c.SMTP.Username
}

// SMTPPassword 获取SMTP密码
func (c *Config) SMTPPassword() string {
	return c.SMTP.Password
}

// SMTPFrom 获取SMTP发件人
func (c *Config) SMTPFrom() string {
	return c.SMTP.From
}

// SMTPTo 获取SMTP收件人
func (c *Config) SMTPTo() string {
	return c.SMTP.To
}

// SMTPEnable 获取SMTP启用状态
func (c *Config) SMTPEnable() bool {
	return c.SMTP.Enable
}

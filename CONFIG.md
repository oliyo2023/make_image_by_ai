# 配置说明

## 环境变量配置

你可以通过环境变量来配置应用程序。以下是所有可用的配置项：

### API Keys
- `MODEL_SCOPE_TOKEN`: ModelScope API 令牌
- `OPENROUTER_API_KEY`: OpenRouter API 密钥

### Cloudflare R2 配置
- `CLOUDFLARE_R2_ACCOUNT_ID`: Cloudflare R2 账户 ID
- `CLOUDFLARE_R2_ACCOUNT_KEY_ID`: R2 API 密钥 ID
- `CLOUDFLARE_R2_ACCOUNT_KEY_SECRET`: R2 API 密钥密码
- `CLOUDFLARE_R2_URL`: R2 端点 URL
- `CLOUDFLARE_R2_BUCKET`: R2 存储桶名称

### 模型配置
- `MODEL_SCOPE_MODEL`: ModelScope 翻译模型
- `DEFAULT_OPENROUTER_MODEL`: 默认的 OpenRouter 图像生成模型

### 服务器配置
- `PORT`: 服务器端口号（默认：8000）
- `IMAGES_DIR`: 图片存储目录（默认：public/static/images）

### 图片压缩配置
- `IMAGE_MAX_WIDTH`: 图片最大宽度（默认：1920）
- `IMAGE_MAX_HEIGHT`: 图片最大高度（默认：1080）
- `IMAGE_QUALITY`: 图片质量（1-100，默认：85）
- `IMAGE_FORMAT`: 压缩格式（jpeg/png，默认：jpeg）
- `IMAGE_ENABLE_RESIZE`: 是否启用调整大小（true/false，默认：true）

### 其他配置
- `MAX_RETRIES`: 最大重试次数（默认：3）
- `TIMEOUT`: 请求超时时间（秒，默认：30）

## 使用示例

### Windows PowerShell
```powershell
$env:MODEL_SCOPE_TOKEN="your-model-scope-token"
$env:OPENROUTER_API_KEY="your-openrouter-api-key"
$env:CLOUDFLARE_R2_ACCOUNT_ID="your-r2-account-id"
$env:CLOUDFLARE_R2_ACCOUNT_KEY_ID="your-r2-key-id"
$env:CLOUDFLARE_R2_ACCOUNT_KEY_SECRET="your-r2-key-secret"
$env:CLOUDFLARE_R2_URL="https://your-account-id.r2.cloudflarestorage.com"
$env:CLOUDFLARE_R2_BUCKET="your-bucket-name"
$env:IMAGE_MAX_WIDTH="1920"
$env:IMAGE_MAX_HEIGHT="1080"
$env:IMAGE_QUALITY="85"
$env:IMAGE_FORMAT="jpeg"
$env:IMAGE_ENABLE_RESIZE="true"
$env:PORT="8080"
go run main.go
```

### Linux/macOS
```bash
export MODEL_SCOPE_TOKEN="your-model-scope-token"
export OPENROUTER_API_KEY="your-openrouter-api-key"
export CLOUDFLARE_R2_ACCOUNT_ID="your-r2-account-id"
export CLOUDFLARE_R2_ACCOUNT_KEY_ID="your-r2-key-id"
export CLOUDFLARE_R2_ACCOUNT_KEY_SECRET="your-r2-key-secret"
export CLOUDFLARE_R2_URL="https://your-account-id.r2.cloudflarestorage.com"
export CLOUDFLARE_R2_BUCKET="your-bucket-name"
export IMAGE_MAX_WIDTH="1920"
export IMAGE_MAX_HEIGHT="1080"
export IMAGE_QUALITY="85"
export IMAGE_FORMAT="jpeg"
export IMAGE_ENABLE_RESIZE="true"
export PORT="8080"
go run main.go
```

### 使用 .env 文件（需要第三方库支持）
```bash
# 安装 godotenv
go get github.com/joho/godotenv

# 创建 .env 文件
MODEL_SCOPE_TOKEN=your-model-scope-token
OPENROUTER_API_KEY=your-openrouter-api-key
PORT=8080
```

## 默认配置

如果没有设置环境变量，程序将使用以下默认值：

```go
ModelScopeToken:       "ms-97290993-2317-4f65-82a9-cad949f6b2a0"
OpenRouterAPIKey:      "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"
ModelScopeModel:       "deepseek-ai/DeepSeek-V3.1"
DefaultOpenRouterModel: "google/gemini-2.5-flash-image-preview:free"
Port:                  8000
ImagesDir:             "public/static/images"
MaxRetries:            3
Timeout:               30
```

## 安全建议

1. **不要将 API 密钥提交到版本控制系统**
2. **在生产环境中使用环境变量或配置文件**
3. **定期轮换 API 密钥**
4. **限制 API 密钥的权限范围**

# TOML 配置管理

项目现已支持 TOML 格式的配置文件管理，提供更直观和结构化的配置方式。

## 配置加载优先级

配置加载遵循以下优先级顺序（高优先级覆盖低优先级）：

1. **环境变量** (最高优先级)
2. **TOML 配置文件**
3. **默认配置** (最低优先级)

## 配置文件

### 配置文件位置

系统会按以下顺序查找配置文件：
- `config.toml` (当前目录)
- `./config.toml` (当前目录)
- `../config.toml` (上级目录)

### 配置文件示例

参考 `config.example.toml` 文件，复制并重命名为 `config.toml` 后修改相应配置。

## 配置结构

### 1. 服务器配置 `[server]`
```toml
[server]
port = 8000                    # 服务器端口
images_dir = "public/static/images"  # 图片存储目录
max_retries = 3               # 最大重试次数
timeout = 30                  # 请求超时时间(秒)
```

### 2. API 密钥配置 `[api_keys]`
```toml
[api_keys]
model_scope_token = "your-token"      # ModelScope API 令牌
openrouter_api_key = "your-api-key"   # OpenRouter API 密钥
```

### 3. 模型配置 `[models]`
```toml
[models]
model_scope_model = "deepseek-ai/DeepSeek-V3.1"
default_openrouter_model = "google/gemini-2.5-flash-image-preview:free"
```

### 4. Cloudflare R2 配置 `[cloudflare_r2]` (可选)
```toml
[cloudflare_r2]
account_id = "your-account-id"
access_key_id = "your-access-key-id"
access_key_secret = "your-access-key-secret"
endpoint = "https://your-account-id.r2.cloudflarestorage.com"
bucket = "your-bucket-name"
```

### 5. Cloudflare D1 配置 `[cloudflare_d1]`
```toml
[cloudflare_d1]
account_id = "your-cloudflare-account-id"    # Cloudflare 账户ID
api_token = "your-cloudflare-api-token"      # Cloudflare API令牌
database_id = "your-d1-database-id"          # D1数据库ID
database_name = "ai_images"                  # 数据库名称
```

**D1配置获取步骤**：
1. 登录 [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. 转到 "Workers & Pages" > "D1 SQL Database"
3. 创建或选择数据库，获取 Database ID
4. 在 "My Profile" > "API Tokens" 中创建 API 令牌
5. 账户ID在右侧栏中可以找到

### 6. 图片处理配置 `[image_processing]`
```toml
[image_processing]
max_width = 1920           # 图片最大宽度
max_height = 1080          # 图片最大高度
quality = 85               # 压缩质量(1-100)
format = "jpeg"            # 压缩格式(jpeg/png)
enable_resize = true       # 是否启用缩放
```

### 7. 日志配置 `[logging]`
```toml
[logging]
level = "info"             # 日志级别(debug/info/warn/error)
format = "text"            # 日志格式(text/json)
file = ""                  # 日志文件路径(空=控制台输出)
```

## 环境变量覆盖

即使使用 TOML 配置文件，仍可通过环境变量覆盖特定配置：

```bash
# 服务器配置
export PORT=8080
export IMAGES_DIR="/custom/path"
export MAX_RETRIES=5
export TIMEOUT=60

# API 密钥
export MODEL_SCOPE_TOKEN="your-token"
export OPENROUTER_API_KEY="your-key"

# 模型配置
export MODEL_SCOPE_MODEL="custom-model"
export DEFAULT_OPENROUTER_MODEL="custom-image-model"

# R2 配置
export CLOUDFLARE_R2_ACCOUNT_ID="account-id"
export CLOUDFLARE_R2_ACCOUNT_KEY_ID="key-id"
export CLOUDFLARE_R2_ACCOUNT_KEY_SECRET="key-secret"
export CLOUDFLARE_R2_URL="https://endpoint"
export CLOUDFLARE_R2_BUCKET="bucket-name"

# D1 配置
export CLOUDFLARE_D1_ACCOUNT_ID="account-id"
export CLOUDFLARE_D1_API_TOKEN="api-token"
export CLOUDFLARE_D1_DATABASE_ID="database-id"
export CLOUDFLARE_D1_DATABASE_NAME="ai_images"

# 图片处理配置
export IMAGE_MAX_WIDTH=2560
export IMAGE_MAX_HEIGHT=1440
export IMAGE_QUALITY=90
export IMAGE_FORMAT=png
export IMAGE_ENABLE_RESIZE=false

# 日志配置
export LOG_LEVEL=debug
export LOG_FORMAT=json
export LOG_FILE="/var/log/ai-image-generator.log"
```

## 快速开始

1. **复制配置文件模板**
   ```bash
   cp config.example.toml config.toml
   ```

2. **编辑配置文件**
   ```bash
   # 编辑 config.toml，设置你的 API 密钥和其他配置
   ```

3. **运行应用**
   ```bash
   go run main.go
   ```

## 配置验证

应用启动时会自动验证配置的有效性：
- 检查端口号是否在有效范围内 (1-65535)
- 验证必需的 API 密钥是否已设置
- 检查图片质量设置是否在有效范围内 (1-100)

如果配置验证失败，会输出警告信息但不会阻止应用启动。

## 迁移指南

从旧版本的环境变量配置迁移到 TOML 配置：

1. 创建 `config.toml` 文件
2. 将环境变量值迁移到对应的 TOML 配置项
3. 删除或保留环境变量（环境变量仍会覆盖 TOML 配置）

配置结构保持向后兼容，无需修改现有代码。
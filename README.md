# AI 图像生成器

一个基于 Go 语言开发的 AI 图像生成服务，支持中文提示词翻译和图像生成。

## 项目结构

```
make_image_by_ai/
├── config/
│   └── config.go          # 配置管理
├── models/
│   └── models.go          # 数据模型定义
├── services/
│   └── image_service.go   # 图像生成服务
├── handlers/
│   └── handlers.go        # HTTP 处理器
├── utils/
│   └── utils.go           # 工具函数
├── main.go                # 主程序入口
├── test_client.go         # 测试客户端
├── go.mod                 # Go 模块文件
├── go.sum                 # Go 依赖校验
├── Makefile               # 构建脚本
├── CONFIG.md              # 配置说明文档
└── public/
    └── static/
        └── images/        # 生成的图片存储目录
```

## 功能特性

- 🎨 **AI 图像生成**：支持多种 AI 模型生成高质量图像
- 🌐 **中文翻译**：自动将中文提示词翻译为英文
- 💾 **本地存储**：生成的图片自动保存到本地
- 🔧 **配置灵活**：支持环境变量配置
- 📱 **RESTful API**：提供完整的 HTTP API 接口
- 🛡️ **错误处理**：完善的错误处理和日志记录

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置环境变量（可选）

优先使用 TOML 配置文件：
```bash
# 复制配置文件模板
cp config.example.toml config.toml
# 编辑 config.toml 设置你的 API 密钥
```

或者使用环境变量（优先级更高）：

```bash
# Windows PowerShell
$env:MODEL_SCOPE_TOKEN="your-model-scope-token"
$env:OPENROUTER_API_KEY="your-openrouter-api-key"
$env:PORT="8000"

# Linux/macOS
export MODEL_SCOPE_TOKEN="your-model-scope-token"
export OPENROUTER_API_KEY="your-openrouter-api-key"
export PORT="8000"
```

### 3. 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8000` 启动。

## API 接口

### 健康检查
```
GET /health
```

### 图像生成
```
POST /generate-image
Content-Type: application/json

{
  "prompt": "一只可爱的小猫",
  "model": "google/gemini-2.5-flash-image-preview:free"
}
```

### 文本翻译
```
POST /translate
Content-Type: application/json

{
  "text": "一只可爱的小猫"
}
```

### 图片列表
```
GET /images
```

### 图片记录查询
```
GET /records?page=1&limit=20&keyword=龙&model=google/gemini-2.5-flash-image-preview:free&date_from=2024-01-01&date_to=2024-12-31
```

查询参数：
- `page`: 页码（默认1）
- `limit`: 每页数量（默认20，最大100）
- `keyword`: 关键词搜索（搜索原始提示词和英文提示词）
- `model`: 模型筛选
- `date_from`: 开始日期（YYYY-MM-DD）
- `date_to`: 结束日期（YYYY-MM-DD）

### 获取单个图片记录
```
GET /records/{id}
```

## 配置说明

项目现已支持 TOML 格式的配置文件管理，提供更直观和结构化的配置方式。

### 配置加载优先级
1. **环境变量** (最高优先级)
2. **TOML 配置文件** 
3. **默认配置** (最低优先级)

详细的配置说明请参考 [TOML_CONFIG.md](TOML_CONFIG.md) 文件。

### 默认配置

- **端口**: 8000
- **图片目录**: `public/static/images`
- **ModelScope 模型**: `deepseek-ai/DeepSeek-V3.1`
- **OpenRouter 模型**: `google/gemini-2.5-flash-image-preview:free`

## 测试

运行测试客户端：

```bash
go run -tags testclient test_client.go
```

## 构建

```bash
# 构建主程序
go build -o ai-image-generator main.go

# 构建测试客户端
go build -tags testclient -o test-client test_client.go
```

## 项目架构

### 模块化设计

- **config**: 配置管理，支持环境变量和默认值
- **models**: 数据模型定义，包含请求和响应结构
- **services**: 业务逻辑层，处理图像生成和翻译
- **handlers**: HTTP 处理器，处理 API 请求
- **utils**: 工具函数，包含文件操作和图片处理

### 依赖注入

项目使用依赖注入模式，便于测试和维护：

```go
// 创建服务实例
config := config.LoadConfig()
imageService := services.NewImageService(config)
handler := handlers.NewHandler(imageService)
```

## 开发说明

### 添加新功能

1. 在 `models/` 中定义数据结构
2. 在 `services/` 中实现业务逻辑
3. 在 `handlers/` 中添加 HTTP 处理器
4. 在 `main.go` 中注册路由

### 错误处理

所有错误都会返回适当的 HTTP 状态码和错误信息：

```json
{
  "error": "错误描述"
}
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
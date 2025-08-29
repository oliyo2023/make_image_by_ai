# Make AI Image

一个基于 Go 语言开发的 AI 图像生成服务，支持中文提示词翻译和多种 AI 模型的图像生成。

## 🚀 项目特色

- **多语言支持**: 自动将中文提示词翻译为英文
- **多模型支持**: 集成 ModelScope 和 OpenRouter 等多种 AI 服务
- **双重存储**: 本地存储 + Cloudflare R2 云存储备份
- **完整记录**: Cloudflare D1 数据库存储生成记录
- **RESTful API**: 标准化的 HTTP API 接口
- **无服务器架构**: 支持 Cloudflare Workers 部署

## 📋 系统功能

### 核心功能
- AI 图像生成：支持多种 AI 模型生成高质量图像
- 中文翻译：自动将中文提示词翻译为英文
- 本地存储：生成的图片自动保存到本地目录
- RESTful API：提供完整的 HTTP API 接口
- 错误处理：完善的错误处理和日志记录

### API 端点
- `GET /health` - 健康检查
- `POST /generate-image` - 生成图像
- `POST /translate` - 翻译提示词
- `GET /images` - 获取图像列表
- `GET /static/*` - 静态文件服务
- `GET /records` - 分页查询图片记录
- `GET /records/{id}` - 获取单个记录

## 🛠️ 技术栈

### 后端 (Go)
- **框架**: Gin v1.9.1
- **AI 服务**: OpenAI API v1.41.1
- **云服务**: AWS SDK v1.55.8
- **工具库**: UUID v1.3.0, nfnt/resize

### 前端支持 (Python)
- **框架**: FastAPI ≥0.104.1
- **服务器**: Uvicorn ≥0.24.0
- **AI 集成**: OpenAI ≥1.3.0
- **数据验证**: Pydantic ≥2.5.0

### 云服务 (Cloudflare)
- **计算**: Cloudflare Workers (TypeScript/Python)
- **存储**: Cloudflare R2
- **数据库**: Cloudflare D1
- **前端**: Cloudflare Pages

## 🚀 快速开始

### 环境要求
- Go 1.19+
- Python 3.x (可选)
- Node.js 18+ (用于 Cloudflare 部署)

### 本地开发

1. **克隆项目**
```bash
git clone https://github.com/YOUR_USERNAME/make_ai_image.git
cd make_ai_image
```

2. **安装依赖**
```bash
go mod tidy
```

3. **配置环境变量**
```bash
# 方法1: 使用配置文件
cp config.example.toml config.toml
# 编辑 config.toml 文件，填入您的 API 密钥

# 方法2: 使用环境变量文件  
cp .env.example .env
# 编辑 .env 文件，填入您的 API 密钥

# 方法3: 直接设置环境变量
export MODEL_SCOPE_TOKEN="your-model-scope-token"
export OPENROUTER_API_KEY="your-openrouter-api-key"
export PORT="8000"
```

**⚠️ 安全提醒**：
- 🔐 **切勿**将真实的 API 密钥提交到代码仓库
- 📝 使用 `.env` 或 `config.toml` 文件存储密钥（已在 .gitignore 中排除）
- 🔄 定期轮换 API 密钥
- 🛡️ 为不同环境使用不同的密钥

4. **运行服务**
```bash
# 开发模式
go run main.go

# 构建并运行
go build -o ai-image-generator main.go
./ai-image-generator
```

服务将在 http://localhost:8000 启动。

### Cloudflare 部署

详细部署说明请参考 [cloudflare-site/DEPLOYMENT.md](cloudflare-site/DEPLOYMENT.md)

## 📁 项目结构

```
├── config/                 # 配置管理
├── handlers/              # HTTP 处理器
├── models/                # 数据模型
├── services/              # 业务服务
│   ├── d1_service.go     # D1 数据库服务
│   ├── image_service.go  # 图像生成服务
│   └── r2_service.go     # R2 存储服务
├── utils/                 # 工具函数
├── cloudflare-site/       # Cloudflare 部署相关
│   ├── worker/           # Workers 代码
│   ├── pages/            # Pages 前端
│   └── DEPLOYMENT.md     # 部署说明
├── public/static/images/  # 本地图片存储
├── main.go               # 主程序
├── config.example.toml   # 配置文件示例
└── README.md             # 项目说明
```

## ⚙️ 配置说明

项目使用 TOML 格式配置文件，支持环境变量覆盖：

### 主要配置段
- `[server]` - 服务器配置
- `[api_keys]` - API 密钥配置
- `[models]` - AI 模型配置
- `[cloudflare_r2]` - R2 存储配置
- `[cloudflare_d1]` - D1 数据库配置
- `[image_processing]` - 图像处理配置
- `[logging]` - 日志配置

详细配置说明请参考 [CONFIG.md](CONFIG.md) 和 [TOML_CONFIG.md](TOML_CONFIG.md)。

## 🔒 安全规范

- 使用环境变量保护敏感信息（API 密钥）
- 禁止在日志中记录图片内容或 base64 数据
- 智能文件命名避免中文字符
- 完善的错误处理和容错机制

## 🧪 测试

```bash
# 运行测试
go test ./...

# 生成测试覆盖率报告
go test -cover ./...
```

## 📚 API 文档

详细的 API 文档请参考 [API_DOCS.md](API_DOCS.md)。

## 🤝 贡献指南

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 问题反馈

如果您遇到问题或有建议，请 [创建 Issue](https://github.com/YOUR_USERNAME/make_ai_image/issues)。

## 🙏 致谢

感谢以下开源项目：
- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [OpenAI Go](https://github.com/sashabaranov/go-openai) - OpenAI API 客户端
- [Cloudflare](https://cloudflare.com) - 边缘计算服务
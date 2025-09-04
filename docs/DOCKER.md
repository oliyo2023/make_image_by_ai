# Docker 部署指南

本文档说明如何使用 Docker 部署 AI 图像生成器项目。

## 快速开始

### 1. 准备环境

确保已安装以下软件：
- Docker
- Docker Compose（可选）

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，填入实际的 API 密钥
# MODEL_SCOPE_TOKEN=your-actual-token
# OPENROUTER_API_KEY=your-actual-key
```

### 3. 创建必要目录

```bash
# 创建图片存储和日志目录
mkdir -p public/static/images logs
```

## 使用方法

### 方法一：使用 Makefile（推荐）

```bash
# 设置 Docker 环境
make setup-docker-env

# 构建 Docker 镜像
make docker-build

# 运行容器
make docker-run

# 查看日志
make docker-logs

# 停止容器
make docker-stop
```

### 方法二：使用 Docker Compose（推荐用于生产）

```bash
# 启动所有服务（包括 Redis）
make docker-compose-up

# 查看日志
make docker-compose-logs

# 停止所有服务
make docker-compose-down
```

### 方法三：直接使用 Docker 命令

```bash
# 构建镜像
docker build -t ai-image-generator:latest .

# 运行容器
docker run -d --name ai-image-generator \\
  -p 8000:8000 \\
  -v $(pwd)/public/static/images:/app/public/static/images \\
  -v $(pwd)/logs:/app/logs \\
  --env-file .env \\
  ai-image-generator:latest

# 查看日志
docker logs -f ai-image-generator

# 停止容器
docker stop ai-image-generator
docker rm ai-image-generator
```

## 环境变量配置

### 必需的环境变量

| 变量名 | 描述 | 示例值 |
|--------|------|--------|
| `MODEL_SCOPE_TOKEN` | ModelScope API 令牌 | `your-model-scope-token` |
| `OPENROUTER_API_KEY` | OpenRouter API 密钥 | `your-openrouter-api-key` |

### 可选的环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `PORT` | 服务端口 | `8000` |
| `IMAGES_DIR` | 图片存储目录 | `/app/public/static/images` |
| `MAX_RETRIES` | 最大重试次数 | `3` |
| `TIMEOUT` | 请求超时时间（秒） | `30` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `LOG_FORMAT` | 日志格式 | `text` |

### Cloudflare 配置（可选）

| 变量名 | 描述 |
|--------|------|
| `CLOUDFLARE_R2_ACCOUNT_ID` | Cloudflare R2 账户 ID |
| `CLOUDFLARE_R2_ACCESS_KEY_ID` | R2 访问密钥 ID |
| `CLOUDFLARE_R2_ACCESS_KEY_SECRET` | R2 访问密钥 |
| `CLOUDFLARE_R2_ENDPOINT` | R2 端点 URL |
| `CLOUDFLARE_R2_BUCKET` | R2 存储桶名称 |
| `CLOUDFLARE_D1_ACCOUNT_ID` | Cloudflare D1 账户 ID |
| `CLOUDFLARE_D1_API_TOKEN` | D1 API 令牌 |
| `CLOUDFLARE_D1_DATABASE_ID` | D1 数据库 ID |

## 访问服务

服务启动后，可以通过以下端点访问：

- 健康检查: http://localhost:8000/health
- 图像生成: http://localhost:8000/generate-image
- 翻译服务: http://localhost:8000/translate
- 图片列表: http://localhost:8000/images
- 静态文件: http://localhost:8000/static/images/

## 数据持久化

### 挂载的目录

- `./public/static/images:/app/public/static/images` - 生成的图片存储
- `./logs:/app/logs` - 应用日志

### Redis 数据（使用 docker-compose 时）

Redis 数据存储在 Docker 卷 `redis_data` 中，确保数据持久化。

## 健康检查

Docker 容器包含内置的健康检查：
- 检查间隔：30秒
- 超时时间：10秒
- 重试次数：3次
- 启动等待：40秒

## 故障排除

### 1. 检查容器状态

```bash
docker ps -a
```

### 2. 查看容器日志

```bash
docker logs ai-image-generator
```

### 3. 进入容器调试

```bash
docker exec -it ai-image-generator sh
```

### 4. 检查健康状态

```bash
docker inspect --format='{{.State.Health.Status}}' ai-image-generator
```

### 5. 常见问题

**问题：容器启动失败**
- 检查环境变量是否正确设置
- 确保端口 8000 未被占用
- 查看容器日志获取详细错误信息

**问题：无法访问生成的图片**
- 确保 `public/static/images` 目录已正确挂载
- 检查目录权限

**问题：API 调用失败**
- 验证 API 密钥是否有效
- 检查网络连接
- 查看应用日志了解具体错误

## 生产部署建议

### 1. 安全考虑

- 使用非 root 用户运行容器（已在 Dockerfile 中配置）
- 定期更新基础镜像
- 使用 secrets 管理敏感信息

### 2. 性能优化

- 根据需求调整容器资源限制
- 使用多阶段构建减小镜像大小
- 配置适当的健康检查参数

### 3. 监控和日志

- 使用外部日志收集系统
- 配置监控和告警
- 定期备份重要数据

### 4. 扩展性

- 使用 Kubernetes 进行容器编排
- 配置负载均衡
- 实现水平扩展

## 清理

### 停止并删除容器

```bash
make docker-stop
# 或
docker-compose down
```

### 清理 Docker 资源

```bash
make docker-clean
```

### 删除所有相关资源

```bash
# 停止服务
make docker-compose-down

# 删除镜像
docker rmi ai-image-generator:latest

# 清理系统
docker system prune -f
```
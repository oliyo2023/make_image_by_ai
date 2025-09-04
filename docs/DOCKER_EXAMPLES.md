# Docker 完整使用示例

本文档提供了详细的Docker使用示例和测试步骤。

## 目录结构

确保您的项目目录包含以下Docker相关文件：

```
make_image_by_ai/
├── Dockerfile              # 标准多阶段构建
├── Dockerfile.minimal      # 最小化镜像
├── .dockerignore           # Docker忽略文件
├── docker-compose.yml      # 容器编排
├── .env.example           # 环境变量模板
├── deploy.sh              # Linux/Mac部署脚本
├── deploy.bat             # Windows部署脚本
├── DOCKER.md              # Docker文档
└── README.md              # 项目说明
```

## 快速开始

### 方法一：使用部署脚本（推荐）

#### Windows用户

```batch
# 1. 设置环境
deploy.bat setup

# 2. 编辑 .env 文件，设置API密钥
notepad .env

# 3. 构建镜像
deploy.bat build

# 4. 运行服务
deploy.bat run

# 5. 查看状态
deploy.bat status

# 6. 查看日志
deploy.bat logs
```

#### Linux/Mac用户

```bash
# 给脚本执行权限
chmod +x deploy.sh

# 1. 设置环境
./deploy.sh setup

# 2. 编辑 .env 文件，设置API密钥
nano .env

# 3. 构建镜像
./deploy.sh build

# 4. 运行服务
./deploy.sh run

# 5. 查看状态
./deploy.sh status

# 6. 查看日志
./deploy.sh logs
```

### 方法二：使用Docker Compose

```bash
# 1. 复制环境变量文件
cp .env.example .env

# 2. 编辑环境变量
nano .env  # 或使用您喜欢的编辑器

# 3. 启动所有服务
docker-compose up -d

# 4. 查看日志
docker-compose logs -f

# 5. 停止服务
docker-compose down
```

### 方法三：直接使用Docker命令

```bash
# 1. 构建镜像
docker build -t ai-image-generator:latest .

# 2. 创建网络（可选）
docker network create ai-network

# 3. 运行容器
docker run -d \
  --name ai-image-generator \
  --network ai-network \
  -p 8000:8000 \
  -v $(pwd)/public/static/images:/app/public/static/images \
  -v $(pwd)/logs:/app/logs \
  --env-file .env \
  --restart unless-stopped \
  ai-image-generator:latest

# 4. 查看日志
docker logs -f ai-image-generator
```

## 环境变量配置示例

### .env 文件示例

```bash
# 必需配置
MODEL_SCOPE_TOKEN=ms-xxxxxxxxxxxxxxxxxxxxxxxxxxxx
OPENROUTER_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxxxxxx

# 服务配置
PORT=8000
IMAGES_DIR=/app/public/static/images
MAX_RETRIES=3
TIMEOUT=30

# 模型配置
MODEL_SCOPE_MODEL=deepseek-ai/DeepSeek-V3.1
DEFAULT_OPENROUTER_MODEL=google/gemini-2.5-flash-image-preview:free

# Cloudflare R2 配置（可选）
CLOUDFLARE_R2_ACCOUNT_ID=your-account-id
CLOUDFLARE_R2_ACCESS_KEY_ID=your-access-key-id
CLOUDFLARE_R2_ACCESS_KEY_SECRET=your-access-key-secret
CLOUDFLARE_R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
CLOUDFLARE_R2_BUCKET=ai-images

# Cloudflare D1 配置（可选）
CLOUDFLARE_D1_ACCOUNT_ID=your-account-id
CLOUDFLARE_D1_API_TOKEN=your-api-token
CLOUDFLARE_D1_DATABASE_ID=your-database-id
CLOUDFLARE_D1_DATABASE_NAME=ai_images

# 图片处理配置
MAX_WIDTH=1920
MAX_HEIGHT=1080
QUALITY=85
FORMAT=jpeg
ENABLE_RESIZE=true

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=text
LOG_FILE=
```

## 测试服务

### 1. 健康检查

```bash
# 使用curl测试健康检查端点
curl http://localhost:8000/health

# 预期响应
{
  "status": "ok",
  "timestamp": "2024-01-20T12:00:00Z"
}
```

### 2. 翻译服务测试

```bash
# 测试中文翻译功能
curl -X POST http://localhost:8000/translate \
  -H "Content-Type: application/json" \
  -d '{
    "text": "一只可爱的小猫咪在花园里玩耍"
  }'

# 预期响应
{
  "original_text": "一只可爱的小猫咪在花园里玩耍",
  "translated_text": "A cute little kitten playing in the garden",
  "translation_time": "2024-01-20T12:00:00Z"
}
```

### 3. 图像生成测试

```bash
# 测试图像生成功能
curl -X POST http://localhost:8000/generate-image \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪",
    "model": "google/gemini-2.5-flash-image-preview:free",
    "width": 512,
    "height": 512
  }'

# 预期响应
{
  "success": true,
  "image_url": "http://localhost:8000/static/images/cute-little-kitten-uuid.jpeg",
  "local_path": "/app/public/static/images/cute-little-kitten-uuid.jpeg",
  "original_prompt": "一只可爱的小猫咪",
  "english_prompt": "A cute little kitten",
  "model": "google/gemini-2.5-flash-image-preview:free",
  "generation_time": "2024-01-20T12:00:00Z"
}
```

### 4. 图片列表测试

```bash
# 获取生成的图片列表
curl http://localhost:8000/images

# 预期响应
{
  "images": [
    {
      "filename": "cute-little-kitten-uuid.jpeg",
      "url": "http://localhost:8000/static/images/cute-little-kitten-uuid.jpeg",
      "created_at": "2024-01-20T12:00:00Z"
    }
  ],
  "total": 1
}
```

## 容器管理

### 查看容器状态

```bash
# 查看运行中的容器
docker ps

# 查看所有容器
docker ps -a

# 查看容器详细信息
docker inspect ai-image-generator

# 查看容器资源使用情况
docker stats ai-image-generator
```

### 容器日志管理

```bash
# 查看实时日志
docker logs -f ai-image-generator

# 查看最近100行日志
docker logs --tail 100 ai-image-generator

# 查看特定时间段的日志
docker logs --since="2024-01-20T10:00:00" ai-image-generator
```

### 进入容器调试

```bash
# 进入运行中的容器
docker exec -it ai-image-generator sh

# 查看容器内文件系统
docker exec ai-image-generator ls -la /app

# 查看环境变量
docker exec ai-image-generator env
```

## 数据备份和恢复

### 备份生成的图片

```bash
# 创建备份目录
mkdir -p backup/$(date +%Y%m%d)

# 备份图片文件
cp -r public/static/images/* backup/$(date +%Y%m%d)/

# 使用tar压缩备份
tar -czf backup/images-$(date +%Y%m%d).tar.gz public/static/images/
```

### 备份日志文件

```bash
# 备份日志
tar -czf backup/logs-$(date +%Y%m%d).tar.gz logs/
```

## 性能优化

### 1. 镜像优化

```bash
# 使用最小化Dockerfile
docker build -f Dockerfile.minimal -t ai-image-generator:minimal .

# 查看镜像大小
docker images ai-image-generator
```

### 2. 容器资源限制

```bash
# 限制CPU和内存使用
docker run -d \
  --name ai-image-generator \
  --cpus="2.0" \
  --memory="2g" \
  -p 8000:8000 \
  ai-image-generator:latest
```

### 3. 使用多阶段构建

```dockerfile
# 在Dockerfile中已实现多阶段构建
# 构建阶段使用完整的golang镜像
# 运行阶段使用最小的alpine镜像
```

## 故障排除

### 常见问题和解决方案

#### 1. 容器启动失败

```bash
# 检查容器日志
docker logs ai-image-generator

# 检查配置文件
docker exec ai-image-generator cat /app/config.example.toml

# 检查环境变量
docker exec ai-image-generator env | grep -E "(MODEL_SCOPE|OPENROUTER)"
```

#### 2. API调用失败

```bash
# 测试网络连接
docker exec ai-image-generator ping -c 3 api.openrouter.ai

# 检查API密钥配置
docker exec ai-image-generator env | grep API_KEY
```

#### 3. 图片生成失败

```bash
# 检查存储目录权限
docker exec ai-image-generator ls -la /app/public/static/images

# 检查磁盘空间
docker exec ai-image-generator df -h
```

#### 4. 内存不足

```bash
# 查看内存使用情况
docker stats ai-image-generator

# 增加内存限制
docker update --memory="4g" ai-image-generator
```

## 监控和维护

### 1. 健康检查监控

```bash
# 编写健康检查脚本
#!/bin/bash
HEALTH_URL="http://localhost:8000/health"
if curl -f $HEALTH_URL > /dev/null 2>&1; then
    echo "服务正常"
else
    echo "服务异常，尝试重启容器"
    docker restart ai-image-generator
fi
```

### 2. 日志轮转

```bash
# 配置Docker日志轮转
docker run -d \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  ai-image-generator:latest
```

### 3. 定期清理

```bash
# 清理未使用的镜像
docker image prune -f

# 清理停止的容器
docker container prune -f

# 清理未使用的卷
docker volume prune -f
```

## 生产部署建议

### 1. 使用Docker Compose进行编排

```yaml
# 生产环境docker-compose.yml
version: '3.8'
services:
  ai-image-generator:
    build: .
    restart: always
    ports:
      - "8000:8000"
    environment:
      - LOG_LEVEL=warn
    volumes:
      - ./data:/app/data
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### 2. 使用反向代理

```nginx
# nginx配置示例
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 3. SSL证书配置

```bash
# 使用Let's Encrypt
certbot --nginx -d your-domain.com
```

这个Docker配置为您的AI图像生成项目提供了完整的容器化解决方案，包括开发、测试和生产环境的支持。
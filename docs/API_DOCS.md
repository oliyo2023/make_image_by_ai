# 绘影(huiying) API 文档

## 概述
绘影(huiying)提供完整的图像生成、翻译和记录管理功能，支持中文提示词自动翻译，并将生成记录保存到Cloudflare D1数据库。

## 基础信息
- **Base URL**: `http://localhost:8000`
- **Content-Type**: `application/json`
- **数据存储**: Cloudflare D1 数据库 + 本地/R2存储

## API 端点

### 1. 健康检查
```http
GET /health
```

**响应示例**:
```json
{
  "status": "ok",
  "timestamp": "2024-08-29 14:30:00",
  "version": "1.0.0"
}
```

### 2. 文本翻译
```http
POST /translate
```

**请求体**:
```json
{
  "text": "一只可爱的小猫"
}
```

**响应示例**:
```json
{
  "translated_text": "A cute little kitten",
  "success": true
}
```

### 3. 图像生成
```http
POST /generate-image
```

**请求体**:
```json
{
  "prompt": "一只可爱的小猫",
  "model": "google/gemini-2.5-flash-image-preview:free"
}
```

**响应示例**:
```json
{
  "created": 1693312800,
  "data": [
    {
      "url": "https://your-bucket.r2.cloudflarestorage.com/images/20240829_143000_abc123_cute_little_kitten.jpg"
    }
  ]
}
```

### 4. 图片记录查询
```http
GET /records
```

**查询参数**:
- `page` (int): 页码，默认1
- `limit` (int): 每页数量，默认20，最大100
- `keyword` (string): 关键词搜索（搜索原始和英文提示词）
- `model` (string): 模型筛选
- `date_from` (string): 开始日期 (YYYY-MM-DD)
- `date_to` (string): 结束日期 (YYYY-MM-DD)

**请求示例**:
```http
GET /records?page=1&limit=10&keyword=猫&model=google/gemini-2.5-flash-image-preview:free&date_from=2024-08-01&date_to=2024-08-31
```

**响应示例**:
```json
{
  "records": [
    {
      "id": 1,
      "original_prompt": "一只可爱的小猫",
      "english_prompt": "A cute little kitten",
      "model": "google/gemini-2.5-flash-image-preview:free",
      "image_url": "https://openrouter.ai/generated/image.jpg",
      "local_path": "/static/images/ai_image_20240829_143000_abc123_cute_little_kitten.jpg",
      "r2_url": "https://your-bucket.r2.cloudflarestorage.com/images/20240829_143000_abc123_cute_little_kitten.jpg",
      "file_size": 245760,
      "width": 1024,
      "height": 1024,
      "format": "jpeg",
      "created_at": "2024-08-29 14:30:00",
      "updated_at": "2024-08-29 14:30:00"
    }
  ],
  "total": 45,
  "page": 1,
  "limit": 10,
  "pages": 5
}
```

### 5. 获取单个图片记录
```http
GET /records/{id}
```

**路径参数**:
- `id` (int): 记录ID

**响应示例**:
```json
{
  "id": 1,
  "original_prompt": "一只可爱的小猫",
  "english_prompt": "A cute little kitten",
  "model": "google/gemini-2.5-flash-image-preview:free",
  "image_url": "https://openrouter.ai/generated/image.jpg",
  "local_path": "/static/images/ai_image_20240829_143000_abc123_cute_little_kitten.jpg",
  "r2_url": "https://your-bucket.r2.cloudflarestorage.com/images/20240829_143000_abc123_cute_little_kitten.jpg",
  "file_size": 245760,
  "width": 1024,
  "height": 1024,
  "format": "jpeg",
  "created_at": "2024-08-29 14:30:00",
  "updated_at": "2024-08-29 14:30:00"
}
```

### 6. 本地图片列表（兼容性）
```http
GET /images
```

**响应示例**:
```json
{
  "images": [
    {
      "filename": "ai_image_20240829_143000_abc123_cute_little_kitten.jpg",
      "url": "/static/images/ai_image_20240829_143000_abc123_cute_little_kitten.jpg",
      "created_time": "2024-08-29 14:30:00",
      "size": 245760
    }
  ]
}
```

## 错误响应
所有错误响应都遵循以下格式：
```json
{
  "error": "错误描述信息"
}
```

**常见HTTP状态码**:
- `400`: 请求参数错误
- `404`: 资源不存在
- `500`: 服务器内部错误

## 功能特性

### 1. 智能文件命名
生成的图片文件名格式：`huiying_image_{timestamp}_{uuid}_{english_keywords}.{format}`

示例：`huiying_image_20240829_143000_abc123_cute_little_kitten.jpg`

### 2. 双重存储
- **本地存储**: 保存到 `public/static/images` 目录
- **云存储**: 自动上传到 Cloudflare R2（如果配置）

### 3. 数据库记录
- 自动保存所有生成记录到 Cloudflare D1 数据库
- 包含原始提示词、翻译结果、模型信息、存储路径等
- 支持全文搜索和筛选

### 4. 图片压缩与优化
- 支持JPEG/PNG格式
- 可配置压缩质量和尺寸
- 智能缩放保持宽高比

## 配置要求

### 必需配置
- `MODEL_SCOPE_TOKEN`: ModelScope API令牌
- `OPENROUTER_API_KEY`: OpenRouter API密钥

### 可选配置
- **Cloudflare R2**: 云存储配置
- **Cloudflare D1**: 数据库配置（强烈建议）

详细配置说明请参考 [TOML_CONFIG.md](TOML_CONFIG.md)

## 使用示例

### curl命令示例
```
# 翻译文本
curl -X POST http://localhost:8000/translate \
  -H "Content-Type: application/json" \
  -d '{"text":"一只可爱的小猫"}'

# 生成图像
curl -X POST http://localhost:8000/generate-image \
  -H "Content-Type: application/json" \
  -d '{"prompt":"一只可爱的小猫","model":"google/gemini-2.5-flash-image-preview:free"}'

# 查询图片记录
curl "http://localhost:8000/records?page=1&limit=5&keyword=猫"

# 获取特定记录
curl "http://localhost:8000/records/1"
```

### JavaScript示例
```
// 生成图像
const response = await fetch('http://localhost:8000/generate-image', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    prompt: '一只可爱的小猫',
    model: 'google/gemini-2.5-flash-image-preview:free'
  })
});

const result = await response.json();
console.log('生成的图片URL:', result.data[0].url);

// 查询记录
const recordsResponse = await fetch('http://localhost:8000/records?keyword=猫&limit=10');
const records = await recordsResponse.json();
console.log('找到', records.total, '条记录');
```

## 注意事项

1. **API限制**: 图像生成依赖外部服务，可能有速率限制
2. **存储空间**: 本地存储需要足够的磁盘空间
3. **数据库**: D1数据库有免费额度限制，超出需付费
4. **文件命名**: 使用英文关键词命名，避免中文字符问题
5. **并发**: 支持多个并发请求，但受到外部API限制

## 常见问题

**Q: 为什么图片记录查询返回空？**
A: 检查D1数据库配置是否正确，确保API令牌有足够权限。

**Q: 图片无法访问？**
A: 确保静态文件目录权限正确，或检查R2存储配置。

**Q: 翻译失败？**
A: 检查ModelScope API令牌是否有效，网络连接是否正常。
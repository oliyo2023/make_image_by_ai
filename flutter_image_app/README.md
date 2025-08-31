# Flutter AI Image Generator App

这是一个使用 Flutter 开发的 AI 图像生成应用，可以将中文提示词翻译成英文并生成相应的图像。

## 功能特性

1. 中文提示词自动翻译成英文
2. AI 图像生成
3. 历史记录查看
4. 清除输入内容

## 配置要求

在运行应用之前，请确保后端服务已经启动并正确配置。

### 后端配置

1. 复制 `config.example.toml` 文件为 `config.toml`
2. 修改 `config.toml` 文件中的配置：
   - `model_scope_token`: ModelScope API 令牌
   - `openrouter_api_key`: OpenRouter API 密钥

### 环境变量

确保后端服务运行在 `http://localhost:8000`，或者修改 `lib/api_service.dart` 中的 `_baseUrl` 常量。

## 安装和运行

1. 确保已安装 Flutter SDK
2. 进入项目目录：
   ```
   cd flutter_image_app
   ```
3. 安装依赖：
   ```
   flutter pub get
   ```

## 运行 Web 版本

```
flutter run -d chrome
```

或者构建生产版本：

```
flutter build web
```

## 运行 Windows 版本

```
flutter run -d windows
```

或者构建生产版本：

```
flutter build windows
```

## 功能说明

### 主页功能

- **中文提示词输入框**：输入中文描述，应用会自动翻译成英文
- **英文提示词输入框**：可直接输入英文提示词，或查看翻译结果
- **生成图片按钮**：根据提示词生成图像
- **清除按钮**：清空所有输入内容和生成的图像
- **查看历史记录按钮**：查看之前生成的图像记录

### 历史记录页面

- 显示之前生成的图像记录
- 包括原始中文提示词、翻译后的英文提示词和生成时间
- 分页浏览功能

## 依赖说明

- `http`: 用于网络请求

## 支持平台

- Web
- Windows

## 注意事项

1. 确保设备已连接网络以使用翻译和图像生成功能
2. 后端服务必须运行在 `http://localhost:8000` 或相应配置的地址
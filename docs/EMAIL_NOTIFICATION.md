# 绘影(huiying)邮件通知功能使用指南

## 功能概述

绘影(huiying)支持在图像生成任务完成后自动发送邮件通知。该功能使用SMTP协议，支持QQ邮箱等主流邮件服务商。

## 配置方法

### 1. 通过配置文件配置

在 `config.toml` 文件中添加以下配置：

```toml
[smtp]
# SMTP 邮件配置 (QQ邮箱示例)
host = "smtp.qq.com"
port = 587
username = "your_qq_email@qq.com"
password = "your_qq_email_auth_code"  # QQ邮箱需要使用授权码而非密码
from = "your_qq_email@qq.com"
to = "recipient@example.com"
enable = true  # 设置为true以启用邮件通知功能
```

### 2. 通过环境变量配置

```bash
export SMTP_HOST="smtp.qq.com"
export SMTP_PORT="587"
export SMTP_USERNAME="your_qq_email@qq.com"
export SMTP_PASSWORD="your_qq_email_auth_code"
export SMTP_FROM="your_qq_email@qq.com"
export SMTP_TO="recipient@example.com"
export SMTP_ENABLE="true"
```

## QQ邮箱配置详细步骤

### 1. 登录QQ邮箱并开启SMTP服务

1. 登录您的QQ邮箱
2. 点击右上角的"设置"
3. 选择"账户"选项卡
4. 找到"POP3/SMTP服务"，点击"开启"
5. 根据提示发送短信验证
6. 验证成功后，会显示授权码

### 2. 获取授权码

1. 在"账户"设置页面，找到"POP3/SMTP服务"
2. 点击"生成授权码"
3. 通过短信验证后，系统会生成一个16位的授权码
4. 请妥善保管此授权码，它将作为SMTP密码使用

### 3. 配置参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| host | SMTP服务器地址 | smtp.qq.com |
| port | SMTP服务器端口 | 587 |
| username | 发件人邮箱地址 | your_qq_email@qq.com |
| password | 邮箱授权码 | xxxxxxxxxxxxxxxx |
| from | 发件人邮箱地址 | your_qq_email@qq.com |
| to | 收件人邮箱地址 | recipient@example.com |
| enable | 是否启用邮件通知 | true |

## 其他邮箱服务商配置

### Gmail

```toml
[smtp]
host = "smtp.gmail.com"
port = 587
username = "your_gmail@gmail.com"
password = "your_gmail_app_password"  # 需要使用应用专用密码
from = "your_gmail@gmail.com"
to = "recipient@example.com"
enable = true
```

### 163邮箱

```toml
[smtp]
host = "smtp.163.com"
port = 25
username = "your_163_email@163.com"
password = "your_163_email_auth_code"
from = "your_163_email@163.com"
to = "recipient@example.com"
enable = true
```

## 功能特点

### 1. 异步发送

邮件发送采用异步方式，不会阻塞图像生成的主流程，确保系统响应速度。

### 2. 双类型通知

- **成功通知**：图像生成成功后发送包含图像信息和链接的邮件
- **错误通知**：图像生成失败时发送包含错误信息的邮件

### 3. HTML格式邮件

邮件采用HTML格式，包含：
- 图像生成的基本信息（提示词、尺寸、格式等）
- 图像访问链接（R2链接或本地链接）
- 美观的排版和格式

### 4. 灵活配置

- 可通过配置文件或环境变量进行配置
- 可随时启用或禁用邮件通知功能
- 支持多种主流邮件服务商

## 测试邮件功能

配置完成后，您可以发送测试邮件来验证配置是否正确：

```bash
# 使用curl测试API
curl -X POST http://localhost:8000/generate-image \
  -H "Content-Type: application/json" \
  -d '{"prompt": "测试绘影邮件通知功能", "model": "google/gemini-2.5-flash-image-preview:free"}'
```

如果配置正确，您将在图像生成完成后收到一封通知邮件。

## 常见问题

### 1. 邮件发送失败

- 检查SMTP配置是否正确
- 确认授权码是否正确
- 检查网络连接是否正常
- 查看系统日志中的错误信息

### 2. 收不到邮件

- 检查垃圾邮件箱
- 确认收件人邮箱地址是否正确
- 检查邮箱服务商的限制策略

### 3. 邮件内容不完整

- 检查系统日志是否有错误信息
- 确认图像记录是否正确保存到数据库

## 安全建议

1. **保护授权码**：不要将授权码提交到代码仓库
2. **使用环境变量**：推荐使用环境变量而非配置文件存储敏感信息
3. **定期更换授权码**：定期更换邮箱授权码以提高安全性
4. **限制收件人**：只发送邮件给授权的收件人
# 🔒 安全配置指南

本文档提供了如何安全配置 AI 图像生成服务的详细指导。

## ⚠️ 重要安全原则

### 1. 绝对禁止的行为
- ❌ **禁止将 API 密钥硬编码在源代码中**
- ❌ **禁止将包含真实密钥的配置文件提交到版本控制**
- ❌ **禁止在日志中输出 API 密钥**
- ❌ **禁止在错误信息中暴露密钥**

### 2. 推荐的安全实践
- ✅ **使用环境变量存储敏感信息**
- ✅ **使用配置文件模板，真实配置文件加入 .gitignore**
- ✅ **为不同环境使用不同的密钥**
- ✅ **定期轮换 API 密钥**
- ✅ **监控 API 密钥使用情况**

## 🛡️ 配置方法

### 方法 1: 环境变量 (推荐)

```bash
# Linux/macOS
export MODEL_SCOPE_TOKEN="your-actual-token"
export OPENROUTER_API_KEY="your-actual-key"

# Windows PowerShell
$env:MODEL_SCOPE_TOKEN="your-actual-token"
$env:OPENROUTER_API_KEY="your-actual-key"

# Windows CMD
set MODEL_SCOPE_TOKEN=your-actual-token
set OPENROUTER_API_KEY=your-actual-key
```

### 方法 2: .env 文件

1. 复制示例文件：
```bash
cp .env.example .env
```

2. 编辑 `.env` 文件：
```env
MODEL_SCOPE_TOKEN=ms-your-actual-token
OPENROUTER_API_KEY=sk-or-v1-your-actual-key
```

3. 确保 `.env` 文件已在 `.gitignore` 中：
```gitignore
*.env
*.env.*
!*.env.example
```

### 方法 3: TOML 配置文件

1. 复制示例文件：
```bash
cp config.example.toml config.toml
```

2. 编辑 `config.toml` 文件：
```toml
[api_keys]
model_scope_token = "ms-your-actual-token"
openrouter_api_key = "sk-or-v1-your-actual-key"
```

3. 确保 `config.toml` 文件已在 `.gitignore` 中：
```gitignore
config.toml
!config.example.toml
```

## 🔑 API 密钥获取指南

### ModelScope API Token
1. 访问 [ModelScope 控制台](https://www.modelscope.cn/)
2. 登录您的账户
3. 进入 "API Token" 页面
4. 生成新的 Token
5. 格式：`ms-xxxxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

### OpenRouter API Key
1. 访问 [OpenRouter 控制台](https://openrouter.ai/)
2. 注册并登录账户
3. 进入 "API Keys" 页面
4. 创建新的 API Key
5. 格式：`sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

### Cloudflare API Token
1. 访问 [Cloudflare 控制台](https://dash.cloudflare.com/)
2. 进入 "My Profile" > "API Tokens"
3. 创建自定义 Token，权限包括：
   - Zone:Zone:Read
   - Zone:DNS:Edit
   - Account:Cloudflare D1:Edit
   - Account:Cloudflare R2:Edit

## 🚨 密钥泄露应急处理

### 如果密钥意外泄露：

1. **立即撤销泄露的密钥**
   - ModelScope: 登录控制台删除 Token
   - OpenRouter: 登录控制台撤销 API Key
   - Cloudflare: 删除泄露的 API Token

2. **生成新的密钥**
   - 按照上述指南重新生成
   - 更新本地配置

3. **检查使用记录**
   - 查看 API 使用日志
   - 检查是否有异常调用

4. **修复代码**
   - 移除硬编码的密钥
   - 提交安全修复
   - 强制推送覆盖历史记录（如必要）

```bash
# 移除 Git 历史中的敏感信息（谨慎使用）
git filter-branch --force --index-filter \
'git rm --cached --ignore-unmatch config/config.go' \
--prune-empty --tag-name-filter cat -- --all

# 强制推送到远程仓库
git push origin --force --all
```

## 🔍 安全检查清单

### 开发环境
- [ ] 配置文件不包含真实密钥
- [ ] .gitignore 包含所有敏感文件
- [ ] 环境变量正确设置
- [ ] 本地测试正常工作

### 代码审查
- [ ] 源代码中无硬编码密钥
- [ ] 日志输出不包含敏感信息
- [ ] 错误处理不暴露密钥
- [ ] 配置加载逻辑正确

### 部署前检查
- [ ] 生产环境变量配置完成
- [ ] API 密钥权限最小化
- [ ] 监控和告警设置
- [ ] 备份和恢复计划

## 📚 相关资源

- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security)
- [Cloudflare Security Documentation](https://developers.cloudflare.com/security/)

---

**记住：安全是一个持续的过程，而不是一次性的任务！**
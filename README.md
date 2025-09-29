# Novel API - AI 图像生成服务

一个基于 Go 语言开发的 AI 图像生成 API 服务，支持 NovelAI Diffusion 模型（v3 和 v4），集成智能翻译功能和多种云存储服务。

## 🚀 功能特性

### 🎨 AI 图像生成
- **支持多个 NovelAI 模型**：
  - `nai-diffusion-3` - NAI Diffusion 3.0
  - `nai-diffusion-furry-3` - NAI Diffusion 3.0 兽人版
  - `nai-diffusion-4-full` - NAI Diffusion 4.0 完整版
  - `nai-diffusion-4-curated-preview` - NAI Diffusion 4.0 精选预览版
  - `nai-diffusion-4-5-curated` - NAI Diffusion 4.5 精选版
  - `nai-diffusion-4-5-full` - NAI Diffusion 4.5 完整版

### 🌐 智能翻译
- **AI 翻译服务**：自动将中文提示词翻译为英文
- **专业提示词优化**：内置 NovelAI 专用提示词优化系统
- **可配置开关**：支持启用/禁用翻译功能

### ☁️ 多云存储支持
- **腾讯云 COS**：支持腾讯云对象存储
- **MinIO**：支持自建或第三方 MinIO 服务
- **Alist**：支持 Alist 网盘聚合服务

### ⚙️ 高度可配置
- **参数可调**：支持图像尺寸、采样器、步数等全参数配置
- **YAML 配置**：使用 `.env` 文件进行集中配置管理
- **多环境支持**：易于部署到不同环境

## 📦 项目结构

```
novel-api重构/
├── main.go                    # 主程序入口
├── go.mod                     # Go 模块依赖管理
├── go.sum                     # 依赖校验文件
├── .env                       # 配置文件
├── api/                       # API 处理模块
│   ├── api_completions.go     # 主要的 API 处理逻辑
│   ├── api_translation.go     # AI 翻译服务
│   ├── api_images.go          # 图像处理工具
│   └── api_config.go          # 配置相关 API
├── config/                    # 配置结构定义
│   └── config.go              # 配置结构体定义
├── models/                    # AI 模型实现
│   ├── nai-diffusion-v3.go    # NAI Diffusion 3.0 实现
│   └── nai-diffusion-v4.go    # NAI Diffusion 4.0 实现
└── upload/                    # 文件上传模块
    ├── uploader.go            # 通用上传接口
    ├── tengxun_cos.go         # 腾讯云 COS 上传器
    ├── minio_cos.go           # MinIO 上传器
    └── alist_cos.go           # Alist 上传器
```

## 🛠️ 安装与配置

### 系统要求
- Go 1.23.0 或更高版本
- 网络连接（用于访问 NovelAI API 和云存储服务）

### 安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd novel-api-go
```

2. **安装依赖**
```bash
go mod tidy
```

3. **配置服务**

复制并编辑 `.env` 配置文件：

```yaml
# 启动端口号
server:
  addr: 3388

# 存储桶选择 (Tengxun/Minio/Alist)
cos:
  backet: Tengxun

# 翻译 API 配置
translation:
  enable: true
  url: https://api.codesphere.chat
  key: sk-your-api-key-here
  model: gpt-4.1-nano
  role: "专业的 NovelAI 提示词翻译专家..."

# 腾讯云 COS 配置
tencent_cos:
  secret_id: "your-secret-id"
  secret_key: "your-secret-key"
  region: "ap-guangzhou"
  bucket: "your-bucket-name"
  base_url: "https://your-bucket.cos.ap-guangzhou.myqcloud.com"

# MinIO 配置
minio:
  endpoint: "your-minio-endpoint.com"
  access_key_id: "your-access-key"
  secret_access_key: "your-secret-key"
  bucket_name: "novel"
  use_ssl: true
  base_url: "https://your-minio-endpoint.com"

# Alist 配置
alist:
  base_url: "http://127.0.0.1:5244"
  token: "your-alist-token"
  path: "/nai"
  username: "admin"
  password: "your-password"

# 图像生成参数
parameters:
  width: 832
  height: 1216
  scale: 5.5
  sampler: "k_euler_ancestral"
  steps: 28
  # ... 更多参数配置
```

4. **启动服务**
```bash
go run main.go
```

服务将在 `http://localhost:3388` 启动。

## 🔧 API 使用

### 图像生成 API

**请求地址**：`POST /v1/chat/completions`

**请求头**：
```
Authorization: Bearer your-novel-ai-token
Content-Type: application/json
```

**请求体**：
```json
{
  "model": "nai-diffusion-4-curated-preview",
  "messages": [
    {
      "role": "user", 
      "content": "一个美丽的女孩，蓝色眼睛，短发"
    }
  ]
}
```

**响应**：
```json
{
  "id": "chatcmpl-xxxxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "nai-diffusion-4-curated-preview",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "https://your-storage.com/path/to/generated-image.png"
      },
      "finish_reason": "stop"
    }
  ]
}
```

### 支持的模型

| 模型名称 | 描述 | 版本 |
|---------|------|------|
| `nai-diffusion-3` | NAI Diffusion 3.0 标准版 | v3 |
| `nai-diffusion-furry-3` | NAI Diffusion 3.0 兽人版 | v3 |
| `nai-diffusion-4-full` | NAI Diffusion 4.0 完整版 | v4 |
| `nai-diffusion-4-curated-preview` | NAI Diffusion 4.0 精选预览版 | v4 |
| `nai-diffusion-4-5-curated` | NAI Diffusion 4.5 精选版 | v4 |
| `nai-diffusion-4-5-full` | NAI Diffusion 4.5 完整版 | v4 |

## ⚡ 核心功能

### 智能翻译系统

当启用翻译功能时，系统会自动：
1. 检测用户输入是否为中文
2. 调用配置的 AI 翻译服务
3. 将中文描述转换为专业的 NovelAI 英文提示词
4. 使用翻译后的提示词生成图像

**翻译示例**：
- 输入：`"一个穿着白色长裙的天使"`
- 输出：`"{1girl},angel,white dress,{detailed eyes},{shine golden eyes},halo,{white wings}"`

### 图像处理流程

1. **请求解析**：解析 OpenAI 兼容的请求格式
2. **翻译处理**：可选的中文到英文提示词翻译
3. **模型路由**：根据模型名称选择对应的处理器
4. **图像生成**：调用 NovelAI API 生成图像
5. **文件上传**：将生成的图像上传到配置的云存储
6. **响应返回**：返回图像访问链接

### 存储服务集成

**腾讯云 COS**：
- 企业级对象存储服务
- 支持 CDN 加速
- 高可用性和安全性

**MinIO**：
- 兼容 Amazon S3 的对象存储
- 支持私有云部署
- 高性能存储解决方案

**Alist**：
- 支持多种网盘服务
- 统一的文件管理界面
- 适合个人和小团队使用

## 🔧 配置说明

### 服务器配置
- `server.addr`：服务监听端口

### 存储配置
- `cos.backet`：选择存储服务类型（Tengxun/Minio/Alist）

### 翻译配置
- `translation.enable`：是否启用翻译功能
- `translation.url`：翻译 API 地址
- `translation.key`：翻译 API 密钥
- `translation.model`：使用的翻译模型
- `translation.role`：翻译提示词模板

### 图像参数配置
- `parameters.width/height`：图像尺寸
- `parameters.scale`：生成比例（0.1-10.0）
- `parameters.sampler`：采样器类型
- `parameters.steps`：生成步数（1-50）
- `parameters.n_samples`：生成图像数量

## 🚀 部署指南

### 本地开发部署
```bash
# 开发模式启动
go run main.go

# 构建二进制文件
go build -o novel-api main.go
./novel-api
```

### Docker 部署
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o novel-api main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/novel-api .
COPY --from=builder /app/.env .
CMD ["./novel-api"]
```

### 生产环境建议
1. 使用反向代理（Nginx）
2. 配置 HTTPS 证书
3. 设置进程管理（systemd/supervisor）
4. 配置日志轮转
5. 监控服务状态

## 📊 监控与日志

### 日志级别
- **INFO**：正常操作日志
- **ERROR**：错误信息
- **DEBUG**：调试信息

### 关键监控指标
- API 请求响应时间
- 图像生成成功率
- 存储服务上传状态
- 翻译服务调用状态

## 🛡️ 安全考虑

1. **API 密钥管理**：
   - 使用环境变量存储敏感信息
   - 定期轮换 API 密钥
   - 限制密钥访问权限

2. **访问控制**：
   - 实施速率限制
   - IP 白名单机制
   - 用户认证授权

3. **数据安全**：
   - HTTPS 传输加密
   - 存储服务访问控制
   - 定期安全审计

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

### 开发规范
1. 遵循 Go 语言编码规范
2. 添加适当的注释和文档
3. 编写单元测试
4. 保持代码简洁和可维护

### 提交流程
1. Fork 项目
2. 创建功能分支
3. 提交代码变更
4. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 💡 常见问题

**Q: 如何切换不同的存储服务？**  
A: 修改 `.env` 文件中的 `cos.backet` 配置，支持 `Tengxun`、`Minio`、`Alist`。

**Q: 翻译功能不工作怎么办？**  
A: 检查 `translation.enable` 是否为 `true`，确认 API 地址和密钥配置正确。

**Q: 生成的图像质量不满意？**  
A: 调整 `parameters` 部分的配置，如增加 `steps` 数值、调整 `scale` 比例等。

**Q: 支持哪些图像格式？**  
A: 目前支持 PNG 和 JPEG 格式的图像生成和上传。

---

🎨 **Happy Creating with Novel API!** 🎨
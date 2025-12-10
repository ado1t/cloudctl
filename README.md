# cloudctl

一个用于管理多云平台资源的命令行工具，支持 Cloudflare 和 AWS。

## 功能特性

- 🌐 **Cloudflare 管理**
  - Zone（域名）管理
  - DNS 记录管理
  - 缓存清除

- ☁️ **AWS 管理**
  - CloudFront CDN 管理
  - CloudFront 缓存失效
  - ACM 证书管理

- 🎨 **用户友好**
  - 彩色输出
  - 表格和 JSON 格式支持
  - 详细的日志级别控制（`-v/-vv/-vvv/-vvvv`）
  - 多账号 Profile 支持

## 快速开始

### 安装

```bash
# 克隆项目
git clone https://github.com/panda2xx/cloudctl.git
cd cloudctl

# 编译
make build

# 或者直接安装到 GOPATH/bin
make install
```

### 配置

创建配置文件 `~/.cloudctl/config.yaml`：

```yaml
# 默认 profile
default_profile:
  cloudflare: cf-prod
  aws: aws-prod

# Cloudflare 配置
cloudflare:
  cf-prod:
    api_token: ${CLOUDFLARE_API_TOKEN}
    email: user@example.com

# AWS 配置
aws:
  aws-prod:
    access_key_id: ${AWS_ACCESS_KEY_ID}
    secret_access_key: ${AWS_SECRET_ACCESS_KEY}
    region: us-east-1
```

参考 `conf/config.example.yaml` 获取完整配置示例。

### 使用示例

#### Cloudflare Zone 管理

```bash
# 列出所有域名
cloudctl cf zone list

# 创建域名
cloudctl cf zone create example.com
```

#### Cloudflare DNS 管理

```bash
# 列出 DNS 记录
cloudctl cf dns list example.com
cloudctl cf dns list example.com --type A

# 创建 DNS 记录
cloudctl cf dns create example.com -t A -n www --content 1.2.3.4
cloudctl cf dns create example.com -t A -n api --content 1.2.3.4 --proxied
cloudctl cf dns create example.com -t CNAME -n blog --content example.com

# 更新 DNS 记录
cloudctl cf dns update example.com <record-id> --content 2.3.4.5
cloudctl cf dns update example.com <record-id> --ttl 3600 --proxied

# 删除 DNS 记录
cloudctl cf dns delete example.com <record-id>

# 批量创建 DNS 记录
cloudctl cf dns batch-create --config dns-records.yaml

# 预览批量操作（不实际执行）
cloudctl cf dns batch-create --config dns-records.yaml --dry-run

# 使用并发加速
cloudctl cf dns batch-create --config dns-records.yaml --concurrency 3
```

#### Cloudflare 缓存管理

```bash
# 清除所有缓存
cloudctl cf cache purge example.com --purge-all

# 清除指定文件
cloudctl cf cache purge example.com --files /path/to/file.html
cloudctl cf cache purge example.com -f /css/style.css,/js/app.js

# 清除指定目录/前缀
cloudctl cf cache purge example.com --prefixes /foo/bar/
cloudctl cf cache purge example.com --prefixes /images/,/static/

# 清除指定主机名
cloudctl cf cache purge example.com --hosts www.example.com,api.example.com

# 清除指定标签（企业版）
cloudctl cf cache purge example.com --tags tag1,tag2
```

#### AWS CloudFront 管理

```bash
# 列出 CloudFront 分发
cloudctl aws cdn list

# 创建缓存失效
cloudctl aws cdn invalidate E1234567890ABC --paths "/index.html,/images/*"
```

#### AWS ACM 证书管理

```bash
# 申请证书
cloudctl aws cert request --domain example.com --san "*.example.com,www.example.com"
```

#### 通用选项

```bash

# 使用不同的 profile
cloudctl cf zone list --profile cf-dev

# 调整日志详细程度
cloudctl cf zone list -v         # WARN 级别
cloudctl cf zone list -vv        # INFO 级别
cloudctl cf zone list -vvv       # DEBUG 级别（最详细，包含 API 请求/响应）

# JSON 格式输出
cloudctl cf zone list --output json
```

## 开发

### 环境要求

- Go 1.25+
- Make

### 开发命令

```bash
# 查看所有可用命令
make help

# 开发模式运行
make dev

# 运行测试
make test

# 代码检查
make check

# 格式化代码
make fmt
```

### 项目结构

```
cloudctl/
├── cmd/cloudctl/       # 主程序入口
├── internal/           # 内部包
│   ├── cloudflare/    # Cloudflare 实现
│   ├── aws/           # AWS 实现
│   ├── config/        # 配置管理
│   └── output/        # 输出格式化
├── pkg/               # 公共包
├── conf/              # 配置示例
├── doc/               # 文档
└── Makefile           # 构建脚本
```

## 文档

详细文档请查看：
- [需求说明](doc/需求说明.md) - 功能需求和设计说明
- [项目规划](doc/项目规划.md) - 分阶段实现计划（10周）

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

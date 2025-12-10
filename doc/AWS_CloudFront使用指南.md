# AWS CloudFront 使用指南

本文档介绍如何使用 cloudctl 管理 AWS CloudFront CDN 分发。

## 前置条件

### 1. 配置 AWS 凭证

在配置文件 `~/.cloudctl/config.yaml` 中添加 AWS 凭证:

```yaml
aws:
  aws-prod:
    access_key_id: ${AWS_ACCESS_KEY_ID}
    secret_access_key: ${AWS_SECRET_ACCESS_KEY}
    region: us-east-1  # CloudFront 使用 us-east-1
```

或者通过环境变量设置:

```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"
```

### 2. IAM 权限要求

确保 AWS IAM 用户具有以下权限:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudfront:ListDistributions",
        "cloudfront:GetDistribution",
        "cloudfront:CreateDistribution",
        "cloudfront:UpdateDistribution",
        "cloudfront:GetDistributionConfig"
      ],
      "Resource": "*"
    }
  ]
}
```

## 命令使用

### 列出所有分发

列出 AWS 账户中的所有 CloudFront 分发:

```bash
# 使用默认 profile
cloudctl aws cdn list

# 使用指定 profile
cloudctl aws cdn list -p aws-prod

# JSON 格式输出
cloudctl aws cdn list -o json
```

**输出示例**:

```
id              domain_name                          status      enabled  aliases  origins
E1234567890ABC  d111111abcdef8.cloudfront.net       Deployed    true     -        example.com
E0987654321XYZ  d222222abcdef8.cloudfront.net       Deployed    true     cdn.example.com  origin.example.com
```

### 获取分发详情

获取指定 CloudFront 分发的详细信息:

```bash
# 获取分发详情
cloudctl aws cdn get E1234567890ABC

# JSON 格式输出
cloudctl aws cdn get E1234567890ABC -o json
```

**输出示例**:

```json
{
  "id": "E1234567890ABC",
  "domain_name": "d111111abcdef8.cloudfront.net",
  "status": "Deployed",
  "enabled": true,
  "aliases": ["cdn.example.com"],
  "origins": [
    {
      "ID": "origin-1",
      "DomainName": "example.com",
      "Path": ""
    }
  ],
  "comment": "My CDN Distribution"
}
```

### 创建分发

创建新的 CloudFront 分发。

#### 基本创建

创建最简单的分发,只需指定源站:

```bash
cloudctl aws cdn create --origin example.com
```

这将创建一个:
- 启用的分发
- 使用 CloudFront 默认域名
- 使用 CloudFront 默认证书
- 价格等级为 PriceClass_100 (仅北美和欧洲)

#### 带自定义域名

使用自定义域名需要配置 SSL 证书:

```bash
cloudctl aws cdn create \
  --origin example.com \
  --aliases cdn.example.com \
  --certificate-arn arn:aws:acm:us-east-1:123456789012:certificate/xxx
```

**注意**: 
- SSL 证书必须在 ACM (AWS Certificate Manager) 中申请
- 证书必须在 `us-east-1` 区域
- 证书必须包含自定义域名

#### 完整配置

创建带有完整配置的分发:

```bash
cloudctl aws cdn create \
  --origin example.com \
  --aliases cdn.example.com,www.example.com \
  --comment "Production CDN" \
  --certificate-arn arn:aws:acm:us-east-1:123456789012:certificate/xxx \
  --price-class PriceClass_All \
  --default-root-object index.html
```

**参数说明**:

- `--origin`: 源站域名 (必需)
- `--aliases` / `-a`: 自定义域名,多个用逗号分隔
- `--comment`: 备注说明
- `--enabled`: 是否启用分发 (默认: true)
- `--certificate-arn`: SSL 证书 ARN
- `--price-class`: 价格等级
  - `PriceClass_100`: 北美和欧洲
  - `PriceClass_200`: 北美、欧洲、亚洲、中东和非洲
  - `PriceClass_All`: 所有边缘节点
- `--default-root-object`: 默认根对象 (如: index.html)

### 更新分发

更新现有 CloudFront 分发的配置:

```bash
# 更新备注
cloudctl aws cdn update E1234567890ABC --comment "Updated comment"

# 启用分发
cloudctl aws cdn update E1234567890ABC --enabled

# 禁用分发
cloudctl aws cdn update E1234567890ABC --disabled
```

**注意**: 
- 配置更新需要时间才能生效
- 分发状态会变为 "InProgress"
- 部署通常需要 10-15 分钟

## 使用场景

### 场景 1: 为网站创建 CDN

```bash
# 1. 创建分发
cloudctl aws cdn create \
  --origin www.example.com \
  --aliases cdn.example.com \
  --certificate-arn arn:aws:acm:us-east-1:123456789012:certificate/xxx \
  --default-root-object index.html

# 2. 等待部署完成 (10-15 分钟)
cloudctl aws cdn get E1234567890ABC

# 3. 配置 DNS
# 在 DNS 提供商处添加 CNAME 记录:
# cdn.example.com -> d111111abcdef8.cloudfront.net
```

### 场景 2: 为静态资源创建 CDN

```bash
# 创建用于静态资源的分发
cloudctl aws cdn create \
  --origin static.example.com \
  --aliases assets.example.com \
  --certificate-arn arn:aws:acm:us-east-1:123456789012:certificate/xxx \
  --comment "Static Assets CDN"
```

### 场景 3: 临时禁用分发

```bash
# 禁用分发 (不删除)
cloudctl aws cdn update E1234567890ABC --disabled

# 稍后重新启用
cloudctl aws cdn update E1234567890ABC --enabled
```

## 常见问题

### Q: 创建分发后多久可以使用?

A: CloudFront 分发的部署通常需要 10-15 分钟。可以使用 `cloudctl aws cdn get <id>` 查看状态,当 status 为 "Deployed" 时即可使用。

### Q: 如何使用自定义域名?

A: 需要:
1. 在 ACM 中申请 SSL 证书 (必须在 us-east-1 区域)
2. 创建分发时指定 `--aliases` 和 `--certificate-arn`
3. 在 DNS 提供商处添加 CNAME 记录指向 CloudFront 域名

### Q: 价格等级如何选择?

A: 
- `PriceClass_100`: 最便宜,仅北美和欧洲,适合主要用户在这些地区
- `PriceClass_200`: 中等价格,覆盖更多地区
- `PriceClass_All`: 最贵,全球所有边缘节点,最佳性能

### Q: 如何查看我的分发列表?

A: 使用 `cloudctl aws cdn list` 命令。

### Q: 更新配置后多久生效?

A: 配置更新通常需要 10-15 分钟才能部署到所有边缘节点。

## 最佳实践

1. **使用有意义的备注**: 使用 `--comment` 参数添加描述性备注,便于管理多个分发

2. **选择合适的价格等级**: 根据用户地理分布选择价格等级,避免不必要的成本

3. **使用 HTTPS**: 始终为自定义域名配置 SSL 证书

4. **监控分发状态**: 创建或更新后使用 `get` 命令监控部署状态

5. **配置默认根对象**: 对于网站,设置 `--default-root-object index.html`

### 创建缓存失效

清除 CloudFront 边缘节点上的缓存内容。

#### 失效单个路径

清除所有缓存:

```bash
cloudctl aws cdn invalidate E1234567890ABC --paths "/*"
```

清除特定文件:

```bash
cloudctl aws cdn invalidate E1234567890ABC --paths "/index.html"
```

#### 失效多个路径

清除多个路径的缓存:

```bash
cloudctl aws cdn invalidate E1234567890ABC --paths "/index.html,/css/*,/js/*"
```

或者:

```bash
cloudctl aws cdn invalidate E1234567890ABC \
  --paths "/images/*" \
  --paths "/css/*" \
  --paths "/js/*"
```

#### 指定 Caller Reference

使用自定义的调用者引用:

```bash
cloudctl aws cdn invalidate E1234567890ABC \
  --paths "/*" \
  --caller-reference "deployment-2024-001"
```

**注意**: 
- 如果不指定 `--caller-reference`，系统会自动生成一个基于时间戳的引用
- 相同的 caller reference 不能重复使用

**输出示例**:

```json
{
  "id": "I2J3K4L5M6N7O8P9Q0",
  "status": "InProgress",
  "create_time": "2024-11-24 14:30:00",
  "caller_reference": "cloudctl-1700832600",
  "paths": ["/*"]
}

✓ 缓存失效请求已创建
失效 ID: I2J3K4L5M6N7O8P9Q0
状态: InProgress

注意: 缓存失效通常需要 10-15 分钟才能完成
```

#### 查询失效状态

查询缓存失效请求的状态:

```bash
# 查询失效状态
cloudctl aws cdn invalidate-status E1234567890ABC I2J3K4L5M6N7O8P9Q0

# JSON 格式输出
cloudctl aws cdn invalidate-status E1234567890ABC I2J3K4L5M6N7O8P9Q0 -o json
```

**输出示例**:

```json
{
  "id": "I2J3K4L5M6N7O8P9Q0",
  "status": "Completed",
  "create_time": "2024-11-24 14:30:00",
  "caller_reference": "cloudctl-1700832600",
  "paths": ["/*"]
}

✓ 缓存失效已完成
```

## 使用场景

### 场景 4: 部署后清除缓存

在部署新版本后清除 CDN 缓存:

```bash
# 1. 部署新版本到源站
# ...

# 2. 清除 CloudFront 缓存
cloudctl aws cdn invalidate E1234567890ABC --paths "/*"

# 3. 等待缓存失效完成
# 通常需要 10-15 分钟
```

### 场景 5: 清除特定资源缓存

只清除特定类型文件的缓存:

```bash
# 清除所有 CSS 和 JS 文件
cloudctl aws cdn invalidate E1234567890ABC --paths "/css/*,/js/*"

# 清除所有图片
cloudctl aws cdn invalidate E1234567890ABC --paths "/images/*"
```

## 常见问题

### Q: 缓存失效需要多久?

A: CloudFront 缓存失效通常需要 10-15 分钟才能完成。可以使用 `cloudctl aws cdn invalidate-status` 命令查看失效状态来确认是否完成。

### Q: 缓存失效有费用吗?

A: AWS 每月提供 1000 次免费的缓存失效路径。超过部分每个路径收费 $0.005。使用通配符 `/*` 算作一个路径。

### Q: 如何清除所有缓存?

A: 使用路径 `/*` 可以清除所有缓存:
```bash
cloudctl aws cdn invalidate E1234567890ABC --paths "/*"
```

### Q: 可以同时失效多少个路径?

A: 单次请求最多可以失效 3000 个路径。建议使用通配符来减少路径数量。

### Q: 失效请求失败了怎么办?

A: 检查:
1. Distribution ID 是否正确
2. 路径格式是否正确（必须以 `/` 开头）
3. AWS 凭证是否有效
4. 是否有足够的权限

## 相关命令

- `cloudctl aws cdn list` - 列出所有分发
- `cloudctl aws cdn get` - 获取分发详情
- `cloudctl aws cdn create` - 创建分发
- `cloudctl aws cdn update` - 更新分发
- `cloudctl aws cdn invalidate` - 创建缓存失效
- `cloudctl aws cdn invalidate-status` - 查询失效状态
- `cloudctl aws cert list` - 列出 ACM 证书
- `cloudctl aws cert request` - 申请新证书
- `cloudctl cf dns create` - 创建 DNS 记录

## 参考资料

- [AWS CloudFront 文档](https://docs.aws.amazon.com/cloudfront/)
- [CloudFront 缓存失效](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Invalidation.html)
- [ACM 证书管理](https://docs.aws.amazon.com/acm/)
- [CloudFront 价格](https://aws.amazon.com/cloudfront/pricing/)

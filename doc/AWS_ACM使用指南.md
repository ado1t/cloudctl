# AWS ACM 证书管理使用指南

本文档介绍如何使用 `cloudctl` 管理 AWS ACM (AWS Certificate Manager) 证书。

## 目录

- [前置要求](#前置要求)
- [配置](#配置)
- [基本操作](#基本操作)
  - [列出证书](#列出证书)
  - [获取证书详情](#获取证书详情)
  - [申请新证书](#申请新证书)
- [使用场景](#使用场景)
- [常见问题](#常见问题)
- [相关命令](#相关命令)

## 前置要求

1. **AWS 账户**: 需要有效的 AWS 账户
2. **IAM 权限**: 需要以下 ACM 权限:
   - `acm:ListCertificates`
   - `acm:DescribeCertificate`
   - `acm:RequestCertificate`
3. **DNS 访问**: 申请证书需要能够添加 DNS 验证记录

## 配置

在 `~/.cloudctl/config.yaml` 中配置 AWS 凭证:

```yaml
default_profile:
  aws: "default"

aws:
  default:
    access_key_id: "YOUR_ACCESS_KEY_ID"
    secret_access_key: "YOUR_SECRET_ACCESS_KEY"
    region: "us-east-1"  # ACM 证书必须在 us-east-1 区域
```

**重要**: CloudFront 使用的 ACM 证书必须在 `us-east-1` 区域申请。

## 基本操作

### 列出证书

列出所有 ACM 证书:

```bash
# 使用默认 profile
cloudctl aws cert list

# 使用指定 profile
cloudctl aws cert list -p aws-prod

# JSON 格式输出
cloudctl aws cert list -o json
```

**输出示例**:

```
ARN                                                          DOMAIN              STATUS   TYPE      IN_USE  NOT_BEFORE           NOT_AFTER
arn:aws:acm:us-east-1:123456789012:certificate/xxx-xxx-xxx  example.com         ISSUED   AMAZON    true    2024-01-01 00:00:00  2025-01-01 00:00:00
arn:aws:acm:us-east-1:123456789012:certificate/yyy-yyy-yyy  *.example.com       ISSUED   AMAZON    false   2024-01-01 00:00:00  2025-01-01 00:00:00
```

### 获取证书详情

获取指定证书的详细信息:

```bash
# 获取证书详情
cloudctl aws cert get arn:aws:acm:us-east-1:123456789012:certificate/xxx-xxx-xxx

# JSON 格式输出
cloudctl aws cert get arn:aws:acm:us-east-1:123456789012:certificate/xxx-xxx-xxx -o json
```

**输出示例**:

```json
{
  "arn": "arn:aws:acm:us-east-1:123456789012:certificate/xxx-xxx-xxx",
  "domain": "example.com",
  "status": "ISSUED",
  "type": "AMAZON_ISSUED",
  "in_use": true,
  "issued_at": "2024-01-01T00:00:00Z",
  "not_before": "2024-01-01T00:00:00Z",
  "not_after": "2025-01-01T00:00:00Z",
  "subject_alt_names": [
    "example.com",
    "www.example.com"
  ],
  "validation_records": [
    {
      "name": "_abc123.example.com",
      "type": "CNAME",
      "value": "_xyz456.acm-validations.aws",
      "status": "SUCCESS"
    }
  ]
}
```

### 申请新证书

#### 批量申请证书

使用 YAML 配置文件批量申请多个证书:

**1. 创建配置文件** (`certificates.yaml`):

```yaml
certificates:
  # 单域名证书
  - domain: example1.com
    san:
      - "*.example1.com"
      - www.example1.com
  
  # 多域名证书
  - domain: example2.com
    san:
      - "*.example2.com"
      - www.example2.com
      - api.example2.com
      - cdn.example2.com
  
  # 通配符证书
  - domain: "*.example3.com"
    san:
      - example3.com
```

**2. 执行批量申请**:

```bash
cloudctl aws cert request -f certificates.yaml
```

**输出示例**:

```
开始批量申请 3 个证书...

============================================================
批量申请完成
============================================================

总计: 3
成功: 3
失败: 0

成功的证书:
------------------------------------------------------------
✓ example1.com
  ARN: arn:aws:acm:us-east-1:123456789012:certificate/xxx-1

✓ example2.com
  ARN: arn:aws:acm:us-east-1:123456789012:certificate/xxx-2

✓ *.example3.com
  ARN: arn:aws:acm:us-east-1:123456789012:certificate/xxx-3

注意: 请为每个证书添加 DNS 验证记录
可以使用以下命令查看证书详情:
  cloudctl aws cert list
```

#### 申请单域名证书

```bash
cloudctl aws cert request -d example.com
```

#### 申请通配符证书

```bash
cloudctl aws cert request -d "*.example.com"
```

#### 申请多域名证书

```bash
# 使用多个 --san 参数
cloudctl aws cert request -d example.com \
  --san www.example.com \
  --san api.example.com

# 或使用逗号分隔
cloudctl aws cert request -d example.com --san "www.example.com,api.example.com"
```

#### 申请通配符 + 根域名证书

```bash
cloudctl aws cert request -d "*.example.com" --san example.com
```

**输出示例**:

```json
{
  "arn": "arn:aws:acm:us-east-1:123456789012:certificate/new-cert-id",
  "domain": "example.com",
  "status": "PENDING_VALIDATION",
  "validation_records": [
    {
      "name": "_abc123.example.com",
      "type": "CNAME",
      "value": "_xyz456.acm-validations.aws",
      "status": "PENDING_VALIDATION"
    }
  ]
}

✓ 证书申请已提交
证书 ARN: arn:aws:acm:us-east-1:123456789012:certificate/new-cert-id
状态: PENDING_VALIDATION

请在 DNS 中添加以下验证记录:

类型: CNAME
名称: _abc123.example.com
值: _xyz456.acm-validations.aws

注意: DNS 验证通常需要几分钟到几小时才能完成
可以使用以下命令检查证书状态:
  cloudctl aws cert get arn:aws:acm:us-east-1:123456789012:certificate/new-cert-id
```

## 使用场景

### 场景 1: 为 CloudFront 申请证书

为 CloudFront 分发申请 SSL 证书:

```bash
# 1. 申请证书 (必须在 us-east-1 区域)
cloudctl aws cert request -d cdn.example.com

# 2. 记录返回的证书 ARN 和 DNS 验证记录

# 3. 在 DNS 中添加验证记录
cloudctl cf dns create example.com \
  --type CNAME \
  --name _abc123.cdn.example.com \
  --content _xyz456.acm-validations.aws

# 4. 等待验证完成 (通常几分钟到几小时)
cloudctl aws cert get arn:aws:acm:us-east-1:123456789012:certificate/xxx

# 5. 验证完成后，创建 CloudFront 分发
cloudctl aws cdn create \
  --origin origin.example.com \
  --aliases cdn.example.com \
  --certificate-arn arn:aws:acm:us-east-1:123456789012:certificate/xxx
```

### 场景 2: 申请通配符证书

为所有子域名申请通配符证书:

```bash
# 申请通配符证书 (同时包含根域名)
cloudctl aws cert request -d "*.example.com" --san example.com

# 这样可以保护:
# - example.com
# - www.example.com
# - api.example.com
# - cdn.example.com
# - 等所有子域名
```

### 场景 3: 批量查看证书状态

查看所有证书并筛选特定状态:

```bash
# 列出所有证书
cloudctl aws cert list -o json | jq '.[] | select(.status == "PENDING_VALIDATION")'

# 查看即将过期的证书 (30 天内)
cloudctl aws cert list -o json | jq '.[] | select(.not_after < "2024-12-31")'
```

### 场景 4: 批量申请多个域名证书

为多个域名批量申请证书:

```bash
# 1. 创建配置文件
cat > certificates.yaml <<EOF
certificates:
  - domain: example1.com
    san: ["*.example1.com"]
  - domain: example2.com
    san: ["*.example2.com"]
  - domain: example3.com
    san: ["*.example3.com"]
EOF

# 2. 批量申请
cloudctl aws cert request -f certificates.yaml

# 3. 查看所有证书
cloudctl aws cert list

# 4. 为每个证书添加 DNS 验证记录
# (可以使用脚本自动化这个过程)
```

### 场景 5: 自动化证书申请

使用脚本自动申请证书并添加 DNS 记录:

```bash
#!/bin/bash

DOMAIN="example.com"

# 申请证书
OUTPUT=$(cloudctl aws cert request -d "$DOMAIN" -o json)

# 提取证书 ARN
CERT_ARN=$(echo "$OUTPUT" | jq -r '.arn')

# 提取 DNS 验证记录
DNS_NAME=$(echo "$OUTPUT" | jq -r '.validation_records[0].name')
DNS_VALUE=$(echo "$OUTPUT" | jq -r '.validation_records[0].value')

# 自动添加 DNS 记录
cloudctl cf dns create "$DOMAIN" \
  --type CNAME \
  --name "$DNS_NAME" \
  --content "$DNS_VALUE"

echo "证书 ARN: $CERT_ARN"
echo "DNS 验证记录已添加"
```

## 常见问题

### Q: 为什么必须在 us-east-1 区域申请证书?

A: CloudFront 是全球服务,它只能使用 us-east-1 区域的 ACM 证书。如果要在其他区域使用证书(如 ELB),可以在相应区域申请。

### Q: DNS 验证需要多久?

A: DNS 验证通常需要几分钟到几小时。AWS 会定期检查 DNS 记录,一旦验证成功,证书状态会变为 `ISSUED`。

### Q: 如何验证证书是否已签发?

A: 使用 `cloudctl aws cert get <arn>` 命令查看证书状态。当 `status` 为 `ISSUED` 时表示证书已签发。

### Q: 证书有效期是多久?

A: ACM 证书有效期为 13 个月(约 395 天),AWS 会在到期前自动续期,无需手动操作。

### Q: 可以申请多少个证书?

A: 默认情况下,每个 AWS 账户每年可以申请 2000 个 ACM 证书。如需更多,可以联系 AWS 支持。

### Q: 如何删除证书?

A: 目前 cloudctl 暂不支持删除证书。可以通过 AWS 控制台或 AWS CLI 删除未使用的证书。

### Q: 通配符证书可以保护根域名吗?

A: 不可以。`*.example.com` 只保护一级子域名(如 `www.example.com`),不保护根域名 `example.com`。需要同时申请根域名:
```bash
cloudctl aws cert request -d "*.example.com" --san example.com
```

### Q: 申请失败怎么办?

A: 检查:
1. 域名格式是否正确
2. AWS 凭证是否有效
3. 是否有足够的 ACM 配额
4. 域名是否已被其他证书使用

## 最佳实践

1. **使用通配符证书**: 对于有多个子域名的情况,使用通配符证书可以简化管理

2. **自动化 DNS 验证**: 如果使用 Cloudflare 等支持 API 的 DNS 服务,可以自动化添加验证记录

3. **监控证书状态**: 定期检查证书状态,确保验证成功

4. **记录证书 ARN**: 申请证书后立即记录 ARN,便于后续使用

5. **使用描述性域名**: 为证书使用清晰的域名,便于管理和识别

## 相关命令

- `cloudctl aws cdn create` - 创建 CloudFront 分发(使用证书)
- `cloudctl aws cdn list` - 列出 CloudFront 分发
- `cloudctl cf dns create` - 创建 DNS 记录(用于验证)
- `cloudctl cf dns list` - 列出 DNS 记录

## 参考资料

- [AWS ACM 文档](https://docs.aws.amazon.com/acm/)
- [ACM 证书验证](https://docs.aws.amazon.com/acm/latest/userguide/dns-validation.html)
- [CloudFront 使用 ACM 证书](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cnames-and-https-requirements.html)
- [ACM 配额](https://docs.aws.amazon.com/acm/latest/userguide/acm-limits.html)

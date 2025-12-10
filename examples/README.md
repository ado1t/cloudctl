# Examples

这个目录包含 cloudctl 各个功能模块的示例和测试程序。

## 目录结构

每个示例程序都在独立的子目录中，避免 main 包冲突：

```
examples/
├── test-log/         # 日志系统测试
│   └── main.go
├── test-output/      # 输出格式化测试
│   └── main.go
├── test-cloudflare/  # Cloudflare 客户端测试
│   └── main.go
└── README.md
```

## 日志系统示例

### test-log

测试日志系统的各种功能，包括日志级别、颜色输出、详细程度等。

**使用方法**:

```bash
# 默认日志级别（ERROR）
go run examples/test-log/main.go

# INFO 级别
go run examples/test-log/main.go -v

# DEBUG 级别
go run examples/test-log/main.go -vv

# DEBUG + 源码位置
go run examples/test-log/main.go -vvv

# 指定日志级别
go run examples/test-log/main.go --log-level debug

# 禁用颜色输出
go run examples/test-log/main.go --no-color

# 静默模式
go run examples/test-log/main.go -q
```

**预期输出**:

```
2025-11-20 09:27:49 [INFO ] 这是 INFO 级别的日志 key2="value2" enabled=true
2025-11-20 09:27:49 [WARN ] 这是 WARN 级别的日志 key3="value3" duration="5s"
2025-11-20 09:27:49 [ERROR] 这是 ERROR 级别的日志 key4="value4" count=100
2025-11-20 09:27:49 [INFO ] 这是带有额外字段的日志 component="test" version="1.0" status="running"
2025-11-20 09:27:49 [INFO ] 数据库连接成功 database.host="localhost" database.port=5432
```

## 输出格式化示例

### test-output

测试输出格式化系统的各种功能，包括表格、JSON、文本格式输出。

**使用方法**:

```bash
# 表格格式（默认）
go run examples/test-output/main.go

# JSON 格式
go run examples/test-output/main.go --output json

# 文本格式
go run examples/test-output/main.go --output text

# 禁用颜色输出
go run examples/test-output/main.go --no-color
```

**预期输出**:

表格格式：
```
name     age  city     
-------  ---  ---------
Alice    30   Beijing  
Bob      25   Shanghai 
Charlie  35   Guangzhou
```

JSON 格式：
```json
[
  {
    "age": 30,
    "city": "Beijing",
    "name": "Alice"
  }
]
```

## Cloudflare 客户端示例

### test-cloudflare

测试 Cloudflare 客户端的初始化、认证和错误处理功能。

**使用方法**:

```bash
# 使用默认 profile
go run examples/test-cloudflare/main.go

# 使用指定 profile
go run examples/test-cloudflare/main.go --profile cf-prod

# 指定配置文件
go run examples/test-cloudflare/main.go --config ~/.cloudctl/config.yaml

# 调试模式
go run examples/test-cloudflare/main.go --log-level debug

# 禁用颜色输出
go run examples/test-cloudflare/main.go --no-color
```

**前置条件**:

需要设置环境变量或配置文件：

```bash
# 方式 1: 环境变量
export CLOUDFLARE_API_TOKEN="your-api-token"

# 方式 2: 配置文件 ~/.cloudctl/config.yaml
cloudflare:
  default:
    api_token: ${CLOUDFLARE_API_TOKEN}
```

**预期输出**:

```
2025-11-21 09:15:00 [INFO ] Cloudflare 客户端测试程序
2025-11-21 09:15:00 [INFO ] 加载配置文件 path="~/.cloudctl/config.yaml"
2025-11-21 09:15:00 [INFO ] 配置加载成功
2025-11-21 09:15:00 [INFO ] 可用的 Cloudflare profiles profiles=["default"]
2025-11-21 09:15:00 [INFO ] 使用默认 profile profile="default"
2025-11-21 09:15:00 [INFO ] 创建 Cloudflare 客户端 profile="default"
2025-11-21 09:15:00 [DEBUG] Cloudflare 客户端初始化成功 profile="default"
2025-11-21 09:15:00 [INFO ] Cloudflare 客户端创建成功
2025-11-21 09:15:00 [INFO ] 测试带重试的 API 操作
2025-11-21 09:15:00 [DEBUG] 执行 API 操作 operation="测试操作" attempt=1 max_attempts=3
2025-11-21 09:15:00 [DEBUG] 执行测试操作
2025-11-21 09:15:00 [INFO ] 操作成功
2025-11-21 09:15:00 [INFO ] 测试错误处理
2025-11-21 09:15:00 [INFO ] === 测试错误类型 ===
2025-11-21 09:15:00 [INFO ] 测试完成
```

## DNS 管理示例

### test-dns

展示如何使用 cloudctl 管理 Cloudflare DNS 记录的完整示例。

**使用方法**:

```bash
# 查看示例说明
go run examples/test-dns/main.go
```

**功能演示**:

1. **列出 DNS 记录**
   ```bash
   cloudctl cf dns list example.com
   cloudctl cf dns list example.com --type A
   cloudctl cf dns list example.com -o json
   ```

2. **创建 DNS 记录**
   ```bash
   # 创建 A 记录
   cloudctl cf dns create example.com -t A -n www --content 1.2.3.4
   
   # 创建 A 记录并启用代理
   cloudctl cf dns create example.com -t A -n api --content 1.2.3.4 --proxied
   
   # 创建 CNAME 记录
   cloudctl cf dns create example.com -t CNAME -n blog --content example.com
   ```

3. **更新 DNS 记录**
   ```bash
   # 更新记录内容
   cloudctl cf dns update example.com <record-id> --content 2.3.4.5
   
   # 更新 TTL
   cloudctl cf dns update example.com <record-id> --ttl 7200
   
   # 启用/禁用代理
   cloudctl cf dns update example.com <record-id> --proxied
   cloudctl cf dns update example.com <record-id> --no-proxied
   ```

4. **删除 DNS 记录**
   ```bash
   cloudctl cf dns delete example.com <record-id>
   ```

5. **批量创建 DNS 记录**
   ```bash
   # 批量创建
   cloudctl cf dns batch-create --config dns-records.yaml
   
   # 预览模式
   cloudctl cf dns batch-create --config dns-records.yaml --dry-run
   
   # 使用并发
   cloudctl cf dns batch-create --config dns-records.yaml --concurrency 3
   ```

**前置条件**:

```bash
# 设置 Cloudflare API Token
export CLOUDFLARE_API_TOKEN="your-api-token"

# 确保已有至少一个 Zone
cloudctl cf zone list
```

## 未来示例

随着项目开发，这里将添加更多示例：

- **test-config/** - 配置系统示例
- **test-aws/** - AWS API 调用示例

## 注意事项

- 所有示例程序都是独立的可执行文件
- 示例程序不会被包含在最终的 cloudctl 二进制文件中
- 示例程序主要用于开发和测试，不要删除

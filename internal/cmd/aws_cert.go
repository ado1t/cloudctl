package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"

	"github.com/ado1t/cloudctl/internal/aws"
	"github.com/ado1t/cloudctl/internal/logger"
)

var (
	// Cert List 参数
	certListProfile string

	// Cert Get 参数
	certGetProfile string

	// Cert Request 参数
	certRequestProfile      string
	certRequestDomain       string
	certRequestSANs         []string
	certRequestConfigFile   string
	certRequestOutputConfig string
)

func init() {
	// 添加 Cert 子命令
	awsCertCmd.AddCommand(certListCmd)
	awsCertCmd.AddCommand(certGetCmd)
	awsCertCmd.AddCommand(certRequestCmd)

	// Cert List 命令参数
	certListCmd.Flags().StringVarP(&certListProfile, "profile", "p", "", "使用指定的 AWS profile")

	// Cert Get 命令参数
	certGetCmd.Flags().StringVarP(&certGetProfile, "profile", "p", "", "使用指定的 AWS profile")

	// Cert Request 命令参数
	certRequestCmd.Flags().StringVarP(&certRequestProfile, "profile", "p", "", "使用指定的 AWS profile")
	certRequestCmd.Flags().StringVarP(&certRequestDomain, "domain", "d", "", "主域名（单个申请时必需）")
	certRequestCmd.Flags().StringSliceVar(&certRequestSANs, "san", []string{}, "备用域名（可选，多个用逗号分隔）")
	certRequestCmd.Flags().StringVarP(&certRequestConfigFile, "config-file", "f", "", "批量申请配置文件（YAML 格式）")
	certRequestCmd.Flags().StringVar(&certRequestOutputConfig, "output-config", "", "输出 DNS 验证记录配置文件路径（用于 Cloudflare DNS 批量创建）")
}

// certListCmd 列出所有证书
var certListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有 ACM 证书",
	Long: `列出 AWS 账户中的所有 ACM 证书。

使用示例:
  cloudctl aws cert list                    # 使用默认 profile
  cloudctl aws cert list -p aws-prod        # 使用指定 profile
  cloudctl aws cert list -o json            # JSON 格式输出`,
	RunE: runCertList,
}

// runCertList 执行 cert list 命令
func runCertList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 创建 AWS 客户端
	client, err := aws.NewClient(certListProfile)
	if err != nil {
		return fmt.Errorf("创建 AWS 客户端失败: %w", err)
	}

	logger.Info("正在列出 ACM 证书...")

	// 列出证书
	certificates, err := client.ListCertificates(ctx)
	if err != nil {
		return fmt.Errorf("列出证书失败: %w", err)
	}

	if len(certificates) == 0 {
		fmt.Println("没有找到证书")
		return nil
	}

	logger.Info("成功列出证书", "count", len(certificates))

	// 格式化输出
	formatter := GetFormatter()

	// 转换为输出格式
	data := make([]map[string]interface{}, len(certificates))
	for i, cert := range certificates {
		data[i] = map[string]interface{}{
			"arn":        cert.ARN,
			"domain":     cert.DomainName,
			"status":     cert.Status,
			"type":       cert.Type,
			"in_use":     cert.InUse,
			"not_before": formatTime(cert.NotBefore),
			"not_after":  formatTime(cert.NotAfter),
		}
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	return nil
}

// certGetCmd 获取证书详情
var certGetCmd = &cobra.Command{
	Use:   "get <certificate-arn>",
	Short: "获取 ACM 证书详情",
	Long: `获取指定 ACM 证书的详细信息。

使用示例:
  cloudctl aws cert get arn:aws:acm:us-east-1:123456789012:certificate/xxx
  cloudctl aws cert get arn:aws:acm:us-east-1:123456789012:certificate/xxx -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runCertGet,
}

// runCertGet 执行 cert get 命令
func runCertGet(cmd *cobra.Command, args []string) error {
	certificateARN := args[0]
	ctx := context.Background()

	// 创建 AWS 客户端
	client, err := aws.NewClient(certGetProfile)
	if err != nil {
		return fmt.Errorf("创建 AWS 客户端失败: %w", err)
	}

	logger.Info("正在获取证书详情...", "arn", certificateARN)

	// 获取证书详情
	cert, err := client.GetCertificate(ctx, certificateARN)
	if err != nil {
		return fmt.Errorf("获取证书详情失败: %w", err)
	}

	logger.Info("成功获取证书详情", "domain", cert.DomainName)

	// 格式化输出
	formatter := GetFormatter()

	// 转换验证记录
	validationRecords := make([]map[string]interface{}, len(cert.ValidationRecords))
	for i, record := range cert.ValidationRecords {
		validationRecords[i] = map[string]interface{}{
			"name":   record.Name,
			"type":   record.Type,
			"value":  record.Value,
			"status": record.Status,
		}
	}

	// 转换为输出格式
	data := map[string]interface{}{
		"arn":                cert.ARN,
		"domain":             cert.DomainName,
		"status":             cert.Status,
		"type":               cert.Type,
		"in_use":             cert.InUse,
		"issued_at":          formatTime(cert.IssuedAt),
		"not_before":         formatTime(cert.NotBefore),
		"not_after":          formatTime(cert.NotAfter),
		"subject_alt_names":  cert.SubjectAltNames,
		"validation_records": validationRecords,
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	return nil
}

// certRequestCmd 申请新证书
var certRequestCmd = &cobra.Command{
	Use:   "request",
	Short: "申请新的 ACM 证书",
	Long: `申请新的 ACM 证书（使用 DNS 验证）。

使用示例:
  # 申请单域名证书
  cloudctl aws cert request -d example.com

  # 申请通配符证书
  cloudctl aws cert request -d "*.example.com"

  # 申请多域名证书
  cloudctl aws cert request -d example.com --san www.example.com --san api.example.com

  # 批量申请证书（使用配置文件）
  cloudctl aws cert request -f certificates.yaml

  # 批量申请证书并生成 DNS 验证记录配置文件
  cloudctl aws cert request -f certificates.yaml --output-config dns-validation.yaml

  # 使用指定 profile
  cloudctl aws cert request -d example.com -p aws-prod`,
	RunE: runCertRequest,
}

// runCertRequest 执行 cert request 命令
func runCertRequest(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 创建 AWS 客户端
	client, err := aws.NewClient(certRequestProfile)
	if err != nil {
		return fmt.Errorf("创建 AWS 客户端失败: %w", err)
	}

	// 判断是批量申请还是单个申请
	if certRequestConfigFile != "" {
		return runBatchCertRequest(ctx, client)
	}

	// 单个申请 - 验证参数
	if certRequestDomain == "" {
		return fmt.Errorf("必须指定域名 (-d/--domain) 或配置文件 (-f/--config-file)")
	}

	logger.Info("正在申请证书...", "domain", certRequestDomain, "sans", certRequestSANs)

	// 准备申请参数
	input := &aws.RequestCertificateInput{
		DomainName:              certRequestDomain,
		SubjectAlternativeNames: certRequestSANs,
	}

	// 申请证书
	cert, err := client.RequestCertificate(ctx, input)
	if err != nil {
		return fmt.Errorf("申请证书失败: %w", err)
	}

	logger.Info("成功申请证书", "arn", cert.ARN)

	// 格式化输出
	formatter := GetFormatter()

	// 转换验证记录
	validationRecords := make([]map[string]interface{}, len(cert.ValidationRecords))
	for i, record := range cert.ValidationRecords {
		validationRecords[i] = map[string]interface{}{
			"name":   record.Name,
			"type":   record.Type,
			"value":  record.Value,
			"status": record.Status,
		}
	}

	// 转换为输出格式
	data := map[string]interface{}{
		"arn":                cert.ARN,
		"domain":             cert.DomainName,
		"status":             cert.Status,
		"validation_records": validationRecords,
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	// 显示提示信息
	fmt.Printf("\n✓ 证书申请已提交\n")
	fmt.Printf("证书 ARN: %s\n", cert.ARN)
	fmt.Printf("状态: %s\n\n", cert.Status)

	if len(cert.ValidationRecords) > 0 {
		fmt.Printf("请在 DNS 中添加以下验证记录:\n\n")
		for _, record := range cert.ValidationRecords {
			fmt.Printf("类型: %s\n", record.Type)
			fmt.Printf("名称: %s\n", record.Name)
			fmt.Printf("值: %s\n\n", record.Value)
		}
		fmt.Printf("注意: DNS 验证通常需要几分钟到几小时才能完成\n")
		fmt.Printf("可以使用以下命令检查证书状态:\n")
		fmt.Printf("  cloudctl aws cert get %s\n", cert.ARN)
	}

	return nil
}

// runBatchCertRequest 执行批量证书申请
func runBatchCertRequest(ctx context.Context, client *aws.Client) error {
	logger.Info("正在读取配置文件...", "file", certRequestConfigFile)

	// 读取配置文件
	data, err := os.ReadFile(certRequestConfigFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var config aws.CertificatesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	if len(config.Certificates) == 0 {
		return fmt.Errorf("配置文件中没有证书申请记录")
	}

	logger.Info("开始批量申请证书", "count", len(config.Certificates))
	fmt.Printf("\n开始批量申请 %d 个证书...\n\n", len(config.Certificates))

	// 批量申请
	result := client.BatchRequestCertificates(ctx, config.Certificates)

	// 显示结果
	separator := strings.Repeat("=", 60)
	fmt.Printf("\n%s\n", separator)
	fmt.Printf("批量申请完成\n")
	fmt.Printf("%s\n\n", separator)
	fmt.Printf("总计: %d\n", result.Total)
	fmt.Printf("成功: %d\n", result.Success)
	fmt.Printf("失败: %d\n\n", result.Failed)

	// 显示详细结果
	if result.Success > 0 {
		line := strings.Repeat("-", 60)
		fmt.Printf("成功的证书:\n")
		fmt.Printf("%s\n", line)
		for _, r := range result.Results {
			if r.Success {
				fmt.Printf("✓ %s\n", r.Domain)
				fmt.Printf("  ARN: %s\n\n", r.ARN)
			}
		}
	}

	if result.Failed > 0 {
		line := strings.Repeat("-", 60)
		fmt.Printf("失败的证书:\n")
		fmt.Printf("%s\n", line)
		for _, r := range result.Results {
			if !r.Success {
				fmt.Printf("✗ %s\n", r.Domain)
				fmt.Printf("  错误: %s\n\n", r.Error)
			}
		}
	}

	fmt.Printf("\n注意: 请为每个证书添加 DNS 验证记录\n")
	fmt.Printf("可以使用以下命令查看证书详情:\n")
	fmt.Printf("  cloudctl aws cert list\n")

	// 如果指定了输出配置文件，生成 DNS 验证记录配置
	if certRequestOutputConfig != "" && result.Success > 0 {
		logger.Info("正在生成 DNS 验证记录配置文件...", "file", certRequestOutputConfig)

		dnsConfig, err := generateDNSConfig(result)
		if err != nil {
			logger.Error("生成 DNS 配置失败", "error", err)
			fmt.Printf("\n警告: 生成 DNS 配置文件失败: %v\n", err)
		} else {
			if err := saveDNSConfigToFile(dnsConfig, certRequestOutputConfig); err != nil {
				logger.Error("保存 DNS 配置文件失败", "error", err)
				fmt.Printf("\n警告: 保存 DNS 配置文件失败: %v\n", err)
			} else {
				logger.Info("成功生成 DNS 配置文件", "file", certRequestOutputConfig)
				fmt.Printf("\n✓ 已生成 DNS 验证记录配置文件: %s\n", certRequestOutputConfig)
				fmt.Printf("可以使用以下命令批量创建 DNS 记录:\n")
				fmt.Printf("  cloudctl cf dns batch-create --config %s\n", certRequestOutputConfig)
			}
		}
	}

	// 如果有失败的,返回错误
	if result.Failed > 0 {
		return fmt.Errorf("有 %d 个证书申请失败", result.Failed)
	}

	return nil
}

// formatTime 格式化时间
func formatTime(t interface{}) string {
	switch v := t.(type) {
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// DNSBatchConfig DNS 批量创建配置（用于 Cloudflare）
type DNSBatchConfig struct {
	Zones []DNSZoneConfig `yaml:"zones"`
}

// DNSZoneConfig 单个 Zone 的 DNS 配置
type DNSZoneConfig struct {
	Zone    string          `yaml:"zone"`
	Records []DNSRecordItem `yaml:"records"`
}

// DNSRecordItem DNS 记录项
type DNSRecordItem struct {
	Type    string `yaml:"type"`
	Name    string `yaml:"name"`
	Content string `yaml:"content"`
	Proxied bool   `yaml:"proxied"`
}

// generateDNSConfig 从批量申请结果生成 DNS 配置
func generateDNSConfig(result *aws.BatchRequestResult) (*DNSBatchConfig, error) {
	if result == nil || len(result.Results) == 0 {
		return nil, fmt.Errorf("没有证书申请结果")
	}

	// 按域名分组验证记录，使用 map 去重
	zoneMap := make(map[string]map[string]DNSRecordItem)

	for _, certResult := range result.Results {
		if !certResult.Success {
			continue
		}

		if len(certResult.ValidationRecords) == 0 {
			logger.Warn("证书没有验证记录", "domain", certResult.Domain)
			continue
		}

		for _, valRecord := range certResult.ValidationRecords {
			// 从验证记录的 Name 中提取域名
			// 例如: _c93b9a33d8bde7910f5e9680040b841c.example.com. -> example.com
			zone := extractZoneFromValidationName(valRecord.Name)
			if zone == "" {
				logger.Warn("无法从验证记录中提取域名", "name", valRecord.Name)
				continue
			}

			// 提取子域名部分
			// 例如: _c93b9a33d8bde7910f5e9680040b841c.example.com. -> _c93b9a33d8bde7910f5e9680040b841c
			name := extractSubdomainFromValidationName(valRecord.Name, zone)

			record := DNSRecordItem{
				Type:    valRecord.Type,
				Name:    name,
				Content: strings.TrimSuffix(valRecord.Value, "."), // 去除 FQDN 尾部的点
				Proxied: false,                                    // ACM 验证记录不能开启代理
			}

			// 使用记录的唯一键去重（zone + name + type + content）
			recordKey := fmt.Sprintf("%s|%s|%s|%s", zone, name, valRecord.Type, valRecord.Value)

			if zoneMap[zone] == nil {
				zoneMap[zone] = make(map[string]DNSRecordItem)
			}
			zoneMap[zone][recordKey] = record
		}
	}

	if len(zoneMap) == 0 {
		return nil, fmt.Errorf("没有可用的 DNS 验证记录")
	}

	// 构建配置
	config := &DNSBatchConfig{
		Zones: make([]DNSZoneConfig, 0, len(zoneMap)),
	}

	for zone, recordMap := range zoneMap {
		// 将 map 转换为切片
		records := make([]DNSRecordItem, 0, len(recordMap))
		for _, record := range recordMap {
			records = append(records, record)
		}

		config.Zones = append(config.Zones, DNSZoneConfig{
			Zone:    zone,
			Records: records,
		})
	}

	return config, nil
}

// extractZoneFromValidationName 从验证记录名称中提取域名
// 例如: _c93b9a33d8bde7910f5e9680040b841c.example.com. -> example.com
//
//	_c93b9a33d8bde7910f5e9680040b841c.sub.example.com. -> sub.example.com
func extractZoneFromValidationName(name string) string {
	// 去除 FQDN 尾部的点
	name = strings.TrimSuffix(name, ".")

	// 移除最前面的验证标识部分
	parts := strings.Split(name, ".")
	if len(parts) <= 1 {
		return ""
	}

	// 跳过第一个部分（验证哈希）
	// 剩余部分就是域名
	return strings.Join(parts[1:], ".")
}

// extractSubdomainFromValidationName 从验证记录名称中提取子域名部分
// 例如: _c93b9a33d8bde7910f5e9680040b841c.example.com., example.com -> _c93b9a33d8bde7910f5e9680040b841c
func extractSubdomainFromValidationName(name, zone string) string {
	// 去除 FQDN 尾部的点
	name = strings.TrimSuffix(name, ".")

	// 移除域名部分，只保留子域名
	if strings.HasSuffix(name, "."+zone) {
		return strings.TrimSuffix(name, "."+zone)
	}
	// 如果完全匹配，说明是根域名验证
	if name == zone {
		return "@"
	}
	return name
}

// saveDNSConfigToFile 保存 DNS 配置到文件
func saveDNSConfigToFile(config *DNSBatchConfig, filepath string) error {
	// 创建文件
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入文件头注释
	header := `# DNS 批量创建配置（由 cloudctl aws cert request 自动生成）
# 使用方法: cloudctl cf dns batch-create --config ` + filepath + `

`
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("写入文件头失败: %w", err)
	}

	// 使用 yaml.v3 的 Encoder，设置缩进为 2 个空格
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2) // 设置缩进为 2 个空格
	defer encoder.Close()

	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	return nil
}

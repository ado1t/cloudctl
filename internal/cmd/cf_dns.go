package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ado1t/cloudctl/internal/cloudflare"
	"github.com/ado1t/cloudctl/internal/logger"
)

func init() {
	// 添加 dns 子命令
	cfDnsCmd.AddCommand(cfDnsListCmd)
	cfDnsCmd.AddCommand(cfDnsCreateCmd)
	cfDnsCmd.AddCommand(cfDnsUpdateCmd)
	cfDnsCmd.AddCommand(cfDnsDeleteCmd)
	cfDnsCmd.AddCommand(cfDnsBatchCreateCmd)

	// dns list 命令参数
	cfDnsListCmd.Flags().StringP("type", "t", "", "过滤记录类型 (A, AAAA, CNAME 等)")

	// dns create 命令参数
	cfDnsCreateCmd.Flags().StringP("type", "t", "", "记录类型 (A, CNAME)")
	cfDnsCreateCmd.Flags().StringP("name", "n", "", "记录名称")
	cfDnsCreateCmd.Flags().String("content", "", "记录内容")
	cfDnsCreateCmd.Flags().Float64("ttl", 1, "TTL (1 = 自动)")
	cfDnsCreateCmd.Flags().Bool("proxied", false, "启用 Cloudflare 代理")
	cfDnsCreateCmd.Flags().String("config", "", "批量操作配置文件 (YAML)")

	// dns batch-create 命令参数
	cfDnsBatchCreateCmd.Flags().String("config", "", "批量操作配置文件 (YAML)")
	cfDnsBatchCreateCmd.Flags().Bool("dry-run", false, "预览模式，不实际执行")
	cfDnsBatchCreateCmd.Flags().Int("concurrency", 1, "并发数 (1-10)")
	cfDnsBatchCreateCmd.MarkFlagRequired("config")

	// dns update 命令参数
	cfDnsUpdateCmd.Flags().String("content", "", "新的记录内容")
	cfDnsUpdateCmd.Flags().Float64("ttl", 0, "新的 TTL")
	cfDnsUpdateCmd.Flags().Bool("proxied", false, "是否启用代理")
	cfDnsUpdateCmd.Flags().Bool("no-proxied", false, "禁用代理")
}

// cfDnsListCmd 列出 DNS 记录
var cfDnsListCmd = &cobra.Command{
	Use:   "list <domain>",
	Short: "列出 DNS 记录",
	Long: `列出指定域名的所有 DNS 记录。

显示信息包括:
  - 记录类型 (A, CNAME 等)
  - 记录名称
  - 记录内容
  - TTL
  - 是否启用代理

使用示例:
  cloudctl cf dns list example.com                    # 列出所有记录
  cloudctl cf dns list example.com --type A           # 只列出 A 记录
  cloudctl cf dns list example.com -t CNAME           # 只列出 CNAME 记录
  cloudctl cf dns list example.com -o json            # JSON 格式输出`,
	Args: cobra.ExactArgs(1),
	RunE: runDNSList,
}

// runDNSList 执行 dns list 命令
func runDNSList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	domain := args[0]

	// 获取参数
	profile, _ := cmd.Flags().GetString("profile")
	recordType, _ := cmd.Flags().GetString("type")

	// 创建 Cloudflare 客户端
	logger.Debug("创建 Cloudflare 客户端", "profile", profile)
	client, err := cloudflare.NewClient(profile, logger.Logger)
	if err != nil {
		logger.Error("创建客户端失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}
	defer client.Close()

	// 获取 Zone ID
	logger.Info("正在查找域名...", "domain", domain)
	zone, err := client.GetZoneByName(ctx, domain)
	if err != nil {
		logger.Error("查找域名失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 列出 DNS 记录
	logger.Info("正在获取 DNS 记录列表...")
	records, err := client.ListDNSRecords(ctx, zone.ID, recordType)
	if err != nil {
		logger.Error("获取 DNS 记录列表失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	if len(records) == 0 {
		logger.Info("未找到任何 DNS 记录")
		fmt.Println("未找到任何 DNS 记录")
		return nil
	}

	// 获取格式化器
	formatter := GetFormatter()

	// 转换为输出格式
	data := make([]map[string]interface{}, len(records))
	for i, record := range records {
		data[i] = map[string]interface{}{
			"type":    record.Type,
			"name":    record.Name,
			"content": record.Content,
			"ttl":     formatTTL(record.TTL),
			"proxied": formatProxied(record.Proxied, record.Proxiable),
			"id":      record.ID,
		}
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		logger.Error("格式化输出失败", "error", err)
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	logger.Info("成功列出 DNS 记录", "total", len(records))
	return nil
}

// cfDnsCreateCmd 创建 DNS 记录
var cfDnsCreateCmd = &cobra.Command{
	Use:   "create <domain>",
	Short: "创建 DNS 记录",
	Long: `在指定域名下创建新的 DNS 记录。

支持的记录类型:
  - A: IPv4 地址
  - AAAA: IPv6 地址
  - CNAME: 别名记录

使用示例:
  # 创建 A 记录
  cloudctl cf dns create example.com -t A -n www --content 1.2.3.4

  # 创建 A 记录并启用代理
  cloudctl cf dns create example.com -t A -n www --content 1.2.3.4 --proxied

  # 创建 CNAME 记录
  cloudctl cf dns create example.com -t CNAME -n blog --content example.com

  # 创建记录并指定 TTL
  cloudctl cf dns create example.com -t A -n api --content 1.2.3.4 --ttl 3600`,
	Args: cobra.ExactArgs(1),
	RunE: runDNSCreate,
}

// runDNSCreate 执行 dns create 命令
func runDNSCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	domain := args[0]

	// 获取参数
	profile, _ := cmd.Flags().GetString("profile")
	recordType, _ := cmd.Flags().GetString("type")
	name, _ := cmd.Flags().GetString("name")
	content, _ := cmd.Flags().GetString("content")
	ttl, _ := cmd.Flags().GetFloat64("ttl")
	proxied, _ := cmd.Flags().GetBool("proxied")

	// 验证记录类型
	recordType = strings.ToUpper(recordType)
	if recordType != "A" && recordType != "AAAA" && recordType != "CNAME" {
		return fmt.Errorf("不支持的记录类型: %s (支持: A, AAAA, CNAME)", recordType)
	}

	// 创建 Cloudflare 客户端
	logger.Debug("创建 Cloudflare 客户端", "profile", profile)
	client, err := cloudflare.NewClient(profile, logger.Logger)
	if err != nil {
		logger.Error("创建客户端失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}
	defer client.Close()

	// 获取 Zone ID
	logger.Info("正在查找域名...", "domain", domain)
	zone, err := client.GetZoneByName(ctx, domain)
	if err != nil {
		logger.Error("查找域名失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 创建 DNS 记录
	logger.Info("正在创建 DNS 记录...", "type", recordType, "name", name)
	record, err := client.CreateDNSRecord(ctx, zone.ID, cloudflare.DNSRecordCreateParams{
		Type:    recordType,
		Name:    name,
		Content: content,
		TTL:     ttl,
		Proxied: proxied,
	})
	if err != nil {
		logger.Error("创建 DNS 记录失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 获取格式化器
	formatter := GetFormatter()

	// 输出结果
	data := map[string]interface{}{
		"id":      record.ID,
		"type":    record.Type,
		"name":    record.Name,
		"content": record.Content,
		"ttl":     formatTTL(record.TTL),
		"proxied": formatProxied(record.Proxied, record.Proxiable),
	}

	if err := formatter.Format(data); err != nil {
		logger.Error("格式化输出失败", "error", err)
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	logger.Info("成功创建 DNS 记录", "record_id", record.ID)
	return nil
}

// cfDnsUpdateCmd 更新 DNS 记录
var cfDnsUpdateCmd = &cobra.Command{
	Use:   "update <domain> <record-id>",
	Short: "更新 DNS 记录",
	Long: `更新指定的 DNS 记录。

可以更新的字段:
  - content: 记录内容
  - ttl: TTL 值
  - proxied: 是否启用代理

使用示例:
  # 更新记录内容
  cloudctl cf dns update example.com abc123 --content 2.3.4.5

  # 更新 TTL
  cloudctl cf dns update example.com abc123 --ttl 3600

  # 启用代理
  cloudctl cf dns update example.com abc123 --proxied

  # 禁用代理
  cloudctl cf dns update example.com abc123 --no-proxied

  # 同时更新多个字段
  cloudctl cf dns update example.com abc123 --content 2.3.4.5 --ttl 3600 --proxied`,
	Args: cobra.ExactArgs(2),
	RunE: runDNSUpdate,
}

// runDNSUpdate 执行 dns update 命令
func runDNSUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	domain := args[0]
	recordID := args[1]

	// 获取参数
	profile, _ := cmd.Flags().GetString("profile")

	// 创建 Cloudflare 客户端
	logger.Debug("创建 Cloudflare 客户端", "profile", profile)
	client, err := cloudflare.NewClient(profile, logger.Logger)
	if err != nil {
		logger.Error("创建客户端失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}
	defer client.Close()

	// 获取 Zone ID
	logger.Info("正在查找域名...", "domain", domain)
	zone, err := client.GetZoneByName(ctx, domain)
	if err != nil {
		logger.Error("查找域名失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 获取现有记录以确定类型
	existingRecord, err := client.GetDNSRecord(ctx, zone.ID, recordID)
	if err != nil {
		logger.Error("获取 DNS 记录失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 构建更新参数
	params := cloudflare.DNSRecordUpdateParams{}

	// 检查是否有要更新的字段
	hasUpdates := false

	if cmd.Flags().Changed("content") {
		content, _ := cmd.Flags().GetString("content")
		params.Content = &content
		hasUpdates = true
	}

	if cmd.Flags().Changed("ttl") {
		ttl, _ := cmd.Flags().GetFloat64("ttl")
		params.TTL = &ttl
		hasUpdates = true
	}

	if cmd.Flags().Changed("proxied") {
		proxied, _ := cmd.Flags().GetBool("proxied")
		params.Proxied = &proxied
		hasUpdates = true
	}

	if cmd.Flags().Changed("no-proxied") {
		proxied := false
		params.Proxied = &proxied
		hasUpdates = true
	}

	if !hasUpdates {
		return fmt.Errorf("请至少指定一个要更新的字段 (--content, --ttl, --proxied, --no-proxied)")
	}

	// 更新 DNS 记录
	logger.Info("正在更新 DNS 记录...", "record_id", recordID)
	record, err := client.UpdateDNSRecord(ctx, zone.ID, recordID, existingRecord.Type, params)
	if err != nil {
		logger.Error("更新 DNS 记录失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 获取格式化器
	formatter := GetFormatter()

	// 输出结果
	data := map[string]interface{}{
		"id":      record.ID,
		"type":    record.Type,
		"name":    record.Name,
		"content": record.Content,
		"ttl":     formatTTL(record.TTL),
		"proxied": formatProxied(record.Proxied, record.Proxiable),
	}

	if err := formatter.Format(data); err != nil {
		logger.Error("格式化输出失败", "error", err)
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	logger.Info("成功更新 DNS 记录", "record_id", record.ID)
	return nil
}

// cfDnsDeleteCmd 删除 DNS 记录
var cfDnsDeleteCmd = &cobra.Command{
	Use:   "delete <domain> <record-id>",
	Short: "删除 DNS 记录",
	Long: `删除指定的 DNS 记录。

注意: 此操作不可逆，请谨慎使用。

使用示例:
  cloudctl cf dns delete example.com abc123`,
	Args: cobra.ExactArgs(2),
	RunE: runDNSDelete,
}

// runDNSDelete 执行 dns delete 命令
func runDNSDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	domain := args[0]
	recordID := args[1]

	// 获取参数
	profile, _ := cmd.Flags().GetString("profile")

	// 创建 Cloudflare 客户端
	logger.Debug("创建 Cloudflare 客户端", "profile", profile)
	client, err := cloudflare.NewClient(profile, logger.Logger)
	if err != nil {
		logger.Error("创建客户端失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}
	defer client.Close()

	// 获取 Zone ID
	logger.Info("正在查找域名...", "domain", domain)
	zone, err := client.GetZoneByName(ctx, domain)
	if err != nil {
		logger.Error("查找域名失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 获取记录信息用于确认
	record, err := client.GetDNSRecord(ctx, zone.ID, recordID)
	if err != nil {
		logger.Error("获取 DNS 记录失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 显示确认信息
	fmt.Printf("确认删除以下 DNS 记录?\n")
	fmt.Printf("  类型: %s\n", record.Type)
	fmt.Printf("  名称: %s\n", record.Name)
	fmt.Printf("  内容: %s\n", record.Content)
	fmt.Printf("  ID: %s\n", record.ID)
	fmt.Print("\n输入 'yes' 确认删除: ")

	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" {
		fmt.Println("已取消删除操作")
		return nil
	}

	// 删除 DNS 记录
	logger.Info("正在删除 DNS 记录...", "record_id", recordID)
	err = client.DeleteDNSRecord(ctx, zone.ID, recordID)
	if err != nil {
		logger.Error("删除 DNS 记录失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 获取格式化器
	formatter := GetFormatter()

	// 输出结果
	data := map[string]interface{}{
		"status":    "deleted",
		"record_id": recordID,
		"type":      record.Type,
		"name":      record.Name,
	}

	if err := formatter.Format(data); err != nil {
		logger.Error("格式化输出失败", "error", err)
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	logger.Info("成功删除 DNS 记录", "record_id", recordID)
	return nil
}

// formatTTL 格式化 TTL 显示
func formatTTL(ttl float64) string {
	if ttl == 1 {
		return "auto"
	}
	return strconv.FormatFloat(ttl, 'f', 0, 64)
}

// formatProxied 格式化 Proxied 显示
func formatProxied(proxied, proxiable bool) string {
	if !proxiable {
		return "N/A"
	}
	if proxied {
		return "yes"
	}
	return "no"
}

// cfDnsBatchCreateCmd 批量创建 DNS 记录
var cfDnsBatchCreateCmd = &cobra.Command{
	Use:   "batch-create",
	Short: "批量创建 DNS 记录",
	Long: `通过 YAML 配置文件批量创建 DNS 记录。

支持多个 zone，每个 zone 可以包含多个记录。
失败时会继续执行其他项，最后汇总结果。

配置文件示例 (dns-records.yaml):
  zones:
    - zone: example1.com
      records:
        - type: A
          name: www
          content: 1.2.3.4
          ttl: 3600
          proxied: true
        - type: CNAME
          name: blog
          content: example1.com
          proxied: true
    
    - zone: example2.com
      records:
        - type: A
          name: www
          content: 2.3.4.5
          ttl: 3600
          proxied: true

使用示例:
  # 批量创建 DNS 记录
  cloudctl cf dns batch-create --config dns-records.yaml

  # 预览模式（不实际执行）
  cloudctl cf dns batch-create --config dns-records.yaml --dry-run

  # 使用并发加速（最多 10 个并发）
  cloudctl cf dns batch-create --config dns-records.yaml --concurrency 3

  # 查看详细日志
  cloudctl cf dns batch-create --config dns-records.yaml -vv`,
	Args: cobra.NoArgs,
	RunE: runDNSBatchCreate,
}

// runDNSBatchCreate 执行批量创建命令
func runDNSBatchCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 获取参数
	profile, _ := cmd.Flags().GetString("profile")
	configFile, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	concurrency, _ := cmd.Flags().GetInt("concurrency")

	// 验证并发数
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > 10 {
		concurrency = 10
	}

	// 加载配置文件
	logger.Info("加载批量操作配置文件", "file", configFile)
	config, err := cloudflare.LoadDNSBatchConfig(configFile)
	if err != nil {
		logger.Error("加载配置文件失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	logger.Info("配置文件加载成功",
		"zones", len(config.Zones),
		"total_records", countTotalRecords(config),
	)

	// 预览模式
	if dryRun {
		fmt.Println("=== 预览模式 ===")
		fmt.Printf("总共 %d 个 zone，%d 条记录\n\n", len(config.Zones), countTotalRecords(config))

		for i, zone := range config.Zones {
			fmt.Printf("Zone %d: %s (%d 条记录)\n", i+1, zone.Zone, len(zone.Records))
			for j, record := range zone.Records {
				fmt.Printf("  %d. %s %s -> %s", j+1, record.Type, record.Name, record.Content)
				if record.TTL > 0 && record.TTL != 1 {
					fmt.Printf(" (TTL: %.0f)", record.TTL)
				}
				if record.Proxied {
					fmt.Print(" [Proxied]")
				}
				fmt.Println()
			}
			fmt.Println()
		}

		fmt.Println("使用 --dry-run=false 执行实际创建")
		return nil
	}

	// 创建 Cloudflare 客户端
	logger.Debug("创建 Cloudflare 客户端", "profile", profile)
	client, err := cloudflare.NewClient(profile, logger.Logger)
	if err != nil {
		logger.Error("创建客户端失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}
	defer client.Close()

	// 进度回调
	progressCallback := func(msg string) {
		logger.Info(msg)
	}

	// 执行批量创建
	var result *cloudflare.DNSBatchResult
	if concurrency > 1 {
		logger.Info("开始并发批量创建", "concurrency", concurrency)
		result, err = client.BatchCreateDNSRecordsConcurrent(ctx, config, concurrency, progressCallback)
	} else {
		logger.Info("开始批量创建")
		result, err = client.BatchCreateDNSRecords(ctx, config, progressCallback)
	}

	if err != nil {
		logger.Error("批量创建失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 输出结果汇总
	fmt.Println("\n=== 批量创建结果 ===")
	fmt.Printf("总耗时: %s\n\n", result.Duration.Round(time.Millisecond))

	fmt.Printf("Zone 统计:\n")
	fmt.Printf("  总数: %d\n", result.TotalZones)
	fmt.Printf("  成功: %d\n", result.SuccessZones)
	fmt.Printf("  失败: %d\n\n", result.FailedZones)

	fmt.Printf("记录统计:\n")
	fmt.Printf("  总数: %d\n", result.TotalRecords)
	fmt.Printf("  成功: %d\n", result.SuccessRecords)
	fmt.Printf("  失败: %d\n\n", result.FailedRecords)

	// 详细结果
	fmt.Println("详细结果:")
	for i, zoneResult := range result.ZoneResults {
		status := "✓"
		if !zoneResult.Success {
			status = "✗"
		}

		fmt.Printf("\n%s Zone %d: %s\n", status, i+1, zoneResult.Zone)

		if zoneResult.Error != nil {
			fmt.Printf("  错误: %v\n", zoneResult.Error)
			continue
		}

		fmt.Printf("  记录: %d 成功, %d 失败\n", zoneResult.SuccessRecords, zoneResult.FailedRecords)

		// 显示失败的记录
		for _, recordResult := range zoneResult.RecordResults {
			if !recordResult.Success {
				fmt.Printf("    ✗ %s %s -> %s: %v\n",
					recordResult.Type, recordResult.Name, recordResult.Content, recordResult.Error)
			}
		}
	}

	// 如果有失败，返回错误
	if result.FailedRecords > 0 || result.FailedZones > 0 {
		fmt.Println("\n部分操作失败，请检查上述错误信息")
		os.Exit(1)
	}

	fmt.Println("\n✓ 所有记录创建成功")
	return nil
}

// countTotalRecords 计算总记录数
func countTotalRecords(config *cloudflare.DNSBatchConfig) int {
	total := 0
	for _, zone := range config.Zones {
		total += len(zone.Records)
	}
	return total
}

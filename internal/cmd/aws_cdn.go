package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ado1t/cloudctl/internal/aws"
	"github.com/ado1t/cloudctl/internal/logger"
)

var (
	// CDN List 参数
	cdnListProfile string

	// CDN Get 参数
	cdnGetProfile string

	// CDN Create 参数
	cdnCreateProfile           string
	cdnCreateOrigin            string
	cdnCreateAliases           []string
	cdnCreateComment           string
	cdnCreateEnabled           bool
	cdnCreateCertificateARN    string
	cdnCreatePriceClass        string
	cdnCreateDefaultRootObject string

	// CDN Update 参数
	cdnUpdateProfile string
	cdnUpdateComment string
	cdnUpdateEnabled *bool

	// CDN Invalidate 参数
	cdnInvalidateProfile         string
	cdnInvalidatePaths           []string
	cdnInvalidateCallerReference string

	// CDN Invalidate Status 参数
	cdnInvalidateStatusProfile string
)

func init() {
	// 添加 CDN 子命令
	awsCdnCmd.AddCommand(cdnListCmd)
	awsCdnCmd.AddCommand(cdnGetCmd)
	awsCdnCmd.AddCommand(cdnCreateCmd)
	awsCdnCmd.AddCommand(cdnUpdateCmd)
	awsCdnCmd.AddCommand(cdnInvalidateCmd)
	awsCdnCmd.AddCommand(cdnInvalidateStatusCmd)

	// CDN List 命令参数
	cdnListCmd.Flags().StringVarP(&cdnListProfile, "profile", "p", "", "使用指定的 AWS profile")

	// CDN Get 命令参数
	cdnGetCmd.Flags().StringVarP(&cdnGetProfile, "profile", "p", "", "使用指定的 AWS profile")

	// CDN Create 命令参数
	cdnCreateCmd.Flags().StringVarP(&cdnCreateProfile, "profile", "p", "", "使用指定的 AWS profile")
	cdnCreateCmd.Flags().StringVar(&cdnCreateOrigin, "origin", "", "源站域名（必需）")
	cdnCreateCmd.Flags().StringSliceVarP(&cdnCreateAliases, "aliases", "a", []string{}, "自定义域名（可选，多个用逗号分隔）")
	cdnCreateCmd.Flags().StringVar(&cdnCreateComment, "comment", "", "备注说明（可选）")
	cdnCreateCmd.Flags().BoolVar(&cdnCreateEnabled, "enabled", true, "是否启用分发（默认: true）")
	cdnCreateCmd.Flags().StringVar(&cdnCreateCertificateARN, "certificate-arn", "", "SSL 证书 ARN（可选，使用自定义域名时需要）")
	cdnCreateCmd.Flags().StringVar(&cdnCreatePriceClass, "price-class", "", "价格等级（可选: PriceClass_100, PriceClass_200, PriceClass_All）")
	cdnCreateCmd.Flags().StringVar(&cdnCreateDefaultRootObject, "default-root-object", "", "默认根对象（可选，如: index.html）")
	cdnCreateCmd.MarkFlagRequired("origin")

	// CDN Update 命令参数
	cdnUpdateCmd.Flags().StringVarP(&cdnUpdateProfile, "profile", "p", "", "使用指定的 AWS profile")
	cdnUpdateCmd.Flags().StringVar(&cdnUpdateComment, "comment", "", "更新备注说明")
	cdnUpdateCmd.Flags().BoolVar(new(bool), "enabled", false, "启用分发")
	cdnUpdateCmd.Flags().BoolVar(new(bool), "disabled", false, "禁用分发")

	// CDN Invalidate 命令参数
	cdnInvalidateCmd.Flags().StringVarP(&cdnInvalidateProfile, "profile", "p", "", "使用指定的 AWS profile")
	cdnInvalidateCmd.Flags().StringSliceVar(&cdnInvalidatePaths, "paths", []string{}, "要失效的路径列表（必需，多个用逗号分隔）")
	cdnInvalidateCmd.Flags().StringVar(&cdnInvalidateCallerReference, "caller-reference", "", "调用者引用（可选，默认自动生成）")
	cdnInvalidateCmd.MarkFlagRequired("paths")

	// CDN Invalidate Status 命令参数
	cdnInvalidateStatusCmd.Flags().StringVarP(&cdnInvalidateStatusProfile, "profile", "p", "", "使用指定的 AWS profile")
}

// cdnListCmd 列出所有 CloudFront 分发
var cdnListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有 CloudFront 分发",
	Long: `列出 AWS 账户中的所有 CloudFront 分发。

使用示例:
  cloudctl aws cdn list                    # 使用默认 profile
  cloudctl aws cdn list -p aws-prod        # 使用指定 profile
  cloudctl aws cdn list -o json            # JSON 格式输出`,
	RunE: runCdnList,
}

// cdnGetCmd 获取指定分发的详细信息
var cdnGetCmd = &cobra.Command{
	Use:   "get <distribution-id>",
	Short: "获取 CloudFront 分发详情",
	Long: `获取指定 CloudFront 分发的详细信息。

使用示例:
  cloudctl aws cdn get E1234567890ABC           # 获取分发详情
  cloudctl aws cdn get E1234567890ABC -o json   # JSON 格式输出`,
	Args: cobra.ExactArgs(1),
	RunE: runCdnGet,
}

// cdnCreateCmd 创建 CloudFront 分发
var cdnCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建 CloudFront 分发",
	Long: `创建新的 CloudFront 分发。

使用示例:
  # 创建基本分发
  cloudctl aws cdn create --origin example.com

  # 创建带自定义域名的分发
  cloudctl aws cdn create --origin example.com \
    --aliases cdn.example.com \
    --certificate-arn arn:aws:acm:us-east-1:123456789012:certificate/xxx

  # 创建完整配置的分发
  cloudctl aws cdn create --origin example.com \
    --aliases cdn.example.com,www.example.com \
    --comment "My CDN Distribution" \
    --certificate-arn arn:aws:acm:us-east-1:123456789012:certificate/xxx \
    --price-class PriceClass_100 \
    --default-root-object index.html`,
	RunE: runCdnCreate,
}

// cdnUpdateCmd 更新 CloudFront 分发
var cdnUpdateCmd = &cobra.Command{
	Use:   "update <distribution-id>",
	Short: "更新 CloudFront 分发",
	Long: `更新 CloudFront 分发配置。

使用示例:
  # 更新备注
  cloudctl aws cdn update E1234567890ABC --comment "Updated comment"

  # 启用分发
  cloudctl aws cdn update E1234567890ABC --enabled

  # 禁用分发
  cloudctl aws cdn update E1234567890ABC --disabled`,
	Args: cobra.ExactArgs(1),
	RunE: runCdnUpdate,
}

// runCdnList 执行 CDN list 命令
func runCdnList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 创建 AWS 客户端
	client, err := aws.NewClient(cdnListProfile)
	if err != nil {
		return fmt.Errorf("创建 AWS 客户端失败: %w", err)
	}

	logger.Info("正在列出 CloudFront 分发...")

	// 列出分发
	distributions, err := client.ListDistributions(ctx)
	if err != nil {
		return fmt.Errorf("列出分发失败: %w", err)
	}

	if len(distributions) == 0 {
		fmt.Println("没有找到任何 CloudFront 分发")
		return nil
	}

	// 格式化输出
	formatter := GetFormatter()

	// 转换为输出格式
	data := make([]map[string]interface{}, len(distributions))
	for i, dist := range distributions {
		aliases := strings.Join(dist.Aliases, ", ")
		if aliases == "" {
			aliases = "-"
		}

		origins := ""
		if len(dist.Origins) > 0 {
			origins = dist.Origins[0].DomainName
			if len(dist.Origins) > 1 {
				origins += fmt.Sprintf(" (+%d)", len(dist.Origins)-1)
			}
		}

		data[i] = map[string]interface{}{
			"id":          dist.ID,
			"domain_name": dist.DomainName,
			"status":      dist.Status,
			"enabled":     dist.Enabled,
			"aliases":     aliases,
			"origins":     origins,
		}
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	return nil
}

// runCdnGet 执行 CDN get 命令
func runCdnGet(cmd *cobra.Command, args []string) error {
	distributionID := args[0]
	ctx := context.Background()

	// 创建 AWS 客户端
	client, err := aws.NewClient(cdnGetProfile)
	if err != nil {
		return fmt.Errorf("创建 AWS 客户端失败: %w", err)
	}

	logger.Info("正在获取分发详情...", "id", distributionID)

	// 获取分发详情
	dist, err := client.GetDistribution(ctx, distributionID)
	if err != nil {
		return fmt.Errorf("获取分发失败: %w", err)
	}

	// 格式化输出
	formatter := GetFormatter()

	// 转换为输出格式
	data := map[string]interface{}{
		"id":          dist.ID,
		"domain_name": dist.DomainName,
		"status":      dist.Status,
		"enabled":     dist.Enabled,
		"aliases":     dist.Aliases,
		"origins":     dist.Origins,
		"comment":     dist.Comment,
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	return nil
}

// runCdnCreate 执行 CDN create 命令
func runCdnCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 验证参数
	if cdnCreateOrigin == "" {
		return fmt.Errorf("必须指定源站域名 (--origin)")
	}

	// 如果有自定义域名但没有证书，给出警告
	if len(cdnCreateAliases) > 0 && cdnCreateCertificateARN == "" {
		logger.Warn("使用自定义域名时建议配置 SSL 证书 (--certificate-arn)")
	}

	// 创建 AWS 客户端
	client, err := aws.NewClient(cdnCreateProfile)
	if err != nil {
		return fmt.Errorf("创建 AWS 客户端失败: %w", err)
	}

	logger.Info("正在创建 CloudFront 分发...", "origin", cdnCreateOrigin)

	// 准备创建参数
	input := &aws.CreateDistributionInput{
		OriginDomain:      cdnCreateOrigin,
		Aliases:           cdnCreateAliases,
		Comment:           cdnCreateComment,
		Enabled:           cdnCreateEnabled,
		CertificateARN:    cdnCreateCertificateARN,
		PriceClass:        cdnCreatePriceClass,
		DefaultRootObject: cdnCreateDefaultRootObject,
	}

	// 创建分发
	dist, err := client.CreateDistribution(ctx, input)
	if err != nil {
		return fmt.Errorf("创建分发失败: %w", err)
	}

	logger.Info("成功创建 CloudFront 分发", "id", dist.ID)

	// 输出结果
	formatter := GetFormatter()

	// 转换为输出格式
	data := map[string]interface{}{
		"id":          dist.ID,
		"domain_name": dist.DomainName,
		"status":      dist.Status,
		"enabled":     dist.Enabled,
		"aliases":     dist.Aliases,
		"comment":     dist.Comment,
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Printf("\n注意: CloudFront 分发部署通常需要 10-15 分钟\n")
	fmt.Printf("可以使用以下命令检查部署状态:\n")
	fmt.Printf("  cloudctl aws cdn get %s\n", dist.ID)

	return nil
}

// runCdnUpdate 执行 CDN update 命令
func runCdnUpdate(cmd *cobra.Command, args []string) error {
	distributionID := args[0]
	ctx := context.Background()

	// 检查是否有更新参数
	hasUpdate := false
	input := &aws.UpdateDistributionInput{
		DistributionID: distributionID,
	}

	if cmd.Flags().Changed("comment") {
		input.Comment = &cdnUpdateComment
		hasUpdate = true
	}

	if cmd.Flags().Changed("enabled") {
		enabled := true
		input.Enabled = &enabled
		hasUpdate = true
	}

	if cmd.Flags().Changed("disabled") {
		disabled := false
		input.Enabled = &disabled
		hasUpdate = true
	}

	if !hasUpdate {
		return fmt.Errorf("请至少指定一个要更新的参数")
	}

	// 创建 AWS 客户端
	client, err := aws.NewClient(cdnUpdateProfile)
	if err != nil {
		return fmt.Errorf("创建 AWS 客户端失败: %w", err)
	}

	logger.Info("正在更新 CloudFront 分发...", "id", distributionID)

	// 更新分发
	dist, err := client.UpdateDistribution(ctx, input)
	if err != nil {
		return fmt.Errorf("更新分发失败: %w", err)
	}

	logger.Info("成功更新 CloudFront 分发", "id", dist.ID)

	// 输出结果
	formatter := GetFormatter()

	// 转换为输出格式
	data := map[string]interface{}{
		"id":      dist.ID,
		"status":  dist.Status,
		"enabled": dist.Enabled,
		"comment": dist.Comment,
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Printf("\n注意: 配置更新需要一定时间才能生效\n")

	return nil
}

// cdnInvalidateCmd 创建缓存失效
var cdnInvalidateCmd = &cobra.Command{
	Use:   "invalidate <distribution-id>",
	Short: "创建 CloudFront 缓存失效",
	Long: `创建 CloudFront 缓存失效请求，清除指定路径的缓存。

使用示例:
  # 失效单个路径
  cloudctl aws cdn invalidate E1234567890ABC --paths "/*"

  # 失效多个路径
  cloudctl aws cdn invalidate E1234567890ABC --paths "/index.html,/css/*,/js/*"

  # 指定 caller reference
  cloudctl aws cdn invalidate E1234567890ABC --paths "/*" --caller-reference "my-invalidation-001"

  # 使用指定 profile
  cloudctl aws cdn invalidate E1234567890ABC --paths "/*" -p aws-prod`,
	Args: cobra.ExactArgs(1),
	RunE: runCdnInvalidate,
}

// runCdnInvalidate 执行 CDN invalidate 命令
func runCdnInvalidate(cmd *cobra.Command, args []string) error {
	distributionID := args[0]
	ctx := context.Background()

	// 验证参数
	if len(cdnInvalidatePaths) == 0 {
		return fmt.Errorf("必须指定至少一个路径 (--paths)")
	}

	// 创建 AWS 客户端
	client, err := aws.NewClient(cdnInvalidateProfile)
	if err != nil {
		return fmt.Errorf("创建 AWS 客户端失败: %w", err)
	}

	logger.Info("正在创建缓存失效...", "distribution_id", distributionID, "paths", cdnInvalidatePaths)

	// 准备创建参数
	input := &aws.CreateInvalidationInput{
		DistributionID:  distributionID,
		Paths:           cdnInvalidatePaths,
		CallerReference: cdnInvalidateCallerReference,
	}

	// 创建缓存失效
	invalidation, err := client.CreateInvalidation(ctx, input)
	if err != nil {
		return fmt.Errorf("创建缓存失效失败: %w", err)
	}

	logger.Info("成功创建缓存失效", "id", invalidation.ID, "status", invalidation.Status)

	// 输出结果
	formatter := GetFormatter()

	// 转换为输出格式
	data := map[string]interface{}{
		"id":               invalidation.ID,
		"status":           invalidation.Status,
		"create_time":      invalidation.CreateTime.Format("2006-01-02 15:04:05"),
		"caller_reference": invalidation.CallerReference,
		"paths":            invalidation.Paths,
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Printf("\n✓ 缓存失效请求已创建\n")
	fmt.Printf("失效 ID: %s\n", invalidation.ID)
	fmt.Printf("状态: %s\n", invalidation.Status)
	fmt.Printf("\n注意: 缓存失效通常需要 10-15 分钟才能完成\n")
	fmt.Printf("可以使用以下命令检查失效状态:\n")
	fmt.Printf("  cloudctl aws cdn invalidate-status %s %s\n", distributionID, invalidation.ID)

	return nil
}

// cdnInvalidateStatusCmd 查询缓存失效状态
var cdnInvalidateStatusCmd = &cobra.Command{
	Use:   "invalidate-status <distribution-id> <invalidation-id>",
	Short: "查询 CloudFront 缓存失效状态",
	Long: `查询 CloudFront 缓存失效请求的状态。

使用示例:
  # 查询失效状态
  cloudctl aws cdn invalidate-status E1234567890ABC I2J3K4L5M6N7O8P9Q0

  # JSON 格式输出
  cloudctl aws cdn invalidate-status E1234567890ABC I2J3K4L5M6N7O8P9Q0 -o json

  # 使用指定 profile
  cloudctl aws cdn invalidate-status E1234567890ABC I2J3K4L5M6N7O8P9Q0 -p aws-prod`,
	Args: cobra.ExactArgs(2),
	RunE: runCdnInvalidateStatus,
}

// runCdnInvalidateStatus 执行 CDN invalidate-status 命令
func runCdnInvalidateStatus(cmd *cobra.Command, args []string) error {
	distributionID := args[0]
	invalidationID := args[1]
	ctx := context.Background()

	// 创建 AWS 客户端
	client, err := aws.NewClient(cdnInvalidateStatusProfile)
	if err != nil {
		return fmt.Errorf("创建 AWS 客户端失败: %w", err)
	}

	logger.Info("正在查询缓存失效状态...", "distribution_id", distributionID, "invalidation_id", invalidationID)

	// 获取失效状态
	invalidation, err := client.GetInvalidation(ctx, distributionID, invalidationID)
	if err != nil {
		return fmt.Errorf("获取缓存失效状态失败: %w", err)
	}

	logger.Info("成功获取缓存失效状态", "id", invalidation.ID, "status", invalidation.Status)

	// 输出结果
	formatter := GetFormatter()

	// 转换为输出格式
	data := map[string]interface{}{
		"id":               invalidation.ID,
		"status":           invalidation.Status,
		"create_time":      invalidation.CreateTime.Format("2006-01-02 15:04:05"),
		"caller_reference": invalidation.CallerReference,
		"paths":            invalidation.Paths,
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	// 根据状态显示提示信息
	fmt.Printf("\n")
	switch invalidation.Status {
	case "InProgress":
		fmt.Printf("⏳ 缓存失效正在进行中...\n")
		fmt.Printf("通常需要 10-15 分钟才能完成\n")
	case "Completed":
		fmt.Printf("✓ 缓存失效已完成\n")
	default:
		fmt.Printf("状态: %s\n", invalidation.Status)
	}

	return nil
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ado1t/cloudctl/internal/cloudflare"
	"github.com/ado1t/cloudctl/internal/logger"
	"github.com/ado1t/cloudctl/internal/output"
)

func init() {
	// 添加 zone 子命令
	cfZoneCmd.AddCommand(cfZoneListCmd)
	cfZoneCmd.AddCommand(cfZoneCreateCmd)
}

// cfZoneListCmd 列出所有 Zone
var cfZoneListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有域名",
	Long: `列出 Cloudflare 账户下的所有域名 (Zone)。

显示信息包括:
  - 域名名称
  - 状态 (active/pending)
  - Zone ID
  - 创建时间

使用示例:
  cloudctl cf zone list                    # 使用默认 profile
  cloudctl cf zone list --profile cf-prod  # 使用指定 profile
  cloudctl cf zone list -o json            # JSON 格式输出`,
	RunE: runZoneList,
}

// runZoneList 执行 zone list 命令
func runZoneList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 获取 profile 参数
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

	// 列出所有 Zone
	logger.Info("正在获取 Zone 列表...")
	zones, err := client.ListZones(ctx)
	if err != nil {
		logger.Error("获取 Zone 列表失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	if len(zones) == 0 {
		logger.Info("未找到任何 Zone")
		fmt.Println("未找到任何域名")
		return nil
	}

	// 获取格式化器
	formatter := GetFormatter()

	// 转换为输出格式
	data := make([]map[string]interface{}, len(zones))
	for i, zone := range zones {
		data[i] = map[string]interface{}{
			"name":       zone.Name,
			"status":     zone.Status,
			"id":         zone.ID,
			"created_on": zone.CreatedOn.Format("2006-01-02 15:04:05"),
		}
	}

	// 输出结果
	if err := formatter.Format(data); err != nil {
		logger.Error("格式化输出失败", "error", err)
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	logger.Info("成功列出 Zone", "total", len(zones))
	return nil
}

// cfZoneCreateCmd 创建 Zone
var cfZoneCreateCmd = &cobra.Command{
	Use:   "create <domain>[,domain...]",
	Short: "创建域名",
	Long: `在 Cloudflare 账户下创建新的域名 (Zone)。

支持同时创建多个域名，使用逗号分隔。

创建成功后会返回:
  - Zone ID
  - Name Server 信息

使用示例:
  cloudctl cf zone create example.com                              # 创建单个域名
  cloudctl cf zone create example.com,test.com,demo.com            # 创建多个域名（逗号分隔）
  cloudctl cf zone create example.com --profile cf-prod            # 使用指定 profile
  cloudctl cf zone create example.com,test.com -o json             # JSON 格式输出`,
	Args: cobra.ExactArgs(1),
	RunE: runZoneCreate,
}

// runZoneCreate 执行 zone create 命令
func runZoneCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 解析逗号分隔的域名
	var domains []string
	for _, arg := range args {
		// 按逗号分割
		parts := strings.Split(arg, ",")
		for _, domain := range parts {
			domain = strings.TrimSpace(domain)
			if domain != "" {
				domains = append(domains, domain)
			}
		}
	}

	if len(domains) == 0 {
		return fmt.Errorf("请提供至少一个域名")
	}

	// 获取 profile 参数
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

	// 批量创建 Zone
	var createdZones []map[string]interface{}
	var failedDomains []string

	for _, domain := range domains {
		logger.Info("正在创建 Zone...", "domain", domain)
		zone, err := client.CreateZone(ctx, domain)
		if err != nil {
			logger.Error("创建 Zone 失败", "domain", domain, "error", err)
			failedDomains = append(failedDomains, domain)
			continue
		}

		logger.Info("成功创建 Zone", "zone_id", zone.ID, "name", zone.Name)

		createdZones = append(createdZones, map[string]interface{}{
			"name":         zone.Name,
			"status":       zone.Status,
			"zone_id":      zone.ID,
			"name_servers": zone.NameServers,
			"created_on":   zone.CreatedOn.Format("2006-01-02 15:04:05"),
		})
	}

	// 如果全部失败，返回错误
	if len(createdZones) == 0 {
		return fmt.Errorf("所有域名创建失败")
	}

	// 获取格式化器
	formatter := GetFormatter()

	// 输出结果
	var outputData interface{}
	if len(createdZones) == 1 {
		// 单个域名，输出为 map
		outputData = createdZones[0]
	} else {
		// 多个域名，输出为数组
		outputData = createdZones
	}

	if err := formatter.Format(outputData); err != nil {
		logger.Error("格式化输出失败", "error", err)
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	// 如果是表格输出，额外显示 Name Server 配置提示
	if _, ok := formatter.(*output.TableFormatter); ok {
		if len(createdZones) == 1 {
			fmt.Println("\n请在域名注册商处配置以下 Name Server:")
			if ns, ok := createdZones[0]["name_servers"].([]string); ok {
				for _, server := range ns {
					fmt.Printf("  - %s\n", server)
				}
			}
		} else {
			fmt.Println("\n请在域名注册商处配置相应的 Name Server")
		}
	}

	// 显示失败的域名
	if len(failedDomains) > 0 {
		fmt.Fprintf(os.Stderr, "\n以下域名创建失败:\n")
		for _, domain := range failedDomains {
			fmt.Fprintf(os.Stderr, "  - %s\n", domain)
		}
	}

	logger.Info("批量创建完成", "成功", len(createdZones), "失败", len(failedDomains))
	return nil
}

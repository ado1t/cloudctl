package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ado1t/cloudctl/internal/cloudflare"
	"github.com/ado1t/cloudctl/internal/logger"
)

func init() {
	// 添加 cache 子命令
	cfCacheCmd.AddCommand(cfCachePurgeCmd)

	// purge 命令参数
	cfCachePurgeCmd.Flags().Bool("purge-all", false, "清除所有缓存")
	cfCachePurgeCmd.Flags().StringSliceP("files", "f", nil, "要清除的文件 URL 列表（逗号分隔）")
	cfCachePurgeCmd.Flags().StringSlice("prefixes", nil, "要清除的 URL 前缀/目录列表（逗号分隔）")
	cfCachePurgeCmd.Flags().StringSliceP("tags", "t", nil, "要清除的 Cache Tag 列表（逗号分隔，企业版功能）")
	cfCachePurgeCmd.Flags().StringSlice("hosts", nil, "要清除的主机名列表（逗号分隔）")
}

// cfCachePurgeCmd 清除缓存命令
var cfCachePurgeCmd = &cobra.Command{
	Use:   "purge <domain>",
	Short: "清除 Cloudflare 缓存",
	Long: `清除 Cloudflare 缓存。

支持多种清除模式：
  1. 清除所有缓存（--purge-all）
  2. 按文件清除（--files）
  3. 按目录/前缀清除（--prefixes）
  4. 按 Cache Tag 清除（--tags，企业版功能）
  5. 按主机名清除（--hosts）

注意：一次只能使用一种清除模式。

使用示例:
  # 清除所有缓存
  cloudctl cf cache purge example.com --purge-all

  # 清除指定文件
  cloudctl cf cache purge example.com --files /path/to/file.html
  cloudctl cf cache purge example.com -f /css/style.css,/js/app.js
  cloudctl cf cache purge example.com -f https://www.example.com/image.png

  # 清除指定目录/前缀（你的需求）
  cloudctl cf cache purge example.com --prefixes /foo/bar/
  cloudctl cf cache purge example.com --prefixes /images/,/static/
  cloudctl cf cache purge example.com --prefixes https://www.51rainbow.xyz/foo/bar/

  # 清除指定主机名
  cloudctl cf cache purge example.com --hosts www.example.com,api.example.com

  # 清除指定标签（企业版）
  cloudctl cf cache purge example.com --tags tag1,tag2

  # 查看详细日志
  cloudctl cf cache purge example.com --purge-all -vv`,
	Args: cobra.ExactArgs(1),
	RunE: runCachePurge,
}

// runCachePurge 执行缓存清除命令
func runCachePurge(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	domain := args[0]

	// 获取参数
	profile, _ := cmd.Flags().GetString("profile")
	purgeAll, _ := cmd.Flags().GetBool("purge-all")
	files, _ := cmd.Flags().GetStringSlice("files")
	prefixes, _ := cmd.Flags().GetStringSlice("prefixes")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	hosts, _ := cmd.Flags().GetStringSlice("hosts")

	logger.Info("准备清除缓存", "domain", domain)

	// 创建 Cloudflare 客户端
	logger.Debug("创建 Cloudflare 客户端", "profile", profile)
	client, err := cloudflare.NewClient(profile, logger.Logger)
	if err != nil {
		logger.Error("创建客户端失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}
	defer client.Close()

	// 构建清除参数
	params := cloudflare.CachePurgeParams{
		PurgeEverything: purgeAll,
		Files:           files,
		Prefixes:        prefixes,
		Tags:            tags,
		Hosts:           hosts,
	}

	// 执行清除
	logger.Debug("执行缓存清除", "params", fmt.Sprintf("%+v", params))
	result, err := client.PurgeCache(ctx, domain, params)
	if err != nil {
		logger.Error("清除缓存失败", "error", err)
		fmt.Fprintln(os.Stderr, cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	// 输出结果
	formatter := GetFormatter()

	// 准备输出数据
	data := map[string]interface{}{
		"success":    result.Success,
		"purge_type": result.PurgeType,
		"message":    result.Message,
	}

	if result.ID != "" {
		data["purge_id"] = result.ID
	}

	if result.ItemCount > 0 {
		data["item_count"] = result.ItemCount
	}

	// 添加详细信息
	switch result.PurgeType {
	case "files":
		data["files"] = strings.Join(files, ", ")
	case "prefixes":
		data["prefixes"] = strings.Join(prefixes, ", ")
	case "tags":
		data["tags"] = strings.Join(tags, ", ")
	case "hosts":
		data["hosts"] = strings.Join(hosts, ", ")
	}

	if err := formatter.Format(data); err != nil {
		logger.Error("格式化输出失败", "error", err)
		fmt.Fprintln(os.Stderr, "格式化输出失败:", err)
		os.Exit(1)
	}

	logger.Info("缓存清除完成", "type", result.PurgeType)
	return nil
}

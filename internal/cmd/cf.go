package cmd

import (
	"github.com/spf13/cobra"
)

// cfCmd 表示 Cloudflare 命令
var cfCmd = &cobra.Command{
	Use:   "cf",
	Short: "Cloudflare 资源管理",
	Long: `管理 Cloudflare 资源，包括域名、DNS 记录和缓存。

可用命令:
  cloudctl cf zone    - 域名管理
  cloudctl cf dns     - DNS 记录管理
  cloudctl cf cache   - 缓存管理

使用示例:
  cloudctl cf zone list              # 列出所有域名
  cloudctl cf zone create example.com # 创建域名
  cloudctl cf dns list example.com   # 列出 DNS 记录
  cloudctl cf cache purge example.com --purge-all # 清除所有缓存`,
}

func init() {
	// 添加 Cloudflare 子命令
	cfCmd.AddCommand(cfZoneCmd)
	cfCmd.AddCommand(cfDnsCmd)
	cfCmd.AddCommand(cfCacheCmd)
}

// cfZoneCmd 表示 zone 命令
var cfZoneCmd = &cobra.Command{
	Use:   "zone",
	Short: "域名管理",
	Long:  `管理 Cloudflare 域名 (Zone)`,
}

// cfDnsCmd 表示 dns 命令
var cfDnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS 记录管理",
	Long:  `管理 Cloudflare DNS 记录`,
}

// cfCacheCmd 表示 cache 命令
var cfCacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "缓存管理",
	Long:  `管理 Cloudflare 缓存`,
}

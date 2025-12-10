package cmd

import (
	"github.com/spf13/cobra"
)

// awsCmd 表示 AWS 命令
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "AWS 资源管理",
	Long: `管理 AWS 资源，包括 CloudFront CDN 和 ACM 证书。

可用命令:
  cloudctl aws cdn    - CloudFront CDN 管理
  cloudctl aws cert   - ACM 证书管理

使用示例:
  cloudctl aws cdn list                           # 列出所有 CloudFront 分发
  cloudctl aws cdn create --origin example.com    # 创建 CloudFront 分发
  cloudctl aws cdn invalidate E123 --paths "/*"   # 创建缓存失效
  cloudctl aws cert list                          # 列出所有证书
  cloudctl aws cert request -d example.com        # 申请证书`,
}

func init() {
	// 添加 AWS 子命令
	awsCmd.AddCommand(awsCdnCmd)
	awsCmd.AddCommand(awsCertCmd)
}

// awsCdnCmd 表示 cdn 命令
var awsCdnCmd = &cobra.Command{
	Use:   "cdn",
	Short: "CloudFront CDN 管理",
	Long:  `管理 AWS CloudFront CDN 分发`,
}

// awsCertCmd 表示 cert 命令
var awsCertCmd = &cobra.Command{
	Use:   "cert",
	Short: "ACM 证书管理",
	Long:  `管理 AWS ACM 证书`,
}

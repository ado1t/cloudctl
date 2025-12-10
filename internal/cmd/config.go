package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ado1t/cloudctl/internal/config"
)

// configCmd 表示 config 命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long: `管理 cloudctl 配置文件和 profile。

可用命令:
  cloudctl config validate       - 验证配置文件
  cloudctl config show           - 显示当前配置
  cloudctl config list-profiles  - 列出所有 profile

使用示例:
  cloudctl config validate              # 验证配置文件格式
  cloudctl config show                  # 显示当前配置
  cloudctl config show --profile cf-dev # 显示指定 profile 的配置
  cloudctl config list-profiles         # 列出所有可用的 profile`,
}

func init() {
	// 添加 config 子命令
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configListProfilesCmd)
}

// configValidateCmd 表示 validate 命令
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "验证配置文件",
	Long:  `验证配置文件格式是否正确，检查必需字段是否存在`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取配置文件路径
		configPath, _ := cmd.Flags().GetString("config")

		// 加载配置
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 验证配置
		if err := config.Validate(cfg); err != nil {
			return fmt.Errorf("配置验证失败: %w", err)
		}

		fmt.Println("✓ 配置文件验证通过")
		return nil
	},
}

// configShowCmd 表示 show 命令
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Long:  `显示当前使用的配置，隐藏敏感信息（API Token 等）`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取配置文件路径
		configPath, _ := cmd.Flags().GetString("config")
		profileName, _ := cmd.Flags().GetString("profile")

		// 加载配置
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 显示配置信息
		fmt.Println("配置信息:")
		fmt.Println("─────────────────────────────────────")

		// 默认 profile
		fmt.Printf("默认 Profile:\n")
		fmt.Printf("  Cloudflare: %s\n", cfg.DefaultProfile.Cloudflare)
		fmt.Printf("  AWS:        %s\n", cfg.DefaultProfile.AWS)
		fmt.Println()

		// 如果指定了 profile，只显示该 profile
		if profileName != "" {
			// 尝试显示 Cloudflare profile
			if cfProfile, err := config.GetCloudflareProfileSafe(profileName); err == nil {
				fmt.Printf("Cloudflare Profile '%s':\n", profileName)
				for k, v := range cfProfile {
					fmt.Printf("  %s: %s\n", k, v)
				}
				fmt.Println()
			}

			// 尝试显示 AWS profile
			if awsProfile, err := config.GetAWSProfileSafe(profileName); err == nil {
				fmt.Printf("AWS Profile '%s':\n", profileName)
				for k, v := range awsProfile {
					fmt.Printf("  %s: %s\n", k, v)
				}
				fmt.Println()
			}
		} else {
			// 显示所有 Cloudflare profiles
			if len(cfg.Cloudflare) > 0 {
				fmt.Println("Cloudflare Profiles:")
				for name := range cfg.Cloudflare {
					isDefault := ""
					if name == cfg.DefaultProfile.Cloudflare {
						isDefault = " (默认)"
					}
					profile, _ := config.GetCloudflareProfileSafe(name)
					fmt.Printf("  %s%s:\n", name, isDefault)
					for k, v := range profile {
						fmt.Printf("    %s: %s\n", k, v)
					}
				}
				fmt.Println()
			}

			// 显示所有 AWS profiles
			if len(cfg.AWS) > 0 {
				fmt.Println("AWS Profiles:")
				for name := range cfg.AWS {
					isDefault := ""
					if name == cfg.DefaultProfile.AWS {
						isDefault = " (默认)"
					}
					profile, _ := config.GetAWSProfileSafe(name)
					fmt.Printf("  %s%s:\n", name, isDefault)
					for k, v := range profile {
						fmt.Printf("    %s: %s\n", k, v)
					}
				}
				fmt.Println()
			}
		}

		// 日志配置
		fmt.Println("日志配置:")
		fmt.Printf("  级别:   %s\n", cfg.Log.Level)
		fmt.Printf("  格式:   %s\n", cfg.Log.Format)
		fmt.Printf("  输出:   %s\n", cfg.Log.Output)
		fmt.Println()

		// 输出配置
		fmt.Println("输出配置:")
		fmt.Printf("  格式:   %s\n", cfg.Output.Format)
		fmt.Printf("  颜色:   %v\n", cfg.Output.Color)
		fmt.Println()

		// API 配置
		fmt.Println("API 配置:")
		fmt.Printf("  超时:   %d 秒\n", cfg.API.Timeout)
		fmt.Printf("  重试:   %v\n", cfg.API.Retry.Enabled)
		if cfg.API.Retry.Enabled {
			fmt.Printf("    最大重试次数: %d\n", cfg.API.Retry.MaxAttempts)
			fmt.Printf("    初始延迟:     %d 秒\n", cfg.API.Retry.InitialDelay)
			fmt.Printf("    最大延迟:     %d 秒\n", cfg.API.Retry.MaxDelay)
		}

		return nil
	},
}

// configListProfilesCmd 表示 list-profiles 命令
var configListProfilesCmd = &cobra.Command{
	Use:   "list-profiles",
	Short: "列出所有 profile",
	Long:  `列出所有可用的 profile，标记当前默认 profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取配置文件路径
		configPath, _ := cmd.Flags().GetString("config")

		// 加载配置
		_, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 获取所有 profiles
		profiles := config.GetAllProfiles()

		if len(profiles) == 0 {
			fmt.Println("未找到任何 profile")
			fmt.Println("\n提示: 请在配置文件中添加 profile 或设置环境变量")
			return nil
		}

		// 按类型分组显示
		cfProfiles := []config.ProfileInfo{}
		awsProfiles := []config.ProfileInfo{}

		for _, p := range profiles {
			switch p.Type {
			case "cloudflare":
				cfProfiles = append(cfProfiles, p)
			case "aws":
				awsProfiles = append(awsProfiles, p)
			}
		}

		// 显示 Cloudflare profiles
		if len(cfProfiles) > 0 {
			fmt.Println("Cloudflare Profiles:")
			fmt.Println("─────────────────────────────────────")
			for _, p := range cfProfiles {
				status := []string{}
				if p.IsDefault {
					status = append(status, "默认")
				}
				if !p.HasCreds {
					status = append(status, "缺少凭证")
				}

				statusStr := ""
				if len(status) > 0 {
					statusStr = " [" + strings.Join(status, ", ") + "]"
				}

				fmt.Printf("  • %s%s\n", p.Name, statusStr)
			}
			fmt.Println()
		}

		// 显示 AWS profiles
		if len(awsProfiles) > 0 {
			fmt.Println("AWS Profiles:")
			fmt.Println("─────────────────────────────────────")
			for _, p := range awsProfiles {
				status := []string{}
				if p.IsDefault {
					status = append(status, "默认")
				}
				if !p.HasCreds {
					status = append(status, "缺少凭证")
				}

				statusStr := ""
				if len(status) > 0 {
					statusStr = " [" + strings.Join(status, ", ") + "]"
				}

				fmt.Printf("  • %s%s\n", p.Name, statusStr)
			}
			fmt.Println()
		}

		fmt.Printf("总计: %d 个 profile\n", len(profiles))
		return nil
	},
}

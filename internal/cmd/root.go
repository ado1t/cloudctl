package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ado1t/cloudctl/internal/config"
	"github.com/ado1t/cloudctl/internal/logger"
	"github.com/ado1t/cloudctl/internal/output"
)

var (
	version   string
	buildTime string
	// 全局格式化器，在 PersistentPreRunE 中初始化
	globalFormatter output.Formatter
)

// rootCmd 表示基础命令
var rootCmd = &cobra.Command{
	Use:   "cloudctl",
	Short: "cloudctl - 多云平台资源管理工具",
	Long: `cloudctl 是一个用于管理多云平台资源的命令行工具。

支持的云平台:
  - Cloudflare (域名、DNS、缓存管理)
  - AWS (CloudFront CDN、ACM 证书管理)

使用示例:
  cloudctl cf zone list              # 列出 Cloudflare 域名
  cloudctl cf dns create example.com # 创建 DNS 记录
  cloudctl aws cdn list              # 列出 CloudFront 分发
  cloudctl aws cert request          # 申请 ACM 证书

更多信息请访问: https://github.com/ado1t/cloudctl`,
	Version: version,
	// 禁用默认的 completion 命令
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},
	PersistentPreRunE: initialize,
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo 设置版本信息
func SetVersionInfo(ver, build string) {
	version = ver
	buildTime = build
	rootCmd.Version = fmt.Sprintf("%s (built at %s)", version, buildTime)
}

// initialize 初始化日志系统和输出格式化器
func initialize(cmd *cobra.Command, args []string) error {
	// 获取标志值
	quiet, _ := cmd.Flags().GetBool("quiet")
	noColor, _ := cmd.Flags().GetBool("no-color")
	logLevel, _ := cmd.Flags().GetString("log-level")
	verbosity, _ := cmd.Flags().GetCount("verbose")
	configPath, _ := cmd.Flags().GetString("config")
	outputFormat, _ := cmd.Flags().GetString("output")

	// 如果是静默模式，不初始化日志
	if quiet {
		return nil
	}

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		// 配置加载失败不影响命令执行，使用默认配置
		cfg = &config.Config{
			Log: config.LogConfig{
				Level:  "error",
				Format: "text",
				Output: "stdout",
			},
		}
	}

	// 确定日志级别和格式
	var level logger.LogLevel
	var addSource bool

	// 优先级: 命令行 log-level > 命令行 verbose > 配置文件
	if logLevel != "" {
		level = logger.ParseLevel(logLevel)
	} else if verbosity > 0 {
		level, addSource = logger.VerbosityToLevel(verbosity)
	} else {
		level = logger.ParseLevel(cfg.Log.Level)
		addSource = cfg.Log.AddSource
	}

	// 确定日志格式
	var format logger.LogFormat
	if cfg.Log.Format == "json" {
		format = logger.FormatJSON
	} else {
		format = logger.FormatText
	}

	// 确定输出目标
	var output *os.File
	switch cfg.Log.Output {
	case "stderr":
		output = os.Stderr
	case "stdout", "":
		output = os.Stdout
	default:
		// 如果是文件路径，打开文件
		f, err := os.OpenFile(cfg.Log.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// 打开文件失败，使用 stdout
			output = os.Stdout
		} else {
			output = f
		}
	}

	// 初始化日志系统
	logger.Init(logger.Config{
		Level:     level,
		Format:    format,
		Output:    output,
		AddSource: addSource,
		NoColor:   noColor || !cfg.Output.Color,
	})

	// 初始化输出格式化器
	initializeFormatter(outputFormat, noColor, quiet, cfg)

	return nil
}

// initializeFormatter 初始化输出格式化器
func initializeFormatter(outputFormat string, noColor, quiet bool, cfg *config.Config) {
	// 确定输出格式
	format := output.ParseFormat(outputFormat)
	if format == "" {
		// 如果命令行未指定，使用配置文件中的格式
		format = output.ParseFormat(cfg.Output.Format)
	}

	// 确定是否启用颜色
	enableColor := !noColor && cfg.Output.Color

	// 创建全局格式化器
	globalFormatter = output.New(output.Config{
		Format:  format,
		Writer:  os.Stdout,
		NoColor: !enableColor,
		Quiet:   quiet,
	})
}

// GetFormatter 获取全局格式化器
func GetFormatter() output.Formatter {
	return globalFormatter
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringP("config", "c", "", "配置文件路径 (默认: ~/.cloudctl/config.yaml)")
	rootCmd.PersistentFlags().StringP("profile", "p", "", "使用的 profile")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "输出格式 (table|json)")
	rootCmd.PersistentFlags().Bool("no-color", false, "禁用颜色输出")
	rootCmd.PersistentFlags().StringP("log-level", "l", "", "日志级别 (debug|info|warn|error)")
	rootCmd.PersistentFlags().CountP("verbose", "v", "详细程度级别 (-v: INFO, -vv: DEBUG, -vvv: DEBUG+源码)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "静默模式，完全不输出日志")

	// 添加子命令
	rootCmd.AddCommand(cfCmd)
	rootCmd.AddCommand(awsCmd)
	rootCmd.AddCommand(configCmd)
}

package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ado1t/cloudctl/internal/logger"
)

// 这是一个用于测试日志功能的示例程序
// 使用方法:
//   go run examples/test_log.go
//   go run examples/test_log.go -v
//   go run examples/test_log.go -vv
//   go run examples/test_log.go -vvv
//   go run examples/test_log.go --log-level debug
//   go run examples/test_log.go --no-color
//   go run examples/test_log.go -q

var (
	logLevel  string
	verbosity int
	noColor   bool
	quiet     bool
)

var rootCmd = &cobra.Command{
	Use:   "test-log",
	Short: "测试日志功能",
	Long:  `这是一个用于测试 cloudctl 日志系统的示例程序。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化日志系统
		initLogger()

		// 输出各个级别的日志
		logger.Debug("这是 DEBUG 级别的日志", "key1", "value1", "number", 42)
		logger.Info("这是 INFO 级别的日志", "key2", "value2", "enabled", true)
		logger.Warn("这是 WARN 级别的日志", "key3", "value3", "duration", "5s")
		logger.Error("这是 ERROR 级别的日志", "key4", "value4", "count", 100)

		// 测试带上下文的日志
		childLogger := logger.With("component", "test", "version", "1.0")
		if childLogger != nil {
			childLogger.Info("这是带有额外字段的日志", "status", "running")
		}

		// 测试分组日志
		groupLogger := logger.WithGroup("database")
		if groupLogger != nil {
			groupLogger.Info("数据库连接成功", "host", "localhost", "port", 5432)
		}
	},
}

func initLogger() {
	if quiet {
		return
	}

	// 确定日志级别
	var level logger.LogLevel
	var addSource bool

	if logLevel != "" {
		level = logger.ParseLevel(logLevel)
	} else if verbosity > 0 {
		level, addSource = logger.VerbosityToLevel(verbosity)
	} else {
		level = logger.LevelError
	}

	// 初始化日志系统
	logger.Init(logger.Config{
		Level:     level,
		Format:    logger.FormatText,
		Output:    os.Stdout,
		AddSource: addSource,
		NoColor:   noColor,
	})
}

func init() {
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "", "日志级别 (debug|info|warn|error)")
	rootCmd.Flags().CountVarP(&verbosity, "verbose", "v", "详细程度级别 (-v: INFO, -vv: DEBUG, -vvv: DEBUG+源码)")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "禁用颜色输出")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "静默模式，完全不输出日志")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ado1t/cloudctl/internal/cloudflare"
	"github.com/ado1t/cloudctl/internal/config"
	"github.com/ado1t/cloudctl/internal/logger"
)

func main() {
	// 命令行参数
	profile := flag.String("profile", "", "Cloudflare profile 名称")
	logLevel := flag.String("log-level", "info", "日志级别 (debug|info|warn|error)")
	noColor := flag.Bool("no-color", false, "禁用颜色输出")
	configPath := flag.String("config", "", "配置文件路径")
	flag.Parse()

	// 初始化日志
	level := logger.ParseLevel(*logLevel)
	logger.Init(logger.Config{
		Level:     level,
		Format:    logger.FormatText,
		Output:    os.Stdout,
		AddSource: false,
		NoColor:   *noColor,
	})
	log := logger.Logger

	log.Info("Cloudflare 客户端测试程序")

	// 加载配置
	log.Info("加载配置文件", "path", getConfigPath(*configPath))
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	log.Info("配置加载成功")

	// 显示可用的 Cloudflare profiles
	cfProfiles, _ := config.ListProfiles()
	if len(cfProfiles) > 0 {
		log.Info("可用的 Cloudflare profiles", "profiles", cfProfiles)
	} else {
		log.Warn("未找到 Cloudflare profiles")
		log.Info("请设置环境变量 CLOUDFLARE_API_TOKEN 或在配置文件中配置")
		os.Exit(1)
	}

	// 确定使用的 profile
	profileName := *profile
	if profileName == "" {
		profileName = cfg.DefaultProfile.Cloudflare
		log.Info("使用默认 profile", "profile", profileName)
	} else {
		log.Info("使用指定 profile", "profile", profileName)
	}

	// 创建 Cloudflare 客户端
	log.Info("创建 Cloudflare 客户端", "profile", profileName)
	client, err := cloudflare.NewClient(profileName, log)
	if err != nil {
		log.Error("创建客户端失败", "error", err)
		if cloudflare.IsAuthError(err) {
			log.Error("认证错误提示", "message", cloudflare.FormatError(err))
		}
		os.Exit(cloudflare.GetExitCode(err))
	}

	log.Info("Cloudflare 客户端创建成功")

	// 测试带重试的操作
	ctx := context.Background()
	log.Info("测试带重试的 API 操作")

	err = client.WithRetry(ctx, "测试操作", func() error {
		log.Debug("执行测试操作")
		// 这里可以调用实际的 API 操作
		// 为了测试，我们只是返回 nil
		return nil
	})

	if err != nil {
		log.Error("操作失败", "error", err)
		log.Error("格式化错误信息", "message", cloudflare.FormatError(err))
		os.Exit(cloudflare.GetExitCode(err))
	}

	log.Info("操作成功")

	// 测试错误类型
	log.Info("测试错误处理")
	testErrors(log)

	// 关闭客户端
	if err := client.Close(); err != nil {
		log.Error("关闭客户端失败", "error", err)
	}

	log.Info("测试完成")
}

// testErrors 测试各种错误类型
func testErrors(log *slog.Logger) {
	log.Info("=== 测试错误类型 ===")

	// 测试认证错误
	authErr := cloudflare.NewAuthError("测试操作", "无效的 API Token")
	log.Info("认证错误",
		"error", authErr,
		"is_auth_error", cloudflare.IsAuthError(authErr),
		"formatted", cloudflare.FormatError(authErr),
		"exit_code", cloudflare.GetExitCode(authErr),
	)

	// 测试资源不存在错误
	notFoundErr := cloudflare.NewNotFoundError("测试操作", "example.com")
	log.Info("资源不存在错误",
		"error", notFoundErr,
		"is_not_found_error", cloudflare.IsNotFoundError(notFoundErr),
		"formatted", cloudflare.FormatError(notFoundErr),
		"exit_code", cloudflare.GetExitCode(notFoundErr),
	)

	// 测试资源冲突错误
	conflictErr := cloudflare.NewConflictError("测试操作", "example.com")
	log.Info("资源冲突错误",
		"error", conflictErr,
		"is_conflict_error", cloudflare.IsConflictError(conflictErr),
		"formatted", cloudflare.FormatError(conflictErr),
		"exit_code", cloudflare.GetExitCode(conflictErr),
	)

	// 测试验证错误
	validationErr := cloudflare.NewValidationError("测试操作", "域名格式不正确")
	log.Info("验证错误",
		"error", validationErr,
		"is_validation_error", cloudflare.IsValidationError(validationErr),
		"formatted", cloudflare.FormatError(validationErr),
		"exit_code", cloudflare.GetExitCode(validationErr),
	)

	// 测试错误封装
	wrappedErr := cloudflare.WrapError(fmt.Errorf("network timeout"), "测试操作")
	log.Info("封装的网络错误",
		"error", wrappedErr,
		"is_network_error", cloudflare.IsNetworkError(wrappedErr),
		"formatted", cloudflare.FormatError(wrappedErr),
		"exit_code", cloudflare.GetExitCode(wrappedErr),
	)
}

// getConfigPath 获取配置文件路径
func getConfigPath(path string) string {
	if path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.cloudctl/config.yaml"
	}
	return home + "/.cloudctl/config.yaml"
}

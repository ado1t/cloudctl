package cloudflare

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cloudflare/cloudflare-go/v2"
	"github.com/cloudflare/cloudflare-go/v2/option"

	"github.com/ado1t/cloudctl/internal/config"
)

// Client Cloudflare 客户端封装
type Client struct {
	api     *cloudflare.Client
	profile string
	logger  *slog.Logger
}

// NewClient 创建新的 Cloudflare 客户端
func NewClient(profileName string, logger *slog.Logger) (*Client, error) {
	// 获取配置
	profile, err := config.GetCloudflareProfile(profileName)
	if err != nil {
		return nil, fmt.Errorf("获取 Cloudflare profile 失败: %w", err)
	}

	// 验证 API Token
	if profile.APIToken == "" {
		return nil, fmt.Errorf("%s", "Cloudflare API Token 未配置")
	}

	// 创建 Cloudflare API 客户端
	api := cloudflare.NewClient(
		option.WithAPIToken(profile.APIToken),
	)

	// 如果没有提供 logger，使用默认的
	if logger == nil {
		logger = slog.Default()
	}

	client := &Client{
		api:     api,
		profile: profileName,
		logger:  logger,
	}

	logger.Debug("Cloudflare 客户端初始化成功",
		"profile", profileName,
	)

	return client, nil
}

// API 返回底层的 Cloudflare API 客户端
func (c *Client) API() *cloudflare.Client {
	return c.api
}

// Profile 返回当前使用的 profile 名称
func (c *Client) Profile() string {
	return c.profile
}

// Logger 返回日志记录器
func (c *Client) Logger() *slog.Logger {
	return c.logger
}

// WithRetry 执行带重试的操作
func (c *Client) WithRetry(ctx context.Context, operation string, fn func() error) error {
	cfg := config.Get()
	if cfg == nil || !cfg.API.Retry.Enabled {
		// 不启用重试，直接执行
		return fn()
	}

	maxAttempts := cfg.API.Retry.MaxAttempts
	initialDelay := time.Duration(cfg.API.Retry.InitialDelay) * time.Second
	maxDelay := time.Duration(cfg.API.Retry.MaxDelay) * time.Second

	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		c.logger.Debug("执行 API 操作",
			"operation", operation,
			"attempt", attempt,
			"max_attempts", maxAttempts,
		)

		err := fn()
		if err == nil {
			if attempt > 1 {
				c.logger.Info("API 操作重试成功",
					"operation", operation,
					"attempt", attempt,
				)
			}
			return nil
		}

		lastErr = err

		// 检查是否应该重试
		if !shouldRetry(err) {
			c.logger.Debug("错误不可重试",
				"operation", operation,
				"error", err,
			)
			return WrapError(err, operation)
		}

		// 如果不是最后一次尝试，等待后重试
		if attempt < maxAttempts {
			c.logger.Warn("API 操作失败，准备重试",
				"operation", operation,
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"delay", delay,
				"error", err,
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// 指数退避
				delay *= 2
				if delay > maxDelay {
					delay = maxDelay
				}
			}
		}
	}

	c.logger.Error("API 操作失败，已达最大重试次数",
		"operation", operation,
		"max_attempts", maxAttempts,
		"error", lastErr,
	)

	return WrapError(lastErr, operation)
}

// shouldRetry 判断错误是否应该重试
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是网络错误或临时错误
	// 这里可以根据 Cloudflare API 的具体错误类型进行判断
	// 暂时对所有错误都进行重试
	return true
}

// Close 关闭客户端（预留方法，当前 cloudflare-go v2 不需要显式关闭）
func (c *Client) Close() error {
	c.logger.Debug("关闭 Cloudflare 客户端", "profile", c.profile)
	return nil
}

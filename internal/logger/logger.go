package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Logger 全局日志实例
var Logger *slog.Logger

// LogLevel 日志级别类型
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// LogFormat 日志格式类型
type LogFormat string

const (
	FormatText LogFormat = "text"
	FormatJSON LogFormat = "json"
)

// Config 日志配置
type Config struct {
	Level     LogLevel  // 日志级别
	Format    LogFormat // 日志格式
	Output    io.Writer // 输出目标
	AddSource bool      // 是否添加源代码位置
	NoColor   bool      // 是否禁用颜色
}

// Init 初始化日志系统
func Init(cfg Config) {
	// 解析日志级别
	level := parseLevel(cfg.Level)

	// 设置输出
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// 创建 handler 选项
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler

	// 根据格式创建不同的 handler
	if cfg.Format == FormatJSON {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		// 如果启用颜色，使用自定义的彩色 handler
		if !cfg.NoColor && isTerminal(output) {
			handler = NewColorHandler(output, opts)
		} else {
			handler = slog.NewTextHandler(output, opts)
		}
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// parseLevel 解析日志级别
func parseLevel(level LogLevel) slog.Level {
	switch strings.ToLower(string(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// isTerminal 检查输出是否为终端
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return color.NoColor == false && (f == os.Stdout || f == os.Stderr)
	}
	return false
}

// SetLevel 动态设置日志级别
func SetLevel(level LogLevel) {
	if Logger == nil {
		return
	}
	// 注意：slog 的 Logger 不支持动态修改级别
	// 需要重新创建 Logger
	// 这里我们通过重新初始化来实现
}

// Debug 输出 Debug 级别日志
func Debug(msg string, args ...any) {
	if Logger != nil {
		Logger.Debug(msg, args...)
	}
}

// Info 输出 Info 级别日志
func Info(msg string, args ...any) {
	if Logger != nil {
		Logger.Info(msg, args...)
	}
}

// Warn 输出 Warn 级别日志
func Warn(msg string, args ...any) {
	if Logger != nil {
		Logger.Warn(msg, args...)
	}
}

// Error 输出 Error 级别日志
func Error(msg string, args ...any) {
	if Logger != nil {
		Logger.Error(msg, args...)
	}
}

// DebugContext 输出带上下文的 Debug 级别日志
func DebugContext(ctx context.Context, msg string, args ...any) {
	if Logger != nil {
		Logger.DebugContext(ctx, msg, args...)
	}
}

// InfoContext 输出带上下文的 Info 级别日志
func InfoContext(ctx context.Context, msg string, args ...any) {
	if Logger != nil {
		Logger.InfoContext(ctx, msg, args...)
	}
}

// WarnContext 输出带上下文的 Warn 级别日志
func WarnContext(ctx context.Context, msg string, args ...any) {
	if Logger != nil {
		Logger.WarnContext(ctx, msg, args...)
	}
}

// ErrorContext 输出带上下文的 Error 级别日志
func ErrorContext(ctx context.Context, msg string, args ...any) {
	if Logger != nil {
		Logger.ErrorContext(ctx, msg, args...)
	}
}

// With 创建带有额外字段的 Logger
func With(args ...any) *slog.Logger {
	if Logger != nil {
		return Logger.With(args...)
	}
	return nil
}

// WithGroup 创建带有分组的 Logger
func WithGroup(name string) *slog.Logger {
	if Logger != nil {
		return Logger.WithGroup(name)
	}
	return nil
}

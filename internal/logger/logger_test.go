package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "文本格式",
			config: Config{
				Level:     LevelInfo,
				Format:    FormatText,
				Output:    &bytes.Buffer{},
				AddSource: false,
				NoColor:   true,
			},
		},
		{
			name: "JSON格式",
			config: Config{
				Level:     LevelDebug,
				Format:    FormatJSON,
				Output:    &bytes.Buffer{},
				AddSource: true,
				NoColor:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.config)
			if Logger == nil {
				t.Error("Logger 不应为 nil")
			}
		})
	}
}

func TestParseLevelInternal(t *testing.T) {
	tests := []struct {
		input    LogLevel
		expected slog.Level
	}{
		{LevelDebug, slog.LevelDebug},
		{LevelInfo, slog.LevelInfo},
		{LevelWarn, slog.LevelWarn},
		{LevelError, slog.LevelError},
		{"invalid", slog.LevelInfo}, // 默认值
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := parseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLevel(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:     LevelDebug,
		Format:    FormatText,
		Output:    buf,
		AddSource: false,
		NoColor:   true,
	})

	// 测试各个级别的日志
	Debug("debug message", "key", "value")
	Info("info message", "key", "value")
	Warn("warn message", "key", "value")
	Error("error message", "key", "value")

	output := buf.String()

	// 验证所有级别的日志都被输出
	if !strings.Contains(output, "debug message") {
		t.Error("Debug 日志未输出")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info 日志未输出")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn 日志未输出")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error 日志未输出")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:     LevelWarn,
		Format:    FormatText,
		Output:    buf,
		AddSource: false,
		NoColor:   true,
	})

	// 测试各个级别的日志
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()

	// Debug 和 Info 不应该被输出
	if strings.Contains(output, "debug message") {
		t.Error("Debug 日志不应该被输出")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info 日志不应该被输出")
	}

	// Warn 和 Error 应该被输出
	if !strings.Contains(output, "warn message") {
		t.Error("Warn 日志应该被输出")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error 日志应该被输出")
	}
}

func TestJSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:     LevelInfo,
		Format:    FormatJSON,
		Output:    buf,
		AddSource: false,
		NoColor:   true,
	})

	Info("test message", "key", "value")

	output := buf.String()

	// 验证 JSON 格式
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Error("JSON 格式不正确")
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Error("JSON 属性不正确")
	}
}

func TestWith(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{
		Level:     LevelInfo,
		Format:    FormatText,
		Output:    buf,
		AddSource: false,
		NoColor:   true,
	})

	// 创建带有额外字段的 logger
	childLogger := With("component", "test")
	if childLogger == nil {
		t.Fatal("With() 返回 nil")
	}

	childLogger.Info("test message")

	output := buf.String()

	// 验证额外字段被包含
	if !strings.Contains(output, "component") {
		t.Error("额外字段未被包含")
	}
}

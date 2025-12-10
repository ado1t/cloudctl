package output

import (
	"io"
)

// Format 定义支持的输出格式
type Format string

const (
	// FormatTable 表格格式
	FormatTable Format = "table"
	// FormatJSON JSON 格式
	FormatJSON Format = "json"
	// FormatYAML YAML 格式（预留）
	FormatYAML Format = "yaml"
	// FormatText 纯文本格式
	FormatText Format = "text"
)

// Formatter 定义输出格式化接口
type Formatter interface {
	// Format 格式化数据并输出
	Format(data interface{}) error
}

// Config 输出配置
type Config struct {
	// Format 输出格式
	Format Format
	// Writer 输出目标
	Writer io.Writer
	// NoColor 是否禁用颜色
	NoColor bool
	// Quiet 静默模式
	Quiet bool
}

// New 创建新的格式化器
func New(cfg Config) Formatter {
	if cfg.Quiet {
		return &NullFormatter{}
	}

	switch cfg.Format {
	case FormatJSON:
		return NewJSONFormatter(cfg.Writer)
	case FormatTable:
		return NewTableFormatter(cfg.Writer, !cfg.NoColor)
	case FormatText:
		return NewTextFormatter(cfg.Writer, !cfg.NoColor)
	default:
		return NewTableFormatter(cfg.Writer, !cfg.NoColor)
	}
}

// NullFormatter 空格式化器，用于静默模式
type NullFormatter struct{}

// Format 实现 Formatter 接口
func (f *NullFormatter) Format(data interface{}) error {
	return nil
}

// IsValidFormat 检查格式是否有效
func IsValidFormat(format string) bool {
	switch Format(format) {
	case FormatTable, FormatJSON, FormatYAML, FormatText:
		return true
	default:
		return false
	}
}

// ParseFormat 解析格式字符串
func ParseFormat(format string) Format {
	switch Format(format) {
	case FormatJSON:
		return FormatJSON
	case FormatYAML:
		return FormatYAML
	case FormatText:
		return FormatText
	case FormatTable:
		return FormatTable
	default:
		return FormatTable
	}
}

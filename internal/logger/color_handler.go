package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
)

// ColorHandler 彩色日志处理器
type ColorHandler struct {
	opts   *slog.HandlerOptions
	output io.Writer
	mu     sync.Mutex
	groups []string
	attrs  []slog.Attr
}

// NewColorHandler 创建彩色日志处理器
func NewColorHandler(output io.Writer, opts *slog.HandlerOptions) *ColorHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ColorHandler{
		opts:   opts,
		output: output,
		groups: []string{},
		attrs:  []slog.Attr{},
	}
}

// Enabled 判断是否启用指定级别的日志
func (h *ColorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle 处理日志记录
func (h *ColorHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 构建日志输出
	buf := make([]byte, 0, 1024)

	// 时间
	if !r.Time.IsZero() {
		buf = append(buf, color.New(color.FgHiBlack).Sprint(r.Time.Format("2006-01-02 15:04:05"))...)
		buf = append(buf, ' ')
	}

	// 日志级别（带颜色）
	levelStr := formatLevel(r.Level)
	buf = append(buf, levelStr...)
	buf = append(buf, ' ')

	// 源代码位置
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			src := fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
			buf = append(buf, color.New(color.FgHiBlack).Sprint(src)...)
			buf = append(buf, ' ')
		}
	}

	// 消息
	buf = append(buf, color.New(color.Bold).Sprint(r.Message)...)

	// 属性
	if len(h.attrs) > 0 || r.NumAttrs() > 0 {
		buf = append(buf, ' ')
		buf = h.appendAttrs(buf, h.attrs)
		r.Attrs(func(a slog.Attr) bool {
			buf = h.appendAttr(buf, a)
			return true
		})
	}

	buf = append(buf, '\n')

	_, err := h.output.Write(buf)
	return err
}

// WithAttrs 添加属性
func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &ColorHandler{
		opts:   h.opts,
		output: h.output,
		groups: h.groups,
		attrs:  newAttrs,
	}
}

// WithGroup 添加分组
func (h *ColorHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &ColorHandler{
		opts:   h.opts,
		output: h.output,
		groups: newGroups,
		attrs:  h.attrs,
	}
}

// formatLevel 格式化日志级别（带颜色）
func formatLevel(level slog.Level) string {
	var levelColor *color.Color
	var levelText string

	switch level {
	case slog.LevelDebug:
		levelColor = color.New(color.FgCyan)
		levelText = "DEBUG"
	case slog.LevelInfo:
		levelColor = color.New(color.FgGreen)
		levelText = "INFO "
	case slog.LevelWarn:
		levelColor = color.New(color.FgYellow)
		levelText = "WARN "
	case slog.LevelError:
		levelColor = color.New(color.FgRed, color.Bold)
		levelText = "ERROR"
	default:
		levelColor = color.New(color.FgWhite)
		levelText = level.String()
	}

	return levelColor.Sprintf("[%s]", levelText)
}

// appendAttrs 添加多个属性
func (h *ColorHandler) appendAttrs(buf []byte, attrs []slog.Attr) []byte {
	for _, a := range attrs {
		buf = h.appendAttr(buf, a)
	}
	return buf
}

// appendAttr 添加单个属性
func (h *ColorHandler) appendAttr(buf []byte, a slog.Attr) []byte {
	// 处理分组
	if len(h.groups) > 0 {
		for _, g := range h.groups {
			buf = append(buf, g...)
			buf = append(buf, '.')
		}
	}

	// 键
	buf = append(buf, color.New(color.FgCyan).Sprint(a.Key)...)
	buf = append(buf, '=')

	// 值
	buf = h.appendValue(buf, a.Value)
	buf = append(buf, ' ')

	return buf
}

// appendValue 添加值
func (h *ColorHandler) appendValue(buf []byte, v slog.Value) []byte {
	switch v.Kind() {
	case slog.KindString:
		buf = append(buf, color.New(color.FgYellow).Sprint(strconv.Quote(v.String()))...)
	case slog.KindInt64:
		buf = append(buf, color.New(color.FgMagenta).Sprint(v.Int64())...)
	case slog.KindUint64:
		buf = append(buf, color.New(color.FgMagenta).Sprint(v.Uint64())...)
	case slog.KindFloat64:
		buf = append(buf, color.New(color.FgMagenta).Sprint(v.Float64())...)
	case slog.KindBool:
		buf = append(buf, color.New(color.FgMagenta).Sprint(v.Bool())...)
	case slog.KindDuration:
		buf = append(buf, color.New(color.FgMagenta).Sprint(v.Duration())...)
	case slog.KindTime:
		buf = append(buf, color.New(color.FgYellow).Sprint(v.Time().Format(time.RFC3339))...)
	case slog.KindGroup:
		attrs := v.Group()
		if len(attrs) > 0 {
			buf = append(buf, '{')
			for i, a := range attrs {
				if i > 0 {
					buf = append(buf, ' ')
				}
				buf = h.appendAttr(buf, a)
			}
			buf = append(buf, '}')
		}
	default:
		buf = append(buf, fmt.Sprint(v.Any())...)
	}
	return buf
}

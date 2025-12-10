package output

import (
	"encoding/json"
	"io"
)

// JSONFormatter JSON 格式化器
type JSONFormatter struct {
	writer io.Writer
	indent bool
}

// NewJSONFormatter 创建 JSON 格式化器
func NewJSONFormatter(writer io.Writer) *JSONFormatter {
	return &JSONFormatter{
		writer: writer,
		indent: true,
	}
}

// NewJSONFormatterCompact 创建紧凑的 JSON 格式化器
func NewJSONFormatterCompact(writer io.Writer) *JSONFormatter {
	return &JSONFormatter{
		writer: writer,
		indent: false,
	}
}

// Format 实现 Formatter 接口
func (f *JSONFormatter) Format(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	if f.indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// FormatCompact 输出紧凑的 JSON
func (f *JSONFormatter) FormatCompact(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = f.writer.Write(bytes)
	if err != nil {
		return err
	}
	_, err = f.writer.Write([]byte("\n"))
	return err
}

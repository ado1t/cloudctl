package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "JSON格式化器",
			config: Config{
				Format: FormatJSON,
				Writer: &bytes.Buffer{},
			},
		},
		{
			name: "表格格式化器",
			config: Config{
				Format: FormatTable,
				Writer: &bytes.Buffer{},
			},
		},
		{
			name: "文本格式化器",
			config: Config{
				Format: FormatText,
				Writer: &bytes.Buffer{},
			},
		},
		{
			name: "静默模式",
			config: Config{
				Format: FormatTable,
				Writer: &bytes.Buffer{},
				Quiet:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := New(tt.config)
			if formatter == nil {
				t.Error("formatter 不应为 nil")
			}

			if tt.config.Quiet {
				if _, ok := formatter.(*NullFormatter); !ok {
					t.Error("静默模式应返回 NullFormatter")
				}
			}
		})
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		format string
		valid  bool
	}{
		{"table", true},
		{"json", true},
		{"yaml", true},
		{"text", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := IsValidFormat(tt.format)
			if result != tt.valid {
				t.Errorf("IsValidFormat(%s) = %v, want %v", tt.format, result, tt.valid)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"json", FormatJSON},
		{"table", FormatTable},
		{"yaml", FormatYAML},
		{"text", FormatText},
		{"invalid", FormatTable}, // 默认值
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseFormat(tt.input)
			if result != tt.expected {
				t.Errorf("ParseFormat(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNullFormatter(t *testing.T) {
	formatter := &NullFormatter{}
	err := formatter.Format(map[string]interface{}{"key": "value"})
	if err != nil {
		t.Errorf("NullFormatter.Format() error = %v", err)
	}
}

func TestJSONFormatter(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "简单map",
			data: map[string]interface{}{
				"name": "test",
				"age":  30,
			},
		},
		{
			name: "map切片",
			data: []map[string]interface{}{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			},
		},
		{
			name: "结构体",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "test",
				Age:  30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := NewJSONFormatter(buf)

			err := formatter.Format(tt.data)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			// 验证输出是有效的 JSON
			var result interface{}
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Errorf("输出不是有效的 JSON: %v", err)
			}
		})
	}
}

func TestTableFormatter(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "map切片",
			data: []map[string]interface{}{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 25},
			},
		},
		{
			name: "单个map",
			data: map[string]interface{}{
				"name": "test",
				"age":  30,
			},
		},
		{
			name: "interface切片",
			data: []interface{}{"item1", "item2", "item3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := NewTableFormatter(buf, false)

			err := formatter.Format(tt.data)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			output := buf.String()
			if output == "" {
				t.Error("输出不应为空")
			}

			// 验证输出包含分隔线
			if !strings.Contains(output, "---") {
				t.Error("表格输出应包含分隔线")
			}
		})
	}
}

func TestTableFormatterWithHeaders(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(buf, false)

	formatter.SetHeaders([]string{"Name", "Age"})
	formatter.AddRow([]string{"Alice", "30"})
	formatter.AddRow([]string{"Bob", "25"})

	err := formatter.Render()
	if err != nil {
		t.Errorf("Render() error = %v", err)
		return
	}

	output := buf.String()
	if !strings.Contains(output, "Name") || !strings.Contains(output, "Age") {
		t.Error("输出应包含表头")
	}
	if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") {
		t.Error("输出应包含数据行")
	}
}

func TestTextFormatter(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "字符串",
			data: "test string",
		},
		{
			name: "map",
			data: map[string]interface{}{
				"name": "test",
				"age":  30,
			},
		},
		{
			name: "map切片",
			data: []map[string]interface{}{
				{"name": "Alice"},
				{"name": "Bob"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := NewTextFormatter(buf, false)

			err := formatter.Format(tt.data)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			output := buf.String()
			if output == "" {
				t.Error("输出不应为空")
			}
		})
	}
}

func TestTableFormatterPrintMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(buf, false)

	formatter.PrintHeader("Test Header")
	formatter.PrintSuccess("Success message")
	formatter.PrintError("Error message")
	formatter.PrintWarning("Warning message")
	formatter.PrintInfo("Info message")

	output := buf.String()
	if !strings.Contains(output, "Test Header") {
		t.Error("输出应包含标题")
	}
	if !strings.Contains(output, "Success message") {
		t.Error("输出应包含成功消息")
	}
	if !strings.Contains(output, "Error message") {
		t.Error("输出应包含错误消息")
	}
}

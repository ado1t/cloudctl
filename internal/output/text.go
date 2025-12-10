package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/fatih/color"
)

// TextFormatter 文本格式化器
type TextFormatter struct {
	writer io.Writer
	color  bool
}

// NewTextFormatter 创建文本格式化器
func NewTextFormatter(writer io.Writer, enableColor bool) *TextFormatter {
	return &TextFormatter{
		writer: writer,
		color:  enableColor,
	}
}

// Format 实现 Formatter 接口
func (f *TextFormatter) Format(data interface{}) error {
	// 处理不同类型的数据
	switch v := data.(type) {
	case []map[string]interface{}:
		return f.formatMapSlice(v)
	case map[string]interface{}:
		return f.formatMap(v)
	case []interface{}:
		return f.formatInterfaceSlice(v)
	case string:
		fmt.Fprintln(f.writer, v)
		return nil
	default:
		return f.formatStruct(data)
	}
}

// formatMapSlice 格式化 map 切片
func (f *TextFormatter) formatMapSlice(data []map[string]interface{}) error {
	for i, item := range data {
		if i > 0 {
			fmt.Fprintln(f.writer)
		}
		if err := f.formatMap(item); err != nil {
			return err
		}
	}
	return nil
}

// formatMap 格式化单个 map
func (f *TextFormatter) formatMap(data map[string]interface{}) error {
	for key, value := range data {
		if f.color {
			color.New(color.FgCyan).Fprintf(f.writer, "%s: ", key)
			fmt.Fprintf(f.writer, "%v\n", value)
		} else {
			fmt.Fprintf(f.writer, "%s: %v\n", key, value)
		}
	}
	return nil
}

// formatInterfaceSlice 格式化 interface{} 切片
func (f *TextFormatter) formatInterfaceSlice(data []interface{}) error {
	// 尝试转换为 map 切片
	mapSlice := make([]map[string]interface{}, 0, len(data))
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			mapSlice = append(mapSlice, m)
		}
	}

	if len(mapSlice) > 0 {
		return f.formatMapSlice(mapSlice)
	}

	// 简单列表输出
	for _, item := range data {
		fmt.Fprintf(f.writer, "- %v\n", item)
	}
	return nil
}

// formatStruct 格式化结构体
func (f *TextFormatter) formatStruct(data interface{}) error {
	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	// 处理指针
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	// 处理切片
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if i > 0 {
				fmt.Fprintln(f.writer)
			}
			if err := f.formatStruct(val.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}

	// 单个结构体
	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			value := val.Field(i)

			// 跳过未导出的字段
			if !field.IsExported() {
				continue
			}

			fieldName := field.Name
			// 使用 json tag 作为字段名（如果存在）
			if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
				fieldName = strings.Split(tag, ",")[0]
			}

			if f.color {
				color.New(color.FgCyan).Fprintf(f.writer, "%s: ", fieldName)
				fmt.Fprintf(f.writer, "%v\n", value.Interface())
			} else {
				fmt.Fprintf(f.writer, "%s: %v\n", fieldName, value.Interface())
			}
		}
		return nil
	}

	// 其他类型，直接输出
	fmt.Fprintf(f.writer, "%v\n", data)
	return nil
}

// Print 打印文本
func (f *TextFormatter) Print(text string) {
	fmt.Fprint(f.writer, text)
}

// Println 打印文本并换行
func (f *TextFormatter) Println(text string) {
	fmt.Fprintln(f.writer, text)
}

// PrintColored 打印带颜色的文本
func (f *TextFormatter) PrintColored(text string, c color.Attribute) {
	if f.color {
		color.New(c).Fprint(f.writer, text)
	} else {
		fmt.Fprint(f.writer, text)
	}
}

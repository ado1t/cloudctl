package output

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/fatih/color"
)

// TableFormatter 表格格式化器
type TableFormatter struct {
	writer  io.Writer
	color   bool
	headers []string
	rows    [][]string
}

// NewTableFormatter 创建表格格式化器
func NewTableFormatter(writer io.Writer, enableColor bool) *TableFormatter {
	return &TableFormatter{
		writer: writer,
		color:  enableColor,
	}
}

// Format 实现 Formatter 接口
func (f *TableFormatter) Format(data interface{}) error {
	// 处理不同类型的数据
	switch v := data.(type) {
	case []map[string]interface{}:
		return f.formatMapSlice(v)
	case map[string]interface{}:
		return f.formatMap(v)
	case []interface{}:
		return f.formatInterfaceSlice(v)
	default:
		return f.formatStruct(data)
	}
}

// formatMapSlice 格式化 map 切片
func (f *TableFormatter) formatMapSlice(data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// 提取表头并按优先级排序，确保列顺序一致
	headers := make([]string, 0)
	for key := range data[0] {
		headers = append(headers, key)
	}

	// 使用自定义排序，常见字段优先
	sort.Slice(headers, func(i, j int) bool {
		priority := map[string]int{
			"name":        1,
			"status":      2,
			"id":          3,
			"zone_id":     3,
			"created_on":  4,
			"modified_on": 5,
		}

		pi, oki := priority[headers[i]]
		pj, okj := priority[headers[j]]

		// 如果都有优先级，按优先级排序
		if oki && okj {
			return pi < pj
		}
		// 如果只有一个有优先级，有优先级的排前面
		if oki {
			return true
		}
		if okj {
			return false
		}
		// 都没有优先级，按字母排序
		return headers[i] < headers[j]
	})

	// 计算每列的最大宽度
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// 准备数据行
	rows := make([][]string, len(data))
	for i, row := range data {
		rowData := make([]string, len(headers))
		for j, header := range headers {
			value := fmt.Sprintf("%v", row[header])
			rowData[j] = value
			if len(value) > colWidths[j] {
				colWidths[j] = len(value)
			}
		}
		rows[i] = rowData
	}

	// 输出表格
	f.printTable(headers, rows, colWidths)
	return nil
}

// formatMap 格式化单个 map（键值对形式）
func (f *TableFormatter) formatMap(data map[string]interface{}) error {
	headers := []string{"Key", "Value"}
	rows := make([][]string, 0, len(data))

	// 计算列宽
	colWidths := []int{len("Key"), len("Value")}
	for key, value := range data {
		valueStr := fmt.Sprintf("%v", value)
		rows = append(rows, []string{key, valueStr})
		if len(key) > colWidths[0] {
			colWidths[0] = len(key)
		}
		if len(valueStr) > colWidths[1] {
			colWidths[1] = len(valueStr)
		}
	}

	f.printTable(headers, rows, colWidths)
	return nil
}

// formatInterfaceSlice 格式化 interface{} 切片
func (f *TableFormatter) formatInterfaceSlice(data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

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
	headers := []string{"Value"}
	rows := make([][]string, len(data))
	colWidths := []int{len("Value")}

	for i, item := range data {
		value := fmt.Sprintf("%v", item)
		rows[i] = []string{value}
		if len(value) > colWidths[0] {
			colWidths[0] = len(value)
		}
	}

	f.printTable(headers, rows, colWidths)
	return nil
}

// formatStruct 格式化结构体
func (f *TableFormatter) formatStruct(data interface{}) error {
	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	// 处理指针
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	// 处理切片
	if val.Kind() == reflect.Slice {
		return f.formatStructSlice(data)
	}

	// 单个结构体，显示为键值对
	if val.Kind() == reflect.Struct {
		headers := []string{"Field", "Value"}
		rows := make([][]string, 0)
		colWidths := []int{len("Field"), len("Value")}

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

			valueStr := fmt.Sprintf("%v", value.Interface())
			rows = append(rows, []string{fieldName, valueStr})

			if len(fieldName) > colWidths[0] {
				colWidths[0] = len(fieldName)
			}
			if len(valueStr) > colWidths[1] {
				colWidths[1] = len(valueStr)
			}
		}

		f.printTable(headers, rows, colWidths)
		return nil
	}

	// 其他类型，直接输出
	fmt.Fprintf(f.writer, "%v\n", data)
	return nil
}

// formatStructSlice 格式化结构体切片
func (f *TableFormatter) formatStructSlice(data interface{}) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Len() == 0 {
		return nil
	}

	// 获取第一个元素的类型
	firstElem := val.Index(0)
	if firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}

	if firstElem.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct slice, got %v", firstElem.Kind())
	}

	typ := firstElem.Type()

	// 提取表头和字段索引
	headers := make([]string, 0)
	fieldIndices := make([]int, 0)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			fieldName = strings.Split(tag, ",")[0]
		}

		headers = append(headers, fieldName)
		fieldIndices = append(fieldIndices, i)
	}

	// 计算列宽
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// 准备数据行
	rows := make([][]string, val.Len())
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		rowData := make([]string, len(fieldIndices))
		for j, idx := range fieldIndices {
			value := fmt.Sprintf("%v", elem.Field(idx).Interface())
			rowData[j] = value
			if len(value) > colWidths[j] {
				colWidths[j] = len(value)
			}
		}
		rows[i] = rowData
	}

	f.printTable(headers, rows, colWidths)
	return nil
}

// printTable 打印表格
func (f *TableFormatter) printTable(headers []string, rows [][]string, colWidths []int) {
	// 打印表头
	f.printRow(headers, colWidths, true)

	// 打印分隔线
	f.printSeparator(colWidths)

	// 打印数据行
	for _, row := range rows {
		f.printRow(row, colWidths, false)
	}
}

// printRow 打印一行
func (f *TableFormatter) printRow(row []string, colWidths []int, isHeader bool) {
	parts := make([]string, len(row))
	for i, cell := range row {
		// 左对齐，填充空格
		parts[i] = cell + strings.Repeat(" ", colWidths[i]-len(cell))
	}

	line := strings.Join(parts, "  ")

	if isHeader && f.color {
		color.New(color.FgCyan, color.Bold).Fprintln(f.writer, line)
	} else {
		fmt.Fprintln(f.writer, line)
	}
}

// printSeparator 打印分隔线
func (f *TableFormatter) printSeparator(colWidths []int) {
	parts := make([]string, len(colWidths))
	for i, width := range colWidths {
		parts[i] = strings.Repeat("-", width)
	}
	fmt.Fprintln(f.writer, strings.Join(parts, "  "))
}

// SetHeaders 设置表头
func (f *TableFormatter) SetHeaders(headers []string) {
	f.headers = headers
}

// AddRow 添加数据行
func (f *TableFormatter) AddRow(row []string) {
	f.rows = append(f.rows, row)
}

// Render 渲染表格
func (f *TableFormatter) Render() error {
	if len(f.headers) == 0 {
		return fmt.Errorf("no headers set")
	}

	// 计算列宽
	colWidths := make([]int, len(f.headers))
	for i, header := range f.headers {
		colWidths[i] = len(header)
	}

	for _, row := range f.rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	f.printTable(f.headers, f.rows, colWidths)
	return nil
}

// PrintHeader 打印带颜色的标题
func (f *TableFormatter) PrintHeader(text string) {
	if f.color {
		color.New(color.FgCyan, color.Bold).Fprintf(f.writer, "%s\n", text)
	} else {
		fmt.Fprintf(f.writer, "%s\n", text)
	}
}

// PrintSuccess 打印成功消息
func (f *TableFormatter) PrintSuccess(text string) {
	if f.color {
		color.New(color.FgGreen).Fprintf(f.writer, "✓ %s\n", text)
	} else {
		fmt.Fprintf(f.writer, "✓ %s\n", text)
	}
}

// PrintError 打印错误消息
func (f *TableFormatter) PrintError(text string) {
	if f.color {
		color.New(color.FgRed).Fprintf(f.writer, "✗ %s\n", text)
	} else {
		fmt.Fprintf(f.writer, "✗ %s\n", text)
	}
}

// PrintWarning 打印警告消息
func (f *TableFormatter) PrintWarning(text string) {
	if f.color {
		color.New(color.FgYellow).Fprintf(f.writer, "⚠ %s\n", text)
	} else {
		fmt.Fprintf(f.writer, "⚠ %s\n", text)
	}
}

// PrintInfo 打印信息消息
func (f *TableFormatter) PrintInfo(text string) {
	if f.color {
		color.New(color.FgBlue).Fprintf(f.writer, "ℹ %s\n", text)
	} else {
		fmt.Fprintf(f.writer, "ℹ %s\n", text)
	}
}

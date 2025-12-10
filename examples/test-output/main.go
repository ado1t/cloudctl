package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ado1t/cloudctl/internal/output"
)

// 这是一个用于测试输出格式化功能的示例程序
// 使用方法:
//   go run examples/test_output.go
//   go run examples/test_output.go --output json
//   go run examples/test_output.go --output text
//   go run examples/test_output.go --no-color

var (
	outputFormat string
	noColor      bool
)

// 示例数据结构
type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

var rootCmd = &cobra.Command{
	Use:   "test-output",
	Short: "测试输出格式化功能",
	Long:  `这是一个用于测试 cloudctl 输出格式化系统的示例程序。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 创建格式化器
		formatter := output.New(output.Config{
			Format:  output.ParseFormat(outputFormat),
			Writer:  os.Stdout,
			NoColor: noColor,
		})

		// 测试表格输出 - map 切片
		cmd.Println("\n=== 测试 1: Map 切片 ===")
		data1 := []map[string]interface{}{
			{"name": "Alice", "age": 30, "city": "Beijing"},
			{"name": "Bob", "age": 25, "city": "Shanghai"},
			{"name": "Charlie", "age": 35, "city": "Guangzhou"},
		}
		formatter.Format(data1)

		// 测试单个 map
		cmd.Println("\n=== 测试 2: 单个 Map ===")
		data2 := map[string]interface{}{
			"name":    "David",
			"age":     28,
			"city":    "Shenzhen",
			"country": "China",
		}
		formatter.Format(data2)

		// 测试结构体切片
		cmd.Println("\n=== 测试 3: 结构体切片 ===")
		data3 := []Person{
			{Name: "Eve", Age: 32, City: "Hangzhou"},
			{Name: "Frank", Age: 29, City: "Nanjing"},
			{Name: "Grace", Age: 27, City: "Chengdu"},
		}
		formatter.Format(data3)

		// 测试单个结构体
		cmd.Println("\n=== 测试 4: 单个结构体 ===")
		data4 := Person{Name: "Henry", Age: 31, City: "Wuhan"}
		formatter.Format(data4)

		// 测试 TableFormatter 的特殊方法
		if tableFormatter, ok := formatter.(*output.TableFormatter); ok {
			cmd.Println("\n=== 测试 5: 表格格式化器特殊方法 ===")
			tableFormatter.PrintHeader("这是一个标题")
			tableFormatter.PrintSuccess("操作成功")
			tableFormatter.PrintError("发生错误")
			tableFormatter.PrintWarning("警告信息")
			tableFormatter.PrintInfo("提示信息")

			cmd.Println("\n=== 测试 6: 手动构建表格 ===")
			tableFormatter.SetHeaders([]string{"ID", "Name", "Status"})
			tableFormatter.AddRow([]string{"1", "Task A", "Completed"})
			tableFormatter.AddRow([]string{"2", "Task B", "In Progress"})
			tableFormatter.AddRow([]string{"3", "Task C", "Pending"})
			tableFormatter.Render()
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "输出格式 (table|json|text)")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "禁用颜色输出")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

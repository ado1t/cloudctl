package cmd

import (
	"testing"

	"github.com/ado1t/cloudctl/internal/config"
	"github.com/ado1t/cloudctl/internal/output"
)

func TestInitializeFormatter(t *testing.T) {
	tests := []struct {
		name         string
		outputFormat string
		noColor      bool
		quiet        bool
		cfg          *config.Config
		wantFormat   output.Format
	}{
		{
			name:         "默认表格格式",
			outputFormat: "table",
			noColor:      false,
			quiet:        false,
			cfg: &config.Config{
				Output: config.OutputConfig{
					Format: "table",
					Color:  true,
				},
			},
			wantFormat: output.FormatTable,
		},
		{
			name:         "JSON格式",
			outputFormat: "json",
			noColor:      false,
			quiet:        false,
			cfg: &config.Config{
				Output: config.OutputConfig{
					Format: "table",
					Color:  true,
				},
			},
			wantFormat: output.FormatJSON,
		},
		{
			name:         "文本格式",
			outputFormat: "text",
			noColor:      false,
			quiet:        false,
			cfg: &config.Config{
				Output: config.OutputConfig{
					Format: "table",
					Color:  true,
				},
			},
			wantFormat: output.FormatText,
		},
		{
			name:         "静默模式",
			outputFormat: "table",
			noColor:      false,
			quiet:        true,
			cfg: &config.Config{
				Output: config.OutputConfig{
					Format: "table",
					Color:  true,
				},
			},
			wantFormat: output.FormatTable,
		},
		{
			name:         "禁用颜色",
			outputFormat: "table",
			noColor:      true,
			quiet:        false,
			cfg: &config.Config{
				Output: config.OutputConfig{
					Format: "table",
					Color:  true,
				},
			},
			wantFormat: output.FormatTable,
		},
		{
			name:         "使用配置文件格式",
			outputFormat: "",
			noColor:      false,
			quiet:        false,
			cfg: &config.Config{
				Output: config.OutputConfig{
					Format: "json",
					Color:  true,
				},
			},
			wantFormat: output.FormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initializeFormatter(tt.outputFormat, tt.noColor, tt.quiet, tt.cfg)

			if globalFormatter == nil {
				t.Error("globalFormatter 不应为 nil")
				return
			}

			// 验证静默模式
			if tt.quiet {
				if _, ok := globalFormatter.(*output.NullFormatter); !ok {
					t.Error("静默模式应返回 NullFormatter")
				}
			}
		})
	}
}

func TestGetFormatter(t *testing.T) {
	// 初始化一个格式化器
	cfg := &config.Config{
		Output: config.OutputConfig{
			Format: "table",
			Color:  true,
		},
	}
	initializeFormatter("table", false, false, cfg)

	formatter := GetFormatter()
	if formatter == nil {
		t.Error("GetFormatter() 不应返回 nil")
	}

	if formatter != globalFormatter {
		t.Error("GetFormatter() 应返回 globalFormatter")
	}
}

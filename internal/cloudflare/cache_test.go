package cloudflare

import (
	"testing"
)

// TestCachePurgeParams 测试缓存清除参数结构
func TestCachePurgeParams(t *testing.T) {
	params := CachePurgeParams{
		PurgeEverything: true,
	}

	if !params.PurgeEverything {
		t.Error("PurgeEverything should be true")
	}
}

// TestCachePurgeParamsFiles 测试文件清除参数
func TestCachePurgeParamsFiles(t *testing.T) {
	params := CachePurgeParams{
		Files: []string{"/path/to/file1.html", "/path/to/file2.css"},
	}

	if len(params.Files) != 2 {
		t.Errorf("Files count = %d, want %d", len(params.Files), 2)
	}

	if params.Files[0] != "/path/to/file1.html" {
		t.Errorf("Files[0] = %s, want %s", params.Files[0], "/path/to/file1.html")
	}
}

// TestCachePurgeParamsPrefixes 测试前缀清除参数
func TestCachePurgeParamsPrefixes(t *testing.T) {
	params := CachePurgeParams{
		Prefixes: []string{"/foo/bar/", "/images/"},
	}

	if len(params.Prefixes) != 2 {
		t.Errorf("Prefixes count = %d, want %d", len(params.Prefixes), 2)
	}

	if params.Prefixes[0] != "/foo/bar/" {
		t.Errorf("Prefixes[0] = %s, want %s", params.Prefixes[0], "/foo/bar/")
	}
}

// TestCachePurgeResult 测试缓存清除结果结构
func TestCachePurgeResult(t *testing.T) {
	result := CachePurgeResult{
		Success:   true,
		ID:        "test-purge-id",
		Message:   "已清除所有缓存",
		PurgeType: "everything",
		ItemCount: 0,
	}

	if !result.Success {
		t.Error("Success should be true")
	}

	if result.ID != "test-purge-id" {
		t.Errorf("ID = %s, want %s", result.ID, "test-purge-id")
	}

	if result.PurgeType != "everything" {
		t.Errorf("PurgeType = %s, want %s", result.PurgeType, "everything")
	}
}

// TestValidatePurgeParams 测试参数验证
func TestValidatePurgeParams(t *testing.T) {
	tests := []struct {
		name    string
		params  CachePurgeParams
		wantErr bool
	}{
		{
			name: "有效 - purge all",
			params: CachePurgeParams{
				PurgeEverything: true,
			},
			wantErr: false,
		},
		{
			name: "有效 - files",
			params: CachePurgeParams{
				Files: []string{"/file1.html", "/file2.css"},
			},
			wantErr: false,
		},
		{
			name: "有效 - prefixes",
			params: CachePurgeParams{
				Prefixes: []string{"/foo/bar/", "/images/"},
			},
			wantErr: false,
		},
		{
			name: "有效 - tags",
			params: CachePurgeParams{
				Tags: []string{"tag1", "tag2"},
			},
			wantErr: false,
		},
		{
			name: "有效 - hosts",
			params: CachePurgeParams{
				Hosts: []string{"www.example.com", "api.example.com"},
			},
			wantErr: false,
		},
		{
			name:    "无效 - 未指定任何模式",
			params:  CachePurgeParams{},
			wantErr: true,
		},
		{
			name: "无效 - 同时指定多种模式",
			params: CachePurgeParams{
				PurgeEverything: true,
				Files:           []string{"/file.html"},
			},
			wantErr: true,
		},
		{
			name: "无效 - 文件数量超限",
			params: CachePurgeParams{
				Files: make([]string, 31),
			},
			wantErr: true,
		},
		{
			name: "无效 - 前缀数量超限",
			params: CachePurgeParams{
				Prefixes: make([]string, 31),
			},
			wantErr: true,
		},
		{
			name: "无效 - 标签数量超限",
			params: CachePurgeParams{
				Tags: make([]string, 31),
			},
			wantErr: true,
		},
		{
			name: "无效 - 主机名数量超限",
			params: CachePurgeParams{
				Hosts: make([]string, 31),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePurgeParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePurgeParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package cloudflare

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadDNSBatchConfig 测试加载批量配置
func TestLoadDNSBatchConfig(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "dns-records.yaml")

	configContent := `zones:
  - zone: example1.com
    records:
      - type: A
        name: www
        content: 1.2.3.4
        ttl: 3600
        proxied: true
      - type: CNAME
        name: blog
        content: example1.com
        proxied: true
  
  - zone: example2.com
    records:
      - type: A
        name: www
        content: 2.3.4.5
        ttl: 3600
        proxied: true
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}

	// 加载配置
	config, err := LoadDNSBatchConfig(configFile)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证配置
	if len(config.Zones) != 2 {
		t.Errorf("Zones 数量 = %d, want %d", len(config.Zones), 2)
	}

	// 验证第一个 zone
	if config.Zones[0].Zone != "example1.com" {
		t.Errorf("Zone[0].Zone = %s, want %s", config.Zones[0].Zone, "example1.com")
	}

	if len(config.Zones[0].Records) != 2 {
		t.Errorf("Zone[0] 记录数 = %d, want %d", len(config.Zones[0].Records), 2)
	}

	// 验证第一条记录
	record := config.Zones[0].Records[0]
	if record.Type != "A" {
		t.Errorf("Record.Type = %s, want %s", record.Type, "A")
	}
	if record.Name != "www" {
		t.Errorf("Record.Name = %s, want %s", record.Name, "www")
	}
	if record.Content != "1.2.3.4" {
		t.Errorf("Record.Content = %s, want %s", record.Content, "1.2.3.4")
	}
	if record.TTL != 3600 {
		t.Errorf("Record.TTL = %f, want %f", record.TTL, 3600.0)
	}
	if !record.Proxied {
		t.Error("Record.Proxied should be true")
	}
}

// TestLoadDNSBatchConfigFileNotFound 测试文件不存在
func TestLoadDNSBatchConfigFileNotFound(t *testing.T) {
	_, err := LoadDNSBatchConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Error("应该返回错误")
	}
}

// TestLoadDNSBatchConfigInvalidYAML 测试无效的 YAML
func TestLoadDNSBatchConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidContent := `zones:
  - zone: example.com
    records:
      - type: A
        name: www
        content: 1.2.3.4
        invalid_field: [unclosed
`

	if err := os.WriteFile(configFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}

	_, err := LoadDNSBatchConfig(configFile)
	if err == nil {
		t.Error("应该返回解析错误")
	}
}

// TestDNSBatchConfigValidate 测试配置验证
func TestDNSBatchConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  DNSBatchConfig
		wantErr bool
	}{
		{
			name: "有效配置",
			config: DNSBatchConfig{
				Zones: []DNSZoneConfig{
					{
						Zone: "example.com",
						Records: []DNSRecordConfig{
							{
								Type:    "A",
								Name:    "www",
								Content: "1.2.3.4",
								TTL:     3600,
								Proxied: true,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "空 zones",
			config: DNSBatchConfig{
				Zones: []DNSZoneConfig{},
			},
			wantErr: true,
		},
		{
			name: "zone 名称为空",
			config: DNSBatchConfig{
				Zones: []DNSZoneConfig{
					{
						Zone: "",
						Records: []DNSRecordConfig{
							{
								Type:    "A",
								Name:    "www",
								Content: "1.2.3.4",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "空 records",
			config: DNSBatchConfig{
				Zones: []DNSZoneConfig{
					{
						Zone:    "example.com",
						Records: []DNSRecordConfig{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "记录类型为空",
			config: DNSBatchConfig{
				Zones: []DNSZoneConfig{
					{
						Zone: "example.com",
						Records: []DNSRecordConfig{
							{
								Type:    "",
								Name:    "www",
								Content: "1.2.3.4",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "记录名称为空",
			config: DNSBatchConfig{
				Zones: []DNSZoneConfig{
					{
						Zone: "example.com",
						Records: []DNSRecordConfig{
							{
								Type:    "A",
								Name:    "",
								Content: "1.2.3.4",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "记录内容为空",
			config: DNSBatchConfig{
				Zones: []DNSZoneConfig{
					{
						Zone: "example.com",
						Records: []DNSRecordConfig{
							{
								Type:    "A",
								Name:    "www",
								Content: "",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "不支持的记录类型",
			config: DNSBatchConfig{
				Zones: []DNSZoneConfig{
					{
						Zone: "example.com",
						Records: []DNSRecordConfig{
							{
								Type:    "MX",
								Name:    "www",
								Content: "mail.example.com",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDNSBatchResult 测试批量结果结构
func TestDNSBatchResult(t *testing.T) {
	result := DNSBatchResult{
		TotalZones:     3,
		TotalRecords:   10,
		SuccessZones:   2,
		SuccessRecords: 8,
		FailedZones:    1,
		FailedRecords:  2,
	}

	if result.TotalZones != 3 {
		t.Errorf("TotalZones = %d, want %d", result.TotalZones, 3)
	}

	if result.SuccessRecords != 8 {
		t.Errorf("SuccessRecords = %d, want %d", result.SuccessRecords, 8)
	}

	if result.FailedRecords != 2 {
		t.Errorf("FailedRecords = %d, want %d", result.FailedRecords, 2)
	}
}

// TestDNSZoneResult 测试 Zone 结果结构
func TestDNSZoneResult(t *testing.T) {
	result := DNSZoneResult{
		Zone:           "example.com",
		Success:        true,
		TotalRecords:   5,
		SuccessRecords: 5,
		FailedRecords:  0,
		RecordResults:  []DNSRecordResult{},
	}

	if result.Zone != "example.com" {
		t.Errorf("Zone = %s, want %s", result.Zone, "example.com")
	}

	if !result.Success {
		t.Error("Success should be true")
	}

	if result.FailedRecords != 0 {
		t.Errorf("FailedRecords = %d, want %d", result.FailedRecords, 0)
	}
}

// TestDNSRecordResult 测试记录结果结构
func TestDNSRecordResult(t *testing.T) {
	result := DNSRecordResult{
		Type:     "A",
		Name:     "www",
		Content:  "1.2.3.4",
		Success:  true,
		RecordID: "test-record-id",
	}

	if result.Type != "A" {
		t.Errorf("Type = %s, want %s", result.Type, "A")
	}

	if result.Name != "www" {
		t.Errorf("Name = %s, want %s", result.Name, "www")
	}

	if !result.Success {
		t.Error("Success should be true")
	}

	if result.RecordID != "test-record-id" {
		t.Errorf("RecordID = %s, want %s", result.RecordID, "test-record-id")
	}
}

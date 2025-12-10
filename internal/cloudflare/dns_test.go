package cloudflare

import (
	"testing"
	"time"
)

// TestDNSRecordInfo 测试 DNSRecordInfo 结构
func TestDNSRecordInfo(t *testing.T) {
	now := time.Now()

	record := DNSRecordInfo{
		ID:         "test-id",
		Type:       "A",
		Name:       "www.example.com",
		Content:    "1.2.3.4",
		TTL:        3600,
		Proxied:    true,
		Proxiable:  true,
		CreatedOn:  now,
		ModifiedOn: now,
	}

	// 验证字段
	if record.ID != "test-id" {
		t.Errorf("ID = %v, want %v", record.ID, "test-id")
	}

	if record.Type != "A" {
		t.Errorf("Type = %v, want %v", record.Type, "A")
	}

	if record.Name != "www.example.com" {
		t.Errorf("Name = %v, want %v", record.Name, "www.example.com")
	}

	if record.Content != "1.2.3.4" {
		t.Errorf("Content = %v, want %v", record.Content, "1.2.3.4")
	}

	if record.TTL != 3600 {
		t.Errorf("TTL = %v, want %v", record.TTL, 3600)
	}

	if !record.Proxied {
		t.Error("Proxied should be true")
	}

	if !record.Proxiable {
		t.Error("Proxiable should be true")
	}
}

// TestDNSRecordCreateParams 测试 DNSRecordCreateParams 结构
func TestDNSRecordCreateParams(t *testing.T) {
	params := DNSRecordCreateParams{
		Type:    "A",
		Name:    "www",
		Content: "1.2.3.4",
		TTL:     3600,
		Proxied: true,
	}

	if params.Type != "A" {
		t.Errorf("Type = %v, want %v", params.Type, "A")
	}

	if params.Name != "www" {
		t.Errorf("Name = %v, want %v", params.Name, "www")
	}

	if params.Content != "1.2.3.4" {
		t.Errorf("Content = %v, want %v", params.Content, "1.2.3.4")
	}

	if params.TTL != 3600 {
		t.Errorf("TTL = %v, want %v", params.TTL, 3600)
	}

	if !params.Proxied {
		t.Error("Proxied should be true")
	}
}

// TestDNSRecordUpdateParams 测试 DNSRecordUpdateParams 结构
func TestDNSRecordUpdateParams(t *testing.T) {
	content := "2.3.4.5"
	ttl := float64(7200)
	proxied := false

	params := DNSRecordUpdateParams{
		Content: &content,
		TTL:     &ttl,
		Proxied: &proxied,
	}

	if params.Content == nil {
		t.Error("Content should not be nil")
	} else if *params.Content != "2.3.4.5" {
		t.Errorf("Content = %v, want %v", *params.Content, "2.3.4.5")
	}

	if params.TTL == nil {
		t.Error("TTL should not be nil")
	} else if *params.TTL != 7200 {
		t.Errorf("TTL = %v, want %v", *params.TTL, 7200)
	}

	if params.Proxied == nil {
		t.Error("Proxied should not be nil")
	} else if *params.Proxied != false {
		t.Errorf("Proxied = %v, want %v", *params.Proxied, false)
	}
}

// TestDNSRecordUpdateParamsPartial 测试部分更新参数
func TestDNSRecordUpdateParamsPartial(t *testing.T) {
	content := "2.3.4.5"

	params := DNSRecordUpdateParams{
		Content: &content,
		TTL:     nil,
		Proxied: nil,
	}

	if params.Content == nil {
		t.Error("Content should not be nil")
	} else if *params.Content != "2.3.4.5" {
		t.Errorf("Content = %v, want %v", *params.Content, "2.3.4.5")
	}

	if params.TTL != nil {
		t.Error("TTL should be nil")
	}

	if params.Proxied != nil {
		t.Error("Proxied should be nil")
	}
}

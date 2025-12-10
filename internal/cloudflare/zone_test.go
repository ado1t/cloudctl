package cloudflare

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/ado1t/cloudctl/internal/config"
)

func TestListZones(t *testing.T) {
	// 跳过需要真实 API Token 的测试
	if os.Getenv("CLOUDFLARE_API_TOKEN") == "" {
		t.Skip("需要设置 CLOUDFLARE_API_TOKEN 环境变量")
	}

	// 加载配置
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建客户端
	client, err := NewClient("", slog.Default())
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 列出 Zone
	ctx := context.Background()
	zones, err := client.ListZones(ctx)
	if err != nil {
		t.Fatalf("列出 Zone 失败: %v", err)
	}

	// 验证结果
	t.Logf("找到 %d 个 Zone", len(zones))
	for _, zone := range zones {
		if zone.ID == "" {
			t.Error("Zone ID 为空")
		}
		if zone.Name == "" {
			t.Error("Zone Name 为空")
		}
		t.Logf("Zone: %s (ID: %s, Status: %s)", zone.Name, zone.ID, zone.Status)
	}
}

func TestGetZone(t *testing.T) {
	// 跳过需要真实 API Token 的测试
	if os.Getenv("CLOUDFLARE_API_TOKEN") == "" {
		t.Skip("需要设置 CLOUDFLARE_API_TOKEN 环境变量")
	}

	// 加载配置
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建客户端
	client, err := NewClient("", slog.Default())
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 先列出所有 Zone
	zones, err := client.ListZones(ctx)
	if err != nil {
		t.Fatalf("列出 Zone 失败: %v", err)
	}

	if len(zones) == 0 {
		t.Skip("没有可用的 Zone 进行测试")
	}

	// 获取第一个 Zone 的详细信息
	zoneID := zones[0].ID
	zone, err := client.GetZone(ctx, zoneID)
	if err != nil {
		t.Fatalf("获取 Zone 失败: %v", err)
	}

	// 验证结果
	if zone.ID != zoneID {
		t.Errorf("Zone ID 不匹配: got %s, want %s", zone.ID, zoneID)
	}
	if zone.Name == "" {
		t.Error("Zone Name 为空")
	}

	t.Logf("Zone: %s (ID: %s, Status: %s)", zone.Name, zone.ID, zone.Status)
}

func TestGetZoneByName(t *testing.T) {
	// 跳过需要真实 API Token 的测试
	if os.Getenv("CLOUDFLARE_API_TOKEN") == "" {
		t.Skip("需要设置 CLOUDFLARE_API_TOKEN 环境变量")
	}

	// 加载配置
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建客户端
	client, err := NewClient("", slog.Default())
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 先列出所有 Zone
	zones, err := client.ListZones(ctx)
	if err != nil {
		t.Fatalf("列出 Zone 失败: %v", err)
	}

	if len(zones) == 0 {
		t.Skip("没有可用的 Zone 进行测试")
	}

	// 通过名称获取第一个 Zone
	zoneName := zones[0].Name
	zone, err := client.GetZoneByName(ctx, zoneName)
	if err != nil {
		t.Fatalf("通过名称获取 Zone 失败: %v", err)
	}

	// 验证结果
	if zone.Name != zoneName {
		t.Errorf("Zone Name 不匹配: got %s, want %s", zone.Name, zoneName)
	}

	t.Logf("Zone: %s (ID: %s, Status: %s)", zone.Name, zone.ID, zone.Status)
}

func TestGetZoneByName_NotFound(t *testing.T) {
	// 跳过需要真实 API Token 的测试
	if os.Getenv("CLOUDFLARE_API_TOKEN") == "" {
		t.Skip("需要设置 CLOUDFLARE_API_TOKEN 环境变量")
	}

	// 加载配置
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建客户端
	client, err := NewClient("", slog.Default())
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 查找不存在的 Zone
	_, err = client.GetZoneByName(ctx, "non-existent-domain-12345.com")
	if err == nil {
		t.Error("期望返回错误，但成功了")
	}

	// 验证是 NotFound 错误
	if !IsNotFoundError(err) {
		t.Errorf("期望 NotFound 错误，得到: %v", err)
	}
}

func TestZoneInfo_Structure(t *testing.T) {
	// 测试 ZoneInfo 结构体
	zone := ZoneInfo{
		ID:          "test-id",
		Name:        "example.com",
		Status:      "active",
		NameServers: []string{"ns1.cloudflare.com", "ns2.cloudflare.com"},
	}

	if zone.ID != "test-id" {
		t.Errorf("ID = %s, want test-id", zone.ID)
	}
	if zone.Name != "example.com" {
		t.Errorf("Name = %s, want example.com", zone.Name)
	}
	if zone.Status != "active" {
		t.Errorf("Status = %s, want active", zone.Status)
	}
	if len(zone.NameServers) != 2 {
		t.Errorf("NameServers length = %d, want 2", len(zone.NameServers))
	}
}

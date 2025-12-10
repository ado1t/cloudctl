package cloudflare

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/ado1t/cloudctl/internal/config"
)

func TestNewClient(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("CLOUDFLARE_API_TOKEN", "test-token-123")
	defer os.Unsetenv("CLOUDFLARE_API_TOKEN")

	// 加载配置
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	tests := []struct {
		name        string
		profileName string
		wantErr     bool
	}{
		{
			name:        "使用默认 profile",
			profileName: "",
			wantErr:     false,
		},
		{
			name:        "使用指定 profile",
			profileName: "default",
			wantErr:     false,
		},
		{
			name:        "使用不存在的 profile",
			profileName: "non-existent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.profileName, slog.Default())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() 返回 nil 客户端")
			}
			if client != nil {
				if err := client.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			}
		})
	}
}

func TestNewClient_NoToken(t *testing.T) {
	// 保存原始环境变量
	originalToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("CLOUDFLARE_API_TOKEN", originalToken)
		}
	}()

	// 确保没有设置 token
	os.Unsetenv("CLOUDFLARE_API_TOKEN")

	// 重新加载配置（清除之前的配置）
	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 如果配置中已经有 token（从配置文件加载），跳过此测试
	if len(cfg.Cloudflare) > 0 {
		for _, profile := range cfg.Cloudflare {
			if profile.APIToken != "" {
				t.Skip("配置文件中已有 API Token，跳过此测试")
				return
			}
		}
	}

	// 尝试创建客户端应该失败
	client, err := NewClient("", slog.Default())
	if err == nil {
		t.Error("期望创建客户端失败，但成功了")
		if client != nil {
			client.Close()
		}
	}
}

func TestClient_WithRetry(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("CLOUDFLARE_API_TOKEN", "test-token-123")
	defer os.Unsetenv("CLOUDFLARE_API_TOKEN")

	// 加载配置
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	client, err := NewClient("", slog.Default())
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("成功的操作", func(t *testing.T) {
		err := client.WithRetry(ctx, "测试操作", func() error {
			return nil
		})
		if err != nil {
			t.Errorf("WithRetry() error = %v, 期望 nil", err)
		}
	})

	t.Run("失败的操作", func(t *testing.T) {
		testErr := NewValidationError("测试", "测试错误")
		err := client.WithRetry(ctx, "测试操作", func() error {
			return testErr
		})
		if err == nil {
			t.Error("WithRetry() 期望返回错误，但返回 nil")
		}
	})
}

func TestClient_Profile(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("CLOUDFLARE_API_TOKEN", "test-token-123")
	defer os.Unsetenv("CLOUDFLARE_API_TOKEN")

	// 加载配置
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	client, err := NewClient("default", slog.Default())
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	if client.Profile() != "default" {
		t.Errorf("Profile() = %v, 期望 default", client.Profile())
	}
}

func TestClient_API(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("CLOUDFLARE_API_TOKEN", "test-token-123")
	defer os.Unsetenv("CLOUDFLARE_API_TOKEN")

	// 加载配置
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	client, err := NewClient("", slog.Default())
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	if client.API() == nil {
		t.Error("API() 返回 nil")
	}
}

func TestClient_Logger(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("CLOUDFLARE_API_TOKEN", "test-token-123")
	defer os.Unsetenv("CLOUDFLARE_API_TOKEN")

	// 加载配置
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	client, err := NewClient("", slog.Default())
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	if client.Logger() == nil {
		t.Error("Logger() 返回 nil")
	}
}

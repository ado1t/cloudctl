package aws

import (
	"testing"

	"github.com/ado1t/cloudctl/internal/config"
)

func TestNewClient(t *testing.T) {
	// 初始化配置
	cfg := &config.Config{
		DefaultProfile: config.DefaultProfile{
			AWS: "test-profile",
		},
		AWS: map[string]config.AWSProfile{
			"test-profile": {
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				Region:          "us-east-1",
			},
		},
	}
	config.SetConfigForTest(cfg)

	tests := []struct {
		name        string
		profile     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "使用指定的 profile",
			profile: "test-profile",
			wantErr: false,
		},
		{
			name:    "使用默认 profile",
			profile: "",
			wantErr: false,
		},
		{
			name:        "profile 不存在",
			profile:     "non-existent",
			wantErr:     true,
			errContains: "不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.profile)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() 应该返回错误")
				}
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("错误信息 %q 不包含 %q", err.Error(), tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() 错误 = %v", err)
				return
			}

			if client == nil {
				t.Error("NewClient() 返回的客户端为 nil")
				return
			}

			if client.CloudFront() == nil {
				t.Error("CloudFront 客户端为 nil")
			}

			expectedProfile := tt.profile
			if expectedProfile == "" {
				expectedProfile = "test-profile"
			}
			if client.Profile() != expectedProfile {
				t.Errorf("Profile() = %v, 期望 %v", client.Profile(), expectedProfile)
			}
		})
	}
}

func TestNewClient_NoConfig(t *testing.T) {
	// 清空配置
	config.SetConfigForTest(nil)

	_, err := NewClient("test")
	if err == nil {
		t.Error("没有配置时应该返回错误")
	}
}

func TestNewClient_MissingCredentials(t *testing.T) {
	tests := []struct {
		name        string
		profile     config.AWSProfile
		errContains string
	}{
		{
			name: "缺少 access_key_id",
			profile: config.AWSProfile{
				AccessKeyID:     "",
				SecretAccessKey: "secret",
				Region:          "us-east-1",
			},
			errContains: "access_key_id",
		},
		{
			name: "缺少 secret_access_key",
			profile: config.AWSProfile{
				AccessKeyID:     "key",
				SecretAccessKey: "",
				Region:          "us-east-1",
			},
			errContains: "secret_access_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				AWS: map[string]config.AWSProfile{
					"test": tt.profile,
				},
			}
			config.SetConfigForTest(cfg)

			_, err := NewClient("test")
			if err == nil {
				t.Error("应该返回错误")
				return
			}

			if !contains(err.Error(), tt.errContains) {
				t.Errorf("错误信息 %q 不包含 %q", err.Error(), tt.errContains)
			}
		})
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

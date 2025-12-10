package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	cfg := getDefaultConfig()

	if cfg == nil {
		t.Fatal("默认配置不应为 nil")
	}

	// 验证默认值
	if cfg.Log.Level != "error" {
		t.Errorf("期望日志级别为 'error'，实际为 '%s'", cfg.Log.Level)
	}

	if cfg.Output.Format != "table" {
		t.Errorf("期望输出格式为 'table'，实际为 '%s'", cfg.Output.Format)
	}

	if cfg.API.Timeout != 30 {
		t.Errorf("期望 API 超时时间为 30，实际为 %d", cfg.API.Timeout)
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)

	if cfg.Log.Level != "error" {
		t.Errorf("期望日志级别为 'error'，实际为 '%s'", cfg.Log.Level)
	}

	if cfg.DefaultProfile.Cloudflare != "default" {
		t.Errorf("期望默认 Cloudflare profile 为 'default'，实际为 '%s'", cfg.DefaultProfile.Cloudflare)
	}

	if cfg.DefaultProfile.AWS != "default" {
		t.Errorf("期望默认 AWS profile 为 'default'，实际为 '%s'", cfg.DefaultProfile.AWS)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "nil 配置",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "有效配置",
			cfg: &Config{
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
				Output: OutputConfig{
					Format: "table",
				},
				API: APIConfig{
					Timeout: 30,
					Retry: RetryConfig{
						Enabled:      true,
						MaxAttempts:  3,
						InitialDelay: 1,
						MaxDelay:     30,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "无效日志级别",
			cfg: &Config{
				Log: LogConfig{
					Level:  "invalid",
					Format: "text",
				},
				Output: OutputConfig{
					Format: "table",
				},
				API: APIConfig{
					Timeout: 30,
				},
			},
			wantErr: true,
		},
		{
			name: "无效输出格式",
			cfg: &Config{
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
				Output: OutputConfig{
					Format: "invalid",
				},
				API: APIConfig{
					Timeout: 30,
				},
			},
			wantErr: true,
		},
		{
			name: "无效 API 超时",
			cfg: &Config{
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
				Output: OutputConfig{
					Format: "table",
				},
				API: APIConfig{
					Timeout: 0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "空字符串",
			input: "",
			want:  "",
		},
		{
			name:  "短字符串",
			input: "short",
			want:  "****",
		},
		{
			name:  "正常长度",
			input: "1234567890abcdef",
			want:  "1234****cdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskSensitiveData(tt.input)
			if got != tt.want {
				t.Errorf("MaskSensitiveData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("CLOUDFLARE_API_TOKEN", "test-token")
	os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	defer func() {
		os.Unsetenv("CLOUDFLARE_API_TOKEN")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	}()

	// 加载配置（不存在配置文件，应该使用环境变量）
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证环境变量是否被正确读取
	if cfg.Cloudflare["default"].APIToken != "test-token" {
		t.Errorf("期望 API Token 为 'test-token'，实际为 '%s'", cfg.Cloudflare["default"].APIToken)
	}

	if cfg.AWS["default"].AccessKeyID != "test-key" {
		t.Errorf("期望 Access Key ID 为 'test-key'，实际为 '%s'", cfg.AWS["default"].AccessKeyID)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
default_profile:
  cloudflare: test-cf
  aws: test-aws

cloudflare:
  test-cf:
    api_token: test-token

aws:
  test-aws:
    access_key_id: test-key
    secret_access_key: test-secret
    region: us-east-1

log:
  level: debug
  format: json

output:
  format: json
  color: false
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建配置文件失败: %v", err)
	}

	// 加载配置
	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证配置
	if cfg.DefaultProfile.Cloudflare != "test-cf" {
		t.Errorf("期望默认 Cloudflare profile 为 'test-cf'，实际为 '%s'", cfg.DefaultProfile.Cloudflare)
	}

	if cfg.Log.Level != "debug" {
		t.Errorf("期望日志级别为 'debug'，实际为 '%s'", cfg.Log.Level)
	}

	if cfg.Output.Format != "json" {
		t.Errorf("期望输出格式为 'json'，实际为 '%s'", cfg.Output.Format)
	}

	if cfg.Output.Color != false {
		t.Errorf("期望颜色输出为 false，实际为 %v", cfg.Output.Color)
	}
}

func TestGetCloudflareProfile(t *testing.T) {
	// 重置全局配置
	globalConfig = &Config{
		DefaultProfile: DefaultProfile{
			Cloudflare: "default",
		},
		Cloudflare: map[string]CFProfile{
			"default": {APIToken: "test-token"},
			"prod":    {APIToken: "prod-token"},
		},
	}
	configLoaded = true

	tests := []struct {
		name        string
		profileName string
		wantToken   string
		wantErr     bool
	}{
		{
			name:        "默认 profile",
			profileName: "",
			wantToken:   "test-token",
			wantErr:     false,
		},
		{
			name:        "指定 profile",
			profileName: "prod",
			wantToken:   "prod-token",
			wantErr:     false,
		},
		{
			name:        "不存在的 profile",
			profileName: "nonexistent",
			wantToken:   "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := GetCloudflareProfile(tt.profileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCloudflareProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && profile.APIToken != tt.wantToken {
				t.Errorf("GetCloudflareProfile() token = %v, want %v", profile.APIToken, tt.wantToken)
			}
		})
	}
}

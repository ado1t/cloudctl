package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var (
	// globalConfig 全局配置实例
	globalConfig *Config
	// configLoaded 配置是否已加载
	configLoaded bool
)

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	if configPath != "" {
		// 使用指定的配置文件
		v.SetConfigFile(configPath)
	} else {
		// 使用默认配置文件路径
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("无法获取用户主目录: %w", err)
		}

		configDir := filepath.Join(home, ".cloudctl")
		v.AddConfigPath(configDir)
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// 设置环境变量前缀
	v.SetEnvPrefix("CLOUDCTL")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定特定的环境变量
	bindEnvVars(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，返回默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return getDefaultConfig(), nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 展开环境变量
	expandEnvVars(&cfg)

	// 应用默认值
	applyDefaults(&cfg)

	// 保存全局配置
	globalConfig = &cfg
	configLoaded = true

	return &cfg, nil
}

// expandEnvVars 展开配置中的环境变量
func expandEnvVars(cfg *Config) {
	// 展开 Cloudflare profiles 中的环境变量
	for name, profile := range cfg.Cloudflare {
		profile.APIToken = os.ExpandEnv(profile.APIToken)
		cfg.Cloudflare[name] = profile
	}

	// 展开 AWS profiles 中的环境变量
	for name, profile := range cfg.AWS {
		profile.AccessKeyID = os.ExpandEnv(profile.AccessKeyID)
		profile.SecretAccessKey = os.ExpandEnv(profile.SecretAccessKey)
		profile.Region = os.ExpandEnv(profile.Region)
		cfg.AWS[name] = profile
	}
}

// bindEnvVars 绑定环境变量
func bindEnvVars(v *viper.Viper) {
	// Cloudflare 环境变量
	v.BindEnv("cloudflare.default.api_token", "CLOUDFLARE_API_TOKEN")

	// AWS 环境变量
	v.BindEnv("aws.default.access_key_id", "AWS_ACCESS_KEY_ID")
	v.BindEnv("aws.default.secret_access_key", "AWS_SECRET_ACCESS_KEY")
	v.BindEnv("aws.default.region", "AWS_REGION")

	// 通用环境变量
	v.BindEnv("log.level", "CLOUDCTL_LOG_LEVEL")
	v.BindEnv("output.format", "CLOUDCTL_OUTPUT_FORMAT")
	v.BindEnv("output.color", "CLOUDCTL_NO_COLOR")
}

// getDefaultConfig 返回默认配置
func getDefaultConfig() *Config {
	cfg := &Config{
		DefaultProfile: DefaultProfile{
			Cloudflare: "default",
			AWS:        "default",
		},
		Cloudflare: make(map[string]CFProfile),
		AWS:        make(map[string]AWSProfile),
		Log: LogConfig{
			Level:     "error",
			Format:    "text",
			Output:    "stdout",
			AddSource: false,
		},
		Output: OutputConfig{
			Format: "table",
			Color:  true,
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
	}

	// 从环境变量读取默认 profile 的凭证
	if token := os.Getenv("CLOUDFLARE_API_TOKEN"); token != "" {
		cfg.Cloudflare["default"] = CFProfile{APIToken: token}
	}

	if accessKey := os.Getenv("AWS_ACCESS_KEY_ID"); accessKey != "" {
		cfg.AWS["default"] = AWSProfile{
			AccessKeyID:     accessKey,
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			Region:          getEnvOrDefault("AWS_REGION", "us-east-1"),
		}
	}

	return cfg
}

// applyDefaults 应用默认值
func applyDefaults(cfg *Config) {
	// 日志配置默认值
	if cfg.Log.Level == "" {
		cfg.Log.Level = "error"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "text"
	}
	if cfg.Log.Output == "" {
		cfg.Log.Output = "stdout"
	}

	// 输出配置默认值
	if cfg.Output.Format == "" {
		cfg.Output.Format = "table"
	}

	// API 配置默认值
	if cfg.API.Timeout == 0 {
		cfg.API.Timeout = 30
	}
	if cfg.API.Retry.MaxAttempts == 0 {
		cfg.API.Retry.MaxAttempts = 3
	}
	if cfg.API.Retry.InitialDelay == 0 {
		cfg.API.Retry.InitialDelay = 1
	}
	if cfg.API.Retry.MaxDelay == 0 {
		cfg.API.Retry.MaxDelay = 30
	}

	// 默认 profile
	if cfg.DefaultProfile.Cloudflare == "" {
		cfg.DefaultProfile.Cloudflare = "default"
	}
	if cfg.DefaultProfile.AWS == "" {
		cfg.DefaultProfile.AWS = "default"
	}
}

// Get 获取全局配置
func Get() *Config {
	if !configLoaded {
		// 如果配置未加载，尝试加载默认配置
		cfg, _ := Load("")
		return cfg
	}
	return globalConfig
}

// GetCloudflareProfile 获取 Cloudflare profile
func GetCloudflareProfile(profileName string) (*CFProfile, error) {
	cfg := Get()
	if cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}

	// 如果未指定 profile，使用默认 profile
	if profileName == "" {
		profileName = cfg.DefaultProfile.Cloudflare
	}

	profile, ok := cfg.Cloudflare[profileName]
	if !ok {
		return nil, fmt.Errorf("Cloudflare profile '%s' 不存在", profileName)
	}

	if profile.APIToken == "" {
		return nil, fmt.Errorf("Cloudflare profile '%s' 缺少 API Token", profileName)
	}

	return &profile, nil
}

// GetAWSProfile 获取 AWS profile
func GetAWSProfile(profileName string) (*AWSProfile, error) {
	cfg := Get()
	if cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}

	// 如果未指定 profile，使用默认 profile
	if profileName == "" {
		profileName = cfg.DefaultProfile.AWS
	}

	profile, ok := cfg.AWS[profileName]
	if !ok {
		return nil, fmt.Errorf("AWS profile '%s' 不存在", profileName)
	}

	if profile.AccessKeyID == "" || profile.SecretAccessKey == "" {
		return nil, fmt.Errorf("AWS profile '%s' 缺少访问凭证", profileName)
	}

	return &profile, nil
}

// ListProfiles 列出所有 profile
func ListProfiles() (cfProfiles []string, awsProfiles []string) {
	cfg := Get()
	if cfg == nil {
		return nil, nil
	}

	// Cloudflare profiles
	for name := range cfg.Cloudflare {
		cfProfiles = append(cfProfiles, name)
	}

	// AWS profiles
	for name := range cfg.AWS {
		awsProfiles = append(awsProfiles, name)
	}

	return cfProfiles, awsProfiles
}

// Validate 验证配置
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("配置为空")
	}

	// 验证日志级别
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[cfg.Log.Level] {
		return fmt.Errorf("无效的日志级别: %s", cfg.Log.Level)
	}

	// 验证日志格式
	validLogFormats := map[string]bool{
		"text": true,
		"json": true,
	}
	if !validLogFormats[cfg.Log.Format] {
		return fmt.Errorf("无效的日志格式: %s", cfg.Log.Format)
	}

	// 验证输出格式
	validOutputFormats := map[string]bool{
		"table": true,
		"json":  true,
	}
	if !validOutputFormats[cfg.Output.Format] {
		return fmt.Errorf("无效的输出格式: %s", cfg.Output.Format)
	}

	// 验证 API 超时时间
	if cfg.API.Timeout <= 0 {
		return fmt.Errorf("API 超时时间必须大于 0")
	}

	// 验证重试配置
	if cfg.API.Retry.Enabled {
		if cfg.API.Retry.MaxAttempts <= 0 {
			return fmt.Errorf("最大重试次数必须大于 0")
		}
		if cfg.API.Retry.InitialDelay <= 0 {
			return fmt.Errorf("初始延迟必须大于 0")
		}
		if cfg.API.Retry.MaxDelay <= 0 {
			return fmt.Errorf("最大延迟必须大于 0")
		}
	}

	return nil
}

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetConfigForTest 设置配置（仅用于测试）
func SetConfigForTest(cfg *Config) {
	globalConfig = cfg
	if cfg != nil {
		configLoaded = true
	} else {
		configLoaded = false
	}
}

package config

// Config 表示完整的配置结构
type Config struct {
	DefaultProfile DefaultProfile        `mapstructure:"default_profile" yaml:"default_profile"`
	Cloudflare     map[string]CFProfile  `mapstructure:"cloudflare" yaml:"cloudflare"`
	AWS            map[string]AWSProfile `mapstructure:"aws" yaml:"aws"`
	Log            LogConfig             `mapstructure:"log" yaml:"log"`
	Output         OutputConfig          `mapstructure:"output" yaml:"output"`
	API            APIConfig             `mapstructure:"api" yaml:"api"`
}

// DefaultProfile 默认 profile 配置
type DefaultProfile struct {
	Cloudflare string `mapstructure:"cloudflare" yaml:"cloudflare"`
	AWS        string `mapstructure:"aws" yaml:"aws"`
}

// CFProfile Cloudflare profile 配置
type CFProfile struct {
	APIToken string `mapstructure:"api_token" yaml:"api_token"`
}

// AWSProfile AWS profile 配置
type AWSProfile struct {
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key" yaml:"secret_access_key"`
	Region          string `mapstructure:"region" yaml:"region"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level     string `mapstructure:"level" yaml:"level"`           // debug, info, warn, error
	Format    string `mapstructure:"format" yaml:"format"`         // text, json
	Output    string `mapstructure:"output" yaml:"output"`         // stdout, stderr, file path
	AddSource bool   `mapstructure:"add_source" yaml:"add_source"` // 是否添加源代码位置信息
}

// OutputConfig 输出配置
type OutputConfig struct {
	Format string `mapstructure:"format" yaml:"format"` // table, json
	Color  bool   `mapstructure:"color" yaml:"color"`   // 是否启用颜色输出
}

// APIConfig API 配置
type APIConfig struct {
	Timeout int         `mapstructure:"timeout" yaml:"timeout"` // 超时时间（秒）
	Retry   RetryConfig `mapstructure:"retry" yaml:"retry"`     // 重试配置
}

// RetryConfig 重试配置
type RetryConfig struct {
	Enabled      bool `mapstructure:"enabled" yaml:"enabled"`             // 是否启用重试
	MaxAttempts  int  `mapstructure:"max_attempts" yaml:"max_attempts"`   // 最大重试次数
	InitialDelay int  `mapstructure:"initial_delay" yaml:"initial_delay"` // 初始延迟（秒）
	MaxDelay     int  `mapstructure:"max_delay" yaml:"max_delay"`         // 最大延迟（秒）
}

package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"

	"github.com/ado1t/cloudctl/internal/config"
	"github.com/ado1t/cloudctl/internal/logger"
)

// Client AWS 客户端封装
type Client struct {
	cfg              aws.Config
	cloudfrontClient *cloudfront.Client
	acmClient        *acm.Client
	profile          string
}

// NewClient 创建新的 AWS 客户端
func NewClient(profile string) (*Client, error) {
	logger.Debug("创建 AWS 客户端", "profile", profile)

	// 获取配置
	cfg := config.Get()
	if cfg == nil {
		return nil, fmt.Errorf("配置未初始化")
	}

	// 如果未指定 profile，使用默认值
	if profile == "" {
		profile = cfg.DefaultProfile.AWS
		if profile == "" {
			return nil, fmt.Errorf("未指定 AWS profile 且未配置默认 profile")
		}
	}

	// 获取 AWS profile 配置
	awsProfile, ok := cfg.AWS[profile]
	if !ok {
		return nil, fmt.Errorf("AWS profile '%s' 不存在", profile)
	}

	// 验证必需的配置项
	if awsProfile.AccessKeyID == "" {
		return nil, fmt.Errorf("AWS profile '%s' 缺少 access_key_id", profile)
	}
	if awsProfile.SecretAccessKey == "" {
		return nil, fmt.Errorf("AWS profile '%s' 缺少 secret_access_key", profile)
	}

	// 创建 AWS 配置
	// CloudFront 是全局服务，不需要指定 region
	awsConfig := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(
			awsProfile.AccessKeyID,
			awsProfile.SecretAccessKey,
			"",
		),
		Region: "us-east-1", // CloudFront 使用 us-east-1
	}

	// 创建 CloudFront 客户端
	cfClient := cloudfront.NewFromConfig(awsConfig)

	// 创建 ACM 客户端 (ACM 证书也必须在 us-east-1 区域)
	acmClient := acm.NewFromConfig(awsConfig)

	logger.Info("AWS 客户端创建成功", "profile", profile)

	return &Client{
		cfg:              awsConfig,
		cloudfrontClient: cfClient,
		acmClient:        acmClient,
		profile:          profile,
	}, nil
}

// CloudFront 返回 CloudFront 客户端
func (c *Client) CloudFront() *cloudfront.Client {
	return c.cloudfrontClient
}

// ACM 返回 ACM 客户端
func (c *Client) ACM() *acm.Client {
	return c.acmClient
}

// Profile 返回当前使用的 profile
func (c *Client) Profile() string {
	return c.profile
}

// Config 返回 AWS 配置
func (c *Client) Config() aws.Config {
	return c.cfg
}

// TestConnection 测试 AWS 连接
func (c *Client) TestConnection(ctx context.Context) error {
	logger.Debug("测试 AWS 连接")

	// 尝试列出 CloudFront 分发（限制为 1 条）
	maxItems := int32(1)
	_, err := c.cloudfrontClient.ListDistributions(ctx, &cloudfront.ListDistributionsInput{
		MaxItems: &maxItems,
	})
	if err != nil {
		return fmt.Errorf("AWS 连接测试失败: %w", err)
	}

	logger.Debug("AWS 连接测试成功")
	return nil
}

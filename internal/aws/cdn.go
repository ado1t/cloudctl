package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"

	"github.com/ado1t/cloudctl/internal/logger"
)

// Distribution CloudFront 分发信息
type Distribution struct {
	ID          string
	DomainName  string
	Status      string
	Enabled     bool
	Origins     []Origin
	Aliases     []string
	Comment     string
	CreatedTime time.Time
}

// Origin 源站信息
type Origin struct {
	ID         string
	DomainName string
	Path       string
}

// ListDistributions 列出所有 CloudFront 分发
func (c *Client) ListDistributions(ctx context.Context) ([]Distribution, error) {
	logger.Debug("列出 CloudFront 分发")

	var distributions []Distribution
	var marker *string

	for {
		input := &cloudfront.ListDistributionsInput{
			Marker: marker,
		}

		output, err := c.cloudfrontClient.ListDistributions(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("列出分发失败: %w", err)
		}

		if output.DistributionList != nil {
			for _, item := range output.DistributionList.Items {
				dist := convertDistributionSummary(item)
				distributions = append(distributions, dist)
			}

			// 检查是否有更多结果
			if output.DistributionList.IsTruncated != nil && *output.DistributionList.IsTruncated {
				marker = output.DistributionList.NextMarker
			} else {
				break
			}
		} else {
			break
		}
	}

	logger.Info("成功列出分发", "count", len(distributions))
	return distributions, nil
}

// GetDistribution 获取指定分发的详细信息
func (c *Client) GetDistribution(ctx context.Context, distributionID string) (*Distribution, error) {
	logger.Debug("获取分发详情", "id", distributionID)

	input := &cloudfront.GetDistributionInput{
		Id: &distributionID,
	}

	output, err := c.cloudfrontClient.GetDistribution(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("获取分发失败: %w", err)
	}

	if output.Distribution == nil {
		return nil, fmt.Errorf("分发不存在")
	}

	dist := convertDistribution(output.Distribution)
	logger.Info("成功获取分发详情", "id", distributionID)
	return &dist, nil
}

// CreateDistributionInput 创建分发的输入参数
type CreateDistributionInput struct {
	OriginDomain      string   // 源站域名
	Aliases           []string // 自定义域名
	Comment           string   // 备注
	Enabled           bool     // 是否启用
	CertificateARN    string   // SSL 证书 ARN（可选）
	PriceClass        string   // 价格等级（可选）
	DefaultRootObject string   // 默认根对象（可选）
}

// CreateDistribution 创建 CloudFront 分发
func (c *Client) CreateDistribution(ctx context.Context, input *CreateDistributionInput) (*Distribution, error) {
	logger.Debug("创建 CloudFront 分发", "origin", input.OriginDomain)

	// 生成唯一的 CallerReference
	callerReference := fmt.Sprintf("cloudctl-%d", time.Now().Unix())

	// 构建源站配置
	originID := "origin-1"
	origins := &types.Origins{
		Quantity: intPtr(1),
		Items: []types.Origin{
			{
				Id:         &originID,
				DomainName: &input.OriginDomain,
				CustomOriginConfig: &types.CustomOriginConfig{
					HTTPPort:             intPtr(80),
					HTTPSPort:            intPtr(443),
					OriginProtocolPolicy: types.OriginProtocolPolicyHttpsOnly,
				},
			},
		},
	}

	// 构建默认缓存行为
	defaultCacheBehavior := &types.DefaultCacheBehavior{
		TargetOriginId:       &originID,
		ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps,
		AllowedMethods: &types.AllowedMethods{
			Quantity: intPtr(7),
			Items: []types.Method{
				types.MethodGet,
				types.MethodHead,
				types.MethodOptions,
				types.MethodPut,
				types.MethodPost,
				types.MethodPatch,
				types.MethodDelete,
			},
			CachedMethods: &types.CachedMethods{
				Quantity: intPtr(2),
				Items: []types.Method{
					types.MethodGet,
					types.MethodHead,
				},
			},
		},
		Compress: boolPtr(true),
		ForwardedValues: &types.ForwardedValues{
			QueryString: boolPtr(true),
			Cookies: &types.CookiePreference{
				Forward: types.ItemSelectionAll,
			},
			Headers: &types.Headers{
				Quantity: intPtr(0),
			},
		},
		MinTTL:     int64Ptr(0),
		DefaultTTL: int64Ptr(86400),
		MaxTTL:     int64Ptr(31536000),
		TrustedSigners: &types.TrustedSigners{
			Enabled:  boolPtr(false),
			Quantity: intPtr(0),
		},
	}

	// 构建分发配置
	distConfig := &types.DistributionConfig{
		CallerReference:      &callerReference,
		Origins:              origins,
		DefaultCacheBehavior: defaultCacheBehavior,
		Comment:              &input.Comment,
		Enabled:              &input.Enabled,
	}

	// 设置自定义域名
	if len(input.Aliases) > 0 {
		distConfig.Aliases = &types.Aliases{
			Quantity: intPtr(len(input.Aliases)),
			Items:    input.Aliases,
		}

		// 如果有自定义域名，需要配置 SSL 证书
		if input.CertificateARN != "" {
			minProto := types.MinimumProtocolVersion("TLSv1.2_2021")
			distConfig.ViewerCertificate = &types.ViewerCertificate{
				ACMCertificateArn:      &input.CertificateARN,
				SSLSupportMethod:       types.SSLSupportMethodSniOnly,
				MinimumProtocolVersion: minProto,
			}
		}
	} else {
		// 没有自定义域名，使用默认证书
		distConfig.ViewerCertificate = &types.ViewerCertificate{
			CloudFrontDefaultCertificate: boolPtr(true),
		}
	}

	// 设置价格等级
	if input.PriceClass != "" {
		distConfig.PriceClass = types.PriceClass(input.PriceClass)
	} else {
		distConfig.PriceClass = types.PriceClassPriceClass100
	}

	// 设置默认根对象
	if input.DefaultRootObject != "" {
		distConfig.DefaultRootObject = &input.DefaultRootObject
	}

	// 创建分发
	createInput := &cloudfront.CreateDistributionInput{
		DistributionConfig: distConfig,
	}

	output, err := c.cloudfrontClient.CreateDistribution(ctx, createInput)
	if err != nil {
		return nil, fmt.Errorf("创建分发失败: %w", err)
	}

	if output.Distribution == nil {
		return nil, fmt.Errorf("创建分发失败: 返回结果为空")
	}

	dist := convertDistribution(output.Distribution)
	logger.Info("成功创建分发", "id", dist.ID)
	return &dist, nil
}

// UpdateDistributionInput 更新分发的输入参数
type UpdateDistributionInput struct {
	DistributionID string
	Comment        *string
	Enabled        *bool
}

// UpdateDistribution 更新 CloudFront 分发
func (c *Client) UpdateDistribution(ctx context.Context, input *UpdateDistributionInput) (*Distribution, error) {
	logger.Debug("更新 CloudFront 分发", "id", input.DistributionID)

	// 先获取当前配置
	getOutput, err := c.cloudfrontClient.GetDistribution(ctx, &cloudfront.GetDistributionInput{
		Id: &input.DistributionID,
	})
	if err != nil {
		return nil, fmt.Errorf("获取分发配置失败: %w", err)
	}

	if getOutput.Distribution == nil || getOutput.Distribution.DistributionConfig == nil {
		return nil, fmt.Errorf("分发不存在")
	}

	// 更新配置
	config := getOutput.Distribution.DistributionConfig
	if input.Comment != nil {
		config.Comment = input.Comment
	}
	if input.Enabled != nil {
		config.Enabled = input.Enabled
	}

	// 执行更新
	updateOutput, err := c.cloudfrontClient.UpdateDistribution(ctx, &cloudfront.UpdateDistributionInput{
		Id:                 &input.DistributionID,
		DistributionConfig: config,
		IfMatch:            getOutput.ETag,
	})
	if err != nil {
		return nil, fmt.Errorf("更新分发失败: %w", err)
	}

	if updateOutput.Distribution == nil {
		return nil, fmt.Errorf("更新分发失败: 返回结果为空")
	}

	dist := convertDistribution(updateOutput.Distribution)
	logger.Info("成功更新分发", "id", input.DistributionID)
	return &dist, nil
}

// convertDistributionSummary 转换分发摘要信息
func convertDistributionSummary(item types.DistributionSummary) Distribution {
	dist := Distribution{
		ID:         safeString(item.Id),
		DomainName: safeString(item.DomainName),
		Status:     safeString(item.Status),
		Enabled:    safeBool(item.Enabled),
		Comment:    safeString(item.Comment),
	}

	// 转换别名
	if item.Aliases != nil && item.Aliases.Items != nil {
		dist.Aliases = item.Aliases.Items
	}

	// 转换源站
	if item.Origins != nil && item.Origins.Items != nil {
		for _, origin := range item.Origins.Items {
			dist.Origins = append(dist.Origins, Origin{
				ID:         safeString(origin.Id),
				DomainName: safeString(origin.DomainName),
				Path:       safeString(origin.OriginPath),
			})
		}
	}

	return dist
}

// convertDistribution 转换完整的分发信息
func convertDistribution(dist *types.Distribution) Distribution {
	result := Distribution{
		ID:         safeString(dist.Id),
		DomainName: safeString(dist.DomainName),
		Status:     safeString(dist.Status),
	}

	if dist.DistributionConfig != nil {
		result.Enabled = safeBool(dist.DistributionConfig.Enabled)
		result.Comment = safeString(dist.DistributionConfig.Comment)

		// 转换别名
		if dist.DistributionConfig.Aliases != nil && dist.DistributionConfig.Aliases.Items != nil {
			result.Aliases = dist.DistributionConfig.Aliases.Items
		}

		// 转换源站
		if dist.DistributionConfig.Origins != nil && dist.DistributionConfig.Origins.Items != nil {
			for _, origin := range dist.DistributionConfig.Origins.Items {
				result.Origins = append(result.Origins, Origin{
					ID:         safeString(origin.Id),
					DomainName: safeString(origin.DomainName),
					Path:       safeString(origin.OriginPath),
				})
			}
		}
	}

	return result
}

// 辅助函数
func intPtr(i int) *int32 {
	v := int32(i)
	return &v
}

func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// Invalidation CloudFront 缓存失效信息
type Invalidation struct {
	ID              string
	Status          string
	CreateTime      time.Time
	CallerReference string
	Paths           []string
}

// CreateInvalidationInput 创建缓存失效的输入参数
type CreateInvalidationInput struct {
	DistributionID  string
	Paths           []string
	CallerReference string // 可选，如果为空则自动生成
}

// CreateInvalidation 创建 CloudFront 缓存失效
func (c *Client) CreateInvalidation(ctx context.Context, input *CreateInvalidationInput) (*Invalidation, error) {
	logger.Debug("创建 CloudFront 缓存失效", "distribution_id", input.DistributionID, "paths", input.Paths)

	// 验证参数
	if input.DistributionID == "" {
		return nil, fmt.Errorf("distribution_id 不能为空")
	}
	if len(input.Paths) == 0 {
		return nil, fmt.Errorf("至少需要指定一个路径")
	}

	// 如果未指定 CallerReference，自动生成
	callerReference := input.CallerReference
	if callerReference == "" {
		callerReference = fmt.Sprintf("cloudctl-%d", time.Now().Unix())
	}

	// 构建失效请求
	quantity := int32(len(input.Paths))
	invalidationBatch := &types.InvalidationBatch{
		CallerReference: &callerReference,
		Paths: &types.Paths{
			Quantity: &quantity,
			Items:    input.Paths,
		},
	}

	// 创建失效请求
	createInput := &cloudfront.CreateInvalidationInput{
		DistributionId:    &input.DistributionID,
		InvalidationBatch: invalidationBatch,
	}

	output, err := c.cloudfrontClient.CreateInvalidation(ctx, createInput)
	if err != nil {
		return nil, fmt.Errorf("创建缓存失效失败: %w", err)
	}

	if output.Invalidation == nil {
		return nil, fmt.Errorf("返回的失效信息为空")
	}

	// 转换结果
	invalidation := &Invalidation{
		ID:              safeString(output.Invalidation.Id),
		Status:          safeString(output.Invalidation.Status),
		CreateTime:      safeTime(output.Invalidation.CreateTime),
		CallerReference: callerReference,
		Paths:           input.Paths,
	}

	logger.Info("成功创建缓存失效", "id", invalidation.ID, "distribution_id", input.DistributionID)
	return invalidation, nil
}

// GetInvalidation 获取缓存失效状态
func (c *Client) GetInvalidation(ctx context.Context, distributionID, invalidationID string) (*Invalidation, error) {
	logger.Debug("获取缓存失效状态", "distribution_id", distributionID, "invalidation_id", invalidationID)

	input := &cloudfront.GetInvalidationInput{
		DistributionId: &distributionID,
		Id:             &invalidationID,
	}

	output, err := c.cloudfrontClient.GetInvalidation(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("获取缓存失效失败: %w", err)
	}

	if output.Invalidation == nil {
		return nil, fmt.Errorf("返回的失效信息为空")
	}

	// 获取路径列表
	var paths []string
	if output.Invalidation.InvalidationBatch != nil &&
		output.Invalidation.InvalidationBatch.Paths != nil {
		paths = output.Invalidation.InvalidationBatch.Paths.Items
	}

	invalidation := &Invalidation{
		ID:              safeString(output.Invalidation.Id),
		Status:          safeString(output.Invalidation.Status),
		CreateTime:      safeTime(output.Invalidation.CreateTime),
		CallerReference: safeString(output.Invalidation.InvalidationBatch.CallerReference),
		Paths:           paths,
	}

	logger.Info("成功获取缓存失效状态", "id", invalidation.ID, "status", invalidation.Status)
	return invalidation, nil
}

// ListInvalidations 列出分发的所有缓存失效
func (c *Client) ListInvalidations(ctx context.Context, distributionID string) ([]Invalidation, error) {
	logger.Debug("列出缓存失效", "distribution_id", distributionID)

	var invalidations []Invalidation
	var marker *string

	for {
		input := &cloudfront.ListInvalidationsInput{
			DistributionId: &distributionID,
			Marker:         marker,
		}

		output, err := c.cloudfrontClient.ListInvalidations(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("列出缓存失效失败: %w", err)
		}

		if output.InvalidationList != nil {
			for _, item := range output.InvalidationList.Items {
				invalidations = append(invalidations, Invalidation{
					ID:         safeString(item.Id),
					Status:     safeString(item.Status),
					CreateTime: safeTime(item.CreateTime),
				})
			}

			// 检查是否有更多结果
			if output.InvalidationList.IsTruncated != nil && *output.InvalidationList.IsTruncated {
				marker = output.InvalidationList.NextMarker
			} else {
				break
			}
		} else {
			break
		}
	}

	logger.Info("成功列出缓存失效", "count", len(invalidations))
	return invalidations, nil
}

func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// DistributionsConfig CloudFront 分发批量创建配置
type DistributionsConfig struct {
	Distributions []DistributionConfig `yaml:"distributions"`
}

// DistributionConfig 单个分发配置
type DistributionConfig struct {
	Name           string           `yaml:"name"`
	Aliases        []string         `yaml:"aliases"`
	CertificateARN string           `yaml:"certificate_arn"`
	WafARN         string           `yaml:"waf_arn"` // 可选
	Origin         OriginConfig     `yaml:"origin"`
	Behaviors      []BehaviorConfig `yaml:"behaviors"`
}

// OriginConfig 源配置
type OriginConfig struct {
	Domain string `yaml:"domain"`
}

// BehaviorConfig 缓存行为配置
type BehaviorConfig struct {
	Priority              int    `yaml:"priority"`
	PathPattern           string `yaml:"path_pattern"`
	ViewerProtocolPolicy  string `yaml:"viewer_protocol_policy"`
	CachePolicy           string `yaml:"cache_policy"`
	OriginRequestPolicy   string `yaml:"origin_request_policy"`
	ResponseHeadersPolicy string `yaml:"response_headers_policy"`
}

// AWS CloudFront 托管策略 ID 映射
var (
	CachePolicyIDs = map[string]string{
		"Managed-CachingDisabled":  "4135ea2d-6df8-44a3-9df3-4b5a84be39ad",
		"Managed-CachingOptimized": "658327ea-f89d-4fab-a63d-7e88639e58f6",
	}

	OriginRequestPolicyIDs = map[string]string{
		"Managed-AllViewer": "216adef6-5c7f-47e4-b989-5492eafa07d3",
	}

	ResponseHeadersPolicyIDs = map[string]string{
		"Managed-SimpleCORS": "60669652-455b-4ae9-85a4-c4c02393f86c",
	}
)

// DistributionCreateResult 单个分发创建结果
type DistributionCreateResult struct {
	Name           string
	Success        bool
	DistributionID string
	DomainName     string
	Error          string
}

// BatchCreateResult 批量创建结果
type BatchCreateResult struct {
	Total   int
	Success int
	Failed  int
	Results []DistributionCreateResult
}

// ValidateCertificates 验证所有证书状态是否为 ISSUED
func (c *Client) ValidateCertificates(ctx context.Context, certificateARNs []string) error {
	logger.Debug("验证证书状态", "count", len(certificateARNs))

	for _, arn := range certificateARNs {
		if arn == "" {
			continue
		}

		cert, err := c.GetCertificate(ctx, arn)
		if err != nil {
			return fmt.Errorf("获取证书失败 %s: %w", arn, err)
		}

		if cert.Status != "ISSUED" {
			return fmt.Errorf("证书 %s 状态不是 ISSUED，当前状态: %s", arn, cert.Status)
		}

		logger.Info("证书验证通过", "arn", arn, "status", cert.Status)
	}

	logger.Info("所有证书验证通过", "count", len(certificateARNs))
	return nil
}

// CreateDistributionWithConfig 创建 CloudFront 分发（使用配置结构）
func (c *Client) CreateDistributionWithConfig(ctx context.Context, config DistributionConfig) (*DistributionCreateResult, error) {
	logger.Debug("开始创建 CloudFront 分发", "name", config.Name)

	// 生成唯一的调用者引用
	callerReference := fmt.Sprintf("%s-%d", config.Name, time.Now().Unix())

	// 构建源配置
	originID := fmt.Sprintf("%s-origin", config.Name)
	origin := &types.Origin{
		Id:         &originID,
		DomainName: &config.Origin.Domain,
		CustomOriginConfig: &types.CustomOriginConfig{
			HTTPPort:             aws.Int32(80),
			HTTPSPort:            aws.Int32(443),
			OriginProtocolPolicy: types.OriginProtocolPolicyHttpOnly, // 使用 HTTP
			OriginSslProtocols: &types.OriginSslProtocols{
				Items:    []types.SslProtocol{types.SslProtocolTLSv12},
				Quantity: aws.Int32(1),
			},
		},
	}

	// 构建缓存行为
	cacheBehaviors, defaultBehavior, err := c.buildCacheBehaviors(config.Behaviors, originID)
	if err != nil {
		return nil, fmt.Errorf("构建缓存行为失败: %w", err)
	}

	// 构建分发配置
	distributionConfig := &types.DistributionConfig{
		CallerReference: &callerReference,
		Comment:         aws.String(config.Name),
		Enabled:         aws.Bool(true),
		Origins: &types.Origins{
			Items:    []types.Origin{*origin},
			Quantity: aws.Int32(1),
		},
		DefaultCacheBehavior: defaultBehavior,
		Aliases: &types.Aliases{
			Items:    config.Aliases,
			Quantity: aws.Int32(int32(len(config.Aliases))),
		},
		ViewerCertificate: &types.ViewerCertificate{
			ACMCertificateArn:      &config.CertificateARN,
			SSLSupportMethod:       types.SSLSupportMethodSniOnly,
			MinimumProtocolVersion: types.MinimumProtocolVersion("TLSv1.2_2021"),
		},
	}

	// 添加额外的缓存行为（如果有）
	if len(cacheBehaviors) > 0 {
		distributionConfig.CacheBehaviors = &types.CacheBehaviors{
			Items:    cacheBehaviors,
			Quantity: aws.Int32(int32(len(cacheBehaviors))),
		}
	}

	// 添加 WAF（如果配置）
	if config.WafARN != "" {
		distributionConfig.WebACLId = &config.WafARN
	}

	// 创建分发
	input := &cloudfront.CreateDistributionInput{
		DistributionConfig: distributionConfig,
	}

	output, err := c.cloudfrontClient.CreateDistribution(ctx, input)
	if err != nil {
		return &DistributionCreateResult{
			Name:    config.Name,
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// 添加 Name 标签
	// CloudFront 分发的 ARN 格式：arn:aws:cloudfront::account-id:distribution/distribution-id
	distributionARN := *output.Distribution.ARN
	tagInput := &cloudfront.TagResourceInput{
		Resource: &distributionARN,
		Tags: &types.Tags{
			Items: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(config.Name),
				},
			},
		},
	}

	if _, err := c.cloudfrontClient.TagResource(ctx, tagInput); err != nil {
		logger.Warn("添加标签失败", "distribution_id", *output.Distribution.Id, "error", err)
		// 标签失败不影响分发创建成功
	}

	result := &DistributionCreateResult{
		Name:           config.Name,
		Success:        true,
		DistributionID: *output.Distribution.Id,
		DomainName:     *output.Distribution.DomainName,
	}

	logger.Info("成功创建 CloudFront 分发",
		"name", config.Name,
		"id", result.DistributionID,
		"domain", result.DomainName)

	return result, nil
}

// buildCacheBehaviors 构建缓存行为配置
func (c *Client) buildCacheBehaviors(behaviors []BehaviorConfig, originID string) ([]types.CacheBehavior, *types.DefaultCacheBehavior, error) {
	if len(behaviors) == 0 {
		return nil, nil, fmt.Errorf("至少需要一个缓存行为配置")
	}

	var cacheBehaviors []types.CacheBehavior
	var defaultBehavior *types.DefaultCacheBehavior

	for _, behavior := range behaviors {
		// 解析策略 ID
		cachePolicyID, err := c.resolvePolicyID(behavior.CachePolicy, CachePolicyIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("解析缓存策略失败: %w", err)
		}

		originRequestPolicyID, err := c.resolvePolicyID(behavior.OriginRequestPolicy, OriginRequestPolicyIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("解析源请求策略失败: %w", err)
		}

		responseHeadersPolicyID, err := c.resolvePolicyID(behavior.ResponseHeadersPolicy, ResponseHeadersPolicyIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("解析响应头策略失败: %w", err)
		}

		// 解析 Viewer Protocol Policy
		viewerProtocolPolicy, err := c.parseViewerProtocolPolicy(behavior.ViewerProtocolPolicy)
		if err != nil {
			return nil, nil, err
		}

		// 优先级 1 且路径为 "*" 的是默认行为
		if behavior.Priority == 1 && behavior.PathPattern == "*" {
			defaultBehavior = &types.DefaultCacheBehavior{
				TargetOriginId:          &originID,
				ViewerProtocolPolicy:    viewerProtocolPolicy,
				CachePolicyId:           &cachePolicyID,
				OriginRequestPolicyId:   &originRequestPolicyID,
				ResponseHeadersPolicyId: &responseHeadersPolicyID,
				Compress:                aws.Bool(true),
				AllowedMethods: &types.AllowedMethods{
					Items:    []types.Method{types.MethodGet, types.MethodHead, types.MethodOptions, types.MethodPut, types.MethodPost, types.MethodPatch, types.MethodDelete},
					Quantity: aws.Int32(7),
					CachedMethods: &types.CachedMethods{
						Items:    []types.Method{types.MethodGet, types.MethodHead},
						Quantity: aws.Int32(2),
					},
				},
			}
		} else {
			// 其他的作为额外的缓存行为
			cacheBehavior := types.CacheBehavior{
				PathPattern:             &behavior.PathPattern,
				TargetOriginId:          &originID,
				ViewerProtocolPolicy:    viewerProtocolPolicy,
				CachePolicyId:           &cachePolicyID,
				OriginRequestPolicyId:   &originRequestPolicyID,
				ResponseHeadersPolicyId: &responseHeadersPolicyID,
				Compress:                aws.Bool(true),
				AllowedMethods: &types.AllowedMethods{
					Items:    []types.Method{types.MethodGet, types.MethodHead, types.MethodOptions, types.MethodPut, types.MethodPost, types.MethodPatch, types.MethodDelete},
					Quantity: aws.Int32(7),
					CachedMethods: &types.CachedMethods{
						Items:    []types.Method{types.MethodGet, types.MethodHead},
						Quantity: aws.Int32(2),
					},
				},
			}
			cacheBehaviors = append(cacheBehaviors, cacheBehavior)
		}
	}

	if defaultBehavior == nil {
		return nil, nil, fmt.Errorf("未找到默认缓存行为（priority=1, path_pattern='*'）")
	}

	return cacheBehaviors, defaultBehavior, nil
}

// resolvePolicyID 解析策略 ID
func (c *Client) resolvePolicyID(policyName string, policyMap map[string]string) (string, error) {
	if policyID, ok := policyMap[policyName]; ok {
		return policyID, nil
	}
	// 如果不在映射表中，假设直接提供的是 ID
	return policyName, nil
}

// parseViewerProtocolPolicy 解析 Viewer Protocol Policy
func (c *Client) parseViewerProtocolPolicy(policy string) (types.ViewerProtocolPolicy, error) {
	switch policy {
	case "redirect-to-https":
		return types.ViewerProtocolPolicyRedirectToHttps, nil
	case "allow-all":
		return types.ViewerProtocolPolicyAllowAll, nil
	case "https-only":
		return types.ViewerProtocolPolicyHttpsOnly, nil
	default:
		return "", fmt.Errorf("不支持的 viewer protocol policy: %s", policy)
	}
}

// BatchCreateDistributions 批量创建 CloudFront 分发
func (c *Client) BatchCreateDistributions(ctx context.Context, configs []DistributionConfig) *BatchCreateResult {
	logger.Debug("批量创建 CloudFront 分发", "count", len(configs))

	result := &BatchCreateResult{
		Total:   len(configs),
		Results: make([]DistributionCreateResult, 0, len(configs)),
	}

	for i, config := range configs {
		logger.Info("创建分发", "progress", fmt.Sprintf("%d/%d", i+1, len(configs)), "name", config.Name)

		distResult, err := c.CreateDistributionWithConfig(ctx, config)
		if err != nil {
			logger.Error("创建分发失败", "name", config.Name, "error", err)
			result.Failed++
			if distResult != nil {
				result.Results = append(result.Results, *distResult)
			}
			continue
		}

		result.Success++
		result.Results = append(result.Results, *distResult)
		logger.Info("成功创建分发", "name", config.Name, "id", distResult.DistributionID)
	}

	logger.Info("批量创建完成", "total", result.Total, "success", result.Success, "failed", result.Failed)
	return result
}

package aws

import (
	"context"
	"fmt"
	"time"

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

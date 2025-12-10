package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"

	"github.com/ado1t/cloudctl/internal/logger"
)

// Certificate ACM 证书信息
type Certificate struct {
	ARN               string
	DomainName        string
	Status            string
	Type              string
	InUse             bool
	IssuedAt          time.Time
	NotBefore         time.Time
	NotAfter          time.Time
	SubjectAltNames   []string
	ValidationRecords []ValidationRecord
}

// ValidationRecord DNS 验证记录
type ValidationRecord struct {
	Name   string
	Type   string
	Value  string
	Status string
}

// RequestCertificateInput 申请证书的输入参数
type RequestCertificateInput struct {
	DomainName              string
	SubjectAlternativeNames []string
}

// CertificatesConfig 批量证书配置
type CertificatesConfig struct {
	Certificates []CertificateRequest `yaml:"certificates"`
}

// CertificateRequest 证书申请请求
type CertificateRequest struct {
	Domain string   `yaml:"domain"`
	SANs   []string `yaml:"san"`
}

// BatchRequestResult 批量申请结果
type BatchRequestResult struct {
	Total   int
	Success int
	Failed  int
	Results []CertificateResult
}

// CertificateResult 单个证书申请结果
type CertificateResult struct {
	Domain            string
	Success           bool
	ARN               string
	Error             string
	ValidationRecords []ValidationRecord
}

// ListCertificates 列出所有证书
func (c *Client) ListCertificates(ctx context.Context) ([]Certificate, error) {
	logger.Debug("列出 ACM 证书")

	var certificates []Certificate
	var nextToken *string

	for {
		input := &acm.ListCertificatesInput{
			NextToken: nextToken,
		}

		output, err := c.acmClient.ListCertificates(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("列出证书失败: %w", err)
		}

		for _, cert := range output.CertificateSummaryList {
			certificates = append(certificates, Certificate{
				ARN:        safeString(cert.CertificateArn),
				DomainName: safeString(cert.DomainName),
				Status:     string(cert.Status),
				Type:       string(cert.Type),
				InUse:      safeBool(cert.InUse),
				NotBefore:  safeTime(cert.NotBefore),
				NotAfter:   safeTime(cert.NotAfter),
			})
		}

		// 检查是否有更多结果
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	logger.Info("成功列出证书", "count", len(certificates))
	return certificates, nil
}

// GetCertificate 获取证书详情
func (c *Client) GetCertificate(ctx context.Context, certificateARN string) (*Certificate, error) {
	logger.Debug("获取证书详情", "arn", certificateARN)

	input := &acm.DescribeCertificateInput{
		CertificateArn: &certificateARN,
	}

	output, err := c.acmClient.DescribeCertificate(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("获取证书详情失败: %w", err)
	}

	if output.Certificate == nil {
		return nil, fmt.Errorf("证书不存在")
	}

	cert := output.Certificate

	// 提取 SubjectAlternativeNames
	var sans []string
	if cert.SubjectAlternativeNames != nil {
		sans = cert.SubjectAlternativeNames
	}

	// 提取验证记录
	var validationRecords []ValidationRecord
	if cert.DomainValidationOptions != nil {
		for _, opt := range cert.DomainValidationOptions {
			if opt.ResourceRecord != nil {
				validationRecords = append(validationRecords, ValidationRecord{
					Name:   safeString(opt.ResourceRecord.Name),
					Type:   string(opt.ResourceRecord.Type),
					Value:  safeString(opt.ResourceRecord.Value),
					Status: string(opt.ValidationStatus),
				})
			}
		}
	}

	// 检查证书是否在使用中 (通过 InUseBy 列表判断)
	inUse := len(cert.InUseBy) > 0

	certificate := &Certificate{
		ARN:               safeString(cert.CertificateArn),
		DomainName:        safeString(cert.DomainName),
		Status:            string(cert.Status),
		Type:              string(cert.Type),
		InUse:             inUse,
		IssuedAt:          safeTime(cert.IssuedAt),
		NotBefore:         safeTime(cert.NotBefore),
		NotAfter:          safeTime(cert.NotAfter),
		SubjectAltNames:   sans,
		ValidationRecords: validationRecords,
	}

	logger.Info("成功获取证书详情", "arn", certificate.ARN)
	return certificate, nil
}

// RequestCertificate 申请新证书
func (c *Client) RequestCertificate(ctx context.Context, input *RequestCertificateInput) (*Certificate, error) {
	logger.Debug("申请 ACM 证书", "domain", input.DomainName, "sans", input.SubjectAlternativeNames)

	// 验证参数
	if input.DomainName == "" {
		return nil, fmt.Errorf("域名不能为空")
	}

	// 构建 SANs 列表 (包含主域名)
	sans := []string{input.DomainName}
	if len(input.SubjectAlternativeNames) > 0 {
		sans = append(sans, input.SubjectAlternativeNames...)
	}

	// 申请证书 (仅支持 DNS 验证)
	requestInput := &acm.RequestCertificateInput{
		DomainName:              &input.DomainName,
		SubjectAlternativeNames: sans,
		ValidationMethod:        types.ValidationMethodDns,
	}

	output, err := c.acmClient.RequestCertificate(ctx, requestInput)
	if err != nil {
		return nil, fmt.Errorf("申请证书失败: %w", err)
	}

	if output.CertificateArn == nil {
		return nil, fmt.Errorf("返回的证书 ARN 为空")
	}

	certificateARN := *output.CertificateArn
	logger.Info("成功申请证书", "arn", certificateARN)

	// 等待并重试获取验证记录 (AWS 生成验证记录需要时间)
	var certificate *Certificate
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			logger.Debug("等待 AWS 生成验证记录...", "retry", i, "max", maxRetries)
			time.Sleep(time.Duration(2+i) * time.Second) // 递增等待时间: 2s, 3s, 4s, 5s, 6s
		} else {
			time.Sleep(2 * time.Second) // 首次等待 2 秒
		}

		cert, err := c.GetCertificate(ctx, certificateARN)
		if err != nil {
			logger.Warn("获取证书详情失败", "error", err, "retry", i+1)
			continue
		}

		// 检查是否有验证记录
		if len(cert.ValidationRecords) > 0 {
			certificate = cert
			logger.Info("成功获取验证记录", "count", len(cert.ValidationRecords))
			break
		}

		logger.Debug("验证记录尚未生成", "retry", i+1)
	}

	// 如果重试后仍然没有获取到验证记录,返回基本信息
	if certificate == nil || len(certificate.ValidationRecords) == 0 {
		logger.Warn("未能获取到验证记录,请稍后使用 cert get 命令查看")
		return &Certificate{
			ARN:        certificateARN,
			DomainName: input.DomainName,
			Status:     "PENDING_VALIDATION",
		}, nil
	}

	return certificate, nil
}

// BatchRequestCertificates 批量申请证书
func (c *Client) BatchRequestCertificates(ctx context.Context, requests []CertificateRequest) *BatchRequestResult {
	logger.Debug("批量申请证书", "count", len(requests))

	result := &BatchRequestResult{
		Total:   len(requests),
		Results: make([]CertificateResult, 0, len(requests)),
	}

	for i, req := range requests {
		logger.Info("申请证书", "progress", fmt.Sprintf("%d/%d", i+1, len(requests)), "domain", req.Domain)

		// 准备申请参数
		input := &RequestCertificateInput{
			DomainName:              req.Domain,
			SubjectAlternativeNames: req.SANs,
		}

		// 申请证书
		cert, err := c.RequestCertificate(ctx, input)
		if err != nil {
			logger.Error("申请证书失败", "domain", req.Domain, "error", err)
			result.Failed++
			result.Results = append(result.Results, CertificateResult{
				Domain:  req.Domain,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		result.Success++
		result.Results = append(result.Results, CertificateResult{
			Domain:            req.Domain,
			Success:           true,
			ARN:               cert.ARN,
			ValidationRecords: cert.ValidationRecords,
		})

		logger.Info("成功申请证书", "domain", req.Domain, "arn", cert.ARN)
	}

	logger.Info("批量申请完成", "total", result.Total, "success", result.Success, "failed", result.Failed)
	return result
}

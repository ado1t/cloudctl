package cloudflare

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go/v2"
	"github.com/cloudflare/cloudflare-go/v2/dns"
)

// DNSRecordInfo DNS 记录信息
type DNSRecordInfo struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Name       string    `json:"name"`
	Content    string    `json:"content"`
	TTL        float64   `json:"ttl"`
	Proxied    bool      `json:"proxied"`
	Proxiable  bool      `json:"proxiable"`
	CreatedOn  time.Time `json:"created_on"`
	ModifiedOn time.Time `json:"modified_on"`
}

// DNSRecordCreateParams DNS 记录创建参数
type DNSRecordCreateParams struct {
	Type    string  // 记录类型: A, CNAME 等
	Name    string  // 记录名称
	Content string  // 记录内容
	TTL     float64 // TTL (1 = auto)
	Proxied bool    // 是否启用 Cloudflare 代理
}

// DNSRecordUpdateParams DNS 记录更新参数
type DNSRecordUpdateParams struct {
	Content *string  // 记录内容
	TTL     *float64 // TTL
	Proxied *bool    // 是否启用 Cloudflare 代理
}

// ListDNSRecords 列出指定 Zone 的所有 DNS 记录
func (c *Client) ListDNSRecords(ctx context.Context, zoneID string, recordType string) ([]DNSRecordInfo, error) {
	var allRecords []DNSRecordInfo

	c.logger.Debug("开始列出 DNS 记录", "zone_id", zoneID, "type", recordType)

	err := c.WithRetry(ctx, "列出 DNS 记录", func() error {
		// 构建查询参数
		params := dns.RecordListParams{
			ZoneID: cloudflare.F(zoneID),
		}

		// 如果指定了类型，添加类型过滤
		if recordType != "" {
			params.Type = cloudflare.F(dns.RecordListParamsType(recordType))
		}

		// 使用分页获取所有记录
		page := float64(1)
		perPage := float64(100)

		for {
			params.Page = cloudflare.F(page)
			params.PerPage = cloudflare.F(perPage)

			c.logger.Debug("获取 DNS 记录列表",
				"page", page,
				"per_page", perPage,
			)

			// 调用 Cloudflare API
			result, err := c.api.DNS.Records.List(ctx, params)
			if err != nil {
				return fmt.Errorf("获取 DNS 记录列表失败: %w", err)
			}

			// 转换为我们的数据结构
			for _, record := range result.Result {
				// Content 是 interface{} 类型，需要转换为字符串
				contentStr := ""
				if record.Content != nil {
					contentStr = fmt.Sprintf("%v", record.Content)
				}

				recordInfo := DNSRecordInfo{
					ID:         record.ID,
					Type:       string(record.Type),
					Name:       record.Name,
					Content:    contentStr,
					TTL:        float64(record.TTL),
					Proxied:    record.Proxied,
					Proxiable:  record.Proxiable,
					CreatedOn:  record.CreatedOn,
					ModifiedOn: record.ModifiedOn,
				}

				allRecords = append(allRecords, recordInfo)
			}

			c.logger.Debug("获取到 DNS 记录",
				"count", len(result.Result),
				"total", len(allRecords),
			)

			// 检查是否还有更多页
			if len(result.Result) < int(perPage) {
				break
			}

			page++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("成功列出 DNS 记录", "zone_id", zoneID, "total", len(allRecords))
	return allRecords, nil
}

// GetDNSRecord 获取指定的 DNS 记录
func (c *Client) GetDNSRecord(ctx context.Context, zoneID, recordID string) (*DNSRecordInfo, error) {
	var recordInfo *DNSRecordInfo

	c.logger.Debug("获取 DNS 记录", "zone_id", zoneID, "record_id", recordID)

	err := c.WithRetry(ctx, "获取 DNS 记录", func() error {
		result, err := c.api.DNS.Records.Get(ctx, recordID, dns.RecordGetParams{
			ZoneID: cloudflare.F(zoneID),
		})
		if err != nil {
			return fmt.Errorf("获取 DNS 记录失败: %w", err)
		}

		// Content 是 interface{} 类型，需要转换为字符串
		contentStr := ""
		if result.Content != nil {
			contentStr = fmt.Sprintf("%v", result.Content)
		}

		recordInfo = &DNSRecordInfo{
			ID:         result.ID,
			Type:       string(result.Type),
			Name:       result.Name,
			Content:    contentStr,
			TTL:        float64(result.TTL),
			Proxied:    result.Proxied,
			Proxiable:  result.Proxiable,
			CreatedOn:  result.CreatedOn,
			ModifiedOn: result.ModifiedOn,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("成功获取 DNS 记录", "record_id", recordID)
	return recordInfo, nil
}

// CreateDNSRecord 创建 DNS 记录
func (c *Client) CreateDNSRecord(ctx context.Context, zoneID string, params DNSRecordCreateParams) (*DNSRecordInfo, error) {
	var recordInfo *DNSRecordInfo

	c.logger.Info("创建 DNS 记录",
		"zone_id", zoneID,
		"type", params.Type,
		"name", params.Name,
		"content", params.Content,
	)

	err := c.WithRetry(ctx, "创建 DNS 记录", func() error {
		// 构建创建参数 - 根据记录类型创建不同的参数
		var recordParam dns.RecordUnionParam

		switch params.Type {
		case "A":
			recordParam = dns.ARecordParam{
				Type:    cloudflare.F(dns.ARecordTypeA),
				Name:    cloudflare.F(params.Name),
				Content: cloudflare.F(params.Content),
				TTL:     cloudflare.F(dns.TTL(params.TTL)),
				Proxied: cloudflare.F(params.Proxied),
			}
		case "AAAA":
			recordParam = dns.AAAARecordParam{
				Type:    cloudflare.F(dns.AAAARecordTypeAAAA),
				Name:    cloudflare.F(params.Name),
				Content: cloudflare.F(params.Content),
				TTL:     cloudflare.F(dns.TTL(params.TTL)),
				Proxied: cloudflare.F(params.Proxied),
			}
		case "CNAME":
			recordParam = dns.CNAMERecordParam{
				Type:    cloudflare.F(dns.CNAMERecordTypeCNAME),
				Name:    cloudflare.F(params.Name),
				Content: cloudflare.F[interface{}](params.Content),
				TTL:     cloudflare.F(dns.TTL(params.TTL)),
				Proxied: cloudflare.F(params.Proxied),
			}
		default:
			return fmt.Errorf("不支持的记录类型: %s", params.Type)
		}

		// 调用 API 创建记录
		result, err := c.api.DNS.Records.New(ctx, dns.RecordNewParams{
			ZoneID: cloudflare.F(zoneID),
			Record: recordParam,
		})
		if err != nil {
			return fmt.Errorf("创建 DNS 记录失败: %w", err)
		}

		// Content 是 interface{} 类型，需要转换为字符串
		contentStr := ""
		if result.Content != nil {
			contentStr = fmt.Sprintf("%v", result.Content)
		}

		recordInfo = &DNSRecordInfo{
			ID:         result.ID,
			Type:       string(result.Type),
			Name:       result.Name,
			Content:    contentStr,
			TTL:        float64(result.TTL),
			Proxied:    result.Proxied,
			Proxiable:  result.Proxiable,
			CreatedOn:  result.CreatedOn,
			ModifiedOn: result.ModifiedOn,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("成功创建 DNS 记录",
		"record_id", recordInfo.ID,
		"type", recordInfo.Type,
		"name", recordInfo.Name,
	)

	return recordInfo, nil
}

// UpdateDNSRecord 更新 DNS 记录
func (c *Client) UpdateDNSRecord(ctx context.Context, zoneID, recordID string, recordType string, params DNSRecordUpdateParams) (*DNSRecordInfo, error) {
	var recordInfo *DNSRecordInfo

	c.logger.Info("更新 DNS 记录", "zone_id", zoneID, "record_id", recordID)

	// 首先获取现有记录信息
	existingRecord, err := c.GetDNSRecord(ctx, zoneID, recordID)
	if err != nil {
		return nil, fmt.Errorf("获取现有记录失败: %w", err)
	}

	// 使用现有值作为默认值
	content := existingRecord.Content
	ttl := existingRecord.TTL
	proxied := existingRecord.Proxied

	// 应用更新的值
	if params.Content != nil {
		content = *params.Content
	}
	if params.TTL != nil {
		ttl = *params.TTL
	}
	if params.Proxied != nil {
		proxied = *params.Proxied
	}

	err = c.WithRetry(ctx, "更新 DNS 记录", func() error {
		// 构建更新参数 - 根据记录类型创建不同的参数
		var recordParam dns.RecordUnionParam

		switch recordType {
		case "A":
			recordParam = dns.ARecordParam{
				Type:    cloudflare.F(dns.ARecordTypeA),
				Name:    cloudflare.F(existingRecord.Name),
				Content: cloudflare.F(content),
				TTL:     cloudflare.F(dns.TTL(ttl)),
				Proxied: cloudflare.F(proxied),
			}
		case "AAAA":
			recordParam = dns.AAAARecordParam{
				Type:    cloudflare.F(dns.AAAARecordTypeAAAA),
				Name:    cloudflare.F(existingRecord.Name),
				Content: cloudflare.F(content),
				TTL:     cloudflare.F(dns.TTL(ttl)),
				Proxied: cloudflare.F(proxied),
			}
		case "CNAME":
			recordParam = dns.CNAMERecordParam{
				Type:    cloudflare.F(dns.CNAMERecordTypeCNAME),
				Name:    cloudflare.F(existingRecord.Name),
				Content: cloudflare.F[interface{}](content),
				TTL:     cloudflare.F(dns.TTL(ttl)),
				Proxied: cloudflare.F(proxied),
			}
		default:
			return fmt.Errorf("不支持的记录类型: %s", recordType)
		}

		// 调用 API 更新记录
		result, err := c.api.DNS.Records.Update(ctx, recordID, dns.RecordUpdateParams{
			ZoneID: cloudflare.F(zoneID),
			Record: recordParam,
		})
		if err != nil {
			return fmt.Errorf("更新 DNS 记录失败: %w", err)
		}

		// Content 是 interface{} 类型，需要转换为字符串
		contentStr := ""
		if result.Content != nil {
			contentStr = fmt.Sprintf("%v", result.Content)
		}

		recordInfo = &DNSRecordInfo{
			ID:         result.ID,
			Type:       string(result.Type),
			Name:       result.Name,
			Content:    contentStr,
			TTL:        float64(result.TTL),
			Proxied:    result.Proxied,
			Proxiable:  result.Proxiable,
			CreatedOn:  result.CreatedOn,
			ModifiedOn: result.ModifiedOn,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("成功更新 DNS 记录", "record_id", recordInfo.ID)
	return recordInfo, nil
}

// DeleteDNSRecord 删除 DNS 记录
func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	c.logger.Info("删除 DNS 记录", "zone_id", zoneID, "record_id", recordID)

	err := c.WithRetry(ctx, "删除 DNS 记录", func() error {
		_, err := c.api.DNS.Records.Delete(ctx, recordID, dns.RecordDeleteParams{
			ZoneID: cloudflare.F(zoneID),
		})
		if err != nil {
			return fmt.Errorf("删除 DNS 记录失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	c.logger.Info("成功删除 DNS 记录", "record_id", recordID)
	return nil
}

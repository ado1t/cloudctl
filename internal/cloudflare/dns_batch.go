package cloudflare

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// DNSBatchConfig 批量 DNS 操作配置
type DNSBatchConfig struct {
	Zones []DNSZoneConfig `yaml:"zones"`
}

// DNSZoneConfig Zone 配置
type DNSZoneConfig struct {
	Zone    string            `yaml:"zone"`
	Records []DNSRecordConfig `yaml:"records"`
}

// DNSRecordConfig DNS 记录配置
type DNSRecordConfig struct {
	Type    string  `yaml:"type"`
	Name    string  `yaml:"name"`
	Content string  `yaml:"content"`
	TTL     float64 `yaml:"ttl,omitempty"`
	Proxied bool    `yaml:"proxied,omitempty"`
}

// DNSBatchResult 批量操作结果
type DNSBatchResult struct {
	TotalZones     int
	TotalRecords   int
	SuccessZones   int
	SuccessRecords int
	FailedZones    int
	FailedRecords  int
	ZoneResults    []DNSZoneResult
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
}

// DNSZoneResult Zone 操作结果
type DNSZoneResult struct {
	Zone           string
	Success        bool
	TotalRecords   int
	SuccessRecords int
	FailedRecords  int
	RecordResults  []DNSRecordResult
	Error          error
}

// DNSRecordResult 记录操作结果
type DNSRecordResult struct {
	Type     string
	Name     string
	Content  string
	Success  bool
	Error    error
	RecordID string
}

// LoadDNSBatchConfig 从 YAML 文件加载批量配置
func LoadDNSBatchConfig(filename string) (*DNSBatchConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config DNSBatchConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// Validate 验证配置
func (c *DNSBatchConfig) Validate() error {
	if len(c.Zones) == 0 {
		return fmt.Errorf("至少需要配置一个 zone")
	}

	for i, zone := range c.Zones {
		if zone.Zone == "" {
			return fmt.Errorf("zone[%d]: zone 名称不能为空", i)
		}

		if len(zone.Records) == 0 {
			return fmt.Errorf("zone[%d] (%s): 至少需要配置一条记录", i, zone.Zone)
		}

		for j, record := range zone.Records {
			if record.Type == "" {
				return fmt.Errorf("zone[%d] (%s) record[%d]: type 不能为空", i, zone.Zone, j)
			}

			if record.Name == "" {
				return fmt.Errorf("zone[%d] (%s) record[%d]: name 不能为空", i, zone.Zone, j)
			}

			if record.Content == "" {
				return fmt.Errorf("zone[%d] (%s) record[%d]: content 不能为空", i, zone.Zone, j)
			}

			// 验证记录类型
			validTypes := map[string]bool{"A": true, "AAAA": true, "CNAME": true}
			if !validTypes[record.Type] {
				return fmt.Errorf("zone[%d] (%s) record[%d]: 不支持的记录类型 %s (支持: A, AAAA, CNAME)",
					i, zone.Zone, j, record.Type)
			}

			// 设置默认 TTL
			if record.TTL == 0 {
				record.TTL = 1 // auto
			}
		}
	}

	return nil
}

// BatchCreateDNSRecords 批量创建 DNS 记录
func (c *Client) BatchCreateDNSRecords(ctx context.Context, config *DNSBatchConfig, progressCallback func(string)) (*DNSBatchResult, error) {
	result := &DNSBatchResult{
		TotalZones:   len(config.Zones),
		TotalRecords: 0,
		StartTime:    time.Now(),
		ZoneResults:  make([]DNSZoneResult, 0, len(config.Zones)),
	}

	// 计算总记录数
	for _, zone := range config.Zones {
		result.TotalRecords += len(zone.Records)
	}

	c.logger.Info("开始批量创建 DNS 记录",
		"total_zones", result.TotalZones,
		"total_records", result.TotalRecords,
	)

	// 逐个处理 zone
	for zoneIdx, zoneConfig := range config.Zones {
		if progressCallback != nil {
			progressCallback(fmt.Sprintf("Processing zone %d/%d: %s",
				zoneIdx+1, result.TotalZones, zoneConfig.Zone))
		}

		zoneResult := c.processZone(ctx, zoneConfig, progressCallback)
		result.ZoneResults = append(result.ZoneResults, zoneResult)

		if zoneResult.Success {
			result.SuccessZones++
		} else {
			result.FailedZones++
		}

		result.SuccessRecords += zoneResult.SuccessRecords
		result.FailedRecords += zoneResult.FailedRecords
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	c.logger.Info("批量创建完成",
		"success_zones", result.SuccessZones,
		"failed_zones", result.FailedZones,
		"success_records", result.SuccessRecords,
		"failed_records", result.FailedRecords,
		"duration", result.Duration,
	)

	return result, nil
}

// processZone 处理单个 zone
func (c *Client) processZone(ctx context.Context, zoneConfig DNSZoneConfig, progressCallback func(string)) DNSZoneResult {
	result := DNSZoneResult{
		Zone:          zoneConfig.Zone,
		TotalRecords:  len(zoneConfig.Records),
		RecordResults: make([]DNSRecordResult, 0, len(zoneConfig.Records)),
	}

	c.logger.Info("处理 zone", "zone", zoneConfig.Zone, "records", result.TotalRecords)

	// 获取 Zone ID
	zone, err := c.GetZoneByName(ctx, zoneConfig.Zone)
	if err != nil {
		c.logger.Error("获取 zone 失败", "zone", zoneConfig.Zone, "error", err)
		result.Success = false
		result.Error = err
		result.FailedRecords = result.TotalRecords
		return result
	}

	// 处理每条记录
	for recordIdx, recordConfig := range zoneConfig.Records {
		if progressCallback != nil {
			progressCallback(fmt.Sprintf("  └─ %s: Creating record %d/%d (%s %s)",
				zoneConfig.Zone, recordIdx+1, result.TotalRecords,
				recordConfig.Type, recordConfig.Name))
		}

		recordResult := c.processRecord(ctx, zone.ID, recordConfig)
		result.RecordResults = append(result.RecordResults, recordResult)

		if recordResult.Success {
			result.SuccessRecords++
		} else {
			result.FailedRecords++
		}
	}

	result.Success = result.FailedRecords == 0

	return result
}

// processRecord 处理单条记录
func (c *Client) processRecord(ctx context.Context, zoneID string, recordConfig DNSRecordConfig) DNSRecordResult {
	result := DNSRecordResult{
		Type:    recordConfig.Type,
		Name:    recordConfig.Name,
		Content: recordConfig.Content,
	}

	c.logger.Debug("创建 DNS 记录",
		"zone_id", zoneID,
		"type", recordConfig.Type,
		"name", recordConfig.Name,
		"content", recordConfig.Content,
	)

	// 创建记录
	record, err := c.CreateDNSRecord(ctx, zoneID, DNSRecordCreateParams(recordConfig))

	if err != nil {
		c.logger.Error("创建 DNS 记录失败",
			"type", recordConfig.Type,
			"name", recordConfig.Name,
			"error", err,
		)
		result.Success = false
		result.Error = err
		return result
	}

	result.Success = true
	result.RecordID = record.ID

	c.logger.Debug("DNS 记录创建成功",
		"record_id", record.ID,
		"type", recordConfig.Type,
		"name", recordConfig.Name,
	)

	return result
}

// BatchCreateDNSRecordsConcurrent 并发批量创建 DNS 记录
func (c *Client) BatchCreateDNSRecordsConcurrent(ctx context.Context, config *DNSBatchConfig, maxConcurrency int, progressCallback func(string)) (*DNSBatchResult, error) {
	result := &DNSBatchResult{
		TotalZones:   len(config.Zones),
		TotalRecords: 0,
		StartTime:    time.Now(),
		ZoneResults:  make([]DNSZoneResult, len(config.Zones)),
	}

	// 计算总记录数
	for _, zone := range config.Zones {
		result.TotalRecords += len(zone.Records)
	}

	c.logger.Info("开始并发批量创建 DNS 记录",
		"total_zones", result.TotalZones,
		"total_records", result.TotalRecords,
		"max_concurrency", maxConcurrency,
	)

	// 使用 WaitGroup 和 channel 控制并发
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	// 并发处理 zones
	for i, zoneConfig := range config.Zones {
		wg.Add(1)
		go func(idx int, cfg DNSZoneConfig) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if progressCallback != nil {
				progressCallback(fmt.Sprintf("Processing zone %d/%d: %s",
					idx+1, result.TotalZones, cfg.Zone))
			}

			result.ZoneResults[idx] = c.processZone(ctx, cfg, progressCallback)
		}(i, zoneConfig)
	}

	wg.Wait()

	// 统计结果
	for _, zoneResult := range result.ZoneResults {
		if zoneResult.Success {
			result.SuccessZones++
		} else {
			result.FailedZones++
		}
		result.SuccessRecords += zoneResult.SuccessRecords
		result.FailedRecords += zoneResult.FailedRecords
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	c.logger.Info("并发批量创建完成",
		"success_zones", result.SuccessZones,
		"failed_zones", result.FailedZones,
		"success_records", result.SuccessRecords,
		"failed_records", result.FailedRecords,
		"duration", result.Duration,
	)

	return result, nil
}

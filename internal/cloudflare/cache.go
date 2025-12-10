package cloudflare

import (
	"context"
	"fmt"

	cloudflare "github.com/cloudflare/cloudflare-go/v2"
	"github.com/cloudflare/cloudflare-go/v2/cache"
)

// CachePurgeParams 缓存清除参数
type CachePurgeParams struct {
	// PurgeEverything 清除所有缓存
	PurgeEverything bool

	// Files 要清除的文件 URL 列表
	Files []string

	// Prefixes 要清除的 URL 前缀列表（目录）
	Prefixes []string

	// Tags 要清除的 Cache Tag 列表（企业版功能）
	Tags []string

	// Hosts 要清除的主机名列表
	Hosts []string
}

// CachePurgeResult 缓存清除结果
type CachePurgeResult struct {
	// Success 是否成功
	Success bool

	// ID 清除任务 ID
	ID string

	// Message 结果消息
	Message string

	// PurgeType 清除类型
	PurgeType string

	// ItemCount 清除的项目数量
	ItemCount int
}

// PurgeCache 清除缓存
func (c *Client) PurgeCache(ctx context.Context, zoneName string, params CachePurgeParams) (*CachePurgeResult, error) {
	c.logger.Debug("清除缓存", "zone", zoneName, "params", fmt.Sprintf("%+v", params))

	// 获取 Zone ID
	zone, err := c.GetZoneByName(ctx, zoneName)
	if err != nil {
		return nil, WrapError(err, "获取 Zone")
	}

	// 验证参数
	if err := validatePurgeParams(params); err != nil {
		return nil, NewValidationError("验证清除参数", err.Error())
	}

	// 构建清除请求
	var result *CachePurgeResult
	var apiErr error

	if params.PurgeEverything {
		// 清除所有缓存
		result, apiErr = c.purgeEverything(ctx, zone.ID)
	} else if len(params.Files) > 0 {
		// 按文件清除
		result, apiErr = c.purgeByFiles(ctx, zone.ID, params.Files)
	} else if len(params.Prefixes) > 0 {
		// 按前缀清除
		result, apiErr = c.purgeByPrefixes(ctx, zone.ID, params.Prefixes)
	} else if len(params.Tags) > 0 {
		// 按标签清除
		result, apiErr = c.purgeByTags(ctx, zone.ID, params.Tags)
	} else if len(params.Hosts) > 0 {
		// 按主机名清除
		result, apiErr = c.purgeByHosts(ctx, zone.ID, params.Hosts)
	}

	if apiErr != nil {
		return nil, WrapError(apiErr, "清除缓存")
	}

	c.logger.Info("缓存清除成功", "zone", zoneName, "type", result.PurgeType)
	return result, nil
}

// purgeEverything 清除所有缓存
func (c *Client) purgeEverything(ctx context.Context, zoneID string) (*CachePurgeResult, error) {
	c.logger.Debug("清除所有缓存", "zone_id", zoneID)

	resp, err := c.api.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(zoneID),
		Body: cache.CachePurgeParamsBodyCachePurgeEverything{
			PurgeEverything: cloudflare.F(true),
		},
	})

	if err != nil {
		return nil, err
	}

	return &CachePurgeResult{
		Success:   true,
		ID:        resp.ID,
		Message:   "已清除所有缓存",
		PurgeType: "everything",
		ItemCount: 0,
	}, nil
}

// purgeByFiles 按文件清除
func (c *Client) purgeByFiles(ctx context.Context, zoneID string, files []string) (*CachePurgeResult, error) {
	c.logger.Debug("按文件清除缓存", "zone_id", zoneID, "count", len(files))

	resp, err := c.api.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(zoneID),
		Body: cache.CachePurgeParamsBodyCachePurgeSingleFile{
			Files: cloudflare.F(files),
		},
	})

	if err != nil {
		return nil, err
	}

	return &CachePurgeResult{
		Success:   true,
		ID:        resp.ID,
		Message:   fmt.Sprintf("已清除 %d 个文件的缓存", len(files)),
		PurgeType: "files",
		ItemCount: len(files),
	}, nil
}

// purgeByPrefixes 按前缀清除
func (c *Client) purgeByPrefixes(ctx context.Context, zoneID string, prefixes []string) (*CachePurgeResult, error) {
	c.logger.Debug("按前缀清除缓存", "zone_id", zoneID, "count", len(prefixes))

	resp, err := c.api.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(zoneID),
		Body: cache.CachePurgeParamsBodyCachePurgeFlexPurgeByPrefixes{
			Prefixes: cloudflare.F(prefixes),
		},
	})

	if err != nil {
		return nil, err
	}

	return &CachePurgeResult{
		Success:   true,
		ID:        resp.ID,
		Message:   fmt.Sprintf("已清除 %d 个前缀的缓存", len(prefixes)),
		PurgeType: "prefixes",
		ItemCount: len(prefixes),
	}, nil
}

// purgeByTags 按标签清除
func (c *Client) purgeByTags(ctx context.Context, zoneID string, tags []string) (*CachePurgeResult, error) {
	c.logger.Debug("按标签清除缓存", "zone_id", zoneID, "count", len(tags))

	resp, err := c.api.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(zoneID),
		Body: cache.CachePurgeParamsBodyCachePurgeFlexPurgeByTags{
			Tags: cloudflare.F(tags),
		},
	})

	if err != nil {
		return nil, err
	}

	return &CachePurgeResult{
		Success:   true,
		ID:        resp.ID,
		Message:   fmt.Sprintf("已清除 %d 个标签的缓存", len(tags)),
		PurgeType: "tags",
		ItemCount: len(tags),
	}, nil
}

// purgeByHosts 按主机名清除
func (c *Client) purgeByHosts(ctx context.Context, zoneID string, hosts []string) (*CachePurgeResult, error) {
	c.logger.Debug("按主机名清除缓存", "zone_id", zoneID, "count", len(hosts))

	resp, err := c.api.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(zoneID),
		Body: cache.CachePurgeParamsBodyCachePurgeFlexPurgeByHostnames{
			Hosts: cloudflare.F(hosts),
		},
	})

	if err != nil {
		return nil, err
	}

	return &CachePurgeResult{
		Success:   true,
		ID:        resp.ID,
		Message:   fmt.Sprintf("已清除 %d 个主机名的缓存", len(hosts)),
		PurgeType: "hosts",
		ItemCount: len(hosts),
	}, nil
}

// validatePurgeParams 验证清除参数
func validatePurgeParams(params CachePurgeParams) error {
	// 统计指定的清除模式数量
	modeCount := 0
	if params.PurgeEverything {
		modeCount++
	}
	if len(params.Files) > 0 {
		modeCount++
	}
	if len(params.Prefixes) > 0 {
		modeCount++
	}
	if len(params.Tags) > 0 {
		modeCount++
	}
	if len(params.Hosts) > 0 {
		modeCount++
	}

	// 必须指定至少一种清除模式
	if modeCount == 0 {
		return fmt.Errorf("必须指定至少一种清除模式：--purge-all, --files, --prefixes, --tags 或 --hosts")
	}

	// 只能指定一种清除模式
	if modeCount > 1 {
		return fmt.Errorf("只能指定一种清除模式，不能同时使用多种模式")
	}

	// 验证文件列表
	if len(params.Files) > 30 {
		return fmt.Errorf("一次最多只能清除 30 个文件")
	}

	// 验证前缀列表
	if len(params.Prefixes) > 30 {
		return fmt.Errorf("一次最多只能清除 30 个前缀")
	}

	// 验证标签列表
	if len(params.Tags) > 30 {
		return fmt.Errorf("一次最多只能清除 30 个标签")
	}

	// 验证主机名列表
	if len(params.Hosts) > 30 {
		return fmt.Errorf("一次最多只能清除 30 个主机名")
	}

	return nil
}

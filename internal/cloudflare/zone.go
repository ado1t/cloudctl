package cloudflare

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go/v2"
	"github.com/cloudflare/cloudflare-go/v2/zones"
)

// ZoneInfo Zone 信息
type ZoneInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	NameServers []string  `json:"name_servers,omitempty"`
	CreatedOn   time.Time `json:"created_on"`
	ModifiedOn  time.Time `json:"modified_on"`
	ActivatedOn time.Time `json:"activated_on,omitempty"`
}

// ListZones 列出所有 Zone
func (c *Client) ListZones(ctx context.Context) ([]ZoneInfo, error) {
	var allZones []ZoneInfo

	c.logger.Debug("开始列出所有 Zone")

	err := c.WithRetry(ctx, "列出 Zone", func() error {
		// 使用分页获取所有 zone
		page := float64(1)
		perPage := float64(50)

		for {
			c.logger.Debug("获取 Zone 列表",
				"page", page,
				"per_page", perPage,
			)

			// 调用 Cloudflare API
			result, err := c.api.Zones.List(ctx, zones.ZoneListParams{
				Page:    cloudflare.F(page),
				PerPage: cloudflare.F(perPage),
			})

			if err != nil {
				return fmt.Errorf("获取 Zone 列表失败: %w", err)
			}

			// 转换为我们的数据结构
			for _, zone := range result.Result {
				// 获取状态 - 根据 ActivatedOn 判断
				status := "pending"
				if !zone.ActivatedOn.IsZero() {
					status = "active"
				}

				zoneInfo := ZoneInfo{
					ID:          zone.ID,
					Name:        zone.Name,
					Status:      status,
					CreatedOn:   zone.CreatedOn,
					ModifiedOn:  zone.ModifiedOn,
					ActivatedOn: zone.ActivatedOn,
				}

				// 添加 Name Servers
				if len(zone.NameServers) > 0 {
					zoneInfo.NameServers = zone.NameServers
				}

				allZones = append(allZones, zoneInfo)
			}

			c.logger.Debug("获取到 Zone",
				"count", len(result.Result),
				"total", len(allZones),
			)

			// 检查是否还有更多页
			// 如果返回的结果数少于每页数量，说明已经是最后一页
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

	c.logger.Info("成功列出所有 Zone", "total", len(allZones))
	return allZones, nil
}

// GetZone 获取指定 Zone 的信息
func (c *Client) GetZone(ctx context.Context, zoneID string) (*ZoneInfo, error) {
	var zoneInfo *ZoneInfo

	c.logger.Debug("获取 Zone 信息", "zone_id", zoneID)

	err := c.WithRetry(ctx, "获取 Zone", func() error {
		result, err := c.api.Zones.Get(ctx, zones.ZoneGetParams{
			ZoneID: cloudflare.F(zoneID),
		})
		if err != nil {
			return fmt.Errorf("获取 Zone 失败: %w", err)
		}

		// 获取状态 - 根据 ActivatedOn 判断
		status := "pending"
		if !result.ActivatedOn.IsZero() {
			status = "active"
		}

		zoneInfo = &ZoneInfo{
			ID:          result.ID,
			Name:        result.Name,
			Status:      status,
			CreatedOn:   result.CreatedOn,
			ModifiedOn:  result.ModifiedOn,
			ActivatedOn: result.ActivatedOn,
		}

		if len(result.NameServers) > 0 {
			zoneInfo.NameServers = result.NameServers
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("成功获取 Zone 信息", "zone_id", zoneID, "name", zoneInfo.Name)
	return zoneInfo, nil
}

// GetZoneByName 通过域名获取 Zone 信息
func (c *Client) GetZoneByName(ctx context.Context, name string) (*ZoneInfo, error) {
	c.logger.Debug("通过域名查找 Zone", "name", name)

	// 列出所有 zone 并查找匹配的
	zones, err := c.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	for _, zone := range zones {
		if zone.Name == name {
			c.logger.Info("找到匹配的 Zone", "name", name, "zone_id", zone.ID)
			return &zone, nil
		}
	}

	return nil, NewNotFoundError("查找 Zone", name)
}

// getAccountID 获取 Account ID
func (c *Client) getAccountID(ctx context.Context) (string, error) {
	// 尝试从现有 Zone 获取 Account ID
	c.logger.Debug("尝试从现有 Zone 获取 Account ID")

	result, err := c.api.Zones.List(ctx, zones.ZoneListParams{
		Page:    cloudflare.F(float64(1)),
		PerPage: cloudflare.F(float64(1)),
	})

	if err == nil && len(result.Result) > 0 {
		accountID := result.Result[0].Account.ID
		c.logger.Debug("从 Zone 获取到 Account ID", "account_id", accountID)
		return accountID, nil
	}

	// 如果没有 Zone，返回错误提示
	return "", fmt.Errorf("无法获取 Account ID，请先在 Cloudflare 控制台创建至少一个 Zone")
}

// CreateZone 创建新的 Zone
func (c *Client) CreateZone(ctx context.Context, name string) (*ZoneInfo, error) {
	var zoneInfo *ZoneInfo

	c.logger.Info("创建 Zone", "name", name)

	// 获取 Account ID
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 Account ID 失败: %w", err)
	}

	c.logger.Debug("使用 Account ID", "account_id", accountID)

	err = c.WithRetry(ctx, "创建 Zone", func() error {
		result, err := c.api.Zones.New(ctx, zones.ZoneNewParams{
			Name: cloudflare.F(name),
			Account: cloudflare.F(zones.ZoneNewParamsAccount{
				ID: cloudflare.F(accountID),
			}),
		})

		if err != nil {
			return fmt.Errorf("创建 Zone 失败: %w", err)
		}

		// 获取状态 - 根据 ActivatedOn 判断
		status := "pending"
		if !result.ActivatedOn.IsZero() {
			status = "active"
		}

		zoneInfo = &ZoneInfo{
			ID:          result.ID,
			Name:        result.Name,
			Status:      status,
			CreatedOn:   result.CreatedOn,
			ModifiedOn:  result.ModifiedOn,
			ActivatedOn: result.ActivatedOn,
		}

		if len(result.NameServers) > 0 {
			zoneInfo.NameServers = result.NameServers
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("成功创建 Zone",
		"zone_id", zoneInfo.ID,
		"name", zoneInfo.Name,
		"status", zoneInfo.Status,
	)

	return zoneInfo, nil
}

package adapter

import (
	"context"

	"github.com/StatusCakeDev/cakectl/internal/domain"
)

// ListUptimeLocations lists the locations uptime checks run from.
func (c *Client) ListUptimeLocations(ctx context.Context) ([]domain.Location, error) {
	res, err := c.client.ListUptimeMonitoringLocations(ctx).Execute()
	return mapAll(res.Data, toLocation), translate("list uptime locations", err)
}

// ListPagespeedLocations lists the locations pagespeed checks run from.
func (c *Client) ListPagespeedLocations(ctx context.Context) ([]domain.Location, error) {
	res, err := c.client.ListPagespeedMonitoringLocations(ctx).Execute()
	return mapAll(res.Data, toLocation), translate("list pagespeed locations", err)
}

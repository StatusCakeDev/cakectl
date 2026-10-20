package adapters

import (
	"context"

	"github.com/StatusCakeDev/statuscake-go"
)

// ListUptimeLocations lists the locations uptime checks run from.
func (c *Client) ListUptimeLocations(ctx context.Context) ([]statuscake.MonitoringLocation, error) {
	res, err := c.client.ListUptimeMonitoringLocations(ctx).Execute()
	return res.Data, translate("list uptime locations", err)
}

// ListPagespeedLocations lists the locations pagespeed checks run from.
func (c *Client) ListPagespeedLocations(ctx context.Context) ([]statuscake.MonitoringLocation, error) {
	res, err := c.client.ListPagespeedMonitoringLocations(ctx).Execute()
	return res.Data, translate("list pagespeed locations", err)
}

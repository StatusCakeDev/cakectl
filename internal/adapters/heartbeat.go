package adapters

import (
	"context"
	"fmt"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// HeartbeatFilter narrows a listing of heartbeat checks. The zero value
// matches every check.
type HeartbeatFilter = UptimeFilter

// ListHeartbeat lists every heartbeat check matching the filter.
func (c *Client) ListHeartbeat(ctx context.Context, filter HeartbeatFilter) ([]statuscake.HeartbeatTestOverview, error) {
	items, err := collect(func(page, limit int32) ([]statuscake.HeartbeatTestOverview, statuscake.Pagination, error) {
		req := c.client.ListHeartbeatTests(ctx).Page(page).Limit(limit)
		if filter.Status != "" {
			req = req.Status(filter.Status)
		}
		if len(filter.Tags) > 0 {
			req = req.Tags(csv(filter.Tags)).Matchany(filter.MatchAny)
		}
		res, err := req.Execute()
		return res.Data, res.Metadata, err
	})
	return items, translate("list heartbeat checks", err)
}

// GetHeartbeat reads a heartbeat check.
func (c *Client) GetHeartbeat(ctx context.Context, id string) (statuscake.HeartbeatTest, error) {
	res, err := c.client.GetHeartbeatTest(ctx, id).Execute()
	return res.Data, translate("get heartbeat check "+id, err)
}

// CreateHeartbeat creates a heartbeat check and returns its ID.
func (c *Client) CreateHeartbeat(ctx context.Context, s spec.Heartbeat) (string, error) {
	if err := s.ValidateCreate(); err != nil {
		return "", fmt.Errorf("invalid heartbeat check: %w", err)
	}

	r := c.client.CreateHeartbeatTest(ctx)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.Period, func(v int32) { r = r.Period(v) })
	apply(s.ContactGroups, func(v []string) { r = r.ContactGroups(v) })
	apply(s.Host, func(v string) { r = r.Host(v) })
	apply(s.Paused, func(v bool) { r = r.Paused(v) })
	apply(s.Tags, func(v []string) { r = r.Tags(v) })

	res, err := r.Execute()
	return res.Data.NewID, translate("create heartbeat check", err)
}

// UpdateHeartbeat changes the fields the spec sets on a heartbeat check.
func (c *Client) UpdateHeartbeat(ctx context.Context, id string, s spec.Heartbeat) error {
	if spec.IsEmpty(s) {
		return ErrNothingToUpdate
	}
	if err := s.ValidateUpdate(); err != nil {
		return fmt.Errorf("invalid heartbeat check: %w", err)
	}

	r := c.client.UpdateHeartbeatTest(ctx, id)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.Period, func(v int32) { r = r.Period(v) })
	apply(s.ContactGroups, func(v []string) { r = r.ContactGroups(v) })
	apply(s.Host, func(v string) { r = r.Host(v) })
	apply(s.Paused, func(v bool) { r = r.Paused(v) })
	apply(s.Tags, func(v []string) { r = r.Tags(v) })

	return translate("update heartbeat check "+id, r.Execute())
}

// DeleteHeartbeat deletes a heartbeat check.
func (c *Client) DeleteHeartbeat(ctx context.Context, id string) error {
	return translate("delete heartbeat check "+id, c.client.DeleteHeartbeatTest(ctx, id).Execute())
}

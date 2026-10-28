package adapter

import (
	"context"
	"fmt"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/domain"
	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// HeartbeatFilter narrows a listing of heartbeat checks. The zero value
// matches every check.
type HeartbeatFilter = UptimeFilter

// ListHeartbeat lists every heartbeat check matching the filter.
func (c *Client) ListHeartbeat(ctx context.Context, filter HeartbeatFilter) ([]domain.HeartbeatOverview, error) {
	items, err := collect(func(page, limit int32) ([]statuscake.HeartbeatTestOverview, statuscake.Pagination, error) {
		req := c.client.ListHeartbeatTests(ctx).Page(page).Limit(limit)
		if filter.Status != "" {
			req = req.Status(filter.Status)
		}
		if len(filter.Tags) > 0 {
			req = req.Tags(csv(filter.Tags)).Matchany(filter.MatchAny)
		}

		// The uptime percentage costs the API work to compute for every check, so
		// it is only requested when the caller wants it. The SDK adds the nouptime
		// parameter whenever it is set, whatever its value, so it must only be set
		// to leave the percentage out.
		if !filter.Uptime {
			req = req.Nouptime(true)
		}

		res, err := req.Execute()
		return res.Data, res.Metadata, err
	})
	if err != nil {
		return nil, translate("list heartbeat checks", err)
	}
	return mapAll(items, toHeartbeatOverview), nil
}

// GetHeartbeat reads a heartbeat check.
func (c *Client) GetHeartbeat(ctx context.Context, id string) (domain.Heartbeat, error) {
	res, err := c.client.GetHeartbeatTest(ctx, id).Execute()
	return toHeartbeat(res.Data), translate("get heartbeat check "+id, err)
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
		return domain.ErrNothingToUpdate
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

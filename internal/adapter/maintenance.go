package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/domain"
	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// ListMaintenance lists every maintenance window, or only those in the given
// state ("active", "paused" or "pending") when one is supplied.
func (c *Client) ListMaintenance(ctx context.Context, state string) ([]domain.MaintenanceWindow, error) {
	items, err := collect(func(page, limit int32) ([]statuscake.MaintenanceWindow, statuscake.Pagination, error) {
		req := c.client.ListMaintenanceWindows(ctx).Page(page).Limit(limit)
		if state != "" {
			req = req.State(state)
		}
		res, err := req.Execute()
		return res.Data, res.Metadata, err
	})
	if err != nil {
		return nil, translate("list maintenance windows", err)
	}
	return mapAll(items, toMaintenanceWindow), nil
}

// GetMaintenance reads a maintenance window.
func (c *Client) GetMaintenance(ctx context.Context, id string) (domain.MaintenanceWindow, error) {
	res, err := c.client.GetMaintenanceWindow(ctx, id).Execute()
	return toMaintenanceWindow(res.Data), translate("get maintenance window "+id, err)
}

// CreateMaintenance creates a maintenance window and returns its ID.
func (c *Client) CreateMaintenance(ctx context.Context, s spec.Maintenance) (string, error) {
	if err := s.ValidateCreate(); err != nil {
		return "", fmt.Errorf("invalid maintenance window: %w", err)
	}

	r := c.client.CreateMaintenanceWindow(ctx)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.StartAt, func(v time.Time) { r = r.Start(v) })
	apply(s.EndAt, func(v time.Time) { r = r.End(v) })
	apply(s.Timezone, func(v string) { r = r.Timezone(v) })
	apply(s.RepeatInterval, func(v string) { r = r.RepeatInterval(statuscake.MaintenanceWindowRepeatInterval(v)) })
	apply(s.Tags, func(v []string) { r = r.Tags(v) })
	apply(s.Tests, func(v []string) { r = r.Tests(v) })

	res, err := r.Execute()
	return res.Data.NewID, translate("create maintenance window", err)
}

// UpdateMaintenance changes the fields the spec sets on a maintenance window.
func (c *Client) UpdateMaintenance(ctx context.Context, id string, s spec.Maintenance) error {
	if spec.IsEmpty(s) {
		return domain.ErrNothingToUpdate
	}
	if err := s.ValidateUpdate(); err != nil {
		return fmt.Errorf("invalid maintenance window: %w", err)
	}

	r := c.client.UpdateMaintenanceWindow(ctx, id)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.StartAt, func(v time.Time) { r = r.Start(v) })
	apply(s.EndAt, func(v time.Time) { r = r.End(v) })
	apply(s.Timezone, func(v string) { r = r.Timezone(v) })
	apply(s.RepeatInterval, func(v string) { r = r.RepeatInterval(statuscake.MaintenanceWindowRepeatInterval(v)) })
	apply(s.Tags, func(v []string) { r = r.Tags(v) })
	apply(s.Tests, func(v []string) { r = r.Tests(v) })

	return translate("update maintenance window "+id, r.Execute())
}

// DeleteMaintenance deletes a maintenance window.
func (c *Client) DeleteMaintenance(ctx context.Context, id string) error {
	return translate("delete maintenance window "+id, c.client.DeleteMaintenanceWindow(ctx, id).Execute())
}

package adapters

import (
	"context"
	"fmt"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// ListPagespeed lists every pagespeed check.
func (c *Client) ListPagespeed(ctx context.Context) ([]statuscake.PagespeedTest, error) {
	items, err := collect(func(page, limit int32) ([]statuscake.PagespeedTest, statuscake.Pagination, error) {
		res, err := c.client.ListPagespeedTests(ctx).Page(page).Limit(limit).Execute()
		return res.Data, res.Metadata, err
	})
	return items, translate("list pagespeed checks", err)
}

// GetPagespeed reads a pagespeed check.
func (c *Client) GetPagespeed(ctx context.Context, id string) (statuscake.PagespeedTest, error) {
	res, err := c.client.GetPagespeedTest(ctx, id).Execute()
	return res.Data, translate("get pagespeed check "+id, err)
}

// CreatePagespeed creates a pagespeed check and returns its ID.
func (c *Client) CreatePagespeed(ctx context.Context, s spec.Pagespeed) (string, error) {
	if err := s.ValidateCreate(); err != nil {
		return "", fmt.Errorf("invalid pagespeed check: %w", err)
	}

	r := c.client.CreatePagespeedTest(ctx)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.WebsiteURL, func(v string) { r = r.WebsiteURL(v) })
	apply(s.CheckRate, func(v int32) { r = r.CheckRate(statuscake.PagespeedTestCheckRate(v)) })
	apply(s.Region, func(v string) { r = r.Region(statuscake.PagespeedTestRegion(v)) })
	apply(s.AlertBigger, func(v int32) { r = r.AlertBigger(v) })
	apply(s.AlertSlower, func(v int64) { r = r.AlertSlower(v) })
	apply(s.AlertSmaller, func(v int32) { r = r.AlertSmaller(v) })
	apply(s.ContactGroups, func(v []string) { r = r.ContactGroups(v) })
	apply(s.Paused, func(v bool) { r = r.Paused(v) })

	res, err := r.Execute()
	return res.Data.NewID, translate("create pagespeed check", err)
}

// UpdatePagespeed changes the fields the spec sets on a pagespeed check.
func (c *Client) UpdatePagespeed(ctx context.Context, id string, s spec.Pagespeed) error {
	if spec.IsEmpty(s) {
		return ErrNothingToUpdate
	}
	if err := s.ValidateUpdate(); err != nil {
		return fmt.Errorf("invalid pagespeed check: %w", err)
	}

	r := c.client.UpdatePagespeedTest(ctx, id)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.WebsiteURL, func(v string) { r = r.WebsiteURL(v) })
	apply(s.CheckRate, func(v int32) { r = r.CheckRate(statuscake.PagespeedTestCheckRate(v)) })
	apply(s.Region, func(v string) { r = r.Region(statuscake.PagespeedTestRegion(v)) })
	apply(s.AlertBigger, func(v int32) { r = r.AlertBigger(v) })
	apply(s.AlertSlower, func(v int64) { r = r.AlertSlower(v) })
	apply(s.AlertSmaller, func(v int32) { r = r.AlertSmaller(v) })
	apply(s.ContactGroups, func(v []string) { r = r.ContactGroups(v) })
	apply(s.Paused, func(v bool) { r = r.Paused(v) })

	return translate("update pagespeed check "+id, r.Execute())
}

// DeletePagespeed deletes a pagespeed check.
func (c *Client) DeletePagespeed(ctx context.Context, id string) error {
	return translate("delete pagespeed check "+id, c.client.DeletePagespeedTest(ctx, id).Execute())
}

// ListPagespeedHistory lists the tests a pagespeed check has run.
func (c *Client) ListPagespeedHistory(ctx context.Context, id string, w Window) ([]statuscake.PagespeedTestHistoryResult, error) {
	req := c.client.ListPagespeedTestHistory(ctx, id)
	if w.Limit > 0 {
		req = req.Limit(int32(w.Limit))
	}
	instant(w.Before, func(v int64) { req = req.Before(v) })
	instant(w.After, func(v int64) { req = req.After(v) })

	res, err := req.Execute()
	return res.Data, translate("list history of pagespeed check "+id, err)
}

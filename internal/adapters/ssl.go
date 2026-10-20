package adapters

import (
	"context"
	"fmt"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// ListSSL lists every SSL check.
func (c *Client) ListSSL(ctx context.Context) ([]statuscake.SSLTest, error) {
	items, err := collect(func(page, limit int32) ([]statuscake.SSLTest, statuscake.Pagination, error) {
		res, err := c.client.ListSslTests(ctx).Page(page).Limit(limit).Execute()
		return res.Data, res.Metadata, err
	})
	return items, translate("list SSL checks", err)
}

// GetSSL reads an SSL check.
func (c *Client) GetSSL(ctx context.Context, id string) (statuscake.SSLTest, error) {
	res, err := c.client.GetSslTest(ctx, id).Execute()
	return res.Data, translate("get SSL check "+id, err)
}

// CreateSSL creates an SSL check and returns its ID.
func (c *Client) CreateSSL(ctx context.Context, s spec.SSL) (string, error) {
	if err := s.ValidateCreate(); err != nil {
		return "", fmt.Errorf("invalid SSL check: %w", err)
	}

	r := c.client.CreateSslTest(ctx)
	apply(s.WebsiteURL, func(v string) { r = r.WebsiteURL(v) })
	apply(s.CheckRate, func(v int32) { r = r.CheckRate(statuscake.SSLTestCheckRate(v)) })
	apply(s.AlertAt, func(v []int32) { r = r.AlertAt(v) })
	apply(s.AlertBroken, func(v bool) { r = r.AlertBroken(v) })
	apply(s.AlertExpiry, func(v bool) { r = r.AlertExpiry(v) })
	apply(s.AlertMixed, func(v bool) { r = r.AlertMixed(v) })
	apply(s.AlertReminder, func(v bool) { r = r.AlertReminder(v) })
	apply(s.ContactGroups, func(v []string) { r = r.ContactGroups(v) })
	apply(s.FollowRedirects, func(v bool) { r = r.FollowRedirects(v) })
	apply(s.Hostname, func(v string) { r = r.Hostname(v) })
	apply(s.Paused, func(v bool) { r = r.Paused(v) })
	apply(s.UserAgent, func(v string) { r = r.UserAgent(v) })

	res, err := r.Execute()
	return res.Data.NewID, translate("create SSL check", err)
}

// UpdateSSL changes the fields the spec sets on an SSL check.
func (c *Client) UpdateSSL(ctx context.Context, id string, s spec.SSL) error {
	if spec.IsEmpty(s) {
		return ErrNothingToUpdate
	}
	if err := s.ValidateUpdate(); err != nil {
		return fmt.Errorf("invalid SSL check: %w", err)
	}

	r := c.client.UpdateSslTest(ctx, id)
	apply(s.WebsiteURL, func(v string) { r = r.WebsiteURL(v) })
	apply(s.CheckRate, func(v int32) { r = r.CheckRate(statuscake.SSLTestCheckRate(v)) })
	apply(s.AlertAt, func(v []int32) { r = r.AlertAt(v) })
	apply(s.AlertBroken, func(v bool) { r = r.AlertBroken(v) })
	apply(s.AlertExpiry, func(v bool) { r = r.AlertExpiry(v) })
	apply(s.AlertMixed, func(v bool) { r = r.AlertMixed(v) })
	apply(s.AlertReminder, func(v bool) { r = r.AlertReminder(v) })
	apply(s.ContactGroups, func(v []string) { r = r.ContactGroups(v) })
	apply(s.FollowRedirects, func(v bool) { r = r.FollowRedirects(v) })
	apply(s.Hostname, func(v string) { r = r.Hostname(v) })
	apply(s.Paused, func(v bool) { r = r.Paused(v) })
	apply(s.UserAgent, func(v string) { r = r.UserAgent(v) })

	return translate("update SSL check "+id, r.Execute())
}

// DeleteSSL deletes an SSL check.
func (c *Client) DeleteSSL(ctx context.Context, id string) error {
	return translate("delete SSL check "+id, c.client.DeleteSslTest(ctx, id).Execute())
}

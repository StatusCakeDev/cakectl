package adapter

import (
	"context"
	"fmt"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/domain"
	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// UptimeFilter narrows a listing of uptime checks. The zero value matches
// every check.
type UptimeFilter struct {
	// Status is "up" or "down".
	Status string
	// Tags restricts the listing to checks carrying these tags.
	Tags []string
	// MatchAny matches checks carrying any of the tags rather than all of them.
	MatchAny bool
	// Uptime asks the API to work out each check's uptime percentage. That is
	// expensive for the API, so it is left out unless it is asked for.
	Uptime bool
}

// ListUptime lists every uptime check matching the filter.
func (c *Client) ListUptime(ctx context.Context, filter UptimeFilter) ([]domain.UptimeOverview, error) {
	items, err := collect(func(page, limit int32) ([]statuscake.UptimeTestOverview, statuscake.Pagination, error) {
		req := c.client.ListUptimeTests(ctx).Page(page).Limit(limit)
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
		return nil, translate("list uptime checks", err)
	}
	return mapAll(items, toUptimeOverview), nil
}

// GetUptime reads an uptime check.
func (c *Client) GetUptime(ctx context.Context, id string) (domain.Uptime, error) {
	res, err := c.client.GetUptimeTest(ctx, id).Execute()
	return toUptime(res.Data), translate("get uptime check "+id, err)
}

// CreateUptime creates an uptime check and returns its ID.
func (c *Client) CreateUptime(ctx context.Context, s spec.Uptime) (string, error) {
	if err := s.ValidateCreate(); err != nil {
		return "", fmt.Errorf("invalid uptime check: %w", err)
	}

	r := c.client.CreateUptimeTest(ctx)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.TestType, func(v string) { r = r.TestType(statuscake.UptimeTestType(v)) })
	apply(s.WebsiteURL, func(v string) { r = r.WebsiteURL(v) })
	apply(s.CheckRate, func(v int32) { r = r.CheckRate(statuscake.UptimeTestCheckRate(v)) })
	apply(s.BasicUsername, func(v string) { r = r.BasicUsername(v) })
	apply(s.BasicPassword, func(v string) { r = r.BasicPassword(v) })
	apply(s.Confirmation, func(v int32) { r = r.Confirmation(v) })
	apply(s.ContactGroups, func(v []string) { r = r.ContactGroups(v) })
	apply(s.CustomHeader, func(v string) { r = r.CustomHeader(v) })
	apply(s.DNSIPs, func(v []string) { r = r.DNSIPs(v) })
	apply(s.DNSServer, func(v string) { r = r.DNSServer(v) })
	apply(s.DoNotFind, func(v bool) { r = r.DoNotFind(v) })
	apply(s.EnableSSLAlert, func(v bool) { r = r.EnableSSLAlert(v) })
	apply(s.FinalEndpoint, func(v string) { r = r.FinalEndpoint(v) })
	apply(s.FindString, func(v string) { r = r.FindString(v) })
	apply(s.FollowRedirects, func(v bool) { r = r.FollowRedirects(v) })
	apply(s.Host, func(v string) { r = r.Host(v) })
	apply(s.IncludeHeader, func(v bool) { r = r.IncludeHeader(v) })
	apply(s.Paused, func(v bool) { r = r.Paused(v) })
	apply(s.Port, func(v int32) { r = r.Port(v) })
	apply(s.PostBody, func(v string) { r = r.PostBody(v) })
	apply(s.PostRaw, func(v string) { r = r.PostRaw(v) })
	apply(s.Regions, func(v []string) { r = r.Regions(v) })
	apply(s.StatusCodes, func(v []string) { r = r.StatusCodes(v) })
	apply(s.Tags, func(v []string) { r = r.Tags(v) })
	apply(s.Timeout, func(v int32) { r = r.Timeout(v) })
	apply(s.TriggerRate, func(v int32) { r = r.TriggerRate(v) })
	apply(s.UseJAR, func(v bool) { r = r.UseJAR(v) })
	apply(s.UserAgent, func(v string) { r = r.UserAgent(v) })

	res, err := r.Execute()
	return res.Data.NewID, translate("create uptime check", err)
}

// UpdateUptime changes the fields the spec sets on an uptime check and leaves
// the rest as they were.
func (c *Client) UpdateUptime(ctx context.Context, id string, s spec.Uptime) error {
	if spec.IsEmpty(s) {
		return domain.ErrNothingToUpdate
	}
	if err := s.ValidateUpdate(); err != nil {
		return fmt.Errorf("invalid uptime check: %w", err)
	}

	r := c.client.UpdateUptimeTest(ctx, id)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.WebsiteURL, func(v string) { r = r.WebsiteURL(v) })
	apply(s.CheckRate, func(v int32) { r = r.CheckRate(statuscake.UptimeTestCheckRate(v)) })
	apply(s.BasicUsername, func(v string) { r = r.BasicUsername(v) })
	apply(s.BasicPassword, func(v string) { r = r.BasicPassword(v) })
	apply(s.Confirmation, func(v int32) { r = r.Confirmation(v) })
	apply(s.ContactGroups, func(v []string) { r = r.ContactGroups(v) })
	apply(s.CustomHeader, func(v string) { r = r.CustomHeader(v) })
	apply(s.DNSIPs, func(v []string) { r = r.DNSIPs(v) })
	apply(s.DNSServer, func(v string) { r = r.DNSServer(v) })
	apply(s.DoNotFind, func(v bool) { r = r.DoNotFind(v) })
	apply(s.EnableSSLAlert, func(v bool) { r = r.EnableSSLAlert(v) })
	apply(s.FinalEndpoint, func(v string) { r = r.FinalEndpoint(v) })
	apply(s.FindString, func(v string) { r = r.FindString(v) })
	apply(s.FollowRedirects, func(v bool) { r = r.FollowRedirects(v) })
	apply(s.Host, func(v string) { r = r.Host(v) })
	apply(s.IncludeHeader, func(v bool) { r = r.IncludeHeader(v) })
	apply(s.Paused, func(v bool) { r = r.Paused(v) })
	apply(s.Port, func(v int32) { r = r.Port(v) })
	apply(s.PostBody, func(v string) { r = r.PostBody(v) })
	apply(s.PostRaw, func(v string) { r = r.PostRaw(v) })
	apply(s.Regions, func(v []string) { r = r.Regions(v) })
	apply(s.StatusCodes, func(v []string) { r = r.StatusCodes(v) })
	apply(s.Tags, func(v []string) { r = r.Tags(v) })
	apply(s.Timeout, func(v int32) { r = r.Timeout(v) })
	apply(s.TriggerRate, func(v int32) { r = r.TriggerRate(v) })
	apply(s.UseJAR, func(v bool) { r = r.UseJAR(v) })
	apply(s.UserAgent, func(v string) { r = r.UserAgent(v) })

	return translate("update uptime check "+id, r.Execute())
}

// DeleteUptime deletes an uptime check.
func (c *Client) DeleteUptime(ctx context.Context, id string) error {
	return translate("delete uptime check "+id, c.client.DeleteUptimeTest(ctx, id).Execute())
}

// ListUptimeAlerts lists the alerts an uptime check has raised.
func (c *Client) ListUptimeAlerts(ctx context.Context, id string, w Window) ([]domain.UptimeAlert, error) {
	req := c.client.ListUptimeTestAlerts(ctx, id)
	if w.Limit > 0 {
		req = req.Limit(int32(w.Limit))
	}
	instant(w.Before, func(v int64) { req = req.Before(v) })
	instant(w.After, func(v int64) { req = req.After(v) })

	res, err := req.Execute()
	return mapAll(res.Data, toUptimeAlert), translate("list alerts of uptime check "+id, err)
}

// ListUptimeHistory lists the tests an uptime check has run.
func (c *Client) ListUptimeHistory(ctx context.Context, id string, w Window) ([]domain.UptimeResult, error) {
	req := c.client.ListUptimeTestHistory(ctx, id)
	if w.Limit > 0 {
		req = req.Limit(int32(w.Limit))
	}
	instant(w.Before, func(v int64) { req = req.Before(v) })
	instant(w.After, func(v int64) { req = req.After(v) })

	res, err := req.Execute()
	return mapAll(res.Data, toUptimeResult), translate("list history of uptime check "+id, err)
}

// ListUptimePeriods lists the periods of up and down time of an uptime check.
func (c *Client) ListUptimePeriods(ctx context.Context, id string, w Window) ([]domain.UptimePeriod, error) {
	req := c.client.ListUptimeTestPeriods(ctx, id)
	if w.Limit > 0 {
		req = req.Limit(int32(w.Limit))
	}
	instant(w.Before, func(v int64) { req = req.Before(v) })
	instant(w.After, func(v int64) { req = req.After(v) })

	res, err := req.Execute()
	return mapAll(res.Data, toUptimePeriod), translate("list periods of uptime check "+id, err)
}

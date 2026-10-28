package adapter_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/adapter"
	"github.com/StatusCakeDev/cakectl/internal/domain"
)

func TestGet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("uptime", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/uptime/{id}", http.StatusOK, map[string]any{"data": uptimeFixture("7")})

		got, err := c.GetUptime(ctx, "7")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if got.ID != "7" || got.Name != "Home page" || string(got.TestType) != "HTTP" || int32(got.CheckRate) != 300 {
			t.Errorf("got %+v, want the fixture's check", got)
		}
		if api.only(t).Path != "/v1/uptime/7" {
			t.Errorf("requested %q", api.only(t).Path)
		}
	})

	t.Run("heartbeat", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/heartbeat/{id}", http.StatusOK, map[string]any{"data": heartbeatFixture("7")})

		got, err := c.GetHeartbeat(ctx, "7")
		if err != nil || got.ID != "7" || got.Period != 600 {
			t.Errorf("GetHeartbeat() = %+v, %v; want the fixture", got, err)
		}
	})

	t.Run("pagespeed", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/pagespeed/{id}", http.StatusOK, map[string]any{"data": pagespeedFixture("7")})

		got, err := c.GetPagespeed(ctx, "7")
		if err != nil || got.ID != "7" || got.Location != "UK" {
			t.Errorf("GetPagespeed() = %+v, %v; want the fixture", got, err)
		}
	})

	t.Run("ssl", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/ssl/{id}", http.StatusOK, map[string]any{"data": sslFixture("7")})

		got, err := c.GetSSL(ctx, "7")
		if err != nil || got.ID != "7" || len(got.AlertAt) != 2 {
			t.Errorf("GetSSL() = %+v, %v; want the fixture", got, err)
		}
	})

	t.Run("contact group", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/contact-groups/{id}", http.StatusOK, map[string]any{"data": contactGroupFixture("7")})

		got, err := c.GetContactGroup(ctx, "7")
		if err != nil || got.ID != "7" || len(got.EmailAddresses) != 1 {
			t.Errorf("GetContactGroup() = %+v, %v; want the fixture", got, err)
		}
	})

	t.Run("maintenance window", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/maintenance-windows/{id}", http.StatusOK, map[string]any{"data": maintenanceFixture("7")})

		got, err := c.GetMaintenance(ctx, "7")
		if err != nil || got.ID != "7" || !got.Start.Equal(maintenanceStart) {
			t.Errorf("GetMaintenance() = %+v, %v; want the fixture", got, err)
		}
	})
}

func TestGetTranslatesNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, c := newAPI(t) // Every route is a 404.

	gets := map[string]func() error{
		"uptime":      func() error { _, err := c.GetUptime(ctx, "x"); return err },
		"heartbeat":   func() error { _, err := c.GetHeartbeat(ctx, "x"); return err },
		"pagespeed":   func() error { _, err := c.GetPagespeed(ctx, "x"); return err },
		"ssl":         func() error { _, err := c.GetSSL(ctx, "x"); return err },
		"contact":     func() error { _, err := c.GetContactGroup(ctx, "x"); return err },
		"maintenance": func() error { _, err := c.GetMaintenance(ctx, "x"); return err },
	}
	for name, get := range gets {
		if err := get(); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: error = %v, want %v", name, err, domain.ErrNotFound)
		}
	}
}

func TestListReadsEveryPage(t *testing.T) {
	t.Parallel()

	api, c := newAPI(t)
	api.serveFunc("GET /v1/ssl", func(r *http.Request) (int, any) {
		switch r.URL.Query().Get("page") {
		case "1":
			return http.StatusOK, page(1, 2, sslFixture("1"), sslFixture("2"))
		case "2":
			return http.StatusOK, page(2, 2, sslFixture("3"))
		default:
			return http.StatusBadRequest, map[string]any{"message": "unexpected page"}
		}
	})

	got, err := c.ListSSL(context.Background())
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	var ids []string
	for _, s := range got {
		ids = append(ids, s.ID)
	}
	if diff := cmp.Diff([]string{"1", "2", "3"}, ids); diff != "" {
		t.Errorf("IDs mismatch (-want +got):\n%s", diff)
	}

	var queries []url.Values
	for _, call := range api.calls() {
		queries = append(queries, call.Query)
	}
	want := []url.Values{
		{"page": {"1"}, "limit": {"100"}},
		{"page": {"2"}, "limit": {"100"}},
	}
	if diff := cmp.Diff(want, queries); diff != "" {
		t.Errorf("queries mismatch (-want +got):\n%s", diff)
	}
}

func TestListOfNothing(t *testing.T) {
	t.Parallel()

	api, c := newAPI(t)
	api.serve("GET /v1/contact-groups", http.StatusOK, page(1, 1))

	got, err := c.ListContactGroups(context.Background())
	if err != nil || len(got) != 0 {
		t.Errorf("ListContactGroups() = %v, %v; want nothing", got, err)
	}
}

func TestListRefusesAnUnboundedListing(t *testing.T) {
	t.Parallel()

	api, c := newAPI(t)
	api.serve("GET /v1/pagespeed", http.StatusOK, page(1, 1000, pagespeedFixture("1")))

	_, err := c.ListPagespeed(context.Background())
	if !errors.Is(err, domain.ErrTruncated) {
		t.Errorf("error = %v, want %v", err, domain.ErrTruncated)
	}
	if n := len(api.calls()); n != 50 {
		t.Errorf("made %d requests, want it to stop at 50", n)
	}
}

func TestListFailureStopsPaging(t *testing.T) {
	t.Parallel()

	api, c := newAPI(t)
	api.serveFunc("GET /v1/heartbeat", func(r *http.Request) (int, any) {
		if r.URL.Query().Get("page") == "1" {
			return http.StatusOK, page(1, 3, heartbeatFixture("1"))
		}
		return http.StatusInternalServerError, map[string]any{"message": "boom"}
	})

	got, err := c.ListHeartbeat(context.Background(), adapter.HeartbeatFilter{})
	if err == nil || got != nil {
		t.Errorf("ListHeartbeat() = %v, %v; want an error and no partial results", got, err)
	}
}

func TestListFilters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("uptime", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/uptime", http.StatusOK, page(1, 1, uptimeOverviewFixture("1")))

		got, err := c.ListUptime(ctx, adapter.UptimeFilter{Status: "down", Tags: []string{"a", "b"}, MatchAny: true})
		if err != nil || len(got) != 1 {
			t.Fatalf("ListUptime() = %v, %v", got, err)
		}

		want := url.Values{"page": {"1"}, "limit": {"100"}, "status": {"down"}, "tags": {"a,b"}, "matchany": {"true"}, "nouptime": {"true"}}
		if diff := cmp.Diff(want, api.only(t).Query); diff != "" {
			t.Errorf("query mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("uptime without a filter sends none", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/uptime", http.StatusOK, page(1, 1))

		if _, err := c.ListUptime(ctx, adapter.UptimeFilter{}); err != nil {
			t.Fatalf("error = %v", err)
		}

		want := url.Values{"page": {"1"}, "limit": {"100"}, "nouptime": {"true"}}
		if diff := cmp.Diff(want, api.only(t).Query); diff != "" {
			t.Errorf("query mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("uptime percentage is only requested when asked for", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/uptime", http.StatusOK, page(1, 1))
		api.serve("GET /v1/heartbeat", http.StatusOK, page(1, 1))

		if _, err := c.ListUptime(ctx, adapter.UptimeFilter{Uptime: true}); err != nil {
			t.Fatalf("ListUptime() error = %v", err)
		}
		if _, err := c.ListHeartbeat(ctx, adapter.HeartbeatFilter{Uptime: true}); err != nil {
			t.Fatalf("ListHeartbeat() error = %v", err)
		}

		want := url.Values{"page": {"1"}, "limit": {"100"}}
		for _, call := range api.calls() {
			if diff := cmp.Diff(want, call.Query); diff != "" {
				t.Errorf("%s query mismatch (-want +got):\n%s", call.Path, diff)
			}
		}
	})

	t.Run("heartbeat", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/heartbeat", http.StatusOK, page(1, 1, heartbeatOverviewFixture("1")))

		if _, err := c.ListHeartbeat(ctx, adapter.HeartbeatFilter{Status: "up", Tags: []string{"x"}}); err != nil {
			t.Fatalf("error = %v", err)
		}

		want := url.Values{"page": {"1"}, "limit": {"100"}, "status": {"up"}, "tags": {"x"}, "matchany": {"false"}, "nouptime": {"true"}}
		if diff := cmp.Diff(want, api.only(t).Query); diff != "" {
			t.Errorf("query mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("maintenance", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/maintenance-windows", http.StatusOK, page(1, 1, maintenanceFixture("1")))

		got, err := c.ListMaintenance(ctx, "active")
		if err != nil || len(got) != 1 {
			t.Fatalf("ListMaintenance() = %v, %v", got, err)
		}

		want := url.Values{"page": {"1"}, "limit": {"100"}, "state": {"active"}}
		if diff := cmp.Diff(want, api.only(t).Query); diff != "" {
			t.Errorf("query mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestHistoryWindows(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	before := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	after := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	window := adapter.Window{Limit: 5, Before: gofp.Some(before), After: gofp.Some(after)}
	wantQuery := url.Values{
		"limit":  {"5"},
		"before": {"1790899200"},
		"after":  {"1790812800"},
	}

	empty := map[string]any{"data": []any{}, "links": map[string]any{"self": "x"}}

	tests := map[string]struct {
		path string
		call func(*adapter.Client) (int, error)
	}{
		"uptime alerts": {"/v1/uptime/7/alerts", func(c *adapter.Client) (int, error) {
			got, err := c.ListUptimeAlerts(ctx, "7", window)
			return len(got), err
		}},
		"uptime history": {"/v1/uptime/7/history", func(c *adapter.Client) (int, error) {
			got, err := c.ListUptimeHistory(ctx, "7", window)
			return len(got), err
		}},
		"uptime periods": {"/v1/uptime/7/periods", func(c *adapter.Client) (int, error) {
			got, err := c.ListUptimePeriods(ctx, "7", window)
			return len(got), err
		}},
		"pagespeed history": {"/v1/pagespeed/7/history", func(c *adapter.Client) (int, error) {
			got, err := c.ListPagespeedHistory(ctx, "7", window)
			return len(got), err
		}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			api, c := newAPI(t)
			api.serve("GET "+tt.path, http.StatusOK, empty)

			if _, err := tt.call(c); err != nil {
				t.Fatalf("error = %v", err)
			}
			if diff := cmp.Diff(wantQuery, api.only(t).Query); diff != "" {
				t.Errorf("query mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHistoryWithoutAWindowSendsNoBounds(t *testing.T) {
	t.Parallel()

	api, c := newAPI(t)
	api.serve("GET /v1/uptime/7/history", http.StatusOK, map[string]any{
		"data":  []any{map[string]any{"created_at": "2026-10-01T00:00:00Z", "status_code": 200, "performance": 120, "location": "UK"}},
		"links": map[string]any{"self": "x"},
	})

	got, err := c.ListUptimeHistory(context.Background(), "7", adapter.Window{})
	if err != nil || len(got) != 1 || got[0].StatusCode == nil || *got[0].StatusCode != 200 {
		t.Fatalf("ListUptimeHistory() = %+v, %v", got, err)
	}
	if q := api.only(t).Query; len(q) != 0 {
		t.Errorf("query = %v, want none", q)
	}
}

func TestHistoryOfAMissingCheckIsNotFound(t *testing.T) {
	t.Parallel()

	_, c := newAPI(t)

	_, err := c.ListUptimePeriods(context.Background(), "x", adapter.Window{})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestLocations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	body := map[string]any{"data": []any{map[string]any{
		"description": "London", "region": "United Kingdom", "region_code": "UK", "status": "up", "ipv4": "192.0.2.1",
	}}}

	t.Run("uptime", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/uptime-locations", http.StatusOK, body)

		got, err := c.ListUptimeLocations(ctx)
		if err != nil || len(got) != 1 || got[0].RegionCode != "UK" {
			t.Errorf("ListUptimeLocations() = %+v, %v", got, err)
		}
	})

	t.Run("pagespeed", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/pagespeed-locations", http.StatusOK, body)

		got, err := c.ListPagespeedLocations(ctx)
		if err != nil || len(got) != 1 || got[0].Description != "London" {
			t.Errorf("ListPagespeedLocations() = %+v, %v", got, err)
		}
	})
}

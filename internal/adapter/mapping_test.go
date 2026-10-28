package adapter_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/StatusCakeDev/cakectl/internal/adapter"
	"github.com/StatusCakeDev/cakectl/internal/domain"
)

// These checks pin the boundary between the SDK and the rest of cakectl: every
// field the API reports arrives in the domain model, and nothing is lost on
// the way. The documents are shaped like the API's JSON, so that decoding is
// exercised along with the mapping.

func ptr[T any](v T) *T { return &v }

func at(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func checkMapping[T any](t *testing.T, got, want T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestDetailsAreMappedIntoDomainModels(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	t.Run("uptime", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/uptime/{id}", http.StatusOK, map[string]any{"data": map[string]any{
			"id": "7", "name": "Home page", "test_type": "HTTP", "website_url": "https://example.com",
			"check_rate": 300, "confirmation": 2, "contact_groups": []string{"1", "2"},
			"custom_header": `{"X":"y"}`, "dns_ips": []string{"1.1.1.1"}, "dns_server": "dns.example",
			"do_not_find": true, "enable_ssl_alert": true, "final_endpoint": "https://example.com/final",
			"find_string": "ok", "follow_redirects": true, "include_header": true, "host": "origin",
			"last_tested_at": "2026-10-01T10:00:00Z", "next_location": "UK", "paused": true, "port": 8080,
			"post_body": "a=b", "post_raw": "{}", "processing": true, "servers": []any{},
			"status": "up", "status_codes": []string{"500", "503"}, "tags": []string{"prod"},
			"timeout": 15, "trigger_rate": 2, "uptime": 99.5, "use_jar": true, "user_agent": "cakectl",
		}})

		got, err := c.GetUptime(ctx, "7")
		checkMapping(t, got, domain.Uptime{
			ID: "7", Name: "Home page", TestType: "HTTP", WebsiteURL: "https://example.com",
			CheckRate: 300, Confirmation: 2, ContactGroups: []string{"1", "2"},
			CustomHeader: ptr(`{"X":"y"}`), DNSIPs: []string{"1.1.1.1"}, DNSServer: ptr("dns.example"),
			DoNotFind: true, EnableSSLAlert: true, FinalEndpoint: ptr("https://example.com/final"),
			FindString: ptr("ok"), FollowRedirects: true, IncludeHeader: true, Host: ptr("origin"),
			LastTested: ptr(at("2026-10-01T10:00:00Z")), NextLocation: ptr("UK"), Paused: true,
			Port: ptr(int32(8080)), PostBody: ptr("a=b"), PostRaw: ptr("{}"), Processing: true,
			Status: "up", StatusCodes: []string{"500", "503"}, Tags: []string{"prod"}, Timeout: 15,
			TriggerRate: 2, Uptime: 99.5, UseJAR: true, UserAgent: ptr("cakectl"),
		}, err)
	})

	t.Run("heartbeat", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/heartbeat/{id}", http.StatusOK, map[string]any{"data": map[string]any{
			"id": "8", "name": "Nightly job", "url": "https://heartbeat.example/abc", "period": 600,
			"contact_groups": []string{"1"}, "host": "worker-1", "last_tested_at": "2026-10-01T10:00:00Z",
			"paused": true, "status": "up", "tags": []string{"batch"}, "uptime": 100,
		}})

		got, err := c.GetHeartbeat(ctx, "8")
		checkMapping(t, got, domain.Heartbeat{
			ID: "8", Name: "Nightly job", WebsiteURL: "https://heartbeat.example/abc", Period: 600,
			ContactGroups: []string{"1"}, Host: ptr("worker-1"), LastTested: ptr(at("2026-10-01T10:00:00Z")),
			Paused: true, Status: "up", Tags: []string{"batch"}, Uptime: 100,
		}, err)
	})

	t.Run("pagespeed", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/pagespeed/{id}", http.StatusOK, map[string]any{"data": map[string]any{
			"id": "9", "name": "Landing", "website_url": "https://example.com", "check_rate": 600,
			"alert_bigger": 100, "alert_slower": 4000, "alert_smaller": 3, "contact_groups": []string{"1"},
			"latest_stats": map[string]any{
				"filesize": 12.5, "has_issue": true, "loadtime": 900, "requests": 40, "latest_issue": "slow",
			},
			"location": "UK", "paused": true,
		}})

		got, err := c.GetPagespeed(ctx, "9")
		checkMapping(t, got, domain.Pagespeed{
			ID: "9", Name: "Landing", WebsiteURL: "https://example.com", CheckRate: 600,
			AlertBigger: 100, AlertSlower: 4000, AlertSmaller: 3, ContactGroups: []string{"1"},
			LatestStats: &domain.PagespeedStats{
				Filesize: 12.5, HasIssue: true, LatestIssue: ptr("slow"), Loadtime: 900, Requests: 40,
			},
			Location: "UK", Paused: true,
		}, err)
	})

	t.Run("ssl", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/ssl/{id}", http.StatusOK, map[string]any{"data": map[string]any{
			"id": "10", "website_url": "https://example.com", "check_rate": 3600, "alert_at": []int{7, 30},
			"alert_broken": true, "alert_expiry": true, "alert_mixed": true, "alert_reminder": true,
			"certificate_score": 90, "certificate_status": "CERT_OK", "cipher": "TLS_AES_256_GCM_SHA384",
			"cipher_score": 95, "contact_groups": []string{"1"}, "issuer_common_name": "R3",
			"follow_redirects": true, "hostname": "example.com",
			"mixed_content": []map[string]any{{"src": "http://x", "type": "img"}, {"src": "http://y", "type": "js"}},
			"paused":        true, "updated_at": "2026-10-02T00:00:00Z", "user_agent": "ua",
			"valid_from": "2026-09-01T00:00:00Z", "valid_until": "2026-12-01T00:00:00Z",
		}})

		got, err := c.GetSSL(ctx, "10")
		checkMapping(t, got, domain.SSL{
			ID: "10", WebsiteURL: "https://example.com", CheckRate: 3600, AlertAt: []int32{7, 30},
			AlertBroken: true, AlertExpiry: true, AlertMixed: true, AlertReminder: true,
			CertificateScore: ptr(int32(90)), CertificateStatus: ptr("CERT_OK"),
			Cipher: ptr("TLS_AES_256_GCM_SHA384"), CipherScore: ptr(int32(95)), ContactGroups: []string{"1"},
			IssuerCommonName: ptr("R3"), FollowRedirects: true, Hostname: ptr("example.com"),
			MixedContent: 2, Paused: true, Updated: ptr(at("2026-10-02T00:00:00Z")), UserAgent: ptr("ua"),
			ValidFrom: ptr(at("2026-09-01T00:00:00Z")), ValidUntil: ptr(at("2026-12-01T00:00:00Z")),
		}, err)
	})

	t.Run("contact group", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/contact-groups/{id}", http.StatusOK, map[string]any{"data": map[string]any{
			"id": "11", "name": "Ops", "email_addresses": []string{"ops@example.com"},
			"integrations": []string{"slack"}, "mobile_numbers": []string{"+447700900000"},
			"ping_url": "https://hooks.example/x",
		}})

		got, err := c.GetContactGroup(ctx, "11")
		checkMapping(t, got, domain.ContactGroup{
			ID: "11", Name: "Ops", EmailAddresses: []string{"ops@example.com"},
			Integrations: []string{"slack"}, MobileNumbers: []string{"+447700900000"},
			PingURL: ptr("https://hooks.example/x"),
		}, err)
	})

	t.Run("maintenance window", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/maintenance-windows/{id}", http.StatusOK, map[string]any{"data": map[string]any{
			"id": "12", "name": "Deploy", "start_at": "2026-11-01T22:00:00Z", "end_at": "2026-11-02T01:00:00Z",
			"repeat_interval": "1w", "state": "pending", "tags": []string{"release"},
			"tests": []string{"7", "8"}, "timezone": "Europe/London",
		}})

		got, err := c.GetMaintenance(ctx, "12")
		checkMapping(t, got, domain.MaintenanceWindow{
			ID: "12", Name: "Deploy", Start: at("2026-11-01T22:00:00Z"), End: at("2026-11-02T01:00:00Z"),
			RepeatInterval: "1w", State: "pending", Tags: []string{"release"}, Tests: []string{"7", "8"},
			Timezone: "Europe/London",
		}, err)
	})
}

func TestListingsAreMappedIntoDomainModels(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	t.Run("uptime", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/uptime", http.StatusOK, page(1, 1, map[string]any{
			"id": "1", "name": "Home", "test_type": "HTTP", "website_url": "https://a.example",
			"check_rate": 300, "contact_groups": []string{"2"}, "paused": true, "status": "down",
			"tags": []string{"prod"}, "uptime": 99.5,
		}))

		got, err := c.ListUptime(ctx, adapter.UptimeFilter{})
		checkMapping(t, got, []domain.UptimeOverview{{
			ID: "1", Name: "Home", TestType: "HTTP", WebsiteURL: "https://a.example", CheckRate: 300,
			ContactGroups: []string{"2"}, Paused: true, Status: "down", Tags: []string{"prod"}, Uptime: ptr(float32(99.5)),
		}}, err)
	})

	t.Run("heartbeat", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		api.serve("GET /v1/heartbeat", http.StatusOK, page(1, 1, map[string]any{
			"id": "2", "name": "Job", "url": "https://heartbeat.example/x", "period": 600,
			"contact_groups": []string{"2"}, "paused": true, "status": "up", "tags": []string{"b"}, "uptime": 100,
		}))

		got, err := c.ListHeartbeat(ctx, adapter.HeartbeatFilter{})
		checkMapping(t, got, []domain.HeartbeatOverview{{
			ID: "2", Name: "Job", Period: 600, ContactGroups: []string{"2"}, Paused: true,
			Status: "up", Tags: []string{"b"}, Uptime: ptr(float32(100)),
		}}, err)
	})

	t.Run("history", func(t *testing.T) {
		t.Parallel()

		api, c := newAPI(t)
		links := map[string]any{"self": "x"}
		api.serve("GET /v1/uptime/{id}/alerts", http.StatusOK, map[string]any{"links": links, "data": []any{
			map[string]any{"id": "a1", "status": "down", "status_code": 503, "triggered_at": "2026-10-01T10:00:00Z"},
		}})
		api.serve("GET /v1/uptime/{id}/history", http.StatusOK, map[string]any{"links": links, "data": []any{
			map[string]any{"created_at": "2026-10-01T10:00:00Z", "location": "UK", "status_code": 200, "performance": 120},
		}})
		api.serve("GET /v1/uptime/{id}/periods", http.StatusOK, map[string]any{"links": links, "data": []any{
			map[string]any{"created_at": "2026-10-01T10:00:00Z", "ended_at": "2026-10-01T11:00:00Z", "status": "up", "duration": 3600000},
		}})
		api.serve("GET /v1/pagespeed/{id}/history", http.StatusOK, map[string]any{"links": links, "data": []any{
			map[string]any{"created_at": "2026-10-01T10:00:00Z", "filesize": 12.25, "har_location": "x", "loadtime": 900, "requests": 40, "throttling": "4G"},
		}})
		api.serve("GET /v1/uptime-locations", http.StatusOK, map[string]any{"data": []any{
			map[string]any{"description": "London", "region": "United Kingdom", "region_code": "UK", "status": "up", "ipv4": "192.0.2.1", "ipv6": "2001:db8::1"},
		}})

		alerts, err := c.ListUptimeAlerts(ctx, "7", adapter.Window{})
		checkMapping(t, alerts, []domain.UptimeAlert{{ID: "a1", Status: "down", StatusCode: 503, Triggered: ptr(at("2026-10-01T10:00:00Z"))}}, err)

		results, err := c.ListUptimeHistory(ctx, "7", adapter.Window{})
		checkMapping(t, results, []domain.UptimeResult{{
			Created: at("2026-10-01T10:00:00Z"), Location: ptr("UK"), StatusCode: ptr(int32(200)), Performance: ptr(int64(120)),
		}}, err)

		periods, err := c.ListUptimePeriods(ctx, "7", adapter.Window{})
		checkMapping(t, periods, []domain.UptimePeriod{{
			Created: at("2026-10-01T10:00:00Z"), Ended: ptr(at("2026-10-01T11:00:00Z")), Status: "up", Duration: ptr(int64(3600000)),
		}}, err)

		speed, err := c.ListPagespeedHistory(ctx, "7", adapter.Window{})
		checkMapping(t, speed, []domain.PagespeedResult{{
			Created: at("2026-10-01T10:00:00Z"), Filesize: 12.25, HARLocation: "x", Loadtime: 900, Requests: 40, Throttling: "4G",
		}}, err)

		locations, err := c.ListUptimeLocations(ctx)
		checkMapping(t, locations, []domain.Location{{
			Description: "London", Region: "United Kingdom", RegionCode: "UK", Status: "up", IPv4: ptr("192.0.2.1"), IPv6: ptr("2001:db8::1"),
		}}, err)
	})
}

package view_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/spf13/pflag"

	cliflag "github.com/tomasbasham/cli-runtime/flag"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/spec"
	"github.com/StatusCakeDev/cakectl/internal/view"
)

// Models are decoded from JSON shaped like the API's, so that the SDK's own
// decoding is exercised along with the view.
func model[T any](t *testing.T, doc string) T {
	t.Helper()

	var v T
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return v
}

var allFormats = []cliflag.Format{cliflag.FormatText, cliflag.FormatJSON, cliflag.FormatPrettyJSON, cliflag.FormatYAML}

func render(t *testing.T, format cliflag.Format, v any) string {
	t.Helper()

	flags := cliflag.NewPrinterFlags(cliflag.FormatTextFlag|cliflag.FormatJSONFlag|cliflag.FormatPrettyJSONFlag|cliflag.FormatYAMLFlag, cliflag.FormatText)
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.AddFlags(fs)
	if err := fs.Parse([]string{"--format", string(format)}); err != nil {
		t.Fatal(err)
	}
	p, err := flags.ToPrinter()
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := p.Print(&buf, v); err != nil {
		t.Fatalf("Print(%s) error = %v", format, err)
	}
	return buf.String()
}

const (
	uptimeDoc = `{
  "id": "7", "name": "Home page", "test_type": "HTTP", "website_url": "https://example.com",
  "check_rate": 300, "confirmation": 2, "contact_groups": ["1", "2"], "custom_header": "{\"X\":\"y\"}",
  "dns_ips": [], "do_not_find": false, "enable_ssl_alert": true, "find_string": "ok",
  "follow_redirects": true, "include_header": false, "host": "origin", "last_tested_at": "2026-10-01T10:00:00Z",
  "paused": false, "port": 8080, "post_body": "a=b", "processing": false, "servers": [],
  "status": "up", "status_codes": ["500", "503"], "tags": ["prod"], "timeout": 15, "trigger_rate": 2,
  "uptime": 99.5, "use_jar": false, "user_agent": "cakectl"
}`

	heartbeatDoc = `{
  "id": "8", "name": "Nightly job", "url": "https://heartbeat.example/SECRET-TOKEN", "period": 600,
  "contact_groups": [], "host": "worker-1", "paused": true, "status": "up", "tags": ["batch"], "uptime": 100
}`

	pagespeedDoc = `{
  "id": "9", "name": "Landing", "website_url": "https://example.com", "check_rate": 600,
  "alert_bigger": 100, "alert_slower": 4000, "alert_smaller": 3, "contact_groups": ["1"],
  "latest_stats": {"filesize": 12.5, "has_issue": true, "loadtime": 900, "requests": 40, "latest_issue": "slow"},
  "location": "UK", "paused": false
}`

	sslDoc = `{
  "id": "10", "website_url": "https://example.com", "check_rate": 3600, "alert_at": [7, 30],
  "alert_broken": true, "alert_expiry": true, "alert_mixed": false, "alert_reminder": true,
  "certificate_status": "CERT_OK", "cipher": "TLS_AES_256_GCM_SHA384", "contact_groups": ["1"],
  "issuer_common_name": "R3", "follow_redirects": true, "hostname": "example.com",
  "mixed_content": [{"src": "http://x", "type": "img"}], "paused": false, "user_agent": "ua",
  "valid_from": "2026-09-01T00:00:00Z", "valid_until": "2026-12-01T00:00:00Z"
}`

	contactGroupDoc = `{
  "id": "11", "name": "Ops", "email_addresses": ["ops@example.com", "dev@example.com"],
  "integrations": ["slack"], "mobile_numbers": ["+447700900000"], "ping_url": "https://hooks.example/SECRET"
}`

	maintenanceDoc = `{
  "id": "12", "name": "Deploy", "start_at": "2026-11-01T22:00:00Z", "end_at": "2026-11-02T01:00:00Z",
  "repeat_interval": "1w", "state": "pending", "tags": ["release"], "tests": ["7", "8"], "timezone": "Europe/London"
}`
)

// TestDescriptionsRoundTripIntoSpecs is the property that makes `describe`
// useful as an editing starting point: what is printed can be read back as the
// same spec, in every structured format.
func TestDescriptionsRoundTripIntoSpecs(t *testing.T) {
	t.Parallel()

	uptime := view.NewUptimeDescription(model[statuscake.UptimeTest](t, uptimeDoc))
	heartbeat := view.NewHeartbeatDescription(model[statuscake.HeartbeatTest](t, heartbeatDoc))
	pagespeed := view.NewPagespeedDescription(model[statuscake.PagespeedTest](t, pagespeedDoc))
	ssl := view.NewSSLDescription(model[statuscake.SSLTest](t, sslDoc))
	contact := view.NewContactGroupDescription(model[statuscake.ContactGroup](t, contactGroupDoc))
	maintenance := view.NewMaintenanceDescription(model[statuscake.MaintenanceWindow](t, maintenanceDoc))

	for _, format := range []cliflag.Format{cliflag.FormatJSON, cliflag.FormatPrettyJSON, cliflag.FormatYAML} {
		t.Run(string(format), func(t *testing.T) {
			t.Parallel()

			roundTrip(t, format, uptime, uptime.Spec)
			roundTrip(t, format, heartbeat, heartbeat.Spec)
			roundTrip(t, format, pagespeed, pagespeed.Spec)
			roundTrip(t, format, ssl, ssl.Spec)
			roundTrip(t, format, contact, contact.Spec)
			roundTrip(t, format, maintenance, maintenance.Spec)
		})
	}
}

func roundTrip[S any](t *testing.T, format cliflag.Format, description any, want S) {
	t.Helper()

	got, err := spec.Decode[S](strings.NewReader(render(t, format, description)))
	if err != nil {
		t.Errorf("%T: reading back %s: %v", want, format, err)
		return
	}

	opts := []cmp.Option{cmp.Exporter(func(typ reflect.Type) bool { return strings.HasPrefix(typ.PkgPath(), "github.com/tomasbasham/gofp") })}
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("%T: %s round trip mismatch (-want +got):\n%s", want, format, diff)
	}
}

func TestDescriptionsCarryTheDocumentedSpecFields(t *testing.T) {
	t.Parallel()

	d := view.NewUptimeDescription(model[statuscake.UptimeTest](t, uptimeDoc))

	if v, _ := d.Spec.CheckRate.TryUnwrap(); v != 300 {
		t.Errorf("check_rate = %v, want 300", d.Spec.CheckRate)
	}
	if v, _ := d.Spec.Port.TryUnwrap(); v != 8080 {
		t.Errorf("port = %v, want 8080", d.Spec.Port)
	}
	// Fields the API returns as null are absent, not set to empty values.
	if d.Spec.DNSServer.IsSome() || d.Spec.FinalEndpoint.IsSome() {
		t.Errorf("absent optional fields were set: %+v", d.Spec)
	}
	// Credentials are write-only.
	if d.Spec.BasicPassword.IsSome() || d.Spec.BasicUsername.IsSome() {
		t.Error("credentials appeared in a description")
	}
}

func TestListsLeaveOutSecretsAndPersonalData(t *testing.T) {
	t.Parallel()

	heartbeats := view.NewHeartbeatList([]statuscake.HeartbeatTestOverview{
		model[statuscake.HeartbeatTestOverview](t, heartbeatDoc),
	})
	contacts := view.NewContactGroupList([]statuscake.ContactGroup{
		model[statuscake.ContactGroup](t, contactGroupDoc),
	})

	tests := map[string]struct {
		list    any
		secrets []string
	}{
		"heartbeats":     {heartbeats, []string{"SECRET-TOKEN", "heartbeat.example"}},
		"contact groups": {contacts, []string{"SECRET", "ops@example.com", "+447700900000", "hooks.example"}},
	}
	for name, tt := range tests {
		for _, format := range allFormats {
			out := render(t, format, tt.list)
			for _, secret := range tt.secrets {
				if strings.Contains(out, secret) {
					t.Errorf("%s as %s contains %q:\n%s", name, format, secret, out)
				}
			}
		}
	}
}

func TestDescriptionsShowWhatTheyAreAskedFor(t *testing.T) {
	t.Parallel()

	// `describe` is the explicit request for the detail, so it shows the
	// heartbeat URL a job needs to ping.
	out := render(t, cliflag.FormatText, view.NewHeartbeatDescription(model[statuscake.HeartbeatTest](t, heartbeatDoc)))
	if !strings.Contains(out, "https://heartbeat.example/SECRET-TOKEN") {
		t.Errorf("heartbeat description does not show its ping URL:\n%s", out)
	}
}

func TestEmptyListsAreEmptyNotNull(t *testing.T) {
	t.Parallel()

	lists := map[string]any{
		"uptime":      view.NewUptimeList(nil),
		"heartbeat":   view.NewHeartbeatList(nil),
		"pagespeed":   view.NewPagespeedList(nil),
		"ssl":         view.NewSSLList(nil),
		"contacts":    view.NewContactGroupList(nil),
		"maintenance": view.NewMaintenanceList(nil),
		"alerts":      view.NewAlertList(nil),
		"history":     view.NewHistoryList(nil),
		"periods":     view.NewPeriodList(nil),
		"results":     view.NewPagespeedResultList(nil),
		"locations":   view.NewLocationList(nil),
	}
	for name, l := range lists {
		if got := render(t, cliflag.FormatPrettyJSON, l); got != "[]\n" {
			t.Errorf("%s as json = %q, want %q", name, got, "[]\n")
		}
		if got := render(t, cliflag.FormatText, l); !strings.HasPrefix(got, "No ") || !strings.HasSuffix(got, " found.\n") {
			t.Errorf("%s as text = %q, want a 'No … found.' message", name, got)
		}
	}
}

func TestUptimeListText(t *testing.T) {
	t.Parallel()

	list := view.NewUptimeList([]statuscake.UptimeTestOverview{
		model[statuscake.UptimeTestOverview](t, `{"id": "1","name":"Home","test_type":"HTTP","website_url":"https://a.example","check_rate":300,"contact_groups":[],"paused":false,"status":"up","tags":[]}`),
		model[statuscake.UptimeTestOverview](t, `{"id": "22","name":"Longer name","test_type":"PING","website_url":"b.example","check_rate":0,"contact_groups":[],"paused":true,"status":"down","tags":[]}`),
	})

	want := `ID   NAME          TYPE   URL                 INTERVAL     STATUS
1    Home          HTTP   https://a.example   5m0s         up
22   Longer name   PING   b.example           continuous   paused
`
	if diff := cmp.Diff(want, render(t, cliflag.FormatText, list)); diff != "" {
		t.Errorf("text mismatch (-want +got):\n%s", diff)
	}
}

func TestUptimeDescriptionText(t *testing.T) {
	t.Parallel()

	out := render(t, cliflag.FormatText, view.NewUptimeDescription(model[statuscake.UptimeTest](t, uptimeDoc)))

	for _, want := range []string{
		"ID:", "7",
		"Name:", "Home page",
		"Interval:", "5m0s",
		"Status:", "up",
		"Uptime:", "99.50%",
		"Last tested:", "2026-10-01T10:00:00Z",
		"Status codes:", "500, 503",
		"DNS server:", // Absent, shown as a dash.
	} {
		if !strings.Contains(out, want) {
			t.Errorf("description does not contain %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "DNS server:") || !strings.Contains(out, "-\n") {
		t.Errorf("absent values are not shown as a dash:\n%s", out)
	}
}

func TestReadOnlyViewsText(t *testing.T) {
	t.Parallel()

	alerts := view.NewAlertList([]statuscake.UptimeTestAlert{
		model[statuscake.UptimeTestAlert](t, `{"id": "a1","status":"down","status_code":503,"triggered_at":"2026-10-01T10:00:00Z"}`),
	})
	history := view.NewHistoryList([]statuscake.UptimeTestHistoryResult{
		model[statuscake.UptimeTestHistoryResult](t, `{"created_at":"2026-10-01T10:00:00Z","location":"UK","status_code":200,"performance":120}`),
		model[statuscake.UptimeTestHistoryResult](t, `{"created_at":"2026-10-01T10:05:00Z"}`),
	})
	periods := view.NewPeriodList([]statuscake.UptimeTestPeriod{
		model[statuscake.UptimeTestPeriod](t, `{"created_at":"2026-10-01T10:00:00Z","ended_at":"2026-10-01T11:00:00Z","status":"up","duration":3600000}`),
	})
	results := view.NewPagespeedResultList([]statuscake.PagespeedTestHistoryResult{
		model[statuscake.PagespeedTestHistoryResult](t, `{"created_at":"2026-10-01T10:00:00Z","filesize":12.25,"har_location":"x","loadtime":900,"requests":40,"throttling":"4G"}`),
	})
	locations := view.NewLocationList([]statuscake.MonitoringLocation{
		model[statuscake.MonitoringLocation](t, `{"description":"London","region":"United Kingdom","region_code":"UK","status":"up","ipv4":"192.0.2.1"}`),
	})

	tests := map[string]struct {
		got  any
		want []string
	}{
		"alerts":    {alerts, []string{"a1", "2026-10-01T10:00:00Z", "down", "503"}},
		"history":   {history, []string{"UK", "200", "120 ms", "2026-10-01T10:05:00Z"}},
		"periods":   {periods, []string{"1h0m0s", "up"}},
		"results":   {results, []string{"900 ms", "40", "12.2 kB", "4G"}},
		"locations": {locations, []string{"UK", "United Kingdom", "London", "192.0.2.1"}},
	}
	for name, tt := range tests {
		out := render(t, cliflag.FormatText, tt.got)
		for _, want := range tt.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s does not contain %q:\n%s", name, want, out)
			}
		}
	}
}

func TestOutcome(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		got  view.Outcome
		text string
		json string
	}{
		"created": {view.Created("uptime", "7"), "uptime 7 created\n", `{
  "resource": "uptime",
  "id": "7",
  "action": "created"
}` + "\n"},
		"updated": {view.Updated("ssl", "8"), "ssl 8 updated\n", `{
  "resource": "ssl",
  "id": "8",
  "action": "updated"
}` + "\n"},
		"deleted": {view.Deleted("contact-group", "9"), "contact-group 9 deleted\n", `{
  "resource": "contact-group",
  "id": "9",
  "action": "deleted"
}` + "\n"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := render(t, cliflag.FormatText, tt.got); got != tt.text {
				t.Errorf("text = %q, want %q", got, tt.text)
			}
			if got := render(t, cliflag.FormatPrettyJSON, tt.got); got != tt.json {
				t.Errorf("json = %q, want %q", got, tt.json)
			}
		})
	}
}

package view_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/spf13/pflag"

	cliflag "github.com/tomasbasham/cli-runtime/flag"

	"github.com/StatusCakeDev/cakectl/internal/domain"
	"github.com/StatusCakeDev/cakectl/internal/spec"
	"github.com/StatusCakeDev/cakectl/internal/view"
)

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

// TestDescriptionsRoundTripIntoSpecs is the property that makes `describe`
// useful as an editing starting point: what is printed can be read back as the
// same spec, in every structured format.
func TestDescriptionsRoundTripIntoSpecs(t *testing.T) {
	t.Parallel()

	uptime := view.NewUptimeDescription(uptime())
	heartbeat := view.NewHeartbeatDescription(heartbeat())
	pagespeed := view.NewPagespeedDescription(pagespeed())
	ssl := view.NewSSLDescription(ssl())
	contact := view.NewContactGroupDescription(contactGroup())
	maintenance := view.NewMaintenanceDescription(maintenance())

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

	d := view.NewUptimeDescription(uptime())

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

	heartbeats := view.NewHeartbeatList([]domain.HeartbeatOverview{
		{ID: "8", Name: "Nightly job", Period: 600, Paused: true, Status: "up", Tags: []string{"batch"}, Uptime: ptr(float32(100))},
	})
	contacts := view.NewContactGroupList([]domain.ContactGroup{contactGroup()})

	tests := map[string]struct {
		list    any
		secrets []string
	}{
		"heartbeats": {
			list:    heartbeats,
			secrets: []string{"SECRET-TOKEN", "heartbeat.example"},
		},
		"contact groups": {
			list:    contacts,
			secrets: []string{"SECRET", "ops@example.com", "+447700900000", "hooks.example"},
		},
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
	out := render(t, cliflag.FormatText, view.NewHeartbeatDescription(heartbeat()))
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

	list := view.NewUptimeList([]domain.UptimeOverview{
		{ID: "1", Name: "Home", TestType: "HTTP", WebsiteURL: "https://a.example", CheckRate: 300, Status: "up"},
		{ID: "22", Name: "Longer name", TestType: "PING", WebsiteURL: "b.example", Paused: true, Status: "down"},
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

	out := render(t, cliflag.FormatText, view.NewUptimeDescription(uptime()))

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

	alerts := view.NewAlertList([]domain.UptimeAlert{
		{ID: "a1", Status: "down", StatusCode: 503, Triggered: ptr(at("2026-10-01T10:00:00Z"))},
	})
	history := view.NewHistoryList([]domain.UptimeResult{
		{Created: at("2026-10-01T10:00:00Z"), Location: ptr("UK"), StatusCode: ptr(int32(200)), Performance: ptr(int64(120))},
		{Created: at("2026-10-01T10:05:00Z")},
	})
	periods := view.NewPeriodList([]domain.UptimePeriod{
		{Created: at("2026-10-01T10:00:00Z"), Ended: ptr(at("2026-10-01T11:00:00Z")), Status: "up", Duration: ptr(int64(3600000))},
	})
	results := view.NewPagespeedResultList([]domain.PagespeedResult{
		{Created: at("2026-10-01T10:00:00Z"), Filesize: 12.25, HARLocation: "x", Loadtime: 900, Requests: 40, Throttling: "4G"},
	})
	locations := view.NewLocationList([]domain.Location{
		{Description: "London", Region: "United Kingdom", RegionCode: "UK", Status: "up", IPv4: ptr("192.0.2.1")},
	})

	tests := map[string]struct {
		got  any
		want []string
	}{
		"alerts": {
			got:  alerts,
			want: []string{"a1", "2026-10-01T10:00:00Z", "down", "503"},
		},
		"history": {
			got:  history,
			want: []string{"UK", "200", "120 ms", "2026-10-01T10:05:00Z"},
		},
		"periods": {
			got:  periods,
			want: []string{"1h0m0s", "up"},
		},
		"results": {
			got:  results,
			want: []string{"900 ms", "40", "12.2 kB", "4G"},
		},
		"locations": {
			got:  locations,
			want: []string{"UK", "United Kingdom", "London", "192.0.2.1"},
		},
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

func TestListsShowUptimeOnlyWhenPresent(t *testing.T) {
	t.Parallel()

	without := view.NewUptimeList([]domain.UptimeOverview{{ID: "1", Name: "Home", Status: "up"}})
	with := view.NewUptimeList([]domain.UptimeOverview{{ID: "1", Name: "Home", Status: "up", Uptime: ptr(float32(99.5))}})
	hbWithout := view.NewHeartbeatList([]domain.HeartbeatOverview{{ID: "1", Name: "Job", Status: "up"}})
	hbWith := view.NewHeartbeatList([]domain.HeartbeatOverview{{ID: "1", Name: "Job", Status: "up", Uptime: ptr(float32(100))}})

	tests := map[string]struct {
		list any
		want bool
	}{
		"uptime without":    {without, false},
		"uptime with":       {with, true},
		"heartbeat without": {hbWithout, false},
		"heartbeat with":    {hbWith, true},
	}
	for name, tt := range tests {
		out := render(t, cliflag.FormatText, tt.list)
		if got := strings.Contains(out, "UPTIME"); got != tt.want {
			t.Errorf("%s: text has UPTIME column = %v, want %v:\n%s", name, got, tt.want, out)
		}
		out = render(t, cliflag.FormatJSON, tt.list)
		if got := strings.Contains(out, `"uptime"`); got != tt.want {
			t.Errorf("%s: JSON has uptime = %v, want %v:\n%s", name, got, tt.want, out)
		}
	}
}

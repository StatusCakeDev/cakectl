package adapter_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/adapter"
	"github.com/StatusCakeDev/cakectl/internal/domain"
	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// Every field of every spec is given a distinct value, so that an adapter that
// wires two fields to each other's parameters, or drops one, is caught.

var (
	maintenanceStart = time.Date(2026, 11, 1, 22, 0, 0, 0, time.UTC)
	maintenanceEnd   = time.Date(2026, 11, 2, 1, 0, 0, 0, time.UTC)

	fullUptime = spec.Uptime{
		Name: gofp.Some("n"), TestType: gofp.Some("HTTP"), WebsiteURL: gofp.Some("u"), CheckRate: gofp.Some(int32(300)),
		BasicUsername: gofp.Some("bu"), BasicPassword: gofp.Some("bp"), Confirmation: gofp.Some(int32(2)),
		ContactGroups: gofp.Some([]string{"c1", "c2"}), CustomHeader: gofp.Some("ch"), DNSIPs: gofp.Some([]string{"1.1.1.1"}),
		DNSServer: gofp.Some("ds"), DoNotFind: gofp.Some(true), EnableSSLAlert: gofp.Some(true), FinalEndpoint: gofp.Some("fe"),
		FindString: gofp.Some("fs"), FollowRedirects: gofp.Some(true), Host: gofp.Some("h"), IncludeHeader: gofp.Some(true),
		Paused: gofp.Some(true), Port: gofp.Some(int32(8080)), PostBody: gofp.Some("pb"), PostRaw: gofp.Some("pr"),
		Regions: gofp.Some([]string{"r1"}), StatusCodes: gofp.Some([]string{"500", "501"}), Tags: gofp.Some([]string{"t1", "t2"}),
		Timeout: gofp.Some(int32(20)), TriggerRate: gofp.Some(int32(3)), UseJAR: gofp.Some(true), UserAgent: gofp.Some("ua"),
	}
	fullUptimeForm = url.Values{
		"name":             {"n"},
		"test_type":        {"HTTP"},
		"website_url":      {"u"},
		"check_rate":       {"300"},
		"basic_username":   {"bu"},
		"basic_password":   {"bp"},
		"confirmation":     {"2"},
		"contact_groups[]": {"c1", "c2"},
		"custom_header":    {"ch"},
		"dns_ips[]":        {"1.1.1.1"},
		"dns_server":       {"ds"},
		"do_not_find":      {"true"},
		"enable_ssl_alert": {"true"},
		"final_endpoint":   {"fe"},
		"find_string":      {"fs"},
		"follow_redirects": {"true"},
		"host":             {"h"},
		"include_header":   {"true"},
		"paused":           {"true"},
		"port":             {"8080"},
		"post_body":        {"pb"},
		"post_raw":         {"pr"},
		"regions[]":        {"r1"},
		"status_codes_csv": {"500,501"},
		"tags[]":           {"t1", "t2"},
		"timeout":          {"20"},
		"trigger_rate":     {"3"},
		"use_jar":          {"true"},
		"user_agent":       {"ua"},
	}

	fullHeartbeat = spec.Heartbeat{
		Name: gofp.Some("n"), Period: gofp.Some(int32(600)), ContactGroups: gofp.Some([]string{"c1"}),
		Host: gofp.Some("h"), Paused: gofp.Some(true), Tags: gofp.Some([]string{"t1"}),
	}
	fullHeartbeatForm = url.Values{
		"name":             {"n"},
		"period":           {"600"},
		"contact_groups[]": {"c1"},
		"host":             {"h"},
		"paused":           {"true"},
		"tags[]":           {"t1"},
	}

	fullPagespeed = spec.Pagespeed{
		Name: gofp.Some("n"), WebsiteURL: gofp.Some("u"), CheckRate: gofp.Some(int32(600)), Region: gofp.Some("UK"),
		AlertBigger: gofp.Some(int32(11)), AlertSlower: gofp.Some(int64(22)), AlertSmaller: gofp.Some(int32(33)),
		ContactGroups: gofp.Some([]string{"c1"}), Paused: gofp.Some(true),
	}
	fullPagespeedForm = url.Values{
		"name":             {"n"},
		"website_url":      {"u"},
		"check_rate":       {"600"},
		"region":           {"UK"},
		"alert_bigger":     {"11"},
		"alert_slower":     {"22"},
		"alert_smaller":    {"33"},
		"contact_groups[]": {"c1"},
		"paused":           {"true"},
	}

	fullSSL = spec.SSL{
		WebsiteURL: gofp.Some("u"), CheckRate: gofp.Some(int32(3600)), AlertAt: gofp.Some([]int32{7, 30}),
		AlertBroken: gofp.Some(true), AlertExpiry: gofp.Some(false), AlertMixed: gofp.Some(true), AlertReminder: gofp.Some(false),
		ContactGroups: gofp.Some([]string{"c1"}), FollowRedirects: gofp.Some(true), Hostname: gofp.Some("hn"),
		Paused: gofp.Some(true), UserAgent: gofp.Some("ua"),
	}
	fullSSLForm = url.Values{
		"website_url":      {"u"},
		"check_rate":       {"3600"},
		"alert_at[]":       {"7", "30"},
		"alert_broken":     {"true"},
		"alert_expiry":     {"false"},
		"alert_mixed":      {"true"},
		"alert_reminder":   {"false"},
		"contact_groups[]": {"c1"},
		"follow_redirects": {"true"},
		"hostname":         {"hn"},
		"paused":           {"true"},
		"user_agent":       {"ua"},
	}

	fullContactGroup = spec.ContactGroup{
		Name: gofp.Some("n"), EmailAddresses: gofp.Some([]string{"a@example.com"}), Integrations: gofp.Some([]string{"i1"}),
		MobileNumbers: gofp.Some([]string{"+44"}), PingURL: gofp.Some("https://example.com/ping"),
	}
	fullContactGroupForm = url.Values{
		"name":              {"n"},
		"email_addresses[]": {"a@example.com"},
		"integrations[]":    {"i1"},
		"mobile_numbers[]":  {"+44"},
		"ping_url":          {"https://example.com/ping"},
	}

	fullMaintenance = spec.Maintenance{
		Name: gofp.Some("n"), StartAt: gofp.Some(maintenanceStart), EndAt: gofp.Some(maintenanceEnd),
		Timezone: gofp.Some("Europe/London"), RepeatInterval: gofp.Some("1w"),
		Tags: gofp.Some([]string{"t1"}), Tests: gofp.Some([]string{"x1"}),
	}
	fullMaintenanceForm = url.Values{
		"name":            {"n"},
		"start_at":        {"2026-11-01T22:00:00Z"},
		"end_at":          {"2026-11-02T01:00:00Z"},
		"timezone":        {"Europe/London"},
		"repeat_interval": {"1w"},
		"tags[]":          {"t1"},
		"tests[]":         {"x1"},
	}
)

// resource describes how to exercise one resource's write operations without
// the test knowing its spec type.
type resource struct {
	name string
	// path is the collection path, without the ID.
	path string
	// create, update and remove perform the operation with the full spec.
	create func(*adapter.Client) (string, error)
	update func(*adapter.Client, string) error
	remove func(*adapter.Client, string) error
	// empty and invalid are specs that must be rejected locally.
	updateEmpty func(*adapter.Client) error
	createEmpty func(*adapter.Client) (string, error)
	// form is what the full spec is expected to send.
	form url.Values
	// one updates a single field and says what that sends.
	updateOne     func(*adapter.Client, string) error
	updateOneForm url.Values
}

func resources() []resource {
	ctx := context.Background()
	return []resource{
		{
			name: "uptime", path: "/v1/uptime", form: fullUptimeForm,
			create:      func(c *adapter.Client) (string, error) { return c.CreateUptime(ctx, fullUptime) },
			createEmpty: func(c *adapter.Client) (string, error) { return c.CreateUptime(ctx, spec.Uptime{}) },
			update: func(c *adapter.Client, id string) error {
				u := fullUptime
				u.TestType = gofp.None[string]() // The type cannot be changed.
				return c.UpdateUptime(ctx, id, u)
			},
			updateEmpty: func(c *adapter.Client) error { return c.UpdateUptime(ctx, "7", spec.Uptime{}) },
			updateOne: func(c *adapter.Client, id string) error {
				return c.UpdateUptime(ctx, id, spec.Uptime{Paused: gofp.Some(false)})
			},
			updateOneForm: url.Values{"paused": {"false"}},
			remove:        func(c *adapter.Client, id string) error { return c.DeleteUptime(ctx, id) },
		},
		{
			name: "heartbeat", path: "/v1/heartbeat", form: fullHeartbeatForm,
			create:      func(c *adapter.Client) (string, error) { return c.CreateHeartbeat(ctx, fullHeartbeat) },
			createEmpty: func(c *adapter.Client) (string, error) { return c.CreateHeartbeat(ctx, spec.Heartbeat{}) },
			update:      func(c *adapter.Client, id string) error { return c.UpdateHeartbeat(ctx, id, fullHeartbeat) },
			updateEmpty: func(c *adapter.Client) error { return c.UpdateHeartbeat(ctx, "7", spec.Heartbeat{}) },
			updateOne: func(c *adapter.Client, id string) error {
				return c.UpdateHeartbeat(ctx, id, spec.Heartbeat{Period: gofp.Some(int32(90))})
			},
			updateOneForm: url.Values{"period": {"90"}},
			remove:        func(c *adapter.Client, id string) error { return c.DeleteHeartbeat(ctx, id) },
		},
		{
			name: "pagespeed", path: "/v1/pagespeed", form: fullPagespeedForm,
			create:      func(c *adapter.Client) (string, error) { return c.CreatePagespeed(ctx, fullPagespeed) },
			createEmpty: func(c *adapter.Client) (string, error) { return c.CreatePagespeed(ctx, spec.Pagespeed{}) },
			update:      func(c *adapter.Client, id string) error { return c.UpdatePagespeed(ctx, id, fullPagespeed) },
			updateEmpty: func(c *adapter.Client) error { return c.UpdatePagespeed(ctx, "7", spec.Pagespeed{}) },
			updateOne: func(c *adapter.Client, id string) error {
				return c.UpdatePagespeed(ctx, id, spec.Pagespeed{Region: gofp.Some("US")})
			},
			updateOneForm: url.Values{"region": {"US"}},
			remove:        func(c *adapter.Client, id string) error { return c.DeletePagespeed(ctx, id) },
		},
		{
			name: "ssl", path: "/v1/ssl", form: fullSSLForm,
			create:      func(c *adapter.Client) (string, error) { return c.CreateSSL(ctx, fullSSL) },
			createEmpty: func(c *adapter.Client) (string, error) { return c.CreateSSL(ctx, spec.SSL{}) },
			update:      func(c *adapter.Client, id string) error { return c.UpdateSSL(ctx, id, fullSSL) },
			updateEmpty: func(c *adapter.Client) error { return c.UpdateSSL(ctx, "7", spec.SSL{}) },
			updateOne: func(c *adapter.Client, id string) error {
				return c.UpdateSSL(ctx, id, spec.SSL{AlertAt: gofp.Some([]int32{1, 2, 3})})
			},
			updateOneForm: url.Values{"alert_at[]": {"1", "2", "3"}},
			remove:        func(c *adapter.Client, id string) error { return c.DeleteSSL(ctx, id) },
		},
		{
			name: "contact-group", path: "/v1/contact-groups", form: fullContactGroupForm,
			create:      func(c *adapter.Client) (string, error) { return c.CreateContactGroup(ctx, fullContactGroup) },
			createEmpty: func(c *adapter.Client) (string, error) { return c.CreateContactGroup(ctx, spec.ContactGroup{}) },
			update:      func(c *adapter.Client, id string) error { return c.UpdateContactGroup(ctx, id, fullContactGroup) },
			updateEmpty: func(c *adapter.Client) error { return c.UpdateContactGroup(ctx, "7", spec.ContactGroup{}) },
			updateOne: func(c *adapter.Client, id string) error {
				return c.UpdateContactGroup(ctx, id, spec.ContactGroup{Name: gofp.Some("renamed")})
			},
			updateOneForm: url.Values{"name": {"renamed"}},
			remove:        func(c *adapter.Client, id string) error { return c.DeleteContactGroup(ctx, id) },
		},
		{
			name: "maintenance-window", path: "/v1/maintenance-windows", form: fullMaintenanceForm,
			create:      func(c *adapter.Client) (string, error) { return c.CreateMaintenance(ctx, fullMaintenance) },
			createEmpty: func(c *adapter.Client) (string, error) { return c.CreateMaintenance(ctx, spec.Maintenance{}) },
			update:      func(c *adapter.Client, id string) error { return c.UpdateMaintenance(ctx, id, fullMaintenance) },
			updateEmpty: func(c *adapter.Client) error { return c.UpdateMaintenance(ctx, "7", spec.Maintenance{}) },
			updateOne: func(c *adapter.Client, id string) error {
				return c.UpdateMaintenance(ctx, id, spec.Maintenance{RepeatInterval: gofp.Some("never")})
			},
			updateOneForm: url.Values{"repeat_interval": {"never"}},
			remove:        func(c *adapter.Client, id string) error { return c.DeleteMaintenance(ctx, id) },
		},
	}
}

func TestCreateSendsEveryField(t *testing.T) {
	t.Parallel()

	for _, r := range resources() {
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()

			api, c := newAPI(t)
			api.serve("POST "+r.path, http.StatusCreated, created("42"))

			id, err := r.create(c)
			if err != nil {
				t.Fatalf("create error = %v", err)
			}
			if id != "42" {
				t.Errorf("create ID = %q, want %q", id, "42")
			}

			got := api.only(t)
			if got.Method != http.MethodPost || got.Path != r.path {
				t.Errorf("request = %s %s, want POST %s", got.Method, got.Path, r.path)
			}
			if diff := cmp.Diff(r.form, got.Form); diff != "" {
				t.Errorf("form mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCreateRejectsAnInvalidSpecWithoutCallingTheAPI(t *testing.T) {
	t.Parallel()

	for _, r := range resources() {
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()

			api, c := newAPI(t)

			_, err := r.createEmpty(c)
			if err == nil || !strings.Contains(err.Error(), "is required") {
				t.Errorf("create error = %v, want a missing field error", err)
			}
			if calls := api.calls(); len(calls) != 0 {
				t.Errorf("served %d requests, want none: %+v", len(calls), calls)
			}
		})
	}
}

func TestCreateFailureIsReportedWithItsCause(t *testing.T) {
	t.Parallel()

	api, c := newAPI(t)
	api.serve("POST /v1/heartbeat", http.StatusBadRequest, map[string]any{
		"message": "invalid", "errors": map[string]any{"period": []string{"too small"}},
	})

	_, err := c.CreateHeartbeat(context.Background(), fullHeartbeat)

	var apiErr statuscake.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusBadRequest {
		t.Fatalf("error = %v, want a 400 APIError", err)
	}
	if errors.Is(err, domain.ErrNotFound) {
		t.Error("a 400 was reported as not found")
	}
}

func TestUpdateSendsEveryField(t *testing.T) {
	t.Parallel()

	for _, r := range resources() {
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()

			api, c := newAPI(t)
			api.serve("PUT "+r.path+"/{id}", http.StatusNoContent, nil)

			if err := r.update(c, "7"); err != nil {
				t.Fatalf("update error = %v", err)
			}

			want := cloneWithout(r.form, "test_type")
			got := api.only(t)
			if got.Method != http.MethodPut || got.Path != r.path+"/7" {
				t.Errorf("request = %s %s, want PUT %s/7", got.Method, got.Path, r.path)
			}
			if diff := cmp.Diff(want, got.Form); diff != "" {
				t.Errorf("form mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUpdateSendsOnlyTheFieldsSupplied(t *testing.T) {
	t.Parallel()

	for _, r := range resources() {
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()

			api, c := newAPI(t)
			api.serve("PUT "+r.path+"/{id}", http.StatusNoContent, nil)

			if err := r.updateOne(c, "7"); err != nil {
				t.Fatalf("update error = %v", err)
			}

			if diff := cmp.Diff(r.updateOneForm, api.only(t).Form); diff != "" {
				t.Errorf("form mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUpdateWithNothingToChangeDoesNotCallTheAPI(t *testing.T) {
	t.Parallel()

	for _, r := range resources() {
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()

			api, c := newAPI(t)

			if err := r.updateEmpty(c); !errors.Is(err, domain.ErrNothingToUpdate) {
				t.Errorf("update error = %v, want %v", err, domain.ErrNothingToUpdate)
			}
			if calls := api.calls(); len(calls) != 0 {
				t.Errorf("served %d requests, want none: %+v", len(calls), calls)
			}
		})
	}
}

func TestUpdateRejectsAnInvalidSpecWithoutCallingTheAPI(t *testing.T) {
	t.Parallel()

	api, c := newAPI(t)

	err := c.UpdateUptime(context.Background(), "7", spec.Uptime{TestType: gofp.Some("HTTP")})
	if err == nil || !strings.Contains(err.Error(), "test_type cannot be changed") {
		t.Errorf("error = %v, want one about test_type", err)
	}
	if calls := api.calls(); len(calls) != 0 {
		t.Errorf("served %d requests, want none: %+v", len(calls), calls)
	}
}

func TestUpdateCanClearAList(t *testing.T) {
	t.Parallel()

	api, c := newAPI(t)
	api.serve("PUT /v1/uptime/{id}", http.StatusNoContent, nil)

	err := c.UpdateUptime(context.Background(), "7", spec.Uptime{ContactGroups: gofp.Some([]string{})})
	if err != nil {
		t.Fatalf("update error = %v", err)
	}

	// An empty value is how the API is told to empty the list; sending nothing
	// would leave it alone.
	want := url.Values{"contact_groups[]": {""}}
	if diff := cmp.Diff(want, api.only(t).Form); diff != "" {
		t.Errorf("form mismatch (-want +got):\n%s", diff)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	for _, r := range resources() {
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()

			api, c := newAPI(t)
			api.serve("DELETE "+r.path+"/{id}", http.StatusNoContent, nil)

			if err := r.remove(c, "7"); err != nil {
				t.Fatalf("delete error = %v", err)
			}

			got := api.only(t)
			if got.Method != http.MethodDelete || got.Path != r.path+"/7" {
				t.Errorf("request = %s %s, want DELETE %s/7", got.Method, got.Path, r.path)
			}
		})
	}
}

func TestWritesTranslateNotFound(t *testing.T) {
	t.Parallel()

	for _, r := range resources() {
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()

			_, c := newAPI(t) // Nothing is declared, so every route is a 404.

			for op, err := range map[string]error{
				"update": r.updateOne(c, "missing"),
				"delete": r.remove(c, "missing"),
			} {
				if !errors.Is(err, domain.ErrNotFound) {
					t.Errorf("%s error = %v, want %v", op, err, domain.ErrNotFound)
				}
				var apiErr statuscake.APIError
				if !errors.As(err, &apiErr) {
					t.Errorf("%s error = %v, want the SDK's error to be kept", op, err)
				}
			}
		})
	}
}

func TestDeleteOtherFailuresAreNotNotFound(t *testing.T) {
	t.Parallel()

	api, c := newAPI(t)
	api.serve("DELETE /v1/uptime/{id}", http.StatusInternalServerError, map[string]any{"message": "boom"})

	err := c.DeleteUptime(context.Background(), "7")
	if err == nil || errors.Is(err, domain.ErrNotFound) {
		t.Errorf("error = %v, want a failure that is not %v", err, domain.ErrNotFound)
	}
}

func TestRejectedTokensAreReportedAsUnauthorized(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()

			api, c := newAPI(t)
			api.serve("GET /v1/uptime", status, map[string]any{"message": "no"})
			api.serve("DELETE /v1/uptime/{id}", status, map[string]any{"message": "no"})

			_, err := c.ListUptime(context.Background(), adapter.UptimeFilter{})
			if !errors.Is(err, domain.ErrUnauthorized) || errors.Is(err, domain.ErrNotFound) {
				t.Errorf("ListUptime() error = %v, want ErrUnauthorized and not ErrNotFound", err)
			}
			if err := c.DeleteUptime(context.Background(), "7"); !errors.Is(err, domain.ErrUnauthorized) {
				t.Errorf("DeleteUptime() error = %v, want one matching %v", err, domain.ErrUnauthorized)
			}
		})
	}
}

func cloneWithout(v url.Values, keys ...string) url.Values {
	out := url.Values{}
	for k, vs := range v {
		out[k] = vs
	}
	for _, k := range keys {
		delete(out, k)
	}
	return out
}

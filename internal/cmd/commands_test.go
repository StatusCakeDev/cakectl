package cmd

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/StatusCakeDev/cakectl/internal/domain"
)

func page(items ...any) map[string]any {
	if items == nil {
		items = []any{}
	}
	return map[string]any{
		"data":     items,
		"metadata": map[string]any{"page": 1, "per_page": 100, "page_count": 1, "total_count": len(items)},
	}
}

func created(id string) map[string]any {
	return map[string]any{"data": map[string]any{"new_id": id}}
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func onlyCall(t *testing.T, api *fakeAPI) call {
	t.Helper()

	calls := api.calls()
	if len(calls) != 1 {
		t.Fatalf("served %d requests, want exactly 1: %+v", len(calls), calls)
	}
	return calls[0]
}

func TestCreateFromFlags(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("POST /v1/uptime", http.StatusCreated, created("42"))
	s := newSession(t, api)

	err := s.run("uptime", "create",
		"--name", "Home page", "--test-type", "HTTP", "--website-url", "https://example.com",
		"--check-rate", "300", "--tags", "prod,web", "--paused=false")
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	want := url.Values{
		"name":        {"Home page"},
		"test_type":   {"HTTP"},
		"website_url": {"https://example.com"},
		"check_rate":  {"300"},
		"tags[]":      {"prod", "web"},
		"paused":      {"false"},
	}
	if diff := cmp.Diff(want, onlyCall(t, api).Form); diff != "" {
		t.Errorf("form mismatch (-want +got):\n%s", diff)
	}
	if got := s.stdout(); got != "uptime 42 created" {
		t.Errorf("output = %q, want %q", got, "uptime 42 created")
	}
}

func TestCreateFromFileWithFlagOverrides(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("POST /v1/uptime", http.StatusCreated, created("42"))
	s := newSession(t, api)

	file := writeFile(t, "uptime.yaml", `
name: From file
test_type: HTTP
website_url: https://example.com
check_rate: 300
contact_groups: ["1"]
`)
	if err := s.run("uptime", "create", "--filename", file, "--name", "From flag"); err != nil {
		t.Fatalf("error = %v", err)
	}

	want := url.Values{
		"name":             {"From flag"},
		"test_type":        {"HTTP"},
		"website_url":      {"https://example.com"},
		"check_rate":       {"300"},
		"contact_groups[]": {"1"},
	}
	if diff := cmp.Diff(want, onlyCall(t, api).Form); diff != "" {
		t.Errorf("form mismatch (-want +got):\n%s", diff)
	}
}

func TestCreateFromStandardInput(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("POST /v1/heartbeat", http.StatusCreated, created("5"))
	s := newSession(t, api)
	s.in.WriteString(`{"name": "Nightly", "period": 600}`)

	if err := s.run("heartbeat", "create", "--filename", "-"); err != nil {
		t.Fatalf("error = %v", err)
	}

	want := url.Values{"name": {"Nightly"}, "period": {"600"}}
	if diff := cmp.Diff(want, onlyCall(t, api).Form); diff != "" {
		t.Errorf("form mismatch (-want +got):\n%s", diff)
	}
}

func TestCreateRejectsBadInputWithoutCallingTheAPI(t *testing.T) {
	tests := map[string]struct {
		args []string
		file string
		want string
	}{
		"missing required fields": {
			args: []string{"uptime", "create", "--name", "x"},
			want: "test_type is required",
		},
		"unknown field in file": {
			args: []string{"uptime", "create", "--filename"},
			file: "nmae: x\n",
			want: "nmae",
		},
		"bad enum from flag": {
			args: []string{"pagespeed", "create", "--name", "x", "--website-url", "u", "--check-rate", "30", "--region", "UK"},
			want: "check_rate must be one of",
		},
		"bad time": {
			args: []string{"maintenance-window", "create", "--start-at", "tomorrow"},
			want: "RFC 3339",
		},
		"missing file": {
			args: []string{"uptime", "create", "--filename", "/no/such/file.yaml"},
			want: "/no/such/file.yaml",
		},
		"empty file": {
			args: []string{"uptime", "create", "--filename"},
			file: "",
			want: "empty",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			api := newFakeAPI(t)
			s := newSession(t, api)

			args := tt.args
			if args[len(args)-1] == "--filename" {
				args = append(args, writeFile(t, "spec.yaml", tt.file))
			}

			err := s.run(args...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want it to mention %q", err, tt.want)
			}
			if calls := api.calls(); len(calls) != 0 {
				t.Errorf("served %d requests, want none", len(calls))
			}
		})
	}
}

func TestUpdateSendsOnlyWhatWasSupplied(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("PUT /v1/ssl/{id}", http.StatusNoContent, nil)
	s := newSession(t, api)

	if err := s.run("ssl", "update", "9", "--alert-at", "1,7", "--paused"); err != nil {
		t.Fatalf("error = %v", err)
	}

	got := onlyCall(t, api)
	want := url.Values{"alert_at[]": {"1", "7"}, "paused": {"true"}}
	if diff := cmp.Diff(want, got.Form); diff != "" {
		t.Errorf("form mismatch (-want +got):\n%s", diff)
	}
	if got.Path != "/v1/ssl/9" {
		t.Errorf("path = %q, want /v1/ssl/9", got.Path)
	}
	if s.stdout() != "ssl 9 updated" {
		t.Errorf("output = %q", s.stdout())
	}
}

func TestUpdateCanClearAList(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("PUT /v1/uptime/{id}", http.StatusNoContent, nil)
	s := newSession(t, api)

	if err := s.run("uptime", "update", "9", "--contact-groups", ""); err != nil {
		t.Fatalf("error = %v", err)
	}

	want := url.Values{"contact_groups[]": {""}}
	if diff := cmp.Diff(want, onlyCall(t, api).Form); diff != "" {
		t.Errorf("form mismatch (-want +got):\n%s", diff)
	}
}

func TestUpdateWithNothingSuppliedSaysHowToProceed(t *testing.T) {
	api := newFakeAPI(t)
	s := newSession(t, api)

	err := s.run("uptime", "update", "9")
	if err == nil || !strings.Contains(err.Error(), "nothing to update") || !strings.Contains(err.Error(), "--filename") {
		t.Errorf("error = %v, want advice on supplying fields", err)
	}
	if calls := api.calls(); len(calls) != 0 {
		t.Errorf("served %d requests, want none", len(calls))
	}
}

func TestUpdateDoesNotAcceptATypeChange(t *testing.T) {
	api := newFakeAPI(t)
	s := newSession(t, api)

	// The flag is not offered on update at all.
	err := s.run("uptime", "update", "9", "--test-type", "TCP")
	if err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("error = %v, want an unknown flag error", err)
	}
}

// TestDescribedResourcesCanBeEditedAndAppliedBack is the workflow the hybrid
// input design exists for: describe to a file, edit, update from the file.
func TestDescribedResourcesCanBeEditedAndAppliedBack(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("GET /v1/maintenance-windows/{id}", http.StatusOK, map[string]any{"data": map[string]any{
		"id": "12", "name": "Deploy", "start_at": "2026-11-01T22:00:00Z", "end_at": "2026-11-02T01:00:00Z",
		"repeat_interval": "1w", "state": "pending", "tags": []string{"release"}, "tests": []string{"7", "8"},
		"timezone": "Europe/London",
	}})
	api.serve("PUT /v1/maintenance-windows/{id}", http.StatusNoContent, nil)

	describe := newSession(t, api)
	if err := describe.run("maintenance-window", "describe", "12", "--format", "yaml"); err != nil {
		t.Fatalf("describe error = %v", err)
	}

	edited := strings.Replace(describe.out.String(), "name: Deploy", "name: Deploy v2", 1)
	if edited == describe.out.String() {
		t.Fatalf("description has no name to edit:\n%s", edited)
	}

	update := newSession(t, api)
	update.in.WriteString(edited)
	if err := update.run("maintenance-window", "update", "12", "--filename", "-"); err != nil {
		t.Fatalf("update error = %v", err)
	}

	calls := api.calls()
	got := calls[len(calls)-1]
	want := url.Values{
		"name":            {"Deploy v2"},
		"start_at":        {"2026-11-01T22:00:00Z"},
		"end_at":          {"2026-11-02T01:00:00Z"},
		"timezone":        {"Europe/London"},
		"repeat_interval": {"1w"},
		"tags[]":          {"release"},
		"tests[]":         {"7", "8"},
	}
	if got.Method != http.MethodPut || got.Path != "/v1/maintenance-windows/12" {
		t.Errorf("request = %s %s", got.Method, got.Path)
	}
	if diff := cmp.Diff(want, got.Form); diff != "" {
		t.Errorf("form mismatch (-want +got):\n%s", diff)
	}
}

func TestDeleteWithYesDoesNotAsk(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("DELETE /v1/contact-groups/{id}", http.StatusNoContent, nil)
	s := newSession(t, api)

	if err := s.run("contact-group", "delete", "3", "--yes"); err != nil {
		t.Fatalf("error = %v", err)
	}

	if got := onlyCall(t, api); got.Method != http.MethodDelete || got.Path != "/v1/contact-groups/3" {
		t.Errorf("request = %s %s", got.Method, got.Path)
	}
	if s.stdout() != "contact-group 3 deleted" {
		t.Errorf("output = %q", s.stdout())
	}
	if s.errOut.Len() != 0 {
		t.Errorf("--yes still prompted: %q", s.errOut)
	}
}

func TestDeleteRefusesWithoutATerminal(t *testing.T) {
	api := newFakeAPI(t)
	s := newSession(t, api)
	// Even an affirmative answer on a pipe must not be taken as consent.
	s.in.WriteString("y\n")

	err := s.run("uptime", "delete", "3")
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Errorf("error = %v, want one telling the user to pass --yes", err)
	}
	if calls := api.calls(); len(calls) != 0 {
		t.Errorf("served %d requests, want none", len(calls))
	}
}

func TestDeleteAsksOnATerminal(t *testing.T) {
	tests := map[string]struct {
		answer  string
		deleted bool
	}{
		"y": {
			answer:  "y\n",
			deleted: true,
		},
		"yes": {
			answer:  "YES\n",
			deleted: true,
		},
		"n": {
			answer:  "n\n",
			deleted: false,
		},
		"empty": {
			answer:  "\n",
			deleted: false,
		},
		"anything": {
			answer:  "maybe\n",
			deleted: false,
		},
		"no answer": {
			answer:  "",
			deleted: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			api := newFakeAPI(t)
			api.serve("DELETE /v1/uptime/{id}", http.StatusNoContent, nil)
			s := newSession(t, api)
			s.opts.isTerminal = func(io.Reader) bool { return true }
			s.in.WriteString(tt.answer)

			err := s.run("uptime", "delete", "3")

			if prompt := s.errOut.String(); !strings.Contains(prompt, "Delete uptime check 3? [y/N]") {
				t.Errorf("prompt = %q, want it to name the resource", prompt)
			}
			if tt.deleted {
				if err != nil || len(api.calls()) != 1 {
					t.Errorf("error = %v, calls = %d; want a delete", err, len(api.calls()))
				}
				return
			}
			if !errors.Is(err, ErrNotConfirmed) || len(api.calls()) != 0 {
				t.Errorf("error = %v, calls = %d; want %v and no delete", err, len(api.calls()), ErrNotConfirmed)
			}
		})
	}
}

func TestDeleteOfAMissingResourceFails(t *testing.T) {
	api := newFakeAPI(t)
	s := newSession(t, api)

	err := s.run("ssl", "delete", "404", "-y")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestListAndFilters(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("GET /v1/uptime", http.StatusOK, page(map[string]any{
		"id": "1", "name": "Home", "test_type": "HTTP", "website_url": "https://example.com",
		"check_rate": 300, "contact_groups": []string{}, "paused": false, "status": "down", "tags": []string{"prod"},
	}))
	s := newSession(t, api)

	if err := s.run("uptime", "list", "--status", "down", "--tags", "prod"); err != nil {
		t.Fatalf("error = %v", err)
	}

	wantQuery := url.Values{"page": {"1"}, "limit": {"100"}, "status": {"down"}, "tags": {"prod"}, "matchany": {"false"}, "nouptime": {"true"}}
	if diff := cmp.Diff(wantQuery, onlyCall(t, api).Query); diff != "" {
		t.Errorf("query mismatch (-want +got):\n%s", diff)
	}

	out := s.out.String()
	for _, want := range []string{"ID", "STATUS", "Home", "https://example.com", "5m0s", "down"} {
		if !strings.Contains(out, want) {
			t.Errorf("list output does not contain %q:\n%s", want, out)
		}
	}
}

func TestListLeavesOutUptimeUnlessAsked(t *testing.T) {
	for _, resource := range []string{"uptime", "heartbeat"} {
		t.Run(resource, func(t *testing.T) {
			api := newFakeAPI(t)
			api.serve("GET /v1/"+resource, http.StatusOK, page())
			s := newSession(t, api)

			if err := s.run(resource, "list"); err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := onlyCall(t, api).Query.Get("nouptime"); got != "true" {
				t.Errorf("nouptime = %q, want true", got)
			}

			api = newFakeAPI(t)
			api.serve("GET /v1/"+resource, http.StatusOK, page())
			s = newSession(t, api)

			if err := s.run(resource, "list", "--uptime"); err != nil {
				t.Fatalf("error = %v", err)
			}

			// The API treats any nouptime parameter as a request to leave the
			// percentage out, whatever its value, so it must be absent.
			if _, has := onlyCall(t, api).Query["nouptime"]; has {
				t.Error("nouptime was sent, want it absent")
			}
		})
	}
}

func TestOutputFormats(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("GET /v1/contact-groups", http.StatusOK, page(map[string]any{
		"id": "3", "name": "Ops", "email_addresses": []string{"ops@example.com"},
		"integrations": []string{}, "mobile_numbers": []string{},
	}))

	tests := map[string]string{
		"text":       "Ops",
		"json":       `"name":"Ops"`,
		"prettyjson": `"name": "Ops"`,
		"yaml":       "name: Ops",
	}
	for format, want := range tests {
		s := newSession(t, api)
		if err := s.run("contact-group", "list", "--format", format); err != nil {
			t.Fatalf("-o %s error = %v", format, err)
		}
		if !strings.Contains(s.out.String(), want) {
			t.Errorf("-o %s output does not contain %q:\n%s", format, want, s.out)
		}
		// The listing never exposes contact details, in any format.
		if strings.Contains(s.out.String(), "ops@example.com") {
			t.Errorf("-o %s output leaked an email address:\n%s", format, s.out)
		}
	}
}

func TestInvalidOutputFormat(t *testing.T) {
	s := newSession(t, newFakeAPI(t))

	err := s.run("uptime", "list", "--format", "xml")
	if err == nil || !strings.Contains(err.Error(), "text|json|prettyjson|yaml") {
		t.Errorf("error = %v, want one listing the formats", err)
	}
}

func TestDescribe(t *testing.T) {
	api := newFakeAPI(t)
	api.serve("GET /v1/heartbeat/{id}", http.StatusOK, map[string]any{"data": map[string]any{
		"id": "8", "name": "Nightly", "url": "https://heartbeat.example/TOKEN", "period": 600,
		"contact_groups": []string{}, "paused": false, "status": "up", "tags": []string{}, "uptime": 100,
	}})

	s := newSession(t, api)
	if err := s.run("heartbeat", "describe", "8"); err != nil {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(s.out.String(), "https://heartbeat.example/TOKEN") {
		t.Errorf("describe does not show the ping URL:\n%s", s.out)
	}

	// `get` is an alias.
	s = newSession(t, api)
	if err := s.run("heartbeat", "get", "8", "--format", "json"); err != nil || !strings.Contains(s.out.String(), `"period":600`) {
		t.Errorf("get --format json = %v:\n%s", err, s.out)
	}
}

func TestDescribeOfAMissingResource(t *testing.T) {
	s := newSession(t, newFakeAPI(t))

	err := s.run("pagespeed", "describe", "404")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestHistoryCommands(t *testing.T) {
	empty := map[string]any{"data": []any{}, "links": map[string]any{"self": "x"}}

	tests := map[string]struct {
		args []string
		path string
	}{
		"alerts": {
			args: []string{"uptime", "alerts", "7"},
			path: "/v1/uptime/7/alerts",
		},
		"history": {
			args: []string{"uptime", "history", "7"},
			path: "/v1/uptime/7/history",
		},
		"periods": {
			args: []string{"uptime", "periods", "7"},
			path: "/v1/uptime/7/periods",
		},
		"pagespeed history": {
			args: []string{"pagespeed", "history", "7"},
			path: "/v1/pagespeed/7/history",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			api := newFakeAPI(t)
			api.serve("GET "+tt.path, http.StatusOK, empty)
			s := newSession(t, api)

			args := append(tt.args, "--limit", "5", "--after", "2026-10-01T00:00:00Z", "--before", "2026-10-02T00:00:00Z")
			if err := s.run(args...); err != nil {
				t.Fatalf("error = %v", err)
			}

			want := url.Values{"limit": {"5"}, "after": {"1790812800"}, "before": {"1790899200"}}
			if diff := cmp.Diff(want, onlyCall(t, api).Query); diff != "" {
				t.Errorf("query mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHistoryRejectsBadWindows(t *testing.T) {
	for _, args := range [][]string{
		{"uptime", "history", "7", "--after", "yesterday"},
		{"uptime", "history", "7", "--before", "2026-10-01"},
		{"pagespeed", "history", "7", "--limit", "-1"},
	} {
		api := newFakeAPI(t)
		s := newSession(t, api)

		if err := s.run(args...); err == nil {
			t.Errorf("%v: error = nil, want one", args)
		}
		if len(api.calls()) != 0 {
			t.Errorf("%v: called the API", args)
		}
	}
}

func TestLocations(t *testing.T) {
	body := map[string]any{"data": []any{map[string]any{
		"description": "London", "region": "United Kingdom", "region_code": "UK", "status": "up",
	}}}

	for _, kind := range []string{"uptime", "pagespeed"} {
		api := newFakeAPI(t)
		api.serve("GET /v1/"+kind+"-locations", http.StatusOK, body)
		s := newSession(t, api)

		if err := s.run("locations", kind); err != nil {
			t.Fatalf("locations %s error = %v", kind, err)
		}
		if !strings.Contains(s.out.String(), "London") {
			t.Errorf("locations %s output:\n%s", kind, s.out)
		}
	}
}

func TestEveryResourceHasTheStandardCommands(t *testing.T) {
	root := NewRootCommandWithArgs(newSession(t, newFakeAPI(t)).opts)

	for _, resource := range []string{"uptime", "heartbeat", "pagespeed", "ssl", "contact-group", "maintenance-window"} {
		cmd, _, err := root.Find([]string{resource})
		if err != nil || cmd.Name() != resource {
			t.Errorf("no %q command: %v", resource, err)
			continue
		}
		for _, sub := range []string{"create", "update", "delete", "describe", "list"} {
			if c, _, err := root.Find([]string{resource, sub}); err != nil || c.Name() != sub {
				t.Errorf("%s has no %q subcommand", resource, sub)
			}
		}
	}
}

func TestCommandsAreGrouped(t *testing.T) {
	root := NewRootCommandWithArgs(newSession(t, newFakeAPI(t)).opts)

	want := map[string]string{
		"uptime": "Monitoring Commands:", "heartbeat": "Monitoring Commands:",
		"pagespeed": "Monitoring Commands:", "ssl": "Monitoring Commands:",
		"contact-group": "Alerting Commands:", "maintenance-window": "Alerting Commands:",
		"locations": "Reference Commands:",
	}
	for name, group := range want {
		cmd, _, _ := root.Find([]string{name})
		if cmd.GroupID != group {
			t.Errorf("%s group = %q, want %q", name, cmd.GroupID, group)
		}
	}
}

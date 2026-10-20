package factories_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/StatusCakeDev/cakectl/internal/factories"
)

func valid() factories.ClientConfig {
	return factories.ClientConfig{
		APIToken:   "tok",
		RPS:        4,
		Retries:    3,
		MinBackoff: time.Second,
		MaxBackoff: 30 * time.Second,
	}
}

func TestCreateClientAcceptsAValidConfig(t *testing.T) {
	t.Parallel()

	client, err := factories.CreateClient(valid())
	if err != nil || client == nil {
		t.Fatalf("CreateClient() = %v, %v; want a client", client, err)
	}
}

func TestCreateClientRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mutate func(*factories.ClientConfig)
		want   string
	}{
		"no token":            {func(c *factories.ClientConfig) { c.APIToken = "" }, "API token is required"},
		"zero rps":            {func(c *factories.ClientConfig) { c.RPS = 0 }, "rps must be greater than zero"},
		"negative retries":    {func(c *factories.ClientConfig) { c.Retries = -1 }, "retries must not be negative"},
		"zero min backoff":    {func(c *factories.ClientConfig) { c.MinBackoff = 0 }, "min backoff must be greater than zero"},
		"max below min":       {func(c *factories.ClientConfig) { c.MaxBackoff = time.Millisecond }, "must not be less than min backoff"},
		"everything wrong":    {func(c *factories.ClientConfig) { *c = factories.ClientConfig{} }, "API token is required"},
		"retries may be zero": {func(c *factories.ClientConfig) { c.Retries = 0 }, ""},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := valid()
			tt.mutate(&cfg)

			_, err := factories.CreateClient(cfg)
			if tt.want == "" {
				if err != nil {
					t.Errorf("CreateClient() error = %v, want none", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("CreateClient() error = %v, want it to mention %q", err, tt.want)
			}
		})
	}
}

func TestCreateClientSendsCredentialsAndUserAgentToTheEndpoint(t *testing.T) {
	t.Parallel()

	var auth, agent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, agent = r.Header.Get("Authorization"), r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(srv.Close)

	cfg := valid()
	cfg.CustomEndpoint = srv.URL
	cfg.Version = "1.2.3"
	cfg.Commit = "abc123"

	client, err := factories.CreateClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListUptimeMonitoringLocations(context.Background()).Execute(); err != nil {
		t.Fatalf("request error = %v", err)
	}

	if auth != "Bearer tok" {
		t.Errorf("Authorization = %q, want %q", auth, "Bearer tok")
	}
	if want := "cakectl/1.2.3 (abc123)"; !strings.Contains(agent, want) {
		t.Errorf("User-Agent = %q, want it to contain %q", agent, want)
	}
}

func TestUserAgentFallsBackToDev(t *testing.T) {
	t.Parallel()

	var agent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(srv.Close)

	cfg := valid()
	cfg.CustomEndpoint = srv.URL
	client, _ := factories.CreateClient(cfg)
	_, _ = client.ListUptimeMonitoringLocations(context.Background()).Execute()

	if !strings.Contains(agent, "cakectl/dev") {
		t.Errorf("User-Agent = %q, want it to contain cakectl/dev", agent)
	}
}

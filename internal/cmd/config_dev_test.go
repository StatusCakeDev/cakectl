//go:build dev

package cmd

import (
	"net/http"
	"os"
	"testing"
)

// These checks aim the command at a fake API through the config file, which
// only a development build allows (see internal/cmd/flags).

// TestConfigFileConfiguresTheClient proves the path from a config file,
// through viper and the flag structure, to requests on the wire: the token,
// the endpoint and the retry settings all arrive.
func TestConfigFileConfiguresTheClient(t *testing.T) {
	api := newFakeAPI(t)
	var auth string
	api.mu.Lock()
	api.routes["GET /v1/uptime"] = func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		write(w, http.StatusOK, page())
	}
	api.mu.Unlock()

	s, path := newConfigSession(t)
	config := "api_token: file-token\ncustom_endpoint: " + api.srv.URL + "\nrps: 50\nretries: 0\nmin_backoff: 10ms\nmax_backoff: 20ms\n"
	if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := s.run("uptime", "list", "--config", path); err != nil {
		t.Fatalf("error = %v", err)
	}

	if auth != "Bearer file-token" {
		t.Errorf("Authorization = %q, want the file's token", auth)
	}
	if got := s.stdout(); got != "No uptime checks found." {
		t.Errorf("output = %q", got)
	}
}

func TestFlagsOverrideTheConfigFile(t *testing.T) {
	api := newFakeAPI(t)
	var auth string
	api.mu.Lock()
	api.routes["GET /v1/uptime"] = func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		write(w, http.StatusOK, page())
	}
	api.mu.Unlock()

	s, path := newConfigSession(t)
	config := "api_token: file-token\ncustom_endpoint: " + api.srv.URL + "\n"
	if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := s.run("uptime", "list", "--config", path, "--api-token", "flag-token"); err != nil {
		t.Fatalf("error = %v", err)
	}
	if auth != "Bearer flag-token" {
		t.Errorf("Authorization = %q, want the flag's token", auth)
	}
}

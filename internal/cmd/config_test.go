package cmd

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tomasbasham/cli-runtime/iooption"
)

// resetViper clears the global configuration that commands load into, so that
// one test's config file cannot leak into the next.
func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

func newConfigSession(t *testing.T) (*session, string) {
	t.Helper()

	resetViper(t)
	streams, in, out, errOut := iooption.NewTestIOStreams()
	opts := NewRootOptions(streams)
	opts.isTerminal = func(io.Reader) bool { return false }

	return &session{in: in, out: out, errOut: errOut, opts: opts}, filepath.Join(t.TempDir(), "cakectl.yaml")
}

func TestAuthLoginSavesTheTokenPrivately(t *testing.T) {
	s, path := newConfigSession(t)

	if err := s.run("auth", "login", "--api-token", "secret-token-1234", "--config", path); err != nil {
		t.Fatalf("error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(got), "api_token: secret-token-1234") {
		t.Fatalf("config file = %q, %v; want the token", got, err)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Errorf("config file mode = %o, want 600", info.Mode().Perm())
	}
	if strings.Contains(s.errOut.String()+s.out.String(), "secret-token-1234") {
		t.Errorf("the token was echoed: %q %q", s.out, s.errOut)
	}
}

func TestAuthLoginReadsTheTokenFromStandardInput(t *testing.T) {
	s, path := newConfigSession(t)
	s.in.WriteString("piped-token-5678\n")

	if err := s.run("auth", "login", "--config", path); err != nil {
		t.Fatalf("error = %v", err)
	}

	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "piped-token-5678") {
		t.Errorf("config file = %q, want the piped token", got)
	}
}

func TestAuthLoginWithoutATokenFails(t *testing.T) {
	s, path := newConfigSession(t)

	err := s.run("auth", "login", "--config", path)
	if err == nil || !strings.Contains(err.Error(), "token is required") {
		t.Errorf("error = %v, want a missing token error", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("a failed login created the config file")
	}
}

func TestConfigSetGetUnsetView(t *testing.T) {
	s, path := newConfigSession(t)

	run := func(args ...string) string {
		t.Helper()
		s.out.Reset()
		s.errOut.Reset()
		if err := s.run(append(args, "--config", path)...); err != nil {
			t.Fatalf("%v: error = %v", args, err)
		}
		return s.out.String()
	}

	run("config", "set", "rps", "8")
	run("config", "set", "api_token", "abcdefghijklmnop")

	if got := run("config", "get", "rps"); got != "8\n" {
		t.Errorf("get rps = %q, want %q", got, "8\n")
	}
	if got := run("config", "get", "api_token"); got != "abcdefghijklmnop\n" {
		t.Errorf("get api_token = %q, want it in full", got)
	}

	view := run("config", "view")
	if !strings.Contains(view, "rps: 8") || !strings.Contains(view, "api_token: ****mnop") || strings.Contains(view, "abcdefgh") {
		t.Errorf("view = %q, want the token masked", view)
	}

	run("config", "unset", "rps")
	if view := run("config", "view"); strings.Contains(view, "rps") {
		t.Errorf("view after unset = %q, want rps gone", view)
	}
}

func TestConfigRejectsBadInput(t *testing.T) {
	for _, args := range [][]string{
		{"config", "set", "colour", "red"},
		{"config", "set", "rps", "fast"},
		{"config", "get", "colour"},
		{"config", "get", "rps"}, // Valid key, but not set.
		{"config", "unset", "colour"},
	} {
		s, path := newConfigSession(t)
		if err := s.run(append(args, "--config", path)...); err == nil {
			t.Errorf("%v: error = nil, want one", args)
		}
	}
}

func TestConfigViewOfNoFile(t *testing.T) {
	s, path := newConfigSession(t)

	if err := s.run("config", "view", "--config", path); err != nil {
		t.Fatalf("error = %v", err)
	}
	if s.out.Len() != 0 || !strings.Contains(s.errOut.String(), "No config file") {
		t.Errorf("out = %q, err = %q; want a note on stderr only", s.out, s.errOut)
	}
}

// TestConfigFileConfiguresTheClient proves the path from a config file,
// through viper and the flag structure, to requests on the wire: the token,
// the endpoint and the retry settings all arrive.
func TestConfigFileConfiguresTheClient(t *testing.T) {
	api := newFakeAPI(t)
	var auth string
	api.mu.Lock()
	api.routes["GET /v1/uptime"] = func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		writeJSON(w, http.StatusOK, page())
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
		writeJSON(w, http.StatusOK, page())
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

func TestMissingTokenIsReportedBeforeAnyRequest(t *testing.T) {
	t.Setenv("STATUSCAKE_API_TOKEN", "")
	s, path := newConfigSession(t)
	if err := os.WriteFile(path, []byte("rps: 4\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := s.run("uptime", "list", "--config", path)
	if err == nil || !strings.Contains(err.Error(), "API token is required") || !strings.Contains(err.Error(), "auth login") {
		t.Errorf("error = %v, want advice to log in", err)
	}
}

func TestUnreadableConfigIsReported(t *testing.T) {
	s, path := newConfigSession(t)
	if err := os.WriteFile(path, []byte("api_token: [unclosed\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := s.run("uptime", "list", "--config", path)
	if err == nil || !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("error = %v, want a config error", err)
	}
}

func TestCommandsThatDoNotNeedAnAPITokenDoNotRequireOne(t *testing.T) {
	t.Setenv("STATUSCAKE_API_TOKEN", "")
	s, path := newConfigSession(t)

	if err := s.run("config", "view", "--config", path); err != nil {
		t.Errorf("config view error = %v", err)
	}
	if err := s.run("--help"); err != nil {
		t.Errorf("--help error = %v", err)
	}
}

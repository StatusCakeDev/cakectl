package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tomasbasham/cli-runtime/iooption"
)

// resetViper clears the global configuration that commands load into, so that
// one test's config file cannot leak into the next, and the environment that
// would override it, so that the machine running the tests cannot either.
func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)

	for _, name := range []string{
		"STATUSCAKE_API_TOKEN",
		"STATUSCAKE_CUSTOM_ENDPOINT",
		"STATUSCAKE_RPS",
		"STATUSCAKE_RETRIES",
		"STATUSCAKE_MIN_BACKOFF",
		"STATUSCAKE_MAX_BACKOFF",
	} {
		t.Setenv(name, "")
	}
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

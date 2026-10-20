package configstore_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/StatusCakeDev/cakectl/internal/configstore"
)

func newStore(t *testing.T) *configstore.Store {
	t.Helper()
	return configstore.New(filepath.Join(t.TempDir(), "nested", "cakectl.yaml"))
}

func TestMissingFileHasNoSettings(t *testing.T) {
	t.Parallel()

	s := newStore(t)

	all, err := s.All()
	if err != nil || len(all) != 0 {
		t.Errorf("All() = %v, %v; want nothing", all, err)
	}
	if _, err := s.Get("api_token"); !errors.Is(err, configstore.ErrNotSet) {
		t.Errorf("Get() error = %v, want %v", err, configstore.ErrNotSet)
	}
}

func TestSetStoresValuesWithTheirTypes(t *testing.T) {
	t.Parallel()

	s := newStore(t)
	for key, value := range map[string]string{
		"api_token":   "tok",
		"rps":         "8",
		"min_backoff": "500ms",
	} {
		if err := s.Set(key, value); err != nil {
			t.Fatalf("Set(%s) error = %v", key, err)
		}
	}

	all, err := s.All()
	if err != nil {
		t.Fatal(err)
	}
	// A number is stored as a number, not as the string typed.
	want := map[string]any{"api_token": "tok", "rps": uint64(8), "min_backoff": "500ms"}
	if diff := cmp.Diff(want, all); diff != "" {
		t.Errorf("All() mismatch (-want +got):\n%s", diff)
	}
}

func TestSetRejectsBadInput(t *testing.T) {
	t.Parallel()

	s := newStore(t)

	if err := s.Set("colour", "red"); !errors.Is(err, configstore.ErrUnknownKey) {
		t.Errorf("Set(unknown) error = %v, want %v", err, configstore.ErrUnknownKey)
	}
	if err := s.Set("rps", "fast"); err == nil {
		t.Error("Set(rps, fast) error = nil, want an error")
	}
	if err := s.Set("max_backoff", "soon"); err == nil {
		t.Error("Set(max_backoff, soon) error = nil, want an error")
	}
	if _, err := os.Stat(s.Path()); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a rejected Set created the file: %v", err)
	}
}

func TestUnset(t *testing.T) {
	t.Parallel()

	s := newStore(t)
	if err := s.Set("api_token", "tok"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("custom_endpoint", "https://example.com"); err != nil {
		t.Fatal(err)
	}

	if err := s.Unset("api_token"); err != nil {
		t.Fatalf("Unset() error = %v", err)
	}
	if err := s.Unset("api_token"); err != nil {
		t.Errorf("Unset() of an unset key error = %v, want none", err)
	}
	if err := s.Unset("colour"); !errors.Is(err, configstore.ErrUnknownKey) {
		t.Errorf("Unset(unknown) error = %v, want %v", err, configstore.ErrUnknownKey)
	}

	all, _ := s.All()
	want := map[string]any{"custom_endpoint": "https://example.com"}
	if diff := cmp.Diff(want, all); diff != "" {
		t.Errorf("All() mismatch (-want +got):\n%s", diff)
	}
}

func TestFileIsPrivate(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("file modes are not meaningful on Windows")
	}

	s := newStore(t)
	if err := s.Set("api_token", "tok"); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(s.Path())
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("file mode = %o, want 600", mode)
	}
	dir, _ := os.Stat(filepath.Dir(s.Path()))
	if mode := dir.Mode().Perm(); mode != 0o700 {
		t.Errorf("directory mode = %o, want 700", mode)
	}
}

func TestCorruptFileIsReportedNotOverwritten(t *testing.T) {
	t.Parallel()

	s := newStore(t)
	if err := os.MkdirAll(filepath.Dir(s.Path()), 0o700); err != nil {
		t.Fatal(err)
	}
	corrupt := []byte("api_token: [unclosed\n")
	if err := os.WriteFile(s.Path(), corrupt, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := s.Set("rps", "4"); err == nil {
		t.Error("Set() error = nil, want a parse error")
	}
	if got, _ := os.ReadFile(s.Path()); string(got) != string(corrupt) {
		t.Errorf("file was changed to %q", got)
	}
}

func TestMask(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"":                 "****",
		"short":            "****",
		"abcdefghijklmnop": "****mnop",
	}
	for in, want := range tests {
		if got := configstore.Mask(in); got != want {
			t.Errorf("Mask(%q) = %q, want %q", in, got, want)
		}
	}
	if !configstore.Secret("api_token") || configstore.Secret("rps") {
		t.Error("Secret() does not identify only the API token")
	}
}

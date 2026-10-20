package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestOSFS_Open(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "check.json")
	if err := os.WriteFile(path, []byte(`{"name":"example"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	// These are the forms of path that os.DirFS rejects but which users of the
	// command line expect to work.
	tests := map[string]string{
		"absolute path":            path,
		"path with parent element": filepath.Join(dir, "..", filepath.Base(dir), "check.json"),
	}

	for name, p := range tests {
		t.Run(name, func(t *testing.T) {
			f, err := osFS{}.Open(p)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer f.Close()

			b, err := io.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}

			if got, want := string(b), `{"name":"example"}`; got != want {
				t.Errorf("contents = %q, want %q", got, want)
			}
		})
	}

	t.Run("missing file returns an error", func(t *testing.T) {
		if _, err := (osFS{}).Open(filepath.Join(dir, "missing.json")); err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

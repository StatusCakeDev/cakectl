package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tomasbasham/cli-runtime/iooption"

	"github.com/StatusCakeDev/cakectl/internal/adapter"
)

// Command tests run the real command tree against a fake StatusCake API
// through the real SDK, so they see what a user's invocation would send.
//
// They are not parallel: building a command binds its flags to viper's global
// instance, which is not safe for concurrent use.

type call struct {
	Method string
	Path   string
	Query  url.Values
	Form   url.Values
}

type fakeAPI struct {
	mu     sync.Mutex
	routes map[string]http.HandlerFunc
	seen   []call
	srv    *httptest.Server
}

func newFakeAPI(t *testing.T) *fakeAPI {
	t.Helper()

	api := &fakeAPI{routes: map[string]http.HandlerFunc{}}
	api.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()

		api.mu.Lock()
		api.seen = append(api.seen, call{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(), Form: r.PostForm})
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			write(w, http.StatusNotFound, map[string]any{"message": "not found"})
		})
		for pattern, h := range api.routes {
			mux.HandleFunc(pattern, h)
		}
		api.mu.Unlock()

		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(api.srv.Close)
	return api
}

func (a *fakeAPI) serve(pattern string, status int, body any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.routes[pattern] = func(w http.ResponseWriter, _ *http.Request) { write(w, status, body) }
}

func (a *fakeAPI) calls() []call {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]call(nil), a.seen...)
}

func write(w http.ResponseWriter, status int, body any) {
	if body == nil {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// session is one invocation of the command tree.
type session struct {
	api    *fakeAPI
	in     *bytes.Buffer
	out    *bytes.Buffer
	errOut *bytes.Buffer
	opts   *RootOptions
}

// newSession prepares a command tree whose API client talks to the fake API.
func newSession(t *testing.T, api *fakeAPI) *session {
	t.Helper()

	streams, in, out, errOut := iooption.NewTestIOStreams()
	opts := NewRootOptions(streams)

	client, err := adapter.New(slog.New(slog.DiscardHandler), adapter.Config{
		Token:      "token",
		RPS:        1000,
		Retries:    0,
		MinBackoff: time.Millisecond,
		MaxBackoff: time.Millisecond,
		Endpoint:   api.srv.URL,
	})
	if err != nil {
		t.Fatalf("adapters.New() error = %v", err)
	}

	opts.api = client
	// Input is a buffer, not a terminal, unless a test says otherwise.
	opts.isTerminal = func(io.Reader) bool { return false }

	return &session{
		api:    api,
		in:     in,
		out:    out,
		errOut: errOut,
		opts:   opts,
	}
}

// run executes the command line and returns its error.
func (s *session) run(args ...string) error {
	cmd := NewRootCommandWithArgs(s.opts)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// stdout returns what the command printed, trimmed of the final newline.
func (s *session) stdout() string {
	return strings.TrimSuffix(s.out.String(), "\n")
}

package spec_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/spec"
)

func TestDecodeReadsYAMLAndJSONIdentically(t *testing.T) {
	t.Parallel()

	want := spec.Uptime{
		Name:          gofp.Some("Home page"),
		TestType:      gofp.Some("HTTP"),
		WebsiteURL:    gofp.Some("https://example.com"),
		CheckRate:     gofp.Some(int32(300)),
		ContactGroups: gofp.Some([]string{"1", "2"}),
		Paused:        gofp.Some(false),
	}

	tests := map[string]string{
		"yaml": `
name: Home page
test_type: HTTP
website_url: https://example.com
check_rate: 300
contact_groups: ["1", "2"]
paused: false
`,
		"json": `{"name":"Home page","test_type":"HTTP","website_url":"https://example.com","check_rate":300,"contact_groups":["1","2"],"paused":false}`,
	}
	for name, doc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := spec.Decode[spec.Uptime](strings.NewReader(doc))
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if diff := cmp.Diff(want, got, gofpOptions); diff != "" {
				t.Errorf("Decode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecodeDistinguishesAbsentFromEmpty(t *testing.T) {
	t.Parallel()

	got, err := spec.Decode[spec.Heartbeat](strings.NewReader("tags: []\nhost: \"\"\n"))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	want := spec.Heartbeat{
		Tags: gofp.Some([]string{}),
		Host: gofp.Some(""),
	}
	if diff := cmp.Diff(want, got, gofpOptions); diff != "" {
		t.Errorf("Decode() mismatch (-want +got):\n%s", diff)
	}
}

func TestDecodeTreatsNullAsAbsent(t *testing.T) {
	t.Parallel()

	got, err := spec.Decode[spec.Heartbeat](strings.NewReader("name: null\nperiod: 60\n"))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got.Name.IsSome() {
		t.Errorf("Name = %v, want absent", got.Name)
	}
}

func TestDecodeReadsTimestamps(t *testing.T) {
	t.Parallel()

	doc := `
name: Deploy
start_at: 2026-11-01T22:00:00Z
end_at: "2026-11-01T23:30:00+01:00"
timezone: Europe/London
`
	got, err := spec.Decode[spec.Maintenance](strings.NewReader(doc))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	start := time.Date(2026, 11, 1, 22, 0, 0, 0, time.UTC)
	if s, _ := got.StartAt.TryUnwrap(); !s.Equal(start) {
		t.Errorf("StartAt = %v, want %v", s, start)
	}
	end := time.Date(2026, 11, 1, 22, 30, 0, 0, time.UTC)
	if e, _ := got.EndAt.TryUnwrap(); !e.Equal(end) {
		t.Errorf("EndAt = %v, want %v", e, end)
	}
}

func TestDecodeAcceptsDescribeEnvelope(t *testing.T) {
	t.Parallel()

	doc := `
id: "123"
spec:
  name: Heartbeat
  period: 600
status:
  status: up
  uptime: 99.9
`
	got, err := spec.Decode[spec.Heartbeat](strings.NewReader(doc))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	want := spec.Heartbeat{Name: gofp.Some("Heartbeat"), Period: gofp.Some(int32(600))}
	if diff := cmp.Diff(want, got, gofpOptions); diff != "" {
		t.Errorf("Decode() mismatch (-want +got):\n%s", diff)
	}
}

func TestDecodeRejectsBadDocuments(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		doc  string
		want string
	}{
		"unknown field": {
			doc:  "name: x\nperiodd: 60\n",
			want: "periodd",
		},
		"unknown field in spec": {
			doc:  "spec:\n  nme: x\n",
			want: "nme",
		},
		"unknown envelope key": {
			doc:  "spec:\n  name: x\nextra: 1\n",
			want: `unexpected key "extra"`,
		},
		"wrong type": {
			doc:  "period: soon\n",
			want: "period",
		},
		"not a mapping": {
			doc:  "- a\n- b\n",
			want: "mapping",
		},
		"scalar document": {
			doc:  "hello\n",
			want: "mapping",
		},
		"malformed": {
			doc:  "name: [unclosed\n",
			want: "parse specification",
		},
		"fractional integer field": {
			doc:  "period: 1.5\n",
			want: "period",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := spec.Decode[spec.Heartbeat](strings.NewReader(tt.doc))
			if err == nil {
				t.Fatalf("Decode(%q) error = nil, want error mentioning %q", tt.doc, tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Decode(%q) error = %q, want it to mention %q", tt.doc, err, tt.want)
			}
		})
	}
}

func TestDecodeRejectsEmptyDocuments(t *testing.T) {
	t.Parallel()

	for _, doc := range []string{"", "  \n\t\n"} {
		_, err := spec.Decode[spec.Heartbeat](strings.NewReader(doc))
		if !errors.Is(err, spec.ErrEmpty) {
			t.Errorf("Decode(%q) error = %v, want %v", doc, err, spec.ErrEmpty)
		}
	}
}

func TestDecodeRejectsOversizedDocuments(t *testing.T) {
	t.Parallel()

	doc := "name: " + strings.Repeat("a", 2<<20)
	_, err := spec.Decode[spec.Heartbeat](strings.NewReader(doc))
	if err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Errorf("Decode() error = %v, want a size error", err)
	}
}

func TestDecodeReportsReadFailures(t *testing.T) {
	t.Parallel()

	_, err := spec.Decode[spec.Heartbeat](errReader{})
	if err == nil || !strings.Contains(err.Error(), "read specification") {
		t.Errorf("Decode() error = %v, want a read error", err)
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	if !spec.IsEmpty(spec.Uptime{}) {
		t.Error("IsEmpty(zero spec) = false, want true")
	}
	if spec.IsEmpty(spec.Uptime{Paused: gofp.Some(false)}) {
		t.Error("IsEmpty(spec with a field set to its zero value) = true, want false")
	}
	if spec.IsEmpty(spec.Uptime{Tags: gofp.Some([]string(nil))}) {
		t.Error("IsEmpty(spec clearing tags) = true, want false")
	}
}

// gofpOptions lets go-cmp look inside [gofp.Option], whose fields are
// unexported and whose payload may be a slice, so it cannot be compared with
// ==.
var gofpOptions = cmp.Exporter(func(t reflect.Type) bool {
	return strings.HasPrefix(t.PkgPath(), "github.com/tomasbasham/gofp")
})

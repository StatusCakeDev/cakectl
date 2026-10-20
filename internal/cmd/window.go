package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/adapters"
)

// windowFlags registers the flags that select a slice of history, and returns
// a function that reads them once parsed.
func windowFlags(fs *pflag.FlagSet) func() (adapters.Window, error) {
	limit := fs.Int("limit", 25, "Maximum number of results")
	before := fs.String("before", "", "Only results before this time (RFC 3339)")
	after := fs.String("after", "", "Only results after this time (RFC 3339)")

	return func() (adapters.Window, error) {
		w := adapters.Window{Limit: *limit}
		if *limit < 0 {
			return w, fmt.Errorf("invalid --limit %d: must not be negative", *limit)
		}

		var err error
		if w.Before, err = parseInstant("before", *before); err != nil {
			return w, err
		}
		if w.After, err = parseInstant("after", *after); err != nil {
			return w, err
		}
		return w, nil
	}
}

func parseInstant(flag, value string) (gofp.Option[time.Time], error) {
	if value == "" {
		return gofp.None[time.Time](), nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return gofp.None[time.Time](), fmt.Errorf("invalid --%s %q: expected an RFC 3339 time such as 2026-11-01T22:00:00Z", flag, value)
	}
	return gofp.Some(t), nil
}

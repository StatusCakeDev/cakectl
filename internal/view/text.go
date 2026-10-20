// Package view turns what the API returns into what the user sees.
//
// Each type is a read model: a flat, stable shape with JSON field names, that
// can be rendered as a table or block of fields for people
// ([printer.TextFormatter]) and encoded as JSON or YAML for scripts. A view
// never includes more than its command needs: lists omit secrets and personal
// data such as heartbeat URLs and contact details, which only an explicit
// `describe` shows.
//
// The description of a resource carries its writable fields as a [spec], in
// the same shape that `-f` reads, so that a description can be edited and
// applied back.
package view

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/tomasbasham/gofp"
)

// none is shown in place of a value that is not set.
const none = "-"

// field is one labelled value of a description.
type field struct {
	label string
	value string
}

// fields renders labelled values as aligned lines.
func fields(fs ...field) []byte {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 8, 2, ' ', 0)
	for _, f := range fs {
		fmt.Fprintf(w, "%s:\t%s\n", f.label, f.value)
	}
	_ = w.Flush()
	return buf.Bytes()
}

// table renders rows under headers, or the empty message when there are none.
func table(headers []string, rows [][]string, empty string) []byte {
	if len(rows) == 0 {
		return []byte(empty + "\n")
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 8, 3, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	_ = w.Flush()
	return buf.Bytes()
}

// show renders an optional value, or a dash when it is absent.
func show[T any](o gofp.Option[T]) string {
	v, ok := o.TryUnwrap()
	if !ok {
		return none
	}
	return format(v)
}

// format renders a value for a person.
func format(v any) string {
	switch x := v.(type) {
	case string:
		if x == "" {
			return none
		}
		return x
	case []string:
		return join(x)
	case []int32:
		parts := make([]string, len(x))
		for i, n := range x {
			parts[i] = fmt.Sprint(n)
		}
		return join(parts)
	case time.Time:
		return x.Format(time.RFC3339)
	default:
		return fmt.Sprint(v)
	}
}

func join(values []string) string {
	if len(values) == 0 {
		return none
	}
	return strings.Join(values, ", ")
}

// text renders a possibly absent string.
func text(s *string) string {
	if s == nil || *s == "" {
		return none
	}
	return *s
}

// instant renders a possibly absent time.
func instant(t *time.Time) string {
	if t == nil {
		return none
	}
	return t.Format(time.RFC3339)
}

// interval renders a number of seconds as a duration such as "5m0s".
func interval(seconds int32) string {
	if seconds == 0 {
		return "continuous"
	}
	return (time.Duration(seconds) * time.Second).String()
}

// percent renders an uptime percentage.
func percent(p *float32) string {
	if p == nil {
		return none
	}
	return fmt.Sprintf("%.2f%%", *p)
}

// status folds the paused flag into a status, since a paused check reports no
// useful health.
func status(paused bool, s string) string {
	if paused {
		return "paused"
	}
	return s
}

// someIf is the option holding v when it is present.
func someIf[T any](v *T) gofp.Option[T] {
	if v == nil {
		return gofp.None[T]()
	}
	return gofp.Some(*v)
}

// list returns a slice that is never nil, so that it is encoded as an empty
// list rather than null.
func list[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

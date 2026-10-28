package spec

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/tomasbasham/gofp"
)

// validator accumulates the problems found in a spec so that all of them can be
// reported together, rather than one per attempt.
type validator struct {
	errs []error
}

func (v *validator) err() error {
	return errors.Join(v.errs...)
}

func (v *validator) fail(format string, args ...any) {
	v.errs = append(v.errs, fmt.Errorf(format, args...))
}

// required records a problem when the option is absent.
func required[T any](v *validator, name string, o gofp.Option[T]) {
	if o.IsNone() {
		v.fail("%s is required", name)
	}
}

// oneOf records a problem when the option is present and not among th allowed
// values.
func oneOf[T comparable](v *validator, name string, o gofp.Option[T], allowed []T) {
	value, ok := o.TryUnwrap()
	if !ok || slices.Contains(allowed, value) {
		return
	}

	names := make([]string, len(allowed))
	for i, a := range allowed {
		names[i] = fmt.Sprint(a)
	}
	v.fail("%s must be one of %s, got %v", name, strings.Join(names, ", "), value)
}

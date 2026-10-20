package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/tomasbasham/gofp"
)

// binder connects command line flags to the fields of a spec.
//
// Each flag is registered with the flag set, and remembers which field it
// overrides. After parsing, [binder.apply] sets exactly the fields whose flags
// the user passed, so a flag left alone never overwrites what a file said and
// never clears a field on update.
type binder[S any] struct {
	flags    *pflag.FlagSet
	override []func(*S) error
}

func newBinder[S any](flags *pflag.FlagSet) *binder[S] {
	return &binder[S]{flags: flags}
}

// apply sets the fields whose flags were passed.
func (b *binder[S]) apply(s *S) error {
	for _, f := range b.override {
		if err := f(s); err != nil {
			return err
		}
	}
	return nil
}

func bindString[S any](b *binder[S], name, usage string, field func(*S) *gofp.Option[string]) {
	v := b.flags.String(name, "", usage)
	b.override = append(b.override, func(s *S) error {
		if b.flags.Changed(name) {
			*field(s) = gofp.Some(*v)
		}
		return nil
	})
}

func bindBool[S any](b *binder[S], name, usage string, field func(*S) *gofp.Option[bool]) {
	v := b.flags.Bool(name, false, usage)
	b.override = append(b.override, func(s *S) error {
		if b.flags.Changed(name) {
			*field(s) = gofp.Some(*v)
		}
		return nil
	})
}

func bindInt32[S any](b *binder[S], name, usage string, field func(*S) *gofp.Option[int32]) {
	v := b.flags.Int32(name, 0, usage)
	b.override = append(b.override, func(s *S) error {
		if b.flags.Changed(name) {
			*field(s) = gofp.Some(*v)
		}
		return nil
	})
}

func bindInt64[S any](b *binder[S], name, usage string, field func(*S) *gofp.Option[int64]) {
	v := b.flags.Int64(name, 0, usage)
	b.override = append(b.override, func(s *S) error {
		if b.flags.Changed(name) {
			*field(s) = gofp.Some(*v)
		}
		return nil
	})
}

func bindStrings[S any](b *binder[S], name, usage string, field func(*S) *gofp.Option[[]string]) {
	v := b.flags.StringSlice(name, nil, usage+" (comma separated; pass an empty value to clear)")
	b.override = append(b.override, func(s *S) error {
		if b.flags.Changed(name) {
			// Copy, so that the spec does not alias the flag's storage.
			*field(s) = gofp.Some(append([]string{}, *v...))
		}
		return nil
	})
}

func bindInt32s[S any](b *binder[S], name, usage string, field func(*S) *gofp.Option[[]int32]) {
	v := b.flags.Int32Slice(name, nil, usage+" (comma separated)")
	b.override = append(b.override, func(s *S) error {
		if b.flags.Changed(name) {
			*field(s) = gofp.Some(append([]int32{}, *v...))
		}
		return nil
	})
}

// bindTime registers a flag holding an RFC 3339 time.
func bindTime[S any](b *binder[S], name, usage string, field func(*S) *gofp.Option[time.Time]) {
	v := b.flags.String(name, "", usage+" (RFC 3339, e.g. 2026-11-01T22:00:00Z)")
	b.override = append(b.override, func(s *S) error {
		if !b.flags.Changed(name) {
			return nil
		}
		t, err := time.Parse(time.RFC3339, *v)
		if err != nil {
			return fmt.Errorf("invalid --%s %q: expected an RFC 3339 time such as 2026-11-01T22:00:00Z", name, *v)
		}
		*field(s) = gofp.Some(t)
		return nil
	})
}

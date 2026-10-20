package spec_test

import (
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
)

// gofpOptions lets go-cmp look inside [gofp.Option], whose fields are
// unexported and whose payload may be a slice, so it cannot be compared with
// ==.
var gofpOptions = cmp.Exporter(func(t reflect.Type) bool {
	return strings.HasPrefix(t.PkgPath(), "github.com/tomasbasham/gofp")
})

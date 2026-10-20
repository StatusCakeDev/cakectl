// Package spec describes the desired state of each StatusCake resource.
//
// A spec is what a person asks for: the fields they may set when creating or
// updating a resource. Every field is a [gofp.Option] so that a spec can say
// "leave this alone" as well as "set this to the zero value", which is what
// makes partial updates and flag overrides possible. A spec holds no knowledge
// of the StatusCake SDK, of files or of flags; those are the concern of the
// layers around it.
package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"

	"github.com/goccy/go-yaml"
)

// maxSize bounds how much of a specification is read, so that pointing the
// command at the wrong file or stream cannot exhaust memory.
const maxSize = 1 << 20

// ErrEmpty is returned when a specification contains no content at all.
var ErrEmpty = errors.New("specification is empty")

// envelopeKeys are the keys, besides "spec", that may accompany a spec in a
// document. They are what `describe` prints alongside the spec, so allowing
// them means its output can be edited and fed straight back in.
var envelopeKeys = []string{"id", "status"}

// Decode reads a single specification from r.
//
// The document may be JSON or YAML. It is decoded strictly: a field the spec
// does not define is an error rather than being silently ignored, because a
// misspelt field would otherwise leave a resource configured differently from
// what the author intended. The fields may sit at the top level, or under a
// "spec" key alongside the "id" and "status" keys that `describe` prints.
func Decode[T any](r io.Reader) (T, error) {
	var zero T

	data, err := io.ReadAll(io.LimitReader(r, maxSize+1))
	if err != nil {
		return zero, fmt.Errorf("read specification: %w", err)
	}
	if len(data) > maxSize {
		return zero, fmt.Errorf("specification is larger than %d bytes", maxSize)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return zero, ErrEmpty
	}

	// JSON is a subset of YAML, so one parser serves both formats. Converting to
	// JSON lets the standard decoder enforce strictness and honour the Option
	// type's own unmarshalling.
	doc, err := yaml.YAMLToJSON(data)
	if err != nil {
		return zero, fmt.Errorf("parse specification: %w", err)
	}

	doc, err = unwrap(doc)
	if err != nil {
		return zero, err
	}

	dec := json.NewDecoder(bytes.NewReader(doc))
	dec.DisallowUnknownFields()

	var v T
	if err := dec.Decode(&v); err != nil {
		return zero, fmt.Errorf("invalid specification: %w", err)
	}
	return v, nil
}

// unwrap returns the content of the "spec" key when the document is wrapped in
// an envelope, and the document unchanged otherwise.
func unwrap(doc []byte) ([]byte, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(doc, &top); err != nil {
		return nil, fmt.Errorf("invalid specification: expected a mapping of fields: %w", err)
	}

	inner, ok := top["spec"]
	if !ok {
		return doc, nil
	}

	for key := range top {
		if key != "spec" && !slices.Contains(envelopeKeys, key) {
			return nil, fmt.Errorf("invalid specification: unexpected key %q beside \"spec\"", key)
		}
	}
	return inner, nil
}

// IsEmpty reports whether no field of the spec has been set.
func IsEmpty[T any](spec T) bool {
	return reflect.ValueOf(spec).IsZero()
}

package view

import (
	"encoding/json"

	"github.com/goccy/go-yaml"
)

// marshalYAML encodes a view as YAML by way of JSON.
//
// The views are shaped by their JSON field names and use [gofp.Option] for the
// fields of a spec, neither of which the YAML encoder understands directly.
// Going through JSON keeps one definition of every field's name and presence,
// and the ordered map keeps the fields in the order they are declared rather
// than alphabetically.
func marshalYAML(v any) ([]byte, error) {
	doc, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var ordered any
	if err := yaml.UnmarshalWithOptions(doc, &ordered, yaml.UseOrderedMap()); err != nil {
		return nil, err
	}
	return yaml.Marshal(ordered)
}

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v UptimeList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v UptimeDescription) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v HeartbeatList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v HeartbeatDescription) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v PagespeedList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v PagespeedDescription) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v SSLList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v SSLDescription) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v ContactGroupList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v ContactGroupDescription) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v MaintenanceList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v MaintenanceDescription) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v AlertList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v HistoryList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v PeriodList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v PagespeedResultList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v LocationList) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

// MarshalYAML encodes the view as YAML, with the same fields as its JSON.
func (v Outcome) MarshalYAML() ([]byte, error) { return marshalYAML(v) }

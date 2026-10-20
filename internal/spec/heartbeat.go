package spec

import "github.com/tomasbasham/gofp"

// Heartbeat is the desired state of a heartbeat check. Period is the number of
// seconds the check waits for a ping before raising an alert.
type Heartbeat struct {
	Name          gofp.Option[string]   `json:"name,omitzero"`
	Period        gofp.Option[int32]    `json:"period,omitzero"`
	ContactGroups gofp.Option[[]string] `json:"contact_groups,omitzero"`
	Host          gofp.Option[string]   `json:"host,omitzero"`
	Paused        gofp.Option[bool]     `json:"paused,omitzero"`
	Tags          gofp.Option[[]string] `json:"tags,omitzero"`
}

// ValidateCreate reports every reason the spec cannot create a check.
func (h Heartbeat) ValidateCreate() error {
	var v validator
	required(&v, "name", h.Name)
	required(&v, "period", h.Period)
	h.validateValues(&v)
	return v.err()
}

// ValidateUpdate reports every reason the spec cannot update a check.
func (h Heartbeat) ValidateUpdate() error {
	var v validator
	h.validateValues(&v)
	return v.err()
}

func (h Heartbeat) validateValues(v *validator) {
	if p, ok := h.Period.TryUnwrap(); ok && p <= 0 {
		v.fail("period must be greater than zero, got %d", p)
	}
}

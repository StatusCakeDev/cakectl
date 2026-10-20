package spec

import (
	"time"

	"github.com/tomasbasham/gofp"
)

// MaintenanceRepeatIntervals are how often a maintenance window can recur.
var MaintenanceRepeatIntervals = []string{"never", "1d", "1w", "2w", "1m"}

// Maintenance is the desired state of a maintenance window. Times are RFC 3339
// and are interpreted in the window's time zone, an IANA name such as
// "Europe/London".
type Maintenance struct {
	Name           gofp.Option[string]    `json:"name,omitzero"`
	StartAt        gofp.Option[time.Time] `json:"start_at,omitzero"`
	EndAt          gofp.Option[time.Time] `json:"end_at,omitzero"`
	Timezone       gofp.Option[string]    `json:"timezone,omitzero"`
	RepeatInterval gofp.Option[string]    `json:"repeat_interval,omitzero"`
	Tags           gofp.Option[[]string]  `json:"tags,omitzero"`
	Tests          gofp.Option[[]string]  `json:"tests,omitzero"`
}

// ValidateCreate reports every reason the spec cannot create a window.
func (m Maintenance) ValidateCreate() error {
	var v validator
	required(&v, "name", m.Name)
	required(&v, "start_at", m.StartAt)
	required(&v, "end_at", m.EndAt)
	required(&v, "timezone", m.Timezone)
	m.validateValues(&v)
	return v.err()
}

// ValidateUpdate reports every reason the spec cannot update a window.
func (m Maintenance) ValidateUpdate() error {
	var v validator
	m.validateValues(&v)
	return v.err()
}

func (m Maintenance) validateValues(v *validator) {
	oneOf(v, "repeat_interval", m.RepeatInterval, MaintenanceRepeatIntervals)

	start, hasStart := m.StartAt.TryUnwrap()
	end, hasEnd := m.EndAt.TryUnwrap()
	if hasStart && hasEnd && !end.After(start) {
		v.fail("end_at must be after start_at")
	}
}

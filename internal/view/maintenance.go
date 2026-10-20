package view

import (
	"time"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// MaintenanceRow is a maintenance window as shown in a listing.
type MaintenanceRow struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	StartAt        time.Time `json:"start_at"`
	EndAt          time.Time `json:"end_at"`
	Timezone       string    `json:"timezone"`
	RepeatInterval string    `json:"repeat_interval"`
	State          string    `json:"state"`
}

// MaintenanceList is a listing of maintenance windows.
type MaintenanceList []MaintenanceRow

// NewMaintenanceList provides the listing of the given windows.
func NewMaintenanceList(windows []statuscake.MaintenanceWindow) MaintenanceList {
	rows := make(MaintenanceList, len(windows))
	for i, w := range windows {
		rows[i] = MaintenanceRow{
			ID: w.ID, Name: w.Name, StartAt: w.Start, EndAt: w.End, Timezone: w.Timezone,
			RepeatInterval: string(w.RepeatInterval), State: string(w.State),
		}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l MaintenanceList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		rows[i] = []string{
			r.ID, r.Name, r.StartAt.Format(time.RFC3339), r.EndAt.Format(time.RFC3339),
			r.RepeatInterval, r.State,
		}
	}
	return table([]string{"ID", "NAME", "START", "END", "REPEAT", "STATE"}, rows, "No maintenance windows found."), nil
}

// MaintenanceStatus is what the API reports about a maintenance window.
type MaintenanceStatus struct {
	State string `json:"state"`
}

// MaintenanceDescription is the full detail of a maintenance window.
type MaintenanceDescription struct {
	ID     string            `json:"id"`
	Spec   spec.Maintenance  `json:"spec"`
	Status MaintenanceStatus `json:"status"`
}

// NewMaintenanceDescription describes a maintenance window.
func NewMaintenanceDescription(w statuscake.MaintenanceWindow) MaintenanceDescription {
	return MaintenanceDescription{
		ID: w.ID,
		Spec: spec.Maintenance{
			Name:           gofp.Some(w.Name),
			StartAt:        gofp.Some(w.Start),
			EndAt:          gofp.Some(w.End),
			Timezone:       gofp.Some(w.Timezone),
			RepeatInterval: gofp.Some(string(w.RepeatInterval)),
			Tags:           gofp.Some(list(w.Tags)),
			Tests:          gofp.Some(list(w.Tests)),
		},
		Status: MaintenanceStatus{State: string(w.State)},
	}
}

// FormatText renders the description as labelled fields.
func (d MaintenanceDescription) FormatText() ([]byte, error) {
	s := d.Spec
	return fields(
		field{"ID", d.ID},
		field{"Name", show(s.Name)},
		field{"State", d.Status.State},
		field{"Start", show(s.StartAt)},
		field{"End", show(s.EndAt)},
		field{"Time zone", show(s.Timezone)},
		field{"Repeat", show(s.RepeatInterval)},
		field{"Checks", show(s.Tests)},
		field{"Tags", show(s.Tags)},
	), nil
}

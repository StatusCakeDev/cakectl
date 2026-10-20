package view

import (
	"time"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// HeartbeatRow is a heartbeat check as shown in a listing. The ping URL
// contains a secret token and is left out.
type HeartbeatRow struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Period        int32    `json:"period"`
	Status        string   `json:"status"`
	Paused        bool     `json:"paused"`
	Uptime        *float32 `json:"uptime,omitempty"`
	ContactGroups []string `json:"contact_groups"`
	Tags          []string `json:"tags"`
}

// HeartbeatList is a listing of heartbeat checks.
type HeartbeatList []HeartbeatRow

// NewHeartbeatList provides the listing of the given checks.
func NewHeartbeatList(tests []statuscake.HeartbeatTestOverview) HeartbeatList {
	rows := make(HeartbeatList, len(tests))
	for i, t := range tests {
		rows[i] = HeartbeatRow{
			ID: t.ID, Name: t.Name, Period: t.Period, Status: string(t.Status), Paused: t.Paused,
			Uptime: t.Uptime, ContactGroups: list(t.ContactGroups), Tags: list(t.Tags),
		}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l HeartbeatList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		rows[i] = []string{r.ID, r.Name, interval(r.Period), status(r.Paused, r.Status)}
	}
	return table([]string{"ID", "NAME", "PERIOD", "STATUS"}, rows, "No heartbeat checks found."), nil
}

// HeartbeatStatus is what the API reports about a heartbeat check. The URL is
// the address the monitored job must ping; it holds a secret token.
type HeartbeatStatus struct {
	URL          string     `json:"url"`
	Status       string     `json:"status"`
	Uptime       float32    `json:"uptime"`
	LastTestedAt *time.Time `json:"last_tested_at,omitempty"`
}

// HeartbeatDescription is the full detail of a heartbeat check.
type HeartbeatDescription struct {
	ID     string          `json:"id"`
	Spec   spec.Heartbeat  `json:"spec"`
	Status HeartbeatStatus `json:"status"`
}

// NewHeartbeatDescription describes a heartbeat check.
func NewHeartbeatDescription(t statuscake.HeartbeatTest) HeartbeatDescription {
	return HeartbeatDescription{
		ID: t.ID,
		Spec: spec.Heartbeat{
			Name:          gofp.Some(t.Name),
			Period:        gofp.Some(t.Period),
			ContactGroups: gofp.Some(list(t.ContactGroups)),
			Host:          someIf(t.Host),
			Paused:        gofp.Some(t.Paused),
			Tags:          gofp.Some(list(t.Tags)),
		},
		Status: HeartbeatStatus{URL: t.WebsiteURL, Status: string(t.Status), Uptime: t.Uptime, LastTestedAt: t.LastTested},
	}
}

// FormatText renders the description as labelled fields.
func (d HeartbeatDescription) FormatText() ([]byte, error) {
	s := d.Spec
	return fields(
		field{"ID", d.ID},
		field{"Name", show(s.Name)},
		field{"Ping URL", format(d.Status.URL)},
		field{"Period", interval(s.Period.UnwrapOr(0))},
		field{"Status", status(s.Paused.UnwrapOr(false), d.Status.Status)},
		field{"Uptime", percent(&d.Status.Uptime)},
		field{"Last tested", instant(d.Status.LastTestedAt)},
		field{"Contact groups", show(s.ContactGroups)},
		field{"Tags", show(s.Tags)},
		field{"Host", show(s.Host)},
	), nil
}

package view

import (
	"slices"
	"time"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/domain"
	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// UptimeRow is an uptime check as shown in a listing.
type UptimeRow struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	TestType      string   `json:"test_type"`
	WebsiteURL    string   `json:"website_url"`
	CheckRate     int32    `json:"check_rate"`
	Status        string   `json:"status"`
	Paused        bool     `json:"paused"`
	Uptime        *float32 `json:"uptime,omitempty"`
	ContactGroups []string `json:"contact_groups"`
	Tags          []string `json:"tags"`
}

// UptimeList is a listing of uptime checks.
type UptimeList []UptimeRow

// NewUptimeList provides the listing of the given checks.
func NewUptimeList(tests []domain.UptimeOverview) UptimeList {
	rows := make(UptimeList, len(tests))
	for i, t := range tests {
		rows[i] = UptimeRow{
			ID: t.ID, Name: t.Name, TestType: string(t.TestType), WebsiteURL: t.WebsiteURL,
			CheckRate: int32(t.CheckRate), Status: string(t.Status), Paused: t.Paused, Uptime: t.Uptime,
			ContactGroups: list(t.ContactGroups), Tags: list(t.Tags),
		}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l UptimeList) FormatText() ([]byte, error) {
	headers := []string{"ID", "NAME", "TYPE", "URL", "INTERVAL", "STATUS"}
	withUptime := slices.ContainsFunc(l, func(r UptimeRow) bool { return r.Uptime != nil })
	if withUptime {
		headers = append(headers, "UPTIME")
	}

	rows := make([][]string, len(l))
	for i, r := range l {
		rows[i] = []string{r.ID, r.Name, r.TestType, r.WebsiteURL, interval(r.CheckRate), status(r.Paused, r.Status)}
		if withUptime {
			rows[i] = append(rows[i], percent(r.Uptime))
		}
	}
	return table(headers, rows, "No uptime checks found."), nil
}

// UptimeStatus is what the API reports about an uptime check, as opposed to
// what was asked for.
type UptimeStatus struct {
	Status       string     `json:"status"`
	Uptime       float32    `json:"uptime"`
	Processing   bool       `json:"processing"`
	LastTestedAt *time.Time `json:"last_tested_at,omitempty"`
	NextLocation *string    `json:"next_location,omitempty"`
}

// UptimeDescription is the full detail of an uptime check.
type UptimeDescription struct {
	ID     string       `json:"id"`
	Spec   spec.Uptime  `json:"spec"`
	Status UptimeStatus `json:"status"`
}

// NewUptimeDescription describes an uptime check. Credentials and monitoring
// regions are never returned by the API, so they are absent from the spec.
func NewUptimeDescription(t domain.Uptime) UptimeDescription {
	return UptimeDescription{
		ID: t.ID,
		Spec: spec.Uptime{
			Name:            gofp.Some(t.Name),
			TestType:        gofp.Some(string(t.TestType)),
			WebsiteURL:      gofp.Some(t.WebsiteURL),
			CheckRate:       gofp.Some(int32(t.CheckRate)),
			Confirmation:    gofp.Some(t.Confirmation),
			ContactGroups:   gofp.Some(list(t.ContactGroups)),
			CustomHeader:    someIf(t.CustomHeader),
			DNSIPs:          gofp.Some(list(t.DNSIPs)),
			DNSServer:       someIf(t.DNSServer),
			DoNotFind:       gofp.Some(t.DoNotFind),
			EnableSSLAlert:  gofp.Some(t.EnableSSLAlert),
			FinalEndpoint:   someIf(t.FinalEndpoint),
			FindString:      someIf(t.FindString),
			FollowRedirects: gofp.Some(t.FollowRedirects),
			Host:            someIf(t.Host),
			IncludeHeader:   gofp.Some(t.IncludeHeader),
			Paused:          gofp.Some(t.Paused),
			Port:            someIf(t.Port),
			PostBody:        someIf(t.PostBody),
			PostRaw:         someIf(t.PostRaw),
			StatusCodes:     gofp.Some(list(t.StatusCodes)),
			Tags:            gofp.Some(list(t.Tags)),
			Timeout:         gofp.Some(t.Timeout),
			TriggerRate:     gofp.Some(t.TriggerRate),
			UseJAR:          gofp.Some(t.UseJAR),
			UserAgent:       someIf(t.UserAgent),
		},
		Status: UptimeStatus{
			Status: string(t.Status), Uptime: t.Uptime, Processing: t.Processing,
			LastTestedAt: t.LastTested, NextLocation: t.NextLocation,
		},
	}
}

// FormatText renders the description as labelled fields.
func (d UptimeDescription) FormatText() ([]byte, error) {
	s := d.Spec
	return fields(
		field{"ID", d.ID},
		field{"Name", show(s.Name)},
		field{"Type", show(s.TestType)},
		field{"URL", show(s.WebsiteURL)},
		field{"Interval", interval(s.CheckRate.UnwrapOr(0))},
		field{"Status", status(s.Paused.UnwrapOr(false), d.Status.Status)},
		field{"Uptime", percent(&d.Status.Uptime)},
		field{"Last tested", instant(d.Status.LastTestedAt)},
		field{"Contact groups", show(s.ContactGroups)},
		field{"Tags", show(s.Tags)},
		field{"Confirmation", show(s.Confirmation)},
		field{"Trigger rate", show(s.TriggerRate)},
		field{"Timeout", show(s.Timeout)},
		field{"Status codes", show(s.StatusCodes)},
		field{"Follow redirects", show(s.FollowRedirects)},
		field{"Find string", show(s.FindString)},
		field{"Do not find", show(s.DoNotFind)},
		field{"Host", show(s.Host)},
		field{"Port", show(s.Port)},
		field{"DNS server", show(s.DNSServer)},
		field{"DNS IPs", show(s.DNSIPs)},
		field{"SSL alerts", show(s.EnableSSLAlert)},
		field{"Custom header", show(s.CustomHeader)},
		field{"User agent", show(s.UserAgent)},
	), nil
}

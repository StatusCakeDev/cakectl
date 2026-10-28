package view

import (
	"fmt"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/domain"
	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// PagespeedRow is a pagespeed check as shown in a listing.
type PagespeedRow struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	WebsiteURL    string   `json:"website_url"`
	CheckRate     int32    `json:"check_rate"`
	Region        string   `json:"region"`
	Paused        bool     `json:"paused"`
	ContactGroups []string `json:"contact_groups"`
}

// PagespeedList is a listing of pagespeed checks.
type PagespeedList []PagespeedRow

// NewPagespeedList provides the listing of the given checks.
func NewPagespeedList(tests []domain.Pagespeed) PagespeedList {
	rows := make(PagespeedList, len(tests))
	for i, t := range tests {
		rows[i] = PagespeedRow{
			ID: t.ID, Name: t.Name, WebsiteURL: t.WebsiteURL, CheckRate: int32(t.CheckRate),
			Region: t.Location, Paused: t.Paused, ContactGroups: list(t.ContactGroups),
		}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l PagespeedList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		rows[i] = []string{r.ID, r.Name, r.WebsiteURL, interval(r.CheckRate), r.Region, status(r.Paused, "active")}
	}
	return table([]string{"ID", "NAME", "URL", "INTERVAL", "REGION", "STATUS"}, rows, "No pagespeed checks found."), nil
}

// PagespeedStats are the results of the latest pagespeed test.
type PagespeedStats struct {
	FilesizeKB  float32 `json:"filesize_kb"`
	LoadtimeMS  int32   `json:"loadtime_ms"`
	Requests    int32   `json:"requests"`
	HasIssue    bool    `json:"has_issue"`
	LatestIssue *string `json:"latest_issue,omitempty"`
}

// PagespeedStatus is what the API reports about a pagespeed check.
type PagespeedStatus struct {
	Latest *PagespeedStats `json:"latest,omitempty"`
}

// PagespeedDescription is the full detail of a pagespeed check.
type PagespeedDescription struct {
	ID     string          `json:"id"`
	Spec   spec.Pagespeed  `json:"spec"`
	Status PagespeedStatus `json:"status"`
}

// NewPagespeedDescription describes a pagespeed check.
func NewPagespeedDescription(t domain.Pagespeed) PagespeedDescription {
	d := PagespeedDescription{
		ID: t.ID,
		Spec: spec.Pagespeed{
			Name:          gofp.Some(t.Name),
			WebsiteURL:    gofp.Some(t.WebsiteURL),
			CheckRate:     gofp.Some(int32(t.CheckRate)),
			Region:        gofp.Some(t.Location),
			AlertBigger:   gofp.Some(t.AlertBigger),
			AlertSlower:   gofp.Some(t.AlertSlower),
			AlertSmaller:  gofp.Some(t.AlertSmaller),
			ContactGroups: gofp.Some(list(t.ContactGroups)),
			Paused:        gofp.Some(t.Paused),
		},
	}
	if s := t.LatestStats; s != nil {
		d.Status.Latest = &PagespeedStats{
			FilesizeKB: s.Filesize, LoadtimeMS: s.Loadtime, Requests: s.Requests,
			HasIssue: s.HasIssue, LatestIssue: s.LatestIssue,
		}
	}
	return d
}

// FormatText renders the description as labelled fields.
func (d PagespeedDescription) FormatText() ([]byte, error) {
	s := d.Spec
	latest := none
	if l := d.Status.Latest; l != nil {
		latest = fmt.Sprintf("%d ms, %d requests, %.1f kB", l.LoadtimeMS, l.Requests, l.FilesizeKB)
		if l.HasIssue {
			latest += " (issue: " + text(l.LatestIssue) + ")"
		}
	}

	return fields(
		field{"ID", d.ID},
		field{"Name", show(s.Name)},
		field{"URL", show(s.WebsiteURL)},
		field{"Interval", interval(s.CheckRate.UnwrapOr(0))},
		field{"Region", show(s.Region)},
		field{"Status", status(s.Paused.UnwrapOr(false), "active")},
		field{"Latest result", latest},
		field{"Alert if bigger than (kB)", show(s.AlertBigger)},
		field{"Alert if slower than (ms)", show(s.AlertSlower)},
		field{"Alert if fewer requests than", show(s.AlertSmaller)},
		field{"Contact groups", show(s.ContactGroups)},
	), nil
}

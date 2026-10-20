package view

import (
	"fmt"
	"strconv"
	"time"

	"github.com/StatusCakeDev/statuscake-go"
)

// AlertRow is one alert an uptime check raised.
type AlertRow struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	StatusCode  int32      `json:"status_code"`
	TriggeredAt *time.Time `json:"triggered_at,omitempty"`
}

// AlertList is a listing of alerts.
type AlertList []AlertRow

// NewAlertList provides the listing of the given alerts.
func NewAlertList(alerts []statuscake.UptimeTestAlert) AlertList {
	rows := make(AlertList, len(alerts))
	for i, a := range alerts {
		rows[i] = AlertRow{ID: a.ID, Status: string(a.Status), StatusCode: a.StatusCode, TriggeredAt: a.Triggered}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l AlertList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		rows[i] = []string{r.ID, instant(r.TriggeredAt), r.Status, strconv.Itoa(int(r.StatusCode))}
	}
	return table([]string{"ID", "TRIGGERED", "STATUS", "CODE"}, rows, "No alerts found."), nil
}

// HistoryRow is one test an uptime check ran.
type HistoryRow struct {
	CreatedAt   time.Time `json:"created_at"`
	Location    *string   `json:"location,omitempty"`
	StatusCode  *int32    `json:"status_code,omitempty"`
	Performance *int64    `json:"performance,omitempty"`
}

// HistoryList is a listing of uptime test results.
type HistoryList []HistoryRow

// NewHistoryList provides the listing of the given results.
func NewHistoryList(results []statuscake.UptimeTestHistoryResult) HistoryList {
	rows := make(HistoryList, len(results))
	for i, r := range results {
		rows[i] = HistoryRow{CreatedAt: r.Created, Location: r.Location, StatusCode: r.StatusCode, Performance: r.Performance}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l HistoryList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		code, perf := none, none
		if r.StatusCode != nil {
			code = strconv.Itoa(int(*r.StatusCode))
		}
		if r.Performance != nil {
			perf = fmt.Sprintf("%d ms", *r.Performance)
		}
		rows[i] = []string{r.CreatedAt.Format(time.RFC3339), text(r.Location), code, perf}
	}
	return table([]string{"TIME", "LOCATION", "CODE", "RESPONSE"}, rows, "No results found."), nil
}

// PeriodRow is a stretch of time an uptime check spent up or down.
type PeriodRow struct {
	CreatedAt time.Time  `json:"created_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Status    string     `json:"status"`
	Duration  *int64     `json:"duration,omitempty"`
}

// PeriodList is a listing of periods.
type PeriodList []PeriodRow

// NewPeriodList provides the listing of the given periods.
func NewPeriodList(periods []statuscake.UptimeTestPeriod) PeriodList {
	rows := make(PeriodList, len(periods))
	for i, p := range periods {
		rows[i] = PeriodRow{CreatedAt: p.Created, EndedAt: p.Ended, Status: string(p.Status), Duration: p.Duration}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l PeriodList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		duration := none
		if r.Duration != nil {
			duration = (time.Duration(*r.Duration) * time.Millisecond).String()
		}
		rows[i] = []string{r.CreatedAt.Format(time.RFC3339), instant(r.EndedAt), r.Status, duration}
	}
	return table([]string{"STARTED", "ENDED", "STATUS", "DURATION"}, rows, "No periods found."), nil
}

// PagespeedResultRow is one test a pagespeed check ran.
type PagespeedResultRow struct {
	CreatedAt  time.Time `json:"created_at"`
	LoadtimeMS int64     `json:"loadtime_ms"`
	Requests   int32     `json:"requests"`
	FilesizeKB float32   `json:"filesize_kb"`
	Throttling string    `json:"throttling"`
	HARURL     string    `json:"har_location"`
}

// PagespeedResultList is a listing of pagespeed test results.
type PagespeedResultList []PagespeedResultRow

// NewPagespeedResultList provides the listing of the given results.
func NewPagespeedResultList(results []statuscake.PagespeedTestHistoryResult) PagespeedResultList {
	rows := make(PagespeedResultList, len(results))
	for i, r := range results {
		rows[i] = PagespeedResultRow{
			CreatedAt: r.Created, LoadtimeMS: r.Loadtime, Requests: r.Requests,
			FilesizeKB: r.Filesize, Throttling: string(r.Throttling), HARURL: r.HARLocation,
		}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l PagespeedResultList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		rows[i] = []string{
			r.CreatedAt.Format(time.RFC3339), fmt.Sprintf("%d ms", r.LoadtimeMS),
			strconv.Itoa(int(r.Requests)), fmt.Sprintf("%.1f kB", r.FilesizeKB), r.Throttling,
		}
	}
	return table([]string{"TIME", "LOAD TIME", "REQUESTS", "SIZE", "THROTTLING"}, rows, "No results found."), nil
}

// LocationRow is a place checks can run from.
type LocationRow struct {
	RegionCode  string  `json:"region_code"`
	Region      string  `json:"region"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	IPv4        *string `json:"ipv4,omitempty"`
	IPv6        *string `json:"ipv6,omitempty"`
}

// LocationList is a listing of monitoring locations.
type LocationList []LocationRow

// NewLocationList provides the listing of the given locations.
func NewLocationList(locations []statuscake.MonitoringLocation) LocationList {
	rows := make(LocationList, len(locations))
	for i, l := range locations {
		rows[i] = LocationRow{
			RegionCode: l.RegionCode, Region: l.Region, Description: l.Description,
			Status: string(l.Status), IPv4: l.IPv4, IPv6: l.IPv6,
		}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l LocationList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		rows[i] = []string{r.RegionCode, r.Region, r.Description, r.Status, text(r.IPv4)}
	}
	return table([]string{"CODE", "REGION", "DESCRIPTION", "STATUS", "IPV4"}, rows, "No locations found."), nil
}

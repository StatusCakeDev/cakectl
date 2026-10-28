package view

import (
	"time"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/domain"
	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// SSLRow is an SSL check as shown in a listing.
type SSLRow struct {
	ID            string     `json:"id"`
	WebsiteURL    string     `json:"website_url"`
	CheckRate     int32      `json:"check_rate"`
	Paused        bool       `json:"paused"`
	ValidUntil    *time.Time `json:"valid_until,omitempty"`
	ContactGroups []string   `json:"contact_groups"`
}

// SSLList is a listing of SSL checks.
type SSLList []SSLRow

// NewSSLList provides the listing of the given checks.
func NewSSLList(tests []domain.SSL) SSLList {
	rows := make(SSLList, len(tests))
	for i, t := range tests {
		rows[i] = SSLRow{
			ID: t.ID, WebsiteURL: t.WebsiteURL, CheckRate: int32(t.CheckRate), Paused: t.Paused,
			ValidUntil: t.ValidUntil, ContactGroups: list(t.ContactGroups),
		}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l SSLList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		rows[i] = []string{r.ID, r.WebsiteURL, interval(r.CheckRate), instant(r.ValidUntil), status(r.Paused, "active")}
	}
	return table([]string{"ID", "URL", "INTERVAL", "EXPIRES", "STATUS"}, rows, "No SSL checks found."), nil
}

// SSLStatus is what the API reports about an SSL check.
type SSLStatus struct {
	CertificateStatus *string    `json:"certificate_status,omitempty"`
	CertificateScore  *int32     `json:"certificate_score,omitempty"`
	Cipher            *string    `json:"cipher,omitempty"`
	CipherScore       *int32     `json:"cipher_score,omitempty"`
	IssuerCommonName  *string    `json:"issuer_common_name,omitempty"`
	ValidFrom         *time.Time `json:"valid_from,omitempty"`
	ValidUntil        *time.Time `json:"valid_until,omitempty"`
	MixedContent      int        `json:"mixed_content"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
}

// SSLDescription is the full detail of an SSL check.
type SSLDescription struct {
	ID     string    `json:"id"`
	Spec   spec.SSL  `json:"spec"`
	Status SSLStatus `json:"status"`
}

// NewSSLDescription describes an SSL check.
func NewSSLDescription(t domain.SSL) SSLDescription {
	return SSLDescription{
		ID: t.ID,
		Spec: spec.SSL{
			WebsiteURL:      gofp.Some(t.WebsiteURL),
			CheckRate:       gofp.Some(int32(t.CheckRate)),
			AlertAt:         gofp.Some(list(t.AlertAt)),
			AlertBroken:     gofp.Some(t.AlertBroken),
			AlertExpiry:     gofp.Some(t.AlertExpiry),
			AlertMixed:      gofp.Some(t.AlertMixed),
			AlertReminder:   gofp.Some(t.AlertReminder),
			ContactGroups:   gofp.Some(list(t.ContactGroups)),
			FollowRedirects: gofp.Some(t.FollowRedirects),
			Hostname:        someIf(t.Hostname),
			Paused:          gofp.Some(t.Paused),
			UserAgent:       someIf(t.UserAgent),
		},
		Status: SSLStatus{
			CertificateStatus: t.CertificateStatus, CertificateScore: t.CertificateScore,
			Cipher: t.Cipher, CipherScore: t.CipherScore, IssuerCommonName: t.IssuerCommonName,
			ValidFrom: t.ValidFrom, ValidUntil: t.ValidUntil, MixedContent: t.MixedContent, UpdatedAt: t.Updated,
		},
	}
}

// FormatText renders the description as labelled fields.
func (d SSLDescription) FormatText() ([]byte, error) {
	s := d.Spec
	return fields(
		field{"ID", d.ID},
		field{"URL", show(s.WebsiteURL)},
		field{"Hostname", show(s.Hostname)},
		field{"Interval", interval(s.CheckRate.UnwrapOr(0))},
		field{"Status", status(s.Paused.UnwrapOr(false), "active")},
		field{"Certificate", text(d.Status.CertificateStatus)},
		field{"Issuer", text(d.Status.IssuerCommonName)},
		field{"Valid from", instant(d.Status.ValidFrom)},
		field{"Valid until", instant(d.Status.ValidUntil)},
		field{"Cipher", text(d.Status.Cipher)},
		field{"Mixed content items", format(d.Status.MixedContent)},
		field{"Alert at (days)", show(s.AlertAt)},
		field{"Alert on broken", show(s.AlertBroken)},
		field{"Alert on expiry", show(s.AlertExpiry)},
		field{"Alert on mixed content", show(s.AlertMixed)},
		field{"Alert reminders", show(s.AlertReminder)},
		field{"Follow redirects", show(s.FollowRedirects)},
		field{"Contact groups", show(s.ContactGroups)},
		field{"User agent", show(s.UserAgent)},
	), nil
}

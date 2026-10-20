package spec

import "github.com/tomasbasham/gofp"

// SSLCheckRates are the intervals, in seconds, at which an SSL check can run.
var SSLCheckRates = []int32{300, 600, 1800, 3600, 86400, 2073600}

// SSL is the desired state of an SSL check. AlertAt lists how many days before
// the certificate expires an alert is raised.
type SSL struct {
	WebsiteURL      gofp.Option[string]   `json:"website_url,omitzero"`
	CheckRate       gofp.Option[int32]    `json:"check_rate,omitzero"`
	AlertAt         gofp.Option[[]int32]  `json:"alert_at,omitzero"`
	AlertBroken     gofp.Option[bool]     `json:"alert_broken,omitzero"`
	AlertExpiry     gofp.Option[bool]     `json:"alert_expiry,omitzero"`
	AlertMixed      gofp.Option[bool]     `json:"alert_mixed,omitzero"`
	AlertReminder   gofp.Option[bool]     `json:"alert_reminder,omitzero"`
	ContactGroups   gofp.Option[[]string] `json:"contact_groups,omitzero"`
	FollowRedirects gofp.Option[bool]     `json:"follow_redirects,omitzero"`
	Hostname        gofp.Option[string]   `json:"hostname,omitzero"`
	Paused          gofp.Option[bool]     `json:"paused,omitzero"`
	UserAgent       gofp.Option[string]   `json:"user_agent,omitzero"`
}

// ValidateCreate reports every reason the spec cannot create a check.
func (s SSL) ValidateCreate() error {
	var v validator
	required(&v, "website_url", s.WebsiteURL)
	required(&v, "check_rate", s.CheckRate)
	required(&v, "alert_at", s.AlertAt)
	s.validateValues(&v)
	return v.err()
}

// ValidateUpdate reports every reason the spec cannot update a check.
func (s SSL) ValidateUpdate() error {
	var v validator
	s.validateValues(&v)
	return v.err()
}

func (s SSL) validateValues(v *validator) {
	oneOf(v, "check_rate", s.CheckRate, SSLCheckRates)
}

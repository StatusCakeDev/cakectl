package spec

import "github.com/tomasbasham/gofp"

// PagespeedCheckRates are the intervals, in seconds, at which a pagespeed
// check can run.
var PagespeedCheckRates = []int32{60, 300, 600, 900, 1800, 3600, 86400}

// PagespeedRegions are the locations a pagespeed check can run from.
var PagespeedRegions = []string{"AU", "CA", "DE", "FR", "IN", "JP", "NL", "SG", "UK", "US", "USW"}

// Pagespeed is the desired state of a pagespeed check. The alert thresholds
// are a page size in kilobytes, a load time in milliseconds and a request
// count; a value of zero disables the alert.
type Pagespeed struct {
	Name          gofp.Option[string]   `json:"name,omitzero"`
	WebsiteURL    gofp.Option[string]   `json:"website_url,omitzero"`
	CheckRate     gofp.Option[int32]    `json:"check_rate,omitzero"`
	Region        gofp.Option[string]   `json:"region,omitzero"`
	AlertBigger   gofp.Option[int32]    `json:"alert_bigger,omitzero"`
	AlertSlower   gofp.Option[int64]    `json:"alert_slower,omitzero"`
	AlertSmaller  gofp.Option[int32]    `json:"alert_smaller,omitzero"`
	ContactGroups gofp.Option[[]string] `json:"contact_groups,omitzero"`
	Paused        gofp.Option[bool]     `json:"paused,omitzero"`
}

// ValidateCreate reports every reason the spec cannot create a check.
func (p Pagespeed) ValidateCreate() error {
	var v validator
	required(&v, "name", p.Name)
	required(&v, "website_url", p.WebsiteURL)
	required(&v, "check_rate", p.CheckRate)
	required(&v, "region", p.Region)
	p.validateValues(&v)
	return v.err()
}

// ValidateUpdate reports every reason the spec cannot update a check.
func (p Pagespeed) ValidateUpdate() error {
	var v validator
	p.validateValues(&v)
	return v.err()
}

func (p Pagespeed) validateValues(v *validator) {
	oneOf(v, "check_rate", p.CheckRate, PagespeedCheckRates)
	oneOf(v, "region", p.Region, PagespeedRegions)
}

package spec

import "github.com/tomasbasham/gofp"

// UptimeTestTypes are the kinds of uptime check that can be created.
var UptimeTestTypes = []string{"DNS", "HEAD", "HTTP", "PING", "SMTP", "SSH", "TCP"}

// UptimeCheckRates are the intervals, in seconds, at which an uptime check can
// run. A rate of zero asks for continuous checking.
var UptimeCheckRates = []int32{0, 30, 60, 300, 900, 1800, 3600, 86400}

// Uptime is the desired state of an uptime check.
//
// Basic authentication credentials and the monitoring regions can be written
// but are never returned by the API, so they cannot appear in the spec of an
// existing check.
type Uptime struct {
	Name            gofp.Option[string]   `json:"name,omitzero"`
	TestType        gofp.Option[string]   `json:"test_type,omitzero"`
	WebsiteURL      gofp.Option[string]   `json:"website_url,omitzero"`
	CheckRate       gofp.Option[int32]    `json:"check_rate,omitzero"`
	BasicUsername   gofp.Option[string]   `json:"basic_username,omitzero"`
	BasicPassword   gofp.Option[string]   `json:"basic_password,omitzero"`
	Confirmation    gofp.Option[int32]    `json:"confirmation,omitzero"`
	ContactGroups   gofp.Option[[]string] `json:"contact_groups,omitzero"`
	CustomHeader    gofp.Option[string]   `json:"custom_header,omitzero"`
	DNSIPs          gofp.Option[[]string] `json:"dns_ips,omitzero"`
	DNSServer       gofp.Option[string]   `json:"dns_server,omitzero"`
	DoNotFind       gofp.Option[bool]     `json:"do_not_find,omitzero"`
	EnableSSLAlert  gofp.Option[bool]     `json:"enable_ssl_alert,omitzero"`
	FinalEndpoint   gofp.Option[string]   `json:"final_endpoint,omitzero"`
	FindString      gofp.Option[string]   `json:"find_string,omitzero"`
	FollowRedirects gofp.Option[bool]     `json:"follow_redirects,omitzero"`
	Host            gofp.Option[string]   `json:"host,omitzero"`
	IncludeHeader   gofp.Option[bool]     `json:"include_header,omitzero"`
	Paused          gofp.Option[bool]     `json:"paused,omitzero"`
	Port            gofp.Option[int32]    `json:"port,omitzero"`
	PostBody        gofp.Option[string]   `json:"post_body,omitzero"`
	PostRaw         gofp.Option[string]   `json:"post_raw,omitzero"`
	Regions         gofp.Option[[]string] `json:"regions,omitzero"`
	StatusCodes     gofp.Option[[]string] `json:"status_codes,omitzero"`
	Tags            gofp.Option[[]string] `json:"tags,omitzero"`
	Timeout         gofp.Option[int32]    `json:"timeout,omitzero"`
	TriggerRate     gofp.Option[int32]    `json:"trigger_rate,omitzero"`
	UseJAR          gofp.Option[bool]     `json:"use_jar,omitzero"`
	UserAgent       gofp.Option[string]   `json:"user_agent,omitzero"`
}

// ValidateCreate reports every reason the spec cannot create a check.
func (u Uptime) ValidateCreate() error {
	var v validator
	required(&v, "name", u.Name)
	required(&v, "test_type", u.TestType)
	required(&v, "website_url", u.WebsiteURL)
	required(&v, "check_rate", u.CheckRate)
	u.validateValues(&v)
	return v.err()
}

// ValidateUpdate reports every reason the spec cannot update a check. The type
// of a check is fixed when it is created.
func (u Uptime) ValidateUpdate() error {
	var v validator
	if u.TestType.IsSome() {
		v.fail("test_type cannot be changed once a check exists")
	}
	u.validateValues(&v)
	return v.err()
}

func (u Uptime) validateValues(v *validator) {
	oneOf(v, "test_type", u.TestType, UptimeTestTypes)
	oneOf(v, "check_rate", u.CheckRate, UptimeCheckRates)
}

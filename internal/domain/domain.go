// Package domain holds cakectl's own view of what StatusCake reports: the read
// models that adapters return and that commands render, and the errors by which
// adapters report what went wrong.
//
// It is deliberately small and owned by cakectl alone. The other StatusCake
// applications see the same API through models of their own, which lets each
// shape the data to its needs and keeps them from depending on one another.
//
// Enumerations that the API defines, such as a check's status, are plain
// strings here: cakectl displays them and never branches on them, so there is
// nothing for a type to protect.
package domain

import (
	"errors"
	"time"
)

var (
	// ErrNotFound reports that the API has no resource with the requested ID.
	ErrNotFound = errors.New("not found")

	// ErrUnauthorized reports that StatusCake rejected the API token, either
	// because it is not valid or because it may not do what was asked.
	ErrUnauthorized = errors.New("credentials rejected")

	// ErrTruncated reports a listing too large to be read in full. It is an
	// error rather than a flag so that a partial listing is never mistaken for
	// a whole one.
	ErrTruncated = errors.New("listing is too large to read in full")

	// ErrNothingToUpdate reports an update that sets no fields.
	ErrNothingToUpdate = errors.New("nothing to update: no fields were supplied")
)

// UptimeOverview is an uptime check as it appears in a listing.
type UptimeOverview struct {
	ID            string
	Name          string
	TestType      string
	WebsiteURL    string
	CheckRate     int32
	ContactGroups []string
	Paused        bool
	Status        string
	Tags          []string
	Uptime        *float32
}

// Uptime is the full detail of an uptime check.
type Uptime struct {
	ID              string
	Name            string
	TestType        string
	WebsiteURL      string
	CheckRate       int32
	Confirmation    int32
	ContactGroups   []string
	CustomHeader    *string
	DNSIPs          []string
	DNSServer       *string
	DoNotFind       bool
	EnableSSLAlert  bool
	FinalEndpoint   *string
	FindString      *string
	FollowRedirects bool
	IncludeHeader   bool
	Host            *string
	LastTested      *time.Time
	NextLocation    *string
	Paused          bool
	Port            *int32
	PostBody        *string
	PostRaw         *string
	Processing      bool
	Status          string
	StatusCodes     []string
	Tags            []string
	Timeout         int32
	TriggerRate     int32
	Uptime          float32
	UseJAR          bool
	UserAgent       *string
}

// UptimeAlert is one alert an uptime check raised.
type UptimeAlert struct {
	ID         string
	Status     string
	StatusCode int32
	Triggered  *time.Time
}

// UptimeResult is one test an uptime check ran.
type UptimeResult struct {
	Created     time.Time
	Location    *string
	Performance *int64
	StatusCode  *int32
}

// UptimePeriod is a stretch of time an uptime check spent up or down. Duration
// is in milliseconds.
type UptimePeriod struct {
	Created  time.Time
	Duration *int64
	Ended    *time.Time
	Status   string
}

// HeartbeatOverview is a heartbeat check as it appears in a listing.
type HeartbeatOverview struct {
	ID            string
	Name          string
	Period        int32
	ContactGroups []string
	Paused        bool
	Status        string
	Tags          []string
	Uptime        *float32
}

// Heartbeat is the full detail of a heartbeat check. WebsiteURL is the address
// the monitored job must ping, and holds a secret token.
type Heartbeat struct {
	ID            string
	Name          string
	WebsiteURL    string
	Period        int32
	ContactGroups []string
	Host          *string
	LastTested    *time.Time
	Paused        bool
	Status        string
	Tags          []string
	Uptime        float32
}

// Pagespeed is a pagespeed check.
type Pagespeed struct {
	ID            string
	Name          string
	WebsiteURL    string
	CheckRate     int32
	AlertBigger   int32
	AlertSlower   int64
	AlertSmaller  int32
	ContactGroups []string
	LatestStats   *PagespeedStats
	Location      string
	Paused        bool
}

// PagespeedStats are the results of a pagespeed check's latest test.
type PagespeedStats struct {
	Filesize    float32
	HasIssue    bool
	LatestIssue *string
	Loadtime    int32
	Requests    int32
}

// PagespeedResult is one test a pagespeed check ran.
type PagespeedResult struct {
	Created     time.Time
	Filesize    float32
	HARLocation string
	Loadtime    int64
	Requests    int32
	Throttling  string
}

// SSL is an SSL check. MixedContent counts the insecure items found on the
// page; their details are not modelled.
type SSL struct {
	ID                string
	WebsiteURL        string
	CheckRate         int32
	AlertAt           []int32
	AlertBroken       bool
	AlertExpiry       bool
	AlertMixed        bool
	AlertReminder     bool
	CertificateScore  *int32
	CertificateStatus *string
	Cipher            *string
	CipherScore       *int32
	ContactGroups     []string
	IssuerCommonName  *string
	FollowRedirects   bool
	Hostname          *string
	MixedContent      int
	Paused            bool
	Updated           *time.Time
	UserAgent         *string
	ValidFrom         *time.Time
	ValidUntil        *time.Time
}

// ContactGroup is a set of people and services that are alerted together.
type ContactGroup struct {
	ID             string
	Name           string
	EmailAddresses []string
	Integrations   []string
	MobileNumbers  []string
	PingURL        *string
}

// MaintenanceWindow is a period during which checks do not alert.
type MaintenanceWindow struct {
	ID             string
	Name           string
	End            time.Time
	RepeatInterval string
	Start          time.Time
	State          string
	Tags           []string
	Tests          []string
	Timezone       string
}

// Location is a place checks can run from.
type Location struct {
	Description string
	IPv4        *string
	IPv6        *string
	Region      string
	RegionCode  string
	Status      string
}

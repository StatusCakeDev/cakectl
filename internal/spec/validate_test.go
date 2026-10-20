package spec_test

import (
	"strings"
	"testing"
	"time"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/spec"
)

type validation struct {
	spec interface {
		ValidateCreate() error
		ValidateUpdate() error
	}
	// create and update are the substrings each error must contain, in any order.
	// A nil slice means the validation must succeed.
	create []string
	update []string
}

func check(t *testing.T, name string, err error, wants []string) {
	t.Helper()

	if len(wants) == 0 {
		if err != nil {
			t.Errorf("%s error = %v, want none", name, err)
		}
		return
	}
	if err == nil {
		t.Fatalf("%s error = nil, want errors mentioning %v", name, wants)
	}
	for _, want := range wants {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%s error = %q, want it to mention %q", name, err, want)
		}
	}
}

func TestValidation(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 11, 1, 22, 0, 0, 0, time.UTC)

	tests := map[string]validation{
		"uptime complete": {
			spec: spec.Uptime{
				Name: gofp.Some("a"), TestType: gofp.Some("HTTP"),
				WebsiteURL: gofp.Some("https://example.com"), CheckRate: gofp.Some(int32(300)),
			},
			update: []string{"test_type cannot be changed"},
		},
		"uptime empty": {
			spec:   spec.Uptime{},
			create: []string{"name is required", "test_type is required", "website_url is required", "check_rate is required"},
		},
		"uptime bad enums": {
			spec: spec.Uptime{
				Name: gofp.Some("a"), TestType: gofp.Some("FTP"),
				WebsiteURL: gofp.Some("x"), CheckRate: gofp.Some(int32(45)),
			},
			create: []string{"test_type must be one of DNS, HEAD", "check_rate must be one of 0, 30"},
			update: []string{"test_type cannot be changed", "check_rate must be one of"},
		},
		"uptime partial update": {
			spec:   spec.Uptime{Paused: gofp.Some(true)},
			create: []string{"name is required"},
		},
		"heartbeat complete": {
			spec: spec.Heartbeat{Name: gofp.Some("a"), Period: gofp.Some(int32(60))},
		},
		"heartbeat empty": {
			spec:   spec.Heartbeat{},
			create: []string{"name is required", "period is required"},
		},
		"heartbeat bad period": {
			spec:   spec.Heartbeat{Name: gofp.Some("a"), Period: gofp.Some(int32(0))},
			create: []string{"period must be greater than zero"},
			update: []string{"period must be greater than zero"},
		},
		"pagespeed complete": {
			spec: spec.Pagespeed{
				Name: gofp.Some("a"), WebsiteURL: gofp.Some("x"),
				CheckRate: gofp.Some(int32(600)), Region: gofp.Some("UK"),
			},
		},
		"pagespeed empty": {
			spec:   spec.Pagespeed{},
			create: []string{"name is required", "website_url is required", "check_rate is required", "region is required"},
		},
		"pagespeed bad enums": {
			spec: spec.Pagespeed{
				Name: gofp.Some("a"), WebsiteURL: gofp.Some("x"),
				CheckRate: gofp.Some(int32(30)), Region: gofp.Some("MARS"),
			},
			create: []string{"check_rate must be one of 60", "region must be one of AU"},
			update: []string{"check_rate must be one of 60", "region must be one of AU"},
		},
		"ssl complete": {
			spec: spec.SSL{
				WebsiteURL: gofp.Some("x"), CheckRate: gofp.Some(int32(86400)),
				AlertAt: gofp.Some([]int32{7, 30}),
			},
		},
		"ssl empty": {
			spec:   spec.SSL{},
			create: []string{"website_url is required", "check_rate is required", "alert_at is required"},
		},
		"ssl bad rate": {
			spec: spec.SSL{
				WebsiteURL: gofp.Some("x"), CheckRate: gofp.Some(int32(60)),
				AlertAt: gofp.Some([]int32{7}),
			},
			create: []string{"check_rate must be one of 300"},
			update: []string{"check_rate must be one of 300"},
		},
		"contact group complete": {
			spec: spec.ContactGroup{Name: gofp.Some("ops")},
		},
		"contact group empty": {
			spec:   spec.ContactGroup{},
			create: []string{"name is required"},
		},
		"maintenance complete": {
			spec: spec.Maintenance{
				Name: gofp.Some("a"), StartAt: gofp.Some(start), EndAt: gofp.Some(start.Add(time.Hour)),
				Timezone: gofp.Some("UTC"), RepeatInterval: gofp.Some("1w"),
			},
		},
		"maintenance empty": {
			spec:   spec.Maintenance{},
			create: []string{"name is required", "start_at is required", "end_at is required", "timezone is required"},
		},
		"maintenance bad window": {
			spec: spec.Maintenance{
				Name: gofp.Some("a"), StartAt: gofp.Some(start), EndAt: gofp.Some(start),
				Timezone: gofp.Some("UTC"), RepeatInterval: gofp.Some("hourly"),
			},
			create: []string{"end_at must be after start_at", "repeat_interval must be one of never"},
			update: []string{"end_at must be after start_at", "repeat_interval must be one of never"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			check(t, "ValidateCreate()", tt.spec.ValidateCreate(), tt.create)
			check(t, "ValidateUpdate()", tt.spec.ValidateUpdate(), tt.update)
		})
	}
}

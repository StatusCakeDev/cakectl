package view_test

import (
	"time"

	"github.com/StatusCakeDev/cakectl/internal/domain"
)

// The models below are what an adapter would hand over for a check that has
// been configured with most of its settings. Optional fields that the API
// reports as null are left nil.

func ptr[T any](v T) *T { return &v }

func at(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func uptime() domain.Uptime {
	return domain.Uptime{
		ID: "7", Name: "Home page", TestType: "HTTP", WebsiteURL: "https://example.com",
		CheckRate: 300, Confirmation: 2, ContactGroups: []string{"1", "2"},
		CustomHeader: ptr(`{"X":"y"}`), DNSIPs: []string{}, EnableSSLAlert: true, FindString: ptr("ok"),
		FollowRedirects: true, Host: ptr("origin"), LastTested: ptr(at("2026-10-01T10:00:00Z")),
		Port: ptr(int32(8080)), PostBody: ptr("a=b"), Status: "up", StatusCodes: []string{"500", "503"},
		Tags: []string{"prod"}, Timeout: 15, TriggerRate: 2, Uptime: 99.5, UserAgent: ptr("cakectl"),
	}
}

func heartbeat() domain.Heartbeat {
	return domain.Heartbeat{
		ID: "8", Name: "Nightly job", WebsiteURL: "https://heartbeat.example/SECRET-TOKEN", Period: 600,
		ContactGroups: []string{}, Host: ptr("worker-1"), Paused: true, Status: "up",
		Tags: []string{"batch"}, Uptime: 100,
	}
}

func pagespeed() domain.Pagespeed {
	return domain.Pagespeed{
		ID: "9", Name: "Landing", WebsiteURL: "https://example.com", CheckRate: 600,
		AlertBigger: 100, AlertSlower: 4000, AlertSmaller: 3, ContactGroups: []string{"1"},
		LatestStats: &domain.PagespeedStats{
			Filesize: 12.5, HasIssue: true, Loadtime: 900, Requests: 40, LatestIssue: ptr("slow"),
		},
		Location: "UK",
	}
}

func ssl() domain.SSL {
	return domain.SSL{
		ID: "10", WebsiteURL: "https://example.com", CheckRate: 3600, AlertAt: []int32{7, 30},
		AlertBroken: true, AlertExpiry: true, AlertReminder: true, CertificateStatus: ptr("CERT_OK"),
		Cipher: ptr("TLS_AES_256_GCM_SHA384"), ContactGroups: []string{"1"}, IssuerCommonName: ptr("R3"),
		FollowRedirects: true, Hostname: ptr("example.com"), MixedContent: 1, UserAgent: ptr("ua"),
		ValidFrom: ptr(at("2026-09-01T00:00:00Z")), ValidUntil: ptr(at("2026-12-01T00:00:00Z")),
	}
}

func contactGroup() domain.ContactGroup {
	return domain.ContactGroup{
		ID: "11", Name: "Ops", EmailAddresses: []string{"ops@example.com", "dev@example.com"},
		Integrations: []string{"slack"}, MobileNumbers: []string{"+447700900000"},
		PingURL: ptr("https://hooks.example/SECRET"),
	}
}

func maintenance() domain.MaintenanceWindow {
	return domain.MaintenanceWindow{
		ID: "12", Name: "Deploy", Start: at("2026-11-01T22:00:00Z"), End: at("2026-11-02T01:00:00Z"),
		RepeatInterval: "1w", State: "pending", Tags: []string{"release"}, Tests: []string{"7", "8"},
		Timezone: "Europe/London",
	}
}

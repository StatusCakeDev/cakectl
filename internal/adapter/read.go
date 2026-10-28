package adapter

import (
	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/domain"
)

// This file is where the SDK's models stop. Everything an adapter returns has
// been copied into a type from the domain package, so that nothing beyond this
// package depends on the SDK.

// mapAll converts every item, keeping a nil slice nil.
func mapAll[S, D any](items []S, convert func(S) D) []D {
	if items == nil {
		return nil
	}

	out := make([]D, len(items))
	for i, it := range items {
		out[i] = convert(it)
	}
	return out
}

func toUptimeOverview(t statuscake.UptimeTestOverview) domain.UptimeOverview {
	return domain.UptimeOverview{
		ID: t.ID, Name: t.Name, TestType: string(t.TestType), WebsiteURL: t.WebsiteURL,
		CheckRate: int32(t.CheckRate), ContactGroups: t.ContactGroups, Paused: t.Paused,
		Status: string(t.Status), Tags: t.Tags, Uptime: t.Uptime,
	}
}

func toUptime(t statuscake.UptimeTest) domain.Uptime {
	return domain.Uptime{
		ID: t.ID, Name: t.Name, TestType: string(t.TestType), WebsiteURL: t.WebsiteURL,
		CheckRate: int32(t.CheckRate), Confirmation: t.Confirmation, ContactGroups: t.ContactGroups,
		CustomHeader: t.CustomHeader, DNSIPs: t.DNSIPs, DNSServer: t.DNSServer,
		DoNotFind: t.DoNotFind, EnableSSLAlert: t.EnableSSLAlert, FinalEndpoint: t.FinalEndpoint,
		FindString: t.FindString, FollowRedirects: t.FollowRedirects, IncludeHeader: t.IncludeHeader,
		Host: t.Host, LastTested: t.LastTested, NextLocation: t.NextLocation, Paused: t.Paused,
		Port: t.Port, PostBody: t.PostBody, PostRaw: t.PostRaw, Processing: t.Processing,
		Status: string(t.Status), StatusCodes: t.StatusCodes, Tags: t.Tags, Timeout: t.Timeout,
		TriggerRate: t.TriggerRate, Uptime: t.Uptime, UseJAR: t.UseJAR, UserAgent: t.UserAgent,
	}
}

func toUptimeAlert(a statuscake.UptimeTestAlert) domain.UptimeAlert {
	return domain.UptimeAlert{ID: a.ID, Status: string(a.Status), StatusCode: a.StatusCode, Triggered: a.Triggered}
}

func toUptimeResult(r statuscake.UptimeTestHistoryResult) domain.UptimeResult {
	return domain.UptimeResult{Created: r.Created, Location: r.Location, Performance: r.Performance, StatusCode: r.StatusCode}
}

func toUptimePeriod(p statuscake.UptimeTestPeriod) domain.UptimePeriod {
	return domain.UptimePeriod{Created: p.Created, Duration: p.Duration, Ended: p.Ended, Status: string(p.Status)}
}

func toHeartbeatOverview(t statuscake.HeartbeatTestOverview) domain.HeartbeatOverview {
	return domain.HeartbeatOverview{
		ID: t.ID, Name: t.Name, Period: t.Period, ContactGroups: t.ContactGroups,
		Paused: t.Paused, Status: string(t.Status), Tags: t.Tags, Uptime: t.Uptime,
	}
}

func toHeartbeat(t statuscake.HeartbeatTest) domain.Heartbeat {
	return domain.Heartbeat{
		ID: t.ID, Name: t.Name, WebsiteURL: t.WebsiteURL, Period: t.Period,
		ContactGroups: t.ContactGroups, Host: t.Host, LastTested: t.LastTested, Paused: t.Paused,
		Status: string(t.Status), Tags: t.Tags, Uptime: t.Uptime,
	}
}

func toPagespeed(t statuscake.PagespeedTest) domain.Pagespeed {
	p := domain.Pagespeed{
		ID: t.ID, Name: t.Name, WebsiteURL: t.WebsiteURL, CheckRate: int32(t.CheckRate),
		AlertBigger: t.AlertBigger, AlertSlower: t.AlertSlower, AlertSmaller: t.AlertSmaller,
		ContactGroups: t.ContactGroups, Location: t.Location, Paused: t.Paused,
	}
	if s := t.LatestStats; s != nil {
		p.LatestStats = &domain.PagespeedStats{
			Filesize: s.Filesize, HasIssue: s.HasIssue, LatestIssue: s.LatestIssue,
			Loadtime: s.Loadtime, Requests: s.Requests,
		}
	}
	return p
}

func toPagespeedResult(r statuscake.PagespeedTestHistoryResult) domain.PagespeedResult {
	return domain.PagespeedResult{
		Created: r.Created, Filesize: r.Filesize, HARLocation: r.HARLocation,
		Loadtime: r.Loadtime, Requests: r.Requests, Throttling: string(r.Throttling),
	}
}

func toSSL(t statuscake.SSLTest) domain.SSL {
	return domain.SSL{
		ID: t.ID, WebsiteURL: t.WebsiteURL, CheckRate: int32(t.CheckRate), AlertAt: t.AlertAt,
		AlertBroken: t.AlertBroken, AlertExpiry: t.AlertExpiry, AlertMixed: t.AlertMixed,
		AlertReminder: t.AlertReminder, CertificateScore: t.CertificateScore,
		CertificateStatus: t.CertificateStatus, Cipher: t.Cipher, CipherScore: t.CipherScore,
		ContactGroups: t.ContactGroups, IssuerCommonName: t.IssuerCommonName,
		FollowRedirects: t.FollowRedirects, Hostname: t.Hostname, MixedContent: len(t.MixedContent),
		Paused: t.Paused, Updated: t.Updated, UserAgent: t.UserAgent,
		ValidFrom: t.ValidFrom, ValidUntil: t.ValidUntil,
	}
}

func toContactGroup(g statuscake.ContactGroup) domain.ContactGroup {
	return domain.ContactGroup{
		ID: g.ID, Name: g.Name, EmailAddresses: g.EmailAddresses, Integrations: g.Integrations,
		MobileNumbers: g.MobileNumbers, PingURL: g.PingURL,
	}
}

func toMaintenanceWindow(w statuscake.MaintenanceWindow) domain.MaintenanceWindow {
	return domain.MaintenanceWindow{
		ID: w.ID, Name: w.Name, End: w.End, RepeatInterval: string(w.RepeatInterval),
		Start: w.Start, State: string(w.State), Tags: w.Tags, Tests: w.Tests, Timezone: w.Timezone,
	}
}

func toLocation(l statuscake.MonitoringLocation) domain.Location {
	return domain.Location{
		Description: l.Description, IPv4: l.IPv4, IPv6: l.IPv6, Region: l.Region,
		RegionCode: l.RegionCode, Status: string(l.Status),
	}
}

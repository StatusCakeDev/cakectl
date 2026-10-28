package adapter_test

// Fixtures are shaped like the API's JSON, not like the SDK's Go types, so
// that decoding is exercised along with the adapter.

func uptimeFixture(id string) map[string]any {
	return map[string]any{
		"id": id, "name": "Home page", "test_type": "HTTP", "website_url": "https://example.com",
		"check_rate": 300, "confirmation": 2, "contact_groups": []string{"1"}, "dns_ips": []string{},
		"do_not_find": false, "enable_ssl_alert": true, "follow_redirects": true, "include_header": false,
		"paused": false, "processing": false, "servers": []any{}, "status": "up", "status_codes": []string{"500"},
		"tags": []string{"prod"}, "timeout": 15, "trigger_rate": 2, "uptime": 99.9, "use_jar": false,
	}
}

func uptimeOverviewFixture(id string) map[string]any {
	return map[string]any{
		"id": id, "name": "Home page", "test_type": "HTTP", "website_url": "https://example.com",
		"check_rate": 300, "contact_groups": []string{}, "paused": false, "status": "down", "tags": []string{},
	}
}

func heartbeatFixture(id string) map[string]any {
	return map[string]any{
		"id": id, "name": "Nightly job", "url": "https://heartbeat.example/abc", "period": 600,
		"contact_groups": []string{}, "paused": false, "status": "up", "tags": []string{}, "uptime": 100,
	}
}

func heartbeatOverviewFixture(id string) map[string]any {
	f := heartbeatFixture(id)
	delete(f, "uptime")
	return f
}

func pagespeedFixture(id string) map[string]any {
	return map[string]any{
		"id": id, "name": "Landing", "website_url": "https://example.com", "check_rate": 600,
		"alert_bigger": 0, "alert_slower": 0, "alert_smaller": 0, "contact_groups": []string{},
		"location": "UK", "paused": false,
	}
}

func sslFixture(id string) map[string]any {
	return map[string]any{
		"id": id, "website_url": "https://example.com", "check_rate": 3600, "alert_at": []int{7, 30},
		"alert_broken": true, "alert_expiry": true, "alert_mixed": false, "alert_reminder": false,
		"contact_groups": []string{}, "follow_redirects": true, "mixed_content": []any{}, "paused": false,
	}
}

func contactGroupFixture(id string) map[string]any {
	return map[string]any{
		"id": id, "name": "Ops", "email_addresses": []string{"ops@example.com"},
		"integrations": []string{}, "mobile_numbers": []string{},
	}
}

func maintenanceFixture(id string) map[string]any {
	return map[string]any{
		"id": id, "name": "Deploy", "start_at": "2026-11-01T22:00:00Z", "end_at": "2026-11-02T01:00:00Z",
		"repeat_interval": "never", "state": "pending", "tags": []string{}, "tests": []string{"1"},
		"timezone": "Europe/London",
	}
}

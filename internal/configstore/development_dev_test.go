//go:build dev

package configstore_test

import "testing"

func TestDevelopmentBuildsHaveACustomEndpointSetting(t *testing.T) {
	t.Parallel()

	if err := newStore(t).Set("custom_endpoint", "https://example.com"); err != nil {
		t.Errorf("Set(custom_endpoint) error = %v", err)
	}
}

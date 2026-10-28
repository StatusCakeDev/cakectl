//go:build !dev

package configstore_test

import (
	"errors"
	"testing"

	"github.com/StatusCakeDev/cakectl/internal/configstore"
)

func TestReleaseBuildsHaveNoCustomEndpointSetting(t *testing.T) {
	t.Parallel()

	if err := newStore(t).Set("custom_endpoint", "https://example.com"); !errors.Is(err, configstore.ErrUnknownKey) {
		t.Errorf("Set(custom_endpoint) error = %v, want %v", err, configstore.ErrUnknownKey)
	}
	for _, key := range configstore.Keys() {
		if key == "custom_endpoint" {
			t.Error("Keys() lists custom_endpoint in a release build")
		}
	}
}

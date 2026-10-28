//go:build !dev

package configstore

// developmentSettings are the settings that exist only in development builds.
// A release build has none.
func developmentSettings() map[string]kind { return nil }

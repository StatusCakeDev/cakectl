//go:build dev

package configstore

// developmentSettings are the settings that exist only in development builds,
// which are made with `go build -tags dev`.
func developmentSettings() map[string]kind {
	return map[string]kind{"custom_endpoint": text}
}

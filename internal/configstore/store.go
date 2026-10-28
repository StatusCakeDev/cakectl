// Package configstore reads and writes cakectl's configuration file.
//
// Loading configuration at start-up is viper's job; this package is for the
// commands that change the file. It keeps the file to the settings cakectl
// knows, stores each with the right type, and writes it readable only by its
// owner, since it may hold an API token.
package configstore

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/goccy/go-yaml"
)

// kind is how a setting's value is parsed and stored.
type kind int

const (
	text kind = iota
	number
	duration
)

// settings are the keys the file may hold, matching the mapstructure names of
// the configuration flags.
//
// Development builds have settings that a release build does not (see
// developmentSettings), so that a released cakectl cannot be pointed away from
// the StatusCake API by editing its config file.
var settings = func() map[string]kind {
	s := map[string]kind{
		"api_token":   text,
		"rps":         number,
		"retries":     number,
		"min_backoff": duration,
		"max_backoff": duration,
	}
	for k, v := range developmentSettings() {
		s[k] = v
	}
	return s
}()

// Keys lists the settings, in alphabetical order.
func Keys() []string {
	keys := make([]string, 0, len(settings))
	for k := range settings {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Secret reports whether a setting's value should not be displayed in full.
func Secret(key string) bool {
	return key == "api_token"
}

// ErrUnknownKey reports a setting that cakectl does not have.
var ErrUnknownKey = errors.New("unknown setting")

// ErrNotSet reports a setting that has no value in the file.
var ErrNotSet = errors.New("setting is not set")

// Store is a configuration file.
type Store struct {
	path string
}

// New provides a [Store] for the file at path. The file need not exist yet.
func New(path string) *Store {
	return &Store{path: path}
}

// Path is where the file lives.
func (s *Store) Path() string {
	return s.path
}

// All reads every setting in the file. A file that does not exist has none.
func (s *Store) All() (map[string]any, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	values := map[string]any{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", s.path, err)
	}
	return values, nil
}

// Get reads one setting.
func (s *Store) Get(key string) (any, error) {
	if _, ok := settings[key]; !ok {
		return nil, unknown(key)
	}

	values, err := s.All()
	if err != nil {
		return nil, err
	}
	v, ok := values[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotSet, key)
	}
	return v, nil
}

// Set writes one setting, parsing the value according to the setting's type.
func (s *Store) Set(key, value string) error {
	k, ok := settings[key]
	if !ok {
		return unknown(key)
	}

	var parsed any
	switch k {
	case number:
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("%s must be a whole number, got %q", key, value)
		}
		parsed = n
	case duration:
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("%s must be a duration such as 500ms or 30s, got %q", key, value)
		}
		parsed = value
	default:
		parsed = value
	}

	values, err := s.All()
	if err != nil {
		return err
	}
	values[key] = parsed
	return s.write(values)
}

// Unset removes one setting. Removing a setting that is not set is not an
// error.
func (s *Store) Unset(key string) error {
	if _, ok := settings[key]; !ok {
		return unknown(key)
	}

	values, err := s.All()
	if err != nil {
		return err
	}
	delete(values, key)
	return s.write(values)
}

func (s *Store) write(values map[string]any) error {
	data, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Write beside the file and rename, so an interrupted write cannot leave a
	// truncated config behind.
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".cakectl-*")
	if err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("write config file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write config file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

func unknown(key string) error {
	return fmt.Errorf("%w %q (known settings: %v)", ErrUnknownKey, key, Keys())
}

// Mask hides all but the last four characters of a secret.
func Mask(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return "****" + secret[len(secret)-4:]
}

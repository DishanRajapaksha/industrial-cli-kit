// Package config provides protocol-agnostic YAML configuration helpers.
package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a YAML duration represented as a Go duration string, such as 5s.
type Duration struct{ time.Duration }

func NewDuration(value time.Duration) Duration { return Duration{Duration: value} }

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("duration must be a string")
	}
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value.Value, err)
	}
	d.Duration = parsed
	return nil
}

func (d Duration) MarshalYAML() (any, error) { return d.Duration.String(), nil }

// LoadYAML decodes a YAML file and rejects unknown fields.
func LoadYAML(path string, destination any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %q: %w", path, err)
	}
	return ParseYAML(data, destination)
}

// ParseYAML decodes a YAML document and rejects unknown fields.
func ParseYAML(data []byte, destination any) error {
	decoder := yaml.NewDecoder(bytesReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	return nil
}

// WriteStarter writes contents with owner-only permissions. It refuses to
// overwrite an existing file unless force is true.
func WriteStarter(path string, contents []byte, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("refusing to overwrite existing file %q; use --force", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat config %q: %w", path, err)
		}
	}
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}
	return nil
}

// ResolveProfile applies a named profile over defaults. The caller supplies
// merge because only the protocol knows the meaning of its configuration fields.
func ResolveProfile[T any](defaults T, profiles map[string]T, requested, defaultName string, merge func(T, T) T) (T, string, error) {
	name := requested
	if name == "" {
		name = defaultName
	}
	if name == "" {
		return defaults, "", nil
	}
	override, ok := profiles[name]
	if !ok {
		var zero T
		return zero, "", fmt.Errorf("profile %q not found", name)
	}
	return merge(defaults, override), name, nil
}

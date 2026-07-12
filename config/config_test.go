package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

type exampleConfig struct {
	Timeout Duration `yaml:"timeout"`
}

func TestParseYAMLRejectsUnknownFields(t *testing.T) {
	var value exampleConfig
	if err := ParseYAML([]byte("unknown: value\n"), &value); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestDurationRoundTrip(t *testing.T) {
	var value exampleConfig
	if err := ParseYAML([]byte("timeout: 5s\n"), &value); err != nil {
		t.Fatal(err)
	}
	if got := value.Timeout.Duration; got != 5*time.Second {
		t.Fatalf("duration = %s", got)
	}
}

func TestWriteStarterRefusesOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := WriteStarter(path, []byte("first\n"), false); err != nil {
		t.Fatal(err)
	}
	if err := WriteStarter(path, []byte("second\n"), false); err == nil {
		t.Fatal("overwrite accepted")
	}
	if err := WriteStarter(path, []byte("second\n"), true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "second\n" {
		t.Fatalf("contents = %q", data)
	}
}

func TestResolveProfile(t *testing.T) {
	type value struct{ Host string }
	got, name, err := ResolveProfile(value{Host: "default"}, map[string]value{"site": {Host: "site"}}, "", "site", func(base, override value) value {
		return override
	})
	if err != nil || name != "site" || got.Host != "site" {
		t.Fatalf("ResolveProfile() = %#v, %q, %v", got, name, err)
	}
}

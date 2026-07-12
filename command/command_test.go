package command

import (
	"reflect"
	"testing"
)

var globalFlags = []Flag{
	{Name: "config", TakesValue: true},
	{Name: "profile", TakesValue: true},
	{Name: "verbose"},
}

func TestNormalizeGlobalFlags(t *testing.T) {
	got, err := NormalizeGlobalFlags([]string{"--config", "site.yaml", "--verbose", "read", "--item", "power"}, globalFlags)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"read", "--config", "site.yaml", "--verbose", "--item", "power"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeGlobalFlags() = %#v, want %#v", got, want)
	}
}

func TestNormalizeGlobalFlagsForRegistryPreservesLeadingArguments(t *testing.T) {
	registry := Registry{
		GlobalFlags: globalFlags,
		Commands: []Command{
			{Name: "read", LeadingArgs: 1},
			{Name: "read-point", LeadingArgs: 1},
		},
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "nested operation",
			args: []string{"--config", "site.yaml", "--verbose", "read", "coils", "--address", "10"},
			want: []string{"read", "coils", "--config", "site.yaml", "--verbose", "--address", "10"},
		},
		{
			name: "arbitrary positional name",
			args: []string{"--profile", "local", "read-point", "active_power", "--format", "json"},
			want: []string{"read-point", "active_power", "--profile", "local", "--format", "json"},
		},
		{
			name: "unknown command",
			args: []string{"--verbose", "unknown", "value"},
			want: []string{"unknown", "--verbose", "value"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeGlobalFlagsForRegistry(test.args, registry)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("NormalizeGlobalFlagsForRegistry() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestNormalizeGlobalFlagsRejectsUnknownPrefixFlag(t *testing.T) {
	if _, err := NormalizeGlobalFlags([]string{"--endpoint", "x", "read"}, globalFlags); err == nil {
		t.Fatal("unknown global flag accepted")
	}
}

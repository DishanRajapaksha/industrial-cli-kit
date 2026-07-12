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

func TestNormalizeGlobalFlagsAllowsExplicitEmptyValues(t *testing.T) {
	flags := append(append([]Flag(nil), globalFlags...), Flag{Name: "path", TakesValue: true, AllowEmpty: true})
	for _, args := range [][]string{
		{"--path=", "read", "tag"},
		{"--path", "", "read", "tag"},
	} {
		got, err := NormalizeGlobalFlags(args, flags)
		if err != nil {
			t.Fatalf("NormalizeGlobalFlags(%#v): %v", args, err)
		}
		want := []string{"read", "--path", "", "tag"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("NormalizeGlobalFlags(%#v) = %#v, want %#v", args, got, want)
		}
	}
}

func TestNormalizeGlobalFlagsRejectsUnexpectedEmptyValues(t *testing.T) {
	if _, err := NormalizeGlobalFlags([]string{"--config=", "read"}, globalFlags); err == nil {
		t.Fatal("empty config value accepted")
	}
}

func TestNormalizeGlobalFlagsForRegistryPreservesLeadingArguments(t *testing.T) {
	registry := Registry{
		GlobalFlags: globalFlags,
		Commands: []Command{
			{Name: "read", LeadingArgs: 1},
			{Name: "read-point", LeadingArgs: 1},
			{Name: "identify", LeadingArgs: 1},
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
			name: "optional positional present",
			args: []string{"--profile", "local", "identify", "ahu", "--device-id", "42"},
			want: []string{"identify", "ahu", "--profile", "local", "--device-id", "42"},
		},
		{
			name: "optional positional omitted",
			args: []string{"--profile", "local", "identify", "--device-id", "42"},
			want: []string{"identify", "--profile", "local", "--device-id", "42"},
		},
		{
			name: "dash positional",
			args: []string{"--profile", "local", "read-point", "-", "--format", "json"},
			want: []string{"read-point", "-", "--profile", "local", "--format", "json"},
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

func TestNormalizeGlobalFlagsForRegistryAppliesCommandPolicy(t *testing.T) {
	registry := Registry{
		GlobalFlags: []Flag{
			{Name: "config", TakesValue: true},
			{Name: "profile", TakesValue: true},
			{Name: "endpoint", TakesValue: true},
			{Name: "verbose"},
		},
		Commands: []Command{
			{Name: "read"},
			{Name: "validate-config", GlobalFlags: []string{"config", "profile", "verbose"}},
			{Name: "init-config", GlobalFlags: []string{}},
		},
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "inherits all",
			args: []string{"--endpoint", "opc.tcp://host:4840", "--verbose", "read"},
			want: []string{"read", "--endpoint", "opc.tcp://host:4840", "--verbose"},
		},
		{
			name: "allows subset",
			args: []string{"--config", "site.yaml", "--endpoint", "opc.tcp://host:4840", "--verbose", "validate-config"},
			want: []string{"validate-config", "--config", "site.yaml", "--verbose"},
		},
		{
			name: "allows none",
			args: []string{"--config", "site.yaml", "--verbose", "init-config"},
			want: []string{"init-config"},
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

func TestFilterGlobalFlagsForRegistryKeepsPrefixAndAppliesPolicy(t *testing.T) {
	registry := Registry{
		GlobalFlags: []Flag{
			{Name: "config", TakesValue: true},
			{Name: "format", TakesValue: true},
			{Name: "timeout", TakesValue: true},
			{Name: "verbose"},
		},
		Commands: []Command{
			{Name: "listen"},
			{Name: "test-connection", GlobalFlags: []string{"config", "timeout", "verbose"}},
			{Name: "init-config", GlobalFlags: []string{}},
		},
	}

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "inherits all globals before command",
			args: []string{"--config", "site.yaml", "--format", "jsonl", "--verbose", "listen", "--duration", "30s"},
			want: []string{"--config", "site.yaml", "--format", "jsonl", "--verbose", "listen", "--duration", "30s"},
		},
		{
			name: "drops unsupported prefix global",
			args: []string{"--config", "site.yaml", "--format", "json", "--timeout", "5s", "test-connection"},
			want: []string{"--config", "site.yaml", "--timeout", "5s", "test-connection"},
		},
		{
			name: "drops every prefix global",
			args: []string{"--config", "site.yaml", "--verbose", "init-config", "--output", "new.yaml"},
			want: []string{"init-config", "--output", "new.yaml"},
		},
		{
			name: "unknown command keeps globals",
			args: []string{"--verbose", "unknown"},
			want: []string{"--verbose", "unknown"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := FilterGlobalFlagsForRegistry(test.args, registry)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("FilterGlobalFlagsForRegistry() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestNormalizeGlobalFlagsRejectsUnknownPrefixFlag(t *testing.T) {
	if _, err := NormalizeGlobalFlags([]string{"--endpoint", "x", "read"}, globalFlags); err == nil {
		t.Fatal("unknown global flag accepted")
	}
}

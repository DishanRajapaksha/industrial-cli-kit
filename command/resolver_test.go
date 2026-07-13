package command

import (
	"reflect"
	"testing"
)

func TestResolveReturnsDeepestNestedCommandAndInheritedPolicy(t *testing.T) {
	registry := Registry{
		GlobalFlags: []Flag{{Name: "config", TakesValue: true}, {Name: "endpoint", TakesValue: true}},
		Commands: []Command{{
			Name:        "send",
			GlobalFlags: []string{"config"},
			Subcommands: []Command{{
				Name: "transaction",
				Subcommands: []Command{{
					Name:        "start",
					LeadingArgs: 1,
				}},
			}},
		}},
	}

	got, ok := Resolve(registry, []string{"send", "transaction", "start", "meter-1", "--yes"})
	if !ok {
		t.Fatal("Resolve() did not match a known command path")
	}
	if want := []string{"send", "transaction", "start"}; !reflect.DeepEqual(got.Path, want) {
		t.Fatalf("path = %#v, want %#v", got.Path, want)
	}
	if got.Command.Name != "start" || got.Next != 3 {
		t.Fatalf("resolution = %#v", got)
	}
	if want := []string{"config"}; !reflect.DeepEqual(got.GlobalFlags, want) {
		t.Fatalf("global policy = %#v, want %#v", got.GlobalFlags, want)
	}
}

func TestResolvePreservesExplicitEmptyNestedPolicy(t *testing.T) {
	registry := Registry{Commands: []Command{{
		Name:        "send",
		GlobalFlags: []string{"config"},
		Subcommands: []Command{{Name: "reset", GlobalFlags: []string{}}},
	}}}

	got, ok := Resolve(registry, []string{"send", "reset"})
	if !ok {
		t.Fatal("Resolve() did not match")
	}
	if got.GlobalFlags == nil || len(got.GlobalFlags) != 0 {
		t.Fatalf("explicit empty policy was not preserved: %#v", got.GlobalFlags)
	}
}

func TestNormalizeGlobalFlagsForRegistryPlacesGlobalsAfterNestedPathAndArguments(t *testing.T) {
	registry := Registry{
		GlobalFlags: []Flag{{Name: "config", TakesValue: true}, {Name: "endpoint", TakesValue: true}, {Name: "verbose"}},
		Commands: []Command{{
			Name:        "send",
			GlobalFlags: []string{"config", "verbose"},
			Subcommands: []Command{{Name: "start", LeadingArgs: 1}},
		}},
	}

	got, err := NormalizeGlobalFlagsForRegistry([]string{
		"--config", "site.yaml", "--endpoint", "tcp://ignored", "--verbose",
		"send", "start", "meter-1", "--yes",
	}, registry)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"send", "start", "meter-1", "--config", "site.yaml", "--verbose", "--yes"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeGlobalFlagsForRegistry() = %#v, want %#v", got, want)
	}
}

func TestFilterGlobalFlagsForRegistryUsesNestedOverride(t *testing.T) {
	registry := Registry{
		GlobalFlags: []Flag{{Name: "config", TakesValue: true}, {Name: "endpoint", TakesValue: true}},
		Commands: []Command{{
			Name:        "send",
			GlobalFlags: []string{"config"},
			Subcommands: []Command{{Name: "status", GlobalFlags: []string{"endpoint"}}},
		}},
	}

	got, err := FilterGlobalFlagsForRegistry([]string{
		"--config", "site.yaml", "--endpoint", "tcp://host", "send", "status",
	}, registry)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"--endpoint", "tcp://host", "send", "status"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterGlobalFlagsForRegistry() = %#v, want %#v", got, want)
	}
}

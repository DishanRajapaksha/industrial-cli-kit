package contracttest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
)

func TestRegistryContractsAcceptAlignedRegistry(t *testing.T) {
	registry := command.Registry{
		Binary: "example-cli",
		GlobalFlags: []command.Flag{{Name: "config", TakesValue: true}},
		Commands: []command.Command{
			{Name: "init-config"},
			{Name: "validate-config"},
			{Name: "test-connection"},
			{Name: "status"},
			{Name: "send", Flags: []command.Flag{{Name: "yes"}}, Subcommands: []command.Command{{Name: "start", Flags: []command.Flag{{Name: "value", TakesValue: true}}}}},
			{Name: "completions"},
			{Name: "help"},
			{Name: "version"},
		},
	}
	names := registry.Names()
	Registry(t, registry, names)
	LifecycleRegistry(t, registry)
	CompletionRegistry(t, registry)
}

func TestRegistryReportsDrift(t *testing.T) {
	recorder := &recordingT{}
	registry := command.Registry{
		Binary: "example-cli",
		GlobalFlags: []command.Flag{{Name: "config"}, {Name: "config"}},
		Commands: []command.Command{
			{Name: "status"},
			{Name: "status"},
			{Name: "ghost"},
		},
	}
	Registry(recorder, registry, []string{"status", "missing"})
	joined := strings.Join(recorder.errors, "\n")
	for _, want := range []string{"duplicate command", "duplicate flag", "missing", "ghost"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("errors missing %q:\n%s", want, joined)
		}
	}
}

func TestLifecycleRegistryReportsMissingCommands(t *testing.T) {
	recorder := &recordingT{}
	LifecycleRegistry(recorder, command.Registry{Commands: []command.Command{{Name: "help"}}})
	if len(recorder.errors) != 6 {
		t.Fatalf("errors = %d, want 6: %v", len(recorder.errors), recorder.errors)
	}
}

type recordingT struct {
	errors []string
}

func (r *recordingT) Helper() {}

func (r *recordingT) Errorf(format string, args ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

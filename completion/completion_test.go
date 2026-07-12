package completion

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
)

var registry = command.Registry{
	Binary:      "example-cli",
	GlobalFlags: []command.Flag{{Name: "config", TakesValue: true}, {Name: "verbose"}},
	Commands: []command.Command{
		{Name: "read", Summary: "Read a value", Flags: []command.Flag{{Name: "node", TakesValue: true}}},
		{Name: "write-point", Summary: "Write a named point", LeadingArgs: 1, Flags: []command.Flag{{Name: "value", TakesValue: true}, {Name: "yes"}}},
		{
			Name:    "send",
			Summary: "Send an operation",
			Flags:   []command.Flag{{Name: "timeout", TakesValue: true}},
			Subcommands: []command.Command{
				{Name: "start", Summary: "Start an operation", Flags: []command.Flag{{Name: "yes"}}},
			},
		},
		{Name: "watch"},
	},
}

func TestWriteBashIncludesCommandsAndFlags(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, "bash", registry); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"read", "watch", "send", "start", "write-point", "--config", "--verbose", "--node", "--timeout", "--value", "--yes"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("script does not include %q: %s", want, out.String())
		}
	}
	if !strings.Contains(out.String(), "send:start") {
		t.Fatalf("bash script does not dispatch nested commands: %s", out.String())
	}
	if strings.Contains(out.String(), "write-point:") {
		t.Fatalf("leading argument command was treated as a nested command: %s", out.String())
	}
}

func TestWriteZshIncludesSummariesAndNestedFlags(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, "zsh", registry); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"read:Read a value", "send:Send an operation", "start:Start an operation", "write-point:Write a named point", "--node", "--value", "--yes"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("script does not include %q: %s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "write-point:") && strings.Contains(out.String(), "write-point:active") {
		t.Fatalf("leading argument command was treated as a nested command: %s", out.String())
	}
}

func TestWriteUsesShellSafeFunctionNameAndRealBinary(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, "bash", registry); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "_example_cli_completion") {
		t.Fatalf("bash function name was not normalised: %s", out.String())
	}
	if !strings.Contains(out.String(), "complete -F _example_cli_completion example-cli") {
		t.Fatalf("bash completion was not registered for the real binary: %s", out.String())
	}
}

func TestWriteRejectsUnknownShell(t *testing.T) {
	if err := Write(&bytes.Buffer{}, "fish", registry); err == nil {
		t.Fatal("unknown shell accepted")
	}
}

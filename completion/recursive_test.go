package completion

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
)

var recursiveRegistry = command.Registry{
	Binary: "example-cli",
	GlobalFlags: []command.Flag{
		{Name: "config", TakesValue: true},
		{Name: "endpoint", TakesValue: true},
		{Name: "verbose"},
	},
	Commands: []command.Command{
		{Name: "read", Summary: "Read a value", Flags: []command.Flag{{Name: "node", TakesValue: true}}},
		{Name: "validate-config", Summary: "Validate configuration", GlobalFlags: []string{"config", "verbose"}},
		{Name: "init-config", Summary: "Initialise configuration", GlobalFlags: []string{}, Flags: []command.Flag{{Name: "output", TakesValue: true}}},
		{Name: "write-point", Summary: "Write a named point", LeadingArgs: 1, Flags: []command.Flag{{Name: "value", TakesValue: true}, {Name: "yes"}}},
		{
			Name:    "send",
			Summary: "Send an operation",
			Flags:   []command.Flag{{Name: "timeout", TakesValue: true}},
			Subcommands: []command.Command{{
				Name:    "transaction",
				Summary: "Send a transaction operation",
				Subcommands: []command.Command{{
					Name:    "start",
					Summary: "Start an operation",
					Flags:   []command.Flag{{Name: "yes"}},
				}},
			}},
		},
		{Name: "watch"},
	},
}

func TestWriteBashIncludesRecursiveCommands(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, "bash", recursiveRegistry); err != nil {
		t.Fatal(err)
	}
	script := out.String()
	for _, want := range []string{
		"send:transaction:start",
		`send:transaction:start) words="--config --endpoint --timeout --verbose --yes" ;;`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("bash recursive completion missing %q: %s", want, script)
		}
	}
}

func TestWriteBashFindsCommandAfterPrefixGlobals(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, "bash", recursiveRegistry); err != nil {
		t.Fatal(err)
	}
	script := out.String()
	for _, want := range []string{
		`for ((i=1; i<COMP_CWORD; i++))`,
		`--config|--endpoint|--node|--output|--timeout|--value) expect_value=1`,
		`candidate="$key:$token"`,
		`"") words="--config --endpoint --verbose init-config read send validate-config watch write-point" ;;`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("bash prefix-global resolver missing %q: %s", want, script)
		}
	}
	if strings.Contains(script, `command="${COMP_WORDS[1]}"`) {
		t.Fatalf("bash completion still assumes a fixed command index: %s", script)
	}
}

func TestWriteZshIncludesRecursiveCommandsAndScansWords(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, "zsh", recursiveRegistry); err != nil {
		t.Fatal(err)
	}
	script := out.String()
	for _, want := range []string{
		"transaction:Send a transaction operation",
		"start:Start an operation",
		"send:transaction:start",
		`for ((i=2; i<CURRENT; i++))`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("zsh recursive completion missing %q: %s", want, script)
		}
	}
	if strings.Contains(script, `command="$words[2]"`) {
		t.Fatalf("zsh completion still assumes a fixed command index: %s", script)
	}
}

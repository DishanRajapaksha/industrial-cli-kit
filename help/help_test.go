package help

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
)

var testRegistry = command.Registry{
	Binary: "example-cli",
	GlobalFlags: []command.Flag{
		{Name: "config", TakesValue: true, Summary: "config file"},
		{Name: "verbose", Summary: "verbose diagnostics"},
	},
	Commands: []command.Command{
		{Name: "status", Summary: "show status"},
		{Name: "send", Summary: "send an operation", Subcommands: []command.Command{
			{Name: "start", Summary: "start operation"},
			{Name: "stop", Summary: "stop operation"},
		}},
	},
}

func TestWriteIncludesRegistryMetadata(t *testing.T) {
	var out bytes.Buffer
	err := Write(&out, testRegistry, Options{
		Description: "Example client.",
		Examples:    []string{"example-cli status"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Example client.",
		"example-cli [global flags] <command> [flags]",
		"example-cli status",
		"status",
		"send start",
		"send stop",
		"--config <value>",
		"--verbose",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("help output missing %q:\n%s", want, out.String())
		}
	}
}

func TestWriteUsesCustomUsage(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, testRegistry, Options{Usage: []string{"example-cli status [flags]"}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "example-cli status [flags]") {
		t.Fatalf("custom usage missing:\n%s", out.String())
	}
	if strings.Contains(out.String(), "example-cli [global flags] <command> [flags]") {
		t.Fatalf("default usage unexpectedly rendered:\n%s", out.String())
	}
}

func TestWriteRejectsMissingBinary(t *testing.T) {
	if err := Write(&bytes.Buffer{}, command.Registry{}, Options{}); err == nil {
		t.Fatal("missing binary accepted")
	}
}

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
	Commands:    []command.Command{{Name: "read"}, {Name: "watch"}},
}

func TestWriteBashIncludesCommandsAndFlags(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, "bash", registry); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"read", "watch", "--config", "--verbose"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("script does not include %q: %s", want, out.String())
		}
	}
}

func TestWriteRejectsUnknownShell(t *testing.T) {
	if err := Write(&bytes.Buffer{}, "fish", registry); err == nil {
		t.Fatal("unknown shell accepted")
	}
}

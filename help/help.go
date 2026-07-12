// Package help renders human-readable usage from a command registry.
package help

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
)

// Options provides CLI-specific prose while command and flag metadata remain
// sourced from the shared registry.
type Options struct {
	Description string
	Usage       []string
	Examples    []string
}

// Write renders a complete root help page.
func Write(w io.Writer, registry command.Registry, options Options) error {
	if strings.TrimSpace(registry.Binary) == "" {
		return fmt.Errorf("registry binary is required")
	}
	if description := strings.TrimSpace(options.Description); description != "" {
		if _, err := fmt.Fprintln(w, description); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	usage := options.Usage
	if len(usage) == 0 {
		usage = []string{registry.Binary + " [global flags] <command> [flags]"}
	}
	if err := writeLines(w, "Usage:", usage); err != nil {
		return err
	}
	if len(options.Examples) > 0 {
		if err := writeLines(w, "Examples:", options.Examples); err != nil {
			return err
		}
	}
	if len(registry.Commands) > 0 {
		if err := WriteCommands(w, registry.Commands); err != nil {
			return err
		}
	}
	if len(registry.GlobalFlags) > 0 {
		if err := WriteGlobalFlags(w, registry.GlobalFlags); err != nil {
			return err
		}
	}
	return nil
}

// WriteCommands renders top-level commands and one nested subcommand level.
func WriteCommands(w io.Writer, commands []command.Command) error {
	if _, err := fmt.Fprintln(w, "Commands:"); err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for _, registered := range commands {
		if err := writeCommand(tw, "  ", registered); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w)
	return err
}

// WriteGlobalFlags renders global flags in registry order.
func WriteGlobalFlags(w io.Writer, flags []command.Flag) error {
	if _, err := fmt.Fprintln(w, "Global flags:"); err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for _, registered := range flags {
		name := "--" + registered.Name
		if registered.TakesValue {
			name += " <value>"
		}
		if _, err := fmt.Fprintf(tw, "  %s\t%s\n", name, registered.Summary); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w)
	return err
}

func writeCommand(w io.Writer, indent string, registered command.Command) error {
	if _, err := fmt.Fprintf(w, "%s%s\t%s\n", indent, registered.Name, registered.Summary); err != nil {
		return err
	}
	for _, nested := range registered.Subcommands {
		if _, err := fmt.Fprintf(w, "%s  %s %s\t%s\n", indent, registered.Name, nested.Name, nested.Summary); err != nil {
			return err
		}
	}
	return nil
}

func writeLines(w io.Writer, heading string, lines []string) error {
	if _, err := fmt.Fprintln(w, heading); err != nil {
		return err
	}
	for _, line := range lines {
		if _, err := fmt.Fprintf(w, "  %s\n", line); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

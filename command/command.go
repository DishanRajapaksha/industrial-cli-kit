// Package command defines reusable command metadata and global-flag parsing.
package command

import (
	"fmt"
	"strings"
)

// Flag describes a command-line flag without its leading dashes.
type Flag struct {
	Name       string
	TakesValue bool
	Summary    string
}

// Command is a public command and optional nested subcommands.
type Command struct {
	Name        string
	Summary     string
	Flags       []Flag
	Subcommands []Command
}

// Registry is the declarative public shape of a CLI.
type Registry struct {
	Binary      string
	GlobalFlags []Flag
	Commands    []Command
}

// Names returns the registered top-level command names in declaration order.
func (r Registry) Names() []string {
	names := make([]string, 0, len(r.Commands))
	for _, command := range r.Commands {
		names = append(names, command.Name)
	}
	return names
}

// NormalizeGlobalFlags moves recognized global flags from before the command to
// immediately after it, allowing command-local flag parsing without changing
// the public invocation form.
func NormalizeGlobalFlags(args []string, flags []Flag) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
	}

	known := make(map[string]Flag, len(flags))
	for _, flag := range flags {
		known["--"+flag.Name] = flag
	}

	var globals []string
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--" {
			if index+1 >= len(args) {
				return nil, fmt.Errorf("command is required after --")
			}
			return appendAfterCommand(args[index+1:], globals), nil
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			return appendAfterCommand(args[index:], globals), nil
		}
		if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
			return args[index:], nil
		}

		name, value, inline := strings.Cut(arg, "=")
		flag, ok := known[name]
		if !ok {
			return nil, fmt.Errorf("unknown global flag %q", name)
		}
		if !flag.TakesValue {
			if inline {
				return nil, fmt.Errorf("%s does not take a value", name)
			}
			globals = append(globals, name)
			continue
		}
		if !inline {
			index++
			if index >= len(args) || strings.HasPrefix(args[index], "-") {
				return nil, fmt.Errorf("%s requires a value", name)
			}
			value = args[index]
		}
		if value == "" {
			return nil, fmt.Errorf("%s requires a value", name)
		}
		globals = append(globals, name, value)
	}
	return nil, fmt.Errorf("command is required")
}

func appendAfterCommand(args, globals []string) []string {
	if len(args) == 0 || len(globals) == 0 {
		return args
	}
	normalized := make([]string, 0, len(args)+len(globals))
	normalized = append(normalized, args[0])
	normalized = append(normalized, globals...)
	normalized = append(normalized, args[1:]...)
	return normalized
}

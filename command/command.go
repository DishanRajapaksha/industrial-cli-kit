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

	// LeadingArgs is the maximum number of positional arguments that should
	// remain before flags. Only actual non-flag tokens are skipped, allowing the
	// same metadata to describe required and optional leading arguments.
	LeadingArgs int
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
	return normalizeGlobalFlags(args, flags, func(_ []string) int { return 1 })
}

// NormalizeGlobalFlagsForRegistry moves recognized global flags after the
// command's declared positional prefix. Unknown commands retain the traditional
// placement immediately after the top-level command so the caller can report
// the unknown command normally.
func NormalizeGlobalFlagsForRegistry(args []string, registry Registry) ([]string, error) {
	return normalizeGlobalFlags(args, registry.GlobalFlags, func(commandArgs []string) int {
		if len(commandArgs) == 0 {
			return 0
		}
		for _, registered := range registry.Commands {
			if registered.Name != commandArgs[0] {
				continue
			}
			index := 1
			for remaining := registered.LeadingArgs; remaining > 0 && index < len(commandArgs); remaining-- {
				if isFlagToken(commandArgs[index]) {
					break
				}
				index++
			}
			return index
		}
		return 1
	})
}

func normalizeGlobalFlags(args []string, flags []Flag, insertionIndex func([]string) int) ([]string, error) {
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
			return insertGlobals(args[index+1:], globals, insertionIndex), nil
		}
		if !isFlagToken(arg) {
			return insertGlobals(args[index:], globals, insertionIndex), nil
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
			if index >= len(args) || isFlagToken(args[index]) {
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

func isFlagToken(value string) bool {
	return value != "-" && strings.HasPrefix(value, "-")
}

func insertGlobals(args, globals []string, insertionIndex func([]string) int) []string {
	if len(args) == 0 || len(globals) == 0 {
		return args
	}
	index := insertionIndex(args)
	if index < 0 {
		index = 0
	}
	if index > len(args) {
		index = len(args)
	}
	normalized := make([]string, 0, len(args)+len(globals))
	normalized = append(normalized, args[:index]...)
	normalized = append(normalized, globals...)
	normalized = append(normalized, args[index:]...)
	return normalized
}

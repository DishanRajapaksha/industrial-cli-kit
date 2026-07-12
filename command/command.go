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

	// AllowEmpty permits an explicitly supplied empty value. It is ignored for
	// boolean flags and should only be set when an empty string has domain
	// meaning, such as an empty CIP route path.
	AllowEmpty bool
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

	// GlobalFlags restricts which registry global flags apply to this command.
	// Nil inherits every global flag. A non-nil empty slice allows none. Names
	// are written without leading dashes.
	GlobalFlags []string
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
	return normalizeGlobalFlags(args, flags, func(commandArgs, globals []string) []string {
		return insertGlobalsAt(commandArgs, globals, 1)
	})
}

// NormalizeGlobalFlagsForRegistry moves recognized global flags after the
// command's declared positional prefix and applies its global-flag policy.
// Unknown commands retain every global flag immediately after the top-level
// command so the caller can report the unknown command normally.
func NormalizeGlobalFlagsForRegistry(args []string, registry Registry) ([]string, error) {
	return normalizeGlobalFlags(args, registry.GlobalFlags, func(commandArgs, globals []string) []string {
		if len(commandArgs) == 0 {
			return commandArgs
		}
		for _, registered := range registry.Commands {
			if registered.Name != commandArgs[0] {
				continue
			}
			globals = filterGlobalArguments(globals, registry.GlobalFlags, registered.GlobalFlags)
			index := 1
			for remaining := registered.LeadingArgs; remaining > 0 && index < len(commandArgs); remaining-- {
				if isFlagToken(commandArgs[index]) {
					break
				}
				index++
			}
			return insertGlobalsAt(commandArgs, globals, index)
		}
		return insertGlobalsAt(commandArgs, globals, 1)
	})
}

// FilterGlobalFlagsForRegistry keeps recognized global flags before the command
// while applying its registry policy. This supports CLIs that parse typed global
// options before dispatching to command-specific flag sets.
func FilterGlobalFlagsForRegistry(args []string, registry Registry) ([]string, error) {
	return normalizeGlobalFlags(args, registry.GlobalFlags, func(commandArgs, globals []string) []string {
		if len(commandArgs) > 0 {
			for _, registered := range registry.Commands {
				if registered.Name == commandArgs[0] {
					globals = filterGlobalArguments(globals, registry.GlobalFlags, registered.GlobalFlags)
					break
				}
			}
		}
		filtered := make([]string, 0, len(globals)+len(commandArgs))
		filtered = append(filtered, globals...)
		filtered = append(filtered, commandArgs...)
		return filtered
	})
}

func normalizeGlobalFlags(args []string, flags []Flag, place func([]string, []string) []string) ([]string, error) {
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
			return place(args[index+1:], globals), nil
		}
		if !isFlagToken(arg) {
			return place(args[index:], globals), nil
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
		if value == "" && !flag.AllowEmpty {
			return nil, fmt.Errorf("%s requires a value", name)
		}
		globals = append(globals, name, value)
	}
	return nil, fmt.Errorf("command is required")
}

func filterGlobalArguments(arguments []string, definitions []Flag, allowed []string) []string {
	if allowed == nil {
		return arguments
	}
	allowedNames := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedNames["--"+strings.TrimPrefix(name, "--")] = struct{}{}
	}
	known := make(map[string]Flag, len(definitions))
	for _, flag := range definitions {
		known["--"+flag.Name] = flag
	}

	filtered := make([]string, 0, len(arguments))
	for index := 0; index < len(arguments); index++ {
		name := arguments[index]
		flag := known[name]
		_, keep := allowedNames[name]
		if keep {
			filtered = append(filtered, name)
		}
		if flag.TakesValue {
			index++
			if index < len(arguments) && keep {
				filtered = append(filtered, arguments[index])
			}
		}
	}
	return filtered
}

func isFlagToken(value string) bool {
	return value != "-" && strings.HasPrefix(value, "-")
}

func insertGlobalsAt(args, globals []string, index int) []string {
	if len(args) == 0 || len(globals) == 0 {
		return args
	}
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

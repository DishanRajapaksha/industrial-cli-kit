// Package completion generates shell completions from a command registry.
package completion

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
)

// Write emits a Bash or Zsh completion script for registry.
func Write(w io.Writer, shell string, registry command.Registry) error {
	switch shell {
	case "bash":
		_, err := io.WriteString(w, bash(registry))
		return err
	case "zsh":
		_, err := io.WriteString(w, zsh(registry))
		return err
	default:
		return fmt.Errorf("unsupported shell %q; expected bash or zsh", shell)
	}
}

func bash(registry command.Registry) string {
	var cases strings.Builder
	for _, cmd := range registry.Commands {
		globals := applicableGlobalFlags(registry.GlobalFlags, cmd.GlobalFlags)
		words := mergeWords(flagNames(globals), flagNames(cmd.Flags))
		if len(cmd.Subcommands) > 0 {
			words = mergeWords(commandNames(cmd.Subcommands), words)
		}
		fmt.Fprintf(&cases, "    %s) words=%q ;;\n", cmd.Name, strings.Join(words, " "))
		for _, sub := range cmd.Subcommands {
			nestedGlobals := globals
			if sub.GlobalFlags != nil {
				nestedGlobals = applicableGlobalFlags(registry.GlobalFlags, sub.GlobalFlags)
			}
			nested := mergeWords(flagNames(nestedGlobals), flagNames(cmd.Flags), flagNames(sub.Flags))
			fmt.Fprintf(&cases, "    %s:%s) words=%q ;;\n", cmd.Name, sub.Name, strings.Join(nested, " "))
		}
	}

	return fmt.Sprintf(`_%[1]s_completion() {
  local cur command nested key words
  cur="${COMP_WORDS[COMP_CWORD]}"
  command="${COMP_WORDS[1]}"
  nested="${COMP_WORDS[2]}"

  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=( $(compgen -W "%[3]s" -- "$cur") )
    return 0
  fi

  key="$command"
  if [ "$COMP_CWORD" -gt 2 ] && [[ "$nested" != -* ]] && [[ " %[6]s " == *" $command "* ]]; then
    key="$command:$nested"
  fi

  case "$key" in
%[4]s    *) words="%[5]s" ;;
  esac
  COMPREPLY=( $(compgen -W "$words" -- "$cur") )
}
complete -F _%[1]s_completion %[2]s
`, shellName(registry.Binary), registry.Binary, strings.Join(registry.Names(), " "), cases.String(), strings.Join(flagNames(registry.GlobalFlags), " "), strings.Join(commandsWithSubcommands(registry.Commands), " "))
}

func zsh(registry command.Registry) string {
	var commandSpecs []string
	var cases strings.Builder
	for _, cmd := range registry.Commands {
		summary := cmd.Summary
		if summary == "" {
			summary = cmd.Name
		}
		commandSpecs = append(commandSpecs, fmt.Sprintf("'%s:%s'", cmd.Name, escapeZsh(summary)))

		globals := applicableGlobalFlags(registry.GlobalFlags, cmd.GlobalFlags)
		flags := mergeWords(flagNames(globals), flagNames(cmd.Flags))
		if len(cmd.Subcommands) > 0 {
			var subSpecs []string
			for _, sub := range cmd.Subcommands {
				subSummary := sub.Summary
				if subSummary == "" {
					subSummary = sub.Name
				}
				subSpecs = append(subSpecs, fmt.Sprintf("'%s:%s'", sub.Name, escapeZsh(subSummary)))
			}
			fmt.Fprintf(&cases, "    %s)\n      if (( CURRENT == 3 )); then\n        _describe 'subcommand' '(%s)'\n      else\n        _values 'flag' %s\n      fi\n      ;;\n", cmd.Name, strings.Join(subSpecs, " "), quoteWords(strings.Join(flags, " ")))
			for _, sub := range cmd.Subcommands {
				nestedGlobals := globals
				if sub.GlobalFlags != nil {
					nestedGlobals = applicableGlobalFlags(registry.GlobalFlags, sub.GlobalFlags)
				}
				nested := mergeWords(flagNames(nestedGlobals), flagNames(cmd.Flags), flagNames(sub.Flags))
				fmt.Fprintf(&cases, "    %s:%s) _values 'flag' %s ;;\n", cmd.Name, sub.Name, quoteWords(strings.Join(nested, " ")))
			}
			continue
		}
		fmt.Fprintf(&cases, "    %s) _values 'flag' %s ;;\n", cmd.Name, quoteWords(strings.Join(flags, " ")))
	}

	return fmt.Sprintf(`#compdef %[1]s

_%[2]s_completion() {
  local command nested key
  command="$words[2]"
  nested="$words[3]"

  if (( CURRENT == 2 )); then
    local -a commands
    commands=(%[3]s)
    _describe 'command' commands
    return
  fi

  key="$command"
  if (( CURRENT > 3 )) && [[ "$nested" != -* ]] && [[ " %[6]s " == *" $command "* ]]; then
    key="$command:$nested"
  fi

  case "$key" in
%[4]s    *) _values 'flag' %[5]s ;;
  esac
}
_%[2]s_completion
`, registry.Binary, shellName(registry.Binary), strings.Join(commandSpecs, " "), cases.String(), quoteWords(strings.Join(flagNames(registry.GlobalFlags), " ")), strings.Join(commandsWithSubcommands(registry.Commands), " "))
}

func commandNames(commands []command.Command) []string {
	names := make([]string, 0, len(commands))
	for _, cmd := range commands {
		names = append(names, cmd.Name)
	}
	return names
}

func commandsWithSubcommands(commands []command.Command) []string {
	names := make([]string, 0, len(commands))
	for _, cmd := range commands {
		if len(cmd.Subcommands) > 0 {
			names = append(names, cmd.Name)
		}
	}
	return names
}

func applicableGlobalFlags(flags []command.Flag, policy []string) []command.Flag {
	if policy == nil {
		return flags
	}
	allowed := make(map[string]struct{}, len(policy))
	for _, name := range policy {
		allowed[strings.TrimPrefix(name, "--")] = struct{}{}
	}
	filtered := make([]command.Flag, 0, len(policy))
	for _, flag := range flags {
		if _, ok := allowed[flag.Name]; ok {
			filtered = append(filtered, flag)
		}
	}
	return filtered
}

func flagNames(flags []command.Flag) []string {
	names := make([]string, 0, len(flags))
	for _, flag := range flags {
		names = append(names, "--"+flag.Name)
	}
	return names
}

func mergeWords(groups ...[]string) []string {
	seen := make(map[string]struct{})
	for _, group := range groups {
		for _, word := range group {
			if word != "" {
				seen[word] = struct{}{}
			}
		}
	}
	words := make([]string, 0, len(seen))
	for word := range seen {
		words = append(words, word)
	}
	sort.Strings(words)
	return words
}

func shellName(binary string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_")
	return replacer.Replace(binary)
}

func quoteWords(words string) string {
	if words == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(words, " ", "' '") + "'"
}

func escapeZsh(value string) string {
	return strings.ReplaceAll(value, "'", "'\\''")
}

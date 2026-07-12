// Package completion generates shell completions from a command registry.
package completion

import (
	"fmt"
	"io"
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
	commands := strings.Join(registry.Names(), " ")
	flags := strings.Join(flagNames(registry.GlobalFlags), " ")
	return fmt.Sprintf(`_%[1]s_completion() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=( $(compgen -W "%[2]s" -- "$cur") )
    return 0
  fi
  COMPREPLY=( $(compgen -W "%[3]s" -- "$cur") )
}
complete -F _%[1]s_completion %[1]s
`, registry.Binary, commands, flags)
}

func zsh(registry command.Registry) string {
	commands := strings.Join(registry.Names(), " ")
	flags := strings.Join(flagNames(registry.GlobalFlags), " ")
	return fmt.Sprintf(`#compdef %[1]s

_%[1]s_completion() {
  if (( CURRENT == 2 )); then
    _values 'command' %[2]s
  else
    _values 'flag' %[3]s
  fi
}
_%[1]s_completion
`, registry.Binary, quoteWords(commands), quoteWords(flags))
}

func flagNames(flags []command.Flag) []string {
	names := make([]string, 0, len(flags))
	for _, flag := range flags {
		names = append(names, "--"+flag.Name)
	}
	return names
}

func quoteWords(words string) string {
	if words == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(words, " ", "' '") + "'"
}

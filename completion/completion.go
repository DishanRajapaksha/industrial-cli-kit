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

type completionContext struct {
	key      string
	children []command.Command
	flags    []string
}

func bash(registry command.Registry) string {
	contexts := buildContexts(registry)
	var cases strings.Builder
	for _, context := range contexts {
		words := mergeWords(commandNames(context.children), context.flags)
		fmt.Fprintf(&cases, "    %s) words=%q ;;\n", context.key, strings.Join(words, " "))
	}

	rootWords := mergeWords(commandNames(registry.Commands), flagNames(registry.GlobalFlags))
	paths := contextKeys(contexts)
	exactValueFlags, inlineValueFlags := valueFlagPatterns(registry)
	valueCases := ""
	if exactValueFlags != "" {
		valueCases = fmt.Sprintf(`    %s) expect_value=1; continue ;;
    %s) continue ;;
`, exactValueFlags, inlineValueFlags)
	}

	return fmt.Sprintf(`_%[1]s_completion() {
  local cur key candidate token words expect_value path_open i
  cur="${COMP_WORDS[COMP_CWORD]}"
  key=""
  expect_value=0
  path_open=1

  for ((i=1; i<COMP_CWORD; i++)); do
    token="${COMP_WORDS[i]}"
    if (( expect_value )); then
      expect_value=0
      continue
    fi
    case "$token" in
%[5]s    esac
    if [[ "$token" == -* ]]; then
      continue
    fi
    if (( ! path_open )); then
      continue
    fi
    if [[ -z "$key" ]]; then
      candidate="$token"
    else
      candidate="$key:$token"
    fi
    case "$candidate" in
      %[6]s) key="$candidate" ;;
      *) path_open=0 ;;
    esac
  done

  if (( expect_value )); then
    COMPREPLY=()
    return 0
  fi

  case "$key" in
    "") words=%[4]q ;;
%[3]s    *) words="" ;;
  esac
  COMPREPLY=( $(compgen -W "$words" -- "$cur") )
}
complete -F _%[1]s_completion %[2]s
`, shellName(registry.Binary), registry.Binary, cases.String(), strings.Join(rootWords, " "), valueCases, commandPathPattern(paths))
}

func zsh(registry command.Registry) string {
	contexts := buildContexts(registry)
	var cases strings.Builder
	for _, context := range contexts {
		values := make([]string, 0, len(context.children)+len(context.flags))
		for _, child := range context.children {
			summary := child.Summary
			if summary == "" {
				summary = child.Name
			}
			values = append(values, child.Name+":"+summary)
		}
		values = append(values, context.flags...)
		label := "flag"
		if len(context.children) > 0 {
			label = "command or flag"
		}
		fmt.Fprintf(&cases, "    %s) _values %s %s ;;\n", context.key, quoteZsh(label), quoteZshWords(values))
	}

	rootValues := make([]string, 0, len(registry.Commands)+len(registry.GlobalFlags))
	for _, registered := range registry.Commands {
		summary := registered.Summary
		if summary == "" {
			summary = registered.Name
		}
		rootValues = append(rootValues, registered.Name+":"+summary)
	}
	rootValues = append(rootValues, flagNames(registry.GlobalFlags)...)

	paths := contextKeys(contexts)
	exactValueFlags, inlineValueFlags := valueFlagPatterns(registry)
	valueCases := ""
	if exactValueFlags != "" {
		valueCases = fmt.Sprintf(`      %s) expect_value=1; continue ;;
      %s) continue ;;
`, exactValueFlags, inlineValueFlags)
	}

	return fmt.Sprintf(`#compdef %[1]s

_%[2]s_completion() {
  local key candidate token expect_value path_open i
  key=""
  expect_value=0
  path_open=1

  for ((i=2; i<CURRENT; i++)); do
    token="${words[i]}"
    if (( expect_value )); then
      expect_value=0
      continue
    fi
    case "$token" in
%[5]s    esac
    if [[ "$token" == -* ]]; then
      continue
    fi
    if (( ! path_open )); then
      continue
    fi
    if [[ -z "$key" ]]; then
      candidate="$token"
    else
      candidate="$key:$token"
    fi
    case "$candidate" in
      %[6]s) key="$candidate" ;;
      *) path_open=0 ;;
    esac
  done

  if (( expect_value )); then
    return 0
  fi

  case "$key" in
    '') _values 'command or global flag' %[4]s ;;
%[3]s  esac
}
_%[2]s_completion
`, registry.Binary, shellName(registry.Binary), cases.String(), quoteZshWords(rootValues), valueCases, commandPathPattern(paths))
}

func buildContexts(registry command.Registry) []completionContext {
	var contexts []completionContext
	walkContexts(&contexts, registry, registry.Commands, nil, nil, registry.GlobalFlags)
	return contexts
}

func walkContexts(contexts *[]completionContext, registry command.Registry, commands []command.Command, path []string, inheritedFlags, globals []command.Flag) {
	for _, registered := range commands {
		currentGlobals := globals
		if registered.GlobalFlags != nil {
			currentGlobals = applicableGlobalFlags(registry.GlobalFlags, registered.GlobalFlags)
		}
		currentPath := append(append([]string(nil), path...), registered.Name)
		currentFlags := mergeWords(flagNames(currentGlobals), flagNames(inheritedFlags), flagNames(registered.Flags))
		*contexts = append(*contexts, completionContext{
			key:      strings.Join(currentPath, ":"),
			children: registered.Subcommands,
			flags:    currentFlags,
		})
		nextInherited := append(append([]command.Flag(nil), inheritedFlags...), registered.Flags...)
		walkContexts(contexts, registry, registered.Subcommands, currentPath, nextInherited, currentGlobals)
	}
}

func contextKeys(contexts []completionContext) []string {
	keys := make([]string, 0, len(contexts))
	for _, context := range contexts {
		keys = append(keys, context.key)
	}
	return keys
}

func commandPathPattern(paths []string) string {
	if len(paths) == 0 {
		return "__industrial_cli_no_command__"
	}
	return strings.Join(paths, "|")
}

func valueFlagPatterns(registry command.Registry) (exact string, inline string) {
	values := map[string]bool{}
	for _, flag := range registry.GlobalFlags {
		if flag.TakesValue {
			values["--"+flag.Name] = true
		}
	}
	collectValueFlags(values, registry.Commands)
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	inlineNames := make([]string, len(names))
	for index, name := range names {
		inlineNames[index] = name + "=*"
	}
	return strings.Join(names, "|"), strings.Join(inlineNames, "|")
}

func collectValueFlags(values map[string]bool, commands []command.Command) {
	for _, registered := range commands {
		for _, flag := range registered.Flags {
			if flag.TakesValue {
				values["--"+flag.Name] = true
			}
		}
		collectValueFlags(values, registered.Subcommands)
	}
}

func commandNames(commands []command.Command) []string {
	names := make([]string, 0, len(commands))
	for _, cmd := range commands {
		names = append(names, cmd.Name)
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

func quoteZshWords(words []string) string {
	if len(words) == 0 {
		return "''"
	}
	quoted := make([]string, len(words))
	for index, word := range words {
		quoted[index] = quoteZsh(word)
	}
	return strings.Join(quoted, " ")
}

func quoteZsh(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

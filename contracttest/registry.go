package contracttest

import (
	"bytes"
	"strings"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
	"github.com/DishanRajapaksha/industrial-cli-kit/completion"
)

// Registry verifies structural invariants shared by all CLI command registries.
// dispatched is the list of top-level names accepted by the real dispatcher.
func Registry(t TestingT, registry command.Registry, dispatched []string) {
	t.Helper()
	if strings.TrimSpace(registry.Binary) == "" {
		t.Errorf("registry binary is empty")
	}

	registered := make(map[string]bool, len(registry.Commands))
	for _, registeredCommand := range registry.Commands {
		if strings.TrimSpace(registeredCommand.Name) == "" {
			t.Errorf("registry contains an empty command name")
			continue
		}
		if registered[registeredCommand.Name] {
			t.Errorf("registry contains duplicate command %q", registeredCommand.Name)
		}
		registered[registeredCommand.Name] = true
		validateFlags(t, registeredCommand.Name, registeredCommand.Flags)

		nested := map[string]bool{}
		for _, subcommand := range registeredCommand.Subcommands {
			if strings.TrimSpace(subcommand.Name) == "" {
				t.Errorf("command %q contains an empty subcommand name", registeredCommand.Name)
				continue
			}
			if nested[subcommand.Name] {
				t.Errorf("command %q contains duplicate subcommand %q", registeredCommand.Name, subcommand.Name)
			}
			nested[subcommand.Name] = true
			validateFlags(t, registeredCommand.Name+" "+subcommand.Name, subcommand.Flags)
		}
	}
	validateFlags(t, "global", registry.GlobalFlags)

	dispatchedSet := make(map[string]bool, len(dispatched))
	for _, name := range dispatched {
		if dispatchedSet[name] {
			t.Errorf("dispatcher list contains duplicate command %q", name)
		}
		dispatchedSet[name] = true
		if !registered[name] {
			t.Errorf("dispatcher command %q is not registered", name)
		}
	}
	for name := range registered {
		if !dispatchedSet[name] {
			t.Errorf("registered command %q is not dispatched", name)
		}
	}
}

// LifecycleRegistry verifies the common lifecycle command names.
func LifecycleRegistry(t TestingT, registry command.Registry) {
	t.Helper()
	required := []string{"init-config", "validate-config", "test-connection", "status", "completions", "help", "version"}
	registered := map[string]bool{}
	for _, registeredCommand := range registry.Commands {
		registered[registeredCommand.Name] = true
	}
	for _, name := range required {
		if !registered[name] {
			t.Errorf("registry missing lifecycle command %q", name)
		}
	}
}

// CompletionRegistry verifies that generated Bash and Zsh scripts expose every
// registered command, nested command, and flag.
func CompletionRegistry(t TestingT, registry command.Registry) {
	t.Helper()
	for _, shell := range []string{"bash", "zsh"} {
		var out bytes.Buffer
		if err := completion.Write(&out, shell, registry); err != nil {
			t.Errorf("generate %s completions: %v", shell, err)
			continue
		}
		script := out.String()
		for _, registeredCommand := range registry.Commands {
			requireContains(t, shell, script, registeredCommand.Name)
			for _, flag := range registeredCommand.Flags {
				requireContains(t, shell, script, "--"+flag.Name)
			}
			for _, subcommand := range registeredCommand.Subcommands {
				requireContains(t, shell, script, subcommand.Name)
				for _, flag := range subcommand.Flags {
					requireContains(t, shell, script, "--"+flag.Name)
				}
			}
		}
		for _, flag := range registry.GlobalFlags {
			requireContains(t, shell, script, "--"+flag.Name)
		}
	}
}

func validateFlags(t TestingT, owner string, flags []command.Flag) {
	seen := map[string]bool{}
	for _, flag := range flags {
		if strings.TrimSpace(flag.Name) == "" {
			t.Errorf("%s contains an empty flag name", owner)
			continue
		}
		if seen[flag.Name] {
			t.Errorf("%s contains duplicate flag --%s", owner, flag.Name)
		}
		seen[flag.Name] = true
	}
}

func requireContains(t TestingT, shell, script, value string) {
	if !strings.Contains(script, value) {
		t.Errorf("%s completion script missing %q", shell, value)
	}
}

package optargs

import (
	"testing"
)

func TestNewParserWithCaseInsensitiveCommands(t *testing.T) {
	parser, err := NewParserWithCaseInsensitiveCommands(
		map[byte]*Flag{}, map[string]*Flag{}, []string{"a", "b"},
	)
	if err != nil {
		t.Fatalf("NewParserWithCaseInsensitiveCommands: %v", err)
	}
	if !parser.config.commandCaseIgnore {
		t.Error("commandCaseIgnore should be true")
	}
	if len(parser.Args) != 2 || parser.Args[0] != "a" || parser.Args[1] != "b" {
		t.Errorf("Args = %v, want [a b]", parser.Args)
	}
}

func TestNewParserWithCaseInsensitiveCommandsParent(t *testing.T) {
	parent := newMinimalParser(t)
	child, err := NewParserWithCaseInsensitiveCommands(
		map[byte]*Flag{}, map[string]*Flag{}, []string{},
	)
	if err != nil {
		t.Fatalf("NewParserWithCaseInsensitiveCommands: %v", err)
	}

	parent.AddCmd("child", child)

	if child.parent != parent {
		t.Error("child should reference parent")
	}
	if !child.config.commandCaseIgnore {
		t.Error("commandCaseIgnore should be true")
	}
}

// registryExecuteTests drives table-driven tests for CommandRegistry.ExecuteCommand.
var registryExecuteTests = []struct {
	name    string
	setup   func(CommandRegistry)
	cmd     string
	args    []string
	wantErr string
}{
	{
		name:    "unknown_command",
		setup:   func(CommandRegistry) {},
		cmd:     "missing",
		args:    []string{"a"},
		wantErr: "unknown command: missing",
	},
	{
		name:    "nil_parser",
		setup:   func(cr CommandRegistry) { cr.AddCmd("nil", nil) },
		cmd:     "nil",
		args:    []string{"a"},
		wantErr: "command nil has no parser",
	},
}

func TestRegistryExecuteCommandErrors(t *testing.T) {
	for _, tt := range registryExecuteTests {
		t.Run(tt.name, func(t *testing.T) {
			cr := NewCommandRegistry()
			tt.setup(cr)

			_, err := cr.ExecuteCommand(tt.cmd, tt.args)
			if err == nil {
				t.Fatal("expected error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Errorf("error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestRegistryExecuteCommandSuccess(t *testing.T) {
	cr := NewCommandRegistry()
	sub := newMinimalParser(t)
	sub.nonOpts = []string{"stale"}
	cr.AddCmd("run", sub)

	got, err := cr.ExecuteCommand("run", []string{"x", "y"})
	if err != nil {
		t.Fatalf("ExecuteCommand: %v", err)
	}
	if got != sub {
		t.Error("returned parser should match registered parser")
	}
	if len(got.Args) != 2 || got.Args[0] != "x" || got.Args[1] != "y" {
		t.Errorf("Args = %v, want [x y]", got.Args)
	}
	if len(got.nonOpts) != 0 {
		t.Errorf("nonOpts = %v, want empty", got.nonOpts)
	}
}

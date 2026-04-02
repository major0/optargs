package optargs

import (
	"strings"
	"testing"
	"testing/quick"
)

// newCmdRootParser creates a root parser with --verbose (-v) and --config (-c)
// flags. Reduces boilerplate in command tests that need a parent parser.
func newCmdRootParser(t *testing.T) *Parser {
	t.Helper()
	p, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "config", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("newCmdRootParser: %v", err)
	}
	return p
}

// newCmdServerParser creates a subcommand parser with --port (-p) and --host
// flags. Reduces boilerplate in command tests that need a child parser.
func newCmdServerParser(t *testing.T) *Parser {
	t.Helper()
	p, err := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
		{Name: "host", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("newCmdServerParser: %v", err)
	}
	return p
}

// newCmdClientParser creates a subcommand parser with --url (-u).
// Reduces boilerplate in command tests that need a client parser.
func newCmdClientParser(t *testing.T) *Parser {
	t.Helper()
	p, err := GetOptLong([]string{}, "u:", []Flag{
		{Name: "url", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("newCmdClientParser: %v", err)
	}
	return p
}

// newMinimalParser creates a parser with no options. Useful for tests that
// only need a bare registry.
func newMinimalParser(t *testing.T) *Parser {
	t.Helper()
	p, err := GetOptLong([]string{}, "", []Flag{})
	if err != nil {
		t.Fatalf("newMinimalParser: %v", err)
	}
	return p
}

func TestBasicCommandRegistration(t *testing.T) {
	rootParser := newCmdRootParser(t)
	serverParser := newCmdServerParser(t)

	registeredParser := rootParser.AddCmd("server", serverParser)
	if registeredParser != serverParser {
		t.Fatal("AddCmd should return the registered parser")
	}

	parser, exists := rootParser.GetCommand("server")
	if !exists {
		t.Fatal("server command not found after registration")
	}
	if parser != serverParser {
		t.Error("retrieved parser doesn't match registered parser")
	}
}

func TestCommandExecution(t *testing.T) {
	root := newCmdRootParser(t)
	server := newCmdServerParser(t)
	root.AddCmd("server", server)

	got, err := root.ExecuteCommand("server", []string{"--port", "8080"})
	if err != nil {
		t.Fatalf("ExecuteCommand: %v", err)
	}
	if got != server {
		t.Error("ExecuteCommand should return the subcommand parser")
	}
	if len(got.Args) != 2 || got.Args[0] != "--port" || got.Args[1] != "8080" {
		t.Errorf("Args = %v, want [--port 8080]", got.Args)
	}
}

// executeCommandErrorTests consolidates error cases from command execution,
// including unknown commands and nil-parser commands.
var executeCommandErrorTests = []struct {
	name    string
	setup   func(*Parser)
	cmd     string
	args    []string
	wantErr string
}{
	{
		name:    "unknown_command",
		setup:   func(*Parser) {},
		cmd:     "nonexistent",
		args:    []string{},
		wantErr: "unknown command: nonexistent",
	},
	{
		name:    "nil_parser",
		setup:   func(p *Parser) { p.AddCmd("help", nil) },
		cmd:     "help",
		args:    []string{},
		wantErr: "command help has no parser",
	},
	{
		name:    "registry_unknown_command",
		setup:   func(*Parser) {},
		cmd:     "missing",
		args:    []string{"a"},
		wantErr: "unknown command: missing",
	},
	{
		name:    "registry_nil_parser",
		setup:   func(p *Parser) { p.AddCmd("nil", nil) },
		cmd:     "nil",
		args:    []string{"a"},
		wantErr: "command nil has no parser",
	},
}

func TestExecuteCommandErrors(t *testing.T) {
	for _, tt := range executeCommandErrorTests {
		t.Run(tt.name, func(t *testing.T) {
			root := newMinimalParser(t)
			tt.setup(root)

			_, err := root.ExecuteCommand(tt.cmd, tt.args)
			if err == nil {
				t.Fatal("expected error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Errorf("error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestExecuteCommandSuccess(t *testing.T) {
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

func TestParentOptionInheritance(t *testing.T) {
	rootParser := newCmdRootParser(t)
	serverParser := newCmdServerParser(t)
	rootParser.AddCmd("server", serverParser)

	args, _, option, err := serverParser.findLongOpt("verbose", []string{})
	if err != nil {
		t.Errorf("findLongOpt(verbose): %v", err)
	}
	if option.Name != "verbose" {
		t.Errorf("option.Name = %q, want %q", option.Name, "verbose")
	}
	if len(args) != 0 {
		t.Errorf("args = %v, want empty", args)
	}
}

func TestShortOptionInheritance(t *testing.T) {
	rootParser := newCmdRootParser(t)
	serverParser := newCmdServerParser(t)
	rootParser.AddCmd("server", serverParser)

	args, word, _, option, err := serverParser.findShortOpt('v', "", []string{})
	if err != nil {
		t.Errorf("findShortOpt('v'): %v", err)
	}
	if option.Name != "v" {
		t.Errorf("option.Name = %q, want %q", option.Name, "v")
	}
	if len(args) != 0 {
		t.Errorf("args = %v, want empty", args)
	}
	if word != "" {
		t.Errorf("word = %q, want empty", word)
	}
}

func TestCommandAliases(t *testing.T) {
	root := newCmdRootParser(t)
	server := newCmdServerParser(t)
	root.AddCmd("server", server)
	if err := root.AddAlias("srv", "server"); err != nil {
		t.Fatalf("AddAlias(srv): %v", err)
	}
	if err := root.AddAlias("s", "server"); err != nil {
		t.Fatalf("AddAlias(s): %v", err)
	}

	for _, name := range []string{"server", "srv", "s"} {
		cmd, exists := root.GetCommand(name)
		if !exists || cmd != server {
			t.Fatalf("GetCommand(%q) failed", name)
		}
	}

	aliases := root.GetAliases(server)
	if len(aliases) != 3 {
		t.Errorf("len(aliases) = %d, want 3", len(aliases))
	}
}

func TestAliasForNonExistentCommand(t *testing.T) {
	rootParser := newMinimalParser(t)

	err := rootParser.AddAlias("srv", "server")
	if err == nil {
		t.Fatal("expected error when aliasing non-existent command")
	}
	if got := err.Error(); got != "command server does not exist" {
		t.Errorf("error = %q, want %q", got, "command server does not exist")
	}
}

func TestCommandInspection(t *testing.T) {
	root := newCmdRootParser(t)
	server := newCmdServerParser(t)
	client := newCmdClientParser(t)
	root.AddCmd("server", server)
	root.AddCmd("client", client)
	_ = root.AddAlias("srv", "server")

	commands := root.ListCommands()
	if len(commands) != 3 {
		t.Errorf("len(commands) = %d, want 3", len(commands))
	}
	if commands["server"] != server || commands["srv"] != server || commands["client"] != client {
		t.Error("command mapping incorrect")
	}
}

// --- Case-insensitive command tests ---

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

var caseInsensitiveLookupTests = []struct {
	name       string
	caseIgnore bool
	lookup     string
	wantFound  bool
}{
	{"insensitive_exact", true, "server", true},
	{"insensitive_upper", true, "SERVER", true},
	{"insensitive_miss", true, "nonexistent", false},
	{"sensitive_exact", false, "server", true},
	{"sensitive_upper", false, "SERVER", false},
}

func TestCommandCaseInsensitiveLookup(t *testing.T) {
	for _, tt := range caseInsensitiveLookupTests {
		t.Run(tt.name, func(t *testing.T) {
			root := newMinimalParser(t)
			sub := newMinimalParser(t)
			root.config.commandCaseIgnore = tt.caseIgnore
			root.AddCmd("server", sub)

			got, exists := root.GetCommand(tt.lookup)
			if exists != tt.wantFound {
				t.Fatalf("GetCommand(%q) exists = %v, want %v", tt.lookup, exists, tt.wantFound)
			}
			if tt.wantFound && got != sub {
				t.Error("GetCommand returned wrong parser")
			}
		})
	}
}

// TestSubcommandOverlappingLongOpts verifies cross-chain prefix matching
// when parent and child parsers register long options with overlapping
// prefixes. The longest matching option name wins regardless of which
// level in the chain registered it.
// TestSubcommandOverlappingLongOpts verifies cross-chain prefix matching.
func TestSubcommandOverlappingLongOpts(t *testing.T) {
	tests := []struct {
		name      string
		gpOpts    []Flag
		parOpts   []Flag
		childOpts []Flag
		childArgs []string
		expected  []Option
	}{
		{
			name:      "child_longer_prefix_wins",
			parOpts:   []Flag{{Name: "out", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "output", HasArg: RequiredArgument}},
			childArgs: []string{"--output=file.txt"},
			expected:  []Option{{Name: "output", Arg: "file.txt", HasArg: true}},
		},
		{
			name:      "parent_longer_prefix_wins",
			parOpts:   []Flag{{Name: "output", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "out", HasArg: RequiredArgument}},
			childArgs: []string{"--output=file.txt"},
			expected:  []Option{{Name: "output", Arg: "file.txt", HasArg: true}},
		},
		{
			name:      "three_level_chain_longest_from_grandparent",
			gpOpts:    []Flag{{Name: "output-format", HasArg: RequiredArgument}},
			parOpts:   []Flag{{Name: "output", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "out", HasArg: RequiredArgument}},
			childArgs: []string{"--output-format=json"},
			expected:  []Option{{Name: "output-format", Arg: "json", HasArg: true}},
		},
		{
			name:      "equals_in_option_name_across_chain",
			parOpts:   []Flag{{Name: "key=val", HasArg: NoArgument}},
			childOpts: []Flag{{Name: "key", HasArg: RequiredArgument}},
			childArgs: []string{"--key=val"},
			expected:  []Option{{Name: "key=val", HasArg: false}},
		},
		{
			name:      "child_own_option_preferred_when_same_length",
			parOpts:   []Flag{{Name: "verbose", HasArg: NoArgument}},
			childOpts: []Flag{{Name: "verbose", HasArg: NoArgument}},
			childArgs: []string{"--verbose"},
			expected:  []Option{{Name: "verbose", HasArg: false}},
		},
		{
			name:      "parent_only_option_resolved_from_child",
			parOpts:   []Flag{{Name: "debug", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "color", HasArg: NoArgument}},
			childArgs: []string{"--debug=trace"},
			expected:  []Option{{Name: "debug", Arg: "trace", HasArg: true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var child *Parser
			if tt.gpOpts != nil {
				child = setupChain3(t, tt.gpOpts, tt.parOpts, tt.childOpts, tt.childArgs)
			} else {
				child = setupChain(t, tt.parOpts, tt.childOpts, tt.childArgs)
			}
			assertOptions(t, requireParsedOptions(t, child), tt.expected)
		})
	}
}

// TestNativeSubcommandDispatch exercises the full dispatch flow:
// root Options() encounters a subcommand name, dispatches via ExecuteCommand,
// then the child parser's Options() resolves both local and inherited options.
// TestNativeSubcommandDispatch exercises the full dispatch flow.
func TestNativeSubcommandDispatch(t *testing.T) {
	t.Run("dispatch_and_inherit", func(t *testing.T) {
		root, _ := GetOptLong([]string{"--verbose", "serve", "--port", "8080"}, "v", []Flag{{Name: "verbose", HasArg: NoArgument}})
		child, _ := GetOptLong([]string{}, "p:", []Flag{{Name: "port", HasArg: RequiredArgument}})
		root.AddCmd("serve", child)

		rootOpts := collectNamedOptions(t, root)
		if _, ok := rootOpts["verbose"]; !ok {
			t.Error("Expected root to yield 'verbose'")
		}
		childOpts := collectNamedOptions(t, child)
		if arg, ok := childOpts["port"]; !ok || arg != "8080" {
			t.Errorf("Expected port=8080, got %v", childOpts)
		}
	})

	t.Run("multi_level_dispatch", func(t *testing.T) {
		root, _ := GetOptLong([]string{"-v", "db", "--name", "mydb", "migrate", "--steps", "3"}, "v", []Flag{{Name: "verbose", HasArg: NoArgument}})
		db, _ := GetOptLong([]string{}, "n:", []Flag{{Name: "name", HasArg: RequiredArgument}})
		root.AddCmd("db", db)
		migrate, _ := GetOptLong([]string{}, "s:", []Flag{{Name: "steps", HasArg: RequiredArgument}})
		db.AddCmd("migrate", migrate)

		collectNamedOptions(t, root)
		dbOpts := collectNamedOptions(t, db)
		if arg, ok := dbOpts["name"]; !ok || arg != "mydb" {
			t.Errorf("Expected db [name=mydb], got %v", dbOpts)
		}
		migrateOpts := collectNamedOptions(t, migrate)
		if arg, ok := migrateOpts["steps"]; !ok || arg != "3" {
			t.Errorf("Expected migrate [steps=3], got %v", migrateOpts)
		}
	})
}

// TestDispatchErrorModes verifies that error modes work correctly through
// the dispatch + inheritance chain.
// TestDispatchErrorModes verifies error modes through dispatch + inheritance.
func TestDispatchErrorModes(t *testing.T) {
	t.Run("silent_child_inherits_parent_option", func(t *testing.T) {
		root, _ := GetOptLong([]string{"sub", "-v", "--file", "test.txt"}, "v", []Flag{{Name: "verbose", HasArg: NoArgument}, {Name: "file", HasArg: RequiredArgument}})
		child, _ := GetOpt([]string{}, ":")
		root.AddCmd("sub", child)
		collectNamedOptions(t, root)
		found := collectNamedOptions(t, child)
		if _, ok := found["v"]; !ok {
			t.Error("Expected child to find parent's -v")
		}
		if arg, ok := found["file"]; !ok || arg != "test.txt" {
			t.Errorf("Expected --file=test.txt, got %v", found)
		}
	})

	t.Run("silent_child_unknown_option_no_log", func(t *testing.T) {
		root, _ := GetOptLong([]string{"sub", "-x"}, "v", []Flag{{Name: "verbose", HasArg: NoArgument}})
		child, _ := GetOpt([]string{}, ":")
		root.AddCmd("sub", child)
		collectNamedOptions(t, root)
		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}
		if foundErr == nil {
			t.Fatal("expected error for unknown option -x")
		}
		if !strings.Contains(foundErr.Error(), "unknown option: x") {
			t.Errorf("Expected 'unknown option: x', got '%s'", foundErr.Error())
		}
	})

	t.Run("verbose_child_missing_parent_arg", func(t *testing.T) {
		root, _ := GetOptLong([]string{"sub", "-f"}, "f:", []Flag{})
		child, _ := GetOpt([]string{}, "")
		root.AddCmd("sub", child)
		collectNamedOptions(t, root)
		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}
		if foundErr == nil {
			t.Fatal("expected error for missing argument")
		}
		if !strings.Contains(foundErr.Error(), "option requires an argument: f") {
			t.Errorf("Expected 'option requires an argument: f', got '%s'", foundErr.Error())
		}
	})
}

// Feature: goarg-optargs-integration, Property 10: Active subcommand detection correctness.
func TestPropertyActiveCommandCorrectness(t *testing.T) {
	// Sub-property A: After dispatching a subcommand, ActiveCommand returns
	// the correct name and parser pointer.
	t.Run("dispatched_returns_correct", func(t *testing.T) {
		f := func(cmdName string) bool {
			// Filter to valid command names (non-empty, no dashes, no spaces).
			if len(cmdName) == 0 || len(cmdName) > 20 {
				return true
			}
			for _, r := range cmdName {
				if r == '-' || r == ' ' || r < 'a' || r > 'z' {
					return true
				}
			}

			root, err := GetOptLong([]string{cmdName}, "", []Flag{})
			if err != nil {
				return false
			}
			child, err := GetOptLong([]string{}, "", []Flag{})
			if err != nil {
				return false
			}
			root.AddCmd(cmdName, child)

			// Drain the iterator to trigger dispatch.
			for range root.Options() {
			}

			name, parser := root.ActiveCommand()
			return name == cmdName && parser == child
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
			t.Error(err)
		}
	})

	// Sub-property B: When no subcommand is dispatched, ActiveCommand
	// returns empty name and nil parser.
	t.Run("no_dispatch_returns_empty", func(t *testing.T) {
		f := func(optArg string) bool {
			if len(optArg) == 0 || len(optArg) > 20 {
				return true
			}
			// Use only plain non-option args that won't match a command.
			for _, r := range optArg {
				if r < 'a' || r > 'z' {
					return true
				}
			}

			root, err := GetOptLong([]string{optArg}, "", []Flag{})
			if err != nil {
				return false
			}
			// No commands registered — nothing to dispatch.
			for range root.Options() {
			}

			name, parser := root.ActiveCommand()
			return name == "" && parser == nil
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
			t.Error(err)
		}
	})

	// Sub-property C: Recursive walk of ActiveCommand produces the full
	// dispatch chain for nested subcommands.
	t.Run("nested_chain_walk", func(t *testing.T) {
		f := func(depth uint8) bool {
			// Limit depth to 1–5 levels.
			n := int(depth%5) + 1

			// Build chain: root → cmd0 → cmd1 → ... → cmdN-1
			names := make([]string, n)
			parsers := make([]*Parser, n+1)
			var err error
			parsers[0], err = GetOptLong([]string{}, "", []Flag{})
			if err != nil {
				return false
			}
			args := make([]string, n)
			for i := range n {
				names[i] = string(rune('a' + i))
				args[i] = names[i]
				parsers[i+1], err = GetOptLong([]string{}, "", []Flag{})
				if err != nil {
					return false
				}
				parsers[i].AddCmd(names[i], parsers[i+1])
			}

			// Set args on root and drain.
			parsers[0].Args = args
			for range parsers[0].Options() {
			}
			// Drain each child's Options() to trigger nested dispatch.
			for i := 1; i < n; i++ {
				for range parsers[i].Options() {
				}
			}

			// Walk the chain via ActiveCommand.
			current := parsers[0]
			for i := range n {
				name, p := current.ActiveCommand()
				if name != names[i] || p != parsers[i+1] {
					return false
				}
				current = p
			}
			// Leaf should have no active command.
			name, p := current.ActiveCommand()
			return name == "" && p == nil
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
			t.Error(err)
		}
	})
}

// activeCommandTests covers specific ActiveCommand scenarios not explored
// by the property test: inherited options, aliases, and edge cases.
var activeCommandTests = []struct {
	name     string
	args     []string
	setup    func(t *testing.T) (*Parser, map[string]*Parser)
	wantName string
	wantNil  bool
}{
	{
		name: "single_subcommand",
		args: []string{"server", "--port", "8080"},
		setup: func(t *testing.T) (*Parser, map[string]*Parser) {
			t.Helper()
			root := newCmdRootParser(t)
			server := newCmdServerParser(t)
			root.AddCmd("server", server)
			return root, map[string]*Parser{"server": server}
		},
		wantName: "server",
	},
	{
		name: "no_subcommand",
		args: []string{"--verbose"},
		setup: func(t *testing.T) (*Parser, map[string]*Parser) {
			t.Helper()
			root := newCmdRootParser(t)
			server := newCmdServerParser(t)
			root.AddCmd("server", server)
			return root, nil
		},
		wantName: "",
		wantNil:  true,
	},
	{
		name: "subcommand_with_inherited_options",
		args: []string{"--verbose", "server", "--port", "9090"},
		setup: func(t *testing.T) (*Parser, map[string]*Parser) {
			t.Helper()
			root := newCmdRootParser(t)
			server := newCmdServerParser(t)
			root.AddCmd("server", server)
			return root, map[string]*Parser{"server": server}
		},
		wantName: "server",
	},
	{
		name: "no_commands_registered",
		args: []string{"anything"},
		setup: func(t *testing.T) (*Parser, map[string]*Parser) {
			t.Helper()
			root := newMinimalParser(t)
			return root, nil
		},
		wantName: "",
		wantNil:  true,
	},
}

func TestActiveCommand(t *testing.T) {
	for _, tt := range activeCommandTests {
		t.Run(tt.name, func(t *testing.T) {
			root, parsers := tt.setup(t)
			root.Args = tt.args

			// Drain root options to trigger dispatch.
			for _, err := range root.Options() {
				if err != nil {
					t.Fatalf("Options() error: %v", err)
				}
			}

			name, parser := root.ActiveCommand()
			if name != tt.wantName {
				t.Errorf("ActiveCommand() name = %q, want %q", name, tt.wantName)
			}
			if tt.wantNil {
				if parser != nil {
					t.Error("ActiveCommand() parser should be nil")
				}
			} else {
				want := parsers[tt.wantName]
				if parser != want {
					t.Error("ActiveCommand() parser does not match expected")
				}
			}
		})
	}
}

func TestActiveCommandNestedChain(t *testing.T) {
	root, _ := GetOptLong(
		[]string{"-v", "db", "--name", "mydb", "migrate", "--steps", "3"},
		"v", []Flag{{Name: "verbose", HasArg: NoArgument}},
	)
	db, _ := GetOptLong([]string{}, "n:", []Flag{{Name: "name", HasArg: RequiredArgument}})
	root.AddCmd("db", db)
	migrate, _ := GetOptLong([]string{}, "s:", []Flag{{Name: "steps", HasArg: RequiredArgument}})
	db.AddCmd("migrate", migrate)

	// Drain each level.
	for _, err := range root.Options() {
		if err != nil {
			t.Fatalf("root Options(): %v", err)
		}
	}
	for _, err := range db.Options() {
		if err != nil {
			t.Fatalf("db Options(): %v", err)
		}
	}
	for _, err := range migrate.Options() {
		if err != nil {
			t.Fatalf("migrate Options(): %v", err)
		}
	}

	// Walk the chain.
	name1, p1 := root.ActiveCommand()
	if name1 != "db" || p1 != db {
		t.Fatalf("root.ActiveCommand() = (%q, %v), want (\"db\", db)", name1, p1)
	}
	name2, p2 := p1.ActiveCommand()
	if name2 != "migrate" || p2 != migrate {
		t.Fatalf("db.ActiveCommand() = (%q, %v), want (\"migrate\", migrate)", name2, p2)
	}
	name3, p3 := p2.ActiveCommand()
	if name3 != "" || p3 != nil {
		t.Errorf("migrate.ActiveCommand() = (%q, %v), want (\"\", nil)", name3, p3)
	}
}

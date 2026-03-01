package optargs

import (
	"testing"
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
	rootParser := newCmdRootParser(t)
	serverParser := newCmdServerParser(t)
	rootParser.AddCmd("server", serverParser)

	executedParser, err := rootParser.ExecuteCommand("server", []string{"--port", "8080"})
	if err != nil {
		t.Errorf("ExecuteCommand: %v", err)
	}
	if executedParser != serverParser {
		t.Error("ExecuteCommand should return the subcommand parser")
	}

	expectedArgs := []string{"--port", "8080"}
	if len(executedParser.Args) != len(expectedArgs) {
		t.Errorf("len(Args) = %d, want %d", len(executedParser.Args), len(expectedArgs))
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

func TestCommandNotFound(t *testing.T) {
	rootParser := newMinimalParser(t)

	_, err := rootParser.ExecuteCommand("nonexistent", []string{})
	if err == nil {
		t.Fatal("expected error for non-existent command")
	}
	if got := err.Error(); got != "unknown command: nonexistent" {
		t.Errorf("error = %q, want %q", got, "unknown command: nonexistent")
	}
}

func TestMultipleCommands(t *testing.T) {
	rootParser := newCmdRootParser(t)
	serverParser := newCmdServerParser(t)

	clientParser, err := GetOptLong([]string{}, "u:", []Flag{
		{Name: "url", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("GetOptLong(client): %v", err)
	}

	rootParser.AddCmd("server", serverParser)
	rootParser.AddCmd("client", clientParser)

	commands := rootParser.ListCommands()
	if len(commands) != 2 {
		t.Errorf("len(commands) = %d, want 2", len(commands))
	}
	if parser, exists := commands["server"]; !exists || parser != serverParser {
		t.Error("server command not found or incorrect parser")
	}
	if parser, exists := commands["client"]; !exists || parser != clientParser {
		t.Error("client command not found or incorrect parser")
	}
}

func TestCommandWithoutParser(t *testing.T) {
	rootParser := newMinimalParser(t)
	rootParser.AddCmd("help", nil)

	_, err := rootParser.ExecuteCommand("help", []string{})
	if err == nil {
		t.Fatal("expected error for command without parser")
	}
	if got := err.Error(); got != "command help has no parser" {
		t.Errorf("error = %q, want %q", got, "command help has no parser")
	}
}

func TestCommandAliases(t *testing.T) {
	rootParser := newCmdRootParser(t)
	serverParser := newCmdServerParser(t)
	rootParser.AddCmd("server", serverParser)

	if err := rootParser.AddAlias("srv", "server"); err != nil {
		t.Fatalf("AddAlias(srv): %v", err)
	}
	if err := rootParser.AddAlias("s", "server"); err != nil {
		t.Fatalf("AddAlias(s): %v", err)
	}

	// All names should resolve to the same parser.
	for _, name := range []string{"server", "srv", "s"} {
		cmd, exists := rootParser.GetCommand(name)
		if !exists {
			t.Fatalf("GetCommand(%q) not found", name)
		}
		if cmd != serverParser {
			t.Errorf("GetCommand(%q) returned wrong parser", name)
		}
	}

	aliases := rootParser.GetAliases(serverParser)
	if len(aliases) != 3 {
		t.Errorf("len(aliases) = %d, want 3", len(aliases))
	}

	aliasSet := make(map[string]bool, len(aliases))
	for _, a := range aliases {
		aliasSet[a] = true
	}
	for _, want := range []string{"server", "srv", "s"} {
		if !aliasSet[want] {
			t.Errorf("alias %q not found in %v", want, aliases)
		}
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

// commandMappingTests drives TestCommandInspection subtests.
var commandMappingTests = []struct {
	name     string
	register func(*Parser, *Parser, *Parser)
	want     map[string]string // command name → parser label
}{
	{
		name: "commands_and_aliases",
		register: func(root, server, client *Parser) {
			root.AddCmd("server", server)
			root.AddCmd("client", client)
			_ = root.AddAlias("srv", "server")
			_ = root.AddAlias("c", "client")
		},
		want: map[string]string{
			"server": "server",
			"srv":    "server",
			"client": "client",
			"c":      "client",
		},
	},
}

func TestCommandInspection(t *testing.T) {
	for _, tt := range commandMappingTests {
		t.Run(tt.name, func(t *testing.T) {
			rootParser := newCmdRootParser(t)
			serverParser := newCmdServerParser(t)
			clientParser, err := GetOptLong([]string{}, "u:", []Flag{
				{Name: "url", HasArg: RequiredArgument},
			})
			if err != nil {
				t.Fatalf("GetOptLong(client): %v", err)
			}

			parsersByLabel := map[string]*Parser{
				"server": serverParser,
				"client": clientParser,
			}

			tt.register(rootParser, serverParser, clientParser)

			commands := rootParser.ListCommands()
			if len(commands) != len(tt.want) {
				t.Errorf("len(commands) = %d, want %d", len(commands), len(tt.want))
			}
			for name, label := range tt.want {
				if parser, exists := commands[name]; !exists || parser != parsersByLabel[label] {
					t.Errorf("command %q mapping incorrect", name)
				}
			}
		})
	}
}

// TestRealWorldCommandHierarchy demonstrates a complete real-world usage
// of the command system with nested commands, aliases, and option inheritance.
func TestRealWorldCommandHierarchy(t *testing.T) {
	rootParser, err := GetOptLong([]string{}, "vhc:", []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "help", HasArg: NoArgument},
		{Name: "config", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("GetOptLong(root): %v", err)
	}

	serverParser, err := GetOptLong([]string{}, "p:H:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
		{Name: "host", HasArg: RequiredArgument},
		{Name: "daemon", HasArg: NoArgument},
	})
	if err != nil {
		t.Fatalf("GetOptLong(server): %v", err)
	}

	clientParser, err := GetOptLong([]string{}, "u:t:", []Flag{
		{Name: "url", HasArg: RequiredArgument},
		{Name: "timeout", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("GetOptLong(client): %v", err)
	}

	dbParser, err := GetOptLong([]string{}, "d:", []Flag{
		{Name: "database", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("GetOptLong(db): %v", err)
	}

	migrateParser, err := GetOptLong([]string{}, "", []Flag{
		{Name: "dry-run", HasArg: NoArgument},
		{Name: "steps", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("GetOptLong(migrate): %v", err)
	}

	rootParser.AddCmd("server", serverParser)
	rootParser.AddCmd("client", clientParser)
	rootParser.AddCmd("database", dbParser)
	dbParser.AddCmd("migrate", migrateParser)

	_ = rootParser.AddAlias("srv", "server")
	_ = rootParser.AddAlias("s", "server")
	_ = rootParser.AddAlias("c", "client")
	_ = rootParser.AddAlias("db", "database")
	_ = dbParser.AddAlias("mig", "migrate")
	_ = dbParser.AddAlias("m", "migrate")

	t.Run("alias_execution", func(t *testing.T) {
		parser, err := rootParser.ExecuteCommand("srv", []string{"--port", "9000"})
		if err != nil {
			t.Errorf("ExecuteCommand(srv): %v", err)
		}
		if parser != serverParser {
			t.Error("alias should resolve to same parser")
		}
	})

	t.Run("nested_alias_execution", func(t *testing.T) {
		parser, err := dbParser.ExecuteCommand("mig", []string{"--dry-run"})
		if err != nil {
			t.Errorf("ExecuteCommand(mig): %v", err)
		}
		if parser != migrateParser {
			t.Error("nested alias should resolve to migrate parser")
		}
	})

	t.Run("multi_level_long_opt_inheritance", func(t *testing.T) {
		_, _, option, err := migrateParser.findLongOpt("config", []string{"test.conf"})
		if err != nil {
			t.Errorf("findLongOpt(config): %v", err)
		}
		if option.Name != "config" {
			t.Errorf("option.Name = %q, want %q", option.Name, "config")
		}
		if option.Arg != "test.conf" {
			t.Errorf("option.Arg = %q, want %q", option.Arg, "test.conf")
		}
	})

	t.Run("full_command_tree_inspection", func(t *testing.T) {
		commands := rootParser.ListCommands()
		// server, srv, s, client, c, database, db = 7
		if len(commands) != 7 {
			t.Errorf("len(commands) = %d, want 7", len(commands))
		}

		wantMapping := []struct {
			name     string
			expected *Parser
		}{
			{"server", serverParser},
			{"srv", serverParser},
			{"s", serverParser},
			{"client", clientParser},
			{"c", clientParser},
			{"database", dbParser},
			{"db", dbParser},
		}

		for _, tc := range wantMapping {
			if parser, exists := commands[tc.name]; !exists || parser != tc.expected {
				t.Errorf("command %q mapping incorrect", tc.name)
			}
		}
	})
}

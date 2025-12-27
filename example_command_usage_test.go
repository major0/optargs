package optargs

import (
	"testing"
)

// TestRealWorldCommandExample demonstrates a complete real-world usage
// of the command system with option inheritance and aliases
func TestRealWorldCommandExample(t *testing.T) {
	// Create root parser with global options
	rootParser, err := GetOptLong([]string{}, "vhc:", []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "help", HasArg: NoArgument},
		{Name: "config", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Create server subcommand parser
	serverParser, err := GetOptLong([]string{}, "p:H:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
		{Name: "host", HasArg: RequiredArgument},
		{Name: "daemon", HasArg: NoArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create server parser: %v", err)
	}

	// Create client subcommand parser
	clientParser, err := GetOptLong([]string{}, "u:t:", []Flag{
		{Name: "url", HasArg: RequiredArgument},
		{Name: "timeout", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create client parser: %v", err)
	}

	// Create database subcommand parser
	dbParser, err := GetOptLong([]string{}, "d:", []Flag{
		{Name: "database", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create db parser: %v", err)
	}

	// Create migrate sub-subcommand parser
	migrateParser, err := GetOptLong([]string{}, "", []Flag{
		{Name: "dry-run", HasArg: NoArgument},
		{Name: "steps", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create migrate parser: %v", err)
	}

	// Register commands
	rootParser.AddCmd("server", serverParser)
	rootParser.AddCmd("client", clientParser)
	rootParser.AddCmd("database", dbParser)
	
	// Register nested command
	dbParser.AddCmd("migrate", migrateParser)

	// Add aliases
	rootParser.AddAlias("srv", "server")
	rootParser.AddAlias("s", "server")
	rootParser.AddAlias("c", "client")
	rootParser.AddAlias("db", "database")
	
	dbParser.AddAlias("mig", "migrate")
	dbParser.AddAlias("m", "migrate")

	// Test 1: Basic command execution
	t.Run("basic_server_command", func(t *testing.T) {
		parser, err := rootParser.ExecuteCommand("server", []string{"--port", "8080", "--host", "0.0.0.0"})
		if err != nil {
			t.Errorf("Server command failed: %v", err)
		}
		if parser != serverParser {
			t.Error("Wrong parser returned")
		}
	})

	// Test 2: Alias usage
	t.Run("server_alias", func(t *testing.T) {
		parser, err := rootParser.ExecuteCommand("srv", []string{"--port", "9000"})
		if err != nil {
			t.Errorf("Server alias command failed: %v", err)
		}
		if parser != serverParser {
			t.Error("Alias should point to same parser")
		}
	})

	// Test 3: Nested command with alias
	t.Run("nested_migrate_command", func(t *testing.T) {
		parser, err := dbParser.ExecuteCommand("mig", []string{"--dry-run"})
		if err != nil {
			t.Errorf("Migrate alias command failed: %v", err)
		}
		if parser != migrateParser {
			t.Error("Nested alias should point to migrate parser")
		}
	})

	// Test 4: Option inheritance - server can use root's verbose
	t.Run("option_inheritance", func(t *testing.T) {
		_, option, err := serverParser.findLongOptWithFallback("verbose", []string{})
		if err != nil {
			t.Errorf("Server couldn't inherit verbose option: %v", err)
		}
		if option.Name != "verbose" {
			t.Errorf("Expected 'verbose', got '%s'", option.Name)
		}
	})

	// Test 5: Multi-level inheritance - migrate can use root's config
	t.Run("multi_level_inheritance", func(t *testing.T) {
		_, option, err := migrateParser.findLongOptWithFallback("config", []string{"test.conf"})
		if err != nil {
			t.Errorf("Migrate couldn't inherit config option: %v", err)
		}
		if option.Name != "config" {
			t.Errorf("Expected 'config', got '%s'", option.Name)
		}
		if option.Arg != "test.conf" {
			t.Errorf("Expected 'test.conf', got '%s'", option.Arg)
		}
	})

	// Test 6: Command inspection
	t.Run("command_inspection", func(t *testing.T) {
		commands := rootParser.ListCommands()
		
		// Should have: server, srv, s, client, c, database, db
		expectedCount := 7
		if len(commands) != expectedCount {
			t.Errorf("Expected %d commands, got %d", expectedCount, len(commands))
		}

		// Test that aliases point to correct parsers
		testCases := []struct {
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

		for _, tc := range testCases {
			if parser, exists := commands[tc.name]; !exists || parser != tc.expected {
				t.Errorf("Command '%s' mapping incorrect", tc.name)
			}
		}
	})

	// Test 7: Get aliases for a parser
	t.Run("get_aliases", func(t *testing.T) {
		aliases := rootParser.GetAliases(serverParser)
		expectedAliases := []string{"server", "srv", "s"}
		
		if len(aliases) != len(expectedAliases) {
			t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(aliases))
		}

		aliasMap := make(map[string]bool)
		for _, alias := range aliases {
			aliasMap[alias] = true
		}

		for _, expected := range expectedAliases {
			if !aliasMap[expected] {
				t.Errorf("Expected alias '%s' not found", expected)
			}
		}
	})
}

// TestCommandSystemArchitecture verifies the key architectural principles
func TestCommandSystemArchitecture(t *testing.T) {
	// Create a simple command hierarchy
	root, _ := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
	})
	
	sub, _ := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
	})
	
	root.AddCmd("sub", sub)

	// Test 1: Commands are stored as simple key-value pairs
	t.Run("simple_key_value_storage", func(t *testing.T) {
		commands := root.ListCommands()
		
		// Should be a simple map[string]*Parser
		if parser, exists := commands["sub"]; !exists || parser != sub {
			t.Error("Commands should be stored as simple key-value pairs")
		}
	})

	// Test 2: Multiple keys can point to same parser (aliases)
	t.Run("multiple_keys_same_parser", func(t *testing.T) {
		root.AddAlias("s", "sub")
		
		commands := root.ListCommands()
		
		if commands["sub"] != commands["s"] {
			t.Error("Aliases should point to the same parser instance")
		}
	})

	// Test 3: Parent relationships are established
	t.Run("parent_relationships", func(t *testing.T) {
		if sub.parent != root {
			t.Error("Parent relationship not established correctly")
		}
	})

	// Test 4: Command registry is inspectable
	t.Run("inspectable_registry", func(t *testing.T) {
		// We can iterate over all commands
		commands := root.ListCommands()
		
		for name, parser := range commands {
			t.Logf("Command '%s' -> Parser %p", name, parser)
		}
		
		// We can find all aliases for a parser
		aliases := root.GetAliases(sub)
		if len(aliases) != 2 { // "sub" and "s"
			t.Errorf("Expected 2 aliases, got %d", len(aliases))
		}
	})
}
package optargs

import (
	"testing"
)

func TestBasicCommandRegistration(t *testing.T) {
	// Create root parser with global options
	rootParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "config", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Create subcommand parser
	serverParser, err := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
		{Name: "host", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create server parser: %v", err)
	}

	// Register subcommand
	registeredParser := rootParser.AddCmd("server", serverParser)

	if registeredParser != serverParser {
		t.Fatal("AddCmd should return the registered parser")
	}

	// Verify command registration
	parser, exists := rootParser.GetCommand("server")
	if !exists {
		t.Fatal("Server command not found after registration")
	}

	if parser != serverParser {
		t.Error("Retrieved parser doesn't match registered parser")
	}
}

func TestCommandExecution(t *testing.T) {
	// Create root parser
	rootParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Create subcommand parser
	serverParser, err := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create server parser: %v", err)
	}

	// Register subcommand
	rootParser.AddCmd("server", serverParser)

	// Execute command
	executedParser, err := rootParser.ExecuteCommand("server", []string{"--port", "8080"})
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	if executedParser != serverParser {
		t.Error("ExecuteCommand should return the subcommand parser")
	}

	// Verify the parser has the correct args
	expectedArgs := []string{"--port", "8080"}
	if len(executedParser.Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(executedParser.Args))
	}
}

func TestParentOptionInheritance(t *testing.T) {
	// Create root parser with global verbose option
	rootParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Create subcommand parser without verbose option
	serverParser, err := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create server parser: %v", err)
	}

	// Register subcommand
	rootParser.AddCmd("server", serverParser)

	// Test that subcommand can find parent's verbose option
	args, option, err := serverParser.findLongOptWithFallback("verbose", []string{})
	if err != nil {
		t.Errorf("Failed to find verbose option in parent: %v", err)
	}

	if option.Name != "verbose" {
		t.Errorf("Expected option name 'verbose', got '%s'", option.Name)
	}

	if len(args) != 0 {
		t.Errorf("Expected empty args, got %v", args)
	}
}

func TestShortOptionInheritance(t *testing.T) {
	// Create root parser with global short option
	rootParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Create subcommand parser without verbose option
	serverParser, err := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create server parser: %v", err)
	}

	// Register subcommand
	rootParser.AddCmd("server", serverParser)

	// Test that subcommand can find parent's short option
	args, word, option, err := serverParser.findShortOptWithFallback('v', "", []string{})
	if err != nil {
		t.Errorf("Failed to find 'v' option in parent: %v", err)
	}

	if option.Name != "v" {
		t.Errorf("Expected option name 'v', got '%s'", option.Name)
	}

	if len(args) != 0 {
		t.Errorf("Expected empty args, got %v", args)
	}

	if word != "" {
		t.Errorf("Expected empty word, got '%s'", word)
	}
}

func TestCommandNotFound(t *testing.T) {
	// Create root parser
	rootParser, err := GetOptLong([]string{}, "", []Flag{})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Try to execute non-existent command
	_, err = rootParser.ExecuteCommand("nonexistent", []string{})
	if err == nil {
		t.Error("Expected error for non-existent command")
	}

	expectedMsg := "unknown command: nonexistent"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestMultipleCommands(t *testing.T) {
	// Create root parser
	rootParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Create multiple subcommand parsers
	serverParser, _ := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
	})
	
	clientParser, _ := GetOptLong([]string{}, "u:", []Flag{
		{Name: "url", HasArg: RequiredArgument},
	})

	// Register multiple commands
	rootParser.AddCmd("server", serverParser)
	rootParser.AddCmd("client", clientParser)

	// Verify both commands exist
	commands := rootParser.ListCommands()
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}

	if parser, exists := commands["server"]; !exists || parser != serverParser {
		t.Error("Server command not found or incorrect parser")
	}

	if parser, exists := commands["client"]; !exists || parser != clientParser {
		t.Error("Client command not found or incorrect parser")
	}
}

func TestCommandWithoutParser(t *testing.T) {
	// Create root parser
	rootParser, err := GetOptLong([]string{}, "", []Flag{})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Register command without parser
	rootParser.AddCmd("help", nil)

	// Try to execute command without parser
	_, err = rootParser.ExecuteCommand("help", []string{})
	if err == nil {
		t.Error("Expected error for command without parser")
	}

	expectedMsg := "command help has no parser"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}
func TestCommandAliases(t *testing.T) {
	// Create root parser
	rootParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Create subcommand parser
	serverParser, err := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create server parser: %v", err)
	}

	// Register main command
	rootParser.AddCmd("server", serverParser)

	// Add aliases
	err = rootParser.AddAlias("srv", "server")
	if err != nil {
		t.Errorf("Failed to add alias: %v", err)
	}

	err = rootParser.AddAlias("s", "server")
	if err != nil {
		t.Errorf("Failed to add second alias: %v", err)
	}

	// Test that all aliases point to the same parser
	serverCmd, exists := rootParser.GetCommand("server")
	if !exists {
		t.Fatal("Original server command not found")
	}

	srvCmd, exists := rootParser.GetCommand("srv")
	if !exists {
		t.Fatal("srv alias not found")
	}

	sCmd, exists := rootParser.GetCommand("s")
	if !exists {
		t.Fatal("s alias not found")
	}

	// All should point to the same parser
	if serverCmd != serverParser || srvCmd != serverParser || sCmd != serverParser {
		t.Error("Aliases don't point to the same parser")
	}

	// Test GetAliases function
	aliases := rootParser.GetAliases(serverParser)
	expectedAliases := []string{"server", "srv", "s"}
	
	if len(aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(aliases))
	}

	// Check that all expected aliases are present
	aliasMap := make(map[string]bool)
	for _, alias := range aliases {
		aliasMap[alias] = true
	}

	for _, expected := range expectedAliases {
		if !aliasMap[expected] {
			t.Errorf("Expected alias '%s' not found", expected)
		}
	}
}

func TestAliasForNonExistentCommand(t *testing.T) {
	// Create root parser
	rootParser, err := GetOptLong([]string{}, "", []Flag{})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Try to create alias for non-existent command
	err = rootParser.AddAlias("srv", "server")
	if err == nil {
		t.Error("Expected error when creating alias for non-existent command")
	}

	expectedMsg := "command server does not exist"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestCommandInspection(t *testing.T) {
	// Create root parser
	rootParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Create multiple parsers
	serverParser, _ := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
	})
	
	clientParser, _ := GetOptLong([]string{}, "u:", []Flag{
		{Name: "url", HasArg: RequiredArgument},
	})

	// Register commands and aliases
	rootParser.AddCmd("server", serverParser)
	rootParser.AddCmd("client", clientParser)
	rootParser.AddAlias("srv", "server")
	rootParser.AddAlias("c", "client")

	// Test command inspection
	commands := rootParser.ListCommands()
	
	// Should have 4 entries: server, srv, client, c
	if len(commands) != 4 {
		t.Errorf("Expected 4 command entries, got %d", len(commands))
	}

	// Verify mappings
	testCases := []struct {
		name     string
		expected *Parser
	}{
		{"server", serverParser},
		{"srv", serverParser},
		{"client", clientParser},
		{"c", clientParser},
	}

	for _, tc := range testCases {
		if parser, exists := commands[tc.name]; !exists || parser != tc.expected {
			t.Errorf("Command '%s' mapping incorrect", tc.name)
		}
	}
}
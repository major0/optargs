package optargs

import (
	"testing"
)

func TestFullCommandIntegration(t *testing.T) {
	// This test demonstrates the full workflow: mycmd mysubcmd --verbose
	// where --verbose is defined in the root parser but used by the subcommand
	
	// Create root parser with global --verbose option
	rootParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "config", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Create server subcommand parser with --port option
	serverParser, err := GetOptLong([]string{}, "p:", []Flag{
		{Name: "port", HasArg: RequiredArgument},
		{Name: "host", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create server parser: %v", err)
	}

	// Register server subcommand
	rootParser.AddCmd("server", serverParser)

	// Test command registration
	retrievedParser, exists := rootParser.GetCommand("server")
	if !exists {
		t.Fatal("Server command not found after registration")
	}

	if retrievedParser != serverParser {
		t.Error("Retrieved parser doesn't match registered parser")
	}

	// Test parent relationship
	if serverParser.parent != rootParser {
		t.Error("Server parser parent not set correctly")
	}

	// Test option inheritance - server parser should be able to find parent's verbose option
	_, option, err := serverParser.findLongOptWithFallback("verbose", []string{})
	if err != nil {
		t.Errorf("Server parser couldn't find parent's verbose option: %v", err)
	}

	if option.Name != "verbose" {
		t.Errorf("Expected option name 'verbose', got '%s'", option.Name)
	}

	// Test command execution
	executedParser, err := rootParser.ExecuteCommand("server", []string{"--port", "8080"})
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	if executedParser != serverParser {
		t.Error("ExecuteCommand should return the server parser")
	}

	// Verify the executed parser has the correct args
	expectedArgs := []string{"--port", "8080"}
	if len(executedParser.Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(executedParser.Args))
	}
}

func TestNestedCommandInheritance(t *testing.T) {
	// Test multiple levels of command nesting with option inheritance
	
	// Root parser: mycmd --verbose
	rootParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "config", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Database subcommand parser: mycmd database --verbose
	dbParser, err := GetOptLong([]string{}, "v", []Flag{
		{Name: "verbose", HasArg: NoArgument}, // Can override parent
		{Name: "database", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create db parser: %v", err)
	}

	// Migrate sub-subcommand parser: mycmd database migrate --dry-run
	migrateParser, err := GetOptLong([]string{}, "", []Flag{
		{Name: "dry-run", HasArg: NoArgument},
		{Name: "steps", HasArg: RequiredArgument},
	})
	if err != nil {
		t.Fatalf("Failed to create migrate parser: %v", err)
	}

	// Register nested commands
	rootParser.AddCmd("database", dbParser)
	dbParser.AddCmd("migrate", migrateParser)

	// Test that nested parsers have correct parent relationships
	if dbParser.parent != rootParser {
		t.Error("Database parser parent not set correctly")
	}

	if migrateParser.parent != dbParser {
		t.Error("Migrate parser parent not set correctly")
	}

	// Test multi-level inheritance: migrate should be able to find root's verbose option
	_, option, err := migrateParser.findLongOptWithFallback("verbose", []string{})
	if err != nil {
		t.Errorf("Migrate parser couldn't find root's verbose option: %v", err)
	}

	if option.Name != "verbose" {
		t.Errorf("Expected option name 'verbose', got '%s'", option.Name)
	}
}

func TestCommandWithGlobalOptions(t *testing.T) {
	// Test that global options work with subcommands
	
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

	// Test that server parser can access parent's verbose option
	_, option, err := serverParser.findLongOptWithFallback("verbose", []string{})
	if err != nil {
		t.Errorf("Server parser couldn't find parent's verbose option: %v", err)
	}

	if option.Name != "verbose" {
		t.Errorf("Expected option name 'verbose', got '%s'", option.Name)
	}

	// Test command execution with arguments
	executedParser, err := rootParser.ExecuteCommand("server", []string{"--port", "8080"})
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	if executedParser != serverParser {
		t.Error("ExecuteCommand should return the server parser")
	}
}
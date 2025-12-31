package optargs

import (
	"testing"
)

// TestCoverageRegressionFix adds tests for uncovered functions to achieve 100% coverage
func TestCoverageRegressionFix(t *testing.T) {
	// Test NewParserWithCaseInsensitiveCommands function (0% coverage)
	t.Run("NewParserWithCaseInsensitiveCommands", func(t *testing.T) {
		shortOpts := map[byte]*Flag{
			'v': {HasArg: NoArgument},
		}
		longOpts := map[string]*Flag{
			"verbose": {HasArg: NoArgument},
		}
		args := []string{"test", "arg"}

		parser, err := NewParserWithCaseInsensitiveCommands(shortOpts, longOpts, args, nil)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// Verify that case insensitive command matching is enabled
		if !parser.config.commandCaseIgnore {
			t.Error("Expected commandCaseIgnore to be true")
		}

		// Verify parser was created with correct options and args
		if len(parser.shortOpts) != 1 {
			t.Errorf("Expected 1 short option, got %d", len(parser.shortOpts))
		}
		if len(parser.longOpts) != 1 {
			t.Errorf("Expected 1 long option, got %d", len(parser.longOpts))
		}
		if len(parser.Args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(parser.Args))
		}
		if parser.Args[0] != "test" || parser.Args[1] != "arg" {
			t.Errorf("Expected args [test, arg], got %v", parser.Args)
		}
	})

	// Test CommandRegistry.ExecuteCommand function (0% coverage)
	t.Run("CommandRegistry_ExecuteCommand", func(t *testing.T) {
		registry := NewCommandRegistry()

		// Test unknown command error
		t.Run("unknown_command", func(t *testing.T) {
			_, err := registry.ExecuteCommand("unknown", []string{"arg1"})
			if err == nil {
				t.Error("Expected error for unknown command")
			}
			if err.Error() != "unknown command: unknown" {
				t.Errorf("Expected 'unknown command: unknown', got %v", err)
			}
		})

		// Test command with nil parser error
		t.Run("nil_parser", func(t *testing.T) {
			registry.AddCmd("nilcmd", nil)
			_, err := registry.ExecuteCommand("nilcmd", []string{"arg1"})
			if err == nil {
				t.Error("Expected error for nil parser")
			}
			if err.Error() != "command nilcmd has no parser" {
				t.Errorf("Expected 'command nilcmd has no parser', got %v", err)
			}
		})

		// Test successful command execution
		t.Run("successful_execution", func(t *testing.T) {
			subParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"original"}, nil)
			if err != nil {
				t.Fatalf("Failed to create subparser: %v", err)
			}

			registry.AddCmd("testcmd", subParser)

			result, err := registry.ExecuteCommand("testcmd", []string{"new", "args"})
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != subParser {
				t.Error("Expected returned parser to be the same as registered parser")
			}

			// Verify that args were updated
			if len(result.Args) != 2 {
				t.Errorf("Expected 2 args, got %d", len(result.Args))
			}
			if result.Args[0] != "new" || result.Args[1] != "args" {
				t.Errorf("Expected args [new, args], got %v", result.Args)
			}

			// Verify that nonOpts was reset
			if len(result.nonOpts) != 0 {
				t.Errorf("Expected empty nonOpts, got %v", result.nonOpts)
			}
		})
	})

	// Test ExecuteCommandCaseInsensitive function coverage
	t.Run("CommandRegistry_ExecuteCommandCaseInsensitive", func(t *testing.T) {
		registry := NewCommandRegistry()

		// Test unknown command error with case insensitive matching
		t.Run("unknown_command_case_insensitive", func(t *testing.T) {
			_, err := registry.ExecuteCommandCaseInsensitive("UNKNOWN", []string{"arg1"}, true)
			if err == nil {
				t.Error("Expected error for unknown command")
			}
			if err.Error() != "unknown command: UNKNOWN" {
				t.Errorf("Expected 'unknown command: UNKNOWN', got %v", err)
			}
		})

		// Test command with nil parser error (case insensitive)
		t.Run("nil_parser_case_insensitive", func(t *testing.T) {
			registry.AddCmd("nilcmd", nil)
			_, err := registry.ExecuteCommandCaseInsensitive("NILCMD", []string{"arg1"}, true)
			if err == nil {
				t.Error("Expected error for nil parser")
			}
			if err.Error() != "command NILCMD has no parser" {
				t.Errorf("Expected 'command NILCMD has no parser', got %v", err)
			}
		})

		// Test successful case insensitive command execution
		t.Run("successful_case_insensitive_execution", func(t *testing.T) {
			subParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"original"}, nil)
			if err != nil {
				t.Fatalf("Failed to create subparser: %v", err)
			}

			registry.AddCmd("TestCmd", subParser)

			// Execute with different case
			result, err := registry.ExecuteCommandCaseInsensitive("TESTCMD", []string{"case", "insensitive"}, true)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != subParser {
				t.Error("Expected returned parser to be the same as registered parser")
			}

			// Verify that args were updated
			if len(result.Args) != 2 {
				t.Errorf("Expected 2 args, got %d", len(result.Args))
			}
			if result.Args[0] != "case" || result.Args[1] != "insensitive" {
				t.Errorf("Expected args [case, insensitive], got %v", result.Args)
			}

			// Verify that nonOpts was reset
			if len(result.nonOpts) != 0 {
				t.Errorf("Expected empty nonOpts, got %v", result.nonOpts)
			}
		})

		// Test case sensitive mode (caseIgnore = false)
		t.Run("case_sensitive_mode", func(t *testing.T) {
			registry := NewCommandRegistry()
			subParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{}, nil)
			if err != nil {
				t.Fatalf("Failed to create subparser: %v", err)
			}

			registry.AddCmd("TestCmd", subParser)

			// Should fail with case sensitive matching
			_, err = registry.ExecuteCommandCaseInsensitive("testcmd", []string{}, false)
			if err == nil {
				t.Error("Expected error for case mismatch in case sensitive mode")
			}

			// Should succeed with exact case
			result, err := registry.ExecuteCommandCaseInsensitive("TestCmd", []string{}, false)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != subParser {
				t.Error("Expected returned parser to be the same as registered parser")
			}
		})
	})

	// Test additional edge cases for complete coverage
	t.Run("additional_edge_cases", func(t *testing.T) {
		// Test NewParserWithCaseInsensitiveCommands with parent
		t.Run("with_parent", func(t *testing.T) {
			parent, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{}, nil)
			if err != nil {
				t.Fatalf("Failed to create parent parser: %v", err)
			}

			child, err := NewParserWithCaseInsensitiveCommands(map[byte]*Flag{}, map[string]*Flag{}, []string{}, parent)
			if err != nil {
				t.Fatalf("Failed to create child parser: %v", err)
			}

			if child.parent != parent {
				t.Error("Expected child parser to have correct parent")
			}
			if !child.config.commandCaseIgnore {
				t.Error("Expected commandCaseIgnore to be true")
			}
		})

		// Test ExecuteCommand with parser that has existing nonOpts
		t.Run("execute_with_existing_nonOpts", func(t *testing.T) {
			registry := NewCommandRegistry()
			subParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{}, nil)
			if err != nil {
				t.Fatalf("Failed to create subparser: %v", err)
			}

			// Set some existing nonOpts
			subParser.nonOpts = []string{"existing1", "existing2"}

			registry.AddCmd("resetcmd", subParser)

			result, err := registry.ExecuteCommand("resetcmd", []string{"new"})
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify nonOpts was reset to empty
			if len(result.nonOpts) != 0 {
				t.Errorf("Expected empty nonOpts after execution, got %v", result.nonOpts)
			}

			// Verify args were set correctly
			if len(result.Args) != 1 || result.Args[0] != "new" {
				t.Errorf("Expected args [new], got %v", result.Args)
			}
		})
	})
}

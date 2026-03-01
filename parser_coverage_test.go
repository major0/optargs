package optargs

import (
	"testing"
)

// TestCoverageCompletion tests parser internal paths: findShortOpt edge cases,
// Options() mode handling, and subcommand dispatch through the iterator.
func TestCoverageCompletion(t *testing.T) {
	// Test findShortOpt unknown argument type error path
	t.Run("findShortOpt_unknown_argument_type", func(t *testing.T) {
		config := ParserConfig{}
		shortOpts := map[byte]*Flag{
			'x': {HasArg: 999}, // Invalid argument type
		}
		longOpts := map[string]*Flag{}
		args := []string{}

		parser, err := NewParser(config, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		_, _, _, _, err = parser.findShortOpt('x', "", []string{})
		if err == nil {
			t.Error("Expected error for unknown argument type")
		}
		if err.Error() != "unknown argument type: 999" {
			t.Errorf("Expected 'unknown argument type: 999', got %v", err)
		}
	})

	// Test case insensitive short option matching
	t.Run("findShortOpt_case_insensitive", func(t *testing.T) {
		config := ParserConfig{shortCaseIgnore: true}
		shortOpts := map[byte]*Flag{
			'v': {HasArg: NoArgument},
			'f': {HasArg: RequiredArgument},
			'o': {HasArg: OptionalArgument},
		}
		longOpts := map[string]*Flag{}
		args := []string{}

		parser, err := NewParser(config, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// Test case insensitive NoArgument
		_, _, _, option, err := parser.findShortOpt('V', "", []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name != "v" {
			t.Errorf("Expected option name 'v', got '%s'", option.Name)
		}

		// Test case insensitive RequiredArgument from word
		_, _, _, option, err = parser.findShortOpt('F', "value", []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Expected option name 'f', got '%s'", option.Name)
		}
		if option.Arg != "value" {
			t.Errorf("Expected option arg 'value', got '%s'", option.Arg)
		}

		// Test case insensitive OptionalArgument from args
		_, _, _, option, err = parser.findShortOpt('O', "", []string{"value"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name != "o" {
			t.Errorf("Expected option name 'o', got '%s'", option.Name)
		}
		if option.Arg != "value" {
			t.Errorf("Expected option arg 'value', got '%s'", option.Arg)
		}
	})

	// Test Options function missing paths
	t.Run("Options_gnu_words_transformation", func(t *testing.T) {
		config := ParserConfig{gnuWords: true}
		shortOpts := map[byte]*Flag{
			'W': {HasArg: RequiredArgument},
		}
		longOpts := map[string]*Flag{}
		args := []string{"-Wfoo"}

		parser, err := NewParser(config, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		for option, err := range parser.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if option.Name != "foo" {
				t.Errorf("Expected option name 'foo' (GNU words transformation), got '%s'", option.Name)
			}
			break
		}
	})

	t.Run("Options_parse_non_opts_mode", func(t *testing.T) {
		config := ParserConfig{parseMode: ParseNonOpts}
		shortOpts := map[byte]*Flag{}
		longOpts := map[string]*Flag{}
		args := []string{"non-option"}

		parser, err := NewParser(config, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		for option, err := range parser.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if option.Name != string(byte(1)) {
				t.Errorf("Expected option name '\\001', got '%s'", option.Name)
			}
			if option.Arg != "non-option" {
				t.Errorf("Expected option arg 'non-option', got '%s'", option.Arg)
			}
			break
		}
	})

	t.Run("Options_posixly_correct_mode", func(t *testing.T) {
		config := ParserConfig{parseMode: ParsePosixlyCorrect}
		shortOpts := map[byte]*Flag{}
		longOpts := map[string]*Flag{}
		args := []string{"non-option", "-v"}

		parser, err := NewParser(config, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// In POSIXLY_CORRECT mode, parsing should stop at first non-option
		count := 0
		for _, err := range parser.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			count++
		}

		// Should not process any options since first arg is non-option
		if count != 0 {
			t.Errorf("Expected 0 options in POSIXLY_CORRECT mode, got %d", count)
		}

		// Args should contain the non-option and the unprocessed option
		if len(parser.Args) != 2 {
			t.Errorf("Expected 2 args remaining, got %d", len(parser.Args))
		}
	})

	t.Run("Options_long_opts_only_mode", func(t *testing.T) {
		config := ParserConfig{longOptsOnly: true}
		shortOpts := map[byte]*Flag{}
		longOpts := map[string]*Flag{
			"verbose": {HasArg: NoArgument},
		}
		args := []string{"-verbose"}

		parser, err := NewParser(config, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		for option, err := range parser.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if option.Name != "verbose" {
				t.Errorf("Expected option name 'verbose', got '%s'", option.Name)
			}
			break
		}
	})

	t.Run("Options_command_execution", func(t *testing.T) {
		config := ParserConfig{}
		shortOpts := map[byte]*Flag{}
		longOpts := map[string]*Flag{}
		args := []string{"subcmd", "arg1"}

		parser, err := NewParser(config, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// Create a subcommand parser
		subParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{})
		if err != nil {
			t.Fatalf("Failed to create subparser: %v", err)
		}

		parser.AddCmd("subcmd", subParser)

		// Should execute command and stop processing
		count := 0
		for _, err := range parser.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			count++
		}

		// Should not yield any options since command was executed
		if count != 0 {
			t.Errorf("Expected 0 options after command execution, got %d", count)
		}

		// Args should be empty after command execution
		if len(parser.Args) != 0 {
			t.Errorf("Expected 0 args after command execution, got %d", len(parser.Args))
		}
	})

	t.Run("Options_command_execution_error", func(t *testing.T) {
		config := ParserConfig{}
		shortOpts := map[byte]*Flag{}
		longOpts := map[string]*Flag{}
		args := []string{"subcmd"}

		parser, err := NewParser(config, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// Register command with nil parser to trigger error
		parser.AddCmd("subcmd", nil)

		for _, err := range parser.Options() {
			if err == nil {
				t.Error("Expected error from command execution")
			}
			break
		}
	})
}

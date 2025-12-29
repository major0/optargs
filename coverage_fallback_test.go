package optargs

import (
	"testing"
)

// TestFindShortOptWithFallbackCoverage tests all code paths in findShortOptWithFallback
// to achieve 100% coverage for this function
func TestFindShortOptWithFallbackCoverage(t *testing.T) {
	// Test case 1: NoArgument option inheritance
	t.Run("NoArgument_option_inheritance", func(t *testing.T) {
		// Create parent parser with a no-argument option
		parentParser, err := GetOpt([]string{}, "v")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		// Set up parent-child relationship
		childParser.parent = parentParser

		// Test finding parent's option
		args, word, option, err := childParser.findShortOptWithFallback('v', "", []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if option.Name != "v" {
			t.Errorf("Expected option name 'v', got '%s'", option.Name)
		}
		if option.HasArg {
			t.Errorf("Expected HasArg false, got true")
		}
		if len(args) != 0 {
			t.Errorf("Expected empty args, got %v", args)
		}
		if word != "" {
			t.Errorf("Expected empty word, got '%s'", word)
		}
	})

	// Test case 2: RequiredArgument with argument from word
	t.Run("RequiredArgument_from_word", func(t *testing.T) {
		// Create parent parser with required argument option
		parentParser, err := GetOpt([]string{}, "f:")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		// Set up parent-child relationship
		childParser.parent = parentParser

		// Test finding parent's option with argument from remaining word
		args, word, option, err := childParser.findShortOptWithFallback('f', "filename.txt", []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Expected option name 'f', got '%s'", option.Name)
		}
		if !option.HasArg {
			t.Errorf("Expected HasArg true, got false")
		}
		if option.Arg != "filename.txt" {
			t.Errorf("Expected arg 'filename.txt', got '%s'", option.Arg)
		}
		if word != "" {
			t.Errorf("Expected empty word, got '%s'", word)
		}
		_ = args // args should be unchanged
	})

	// Test case 3: RequiredArgument with argument from next arg
	t.Run("RequiredArgument_from_next_arg", func(t *testing.T) {
		// Create parent parser with required argument option
		parentParser, err := GetOpt([]string{}, "f:")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		// Set up parent-child relationship
		childParser.parent = parentParser

		// Test finding parent's option with argument from next arg
		args, word, option, err := childParser.findShortOptWithFallback('f', "", []string{"filename.txt", "other"})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Expected option name 'f', got '%s'", option.Name)
		}
		if !option.HasArg {
			t.Errorf("Expected HasArg true, got false")
		}
		if option.Arg != "filename.txt" {
			t.Errorf("Expected arg 'filename.txt', got '%s'", option.Arg)
		}
		if len(args) != 1 || args[0] != "other" {
			t.Errorf("Expected args ['other'], got %v", args)
		}
		if word != "" {
			t.Errorf("Expected empty word, got '%s'", word)
		}
	})

	// Test case 4: RequiredArgument with missing argument (error case)
	t.Run("RequiredArgument_missing_argument", func(t *testing.T) {
		// Create parent parser with required argument option
		parentParser, err := GetOpt([]string{}, "f:")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		// Set up parent-child relationship
		childParser.parent = parentParser

		// Test finding parent's option with missing argument
		_, _, _, err = childParser.findShortOptWithFallback('f', "", []string{})
		if err == nil {
			t.Errorf("Expected error for missing required argument, got nil")
		}
		if err != nil && err.Error() != "option requires an argument: f" {
			t.Errorf("Expected 'option requires an argument: f', got '%s'", err.Error())
		}
	})

	// Test case 5: OptionalArgument with argument from word
	t.Run("OptionalArgument_from_word", func(t *testing.T) {
		// Create parent parser with optional argument option
		parentParser, err := GetOpt([]string{}, "f::")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		// Set up parent-child relationship
		childParser.parent = parentParser

		// Test finding parent's option with argument from remaining word
		args, word, option, err := childParser.findShortOptWithFallback('f', "filename.txt", []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Expected option name 'f', got '%s'", option.Name)
		}
		if !option.HasArg {
			t.Errorf("Expected HasArg true, got false")
		}
		if option.Arg != "filename.txt" {
			t.Errorf("Expected arg 'filename.txt', got '%s'", option.Arg)
		}
		if word != "" {
			t.Errorf("Expected empty word, got '%s'", word)
		}
		_ = args // args should be unchanged
	})

	// Test case 6: OptionalArgument with argument from next arg
	t.Run("OptionalArgument_from_next_arg", func(t *testing.T) {
		// Create parent parser with optional argument option
		parentParser, err := GetOpt([]string{}, "f::")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		// Set up parent-child relationship
		childParser.parent = parentParser

		// Test finding parent's option with argument from next arg
		args, word, option, err := childParser.findShortOptWithFallback('f', "", []string{"filename.txt", "other"})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Expected option name 'f', got '%s'", option.Name)
		}
		if !option.HasArg {
			t.Errorf("Expected HasArg true, got false")
		}
		if option.Arg != "filename.txt" {
			t.Errorf("Expected arg 'filename.txt', got '%s'", option.Arg)
		}
		if len(args) != 1 || args[0] != "other" {
			t.Errorf("Expected args ['other'], got %v", args)
		}
		if word != "" {
			t.Errorf("Expected empty word, got '%s'", word)
		}
	})

	// Test case 7: OptionalArgument with no argument
	t.Run("OptionalArgument_no_argument", func(t *testing.T) {
		// Create parent parser with optional argument option
		parentParser, err := GetOpt([]string{}, "f::")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		// Set up parent-child relationship
		childParser.parent = parentParser

		// Test finding parent's option with no argument
		args, word, option, err := childParser.findShortOptWithFallback('f', "", []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Expected option name 'f', got '%s'", option.Name)
		}
		if option.HasArg {
			t.Errorf("Expected HasArg false, got true")
		}
		if option.Arg != "" {
			t.Errorf("Expected empty arg, got '%s'", option.Arg)
		}
		if len(args) != 0 {
			t.Errorf("Expected empty args, got %v", args)
		}
		if word != "" {
			t.Errorf("Expected empty word, got '%s'", word)
		}
	})

	// Test case 8: Multi-level parent fallback chain
	t.Run("Multi_level_parent_fallback", func(t *testing.T) {
		// Create grandparent parser with option
		grandparentParser, err := GetOpt([]string{}, "g")
		if err != nil {
			t.Fatalf("Failed to create grandparent parser: %v", err)
		}

		// Create parent parser without the option
		parentParser, err := GetOpt([]string{}, "p")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}
		parentParser.parent = grandparentParser

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "c")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}
		childParser.parent = parentParser

		// Test finding grandparent's option through parent fallback chain
		args, word, option, err := childParser.findShortOptWithFallback('g', "", []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if option.Name != "g" {
			t.Errorf("Expected option name 'g', got '%s'", option.Name)
		}
		if option.HasArg {
			t.Errorf("Expected HasArg false, got true")
		}
		_ = args
		_ = word
	})

	// Test case 9: Option not found anywhere in chain
	t.Run("Option_not_found_in_chain", func(t *testing.T) {
		// Create parent parser without the option we're looking for
		parentParser, err := GetOpt([]string{}, "p")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "c")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}
		childParser.parent = parentParser

		// Test finding non-existent option
		_, _, _, err = childParser.findShortOptWithFallback('x', "", []string{})
		if err == nil {
			t.Errorf("Expected error for unknown option, got nil")
		}
		// The error should be from the original findShortOpt call
		if err != nil && err.Error() != "unknown option: x" {
			t.Errorf("Expected 'unknown option: x', got '%s'", err.Error())
		}
	})
}

// TestFindShortOptCoverage tests the missing 2.5% coverage in findShortOpt
func TestFindShortOptCoverage(t *testing.T) {
	// The missing coverage is likely in edge cases or error paths
	// Let me check what specific lines are not covered by testing various scenarios

	t.Run("Invalid_option_character", func(t *testing.T) {
		parser, err := GetOpt([]string{}, "abc")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// Test with invalid option character (dash)
		_, _, _, err = parser.findShortOpt('-', "", []string{})
		if err == nil {
			t.Errorf("Expected error for invalid option character '-', got nil")
		}
		if err != nil && err.Error() != "invalid option: -" {
			t.Errorf("Expected 'invalid option: -', got '%s'", err.Error())
		}
	})

	t.Run("Unknown_option_character", func(t *testing.T) {
		parser, err := GetOpt([]string{}, "abc")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// Test with unknown option character
		_, _, _, err = parser.findShortOpt('z', "", []string{})
		if err == nil {
			t.Errorf("Expected error for unknown option character 'z', got nil")
		}
		if err != nil && err.Error() != "unknown option: z" {
			t.Errorf("Expected 'unknown option: z', got '%s'", err.Error())
		}
	})

	t.Run("Required_argument_missing", func(t *testing.T) {
		parser, err := GetOpt([]string{}, "f:")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// Test required argument missing
		_, _, _, err = parser.findShortOpt('f', "", []string{})
		if err == nil {
			t.Errorf("Expected error for missing required argument, got nil")
		}
		if err != nil && err.Error() != "option requires an argument: f" {
			t.Errorf("Expected 'option requires an argument: f', got '%s'", err.Error())
		}
	})
}

// TestFindShortOptWithFallbackEdgeCases tests the remaining edge cases for 100% coverage
func TestFindShortOptWithFallbackEdgeCases(t *testing.T) {
	// Test case: Unknown argument type (should never happen in normal usage, but for coverage)
	t.Run("Unknown_argument_type", func(t *testing.T) {
		// Create parent parser
		parentParser, err := GetOpt([]string{}, "f")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		// Create child parser
		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		// Set up parent-child relationship
		childParser.parent = parentParser

		// Manually corrupt the parent's flag to have an invalid HasArg value
		// This is a bit hacky but necessary to test the default case
		parentParser.shortOpts['f'] = &Flag{Name: "f", HasArg: ArgType(999)} // Invalid ArgType

		// Test finding parent's option with invalid argument type
		_, _, _, err = childParser.findShortOptWithFallback('f', "", []string{})
		if err == nil {
			t.Errorf("Expected error for unknown argument type, got nil")
		}
		if err != nil && !containsSubstring(err.Error(), "unknown argument type") {
			t.Errorf("Expected error containing 'unknown argument type', got '%s'", err.Error())
		}
	})

	// Test case: Parent doesn't have the option either, continue fallback chain
	t.Run("Parent_doesnt_have_option_fallback_chain", func(t *testing.T) {
		// Create grandparent parser with the option
		grandparentParser, err := GetOpt([]string{}, "g")
		if err != nil {
			t.Fatalf("Failed to create grandparent parser: %v", err)
		}

		// Create parent parser without the option we're looking for
		parentParser, err := GetOpt([]string{}, "p")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}
		parentParser.parent = grandparentParser

		// Create child parser without the option
		childParser, err := GetOpt([]string{}, "c")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}
		childParser.parent = parentParser

		// Test finding option that parent doesn't have but grandparent does
		// This should trigger the fallback chain continuation
		args, word, option, err := childParser.findShortOptWithFallback('g', "", []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if option.Name != "g" {
			t.Errorf("Expected option name 'g', got '%s'", option.Name)
		}
		_ = args
		_ = word
	})
}

// Helper function to check if a string contains a substring (renamed to avoid conflict)
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

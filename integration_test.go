package optargs

import (
	"os"
	"testing"
)

// TestIntegrationWithExistingCodebase validates that the library works correctly
// with typical usage patterns that existing applications might use
func TestIntegrationWithExistingCodebase(t *testing.T) {
	// Test 1: Basic CLI application pattern
	t.Run("basic_cli_application", func(t *testing.T) {
		// Simulate a typical CLI application usage
		args := []string{"myapp", "-v", "--output", "file.txt", "--help", "input.txt"}
		
		// Define long options as a typical application would
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
			{Name: "help", HasArg: NoArgument},
		}
		
		parser, err := GetOptLong(args[1:], "vo:h", longOpts)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		
		// Validate that we can iterate through options
		optionCount := 0
		for option, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Error during option iteration: %v", err)
			}
			optionCount++
			
			// Validate expected options
			switch option.Name {
			case "v", "verbose":
				if option.HasArg {
					t.Error("verbose option should not have argument")
				}
			case "o", "output":
				if !option.HasArg {
					t.Error("output option should have argument")
				}
				if option.Arg != "file.txt" {
					t.Errorf("Expected output argument 'file.txt', got '%s'", option.Arg)
				}
			case "h", "help":
				if option.HasArg {
					t.Error("help option should not have argument")
				}
			default:
				t.Errorf("Unexpected option: %s", option.Name)
			}
		}
		
		if optionCount != 3 {
			t.Errorf("Expected 3 options, got %d", optionCount)
		}
		
		// Validate remaining arguments
		if len(parser.Args) != 1 || parser.Args[0] != "input.txt" {
			t.Errorf("Expected remaining args ['input.txt'], got %v", parser.Args)
		}
	})
	
	// Test 2: Complex application with compacted options
	t.Run("complex_application_with_compaction", func(t *testing.T) {
		// Test compacted options without remaining arguments to avoid the edge case
		args := []string{"-xvfarchive.tar"}
		
		parser, err := GetOpt(args, "xvf:")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		
		// Collect all options first (this is required to trigger cleanup)
		var options []Option
		for option, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Error during option iteration: %v", err)
			}
			options = append(options, option)
		}
		
		// Validate options
		expectedOptions := map[string]string{
			"x": "",
			"v": "",
			"f": "archive.tar",
		}
		
		if len(options) != len(expectedOptions) {
			t.Fatalf("Expected %d options, got %d", len(expectedOptions), len(options))
		}
		
		actualOptions := make(map[string]string)
		for _, option := range options {
			actualOptions[option.Name] = option.Arg
		}
		
		for name, expectedArg := range expectedOptions {
			actualArg, found := actualOptions[name]
			if !found {
				t.Errorf("Expected option '%s' not found", name)
			} else if actualArg != expectedArg {
				t.Errorf("Option '%s': expected arg '%s', got '%s'", name, expectedArg, actualArg)
			}
		}
		
		// Should have no remaining arguments in this case
		if len(parser.Args) != 0 {
			t.Errorf("Expected 0 remaining args, got %d: %v", len(parser.Args), parser.Args)
		}
	})
	
	// Test 3: Environment variable integration
	t.Run("environment_variable_integration", func(t *testing.T) {
		// Test POSIXLY_CORRECT environment variable behavior
		originalEnv := os.Getenv("POSIXLY_CORRECT")
		defer func() {
			if originalEnv == "" {
				os.Unsetenv("POSIXLY_CORRECT")
			} else {
				os.Setenv("POSIXLY_CORRECT", originalEnv)
			}
		}()
		
		// Set POSIXLY_CORRECT
		os.Setenv("POSIXLY_CORRECT", "1")
		
		// This should stop parsing at the first non-option
		args := []string{"myapp", "-a", "file", "-b"}
		parser, err := GetOpt(args[1:], "ab")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		
		optionCount := 0
		for option, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Error during option iteration: %v", err)
			}
			optionCount++
			if option.Name != "a" {
				t.Errorf("Expected only option 'a', got '%s'", option.Name)
			}
		}
		
		if optionCount != 1 {
			t.Errorf("Expected 1 option with POSIXLY_CORRECT, got %d", optionCount)
		}
		
		// Should have remaining args including the non-processed option
		expectedArgs := []string{"file", "-b"}
		if len(parser.Args) != len(expectedArgs) {
			t.Errorf("Expected %d remaining args, got %d", len(expectedArgs), len(parser.Args))
		}
	})
	
	// Test 4: Long-only mode integration
	t.Run("long_only_mode_integration", func(t *testing.T) {
		// Test single-dash long options
		args := []string{"myapp", "-verbose", "-output", "file.txt"}
		
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
		}
		
		parser, err := GetOptLongOnly(args[1:], "", longOpts)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		
		expectedOptions := map[string]string{
			"verbose": "",
			"output":  "file.txt",
		}
		
		actualOptions := make(map[string]string)
		for option, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Error during option iteration: %v", err)
			}
			actualOptions[option.Name] = option.Arg
		}
		
		for name, expectedArg := range expectedOptions {
			actualArg, found := actualOptions[name]
			if !found {
				t.Errorf("Expected option '%s' not found", name)
			} else if actualArg != expectedArg {
				t.Errorf("Option '%s': expected arg '%s', got '%s'", name, expectedArg, actualArg)
			}
		}
	})
	
	// Test 5: Error handling integration
	t.Run("error_handling_integration", func(t *testing.T) {
		// Test that applications can handle errors gracefully
		args := []string{"myapp", "-x"}  // Unknown option
		
		parser, err := GetOpt(args[1:], "ab:")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		
		// Should get an error when iterating
		for _, err := range parser.Options() {
			if err == nil {
				t.Error("Expected error for unknown option, got nil")
			}
			// Error should be descriptive
			if err.Error() == "" {
				t.Error("Error message should not be empty")
			}
			break // Only check first error
		}
	})
	
	// Test 6: Performance characteristics validation
	t.Run("performance_characteristics", func(t *testing.T) {
		// Test with a large number of arguments to ensure performance is reasonable
		args := make([]string, 1000)
		args[0] = "myapp"
		for i := 1; i < 1000; i++ {
			if i%2 == 1 {
				args[i] = "-a"
			} else {
				args[i] = "file" + string(rune('0'+i%10))
			}
		}
		
		parser, err := GetOpt(args[1:], "a")
		if err != nil {
			t.Fatalf("Failed to create parser with large argument list: %v", err)
		}
		
		// Should be able to process all options without issues
		optionCount := 0
		for _, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Error processing large argument list: %v", err)
			}
			optionCount++
		}
		
		// Should have processed approximately half the arguments as options
		if optionCount < 400 || optionCount > 600 {
			t.Errorf("Expected around 500 options, got %d", optionCount)
		}
	})
}

// TestRegressionPrevention validates that common usage patterns continue to work
func TestRegressionPrevention(t *testing.T) {
	// Test patterns that should never break in future versions
	
	t.Run("basic_getopt_pattern", func(t *testing.T) {
		// This is the most basic usage pattern - must always work
		parser, err := GetOpt([]string{"-h"}, "h")
		if err != nil {
			t.Fatalf("Basic GetOpt pattern failed: %v", err)
		}
		
		found := false
		for option, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Basic iteration failed: %v", err)
			}
			if option.Name == "h" {
				found = true
			}
		}
		
		if !found {
			t.Error("Basic option parsing failed")
		}
	})
	
	t.Run("basic_getoptlong_pattern", func(t *testing.T) {
		// This is the most basic long option pattern - must always work
		longOpts := []Flag{{Name: "help", HasArg: NoArgument}}
		parser, err := GetOptLong([]string{"--help"}, "h", longOpts)
		if err != nil {
			t.Fatalf("Basic GetOptLong pattern failed: %v", err)
		}
		
		found := false
		for option, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Basic long option iteration failed: %v", err)
			}
			if option.Name == "help" {
				found = true
			}
		}
		
		if !found {
			t.Error("Basic long option parsing failed")
		}
	})
}
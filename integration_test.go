package optargs

import (
	"testing"
)

// collectNamedOptions iterates a parser and returns a map of option name → arg value.
func collectNamedOptions(t *testing.T, p *Parser) map[string]string {
	t.Helper()
	result := make(map[string]string)
	for opt, err := range p.Options() {
		if err != nil {
			t.Fatalf("unexpected error during iteration: %v", err)
		}
		result[opt.Name] = opt.Arg
	}
	return result
}

// TestIntegrationWithExistingCodebase validates that the library works correctly
// with typical usage patterns that existing applications might use.
func TestIntegrationWithExistingCodebase(t *testing.T) {
	t.Run("basic_cli_application", func(t *testing.T) {
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
			{Name: "help", HasArg: NoArgument},
		}

		parser, err := GetOptLong([]string{"-v", "--output", "file.txt", "--help", "input.txt"}, "vo:h", longOpts)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		optionCount := 0
		for opt, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Error during option iteration: %v", err)
			}
			optionCount++

			switch opt.Name {
			case "v", "verbose":
				if opt.HasArg {
					t.Error("verbose option should not have argument")
				}
			case "o", "output":
				if !opt.HasArg {
					t.Error("output option should have argument")
				}
				if opt.Arg != "file.txt" {
					t.Errorf("Expected output argument 'file.txt', got '%s'", opt.Arg)
				}
			case "h", "help":
				if opt.HasArg {
					t.Error("help option should not have argument")
				}
			default:
				t.Errorf("Unexpected option: %s", opt.Name)
			}
		}

		if optionCount != 3 {
			t.Errorf("Expected 3 options, got %d", optionCount)
		}

		if len(parser.Args) != 1 || parser.Args[0] != "input.txt" {
			t.Errorf("Expected remaining args ['input.txt'], got %v", parser.Args)
		}
	})

	t.Run("compacted_short_options", func(t *testing.T) {
		parser, err := GetOpt([]string{"-xvfarchive.tar"}, "xvf:")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		got := collectNamedOptions(t, parser)

		expected := map[string]string{
			"x": "",
			"v": "",
			"f": "archive.tar",
		}
		if len(got) != len(expected) {
			t.Fatalf("Expected %d options, got %d", len(expected), len(got))
		}
		for name, wantArg := range expected {
			if gotArg, ok := got[name]; !ok {
				t.Errorf("Expected option '%s' not found", name)
			} else if gotArg != wantArg {
				t.Errorf("Option '%s': expected arg '%s', got '%s'", name, wantArg, gotArg)
			}
		}

		if len(parser.Args) != 0 {
			t.Errorf("Expected 0 remaining args, got %d: %v", len(parser.Args), parser.Args)
		}
	})

	t.Run("posixly_correct_env", func(t *testing.T) {
		t.Setenv("POSIXLY_CORRECT", "1")

		parser, err := GetOpt([]string{"-a", "file", "-b"}, "ab")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		got := collectNamedOptions(t, parser)
		if len(got) != 1 {
			t.Errorf("Expected 1 option with POSIXLY_CORRECT, got %d", len(got))
		}
		if _, ok := got["a"]; !ok {
			t.Error("Expected option 'a' to be parsed")
		}

		expectedArgs := []string{"file", "-b"}
		if len(parser.Args) != len(expectedArgs) {
			t.Errorf("Expected %d remaining args, got %d", len(expectedArgs), len(parser.Args))
		}
	})

	t.Run("long_only_mode", func(t *testing.T) {
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
		}

		parser, err := GetOptLongOnly([]string{"-verbose", "-output", "file.txt"}, "", longOpts)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		got := collectNamedOptions(t, parser)

		expected := map[string]string{
			"verbose": "",
			"output":  "file.txt",
		}
		for name, wantArg := range expected {
			if gotArg, ok := got[name]; !ok {
				t.Errorf("Expected option '%s' not found", name)
			} else if gotArg != wantArg {
				t.Errorf("Option '%s': expected arg '%s', got '%s'", name, wantArg, gotArg)
			}
		}
	})

	t.Run("error_handling", func(t *testing.T) {
		parser, err := GetOpt([]string{"-x"}, "ab:")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		for _, err := range parser.Options() {
			if err == nil {
				t.Error("Expected error for unknown option, got nil")
			}
			if err.Error() == "" {
				t.Error("Error message should not be empty")
			}
			break
		}
	})
}

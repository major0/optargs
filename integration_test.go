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

// assertNamedOptions compares a got map against expected option name → arg pairs.
func assertNamedOptions(t *testing.T, got, expected map[string]string) {
	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf("Expected %d options, got %d: %v", len(expected), len(got), got)
	}
	for name, wantArg := range expected {
		gotArg, ok := got[name]
		if !ok {
			t.Errorf("Expected option '%s' not found", name)
		} else if gotArg != wantArg {
			t.Errorf("Option '%s': expected arg '%s', got '%s'", name, wantArg, gotArg)
		}
	}
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

		got := collectNamedOptions(t, parser)
		assertNamedOptions(t, got, map[string]string{
			"v":      "",
			"output": "file.txt",
			"help":   "",
		})

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
		assertNamedOptions(t, got, map[string]string{
			"x": "",
			"v": "",
			"f": "archive.tar",
		})

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
		assertNamedOptions(t, got, map[string]string{
			"a": "",
		})

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
		assertNamedOptions(t, got, map[string]string{
			"verbose": "",
			"output":  "file.txt",
		})
	})

	t.Run("error_handling", func(t *testing.T) {
		parser, err := GetOpt([]string{"-x"}, "ab:")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		for _, err := range parser.Options() {
			if err == nil {
				t.Error("expected error for unknown option, got nil")
			}
			if err.Error() == "" {
				t.Error("Error message should not be empty")
			}
			break
		}
	})
}

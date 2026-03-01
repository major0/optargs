package optargs

import (
	"strings"
	"testing"
)

// TestFindShortOptCoverage tests all code paths in findShortOpt
func TestFindShortOptCoverage(t *testing.T) {
	t.Run("NoArgument_option_inheritance", func(t *testing.T) {
		parentParser, err := GetOpt([]string{}, "v")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		parentParser.AddCmd("child", childParser)

		args, word, _, option, err := childParser.findShortOpt('v', "", []string{})
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

	t.Run("RequiredArgument_from_word", func(t *testing.T) {
		parentParser, err := GetOpt([]string{}, "f:")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		parentParser.AddCmd("child", childParser)

		args, word, _, option, err := childParser.findShortOpt('f', "filename.txt", []string{})
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
		_ = args
	})

	t.Run("RequiredArgument_from_next_arg", func(t *testing.T) {
		parentParser, err := GetOpt([]string{}, "f:")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		parentParser.AddCmd("child", childParser)

		args, word, _, option, err := childParser.findShortOpt('f', "", []string{"filename.txt", "other"})
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

	t.Run("RequiredArgument_missing_argument", func(t *testing.T) {
		parentParser, err := GetOpt([]string{}, "f:")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		parentParser.AddCmd("child", childParser)

		_, _, _, _, err = childParser.findShortOpt('f', "", []string{})
		if err == nil {
			t.Errorf("Expected error for missing required argument, got nil")
		}
		if err != nil && err.Error() != "option requires an argument: f" {
			t.Errorf("Expected 'option requires an argument: f', got '%s'", err.Error())
		}
	})

	t.Run("OptionalArgument_from_word", func(t *testing.T) {
		parentParser, err := GetOpt([]string{}, "f::")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		parentParser.AddCmd("child", childParser)

		args, word, _, option, err := childParser.findShortOpt('f', "filename.txt", []string{})
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
		_ = args
	})

	t.Run("OptionalArgument_from_next_arg", func(t *testing.T) {
		parentParser, err := GetOpt([]string{}, "f::")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		parentParser.AddCmd("child", childParser)

		args, word, _, option, err := childParser.findShortOpt('f', "", []string{"filename.txt", "other"})
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

	t.Run("OptionalArgument_no_argument", func(t *testing.T) {
		parentParser, err := GetOpt([]string{}, "f::")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		parentParser.AddCmd("child", childParser)

		args, word, _, option, err := childParser.findShortOpt('f', "", []string{})
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

	t.Run("Multi_level_parent_fallback", func(t *testing.T) {
		grandparentParser, err := GetOpt([]string{}, "g")
		if err != nil {
			t.Fatalf("Failed to create grandparent parser: %v", err)
		}

		parentParser, err := GetOpt([]string{}, "p")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}
		grandparentParser.AddCmd("parent", parentParser)

		childParser, err := GetOpt([]string{}, "c")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}
		parentParser.AddCmd("child", childParser)

		args, word, _, option, err := childParser.findShortOpt('g', "", []string{})
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

	t.Run("Option_not_found_in_chain", func(t *testing.T) {
		parentParser, err := GetOpt([]string{}, "p")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		childParser, err := GetOpt([]string{}, "c")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}
		parentParser.AddCmd("child", childParser)

		_, _, _, _, err = childParser.findShortOpt('x', "", []string{})
		if err == nil {
			t.Errorf("Expected error for unknown option, got nil")
		}
		if err != nil && err.Error() != "unknown option: x" {
			t.Errorf("Expected 'unknown option: x', got '%s'", err.Error())
		}
	})
}

// TestFindShortOptDirectCoverage tests edge cases in findShortOpt
func TestFindShortOptDirectCoverage(t *testing.T) {
	t.Run("Invalid_option_character", func(t *testing.T) {
		parser, err := GetOpt([]string{}, "abc")
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		_, _, _, _, err = parser.findShortOpt('-', "", []string{})
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

		_, _, _, _, err = parser.findShortOpt('z', "", []string{})
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

		_, _, _, _, err = parser.findShortOpt('f', "", []string{})
		if err == nil {
			t.Errorf("Expected error for missing required argument, got nil")
		}
		if err != nil && err.Error() != "option requires an argument: f" {
			t.Errorf("Expected 'option requires an argument: f', got '%s'", err.Error())
		}
	})
}

// TestFindShortOptEdgeCases tests remaining edge cases
func TestFindShortOptEdgeCases(t *testing.T) {
	t.Run("Unknown_argument_type", func(t *testing.T) {
		parentParser, err := GetOpt([]string{}, "f")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}

		childParser, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}

		parentParser.AddCmd("child", childParser)

		// Corrupt the parent's flag to have an invalid HasArg value
		parentParser.shortOpts['f'] = &Flag{Name: "f", HasArg: ArgType(999)}

		_, _, _, _, err = childParser.findShortOpt('f', "", []string{})
		if err == nil {
			t.Errorf("Expected error for unknown argument type, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "unknown argument type") {
			t.Errorf("Expected error containing 'unknown argument type', got '%s'", err.Error())
		}
	})

	t.Run("Parent_doesnt_have_option_fallback_chain", func(t *testing.T) {
		grandparentParser, err := GetOpt([]string{}, "g")
		if err != nil {
			t.Fatalf("Failed to create grandparent parser: %v", err)
		}

		parentParser, err := GetOpt([]string{}, "p")
		if err != nil {
			t.Fatalf("Failed to create parent parser: %v", err)
		}
		grandparentParser.AddCmd("parent", parentParser)

		childParser, err := GetOpt([]string{}, "c")
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}
		parentParser.AddCmd("child", childParser)

		args, word, _, option, err := childParser.findShortOpt('g', "", []string{})
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

// TestFallbackErrorModesThroughChain verifies that the originating parser's
// error mode controls error reporting, not the parent's.
func TestFallbackErrorModesThroughChain(t *testing.T) {
	// Helper: create a parser with explicit error mode
	makeParser := func(t *testing.T, optstring string, longOpts []Flag, args []string) *Parser {
		t.Helper()
		var p *Parser
		var err error
		if len(longOpts) > 0 {
			p, err = GetOptLong(args, optstring, longOpts)
		} else {
			p, err = GetOpt(args, optstring)
		}
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}
		return p
	}

	t.Run("silent_child_verbose_parent_short_unknown", func(t *testing.T) {
		// Parent: verbose (no ':' prefix), has option 'v'
		parent := makeParser(t, "v", nil, []string{})
		// Child: silent (':' prefix), no options
		child := makeParser(t, ":", nil, []string{"-x"})
		parent.AddCmd("child", child)

		// Unknown option 'x' — not in child or parent.
		// Child is silent, so no slog.Error should fire.
		// We verify the error is returned correctly.
		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "unknown option: x" {
					t.Errorf("Expected 'unknown option: x', got '%s'", err.Error())
				}
				return
			}
		}
		t.Error("Expected error for unknown option 'x'")
	})

	t.Run("silent_child_verbose_parent_short_found_in_parent", func(t *testing.T) {
		// Parent: verbose, has option 'v'
		parent := makeParser(t, "v", nil, []string{})
		// Child: silent, no options
		child := makeParser(t, ":", nil, []string{"-v"})
		parent.AddCmd("child", child)

		found := false
		for opt, err := range child.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				continue
			}
			if opt.Name == "v" {
				found = true
			}
		}
		if !found {
			t.Error("Expected to find parent option 'v'")
		}
	})

	t.Run("silent_child_verbose_parent_long_unknown", func(t *testing.T) {
		parentLong := []Flag{{Name: "verbose", HasArg: NoArgument}}
		parent := makeParser(t, "", parentLong, []string{})
		child := makeParser(t, ":", nil, []string{"--unknown"})
		parent.AddCmd("child", child)

		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "unknown option: unknown" {
					t.Errorf("Expected 'unknown option: unknown', got '%s'", err.Error())
				}
				return
			}
		}
		t.Error("Expected error for unknown long option")
	})

	t.Run("silent_child_verbose_parent_long_found_in_parent", func(t *testing.T) {
		parentLong := []Flag{{Name: "verbose", HasArg: NoArgument}}
		parent := makeParser(t, "", parentLong, []string{})
		child := makeParser(t, ":", nil, []string{"--verbose"})
		parent.AddCmd("child", child)

		found := false
		for opt, err := range child.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				continue
			}
			if opt.Name == "verbose" {
				found = true
			}
		}
		if !found {
			t.Error("Expected to find parent long option 'verbose'")
		}
	})

	t.Run("silent_child_verbose_parent_missing_arg_in_parent", func(t *testing.T) {
		// Parent: verbose, has 'f' requiring argument
		parent := makeParser(t, "f:", nil, []string{})
		// Child: silent, no options, passes -f with no arg
		child := makeParser(t, ":", nil, []string{"-f"})
		parent.AddCmd("child", child)

		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "option requires an argument: f" {
					t.Errorf("Expected 'option requires an argument: f', got '%s'", err.Error())
				}
				return
			}
		}
		t.Error("Expected error for missing argument")
	})

	t.Run("silent_child_verbose_parent_long_missing_arg_in_parent", func(t *testing.T) {
		parentLong := []Flag{{Name: "file", HasArg: RequiredArgument}}
		parent := makeParser(t, "", parentLong, []string{})
		child := makeParser(t, ":", nil, []string{"--file"})
		parent.AddCmd("child", child)

		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "option requires an argument: file" {
					t.Errorf("Expected 'option requires an argument: file', got '%s'", err.Error())
				}
				return
			}
		}
		t.Error("Expected error for missing long option argument")
	})

	t.Run("verbose_child_silent_parent_unknown", func(t *testing.T) {
		// Parent: silent
		parent := makeParser(t, ":", nil, []string{})
		// Child: verbose (default)
		child := makeParser(t, "", nil, []string{"-x"})
		parent.AddCmd("child", child)

		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "unknown option: x" {
					t.Errorf("Expected 'unknown option: x', got '%s'", err.Error())
				}
				return
			}
		}
		t.Error("Expected error for unknown option")
	})

	t.Run("multi_level_silent_child_verbose_ancestors", func(t *testing.T) {
		// Root: verbose, has 'r'
		root := makeParser(t, "r", nil, []string{})
		// Mid: verbose, has 'm'
		mid := makeParser(t, "m", nil, []string{})
		root.AddCmd("mid", mid)
		// Leaf: silent, no options
		leaf := makeParser(t, ":", nil, []string{"-r", "-m", "-x"})
		mid.AddCmd("leaf", leaf)

		foundR := false
		foundM := false
		var lastErr error
		for opt, err := range leaf.Options() {
			if err != nil {
				lastErr = err
				continue
			}
			if opt.Name == "r" {
				foundR = true
			}
			if opt.Name == "m" {
				foundM = true
			}
		}
		if !foundR {
			t.Error("Expected to find root option 'r'")
		}
		if !foundM {
			t.Error("Expected to find mid option 'm'")
		}
		if lastErr == nil {
			t.Error("Expected error for unknown option 'x'")
		}
	})
}

// TestMultiLevelInheritanceViaIterator tests option inheritance through multiple
// levels using the Options() iterator (not direct fallback calls).
func TestMultiLevelInheritanceViaIterator(t *testing.T) {
	t.Run("short_and_long_opts_4_levels", func(t *testing.T) {
		rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'r': {Name: "r", HasArg: NoArgument},
		}, map[string]*Flag{
			"root": {Name: "root", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("Failed to create root parser: %v", err)
		}

		level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'a': {Name: "a", HasArg: NoArgument},
		}, map[string]*Flag{
			"level1": {Name: "level1", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("Failed to create level1 parser: %v", err)
		}
		rootParser.AddCmd("level1", level1Parser)

		level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'b': {Name: "b", HasArg: NoArgument},
		}, map[string]*Flag{
			"level2": {Name: "level2", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("Failed to create level2 parser: %v", err)
		}
		level1Parser.AddCmd("level2", level2Parser)

		level3Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'c': {Name: "c", HasArg: NoArgument},
		}, map[string]*Flag{
			"level3": {Name: "level3", HasArg: NoArgument},
		}, []string{"-r", "-a", "-b", "-c"})
		if err != nil {
			t.Fatalf("Failed to create level3 parser: %v", err)
		}
		level2Parser.AddCmd("level3", level3Parser)

		foundOptions := make(map[string]bool)
		for option, err := range level3Parser.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				continue
			}
			foundOptions[option.Name] = true
		}

		for _, expected := range []string{"r", "a", "b", "c"} {
			if !foundOptions[expected] {
				t.Errorf("Expected to find option '%s' but didn't", expected)
			}
		}
	})

	t.Run("inherited_options_with_arguments", func(t *testing.T) {
		rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'v': {Name: "v", HasArg: RequiredArgument},
		}, map[string]*Flag{}, []string{})
		if err != nil {
			t.Fatalf("Failed to create root parser: %v", err)
		}

		level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'o': {Name: "o", HasArg: OptionalArgument},
		}, map[string]*Flag{}, []string{})
		if err != nil {
			t.Fatalf("Failed to create level1 parser: %v", err)
		}
		rootParser.AddCmd("level1", level1Parser)

		level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'f': {Name: "f", HasArg: RequiredArgument},
		}, map[string]*Flag{}, []string{"-v", "verbose", "-o", "optional", "-f", "file"})
		if err != nil {
			t.Fatalf("Failed to create level2 parser: %v", err)
		}
		level1Parser.AddCmd("level2", level2Parser)

		expected := map[string]string{
			"v": "verbose",
			"o": "optional",
			"f": "file",
		}

		found := make(map[string]string)
		for option, err := range level2Parser.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				continue
			}
			found[option.Name] = option.Arg
		}

		for name, arg := range expected {
			if foundArg, exists := found[name]; !exists {
				t.Errorf("Expected to find option '%s' but didn't", name)
			} else if foundArg != arg {
				t.Errorf("Expected option '%s' arg '%s', got '%s'", name, arg, foundArg)
			}
		}
	})

	t.Run("inherited_long_options", func(t *testing.T) {
		rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
			"verbose": {Name: "verbose", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("Failed to create root parser: %v", err)
		}

		level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
			"output": {Name: "output", HasArg: RequiredArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("Failed to create level1 parser: %v", err)
		}
		rootParser.AddCmd("level1", level1Parser)

		level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
			"file": {Name: "file", HasArg: RequiredArgument},
		}, []string{"--verbose", "--output", "out.txt", "--file", "input.txt"})
		if err != nil {
			t.Fatalf("Failed to create level2 parser: %v", err)
		}
		level1Parser.AddCmd("level2", level2Parser)

		expected := map[string]string{
			"verbose": "",
			"output":  "out.txt",
			"file":    "input.txt",
		}

		found := make(map[string]string)
		for option, err := range level2Parser.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				continue
			}
			found[option.Name] = option.Arg
		}

		for name, arg := range expected {
			if foundArg, exists := found[name]; !exists {
				t.Errorf("Expected to find option '%s' but didn't", name)
			} else if foundArg != arg {
				t.Errorf("Expected option '%s' arg '%s', got '%s'", name, arg, foundArg)
			}
		}
	})
}

// TestParentChainMissingArgDeferral verifies that when a child parser with a parent
// encounters a missing required argument, the error is deferred (not logged by the child).
func TestParentChainMissingArgDeferral(t *testing.T) {
	t.Run("long_opt_missing_arg_with_parent", func(t *testing.T) {
		root, err := GetOptLong([]string{"sub", "--file"}, ":", []Flag{
			{Name: "file", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		child, err := GetOpt([]string{}, ":")
		if err != nil {
			t.Fatalf("Failed to create child: %v", err)
		}
		root.AddCmd("sub", child)

		// Dispatch
		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("Root error: %v", err)
			}
		}

		// Child iterates — --file found in parent but missing arg
		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Error("Expected error for missing long option argument via parent chain")
		}
		if foundErr != nil && !strings.Contains(foundErr.Error(), "option requires an argument") {
			t.Errorf("Expected 'option requires an argument', got '%s'", foundErr.Error())
		}
	})

	t.Run("short_opt_missing_arg_with_parent", func(t *testing.T) {
		root, err := GetOpt([]string{"sub", "-f"}, ":f:")
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		child, err := GetOpt([]string{}, ":")
		if err != nil {
			t.Fatalf("Failed to create child: %v", err)
		}
		root.AddCmd("sub", child)

		// Dispatch
		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("Root error: %v", err)
			}
		}

		// Child iterates — -f found in parent but missing arg
		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Error("Expected error for missing short option argument via parent chain")
		}
		if foundErr != nil && !strings.Contains(foundErr.Error(), "option requires an argument") {
			t.Errorf("Expected 'option requires an argument', got '%s'", foundErr.Error())
		}
	})
}

// TestChildOwnOptionMissingArgWithParent verifies that when a child parser's own
// option (not inherited) requires an argument and none is provided, the error is
// deferred via the parent-chain path.
func TestChildOwnOptionMissingArgWithParent(t *testing.T) {
	t.Run("child_long_opt_missing_arg", func(t *testing.T) {
		root, err := GetOpt([]string{"sub", "--port"}, ":")
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		child, err := GetOptLong([]string{}, ":", []Flag{
			{Name: "port", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create child: %v", err)
		}
		root.AddCmd("sub", child)

		// Dispatch
		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("Root error: %v", err)
			}
		}

		// Child iterates — --port is child's own option but arg is missing
		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Error("Expected error for missing argument on child's own long option")
		}
	})

	t.Run("child_short_opt_missing_arg", func(t *testing.T) {
		root, err := GetOpt([]string{"sub", "-p"}, ":")
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		child, err := GetOpt([]string{}, ":p:")
		if err != nil {
			t.Fatalf("Failed to create child: %v", err)
		}
		root.AddCmd("sub", child)

		// Dispatch
		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("Root error: %v", err)
			}
		}

		// Child iterates — -p is child's own option but arg is missing
		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Error("Expected error for missing argument on child's own short option")
		}
	})
}

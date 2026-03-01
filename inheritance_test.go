package optargs

import (
	"strings"
	"testing"
)

// childOf creates a child parser linked to a parent via AddCmd.
func childOf(t *testing.T, parentOpts, childOpts string) (*Parser, *Parser) {
	t.Helper()
	parent, err := GetOpt([]string{}, parentOpts)
	if err != nil {
		t.Fatalf("parent parser: %v", err)
	}
	child, err := GetOpt([]string{}, childOpts)
	if err != nil {
		t.Fatalf("child parser: %v", err)
	}
	parent.AddCmd("child", child)
	return parent, child
}

// TestFindShortOptCoverage tests all code paths in findShortOpt via
// parent-chain inheritance.
func TestFindShortOptCoverage(t *testing.T) {
	t.Run("NoArgument_option_inheritance", func(t *testing.T) {
		_, child := childOf(t, "v", "")

		args, word, _, option, err := child.findShortOpt('v', "", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if option.Name != "v" {
			t.Errorf("Name = %q, want %q", option.Name, "v")
		}
		if option.HasArg {
			t.Error("HasArg = true, want false")
		}
		if len(args) != 0 {
			t.Errorf("args = %v, want empty", args)
		}
		if word != "" {
			t.Errorf("word = %q, want empty", word)
		}
	})

	t.Run("RequiredArgument_from_word", func(t *testing.T) {
		_, child := childOf(t, "f:", "")

		_, _, _, option, err := child.findShortOpt('f', "filename.txt", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Name = %q, want %q", option.Name, "f")
		}
		if !option.HasArg {
			t.Error("HasArg = false, want true")
		}
		if option.Arg != "filename.txt" {
			t.Errorf("Arg = %q, want %q", option.Arg, "filename.txt")
		}
	})

	t.Run("RequiredArgument_from_next_arg", func(t *testing.T) {
		_, child := childOf(t, "f:", "")

		args, _, _, option, err := child.findShortOpt('f', "", []string{"filename.txt", "other"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Name = %q, want %q", option.Name, "f")
		}
		if !option.HasArg {
			t.Error("HasArg = false, want true")
		}
		if option.Arg != "filename.txt" {
			t.Errorf("Arg = %q, want %q", option.Arg, "filename.txt")
		}
		if len(args) != 1 || args[0] != "other" {
			t.Errorf("args = %v, want [other]", args)
		}
	})

	t.Run("RequiredArgument_missing_argument", func(t *testing.T) {
		_, child := childOf(t, "f:", "")

		_, _, _, _, err := child.findShortOpt('f', "", []string{})
		if err == nil {
			t.Fatal("expected error for missing required argument")
		}
		if err.Error() != "option requires an argument: f" {
			t.Errorf("error = %q, want %q", err.Error(), "option requires an argument: f")
		}
	})

	t.Run("OptionalArgument_from_word", func(t *testing.T) {
		_, child := childOf(t, "f::", "")

		_, _, _, option, err := child.findShortOpt('f', "filename.txt", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Name = %q, want %q", option.Name, "f")
		}
		if !option.HasArg {
			t.Error("HasArg = false, want true")
		}
		if option.Arg != "filename.txt" {
			t.Errorf("Arg = %q, want %q", option.Arg, "filename.txt")
		}
	})

	t.Run("OptionalArgument_from_next_arg", func(t *testing.T) {
		_, child := childOf(t, "f::", "")

		args, _, _, option, err := child.findShortOpt('f', "", []string{"filename.txt", "other"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Name = %q, want %q", option.Name, "f")
		}
		if !option.HasArg {
			t.Error("HasArg = false, want true")
		}
		if option.Arg != "filename.txt" {
			t.Errorf("Arg = %q, want %q", option.Arg, "filename.txt")
		}
		if len(args) != 1 || args[0] != "other" {
			t.Errorf("args = %v, want [other]", args)
		}
	})

	t.Run("OptionalArgument_no_argument", func(t *testing.T) {
		_, child := childOf(t, "f::", "")

		args, word, _, option, err := child.findShortOpt('f', "", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if option.Name != "f" {
			t.Errorf("Name = %q, want %q", option.Name, "f")
		}
		if option.HasArg {
			t.Error("HasArg = true, want false")
		}
		if option.Arg != "" {
			t.Errorf("Arg = %q, want empty", option.Arg)
		}
		if len(args) != 0 {
			t.Errorf("args = %v, want empty", args)
		}
		if word != "" {
			t.Errorf("word = %q, want empty", word)
		}
	})

	t.Run("Multi_level_parent_fallback", func(t *testing.T) {
		grandparent, err := GetOpt([]string{}, "g")
		if err != nil {
			t.Fatalf("grandparent parser: %v", err)
		}
		parent, err := GetOpt([]string{}, "p")
		if err != nil {
			t.Fatalf("parent parser: %v", err)
		}
		grandparent.AddCmd("parent", parent)
		child, err := GetOpt([]string{}, "c")
		if err != nil {
			t.Fatalf("child parser: %v", err)
		}
		parent.AddCmd("child", child)

		_, _, _, option, err := child.findShortOpt('g', "", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if option.Name != "g" {
			t.Errorf("Name = %q, want %q", option.Name, "g")
		}
		if option.HasArg {
			t.Error("HasArg = true, want false")
		}
	})

	t.Run("Option_not_found_in_chain", func(t *testing.T) {
		_, child := childOf(t, "p", "c")

		_, _, _, _, err := child.findShortOpt('x', "", []string{})
		if err == nil {
			t.Fatal("expected error for unknown option")
		}
		if err.Error() != "unknown option: x" {
			t.Errorf("error = %q, want %q", err.Error(), "unknown option: x")
		}
	})
}

// TestFindShortOptDirectCoverage tests findShortOpt error paths on a
// single parser (no inheritance chain).
func TestFindShortOptDirectCoverage(t *testing.T) {
	tests := []struct {
		name    string
		char    byte
		wantErr string
	}{
		{"invalid_option_dash", '-', "invalid option: -"},
		{"unknown_option", 'z', "unknown option: z"},
	}

	parser, err := GetOpt([]string{}, "abc")
	if err != nil {
		t.Fatalf("parser: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, _, err := parser.findShortOpt(tt.char, "", []string{})
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}

	t.Run("required_argument_missing", func(t *testing.T) {
		p, err := GetOpt([]string{}, "f:")
		if err != nil {
			t.Fatalf("parser: %v", err)
		}
		_, _, _, _, err = p.findShortOpt('f', "", []string{})
		if err == nil {
			t.Fatal("expected error for missing required argument")
		}
		if err.Error() != "option requires an argument: f" {
			t.Errorf("error = %q, want %q", err.Error(), "option requires an argument: f")
		}
	})
}

// TestFindShortOptEdgeCases tests remaining edge cases in findShortOpt
// via inheritance.
func TestFindShortOptEdgeCases(t *testing.T) {
	t.Run("Unknown_argument_type", func(t *testing.T) {
		parent, child := childOf(t, "f", "")

		// Corrupt the parent's flag to have an invalid HasArg value.
		parent.shortOpts['f'] = &Flag{Name: "f", HasArg: ArgType(999)}

		_, _, _, _, err := child.findShortOpt('f', "", []string{})
		if err == nil {
			t.Fatal("expected error for unknown argument type")
		}
		if !strings.Contains(err.Error(), "unknown argument type") {
			t.Errorf("error = %q, want containing %q", err.Error(), "unknown argument type")
		}
	})
}

// TestFallbackErrorModesThroughChain verifies that the originating parser's
// error mode controls error reporting, not the parent's.
func TestFallbackErrorModesThroughChain(t *testing.T) {
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
			t.Fatalf("parser: %v", err)
		}
		return p
	}

	t.Run("silent_child_verbose_parent_short_unknown", func(t *testing.T) {
		parent := makeParser(t, "v", nil, []string{})
		child := makeParser(t, ":", nil, []string{"-x"})
		parent.AddCmd("child", child)

		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "unknown option: x" {
					t.Errorf("error = %q, want %q", err.Error(), "unknown option: x")
				}
				return
			}
		}
		t.Error("expected error for unknown option 'x'")
	})

	t.Run("silent_child_verbose_parent_short_found_in_parent", func(t *testing.T) {
		parent := makeParser(t, "v", nil, []string{})
		child := makeParser(t, ":", nil, []string{"-v"})
		parent.AddCmd("child", child)

		found := false
		for opt, err := range child.Options() {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			if opt.Name == "v" {
				found = true
			}
		}
		if !found {
			t.Error("expected to find parent option 'v'")
		}
	})

	t.Run("silent_child_verbose_parent_long_unknown", func(t *testing.T) {
		parent := makeParser(t, "", []Flag{{Name: "verbose", HasArg: NoArgument}}, []string{})
		child := makeParser(t, ":", nil, []string{"--unknown"})
		parent.AddCmd("child", child)

		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "unknown option: unknown" {
					t.Errorf("error = %q, want %q", err.Error(), "unknown option: unknown")
				}
				return
			}
		}
		t.Error("expected error for unknown long option")
	})

	t.Run("silent_child_verbose_parent_long_found_in_parent", func(t *testing.T) {
		parent := makeParser(t, "", []Flag{{Name: "verbose", HasArg: NoArgument}}, []string{})
		child := makeParser(t, ":", nil, []string{"--verbose"})
		parent.AddCmd("child", child)

		found := false
		for opt, err := range child.Options() {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			if opt.Name == "verbose" {
				found = true
			}
		}
		if !found {
			t.Error("expected to find parent long option 'verbose'")
		}
	})

	t.Run("silent_child_verbose_parent_missing_arg_in_parent", func(t *testing.T) {
		parent := makeParser(t, "f:", nil, []string{})
		child := makeParser(t, ":", nil, []string{"-f"})
		parent.AddCmd("child", child)

		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "option requires an argument: f" {
					t.Errorf("error = %q, want %q", err.Error(), "option requires an argument: f")
				}
				return
			}
		}
		t.Error("expected error for missing argument")
	})

	t.Run("silent_child_verbose_parent_long_missing_arg_in_parent", func(t *testing.T) {
		parent := makeParser(t, "", []Flag{{Name: "file", HasArg: RequiredArgument}}, []string{})
		child := makeParser(t, ":", nil, []string{"--file"})
		parent.AddCmd("child", child)

		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "option requires an argument: file" {
					t.Errorf("error = %q, want %q", err.Error(), "option requires an argument: file")
				}
				return
			}
		}
		t.Error("expected error for missing long option argument")
	})

	t.Run("verbose_child_silent_parent_unknown", func(t *testing.T) {
		parent := makeParser(t, ":", nil, []string{})
		child := makeParser(t, "", nil, []string{"-x"})
		parent.AddCmd("child", child)

		for _, err := range child.Options() {
			if err != nil {
				if err.Error() != "unknown option: x" {
					t.Errorf("error = %q, want %q", err.Error(), "unknown option: x")
				}
				return
			}
		}
		t.Error("expected error for unknown option")
	})

	t.Run("multi_level_silent_child_verbose_ancestors", func(t *testing.T) {
		root := makeParser(t, "r", nil, []string{})
		mid := makeParser(t, "m", nil, []string{})
		root.AddCmd("mid", mid)
		leaf := makeParser(t, ":", nil, []string{"-r", "-m", "-x"})
		mid.AddCmd("leaf", leaf)

		foundR, foundM := false, false
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
			t.Error("expected to find root option 'r'")
		}
		if !foundM {
			t.Error("expected to find mid option 'm'")
		}
		if lastErr == nil {
			t.Error("expected error for unknown option 'x'")
		}
	})
}

// TestMultiLevelInheritanceViaIterator tests option inheritance through
// multiple levels using the Options() iterator.
func TestMultiLevelInheritanceViaIterator(t *testing.T) {
	t.Run("short_and_long_opts_4_levels", func(t *testing.T) {
		rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'r': {Name: "r", HasArg: NoArgument},
		}, map[string]*Flag{
			"root": {Name: "root", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("root parser: %v", err)
		}

		level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'a': {Name: "a", HasArg: NoArgument},
		}, map[string]*Flag{
			"level1": {Name: "level1", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("level1 parser: %v", err)
		}
		rootParser.AddCmd("level1", level1Parser)

		level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'b': {Name: "b", HasArg: NoArgument},
		}, map[string]*Flag{
			"level2": {Name: "level2", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("level2 parser: %v", err)
		}
		level1Parser.AddCmd("level2", level2Parser)

		level3Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'c': {Name: "c", HasArg: NoArgument},
		}, map[string]*Flag{
			"level3": {Name: "level3", HasArg: NoArgument},
		}, []string{"-r", "-a", "-b", "-c"})
		if err != nil {
			t.Fatalf("level3 parser: %v", err)
		}
		level2Parser.AddCmd("level3", level3Parser)

		foundOptions := make(map[string]bool)
		for option, err := range level3Parser.Options() {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			foundOptions[option.Name] = true
		}

		for _, expected := range []string{"r", "a", "b", "c"} {
			if !foundOptions[expected] {
				t.Errorf("missing option %q", expected)
			}
		}
	})

	t.Run("inherited_options_with_arguments", func(t *testing.T) {
		rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'v': {Name: "v", HasArg: RequiredArgument},
		}, map[string]*Flag{}, []string{})
		if err != nil {
			t.Fatalf("root parser: %v", err)
		}

		level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'o': {Name: "o", HasArg: OptionalArgument},
		}, map[string]*Flag{}, []string{})
		if err != nil {
			t.Fatalf("level1 parser: %v", err)
		}
		rootParser.AddCmd("level1", level1Parser)

		level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'f': {Name: "f", HasArg: RequiredArgument},
		}, map[string]*Flag{}, []string{"-v", "verbose", "-o", "optional", "-f", "file"})
		if err != nil {
			t.Fatalf("level2 parser: %v", err)
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
				t.Errorf("unexpected error: %v", err)
				continue
			}
			found[option.Name] = option.Arg
		}

		for name, arg := range expected {
			if foundArg, exists := found[name]; !exists {
				t.Errorf("missing option %q", name)
			} else if foundArg != arg {
				t.Errorf("option %q: Arg = %q, want %q", name, foundArg, arg)
			}
		}
	})

	t.Run("inherited_long_options", func(t *testing.T) {
		rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
			"verbose": {Name: "verbose", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("root parser: %v", err)
		}

		level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
			"output": {Name: "output", HasArg: RequiredArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("level1 parser: %v", err)
		}
		rootParser.AddCmd("level1", level1Parser)

		level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
			"file": {Name: "file", HasArg: RequiredArgument},
		}, []string{"--verbose", "--output", "out.txt", "--file", "input.txt"})
		if err != nil {
			t.Fatalf("level2 parser: %v", err)
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
				t.Errorf("unexpected error: %v", err)
				continue
			}
			found[option.Name] = option.Arg
		}

		for name, arg := range expected {
			if foundArg, exists := found[name]; !exists {
				t.Errorf("missing option %q", name)
			} else if foundArg != arg {
				t.Errorf("option %q: Arg = %q, want %q", name, foundArg, arg)
			}
		}
	})
}

// TestParentChainMissingArgDeferral verifies that when a child parser
// with a parent encounters a missing required argument, the error is
// deferred (not logged by the child).
func TestParentChainMissingArgDeferral(t *testing.T) {
	t.Run("long_opt_missing_arg_with_parent", func(t *testing.T) {
		root, err := GetOptLong([]string{"sub", "--file"}, ":", []Flag{
			{Name: "file", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("root: %v", err)
		}

		child, err := GetOpt([]string{}, ":")
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		root.AddCmd("sub", child)

		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("root error: %v", err)
			}
		}

		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Fatal("expected error for missing long option argument via parent chain")
		}
		if !strings.Contains(foundErr.Error(), "option requires an argument") {
			t.Errorf("error = %q, want containing %q", foundErr.Error(), "option requires an argument")
		}
	})

	t.Run("short_opt_missing_arg_with_parent", func(t *testing.T) {
		root, err := GetOpt([]string{"sub", "-f"}, ":f:")
		if err != nil {
			t.Fatalf("root: %v", err)
		}

		child, err := GetOpt([]string{}, ":")
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		root.AddCmd("sub", child)

		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("root error: %v", err)
			}
		}

		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Fatal("expected error for missing short option argument via parent chain")
		}
		if !strings.Contains(foundErr.Error(), "option requires an argument") {
			t.Errorf("error = %q, want containing %q", foundErr.Error(), "option requires an argument")
		}
	})
}

// TestChildOwnOptionMissingArgWithParent verifies that when a child
// parser's own option requires an argument and none is provided, the
// error is deferred via the parent-chain path.
func TestChildOwnOptionMissingArgWithParent(t *testing.T) {
	t.Run("child_long_opt_missing_arg", func(t *testing.T) {
		root, err := GetOpt([]string{"sub", "--port"}, ":")
		if err != nil {
			t.Fatalf("root: %v", err)
		}

		child, err := GetOptLong([]string{}, ":", []Flag{
			{Name: "port", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		root.AddCmd("sub", child)

		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("root error: %v", err)
			}
		}

		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Fatal("expected error for missing argument on child's own long option")
		}
	})

	t.Run("child_short_opt_missing_arg", func(t *testing.T) {
		root, err := GetOpt([]string{"sub", "-p"}, ":")
		if err != nil {
			t.Fatalf("root: %v", err)
		}

		child, err := GetOpt([]string{}, ":p:")
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		root.AddCmd("sub", child)

		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("root error: %v", err)
			}
		}

		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Fatal("expected error for missing argument on child's own short option")
		}
	})
}

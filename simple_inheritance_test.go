package optargs

import (
	"testing"
)

// TestSimpleInheritance tests basic parent-child inheritance
func TestSimpleInheritance(t *testing.T) {
	// Create parent parser with option 'p'
	parentShortOpts := map[byte]*Flag{
		'p': {Name: "p", HasArg: NoArgument},
	}

	parentParser, err := NewParser(ParserConfig{}, parentShortOpts, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("Failed to create parent parser: %v", err)
	}

	// Create child parser with option 'c' and parent option 'p' in args
	childShortOpts := map[byte]*Flag{
		'c': {Name: "c", HasArg: NoArgument},
	}

	childParser, err := NewParser(ParserConfig{}, childShortOpts, map[string]*Flag{}, []string{"-p", "-c"}, parentParser)
	if err != nil {
		t.Fatalf("Failed to create child parser: %v", err)
	}

	// Test that child can access both its own option and parent option
	foundOptions := make(map[string]bool)

	for option, err := range childParser.Options() {
		if err != nil {
			t.Errorf("Unexpected error parsing option: %v", err)
			continue
		}

		foundOptions[option.Name] = true
		t.Logf("Found option: %s", option.Name)
	}

	// Check that both options were found
	if !foundOptions["c"] {
		t.Error("Expected to find child option 'c' but didn't")
	}

	if !foundOptions["p"] {
		t.Error("Expected to find parent option 'p' but didn't")
	}
}

// TestDirectFallback tests the fallback methods directly
func TestDirectFallback(t *testing.T) {
	// Create parent parser with option 'p'
	parentShortOpts := map[byte]*Flag{
		'p': {Name: "p", HasArg: NoArgument},
	}

	parentParser, err := NewParser(ParserConfig{}, parentShortOpts, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("Failed to create parent parser: %v", err)
	}

	// Create child parser without option 'p'
	childShortOpts := map[byte]*Flag{
		'c': {Name: "c", HasArg: NoArgument},
	}

	childParser, err := NewParser(ParserConfig{}, childShortOpts, map[string]*Flag{}, []string{}, parentParser)
	if err != nil {
		t.Fatalf("Failed to create child parser: %v", err)
	}

	// Test direct fallback for option 'p' (should find in parent)
	_, _, option, err := childParser.findShortOptWithFallback('p', "", []string{})
	if err != nil {
		t.Errorf("Expected to find option 'p' in parent but got error: %v", err)
	} else if option.Name != "p" {
		t.Errorf("Expected option name 'p', got '%s'", option.Name)
	} else {
		t.Logf("Successfully found parent option 'p': %+v", option)
	}

	// Test direct fallback for option 'c' (should find in child)
	_, _, option, err = childParser.findShortOptWithFallback('c', "", []string{})
	if err != nil {
		t.Errorf("Expected to find option 'c' in child but got error: %v", err)
	} else if option.Name != "c" {
		t.Errorf("Expected option name 'c', got '%s'", option.Name)
	} else {
		t.Logf("Successfully found child option 'c': %+v", option)
	}

	// Test direct fallback for non-existent option 'x' (should fail)
	_, _, option, err = childParser.findShortOptWithFallback('x', "", []string{})
	if err == nil {
		t.Error("Expected error for non-existent option 'x' but got none")
	} else {
		t.Logf("Correctly got error for non-existent option 'x': %v", err)
	}
}

// TestDebugInheritanceSimple tests the inheritance with debug output
func TestDebugInheritanceSimple(t *testing.T) {
	// Create the same 4-level hierarchy as the failing test
	rootShortOpts := map[byte]*Flag{
		'r': {Name: "r", HasArg: NoArgument},
	}

	level1ShortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
	}

	level2ShortOpts := map[byte]*Flag{
		'b': {Name: "b", HasArg: NoArgument},
	}

	level3ShortOpts := map[byte]*Flag{
		'c': {Name: "c", HasArg: NoArgument},
	}

	// Create the parser hierarchy
	rootParser, err := NewParser(ParserConfig{}, rootShortOpts, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	level1Parser, err := NewParser(ParserConfig{}, level1ShortOpts, map[string]*Flag{}, []string{}, rootParser)
	if err != nil {
		t.Fatalf("Failed to create level1 parser: %v", err)
	}

	level2Parser, err := NewParser(ParserConfig{}, level2ShortOpts, map[string]*Flag{}, []string{}, level1Parser)
	if err != nil {
		t.Fatalf("Failed to create level2 parser: %v", err)
	}

	level3Parser, err := NewParser(ParserConfig{}, level3ShortOpts, map[string]*Flag{}, []string{"-r", "-a", "-b", "-c"}, level2Parser)
	if err != nil {
		t.Fatalf("Failed to create level3 parser: %v", err)
	}

	// Debug the parser hierarchy
	t.Logf("Root parser options: %v", getOptionNames(rootParser.shortOpts))
	t.Logf("Level1 parser options: %v, parent: %v", getOptionNames(level1Parser.shortOpts), level1Parser.parent != nil)
	t.Logf("Level2 parser options: %v, parent: %v", getOptionNames(level2Parser.shortOpts), level2Parser.parent != nil)
	t.Logf("Level3 parser options: %v, parent: %v", getOptionNames(level3Parser.shortOpts), level3Parser.parent != nil)

	// Test direct fallback for each option
	testOptions := []struct {
		option byte
		name   string
	}{
		{'c', "level3 own option"},
		{'b', "level2 inherited option"},
		{'a', "level1 inherited option"},
		{'r', "root inherited option"},
	}

	for _, test := range testOptions {
		t.Run(test.name, func(t *testing.T) {
			_, _, option, err := level3Parser.findShortOptWithFallback(test.option, "", []string{})
			if err != nil {
				t.Errorf("Expected to find option '%c' but got error: %v", test.option, err)
			} else {
				t.Logf("Found option '%c' via fallback: %+v", test.option, option)
			}
		})
	}

	// Test full parsing
	t.Run("FullParsing", func(t *testing.T) {
		foundOptions := make(map[string]bool)

		for option, err := range level3Parser.Options() {
			if err != nil {
				t.Errorf("Unexpected error parsing option: %v", err)
				continue
			}

			foundOptions[option.Name] = true
			t.Logf("Found option during parsing: %s", option.Name)
		}

		expectedOptions := []string{"c", "b", "a", "r"}
		for _, expected := range expectedOptions {
			if !foundOptions[expected] {
				t.Errorf("Expected to find option '%s' but didn't", expected)
			}
		}
	})
}

// Helper function to get option names from a map
func getOptionNames(opts map[byte]*Flag) []string {
	var names []string
	for key := range opts {
		names = append(names, string(key))
	}
	return names
}

package optargs

import (
	"strings"
	"testing"
)

// TestInfiniteSubcommandInheritance tests option inheritance through multiple levels of subcommands
func TestInfiniteSubcommandInheritance(t *testing.T) {
	// Create a 4-level hierarchy: root -> level1 -> level2 -> level3
	// Each level has its own options, and child levels should inherit from all parents

	// Level 3 (deepest) - has option 'c'
	level3ShortOpts := map[byte]*Flag{
		'c': {Name: "c", HasArg: NoArgument},
	}
	level3LongOpts := map[string]*Flag{
		"level3": {Name: "level3", HasArg: NoArgument},
	}

	// Level 2 - has option 'b' and will be parent of level3
	level2ShortOpts := map[byte]*Flag{
		'b': {Name: "b", HasArg: NoArgument},
	}
	level2LongOpts := map[string]*Flag{
		"level2": {Name: "level2", HasArg: NoArgument},
	}

	// Level 1 - has option 'a' and will be parent of level2
	level1ShortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
	}
	level1LongOpts := map[string]*Flag{
		"level1": {Name: "level1", HasArg: NoArgument},
	}

	// Root level - has option 'r' and will be parent of level1
	rootShortOpts := map[byte]*Flag{
		'r': {Name: "r", HasArg: NoArgument},
	}
	rootLongOpts := map[string]*Flag{
		"root": {Name: "root", HasArg: NoArgument},
	}

	// Create the parser hierarchy
	rootParser, err := NewParser(ParserConfig{}, rootShortOpts, rootLongOpts, []string{}, nil)
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	level1Parser, err := NewParser(ParserConfig{}, level1ShortOpts, level1LongOpts, []string{}, rootParser)
	if err != nil {
		t.Fatalf("Failed to create level1 parser: %v", err)
	}

	level2Parser, err := NewParser(ParserConfig{}, level2ShortOpts, level2LongOpts, []string{}, level1Parser)
	if err != nil {
		t.Fatalf("Failed to create level2 parser: %v", err)
	}

	_, err = NewParser(ParserConfig{}, level3ShortOpts, level3LongOpts, []string{"-r", "-a", "-b", "-c"}, level2Parser)
	if err != nil {
		t.Fatalf("Failed to create level3 parser: %v", err)
	}

	// Test that level3 can access options from all parent levels
	tests := []struct {
		name          string
		option        byte
		expectedName  string
		shouldFind    bool
		expectedLevel string
	}{
		{"level3 own option", 'c', "c", true, "level3"},
		{"level2 inherited option", 'b', "b", true, "level2"},
		{"level1 inherited option", 'a', "a", true, "level1"},
		{"root inherited option", 'r', "r", true, "root"},
		{"non-existent option", 'x', "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh level3Parser for each test to avoid argument consumption
			freshLevel3Parser, err := NewParser(ParserConfig{}, level3ShortOpts, level3LongOpts, []string{"-r", "-a", "-b", "-c"}, level2Parser)
			if err != nil {
				t.Fatalf("Failed to create fresh level3 parser: %v", err)
			}

			// Test the option parsing through the inheritance chain
			var foundOption Option

			// Parse options from fresh level3 parser
			for option, err := range freshLevel3Parser.Options() {
				if err != nil {
					if !tt.shouldFind && option.Name == string(tt.option) {
						// Expected error for non-existent option
						break
					}
					continue
				}

				t.Logf("Found option during parsing: %s (looking for %c)", option.Name, tt.option)
				if option.Name == string(tt.option) {
					foundOption = option
					break
				}
			}

			if tt.shouldFind {
				if foundOption.Name == "" {
					t.Errorf("Expected to find option '%c' but didn't", tt.option)
				} else if foundOption.Name != tt.expectedName {
					t.Errorf("Expected option name '%s', got '%s'", tt.expectedName, foundOption.Name)
				}
			} else {
				if foundOption.Name != "" {
					t.Errorf("Expected not to find option '%c' but found '%s'", tt.option, foundOption.Name)
				}
			}
		})
	}
}

// TestInfiniteInheritanceWithArguments tests inheritance with options that take arguments
func TestInfiniteInheritanceWithArguments(t *testing.T) {
	// Create a 3-level hierarchy with options that take arguments

	// Level 2 (deepest) - has option 'f' with required argument
	level2ShortOpts := map[byte]*Flag{
		'f': {Name: "f", HasArg: RequiredArgument},
	}

	// Level 1 - has option 'o' with optional argument
	level1ShortOpts := map[byte]*Flag{
		'o': {Name: "o", HasArg: OptionalArgument},
	}

	// Root level - has option 'v' with required argument
	rootShortOpts := map[byte]*Flag{
		'v': {Name: "v", HasArg: RequiredArgument},
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

	level2Parser, err := NewParser(ParserConfig{}, level2ShortOpts, map[string]*Flag{}, []string{"-v", "verbose", "-o", "optional", "-f", "file"}, level1Parser)
	if err != nil {
		t.Fatalf("Failed to create level2 parser: %v", err)
	}

	// Test parsing options with arguments through inheritance
	expectedOptions := map[string]string{
		"v": "verbose",
		"o": "optional",
		"f": "file",
	}

	foundOptions := make(map[string]string)

	for option, err := range level2Parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error parsing option: %v", err)
			continue
		}

		foundOptions[option.Name] = option.Arg
	}

	for expectedName, expectedArg := range expectedOptions {
		if foundArg, exists := foundOptions[expectedName]; !exists {
			t.Errorf("Expected to find option '%s' but didn't", expectedName)
		} else if foundArg != expectedArg {
			t.Errorf("Expected option '%s' to have argument '%s', got '%s'", expectedName, expectedArg, foundArg)
		}
	}
}

// TestInfiniteInheritanceLongOptions tests inheritance with long options
func TestInfiniteInheritanceLongOptions(t *testing.T) {
	// Create a 3-level hierarchy with long options

	// Level 2 (deepest) - has long option 'file'
	level2LongOpts := map[string]*Flag{
		"file": {Name: "file", HasArg: RequiredArgument},
	}

	// Level 1 - has long option 'output'
	level1LongOpts := map[string]*Flag{
		"output": {Name: "output", HasArg: RequiredArgument},
	}

	// Root level - has long option 'verbose'
	rootLongOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
	}

	// Create the parser hierarchy
	rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, rootLongOpts, []string{}, nil)
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, level1LongOpts, []string{}, rootParser)
	if err != nil {
		t.Fatalf("Failed to create level1 parser: %v", err)
	}

	level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, level2LongOpts, []string{"--verbose", "--output", "out.txt", "--file", "input.txt"}, level1Parser)
	if err != nil {
		t.Fatalf("Failed to create level2 parser: %v", err)
	}

	// Test parsing long options through inheritance
	expectedOptions := map[string]string{
		"verbose": "",
		"output":  "out.txt",
		"file":    "input.txt",
	}

	foundOptions := make(map[string]string)

	for option, err := range level2Parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error parsing option: %v", err)
			continue
		}

		foundOptions[option.Name] = option.Arg
	}

	for expectedName, expectedArg := range expectedOptions {
		if foundArg, exists := foundOptions[expectedName]; !exists {
			t.Errorf("Expected to find option '%s' but didn't", expectedName)
		} else if foundArg != expectedArg {
			t.Errorf("Expected option '%s' to have argument '%s', got '%s'", expectedName, expectedArg, foundArg)
		}
	}
}

// TestInfiniteInheritanceErrorHandling tests error handling in inheritance chain
func TestInfiniteInheritanceErrorHandling(t *testing.T) {
	// Create a simple 2-level hierarchy
	rootShortOpts := map[byte]*Flag{
		'v': {Name: "v", HasArg: RequiredArgument},
	}

	rootParser, err := NewParser(ParserConfig{}, rootShortOpts, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("Failed to create root parser: %v", err)
	}

	// Child parser with an option that requires an argument but doesn't get one
	childParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"-v"}, rootParser)
	if err != nil {
		t.Fatalf("Failed to create child parser: %v", err)
	}

	// Test that error handling works correctly through inheritance
	var foundError error
	for _, err := range childParser.Options() {
		if err != nil {
			foundError = err
			break
		}
	}

	if foundError == nil {
		t.Error("Expected error for option requiring argument but got none")
	} else if !strings.Contains(foundError.Error(), "option requires an argument") {
		t.Errorf("Expected 'option requires an argument' error, got: %v", foundError)
	}
}

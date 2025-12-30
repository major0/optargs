package optargs

import (
	"strings"
	"testing"
)

// TestFindShortOptParentFallbackPath tests the specific error path in findShortOpt
// when a parser has a parent but the option is not found locally.
// This covers the missing coverage in the final error handling paths.
func TestFindShortOptParentFallbackPath(t *testing.T) {
	// Create parent parser with an option
	parentParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'p': {Name: "parent", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("Failed to create parent parser: %v", err)
	}

	// Create child parser without the option
	childParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'c': {Name: "child", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{}, parentParser)
	if err != nil {
		t.Fatalf("Failed to create child parser: %v", err)
	}

	// Set up parent-child relationship
	childParser.parent = parentParser

	// Test case 1: Call findShortOpt directly (not findShortOptWithFallback)
	// on child parser for an option that doesn't exist in child but exists in parent
	// This should return an error without logging (the parent fallback path)
	_, _, _, err = childParser.findShortOpt('p', "", []string{})
	if err == nil {
		t.Error("Expected error when calling findShortOpt directly on child parser for parent's option")
	}
	if !strings.Contains(err.Error(), "unknown option: p") {
		t.Errorf("Expected 'unknown option: p' error, got: %v", err)
	}

	// Test case 2: Call findShortOpt directly on child parser for non-existent option
	// This should also return an error without logging (the parent fallback path)
	_, _, _, err = childParser.findShortOpt('x', "", []string{})
	if err == nil {
		t.Error("Expected error when calling findShortOpt directly on child parser for non-existent option")
	}
	if !strings.Contains(err.Error(), "unknown option: x") {
		t.Errorf("Expected 'unknown option: x' error, got: %v", err)
	}

	// Test case 3: Call findShortOpt directly on parser with no parent
	// This should log the error (the no-parent path)
	_, _, _, err = parentParser.findShortOpt('x', "", []string{})
	if err == nil {
		t.Error("Expected error when calling findShortOpt on parser with no parent for non-existent option")
	}
	if !strings.Contains(err.Error(), "unknown option: x") {
		t.Errorf("Expected 'unknown option: x' error, got: %v", err)
	}
}

// TestFindShortOptAllErrorPaths ensures all error paths in findShortOpt are covered
func TestFindShortOptAllErrorPaths(t *testing.T) {
	// Test the dash character error path
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	_, _, _, err = parser.findShortOpt('-', "", []string{})
	if err == nil {
		t.Error("Expected error for dash character")
	}
	if !strings.Contains(err.Error(), "invalid option: -") {
		t.Errorf("Expected 'invalid option: -' error, got: %v", err)
	}

	// Test the unknown argument type error path (default case in switch)
	parser.shortOpts = map[byte]*Flag{
		'x': {Name: "x", HasArg: ArgType(999)}, // Invalid argument type
	}

	_, _, _, err = parser.findShortOpt('x', "", []string{})
	if err == nil {
		t.Error("Expected error for unknown argument type")
	}
	if !strings.Contains(err.Error(), "unknown argument type: 999") {
		t.Errorf("Expected 'unknown argument type: 999' error, got: %v", err)
	}
}

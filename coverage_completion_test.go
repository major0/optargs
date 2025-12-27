package optargs

import (
	"testing"
)

// TestFindShortOptDefaultCase tests the default case in findShortOpt switch statement
// This covers the "unknown argument type" error path that was missing coverage
func TestFindShortOptDefaultCase(t *testing.T) {
	parser := &Parser{
		shortOpts: map[byte]*Flag{
			'x': {Name: "x", HasArg: 99}, // Invalid HasArg value to trigger default case
		},
		config: ParserConfig{},
	}

	_, _, _, err := parser.findShortOpt('x', "", []string{})
	if err == nil {
		t.Error("Expected error for unknown argument type")
	}
	
	expectedMsg := "unknown argument type: 99"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestOptionsCommandExecution tests command execution path in Options iterator
// This covers the command execution branch that was missing coverage
func TestOptionsCommandExecution(t *testing.T) {
	// Create a subcommand parser
	subParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		's': {Name: "s", HasArg: NoArgument},
	}, map[string]*Flag{
		"sub-flag": {Name: "sub-flag", HasArg: NoArgument},
	}, []string{})
	if err != nil {
		t.Fatalf("Failed to create subparser: %v", err)
	}
	
	// Create main parser with command
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"subcmd", "--sub-flag"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	parser.AddCmd("subcmd", subParser)
	
	// Iterate through options - this should execute the command
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
		// The command execution should handle the sub-flag
		if option.Name == "" {
			break // Command execution completed
		}
	}
	
	// After command execution, Args should be empty
	if len(parser.Args) != 0 {
		t.Errorf("Expected Args to be empty after command execution, got %v", parser.Args)
	}
}

// TestOptionsParseNonOptsMode tests ParseNonOpts mode in Options iterator
// This covers the ParseNonOpts branch that was missing coverage
func TestOptionsParseNonOptsMode(t *testing.T) {
	parser, err := NewParser(ParserConfig{parseMode: ParseNonOpts}, map[byte]*Flag{}, map[string]*Flag{}, []string{"non-option-arg"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionFound := false
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		// In ParseNonOpts mode, non-options are yielded as options with Name = byte(1)
		if option.Name == string(byte(1)) && option.Arg == "non-option-arg" {
			optionFound = true
		}
		break // Only check first option
	}
	
	if !optionFound {
		t.Error("Expected non-option to be yielded as option in ParseNonOpts mode")
	}
}

// TestOptionsLongOptsOnlyMode tests longOptsOnly mode with short option syntax
// This covers the longOptsOnly branch in short option handling
func TestOptionsLongOptsOnlyMode(t *testing.T) {
	parser, err := NewParser(ParserConfig{longOptsOnly: true}, map[byte]*Flag{}, map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
	}, []string{"-verbose"}) // Short syntax but treated as long option
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionFound := false
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if option.Name == "verbose" {
			optionFound = true
		}
		break
	}
	
	if !optionFound {
		t.Error("Expected long option to be found in longOptsOnly mode")
	}
}

// TestOptionsGnuWordsTransformation tests the -W option transformation
// This covers the GNU words transformation branch
func TestOptionsGnuWordsTransformation(t *testing.T) {
	parser, err := NewParser(ParserConfig{gnuWords: true}, map[byte]*Flag{
		'W': {Name: "W", HasArg: RequiredArgument},
	}, map[string]*Flag{}, []string{"-W", "foo"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionFound := false
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		// -W foo should be transformed to --foo
		if option.Name == "foo" && option.Arg == "foo" {
			optionFound = true
		}
		break
	}
	
	if !optionFound {
		t.Error("Expected -W option to be transformed to long option name")
	}
}

// TestOptionsPosixlyCorrectMode tests POSIXLY_CORRECT parsing mode
// This covers the ParsePosixlyCorrect branch
func TestOptionsPosixlyCorrectMode(t *testing.T) {
	parser, err := NewParser(ParserConfig{parseMode: ParsePosixlyCorrect}, map[byte]*Flag{}, map[string]*Flag{
		"flag": {Name: "flag", HasArg: NoArgument},
	}, []string{"non-option", "--flag"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	// In POSIXLY_CORRECT mode, parsing should stop at first non-option
	optionCount := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
	}
	
	// Should not have processed any options due to early break
	if optionCount > 0 {
		t.Error("Expected no options to be processed in POSIXLY_CORRECT mode with leading non-option")
	}
	
	// Args should contain both the non-option and the unprocessed flag
	if len(parser.Args) != 2 {
		t.Errorf("Expected 2 args remaining, got %d: %v", len(parser.Args), parser.Args)
	}
}

// TestOptionsYieldFalseEarlyReturn tests early return when yield returns false
// This covers the yield return value checking branches
func TestOptionsYieldFalseEarlyReturn(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-ab"}) // Two options in one argument
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
		
		// Return false after first option to test early return
		if optionCount == 1 {
			if option.Name != "a" {
				t.Errorf("Expected first option to be 'a', got '%s'", option.Name)
			}
			break // This simulates yield returning false
		}
	}
	
	// Should have only processed one option due to early return
	if optionCount != 1 {
		t.Errorf("Expected exactly 1 option to be processed, got %d", optionCount)
	}
}

// TestFindShortOptCaseInsensitiveMatch tests case insensitive matching in findShortOpt
// This may cover additional branches in the case matching logic
func TestFindShortOptCaseInsensitiveMatch(t *testing.T) {
	parser, err := NewParser(ParserConfig{shortCaseIgnore: true}, map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test case insensitive match
	_, _, option, err := parser.findShortOpt('A', "", []string{})
	if err != nil {
		t.Errorf("Unexpected error with case insensitive match: %v", err)
	}
	if option.Name != "a" {
		t.Errorf("Expected option name 'a', got '%s'", option.Name)
	}
}

// TestOptionsIteratorEarlyBreak tests various early break conditions in Options iterator
func TestOptionsIteratorEarlyBreak(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-a", "non-option"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	count := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		count++
		if count == 1 {
			if option.Name != "a" {
				t.Errorf("Expected first option 'a', got '%s'", option.Name)
			}
			// Break early to test cleanup logic
			break
		}
	}

	// Should have processed exactly one option
	if count != 1 {
		t.Errorf("Expected 1 option processed, got %d", count)
	}
}

// TestOptionsCommandNotFound tests command lookup failure path
func TestOptionsCommandNotFound(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"unknown-command"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	// Add a command registry but not the command we're looking for
	subParser, _ := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{})
	parser.AddCmd("known-command", subParser)

	// This should not execute a command, should treat as non-option
	optionCount := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
	}

	// Should have processed the unknown command as a non-option
	if len(parser.Args) == 0 {
		t.Error("Expected unknown command to remain in Args")
	}
}

// TestFindShortOptDashError tests the specific error case for '-' character
func TestFindShortOptDashError(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	_, _, _, err = parser.findShortOpt('-', "", []string{})
	if err == nil {
		t.Error("Expected error for '-' character as short option")
	}
	
	expectedMsg := "invalid option: -"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestOptionsDefaultCaseWithoutCommands tests the default case when no commands are registered
func TestOptionsDefaultCaseWithoutCommands(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"plain-arg"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Don't add any commands - this should hit the default case without command checking
	optionCount := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
	}

	// Should have processed the argument as a non-option
	if len(parser.Args) == 0 {
		t.Error("Expected plain argument to remain in Args")
	}
}
// TestFindShortOptAllBranches tests all remaining branches in findShortOpt
func TestFindShortOptAllBranches(t *testing.T) {
	// Test OptionalArgument with args available
	parser1, _ := NewParser(ParserConfig{}, map[byte]*Flag{
		'o': {Name: "o", HasArg: OptionalArgument},
	}, map[string]*Flag{}, []string{})
	
	// Test optional argument taken from args (not word)
	_, _, option, err := parser1.findShortOpt('o', "", []string{"value", "extra"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !option.HasArg || option.Arg != "value" {
		t.Errorf("Expected optional arg 'value', got HasArg=%v, Arg='%s'", option.HasArg, option.Arg)
	}

	// Test case where option doesn't match in case-sensitive mode
	parser2, _ := NewParser(ParserConfig{shortCaseIgnore: false}, map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{})
	
	// This should continue through the loop and not find 'A' (case sensitive)
	_, _, _, err = parser2.findShortOpt('A', "", []string{})
	if err == nil {
		t.Error("Expected error for case mismatch in case-sensitive mode")
	}
}

// TestOptionsComplexIteratorFlow tests complex flow through Options iterator
func TestOptionsComplexIteratorFlow(t *testing.T) {
	// Test a complex scenario that exercises multiple branches
	parser, _ := NewParser(ParserConfig{}, map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
		'f': {Name: "f", HasArg: RequiredArgument},
	}, map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
		"file":    {Name: "file", HasArg: RequiredArgument},
	}, []string{"-vf", "filename", "--verbose", "--file=test.txt", "--", "remaining", "args"})
	
	expectedOptions := []struct {
		name   string
		hasArg bool
		arg    string
	}{
		{"v", false, ""},
		{"f", true, "filename"},
		{"verbose", false, ""},
		{"file", true, "test.txt"},
	}
	
	optionIndex := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}
		
		if optionIndex >= len(expectedOptions) {
			t.Errorf("Too many options processed")
			break
		}
		
		expected := expectedOptions[optionIndex]
		if option.Name != expected.name {
			t.Errorf("Option %d: expected name '%s', got '%s'", optionIndex, expected.name, option.Name)
		}
		if option.HasArg != expected.hasArg {
			t.Errorf("Option %d: expected HasArg %v, got %v", optionIndex, expected.hasArg, option.HasArg)
		}
		if option.Arg != expected.arg {
			t.Errorf("Option %d: expected arg '%s', got '%s'", optionIndex, expected.arg, option.Arg)
		}
		
		optionIndex++
	}
	
	if optionIndex != len(expectedOptions) {
		t.Errorf("Expected %d options, processed %d", len(expectedOptions), optionIndex)
	}
	
	// Check that remaining args are correct
	expectedRemaining := []string{"remaining", "args"}
	if len(parser.Args) != len(expectedRemaining) {
		t.Errorf("Expected %d remaining args, got %d", len(expectedRemaining), len(parser.Args))
	}
	for i, arg := range expectedRemaining {
		if i < len(parser.Args) && parser.Args[i] != arg {
			t.Errorf("Remaining arg %d: expected '%s', got '%s'", i, arg, parser.Args[i])
		}
	}
}
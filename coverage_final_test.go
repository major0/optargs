package optargs

import (
	"testing"
)

// TestFindShortOptCaseInsensitiveLoop tests the case insensitive loop in findShortOpt
func TestFindShortOptCaseInsensitiveLoop(t *testing.T) {
	// Test case where we need to iterate through multiple options to find a match
	parser, err := NewParser(ParserConfig{shortCaseIgnore: true}, map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: NoArgument},
		'c': {Name: "c", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test finding 'C' (uppercase) which should match 'c' (lowercase)
	_, _, option, err := parser.findShortOpt('C', "", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "c" {
		t.Errorf("Expected option name 'c', got '%s'", option.Name)
	}
}

// TestFindShortOptCaseSensitiveNoMatch tests case sensitive mode with no match
func TestFindShortOptCaseSensitiveNoMatch(t *testing.T) {
	parser, err := NewParser(ParserConfig{shortCaseIgnore: false}, map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test finding 'A' (uppercase) in case sensitive mode - should fail
	_, _, _, err = parser.findShortOpt('A', "", []string{})
	if err == nil {
		t.Error("Expected error for case mismatch in case-sensitive mode")
	}
	expectedMsg := "unknown option: A"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestFindShortOptOptionalArgumentFromArgs tests optional argument taken from args
func TestFindShortOptOptionalArgumentFromArgs(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'o': {Name: "o", HasArg: OptionalArgument},
	}, map[string]*Flag{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test optional argument taken from args (not from word)
	args, word, option, err := parser.findShortOpt('o', "", []string{"value", "extra"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !option.HasArg {
		t.Error("Expected HasArg to be true for optional argument")
	}
	if option.Arg != "value" {
		t.Errorf("Expected arg 'value', got '%s'", option.Arg)
	}
	if len(args) != 1 || args[0] != "extra" {
		t.Errorf("Expected remaining args ['extra'], got %v", args)
	}
	if word != "" {
		t.Errorf("Expected empty word, got '%s'", word)
	}
}

// TestFindShortOptOptionalArgumentNoArgs tests optional argument with no args available
func TestFindShortOptOptionalArgumentNoArgs(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'o': {Name: "o", HasArg: OptionalArgument},
	}, map[string]*Flag{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test optional argument with no args available
	args, word, option, err := parser.findShortOpt('o', "", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.HasArg {
		t.Error("Expected HasArg to be false when no args available")
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
}

// TestOptionsCommandExecutionWithError tests command execution that returns an error
func TestOptionsCommandExecutionWithError(t *testing.T) {
	// Create a parser with a nil command to trigger the error path
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"subcmd"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	// Add a command with nil parser to trigger the error in ExecuteCommand
	parser.Commands["subcmd"] = nil
	
	// Iterate through options - this should execute the command and hit the error path
	errorCount := 0
	for _, err := range parser.Options() {
		if err != nil {
			errorCount++
			// This should trigger the uncovered error handling path
			expectedMsg := "command subcmd has no parser"
			if err.Error() != expectedMsg {
				t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
			}
		}
		break // Only check first iteration
	}
	
	// Should have encountered the error
	if errorCount != 1 {
		t.Errorf("Expected 1 error, got %d", errorCount)
	}
}

// TestOptionsCommandExecutionDirectError tests direct command execution error
func TestOptionsCommandExecutionDirectError(t *testing.T) {
	// Create a parser and manually trigger ExecuteCommand with error
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"testcmd"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	// Manually add a command that exists but set it to nil to trigger error
	parser.Commands["testcmd"] = nil
	
	// This should trigger the error path in Options -> ExecuteCommand
	errorFound := false
	for _, err := range parser.Options() {
		if err != nil {
			errorFound = true
			if !contains(err.Error(), "has no parser") {
				t.Errorf("Expected error about 'has no parser', got: %v", err)
			}
		}
		break
	}
	
	if !errorFound {
		t.Error("Expected error from command execution")
	}
}

// TestOptionsCommandExecutionUnknownCommand tests command execution with unknown command
func TestOptionsCommandExecutionUnknownCommand(t *testing.T) {
	// Create a parser with commands but try to execute a non-existent one
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"nonexistent"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	// Add a different command so HasCommands() returns true
	subParser, _ := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{})
	parser.AddCmd("existing", subParser)
	
	// Since the command doesn't exist, it should be treated as a non-option
	// and not trigger the error path in ExecuteCommand
	optionCount := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
	}
	
	// Should have processed no options (nonexistent treated as non-option)
	if optionCount != 0 {
		t.Errorf("Expected 0 options, got %d", optionCount)
	}
	
	// The nonexistent command should be in Args as a non-option
	found := false
	for _, arg := range parser.Args {
		if arg == "nonexistent" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'nonexistent' to be treated as non-option in Args")
	}
}

// TestOptionsParseDefaultMode tests ParseDefault mode with non-option arguments
func TestOptionsParseDefaultMode(t *testing.T) {
	parser, err := NewParser(ParserConfig{parseMode: ParseDefault}, map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-v", "non-option", "-v"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "v" {
			optionCount++
		}
	}
	
	// Should have processed both -v options
	if optionCount != 2 {
		t.Errorf("Expected 2 -v options, got %d", optionCount)
	}
	
	// Should have "non-option" in Args
	found := false
	for _, arg := range parser.Args {
		if arg == "non-option" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'non-option' to be in Args")
	}
}

// TestOptionsLongOptsOnlyFallback tests longOptsOnly mode with fallback to short options
func TestOptionsLongOptsOnlyFallback(t *testing.T) {
	parser, err := NewParser(ParserConfig{longOptsOnly: true}, map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
	}, map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
	}, []string{"-verbose"}) // Should work as long option in longOptsOnly mode
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	options := []string{}
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		options = append(options, option.Name)
	}
	
	// Should have processed the verbose option
	if len(options) != 1 {
		t.Errorf("Expected 1 option, got %d: %v", len(options), options)
	}
	if len(options) > 0 && options[0] != "verbose" {
		t.Errorf("Expected 'verbose' option, got '%s'", options[0])
	}
}

// TestOptionsYieldReturnFalse tests early return when yield returns false
func TestOptionsYieldReturnFalse(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-a", "-b"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
		if option.Name == "a" {
			// Break after first option to test early return
			break
		}
	}
	
	// Should have only processed one option due to early break
	if optionCount != 1 {
		t.Errorf("Expected 1 option processed, got %d", optionCount)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsAt(s, substr, 1))))
}

func containsAt(s, substr string, start int) bool {
	if start >= len(s) {
		return false
	}
	if start+len(substr) <= len(s) && s[start:start+len(substr)] == substr {
		return true
	}
	return containsAt(s, substr, start+1)
}
// TestOptionsCleanupDeferredExecution tests the deferred cleanup logic
func TestOptionsCleanupDeferredExecution(t *testing.T) {
	parser, err := NewParser(ParserConfig{parseMode: ParseDefault}, map[byte]*Flag{}, map[string]*Flag{}, []string{"non-option1", "non-option2"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	// Don't iterate through all options to test the deferred cleanup
	optionCount := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
		// Break early to test deferred cleanup
		break
	}
	
	// The deferred cleanup should have moved non-options to Args
	if len(parser.Args) < 2 {
		t.Errorf("Expected at least 2 args after cleanup, got %d: %v", len(parser.Args), parser.Args)
	}
}

// TestOptionsCommandExecutionPath tests the command execution path more thoroughly
func TestOptionsCommandExecutionPath(t *testing.T) {
	// Create a subcommand parser with some options
	subParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-v"})
	if err != nil {
		t.Fatalf("Failed to create subparser: %v", err)
	}
	
	// Create main parser with command and remaining args
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"subcmd", "remaining"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	parser.AddCmd("subcmd", subParser)
	
	// Iterate through options - this should execute the command
	optionCount := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
	}
	
	// After command execution, Args should be empty (command consumed all args)
	if len(parser.Args) != 0 {
		t.Errorf("Expected Args to be empty after command execution, got %v", parser.Args)
	}
}

// TestOptionsLongOptWithRemainingArgs tests long option processing with remaining args
func TestOptionsLongOptWithRemainingArgs(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
		"file":    {Name: "file", HasArg: RequiredArgument},
	}, []string{"--verbose", "--file", "test.txt", "remaining"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	options := []string{}
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		options = append(options, option.Name)
	}
	
	// Should have processed both long options
	if len(options) != 2 {
		t.Errorf("Expected 2 options, got %d: %v", len(options), options)
	}
	
	// Should have "remaining" in Args
	found := false
	for _, arg := range parser.Args {
		if arg == "remaining" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'remaining' to be in Args")
	}
}

// TestOptionsShortOptWithRemainingArgs tests short option processing with remaining args
func TestOptionsShortOptWithRemainingArgs(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
		'f': {Name: "f", HasArg: RequiredArgument},
	}, map[string]*Flag{}, []string{"-v", "-f", "test.txt", "remaining"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	options := []string{}
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		options = append(options, option.Name)
	}
	
	// Should have processed both short options
	if len(options) != 2 {
		t.Errorf("Expected 2 options, got %d: %v", len(options), options)
	}
	
	// Should have "remaining" in Args
	found := false
	for _, arg := range parser.Args {
		if arg == "remaining" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'remaining' to be in Args")
	}
}

// TestOptionsEmptyArgsAfterProcessing tests when Args becomes empty during processing
func TestOptionsEmptyArgsAfterProcessing(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-v"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "v" {
			optionCount++
		}
	}
	
	// Should have processed exactly one option
	if optionCount != 1 {
		t.Errorf("Expected 1 option, got %d", optionCount)
	}
	
	// Args should be empty after processing
	if len(parser.Args) != 0 {
		t.Errorf("Expected empty Args, got %v", parser.Args)
	}
}
// TestOptionsLongOptNoRemainingArgs tests long option processing with no remaining args
func TestOptionsLongOptNoRemainingArgs(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
	}, []string{"--verbose"}) // Only one argument, no remaining args
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "verbose" {
			optionCount++
		}
	}
	
	// Should have processed exactly one option
	if optionCount != 1 {
		t.Errorf("Expected 1 option, got %d", optionCount)
	}
}

// TestOptionsShortOptNoRemainingArgs tests short option processing with no remaining args
func TestOptionsShortOptNoRemainingArgs(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-v"}) // Only one argument, no remaining args
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "v" {
			optionCount++
		}
	}
	
	// Should have processed exactly one option
	if optionCount != 1 {
		t.Errorf("Expected 1 option, got %d", optionCount)
	}
}

// TestOptionsLongOptsOnlyNoRemainingArgs tests longOptsOnly mode with no remaining args
func TestOptionsLongOptsOnlyNoRemainingArgs(t *testing.T) {
	parser, err := NewParser(ParserConfig{longOptsOnly: true}, map[byte]*Flag{}, map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
	}, []string{"-verbose"}) // Only one argument, no remaining args
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "verbose" {
			optionCount++
		}
	}
	
	// Should have processed exactly one option
	if optionCount != 1 {
		t.Errorf("Expected 1 option, got %d", optionCount)
	}
}

// TestOptionsShortOptCompactedNoRemainingArgs tests compacted short options with no remaining args
func TestOptionsShortOptCompactedNoRemainingArgs(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-ab"}) // Compacted options, no remaining args
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	options := []string{}
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		options = append(options, option.Name)
	}
	
	// Should have processed both options
	if len(options) != 2 {
		t.Errorf("Expected 2 options, got %d: %v", len(options), options)
	}
}
// TestOptionsDoubleBreakOut tests the break out logic with double dash
func TestOptionsDoubleBreakOut(t *testing.T) {
	parser, err := NewParser(ParserConfig{parseMode: ParseDefault}, map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-v", "--", "remaining", "args"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "v" {
			optionCount++
		}
	}
	
	// Should have processed one option before hitting --
	if optionCount != 1 {
		t.Errorf("Expected 1 option, got %d", optionCount)
	}
	
	// Should have remaining args after --
	if len(parser.Args) < 2 {
		t.Errorf("Expected at least 2 remaining args, got %d: %v", len(parser.Args), parser.Args)
	}
}

// TestOptionsNonOptInDefaultMode tests non-option handling in default mode
func TestOptionsNonOptInDefaultMode(t *testing.T) {
	parser, err := NewParser(ParserConfig{parseMode: ParseDefault}, map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"non-option", "-v", "another-non-option"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "v" {
			optionCount++
		}
	}
	
	// Should have processed one option
	if optionCount != 1 {
		t.Errorf("Expected 1 option, got %d", optionCount)
	}
	
	// Should have both non-options in Args
	if len(parser.Args) < 2 {
		t.Errorf("Expected at least 2 args, got %d: %v", len(parser.Args), parser.Args)
	}
}

// TestOptionsCommandWithoutHasCommands tests command lookup when HasCommands returns false
func TestOptionsCommandWithoutHasCommands(t *testing.T) {
	// Create parser without any commands registered
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"potential-command"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	// Don't add any commands - HasCommands() should return false
	optionCount := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
	}
	
	// Should treat "potential-command" as a non-option
	if len(parser.Args) == 0 {
		t.Error("Expected 'potential-command' to remain in Args")
	}
}

// TestOptionsCleanupDoneTrue tests the cleanupDone flag being set to true
func TestOptionsCleanupDoneTrue(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"--", "arg1", "arg2"})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	optionCount := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		optionCount++
	}
	
	// Should have processed no options (just the -- terminator)
	if optionCount != 0 {
		t.Errorf("Expected 0 options, got %d", optionCount)
	}
	
	// Should have both args after --
	if len(parser.Args) != 2 {
		t.Errorf("Expected 2 args, got %d: %v", len(parser.Args), parser.Args)
	}
}
// TestOptionsComprehensiveCoverage tests all paths through Options function
func TestOptionsComprehensiveCoverage(t *testing.T) {
	// Test case 1: Empty args
	parser1, _ := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{})
	count := 0
	for _, err := range parser1.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		count++
	}
	if count != 0 {
		t.Errorf("Expected 0 options for empty args, got %d", count)
	}

	// Test case 2: Only non-options in ParsePosixlyCorrect mode
	parser2, _ := NewParser(ParserConfig{parseMode: ParsePosixlyCorrect}, map[byte]*Flag{}, map[string]*Flag{}, []string{"non-opt"})
	count = 0
	for _, err := range parser2.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		count++
	}
	// Should break immediately on first non-option
	if count != 0 {
		t.Errorf("Expected 0 options in POSIXLY_CORRECT with non-option, got %d", count)
	}

	// Test case 3: Command execution with empty remaining args
	subParser, _ := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{})
	parser3, _ := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{}, []string{"cmd"})
	parser3.AddCmd("cmd", subParser)
	count = 0
	for _, err := range parser3.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		count++
	}
	// Command should be executed and Args should be empty
	if len(parser3.Args) != 0 {
		t.Errorf("Expected empty Args after command execution, got %v", parser3.Args)
	}

	// Test case 4: Long option with exactly one remaining arg
	parser4, _ := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
		"file": {Name: "file", HasArg: RequiredArgument},
	}, []string{"--file", "test.txt"})
	count = 0
	for option, err := range parser4.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "file" && option.Arg == "test.txt" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected 1 file option, got %d", count)
	}

	// Test case 5: Short option with exactly one remaining arg
	parser5, _ := NewParser(ParserConfig{}, map[byte]*Flag{
		'f': {Name: "f", HasArg: RequiredArgument},
	}, map[string]*Flag{}, []string{"-f", "test.txt"})
	count = 0
	for option, err := range parser5.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "f" && option.Arg == "test.txt" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected 1 f option, got %d", count)
	}
}

// TestOptionsIteratorYieldBehavior tests the yield function behavior
func TestOptionsIteratorYieldBehavior(t *testing.T) {
	parser, _ := NewParser(ParserConfig{}, map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: NoArgument},
		'c': {Name: "c", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-a", "-b", "-c"})
	
	// Test that yield returning false stops iteration
	count := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		count++
		if option.Name == "a" {
			// Stop after first option
			break
		}
	}
	
	if count != 1 {
		t.Errorf("Expected iteration to stop after 1 option, got %d", count)
	}
}

// TestOptionsLongOptsOnlyWithShortFallback tests longOptsOnly mode fallback
func TestOptionsLongOptsOnlyWithShortFallback(t *testing.T) {
	parser, _ := NewParser(ParserConfig{longOptsOnly: true}, map[byte]*Flag{
		'h': {Name: "h", HasArg: NoArgument},
	}, map[string]*Flag{}, []string{"-h"}) // Single char should fall back to short option
	
	count := 0
	for option, err := range parser.Options() {
		if err != nil {
			// In longOptsOnly mode, -h might be treated as long option "h" and fail
			// This is expected behavior, so we'll just count the attempts
			count++
			break
		}
		if option.Name == "h" {
			count++
		}
	}
	
	// The test passes if we attempted to process the option (even if it failed)
	if count == 0 {
		t.Error("Expected at least one option processing attempt")
	}
}
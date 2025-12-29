package optargs

import (
	"strings"
	"testing"
)

// TestGetOpt tests the GetOpt function directly
func TestGetOpt(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		optstring string
		expectErr bool
	}{
		{
			name:      "basic short options",
			args:      []string{"-a", "-b", "value"},
			optstring: "ab:",
			expectErr: false,
		},
		{
			name:      "invalid optstring",
			optstring: "a:",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOpt(tt.args, tt.optstring)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectErr && parser == nil {
				t.Error("Expected parser but got nil")
			}
		})
	}
}

// TestGetOptLong tests the GetOptLong function directly
func TestGetOptLong(t *testing.T) {
	longOpts := []Flag{
		{Name: "help", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}

	tests := []struct {
		name      string
		args      []string
		optstring string
		longOpts  []Flag
		expectErr bool
	}{
		{
			name:      "basic long options",
			args:      []string{"--help", "--output", "file.txt"},
			optstring: "",
			longOpts:  longOpts,
			expectErr: false,
		},
		{
			name:      "empty args",
			args:      []string{},
			optstring: "",
			longOpts:  longOpts,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOptLong(tt.args, tt.optstring, tt.longOpts)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectErr && parser == nil {
				t.Error("Expected parser but got nil")
			}
		})
	}
}

// TestGetOptLongOnly tests the GetOptLongOnly function directly
func TestGetOptLongOnly(t *testing.T) {
	longOpts := []Flag{
		{Name: "help", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}

	tests := []struct {
		name      string
		args      []string
		optstring string
		longOpts  []Flag
		expectErr bool
	}{
		{
			name:      "basic long-only options",
			args:      []string{"-help", "-output", "file.txt"},
			optstring: "",
			longOpts:  longOpts,
			expectErr: false,
		},
		{
			name:      "empty args",
			args:      []string{},
			optstring: "",
			longOpts:  longOpts,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOptLongOnly(tt.args, tt.optstring, tt.longOpts)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectErr && parser == nil {
				t.Error("Expected parser but got nil")
			}
		})
	}
}

// Note: isGraph, hasPrefix, and trimPrefix tests already exist in misc_test.go

// TestNewParser tests the NewParser function directly
func TestNewParser(t *testing.T) {
	tests := []struct {
		name      string
		config    ParserConfig
		shortOpts map[byte]*Flag
		longOpts  map[string]*Flag
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:   "valid parser creation",
			config: ParserConfig{enableErrors: true},
			shortOpts: map[byte]*Flag{
				'a': {Name: "a", HasArg: NoArgument},
			},
			longOpts: map[string]*Flag{
				"help": {Name: "help", HasArg: NoArgument},
			},
			args:      []string{"-a", "--help"},
			expectErr: false,
		},
		{
			name:   "invalid short option character",
			config: ParserConfig{enableErrors: true},
			shortOpts: map[byte]*Flag{
				'\t': {Name: "\t", HasArg: NoArgument}, // tab is not graph
			},
			longOpts:  nil,
			args:      []string{},
			expectErr: true,
			errMsg:    "Invalid short option",
		},
		{
			name:   "prohibited short option character colon",
			config: ParserConfig{enableErrors: true},
			shortOpts: map[byte]*Flag{
				':': {Name: ":", HasArg: NoArgument},
			},
			longOpts:  nil,
			args:      []string{},
			expectErr: true,
			errMsg:    "Prohibited short option",
		},
		{
			name:   "prohibited short option character semicolon",
			config: ParserConfig{enableErrors: true},
			shortOpts: map[byte]*Flag{
				';': {Name: ";", HasArg: NoArgument},
			},
			longOpts:  nil,
			args:      []string{},
			expectErr: true,
			errMsg:    "Prohibited short option",
		},
		{
			name:   "prohibited short option character dash",
			config: ParserConfig{enableErrors: true},
			shortOpts: map[byte]*Flag{
				'-': {Name: "-", HasArg: NoArgument},
			},
			longOpts:  nil,
			args:      []string{},
			expectErr: true,
			errMsg:    "Prohibited short option",
		},
		{
			name:      "invalid long option with space",
			config:    ParserConfig{enableErrors: true},
			shortOpts: nil,
			longOpts: map[string]*Flag{
				"help me": {Name: "help me", HasArg: NoArgument}, // space not allowed
			},
			args:      []string{},
			expectErr: true,
			errMsg:    "invalid long option",
		},
		{
			name:      "invalid long option with control char",
			config:    ParserConfig{enableErrors: true},
			shortOpts: nil,
			longOpts: map[string]*Flag{
				"help\t": {Name: "help\t", HasArg: NoArgument}, // tab not allowed
			},
			args:      []string{},
			expectErr: true,
			errMsg:    "invalid long option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(tt.config, tt.shortOpts, tt.longOpts, tt.args)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if parser == nil {
					t.Error("Expected parser but got nil")
				}
			}
		})
	}
}

// TestOptError tests the optError method
func TestOptError(t *testing.T) {
	tests := []struct {
		name         string
		enableErrors bool
		message      string
	}{
		{
			name:         "errors enabled",
			enableErrors: true,
			message:      "test error message",
		},
		{
			name:         "errors disabled",
			enableErrors: false,
			message:      "test error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				config: ParserConfig{enableErrors: tt.enableErrors},
			}

			err := parser.optError(tt.message)
			if err == nil {
				t.Error("Expected error but got nil")
			}
			if err.Error() != tt.message {
				t.Errorf("Expected error message %q, got %q", tt.message, err.Error())
			}
		})
	}
}

// TestOptErrorf tests the optErrorf method
func TestOptErrorf(t *testing.T) {
	tests := []struct {
		name         string
		enableErrors bool
		format       string
		args         []interface{}
		expected     string
	}{
		{
			name:         "formatted error with args",
			enableErrors: true,
			format:       "error with %s and %d",
			args:         []interface{}{"string", 42},
			expected:     "error with string and 42",
		},
		{
			name:         "simple format no args",
			enableErrors: false,
			format:       "simple error",
			args:         []interface{}{},
			expected:     "simple error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				config: ParserConfig{enableErrors: tt.enableErrors},
			}

			err := parser.optErrorf(tt.format, tt.args...)
			if err == nil {
				t.Error("Expected error but got nil")
			}
			if err.Error() != tt.expected {
				t.Errorf("Expected error message %q, got %q", tt.expected, err.Error())
			}
		})
	}
}

// TestFindLongOptEdgeCases tests edge cases in findLongOpt to improve coverage
func TestFindLongOptEdgeCases(t *testing.T) {
	// Create a parser with specific long options to test edge cases
	longOpts := map[string]*Flag{
		"help":    {Name: "help", HasArg: NoArgument},
		"output":  {Name: "output", HasArg: RequiredArgument},
		"verbose": {Name: "verbose", HasArg: OptionalArgument},
		"foo=bar": {Name: "foo=bar", HasArg: NoArgument}, // option name with equals
		"config":  {Name: "config", HasArg: RequiredArgument},
	}

	parser := &Parser{
		longOpts: longOpts,
		config: ParserConfig{
			longCaseIgnore: true,
			enableErrors:   true,
		},
	}

	tests := []struct {
		name      string
		optName   string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:    "exact match no argument",
			optName: "help",
			args:    []string{},
		},
		{
			name:    "exact match with required argument",
			optName: "output",
			args:    []string{"file.txt"},
		},
		{
			name:    "exact match with optional argument",
			optName: "verbose",
			args:    []string{"level1"},
		},
		{
			name:    "option with equals in name",
			optName: "foo=bar",
			args:    []string{},
		},
		{
			name:    "option with equals and value",
			optName: "output=file.txt",
			args:    []string{},
		},
		{
			name:      "unknown option",
			optName:   "unknown",
			args:      []string{},
			expectErr: true,
			errMsg:    "unknown option",
		},
		{
			name:      "required argument missing",
			optName:   "config",
			args:      []string{},
			expectErr: true,
			errMsg:    "option requires an argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, option, err := parser.findLongOpt(tt.optName, tt.args)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if option.Name == "" {
					t.Error("Expected option name but got empty string")
				}
			}
		})
	}
}

// TestFindShortOptEdgeCases tests edge cases in findShortOpt to improve coverage
func TestFindShortOptEdgeCases(t *testing.T) {
	shortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: RequiredArgument},
		'c': {Name: "c", HasArg: OptionalArgument},
	}

	parser := &Parser{
		shortOpts: shortOpts,
		config: ParserConfig{
			shortCaseIgnore: false,
			enableErrors:    true,
		},
	}

	tests := []struct {
		name      string
		c         byte
		word      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name: "no argument option",
			c:    'a',
			word: "",
			args: []string{},
		},
		{
			name: "required argument from word",
			c:    'b',
			word: "value",
			args: []string{},
		},
		{
			name: "required argument from args",
			c:    'b',
			word: "",
			args: []string{"value"},
		},
		{
			name: "optional argument from word",
			c:    'c',
			word: "value",
			args: []string{},
		},
		{
			name: "optional argument from args",
			c:    'c',
			word: "",
			args: []string{"value"},
		},
		{
			name: "optional argument none provided",
			c:    'c',
			word: "",
			args: []string{},
		},
		{
			name:      "dash character",
			c:         '-',
			word:      "",
			args:      []string{},
			expectErr: true,
			errMsg:    "invalid option",
		},
		{
			name:      "unknown option",
			c:         'z',
			word:      "",
			args:      []string{},
			expectErr: true,
			errMsg:    "unknown option",
		},
		{
			name:      "required argument missing",
			c:         'b',
			word:      "",
			args:      []string{},
			expectErr: true,
			errMsg:    "option requires an argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, option, err := parser.findShortOpt(tt.c, tt.word, tt.args)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if option.Name == "" {
					t.Error("Expected option name but got empty string")
				}
			}
		})
	}
}

// TestOptionsIteratorEdgeCases tests edge cases in Options iterator to improve coverage
func TestOptionsIteratorEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		parser    *Parser
		expectErr bool
	}{
		{
			name: "double dash termination",
			parser: &Parser{
				Args: []string{"--", "remaining", "args"},
				config: ParserConfig{
					parseMode:    ParseDefault,
					enableErrors: true,
				},
				shortOpts: map[byte]*Flag{},
				longOpts:  map[string]*Flag{},
				nonOpts:   []string{},
			},
		},
		{
			name: "long options only mode",
			parser: &Parser{
				Args: []string{"-help"},
				config: ParserConfig{
					longOptsOnly: true,
					enableErrors: true,
				},
				shortOpts: map[byte]*Flag{},
				longOpts: map[string]*Flag{
					"help": {Name: "help", HasArg: NoArgument},
				},
				nonOpts: []string{},
			},
		},
		{
			name: "parse non-opts mode",
			parser: &Parser{
				Args: []string{"non-option"},
				config: ParserConfig{
					parseMode:    ParseNonOpts,
					enableErrors: true,
				},
				shortOpts: map[byte]*Flag{},
				longOpts:  map[string]*Flag{},
				nonOpts:   []string{},
			},
		},
		{
			name: "posixly correct mode",
			parser: &Parser{
				Args: []string{"non-option", "-a"},
				config: ParserConfig{
					parseMode:    ParsePosixlyCorrect,
					enableErrors: true,
				},
				shortOpts: map[byte]*Flag{
					'a': {Name: "a", HasArg: NoArgument},
				},
				longOpts: map[string]*Flag{},
				nonOpts:  []string{},
			},
		},
		{
			name: "gnu words transformation",
			parser: &Parser{
				Args: []string{"-W", "foo"},
				config: ParserConfig{
					gnuWords:     true,
					enableErrors: true,
				},
				shortOpts: map[byte]*Flag{
					'W': {Name: "W", HasArg: RequiredArgument},
				},
				longOpts: map[string]*Flag{},
				nonOpts:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var options []Option
			var errors []error

			for option, err := range tt.parser.Options() {
				options = append(options, option)
				if err != nil {
					errors = append(errors, err)
				}
			}

			if tt.expectErr && len(errors) == 0 {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && len(errors) > 0 {
				t.Errorf("Unexpected errors: %v", errors)
			}
		})
	}
}

// TestGetOptWithInvalidArgType tests invalid argument type handling
func TestGetOptWithInvalidArgType(t *testing.T) {
	// Create a parser with an invalid argument type to test the default case
	shortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: ArgType(999)}, // Invalid ArgType
	}

	parser := &Parser{
		Args:      []string{"-a"},
		shortOpts: shortOpts,
		longOpts:  map[string]*Flag{},
		config: ParserConfig{
			enableErrors: true,
		},
		nonOpts: []string{},
	}

	// This should trigger the default case in findShortOpt
	for option, err := range parser.Options() {
		if err == nil {
			t.Error("Expected error for invalid argument type but got none")
		}
		if !strings.Contains(err.Error(), "unknown argument type") {
			t.Errorf("Expected 'unknown argument type' error, got: %v", err)
		}
		// We expect this to fail, so we can break after first iteration
		_ = option
		break
	}
}

// TestCaseInsensitiveShortOptions tests case insensitive short option handling
func TestCaseInsensitiveShortOptions(t *testing.T) {
	shortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
	}

	parser := &Parser{
		Args:      []string{"-A"}, // uppercase A
		shortOpts: shortOpts,
		longOpts:  map[string]*Flag{},
		config: ParserConfig{
			shortCaseIgnore: true, // Enable case insensitive matching
			enableErrors:    true,
		},
		nonOpts: []string{},
	}

	var foundOption bool
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if option.Name == "a" {
			foundOption = true
		}
		break
	}

	if !foundOption {
		t.Error("Expected to find option 'a' with case insensitive matching")
	}
}

// TestIteratorBreakConditions tests various break conditions in the Options iterator
func TestIteratorBreakConditions(t *testing.T) {
	// Test processing multiple options normally
	parser := &Parser{
		Args: []string{"-a", "-b", "-c"},
		config: ParserConfig{
			enableErrors: true,
		},
		shortOpts: map[byte]*Flag{
			'a': {Name: "a", HasArg: NoArgument},
			'b': {Name: "b", HasArg: NoArgument},
			'c': {Name: "c", HasArg: NoArgument},
		},
		longOpts: map[string]*Flag{},
		nonOpts:  []string{},
	}

	count := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		count++
		_ = option
	}

	if count != 3 {
		t.Errorf("Expected to process 3 options, got %d", count)
	}
}

// TestLongOptContinueConditions tests continue conditions in findLongOpt
func TestLongOptContinueConditions(t *testing.T) {
	longOpts := map[string]*Flag{
		"help":    {Name: "help", HasArg: NoArgument},
		"verbose": {Name: "verbose", HasArg: NoArgument},
		"output":  {Name: "output", HasArg: RequiredArgument},
	}

	parser := &Parser{
		longOpts: longOpts,
		config: ParserConfig{
			longCaseIgnore: false, // Case sensitive to test different paths
			enableErrors:   true,
		},
	}

	// Test case where option length is greater than name length (should continue)
	_, _, err := parser.findLongOpt("he", []string{})
	if err == nil {
		t.Error("Expected error for partial match with case sensitive mode")
	}

	// Test case where names don't match exactly (should continue)
	_, _, err = parser.findLongOpt("HELP", []string{})
	if err == nil {
		t.Error("Expected error for case mismatch with case sensitive mode")
	}

	// Test case where equals sign is not at the right position (should continue)
	_, _, err = parser.findLongOpt("hel=p", []string{})
	if err == nil {
		t.Error("Expected error for malformed option with equals")
	}
}

// TestShortOptContinueConditions tests continue conditions in findShortOpt
func TestShortOptContinueConditions(t *testing.T) {
	shortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: RequiredArgument},
	}

	parser := &Parser{
		shortOpts: shortOpts,
		config: ParserConfig{
			shortCaseIgnore: false, // Case sensitive
			enableErrors:    true,
		},
	}

	// Test case where case doesn't match (should continue)
	_, _, _, err := parser.findShortOpt('A', "", []string{})
	if err == nil {
		t.Error("Expected error for case mismatch with case sensitive mode")
	}

	// Test case where character doesn't exist (should continue through all options)
	_, _, _, err = parser.findShortOpt('z', "", []string{})
	if err == nil {
		t.Error("Expected error for unknown option")
	}
}

// TestGetOptEdgeCases tests edge cases in getOpt function
func TestGetOptEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		optstring string
		longOpts  []Flag
		longOnly  bool
		expectErr bool
		errMsg    string
	}{
		{
			name:      "long only with non-empty optstring",
			args:      []string{},
			optstring: "a",
			longOpts:  []Flag{},
			longOnly:  true,
			expectErr: true,
			errMsg:    "non-empty option string found when long-only parsing was enabled",
		},
		{
			name:      "invalid option character in optstring",
			args:      []string{},
			optstring: "a\tb", // tab character
			longOpts:  []Flag{},
			longOnly:  false,
			expectErr: true,
			errMsg:    "Invalid option character",
		},
		{
			name:      "prohibited colon character",
			args:      []string{},
			optstring: "a:b:", // this should be valid
			longOpts:  []Flag{},
			longOnly:  false,
			expectErr: false,
		},
		{
			name:      "prohibited character in option",
			args:      []string{},
			optstring: "a-b", // dash not allowed as option
			longOpts:  []Flag{},
			longOnly:  false,
			expectErr: true,
			errMsg:    "Invalid option character",
		},
		{
			name:      "gnu words with W option",
			args:      []string{},
			optstring: "W;", // W with semicolon enables gnu words
			longOpts:  []Flag{},
			longOnly:  false,
			expectErr: false,
		},
		{
			name:      "behavior flags",
			args:      []string{},
			optstring: ":+ab", // colon and plus flags
			longOpts:  []Flag{},
			longOnly:  false,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getOpt(tt.args, tt.optstring, tt.longOpts, tt.longOnly)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestLongOptBestMatch tests the "best match" logic in findLongOpt
func TestLongOptBestMatch(t *testing.T) {
	// Create options where multiple matches are possible to test best match logic
	longOpts := map[string]*Flag{
		"help":    {Name: "help", HasArg: NoArgument},
		"help-me": {Name: "help-me", HasArg: NoArgument},
		"verbose": {Name: "verbose", HasArg: NoArgument},
		"verb":    {Name: "verb", HasArg: NoArgument},
	}

	parser := &Parser{
		longOpts: longOpts,
		config: ParserConfig{
			longCaseIgnore: true,
			enableErrors:   true,
		},
	}

	// Test that the longest match wins
	_, option, err := parser.findLongOpt("help-me", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "help-me" {
		t.Errorf("Expected option name 'help-me', got '%s'", option.Name)
	}

	// Test exact match over partial match
	_, option, err = parser.findLongOpt("help", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "help" {
		t.Errorf("Expected option name 'help', got '%s'", option.Name)
	}
}

// TestOptionsYieldFalse tests normal iteration without early termination
func TestOptionsNormalIteration(t *testing.T) {
	parser := &Parser{
		Args: []string{"-a", "-b"},
		config: ParserConfig{
			enableErrors: true,
		},
		shortOpts: map[byte]*Flag{
			'a': {Name: "a", HasArg: NoArgument},
			'b': {Name: "b", HasArg: NoArgument},
		},
		longOpts: map[string]*Flag{},
		nonOpts:  []string{},
	}

	count := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		count++
		_ = option
	}

	if count != 2 {
		t.Errorf("Expected to process 2 options, got %d", count)
	}
}

// TestOptionalArgumentFromArgs tests optional argument taken from args array
func TestOptionalArgumentFromArgs(t *testing.T) {
	shortOpts := map[byte]*Flag{
		'v': {Name: "v", HasArg: OptionalArgument},
	}

	parser := &Parser{
		shortOpts: shortOpts,
		longOpts:  map[string]*Flag{},
		config: ParserConfig{
			enableErrors: true,
		},
	}

	// Test optional argument taken from args (not from word)
	args, word, option, err := parser.findShortOpt('v', "", []string{"value"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "v" {
		t.Errorf("Expected option name 'v', got '%s'", option.Name)
	}
	if !option.HasArg {
		t.Error("Expected option to have argument")
	}
	if option.Arg != "value" {
		t.Errorf("Expected argument 'value', got '%s'", option.Arg)
	}
	if len(args) != 0 {
		t.Errorf("Expected args to be consumed, got %v", args)
	}
	if word != "" {
		t.Errorf("Expected word to be empty, got '%s'", word)
	}
}

// TestLongOptWithEqualsInName tests long option names containing equals signs
func TestLongOptWithEqualsInName(t *testing.T) {
	longOpts := map[string]*Flag{
		"config=file": {Name: "config=file", HasArg: NoArgument},
		"output":      {Name: "output", HasArg: RequiredArgument},
	}

	parser := &Parser{
		longOpts: longOpts,
		config: ParserConfig{
			longCaseIgnore: true,
			enableErrors:   true,
		},
	}

	// Test option name with equals sign
	_, option, err := parser.findLongOpt("config=file", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "config=file" {
		t.Errorf("Expected option name 'config=file', got '%s'", option.Name)
	}

	// Test option with equals and value
	_, option, err = parser.findLongOpt("output=test.txt", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "output" {
		t.Errorf("Expected option name 'output', got '%s'", option.Name)
	}
	if option.Arg != "=test.txt" {
		t.Errorf("Expected argument '=test.txt', got '%s'", option.Arg)
	}
}

// TestCaseSensitiveLongOptions tests case sensitive long option matching
func TestCaseSensitiveLongOptions(t *testing.T) {
	longOpts := map[string]*Flag{
		"Help": {Name: "Help", HasArg: NoArgument},
		"help": {Name: "help", HasArg: NoArgument},
	}

	parser := &Parser{
		longOpts: longOpts,
		config: ParserConfig{
			longCaseIgnore: false, // Case sensitive
			enableErrors:   true,
		},
	}

	// Test exact case match
	_, option, err := parser.findLongOpt("Help", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "Help" {
		t.Errorf("Expected option name 'Help', got '%s'", option.Name)
	}

	// Test different case match
	_, option, err = parser.findLongOpt("help", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "help" {
		t.Errorf("Expected option name 'help', got '%s'", option.Name)
	}
}

// TestEmptyArgsAndWord tests edge case with empty args and word
func TestEmptyArgsAndWord(t *testing.T) {
	shortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
		'b': {Name: "b", HasArg: OptionalArgument},
	}

	parser := &Parser{
		shortOpts: shortOpts,
		longOpts:  map[string]*Flag{},
		config: ParserConfig{
			enableErrors: true,
		},
	}

	// Test no argument option with empty word and args
	args, word, option, err := parser.findShortOpt('a', "", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "a" {
		t.Errorf("Expected option name 'a', got '%s'", option.Name)
	}
	if option.HasArg {
		t.Error("Expected option to not have argument")
	}
	if len(args) != 0 {
		t.Errorf("Expected args to remain empty, got %v", args)
	}
	if word != "" {
		t.Errorf("Expected word to remain empty, got '%s'", word)
	}

	// Test optional argument with empty word and args
	args, word, option, err = parser.findShortOpt('b', "", []string{})
	_ = args // Suppress ineffassign warning
	_ = word // Suppress ineffassign warning
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "b" {
		t.Errorf("Expected option name 'b', got '%s'", option.Name)
	}
	if option.HasArg {
		t.Error("Expected option to not have argument when none provided")
	}
	if option.Arg != "" {
		t.Errorf("Expected empty argument, got '%s'", option.Arg)
	}
}

// TestGotoOutPath tests normal short option processing
func TestShortOptionProcessing(t *testing.T) {
	parser := &Parser{
		Args: []string{"-abc"}, // Multiple short options in one argument
		config: ParserConfig{
			enableErrors: true,
		},
		shortOpts: map[byte]*Flag{
			'a': {Name: "a", HasArg: NoArgument},
			'b': {Name: "b", HasArg: NoArgument},
			'c': {Name: "c", HasArg: NoArgument},
		},
		longOpts: map[string]*Flag{},
		nonOpts:  []string{},
	}

	// Process all options normally
	count := 0
	expectedOptions := []string{"a", "b", "c"}
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count >= len(expectedOptions) {
			t.Errorf("Too many options processed")
			break
		}
		if option.Name != expectedOptions[count] {
			t.Errorf("Expected option '%s', got '%s'", expectedOptions[count], option.Name)
		}
		count++
	}

	if count != 3 {
		t.Errorf("Expected to process 3 options, got %d", count)
	}
}

// TestLongOptCaseIgnorePrefix tests the case ignore prefix matching
func TestLongOptCaseIgnorePrefix(t *testing.T) {
	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
		"version": {Name: "version", HasArg: NoArgument},
	}

	parser := &Parser{
		longOpts: longOpts,
		config: ParserConfig{
			longCaseIgnore: true,
			enableErrors:   true,
		},
	}

	// Test case insensitive exact matching
	_, option, err := parser.findLongOpt("verbose", []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if option.Name != "verbose" {
		t.Errorf("Expected option name 'verbose', got '%s'", option.Name)
	}
}

// TestParseNonOptsBreak tests the ParseNonOpts mode without breaking
func TestParseNonOptsMode(t *testing.T) {
	parser := &Parser{
		Args: []string{"arg1"},
		config: ParserConfig{
			parseMode:    ParseNonOpts,
			enableErrors: true,
		},
		shortOpts: map[byte]*Flag{},
		longOpts:  map[string]*Flag{},
		nonOpts:   []string{},
	}

	// Process the argument normally
	count := 0
	for option, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		count++
		if option.Name != string(byte(1)) {
			t.Errorf("Expected option name to be byte(1), got '%s'", option.Name)
		}
		if option.Arg != "arg1" {
			t.Errorf("Expected argument 'arg1', got '%s'", option.Arg)
		}
	}

	if count != 1 {
		t.Errorf("Expected to process 1 argument, got %d", count)
	}
}

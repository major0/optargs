package goarg

import (
	"bytes"
	"strings"
	"testing"
)

// TestHelpTextFormatting tests that help text formatting matches alexflint/go-arg exactly
func TestHelpTextFormatting(t *testing.T) {
	testCases := []struct {
		name           string
		testStruct     interface{}
		config         Config
		expectedParts  []string // Parts that should be in the help text
		forbiddenParts []string // Parts that should NOT be in the help text
	}{
		{
			name: "basic options formatting",
			testStruct: &struct {
				Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
				Count   int    `arg:"-c,--count" help:"number of items"`
				Input   string `arg:"--input" help:"input file path"`
			}{},
			config: Config{
				Program:     "testapp",
				Description: "A test application",
				Version:     "1.0.0",
			},
			expectedParts: []string{
				"Usage: testapp [OPTIONS]",
				"A test application",
				"Options:",
				"-v, --verbose",
				"enable verbose output",
				"-c, --count COUNT",
				"number of items",
				"--input INPUT",
				"input file path",
				"-h, --help",
				"show this help message and exit",
				"Version: 1.0.0",
			},
		},
		{
			name: "positional arguments formatting",
			testStruct: &struct {
				Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
				Source  string `arg:"positional,required" help:"source file"`
				Dest    string `arg:"positional" help:"destination file"`
			}{},
			config: Config{Program: "copy"},
			expectedParts: []string{
				"Usage: copy [OPTIONS] SOURCE [DEST]",
				"Positional arguments:",
				"SOURCE",
				"source file",
				"DEST",
				"destination file",
				"Options:",
				"-v, --verbose",
			},
		},
		{
			name: "subcommands formatting",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose" help:"enable verbose output"`
				Server  *struct {
					Port int `arg:"-p,--port" help:"server port"`
				} `arg:"subcommand:server" help:"run server"`
				Client *struct {
					URL string `arg:"-u,--url" help:"client URL"`
				} `arg:"subcommand:client" help:"run client"`
			}{},
			config: Config{
				Program:     "myapp",
				Description: "Multi-command application",
			},
			expectedParts: []string{
				"Usage: myapp COMMAND [OPTIONS]",
				"Multi-command application",
				"Options:",
				"-v, --verbose",
				"Commands:",
				"server",
				"run server",
				"client",
				"run client",
			},
		},
		{
			name: "minimal configuration",
			testStruct: &struct {
				Flag bool `arg:"--flag"`
			}{},
			config: Config{}, // Empty config
			expectedParts: []string{
				"Usage:",
				"[OPTIONS]",
				"Options:",
				"--flag",
				"-h, --help",
			},
		},
		{
			name:       "no options struct",
			testStruct: &struct{}{},
			config:     Config{Program: "empty"},
			expectedParts: []string{
				"Usage: empty",
			},
			forbiddenParts: []string{
				"Options:",
				"[OPTIONS]",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewParser(tc.config, tc.testStruct)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			var buf bytes.Buffer
			parser.WriteHelp(&buf)
			helpText := buf.String()

			t.Logf("Generated help text:\n%s", helpText)

			// Check expected parts
			for _, part := range tc.expectedParts {
				if !strings.Contains(helpText, part) {
					t.Errorf("Help text should contain %q", part)
				}
			}

			// Check forbidden parts
			for _, part := range tc.forbiddenParts {
				if strings.Contains(helpText, part) {
					t.Errorf("Help text should NOT contain %q", part)
				}
			}

			// Validate general formatting
			lines := strings.Split(helpText, "\n")
			if len(lines) == 0 {
				t.Error("Help text should not be empty")
			}

			// First line should be usage
			if !strings.HasPrefix(lines[0], "Usage: ") {
				t.Errorf("First line should start with 'Usage: ', got: %q", lines[0])
			}
		})
	}
}

// TestUsageStringGeneration tests usage string generation
func TestUsageStringGeneration(t *testing.T) {
	testCases := []struct {
		name       string
		testStruct interface{}
		config     Config
		expected   string
	}{
		{
			name: "basic options",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
				Count   int  `arg:"-c,--count"`
			}{},
			config:   Config{Program: "test"},
			expected: "Usage: test [OPTIONS]\n",
		},
		{
			name: "positional arguments",
			testStruct: &struct {
				Source string `arg:"positional,required"`
				Dest   string `arg:"positional"`
			}{},
			config:   Config{Program: "copy"},
			expected: "Usage: copy SOURCE [DEST]\n",
		},
		{
			name: "mixed options and positionals",
			testStruct: &struct {
				Verbose bool   `arg:"-v,--verbose"`
				Source  string `arg:"positional,required"`
				Dest    string `arg:"positional"`
			}{},
			config:   Config{Program: "app"},
			expected: "Usage: app [OPTIONS] SOURCE [DEST]\n",
		},
		{
			name: "subcommands",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
				Server  *struct {
					Port int `arg:"-p,--port"`
				} `arg:"subcommand:server"`
			}{},
			config:   Config{Program: "myapp"},
			expected: "Usage: myapp COMMAND [OPTIONS]\n",
		},
		{
			name:       "empty struct",
			testStruct: &struct{}{},
			config:     Config{Program: "empty"},
			expected:   "Usage: empty\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewParser(tc.config, tc.testStruct)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			var buf bytes.Buffer
			parser.WriteUsage(&buf)
			usageText := buf.String()

			if usageText != tc.expected {
				t.Errorf("Expected usage %q, got %q", tc.expected, usageText)
			}
		})
	}
}

// TestErrorMessageFormatAndContent tests error message format and content
func TestErrorMessageFormatAndContent(t *testing.T) {
	testCases := []struct {
		name               string
		testStruct         interface{}
		args               []string
		expectedError      string
		errorShouldContain []string
	}{
		{
			name: "unknown long option",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
			}{},
			args:          []string{"--unknown"},
			expectedError: "unrecognized argument: --unknown",
		},
		{
			name: "unknown short option",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:          []string{"-x"},
			expectedError: "unrecognized argument: -x",
		},
		{
			name: "option requires argument",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:               []string{"--count"},
			errorShouldContain: []string{"option requires an argument", "--count"},
		},
		{
			name: "missing required option",
			testStruct: &struct {
				Input string `arg:"--input,required"`
			}{},
			args:               []string{},
			errorShouldContain: []string{"required argument missing"},
		},
		{
			name: "invalid type conversion",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:               []string{"--count", "invalid"},
			errorShouldContain: []string{"invalid argument"},
		},
		{
			name: "missing required positional",
			testStruct: &struct {
				Source string `arg:"positional,required"`
			}{},
			args:               []string{},
			errorShouldContain: []string{"required", "Source"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewParser(Config{}, tc.testStruct)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			err = parser.Parse(tc.args)
			if err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			errMsg := err.Error()
			t.Logf("Got error: %v", err)

			if tc.expectedError != "" {
				if errMsg != tc.expectedError {
					t.Errorf("Expected exact error %q, got %q", tc.expectedError, errMsg)
				}
			}

			for _, shouldContain := range tc.errorShouldContain {
				if !strings.Contains(errMsg, shouldContain) {
					t.Errorf("Error should contain %q, got %q", shouldContain, errMsg)
				}
			}

			// Validate error message properties
			if strings.TrimSpace(errMsg) == "" {
				t.Error("Error message should not be empty")
			}

			// Should not contain internal implementation details
			forbiddenPhrases := []string{
				"parsing error:",
				"failed to set field",
				"OptArgs Core",
			}
			for _, phrase := range forbiddenPhrases {
				if strings.Contains(errMsg, phrase) {
					t.Errorf("Error message should not contain internal phrase %q", phrase)
				}
			}
		})
	}
}

// TestSubcommandHelpGeneration tests subcommand help generation
func TestSubcommandHelpGeneration(t *testing.T) {
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type ClientCmd struct {
		URL     string `arg:"-u,--url" help:"client URL"`
		Timeout int    `arg:"-t,--timeout" default:"30" help:"timeout in seconds"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Debug   bool       `arg:"-d,--debug" help:"enable debug output"`
		Server  *ServerCmd `arg:"subcommand:server" help:"run server mode"`
		Client  *ClientCmd `arg:"subcommand:client" help:"run client mode"`
	}

	config := Config{
		Program:     "myapp",
		Description: "A multi-command application",
		Version:     "2.0.0",
	}

	parser, err := NewParser(config, &RootCmd{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	var buf bytes.Buffer
	parser.WriteHelp(&buf)
	helpText := buf.String()

	t.Logf("Subcommand help text:\n%s", helpText)

	// Test expected components for subcommand help
	expectedParts := []string{
		"Usage: myapp COMMAND [OPTIONS]",
		"A multi-command application",
		"Options:",
		"-v, --verbose",
		"enable verbose output",
		"-d, --debug",
		"enable debug output",
		"-h, --help",
		"Commands:",
		"server",
		"run server mode",
		"client",
		"run client mode",
		"Version: 2.0.0",
	}

	for _, part := range expectedParts {
		if !strings.Contains(helpText, part) {
			t.Errorf("Subcommand help should contain %q", part)
		}
	}

	// Test that COMMAND appears in usage when subcommands are present
	if !strings.Contains(helpText, "COMMAND") {
		t.Error("Usage should contain COMMAND when subcommands are present")
	}

	// Test that Commands section is present
	if !strings.Contains(helpText, "Commands:") {
		t.Error("Help should contain Commands section when subcommands are present")
	}

	// Test proper alignment and formatting
	lines := strings.Split(helpText, "\n")
	var commandsSection bool
	for _, line := range lines {
		if strings.Contains(line, "Commands:") {
			commandsSection = true
			continue
		}
		if commandsSection && strings.TrimSpace(line) == "" {
			break // End of commands section
		}
		if commandsSection && strings.TrimSpace(line) != "" {
			// Command lines should be properly indented
			if !strings.HasPrefix(line, "  ") {
				t.Errorf("Command line should be indented: %q", line)
			}
		}
	}
}

// TestHelpGeneratorDirectly tests the HelpGenerator directly
func TestHelpGeneratorDirectly(t *testing.T) {
	type TestStruct struct {
		Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int    `arg:"-c,--count" help:"number of items"`
		Input   string `arg:"--input" help:"input file"`
	}

	parser, err := NewParser(Config{}, &TestStruct{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test HelpGenerator directly
	config := Config{
		Program:     "testapp",
		Description: "Test application",
		Version:     "1.0.0",
	}

	helpGenerator := NewHelpGenerator(parser.metadata, config)

	// Test WriteHelp
	var helpBuf bytes.Buffer
	err = helpGenerator.WriteHelp(&helpBuf)
	if err != nil {
		t.Errorf("WriteHelp should not return error: %v", err)
	}

	helpText := helpBuf.String()
	if !strings.Contains(helpText, "testapp") {
		t.Error("Help text should contain program name")
	}

	// Test WriteUsage
	var usageBuf bytes.Buffer
	err = helpGenerator.WriteUsage(&usageBuf)
	if err != nil {
		t.Errorf("WriteUsage should not return error: %v", err)
	}

	usageText := usageBuf.String()
	if !strings.HasPrefix(usageText, "Usage: testapp") {
		t.Errorf("Usage should start with program name, got: %q", usageText)
	}
}

// TestHelpWithDefaults tests help generation with default values
func TestHelpWithDefaults(t *testing.T) {
	type TestStruct struct {
		Port    int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host    string `arg:"-h,--host" default:"localhost" help:"server host"`
		Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
	}

	config := Config{Program: "server"}
	parser, err := NewParser(config, &TestStruct{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	var buf bytes.Buffer
	parser.WriteHelp(&buf)
	helpText := buf.String()

	t.Logf("Help with defaults:\n%s", helpText)

	// Check that default values are shown
	expectedParts := []string{
		"-p, --port PORT",
		"server port",
		"(default: 8080)",
		"-h, --host HOST",
		"server host",
		"(default: localhost)",
		"-v, --verbose",
		"enable verbose output",
	}

	for _, part := range expectedParts {
		if !strings.Contains(helpText, part) {
			t.Errorf("Help text should contain %q", part)
		}
	}

	// Boolean options should not show default values
	if strings.Contains(helpText, "(default: false)") {
		t.Error("Boolean options should not show default values")
	}
}

// TestErrorHandlingEdgeCases tests edge cases in error handling
func TestErrorHandlingEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		testStruct  interface{}
		args        []string
		expectError bool
	}{
		{
			name: "empty args with no required fields",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
			}{},
			args:        []string{},
			expectError: false,
		},
		{
			name: "valid args",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:        []string{"--count", "42"},
			expectError: false,
		},
		{
			name: "multiple unknown options",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
			}{},
			args:        []string{"--unknown1", "--unknown2"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewParser(Config{}, tc.testStruct)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			err = parser.Parse(tc.args)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

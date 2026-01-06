package goarg

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestHelpGeneration(t *testing.T) {
	type TestCmd struct {
		Verbose bool     `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int      `arg:"-c,--count" help:"number of items"`
		Input   string   `arg:"--input,required" help:"input file path"`
		Files   []string `arg:"positional" help:"files to process"`
	}

	config := Config{
		Program:     "testapp",
		Description: "A test application for help generation",
		Version:     "1.0.0",
	}

	parser, err := NewParser(config, &TestCmd{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	var buf bytes.Buffer
	parser.WriteHelp(&buf)

	helpText := buf.String()
	t.Logf("Generated help text:\n%s", helpText)

	// Verify key components are present
	if !strings.Contains(helpText, "Usage: testapp") {
		t.Error("Help text should contain usage line")
	}
	if !strings.Contains(helpText, "A test application for help generation") {
		t.Error("Help text should contain description")
	}
	if !strings.Contains(helpText, "Positional arguments:") {
		t.Error("Help text should contain positional arguments section")
	}
	if !strings.Contains(helpText, "Options:") {
		t.Error("Help text should contain options section")
	}
	if !strings.Contains(helpText, "-v, --verbose") {
		t.Error("Help text should contain verbose option")
	}
	if !strings.Contains(helpText, "enable verbose output") {
		t.Error("Help text should contain verbose help text")
	}
	if !strings.Contains(helpText, "Version: 1.0.0") {
		t.Error("Help text should contain version")
	}
}

func TestUsageGeneration(t *testing.T) {
	type TestCmd struct {
		Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
		Input   string `arg:"--input,required" help:"input file path"`
		Source  string `arg:"positional,required" help:"source file"`
		Dest    string `arg:"positional" help:"destination file"`
	}

	config := Config{
		Program: "testapp",
	}

	parser, err := NewParser(config, &TestCmd{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	var buf bytes.Buffer
	parser.WriteUsage(&buf)

	usageText := buf.String()
	t.Logf("Generated usage text:\n%s", usageText)

	// Verify usage format
	if !strings.Contains(usageText, "Usage: testapp [OPTIONS] SOURCE [DEST]") {
		t.Errorf("Usage text should contain correct format, got: %s", usageText)
	}
}

func TestHelpWithSubcommands(t *testing.T) {
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type ClientCmd struct {
		URL string `arg:"-u,--url" help:"client URL"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Server  *ServerCmd `arg:"subcommand:server" help:"run server"`
		Client  *ClientCmd `arg:"subcommand:client" help:"run client"`
	}

	config := Config{
		Program:     "myapp",
		Description: "A multi-command application",
	}

	parser, err := NewParser(config, &RootCmd{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	var buf bytes.Buffer
	parser.WriteHelp(&buf)

	helpText := buf.String()
	t.Logf("Generated help text with subcommands:\n%s", helpText)

	// Verify subcommand components
	if !strings.Contains(helpText, "Usage: myapp COMMAND") {
		t.Error("Help text should contain COMMAND in usage")
	}
	if !strings.Contains(helpText, "Commands:") {
		t.Error("Help text should contain commands section")
	}
	if !strings.Contains(helpText, "server") {
		t.Error("Help text should contain server command")
	}
	if !strings.Contains(helpText, "run server") {
		t.Error("Help text should contain server help text")
	}
	if !strings.Contains(helpText, "client") {
		t.Error("Help text should contain client command")
	}
}

func TestErrorTranslation(t *testing.T) {
	translator := &ErrorTranslator{}

	testCases := []struct {
		name     string
		input    error
		context  ParseContext
		expected string
	}{
		{
			name:     "unknown option",
			input:    fmt.Errorf("unknown option --invalid"),
			context:  ParseContext{},
			expected: "unrecognized argument: --invalid",
		},
		{
			name:     "option requires argument",
			input:    fmt.Errorf("option requires an argument --count"),
			context:  ParseContext{},
			expected: "option requires an argument: --count",
		},
		{
			name:     "missing required field",
			input:    fmt.Errorf("missing required field"),
			context:  ParseContext{FieldName: "input"},
			expected: "required argument missing: input",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := translator.TranslateError(tc.input, tc.context)
			if result.Error() != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result.Error())
			}
		})
	}
}

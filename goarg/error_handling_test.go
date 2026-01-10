package goarg

import (
	"fmt"
	"strings"
	"testing"
)

func TestErrorHandlingIntegration(t *testing.T) {
	testCases := []struct {
		name          string
		testStruct    interface{}
		args          []string
		expectError   bool
		errorContains string
	}{
		{
			name: "unknown option",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
			}{},
			args:          []string{"--unknown"},
			expectError:   true,
			errorContains: "unrecognized argument",
		},
		{
			name: "missing required argument",
			testStruct: &struct {
				Input string `arg:"--input,required"`
			}{},
			args:          []string{},
			expectError:   true,
			errorContains: "required",
		},
		{
			name: "option requires argument",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:          []string{"--count"},
			expectError:   true,
			errorContains: "requires an argument",
		},
		{
			name: "invalid type conversion",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:          []string{"--count", "invalid"},
			expectError:   true,
			errorContains: "invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewParser(Config{}, tc.testStruct)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			err = parser.Parse(tc.args)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain %q, got: %v", tc.errorContains, err)
				}

				t.Logf("Got expected error: %v", err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestErrorTranslationConsistency(t *testing.T) {
	// Test that the same error conditions produce consistent error messages
	type TestStruct struct {
		Input string `arg:"--input,required"`
	}

	parser, err := NewParser(Config{}, &TestStruct{})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test the same error multiple times
	var firstError string
	for i := 0; i < 3; i++ {
		err := parser.Parse([]string{})
		if err == nil {
			t.Errorf("Expected error on iteration %d", i)
			continue
		}

		if i == 0 {
			firstError = err.Error()
		} else if err.Error() != firstError {
			t.Errorf("Error message inconsistent on iteration %d:\nFirst: %s\nCurrent: %s",
				i, firstError, err.Error())
		}
	}
}

func TestSubcommandErrorHandling(t *testing.T) {
	type ServerCmd struct {
		Port int `arg:"-p,--port,required"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose"`
		Server  *ServerCmd `arg:"subcommand:server"`
	}

	testCases := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
	}{
		{
			name:          "missing required subcommand option",
			args:          []string{"server"},
			expectError:   true,
			errorContains: "port", // The actual error message contains "port"
		},
		{
			name:          "unknown subcommand option",
			args:          []string{"server", "--unknown"},
			expectError:   true,
			errorContains: "unknown", // The actual error message contains "unknown"
		},
		{
			name:        "valid subcommand",
			args:        []string{"server", "--port", "8080"},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewParser(Config{}, &RootCmd{})
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			err = parser.Parse(tc.args)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain %q, got: %v", tc.errorContains, err)
				}

				t.Logf("Got expected error: %v", err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestErrorContextInformation(t *testing.T) {
	// Test that error context is properly passed to the translator
	translator := &ErrorTranslator{}

	testCases := []struct {
		name     string
		input    error
		context  ParseContext
		expected string
	}{
		{
			name:     "field context in error",
			input:    fmt.Errorf("missing required field"),
			context:  ParseContext{FieldName: "input"},
			expected: "required argument missing: input",
		},
		{
			name:     "no field context",
			input:    fmt.Errorf("missing required field"),
			context:  ParseContext{},
			expected: "required argument missing",
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

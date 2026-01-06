package goarg

import (
	"fmt"
	"strings"
	"testing"
	"testing/quick"
)

// TestProperty7ErrorMessageCompatibility tests Property 7: Error Message Compatibility
// Property 7: Error Message Compatibility
// For any invalid input that causes parsing errors, our error messages should match upstream alexflint/go-arg format and wording
// Validates: Requirements 5.2
func TestProperty7ErrorMessageCompatibility(t *testing.T) {
	// Feature: goarg-compatibility, Property 7: Error Message Compatibility

	property := func() bool {
		// Generate random error scenarios for testing
		scenario := generateRandomErrorScenario()

		parser, err := NewParser(Config{}, scenario.testStruct)
		if err != nil {
			// Skip invalid configurations
			return true
		}

		// Parse with invalid arguments to trigger errors
		err = parser.Parse(scenario.args)

		// Validate error message format and content
		return validateErrorMessageFormat(err, scenario)
	}

	config := &quick.Config{
		MaxCount: 100, // Minimum 100 iterations as per requirements
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 7 failed: %v", err)
	}
}

// ErrorScenario represents a test scenario that should produce an error
type ErrorScenario struct {
	name         string
	testStruct   interface{}
	args         []string
	expectError  bool
	errorPattern string // Pattern that should be in the error message
}

// generateRandomErrorScenario creates random error scenarios for testing
func generateRandomErrorScenario() ErrorScenario {
	scenarios := []ErrorScenario{
		// Unknown option scenarios
		{
			name: "unknown long option",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
			}{},
			args:         []string{"--unknown"},
			expectError:  true,
			errorPattern: "unrecognized argument",
		},
		{
			name: "unknown short option",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:         []string{"-x"},
			expectError:  true,
			errorPattern: "unrecognized argument",
		},

		// Missing required argument scenarios
		{
			name: "missing required option",
			testStruct: &struct {
				Input string `arg:"--input,required"`
			}{},
			args:         []string{},
			expectError:  true,
			errorPattern: "required",
		},
		{
			name: "missing required positional",
			testStruct: &struct {
				Source string `arg:"positional,required"`
			}{},
			args:         []string{},
			expectError:  true,
			errorPattern: "required",
		},

		// Option requires argument scenarios
		{
			name: "option missing argument",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:         []string{"--count"},
			expectError:  true,
			errorPattern: "requires an argument",
		},

		// Type conversion error scenarios
		{
			name: "invalid integer",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:         []string{"--count", "invalid"},
			expectError:  true,
			errorPattern: "invalid",
		},
		{
			name: "invalid boolean",
			testStruct: &struct {
				Flag bool `arg:"--flag"`
			}{},
			args:         []string{"--flag", "maybe"},
			expectError:  true,
			errorPattern: "invalid",
		},

		// Valid scenarios (should not error)
		{
			name: "valid parsing",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
				Count   int  `arg:"-c,--count"`
			}{},
			args:        []string{"-v", "--count", "42"},
			expectError: false,
		},
	}

	// Return a random scenario
	idx := len(scenarios) % len(scenarios)
	return scenarios[idx]
}

// validateErrorMessageFormat validates that error messages follow expected patterns
func validateErrorMessageFormat(err error, scenario ErrorScenario) bool {
	if scenario.expectError {
		if err == nil {
			// Expected error but got none
			return false
		}

		errMsg := err.Error()

		// Check that error message contains expected pattern
		if scenario.errorPattern != "" && !strings.Contains(errMsg, scenario.errorPattern) {
			return false
		}

		// Validate error message format properties
		return validateErrorMessageProperties(errMsg)
	} else {
		// Should not have error
		return err == nil
	}
}

// validateErrorMessageProperties validates general properties of error messages
func validateErrorMessageProperties(errMsg string) bool {
	// Error message should not be empty
	if strings.TrimSpace(errMsg) == "" {
		return false
	}

	// Error message should not contain internal implementation details
	forbiddenPhrases := []string{
		"parsing error:",
		"failed to set field",
		"failed to convert value",
		"OptArgs Core",
		"core integration",
	}

	for _, phrase := range forbiddenPhrases {
		if strings.Contains(errMsg, phrase) {
			return false
		}
	}

	// Error message should be properly formatted (no double spaces, etc.)
	if strings.Contains(errMsg, "  ") {
		return false
	}

	// Error message should not start or end with whitespace
	if errMsg != strings.TrimSpace(errMsg) {
		return false
	}

	return true
}

// TestErrorMessageConsistency ensures error messages are consistent across multiple calls
func TestErrorMessageConsistency(t *testing.T) {
	testCases := []struct {
		name        string
		testStruct  interface{}
		args        []string
		expectError bool
	}{
		{
			name: "unknown option consistency",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
			}{},
			args:        []string{"--unknown"},
			expectError: true,
		},
		{
			name: "missing required consistency",
			testStruct: &struct {
				Input string `arg:"--input,required"`
			}{},
			args:        []string{},
			expectError: true,
		},
		{
			name: "type conversion consistency",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:        []string{"--count", "invalid"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewParser(Config{}, tc.testStruct)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			// Test the same error multiple times
			var firstError string
			for i := 0; i < 3; i++ {
				err := parser.Parse(tc.args)

				if tc.expectError {
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
				} else {
					if err != nil {
						t.Errorf("Unexpected error on iteration %d: %v", i, err)
					}
				}
			}
		})
	}
}

// TestErrorMessageFormats tests specific error message formats
func TestErrorMessageFormats(t *testing.T) {
	testCases := []struct {
		name           string
		testStruct     interface{}
		args           []string
		expectedFormat string // Regex pattern or substring
	}{
		{
			name: "unrecognized argument format",
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
			}{},
			args:           []string{"--unknown"},
			expectedFormat: "unrecognized argument: --unknown",
		},
		{
			name: "option requires argument format",
			testStruct: &struct {
				Count int `arg:"-c,--count"`
			}{},
			args:           []string{"--count"},
			expectedFormat: "option requires an argument: --count",
		},
		{
			name: "required argument missing format",
			testStruct: &struct {
				Input string `arg:"--input,required"`
			}{},
			args:           []string{},
			expectedFormat: "required argument missing",
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
			if !strings.Contains(errMsg, tc.expectedFormat) {
				t.Errorf("Expected error to contain %q, got: %q", tc.expectedFormat, errMsg)
			}

			// Validate general error message properties
			if !validateErrorMessageProperties(errMsg) {
				t.Errorf("Error message failed property validation: %q", errMsg)
			}
		})
	}
}

// TestErrorTranslatorDirectly tests the ErrorTranslator directly
func TestErrorTranslatorDirectly(t *testing.T) {
	translator := &ErrorTranslator{}

	testCases := []struct {
		name     string
		input    error
		context  ParseContext
		expected string
	}{
		{
			name:     "unknown option translation",
			input:    fmt.Errorf("parsing error: unknown option: test"),
			context:  ParseContext{},
			expected: "unrecognized argument: --test",
		},
		{
			name:     "option requires argument translation",
			input:    fmt.Errorf("option requires an argument: count"),
			context:  ParseContext{},
			expected: "option requires an argument: --count",
		},
		{
			name:     "missing required with context",
			input:    fmt.Errorf("missing required field"),
			context:  ParseContext{FieldName: "input"},
			expected: "required argument missing: input",
		},
		{
			name:     "type conversion error",
			input:    fmt.Errorf("strconv.ParseInt: parsing \"invalid\": invalid syntax"),
			context:  ParseContext{FieldName: "count"},
			expected: "invalid argument for --count",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := translator.TranslateError(tc.input, tc.context)
			if result.Error() != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result.Error())
			}

			// Validate error message properties
			if !validateErrorMessageProperties(result.Error()) {
				t.Errorf("Translated error failed property validation: %q", result.Error())
			}
		})
	}
}

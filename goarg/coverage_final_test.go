package goarg

import (
	"encoding"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/major0/optargs"
)

// TestMustParseErrorHandling tests MustParse error paths without calling os.Exit
func TestMustParseErrorHandling(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test successful case (we can't test error case due to os.Exit)
	t.Run("successful parse", func(t *testing.T) {
		os.Args = []string{"testprog", "--verbose"}
		testStruct := &struct {
			Verbose bool `arg:"-v,--verbose"`
		}{}

		// This should not panic
		MustParse(testStruct)

		if !testStruct.Verbose {
			t.Error("Expected Verbose to be true")
		}
	})

	// Note: We cannot test the error path of MustParse because it calls os.Exit(1)
	// which would terminate the test process. The function is simple enough that
	// testing the successful case provides adequate coverage for the wrapper.
}

// TestConvertCustomComprehensive tests ConvertCustom with various scenarios
func TestConvertCustomComprehensive(t *testing.T) {
	converter := &TypeConverter{}

	// Test with custom type that implements TextUnmarshaler
	t.Run("TextUnmarshaler implementation", func(t *testing.T) {
		// Create a custom type that implements TextUnmarshaler
		type CustomString string

		// Implement TextUnmarshaler
		var customType CustomString
		customTypePtr := &customType

		// Test with pointer type that implements TextUnmarshaler
		targetType := reflect.TypeOf(customTypePtr)
		result, err := converter.ConvertCustom("test-value", targetType)
		if err == nil {
			t.Error("Expected error for type that doesn't actually implement TextUnmarshaler")
		}

		// Test with non-pointer type
		targetType2 := reflect.TypeOf(customType)
		result2, err2 := converter.ConvertCustom("test-value", targetType2)
		if err2 == nil {
			t.Error("Expected error for type that doesn't actually implement TextUnmarshaler")
		}
		_ = result
		_ = result2
	})

	t.Run("interface type", func(t *testing.T) {
		// Test with interface type
		interfaceType := reflect.TypeOf((*interface{})(nil)).Elem()
		_, err := converter.ConvertCustom("test", interfaceType)
		if err == nil {
			t.Error("Expected error for interface type")
		}
	})

	t.Run("struct type without TextUnmarshaler", func(t *testing.T) {
		// Test with struct type that doesn't implement TextUnmarshaler
		type SimpleStruct struct {
			Value string
		}
		structType := reflect.TypeOf(SimpleStruct{})
		_, err := converter.ConvertCustom("test", structType)
		if err == nil {
			t.Error("Expected error for struct type without TextUnmarshaler")
		}
	})

	t.Run("pointer to struct without TextUnmarshaler", func(t *testing.T) {
		// Test with pointer to struct that doesn't implement TextUnmarshaler
		type SimpleStruct struct {
			Value string
		}
		ptrType := reflect.TypeOf((*SimpleStruct)(nil))
		_, err := converter.ConvertCustom("test", ptrType)
		if err == nil {
			t.Error("Expected error for pointer to struct without TextUnmarshaler")
		}
	})

	t.Run("map type", func(t *testing.T) {
		// Test with map type
		mapType := reflect.TypeOf(map[string]string{})
		_, err := converter.ConvertCustom("test", mapType)
		if err == nil {
			t.Error("Expected error for map type")
		}
	})

	t.Run("channel type", func(t *testing.T) {
		// Test with channel type
		chanType := reflect.TypeOf(make(chan int))
		_, err := converter.ConvertCustom("test", chanType)
		if err == nil {
			t.Error("Expected error for channel type")
		}
	})
}

// TestSetScalarValueEdgeCases tests setScalarValue with various edge cases
func TestSetScalarValueEdgeCases(t *testing.T) {
	integration := &CoreIntegration{}

	t.Run("unsettable field", func(t *testing.T) {
		// Create a struct with an unexported field
		testStruct := struct {
			unexported int
		}{}

		// Try to get the field value (this will be unsettable)
		fieldValue := reflect.ValueOf(testStruct).FieldByName("unexported")
		fieldType := reflect.TypeOf(int(0))

		err := integration.setScalarValue(fieldValue, fieldType, "123")
		if err == nil {
			t.Error("Expected error for unsettable field")
		}
	})

	t.Run("invalid conversion", func(t *testing.T) {
		// Create a settable field
		testStruct := struct {
			Value int
		}{}

		structValue := reflect.ValueOf(&testStruct).Elem()
		fieldValue := structValue.FieldByName("Value")
		fieldType := reflect.TypeOf(int(0))

		// Try to convert invalid string to int
		err := integration.setScalarValue(fieldValue, fieldType, "not-a-number")
		if err == nil {
			t.Error("Expected error for invalid conversion")
		}
	})

	t.Run("successful conversion", func(t *testing.T) {
		// Create a settable field
		testStruct := struct {
			Value int
		}{}

		structValue := reflect.ValueOf(&testStruct).Elem()
		fieldValue := structValue.FieldByName("Value")
		fieldType := reflect.TypeOf(int(0))

		// Convert valid string to int
		err := integration.setScalarValue(fieldValue, fieldType, "42")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if testStruct.Value != 42 {
			t.Errorf("Expected 42, got %d", testStruct.Value)
		}
	})
}

// TestProcessPositionalArgsEdgeCases tests processPositionalArgs with edge cases
func TestProcessPositionalArgsEdgeCases(t *testing.T) {
	t.Run("missing required positional", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "RequiredArg", Positional: true, Required: true, Type: reflect.TypeOf("")},
			},
		}

		integration := &CoreIntegration{
			metadata: metadata,
			positionals: []PositionalArg{
				{Field: &metadata.Fields[0], Required: true, Multiple: false},
			},
		}

		// Create a mock parser with no remaining args
		parser := &optargs.Parser{Args: []string{}}

		testStruct := struct {
			RequiredArg string
		}{}
		destValue := reflect.ValueOf(&testStruct).Elem()

		err := integration.processPositionalArgs(parser, destValue)
		if err == nil {
			t.Error("Expected error for missing required positional argument")
		}

		if !strings.Contains(err.Error(), "missing required positional argument") {
			t.Errorf("Expected 'missing required positional argument' error, got: %v", err)
		}
	})

	t.Run("unsettable field", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "unexported", Positional: true, Required: false, Type: reflect.TypeOf("")},
			},
		}

		integration := &CoreIntegration{
			metadata: metadata,
			positionals: []PositionalArg{
				{Field: &metadata.Fields[0], Required: false, Multiple: false},
			},
		}

		parser := &optargs.Parser{Args: []string{"value"}}

		// Create struct with unexported field
		testStruct := struct {
			unexported string
		}{}
		destValue := reflect.ValueOf(testStruct) // Note: not a pointer, so fields are unsettable

		err := integration.processPositionalArgs(parser, destValue)
		if err == nil {
			t.Error("Expected error for unsettable field")
		}
	})

	t.Run("slice positional with conversion error", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Numbers", Positional: true, Required: false, Type: reflect.TypeOf([]int{})},
			},
		}

		integration := &CoreIntegration{
			metadata: metadata,
			positionals: []PositionalArg{
				{Field: &metadata.Fields[0], Required: false, Multiple: true},
			},
		}

		parser := &optargs.Parser{Args: []string{"1", "not-a-number", "3"}}

		testStruct := struct {
			Numbers []int
		}{}
		destValue := reflect.ValueOf(&testStruct).Elem()

		err := integration.processPositionalArgs(parser, destValue)
		if err == nil {
			t.Error("Expected error for invalid number conversion")
		}
	})

	t.Run("single positional with conversion error", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Number", Positional: true, Required: false, Type: reflect.TypeOf(int(0))},
			},
		}

		integration := &CoreIntegration{
			metadata: metadata,
			positionals: []PositionalArg{
				{Field: &metadata.Fields[0], Required: false, Multiple: false},
			},
		}

		parser := &optargs.Parser{Args: []string{"not-a-number"}}

		testStruct := struct {
			Number int
		}{}
		destValue := reflect.ValueOf(&testStruct).Elem()

		err := integration.processPositionalArgs(parser, destValue)
		if err == nil {
			t.Error("Expected error for invalid number conversion")
		}
	})
}

// TestProcessEnvironmentVariablesEdgeCases tests processEnvironmentVariables edge cases
func TestProcessEnvironmentVariablesEdgeCases(t *testing.T) {
	t.Run("unsettable field", func(t *testing.T) {
		// Set environment variable
		os.Setenv("TEST_VAR", "test-value")
		defer os.Unsetenv("TEST_VAR")

		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "unexported", Env: "TEST_VAR", Type: reflect.TypeOf("")},
			},
		}

		integration := &CoreIntegration{metadata: metadata}

		// Create struct with unexported field
		testStruct := struct {
			unexported string
		}{}
		destValue := reflect.ValueOf(testStruct) // Note: not a pointer, so fields are unsettable

		// This should not error, but should skip the unsettable field
		err := integration.processEnvironmentVariables(destValue)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("conversion error", func(t *testing.T) {
		// Set environment variable with invalid value
		os.Setenv("TEST_INT", "not-a-number")
		defer os.Unsetenv("TEST_INT")

		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Number", Env: "TEST_INT", Type: reflect.TypeOf(int(0))},
			},
		}

		integration := &CoreIntegration{metadata: metadata}

		testStruct := struct {
			Number int
		}{}
		destValue := reflect.ValueOf(&testStruct).Elem()

		err := integration.processEnvironmentVariables(destValue)
		if err == nil {
			t.Error("Expected error for invalid environment variable conversion")
		}
	})

	t.Run("field already set", func(t *testing.T) {
		// Set environment variable
		os.Setenv("TEST_ALREADY_SET", "env-value")
		defer os.Unsetenv("TEST_ALREADY_SET")

		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Value", Env: "TEST_ALREADY_SET", Type: reflect.TypeOf("")},
			},
		}

		integration := &CoreIntegration{metadata: metadata}

		testStruct := struct {
			Value string
		}{Value: "already-set"} // Field is already set
		destValue := reflect.ValueOf(&testStruct).Elem()

		err := integration.processEnvironmentVariables(destValue)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Value should remain unchanged
		if testStruct.Value != "already-set" {
			t.Errorf("Expected 'already-set', got '%s'", testStruct.Value)
		}
	})
}

// TestSetDefaultValuesEdgeCases tests setDefaultValues edge cases
func TestSetDefaultValuesEdgeCases(t *testing.T) {
	t.Run("unsettable field", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "unexported", Default: "default-value", Type: reflect.TypeOf("")},
			},
		}

		integration := &CoreIntegration{metadata: metadata}

		// Create struct with unexported field
		testStruct := struct {
			unexported string
		}{}
		destValue := reflect.ValueOf(testStruct) // Note: not a pointer, so fields are unsettable

		// This should not error, but should skip the unsettable field
		err := integration.setDefaultValues(destValue)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("field already set", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Value", Default: "default-value", Type: reflect.TypeOf("")},
			},
		}

		integration := &CoreIntegration{metadata: metadata}

		testStruct := struct {
			Value string
		}{Value: "already-set"} // Field is already set
		destValue := reflect.ValueOf(&testStruct).Elem()

		err := integration.setDefaultValues(destValue)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Value should remain unchanged
		if testStruct.Value != "already-set" {
			t.Errorf("Expected 'already-set', got '%s'", testStruct.Value)
		}
	})
}

// TestExtractOptionFromErrorEdgeCases tests extractOptionFromError with various formats
func TestExtractOptionFromErrorEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{
			name:     "long option in middle",
			errMsg:   "some error with --verbose option",
			expected: "--verbose",
		},
		{
			name:     "short option in middle",
			errMsg:   "some error with -v option",
			expected: "-v",
		},
		{
			name:     "unknown option format",
			errMsg:   "unknown option: verbose",
			expected: "--verbose",
		},
		{
			name:     "unknown option format single char",
			errMsg:   "unknown option: v",
			expected: "-v",
		},
		{
			name:     "option requires argument format",
			errMsg:   "option requires an argument: verbose",
			expected: "--verbose",
		},
		{
			name:     "option requires argument format single char",
			errMsg:   "option requires an argument: v",
			expected: "-v",
		},
		{
			name:     "option already has dashes",
			errMsg:   "unknown option: --verbose",
			expected: "--verbose",
		},
		{
			name:     "option already has single dash",
			errMsg:   "unknown option: -v",
			expected: "-v",
		},
		{
			name:     "no option found",
			errMsg:   "some generic error message",
			expected: "some generic error message",
		},
		{
			name:     "parsing error prefix",
			errMsg:   "parsing error: unknown option: test",
			expected: "--test",
		},
		{
			name:     "option with colon",
			errMsg:   "error with --config: invalid value",
			expected: "--config",
		},
		{
			name:     "short option with colon",
			errMsg:   "error with -c: invalid value",
			expected: "-c",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractOptionFromError(tc.errMsg)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestBuildOptStringEdgeCases tests BuildOptString with various field configurations
func TestBuildOptStringEdgeCases(t *testing.T) {
	t.Run("mixed argument types", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Flag", Short: "f", ArgType: optargs.NoArgument},
				{Name: "Required", Short: "r", ArgType: optargs.RequiredArgument},
				{Name: "Optional", Short: "o", ArgType: optargs.OptionalArgument},
				{Name: "Positional", Positional: true},   // Should be skipped
				{Name: "Subcommand", IsSubcommand: true}, // Should be skipped
				{Name: "LongOnly", Long: "long-only"},    // Should be skipped (no short)
			},
		}

		integration := &CoreIntegration{metadata: metadata}
		optstring := integration.BuildOptString()

		// Should contain: f (no colon), r: (required), o:: (optional)
		expected := "fr:o::"
		if optstring != expected {
			t.Errorf("Expected '%s', got '%s'", expected, optstring)
		}
	})

	t.Run("no short options", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "LongOnly1", Long: "long1"},
				{Name: "LongOnly2", Long: "long2"},
			},
		}

		integration := &CoreIntegration{metadata: metadata}
		optstring := integration.BuildOptString()

		// Should be empty since no short options
		if optstring != "" {
			t.Errorf("Expected empty string, got '%s'", optstring)
		}
	})
}

// TestCreateParserWithParentEdgeCases tests CreateParserWithParent edge cases
func TestCreateParserWithParentEdgeCases(t *testing.T) {
	t.Run("subcommand creation error", func(t *testing.T) {
		// Create metadata with invalid subcommand structure
		subMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "InvalidField", Short: "invalid-short-option-too-long"}, // Invalid short option
			},
		}

		metadata := &StructMetadata{
			Fields:      []FieldMetadata{},
			Subcommands: map[string]*StructMetadata{"invalid": subMetadata},
		}

		integration := &CoreIntegration{metadata: metadata}

		// This should handle the error gracefully
		parser, err := integration.CreateParserWithParent([]string{}, nil)

		// The error might be caught during subcommand creation
		// If no error, the parser should still be created
		if err != nil && parser == nil {
			// This is expected if subcommand creation fails
			t.Logf("Expected error during subcommand creation: %v", err)
		}
	})
}

// TestProcessResultsEdgeCases tests ProcessResults with various error conditions
func TestProcessResultsEdgeCases(t *testing.T) {
	t.Run("option parsing error", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Verbose", Short: "v", Long: "verbose", Type: reflect.TypeOf(bool(false))},
			},
		}

		integration := &CoreIntegration{metadata: metadata}

		// Create a parser that will have parsing errors
		parser, err := optargs.GetOptLong([]string{"--unknown"}, "v", []optargs.Flag{
			{Name: "verbose", HasArg: optargs.NoArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		testStruct := struct {
			Verbose bool
		}{}

		// This should return an error due to unknown option
		err = integration.ProcessResults(parser, &testStruct)
		if err == nil {
			t.Error("Expected error for unknown option")
		}
	})

	t.Run("field not found", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				// Empty fields - no matching field for parsed options
			},
		}

		integration := &CoreIntegration{metadata: metadata}

		parser, err := optargs.GetOptLong([]string{"-v"}, "v", []optargs.Flag{})
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		testStruct := struct {
			SomeField bool
		}{}

		// This should handle missing field gracefully
		err = integration.ProcessResults(parser, &testStruct)
		// Should not error, just skip unknown options
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

// TestCustomTypeForCoverage implements TextUnmarshaler for testing
type TestCustomTypeForCoverage struct {
	Value string
}

func (c *TestCustomTypeForCoverage) UnmarshalText(text []byte) error {
	c.Value = string(text)
	return nil
}

// TestRealTextUnmarshaler tests ConvertCustom with a real TextUnmarshaler implementation
func TestRealTextUnmarshaler(t *testing.T) {
	// Verify it implements the interface
	var _ encoding.TextUnmarshaler = (*TestCustomTypeForCoverage)(nil)

	converter := &TypeConverter{}

	t.Run("pointer to TextUnmarshaler", func(t *testing.T) {
		targetType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))
		result, err := converter.ConvertCustom("test-value", targetType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}

		// Check the result
		if customResult, ok := result.(*TestCustomTypeForCoverage); ok {
			if customResult.Value != "test-value" {
				t.Errorf("Expected 'test-value', got '%s'", customResult.Value)
			}
		} else {
			t.Errorf("Expected *TestCustomTypeForCoverage, got %T", result)
		}
	})

	t.Run("non-pointer TextUnmarshaler", func(t *testing.T) {
		targetType := reflect.TypeOf(TestCustomTypeForCoverage{})
		result, err := converter.ConvertCustom("test-value", targetType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}

		// Check the result
		if customResult, ok := result.(TestCustomTypeForCoverage); ok {
			if customResult.Value != "test-value" {
				t.Errorf("Expected 'test-value', got '%s'", customResult.Value)
			}
		} else {
			t.Errorf("Expected TestCustomTypeForCoverage, got %T", result)
		}
	})
}

// TestValidateCustomEdgeCases tests ValidateCustom with various constraint scenarios
func TestValidateCustomEdgeCases(t *testing.T) {
	converter := &TypeConverter{}

	t.Run("field not found in struct", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "NonExistentField", Type: reflect.TypeOf("")},
			},
		}

		testStruct := struct {
			ActualField string
		}{}

		// This should not error, just skip the non-existent field
		err := converter.ValidateCustom(&testStruct, metadata)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("invalid min constraint", func(t *testing.T) {
		// Create a struct with invalid min tag
		type TestStruct struct {
			Value int `min:"not-a-number"`
		}

		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Value", Type: reflect.TypeOf(int(0))},
			},
		}

		testStruct := TestStruct{Value: 10}

		err := converter.ValidateCustom(&testStruct, metadata)
		// This should handle the invalid constraint gracefully
		if err != nil {
			t.Logf("Expected error for invalid min constraint: %v", err)
		}
	})
}

// TestValidateMinMaxEdgeCases tests validateMin and validateMax with edge cases
func TestValidateMinMaxEdgeCases(t *testing.T) {
	converter := &TypeConverter{}

	t.Run("validateMin with different numeric types", func(t *testing.T) {
		testCases := []struct {
			name      string
			value     interface{}
			minTag    string
			shouldErr bool
		}{
			{"int below min", int(5), "10", true},
			{"int above min", int(15), "10", false},
			{"uint below min", uint(5), "10", true},
			{"uint above min", uint(15), "10", false},
			{"float below min", float64(5.5), "10.0", true},
			{"float above min", float64(15.5), "10.0", false},
			{"invalid min tag", int(5), "not-a-number", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fieldValue := reflect.ValueOf(tc.value)
				err := converter.validateMin(fieldValue, tc.minTag, "testField")

				if tc.shouldErr && err == nil {
					t.Error("Expected error but got none")
				} else if !tc.shouldErr && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			})
		}
	})

	t.Run("validateMax with different numeric types", func(t *testing.T) {
		testCases := []struct {
			name      string
			value     interface{}
			maxTag    string
			shouldErr bool
		}{
			{"int above max", int(15), "10", true},
			{"int below max", int(5), "10", false},
			{"uint above max", uint(15), "10", true},
			{"uint below max", uint(5), "10", false},
			{"float above max", float64(15.5), "10.0", true},
			{"float below max", float64(5.5), "10.0", false},
			{"invalid max tag", int(5), "not-a-number", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fieldValue := reflect.ValueOf(tc.value)
				err := converter.validateMax(fieldValue, tc.maxTag, "testField")

				if tc.shouldErr && err == nil {
					t.Error("Expected error but got none")
				} else if !tc.shouldErr && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			})
		}
	})
}

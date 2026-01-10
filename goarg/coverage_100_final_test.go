package goarg

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/major0/optargs"
)

// TestCustomTypeForCoverage implements TextUnmarshaler for testing
type TestCustomTypeForCoverage struct {
	Value string
}

func (c *TestCustomTypeForCoverage) UnmarshalText(text []byte) error {
	c.Value = string(text)
	return nil
}

// ErrorUnmarshaler is a type that implements TextUnmarshaler but returns an error
type ErrorUnmarshaler struct {
	Value string
}

func (e *ErrorUnmarshaler) UnmarshalText(text []byte) error {
	return fmt.Errorf("intentional unmarshal error")
}

// TestMustParseErrorDetection tests MustParse error detection without os.Exit
func TestMustParseErrorDetection(t *testing.T) {
	// We can test the error detection path by checking if Parse would return an error
	// This gives us coverage on the error detection logic without calling os.Exit
	t.Run("error detection path", func(t *testing.T) {
		testStruct := &struct {
			Required string `arg:"--required,required"`
		}{}

		// Save original args
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Set args that will cause an error (missing required field)
		os.Args = []string{"testprog"}

		// Test that Parse would return an error (this covers the error detection logic)
		err := Parse(testStruct)
		if err == nil {
			t.Error("Expected error for missing required field")
		}

		// Note: We cannot test the actual os.Exit(1) call in MustParse
		// because it would terminate the test process. The function is simple:
		// if err := Parse(dest); err != nil {
		//     fmt.Fprintln(os.Stderr, err)
		//     os.Exit(1)
		// }
		// Testing Parse() with the same input gives us confidence that
		// MustParse would detect the error and call os.Exit(1).
	})
}

// TestConvertCustomCompleteEdgeCases tests all remaining ConvertCustom paths
func TestConvertCustomCompleteEdgeCases(t *testing.T) {
	converter := &TypeConverter{}

	// Test with interface{} type
	t.Run("interface type", func(t *testing.T) {
		interfaceType := reflect.TypeOf((*interface{})(nil)).Elem()
		_, err := converter.ConvertCustom("test", interfaceType)
		if err == nil {
			t.Error("Expected error for interface{} type")
		}
	})

	// Test with function type
	t.Run("function type", func(t *testing.T) {
		funcType := reflect.TypeOf(func() {})
		_, err := converter.ConvertCustom("test", funcType)
		if err == nil {
			t.Error("Expected error for function type")
		}
	})

	// Test with channel type
	t.Run("channel type", func(t *testing.T) {
		chanType := reflect.TypeOf(make(chan int))
		_, err := converter.ConvertCustom("test", chanType)
		if err == nil {
			t.Error("Expected error for channel type")
		}
	})

	// Test with map type
	t.Run("map type", func(t *testing.T) {
		mapType := reflect.TypeOf(map[string]int{})
		_, err := converter.ConvertCustom("test", mapType)
		if err == nil {
			t.Error("Expected error for map type")
		}
	})

	// Test with array type
	t.Run("array type", func(t *testing.T) {
		arrayType := reflect.TypeOf([5]int{})
		_, err := converter.ConvertCustom("test", arrayType)
		if err == nil {
			t.Error("Expected error for array type")
		}
	})

	// Test with struct type that doesn't implement TextUnmarshaler
	t.Run("struct without TextUnmarshaler", func(t *testing.T) {
		structType := reflect.TypeOf(struct{ Value string }{})
		_, err := converter.ConvertCustom("test", structType)
		if err == nil {
			t.Error("Expected error for struct without TextUnmarshaler")
		}
		if !strings.Contains(err.Error(), "does not implement encoding.TextUnmarshaler") {
			t.Errorf("Expected TextUnmarshaler error, got: %v", err)
		}
	})

	// Test with pointer to struct that doesn't implement TextUnmarshaler
	t.Run("pointer to struct without TextUnmarshaler", func(t *testing.T) {
		ptrType := reflect.TypeOf((*struct{ Value string })(nil))
		_, err := converter.ConvertCustom("test", ptrType)
		if err == nil {
			t.Error("Expected error for pointer to struct without TextUnmarshaler")
		}
	})
}

// TestProcessOptionsWithInheritanceCompleteScenarios tests all inheritance paths
func TestProcessOptionsWithInheritanceCompleteScenarios(t *testing.T) {
	t.Run("complex inheritance with field setting errors", func(t *testing.T) {
		// Create parent metadata with an option
		parentMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "ParentFlag", Short: "p", Long: "parent", ArgType: optargs.NoArgument, Type: reflect.TypeOf(bool(false))},
			},
		}

		// Create subcommand metadata with an option
		subMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "SubFlag", Short: "s", Long: "sub", ArgType: optargs.NoArgument, Type: reflect.TypeOf(bool(false))},
			},
		}

		// Create structs
		parentStruct := struct {
			ParentFlag bool
		}{}
		subStruct := struct {
			SubFlag bool
		}{}

		// Create parser
		parser := &Parser{
			config:   Config{Program: "test"},
			dest:     &parentStruct,
			metadata: parentMetadata,
		}

		// Create parent integration
		parentIntegration := &CoreIntegration{
			metadata:    parentMetadata,
			shortOpts:   make(map[byte]*optargs.Flag),
			longOpts:    make(map[string]*optargs.Flag),
			positionals: []PositionalArg{},
		}
		parentIntegration.BuildLongOpts()

		// Create a subcommand parser with both parent and subcommand options
		subParser, err := optargs.GetOptLong([]string{"-p", "-s"}, "ps", []optargs.Flag{
			{Name: "parent", HasArg: optargs.NoArgument},
			{Name: "sub", HasArg: optargs.NoArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create subparser: %v", err)
		}

		// This should process both inherited and subcommand options
		err = parser.processOptionsWithInheritance(subParser, parentIntegration, subMetadata, &subStruct)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify both flags were set
		if !parentStruct.ParentFlag {
			t.Error("Expected parent flag to be set")
		}
		if !subStruct.SubFlag {
			t.Error("Expected sub flag to be set")
		}
	})

	t.Run("inheritance with unknown option", func(t *testing.T) {
		// Create empty metadata (no fields)
		parentMetadata := &StructMetadata{Fields: []FieldMetadata{}}
		subMetadata := &StructMetadata{Fields: []FieldMetadata{}}

		parentStruct := struct{}{}
		subStruct := struct{}{}

		parser := &Parser{
			config:   Config{Program: "test"},
			dest:     &parentStruct,
			metadata: parentMetadata,
		}

		parentIntegration := &CoreIntegration{
			metadata:    parentMetadata,
			shortOpts:   make(map[byte]*optargs.Flag),
			longOpts:    make(map[string]*optargs.Flag),
			positionals: []PositionalArg{},
		}

		// Create a subcommand parser with an unknown option
		subParser, err := optargs.GetOptLong([]string{"-x"}, "x", []optargs.Flag{
			{Name: "unknown", HasArg: optargs.NoArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create subparser: %v", err)
		}

		// This should handle unknown options gracefully
		err = parser.processOptionsWithInheritance(subParser, parentIntegration, subMetadata, &subStruct)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("inheritance with field setting error", func(t *testing.T) {
		// Create metadata with a field that will cause conversion error
		subMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Number", Short: "n", Long: "number", ArgType: optargs.RequiredArgument, Type: reflect.TypeOf(int(0))},
			},
		}

		parentMetadata := &StructMetadata{Fields: []FieldMetadata{}}

		parentStruct := struct{}{}
		subStruct := struct {
			Number int
		}{}

		parser := &Parser{
			config:   Config{Program: "test"},
			dest:     &parentStruct,
			metadata: parentMetadata,
		}

		parentIntegration := &CoreIntegration{
			metadata:    parentMetadata,
			shortOpts:   make(map[byte]*optargs.Flag),
			longOpts:    make(map[string]*optargs.Flag),
			positionals: []PositionalArg{},
		}

		// Create a subcommand parser with invalid number
		subParser, err := optargs.GetOptLong([]string{"-n", "not-a-number"}, "n:", []optargs.Flag{
			{Name: "number", HasArg: optargs.RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create subparser: %v", err)
		}

		// This should return an error due to invalid number conversion
		err = parser.processOptionsWithInheritance(subParser, parentIntegration, subMetadata, &subStruct)
		if err == nil {
			t.Error("Expected error for invalid number conversion")
		}
	})

	t.Run("inheritance with validation error", func(t *testing.T) {
		// Create metadata with required field
		subMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Required", Long: "required", Required: true, Type: reflect.TypeOf("")},
			},
		}

		parentMetadata := &StructMetadata{Fields: []FieldMetadata{}}

		parentStruct := struct{}{}
		subStruct := struct {
			Required string
		}{}

		parser := &Parser{
			config:   Config{Program: "test"},
			dest:     &parentStruct,
			metadata: parentMetadata,
		}

		parentIntegration := &CoreIntegration{
			metadata:    parentMetadata,
			shortOpts:   make(map[byte]*optargs.Flag),
			longOpts:    make(map[string]*optargs.Flag),
			positionals: []PositionalArg{},
		}

		// Create a subcommand parser with no options (missing required field)
		subParser, err := optargs.GetOptLong([]string{}, "", []optargs.Flag{})
		if err != nil {
			t.Fatalf("Failed to create subparser: %v", err)
		}

		// This should return an error due to missing required field
		err = parser.processOptionsWithInheritance(subParser, parentIntegration, subMetadata, &subStruct)
		if err == nil {
			t.Error("Expected error for missing required field")
		}
	})
}

// TestRemainingUncoveredPaths tests any other uncovered paths
func TestRemainingUncoveredPaths(t *testing.T) {
	// Test NewParser with invalid destination
	t.Run("NewParser with nil destination", func(t *testing.T) {
		_, err := NewParser(Config{}, nil)
		if err == nil {
			t.Error("Expected error for nil destination")
		}
	})

	t.Run("NewParser with non-pointer destination", func(t *testing.T) {
		_, err := NewParser(Config{}, struct{}{})
		if err == nil {
			t.Error("Expected error for non-pointer destination")
		}
	})

	t.Run("NewParser with pointer to non-struct", func(t *testing.T) {
		var i int
		_, err := NewParser(Config{}, &i)
		if err == nil {
			t.Error("Expected error for pointer to non-struct")
		}
	})

	// Test Parse with subcommand not found
	t.Run("Parse with unknown subcommand", func(t *testing.T) {
		testStruct := &struct {
			Server *struct {
				Port int `arg:"--port"`
			} `arg:"subcommand:server"`
		}{}

		parser, err := NewParser(Config{}, testStruct)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// Try to parse with unknown subcommand
		err = parser.Parse([]string{"unknown-command"})
		// This should not error, just process as regular parsing
		if err != nil {
			t.Logf("Got error for unknown subcommand: %v", err)
		}
	})

	// Test findSubcommand with case insensitive match
	t.Run("findSubcommand case insensitive", func(t *testing.T) {
		metadata := &StructMetadata{
			Subcommands: map[string]*StructMetadata{
				"server": {Fields: []FieldMetadata{}},
			},
		}

		parser := &Parser{metadata: metadata}

		// Test exact match
		subMeta, name := parser.findSubcommand("server")
		if subMeta == nil || name != "server" {
			t.Error("Expected exact match for 'server'")
		}

		// Test case insensitive match
		subMeta, name = parser.findSubcommand("SERVER")
		if subMeta == nil || name != "server" {
			t.Error("Expected case insensitive match for 'SERVER'")
		}

		// Test no match
		subMeta, name = parser.findSubcommand("unknown")
		if subMeta != nil || name != "" {
			t.Error("Expected no match for 'unknown'")
		}
	})
}

// TestComplexTextUnmarshalerScenarios tests complex TextUnmarshaler scenarios
func TestComplexTextUnmarshalerScenarios(t *testing.T) {
	converter := &TypeConverter{}

	// Test with a type that implements TextUnmarshaler but returns error
	t.Run("TextUnmarshaler with error", func(t *testing.T) {
		ptrType := reflect.TypeOf((*ErrorUnmarshaler)(nil))
		_, err := converter.ConvertCustom("test", ptrType)
		if err == nil {
			t.Error("Expected error from TextUnmarshaler")
		}
	})

	// Test the specific paths in ConvertCustom for complete coverage
	t.Run("ConvertCustom all paths", func(t *testing.T) {
		// Test with our working TextUnmarshaler
		ptrType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))
		result, err := converter.ConvertCustom("test-value", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}

		// Test with value type
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})
		result, err = converter.ConvertCustom("test-value", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
	})
}

// TestEdgeCaseErrorTranslation tests edge cases in error translation
func TestEdgeCaseErrorTranslation(t *testing.T) {
	translator := &ErrorTranslator{}

	// Test with nil error
	t.Run("nil error", func(t *testing.T) {
		result := translator.TranslateError(nil, ParseContext{})
		if result != nil {
			t.Error("Expected nil for nil error")
		}
	})

	// Test with complex nested error
	t.Run("complex nested error", func(t *testing.T) {
		err := fmt.Errorf("outer: middle: inner: actual error")
		context := ParseContext{FieldName: "test"}
		result := translator.TranslateError(err, context)
		if result == nil {
			t.Error("Expected non-nil result")
		}
	})

	// Test extractOptionFromError with various edge cases
	t.Run("extractOptionFromError edge cases", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"parsing error: --test", "--test"},
			{"parsing error: -t", "-t"},
			{"no option here", "no option here"},
			{"--option at end", "--option"},
			{"-o at end", "-o"},
			{"unknown option: test", "--test"},
			{"unknown option: t", "-t"},
			{"option requires an argument: test", "--test"},
			{"option requires an argument: t", "-t"},
		}

		for _, tc := range testCases {
			result := extractOptionFromError(tc.input)
			if result != tc.expected {
				t.Errorf("For input '%s', expected '%s', got '%s'", tc.input, tc.expected, result)
			}
		}
	})
}

// Ensure our test types are properly defined
var _ encoding.TextUnmarshaler = (*TestCustomTypeForCoverage)(nil)
var _ encoding.TextUnmarshaler = (*ErrorUnmarshaler)(nil)

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

	t.Run("field setting error", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Number", Short: "n", Long: "number", Type: reflect.TypeOf(int(0))},
			},
		}

		integration := &CoreIntegration{metadata: metadata}

		parser, err := optargs.GetOptLong([]string{"-n", "not-a-number"}, "n:", []optargs.Flag{
			{Name: "number", HasArg: optargs.RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		testStruct := struct {
			Number int
		}{}

		err = integration.ProcessResults(parser, &testStruct)
		if err == nil {
			t.Error("Expected error for invalid number conversion")
		}
	})
}

// Ensure our test types implement the interface correctly
var _ encoding.TextUnmarshaler = (*TestCustomTypeForCoverage)(nil)
var _ encoding.TextUnmarshaler = (*ErrorUnmarshaler)(nil)

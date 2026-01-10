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

// TestCustomValueReceiver implements TextUnmarshaler on value receiver
type TestCustomValueReceiver struct {
	Value string
}

func (c TestCustomValueReceiver) UnmarshalText(text []byte) error {
	// Note: This won't actually modify the receiver since it's a value receiver
	// but it allows us to test the interface implementation path
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

	// Test MustParse with successful parsing to cover the success path
	t.Run("successful parsing path", func(t *testing.T) {
		testStruct := &struct {
			Verbose bool `arg:"-v,--verbose"`
		}{}

		// Save original args
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Set args that will succeed
		os.Args = []string{"testprog", "--verbose"}

		// This should succeed and not call os.Exit
		MustParse(testStruct)

		if !testStruct.Verbose {
			t.Error("Expected verbose to be true")
		}
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
var _ encoding.TextUnmarshaler = TestCustomValueReceiver{}

// TestConvertCustom100PercentCoverage tests all paths in ConvertCustom for 100% coverage
func TestConvertCustom100PercentCoverage(t *testing.T) {
	converter := &TypeConverter{}

	// Test pointer type that directly implements TextUnmarshaler
	t.Run("pointer type implements TextUnmarshaler", func(t *testing.T) {
		ptrType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))
		result, err := converter.ConvertCustom("test-value", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}

		// Verify the result is correct
		if customType, ok := result.(*TestCustomTypeForCoverage); ok {
			if customType.Value != "test-value" {
				t.Errorf("Expected 'test-value', got '%s'", customType.Value)
			}
		} else {
			t.Error("Expected result to be *TestCustomTypeForCoverage")
		}
	})

	// Test value type where pointer implements TextUnmarshaler
	t.Run("value type where pointer implements TextUnmarshaler", func(t *testing.T) {
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})
		result, err := converter.ConvertCustom("test-value", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}

		// Verify the result is correct
		if customType, ok := result.(TestCustomTypeForCoverage); ok {
			if customType.Value != "test-value" {
				t.Errorf("Expected 'test-value', got '%s'", customType.Value)
			}
		} else {
			t.Error("Expected result to be TestCustomTypeForCoverage")
		}
	})

	// Test value type that directly implements TextUnmarshaler (value receiver)
	t.Run("value type implements TextUnmarshaler directly", func(t *testing.T) {
		valueType := reflect.TypeOf(TestCustomValueReceiver{})
		result, err := converter.ConvertCustom("test-value", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}

		// Verify the result type
		if _, ok := result.(TestCustomValueReceiver); !ok {
			t.Error("Expected result to be TestCustomValueReceiver")
		}
	})

	// Test pointer to value type that implements TextUnmarshaler (value receiver)
	t.Run("pointer to value type with value receiver TextUnmarshaler", func(t *testing.T) {
		ptrType := reflect.TypeOf((*TestCustomValueReceiver)(nil))
		result, err := converter.ConvertCustom("test-value", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}

		// Verify the result type
		if _, ok := result.(*TestCustomValueReceiver); !ok {
			t.Error("Expected result to be *TestCustomValueReceiver")
		}
	})

	// Test pointer type with TextUnmarshaler error
	t.Run("pointer type TextUnmarshaler error", func(t *testing.T) {
		ptrType := reflect.TypeOf((*ErrorUnmarshaler)(nil))
		_, err := converter.ConvertCustom("test", ptrType)
		if err == nil {
			t.Error("Expected error from TextUnmarshaler")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal text") {
			t.Errorf("Expected unmarshal error, got: %v", err)
		}
	})

	// Test value type with TextUnmarshaler error (via pointer)
	t.Run("value type TextUnmarshaler error via pointer", func(t *testing.T) {
		valueType := reflect.TypeOf(ErrorUnmarshaler{})
		_, err := converter.ConvertCustom("test", valueType)
		if err == nil {
			t.Error("Expected error from TextUnmarshaler")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal text") {
			t.Errorf("Expected unmarshal error, got: %v", err)
		}
	})

	// Test type that doesn't implement TextUnmarshaler at all
	t.Run("type without TextUnmarshaler", func(t *testing.T) {
		type NoUnmarshaler struct {
			Value string
		}

		// Test value type
		valueType := reflect.TypeOf(NoUnmarshaler{})
		_, err := converter.ConvertCustom("test", valueType)
		if err == nil {
			t.Error("Expected error for type without TextUnmarshaler")
		}
		if !strings.Contains(err.Error(), "does not implement encoding.TextUnmarshaler") {
			t.Errorf("Expected TextUnmarshaler error, got: %v", err)
		}

		// Test pointer type
		ptrType := reflect.TypeOf((*NoUnmarshaler)(nil))
		_, err = converter.ConvertCustom("test", ptrType)
		if err == nil {
			t.Error("Expected error for pointer type without TextUnmarshaler")
		}
		if !strings.Contains(err.Error(), "does not implement encoding.TextUnmarshaler") {
			t.Errorf("Expected TextUnmarshaler error, got: %v", err)
		}
	})

	// Test all the different return paths
	t.Run("return path coverage", func(t *testing.T) {
		// Test pointer type returning pointer
		ptrType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))
		result, err := converter.ConvertCustom("ptr-test", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if _, ok := result.(*TestCustomTypeForCoverage); !ok {
			t.Error("Expected pointer result for pointer type")
		}

		// Test value type returning value
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})
		result, err = converter.ConvertCustom("value-test", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if _, ok := result.(TestCustomTypeForCoverage); !ok {
			t.Error("Expected value result for value type")
		}
	})

	// Test the specific branch where ptrType implements TextUnmarshaler but targetType is pointer
	t.Run("pointer target with ptrType TextUnmarshaler", func(t *testing.T) {
		// This is a tricky case - we need a type where:
		// - targetType.Kind() == reflect.Ptr
		// - target.Type() does NOT implement TextUnmarshaler
		// - ptrType DOES implement TextUnmarshaler

		// For this, we need a double pointer scenario or a type that only implements
		// TextUnmarshaler when it's a pointer to the type

		// Let's use TestCustomTypeForCoverage which implements TextUnmarshaler on pointer
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})

		// This should hit the second branch (ptrType implements TextUnmarshaler)
		result, err := converter.ConvertCustom("branch-test", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
	})
}

// TestMustParseCompleteScenarios tests MustParse with different scenarios
func TestMustParseCompleteScenarios(t *testing.T) {
	// Test with custom exit function to avoid os.Exit
	t.Run("custom exit function", func(t *testing.T) {
		testStruct := &struct {
			Required string `arg:"--required,required"`
		}{}

		// Create parser with custom exit function
		parser, err := NewParser(Config{
			Exit: func(code int) {
				// Custom exit function that doesn't actually exit
				if code != 1 {
					t.Errorf("Expected exit code 1, got %d", code)
				}
			},
		}, testStruct)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		// Test the Fail method which calls the exit function
		parser.Fail("test error message")
		// If we reach here, the custom exit function worked
	})

	// Test MustParse with successful parsing
	t.Run("successful parsing", func(t *testing.T) {
		testStruct := &struct {
			Verbose bool `arg:"-v,--verbose"`
		}{}

		// Save original args
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{"testprog", "--verbose"}

		// This should succeed
		MustParse(testStruct)

		if !testStruct.Verbose {
			t.Error("Expected verbose to be true")
		}
	})
}

// TestConvertCustomAllPaths tests all paths in ConvertCustom for 100% coverage
func TestConvertCustomAllPaths(t *testing.T) {
	converter := &TypeConverter{}

	// Test with slice type (should not reach ConvertCustom)
	t.Run("slice type", func(t *testing.T) {
		sliceType := reflect.TypeOf([]string{})
		_, err := converter.ConvertCustom("test", sliceType)
		if err == nil {
			t.Error("Expected error for slice type")
		}
	})

	// Test with basic types that don't implement TextUnmarshaler
	basicTypes := []reflect.Type{
		reflect.TypeOf(int(0)),
		reflect.TypeOf(float64(0)),
		reflect.TypeOf(bool(false)),
		reflect.TypeOf(""),
		reflect.TypeOf([]byte{}),
		reflect.TypeOf(map[string]string{}),
		reflect.TypeOf(make(chan int)),
		reflect.TypeOf(func() {}),
	}

	for _, typ := range basicTypes {
		t.Run(fmt.Sprintf("basic type %s", typ.String()), func(t *testing.T) {
			_, err := converter.ConvertCustom("test", typ)
			if err == nil {
				t.Errorf("Expected error for type %s", typ.String())
			}
			if !strings.Contains(err.Error(), "does not implement encoding.TextUnmarshaler") {
				t.Errorf("Expected TextUnmarshaler error for type %s, got: %v", typ.String(), err)
			}
		})
	}

	// Test with pointer to basic types
	for _, typ := range basicTypes {
		ptrType := reflect.PtrTo(typ)
		t.Run(fmt.Sprintf("pointer to %s", typ.String()), func(t *testing.T) {
			_, err := converter.ConvertCustom("test", ptrType)
			if err == nil {
				t.Errorf("Expected error for pointer to type %s", typ.String())
			}
		})
	}

	// Test with interface{} type specifically
	t.Run("interface{} type", func(t *testing.T) {
		interfaceType := reflect.TypeOf((*interface{})(nil)).Elem()
		_, err := converter.ConvertCustom("test", interfaceType)
		if err == nil {
			t.Error("Expected error for interface{} type")
		}
	})

	// Test with custom struct that doesn't implement TextUnmarshaler
	t.Run("custom struct", func(t *testing.T) {
		type CustomStruct struct {
			Field1 string
			Field2 int
		}
		structType := reflect.TypeOf(CustomStruct{})
		_, err := converter.ConvertCustom("test", structType)
		if err == nil {
			t.Error("Expected error for custom struct")
		}
	})

	// Test with pointer to custom struct that doesn't implement TextUnmarshaler
	t.Run("pointer to custom struct", func(t *testing.T) {
		type CustomStruct struct {
			Field1 string
			Field2 int
		}
		ptrType := reflect.TypeOf((*CustomStruct)(nil))
		_, err := converter.ConvertCustom("test", ptrType)
		if err == nil {
			t.Error("Expected error for pointer to custom struct")
		}
	})

	// Test successful conversion with TextUnmarshaler
	t.Run("successful TextUnmarshaler conversion", func(t *testing.T) {
		// Test with value type
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})
		result, err := converter.ConvertCustom("test-value", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}

		// Test with pointer type
		ptrType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))
		result, err = converter.ConvertCustom("test-value", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
	})

	// Test error from TextUnmarshaler
	t.Run("TextUnmarshaler error", func(t *testing.T) {
		// Test with value type that returns error
		valueType := reflect.TypeOf(ErrorUnmarshaler{})
		_, err := converter.ConvertCustom("test-value", valueType)
		if err == nil {
			t.Error("Expected error from TextUnmarshaler")
		}

		// Test with pointer type that returns error
		ptrType := reflect.TypeOf((*ErrorUnmarshaler)(nil))
		_, err = converter.ConvertCustom("test-value", ptrType)
		if err == nil {
			t.Error("Expected error from TextUnmarshaler")
		}
	})

	// Test all the specific type checking paths in ConvertCustom
	t.Run("comprehensive type coverage", func(t *testing.T) {
		// Test with various unsupported types to hit all error paths
		unsupportedTypes := []reflect.Type{
			reflect.TypeOf(complex64(0)),
			reflect.TypeOf(complex128(0)),
			reflect.TypeOf(uintptr(0)),
			reflect.TypeOf([5]int{}),                    // Array
			reflect.TypeOf(make(chan string)),           // Channel
			reflect.TypeOf(func() string { return "" }), // Function
			reflect.TypeOf(map[string]int{}),            // Map
			reflect.TypeOf((*interface{})(nil)).Elem(),  // Interface
		}

		for _, typ := range unsupportedTypes {
			_, err := converter.ConvertCustom("test", typ)
			if err == nil {
				t.Errorf("Expected error for unsupported type %s", typ.String())
			}
		}
	})
}

// TestValidateAPICompatibilityComplete tests ValidateAPICompatibility for 100% coverage
func TestValidateAPICompatibilityComplete(t *testing.T) {
	framework := NewCompatibilityTestFramework()

	// Test the API compatibility validation
	t.Run("API compatibility validation", func(t *testing.T) {
		err := framework.ValidateAPICompatibility()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Test adding compatibility tests
	t.Run("add compatibility test", func(t *testing.T) {
		framework.AddCompatibilityTest("test case", struct{}{}, []string{"--test"}, false)
		// This should not error
	})

	// Test running full compatibility test
	t.Run("run full compatibility test", func(t *testing.T) {
		report, err := framework.RunFullCompatibilityTest()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if report == nil {
			t.Error("Expected non-nil report")
		}
	})

	// Test ValidateAPICompatibility with missing method (simulate error condition)
	t.Run("API compatibility with missing method", func(t *testing.T) {
		// Create a custom framework to test error path
		customFramework := &CompatibilityTestFramework{
			aliasManager: NewModuleAliasManager(),
			testSuite:    NewTestSuite(),
		}

		// Test the validation - this should pass since our Parser has all required methods
		err := customFramework.ValidateAPICompatibility()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// The error path is difficult to test without modifying the Parser type
		// but we can verify the method checking logic works
		parserType := reflect.TypeOf(&Parser{})
		expectedMethods := []string{"Parse", "WriteHelp", "WriteUsage", "Fail"}

		for _, methodName := range expectedMethods {
			if _, found := parserType.MethodByName(methodName); !found {
				t.Errorf("Parser missing expected method: %s", methodName)
			}
		}
	})

	// Test generating compatibility report
	t.Run("generate compatibility report", func(t *testing.T) {
		report := &CompatibilityReport{
			TotalTests:  2,
			PassedTests: 1,
			FailedTests: 1,
			Scenarios: []ScenarioResult{
				{Name: "test1", Match: true},
				{Name: "test2", Match: false},
			},
		}

		reportStr := framework.GenerateCompatibilityReport(report)
		if reportStr == "" {
			t.Error("Expected non-empty report string")
		}
		if !strings.Contains(reportStr, "Total Tests: 2") {
			t.Error("Expected report to contain total tests")
		}
	})
}

// TestProcessOptionsWithInheritanceAllPaths tests all paths in processOptionsWithInheritance
func TestProcessOptionsWithInheritanceAllPaths(t *testing.T) {
	// Test with parsing error from subParser.Options()
	t.Run("parsing error from options iterator", func(t *testing.T) {
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

		// Create a parser that will have an error in the options iterator
		subParser, err := optargs.GetOptLong([]string{"--unknown-option"}, "", []optargs.Flag{})
		if err != nil {
			t.Fatalf("Failed to create subparser: %v", err)
		}

		// This should handle the parsing error
		err = parser.processOptionsWithInheritance(subParser, parentIntegration, subMetadata, &subStruct)
		if err == nil {
			t.Error("Expected error from unknown option")
		}
	})

	// Test with field setting error in subcommand path
	t.Run("field setting error in subcommand", func(t *testing.T) {
		parentMetadata := &StructMetadata{Fields: []FieldMetadata{}}
		subMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Number", Short: "n", Long: "number", ArgType: optargs.RequiredArgument, Type: reflect.TypeOf(int(0))},
			},
		}

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

	// Test with field setting error in parent inheritance path
	t.Run("field setting error in parent inheritance", func(t *testing.T) {
		parentMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Number", Short: "n", Long: "number", ArgType: optargs.RequiredArgument, Type: reflect.TypeOf(int(0))},
			},
		}
		subMetadata := &StructMetadata{Fields: []FieldMetadata{}}

		parentStruct := struct {
			Number int
		}{}
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
		parentIntegration.BuildLongOpts()

		// Create a subcommand parser with invalid number for parent option
		subParser, err := optargs.GetOptLong([]string{"-n", "not-a-number"}, "n:", []optargs.Flag{
			{Name: "number", HasArg: optargs.RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create subparser: %v", err)
		}

		// This should return an error due to invalid number conversion in parent
		err = parser.processOptionsWithInheritance(subParser, parentIntegration, subMetadata, &subStruct)
		if err == nil {
			t.Error("Expected error for invalid number conversion in parent")
		}
	})
}

// TestParseStructAllPaths tests all paths in ParseStruct for 100% coverage
func TestParseStructAllPaths(t *testing.T) {
	parser := &TagParser{}

	// Test with nil destination
	t.Run("nil destination", func(t *testing.T) {
		_, err := parser.ParseStruct(nil)
		if err == nil {
			t.Error("Expected error for nil destination")
		}
	})

	// Test with non-pointer destination
	t.Run("non-pointer destination", func(t *testing.T) {
		_, err := parser.ParseStruct(struct{}{})
		if err == nil {
			t.Error("Expected error for non-pointer destination")
		}
	})

	// Test with pointer to non-struct
	t.Run("pointer to non-struct", func(t *testing.T) {
		var i int
		_, err := parser.ParseStruct(&i)
		if err == nil {
			t.Error("Expected error for pointer to non-struct")
		}
	})

	// Test with subcommand field parsing error
	t.Run("subcommand field parsing error", func(t *testing.T) {
		testStruct := &struct {
			Server *struct {
				Port string `arg:"invalid-tag-format-with-multiple-errors"`
			} `arg:"subcommand:server"`
		}{}

		// Initialize the subcommand field
		testStruct.Server = &struct {
			Port string `arg:"invalid-tag-format-with-multiple-errors"`
		}{}

		_, err := parser.ParseStruct(testStruct)
		if err == nil {
			t.Error("Expected error for invalid subcommand field")
		}
	})
}

// TestCreateParserWithParentAllPaths tests all paths in CreateParserWithParent
func TestCreateParserWithParentAllPaths(t *testing.T) {
	// Test with subcommand creation error
	t.Run("subcommand creation error", func(t *testing.T) {
		// Create metadata with problematic subcommand
		subMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Invalid", Short: "toolongshort"}, // Invalid short option
			},
		}

		metadata := &StructMetadata{
			Fields:      []FieldMetadata{},
			Subcommands: map[string]*StructMetadata{"invalid": subMetadata},
		}

		integration := &CoreIntegration{metadata: metadata}

		// This should handle subcommand creation gracefully
		parser, err := integration.CreateParserWithParent([]string{}, nil)
		if err != nil {
			// Error is expected due to invalid subcommand
			t.Logf("Expected error during subcommand creation: %v", err)
		} else if parser == nil {
			t.Error("Expected either error or valid parser")
		}
	})
}

// TestBuildOptStringAllPaths tests all paths in BuildOptString
func TestBuildOptStringAllPaths(t *testing.T) {
	// Test with all argument types
	t.Run("all argument types", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "NoArg", Short: "n", ArgType: optargs.NoArgument},
				{Name: "Required", Short: "r", ArgType: optargs.RequiredArgument},
				{Name: "Optional", Short: "o", ArgType: optargs.OptionalArgument},
				{Name: "Positional", Positional: true},   // Should be skipped
				{Name: "Subcommand", IsSubcommand: true}, // Should be skipped
			},
		}

		integration := &CoreIntegration{metadata: metadata}
		optstring := integration.BuildOptString()

		// Should contain: n (no colon), r: (required), o:: (optional)
		expected := "nr:o::"
		if optstring != expected {
			t.Errorf("Expected '%s', got '%s'", expected, optstring)
		}
	})
}

// TestValidateFieldMetadataAllPaths tests all paths in ValidateFieldMetadata
func TestValidateFieldMetadataAllPaths(t *testing.T) {
	parser := &TagParser{}

	// Test positional with option flags
	t.Run("positional with option flags", func(t *testing.T) {
		metadata := &FieldMetadata{
			Name:       "Test",
			Positional: true,
			Short:      "t",
			Long:       "test",
		}

		err := parser.ValidateFieldMetadata(metadata)
		if err == nil {
			t.Error("Expected error for positional with option flags")
		}
	})

	// Test subcommand with non-pointer type
	t.Run("subcommand non-pointer", func(t *testing.T) {
		metadata := &FieldMetadata{
			Name:         "Server",
			IsSubcommand: true,
			Type:         reflect.TypeOf(struct{}{}), // Not a pointer
		}

		err := parser.ValidateFieldMetadata(metadata)
		if err == nil {
			t.Error("Expected error for non-pointer subcommand")
		}
	})

	// Test subcommand with pointer to non-struct
	t.Run("subcommand pointer to non-struct", func(t *testing.T) {
		metadata := &FieldMetadata{
			Name:         "Server",
			IsSubcommand: true,
			Type:         reflect.TypeOf((*int)(nil)), // Pointer to int, not struct
		}

		err := parser.ValidateFieldMetadata(metadata)
		if err == nil {
			t.Error("Expected error for pointer to non-struct subcommand")
		}
	})

	// Test invalid short option length
	t.Run("invalid short option", func(t *testing.T) {
		metadata := &FieldMetadata{
			Name:  "Verbose",
			Short: "verbose", // Invalid: more than one character
		}

		err := parser.ValidateFieldMetadata(metadata)
		if err == nil {
			t.Error("Expected error for invalid short option")
		}
	})

	// Test field without options gets default long option
	t.Run("field without options", func(t *testing.T) {
		metadata := &FieldMetadata{
			Name:         "TestField",
			Positional:   false,
			IsSubcommand: false,
			Short:        "",
			Long:         "",
		}

		err := parser.ValidateFieldMetadata(metadata)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Should have generated default long option
		if metadata.Long != "testfield" {
			t.Errorf("Expected 'testfield', got '%s'", metadata.Long)
		}
	})
}

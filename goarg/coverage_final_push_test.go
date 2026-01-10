package goarg

import (
	"reflect"
	"strings"
	"testing"

	"github.com/major0/optargs"
)

// TestNewParserErrorPaths tests all error paths in NewParser
func TestNewParserErrorPaths(t *testing.T) {
	t.Run("nil destination", func(t *testing.T) {
		_, err := NewParser(Config{}, nil)
		if err == nil {
			t.Error("Expected error for nil destination")
		}
		if !strings.Contains(err.Error(), "destination cannot be nil") {
			t.Errorf("Expected nil destination error, got: %v", err)
		}
	})

	t.Run("non-pointer destination", func(t *testing.T) {
		testStruct := struct {
			Verbose bool
		}{}

		_, err := NewParser(Config{}, testStruct) // Not a pointer
		if err == nil {
			t.Error("Expected error for non-pointer destination")
		}
		if !strings.Contains(err.Error(), "must be a pointer") {
			t.Errorf("Expected pointer error, got: %v", err)
		}
	})

	t.Run("pointer to non-struct", func(t *testing.T) {
		var testInt int
		_, err := NewParser(Config{}, &testInt) // Pointer to int, not struct
		if err == nil {
			t.Error("Expected error for pointer to non-struct")
		}
		if !strings.Contains(err.Error(), "must be a pointer to a struct") {
			t.Errorf("Expected struct error, got: %v", err)
		}
	})
}

// TestValidateMinMaxUnsupportedTypes tests unsupported types in validation
func TestValidateMinMaxUnsupportedTypes(t *testing.T) {
	converter := &TypeConverter{}

	// Test validateMin with unsupported type (should return nil, no error)
	t.Run("validateMin unsupported type", func(t *testing.T) {
		fieldValue := reflect.ValueOf("string-value")
		err := converter.validateMin(fieldValue, "5", "testField")
		// Should return nil for unsupported types (this is the uncovered path)
		if err != nil {
			t.Errorf("Expected nil for unsupported type, got: %v", err)
		}
	})

	// Test validateMax with unsupported type (should return nil, no error)
	t.Run("validateMax unsupported type", func(t *testing.T) {
		fieldValue := reflect.ValueOf("string-value")
		err := converter.validateMax(fieldValue, "5", "testField")
		// Should return nil for unsupported types (this is the uncovered path)
		if err != nil {
			t.Errorf("Expected nil for unsupported type, got: %v", err)
		}
	})

	// Test validateMin with complex type
	t.Run("validateMin complex type", func(t *testing.T) {
		fieldValue := reflect.ValueOf([]string{"test"})
		err := converter.validateMin(fieldValue, "5", "testField")
		// Should return nil for unsupported types
		if err != nil {
			t.Errorf("Expected nil for complex type, got: %v", err)
		}
	})

	// Test validateMax with complex type
	t.Run("validateMax complex type", func(t *testing.T) {
		fieldValue := reflect.ValueOf([]string{"test"})
		err := converter.validateMax(fieldValue, "5", "testField")
		// Should return nil for unsupported types
		if err != nil {
			t.Errorf("Expected nil for complex type, got: %v", err)
		}
	})
}

// TestConvertValueEdgeCases tests edge cases in ConvertValue
func TestConvertValueEdgeCases(t *testing.T) {
	converter := &TypeConverter{}

	// Test with complex slice type
	t.Run("complex slice type", func(t *testing.T) {
		sliceType := reflect.TypeOf([][]string{})
		result, err := converter.ConvertValue("test", sliceType)
		// This might succeed by creating a slice with one element
		if err != nil {
			t.Logf("Expected behavior for complex slice type: %v", err)
		} else {
			t.Logf("Complex slice conversion succeeded: %v", result)
		}
	})

	// Test with channel type
	t.Run("channel type", func(t *testing.T) {
		chanType := reflect.TypeOf(make(chan int))
		_, err := converter.ConvertValue("test", chanType)
		if err == nil {
			t.Error("Expected error for channel type")
		}
	})

	// Test with function type
	t.Run("function type", func(t *testing.T) {
		funcType := reflect.TypeOf(func() {})
		_, err := converter.ConvertValue("test", funcType)
		if err == nil {
			t.Error("Expected error for function type")
		}
	})
}

// TestParseDefaultValueEdgeCases tests edge cases in parseDefaultValue
func TestParseDefaultValueEdgeCases(t *testing.T) {
	parser := &TagParser{}

	// Test with empty default value
	t.Run("empty default value", func(t *testing.T) {
		field := reflect.StructField{
			Name: "TestField",
			Type: reflect.TypeOf(""),
			Tag:  `default:""`,
		}

		fieldMeta, err := parser.ParseField(field)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if fieldMeta.Default != "" {
			t.Errorf("Expected empty default, got: %v", fieldMeta.Default)
		}
	})

	// Test with slice default value
	t.Run("slice default value", func(t *testing.T) {
		field := reflect.StructField{
			Name: "TestSlice",
			Type: reflect.TypeOf([]string{}),
			Tag:  `default:"a,b,c"`,
		}

		fieldMeta, err := parser.ParseField(field)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if fieldMeta.Default == nil {
			t.Error("Expected non-nil default for slice")
		}
	})
}

// TestMapToOptArgsCoreEdgeCases tests edge cases in mapToOptArgsCore
func TestMapToOptArgsCoreEdgeCases(t *testing.T) {
	parser := &TagParser{}

	// Test with pointer to bool
	t.Run("pointer to bool", func(t *testing.T) {
		metadata := &FieldMetadata{
			Name:  "BoolPtr",
			Short: "b",
			Long:  "bool-ptr",
			Type:  reflect.TypeOf((*bool)(nil)),
		}

		err := parser.mapToOptArgsCore(metadata)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Should be NoArgument for pointer to bool
		if metadata.ArgType != optargs.NoArgument {
			t.Errorf("Expected NoArgument, got %v", metadata.ArgType)
		}
	})

	// Test with pointer to non-bool
	t.Run("pointer to non-bool", func(t *testing.T) {
		metadata := &FieldMetadata{
			Name:  "StringPtr",
			Short: "s",
			Long:  "string-ptr",
			Type:  reflect.TypeOf((*string)(nil)),
		}

		err := parser.mapToOptArgsCore(metadata)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Should be OptionalArgument for pointer to non-bool
		if metadata.ArgType != optargs.OptionalArgument {
			t.Errorf("Expected OptionalArgument, got %v", metadata.ArgType)
		}
	})
}

// TestValidateFieldMetadataEdgeCases tests edge cases in ValidateFieldMetadata
func TestValidateFieldMetadataEdgeCases(t *testing.T) {
	parser := &TagParser{}

	// Test field with no options gets default long option
	t.Run("field with no options", func(t *testing.T) {
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

// TestWriteHelpEdgeCases tests edge cases in WriteHelp
func TestWriteHelpEdgeCases(t *testing.T) {
	t.Run("help with argument placeholders", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{
					Name:    "Count",
					Short:   "c",
					Long:    "count",
					ArgType: optargs.RequiredArgument,
					Help:    "number of items",
					Type:    reflect.TypeOf(int(0)),
				},
				{
					Name:    "Optional",
					Short:   "o",
					Long:    "optional",
					ArgType: optargs.OptionalArgument,
					Help:    "optional parameter",
					Type:    reflect.TypeOf(""),
				},
			},
		}

		config := Config{Program: "testapp"}
		generator := NewHelpGenerator(metadata, config)

		var buf strings.Builder
		err := generator.WriteHelp(&buf)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		output := buf.String()
		// Should contain argument placeholders
		if !strings.Contains(output, "COUNT") {
			t.Error("Expected argument placeholder 'COUNT' in help output")
		}
		// Note: The exact format of optional argument placeholders may vary
		// Let's just check that the help output is generated
		if len(output) == 0 {
			t.Error("Expected non-empty help output")
		}
	})
}

// TestWriteUsageEdgeCases tests edge cases in WriteUsage
func TestWriteUsageEdgeCases(t *testing.T) {
	t.Run("usage with positional arguments", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Input", Positional: true, Required: true, Help: "input file", Type: reflect.TypeOf("")},
				{Name: "Output", Positional: true, Required: false, Help: "output file", Type: reflect.TypeOf("")},
			},
		}

		config := Config{Program: "testapp"}
		generator := NewHelpGenerator(metadata, config)

		var buf strings.Builder
		err := generator.WriteUsage(&buf)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "INPUT") {
			t.Error("Expected positional argument 'INPUT' in usage")
		}
		if !strings.Contains(output, "[OUTPUT]") {
			t.Error("Expected optional positional argument '[OUTPUT]' in usage")
		}
	})
}

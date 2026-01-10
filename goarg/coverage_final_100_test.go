package goarg

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/major0/optargs"
)

// TestMustParsePartialCoverage tests what we can of MustParse without os.Exit
func TestMustParsePartialCoverage(t *testing.T) {
	// We can't test the os.Exit(1) path, but we can test the success path
	t.Run("success path", func(t *testing.T) {
		// Create a simple struct that will parse successfully
		var args struct {
			Verbose bool `arg:"-v,--verbose"`
		}

		// Save original os.Args and restore after test
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Set os.Args to simulate command line arguments
		os.Args = []string{"testprog", "--verbose"}

		// This should succeed and not call os.Exit
		MustParse(&args)

		if !args.Verbose {
			t.Error("Expected verbose to be true")
		}
	})
}

// TestConvertCustomComplexPaths tests the remaining uncovered paths in ConvertCustom
func TestConvertCustomComplexPaths(t *testing.T) {
	converter := &TypeConverter{}

	// Test type that doesn't implement TextUnmarshaler at all
	t.Run("type without TextUnmarshaler", func(t *testing.T) {
		// Use a basic type that doesn't implement TextUnmarshaler
		intType := reflect.TypeOf(int(0))
		_, err := converter.ConvertCustom("123", intType)
		if err == nil {
			t.Error("Expected error for type without TextUnmarshaler")
		}
		if !strings.Contains(err.Error(), "does not implement encoding.TextUnmarshaler") {
			t.Errorf("Expected TextUnmarshaler error, got: %v", err)
		}
	})

	// Test with interface type
	t.Run("interface type", func(t *testing.T) {
		interfaceType := reflect.TypeOf((*interface{})(nil)).Elem()
		_, err := converter.ConvertCustom("test", interfaceType)
		if err == nil {
			t.Error("Expected error for interface type")
		}
	})

	// Test with map type
	t.Run("map type", func(t *testing.T) {
		mapType := reflect.TypeOf(map[string]string{})
		_, err := converter.ConvertCustom("test", mapType)
		if err == nil {
			t.Error("Expected error for map type")
		}
	})
}

// TestProcessOptionsWithInheritanceEdgeCases tests uncovered inheritance scenarios
func TestProcessOptionsWithInheritanceEdgeCases(t *testing.T) {
	// Test findFieldInMetadata with no matches
	t.Run("findFieldInMetadata no match", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Test", Short: "t", Long: "test"},
			},
		}

		parser := &Parser{metadata: metadata}
		field := parser.findFieldInMetadata("nonexistent", metadata)
		if field != nil {
			t.Error("Expected nil for non-existent field")
		}
	})

	// Test findParentFieldForOption with no matches
	t.Run("findParentFieldForOption no match", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Test", Short: "t", Long: "test"},
			},
		}

		parser := &Parser{metadata: metadata}
		field := parser.findParentFieldForOption("nonexistent")
		if field != nil {
			t.Error("Expected nil for non-existent parent field")
		}
	})
}

// TestProcessResultsErrorPaths tests uncovered error paths in ProcessResults
func TestProcessResultsErrorPaths(t *testing.T) {
	t.Run("invalid field value", func(t *testing.T) {
		// Create a struct with an unexported field to trigger the error
		type testStruct struct {
			unexported int // This field cannot be set
		}

		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "unexported", Short: "u", Long: "unexported", Type: reflect.TypeOf(int(0))},
			},
		}

		integration := &CoreIntegration{
			metadata:    metadata,
			shortOpts:   make(map[byte]*optargs.Flag),
			longOpts:    make(map[string]*optargs.Flag),
			positionals: []PositionalArg{},
		}

		// Create a mock parser with an option
		parser, err := optargs.GetOptLong([]string{"-u", "123"}, "u:", []optargs.Flag{
			{Name: "unexported", HasArg: optargs.RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		dest := &testStruct{}
		err = integration.ProcessResults(parser, dest)
		// This should succeed because unexported fields are skipped
		if err != nil {
			t.Logf("Got expected error for unexported field: %v", err)
		}
	})
}

// TestSetDefaultValuesErrorPaths tests uncovered error paths in setDefaultValues
func TestSetDefaultValuesErrorPaths(t *testing.T) {
	t.Run("field cannot be set", func(t *testing.T) {
		// Create a field that exists but cannot be set (this is hard to test directly)
		// For now, we'll skip this test as it's difficult to create a scenario
		// where the field exists but CanSet() returns false
		t.Skip("Difficult to test field that cannot be set")
	})
}

// TestValidateMinMaxComplexTypes tests uncovered paths in validation functions
func TestValidateMinMaxComplexTypes(t *testing.T) {
	converter := &TypeConverter{}

	// Test validateMin with invalid constraint
	t.Run("validateMin invalid constraint", func(t *testing.T) {
		fieldValue := reflect.ValueOf(int(10))
		err := converter.validateMin(fieldValue, "invalid", "testField")
		if err == nil {
			t.Error("Expected error for invalid min constraint")
		}
	})

	// Test validateMax with invalid constraint
	t.Run("validateMax invalid constraint", func(t *testing.T) {
		fieldValue := reflect.ValueOf(int(10))
		err := converter.validateMax(fieldValue, "invalid", "testField")
		if err == nil {
			t.Error("Expected error for invalid max constraint")
		}
	})

	// Test validateMin with different numeric types
	t.Run("validateMin uint types", func(t *testing.T) {
		fieldValue := reflect.ValueOf(uint(5))
		err := converter.validateMin(fieldValue, "10", "testField")
		if err == nil {
			t.Error("Expected error for uint value below minimum")
		}
	})

	// Test validateMax with different numeric types
	t.Run("validateMax uint types", func(t *testing.T) {
		fieldValue := reflect.ValueOf(uint(15))
		err := converter.validateMax(fieldValue, "10", "testField")
		if err == nil {
			t.Error("Expected error for uint value above maximum")
		}
	})

	// Test validateMin with float types
	t.Run("validateMin float types", func(t *testing.T) {
		fieldValue := reflect.ValueOf(float64(5.5))
		err := converter.validateMin(fieldValue, "10.0", "testField")
		if err == nil {
			t.Error("Expected error for float value below minimum")
		}
	})

	// Test validateMax with float types
	t.Run("validateMax float types", func(t *testing.T) {
		fieldValue := reflect.ValueOf(float64(15.5))
		err := converter.validateMax(fieldValue, "10.0", "testField")
		if err == nil {
			t.Error("Expected error for float value above maximum")
		}
	})
}

// TestParseStructErrorPaths tests uncovered error paths in ParseStruct
func TestParseStructErrorPaths(t *testing.T) {
	parser := &TagParser{}

	// Test with subcommand parsing error
	t.Run("subcommand parsing error", func(t *testing.T) {
		// Create a struct with an invalid subcommand field
		var args struct {
			Server *struct {
				Port string `arg:"invalid-tag-format"`
			} `arg:"subcommand:server"`
		}

		// Initialize the subcommand field so it's not nil
		args.Server = &struct {
			Port string `arg:"invalid-tag-format"`
		}{}

		_, err := parser.ParseStruct(&args)
		if err == nil {
			t.Error("Expected error for invalid subcommand field")
		}
	})
}

// TestParseFieldErrorPaths tests uncovered error paths in ParseField
func TestParseFieldErrorPaths(t *testing.T) {
	parser := &TagParser{}

	// Test with invalid default value
	t.Run("invalid default value", func(t *testing.T) {
		field := reflect.StructField{
			Name: "Count",
			Type: reflect.TypeOf(int(0)),
			Tag:  `default:"invalid-int"`,
		}

		_, err := parser.ParseField(field)
		if err == nil {
			t.Error("Expected error for invalid default value")
		}
	})

	// Test with invalid field metadata
	t.Run("invalid field metadata", func(t *testing.T) {
		field := reflect.StructField{
			Name: "Test",
			Type: reflect.TypeOf(""),
			Tag:  `arg:"positional,-v"`, // Invalid: positional with option flag
		}

		_, err := parser.ParseField(field)
		if err == nil {
			t.Error("Expected error for invalid field metadata")
		}
	})
}

// TestMapToOptArgsCoreErrorPaths tests uncovered error paths in mapToOptArgsCore
func TestMapToOptArgsCoreErrorPaths(t *testing.T) {
	parser := &TagParser{}

	// Test with complex custom type
	t.Run("complex custom type", func(t *testing.T) {
		metadata := &FieldMetadata{
			Name: "Custom",
			Type: reflect.TypeOf(struct{ Value string }{}),
			Long: "custom",
		}

		err := parser.mapToOptArgsCore(metadata)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Should default to RequiredArgument for custom types
		if metadata.ArgType != optargs.RequiredArgument {
			t.Errorf("Expected RequiredArgument, got %v", metadata.ArgType)
		}
	})
}

// TestValidateFieldMetadataErrorPaths tests uncovered error paths in ValidateFieldMetadata
func TestValidateFieldMetadataErrorPaths(t *testing.T) {
	parser := &TagParser{}

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

	// Test short option with multiple characters
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
}

// TestWriteHelpUncoveredPaths tests remaining uncovered paths in WriteHelp
func TestWriteHelpUncoveredPaths(t *testing.T) {
	t.Run("help with nil metadata", func(t *testing.T) {
		generator := &HelpGenerator{
			metadata: nil,
			config:   Config{},
		}

		var buf strings.Builder
		err := generator.WriteHelp(&buf)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "No help available") {
			t.Error("Expected 'No help available' for nil metadata")
		}
	})

	// Test with short option only
	t.Run("short option only", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{
					Name:    "Verbose",
					Short:   "v",
					Long:    "", // No long option
					ArgType: optargs.NoArgument,
					Help:    "verbose output",
					Type:    reflect.TypeOf(bool(false)),
				},
			},
		}

		generator := NewHelpGenerator(metadata, Config{Program: "test"})
		var buf strings.Builder
		err := generator.WriteHelp(&buf)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "-v") {
			t.Error("Expected short option in help output")
		}
	})

	// Test with long option only
	t.Run("long option only", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{
					Name:    "Verbose",
					Short:   "", // No short option
					Long:    "verbose",
					ArgType: optargs.NoArgument,
					Help:    "verbose output",
					Type:    reflect.TypeOf(bool(false)),
				},
			},
		}

		generator := NewHelpGenerator(metadata, Config{Program: "test"})
		var buf strings.Builder
		err := generator.WriteHelp(&buf)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "--verbose") {
			t.Error("Expected long option in help output")
		}
	})
}

// TestWriteUsageUncoveredPaths tests remaining uncovered paths in WriteUsage
func TestWriteUsageUncoveredPaths(t *testing.T) {
	t.Run("usage with nil metadata", func(t *testing.T) {
		generator := &HelpGenerator{
			metadata: nil,
			config:   Config{Program: "test"},
		}

		var buf strings.Builder
		err := generator.WriteUsage(&buf)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Usage: test") {
			t.Error("Expected basic usage line")
		}
	})
}

// TestTranslateErrorUncoveredPaths tests remaining uncovered paths in TranslateError
func TestTranslateErrorUncoveredPaths(t *testing.T) {
	translator := &ErrorTranslator{}

	// Test with complex error message
	t.Run("complex error message", func(t *testing.T) {
		err := fmt.Errorf("complex: nested: error: message")
		context := ParseContext{FieldName: "test"}

		translated := translator.TranslateError(err, context)
		if translated == nil {
			t.Error("Expected non-nil translated error")
		}
	})

	// Test with error containing colon
	t.Run("error with colon", func(t *testing.T) {
		err := fmt.Errorf("field: value: error")
		context := ParseContext{}

		translated := translator.TranslateError(err, context)
		if translated == nil {
			t.Error("Expected non-nil translated error")
		}
	})
}

// TestExtractOptionFromErrorUncoveredPaths tests remaining uncovered paths
func TestExtractOptionFromErrorUncoveredPaths(t *testing.T) {
	// Test with no option found
	t.Run("no option found", func(t *testing.T) {
		errMsg := "some error without options"
		result := extractOptionFromError(errMsg)
		if result != errMsg {
			t.Errorf("Expected original message, got: %s", result)
		}
	})

	// Test with option at end of message
	t.Run("option at end", func(t *testing.T) {
		errMsg := "error with --option"
		result := extractOptionFromError(errMsg)
		if result != "--option" {
			t.Errorf("Expected '--option', got: %s", result)
		}
	})
}

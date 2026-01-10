package goarg

import (
	"os"
	"reflect"
	"testing"

	"github.com/major0/optargs"
)

// TestConvertCustomSpecificPaths tests the specific uncovered paths in ConvertCustom
func TestConvertCustomSpecificPaths(t *testing.T) {
	converter := &TypeConverter{}

	// Test the second path where ptrType implements TextUnmarshaler
	// and targetType.Kind() == reflect.Ptr (line ~533)
	t.Run("pointer type in second implementation path", func(t *testing.T) {
		// Use our TestCustomTypeForCoverage which implements TextUnmarshaler on *T
		// Test with pointer type to trigger the second path
		ptrType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))

		result, err := converter.ConvertCustom("test-value", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}

		// This should return the pointer interface (line ~533)
		if customResult, ok := result.(*TestCustomTypeForCoverage); ok {
			if customResult.Value != "test-value" {
				t.Errorf("Expected 'test-value', got '%s'", customResult.Value)
			}
		} else {
			t.Errorf("Expected *TestCustomTypeForCoverage, got %T", result)
		}
	})

	// Test the second path where ptrType implements TextUnmarshaler
	// and targetType.Kind() != reflect.Ptr (line ~535)
	t.Run("non-pointer type in second implementation path", func(t *testing.T) {
		// Use our TestCustomTypeForCoverage which implements TextUnmarshaler on *T
		// Test with non-pointer type to trigger the second path
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})

		result, err := converter.ConvertCustom("test-value", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}

		// This should return the element interface (line ~535)
		if customResult, ok := result.(TestCustomTypeForCoverage); ok {
			if customResult.Value != "test-value" {
				t.Errorf("Expected 'test-value', got '%s'", customResult.Value)
			}
		} else {
			t.Errorf("Expected TestCustomTypeForCoverage, got %T", result)
		}
	})
}

// TestMustParseSuccessPath tests the success path of MustParse
func TestMustParseSuccessPath(t *testing.T) {
	// We can only test the success path since the error path calls os.Exit
	// This test ensures we get coverage on the success path
	t.Run("successful parse", func(t *testing.T) {
		// This test is already covered by existing tests, but we include it
		// to ensure the success path is covered
		testStruct := &struct {
			Verbose bool `arg:"-v,--verbose"`
		}{}

		// Save original args
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{"testprog", "--verbose"}

		// This should not panic
		MustParse(testStruct)

		if !testStruct.Verbose {
			t.Error("Expected Verbose to be true")
		}
	})
}

// TestProcessOptionsInheritanceSpecificPaths tests specific uncovered paths
func TestProcessOptionsInheritanceSpecificPaths(t *testing.T) {
	t.Run("option with HasArg true in subcommand path", func(t *testing.T) {
		// Create parent metadata
		parentMetadata := &StructMetadata{
			Fields: []FieldMetadata{},
		}

		// Create subcommand metadata with option that has argument
		subMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Config", Short: "c", Long: "config", ArgType: optargs.RequiredArgument, Type: reflect.TypeOf("")},
			},
		}

		// Create structs
		parentStruct := struct{}{}
		subStruct := struct {
			Config string
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

		// Create a subcommand parser with option that has argument
		subParser, err := optargs.GetOptLong([]string{"-c", "config.json"}, "c:", []optargs.Flag{
			{Name: "config", HasArg: optargs.RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create subparser: %v", err)
		}

		// This should process the option with its argument
		err = parser.processOptionsWithInheritance(subParser, parentIntegration, subMetadata, &subStruct)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if subStruct.Config != "config.json" {
			t.Errorf("Expected 'config.json', got '%s'", subStruct.Config)
		}
	})

	t.Run("option with HasArg true in parent inheritance path", func(t *testing.T) {
		// Create parent metadata with option that has argument
		parentMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Config", Short: "c", Long: "config", ArgType: optargs.RequiredArgument, Type: reflect.TypeOf("")},
			},
		}

		// Create subcommand metadata without the config option
		subMetadata := &StructMetadata{
			Fields: []FieldMetadata{},
		}

		// Create structs
		parentStruct := struct {
			Config string
		}{}
		subStruct := struct{}{}

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

		// Create a subcommand parser with parent option that has argument
		subParser, err := optargs.GetOptLong([]string{"-c", "config.json"}, "c:", []optargs.Flag{
			{Name: "config", HasArg: optargs.RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create subparser: %v", err)
		}

		// This should process the inherited parent option with its argument
		err = parser.processOptionsWithInheritance(subParser, parentIntegration, subMetadata, &subStruct)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if parentStruct.Config != "config.json" {
			t.Errorf("Expected 'config.json', got '%s'", parentStruct.Config)
		}
	})
}

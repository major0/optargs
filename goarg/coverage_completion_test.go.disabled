package goarg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/major0/optargs"
)

// TestParse tests the global Parse function
func TestParse(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	testCases := []struct {
		name        string
		args        []string
		testStruct  interface{}
		expectError bool
	}{
		{
			name: "successful parse",
			args: []string{"testprog", "--verbose", "--count", "42"},
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
				Count   int  `arg:"-c,--count"`
			}{},
			expectError: false,
		},
		{
			name: "parse with error",
			args: []string{"testprog", "--unknown"},
			testStruct: &struct {
				Verbose bool `arg:"-v,--verbose"`
			}{},
			expectError: true,
		},
		{
			name:        "nil destination",
			args:        []string{"testprog"},
			testStruct:  nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set os.Args for the test
			os.Args = tc.args

			err := Parse(tc.testStruct)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestMustParse tests the global MustParse function
func TestMustParse(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Test successful parse case
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

	// Note: We can't easily test the error cases of MustParse because it calls os.Exit(1)
	// which would terminate the test process. The function is simple enough that testing
	// the successful case provides adequate coverage for the MustParse wrapper.
}

// TestParserFail tests the Parser.Fail method
func TestParserFail(t *testing.T) {
	testStruct := &struct {
		Verbose bool `arg:"-v,--verbose"`
		Count   int  `arg:"-c,--count"`
	}{}

	parser, err := NewParser(Config{Program: "testapp"}, testStruct)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Capture stderr
	r, w, _ := os.Pipe()
	originalStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = originalStderr }()

	// Track if exit was called
	exitCalled := false
	exitCode := 0

	// Mock the exit function
	parser.config.Exit = func(code int) {
		exitCalled = true
		exitCode = code
		panic("exit called") // Use panic to stop execution
	}

	// Use recover to catch the panic from mocked exit
	defer func() {
		if r := recover(); r != nil {
			if r != "exit called" {
				panic(r) // Re-panic if it's not our expected panic
			}
		}
	}()

	// Test the Fail method
	parser.Fail("test error message")

	// Close write end and read stderr
	w.Close()
	var stderrBuf bytes.Buffer
	io.Copy(&stderrBuf, r)

	// Verify exit was called
	if !exitCalled {
		t.Error("Expected Fail to call exit function")
	}

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Verify error message and usage were written to stderr
	stderrOutput := stderrBuf.String()
	if !strings.Contains(stderrOutput, "test error message") {
		t.Error("Expected error message in stderr output")
	}

	if !strings.Contains(stderrOutput, "Usage:") {
		t.Error("Expected usage information in stderr output")
	}
}

// TestGenerateCompatibilityReport tests the uncovered function
func TestGenerateCompatibilityReport(t *testing.T) {
	framework := NewCompatibilityTestFramework()

	// Add a simple test
	testStruct := &struct {
		Verbose bool `arg:"-v,--verbose"`
	}{}
	framework.AddCompatibilityTest("test", testStruct, []string{"--verbose"}, false)

	// Create a mock report
	report := &CompatibilityReport{
		TotalTests:  1,
		PassedTests: 1,
		FailedTests: 0,
	}

	// Test the report generation
	reportStr := framework.GenerateCompatibilityReport(report)

	// The function should return some kind of report
	if reportStr == "" {
		t.Error("Expected non-empty compatibility report")
	}

	if !strings.Contains(reportStr, "Compatibility Test Report") {
		t.Error("Expected report header in output")
	}
}

// TestUncoveredHelpFunctions tests functions with low coverage in help.go
func TestUncoveredHelpFunctions(t *testing.T) {
	t.Run("WriteHelp edge cases", func(t *testing.T) {
		// Test with nil metadata
		generator := &HelpGenerator{metadata: nil, config: Config{Program: "test"}}

		var buf bytes.Buffer
		err := generator.WriteHelp(&buf)
		if err != nil {
			t.Errorf("WriteHelp should not error with nil metadata: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "No help available") {
			t.Error("Expected 'No help available' message for nil metadata")
		}
	})

	t.Run("TranslateError edge cases", func(t *testing.T) {
		translator := &ErrorTranslator{}

		// Test with nil error
		result := translator.TranslateError(nil, ParseContext{})
		if result != nil {
			t.Error("Expected nil result for nil error")
		}

		// Test with various error message formats
		testErrors := []struct {
			input   error
			context ParseContext
		}{
			{
				input:   fmt.Errorf("parsing error: unknown option: test"),
				context: ParseContext{},
			},
			{
				input:   fmt.Errorf("option requires an argument: test"),
				context: ParseContext{},
			},
			{
				input:   fmt.Errorf("some other error"),
				context: ParseContext{},
			},
		}

		for _, tc := range testErrors {
			result := translator.TranslateError(tc.input, tc.context)
			if result == nil {
				t.Errorf("Expected non-nil result for error: %v", tc.input)
			}
		}
	})
}

// TestUncoveredTypesFunctions tests functions with low coverage in types.go
func TestUncoveredTypesFunctions(t *testing.T) {
	converter := &TypeConverter{}

	t.Run("ConvertCustom edge cases", func(t *testing.T) {
		// Test with non-unmarshaler type
		_, err := converter.ConvertCustom("123", reflect.TypeOf(int(0)))
		if err == nil {
			t.Error("Expected error for non-unmarshaler type")
		}

		// Test with pointer to non-unmarshaler type
		intType := reflect.TypeOf((*int)(nil))
		_, err = converter.ConvertCustom("123", intType)
		if err == nil {
			t.Error("Expected error for pointer to non-unmarshaler type")
		}

		// Test with interface type that doesn't implement TextUnmarshaler
		interfaceType := reflect.TypeOf((*interface{})(nil)).Elem()
		_, err = converter.ConvertCustom("123", interfaceType)
		if err == nil {
			t.Error("Expected error for interface type")
		}
	})
}

// TestParseArgsEdgeCases tests ParseArgs function edge cases
func TestParseArgsEdgeCases(t *testing.T) {
	t.Run("ParseArgs with nil destination", func(t *testing.T) {
		err := ParseArgs(nil, []string{"--verbose"})
		if err == nil {
			t.Error("Expected error for nil destination")
		}
	})

	t.Run("ParseArgs with invalid struct", func(t *testing.T) {
		err := ParseArgs("not a struct", []string{"--verbose"})
		if err == nil {
			t.Error("Expected error for non-struct destination")
		}
	})

	t.Run("ParseArgs successful", func(t *testing.T) {
		testStruct := &struct {
			Verbose bool `arg:"-v,--verbose"`
		}{}

		err := ParseArgs(testStruct, []string{"--verbose"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !testStruct.Verbose {
			t.Error("Expected Verbose to be true")
		}
	})
}

// TestCoreIntegrationEdgeCases tests core integration functions with low coverage
func TestCoreIntegrationEdgeCases(t *testing.T) {
	t.Run("setFieldValue edge cases", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Items", ArgType: optargs.RequiredArgument},
				{Name: "Flag", ArgType: optargs.NoArgument},
			},
		}
		integration := &CoreIntegration{metadata: metadata}

		// Test with slice field
		testStruct2 := &struct {
			Items []string `arg:"--items"`
		}{}

		fieldValue2 := reflect.ValueOf(testStruct2).Elem().FieldByName("Items")
		fieldMeta2 := &FieldMetadata{
			Name:    "Items",
			ArgType: optargs.RequiredArgument,
			Type:    reflect.TypeOf([]string{}),
		}

		err := integration.setFieldValue(fieldValue2, fieldMeta2, "item1,item2")
		if err != nil {
			t.Errorf("Unexpected error for slice field: %v", err)
		}

		// Test with boolean field and no argument
		testStruct3 := &struct {
			Flag bool `arg:"--flag"`
		}{}

		fieldValue3 := reflect.ValueOf(testStruct3).Elem().FieldByName("Flag")
		fieldMeta3 := &FieldMetadata{
			Name:    "Flag",
			ArgType: optargs.NoArgument,
			Type:    reflect.TypeOf(bool(false)),
		}

		err = integration.setFieldValue(fieldValue3, fieldMeta3, "")
		if err != nil {
			t.Errorf("Unexpected error for boolean field: %v", err)
		}

		if !testStruct3.Flag {
			t.Error("Expected Flag to be true")
		}

		// Test with unsettable field (this should return an error)
		testStruct := struct {
			unexported int
		}{}

		fieldValue := reflect.ValueOf(testStruct).FieldByName("unexported")
		fieldMeta := &FieldMetadata{
			Name:    "unexported",
			ArgType: optargs.RequiredArgument,
			Type:    reflect.TypeOf(int(0)),
		}

		err = integration.setFieldValue(fieldValue, fieldMeta, "123")
		if err == nil {
			t.Error("Expected error for unsettable field")
		}
	})

	t.Run("findFieldForOption edge cases", func(t *testing.T) {
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "Verbose", Short: "v", Long: "verbose"},
				{Name: "Count", Short: "c", Long: "count"},
			},
		}

		integration := &CoreIntegration{metadata: metadata}

		// Test finding by short option
		option := optargs.Option{Name: "v"}
		field, _ := integration.findFieldForOption(option, reflect.TypeOf(""))
		if field == nil || field.Name != "Verbose" {
			t.Error("Expected to find Verbose field by short option")
		}

		// Test finding by long option
		option2 := optargs.Option{Name: "count"}
		field2, _ := integration.findFieldForOption(option2, reflect.TypeOf(""))
		if field2 == nil || field2.Name != "Count" {
			t.Error("Expected to find Count field by long option")
		}

		// Test not finding non-existent option
		option3 := optargs.Option{Name: "nonexistent"}
		field3, _ := integration.findFieldForOption(option3, reflect.TypeOf(""))
		if field3 != nil {
			t.Error("Expected nil for non-existent option")
		}
	})
}

// TestTagParsingEdgeCases tests tag parsing functions with low coverage
func TestTagParsingEdgeCases(t *testing.T) {
	t.Run("mapToOptArgsCore edge cases", func(t *testing.T) {
		parser := &TagParser{}

		// Test with field that has no short or long options
		field := FieldMetadata{
			Name:       "Field",
			Short:      "",
			Long:       "",
			ArgType:    optargs.NoArgument,
			Positional: true,
			Type:       reflect.TypeOf(""),
		}

		err := parser.mapToOptArgsCore(&field)
		if err != nil {
			t.Errorf("Unexpected error for positional field: %v", err)
		}

		// Test with field that has only short option
		field2 := FieldMetadata{
			Name:    "Field2",
			Short:   "f",
			Long:    "",
			ArgType: optargs.NoArgument,
			Type:    reflect.TypeOf(bool(false)),
		}

		err = parser.mapToOptArgsCore(&field2)
		if err != nil {
			t.Errorf("Unexpected error for field with short option: %v", err)
		}

		// Test with field that has only long option
		field3 := FieldMetadata{
			Name:    "Field3",
			Short:   "",
			Long:    "field3",
			ArgType: optargs.NoArgument,
			Type:    reflect.TypeOf(bool(false)),
		}

		err = parser.mapToOptArgsCore(&field3)
		if err != nil {
			t.Errorf("Unexpected error for field with long option: %v", err)
		}
	})

	t.Run("parseDefaultValue edge cases", func(t *testing.T) {
		parser := &TagParser{}

		// Test with empty default value
		result, err := parser.parseDefaultValue("", reflect.TypeOf(""))
		if err != nil {
			t.Errorf("Unexpected error for empty default: %v", err)
		}
		if result != "" {
			t.Error("Expected empty string for empty default")
		}

		// Test with slice default value
		result, err = parser.parseDefaultValue("a,b,c", reflect.TypeOf([]string{}))
		if err != nil {
			t.Errorf("Unexpected error for slice default: %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result for slice default")
		}

		// Test with invalid type for default (channel type)
		result, err = parser.parseDefaultValue("invalid", reflect.TypeOf(make(chan int)))
		if err != nil {
			t.Errorf("Unexpected error for unsupported type: %v", err)
		}
		// For unsupported types, it returns the string as-is
		if result != "invalid" {
			t.Error("Expected string result for unsupported type")
		}
	})

	t.Run("GetEnvironmentValue edge cases", func(t *testing.T) {
		parser := &TagParser{}

		// Test with field that has no env tag
		field := FieldMetadata{Name: "Field", Env: ""}
		value, found := parser.GetEnvironmentValue(&field)
		if found {
			t.Error("Expected not found for field with no env tag")
		}
		if value != "" {
			t.Error("Expected empty value for field with no env tag")
		}

		// Test with field that has env tag but no environment variable set
		field2 := FieldMetadata{Name: "Field2", Env: "NONEXISTENT_ENV_VAR"}
		value2, found2 := parser.GetEnvironmentValue(&field2)
		if found2 {
			t.Error("Expected not found for non-existent environment variable")
		}
		if value2 != "" {
			t.Error("Expected empty value for non-existent environment variable")
		}
	})
}

// TestUncoveredParserFunctions tests functions with low coverage in parser.go
func TestUncoveredParserFunctions(t *testing.T) {
	t.Run("NewParser edge cases", func(t *testing.T) {
		// Test with invalid struct
		_, err := NewParser(Config{}, "not a struct")
		if err == nil {
			t.Error("Expected error for non-struct destination")
		}

		// Test with nil destination
		_, err = NewParser(Config{}, nil)
		if err == nil {
			t.Error("Expected error for nil destination")
		}
	})

	t.Run("findParentFieldForOption edge cases", func(t *testing.T) {
		parser := &Parser{
			metadata: &StructMetadata{
				Fields: []FieldMetadata{
					{Name: "Verbose", Short: "v", Long: "verbose"},
				},
			},
		}

		// Test with existing field (should find it)
		field := parser.findParentFieldForOption("v")
		if field == nil {
			t.Error("Expected to find field 'v' in parent metadata")
		}

		// Test with non-existent field
		field = parser.findParentFieldForOption("nonexistent")
		if field != nil {
			t.Error("Expected nil field for non-existent option")
		}
	})
}

// TestAdditionalCoverageImprovements adds more tests to improve coverage
func TestAdditionalCoverageImprovements(t *testing.T) {
	t.Run("ConvertCustom comprehensive coverage", func(t *testing.T) {
		converter := &TypeConverter{}

		// Test with custom type that implements TextUnmarshaler
		type CustomType struct {
			Value string
		}

		// We can't easily test the TextUnmarshaler path without a real implementation
		// So let's test the error paths instead

		// Test with struct type (should return error)
		structType := reflect.TypeOf(struct{}{})
		_, err := converter.ConvertCustom("test", structType)
		if err == nil {
			t.Error("Expected error for struct type that doesn't implement TextUnmarshaler")
		}

		// Test with map type (should return error)
		mapType := reflect.TypeOf(map[string]string{})
		_, err = converter.ConvertCustom("test", mapType)
		if err == nil {
			t.Error("Expected error for map type that doesn't implement TextUnmarshaler")
		}
	})

	t.Run("mapToOptArgsCore comprehensive coverage", func(t *testing.T) {
		parser := &TagParser{}

		// Test with different field types to improve coverage
		testCases := []struct {
			name      string
			fieldType reflect.Type
			short     string
			long      string
		}{
			{"string field", reflect.TypeOf(""), "s", "string"},
			{"int field", reflect.TypeOf(int(0)), "i", "int"},
			{"bool field", reflect.TypeOf(bool(false)), "b", "bool"},
			{"slice field", reflect.TypeOf([]string{}), "l", "list"},
			{"pointer field", reflect.TypeOf((*string)(nil)), "p", "pointer"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				field := FieldMetadata{
					Name:  tc.name,
					Short: tc.short,
					Long:  tc.long,
					Type:  tc.fieldType,
				}

				err := parser.mapToOptArgsCore(&field)
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.name, err)
				}
			})
		}

		// Test with subcommand field
		subcommandField := FieldMetadata{
			Name:         "SubCmd",
			Type:         reflect.TypeOf(struct{}{}),
			IsSubcommand: true,
		}

		err := parser.mapToOptArgsCore(&subcommandField)
		if err != nil {
			t.Errorf("Unexpected error for subcommand field: %v", err)
		}
	})

	t.Run("GenerateCompatibilityReport comprehensive coverage", func(t *testing.T) {
		framework := NewCompatibilityTestFramework()

		// Add multiple tests to improve coverage
		testStruct1 := &struct {
			Verbose bool `arg:"-v,--verbose"`
		}{}
		framework.AddCompatibilityTest("test1", testStruct1, []string{"--verbose"}, false)

		testStruct2 := &struct {
			Count int `arg:"-c,--count"`
		}{}
		framework.AddCompatibilityTest("test2", testStruct2, []string{"--count", "42"}, false)

		// Create a report with different scenarios
		report := &CompatibilityReport{
			TotalTests:  2,
			PassedTests: 1,
			FailedTests: 1,
			Scenarios: []ScenarioResult{
				{Name: "test1", Match: true},
				{Name: "test2", Match: false},
			},
		}

		// Test the report generation with failures
		reportStr := framework.GenerateCompatibilityReport(report)

		if reportStr == "" {
			t.Error("Expected non-empty compatibility report")
		}

		if !strings.Contains(reportStr, "Compatibility Test Report") {
			t.Error("Expected report header in output")
		}

		if !strings.Contains(reportStr, "Failed: 1") {
			t.Error("Expected failed tests count in output")
		}

		if !strings.Contains(reportStr, "test2: Results differ") {
			t.Error("Expected failed scenario in output")
		}
	})

	t.Run("parseDefaultValue additional coverage", func(t *testing.T) {
		parser := &TagParser{}

		// Test with invalid values for different types to improve error path coverage
		testCases := []struct {
			name        string
			defaultStr  string
			fieldType   reflect.Type
			expectError bool
		}{
			{"invalid int", "not-a-number", reflect.TypeOf(int(0)), true},
			{"invalid uint", "not-a-number", reflect.TypeOf(uint(0)), true},
			{"invalid float", "not-a-number", reflect.TypeOf(float64(0)), true},
			{"invalid bool", "not-a-bool", reflect.TypeOf(bool(false)), true},
			{"valid int", "42", reflect.TypeOf(int(0)), false},
			{"valid uint", "42", reflect.TypeOf(uint(0)), false},
			{"valid float", "3.14", reflect.TypeOf(float64(0)), false},
			{"valid bool", "true", reflect.TypeOf(bool(false)), false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := parser.parseDefaultValue(tc.defaultStr, tc.fieldType)

				if tc.expectError {
					if err == nil {
						t.Errorf("Expected error for %s", tc.name)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for %s: %v", tc.name, err)
					}
					if result == nil {
						t.Errorf("Expected non-nil result for %s", tc.name)
					}
				}
			})
		}

		// Test slice with invalid elements
		_, err := parser.parseDefaultValue("1,not-a-number,3", reflect.TypeOf([]int{}))
		if err == nil {
			t.Error("Expected error for slice with invalid elements")
		}
	})

	t.Run("setScalarValue additional coverage", func(t *testing.T) {
		integration := &CoreIntegration{}

		// Test with different scalar types
		testCases := []struct {
			name      string
			fieldType reflect.Type
			value     string
		}{
			{"string", reflect.TypeOf(""), "test"},
			{"int", reflect.TypeOf(int(0)), "42"},
			{"bool", reflect.TypeOf(bool(false)), "true"},
			{"float", reflect.TypeOf(float64(0)), "3.14"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Create a struct with the field type
				structType := reflect.StructOf([]reflect.StructField{
					{
						Name: "TestField",
						Type: tc.fieldType,
						Tag:  `arg:"--test"`,
					},
				})

				structValue := reflect.New(structType).Elem()
				fieldValue := structValue.FieldByName("TestField")

				err := integration.setScalarValue(fieldValue, tc.fieldType, tc.value)
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.name, err)
				}
			})
		}
	})
}

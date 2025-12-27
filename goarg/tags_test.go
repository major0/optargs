package goarg

import (
	"reflect"
	"testing"

	"github.com/major0/optargs"
)

func TestTagParser_ParseField(t *testing.T) {
	parser := &TagParser{}

	tests := []struct {
		name     string
		field    reflect.StructField
		expected FieldMetadata
		wantErr  bool
	}{
		{
			name: "basic_short_and_long_option",
			field: reflect.StructField{
				Name: "Verbose",
				Type: reflect.TypeOf(false),
				Tag:  `arg:"-v,--verbose" help:"enable verbose output"`,
			},
			expected: FieldMetadata{
				Name:     "Verbose",
				Type:     reflect.TypeOf(false),
				Tag:      `arg:"-v,--verbose" help:"enable verbose output"`,
				Short:    "v",
				Long:     "verbose",
				Help:     "enable verbose output",
				ArgType:  optargs.NoArgument,
				CoreFlag: &optargs.Flag{Name: "verbose", HasArg: optargs.NoArgument},
			},
		},
		{
			name: "long_option_only",
			field: reflect.StructField{
				Name: "Count",
				Type: reflect.TypeOf(0),
				Tag:  `arg:"--count" help:"number of items"`,
			},
			expected: FieldMetadata{
				Name:     "Count",
				Type:     reflect.TypeOf(0),
				Tag:      `arg:"--count" help:"number of items"`,
				Long:     "count",
				Help:     "number of items",
				ArgType:  optargs.RequiredArgument,
				CoreFlag: &optargs.Flag{Name: "count", HasArg: optargs.RequiredArgument},
			},
		},
		{
			name: "short_option_only",
			field: reflect.StructField{
				Name: "Debug",
				Type: reflect.TypeOf(false),
				Tag:  `arg:"-d"`,
			},
			expected: FieldMetadata{
				Name:     "Debug",
				Type:     reflect.TypeOf(false),
				Tag:      `arg:"-d"`,
				Short:    "d",
				ArgType:  optargs.NoArgument,
				CoreFlag: &optargs.Flag{Name: "d", HasArg: optargs.NoArgument},
			},
		},
		{
			name: "required_option",
			field: reflect.StructField{
				Name: "Input",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"--input,required" help:"input file path"`,
			},
			expected: FieldMetadata{
				Name:     "Input",
				Type:     reflect.TypeOf(""),
				Tag:      `arg:"--input,required" help:"input file path"`,
				Long:     "input",
				Help:     "input file path",
				Required: true,
				ArgType:  optargs.RequiredArgument,
				CoreFlag: &optargs.Flag{Name: "input", HasArg: optargs.RequiredArgument},
			},
		},
		{
			name: "positional_argument",
			field: reflect.StructField{
				Name: "Files",
				Type: reflect.TypeOf([]string{}),
				Tag:  `arg:"positional" help:"files to process"`,
			},
			expected: FieldMetadata{
				Name:       "Files",
				Type:       reflect.TypeOf([]string{}),
				Tag:        `arg:"positional" help:"files to process"`,
				Help:       "files to process",
				Positional: true,
				ArgType:    optargs.NoArgument, // Positional args don't use OptArgs flags
			},
		},
		{
			name: "environment_variable",
			field: reflect.StructField{
				Name: "Token",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"--token,env:API_TOKEN" help:"API token"`,
			},
			expected: FieldMetadata{
				Name:     "Token",
				Type:     reflect.TypeOf(""),
				Tag:      `arg:"--token,env:API_TOKEN" help:"API token"`,
				Long:     "token",
				Help:     "API token",
				Env:      "API_TOKEN",
				ArgType:  optargs.RequiredArgument,
				CoreFlag: &optargs.Flag{Name: "token", HasArg: optargs.RequiredArgument},
			},
		},
		{
			name: "default_value",
			field: reflect.StructField{
				Name: "Port",
				Type: reflect.TypeOf(0),
				Tag:  `arg:"-p,--port" default:"8080" help:"server port"`,
			},
			expected: FieldMetadata{
				Name:     "Port",
				Type:     reflect.TypeOf(0),
				Tag:      `arg:"-p,--port" default:"8080" help:"server port"`,
				Short:    "p",
				Long:     "port",
				Help:     "server port",
				Default:  int64(8080),
				ArgType:  optargs.RequiredArgument,
				CoreFlag: &optargs.Flag{Name: "port", HasArg: optargs.RequiredArgument},
			},
		},
		{
			name: "no_arg_tag_generates_default",
			field: reflect.StructField{
				Name: "Output",
				Type: reflect.TypeOf(""),
				Tag:  `help:"output file"`,
			},
			expected: FieldMetadata{
				Name:     "Output",
				Type:     reflect.TypeOf(""),
				Tag:      `help:"output file"`,
				Long:     "output", // Generated from field name
				Help:     "output file",
				ArgType:  optargs.RequiredArgument,
				CoreFlag: &optargs.Flag{Name: "output", HasArg: optargs.RequiredArgument},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseField(tt.field)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseField() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseField() unexpected error: %v", err)
				return
			}

			// Compare all fields except CoreFlag (which needs special handling)
			if result.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
			}
			if result.Short != tt.expected.Short {
				t.Errorf("Short = %v, want %v", result.Short, tt.expected.Short)
			}
			if result.Long != tt.expected.Long {
				t.Errorf("Long = %v, want %v", result.Long, tt.expected.Long)
			}
			if result.Help != tt.expected.Help {
				t.Errorf("Help = %v, want %v", result.Help, tt.expected.Help)
			}
			if result.Required != tt.expected.Required {
				t.Errorf("Required = %v, want %v", result.Required, tt.expected.Required)
			}
			if result.Positional != tt.expected.Positional {
				t.Errorf("Positional = %v, want %v", result.Positional, tt.expected.Positional)
			}
			if result.Env != tt.expected.Env {
				t.Errorf("Env = %v, want %v", result.Env, tt.expected.Env)
			}
			if result.ArgType != tt.expected.ArgType {
				t.Errorf("ArgType = %v, want %v", result.ArgType, tt.expected.ArgType)
			}

			// Compare default values
			if !reflect.DeepEqual(result.Default, tt.expected.Default) {
				t.Errorf("Default = %v, want %v", result.Default, tt.expected.Default)
			}

			// Compare CoreFlag if expected
			if tt.expected.CoreFlag != nil {
				if result.CoreFlag == nil {
					t.Errorf("CoreFlag is nil, expected %+v", tt.expected.CoreFlag)
				} else {
					if result.CoreFlag.Name != tt.expected.CoreFlag.Name {
						t.Errorf("CoreFlag.Name = %v, want %v", result.CoreFlag.Name, tt.expected.CoreFlag.Name)
					}
					if result.CoreFlag.HasArg != tt.expected.CoreFlag.HasArg {
						t.Errorf("CoreFlag.HasArg = %v, want %v", result.CoreFlag.HasArg, tt.expected.CoreFlag.HasArg)
					}
				}
			}
		})
	}
}

func TestTagParser_ParseStruct(t *testing.T) {
	parser := &TagParser{}

	// Test struct with various field types
	type TestStruct struct {
		Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int    `arg:"-c,--count" help:"number of items"`
		Input   string `arg:"--input,required" help:"input file path"`
		Files   []string `arg:"positional" help:"files to process"`
	}

	var testStruct TestStruct
	metadata, err := parser.ParseStruct(&testStruct)
	if err != nil {
		t.Fatalf("ParseStruct() unexpected error: %v", err)
	}

	if len(metadata.Fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(metadata.Fields))
	}

	// Check that we have the expected fields
	fieldNames := make(map[string]bool)
	for _, field := range metadata.Fields {
		fieldNames[field.Name] = true
	}

	expectedFields := []string{"Verbose", "Count", "Input", "Files"}
	for _, expected := range expectedFields {
		if !fieldNames[expected] {
			t.Errorf("Missing expected field: %s", expected)
		}
	}

	// Check specific field properties
	for _, field := range metadata.Fields {
		switch field.Name {
		case "Verbose":
			if field.Short != "v" || field.Long != "verbose" {
				t.Errorf("Verbose field: expected short='v', long='verbose', got short='%s', long='%s'", field.Short, field.Long)
			}
		case "Input":
			if !field.Required {
				t.Errorf("Input field should be required")
			}
		case "Files":
			if !field.Positional {
				t.Errorf("Files field should be positional")
			}
		}
	}
}

func TestTagParser_ErrorCases(t *testing.T) {
	parser := &TagParser{}

	tests := []struct {
		name  string
		field reflect.StructField
	}{
		{
			name: "invalid_short_option_too_long",
			field: reflect.StructField{
				Name: "Test",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"-verbose"`, // Invalid: short option with multiple chars
			},
		},
		{
			name: "positional_with_flags",
			field: reflect.StructField{
				Name: "Test",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"positional,-v"`, // Invalid: positional can't have flags
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseField(tt.field)
			if err == nil {
				t.Errorf("ParseField() expected error for invalid input, got nil")
			}
		})
	}
}
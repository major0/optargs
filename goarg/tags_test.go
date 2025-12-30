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
		Verbose bool     `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int      `arg:"-c,--count" help:"number of items"`
		Input   string   `arg:"--input,required" help:"input file path"`
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

func TestTagParser_SubcommandProcessing(t *testing.T) {
	parser := &TagParser{}

	// Define subcommand structs
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type ClientCmd struct {
		URL     string `arg:"--url,required" help:"server URL"`
		Timeout int    `arg:"--timeout" default:"30" help:"timeout in seconds"`
	}

	// Test struct with subcommands
	type AppWithSubcommands struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Server  *ServerCmd `arg:"subcommand:server" help:"run server"`
		Client  *ClientCmd `arg:"subcommand:client" help:"run client"`
		Default *ServerCmd `arg:"subcommand" help:"default subcommand"`
	}

	tests := []struct {
		name     string
		field    reflect.StructField
		expected FieldMetadata
		wantErr  bool
	}{
		{
			name: "explicit_subcommand_name",
			field: reflect.StructField{
				Name: "Server",
				Type: reflect.TypeOf((*ServerCmd)(nil)),
				Tag:  `arg:"subcommand:server" help:"run server"`,
			},
			expected: FieldMetadata{
				Name:           "Server",
				Type:           reflect.TypeOf((*ServerCmd)(nil)),
				Tag:            `arg:"subcommand:server" help:"run server"`,
				Help:           "run server",
				IsSubcommand:   true,
				SubcommandName: "server",
			},
		},
		{
			name: "default_subcommand_name",
			field: reflect.StructField{
				Name: "Default",
				Type: reflect.TypeOf((*ServerCmd)(nil)),
				Tag:  `arg:"subcommand" help:"default subcommand"`,
			},
			expected: FieldMetadata{
				Name:           "Default",
				Type:           reflect.TypeOf((*ServerCmd)(nil)),
				Tag:            `arg:"subcommand" help:"default subcommand"`,
				Help:           "default subcommand",
				IsSubcommand:   true,
				SubcommandName: "default",
			},
		},
		{
			name: "subcommand_non_pointer_struct_error",
			field: reflect.StructField{
				Name: "Invalid",
				Type: reflect.TypeOf(ServerCmd{}), // Not a pointer
				Tag:  `arg:"subcommand:invalid"`,
			},
			wantErr: true,
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

			if result.IsSubcommand != tt.expected.IsSubcommand {
				t.Errorf("IsSubcommand = %v, want %v", result.IsSubcommand, tt.expected.IsSubcommand)
			}
			if result.SubcommandName != tt.expected.SubcommandName {
				t.Errorf("SubcommandName = %v, want %v", result.SubcommandName, tt.expected.SubcommandName)
			}
		})
	}

	// Test full struct parsing with subcommands
	t.Run("full_struct_with_subcommands", func(t *testing.T) {
		var app AppWithSubcommands
		metadata, err := parser.ParseStruct(&app)
		if err != nil {
			t.Fatalf("ParseStruct() unexpected error: %v", err)
		}

		// Should have 1 regular field (Verbose) and 3 subcommands
		if len(metadata.Fields) != 1 {
			t.Errorf("Expected 1 regular field, got %d", len(metadata.Fields))
		}
		if len(metadata.Subcommands) != 3 {
			t.Errorf("Expected 3 subcommands, got %d", len(metadata.Subcommands))
		}

		// Check subcommand names
		expectedSubcommands := []string{"server", "client", "default"}
		for _, name := range expectedSubcommands {
			if _, exists := metadata.Subcommands[name]; !exists {
				t.Errorf("Missing expected subcommand: %s", name)
			}
		}

		// Check that subcommand metadata is parsed correctly
		serverMeta := metadata.Subcommands["server"]
		if serverMeta == nil {
			t.Fatal("Server subcommand metadata is nil")
		}
		if len(serverMeta.Fields) != 2 {
			t.Errorf("Server subcommand should have 2 fields, got %d", len(serverMeta.Fields))
		}
	})
}

func TestTagParser_PositionalArguments(t *testing.T) {
	parser := &TagParser{}

	tests := []struct {
		name     string
		field    reflect.StructField
		expected FieldMetadata
		wantErr  bool
	}{
		{
			name: "string_slice_positional",
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
			},
		},
		{
			name: "int_slice_positional",
			field: reflect.StructField{
				Name: "Numbers",
				Type: reflect.TypeOf([]int{}),
				Tag:  `arg:"positional" help:"list of numbers"`,
			},
			expected: FieldMetadata{
				Name:       "Numbers",
				Type:       reflect.TypeOf([]int{}),
				Tag:        `arg:"positional" help:"list of numbers"`,
				Help:       "list of numbers",
				Positional: true,
			},
		},
		{
			name: "required_positional",
			field: reflect.StructField{
				Name: "Input",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"positional,required" help:"input file"`,
			},
			expected: FieldMetadata{
				Name:       "Input",
				Type:       reflect.TypeOf(""),
				Tag:        `arg:"positional,required" help:"input file"`,
				Help:       "input file",
				Positional: true,
				Required:   true,
			},
		},
		{
			name: "positional_with_env",
			field: reflect.StructField{
				Name: "Config",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"positional,env:CONFIG_FILE" help:"config file"`,
			},
			expected: FieldMetadata{
				Name:       "Config",
				Type:       reflect.TypeOf(""),
				Tag:        `arg:"positional,env:CONFIG_FILE" help:"config file"`,
				Help:       "config file",
				Positional: true,
				Env:        "CONFIG_FILE",
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

			if result.Positional != tt.expected.Positional {
				t.Errorf("Positional = %v, want %v", result.Positional, tt.expected.Positional)
			}
			if result.Required != tt.expected.Required {
				t.Errorf("Required = %v, want %v", result.Required, tt.expected.Required)
			}
			if result.Env != tt.expected.Env {
				t.Errorf("Env = %v, want %v", result.Env, tt.expected.Env)
			}
		})
	}
}

func TestTagParser_EnvironmentVariables(t *testing.T) {
	parser := &TagParser{}

	tests := []struct {
		name     string
		field    reflect.StructField
		expected FieldMetadata
	}{
		{
			name: "env_with_long_option",
			field: reflect.StructField{
				Name: "Token",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"--token,env:API_TOKEN" help:"API token"`,
			},
			expected: FieldMetadata{
				Name: "Token",
				Long: "token",
				Env:  "API_TOKEN",
				Help: "API token",
			},
		},
		{
			name: "env_with_short_and_long",
			field: reflect.StructField{
				Name: "Password",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"-p,--password,env:PASSWORD" help:"password"`,
			},
			expected: FieldMetadata{
				Name:  "Password",
				Short: "p",
				Long:  "password",
				Env:   "PASSWORD",
				Help:  "password",
			},
		},
		{
			name: "env_only",
			field: reflect.StructField{
				Name: "Secret",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"env:SECRET_KEY" help:"secret key"`,
			},
			expected: FieldMetadata{
				Name: "Secret",
				Long: "secret", // Generated from field name
				Env:  "SECRET_KEY",
				Help: "secret key",
			},
		},
		{
			name: "env_with_required",
			field: reflect.StructField{
				Name: "Database",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"--db,env:DATABASE_URL,required" help:"database URL"`,
			},
			expected: FieldMetadata{
				Name:     "Database",
				Long:     "db",
				Env:      "DATABASE_URL",
				Required: true,
				Help:     "database URL",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseField(tt.field)
			if err != nil {
				t.Errorf("ParseField() unexpected error: %v", err)
				return
			}

			if result.Env != tt.expected.Env {
				t.Errorf("Env = %v, want %v", result.Env, tt.expected.Env)
			}
			if result.Long != tt.expected.Long {
				t.Errorf("Long = %v, want %v", result.Long, tt.expected.Long)
			}
			if result.Short != tt.expected.Short {
				t.Errorf("Short = %v, want %v", result.Short, tt.expected.Short)
			}
			if result.Required != tt.expected.Required {
				t.Errorf("Required = %v, want %v", result.Required, tt.expected.Required)
			}
		})
	}

	// Test environment variable retrieval
	t.Run("environment_variable_retrieval", func(t *testing.T) {
		field := reflect.StructField{
			Name: "TestEnv",
			Type: reflect.TypeOf(""),
			Tag:  `arg:"--test,env:TEST_VAR"`,
		}

		result, err := parser.ParseField(field)
		if err != nil {
			t.Fatalf("ParseField() unexpected error: %v", err)
		}

		// Test GetEnvironmentValue method
		// First, ensure the env var doesn't exist
		value, exists := parser.GetEnvironmentValue(result)
		if exists {
			t.Errorf("Expected TEST_VAR to not exist, but got value: %s", value)
		}

		// Set the environment variable and test again
		t.Setenv("TEST_VAR", "test_value")
		value, exists = parser.GetEnvironmentValue(result)
		if !exists {
			t.Errorf("Expected TEST_VAR to exist after setting")
		}
		if value != "test_value" {
			t.Errorf("Expected value 'test_value', got '%s'", value)
		}
	})
}

func TestTagParser_DefaultValues(t *testing.T) {
	parser := &TagParser{}

	tests := []struct {
		name     string
		field    reflect.StructField
		expected interface{}
		wantErr  bool
	}{
		{
			name: "string_default",
			field: reflect.StructField{
				Name: "Name",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"--name" default:"test"`,
			},
			expected: "test",
		},
		{
			name: "int_default",
			field: reflect.StructField{
				Name: "Count",
				Type: reflect.TypeOf(0),
				Tag:  `arg:"--count" default:"42"`,
			},
			expected: int64(42),
		},
		{
			name: "int64_default",
			field: reflect.StructField{
				Name: "Size",
				Type: reflect.TypeOf(int64(0)),
				Tag:  `arg:"--size" default:"1024"`,
			},
			expected: int64(1024),
		},
		{
			name: "float64_default",
			field: reflect.StructField{
				Name: "Rate",
				Type: reflect.TypeOf(float64(0)),
				Tag:  `arg:"--rate" default:"3.14"`,
			},
			expected: float64(3.14),
		},
		{
			name: "bool_default_true",
			field: reflect.StructField{
				Name: "Enabled",
				Type: reflect.TypeOf(false),
				Tag:  `arg:"--enabled" default:"true"`,
			},
			expected: true,
		},
		{
			name: "bool_default_false",
			field: reflect.StructField{
				Name: "Disabled",
				Type: reflect.TypeOf(false),
				Tag:  `arg:"--disabled" default:"false"`,
			},
			expected: false,
		},
		{
			name: "string_slice_default",
			field: reflect.StructField{
				Name: "Items",
				Type: reflect.TypeOf([]string{}),
				Tag:  `arg:"--items" default:"a,b,c"`,
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "int_slice_default",
			field: reflect.StructField{
				Name: "Numbers",
				Type: reflect.TypeOf([]int{}),
				Tag:  `arg:"--numbers" default:"1,2,3"`,
			},
			expected: []int{1, 2, 3},
		},
		{
			name: "empty_slice_default",
			field: reflect.StructField{
				Name: "Empty",
				Type: reflect.TypeOf([]string{}),
				Tag:  `arg:"--empty" default:""`,
			},
			expected: []string{},
		},
		{
			name: "invalid_int_default",
			field: reflect.StructField{
				Name: "BadInt",
				Type: reflect.TypeOf(0),
				Tag:  `arg:"--bad" default:"not_a_number"`,
			},
			wantErr: true,
		},
		{
			name: "invalid_bool_default",
			field: reflect.StructField{
				Name: "BadBool",
				Type: reflect.TypeOf(false),
				Tag:  `arg:"--bad" default:"maybe"`,
			},
			wantErr: true,
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

			if !reflect.DeepEqual(result.Default, tt.expected) {
				t.Errorf("Default = %v (type %T), want %v (type %T)",
					result.Default, result.Default, tt.expected, tt.expected)
			}
		})
	}
}

func TestTagParser_ComplexTagFormats(t *testing.T) {
	parser := &TagParser{}

	tests := []struct {
		name     string
		field    reflect.StructField
		expected FieldMetadata
		wantErr  bool
	}{
		{
			name: "all_options_combined",
			field: reflect.StructField{
				Name: "Config",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"-c,--config,env:CONFIG_FILE,required" default:"/etc/app.conf" help:"configuration file"`,
			},
			expected: FieldMetadata{
				Name:     "Config",
				Short:    "c",
				Long:     "config",
				Env:      "CONFIG_FILE",
				Required: true,
				Default:  "/etc/app.conf",
				Help:     "configuration file",
			},
		},
		{
			name: "multiple_env_vars_invalid",
			field: reflect.StructField{
				Name: "Invalid",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"env:VAR1,env:VAR2"`, // Multiple env vars not supported
			},
			expected: FieldMetadata{
				Name: "Invalid",
				Long: "invalid", // Generated from field name
				Env:  "VAR2",    // Last one wins
			},
		},
		{
			name: "whitespace_in_tags",
			field: reflect.StructField{
				Name: "Spaced",
				Type: reflect.TypeOf(""),
				Tag:  `arg:" -s , --spaced , required " help:"  spaced tags  "`,
			},
			expected: FieldMetadata{
				Name:     "Spaced",
				Short:    "s",
				Long:     "spaced",
				Required: true,
				Help:     "  spaced tags  ", // Help preserves whitespace
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

			if result.Short != tt.expected.Short {
				t.Errorf("Short = %v, want %v", result.Short, tt.expected.Short)
			}
			if result.Long != tt.expected.Long {
				t.Errorf("Long = %v, want %v", result.Long, tt.expected.Long)
			}
			if result.Env != tt.expected.Env {
				t.Errorf("Env = %v, want %v", result.Env, tt.expected.Env)
			}
			if result.Required != tt.expected.Required {
				t.Errorf("Required = %v, want %v", result.Required, tt.expected.Required)
			}
			if result.Help != tt.expected.Help {
				t.Errorf("Help = %v, want %v", result.Help, tt.expected.Help)
			}
			if !reflect.DeepEqual(result.Default, tt.expected.Default) {
				t.Errorf("Default = %v, want %v", result.Default, tt.expected.Default)
			}
		})
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
		{
			name: "subcommand_not_pointer_to_struct",
			field: reflect.StructField{
				Name: "BadSubcmd",
				Type: reflect.TypeOf(""), // Should be pointer to struct
				Tag:  `arg:"subcommand:bad"`,
			},
		},
		{
			name: "unknown_tag_format",
			field: reflect.StructField{
				Name: "Unknown",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"unknown_format"`,
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

func TestTagParser_ParseStruct_ErrorCases(t *testing.T) {
	parser := &TagParser{}

	tests := []struct {
		name    string
		dest    interface{}
		wantErr bool
	}{
		{
			name:    "nil_destination",
			dest:    nil,
			wantErr: true,
		},
		{
			name:    "non_pointer_destination",
			dest:    struct{}{},
			wantErr: true,
		},
		{
			name:    "pointer_to_non_struct",
			dest:    new(string),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseStruct(tt.dest)
			if tt.wantErr && err == nil {
				t.Errorf("ParseStruct() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ParseStruct() unexpected error: %v", err)
			}
		})
	}
}

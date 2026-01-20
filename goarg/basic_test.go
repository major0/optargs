package goarg

import (
	"os"
	"reflect"
	"testing"
)

func TestBasicParsing(t *testing.T) {
	type BasicCmd struct {
		Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int    `arg:"-c,--count" help:"number of items"`
		Input   string `arg:"--input,required" help:"input file path"`
		Port    int    `arg:"-p,--port" default:"8080" help:"server port"`
	}

	tests := []struct {
		name     string
		args     []string
		expected BasicCmd
		wantErr  bool
	}{
		{
			name: "short_options",
			args: []string{"-v", "-c", "42", "--input", "test.txt"},
			expected: BasicCmd{
				Verbose: true,
				Count:   42,
				Input:   "test.txt",
				Port:    8080, // default
			},
		},
		{
			name: "long_options",
			args: []string{"--verbose", "--count", "100", "--input", "data.txt", "--port", "9000"},
			expected: BasicCmd{
				Verbose: true,
				Count:   100,
				Input:   "data.txt",
				Port:    9000,
			},
		},
		{
			name: "mixed_options",
			args: []string{"-v", "--count", "50", "--input", "mixed.txt", "-p", "3000"},
			expected: BasicCmd{
				Verbose: true,
				Count:   50,
				Input:   "mixed.txt",
				Port:    3000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd BasicCmd
			err := ParseArgs(&cmd, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseArgs() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseArgs() unexpected error: %v", err)
				return
			}

			if cmd.Verbose != tt.expected.Verbose {
				t.Errorf("Verbose = %v, want %v", cmd.Verbose, tt.expected.Verbose)
			}
			if cmd.Count != tt.expected.Count {
				t.Errorf("Count = %v, want %v", cmd.Count, tt.expected.Count)
			}
			if cmd.Input != tt.expected.Input {
				t.Errorf("Input = %v, want %v", cmd.Input, tt.expected.Input)
			}
			if cmd.Port != tt.expected.Port {
				t.Errorf("Port = %v, want %v", cmd.Port, tt.expected.Port)
			}
		})
	}
}

func TestPositionalArguments(t *testing.T) {
	type CopyCmd struct {
		Verbose bool     `arg:"-v,--verbose" help:"enable verbose output"`
		Source  string   `arg:"positional,required" help:"source file"`
		Dest    string   `arg:"positional,required" help:"destination file"`
		Files   []string `arg:"positional" help:"additional files"`
	}

	tests := []struct {
		name     string
		args     []string
		expected CopyCmd
		wantErr  bool
	}{
		{
			name: "required_positionals_only",
			args: []string{"src.txt", "dst.txt"},
			expected: CopyCmd{
				Source: "src.txt",
				Dest:   "dst.txt",
				Files:  []string{},
			},
		},
		{
			name: "with_additional_files",
			args: []string{"src.txt", "dst.txt", "file1.txt", "file2.txt"},
			expected: CopyCmd{
				Source: "src.txt",
				Dest:   "dst.txt",
				Files:  []string{"file1.txt", "file2.txt"},
			},
		},
		{
			name: "with_options_and_positionals",
			args: []string{"-v", "src.txt", "dst.txt"},
			expected: CopyCmd{
				Verbose: true,
				Source:  "src.txt",
				Dest:    "dst.txt",
				Files:   []string{},
			},
		},
		{
			name:    "missing_required_positional",
			args:    []string{"src.txt"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd CopyCmd
			err := ParseArgs(&cmd, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseArgs() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseArgs() unexpected error: %v", err)
				return
			}

			if cmd.Verbose != tt.expected.Verbose {
				t.Errorf("Verbose = %v, want %v", cmd.Verbose, tt.expected.Verbose)
			}
			if cmd.Source != tt.expected.Source {
				t.Errorf("Source = %v, want %v", cmd.Source, tt.expected.Source)
			}
			if cmd.Dest != tt.expected.Dest {
				t.Errorf("Dest = %v, want %v", cmd.Dest, tt.expected.Dest)
			}
			if !reflect.DeepEqual(cmd.Files, tt.expected.Files) {
				t.Errorf("Files = %v, want %v", cmd.Files, tt.expected.Files)
			}
		})
	}
}

func TestEnvironmentVariables(t *testing.T) {
	type EnvCmd struct {
		Token string `arg:"--token,env:API_TOKEN" help:"API token"`
		Debug bool   `arg:"--debug,env:DEBUG" help:"enable debug mode"`
		Port  int    `arg:"--port,env:PORT" default:"8080" help:"server port"`
	}

	// Set environment variables for testing
	os.Setenv("API_TOKEN", "test-token-123")
	os.Setenv("DEBUG", "true")
	defer func() {
		os.Unsetenv("API_TOKEN")
		os.Unsetenv("DEBUG")
	}()

	tests := []struct {
		name     string
		args     []string
		expected EnvCmd
	}{
		{
			name: "use_environment_variables",
			args: []string{},
			expected: EnvCmd{
				Token: "test-token-123",
				Debug: true,
				Port:  8080, // Default value
			},
		},
		{
			name: "command_line_overrides_env",
			args: []string{"--token", "cli-token", "--port", "9000"},
			expected: EnvCmd{
				Token: "cli-token", // CLI overrides env
				Debug: true,        // From env
				Port:  9000,        // CLI overrides default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd EnvCmd
			err := ParseArgs(&cmd, tt.args)
			if err != nil {
				t.Errorf("ParseArgs() unexpected error: %v", err)
				return
			}

			if cmd.Token != tt.expected.Token {
				t.Errorf("Token = %v, want %v", cmd.Token, tt.expected.Token)
			}
			if cmd.Debug != tt.expected.Debug {
				t.Errorf("Debug = %v, want %v", cmd.Debug, tt.expected.Debug)
			}
			if cmd.Port != tt.expected.Port {
				t.Errorf("Port = %v, want %v", cmd.Port, tt.expected.Port)
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	type DefaultCmd struct {
		StringVal string   `arg:"--string" default:"default-string" help:"string value"`
		IntVal    int      `arg:"--int" default:"42" help:"integer value"`
		FloatVal  float64  `arg:"--float" default:"3.14" help:"float value"`
		BoolVal   bool     `arg:"--bool" default:"true" help:"boolean value"`
		SliceVal  []string `arg:"--slice" help:"slice value"` // No default - upstream doesn't support slice defaults
	}

	tests := []struct {
		name     string
		args     []string
		expected DefaultCmd
	}{
		{
			name: "use_all_defaults",
			args: []string{},
			expected: DefaultCmd{
				StringVal: "default-string",
				IntVal:    42,
				FloatVal:  3.14,
				BoolVal:   true,
				SliceVal:  nil, // No default for slices
			},
		},
		{
			name: "override_some_defaults",
			args: []string{"--string", "custom", "--int", "100"},
			expected: DefaultCmd{
				StringVal: "custom", // Overridden
				IntVal:    100,      // Overridden
				FloatVal:  3.14,     // Default
				BoolVal:   true,     // Default
				SliceVal:  nil,      // No default for slices
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd DefaultCmd
			err := ParseArgs(&cmd, tt.args)
			if err != nil {
				t.Errorf("ParseArgs() unexpected error: %v", err)
				return
			}

			if cmd.StringVal != tt.expected.StringVal {
				t.Errorf("StringVal = %v, want %v", cmd.StringVal, tt.expected.StringVal)
			}
			if cmd.IntVal != tt.expected.IntVal {
				t.Errorf("IntVal = %v, want %v", cmd.IntVal, tt.expected.IntVal)
			}
			if cmd.FloatVal != tt.expected.FloatVal {
				t.Errorf("FloatVal = %v, want %v", cmd.FloatVal, tt.expected.FloatVal)
			}
			if cmd.BoolVal != tt.expected.BoolVal {
				t.Errorf("BoolVal = %v, want %v", cmd.BoolVal, tt.expected.BoolVal)
			}
			if !reflect.DeepEqual(cmd.SliceVal, tt.expected.SliceVal) {
				t.Errorf("SliceVal = %v, want %v", cmd.SliceVal, tt.expected.SliceVal)
			}
		})
	}
}

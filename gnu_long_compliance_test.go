package optargs

import (
	"testing"
)

// TestGNULongOptionSyntax tests GNU long option syntax compliance.
func TestGNULongOptionSyntax(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
	}

	tests := []struct {
		name     string
		args     []string
		expected []Option
	}{
		{
			name: "long option no argument",
			args: []string{"--verbose"},
			expected: []Option{
				{Name: "verbose", HasArg: false, Arg: ""},
			},
		},
		{
			name: "long option with equals syntax",
			args: []string{"--output=file.txt"},
			expected: []Option{
				{Name: "output", HasArg: true, Arg: "file.txt"},
			},
		},
		{
			name: "long option with space syntax",
			args: []string{"--output", "file.txt"},
			expected: []Option{
				{Name: "output", HasArg: true, Arg: "file.txt"},
			},
		},
		{
			name: "optional argument provided with equals",
			args: []string{"--config=debug"},
			expected: []Option{
				{Name: "config", HasArg: true, Arg: "debug"},
			},
		},
		{
			name: "optional argument not provided",
			args: []string{"--config"},
			expected: []Option{
				{Name: "config", HasArg: false, Arg: ""},
			},
		},
		{
			name: "empty argument with equals",
			args: []string{"--output="},
			expected: []Option{
				{Name: "output", HasArg: true, Arg: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOptLong(tt.args, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong failed: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, parser), tt.expected)
		})
	}
}

// TestGNULongOptionPartialMatching tests that partial matching is not supported.
func TestGNULongOptionPartialMatching(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "version", HasArg: NoArgument},
		{Name: "help", HasArg: NoArgument},
	}

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "partial match hel fails",
			args:      []string{"--hel"},
			expectErr: true,
		},
		{
			name:      "partial match ver fails",
			args:      []string{"--ver"},
			expectErr: true,
		},
		{
			name:      "exact match works",
			args:      []string{"--verbose"},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOptLong(tt.args, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong failed: %v", err)
			}

			var optErr error
			for _, err := range parser.Options() {
				if err != nil {
					optErr = err
					break
				}
			}

			if tt.expectErr && optErr == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && optErr != nil {
				t.Fatalf("unexpected error: %v", optErr)
			}
		})
	}
}

// TestGNULongOptionCaseSensitivity tests case-insensitive long option matching.
func TestGNULongOptionCaseSensitivity(t *testing.T) {
	longOpts := []Flag{
		{Name: "Verbose", HasArg: NoArgument},
		{Name: "OUTPUT", HasArg: RequiredArgument},
	}

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "lowercase matches Verbose",
			args:      []string{"--verbose"},
			expectErr: false,
		},
		{
			name:      "lowercase matches OUTPUT",
			args:      []string{"--output=file.txt"},
			expectErr: false,
		},
		{
			name:      "exact case matches Verbose",
			args:      []string{"--Verbose"},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOptLong(tt.args, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong failed: %v", err)
			}

			var optErr error
			for _, err := range parser.Options() {
				if err != nil {
					optErr = err
					break
				}
			}

			if tt.expectErr && optErr == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && optErr != nil {
				t.Fatalf("unexpected error: %v", optErr)
			}
		})
	}
}

// TestGNULongOnlyMode tests getopt_long_only functionality.
func TestGNULongOnlyMode(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}

	tests := []struct {
		name      string
		optstring string
		args      []string
		expected  []Option
		expectErr bool
	}{
		{
			name:      "single dash long option",
			optstring: "",
			args:      []string{"-verbose"},
			expected: []Option{
				{Name: "verbose", HasArg: false, Arg: ""},
			},
		},
		{
			name:      "double dash still works",
			optstring: "",
			args:      []string{"--verbose"},
			expected: []Option{
				{Name: "verbose", HasArg: false, Arg: ""},
			},
		},
		{
			name:      "single dash with argument",
			optstring: "",
			args:      []string{"-output=file.txt"},
			expected: []Option{
				{Name: "output", HasArg: true, Arg: "file.txt"},
			},
		},
		{
			name:      "non-empty optstring with short fallback",
			optstring: "v",
			args:      []string{"-v"},
			expected: []Option{
				{Name: "v", HasArg: false, Arg: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOptLongOnly(tt.args, tt.optstring, longOpts)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetOptLongOnly failed: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, parser), tt.expected)
		})
	}
}

// TestGNULongOptionComplexNames tests long options with complex names.
func TestGNULongOptionComplexNames(t *testing.T) {
	longOpts := []Flag{
		{Name: "foo=bar", HasArg: NoArgument},
		{Name: "with-dashes", HasArg: RequiredArgument},
		{Name: "under_scores", HasArg: OptionalArgument},
		{Name: "123numbers", HasArg: NoArgument},
	}

	tests := []struct {
		name     string
		args     []string
		expected []Option
	}{
		{
			name: "option name with equals",
			args: []string{"--foo=bar"},
			expected: []Option{
				{Name: "foo=bar", HasArg: false, Arg: ""},
			},
		},
		{
			name: "option name with dashes",
			args: []string{"--with-dashes=value"},
			expected: []Option{
				{Name: "with-dashes", HasArg: true, Arg: "value"},
			},
		},
		{
			name: "option name with underscores",
			args: []string{"--under_scores=value"},
			expected: []Option{
				{Name: "under_scores", HasArg: true, Arg: "value"},
			},
		},
		{
			name: "option name with numbers",
			args: []string{"--123numbers"},
			expected: []Option{
				{Name: "123numbers", HasArg: false, Arg: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOptLong(tt.args, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong failed: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, parser), tt.expected)
		})
	}
}

// TestGNUMixedShortLongOptions tests mixing short and long options.
func TestGNUMixedShortLongOptions(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}

	tests := []struct {
		name      string
		optstring string
		args      []string
		expected  []Option
	}{
		{
			name:      "mixed short and long options",
			optstring: "vo:",
			args:      []string{"-v", "--output=file.txt", "-o", "other.txt"},
			expected: []Option{
				{Name: "v", HasArg: false, Arg: ""},
				{Name: "output", HasArg: true, Arg: "file.txt"},
				{Name: "o", HasArg: true, Arg: "other.txt"},
			},
		},
		{
			name:      "compacted short options with long options",
			optstring: "abc",
			args:      []string{"-ab", "--verbose", "-c"},
			expected: []Option{
				{Name: "a", HasArg: false, Arg: ""},
				{Name: "b", HasArg: false, Arg: ""},
				{Name: "verbose", HasArg: false, Arg: ""},
				{Name: "c", HasArg: false, Arg: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOptLong(tt.args, tt.optstring, longOpts)
			if err != nil {
				t.Fatalf("GetOptLong failed: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, parser), tt.expected)
		})
	}
}

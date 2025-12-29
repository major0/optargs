package optargs

import (
	"os"
	"testing"
)

// TestPOSIXShortOptionCompaction tests POSIX-compliant short option compaction
// This validates that -abc is equivalent to -a -b -c
func TestPOSIXShortOptionCompaction(t *testing.T) {
	tests := []struct {
		name      string
		optstring string
		args      []string
		expected  []Option
	}{
		{
			name:      "basic compaction no args",
			optstring: "abc",
			args:      []string{"-abc"},
			expected: []Option{
				{Name: "a", HasArg: false, Arg: ""},
				{Name: "b", HasArg: false, Arg: ""},
				{Name: "c", HasArg: false, Arg: ""},
			},
		},
		{
			name:      "compaction with required arg",
			optstring: "ab:c",
			args:      []string{"-abfoo"},
			expected: []Option{
				{Name: "a", HasArg: false, Arg: ""},
				{Name: "b", HasArg: true, Arg: "foo"},
			},
		},
		{
			name:      "compaction with optional arg",
			optstring: "ab::c",
			args:      []string{"-abfoo"},
			expected: []Option{
				{Name: "a", HasArg: false, Arg: ""},
				{Name: "b", HasArg: true, Arg: "foo"},
			},
		},
		{
			name:      "compaction with optional arg empty",
			optstring: "ab::c",
			args:      []string{"-ab"},
			expected: []Option{
				{Name: "a", HasArg: false, Arg: ""},
				{Name: "b", HasArg: false, Arg: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt failed: %v", err)
			}

			var options []Option
			for opt, err := range parser.Options() {
				if err != nil {
					t.Fatalf("Options iteration failed: %v", err)
				}
				options = append(options, opt)
			}

			if len(options) != len(tt.expected) {
				t.Fatalf("Expected %d options, got %d", len(tt.expected), len(options))
			}

			for i, expected := range tt.expected {
				if options[i].Name != expected.Name {
					t.Errorf("Option %d: expected name %s, got %s", i, expected.Name, options[i].Name)
				}
				if options[i].HasArg != expected.HasArg {
					t.Errorf("Option %d: expected HasArg %t, got %t", i, expected.HasArg, options[i].HasArg)
				}
				if options[i].Arg != expected.Arg {
					t.Errorf("Option %d: expected arg %s, got %s", i, expected.Arg, options[i].Arg)
				}
			}
		})
	}
}

// TestPOSIXArgumentHandling tests POSIX-compliant argument handling
func TestPOSIXArgumentHandling(t *testing.T) {
	tests := []struct {
		name      string
		optstring string
		args      []string
		expected  []Option
		expectErr bool
	}{
		{
			name:      "required argument provided inline",
			optstring: "a:",
			args:      []string{"-afoo"},
			expected:  []Option{{Name: "a", HasArg: true, Arg: "foo"}},
		},
		{
			name:      "required argument provided separate",
			optstring: "a:",
			args:      []string{"-a", "foo"},
			expected:  []Option{{Name: "a", HasArg: true, Arg: "foo"}},
		},
		{
			name:      "required argument missing",
			optstring: ":a:", // silent error mode
			args:      []string{"-a"},
			expected:  []Option{{Name: "a", HasArg: false, Arg: ""}},
			expectErr: true,
		},
		{
			name:      "optional argument provided inline",
			optstring: "a::",
			args:      []string{"-afoo"},
			expected:  []Option{{Name: "a", HasArg: true, Arg: "foo"}},
		},
		{
			name:      "optional argument not provided",
			optstring: "a::",
			args:      []string{"-a"},
			expected:  []Option{{Name: "a", HasArg: false, Arg: ""}},
		},
		{
			name:      "negative argument accepted",
			optstring: "a:",
			args:      []string{"-a", "-123"},
			expected:  []Option{{Name: "a", HasArg: true, Arg: "-123"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt failed: %v", err)
			}

			var options []Option
			var optErr error
			for opt, err := range parser.Options() {
				if err != nil {
					optErr = err
					if !tt.expectErr {
						t.Fatalf("Unexpected error: %v", err)
					}
				}
				options = append(options, opt)
			}

			if tt.expectErr && optErr == nil {
				t.Fatal("Expected error but got none")
			}

			if len(options) != len(tt.expected) {
				t.Fatalf("Expected %d options, got %d", len(tt.expected), len(options))
			}

			for i, expected := range tt.expected {
				if options[i].Name != expected.Name {
					t.Errorf("Option %d: expected name %s, got %s", i, expected.Name, options[i].Name)
				}
				if options[i].HasArg != expected.HasArg {
					t.Errorf("Option %d: expected HasArg %t, got %t", i, expected.HasArg, options[i].HasArg)
				}
				if options[i].Arg != expected.Arg {
					t.Errorf("Option %d: expected arg %s, got %s", i, expected.Arg, options[i].Arg)
				}
			}
		})
	}
}

// TestPOSIXOptionTermination tests POSIX -- termination behavior
func TestPOSIXOptionTermination(t *testing.T) {
	tests := []struct {
		name          string
		optstring     string
		args          []string
		expected      []Option
		remainingArgs []string
	}{
		{
			name:          "double dash stops parsing",
			optstring:     "abc",
			args:          []string{"-a", "--", "-b", "-c"},
			expected:      []Option{{Name: "a", HasArg: false, Arg: ""}},
			remainingArgs: []string{"-b", "-c"},
		},
		{
			name:          "double dash with no options before",
			optstring:     "abc",
			args:          []string{"--", "-a", "-b"},
			expected:      []Option{},
			remainingArgs: []string{"-a", "-b"},
		},
		{
			name:          "double dash with arguments",
			optstring:     "a:",
			args:          []string{"-a", "foo", "--", "-a", "bar"},
			expected:      []Option{{Name: "a", HasArg: true, Arg: "foo"}},
			remainingArgs: []string{"-a", "bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt failed: %v", err)
			}

			var options []Option
			for opt, err := range parser.Options() {
				if err != nil {
					t.Fatalf("Options iteration failed: %v", err)
				}
				options = append(options, opt)
			}

			if len(options) != len(tt.expected) {
				t.Fatalf("Expected %d options, got %d", len(tt.expected), len(options))
			}

			for i, expected := range tt.expected {
				if options[i].Name != expected.Name {
					t.Errorf("Option %d: expected name %s, got %s", i, expected.Name, options[i].Name)
				}
				if options[i].HasArg != expected.HasArg {
					t.Errorf("Option %d: expected HasArg %t, got %t", i, expected.HasArg, options[i].HasArg)
				}
				if options[i].Arg != expected.Arg {
					t.Errorf("Option %d: expected arg %s, got %s", i, expected.Arg, options[i].Arg)
				}
			}

			if len(parser.Args) != len(tt.remainingArgs) {
				t.Fatalf("Expected %d remaining args, got %d", len(tt.remainingArgs), len(parser.Args))
			}

			for i, expected := range tt.remainingArgs {
				if parser.Args[i] != expected {
					t.Errorf("Remaining arg %d: expected %s, got %s", i, expected, parser.Args[i])
				}
			}
		})
	}
}

// TestPOSIXParseMode tests different POSIX parsing modes
func TestPOSIXParseMode(t *testing.T) {
	tests := []struct {
		name          string
		optstring     string
		args          []string
		expected      []Option
		remainingArgs []string
	}{
		{
			name:          "default mode reorders arguments",
			optstring:     "a",
			args:          []string{"file1", "-a", "file2"},
			expected:      []Option{{Name: "a", HasArg: false, Arg: ""}},
			remainingArgs: []string{"file1", "file2"},
		},
		{
			name:          "posixly correct mode stops at first non-option",
			optstring:     "+a",
			args:          []string{"file1", "-a", "file2"},
			expected:      []Option{},
			remainingArgs: []string{"file1", "-a", "file2"},
		},
		{
			name:      "non-opts mode treats non-options as arguments to option 1",
			optstring: "-a",
			args:      []string{"-a", "file1", "-a"},
			expected: []Option{
				{Name: "a", HasArg: false, Arg: ""},
				{Name: string(byte(1)), HasArg: false, Arg: "file1"},
				{Name: "a", HasArg: false, Arg: ""},
			},
			remainingArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt failed: %v", err)
			}

			var options []Option
			for opt, err := range parser.Options() {
				if err != nil {
					t.Fatalf("Options iteration failed: %v", err)
				}
				options = append(options, opt)
			}

			if len(options) != len(tt.expected) {
				t.Fatalf("Expected %d options, got %d", len(tt.expected), len(options))
			}

			for i, expected := range tt.expected {
				if options[i].Name != expected.Name {
					t.Errorf("Option %d: expected name %s, got %s", i, expected.Name, options[i].Name)
				}
				if options[i].HasArg != expected.HasArg {
					t.Errorf("Option %d: expected HasArg %t, got %t", i, expected.HasArg, options[i].HasArg)
				}
				if options[i].Arg != expected.Arg {
					t.Errorf("Option %d: expected arg %s, got %s", i, expected.Arg, options[i].Arg)
				}
			}

			if len(parser.Args) != len(tt.remainingArgs) {
				t.Fatalf("Expected %d remaining args, got %d", len(tt.remainingArgs), len(parser.Args))
			}

			for i, expected := range tt.remainingArgs {
				if parser.Args[i] != expected {
					t.Errorf("Remaining arg %d: expected %s, got %s", i, expected, parser.Args[i])
				}
			}
		})
	}
}

// TestPOSIXErrorHandling tests POSIX-compliant error handling
func TestPOSIXErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		optstring  string
		args       []string
		expectErr  bool
		silentMode bool
	}{
		{
			name:       "unknown option with error reporting",
			optstring:  "a",
			args:       []string{"-b"},
			expectErr:  true,
			silentMode: false,
		},
		{
			name:       "unknown option with silent mode",
			optstring:  ":a",
			args:       []string{"-b"},
			expectErr:  true,
			silentMode: true,
		},
		{
			name:       "missing required argument with error reporting",
			optstring:  "a:",
			args:       []string{"-a"},
			expectErr:  true,
			silentMode: false,
		},
		{
			name:       "missing required argument with silent mode",
			optstring:  ":a:",
			args:       []string{"-a"},
			expectErr:  true,
			silentMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt failed: %v", err)
			}

			var optErr error
			for _, err := range parser.Options() {
				if err != nil {
					optErr = err
					break
				}
			}

			if tt.expectErr && optErr == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.expectErr && optErr != nil {
				t.Fatalf("Unexpected error: %v", optErr)
			}

			// Verify silent mode configuration
			if tt.silentMode && parser.config.enableErrors {
				t.Error("Expected silent mode but error reporting is enabled")
			}
			if !tt.silentMode && !parser.config.enableErrors {
				t.Error("Expected error reporting but silent mode is enabled")
			}
		})
	}
}

// TestGNUExtensions tests GNU extensions to POSIX
func TestGNUExtensions(t *testing.T) {
	tests := []struct {
		name      string
		optstring string
		args      []string
		expected  []Option
	}{
		{
			name:      "GNU W extension",
			optstring: "W;a",
			args:      []string{"-W", "foo"},
			expected:  []Option{{Name: "foo", HasArg: true, Arg: "foo"}},
		},
		{
			name:      "GNU W extension with argument",
			optstring: "W;a:",
			args:      []string{"-W", "a=bar"},
			expected:  []Option{{Name: "a=bar", HasArg: true, Arg: "a=bar"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt failed: %v", err)
			}

			var options []Option
			for opt, err := range parser.Options() {
				if err != nil {
					t.Fatalf("Options iteration failed: %v", err)
				}
				options = append(options, opt)
			}

			if len(options) != len(tt.expected) {
				t.Fatalf("Expected %d options, got %d", len(tt.expected), len(options))
			}

			for i, expected := range tt.expected {
				if options[i].Name != expected.Name {
					t.Errorf("Option %d: expected name %s, got %s", i, expected.Name, options[i].Name)
				}
				if options[i].HasArg != expected.HasArg {
					t.Errorf("Option %d: expected HasArg %t, got %t", i, expected.HasArg, options[i].HasArg)
				}
				if options[i].Arg != expected.Arg {
					t.Errorf("Option %d: expected arg %s, got %s", i, expected.Arg, options[i].Arg)
				}
			}
		})
	}
}

// TestPOSIXCharacterValidation tests POSIX character validation rules
func TestPOSIXCharacterValidation(t *testing.T) {
	// Test valid printable ASCII characters (excluding reserved ones)
	for i := 33; i <= 126; i++ { // printable ASCII range
		c := byte(i)
		// Skip reserved characters and prefix characters
		if c == ':' || c == ';' || c == '-' || c == '+' {
			continue
		}

		optstring := string(c)
		args := []string{"-" + string(c)}

		parser, err := GetOpt(args, optstring)
		if err != nil {
			t.Errorf("Valid character %c should be accepted, got error: %v", c, err)
			continue
		}

		var options []Option
		for opt, err := range parser.Options() {
			if err != nil {
				t.Errorf("Valid character %c parsing failed: %v", c, err)
				break
			}
			options = append(options, opt)
		}

		if len(options) != 1 || options[0].Name != string(c) {
			t.Errorf("Valid character %c not parsed correctly", c)
		}
	}

	// Test invalid characters as option characters
	invalidTests := []struct {
		optstring  string
		shouldFail bool
		desc       string
	}{
		{";", true, "semicolon as option character"},
		{"a-", true, "dash as option character"},
		{"a;", true, "semicolon as option character"},
		{"a:", false, "colon as argument modifier (valid)"},
		{":", false, "colon as prefix (valid silent mode)"},
		{"+a", false, "plus as prefix (valid)"},
		{"-a", false, "dash as prefix (valid)"},
	}

	for _, test := range invalidTests {
		_, err := GetOpt(nil, test.optstring)
		if test.shouldFail && err == nil {
			t.Errorf("%s should be rejected", test.desc)
		}
		if !test.shouldFail && err != nil {
			t.Errorf("%s should be accepted, got error: %v", test.desc, err)
		}
	}
}

// TestPOSIXLYCORRECTEnvironmentVariable tests POSIXLY_CORRECT environment variable behavior
func TestPOSIXLYCORRECTEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name          string
		optstring     string
		args          []string
		envValue      string
		expected      []Option
		remainingArgs []string
	}{
		{
			name:          "without POSIXLY_CORRECT processes all options",
			optstring:     "a",
			args:          []string{"file1", "-a", "file2"},
			envValue:      "",
			expected:      []Option{{Name: "a", HasArg: false, Arg: ""}},
			remainingArgs: []string{"file1", "file2"},
		},
		{
			name:          "with POSIXLY_CORRECT stops at first non-option",
			optstring:     "a",
			args:          []string{"file1", "-a", "file2"},
			envValue:      "1",
			expected:      []Option{},
			remainingArgs: []string{"file1", "-a", "file2"},
		},
		{
			name:      "POSIXLY_CORRECT with options first",
			optstring: "ab",
			args:      []string{"-a", "-b", "file1", "-a"},
			envValue:  "1",
			expected: []Option{
				{Name: "a", HasArg: false, Arg: ""},
				{Name: "b", HasArg: false, Arg: ""},
			},
			remainingArgs: []string{"file1", "-a"},
		},
		{
			name:          "plus prefix overrides environment variable",
			optstring:     "+a",
			args:          []string{"file1", "-a", "file2"},
			envValue:      "", // Even without env var, + prefix should work
			expected:      []Option{},
			remainingArgs: []string{"file1", "-a", "file2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment variable
			originalValue := os.Getenv("POSIXLY_CORRECT")
			defer func() {
				if originalValue == "" {
					os.Unsetenv("POSIXLY_CORRECT")
				} else {
					os.Setenv("POSIXLY_CORRECT", originalValue)
				}
			}()

			// Set test environment variable
			if tt.envValue == "" {
				os.Unsetenv("POSIXLY_CORRECT")
			} else {
				os.Setenv("POSIXLY_CORRECT", tt.envValue)
			}

			parser, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt failed: %v", err)
			}

			var options []Option
			for opt, err := range parser.Options() {
				if err != nil {
					t.Fatalf("Options iteration failed: %v", err)
				}
				options = append(options, opt)
			}

			if len(options) != len(tt.expected) {
				t.Fatalf("Expected %d options, got %d", len(tt.expected), len(options))
			}

			for i, expected := range tt.expected {
				if options[i].Name != expected.Name {
					t.Errorf("Option %d: expected name %s, got %s", i, expected.Name, options[i].Name)
				}
				if options[i].HasArg != expected.HasArg {
					t.Errorf("Option %d: expected HasArg %t, got %t", i, expected.HasArg, options[i].HasArg)
				}
				if options[i].Arg != expected.Arg {
					t.Errorf("Option %d: expected arg %s, got %s", i, expected.Arg, options[i].Arg)
				}
			}

			if len(parser.Args) != len(tt.remainingArgs) {
				t.Fatalf("Expected %d remaining args, got %d", len(tt.remainingArgs), len(parser.Args))
			}

			for i, expected := range tt.remainingArgs {
				if parser.Args[i] != expected {
					t.Errorf("Remaining arg %d: expected %s, got %s", i, expected, parser.Args[i])
				}
			}
		})
	}
}

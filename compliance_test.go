package optargs

import (
	"testing"
)

// ---------------------------------------------------------------------------
// POSIX compliance tests
// ---------------------------------------------------------------------------

// TestPOSIXShortOptionCompaction tests POSIX-compliant short option compaction.
// This validates that -abc is equivalent to -a -b -c.
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
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
			},
		},
		{
			name:      "compaction with required arg",
			optstring: "ab:c",
			args:      []string{"-abfoo"},
			expected: []Option{
				{Name: "a"},
				{Name: "b", HasArg: true, Arg: "foo"},
			},
		},
		{
			name:      "compaction with optional arg",
			optstring: "ab::c",
			args:      []string{"-abfoo"},
			expected: []Option{
				{Name: "a"},
				{Name: "b", HasArg: true, Arg: "foo"},
			},
		},
		{
			name:      "compaction with optional arg empty",
			optstring: "ab::c",
			args:      []string{"-ab"},
			expected: []Option{
				{Name: "a"},
				{Name: "b"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt failed: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, parser), tt.expected)
		})
	}
}

// TestPOSIXArgumentHandling tests POSIX-compliant argument handling.
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
			expected:  []Option{{Name: "a"}},
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
			expected:  []Option{{Name: "a"}},
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
						t.Fatalf("unexpected error: %v", err)
					}
				}
				options = append(options, opt)
			}

			if tt.expectErr && optErr == nil {
				t.Fatal("expected error but got none")
			}

			assertOptions(t, options, tt.expected)
		})
	}
}

// TestPOSIXOptionTermination tests POSIX -- termination behavior.
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
			expected:      []Option{{Name: "a"}},
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
			assertOptions(t, requireParsedOptions(t, parser), tt.expected)
			assertArgs(t, parser.Args, tt.remainingArgs)
		})
	}
}

// TestPOSIXParseMode tests different POSIX parsing modes.
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
			expected:      []Option{{Name: "a"}},
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
				{Name: "a"},
				{Name: string(byte(1)), Arg: "file1"},
				{Name: "a"},
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
			assertOptions(t, requireParsedOptions(t, parser), tt.expected)
			assertArgs(t, parser.Args, tt.remainingArgs)
		})
	}
}

// TestPOSIXErrorHandling tests POSIX-compliant error handling.
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
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && optErr != nil {
				t.Fatalf("unexpected error: %v", optErr)
			}

			if tt.silentMode && parser.config.enableErrors {
				t.Error("expected silent mode but error reporting is enabled")
			}
			if !tt.silentMode && !parser.config.enableErrors {
				t.Error("expected error reporting but silent mode is enabled")
			}
		})
	}
}

// TestGNUExtensions tests GNU extensions to POSIX.
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
			assertOptions(t, requireParsedOptions(t, parser), tt.expected)
		})
	}
}

// TestPOSIXCharacterValidation tests POSIX character validation rules.
func TestPOSIXCharacterValidation(t *testing.T) {
	// Test valid printable ASCII characters (excluding reserved ones)
	for i := 33; i <= 126; i++ { // printable ASCII range
		c := byte(i)
		if c == ':' || c == ';' || c == '-' || c == '+' {
			continue
		}

		parser, err := GetOpt([]string{"-" + string(c)}, string(c))
		if err != nil {
			t.Errorf("valid character %c should be accepted, got error: %v", c, err)
			continue
		}

		options := requireParsedOptions(t, parser)
		if len(options) != 1 || options[0].Name != string(c) {
			t.Errorf("valid character %c not parsed correctly", c)
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

// TestPOSIXLYCORRECTEnvironmentVariable tests POSIXLY_CORRECT environment
// variable behavior.
// ---------------------------------------------------------------------------
// GNU long option compliance tests
// ---------------------------------------------------------------------------

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
			requireFirstOptError(t, parser, tt.expectErr)
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
			requireFirstOptError(t, parser, tt.expectErr)
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

// ---------------------------------------------------------------------------
// Equals delimiter compliance tests
// ---------------------------------------------------------------------------

// TestEqualsDelimiterStripping verifies that the = delimiter is not
// included in the arg value when using --option=value syntax.
func TestEqualsDelimiterStripping(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		longOpts []Flag
		expected []Option
	}{
		{
			name:     "required arg with equals",
			args:     []string{"--file=input.txt"},
			longOpts: []Flag{{Name: "file", HasArg: RequiredArgument}},
			expected: []Option{{Name: "file", Arg: "input.txt", HasArg: true}},
		},
		{
			name:     "optional arg with equals",
			args:     []string{"--config=debug"},
			longOpts: []Flag{{Name: "config", HasArg: OptionalArgument}},
			expected: []Option{{Name: "config", Arg: "debug", HasArg: true}},
		},
		{
			name:     "empty arg with equals",
			args:     []string{"--output="},
			longOpts: []Flag{{Name: "output", HasArg: RequiredArgument}},
			expected: []Option{{Name: "output", Arg: "", HasArg: true}},
		},
		{
			name:     "negative number arg",
			args:     []string{"--count=-5"},
			longOpts: []Flag{{Name: "count", HasArg: RequiredArgument}},
			expected: []Option{{Name: "count", Arg: "-5", HasArg: true}},
		},
		{
			name:     "arg containing multiple equals",
			args:     []string{"--query=key=value=extra"},
			longOpts: []Flag{{Name: "query", HasArg: RequiredArgument}},
			expected: []Option{{Name: "query", Arg: "key=value=extra", HasArg: true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong(tt.args, "", tt.longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, p), tt.expected)
		})
	}
}

// TestEdgeCaseMalformedOptions tests malformed option inputs.
func TestEdgeCaseMalformedOptions(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		optstring string
		expectErr bool
	}{
		{
			name:      "double dash terminates parsing",
			args:      []string{"--"},
			optstring: "a",
		},
		{
			name:      "triple dash errors on invalid dash character",
			args:      []string{"---"},
			optstring: "a",
			expectErr: true,
		},
		{
			name:      "unknown long option with equals but no value",
			args:      []string{"--opt="},
			optstring: "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireParseError(t, tt.args, tt.optstring, nil)
			if tt.expectErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestEdgeCaseArgumentHandling tests edge cases in argument handling.
func TestEdgeCaseArgumentHandling(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		optstring string
		expectErr bool
	}{
		{
			name:      "missing required argument at end of args",
			args:      []string{"-a"},
			optstring: ":a:",
			expectErr: true,
		},
		{
			name:      "empty string is valid argument",
			args:      []string{"-a", ""},
			optstring: "a:",
		},
		{
			name:      "whitespace is valid argument",
			args:      []string{"-a", "   "},
			optstring: "a:",
		},
		{
			name:      "special characters are valid arguments",
			args:      []string{"-a", "!@#$%^&*()"},
			optstring: "a:",
		},
		{
			name:      "unicode is valid argument",
			args:      []string{"-a", "café"},
			optstring: "a:",
		},
		{
			name:      "very long argument",
			args:      []string{"-a", string(make([]byte, 10000))},
			optstring: "a:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireParseError(t, tt.args, tt.optstring, nil)
			if tt.expectErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

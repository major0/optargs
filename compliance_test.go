package optargs

import (
	"os"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// requireParsedOptions collects all options from a parser, failing the test on
// any iteration error.
func requireParsedOptions(t *testing.T, parser *Parser) []Option {
	t.Helper()
	var options []Option
	for opt, err := range parser.Options() {
		if err != nil {
			t.Fatalf("Options iteration failed: %v", err)
		}
		options = append(options, opt)
	}
	return options
}

// assertOptions compares actual options against expected, checking Name,
// HasArg, and Arg for each element.
func assertOptions(t *testing.T, got, want []Option) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d options, got %d", len(want), len(got))
	}
	for i, w := range want {
		if got[i].Name != w.Name {
			t.Errorf("option %d: expected name %s, got %s", i, w.Name, got[i].Name)
		}
		if got[i].HasArg != w.HasArg {
			t.Errorf("option %d: expected HasArg %t, got %t", i, w.HasArg, got[i].HasArg)
		}
		if got[i].Arg != w.Arg {
			t.Errorf("option %d: expected arg %s, got %s", i, w.Arg, got[i].Arg)
		}
	}
}

// assertArgs compares remaining positional arguments against expected values.
func assertArgs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d remaining args, got %d", len(want), len(got))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("remaining arg %d: expected %s, got %s", i, w, got[i])
		}
	}
}

// requireFirstOptError iterates parser options and returns the first error, if
// any. It fails the test only when the error expectation is violated.
func requireFirstOptError(t *testing.T, parser *Parser, expectErr bool) {
	t.Helper()
	var optErr error
	for _, err := range parser.Options() {
		if err != nil {
			optErr = err
			break
		}
	}
	if expectErr && optErr == nil {
		t.Fatal("expected error but got none")
	}
	if !expectErr && optErr != nil {
		t.Fatalf("unexpected error: %v", optErr)
	}
}

// setupChain creates a parent→child parser chain. Parent gets empty args;
// child gets the provided args. Returns the child parser.
func setupChain(t *testing.T, parentOpts, childOpts []Flag, childArgs []string) *Parser {
	t.Helper()
	parent, err := GetOptLong([]string{}, "", parentOpts)
	if err != nil {
		t.Fatalf("parent: %v", err)
	}
	child, err := GetOptLong(childArgs, "", childOpts)
	if err != nil {
		t.Fatalf("child: %v", err)
	}
	parent.AddCmd("sub", child)
	return child
}

// setupChain3 creates a grandparent→parent→child parser chain. Only the
// child receives args. Returns the child parser.
func setupChain3(t *testing.T, gpOpts, parOpts, childOpts []Flag, childArgs []string) *Parser {
	t.Helper()
	gp, err := GetOptLong([]string{}, "", gpOpts)
	if err != nil {
		t.Fatalf("grandparent: %v", err)
	}
	par, err := GetOptLong([]string{}, "", parOpts)
	if err != nil {
		t.Fatalf("parent: %v", err)
	}
	child, err := GetOptLong(childArgs, "", childOpts)
	if err != nil {
		t.Fatalf("child: %v", err)
	}
	gp.AddCmd("mid", par)
	par.AddCmd("leaf", child)
	return child
}

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
			expected:      []Option{{Name: "a"}},
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
				{Name: "a"},
				{Name: "b"},
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
			originalValue := os.Getenv("POSIXLY_CORRECT")
			defer func() {
				if originalValue == "" {
					_ = os.Unsetenv("POSIXLY_CORRECT")
				} else {
					_ = os.Setenv("POSIXLY_CORRECT", originalValue)
				}
			}()

			if tt.envValue == "" {
				_ = os.Unsetenv("POSIXLY_CORRECT")
			} else {
				_ = os.Setenv("POSIXLY_CORRECT", tt.envValue)
			}

			parser, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt failed: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, parser), tt.expected)
			assertArgs(t, parser.Args, tt.remainingArgs)
		})
	}
}

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

// TestOverlappingOptionNames verifies longest-prefix-first matching when
// multiple registered option names share a common prefix.
func TestOverlappingOptionNames(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		longOpts []Flag
		expected []Option
	}{
		{
			// "foo" and "foobar" registered. Input: --foo=val
			// "foo" is exact prefix + '=' boundary → match "foo", arg "val"
			name: "exact_match_wins_over_prefix",
			args: []string{"--foo=val"},
			longOpts: []Flag{
				{Name: "foo", HasArg: RequiredArgument},
				{Name: "foobar", HasArg: RequiredArgument},
			},
			expected: []Option{{Name: "foo", Arg: "val", HasArg: true}},
		},
		{
			// "foo" and "foobar" registered. Input: --foobar=val
			// "foobar" is longest prefix + '=' boundary → match "foobar"
			name: "longer_prefix_wins_at_equals_boundary",
			args: []string{"--foobar=val"},
			longOpts: []Flag{
				{Name: "foo", HasArg: RequiredArgument},
				{Name: "foobar", HasArg: RequiredArgument},
			},
			expected: []Option{{Name: "foobar", Arg: "val", HasArg: true}},
		},
		{
			// "foo" (RequiredArgument) and "foobar" (NoArgument) registered.
			// Input: --foo=baz → "foo" matches at '=' boundary, arg "baz"
			name: "fallback_to_shorter_when_no_equals_boundary",
			args: []string{"--foo=baz"},
			longOpts: []Flag{
				{Name: "foo", HasArg: RequiredArgument},
				{Name: "foobar", HasArg: NoArgument},
			},
			expected: []Option{{Name: "foo", Arg: "baz", HasArg: true}},
		},
		{
			// "o", "out", "output" registered. Input: --output=file.txt
			// Longest match "output" at '=' boundary → match "output"
			name: "three_level_prefix_chain",
			args: []string{"--output=file.txt"},
			longOpts: []Flag{
				{Name: "o", HasArg: RequiredArgument},
				{Name: "out", HasArg: RequiredArgument},
				{Name: "output", HasArg: RequiredArgument},
			},
			expected: []Option{{Name: "output", Arg: "file.txt", HasArg: true}},
		},
		{
			// "o", "out", "output" registered. Input: --out=file.txt
			// "out" matches at '=' boundary → match "out"
			name: "three_level_mid_match",
			args: []string{"--out=file.txt"},
			longOpts: []Flag{
				{Name: "o", HasArg: RequiredArgument},
				{Name: "out", HasArg: RequiredArgument},
				{Name: "output", HasArg: RequiredArgument},
			},
			expected: []Option{{Name: "out", Arg: "file.txt", HasArg: true}},
		},
		{
			// "o", "out", "output" registered. Input: --o=file.txt
			// "o" matches at '=' boundary → match "o"
			name: "three_level_shortest_match",
			args: []string{"--o=file.txt"},
			longOpts: []Flag{
				{Name: "o", HasArg: RequiredArgument},
				{Name: "out", HasArg: RequiredArgument},
				{Name: "output", HasArg: RequiredArgument},
			},
			expected: []Option{{Name: "o", Arg: "file.txt", HasArg: true}},
		},
		{
			// "foo" (RequiredArgument), "foo=bar" (NoArgument)
			// Input: --foo=bar → longest "foo=bar" is exact match (NoArgument)
			name: "noarg_longest_skips_to_shorter_with_arg",
			args: []string{"--foo=bar"},
			longOpts: []Flag{
				{Name: "foo", HasArg: RequiredArgument},
				{Name: "foo=bar", HasArg: NoArgument},
			},
			expected: []Option{{Name: "foo=bar", HasArg: false}},
		},
		{
			// "foo=bar" (RequiredArgument) and "foo" (RequiredArgument)
			// Input: --foo=bar=baz → longest match "foo=bar" at '=' boundary, arg "baz"
			name: "equals_in_name_with_arg",
			args: []string{"--foo=bar=baz"},
			longOpts: []Flag{
				{Name: "foo", HasArg: RequiredArgument},
				{Name: "foo=bar", HasArg: RequiredArgument},
			},
			expected: []Option{{Name: "foo=bar", Arg: "baz", HasArg: true}},
		},
		{
			// Only "foo" registered (RequiredArgument)
			// Input: --foo=bar=baz → match "foo", arg "bar=baz"
			name: "shorter_name_when_longer_not_registered",
			args: []string{"--foo=bar=baz"},
			longOpts: []Flag{
				{Name: "foo", HasArg: RequiredArgument},
			},
			expected: []Option{{Name: "foo", Arg: "bar=baz", HasArg: true}},
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

	t.Run("noarg_skips_to_shorter_candidate", func(t *testing.T) {
		// "output" (NoArgument), "out" (RequiredArgument)
		// Input: --output=file → "output" is exact match but NoArgument
		// with '=' present → error.
		longOpts := []Flag{
			{Name: "output", HasArg: NoArgument},
			{Name: "out", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--output=file"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		var gotErr bool
		for _, err := range p.Options() {
			if err != nil {
				gotErr = true
			}
		}
		if !gotErr {
			t.Error("expected error for NoArgument option with =value, got none")
		}
	})
}

// TestSubcommandOverlappingLongOpts verifies longest-prefix matching across
// parent-child parser chains where both parent and child register options
// that share prefixes.
func TestSubcommandOverlappingLongOpts(t *testing.T) {
	tests := []struct {
		name      string
		gpOpts    []Flag // nil means 2-level chain
		parOpts   []Flag // parent opts (or grandparent for 3-level)
		childOpts []Flag
		childArgs []string
		expected  []Option
	}{
		{
			// Parent: "out", Child: "output". Input: --output=file.txt
			// Child's "output" is longest match → match "output"
			name:      "child_longer_prefix_wins_over_parent_shorter",
			parOpts:   []Flag{{Name: "out", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "output", HasArg: RequiredArgument}},
			childArgs: []string{"--output=file.txt"},
			expected:  []Option{{Name: "output", Arg: "file.txt", HasArg: true}},
		},
		{
			// Parent: "output", Child: "out". Input: --output=file.txt
			// Parent's "output" is longest match → match "output"
			name:      "parent_longer_prefix_wins_over_child_shorter",
			parOpts:   []Flag{{Name: "output", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "out", HasArg: RequiredArgument}},
			childArgs: []string{"--output=file.txt"},
			expected:  []Option{{Name: "output", Arg: "file.txt", HasArg: true}},
		},
		{
			// Parent: "output", Child: "out". Input: --out=val
			// "out" matches at '=' boundary → match "out"
			name:      "child_matches_shorter_when_input_is_shorter",
			parOpts:   []Flag{{Name: "output", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "out", HasArg: RequiredArgument}},
			childArgs: []string{"--out=val"},
			expected:  []Option{{Name: "out", Arg: "val", HasArg: true}},
		},
		{
			// GP: "output-format", Parent: "output", Child: "out"
			// Input: --output-format=json → GP's "output-format" is longest
			name:      "three_level_chain_longest_from_grandparent",
			gpOpts:    []Flag{{Name: "output-format", HasArg: RequiredArgument}},
			parOpts:   []Flag{{Name: "output", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "out", HasArg: RequiredArgument}},
			childArgs: []string{"--output-format=json"},
			expected:  []Option{{Name: "output-format", Arg: "json", HasArg: true}},
		},
		{
			// Same chain. Input: --output=json → Parent's "output" matches
			name:      "three_level_chain_mid_match",
			gpOpts:    []Flag{{Name: "output-format", HasArg: RequiredArgument}},
			parOpts:   []Flag{{Name: "output", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "out", HasArg: RequiredArgument}},
			childArgs: []string{"--output=json"},
			expected:  []Option{{Name: "output", Arg: "json", HasArg: true}},
		},
		{
			// Parent: "key=val" (NoArgument), Child: "key" (RequiredArgument)
			// Input: --key=val → Parent's "key=val" is longest exact match
			name:      "equals_in_option_name_across_chain",
			parOpts:   []Flag{{Name: "key=val", HasArg: NoArgument}},
			childOpts: []Flag{{Name: "key", HasArg: RequiredArgument}},
			childArgs: []string{"--key=val"},
			expected:  []Option{{Name: "key=val", HasArg: false}},
		},
		{
			// Parent: "key=val" (RequiredArgument), Child: "key" (RequiredArgument)
			// Input: --key=val=extra → Parent's "key=val" matches, arg "extra"
			name:      "equals_in_name_with_arg_across_chain",
			parOpts:   []Flag{{Name: "key=val", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "key", HasArg: RequiredArgument}},
			childArgs: []string{"--key=val=extra"},
			expected:  []Option{{Name: "key=val", Arg: "extra", HasArg: true}},
		},
		{
			// Both parent and child register "verbose" (NoArgument).
			// Child's own should be found (child is searched first).
			name:      "child_own_option_preferred_when_same_length",
			parOpts:   []Flag{{Name: "verbose", HasArg: NoArgument}},
			childOpts: []Flag{{Name: "verbose", HasArg: NoArgument}},
			childArgs: []string{"--verbose"},
			expected:  []Option{{Name: "verbose", HasArg: false}},
		},
		{
			// Parent: "debug" (RequiredArgument), Child: no overlap
			// Input: --debug=trace → resolved via parent chain
			name:      "parent_only_option_resolved_from_child",
			parOpts:   []Flag{{Name: "debug", HasArg: RequiredArgument}},
			childOpts: []Flag{{Name: "color", HasArg: NoArgument}},
			childArgs: []string{"--debug=trace"},
			expected:  []Option{{Name: "debug", Arg: "trace", HasArg: true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var child *Parser
			if tt.gpOpts != nil {
				child = setupChain3(t, tt.gpOpts, tt.parOpts, tt.childOpts, tt.childArgs)
			} else {
				child = setupChain(t, tt.parOpts, tt.childOpts, tt.childArgs)
			}
			assertOptions(t, requireParsedOptions(t, child), tt.expected)
		})
	}
}

// TestMultipleOptionsWithOverlappingPrefixes exercises a single parser with
// many overlapping prefixes to stress the sort + iterate logic.
func TestMultipleOptionsWithOverlappingPrefixes(t *testing.T) {
	longOpts := []Flag{
		{Name: "v", HasArg: RequiredArgument},
		{Name: "ve", HasArg: RequiredArgument},
		{Name: "ver", HasArg: RequiredArgument},
		{Name: "verb", HasArg: RequiredArgument},
		{Name: "verbo", HasArg: RequiredArgument},
		{Name: "verbos", HasArg: RequiredArgument},
		{Name: "verbose", HasArg: RequiredArgument},
	}

	tests := []struct {
		input    string
		expected []Option
	}{
		{"--v=1", []Option{{Name: "v", Arg: "1", HasArg: true}}},
		{"--ve=2", []Option{{Name: "ve", Arg: "2", HasArg: true}}},
		{"--ver=3", []Option{{Name: "ver", Arg: "3", HasArg: true}}},
		{"--verb=4", []Option{{Name: "verb", Arg: "4", HasArg: true}}},
		{"--verbo=5", []Option{{Name: "verbo", Arg: "5", HasArg: true}}},
		{"--verbos=6", []Option{{Name: "verbos", Arg: "6", HasArg: true}}},
		{"--verbose=7", []Option{{Name: "verbose", Arg: "7", HasArg: true}}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p, err := GetOptLong([]string{tt.input}, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, p), tt.expected)
		})
	}
}

// TestTripleEqualsOverlap exercises three options registered simultaneously
// where each name is a prefix of the next, with '=' embedded in the names:
//
//	"foo"         OptionalArgument  (e.g. --foo=arg)
//	"foo=bar"     OptionalArgument  (e.g. --foo=bar=arg)
//	"foo=bar=arg" NoArgument
//
// Every input must resolve to the longest matching registered option.
func TestTripleEqualsOverlap(t *testing.T) {
	longOpts := []Flag{
		{Name: "foo", HasArg: OptionalArgument},
		{Name: "foo=bar", HasArg: OptionalArgument},
		{Name: "foo=bar=arg", HasArg: NoArgument},
	}

	tests := []struct {
		name     string
		input    string
		expected []Option
	}{
		{
			name:     "exact foo=bar=arg matches NoArgument",
			input:    "--foo=bar=arg",
			expected: []Option{{Name: "foo=bar=arg", HasArg: false}},
		},
		{
			name:     "foo=bar with equals arg",
			input:    "--foo=bar=something",
			expected: []Option{{Name: "foo=bar", Arg: "something", HasArg: true}},
		},
		{
			name:     "foo with equals arg",
			input:    "--foo=qux",
			expected: []Option{{Name: "foo", Arg: "qux", HasArg: true}},
		},
		{
			name:     "foo=bar exact no trailing equals",
			input:    "--foo=bar",
			expected: []Option{{Name: "foo=bar", HasArg: false}},
		},
		{
			name:     "foo exact no trailing equals",
			input:    "--foo",
			expected: []Option{{Name: "foo", HasArg: false}},
		},
		{
			name:     "foo=bar=arg=extra skips NoArgument to foo=bar",
			input:    "--foo=bar=arg=extra",
			expected: []Option{{Name: "foo=bar", Arg: "arg=extra", HasArg: true}},
		},
		{
			name:     "foo=bar= empty arg after foo=bar",
			input:    "--foo=bar=",
			expected: []Option{{Name: "foo=bar", Arg: "", HasArg: true}},
		},
		{
			name:     "foo= empty arg after foo",
			input:    "--foo=",
			expected: []Option{{Name: "foo", Arg: "", HasArg: true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong([]string{tt.input}, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, p), tt.expected)
		})
	}
}

// TestObscureLongOptCharacters exercises long option names containing
// characters that are valid isgraph() but unusual: brackets, braces,
// dots, colons, tildes, etc. Per POSIX/GNU convention, any isgraph()
// character is valid in a long option name.
func TestObscureLongOptCharacters(t *testing.T) {
	tests := []struct {
		name     string
		optName  string
		hasArg   ArgType
		input    []string
		expected []Option
	}{
		// Bracket-style: --config[key]
		{
			name: "brackets space arg", optName: "config[key]",
			hasArg: RequiredArgument, input: []string{"--config[key]", "val"},
			expected: []Option{{Name: "config[key]", Arg: "val", HasArg: true}},
		},
		{
			name: "brackets equals arg", optName: "config[key]",
			hasArg: RequiredArgument, input: []string{"--config[key]=val"},
			expected: []Option{{Name: "config[key]", Arg: "val", HasArg: true}},
		},
		// Brace-style: --data{category.key}
		{
			name: "braces space arg", optName: "data{category.key}",
			hasArg: RequiredArgument, input: []string{"--data{category.key}", "val"},
			expected: []Option{{Name: "data{category.key}", Arg: "val", HasArg: true}},
		},
		{
			name: "braces equals arg", optName: "data{category.key}",
			hasArg: RequiredArgument, input: []string{"--data{category.key}=val"},
			expected: []Option{{Name: "data{category.key}", Arg: "val", HasArg: true}},
		},
		// Colon-style: --command:arg
		{
			name: "colon space arg", optName: "command:arg",
			hasArg: RequiredArgument, input: []string{"--command:arg", "value"},
			expected: []Option{{Name: "command:arg", Arg: "value", HasArg: true}},
		},
		{
			name: "colon equals arg", optName: "command:arg",
			hasArg: RequiredArgument, input: []string{"--command:arg=value"},
			expected: []Option{{Name: "command:arg", Arg: "value", HasArg: true}},
		},
		// Dot-style: --section.key
		{
			name: "dot space arg", optName: "section.key",
			hasArg: RequiredArgument, input: []string{"--section.key", "value"},
			expected: []Option{{Name: "section.key", Arg: "value", HasArg: true}},
		},
		{
			name: "dot equals arg", optName: "section.key",
			hasArg: RequiredArgument, input: []string{"--section.key=value"},
			expected: []Option{{Name: "section.key", Arg: "value", HasArg: true}},
		},
		// Tilde: --path~backup
		{
			name: "tilde space arg", optName: "path~backup",
			hasArg: RequiredArgument, input: []string{"--path~backup", "/tmp"},
			expected: []Option{{Name: "path~backup", Arg: "/tmp", HasArg: true}},
		},
		{
			name: "tilde equals arg", optName: "path~backup",
			hasArg: RequiredArgument, input: []string{"--path~backup=/tmp"},
			expected: []Option{{Name: "path~backup", Arg: "/tmp", HasArg: true}},
		},
		// Plus: --level+1
		{
			name: "plus space arg", optName: "level+1",
			hasArg: RequiredArgument, input: []string{"--level+1", "high"},
			expected: []Option{{Name: "level+1", Arg: "high", HasArg: true}},
		},
		{
			name: "plus equals arg", optName: "level+1",
			hasArg: RequiredArgument, input: []string{"--level+1=high"},
			expected: []Option{{Name: "level+1", Arg: "high", HasArg: true}},
		},
		// At-sign: --user@host
		{
			name: "at space arg", optName: "user@host",
			hasArg: RequiredArgument, input: []string{"--user@host", "root"},
			expected: []Option{{Name: "user@host", Arg: "root", HasArg: true}},
		},
		{
			name: "at equals arg", optName: "user@host",
			hasArg: RequiredArgument, input: []string{"--user@host=root"},
			expected: []Option{{Name: "user@host", Arg: "root", HasArg: true}},
		},
		// NoArgument with obscure chars
		{
			name: "brackets no arg", optName: "flag[x]",
			hasArg: NoArgument, input: []string{"--flag[x]"},
			expected: []Option{{Name: "flag[x]", HasArg: false}},
		},
		// OptionalArgument with obscure chars
		{
			name: "braces optional with equals", optName: "opt{a.b}",
			hasArg: OptionalArgument, input: []string{"--opt{a.b}=yes"},
			expected: []Option{{Name: "opt{a.b}", Arg: "yes", HasArg: true}},
		},
		{
			name: "braces optional without arg", optName: "opt{a.b}",
			hasArg: OptionalArgument, input: []string{"--opt{a.b}"},
			expected: []Option{{Name: "opt{a.b}", HasArg: false}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong(tt.input, "", []Flag{
				{Name: tt.optName, HasArg: tt.hasArg},
			})
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, p), tt.expected)
		})
	}
}

// TestObscureCharOverlappingPrefixes tests longest-prefix matching when
// obscure-character option names overlap with shorter prefixes.
func TestObscureCharOverlappingPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		longOpts []Flag
		args     []string
		expected []Option
	}{
		{
			// "config" and "config[key]" both registered.
			// Input: --config[key]=val → longest match "config[key]"
			name:     "bracket_prefix_overlap",
			longOpts: []Flag{{Name: "config", HasArg: RequiredArgument}, {Name: "config[key]", HasArg: RequiredArgument}},
			args:     []string{"--config[key]=val"},
			expected: []Option{{Name: "config[key]", Arg: "val", HasArg: true}},
		},
		{
			// "config" and "config[key]" both registered.
			// Input: --config=val → match "config"
			name:     "bracket_falls_back_to_shorter",
			longOpts: []Flag{{Name: "config", HasArg: RequiredArgument}, {Name: "config[key]", HasArg: RequiredArgument}},
			args:     []string{"--config=val"},
			expected: []Option{{Name: "config", Arg: "val", HasArg: true}},
		},
		{
			// "cmd" and "cmd:sub" both registered.
			// Input: --cmd:sub=val → longest match "cmd:sub"
			name:     "colon_prefix_overlap",
			longOpts: []Flag{{Name: "cmd", HasArg: RequiredArgument}, {Name: "cmd:sub", HasArg: RequiredArgument}},
			args:     []string{"--cmd:sub=val"},
			expected: []Option{{Name: "cmd:sub", Arg: "val", HasArg: true}},
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

	t.Run("three_level_obscure_overlap_across_chain", func(t *testing.T) {
		// GP: "data{cat.key}", Parent: "data{cat}", Child: "data"
		// Input: --data{cat.key}=val → GP's "data{cat.key}" is longest
		child := setupChain3(t,
			[]Flag{{Name: "data{cat.key}", HasArg: RequiredArgument}},
			[]Flag{{Name: "data{cat}", HasArg: RequiredArgument}},
			[]Flag{{Name: "data", HasArg: RequiredArgument}},
			[]string{"--data{cat.key}=val"},
		)
		assertOptions(t, requireParsedOptions(t, child), []Option{
			{Name: "data{cat.key}", Arg: "val", HasArg: true},
		})
	})
}

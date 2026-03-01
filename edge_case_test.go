package optargs

import (
	"testing"
)

// requireParseError creates a parser and iterates its options, returning
// the first iteration error or nil. It fails the test if parser creation
// itself fails.
func requireParseError(t *testing.T, args []string, optstring string, longOpts []Flag) error {
	t.Helper()

	var parser *Parser
	var err error

	if longOpts != nil {
		parser, err = GetOptLong(args, optstring, longOpts)
	} else {
		parser, err = GetOpt(args, optstring)
	}
	if err != nil {
		return err
	}

	for _, err := range parser.Options() {
		if err != nil {
			return err
		}
	}
	return nil
}

// TestEdgeCaseEmptyInputs tests edge cases with empty inputs.
func TestEdgeCaseEmptyInputs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		optstring string
		longOpts  []Flag
		expectErr bool
	}{
		{
			name:      "empty args and optstring",
			args:      []string{},
			optstring: "",
		},
		{
			name:      "empty args with optstring",
			args:      []string{},
			optstring: "abc",
		},
		{
			name:      "nil args",
			args:      nil,
			optstring: "abc",
		},
		{
			name:      "empty optstring with args",
			args:      []string{"-a"},
			optstring: "",
			expectErr: true,
		},
		{
			name:      "empty long opts",
			args:      []string{"--verbose"},
			optstring: "",
			longOpts:  []Flag{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireParseError(t, tt.args, tt.optstring, tt.longOpts)

			if tt.expectErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
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

// TestEdgeCaseBoundaryValues tests boundary value conditions.
func TestEdgeCaseBoundaryValues(t *testing.T) {
	tests := []struct {
		name      string
		optstring string
		expectErr bool
	}{
		{
			name:      "all valid alphanumeric characters",
			optstring: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},
		{
			name:      "mixed argument types",
			optstring: "a:b::c",
		},
		{
			name:      "too many colons rejected",
			optstring: "a::::",
			expectErr: true,
		},
		{
			name:      "multiple behavior flags",
			optstring: ":+-abc",
		},
		{
			name:      "GNU words extension",
			optstring: "W;abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetOpt(nil, tt.optstring)

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

// TestEdgeCaseLongOptionHandling tests edge cases in long option handling.
func TestEdgeCaseLongOptionHandling(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
		{Name: "a", HasArg: NoArgument},
		{Name: "123", HasArg: NoArgument},
		{Name: "foo-bar", HasArg: NoArgument},
		{Name: "foo_bar", HasArg: NoArgument},
		{Name: "foo=bar", HasArg: NoArgument},
	}

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "single character long option",
			args: []string{"--a"},
		},
		{
			name: "numeric long option",
			args: []string{"--123"},
		},
		{
			name: "hyphenated long option",
			args: []string{"--foo-bar"},
		},
		{
			name: "underscored long option",
			args: []string{"--foo_bar"},
		},
		{
			name: "long option with equals in name",
			args: []string{"--foo=bar"},
		},
		{
			name:      "unknown long option with multiple equals",
			args:      []string{"--foo=bar=baz"},
			expectErr: true,
		},
		{
			name:      "empty long option name with value",
			args:      []string{"--=value"},
			expectErr: true,
		},
		{
			name:      "long option with only equals",
			args:      []string{"--="},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireParseError(t, tt.args, "", longOpts)

			if tt.expectErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestEdgeCaseErrorPropagation tests error propagation paths.
func TestEdgeCaseErrorPropagation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		optstring string
		longOpts  []Flag
		expectErr bool
	}{
		{
			name:      "null character in optstring rejected",
			args:      []string{},
			optstring: "a\x00b",
			expectErr: true,
		},
		{
			name:      "null character in long option name rejected",
			args:      []string{},
			optstring: "",
			longOpts:  []Flag{{Name: "test\x00", HasArg: NoArgument}},
			expectErr: true,
		},
		{
			name:      "short and long options coexist",
			args:      []string{"-v", "--verbose"},
			optstring: "v",
			longOpts:  []Flag{{Name: "verbose", HasArg: NoArgument}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireParseError(t, tt.args, tt.optstring, tt.longOpts)

			if tt.expectErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestEdgeCaseMemoryAndPerformance tests allocation patterns with large inputs.
func TestEdgeCaseMemoryAndPerformance(t *testing.T) {
	t.Run("large argument list", func(t *testing.T) {
		args := make([]string, 1000)
		for i := range args {
			args[i] = "-a"
		}

		parser, err := GetOpt(args, "a")
		if err != nil {
			t.Fatalf("GetOpt: %v", err)
		}

		count := 0
		for _, err := range parser.Options() {
			if err != nil {
				t.Fatalf("Options iteration: %v", err)
			}
			count++
		}

		if count != 1000 {
			t.Errorf("expected 1000 options, got %d", count)
		}
	})

	t.Run("large optstring", func(t *testing.T) {
		buf := make([]byte, 0, 62)
		for c := byte('a'); c <= 'z'; c++ {
			buf = append(buf, c)
		}
		for c := byte('A'); c <= 'Z'; c++ {
			buf = append(buf, c)
		}
		for c := byte('0'); c <= '9'; c++ {
			buf = append(buf, c)
		}

		if _, err := GetOpt(nil, string(buf)); err != nil {
			t.Fatalf("GetOpt: %v", err)
		}
	})

	t.Run("many long options", func(t *testing.T) {
		longOpts := make([]Flag, 100)
		for i := range longOpts {
			longOpts[i] = Flag{
				Name:   string(rune('a'+i%26)) + string(rune('a'+(i/26)%26)),
				HasArg: NoArgument,
			}
		}

		if _, err := GetOptLong(nil, "", longOpts); err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
	})
}

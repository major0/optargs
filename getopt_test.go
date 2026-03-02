package optargs

import (
	"testing"
)

// genShortOpts generates all permutations of the input characters, each
// followed by no colons, one colon, or two colons.
func genShortOpts(input string, index int, current string, result *[]string) {
	if index == len(input) {
		*result = append(*result, current)
		return
	}

	word := current + string(input[index])
	index++

	genShortOpts(input, index, word, result)
	genShortOpts(input, index, word+":", result)
	genShortOpts(input, index, word+"::", result)
}

// allShortOptPermutations returns every colon-suffix permutation of opts.
func allShortOptPermutations(opts string) []string {
	var permutations []string
	genShortOpts(opts, 0, "", &permutations)
	return permutations
}

// TestShortOptsGraph validates that every isgraph() character allowed by
// the spec is usable as a short option.
func TestShortOptsGraph(t *testing.T) {
	for i := 0; i < 127; i++ {
		if !isGraph(byte(i)) {
			continue
		}

		// Disallowed by the spec
		switch byte(i) {
		case ':', ';', '-':
			continue
		}

		// Prefix the optstring with a non-config character so we
		// actually test the character we are passing. POSIX allows
		// overwriting optstring configs for existing characters, so
		// passing "aa" as the optstring is fine.
		optstring := "a" + string(byte(i))

		args := []string{"-" + string(byte(i))}
		getopt, err := GetOpt(args, optstring)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		shortOpt, ok := getopt.shortOpts[byte(i)]
		if !ok {
			t.Errorf("Did not find option %c in shortOpts", byte(i))
		} else if shortOpt.Name != string(byte(i)) {
			t.Errorf("Expected option %c, got %s", byte(i), getopt.shortOpts[byte(i)].Name)
		}
	}
}

// TestShortOpts validates the default parse mode across all optstring
// permutations.
func TestShortOpts(t *testing.T) {
	for _, opts := range allShortOptPermutations("ab") {
		getopt, err := GetOpt(nil, opts)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if getopt.config.parseMode != ParseDefault {
			t.Errorf("Expected parseMode to be %d, got %d", ParseDefault, getopt.config.parseMode)
		}
	}
}

// An empty optstring is required to be supported for POSIX compatibility.
// TestNoOptions validates that each constructor accepts empty/no options.
func TestNoOptions(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"GetOpt empty optstring", func() error {
			_, err := GetOpt(nil, "")
			return err
		}},
		{"GetOptLong no short or long", func() error {
			_, err := GetOptLong(nil, "", nil)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

// A `:` appearing in the optstring prefix before any valid option
// characters disables automatic error reporting by GetOpt(). Per POSIX,
// we consume any number of prefix characters, toggling parser mode and
// error mode settings as we go.
func TestShortOptsDisableErrors(t *testing.T) {
	for _, opts := range allShortOptPermutations("ab") {
		optstring := ":" + opts
		getopt, err := GetOpt(nil, optstring)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if getopt.config.enableErrors {
			t.Errorf("Expected enableErrors to be false, got true")
		}
	}
}

// parseModeTests covers prefix combinations that toggle the parser mode.
// The last prefix character in the string sets the final mode.
var parseModeTests = []struct {
	name   string
	prefix string
	mode   ParseMode
}{
	{name: "plus sets PosixlyCorrect", prefix: "+", mode: ParsePosixlyCorrect},
	{name: "minus-plus sets PosixlyCorrect", prefix: "-+", mode: ParsePosixlyCorrect},
	{name: "minus sets NonOpts", prefix: "-", mode: ParseNonOpts},
	{name: "plus-minus sets NonOpts", prefix: "+-", mode: ParseNonOpts},
}

// TestShortOptsParseMode validates that prefix characters correctly toggle
// the parser mode across all optstring permutations.
func TestShortOptsParseMode(t *testing.T) {
	for _, tt := range parseModeTests {
		t.Run(tt.name, func(t *testing.T) {
			for _, opts := range allShortOptPermutations("ab") {
				optstring := tt.prefix + opts
				getopt, err := GetOpt(nil, optstring)
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if getopt.config.parseMode != tt.mode {
					t.Errorf("optstring %q: expected parseMode %d, got %d", optstring, tt.mode, getopt.config.parseMode)
				}
			}
		})
	}
}

// invalidOptstrings covers optstrings that must be rejected by the parser.
var invalidOptstrings = []struct {
	name      string
	optstring string
}{
	{name: "semicolon prefix", optstring: ";ab:"},
	{name: "dash in options", optstring: "ab-"},
	{name: "semicolon in options", optstring: "ab;"},
	{name: "triple colon", optstring: "a:::"},
}

// TestShortOptsInvalid validates that prohibited optstrings produce errors.
// TestOptstringInvalid validates that prohibited optstrings produce errors
// across all constructors that accept an optstring.
func TestOptstringInvalid(t *testing.T) {
	type constructor struct {
		name string
		fn   func(optstring string) error
	}
	constructors := []constructor{
		{"GetOpt", func(s string) error { _, err := GetOpt(nil, s); return err }},
		{"GetOptLongOnly", func(s string) error { _, err := GetOptLongOnly(nil, s, nil); return err }},
	}

	for _, tt := range invalidOptstrings {
		for _, ctor := range constructors {
			t.Run(ctor.name+"/"+tt.name, func(t *testing.T) {
				if err := ctor.fn(tt.optstring); err == nil {
					t.Errorf("expected error for optstring %q, got nil", tt.optstring)
				}
			})
		}
	}
}

// The `;` is never allowed in the optstring unless it follows `W`.
// This is a GNU extension to POSIX.
func TestShortOptsGnuWords(t *testing.T) {
	for _, opts := range allShortOptPermutations("ab") {
		optstring := opts + "W;"
		getopt, err := GetOpt(nil, optstring)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if !getopt.config.gnuWords {
			t.Errorf("Expected gnuWords to be true, got false")
		}
	}
}

func TestShortOptsFlags(t *testing.T) {
	getopt, err := GetOpt(nil, "ab:c::")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(getopt.shortOpts) != 3 {
		t.Errorf("Expected shortOpts to be 3, got %d", len(getopt.shortOpts))
	}

	for c, opt := range getopt.shortOpts {
		if opt.Name != string(c) {
			t.Errorf("Expected shortOpts[%c].Name to be '%c', got %s", c, c, opt.Name)
		}
	}

	if getopt.shortOpts['a'].HasArg != NoArgument {
		t.Errorf("Expected shortOpts['a'].HasArg to be NoArgument, got %d", getopt.shortOpts['a'].HasArg)
	}

	if getopt.shortOpts['b'].HasArg != RequiredArgument {
		t.Errorf("Expected shortOpts['b'].HasArg to be RequiredArgument, got %d", getopt.shortOpts['b'].HasArg)
	}

	if getopt.shortOpts['c'].HasArg != OptionalArgument {
		t.Errorf("Expected shortOpts['c'].HasArg to be OptionalArgument, got %d", getopt.shortOpts['c'].HasArg)
	}
}

func TestShortOptNoArgIntegration(t *testing.T) {
	args := []string{"-c"}
	getopt, err := GetOpt(args, "abc")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	for opt, err := range getopt.Options() {
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if opt.Name != "c" {
			t.Errorf("Expected option %c, got %s", 'c', opt.Name)
		}
	}
}

func TestShortOptOptionalIntegration(t *testing.T) {
	optstring := "a::"
	tests := []struct {
		label  string
		args   []string
		name   string
		arg    string
		hasArg bool
	}{
		{"inline arg", []string{"-afoo"}, "a", "foo", true},
		{"separate arg", []string{"-a", "bar"}, "a", "bar", true},
		{"negative arg", []string{"-a", "-1"}, "a", "-1", true},
		{"no arg", []string{"-a"}, "a", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			getopt, err := GetOpt(tt.args, optstring)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			for opt, err := range getopt.Options() {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if opt.Name != tt.name {
					t.Errorf("Expected option %s, got %s", tt.name, opt.Name)
				}
				if opt.HasArg != tt.hasArg {
					t.Errorf("Expected HasArg to be %t, got %t", tt.hasArg, opt.HasArg)
				}
				if opt.Arg != tt.arg {
					t.Errorf("Expected Arg to be %s, got %s", tt.arg, opt.Arg)
				}
			}
		})
	}
}

func TestShortOptRequiredIntegration(t *testing.T) {
	// Disable automatic error reporting to avoid polluting test output
	optstring := ":a:"
	tests := []struct {
		label     string
		args      []string
		name      string
		arg       string
		hasArg    bool
		expectErr bool
	}{
		{"inline arg", []string{"-afoo"}, "a", "foo", true, false},
		{"separate arg", []string{"-a", "bar"}, "a", "bar", true, false},
		{"negative arg", []string{"-a", "-1"}, "a", "-1", true, false},
		{"missing required arg", []string{"-a"}, "a", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			getopt, err := GetOpt(tt.args, optstring)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			for opt, err := range getopt.Options() {
				if tt.expectErr && err == nil {
					t.Errorf("Expected an error for args %v", tt.args)
				}
				if !tt.expectErr && err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if opt.Name != tt.name {
					t.Errorf("Expected option %s, got %s", tt.name, opt.Name)
				}
				if opt.HasArg != tt.hasArg {
					t.Errorf("Expected HasArg to be %t, got %t", tt.hasArg, opt.HasArg)
				}
				if opt.Arg != tt.arg {
					t.Errorf("Expected Arg to be %s, got %s", tt.arg, opt.Arg)
				}
			}
		})
	}
}

func BenchmarkShortOpts(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetOpt(nil, ":-+ab:c::W;")
		if err != nil {
			b.Errorf("unexpected error: %s", err)
		}
	}
}

// TestLongOnlyNoShortFallback validates that long-only mode with no short
// options falls back to error on unknown option.
func TestLongOnlyNoShortFallback(t *testing.T) {
	longopts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
	}
	parser, err := GetOptLongOnly([]string{"-unknown"}, ":", longopts)
	if err != nil {
		t.Fatalf("Unexpected parser creation error: %v", err)
	}

	count := 0
	for _, err := range parser.Options() {
		count++
		if err == nil {
			t.Error("expected error for unrecognized long-only option with no short fallback")
		}
	}
	if count == 0 {
		t.Error("expected at least one iteration from Options()")
	}
}

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

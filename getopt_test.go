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
			t.Errorf("Unexpected error: %s", err)
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
			t.Errorf("Unexpected error: %s", err)
		}

		if getopt.config.parseMode != ParseDefault {
			t.Errorf("Expected parseMode to be %d, got %d", ParseDefault, getopt.config.parseMode)
		}
	}
}

// An empty optstring is required to be supported for POSIX compatibility.
func TestShortOptsNone(t *testing.T) {
	_, err := GetOpt(nil, "")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
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
			t.Errorf("Unexpected error: %s", err)
		}

		if getopt.config.enableErrors != false {
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
					t.Errorf("Unexpected error: %s", err)
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
func TestShortOptsInvalid(t *testing.T) {
	for _, tt := range invalidOptstrings {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetOpt(nil, tt.optstring)
			if err == nil {
				t.Errorf("Expected error for optstring %q, got nil", tt.optstring)
			}
		})
	}
}

// The `;` is never allowed in the optstring unless it follows `W`.
// This is a GNU extension to POSIX.
func TestShortOptsGnuWords(t *testing.T) {
	for _, opts := range allShortOptPermutations("ab") {
		optstring := opts + "W;"
		getopt, err := GetOpt(nil, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if getopt.config.gnuWords != true {
			t.Errorf("Expected gnuWords to be true, got false")
		}
	}
}

func TestShortOptsFlags(t *testing.T) {
	getopt, err := GetOpt(nil, "ab:c::")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
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
		t.Errorf("Unexpected error: %s", err)
	}

	for opt, err := range getopt.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if opt.Name != "c" {
			t.Errorf("Expected option %c, got %s", 'c', opt.Name)
		}
	}
}

func TestShortOptOptionalIntegration(t *testing.T) {
	optstring := "a::"
	tests := []struct {
		args   []string
		name   string
		arg    string
		hasArg bool
	}{
		{[]string{"-afoo"}, "a", "foo", true},
		{[]string{"-a", "bar"}, "a", "bar", true},
		{[]string{"-a", "-1"}, "a", "-1", true},
		{[]string{"-a"}, "a", "", false},
	}

	for _, tt := range tests {
		getopt, err := GetOpt(tt.args, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		for opt, err := range getopt.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
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
	}
}

func TestShortOptRequiredIntegration(t *testing.T) {
	// Disable automatic error reporting to avoid polluting test output
	optstring := ":a:"
	tests := []struct {
		args      []string
		name      string
		arg       string
		hasArg    bool
		expectErr bool
	}{
		{[]string{"-afoo"}, "a", "foo", true, false},
		{[]string{"-a", "bar"}, "a", "bar", true, false},
		{[]string{"-a", "-1"}, "a", "-1", true, false},
		{[]string{"-a"}, "a", "", false, true},
	}

	for _, tt := range tests {
		getopt, err := GetOpt(tt.args, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		for opt, err := range getopt.Options() {
			if tt.expectErr && err == nil {
				t.Errorf("Expected an error for args %v", tt.args)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %s", err)
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
	}
}

func BenchmarkShortOpts(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetOpt(nil, ":-+ab:c::W;")
		if err != nil {
			b.Errorf("Unexpected error: %s", err)
		}
	}
}

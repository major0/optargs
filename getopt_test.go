package optargs

import (
	"testing"
)

// Generate all possible permutations of `ab` followed by no colons,
// 1 colon, or 2 colons
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

func GenShortOpts(opts string) []string {
	var permutations []string
	genShortOpts(opts, 0, "", &permutations)
	return permutations
}

// Test ot make certain that "any" allowed isGraph() character is usable
// as a short option
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

		// prefix the optstring with a non-config character so we
		// can _actually_ test the characer we are passing into
		// the optstring. POSIX allows us to overwrite optstring
		// configs for existing characters, so it wont matter if
		// we pass "aa" as the optstring.
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

// Validate our default parse mode
func TestShortOpts(t *testing.T) {
	opstrings := GenShortOpts("ab")

	for _, opts := range opstrings {
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

// A `:` appearing in the optstring prefix _before_ any valid option
// characters will act to disable automatic error reporting by
// `GetOpt()`. Per our POSIX defined behavior, that means we consume
// any number of prefix characters, e.g. `+-:+-:+-:+-+-+-` toggling
// the necessary parser mode and error mode settings as we go.
// Curiously, there is no reserved character for re-enabling error
// reporting or the default parser mode. It is up to the user to not
// mess that up.
func TestShortOptsDisableErrors(t *testing.T) {
	for _, opts := range GenShortOpts("ab") {
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

// Test to make certain we parse through _all_ of the valid prefix
// characters toggling the parser mode. This means we should be
// able to have a prefix string of `+-+-+-` and the last character
// in the string sets the final mode of the parser.
// Note: It is fine to have an optstring that _only_ comprises of
// parser mode flags. This allows us to use the optstring to toggle
// the parser mode for doing `getopt_long_only()` handling.
func TestShortOptsPosixMode(t *testing.T) {
	for _, opts := range GenShortOpts("ab") {
		optstring := "+" + opts // Parse Mode: PosoxlyCorrect
		getopt, err := GetOpt(nil, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if getopt.config.parseMode != ParsePosixlyCorrect {
			t.Errorf("Expected parseMode to be %d, got %d", ParsePosixlyCorrect, getopt.config.parseMode)
		}

		optstring = "-+" + opts // Parse Mode: non-Option parsing
		getopt, err = GetOpt(nil, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if getopt.config.parseMode != ParsePosixlyCorrect {
			t.Errorf("Expected parseMode to be %d, got %d", ParsePosixlyCorrect, getopt.config.parseMode)
		}
	}
}

// Similar to TestShortOptsPosixMode(), only we want the final parser mode
// to be ParseNonOpts.
func TestShortOptsNonOptMode(t *testing.T) {
	for _, opts := range GenShortOpts("ab") {
		optstring := "-" + opts // Parse Mode: non-Option parsing
		getopt, err := GetOpt(nil, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if getopt.config.parseMode != ParseNonOpts {
			t.Errorf("Expected parseMode to be %d, got %d", ParseNonOpts, getopt.config.parseMode)
		}

		optstring = "+-" + opts // Parse Mode: non-Option parsing
		getopt, err = GetOpt(nil, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if getopt.config.parseMode != ParseNonOpts {
			t.Errorf("Expected parseMode to be %d, got %d", ParseNonOpts, getopt.config.parseMode)
		}
	}
}

// Disalllow `;` as a prefix character, or really any character in the
// optstring unless it follows `W`.
func TestShortOptsInvalidPrefix(t *testing.T) {
	_, err := GetOpt(nil, ";ab:")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// Disallow `-` as an option string. POSIX disallows this as it breaks the
// handling of `--` which is reserved for stopping all parsing of the CLI.
func TestShortOptsInvalidChar1(t *testing.T) {
	_, err := GetOpt(nil, "ab-")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// Disallow `;` as an option string unless it follows `W`
func TestShortOptsInvalidChar2(t *testing.T) {
	_, err := GetOpt(nil, "ab;")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// By testing `a:::` we expect the parser to split the tokens into
// `a::` and `:`, which _should_ generate an error as `-:` is not
// allowed to be an option, though `-=` is allowed.
func TestShortOptsInvalidChar3(t *testing.T) {
	_, err := GetOpt(nil, "a:::")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// The `;` is never allowed to appear in the opstring _unless_
// it follows `W`. This is a GNU extension to POSIX.
func TestShortOptsGnuWords(t *testing.T) {
	for _, opts := range GenShortOpts("ab") {
		optstring := opts + "W;" // Enable GNU word parsing
		getopt, err := GetOpt(nil, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if getopt.config.gnuWords != true {
			t.Errorf("Expected gnuWords to be false, got true")
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
	var tests = []struct {
		args []string
		s    string
		a    string
		t    bool
	}{
		{[]string{"-afoo"}, "a", "foo", true},
		{[]string{"-a", "bar"}, "a", "bar", true},
		{[]string{"-a", "-1"}, "a", "-1", true},
		{[]string{"-a"}, "a", "", false},
	}

	for _, test := range tests {
		getopt, err := GetOpt(test.args, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		for opt, err := range getopt.Options() {
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}
			if opt.Name != test.s {
				t.Errorf("Expected option %s, got %s", test.s, opt.Name)
			}
			if opt.HasArg != test.t {
				t.Errorf("Expected HasArg to be %t, got %t", test.t, opt.HasArg)
			}
			if opt.Arg != test.a {
				t.Errorf("Expected Arg to be %s, got %s", test.a, opt.Arg)
			}
		}
	}
}

func TestShortOptRequiredIntegration(t *testing.T) {
	// We disable automatic error reporting to avoid polluting the
	// test output
	optstring := ":a:"
	var tests = []struct {
		args []string
		s    string
		a    string
		t    bool
	}{
		{[]string{"-afoo"}, "a", "foo", true},
		{[]string{"-a", "bar"}, "a", "bar", true},
		{[]string{"-a", "-1"}, "a", "-1", true},
		{[]string{"-a"}, "a", "", false},
	}

	for _, test := range tests {
		getopt, err := GetOpt(test.args, optstring)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		for opt, err := range getopt.Options() {
			if err != nil {
				if test.a != "" {
					t.Errorf("Unexpected error: %s", err)
				}
			} else {
				if test.a == "" {
					t.Errorf("Expected an error")
				}
			}

			if opt.Name != test.s {
				t.Errorf("Expected option %s, got %s", test.s, opt.Name)
			}
			if opt.HasArg != test.t {
				t.Errorf("Expected HasArg to be %t, got %t", test.t, opt.HasArg)
			}
			if opt.Arg != test.a {
				t.Errorf("Expected Arg to be %s, got %s", test.a, opt.Arg)
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

package optargs

import (
	"strings"
	"testing"
	"testing/quick"
)

// graphChars returns every byte value for which isGraph reports true.
var graphChars = func() []byte {
	var out []byte
	for i := range 256 {
		if isGraph(byte(i)) {
			out = append(out, byte(i))
		}
	}
	return out
}()

// argTypeForIndex cycles through NoArgument, RequiredArgument,
// OptionalArgument based on the byte value. Used by init tests to
// exercise all three argument types without duplicating the switch.
func argTypeForIndex(b byte) ArgType {
	switch b % 3 {
	case 1:
		return RequiredArgument
	case 2:
		return OptionalArgument
	default:
		return NoArgument
	}
}

// newTestParser creates a parser with a single short option 'a' (NoArgument)
// and a single long option "a" (NoArgument). Reduces boilerplate in tests
// that only need a minimal option set.
func newTestParser(t *testing.T, config ParserConfig, args []string) *Parser {
	t.Helper()
	shopts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument},
	}
	lopts := map[string]*Flag{
		"a": {Name: "a", HasArg: NoArgument},
	}
	p, err := NewParser(config, shopts, lopts, args)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	return p
}

func TestParserInit(t *testing.T) {
	_, err := NewParser(ParserConfig{}, nil, nil, nil)
	if err != nil {
		t.Errorf("NewParser: %v", err)
	}
}

func TestParserInitShortOpts(t *testing.T) {
	shortOpts := make(map[byte]*Flag)
	for _, c := range graphChars {
		switch c {
		case ':', ';', '-':
			continue
		}
		shortOpts[c] = &Flag{Name: string(c), HasArg: argTypeForIndex(c)}
	}
	_, err := NewParser(ParserConfig{}, shortOpts, nil, nil)
	if err != nil {
		t.Errorf("NewParser: %v", err)
	}
}

var invalidShortOptTests = []struct {
	name string
	key  byte
}{
	{name: "colon", key: ':'},
	{name: "semicolon", key: ';'},
	{name: "dash", key: '-'},
}

func TestParserInitInvalidShortOpts(t *testing.T) {
	for _, tt := range invalidShortOptTests {
		t.Run(tt.name, func(t *testing.T) {
			shortOpts := map[byte]*Flag{
				tt.key: {Name: string(tt.key), HasArg: NoArgument},
			}
			_, err := NewParser(ParserConfig{}, shortOpts, nil, nil)
			if err == nil {
				t.Errorf("expected error for short option %q", tt.key)
			}
		})
	}
}

func TestParserInitNotIsGraphShortOpts(t *testing.T) {
	shortOpts := map[byte]*Flag{
		' ': {Name: " ", HasArg: NoArgument},
	}
	_, err := NewParser(ParserConfig{}, shortOpts, nil, nil)
	if err == nil {
		t.Error("expected error for non-graphic short option")
	}
}

func TestParserInitNotIsGraphLongOpts(t *testing.T) {
	longOpts := map[string]*Flag{
		" ": {Name: " ", HasArg: NoArgument},
	}
	_, err := NewParser(ParserConfig{}, nil, longOpts, nil)
	if err == nil {
		t.Error("expected error for non-graphic long option")
	}
}

func TestParserInitLongOpts(t *testing.T) {
	longOpts := make(map[string]*Flag)
	for _, c := range graphChars {
		s := string(rune(c))
		longOpts[s] = &Flag{Name: s, HasArg: argTypeForIndex(c)}
	}
	_, err := NewParser(ParserConfig{}, nil, longOpts, nil)
	if err != nil {
		t.Errorf("NewParser: %v", err)
	}
}

func TestParserPosixBreak(t *testing.T) {
	parser := newTestParser(t, ParserConfig{}, []string{"--", "-a", "--a"})

	for opt, err := range parser.Options() {
		if err != nil {
			t.Errorf("Options: %v", err)
		}
		t.Errorf("unexpected option %s", opt.Name)
	}

	if parser.Args[0] != "-a" {
		t.Errorf("Args[0] = %q, want %q", parser.Args[0], "-a")
	}
}

func TestParserNonOptShift(t *testing.T) {
	parser := newTestParser(t, ParserConfig{}, []string{"param", "-a", "--a"})

	for opt, err := range parser.Options() {
		if err != nil {
			t.Errorf("Options: %v", err)
		}
		if opt.Name != "a" {
			t.Errorf("option = %q, want %q", opt.Name, "a")
		}
	}

	if parser.Args[0] != "param" {
		t.Errorf("Args[0] = %q, want %q", parser.Args[0], "param")
	}
}

func TestParserNonOpt(t *testing.T) {
	parser := newTestParser(t, ParserConfig{parseMode: ParseNonOpts}, []string{"-a", "param", "--a"})

	for opt, err := range parser.Options() {
		if err != nil {
			t.Errorf("Options: %v", err)
		}
		switch opt.Name {
		case "a":
			// expected
		case string(byte(1)):
			if opt.Arg != "param" {
				t.Errorf("non-opt arg = %q, want %q", opt.Arg, "param")
			}
		default:
			t.Errorf("unexpected option %q", opt.Name)
		}
	}

	if len(parser.Args) != 0 {
		t.Errorf("len(Args) = %d, want 0", len(parser.Args))
	}
}

func TestParserPosixNonOpt(t *testing.T) {
	parser := newTestParser(t, ParserConfig{parseMode: ParsePosixlyCorrect}, []string{"param", "-a", "--a"})

	for opt, err := range parser.Options() {
		if err != nil {
			t.Errorf("Options: %v", err)
		}
		t.Errorf("unexpected option %s", opt.Name)
	}

	if parser.Args[0] != "param" {
		t.Errorf("Args[0] = %q, want %q", parser.Args[0], "param")
	}
	if parser.Args[1] != "-a" {
		t.Errorf("Args[1] = %q, want %q", parser.Args[1], "-a")
	}
}

func TestParserLongOptsLongPrefix(t *testing.T) {
	longOpts := map[string]*Flag{
		"foobar": {Name: "foobar", HasArg: RequiredArgument},
	}
	parser, err := NewParser(ParserConfig{}, nil, longOpts, []string{"--foo"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	for _, err := range parser.Options() {
		if err == nil {
			t.Error("expected error for ambiguous prefix --foo")
		}
	}
}

func TestParserLongOptsPrefix(t *testing.T) {
	longOpts := map[string]*Flag{
		"foo":  {Name: "foo", HasArg: RequiredArgument},
		"foo=": {Name: "foo=", HasArg: NoArgument},
	}
	parser, err := NewParser(ParserConfig{}, nil, longOpts, []string{"--foo="})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	for opt, err := range parser.Options() {
		if err != nil {
			t.Errorf("Options: %v", err)
		}
		if opt.Name != "foo=" {
			t.Errorf("option = %q, want %q", opt.Name, "foo=")
		}
	}
}

func TestParserLongOptsProvidedArgs(t *testing.T) {
	longOpts := map[string]*Flag{
		"foo": {Name: "foo", HasArg: RequiredArgument},
		"boo": {Name: "boo", HasArg: OptionalArgument},
	}
	tests := []struct {
		name string
		args []string
	}{
		{"required with equals", []string{"--foo=bar"}},
		{"required with space", []string{"--foo", "bar"}},
		{"optional with equals", []string{"--boo=baz"}},
		{"optional with space", []string{"--boo", "baz"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(ParserConfig{}, nil, longOpts, tt.args)
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}
			for opt, err := range parser.Options() {
				if err != nil {
					t.Errorf("Options: %v", err)
				}
				if !opt.HasArg {
					t.Error("HasArg = false, want true")
				}
			}
		})
	}
}

func TestParserLongOptsMissingOptArg(t *testing.T) {
	longOpts := map[string]*Flag{
		"foo": {Name: "foo", HasArg: OptionalArgument},
	}
	parser, err := NewParser(ParserConfig{}, nil, longOpts, []string{"--foo"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	count := 0
	for opt, err := range parser.Options() {
		count++
		if err != nil {
			t.Errorf("Options: %v", err)
		}
		if opt.HasArg {
			t.Error("HasArg = true, want false")
		}
	}
	if count != 1 {
		t.Errorf("option count = %d, want 1", count)
	}
}

// TestOptionalArgumentLookahead is a regression test for the bug where
// OptionalArgument long opts unconditionally consumed the next argument,
// even when it was another option or the -- terminator.
func TestOptionalArgumentLookahead(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantArg string // expected arg for --opt; empty means no arg
		wantHas bool   // expected HasArg for --opt
	}{
		{
			name:    "space delimited arg",
			args:    []string{"--opt", "a"},
			wantArg: "a",
			wantHas: true,
		},
		{
			name:    "equals delimited arg",
			args:    []string{"--opt=a"},
			wantArg: "a",
			wantHas: true,
		},
		{
			name:    "equals delimited arg starting with dash",
			args:    []string{"--opt=--"},
			wantArg: "--",
			wantHas: true,
		},
		{
			name:    "no arg followed by short option",
			args:    []string{"--opt", "-a"},
			wantArg: "",
			wantHas: false,
		},
		{
			name:    "no arg followed by long option",
			args:    []string{"--opt", "--arg"},
			wantArg: "",
			wantHas: false,
		},
		{
			name:    "no arg followed by double-hyphen terminator",
			args:    []string{"--opt", "--"},
			wantArg: "",
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			longOpts := map[string]*Flag{
				"opt": {Name: "opt", HasArg: OptionalArgument},
				"arg": {Name: "arg", HasArg: NoArgument},
			}
			shortOpts := map[byte]*Flag{
				'a': {Name: "a", HasArg: NoArgument},
			}
			parser, err := NewParser(ParserConfig{}, shortOpts, longOpts, tt.args)
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}

			var gotOpt Option
			found := false
			for opt, err := range parser.Options() {
				if err != nil {
					t.Fatalf("Options: %v", err)
				}
				if opt.Name == "opt" {
					gotOpt = opt
					found = true
				}
			}

			if !found {
				t.Fatal("--opt not found in parsed options")
			}
			if gotOpt.HasArg != tt.wantHas {
				t.Errorf("HasArg = %t, want %t", gotOpt.HasArg, tt.wantHas)
			}
			if gotOpt.Arg != tt.wantArg {
				t.Errorf("Arg = %q, want %q", gotOpt.Arg, tt.wantArg)
			}
		})
	}
}

// --- Tests merged from parser_coverage_test.go ---

// TestFindShortOptUnknownArgType verifies the error path when a Flag has
// an invalid HasArg value.
func TestFindShortOptUnknownArgType(t *testing.T) {
	shortOpts := map[byte]*Flag{
		'x': {HasArg: 999},
	}
	parser, err := NewParser(ParserConfig{}, shortOpts, nil, nil)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	_, _, _, _, err = parser.findShortOpt('x', "", nil)
	if err == nil {
		t.Fatal("expected error for unknown argument type")
	}
	if err.Error() != "unknown argument type: 999" {
		t.Errorf("error = %q, want %q", err.Error(), "unknown argument type: 999")
	}
}

var caseInsensitiveShortOptTests = []struct {
	name     string
	char     byte
	word     string
	args     []string
	wantName string
	wantArg  string
}{
	{name: "no_argument", char: 'V', wantName: "v"},
	{name: "required_argument_from_word", char: 'F', word: "value", wantName: "f", wantArg: "value"},
	{name: "optional_argument_from_args", char: 'O', args: []string{"value"}, wantName: "o", wantArg: "value"},
}

// TestFindShortOptCaseInsensitive verifies case-insensitive short option
// matching across all three argument types.
func TestFindShortOptCaseInsensitive(t *testing.T) {
	shortOpts := map[byte]*Flag{
		'v': {HasArg: NoArgument},
		'f': {HasArg: RequiredArgument},
		'o': {HasArg: OptionalArgument},
	}
	parser, err := NewParser(ParserConfig{shortCaseIgnore: true}, shortOpts, nil, nil)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	for _, tt := range caseInsensitiveShortOptTests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, option, err := parser.findShortOpt(tt.char, tt.word, tt.args)
			if err != nil {
				t.Fatalf("findShortOpt: %v", err)
			}
			if option.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", option.Name, tt.wantName)
			}
			if option.Arg != tt.wantArg {
				t.Errorf("Arg = %q, want %q", option.Arg, tt.wantArg)
			}
		})
	}
}

// TestOptionsGNUWordsTransformation verifies that -Wfoo is transformed
// into option name "foo" when gnuWords mode is enabled.
func TestOptionsGNUWordsTransformation(t *testing.T) {
	shortOpts := map[byte]*Flag{
		'W': {HasArg: RequiredArgument},
	}
	parser, err := NewParser(ParserConfig{gnuWords: true}, shortOpts, nil, []string{"-Wfoo"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	for option, err := range parser.Options() {
		if err != nil {
			t.Fatalf("Options: %v", err)
		}
		if option.Name != "foo" {
			t.Errorf("Name = %q, want %q (GNU words transformation)", option.Name, "foo")
		}
		break
	}
}

// TestOptionsParseNonOptsMode verifies that non-option arguments are
// yielded as synthetic options with character code 1.
func TestOptionsParseNonOptsMode(t *testing.T) {
	parser, err := NewParser(ParserConfig{parseMode: ParseNonOpts}, nil, nil, []string{"non-option"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	for option, err := range parser.Options() {
		if err != nil {
			t.Fatalf("Options: %v", err)
		}
		if option.Name != string(byte(1)) {
			t.Errorf("Name = %q, want %q", option.Name, string(byte(1)))
		}
		if option.Arg != "non-option" {
			t.Errorf("Arg = %q, want %q", option.Arg, "non-option")
		}
		break
	}
}

// TestOptionsPosixlyCorrectMode verifies that parsing stops at the first
// non-option argument and remaining args are preserved.
func TestOptionsPosixlyCorrectMode(t *testing.T) {
	parser, err := NewParser(ParserConfig{parseMode: ParsePosixlyCorrect}, nil, nil, []string{"non-option", "-v"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	count := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Fatalf("Options: %v", err)
		}
		count++
	}

	if count != 0 {
		t.Errorf("option count = %d, want 0", count)
	}
	if len(parser.Args) != 2 {
		t.Errorf("len(Args) = %d, want 2", len(parser.Args))
	}
}

// TestOptionsLongOptsOnlyMode verifies that single-dash arguments are
// matched as long options when longOptsOnly is enabled.
func TestOptionsLongOptsOnlyMode(t *testing.T) {
	longOpts := map[string]*Flag{
		"verbose": {HasArg: NoArgument},
	}
	parser, err := NewParser(ParserConfig{longOptsOnly: true}, nil, longOpts, []string{"-verbose"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	for option, err := range parser.Options() {
		if err != nil {
			t.Fatalf("Options: %v", err)
		}
		if option.Name != "verbose" {
			t.Errorf("Name = %q, want %q", option.Name, "verbose")
		}
		break
	}
}

// TestOptionsCommandExecution verifies that subcommand dispatch works
// through the Options iterator.
func TestOptionsCommandExecution(t *testing.T) {
	parser, _ := NewParser(ParserConfig{}, nil, nil, []string{"subcmd", "arg1"})
	subParser, _ := NewParser(ParserConfig{}, nil, nil, nil)
	parser.AddCmd("subcmd", subParser)

	count := 0
	for _, err := range parser.Options() {
		if err != nil {
			t.Fatalf("Options: %v", err)
		}
		count++
	}
	if count != 0 {
		t.Errorf("option count = %d, want 0", count)
	}
}

// TestOptionsCommandExecutionError verifies that a nil subcommand parser
// produces an error through the Options iterator.
func TestOptionsCommandExecutionError(t *testing.T) {
	parser, err := NewParser(ParserConfig{}, nil, nil, []string{"subcmd"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	parser.AddCmd("subcmd", nil)

	for _, err := range parser.Options() {
		if err == nil {
			t.Fatal("expected error from nil command parser")
		}
		break
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

// --- Tests moved from compliance_test.go (unique long-option and prefix-overlap tests) ---

func TestGNULongOptionComplexNames(t *testing.T) {
	longOpts := []Flag{
		{Name: "foo=bar", HasArg: NoArgument},
		{Name: "with-dashes", HasArg: RequiredArgument},
		{Name: "under_scores", HasArg: OptionalArgument},
		{Name: "123numbers", HasArg: NoArgument},
		{Name: "a", HasArg: NoArgument},
	}
	tests := []struct {
		name string
		args []string
		want []Option
	}{
		{"equals in name", []string{"--foo=bar"}, []Option{{Name: "foo=bar"}}},
		{"dashes", []string{"--with-dashes=value"}, []Option{{Name: "with-dashes", Arg: "value", HasArg: true}}},
		{"underscores", []string{"--under_scores=value"}, []Option{{Name: "under_scores", Arg: "value", HasArg: true}}},
		{"numbers", []string{"--123numbers"}, []Option{{Name: "123numbers"}}},
		{"single char long", []string{"--a"}, []Option{{Name: "a"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong(tt.args, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, p), tt.want)
		})
	}
}

func TestOverlappingOptionNames(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		longOpts []Flag
		expected []Option
	}{
		{
			name:     "exact_match_wins_over_prefix",
			args:     []string{"--foo=val"},
			longOpts: []Flag{{Name: "foo", HasArg: RequiredArgument}, {Name: "foobar", HasArg: RequiredArgument}},
			expected: []Option{{Name: "foo", Arg: "val", HasArg: true}},
		},
		{
			name:     "longer_prefix_wins_at_equals_boundary",
			args:     []string{"--foobar=val"},
			longOpts: []Flag{{Name: "foo", HasArg: RequiredArgument}, {Name: "foobar", HasArg: RequiredArgument}},
			expected: []Option{{Name: "foobar", Arg: "val", HasArg: true}},
		},
		{
			name:     "three_level_prefix_chain",
			args:     []string{"--output=file.txt"},
			longOpts: []Flag{
				{Name: "o", HasArg: RequiredArgument},
				{Name: "out", HasArg: RequiredArgument},
				{Name: "output", HasArg: RequiredArgument},
			},
			expected: []Option{{Name: "output", Arg: "file.txt", HasArg: true}},
		},
		{
			name:     "three_level_mid_match",
			args:     []string{"--out=file.txt"},
			longOpts: []Flag{
				{Name: "o", HasArg: RequiredArgument},
				{Name: "out", HasArg: RequiredArgument},
				{Name: "output", HasArg: RequiredArgument},
			},
			expected: []Option{{Name: "out", Arg: "file.txt", HasArg: true}},
		},
		{
			name:     "noarg_longest_skips_to_shorter_with_arg",
			args:     []string{"--foo=bar"},
			longOpts: []Flag{{Name: "foo", HasArg: RequiredArgument}, {Name: "foo=bar", HasArg: NoArgument}},
			expected: []Option{{Name: "foo=bar", HasArg: false}},
		},
		{
			name:     "equals_in_name_with_arg",
			args:     []string{"--foo=bar=baz"},
			longOpts: []Flag{{Name: "foo", HasArg: RequiredArgument}, {Name: "foo=bar", HasArg: RequiredArgument}},
			expected: []Option{{Name: "foo=bar", Arg: "baz", HasArg: true}},
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

func TestMultipleOptionsWithOverlappingPrefixes(t *testing.T) {
	longOpts := []Flag{
		{Name: "v", HasArg: RequiredArgument},
		{Name: "ve", HasArg: RequiredArgument},
		{Name: "ver", HasArg: RequiredArgument},
		{Name: "verbose", HasArg: RequiredArgument},
	}
	tests := []struct {
		input string
		want  []Option
	}{
		{"--v=1", []Option{{Name: "v", Arg: "1", HasArg: true}}},
		{"--ve=2", []Option{{Name: "ve", Arg: "2", HasArg: true}}},
		{"--ver=3", []Option{{Name: "ver", Arg: "3", HasArg: true}}},
		{"--verbose=7", []Option{{Name: "verbose", Arg: "7", HasArg: true}}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p, err := GetOptLong([]string{tt.input}, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, p), tt.want)
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
// TestTripleEqualsOverlap exercises three options with '=' embedded in names.
func TestTripleEqualsOverlap(t *testing.T) {
	longOpts := []Flag{
		{Name: "foo", HasArg: OptionalArgument},
		{Name: "foo=bar", HasArg: OptionalArgument},
		{Name: "foo=bar=arg", HasArg: NoArgument},
	}
	tests := []struct {
		name  string
		input string
		want  []Option
	}{
		{"exact foo=bar=arg matches NoArgument", "--foo=bar=arg", []Option{{Name: "foo=bar=arg"}}},
		{"foo=bar with equals arg", "--foo=bar=something", []Option{{Name: "foo=bar", Arg: "something", HasArg: true}}},
		{"foo with equals arg", "--foo=qux", []Option{{Name: "foo", Arg: "qux", HasArg: true}}},
		{"foo=bar exact no trailing equals", "--foo=bar", []Option{{Name: "foo=bar"}}},
		{"foo exact no trailing equals", "--foo", []Option{{Name: "foo"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong([]string{tt.input}, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			assertOptions(t, requireParsedOptions(t, p), tt.want)
		})
	}
}

// TestObscureLongOptCharacters exercises long option names containing
// characters that are valid isgraph() but unusual: brackets, braces,
// dots, colons, tildes, etc. Per POSIX/GNU convention, any isgraph()
// character is valid in a long option name.
// TestObscureLongOptCharacters exercises long option names containing
// characters that are valid isgraph() but unusual: brackets, braces,
// dots, colons, tildes, etc.
func TestObscureLongOptCharacters(t *testing.T) {
	tests := []struct {
		name     string
		optName  string
		hasArg   ArgType
		input    []string
		expected []Option
	}{
		{
			name:     "brackets equals arg",
			optName:  "config[key]",
			hasArg:   RequiredArgument,
			input:    []string{"--config[key]=val"},
			expected: []Option{{Name: "config[key]", Arg: "val", HasArg: true}},
		},
		{
			name:     "braces space arg",
			optName:  "data{category.key}",
			hasArg:   RequiredArgument,
			input:    []string{"--data{category.key}", "val"},
			expected: []Option{{Name: "data{category.key}", Arg: "val", HasArg: true}},
		},
		{
			name:     "colon equals arg",
			optName:  "command:arg",
			hasArg:   RequiredArgument,
			input:    []string{"--command:arg=value"},
			expected: []Option{{Name: "command:arg", Arg: "value", HasArg: true}},
		},
		{
			name:     "dot space arg",
			optName:  "section.key",
			hasArg:   RequiredArgument,
			input:    []string{"--section.key", "value"},
			expected: []Option{{Name: "section.key", Arg: "value", HasArg: true}},
		},
		{
			name:     "tilde equals arg",
			optName:  "path~backup",
			hasArg:   RequiredArgument,
			input:    []string{"--path~backup=/tmp"},
			expected: []Option{{Name: "path~backup", Arg: "/tmp", HasArg: true}},
		},
		{
			name:     "plus space arg",
			optName:  "level+1",
			hasArg:   RequiredArgument,
			input:    []string{"--level+1", "high"},
			expected: []Option{{Name: "level+1", Arg: "high", HasArg: true}},
		},
		{
			name:     "at equals arg",
			optName:  "user@host",
			hasArg:   RequiredArgument,
			input:    []string{"--user@host=root"},
			expected: []Option{{Name: "user@host", Arg: "root", HasArg: true}},
		},
		{
			name:     "brackets no arg",
			optName:  "flag[x]",
			hasArg:   NoArgument,
			input:    []string{"--flag[x]"},
			expected: []Option{{Name: "flag[x]", HasArg: false}},
		},
		{
			name:     "braces optional with equals",
			optName:  "opt{a.b}",
			hasArg:   OptionalArgument,
			input:    []string{"--opt{a.b}=yes"},
			expected: []Option{{Name: "opt{a.b}", Arg: "yes", HasArg: true}},
		},
		{
			name:     "braces optional without arg",
			optName:  "opt{a.b}",
			hasArg:   OptionalArgument,
			input:    []string{"--opt{a.b}"},
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
// TestObscureCharOverlappingPrefixes tests longest-prefix matching with obscure characters.
func TestObscureCharOverlappingPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		longOpts []Flag
		args     []string
		expected []Option
	}{
		{
			name: "bracket_prefix_overlap",
			longOpts: []Flag{
				{Name: "config", HasArg: RequiredArgument},
				{Name: "config[key]", HasArg: RequiredArgument},
			},
			args:     []string{"--config[key]=val"},
			expected: []Option{{Name: "config[key]", Arg: "val", HasArg: true}},
		},
		{
			name: "bracket_falls_back_to_shorter",
			longOpts: []Flag{
				{Name: "config", HasArg: RequiredArgument},
				{Name: "config[key]", HasArg: RequiredArgument},
			},
			args:     []string{"--config=val"},
			expected: []Option{{Name: "config", Arg: "val", HasArg: true}},
		},
		{
			name: "colon_prefix_overlap",
			longOpts: []Flag{
				{Name: "cmd", HasArg: RequiredArgument},
				{Name: "cmd:sub", HasArg: RequiredArgument},
			},
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
		child := setupChain3(t,
			[]Flag{{Name: "data{cat.key}", HasArg: RequiredArgument}},
			[]Flag{{Name: "data{cat}", HasArg: RequiredArgument}},
			[]Flag{{Name: "data", HasArg: RequiredArgument}},
			[]string{"--data{cat.key}=val"},
		)
		assertOptions(t, requireParsedOptions(t, child), []Option{{Name: "data{cat.key}", Arg: "val", HasArg: true}})
	})
}

// --- Unique rows merged from TestEdgeCaseLongOptionHandling ---

// TestEdgeCaseLongOptionErrors tests error cases for long option handling
// that are not covered by other long option tests: unknown options with
// multiple equals signs, empty option names, and equals-only options.
func TestEdgeCaseLongOptionErrors(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
		{Name: "foo=bar", HasArg: NoArgument},
	}

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
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

// TestMultiLevelInheritanceViaIterator tests option inheritance through
// multiple levels using the Options() iterator.
// TestMultiLevelInheritanceViaIterator tests option inheritance through
// multiple levels using the Options() iterator.
func TestMultiLevelInheritanceViaIterator(t *testing.T) {
	t.Run("short_and_long_opts_4_levels", func(t *testing.T) {
		root, _ := NewParser(ParserConfig{},
			map[byte]*Flag{'r': {Name: "r", HasArg: NoArgument}},
			map[string]*Flag{"root": {Name: "root", HasArg: NoArgument}},
			[]string{})
		l1, _ := NewParser(ParserConfig{},
			map[byte]*Flag{'a': {Name: "a", HasArg: NoArgument}},
			map[string]*Flag{"level1": {Name: "level1", HasArg: NoArgument}},
			[]string{})
		l2, _ := NewParser(ParserConfig{},
			map[byte]*Flag{'b': {Name: "b", HasArg: NoArgument}},
			map[string]*Flag{"level2": {Name: "level2", HasArg: NoArgument}},
			[]string{})
		l3, _ := NewParser(ParserConfig{},
			map[byte]*Flag{'c': {Name: "c", HasArg: NoArgument}},
			map[string]*Flag{"level3": {Name: "level3", HasArg: NoArgument}},
			[]string{"-r", "-a", "-b", "-c"})
		root.AddCmd("l1", l1)
		l1.AddCmd("l2", l2)
		l2.AddCmd("l3", l3)

		found := make(map[string]bool)
		for opt, err := range l3.Options() {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			found[opt.Name] = true
		}
		for _, want := range []string{"r", "a", "b", "c"} {
			if !found[want] {
				t.Errorf("missing option %q", want)
			}
		}
	})

	t.Run("inherited_options_with_arguments", func(t *testing.T) {
		root, _ := NewParser(ParserConfig{},
			map[byte]*Flag{'v': {Name: "v", HasArg: RequiredArgument}},
			nil, []string{})
		l1, _ := NewParser(ParserConfig{},
			map[byte]*Flag{'o': {Name: "o", HasArg: OptionalArgument}},
			nil, []string{})
		l2, _ := NewParser(ParserConfig{},
			map[byte]*Flag{'f': {Name: "f", HasArg: RequiredArgument}},
			nil, []string{"-v", "verbose", "-o", "optional", "-f", "file"})
		root.AddCmd("l1", l1)
		l1.AddCmd("l2", l2)

		expected := map[string]string{"v": "verbose", "o": "optional", "f": "file"}
		found := make(map[string]string)
		for opt, err := range l2.Options() {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			found[opt.Name] = opt.Arg
		}
		for name, arg := range expected {
			if found[name] != arg {
				t.Errorf("option %q: got %q, want %q", name, found[name], arg)
			}
		}
	})

	t.Run("inherited_long_options", func(t *testing.T) {
		root, _ := NewParser(ParserConfig{}, nil,
			map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}},
			[]string{})
		l1, _ := NewParser(ParserConfig{}, nil,
			map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument}},
			[]string{})
		l2, _ := NewParser(ParserConfig{}, nil,
			map[string]*Flag{"file": {Name: "file", HasArg: RequiredArgument}},
			[]string{"--verbose", "--output", "out.txt", "--file", "input.txt"})
		root.AddCmd("l1", l1)
		l1.AddCmd("l2", l2)

		expected := map[string]string{"verbose": "", "output": "out.txt", "file": "input.txt"}
		found := make(map[string]string)
		for opt, err := range l2.Options() {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			found[opt.Name] = opt.Arg
		}
		for name, arg := range expected {
			if found[name] != arg {
				t.Errorf("option %q: got %q, want %q", name, found[name], arg)
			}
		}
	})
}

// TestFindShortOptEdgeCases tests remaining edge cases in findShortOpt
// via inheritance.
func TestFindShortOptEdgeCases(t *testing.T) {
	t.Run("Unknown_argument_type", func(t *testing.T) {
		parent, child := childOf(t, "f", "")

		// Corrupt the parent's flag to have an invalid HasArg value.
		parent.shortOpts['f'] = &Flag{Name: "f", HasArg: ArgType(999)}

		_, _, _, _, err := child.findShortOpt('f', "", []string{})
		if err == nil {
			t.Fatal("expected error for unknown argument type")
		}
		if !strings.Contains(err.Error(), "unknown argument type") {
			t.Errorf("error = %q, want containing %q", err.Error(), "unknown argument type")
		}
	})
}

// TestFindShortOptDirectErrors tests findShortOpt error paths on a
// single parser (no inheritance chain).
func TestFindShortOptDirectErrors(t *testing.T) {
	tests := []struct {
		name    string
		char    byte
		wantErr string
	}{
		{"invalid_option_dash", '-', "invalid option: -"},
		{"unknown_option", 'z', "unknown option: z"},
	}

	parser, err := GetOpt([]string{}, "abc")
	if err != nil {
		t.Fatalf("parser: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, _, err := parser.findShortOpt(tt.char, "", []string{})
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// Feature: goarg-optargs-integration, Property 7: Flag and Parser metadata round-trip
// Validates: Requirements 8.1, 8.2, 8.3, 8.6, 8.7, 8.8, 8.9
//
// For any set of metadata values, setting them at registration time and
// reading them back by walking the parser tree returns the same values.

func TestPropertyMetadataRoundTrip(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	t.Run("flag_metadata", func(t *testing.T) {
		f := func(help, argName, defaultVal string) bool {
			flag := &Flag{
				Name:         "test",
				HasArg:       RequiredArgument,
				Help:         help,
				ArgName:      argName,
				DefaultValue: defaultVal,
			}
			p, err := NewParser(ParserConfig{}, nil, map[string]*Flag{"test": flag}, nil)
			if err != nil {
				return false
			}
			got := p.longOpts["test"]
			return got.Help == help && got.ArgName == argName && got.DefaultValue == defaultVal
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("parser_metadata_via_addcmd", func(t *testing.T) {
		f := func(name, desc string) bool {
			// Filter out names with whitespace or non-graphic chars
			// since those would fail command lookup.
			for _, r := range name {
				if r <= ' ' || r > '~' {
					return true // skip, not a valid command name
				}
			}
			if name == "" {
				return true
			}
			parent, err := NewParser(ParserConfig{}, nil, nil, nil)
			if err != nil {
				return false
			}
			child, err := NewParser(ParserConfig{}, nil, nil, nil)
			if err != nil {
				return false
			}
			child.Description = desc
			parent.AddCmd(name, child)

			got, exists := parent.GetCommand(name)
			if !exists {
				return false
			}
			return got.Name == name && got.Description == desc
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// Feature: goarg-optargs-integration, Property 8: Peer link bidirectional invariant
// Validates: Requirements 8.4, 8.5
//
// If shortFlag.Peer == longFlag, then longFlag.Peer == shortFlag —
// one dereference each way.

func TestPropertyPeerLinkBidirectional(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	f := func(help string) bool {
		shortFlag := &Flag{Name: "v", HasArg: NoArgument, Help: help}
		longFlag := &Flag{Name: "verbose", HasArg: NoArgument, Help: help}
		shortFlag.Peer = longFlag
		longFlag.Peer = shortFlag

		p, err := NewParser(ParserConfig{},
			map[byte]*Flag{'v': shortFlag},
			map[string]*Flag{"verbose": longFlag},
			nil,
		)
		if err != nil {
			return false
		}

		s := p.shortOpts['v']
		l := p.longOpts["verbose"]
		return s.Peer == l && l.Peer == s
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

// ---------------------------------------------------------------------------
// Unit tests — Flag and Parser metadata fields
// Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7, 8.8, 8.9
// ---------------------------------------------------------------------------

func TestFlagMetadataFields(t *testing.T) {
	tests := []struct {
		name         string
		help         string
		argName      string
		defaultValue string
	}{
		{"all fields set", "enable verbose output", "LEVEL", "0"},
		{"empty help", "", "FILE", "/dev/stdin"},
		{"empty argname", "show version", "", ""},
		{"all empty", "", "", ""},
		{"unicode help", "启用详细输出", "文件", "默认"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := &Flag{
				Name:         "test",
				HasArg:       RequiredArgument,
				Help:         tt.help,
				ArgName:      tt.argName,
				DefaultValue: tt.defaultValue,
			}
			p, err := NewParser(ParserConfig{}, nil, map[string]*Flag{"test": flag}, nil)
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}
			got := p.longOpts["test"]
			if got.Help != tt.help {
				t.Errorf("Help = %q, want %q", got.Help, tt.help)
			}
			if got.ArgName != tt.argName {
				t.Errorf("ArgName = %q, want %q", got.ArgName, tt.argName)
			}
			if got.DefaultValue != tt.defaultValue {
				t.Errorf("DefaultValue = %q, want %q", got.DefaultValue, tt.defaultValue)
			}
		})
	}
}

func TestParserMetadataViaAddCmd(t *testing.T) {
	tests := []struct {
		name        string
		cmdName     string
		description string
	}{
		{"basic subcommand", "serve", "start the server"},
		{"empty description", "build", ""},
		{"unicode", "构建", "构建项目"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent, err := NewParser(ParserConfig{}, nil, nil, nil)
			if err != nil {
				t.Fatalf("NewParser parent: %v", err)
			}
			child, err := NewParser(ParserConfig{}, nil, nil, nil)
			if err != nil {
				t.Fatalf("NewParser child: %v", err)
			}
			child.Description = tt.description
			parent.AddCmd(tt.cmdName, child)

			got, exists := parent.GetCommand(tt.cmdName)
			if !exists {
				t.Fatalf("command %q not found", tt.cmdName)
			}
			if got.Name != tt.cmdName {
				t.Errorf("Name = %q, want %q", got.Name, tt.cmdName)
			}
			if got.Description != tt.description {
				t.Errorf("Description = %q, want %q", got.Description, tt.description)
			}
		})
	}
}

func TestPeerLinkSetAndVerified(t *testing.T) {
	shortFlag := &Flag{Name: "v", HasArg: NoArgument, Help: "verbose"}
	longFlag := &Flag{Name: "verbose", HasArg: NoArgument, Help: "verbose"}
	shortFlag.Peer = longFlag
	longFlag.Peer = shortFlag

	p, err := NewParser(ParserConfig{},
		map[byte]*Flag{'v': shortFlag},
		map[string]*Flag{"verbose": longFlag},
		nil,
	)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	s := p.shortOpts['v']
	l := p.longOpts["verbose"]

	if s.Peer != l {
		t.Error("short.Peer does not point to long flag")
	}
	if l.Peer != s {
		t.Error("long.Peer does not point to short flag")
	}
	if s.Peer.Name != "verbose" {
		t.Errorf("short.Peer.Name = %q, want %q", s.Peer.Name, "verbose")
	}
	if l.Peer.Name != "v" {
		t.Errorf("long.Peer.Name = %q, want %q", l.Peer.Name, "v")
	}
}

func TestPeerNilForFlagsWithoutCounterpart(t *testing.T) {
	shortOnly := &Flag{Name: "v", HasArg: NoArgument}
	longOnly := &Flag{Name: "output", HasArg: RequiredArgument}

	p, err := NewParser(ParserConfig{},
		map[byte]*Flag{'v': shortOnly},
		map[string]*Flag{"output": longOnly},
		nil,
	)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	if p.shortOpts['v'].Peer != nil {
		t.Error("short-only flag should have nil Peer")
	}
	if p.longOpts["output"].Peer != nil {
		t.Error("long-only flag should have nil Peer")
	}
}

func TestAddCmdNilParser(t *testing.T) {
	parent, err := NewParser(ParserConfig{}, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	// AddCmd with nil parser should not panic
	parent.AddCmd("nilcmd", nil)

	_, exists := parent.GetCommand("nilcmd")
	if !exists {
		t.Error("nil command should still be registered")
	}
}

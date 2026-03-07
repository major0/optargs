package optargs

import (
	"strings"
	"testing"
)

// graphChars returns every byte value for which isGraph reports true.
var graphChars = func() []byte {
	var out []byte
	for i := 0; i <= 255; i++ {
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

var longOptPrefixTests = []struct {
	name    string
	arg     string
	wantOpt string
	wantArg bool
}{
	{name: "exact match with equals", arg: "--foo=", wantOpt: "foo=", wantArg: false},
}

func TestParserLongOptsPrefix(t *testing.T) {
	longOpts := map[string]*Flag{
		"foo":  {Name: "foo", HasArg: RequiredArgument},
		"foo=": {Name: "foo=", HasArg: NoArgument},
	}

	for _, tt := range longOptPrefixTests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(ParserConfig{}, nil, longOpts, []string{tt.arg})
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}
			count := 0
			for opt, err := range parser.Options() {
				count++
				if err != nil {
					t.Errorf("Options: %v", err)
				}
				if opt.Name != tt.wantOpt {
					t.Errorf("option = %q, want %q", opt.Name, tt.wantOpt)
				}
				if opt.HasArg != tt.wantArg {
					t.Errorf("HasArg = %v, want %v", opt.HasArg, tt.wantArg)
				}
			}
			if count != 1 {
				t.Errorf("option count = %d, want 1", count)
			}
		})
	}
}

var longOptArgTests = []struct {
	name    string
	args    []string
	wantArg bool
}{
	{name: "required with equals", args: []string{"--foo=bar"}, wantArg: true},
	{name: "required with space", args: []string{"--foo", "bar"}, wantArg: true},
	{name: "optional with equals", args: []string{"--boo=baz"}, wantArg: true},
	{name: "optional with space", args: []string{"--boo", "baz"}, wantArg: true},
}

func TestParserLongOptsProvidedArgs(t *testing.T) {
	longOpts := map[string]*Flag{
		"foo": {Name: "foo", HasArg: RequiredArgument},
		"boo": {Name: "boo", HasArg: OptionalArgument},
	}

	for _, tt := range longOptArgTests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(ParserConfig{}, nil, longOpts, tt.args)
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}
			for opt, err := range parser.Options() {
				if err != nil {
					t.Errorf("Options: %v", err)
				}
				if opt.HasArg != tt.wantArg {
					t.Errorf("HasArg = %v, want %v", opt.HasArg, tt.wantArg)
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
	parser, err := NewParser(ParserConfig{}, nil, nil, []string{"subcmd", "arg1"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	subParser, err := NewParser(ParserConfig{}, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewParser (sub): %v", err)
	}
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
	if len(parser.Args) != 0 {
		t.Errorf("len(Args) = %d, want 0", len(parser.Args))
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
		{
			name: "single character long option",
			args: []string{"--a"},
			expected: []Option{
				{Name: "a", HasArg: false, Arg: ""},
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

// --- Unique rows merged from TestEdgeCaseLongOptionHandling ---

// TestEdgeCaseLongOptionErrors tests error cases for long option handling
// that are not covered by other long option tests: unknown options with
// multiple equals signs, empty option names, and equals-only options.
func TestEdgeCaseLongOptionErrors(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
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
func TestMultiLevelInheritanceViaIterator(t *testing.T) {
	t.Run("short_and_long_opts_4_levels", func(t *testing.T) {
		rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'r': {Name: "r", HasArg: NoArgument},
		}, map[string]*Flag{
			"root": {Name: "root", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("root parser: %v", err)
		}

		level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'a': {Name: "a", HasArg: NoArgument},
		}, map[string]*Flag{
			"level1": {Name: "level1", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("level1 parser: %v", err)
		}
		rootParser.AddCmd("level1", level1Parser)

		level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'b': {Name: "b", HasArg: NoArgument},
		}, map[string]*Flag{
			"level2": {Name: "level2", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("level2 parser: %v", err)
		}
		level1Parser.AddCmd("level2", level2Parser)

		level3Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'c': {Name: "c", HasArg: NoArgument},
		}, map[string]*Flag{
			"level3": {Name: "level3", HasArg: NoArgument},
		}, []string{"-r", "-a", "-b", "-c"})
		if err != nil {
			t.Fatalf("level3 parser: %v", err)
		}
		level2Parser.AddCmd("level3", level3Parser)

		foundOptions := make(map[string]bool)
		for option, err := range level3Parser.Options() {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			foundOptions[option.Name] = true
		}

		for _, expected := range []string{"r", "a", "b", "c"} {
			if !foundOptions[expected] {
				t.Errorf("missing option %q", expected)
			}
		}
	})

	t.Run("inherited_options_with_arguments", func(t *testing.T) {
		rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'v': {Name: "v", HasArg: RequiredArgument},
		}, map[string]*Flag{}, []string{})
		if err != nil {
			t.Fatalf("root parser: %v", err)
		}

		level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'o': {Name: "o", HasArg: OptionalArgument},
		}, map[string]*Flag{}, []string{})
		if err != nil {
			t.Fatalf("level1 parser: %v", err)
		}
		rootParser.AddCmd("level1", level1Parser)

		level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{
			'f': {Name: "f", HasArg: RequiredArgument},
		}, map[string]*Flag{}, []string{"-v", "verbose", "-o", "optional", "-f", "file"})
		if err != nil {
			t.Fatalf("level2 parser: %v", err)
		}
		level1Parser.AddCmd("level2", level2Parser)

		expected := map[string]string{
			"v": "verbose",
			"o": "optional",
			"f": "file",
		}

		found := make(map[string]string)
		for option, err := range level2Parser.Options() {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			found[option.Name] = option.Arg
		}

		for name, arg := range expected {
			if foundArg, exists := found[name]; !exists {
				t.Errorf("missing option %q", name)
			} else if foundArg != arg {
				t.Errorf("option %q: Arg = %q, want %q", name, foundArg, arg)
			}
		}
	})

	t.Run("inherited_long_options", func(t *testing.T) {
		rootParser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
			"verbose": {Name: "verbose", HasArg: NoArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("root parser: %v", err)
		}

		level1Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
			"output": {Name: "output", HasArg: RequiredArgument},
		}, []string{})
		if err != nil {
			t.Fatalf("level1 parser: %v", err)
		}
		rootParser.AddCmd("level1", level1Parser)

		level2Parser, err := NewParser(ParserConfig{}, map[byte]*Flag{}, map[string]*Flag{
			"file": {Name: "file", HasArg: RequiredArgument},
		}, []string{"--verbose", "--output", "out.txt", "--file", "input.txt"})
		if err != nil {
			t.Fatalf("level2 parser: %v", err)
		}
		level1Parser.AddCmd("level2", level2Parser)

		expected := map[string]string{
			"verbose": "",
			"output":  "out.txt",
			"file":    "input.txt",
		}

		found := make(map[string]string)
		for option, err := range level2Parser.Options() {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				continue
			}
			found[option.Name] = option.Arg
		}

		for name, arg := range expected {
			if foundArg, exists := found[name]; !exists {
				t.Errorf("missing option %q", name)
			} else if foundArg != arg {
				t.Errorf("option %q: Arg = %q, want %q", name, foundArg, arg)
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

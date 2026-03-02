package optargs

import (
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

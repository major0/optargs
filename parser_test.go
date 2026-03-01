package optargs

import (
	"testing"
)

// graphChars returns every byte value for which isGraph reports true.
var graphChars = func() []byte {
	var out []byte
	for i := 0; i < 255; i++ {
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

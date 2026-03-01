package optargs

import (
	"testing"
)

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

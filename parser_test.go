package optargs

import (
	"testing"
)

func genOpts() []byte {
	var opts []byte
	for i := 0; i < 255; i++ {
		if !isGraph(byte(i)) {
			continue
		}
	}
	return opts
}

func TestParserInit(t *testing.T) {
	_, err := NewParser(ParserConfig{}, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestParserInitShortOpts(t *testing.T) {
	var shortOpts = make(map[byte]*Flag)
	for c, i := range genOpts() {
		switch byte(i) {
		case ':', ';', '-':
			continue
		}

		var hasArg ArgType
		switch i % 3 {
		case 0:
			hasArg = NoArgument
		case 1:
			hasArg = RequiredArgument
		case 2:
			hasArg = OptionalArgument
		}

		shortOpts[byte(c)] = &Flag{Name: string(byte(c)), HasArg: hasArg}
	}
	_, err := NewParser(ParserConfig{}, shortOpts, nil, nil, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}
func TestParserInitInvalidShortOpts(t *testing.T) {
	var shortOpts = map[byte]*Flag{
		':': &Flag{Name: ":", HasArg: NoArgument},
		';': &Flag{Name: ";", HasArg: NoArgument},
		'-': &Flag{Name: "-", HasArg: NoArgument},
	}

	_, err := NewParser(ParserConfig{}, shortOpts, nil, nil, nil)
	if err == nil {
		t.Errorf("Expected invalid option error")
	}
}

func TestParserInitNotIsGraphShortOpts(t *testing.T) {
	var shortOpts = map[byte]*Flag{
		' ': &Flag{Name: " ", HasArg: NoArgument},
	}

	_, err := NewParser(ParserConfig{}, shortOpts, nil, nil, nil)
	if err == nil {
		t.Errorf("Expected invalid option error")
	}
}

func TestParserInitNotIsGraphLongOpts(t *testing.T) {
	var options = map[string]*Flag{
		" ": &Flag{Name: " ", HasArg: NoArgument},
	}

	_, err := NewParser(ParserConfig{}, nil, options, nil, nil)
	if err == nil {
		t.Errorf("Expected invalid option error")
	}
}

func TestParserInitLongOpts(t *testing.T) {
	var longOpts = make(map[string]*Flag)
	for c, i := range genOpts() {
		s := string(byte(c))
		var hasArg ArgType
		switch i % 3 {
		case 0:
			hasArg = NoArgument
		case 1:
			hasArg = RequiredArgument
		case 2:
			hasArg = OptionalArgument
		}

		longOpts[s] = &Flag{Name: s, HasArg: hasArg}
	}
	_, err := NewParser(ParserConfig{}, nil, longOpts, nil, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestParserPosixBreak(t *testing.T) {
	shopts := map[byte]*Flag{
		'a': &Flag{Name: "a", HasArg: NoArgument},
	}
	lopts := map[string]*Flag{
		"a": &Flag{Name: "a", HasArg: NoArgument},
	}
	args := []string{"--", "-a", "--a"}

	parser, err := NewParser(ParserConfig{}, shopts, lopts, args, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	for opt, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		t.Errorf("Unexpected option %s", opt.Name)
	}

	if parser.Args[0] != "-a" {
		t.Errorf("Unexpected argument %s", parser.Args[0])
	}
}

func TestParserNonOptShift(t *testing.T) {
	shopts := map[byte]*Flag{
		'a': &Flag{Name: "a", HasArg: NoArgument},
	}
	lopts := map[string]*Flag{
		"a": &Flag{Name: "a", HasArg: NoArgument},
	}
	args := []string{"param", "-a", "--a"}

	parser, err := NewParser(ParserConfig{}, shopts, lopts, args, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	for opt, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if opt.Name != "a" {
			t.Errorf("Unexpected option %s", opt.Name)
		}
	}

	if parser.Args[0] != "param" {
		t.Errorf("Unexpected argument %s", parser.Args[0])
	}
}

func TestParserNonOpt(t *testing.T) {
	shopts := map[byte]*Flag{
		'a': &Flag{Name: "a", HasArg: NoArgument},
	}
	lopts := map[string]*Flag{
		"a": &Flag{Name: "a", HasArg: NoArgument},
	}
	args := []string{"-a", "param", "--a"}

	parser, err := NewParser(ParserConfig{parseMode: ParseNonOpts}, shopts, lopts, args, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	for opt, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		switch opt.Name {
		case "a":
			// Expected option, no action needed

		case string(byte(1)):
			if opt.Arg != "param" {
				t.Errorf("Unexpected argument %s", opt.Arg)
			}

		default:
			t.Errorf("Unexpected option '%s'", opt.Name)
		}
	}

	if len(parser.Args) != 0 {
		t.Errorf("Unexpected argument %s", parser.Args[0])
	}
}

func TestParserPosixNonOpt(t *testing.T) {
	shopts := map[byte]*Flag{
		'a': &Flag{Name: "a", HasArg: NoArgument},
	}
	lopts := map[string]*Flag{
		"a": &Flag{Name: "a", HasArg: NoArgument},
	}
	args := []string{"param", "-a", "--a"}

	parser, err := NewParser(ParserConfig{parseMode: ParsePosixlyCorrect}, shopts, lopts, args, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	for opt, err := range parser.Options() {
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		t.Errorf("Unexpected option %s", opt.Name)
	}

	if parser.Args[0] != "param" {
		t.Errorf("Unexpected argument %s", parser.Args[0])
	}

	if parser.Args[1] != "-a" {
		t.Errorf("Unexpected argument %s", parser.Args[1])
	}

	if parser.Args[1] != "-a" {
		t.Errorf("Unexpected argument %s", parser.Args[1])
	}
}

func TestParserLongOptsLongPrefix(t *testing.T) {
	var longOpts = map[string]*Flag{
		"foobar": &Flag{Name: "foobar", HasArg: RequiredArgument},
	}

	args := []string{"--foo"}
	parser, err := NewParser(ParserConfig{}, nil, longOpts, args, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	for _, err := range parser.Options() {
		if err == nil {
			t.Errorf("Expected unknown option for --foo")
		}
	}
}

func TestParserLongOptsPrefix(t *testing.T) {
	var longOpts = map[string]*Flag{
		"foo":  &Flag{Name: "foo", HasArg: RequiredArgument},
		"foo=": &Flag{Name: "foo=", HasArg: NoArgument},
	}

	args := []string{"--foo="}
	parser, err := NewParser(ParserConfig{}, nil, longOpts, args, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	count := 0
	for opt, err := range parser.Options() {
		count++
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if opt.Name != "foo=" {
			t.Errorf("Expected option %s, got %s", "foo=", opt.Name)
		}
		if opt.HasArg {
			t.Errorf("Expected HasArg to be false")
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 options, got %d", count)
	}
}

func TestParserLongOptsProvidedArgs(t *testing.T) {
	var longOpts = map[string]*Flag{
		"foo": &Flag{Name: "foo", HasArg: RequiredArgument},
		"boo": &Flag{Name: "boo", HasArg: OptionalArgument},
	}

	args := []string{"--foo=bar", "--foo", "bar", "--boo=baz", "--boo", "baz"}
	parser, err := NewParser(ParserConfig{}, nil, longOpts, args, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	count := 0
	for opt, err := range parser.Options() {
		count++
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if !opt.HasArg {
			t.Errorf("Expected HasArg to be true")
		}
	}

	if count != 4 {
		t.Errorf("Expected 4 options, got %d", count)
	}
}

func TestParserLongOptsMissingOptArg(t *testing.T) {
	var longOpts = map[string]*Flag{
		"foo": &Flag{Name: "foo", HasArg: OptionalArgument},
	}

	args := []string{"--foo"}
	parser, err := NewParser(ParserConfig{}, nil, longOpts, args, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	count := 0
	for opt, err := range parser.Options() {
		count++
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if opt.HasArg {
			t.Errorf("Expected HasArg to be false")
		}
	}

	if count != 1 {
		t.Errorf("Expected 2 options, got %d", count)
	}
}

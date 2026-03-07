package optargs

import "testing"

// collectOpts iterates a parser and returns all successfully parsed options.
// Returns nil on the first iteration error.
func collectOpts(p *Parser) []Option {
	var out []Option
	for opt, err := range p.Options() {
		if err != nil {
			return nil
		}
		out = append(out, opt)
	}
	return out
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

// requireParsedOptions collects all options from a parser, failing the test on
// any iteration error.
func requireParsedOptions(t *testing.T, parser *Parser) []Option {
	t.Helper()
	var options []Option
	for opt, err := range parser.Options() {
		if err != nil {
			t.Fatalf("Options iteration failed: %v", err)
		}
		options = append(options, opt)
	}
	return options
}

// assertOptions compares actual options against expected, checking Name,
// HasArg, and Arg for each element.
func assertOptions(t *testing.T, got, want []Option) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d options, got %d", len(want), len(got))
	}
	for i, w := range want {
		if got[i].Name != w.Name {
			t.Errorf("option %d: expected name %s, got %s", i, w.Name, got[i].Name)
		}
		if got[i].HasArg != w.HasArg {
			t.Errorf("option %d: expected HasArg %t, got %t", i, w.HasArg, got[i].HasArg)
		}
		if got[i].Arg != w.Arg {
			t.Errorf("option %d: expected arg %s, got %s", i, w.Arg, got[i].Arg)
		}
	}
}

// assertArgs compares remaining positional arguments against expected values.
func assertArgs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d remaining args, got %d", len(want), len(got))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("remaining arg %d: expected %s, got %s", i, w, got[i])
		}
	}
}

// requireFirstOptError iterates parser options and returns the first error, if
// any. It fails the test only when the error expectation is violated.
func requireFirstOptError(t *testing.T, parser *Parser, expectErr bool) {
	t.Helper()
	var optErr error
	for _, err := range parser.Options() {
		if err != nil {
			optErr = err
			break
		}
	}
	if expectErr && optErr == nil {
		t.Fatal("expected error but got none")
	}
	if !expectErr && optErr != nil {
		t.Fatalf("unexpected error: %v", optErr)
	}
}

// setupChain creates a parent→child parser chain. Parent gets empty args;
// child gets the provided args. Returns the child parser.
func setupChain(t *testing.T, parentOpts, childOpts []Flag, childArgs []string) *Parser {
	t.Helper()
	parent, err := GetOptLong([]string{}, "", parentOpts)
	if err != nil {
		t.Fatalf("parent: %v", err)
	}
	child, err := GetOptLong(childArgs, "", childOpts)
	if err != nil {
		t.Fatalf("child: %v", err)
	}
	parent.AddCmd("sub", child)
	return child
}

// setupChain3 creates a grandparent→parent→child parser chain. Only the
// child receives args. Returns the child parser.
func setupChain3(t *testing.T, gpOpts, parOpts, childOpts []Flag, childArgs []string) *Parser {
	t.Helper()
	gp, err := GetOptLong([]string{}, "", gpOpts)
	if err != nil {
		t.Fatalf("grandparent: %v", err)
	}
	par, err := GetOptLong([]string{}, "", parOpts)
	if err != nil {
		t.Fatalf("parent: %v", err)
	}
	child, err := GetOptLong(childArgs, "", childOpts)
	if err != nil {
		t.Fatalf("child: %v", err)
	}
	gp.AddCmd("mid", par)
	par.AddCmd("leaf", child)
	return child
}

// childOf creates a child parser linked to a parent via AddCmd.
func childOf(t *testing.T, parentOpts, childOpts string) (*Parser, *Parser) {
	t.Helper()
	parent, err := GetOpt([]string{}, parentOpts)
	if err != nil {
		t.Fatalf("parent parser: %v", err)
	}
	child, err := GetOpt([]string{}, childOpts)
	if err != nil {
		t.Fatalf("child parser: %v", err)
	}
	parent.AddCmd("child", child)
	return parent, child
}

// childErr drains root.Options() (failing on error), then returns the
// first error from child.Options().
func childErr(t *testing.T, root, child *Parser) error {
	t.Helper()
	for _, err := range root.Options() {
		if err != nil {
			t.Fatalf("root error: %v", err)
		}
	}
	var first error
	for _, err := range child.Options() {
		if err != nil && first == nil {
			first = err
		}
	}
	return first
}

// collectNamedOptions iterates a parser and returns a map of option name → arg value.
func collectNamedOptions(t *testing.T, p *Parser) map[string]string {
	t.Helper()
	result := make(map[string]string)
	for opt, err := range p.Options() {
		if err != nil {
			t.Fatalf("unexpected error during iteration: %v", err)
		}
		result[opt.Name] = opt.Arg
	}
	return result
}

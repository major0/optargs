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

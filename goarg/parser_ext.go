//go:build goarg_ext

package goarg

import "io"

// ParserExt provides extended parser capabilities beyond base go-arg
// compatibility. Only available when built with -tags goarg_ext.
type ParserExt struct {
	*Parser
}

// NewParserExt creates an extended parser wrapping the base parser.
func NewParserExt(config Config, dest interface{}) (*ParserExt, error) {
	p, err := NewParser(config, dest)
	if err != nil {
		return nil, err
	}
	return &ParserExt{Parser: p}, nil
}

// WriteHelpErr is like WriteHelp but returns an error if writing fails.
func (pe *ParserExt) WriteHelpErr(w io.Writer) error {
	hg := NewHelpGenerator(pe.metadata, pe.config)
	return hg.WriteHelp(w)
}

// WriteUsageErr is like WriteUsage but returns an error if writing fails.
func (pe *ParserExt) WriteUsageErr(w io.Writer) error {
	hg := NewHelpGenerator(pe.metadata, pe.config)
	return hg.WriteUsage(w)
}

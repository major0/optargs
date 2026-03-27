package goarg

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Subcommand returns the active subcommand destination struct, or nil
// if no subcommand was invoked. For nested subcommands, returns the
// leaf (deepest) subcommand struct.
func (p *Parser) Subcommand() interface{} {
	return p.subcommandDest
}

// SubcommandNames returns the chain of invoked subcommand names.
// Returns nil if no subcommand was invoked.
func (p *Parser) SubcommandNames() []string {
	return p.subcommandNames
}

// FailSubcommand prints an error message with subcommand context and exits.
// The subcommand path identifies which subcommand the error applies to.
func (p *Parser) FailSubcommand(msg string, subcommand ...string) error {
	if len(subcommand) == 0 {
		p.Fail(msg)
		return nil
	}

	meta, err := p.lookupSubcommandMetadata(subcommand)
	if err != nil {
		return err
	}

	fmt.Fprintln(p.output(), msg)
	hg := NewHelpGenerator(meta, p.config)
	hg.WriteUsage(p.output()) //nolint:errcheck
	p.config.Exit(1)
	return nil
}

// WriteHelpForSubcommand writes help text for a specific subcommand path.
func (p *Parser) WriteHelpForSubcommand(w io.Writer, subcommand ...string) error {
	meta, err := p.lookupSubcommandMetadata(subcommand)
	if err != nil {
		return err
	}
	hg := NewHelpGenerator(meta, p.config)
	return hg.WriteHelp(w)
}

// WriteUsageForSubcommand writes usage text for a specific subcommand path.
func (p *Parser) WriteUsageForSubcommand(w io.Writer, subcommand ...string) error {
	meta, err := p.lookupSubcommandMetadata(subcommand)
	if err != nil {
		return err
	}
	hg := NewHelpGenerator(meta, p.config)
	return hg.WriteUsage(w)
}

// lookupSubcommandMetadata walks the metadata tree to find the metadata
// for a subcommand path.
func (p *Parser) lookupSubcommandMetadata(path []string) (*StructMetadata, error) {
	meta := p.metadata
	for _, name := range path {
		found := false
		for cmdName, subMeta := range meta.Subcommands {
			if strings.EqualFold(cmdName, name) {
				meta = subMeta
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown subcommand: %s", name)
		}
	}
	return meta, nil
}

// recordSubcommandChain walks the core parser's ActiveCommand chain and
// records the subcommand names and leaf dest struct.
func (p *Parser) recordSubcommandChain(destValue reflect.Value, ci *CoreIntegration) {
	p.subcommandNames = nil
	p.subcommandDest = nil

	currentParser := p.coreParser
	currentDest := destValue
	currentMeta := ci.metadata

	for {
		name, child := currentParser.ActiveCommand()
		if name == "" || child == nil {
			break
		}
		p.subcommandNames = append(p.subcommandNames, name)

		// Find the struct field for this subcommand
		fv, subMeta, err := (&CoreIntegration{metadata: currentMeta}).findSubcommandField(currentDest, name)
		if err != nil {
			break
		}
		if fv.Kind() == reflect.Ptr && !fv.IsNil() {
			p.subcommandDest = fv.Interface()
			currentDest = fv.Elem()
		}
		currentMeta = subMeta
		currentParser = child
	}
}

// output returns the configured output writer, defaulting to os.Stderr.
func (p *Parser) output() io.Writer {
	if p.config.Out != nil {
		return p.config.Out
	}
	return defaultOutput
}

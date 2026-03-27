package goarg

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/major0/optargs"
)

// Parser provides the main parsing interface - identical to alexflint/go-arg
type Parser struct {
	config   Config
	dest     interface{}
	metadata *StructMetadata

	// Direct OptArgs Core integration
	coreParser *optargs.Parser

	// Error translation
	errorTranslator *ErrorTranslator
}

// Config matches alexflint/go-arg configuration options exactly
type Config struct {
	Program     string
	Description string
	Version     string
	Epilogue    string
	IgnoreEnv   bool
	// Additional fields for full alexflint/go-arg compatibility
	IgnoreDefault bool
	Exit          func(int)
}

// Parse parses command line arguments into the destination struct
func Parse(dest interface{}) error {
	parser, err := NewParser(Config{}, dest)
	if err != nil {
		return err
	}
	return parser.Parse(os.Args[1:])
}

// ParseArgs parses the provided arguments into the destination struct
func ParseArgs(dest interface{}, args []string) error {
	parser, err := NewParser(Config{}, dest)
	if err != nil {
		return err
	}
	return parser.Parse(args)
}

// MustParse parses command line arguments, prints help/version on the
// corresponding sentinel errors, and exits on any error. Returns the
// parser on success so callers can inspect subcommand state.
func MustParse(dest interface{}) *Parser {
	p, err := NewParser(Config{}, dest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
		return nil
	}
	err = p.Parse(os.Args[1:])
	p.handleMustParseError(err)
	return p
}

// NewParser creates a new parser with the given configuration
func NewParser(config Config, dest interface{}) (*Parser, error) {
	if dest == nil {
		return nil, fmt.Errorf("destination cannot be nil")
	}

	// Validate that dest is a pointer to a struct
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("destination must be a pointer to a struct, got %T", dest)
	}

	destElem := destValue.Elem()
	if destElem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("destination must be a pointer to a struct, got pointer to %s", destElem.Kind())
	}

	// Parse struct metadata
	tagParser := &TagParser{}
	metadata, err := tagParser.ParseStruct(dest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse struct: %w", err)
	}

	// Detect Versioned/Described/Epilogued interfaces on dest struct
	if v, ok := dest.(Versioned); ok && config.Version == "" {
		config.Version = v.Version()
	}
	if d, ok := dest.(Described); ok && config.Description == "" {
		config.Description = d.Description()
	}
	if e, ok := dest.(Epilogued); ok && config.Epilogue == "" {
		config.Epilogue = e.Epilogue()
	}

	// Set default exit function if not provided
	if config.Exit == nil {
		config.Exit = os.Exit
	}

	return &Parser{
		config:          config,
		dest:            dest,
		metadata:        metadata,
		errorTranslator: &ErrorTranslator{},
	}, nil
}

// Parse parses the given arguments
func (p *Parser) Parse(args []string) error {
	if args == nil {
		args = os.Args[1:]
	}

	ci := &CoreIntegration{
		metadata: p.metadata,
		config:   p.config,
	}
	destValue := reflect.ValueOf(p.dest).Elem()

	// Build parser with Handle callbacks
	coreParser, err := ci.CreateParserWithHandlers(args, destValue)
	if err != nil {
		return p.translateError(err, "")
	}

	// Register subcommands
	if err := ci.RegisterSubcommands(coreParser, destValue); err != nil {
		return p.translateError(err, "")
	}

	p.coreParser = coreParser

	// Iterate — Handle callbacks fire automatically
	for _, err := range coreParser.Options() {
		if err != nil {
			// Sentinel errors pass through without translation
			if errors.Is(err, ErrHelp) || errors.Is(err, ErrVersion) {
				return err
			}
			return p.translateError(err, "")
		}
	}

	// Subcommand dispatch: use core's ActiveCommand() to detect which
	// subcommand was dispatched, iterate its Options(), run PostParse,
	// and walk recursively for nested subcommands.
	if len(p.metadata.Subcommands) > 0 {
		invokedName, childParser := coreParser.ActiveCommand()

		if invokedName != "" && childParser != nil {
			if err := ci.dispatchSubcommand(childParser, invokedName, destValue, p); err != nil {
				return err
			}
		}

		// Nil out non-invoked subcommand fields so callers can detect
		// which subcommand was selected.
		for name := range p.metadata.Subcommands {
			if strings.EqualFold(name, invokedName) {
				continue
			}
			fv, _, err := ci.findSubcommandField(destValue, name)
			if err != nil {
				continue
			}
			if fv.Kind() == reflect.Ptr {
				fv.Set(reflect.Zero(fv.Type()))
			}
		}
	}

	// Post-parse: positionals, env vars, defaults, required validation
	return p.translateError(ci.PostParse(coreParser, destValue), "")
}

// WriteHelp writes help text to the provided writer
func (p *Parser) WriteHelp(w io.Writer) {
	helpGenerator := NewHelpGenerator(p.metadata, p.config)
	helpGenerator.WriteHelp(w) //nolint:errcheck // matches upstream go-arg API (no error return)
}

// WriteUsage writes usage text to the provided writer
func (p *Parser) WriteUsage(w io.Writer) {
	helpGenerator := NewHelpGenerator(p.metadata, p.config)
	helpGenerator.WriteUsage(w) //nolint:errcheck // matches upstream go-arg API (no error return)
}

// Fail prints an error message and exits
func (p *Parser) Fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	p.WriteUsage(os.Stderr)
	p.config.Exit(1)
}

// MustParse parses the given arguments, prints help/version on the
// corresponding sentinel errors, and exits on any error.
func (p *Parser) MustParse(args []string) {
	err := p.Parse(args)
	p.handleMustParseError(err)
}

// handleMustParseError handles the result of Parse for MustParse callers.
// ErrHelp prints help and exits 0, ErrVersion prints version and exits 0,
// any other error prints the error with usage and exits 1.
func (p *Parser) handleMustParseError(err error) {
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, ErrHelp):
		p.WriteHelp(os.Stdout)
		p.config.Exit(0)
	case errors.Is(err, ErrVersion):
		fmt.Fprintln(os.Stdout, p.config.Version)
		p.config.Exit(0)
	default:
		fmt.Fprintln(os.Stderr, err)
		p.WriteUsage(os.Stderr)
		p.config.Exit(1)
	}
}

// translateError translates an error using the error translator with context
func (p *Parser) translateError(err error, fieldName string) error {
	if err == nil {
		return nil
	}

	context := ParseContext{
		StructType: reflect.TypeOf(p.dest).Elem(),
		FieldName:  fieldName,
	}

	return p.errorTranslator.TranslateError(err, context)
}

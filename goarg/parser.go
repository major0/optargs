package goarg

import (
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
	shortOpts  map[byte]*optargs.Flag
	longOpts   map[string]*optargs.Flag

	// Error translation
	errorTranslator *ErrorTranslator
}

// Config matches alexflint/go-arg configuration options exactly
type Config struct {
	Program     string
	Description string
	Version     string
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

// MustParse parses command line arguments and panics on error
func MustParse(dest interface{}) {
	if err := Parse(dest); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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

	// Set default exit function if not provided
	if config.Exit == nil {
		config.Exit = os.Exit
	}

	return &Parser{
		config:          config,
		dest:            dest,
		metadata:        metadata,
		shortOpts:       make(map[byte]*optargs.Flag),
		longOpts:        make(map[string]*optargs.Flag),
		errorTranslator: &ErrorTranslator{},
	}, nil
}

// Parse parses the given arguments
func (p *Parser) Parse(args []string) error {
	if args == nil {
		args = os.Args[1:]
	}

	ci := &CoreIntegration{
		metadata:    p.metadata,
		shortOpts:   p.shortOpts,
		longOpts:    p.longOpts,
		positionals: []PositionalArg{},
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
			return p.translateError(err, "")
		}
	}

	// Subcommand dispatch: detect which subcommand was invoked and run
	// its Options() iteration + PostParse. Non-invoked subcommand fields
	// are reset to nil so callers can detect the active subcommand.
	if len(p.metadata.Subcommands) > 0 {
		invokedName := ""
		// Scan original args for a subcommand name (case-insensitive),
		// matching the same detection strategy as the previous implementation.
		for _, arg := range args {
			if strings.HasPrefix(arg, "-") {
				continue
			}
			if _, _, err := ci.findSubcommandField(destValue, arg); err == nil {
				invokedName = arg
				break
			}
		}

		if invokedName != "" {
			fieldValue, subMeta, _ := ci.findSubcommandField(destValue, invokedName)

			// Get the child parser registered by RegisterSubcommands
			childParser, ok := coreParser.GetCommand(invokedName)
			if !ok {
				return p.translateError(fmt.Errorf("subcommand parser not found for %s", invokedName), invokedName)
			}

			// Iterate child Options() — Handle callbacks fire for subcommand options
			for _, err := range childParser.Options() {
				if err != nil {
					return p.translateError(err, "")
				}
			}

			// PostParse on the subcommand
			subDestValue := fieldValue.Elem()
			childCI := &CoreIntegration{
				metadata:    subMeta,
				shortOpts:   make(map[byte]*optargs.Flag),
				longOpts:    make(map[string]*optargs.Flag),
				positionals: []PositionalArg{},
			}
			childCI.buildPositionalArgs()
			if err := childCI.PostParse(childParser, subDestValue); err != nil {
				return p.translateError(err, "")
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
	helpGenerator.WriteHelp(w)
}

// WriteUsage writes usage text to the provided writer
func (p *Parser) WriteUsage(w io.Writer) {
	helpGenerator := NewHelpGenerator(p.metadata, p.config)
	helpGenerator.WriteUsage(w)
}

// Fail prints an error message and exits
func (p *Parser) Fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	p.WriteUsage(os.Stderr)
	p.config.Exit(1)
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

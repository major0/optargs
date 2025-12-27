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
		config:    config,
		dest:      dest,
		metadata:  metadata,
		shortOpts: make(map[byte]*optargs.Flag),
		longOpts:  make(map[string]*optargs.Flag),
	}, nil
}

// Parse parses the given arguments
func (p *Parser) Parse(args []string) error {
	if args == nil {
		args = os.Args[1:]
	}

	// Create core integration
	coreIntegration := &CoreIntegration{
		metadata:    p.metadata,
		shortOpts:   p.shortOpts,
		longOpts:    p.longOpts,
		positionals: []PositionalArg{},
	}

	// Build OptArgs Core parser with command support
	coreParser, err := coreIntegration.CreateParser(args)
	if err != nil {
		return err
	}

	p.coreParser = coreParser

	// Check if we have subcommands and if a subcommand was invoked
	if len(p.metadata.Subcommands) > 0 && len(args) > 0 {
		// Look for subcommand in arguments
		for i, arg := range args {
			// Try case insensitive lookup for subcommands
			subMetadata, subcommandName := p.findSubcommand(arg)
			if subMetadata != nil {
				// Found subcommand, dispatch to it
				subParser, err := coreParser.Commands.ExecuteCommandCaseInsensitive(subcommandName, args[i+1:], true) // Always use case insensitive for go-arg
				if err != nil {
					return fmt.Errorf("failed to execute subcommand %s: %w", arg, err)
				}
				
				// Process the subcommand's results
				subIntegration := &CoreIntegration{
					metadata:    subMetadata,
					shortOpts:   make(map[byte]*optargs.Flag),
					longOpts:    make(map[string]*optargs.Flag),
					positionals: []PositionalArg{},
				}
				
				// Get the subcommand field from the destination struct
				destValue := reflect.ValueOf(p.dest).Elem()
				var subcommandField reflect.Value
				for j := 0; j < destValue.NumField(); j++ {
					field := destValue.Type().Field(j)
					fieldMeta, _ := (&TagParser{}).ParseField(field)
					if fieldMeta.IsSubcommand && strings.EqualFold(fieldMeta.SubcommandName, arg) {
						subcommandField = destValue.Field(j)
						break
					}
				}
				
				if !subcommandField.IsValid() {
					return fmt.Errorf("subcommand field not found for %s", arg)
				}
				
				// Ensure subcommand field is initialized
				if subcommandField.IsNil() {
					subcommandField.Set(reflect.New(subcommandField.Type().Elem()))
				}
				
				// Process subcommand results
				return subIntegration.ProcessResults(subParser, subcommandField.Interface())
			}
		}
	}

	// No subcommand found, process as regular parsing
	return coreIntegration.ProcessResults(coreParser, p.dest)
}

// WriteHelp writes help text to the provided writer
func (p *Parser) WriteHelp(w io.Writer) {
	if p.metadata == nil {
		fmt.Fprintln(w, "No help available")
		return
	}

	// Generate help text compatible with alexflint/go-arg format
	program := p.config.Program
	if program == "" {
		program = os.Args[0]
	}

	fmt.Fprintf(w, "Usage: %s", program)
	
	// Add options placeholder
	if len(p.metadata.Fields) > 0 {
		fmt.Fprint(w, " [OPTIONS]")
	}

	fmt.Fprintln(w)

	// Add description if available
	if p.config.Description != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, p.config.Description)
	}

	// Add options section
	if len(p.metadata.Fields) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		
		for _, field := range p.metadata.Fields {
			if field.Positional {
				continue
			}
			
			var optStr string
			if field.Short != "" && field.Long != "" {
				optStr = fmt.Sprintf("  -%s, --%s", field.Short, field.Long)
			} else if field.Short != "" {
				optStr = fmt.Sprintf("  -%s", field.Short)
			} else if field.Long != "" {
				optStr = fmt.Sprintf("  --%s", field.Long)
			}
			
			if field.Help != "" {
				fmt.Fprintf(w, "%-20s %s\n", optStr, field.Help)
			} else {
				fmt.Fprintln(w, optStr)
			}
		}
	}

	// Add version if available
	if p.config.Version != "" {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Version: %s\n", p.config.Version)
	}
}

// WriteUsage writes usage text to the provided writer
func (p *Parser) WriteUsage(w io.Writer) {
	program := p.config.Program
	if program == "" {
		program = os.Args[0]
	}

	fmt.Fprintf(w, "Usage: %s", program)
	
	// Add options placeholder
	if p.metadata != nil && len(p.metadata.Fields) > 0 {
		fmt.Fprint(w, " [OPTIONS]")
		
		// Add positional arguments
		for _, field := range p.metadata.Fields {
			if field.Positional {
				if field.Required {
					fmt.Fprintf(w, " %s", field.Name)
				} else {
					fmt.Fprintf(w, " [%s]", field.Name)
				}
			}
		}
	}

	fmt.Fprintln(w)
}

// Fail prints an error message and exits
func (p *Parser) Fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	p.WriteUsage(os.Stderr)
	p.config.Exit(1)
}

// findSubcommand performs case insensitive lookup for subcommands
func (p *Parser) findSubcommand(name string) (*StructMetadata, string) {
	// First try exact match
	if metadata, exists := p.metadata.Subcommands[name]; exists {
		return metadata, name
	}
	
	// Then try case insensitive match
	for cmdName, metadata := range p.metadata.Subcommands {
		if strings.EqualFold(cmdName, name) {
			return metadata, cmdName
		}
	}
	
	return nil, ""
}
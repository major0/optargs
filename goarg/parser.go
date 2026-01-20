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

	// Validate compatibility with upstream alexflint/go-arg
	if err := validateUpstreamCompatibility(metadata, destElem.Type()); err != nil {
		return nil, err
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
		return p.translateError(err, "")
	}

	p.coreParser = coreParser

	// Check if we have subcommands - let OptArgs Core handle the full parsing including global flags
	if len(p.metadata.Subcommands) > 0 {
		// Look for subcommand in arguments to see if one was invoked
		var subcommandFound string
		var subMetadata *StructMetadata
		var unknownSubcommand string

		// Find the first non-flag argument - this would be the subcommand
		for i, arg := range args {
			// Skip flags and their arguments
			if strings.HasPrefix(arg, "-") {
				// Skip this flag and potentially its argument
				continue
			}

			// Skip arguments that follow flags requiring values
			if i > 0 && strings.HasPrefix(args[i-1], "-") {
				// Check if the previous flag requires an argument
				// This is a simplified check - in practice we'd need to know which flags require args
				prevFlag := args[i-1]
				if !strings.Contains(prevFlag, "=") && !isBooleanFlag(prevFlag, p.metadata) {
					// Previous flag likely requires an argument, so this arg is its value
					continue
				}
			}

			// This looks like a potential subcommand (first non-flag, non-flag-value argument)
			if metadata, cmdName := p.findSubcommand(arg); metadata != nil {
				subcommandFound = cmdName
				subMetadata = metadata
				break
			} else {
				// This is an unknown subcommand
				unknownSubcommand = arg
				break
			}
		}

		// If we found an unknown subcommand, return error to match upstream
		if unknownSubcommand != "" && subcommandFound == "" {
			return fmt.Errorf("Parse error: invalid subcommand: %s", unknownSubcommand)
		}

		if subcommandFound != "" {
			// Found subcommand, let OptArgs Core parse the entire argument list
			// This allows global flags before the subcommand to be parsed correctly
			subParser, err := coreParser.Commands.ExecuteCommandCaseInsensitive(subcommandFound, args, true) // Pass full args, not just subcommand args
			if err != nil {
				return p.translateError(err, subcommandFound)
			}

			// Get the subcommand field from the destination struct
			destValue := reflect.ValueOf(p.dest).Elem()
			var subcommandField reflect.Value
			for j := 0; j < destValue.NumField(); j++ {
				field := destValue.Type().Field(j)
				fieldMeta, _ := (&TagParser{}).ParseField(field)
				if fieldMeta.IsSubcommand && strings.EqualFold(fieldMeta.SubcommandName, subcommandFound) {
					subcommandField = destValue.Field(j)
					break
				}
			}

			if !subcommandField.IsValid() {
				return p.translateError(fmt.Errorf("subcommand field not found for %s", subcommandFound), subcommandFound)
			}

			// Ensure subcommand field is initialized
			if subcommandField.IsNil() {
				subcommandField.Set(reflect.New(subcommandField.Type().Elem()))
			}

			// Process all options from the subcommand parser in a single pass
			// This handles both inherited options (set on parent) and subcommand options (set on subcommand)
			return p.translateError(p.processOptionsWithInheritance(subParser, coreIntegration, subMetadata, subcommandField.Interface()), "")
		}
	}

	// No subcommand found, process as regular parsing
	return p.translateError(coreIntegration.ProcessResults(coreParser, p.dest), "")
}

// processOptionsWithInheritance processes options from subcommand parser, handling inheritance in a single pass
func (p *Parser) processOptionsWithInheritance(subParser *optargs.Parser, parentIntegration *CoreIntegration, subMetadata *StructMetadata, subcommandDest interface{}) error {
	// Create integration for subcommand
	subIntegration := &CoreIntegration{
		metadata:    subMetadata,
		shortOpts:   make(map[byte]*optargs.Flag),
		longOpts:    make(map[string]*optargs.Flag),
		positionals: []PositionalArg{},
	}

	// Build subcommand option mappings and positional arguments
	subIntegration.BuildLongOpts()
	subIntegration.buildPositionalArgs() // This was missing!

	destValue := reflect.ValueOf(p.dest).Elem()
	subDestValue := reflect.ValueOf(subcommandDest).Elem()

	// Process all options from the subcommand parser in a single pass
	for option, err := range subParser.Options() {
		if err != nil {
			return p.translateError(err, "")
		}

		// First, try to find the option in the subcommand metadata
		subField := p.findFieldInMetadata(option.Name, subMetadata)
		if subField != nil {
			// This is a subcommand option, set it on the subcommand struct
			fieldValue := subDestValue.FieldByName(subField.Name)
			if fieldValue.IsValid() && fieldValue.CanSet() {
				var arg string
				if option.HasArg {
					arg = option.Arg
				}

				if err := subIntegration.setFieldValue(fieldValue, subField, arg); err != nil {
					return p.translateError(err, subField.Name)
				}
			}
		} else {
			// Not found in subcommand, check if it's a parent option (inherited)
			parentField := p.findParentFieldForOption(option.Name)
			if parentField != nil {
				// This is an inherited parent option, set it on the parent struct
				fieldValue := destValue.FieldByName(parentField.Name)
				if fieldValue.IsValid() && fieldValue.CanSet() {
					var arg string
					if option.HasArg {
						arg = option.Arg
					}

					if err := parentIntegration.setFieldValue(fieldValue, parentField, arg); err != nil {
						return p.translateError(err, parentField.Name)
					}
				}
			}
			// If not found in either parent or subcommand, it's an unknown option (already handled by OptArgs Core)
		}
	}

	// Process positional arguments for subcommand
	if err := subIntegration.processPositionalArgs(subParser, subDestValue); err != nil {
		return p.translateError(err, "")
	}

	// Process environment variables for subcommand
	if err := subIntegration.processEnvironmentVariables(subDestValue); err != nil {
		return p.translateError(err, "")
	}

	// Set default values for subcommand
	if err := subIntegration.setDefaultValues(subDestValue); err != nil {
		return p.translateError(err, "")
	}

	// Validate required fields for subcommand
	typeConverter := &TypeConverter{}
	if err := typeConverter.ValidateRequired(subcommandDest, subMetadata); err != nil {
		return p.translateError(err, "")
	}

	// Process nested subcommands if any were invoked
	if len(subMetadata.Subcommands) > 0 {
		// Check if a nested subcommand was invoked by looking at the subcommand parser's commands
		if subParser.Commands != nil {
			// Look for executed subcommands in the subcommand parser
			for nestedCmdName, nestedMetadata := range subMetadata.Subcommands {
				// Check if this nested command exists
				if nestedParser, exists := subParser.Commands.GetCommand(nestedCmdName); exists && nestedParser != nil {
					// Check if this nested command should be processed by looking at its arguments
					// If the nested parser has arguments, it means this command was invoked
					if len(nestedParser.Args) > 0 {
						// Find the corresponding field in the subcommand struct
						for i := 0; i < subDestValue.NumField(); i++ {
							field := subDestValue.Type().Field(i)
							fieldMeta, _ := (&TagParser{}).ParseField(field)
							if fieldMeta.IsSubcommand && strings.EqualFold(fieldMeta.SubcommandName, nestedCmdName) {
								nestedField := subDestValue.Field(i)

								// Initialize the nested subcommand field if it's nil
								if nestedField.IsNil() {
									nestedField.Set(reflect.New(nestedField.Type().Elem()))
								}

								// Recursively process the nested subcommand
								if err := p.processOptionsWithInheritance(nestedParser, subIntegration, nestedMetadata, nestedField.Interface()); err != nil {
									return p.translateError(err, nestedCmdName)
								}
								break
							}
						}
					}
				}
			}
		}
	}

	// Process environment variables and defaults for parent as well
	if err := parentIntegration.processEnvironmentVariables(destValue); err != nil {
		return p.translateError(err, "")
	}

	if err := parentIntegration.setDefaultValues(destValue); err != nil {
		return p.translateError(err, "")
	}

	return nil
}

// findFieldInMetadata finds a field in the given metadata that matches the option name
func (p *Parser) findFieldInMetadata(optionName string, metadata *StructMetadata) *FieldMetadata {
	for _, field := range metadata.Fields {
		if field.Short == optionName || field.Long == optionName {
			return &field
		}
	}
	return nil
}

// findParentFieldForOption finds a field in the parent metadata that matches the option name
func (p *Parser) findParentFieldForOption(optionName string) *FieldMetadata {
	for _, field := range p.metadata.Fields {
		if field.Short == optionName || field.Long == optionName {
			return &field
		}
	}
	return nil
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

// isBooleanFlag checks if a flag is a boolean flag that doesn't require an argument
func isBooleanFlag(flag string, metadata *StructMetadata) bool {
	// Remove leading dashes
	flagName := strings.TrimPrefix(flag, "--")
	flagName = strings.TrimPrefix(flagName, "-")

	// Look for this flag in metadata
	for _, field := range metadata.Fields {
		if field.Short == flagName || field.Long == flagName {
			return field.Type.Kind() == reflect.Bool
		}
	}
	return false
}

// validateUpstreamCompatibility validates that the struct is compatible with upstream alexflint/go-arg
func validateUpstreamCompatibility(metadata *StructMetadata, structType reflect.Type) error {
	// Check for slice fields with default values - upstream doesn't support this
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Type.Kind() == reflect.Slice {
			// Check if this field has a default value
			if defaultTag, exists := field.Tag.Lookup("default"); exists && defaultTag != "" {
				return fmt.Errorf("%s.%s: default values are not supported for slice or map fields", structType.Name(), field.Name)
			}
		}
	}
	return nil
}

package goarg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/major0/optargs"
)

// CoreIntegration handles direct translation to OptArgs Core
type CoreIntegration struct {
	metadata    *StructMetadata
	shortOpts   map[byte]*optargs.Flag
	longOpts    map[string]*optargs.Flag
	positionals []PositionalArg
}

// PositionalArg represents a positional argument
type PositionalArg struct {
	Field    *FieldMetadata
	Required bool
	Multiple bool
}

// BuildOptString builds the optstring for OptArgs Core
func (ci *CoreIntegration) BuildOptString() string {
	var optstring strings.Builder

	for _, field := range ci.metadata.Fields {
		if field.Positional || field.IsSubcommand {
			continue
		}

		if field.Short != "" {
			optstring.WriteString(field.Short)

			// Add colon for required arguments
			if field.ArgType == optargs.RequiredArgument {
				optstring.WriteString(":")
			} else if field.ArgType == optargs.OptionalArgument {
				optstring.WriteString("::")
			}
		}
	}

	return optstring.String()
}

// BuildLongOpts builds the long options for OptArgs Core
func (ci *CoreIntegration) BuildLongOpts() []optargs.Flag {
	var longOpts []optargs.Flag

	for _, field := range ci.metadata.Fields {
		if field.Positional || field.IsSubcommand {
			continue
		}

		if field.Long != "" {
			flag := optargs.Flag{
				Name:   field.Long,
				HasArg: field.ArgType,
			}
			longOpts = append(longOpts, flag)

			// Store mapping for later processing
			if field.Short != "" {
				ci.shortOpts[field.Short[0]] = &flag
			}
			ci.longOpts[field.Long] = &flag
		}
	}

	return longOpts
}

// CreateParser creates an OptArgs Core parser with command support
func (ci *CoreIntegration) CreateParser(args []string) (*optargs.Parser, error) {
	return ci.CreateParserWithParent(args, nil)
}

// CreateParserWithParent creates an OptArgs Core parser with command support and parent relationship
func (ci *CoreIntegration) CreateParserWithParent(args []string, parent *optargs.Parser) (*optargs.Parser, error) {
	// Build positional arguments
	ci.buildPositionalArgs()

	// Build option string and long options
	optstring := ci.BuildOptString()
	longOpts := ci.BuildLongOpts()

	// Create OptArgs Core parser
	parser, err := optargs.GetOptLong(args, optstring, longOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create OptArgs parser: %w", err)
	}

	// Set up parent relationship for option inheritance
	if parent != nil {
		// The parent relationship is set automatically by OptArgs Core's AddCmd method
		// We don't need to set it manually here
	}

	// Register subcommands using OptArgs Core command system
	if len(ci.metadata.Subcommands) > 0 {
		for cmdName, subMetadata := range ci.metadata.Subcommands {
			// Create integration for subcommand
			subIntegration := &CoreIntegration{
				metadata:    subMetadata,
				shortOpts:   make(map[byte]*optargs.Flag),
				longOpts:    make(map[string]*optargs.Flag),
				positionals: []PositionalArg{},
			}

			// Create subcommand parser with parent relationship for option inheritance
			subParser, err := subIntegration.CreateParserWithParent([]string{}, parser)
			if err != nil {
				return nil, fmt.Errorf("failed to create subcommand parser for %s: %w", cmdName, err)
			}

			// Register subcommand with main parser (this sets parent relationship in OptArgs Core)
			parser.AddCmd(cmdName, subParser)
		}
	}

	return parser, nil
}

// ProcessResults processes parsing results from OptArgs Core
func (ci *CoreIntegration) ProcessResults(parser *optargs.Parser, dest interface{}) error {
	destValue := reflect.ValueOf(dest).Elem()
	destType := destValue.Type()

	// Process parsed options using the iterator
	for option, err := range parser.Options() {
		if err != nil {
			return fmt.Errorf("parsing error: %w", err)
		}

		// Find the corresponding field
		field, err := ci.findFieldForOption(option, destType)
		if err != nil {
			// If option not found in current parser, it might be handled by parent
			// This is expected behavior with command inheritance
			continue
		}

		if field == nil {
			continue // Skip unknown options
		}

		// Set the field value
		fieldValue := destValue.FieldByName(field.Name)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			return fmt.Errorf("cannot set field %s", field.Name)
		}

		var arg string
		if option.HasArg {
			arg = option.Arg
		}

		if err := ci.setFieldValue(fieldValue, field, arg); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	// Process positional arguments
	if err := ci.processPositionalArgs(parser, destValue); err != nil {
		return fmt.Errorf("failed to process positional arguments: %w", err)
	}

	// Process environment variables
	if err := ci.processEnvironmentVariables(destValue); err != nil {
		return fmt.Errorf("failed to process environment variables: %w", err)
	}

	// Set default values for unset fields
	if err := ci.setDefaultValues(destValue); err != nil {
		return fmt.Errorf("failed to set default values: %w", err)
	}

	// Validate required fields
	typeConverter := &TypeConverter{}
	if err := typeConverter.ValidateRequired(dest, ci.metadata); err != nil {
		return err
	}

	return nil
}

// buildPositionalArgs builds the list of positional arguments
func (ci *CoreIntegration) buildPositionalArgs() {
	ci.positionals = []PositionalArg{}

	for _, field := range ci.metadata.Fields {
		if field.Positional {
			positional := PositionalArg{
				Field:    &field,
				Required: field.Required,
				Multiple: field.Type.Kind() == reflect.Slice,
			}
			ci.positionals = append(ci.positionals, positional)
		}
	}
}

// findFieldForOption finds the field metadata for a given option
func (ci *CoreIntegration) findFieldForOption(option optargs.Option, destType reflect.Type) (*FieldMetadata, error) {
	// Look for matching field by option name
	for _, field := range ci.metadata.Fields {
		if field.Short == option.Name || field.Long == option.Name {
			return &field, nil
		}
	}

	// Option not found in current metadata - this is expected with command inheritance
	return nil, nil
}

// setFieldValue sets a field value based on the parsed argument
func (ci *CoreIntegration) setFieldValue(fieldValue reflect.Value, field *FieldMetadata, arg string) error {
	typeConverter := &TypeConverter{}

	// Handle boolean flags specially - they are set to true when present without argument
	if field.Type.Kind() == reflect.Bool && arg == "" {
		fieldValue.SetBool(true)
		return nil
	}

	// Handle slice types - append to existing values
	if field.Type.Kind() == reflect.Slice {
		// Convert the argument to the element type
		elemType := field.Type.Elem()
		converted, err := typeConverter.ConvertValue(arg, elemType)
		if err != nil {
			return fmt.Errorf("failed to convert slice element: %w", err)
		}

		// Append to existing slice
		newSlice := reflect.Append(fieldValue, reflect.ValueOf(converted))
		fieldValue.Set(newSlice)
		return nil
	}

	// For all other types, use the type converter
	converted, err := typeConverter.ConvertValue(arg, field.Type)
	if err != nil {
		return fmt.Errorf("failed to convert value '%s' for field %s: %w", arg, field.Name, err)
	}

	return typeConverter.SetField(fieldValue, converted)
}

// setScalarValue sets a scalar value (helper for slice and pointer handling)
func (ci *CoreIntegration) setScalarValue(fieldValue reflect.Value, fieldType reflect.Type, arg string) error {
	typeConverter := &TypeConverter{}

	converted, err := typeConverter.ConvertValue(arg, fieldType)
	if err != nil {
		return fmt.Errorf("failed to convert scalar value: %w", err)
	}

	return typeConverter.SetField(fieldValue, converted)
}

// processPositionalArgs processes positional arguments from remaining args
func (ci *CoreIntegration) processPositionalArgs(parser *optargs.Parser, destValue reflect.Value) error {
	// Get remaining arguments after option parsing
	remainingArgs := parser.Args
	argIndex := 0

	for _, positional := range ci.positionals {
		field := positional.Field
		fieldValue := destValue.FieldByName(field.Name)

		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			return fmt.Errorf("cannot set positional field %s", field.Name)
		}

		if positional.Multiple {
			// For slice types, consume all remaining arguments
			slice := reflect.MakeSlice(field.Type, 0, len(remainingArgs)-argIndex)

			for argIndex < len(remainingArgs) {
				elemType := field.Type.Elem()
				elemValue := reflect.New(elemType).Elem()

				if err := ci.setScalarValue(elemValue, elemType, remainingArgs[argIndex]); err != nil {
					return fmt.Errorf("failed to set positional argument %d: %w", argIndex, err)
				}

				slice = reflect.Append(slice, elemValue)
				argIndex++
			}

			fieldValue.Set(slice)
		} else {
			// For single values, consume one argument
			if argIndex >= len(remainingArgs) {
				if positional.Required {
					return fmt.Errorf("missing required positional argument: %s", field.Name)
				}
				continue
			}

			if err := ci.setScalarValue(fieldValue, field.Type, remainingArgs[argIndex]); err != nil {
				return fmt.Errorf("failed to set positional argument %s: %w", field.Name, err)
			}

			argIndex++
		}
	}

	return nil
}

// processEnvironmentVariables processes environment variable fallbacks
func (ci *CoreIntegration) processEnvironmentVariables(destValue reflect.Value) error {
	for _, field := range ci.metadata.Fields {
		if field.Env == "" {
			continue
		}

		fieldValue := destValue.FieldByName(field.Name)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			continue
		}

		// Only set from environment if field is not already set
		if ci.isFieldSet(fieldValue, field.Type) {
			continue
		}

		envValue, exists := ci.getEnvironmentValue(&field)
		if !exists {
			continue
		}

		if err := ci.setScalarValue(fieldValue, field.Type, envValue); err != nil {
			return fmt.Errorf("failed to set environment variable %s for field %s: %w", field.Env, field.Name, err)
		}
	}

	return nil
}

// setDefaultValues sets default values for unset fields
func (ci *CoreIntegration) setDefaultValues(destValue reflect.Value) error {
	typeConverter := &TypeConverter{}

	for _, field := range ci.metadata.Fields {
		fieldValue := destValue.FieldByName(field.Name)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			continue
		}

		// Only set default if field is not already set
		if ci.isFieldSet(fieldValue, field.Type) {
			continue
		}

		// Get default value from struct tag
		defaultValue := typeConverter.GetDefault(reflect.StructField{
			Name: field.Name,
			Type: field.Type,
			Tag:  reflect.StructTag(field.Tag),
		})

		if defaultValue != nil {
			if err := typeConverter.SetField(fieldValue, defaultValue); err != nil {
				return fmt.Errorf("failed to set default value for field %s: %w", field.Name, err)
			}
		}
	}

	return nil
}

// isFieldSet checks if a field has been set (not zero value)
func (ci *CoreIntegration) isFieldSet(fieldValue reflect.Value, fieldType reflect.Type) bool {
	zero := reflect.Zero(fieldType)
	return !reflect.DeepEqual(fieldValue.Interface(), zero.Interface())
}

// getEnvironmentValue gets the value from environment variable
func (ci *CoreIntegration) getEnvironmentValue(field *FieldMetadata) (string, bool) {
	tagParser := &TagParser{}
	return tagParser.GetEnvironmentValue(field)
}

// buildShortOptMap builds a map from struct metadata short options to Flag pointers.
// Each flag's HasArg is set based on the field's ArgType. Short and long flags for
// the same field share the same *optargs.Flag pointer so Handle is set once.
func (ci *CoreIntegration) buildShortOptMap() map[byte]*optargs.Flag {
	shortOpts := make(map[byte]*optargs.Flag)

	for i := range ci.metadata.Fields {
		field := &ci.metadata.Fields[i]
		if field.Positional || field.IsSubcommand || field.Short == "" {
			continue
		}

		flag := &optargs.Flag{
			Name:   field.Short,
			HasArg: field.ArgType,
		}
		shortOpts[field.Short[0]] = flag

		// If this field also has a long option, store the shared pointer
		// so buildLongOptMap can reuse it.
		field.CoreFlag = flag
	}

	return shortOpts
}

// buildLongOptMap builds a map from struct metadata long options to Flag pointers.
// For fields that have both short and long options, the same *optargs.Flag pointer
// from buildShortOptMap is reused so Handle is set once.
func (ci *CoreIntegration) buildLongOptMap() map[string]*optargs.Flag {
	longOpts := make(map[string]*optargs.Flag)

	for i := range ci.metadata.Fields {
		field := &ci.metadata.Fields[i]
		if field.Positional || field.IsSubcommand || field.Long == "" {
			continue
		}

		if field.CoreFlag != nil {
			// Reuse the shared pointer created by buildShortOptMap.
			longOpts[field.Long] = field.CoreFlag
		} else {
			longOpts[field.Long] = &optargs.Flag{
				Name:   field.Long,
				HasArg: field.ArgType,
			}
		}
	}

	return longOpts
}

// makeHandler returns a Handle callback that sets the struct field value when
// the option is parsed. Boolean flags with no argument are set to true, slice
// fields append the converted element, and all other types use
// TypeConverter.ConvertValue + TypeConverter.SetField via setFieldValue.
func (ci *CoreIntegration) makeHandler(field *FieldMetadata, destValue reflect.Value) func(string, string) error {
	return func(name, arg string) error {
		fieldValue := destValue.FieldByName(field.Name)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			return fmt.Errorf("cannot set field %s", field.Name)
		}
		return ci.setFieldValue(fieldValue, field, arg)
	}
}

// CreateParserWithHandlers builds an OptArgs parser with Handle callbacks
// wired to each flag. It builds short/long opt maps, sets Handle on each
// flag via makeHandler, creates the parser with case-insensitive commands,
// and prepares positional arg metadata. It does NOT register subcommands.
func (ci *CoreIntegration) CreateParserWithHandlers(args []string, destValue reflect.Value) (*optargs.Parser, error) {
	shortOpts := ci.buildShortOptMap()
	longOpts := ci.buildLongOptMap()

	// Set Handle callbacks on each flag. Because short and long flags for
	// the same field share the same *optargs.Flag pointer, iterating over
	// the metadata fields and setting Handle on whichever flag we find
	// ensures each flag gets its handler exactly once.
	for i := range ci.metadata.Fields {
		field := &ci.metadata.Fields[i]
		if field.Positional || field.IsSubcommand {
			continue
		}

		handler := ci.makeHandler(field, destValue)

		// Find the flag for this field — prefer the short opt pointer
		// (which is shared with long), otherwise use the long opt pointer.
		if field.Short != "" {
			if f, ok := shortOpts[field.Short[0]]; ok {
				f.Handle = handler
			}
		} else if field.Long != "" {
			if f, ok := longOpts[field.Long]; ok {
				f.Handle = handler
			}
		}
	}

	parser, err := optargs.NewParserWithCaseInsensitiveCommands(shortOpts, longOpts, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create OptArgs parser: %w", err)
	}

	ci.buildPositionalArgs()

	return parser, nil
}

// findSubcommandField finds the struct field for a subcommand by name
// (case-insensitive). It returns the field's reflect.Value, the subcommand's
// StructMetadata, and an error if the subcommand is not found.
func (ci *CoreIntegration) findSubcommandField(destValue reflect.Value, name string) (reflect.Value, *StructMetadata, error) {
	// Find the matching subcommand metadata (case-insensitive).
	var matchedName string
	var matchedMeta *StructMetadata
	for cmdName, meta := range ci.metadata.Subcommands {
		if strings.EqualFold(cmdName, name) {
			matchedName = cmdName
			matchedMeta = meta
			break
		}
	}
	if matchedMeta == nil {
		return reflect.Value{}, nil, fmt.Errorf("unknown subcommand: %s", name)
	}

	// Find the struct field whose subcommand name matches.
	tp := &TagParser{}
	destType := destValue.Type()
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		if !field.IsExported() {
			continue
		}
		fieldMeta, err := tp.ParseField(field)
		if err != nil || !fieldMeta.IsSubcommand {
			continue
		}
		if strings.EqualFold(fieldMeta.SubcommandName, matchedName) {
			return destValue.Field(i), matchedMeta, nil
		}
	}

	return reflect.Value{}, nil, fmt.Errorf("subcommand field not found for %s", name)
}

// RegisterSubcommands iterates ci.metadata.Subcommands, creates a child
// CoreIntegration for each, calls CreateParserWithHandlers on the child,
// registers via coreParser.AddCmd, and recursively registers nested
// subcommands.
func (ci *CoreIntegration) RegisterSubcommands(coreParser *optargs.Parser, destValue reflect.Value) error {
	for name, subMeta := range ci.metadata.Subcommands {
		fieldValue, _, err := ci.findSubcommandField(destValue, name)
		if err != nil {
			return fmt.Errorf("failed to find subcommand field for %s: %w", name, err)
		}

		// If the field is a pointer, allocate and dereference so we can
		// set fields on the underlying struct.
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			fieldValue = fieldValue.Elem()
		}

		child := &CoreIntegration{
			metadata:  subMeta,
			shortOpts: make(map[byte]*optargs.Flag),
			longOpts:  make(map[string]*optargs.Flag),
		}

		childParser, err := child.CreateParserWithHandlers([]string{}, fieldValue)
		if err != nil {
			return fmt.Errorf("failed to create parser for subcommand %s: %w", name, err)
		}

		coreParser.AddCmd(name, childParser)

		// Recursively register nested subcommands.
		if err := child.RegisterSubcommands(childParser, fieldValue); err != nil {
			return fmt.Errorf("failed to register nested subcommands for %s: %w", name, err)
		}
	}
	return nil
}

// PostParse executes the complete post-parse sequence: positional argument
// processing, environment variable resolution, default value application,
// and required field validation.
func (ci *CoreIntegration) PostParse(coreParser *optargs.Parser, destValue reflect.Value) error {
	if err := ci.processPositionalArgs(coreParser, destValue); err != nil {
		return err
	}
	if err := ci.processEnvironmentVariables(destValue); err != nil {
		return err
	}
	if err := ci.setDefaultValues(destValue); err != nil {
		return err
	}
	tc := &TypeConverter{}
	return tc.ValidateRequired(destValue.Addr().Interface(), ci.metadata)
}

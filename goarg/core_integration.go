package goarg

import (
	"fmt"
	"os"
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

// buildPositionalArgs builds the list of positional arguments
func (ci *CoreIntegration) buildPositionalArgs() {
	ci.positionals = make([]PositionalArg, 0, len(ci.metadata.Positionals))

	for i := range ci.metadata.Positionals {
		field := &ci.metadata.Positionals[i]
		ci.positionals = append(ci.positionals, PositionalArg{
			Field:    field,
			Required: field.Required,
			Multiple: field.Type.Kind() == reflect.Slice,
		})
	}
}

// setFieldValue sets a field value based on the parsed argument using
// optargs.Convert for all type conversion.
func (ci *CoreIntegration) setFieldValue(fieldValue reflect.Value, field *FieldMetadata, arg string) error {
	// Handle boolean flags specially - they are set to true when present without argument
	if field.Type.Kind() == reflect.Bool && arg == "" {
		fieldValue.SetBool(true)
		return nil
	}

	// Handle slice types - append to existing values
	if field.Type.Kind() == reflect.Slice {
		elemType := field.Type.Elem()
		converted, err := optargs.Convert(arg, elemType)
		if err != nil {
			return fmt.Errorf("failed to convert slice element: %w", err)
		}
		fieldValue.Set(reflect.Append(fieldValue, reflect.ValueOf(converted)))
		return nil
	}

	// For all other types, delegate to core
	converted, err := optargs.Convert(arg, field.Type)
	if err != nil {
		return fmt.Errorf("failed to convert value '%s' for field %s: %w", arg, field.Name, err)
	}

	fieldValue.Set(reflect.ValueOf(converted))
	return nil
}

// setScalarValue converts a string to the target type using optargs.Convert
// and sets the field value.
func (ci *CoreIntegration) setScalarValue(fieldValue reflect.Value, fieldType reflect.Type, arg string) error {
	converted, err := optargs.Convert(arg, fieldType)
	if err != nil {
		return fmt.Errorf("failed to convert scalar value: %w", err)
	}

	fieldValue.Set(reflect.ValueOf(converted))
	return nil
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
			elemType := field.Type.Elem()
			slice := reflect.MakeSlice(field.Type, 0, len(remainingArgs)-argIndex)

			for argIndex < len(remainingArgs) {
				converted, err := optargs.Convert(remainingArgs[argIndex], elemType)
				if err != nil {
					return fmt.Errorf("failed to set positional argument %d: %w", argIndex, err)
				}
				slice = reflect.Append(slice, reflect.ValueOf(converted))
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

		envValue, exists := os.LookupEnv(field.Env)
		if !exists {
			continue
		}

		if err := ci.setScalarValue(fieldValue, field.Type, envValue); err != nil {
			return fmt.Errorf("failed to set environment variable %s for field %s: %w", field.Env, field.Name, err)
		}
	}

	return nil
}

// setDefaultValues sets default values for unset fields using optargs.Convert
// and optargs.ConvertSlice for type conversion.
func (ci *CoreIntegration) setDefaultValues(destValue reflect.Value) error {
	for _, field := range ci.metadata.Fields {
		fieldValue := destValue.FieldByName(field.Name)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			continue
		}

		// Only set default if field is not already set
		if ci.isFieldSet(fieldValue, field.Type) {
			continue
		}

		// Get default value string from struct tag
		defaultTag, exists := reflect.StructTag(field.Tag).Lookup("default")
		if !exists {
			continue
		}

		if field.Type.Kind() == reflect.Slice {
			converted, err := optargs.ConvertSlice(defaultTag, field.Type)
			if err != nil {
				return fmt.Errorf("failed to set default value for field %s: %w", field.Name, err)
			}
			fieldValue.Set(reflect.ValueOf(converted))
		} else {
			converted, err := optargs.Convert(defaultTag, field.Type)
			if err != nil {
				return fmt.Errorf("failed to set default value for field %s: %w", field.Name, err)
			}
			fieldValue.Set(reflect.ValueOf(converted))
		}
	}

	return nil
}

// isFieldSet checks if a field has been set (not zero value)
func (ci *CoreIntegration) isFieldSet(fieldValue reflect.Value, fieldType reflect.Type) bool {
	return !isZeroValue(fieldValue)
}

// buildShortOptMap builds a map from struct metadata short options to Flag pointers.
// Each flag's HasArg is set based on the field's ArgType. Short and long flags for
// the same field share the same *optargs.Flag pointer so Handle is set once.
func (ci *CoreIntegration) buildShortOptMap() map[byte]*optargs.Flag {
	shortOpts := make(map[byte]*optargs.Flag)

	for i := range ci.metadata.Options {
		field := &ci.metadata.Options[i]
		if field.Short == "" {
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

	for i := range ci.metadata.Options {
		field := &ci.metadata.Options[i]
		if field.Long == "" {
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
// fields append the converted element, and all other types delegate to
// optargs.Convert via setFieldValue.
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
	for i := range ci.metadata.Options {
		field := &ci.metadata.Options[i]

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
	// Try direct lookup first via the pre-built field name index.
	if fieldName, ok := ci.metadata.SubcommandFields[name]; ok {
		subMeta := ci.metadata.Subcommands[name]
		if subMeta == nil {
			return reflect.Value{}, nil, fmt.Errorf("subcommand metadata not found for %s", name)
		}
		fv := destValue.FieldByName(fieldName)
		if !fv.IsValid() {
			return reflect.Value{}, nil, fmt.Errorf("subcommand field not found for %s", name)
		}
		return fv, subMeta, nil
	}

	// Fall back to case-insensitive scan of the index.
	for cmdName, fieldName := range ci.metadata.SubcommandFields {
		if strings.EqualFold(cmdName, name) {
			subMeta := ci.metadata.Subcommands[cmdName]
			if subMeta == nil {
				return reflect.Value{}, nil, fmt.Errorf("subcommand metadata not found for %s", cmdName)
			}
			fv := destValue.FieldByName(fieldName)
			if !fv.IsValid() {
				return reflect.Value{}, nil, fmt.Errorf("subcommand field not found for %s", cmdName)
			}
			return fv, subMeta, nil
		}
	}

	return reflect.Value{}, nil, fmt.Errorf("unknown subcommand: %s", name)
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
	return validateRequired(destValue.Addr().Interface(), ci.metadata)
}

// validateRequired validates that all required fields have been set.
func validateRequired(dest interface{}, metadata *StructMetadata) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	destElem := destValue.Elem()
	if destElem.Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a pointer to a struct")
	}

	for _, field := range metadata.Fields {
		if !field.Required {
			continue
		}

		fieldValue := destElem.FieldByName(field.Name)
		if !fieldValue.IsValid() {
			continue
		}

		if isZeroValue(fieldValue) {
			if field.Long != "" {
				return fmt.Errorf("--%s is required", field.Long)
			} else if field.Short != "" {
				return fmt.Errorf("-%s is required", field.Short)
			}
			return fmt.Errorf("%s is required", field.Name)
		}
	}

	return nil
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Chan:
		return v.IsNil() || v.Len() == 0
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

package goarg

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/major0/optargs"
)

// CoreIntegration handles direct translation to OptArgs Core.
type CoreIntegration struct {
	metadata    *StructMetadata
	config      Config
	positionals []PositionalArg
	setFields   map[int]bool // tracks field indices explicitly set during parsing
	flagBuilder *FlagBuilder
}

// PositionalArg represents a positional argument.
type PositionalArg struct {
	Field    *FieldMetadata
	Required bool
	Multiple bool
}

// buildPositionalArgs builds the list of positional arguments.
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

// processPositionalArgs processes positional arguments from remaining args.
func (ci *CoreIntegration) processPositionalArgs(parser *optargs.Parser, destValue reflect.Value) error {
	remainingArgs := parser.Args
	argIndex := 0

	for _, positional := range ci.positionals {
		field := positional.Field
		fieldValue := fieldByMeta(destValue, field)

		if !fieldValue.CanSet() {
			return fmt.Errorf("cannot set positional field %s", field.Name)
		}

		tv, err := typedValueForField(fieldValue, field)
		if err != nil {
			return fmt.Errorf("positional field %s: %w", field.Name, err)
		}

		if positional.Multiple { //nolint:nestif // multiple-positional setup requires conditional slice init + flag registration
			// Initialize to empty slice (not nil) for consistency with
			// the old reflect.MakeSlice behavior.
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.MakeSlice(field.Type, 0, 0))
			}
			for argIndex < len(remainingArgs) {
				if err := tv.Set(remainingArgs[argIndex]); err != nil {
					return fmt.Errorf("failed to set positional argument %d: %w", argIndex, err)
				}
				argIndex++
			}
		} else {
			if argIndex >= len(remainingArgs) {
				if positional.Required {
					return fmt.Errorf("missing required positional argument: %s", field.Name)
				}
				continue
			}

			if err := tv.Set(remainingArgs[argIndex]); err != nil {
				return fmt.Errorf("failed to set positional argument %s: %w", field.Name, err)
			}
			argIndex++
		}
	}

	return nil
}

// processEnvironmentVariables processes environment variable fallbacks.
func (ci *CoreIntegration) processEnvironmentVariables(destValue reflect.Value) error {
	for i := range ci.metadata.Fields {
		field := &ci.metadata.Fields[i]
		if field.Env == "" {
			continue
		}

		fieldValue := fieldByMeta(destValue, field)
		if !fieldValue.CanSet() {
			continue
		}

		if ci.isFieldSet(fieldValue) {
			continue
		}

		envName := field.Env
		if ci.config.EnvPrefix != "" {
			envName = ci.config.EnvPrefix + envName
		}

		envValue, exists := os.LookupEnv(envName)
		if !exists {
			continue
		}

		tv, err := typedValueForField(fieldValue, field)
		if err != nil {
			return fmt.Errorf("env var %s for field %s: %w", field.Env, field.Name, err)
		}
		if err := tv.Set(envValue); err != nil {
			return fmt.Errorf("failed to set environment variable %s for field %s: %w", field.Env, field.Name, err)
		}
	}

	return nil
}

// setDefaultValues sets default values for unset fields via TypedValue.Set().
// Uses pre-parsed HasDefault and DefaultTag from struct metadata.
func (ci *CoreIntegration) setDefaultValues(destValue reflect.Value) error {
	for i := range ci.metadata.Fields {
		field := &ci.metadata.Fields[i]
		if !field.HasDefault {
			continue
		}

		fieldValue := fieldByMeta(destValue, field)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			continue
		}

		// Skip fields explicitly set during parsing (including negatable zero-clear)
		if ci.setFields[field.FieldIndex] {
			continue
		}

		if ci.isFieldSet(fieldValue) {
			continue
		}

		tv, err := typedValueForField(fieldValue, field)
		if err != nil {
			return fmt.Errorf("default for field %s: %w", field.Name, err)
		}
		if err := tv.Set(field.DefaultTag); err != nil {
			return fmt.Errorf("failed to set default value for field %s: %w", field.Name, err)
		}
	}

	return nil
}

// isFieldSet checks if a field has been set (not zero value).
func (ci *CoreIntegration) isFieldSet(fieldValue reflect.Value) bool {
	return !isZeroValue(fieldValue)
}

// fieldByMeta returns the reflect.Value for a field using the cached index
// when available (FieldIndex >= 0), falling back to FieldByName for fields
// inherited from embedded structs (FieldIndex == -1).
func fieldByMeta(destValue reflect.Value, field *FieldMetadata) reflect.Value {
	if field.FieldIndex >= 0 {
		return destValue.Field(field.FieldIndex)
	}
	return destValue.FieldByName(field.Name)
}

// formatDefault returns the display string for a field's default value.
func formatDefault(field *FieldMetadata) string {
	if field.Default == nil {
		return ""
	}
	return fmt.Sprintf("%v", field.Default)
}

// CreateParserWithHandlers builds an OptArgs parser with Handle callbacks
// wired to each flag in a single pass. It creates the parser with
// case-insensitive commands and prepares positional arg metadata.
// It does NOT register subcommands.
func (ci *CoreIntegration) CreateParserWithHandlers(args []string, destValue reflect.Value) (*optargs.Parser, error) {
	ci.flagBuilder = &FlagBuilder{metadata: ci.metadata, config: ci.config}
	shortOpts, longOpts, err := ci.flagBuilder.Build(destValue)
	if err != nil {
		return nil, fmt.Errorf("failed to build flags: %w", err)
	}
	ci.setFields = ci.flagBuilder.SetFields()

	// Register builtin -h/--help flag (returns ErrHelp when parsed).
	helpFlag := &optargs.Flag{
		Name:   "h",
		HasArg: optargs.NoArgument,
		Help:   "display this help and exit",
		Handle: func(_, _ string) error { return ErrHelp },
	}
	helpLong := &optargs.Flag{
		Name:   "help",
		HasArg: optargs.NoArgument,
		Help:   "display this help and exit",
		Peer:   helpFlag,
		Handle: func(_, _ string) error { return ErrHelp },
	}
	helpFlag.Peer = helpLong
	if shortOpts['h'] == nil {
		shortOpts['h'] = helpFlag
	}
	if longOpts["help"] == nil {
		longOpts["help"] = helpLong
	}

	// Register builtin --version flag if version is configured.
	if ci.config.Version != "" {
		if longOpts["version"] == nil {
			longOpts["version"] = &optargs.Flag{
				Name:   "version",
				HasArg: optargs.NoArgument,
				Help:   "display version and exit",
				Handle: func(_, _ string) error { return ErrVersion },
			}
		}
	}

	parser, err := optargs.NewParserWithCaseInsensitiveCommands(shortOpts, longOpts, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create OptArgs parser: %w", err)
	}

	if ci.config.StrictSubcommands {
		parser.SetStrictSubcommands(true)
	}

	ci.buildPositionalArgs()

	return parser, nil
}

// findSubcommandField finds the struct field for a subcommand by name
// (case-insensitive). It returns the field's reflect.Value, the subcommand's
// StructMetadata, and an error if the subcommand is not found.
func (ci *CoreIntegration) findSubcommandField(destValue reflect.Value, name string) (reflect.Value, *StructMetadata, error) {
	// Try direct lookup first via the pre-built field index.
	if idx, ok := ci.metadata.SubcommandFieldIdx[name]; ok {
		subMeta := ci.metadata.Subcommands[name]
		if subMeta == nil {
			return reflect.Value{}, nil, fmt.Errorf("subcommand metadata not found for %s", name)
		}
		fv := destValue.Field(idx)
		if !fv.IsValid() {
			return reflect.Value{}, nil, fmt.Errorf("subcommand field not found for %s", name)
		}
		return fv, subMeta, nil
	}

	// Fall back to case-insensitive scan of the index.
	for cmdName, idx := range ci.metadata.SubcommandFieldIdx {
		if strings.EqualFold(cmdName, name) {
			subMeta := ci.metadata.Subcommands[cmdName]
			if subMeta == nil {
				return reflect.Value{}, nil, fmt.Errorf("subcommand metadata not found for %s", cmdName)
			}
			fv := destValue.Field(idx)
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
			metadata: subMeta,
			config:   ci.config,
		}

		childParser, err := child.CreateParserWithHandlers([]string{}, fieldValue)
		if err != nil {
			return fmt.Errorf("failed to create parser for subcommand %s: %w", name, err)
		}

		coreParser.AddCmd(name, childParser)

		// Set description from subcommand help metadata.
		if help, ok := ci.metadata.SubcommandHelp[name]; ok {
			childParser.Description = help
		}

		// Recursively register nested subcommands.
		if err := child.RegisterSubcommands(childParser, fieldValue); err != nil {
			return fmt.Errorf("failed to register nested subcommands for %s: %w", name, err)
		}
	}
	return nil
}

// dispatchSubcommand iterates the child parser's Options(), runs PostParse
// on the subcommand struct, and recursively walks ActiveCommand() for
// nested subcommands.
func (ci *CoreIntegration) dispatchSubcommand(childParser *optargs.Parser, invokedName string, destValue reflect.Value, p *Parser) error {
	fieldValue, subMeta, err := ci.findSubcommandField(destValue, invokedName)
	if err != nil {
		return p.translateError(err, invokedName)
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
		metadata:  subMeta,
		config:    ci.config,
		setFields: make(map[int]bool),
	}
	childCI.buildPositionalArgs()
	if err := childCI.PostParse(childParser, subDestValue); err != nil {
		return p.translateError(err, "")
	}

	// Recursively dispatch nested subcommands via ActiveCommand()
	nestedName, nestedParser := childParser.ActiveCommand()
	if nestedName != "" && nestedParser != nil {
		return childCI.dispatchSubcommand(nestedParser, nestedName, subDestValue, p)
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
	if !ci.config.IgnoreEnv {
		if err := ci.processEnvironmentVariables(destValue); err != nil {
			return err
		}
	}
	if !ci.config.IgnoreDefault {
		if err := ci.setDefaultValues(destValue); err != nil {
			return err
		}
	}
	return validateRequired(destValue.Addr().Interface(), ci.metadata)
}

// validateRequired validates that all required fields have been set.
func validateRequired(dest any, metadata *StructMetadata) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return errors.New("destination must be a pointer")
	}

	destElem := destValue.Elem()
	if destElem.Kind() != reflect.Struct {
		return errors.New("destination must be a pointer to a struct")
	}

	for i := range metadata.Fields {
		field := &metadata.Fields[i]
		if !field.Required {
			continue
		}

		fieldValue := fieldByMeta(destElem, field)
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
		for i := range v.Len() {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := range v.NumField() {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

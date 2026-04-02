package goarg

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/major0/optargs"
)

// CoreIntegration handles direct translation to OptArgs Core
type CoreIntegration struct {
	metadata    *StructMetadata
	config      Config
	positionals []PositionalArg
	setFields   map[int]bool // tracks field indices explicitly set during parsing
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

// Cached reflect.Type for time.Duration and TextUnmarshaler interface.
var (
	durationType         = reflect.TypeOf(time.Duration(0))
	durationSliceType    = reflect.TypeOf([]time.Duration{})
	textUnmarshalerIface = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// typedValueForField creates an optargs.TypedValue backed by a pointer to
// the struct field's storage. Type dispatch happens once here at setup time;
// the returned TypedValue handles all subsequent Set() calls.
func typedValueForField(fieldValue reflect.Value, field *FieldMetadata) (optargs.TypedValue, error) {
	ft := field.Type

	// Pointer types: wrap in a ptrValue that allocates on first Set().
	if ft.Kind() == reflect.Ptr {
		return &ptrValue{fieldValue: fieldValue, elemType: ft.Elem(), field: field}, nil
	}

	// TextUnmarshaler takes priority over kind-based dispatch — user-defined
	// types (e.g., net.IP which is []byte) must be handled here before the
	// slice/scalar switch below.
	ptrType := reflect.PointerTo(ft)
	if ptrType.Implements(textUnmarshalerIface) {
		dest := fieldValue.Addr().Interface().(encoding.TextUnmarshaler)
		var val encoding.TextMarshaler
		if m, ok := dest.(encoding.TextMarshaler); ok {
			val = m
		}
		return optargs.NewTextValue(val, dest), nil
	}

	// time.Duration must be checked before int64 (same Kind).
	if ft == durationType {
		p := fieldValue.Addr().Interface().(*time.Duration)
		return optargs.NewDurationValue(*p, p), nil
	}

	// Scalar types.
	switch ft.Kind() {
	case reflect.String:
		p := fieldValue.Addr().Interface().(*string)
		return optargs.NewStringValue(*p, p), nil
	case reflect.Bool:
		p := fieldValue.Addr().Interface().(*bool)
		return optargs.NewBoolValue(*p, p), nil
	case reflect.Int:
		p := fieldValue.Addr().Interface().(*int)
		return optargs.NewIntValue(*p, p), nil
	case reflect.Int8:
		p := fieldValue.Addr().Interface().(*int8)
		return optargs.NewInt8Value(*p, p), nil
	case reflect.Int16:
		p := fieldValue.Addr().Interface().(*int16)
		return optargs.NewInt16Value(*p, p), nil
	case reflect.Int32:
		p := fieldValue.Addr().Interface().(*int32)
		return optargs.NewInt32Value(*p, p), nil
	case reflect.Int64:
		p := fieldValue.Addr().Interface().(*int64)
		return optargs.NewInt64Value(*p, p), nil
	case reflect.Uint:
		p := fieldValue.Addr().Interface().(*uint)
		return optargs.NewUintValue(*p, p), nil
	case reflect.Uint8:
		p := fieldValue.Addr().Interface().(*uint8)
		return optargs.NewUint8Value(*p, p), nil
	case reflect.Uint16:
		p := fieldValue.Addr().Interface().(*uint16)
		return optargs.NewUint16Value(*p, p), nil
	case reflect.Uint32:
		p := fieldValue.Addr().Interface().(*uint32)
		return optargs.NewUint32Value(*p, p), nil
	case reflect.Uint64:
		p := fieldValue.Addr().Interface().(*uint64)
		return optargs.NewUint64Value(*p, p), nil
	case reflect.Float32:
		p := fieldValue.Addr().Interface().(*float32)
		return optargs.NewFloat32Value(*p, p), nil
	case reflect.Float64:
		p := fieldValue.Addr().Interface().(*float64)
		return optargs.NewFloat64Value(*p, p), nil

	case reflect.Slice:
		return typedValueForSlice(fieldValue, ft)

	case reflect.Map:
		return typedValueForMap(fieldValue, ft)
	}

	return nil, fmt.Errorf("unsupported type %s for field %s", ft, field.Name)
}

// typedValueForSlice handles slice field types.
func typedValueForSlice(fieldValue reflect.Value, ft reflect.Type) (optargs.TypedValue, error) {
	// []time.Duration must be checked before []int64.
	if ft == durationSliceType {
		p := fieldValue.Addr().Interface().(*[]time.Duration)
		return optargs.NewDurationSliceValue(*p, p), nil
	}

	switch ft.Elem().Kind() {
	case reflect.String:
		p := fieldValue.Addr().Interface().(*[]string)
		return optargs.NewStringSliceValue(*p, p), nil
	case reflect.Bool:
		p := fieldValue.Addr().Interface().(*[]bool)
		return optargs.NewBoolSliceValue(*p, p), nil
	case reflect.Int:
		p := fieldValue.Addr().Interface().(*[]int)
		return optargs.NewIntSliceValue(*p, p), nil
	case reflect.Int32:
		p := fieldValue.Addr().Interface().(*[]int32)
		return optargs.NewInt32SliceValue(*p, p), nil
	case reflect.Int64:
		p := fieldValue.Addr().Interface().(*[]int64)
		return optargs.NewInt64SliceValue(*p, p), nil
	case reflect.Uint:
		p := fieldValue.Addr().Interface().(*[]uint)
		return optargs.NewUintSliceValue(*p, p), nil
	case reflect.Float32:
		p := fieldValue.Addr().Interface().(*[]float32)
		return optargs.NewFloat32SliceValue(*p, p), nil
	case reflect.Float64:
		p := fieldValue.Addr().Interface().(*[]float64)
		return optargs.NewFloat64SliceValue(*p, p), nil
	}

	return nil, fmt.Errorf("unsupported slice element type: %s", ft.Elem())
}

// typedValueForMap handles map field types.
func typedValueForMap(fieldValue reflect.Value, ft reflect.Type) (optargs.TypedValue, error) {
	if ft.Key().Kind() != reflect.String {
		return nil, fmt.Errorf("unsupported map key type: %s", ft.Key())
	}

	switch ft.Elem().Kind() {
	case reflect.String:
		p := fieldValue.Addr().Interface().(*map[string]string)
		return optargs.NewStringToStringValue(*p, p), nil
	case reflect.Int:
		p := fieldValue.Addr().Interface().(*map[string]int)
		return optargs.NewStringToIntValue(*p, p), nil
	case reflect.Int64:
		p := fieldValue.Addr().Interface().(*map[string]int64)
		return optargs.NewStringToInt64Value(*p, p), nil
	}

	return nil, fmt.Errorf("unsupported map value type: %s", ft.Elem())
}

// ptrValue wraps a pointer field. Allocates the pointed-to value on first
// Set() so that unset pointer fields remain nil.
type ptrValue struct {
	fieldValue reflect.Value
	elemType   reflect.Type
	field      *FieldMetadata
	inner      optargs.TypedValue // created lazily on first Set()
}

func (v *ptrValue) Set(s string) error {
	if v.inner == nil {
		// Allocate the pointer and create the inner TypedValue.
		v.fieldValue.Set(reflect.New(v.elemType))
		elemField := &FieldMetadata{
			Name:       v.field.Name,
			FieldIndex: v.field.FieldIndex,
			Type:       v.elemType,
		}
		var err error
		v.inner, err = typedValueForField(v.fieldValue.Elem(), elemField)
		if err != nil {
			return err
		}
	}
	return v.inner.Set(s)
}

func (v *ptrValue) String() string {
	if v.fieldValue.IsNil() {
		return ""
	}
	if v.inner != nil {
		return v.inner.String()
	}
	return ""
}

func (v *ptrValue) Type() string {
	return v.elemType.String()
}

func (v *ptrValue) IsBoolFlag() bool {
	return v.elemType.Kind() == reflect.Bool
}

// processPositionalArgs processes positional arguments from remaining args
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

		if positional.Multiple {
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

// processEnvironmentVariables processes environment variable fallbacks
func (ci *CoreIntegration) processEnvironmentVariables(destValue reflect.Value) error {
	for _, field := range ci.metadata.Fields {
		if field.Env == "" {
			continue
		}

		fieldValue := fieldByMeta(destValue, &field)
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

		tv, err := typedValueForField(fieldValue, &field)
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
	for _, field := range ci.metadata.Fields {
		if !field.HasDefault {
			continue
		}

		fieldValue := fieldByMeta(destValue, &field)
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

		tv, err := typedValueForField(fieldValue, &field)
		if err != nil {
			return fmt.Errorf("default for field %s: %w", field.Name, err)
		}
		if err := tv.Set(field.DefaultTag); err != nil {
			return fmt.Errorf("failed to set default value for field %s: %w", field.Name, err)
		}
	}

	return nil
}

// isFieldSet checks if a field has been set (not zero value)
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

// makeHandler returns a Handle callback that sets the struct field value when
// the option is parsed. Creates a TypedValue at setup time and captures it
// in the closure — the handler just calls Set().
func (ci *CoreIntegration) makeHandler(field *FieldMetadata, destValue reflect.Value) (func(string, string) error, error) {
	fieldValue := fieldByMeta(destValue, field)
	if !fieldValue.CanSet() {
		return nil, fmt.Errorf("cannot set field %s", field.Name)
	}
	tv, err := typedValueForField(fieldValue, field)
	if err != nil {
		return nil, err
	}
	idx := field.FieldIndex
	return func(name, arg string) error {
		if arg == "" {
			if _, ok := tv.(optargs.BoolValuer); ok {
				if err := tv.Set("true"); err != nil {
					return err
				}
				ci.setFields[idx] = true
				return nil
			}
		}
		if err := tv.Set(arg); err != nil {
			return err
		}
		ci.setFields[idx] = true
		return nil
	}, nil
}

// makeBoolPrefixHandler returns a handler for a prefixed boolean option
// (e.g. --enable-shared). The val argument controls whether the field is
// set to true or false.
func (ci *CoreIntegration) makeBoolPrefixHandler(field *FieldMetadata, destValue reflect.Value, val bool) func(string, string) error {
	return func(_, _ string) error {
		fv := fieldByMeta(destValue, field)
		fv.SetBool(val)
		ci.setFields[field.FieldIndex] = true
		return nil
	}
}

// makeNegatableHandler returns a handler for --no-<name> on a non-boolean field.
// Clears the field to its type's zero value via reflect.Zero.
func (ci *CoreIntegration) makeNegatableHandler(field *FieldMetadata, destValue reflect.Value) func(string, string) error {
	return func(_, _ string) error {
		fv := fieldByMeta(destValue, field)
		fv.Set(reflect.Zero(fv.Type()))
		ci.setFields[field.FieldIndex] = true
		return nil
	}
}

// buildFlags builds short and long option maps in a single pass over
// metadata.Options. For fields with both short and long names, a single
// shared *Flag is created. Handlers and Peer links are wired inline,
// eliminating the separate buildShortOptMap / buildLongOptMap /
// Peer-linking / Handle-setting passes.
func (ci *CoreIntegration) buildFlags(destValue reflect.Value) (map[byte]*optargs.Flag, map[string]*optargs.Flag, error) {
	nOpts := len(ci.metadata.Options)
	shortOpts := make(map[byte]*optargs.Flag, nOpts)
	longOpts := make(map[string]*optargs.Flag, nOpts)

	for i := range ci.metadata.Options {
		field := &ci.metadata.Options[i]
		handler, err := ci.makeHandler(field, destValue)
		if err != nil {
			return nil, nil, fmt.Errorf("field %s: %w", field.Name, err)
		}
		argName := strings.ToUpper(field.Name)
		defVal := formatDefault(field)

		hasShort := field.Short != ""
		hasLong := field.Long != ""

		if hasShort && hasLong {
			// Both short and long — one shared Flag in both maps.
			flag := &optargs.Flag{
				Name:         field.Short,
				HasArg:       field.ArgType,
				Help:         field.Help,
				ArgName:      argName,
				DefaultValue: defVal,
				Handle:       handler,
			}
			shortOpts[field.Short[0]] = flag
			longOpts[field.Long] = flag
		} else if hasShort {
			shortOpts[field.Short[0]] = &optargs.Flag{
				Name:         field.Short,
				HasArg:       field.ArgType,
				Help:         field.Help,
				ArgName:      argName,
				DefaultValue: defVal,
				Handle:       handler,
			}
		} else if hasLong {
			longOpts[field.Long] = &optargs.Flag{
				Name:         field.Long,
				HasArg:       field.ArgType,
				Help:         field.Help,
				ArgName:      argName,
				DefaultValue: defVal,
				Handle:       handler,
			}
		}

		// Register prefix pair options for boolean fields (always NoArgument)
		if hasLong {
			for _, pp := range field.Prefixes {
				trueName := pp.True + "-" + field.Long
				falseName := pp.False + "-" + field.Long
				longOpts[trueName] = &optargs.Flag{
					Name:   trueName,
					HasArg: optargs.NoArgument,
					Handle: ci.makeBoolPrefixHandler(field, destValue, true),
				}
				longOpts[falseName] = &optargs.Flag{
					Name:   falseName,
					HasArg: optargs.NoArgument,
					Handle: ci.makeBoolPrefixHandler(field, destValue, false),
				}
			}

			// Register --no-<name> for negatable non-boolean fields
			if field.Negatable && field.Type.Kind() != reflect.Bool {
				negName := "no-" + field.Long
				longOpts[negName] = &optargs.Flag{
					Name:   negName,
					HasArg: optargs.NoArgument,
					Handle: ci.makeNegatableHandler(field, destValue),
				}
			}
		}
	}

	return shortOpts, longOpts, nil
}

// CreateParserWithHandlers builds an OptArgs parser with Handle callbacks
// wired to each flag in a single pass. It creates the parser with
// case-insensitive commands and prepares positional arg metadata.
// It does NOT register subcommands.
func (ci *CoreIntegration) CreateParserWithHandlers(args []string, destValue reflect.Value) (*optargs.Parser, error) {
	ci.setFields = make(map[int]bool)
	shortOpts, longOpts, err := ci.buildFlags(destValue)
	if err != nil {
		return nil, fmt.Errorf("failed to build flags: %w", err)
	}

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

		fieldValue := fieldByMeta(destElem, &field)
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

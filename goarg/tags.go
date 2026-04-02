package goarg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/major0/optargs"
)

// StructMetadata represents parsed struct information
type StructMetadata struct {
	Fields             []FieldMetadata
	Options            []FieldMetadata // non-positional, non-subcommand, has CLI flag
	Positionals        []FieldMetadata // positional fields, in declaration order
	EnvOnly            []FieldMetadata // env-only fields (no CLI flag)
	Subcommands        map[string]*StructMetadata
	SubcommandHelp     map[string]string // Maps subcommand name to help text
	SubcommandFields   map[string]string // Maps subcommand name to struct field name
	SubcommandFieldIdx map[string]int    // Maps subcommand name to struct field index
}

// PrefixPair represents a true/false prefix pair for a boolean field.
// Duplicated from pflags to avoid cross-module dependency.
type PrefixPair struct {
	True  string // e.g. "enable"
	False string // e.g. "disable"
}

// FieldMetadata represents a single struct field's CLI mapping
type FieldMetadata struct {
	Name       string
	FieldIndex int // struct field index for reflect.Value.Field(i) — avoids FieldByName
	Type       reflect.Type
	Tag        string
	Short      string
	Long       string
	Help       string
	Required   bool
	Positional bool
	Env        string
	Default    interface{}
	DefaultTag string // raw default tag string, pre-parsed
	HasDefault bool   // true when a `default:` tag is present (even if empty)

	// Subcommand support
	IsSubcommand   bool
	SubcommandName string

	// Prefix pairs and negatable support
	Prefixes  []PrefixPair // boolean prefix pairs from `prefix` struct tag
	Negatable bool         // non-boolean field supports --no-<name>

	// Direct OptArgs Core mapping
	CoreFlag *optargs.Flag
	ArgType  optargs.ArgType
}

// TagParser processes struct tags - identical behavior to alexflint/go-arg
type TagParser struct{}

// ParseStruct parses a struct and returns its metadata
func (tp *TagParser) ParseStruct(dest interface{}) (*StructMetadata, error) {
	if dest == nil {
		return nil, fmt.Errorf("destination cannot be nil")
	}

	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("destination must be a pointer to a struct, got %T", dest)
	}

	destElem := destValue.Elem()
	if destElem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("destination must be a pointer to a struct, got pointer to %s", destElem.Kind())
	}

	structType := destElem.Type()
	metadata := &StructMetadata{
		Fields:             []FieldMetadata{},
		Options:            []FieldMetadata{},
		Positionals:        []FieldMetadata{},
		Subcommands:        make(map[string]*StructMetadata),
		SubcommandHelp:     make(map[string]string),
		SubcommandFields:   make(map[string]string),
		SubcommandFieldIdx: make(map[string]int),
	}

	// Parse each field in the struct
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported non-embedded fields. Embedded structs may
		// contain exported fields even if the embedding field itself
		// is unexported — recurse into those.
		if !field.IsExported() && !field.Anonymous {
			continue
		}

		// Embedded (anonymous) struct: recurse into its fields.
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embedded := destElem.Field(i).Addr().Interface()
			subMeta, err := tp.ParseStruct(embedded)
			if err != nil {
				return nil, fmt.Errorf("failed to parse embedded struct %s: %w", field.Name, err)
			}
			// Mark embedded fields with FieldIndex = -1 so callers
			// fall back to FieldByName for these (their indices are
			// relative to the embedded struct, not the parent).
			for j := range subMeta.Fields {
				subMeta.Fields[j].FieldIndex = -1
			}
			for j := range subMeta.Options {
				subMeta.Options[j].FieldIndex = -1
			}
			for j := range subMeta.Positionals {
				subMeta.Positionals[j].FieldIndex = -1
			}
			for j := range subMeta.EnvOnly {
				subMeta.EnvOnly[j].FieldIndex = -1
			}
			metadata.Fields = append(metadata.Fields, subMeta.Fields...)
			metadata.Options = append(metadata.Options, subMeta.Options...)
			metadata.Positionals = append(metadata.Positionals, subMeta.Positionals...)
			metadata.EnvOnly = append(metadata.EnvOnly, subMeta.EnvOnly...)
			for k, v := range subMeta.Subcommands {
				metadata.Subcommands[k] = v
			}
			for k, v := range subMeta.SubcommandHelp {
				metadata.SubcommandHelp[k] = v
			}
			for k, v := range subMeta.SubcommandFields {
				metadata.SubcommandFields[k] = v
			}
			for k, v := range subMeta.SubcommandFieldIdx {
				metadata.SubcommandFieldIdx[k] = v
			}
			continue
		}

		// Skip unexported non-anonymous fields
		if !field.IsExported() {
			continue
		}

		fieldMetadata, err := tp.ParseField(field, i)
		if err != nil {
			return nil, fmt.Errorf("failed to parse field %s: %w", field.Name, err)
		}

		// Handle subcommands
		if fieldMetadata.IsSubcommand {
			subcommandName := fieldMetadata.SubcommandName
			if subcommandName == "" {
				subcommandName = strings.ToLower(field.Name)
			}

			// Record the struct field name for O(1) lookup later.
			metadata.SubcommandFields[subcommandName] = field.Name
			metadata.SubcommandFieldIdx[subcommandName] = i

			// Parse the subcommand struct for metadata only
			fieldValue := destElem.Field(i)
			if fieldValue.Kind() == reflect.Ptr {
				var subInstance interface{}
				wasNil := fieldValue.IsNil()

				if wasNil {
					// Create a temporary instance for parsing metadata only
					tempInstance := reflect.New(field.Type.Elem())
					subInstance = tempInstance.Interface()
				} else {
					subInstance = fieldValue.Interface()
				}

				subMetadata, err := tp.ParseStruct(subInstance)
				if err != nil {
					return nil, fmt.Errorf("failed to parse subcommand %s: %w", subcommandName, err)
				}
				metadata.Subcommands[subcommandName] = subMetadata

				// Store the help text for this subcommand
				metadata.SubcommandHelp[subcommandName] = fieldMetadata.Help

				// If the field was originally nil, keep it nil (don't persist the temp instance)
				// The subcommand will only be initialized when actually invoked
			}
		} else {
			metadata.Fields = append(metadata.Fields, *fieldMetadata)
			if fieldMetadata.Positional {
				metadata.Positionals = append(metadata.Positionals, *fieldMetadata)
			} else if fieldMetadata.Short == "" && fieldMetadata.Long == "" && fieldMetadata.Env != "" {
				metadata.EnvOnly = append(metadata.EnvOnly, *fieldMetadata)
			} else {
				metadata.Options = append(metadata.Options, *fieldMetadata)
			}
		}
	}

	return metadata, nil
}

// ParseField parses a single struct field and returns its metadata
func (tp *TagParser) ParseField(field reflect.StructField, fieldIndex int) (*FieldMetadata, error) {
	metadata := &FieldMetadata{
		Name:       field.Name,
		FieldIndex: fieldIndex,
		Type:       field.Type,
		Tag:        string(field.Tag),
	}

	// Parse the 'arg' tag
	argTag := field.Tag.Get("arg")
	if argTag != "" {
		if err := tp.parseArgTag(metadata, argTag); err != nil {
			return nil, fmt.Errorf("invalid arg tag for field %s: %w", field.Name, err)
		}
	}

	// Parse the 'help' tag
	metadata.Help = field.Tag.Get("help")

	// Parse the 'default' tag — use Lookup once to detect presence and value.
	if defaultTag, exists := field.Tag.Lookup("default"); exists {
		metadata.HasDefault = true
		metadata.DefaultTag = defaultTag
		defaultValue, err := tp.parseDefaultValue(defaultTag, field.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid default value for field %s: %w", field.Name, err)
		}
		metadata.Default = defaultValue
	}

	// Parse the 'env' tag — only if not already set from the arg tag
	if metadata.Env == "" {
		metadata.Env = field.Tag.Get("env")
	}

	// Parse the 'prefix' tag — boolean prefix pairs
	if prefixTag := field.Tag.Get("prefix"); prefixTag != "" {
		if field.Type.Kind() != reflect.Bool {
			return nil, fmt.Errorf("prefix tag on non-boolean field %q", field.Name)
		}
		for _, pair := range strings.Split(prefixTag, ";") {
			parts := strings.SplitN(pair, ",", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid prefix pair: %q (expected \"true,false\")", pair)
			}
			metadata.Prefixes = append(metadata.Prefixes, PrefixPair{
				True:  strings.TrimSpace(parts[0]),
				False: strings.TrimSpace(parts[1]),
			})
		}
	}

	// Parse the 'negatable' tag — silently ignored on boolean fields
	if _, exists := field.Tag.Lookup("negatable"); exists && field.Type.Kind() != reflect.Bool {
		metadata.Negatable = true
	}

	// Validate field metadata
	if err := tp.ValidateFieldMetadata(metadata); err != nil {
		return nil, fmt.Errorf("invalid field metadata for %s: %w", field.Name, err)
	}

	// Determine OptArgs Core mapping
	if err := tp.mapToOptArgsCore(metadata); err != nil {
		return nil, fmt.Errorf("failed to map field %s to OptArgs Core: %w", field.Name, err)
	}

	return metadata, nil
}

// parseArgTag parses the 'arg' struct tag and populates metadata
func (tp *TagParser) parseArgTag(metadata *FieldMetadata, argTag string) error {
	// Handle different arg tag formats:
	// 1. "-v,--verbose" - short and long options
	// 2. "--verbose" - long option only
	// 3. "-v" - short option only
	// 4. "positional" - positional argument
	// 5. "required" - required option
	// 6. "subcommand:name" - subcommand
	// 7. "subcommand" - subcommand with default name
	// 8. "env:VAR_NAME" - environment variable (can be combined)

	parts := strings.Split(argTag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "" {
			continue
		}

		// Handle special keywords
		switch {
		case part == "positional":
			metadata.Positional = true
		case part == "required":
			metadata.Required = true
		case part == "subcommand":
			metadata.IsSubcommand = true
			// Use field name as subcommand name if not specified
			metadata.SubcommandName = strings.ToLower(metadata.Name)
		case strings.HasPrefix(part, "subcommand:"):
			metadata.IsSubcommand = true
			metadata.SubcommandName = strings.TrimPrefix(part, "subcommand:")
		case strings.HasPrefix(part, "env:"):
			metadata.Env = strings.TrimPrefix(part, "env:")
		case part == "env":
			// Bare "env" — auto-derive env var name from field name in SCREAMING_SNAKE_CASE.
			metadata.Env = toScreamingSnake(metadata.Name)
		case part == "separate":
			// "separate" changes slice behavior from greedy multi-value to
			// one-value-per-flag. Our POSIX-based parser already uses this
			// semantics by default, so this is a no-op — accepted for
			// upstream compatibility.
			continue
		case strings.HasPrefix(part, "--"):
			// Long option
			metadata.Long = strings.TrimPrefix(part, "--")
		case strings.HasPrefix(part, "-") && len(part) == 2:
			// Short option (single character)
			metadata.Short = strings.TrimPrefix(part, "-")
		case strings.HasPrefix(part, "-"):
			// Invalid short option (more than one character)
			return fmt.Errorf("invalid short option format: %s (short options must be single characters)", part)
		default:
			// Unknown format
			return fmt.Errorf("unknown arg tag format: %s", part)
		}
	}

	return nil
}

// parseDefaultValue parses a default value string into the appropriate type
// using optargs.Convert and optargs.ConvertSlice.
func (tp *TagParser) parseDefaultValue(defaultStr string, fieldType reflect.Type) (interface{}, error) {
	if fieldType.Kind() == reflect.Slice {
		return optargs.ConvertSlice(defaultStr, fieldType)
	}
	return optargs.Convert(defaultStr, fieldType)
}

// mapToOptArgsCore maps field metadata to OptArgs Core structures
func (tp *TagParser) mapToOptArgsCore(metadata *FieldMetadata) error {
	if metadata.Positional || metadata.IsSubcommand {
		// Positional arguments and subcommands don't map to OptArgs Core flags
		return nil
	}

	// Determine argument type based on field type
	var argType optargs.ArgType
	switch metadata.Type.Kind() {
	case reflect.Bool:
		argType = optargs.NoArgument
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		argType = optargs.RequiredArgument
	case reflect.Slice:
		argType = optargs.RequiredArgument
	case reflect.Map:
		argType = optargs.RequiredArgument
	case reflect.Ptr:
		// For pointer types, check the underlying type
		elemType := metadata.Type.Elem()
		switch elemType.Kind() {
		case reflect.Bool:
			argType = optargs.NoArgument
		default:
			argType = optargs.OptionalArgument
		}
	default:
		// For custom types, assume required argument
		argType = optargs.RequiredArgument
	}

	metadata.ArgType = argType

	// Create OptArgs Core flag if we have option names
	if metadata.Long != "" || metadata.Short != "" {
		flagName := metadata.Long
		if flagName == "" {
			flagName = metadata.Short
		}

		metadata.CoreFlag = &optargs.Flag{
			Name:   flagName,
			HasArg: argType,
		}
	}

	return nil
}

// ValidateFieldMetadata validates that field metadata is consistent and complete
func (tp *TagParser) ValidateFieldMetadata(metadata *FieldMetadata) error {
	// Positional arguments cannot have short/long options
	if metadata.Positional && (metadata.Short != "" || metadata.Long != "") {
		return fmt.Errorf("positional argument cannot have option flags")
	}

	// Subcommands must be pointer to struct
	if metadata.IsSubcommand {
		if metadata.Type.Kind() != reflect.Ptr || metadata.Type.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("subcommand field must be pointer to struct")
		}
	}

	// Options should have at least one flag (short or long), unless
	// the field is env-only (has Env but no explicit flags).
	if !metadata.Positional && !metadata.IsSubcommand && metadata.Short == "" && metadata.Long == "" {
		if metadata.Env == "" {
			// Generate default long option from field name
			metadata.Long = strings.ToLower(metadata.Name)
		}
		// else: env-only field, no CLI flag generated
	}

	// Validate short option is single character
	if metadata.Short != "" && len(metadata.Short) != 1 {
		return fmt.Errorf("short option must be single character, got: %s", metadata.Short)
	}

	return nil
}

// toScreamingSnake converts a CamelCase or mixedCase name to SCREAMING_SNAKE_CASE.
// Examples: "Workers" → "WORKERS", "NumWorkers" → "NUM_WORKERS", "APIToken" → "API_TOKEN".
func toScreamingSnake(name string) string {
	var result []byte
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Insert underscore before uppercase if previous char was lowercase
			// or if next char is lowercase (handles "APIToken" → "API_TOKEN").
			prev := name[i-1]
			if prev >= 'a' && prev <= 'z' {
				result = append(result, '_')
			} else if i+1 < len(name) && name[i+1] >= 'a' && name[i+1] <= 'z' {
				result = append(result, '_')
			}
		}
		if r >= 'a' && r <= 'z' {
			result = append(result, byte(r-32)) // to uppercase
		} else {
			result = append(result, byte(r))
		}
	}
	return string(result)
}

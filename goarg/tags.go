package goarg

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/major0/optargs"
)

// StructMetadata represents parsed struct information
type StructMetadata struct {
	Fields      []FieldMetadata
	Subcommands map[string]*StructMetadata
	Program     string
	Description string
	Version     string
}

// FieldMetadata represents a single struct field's CLI mapping
type FieldMetadata struct {
	Name       string
	Type       reflect.Type
	Tag        string
	Short      string
	Long       string
	Help       string
	Required   bool
	Positional bool
	Env        string
	Default    interface{}

	// Subcommand support
	IsSubcommand   bool
	SubcommandName string

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
		Fields:      []FieldMetadata{},
		Subcommands: make(map[string]*StructMetadata),
	}

	// Parse each field in the struct
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldMetadata, err := tp.ParseField(field)
		if err != nil {
			return nil, fmt.Errorf("failed to parse field %s: %w", field.Name, err)
		}

		// Handle subcommands
		if fieldMetadata.IsSubcommand {
			subcommandName := fieldMetadata.SubcommandName
			if subcommandName == "" {
				subcommandName = strings.ToLower(field.Name)
			}

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

				// If the field was originally nil, keep it nil (don't persist the temp instance)
				// The subcommand will only be initialized when actually invoked
			}
		} else {
			metadata.Fields = append(metadata.Fields, *fieldMetadata)
		}
	}

	return metadata, nil
}

// ParseField parses a single struct field and returns its metadata
func (tp *TagParser) ParseField(field reflect.StructField) (*FieldMetadata, error) {
	metadata := &FieldMetadata{
		Name: field.Name,
		Type: field.Type,
		Tag:  string(field.Tag),
	}

	// Parse the 'arg' tag
	argTag := field.Tag.Get("arg")
	if argTag != "" {
		if err := tp.parseArgTag(metadata, argTag); err != nil {
			return nil, fmt.Errorf("invalid arg tag for field %s: %w", field.Name, err)
		}
	}

	// Parse the 'help' tag
	helpTag := field.Tag.Get("help")
	if helpTag != "" {
		metadata.Help = helpTag
	}

	// Parse the 'default' tag
	defaultTag := field.Tag.Get("default")
	if defaultTag != "" {
		defaultValue, err := tp.parseDefaultValue(defaultTag, field.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid default value for field %s: %w", field.Name, err)
		}
		metadata.Default = defaultValue
	}

	// Parse the 'env' tag
	envTag := field.Tag.Get("env")
	if envTag != "" {
		metadata.Env = envTag
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
func (tp *TagParser) parseDefaultValue(defaultStr string, fieldType reflect.Type) (interface{}, error) {
	switch fieldType.Kind() {
	case reflect.String:
		return defaultStr, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(defaultStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int default value: %s", defaultStr)
		}
		return val, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(defaultStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid uint default value: %s", defaultStr)
		}
		return val, nil
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(defaultStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float default value: %s", defaultStr)
		}
		return val, nil
	case reflect.Bool:
		val, err := strconv.ParseBool(defaultStr)
		if err != nil {
			return nil, fmt.Errorf("invalid bool default value: %s", defaultStr)
		}
		return val, nil
	case reflect.Slice:
		// For slices, split by comma
		if defaultStr == "" {
			return reflect.MakeSlice(fieldType, 0, 0).Interface(), nil
		}
		parts := strings.Split(defaultStr, ",")
		slice := reflect.MakeSlice(fieldType, len(parts), len(parts))
		elemType := fieldType.Elem()

		for i, part := range parts {
			part = strings.TrimSpace(part)
			elemVal, err := tp.parseDefaultValue(part, elemType)
			if err != nil {
				return nil, fmt.Errorf("invalid slice element default value: %s", part)
			}
			slice.Index(i).Set(reflect.ValueOf(elemVal))
		}
		return slice.Interface(), nil
	default:
		// For other types, return as string and let type conversion handle it
		return defaultStr, nil
	}
}

// mapToOptArgsCore maps field metadata to OptArgs Core structures
func (tp *TagParser) mapToOptArgsCore(metadata *FieldMetadata) error {
	if metadata.Positional || metadata.IsSubcommand {
		// Positional arguments and subcommands don't map to OptArgs Core flags
		return nil
	}

	// Determine argument type based on field type
	argType := optargs.NoArgument
	switch metadata.Type.Kind() {
	case reflect.Bool:
		argType = optargs.NoArgument
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		argType = optargs.RequiredArgument
	case reflect.Slice:
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

	// Required positional arguments are valid
	if metadata.Positional && metadata.Required {
		// This is valid
	}

	// Options should have at least one flag (short or long)
	if !metadata.Positional && !metadata.IsSubcommand && metadata.Short == "" && metadata.Long == "" {
		// Generate default long option from field name
		metadata.Long = strings.ToLower(metadata.Name)
	}

	// Validate short option is single character
	if metadata.Short != "" && len(metadata.Short) != 1 {
		return fmt.Errorf("short option must be single character, got: %s", metadata.Short)
	}

	return nil
}

// GetEnvironmentValue gets the value from environment variable if specified
func (tp *TagParser) GetEnvironmentValue(metadata *FieldMetadata) (string, bool) {
	if metadata.Env == "" {
		return "", false
	}

	value, exists := os.LookupEnv(metadata.Env)
	return value, exists
}

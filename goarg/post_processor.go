package goarg

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/major0/optargs"
)

// PostProcessor handles positional args, env vars, defaults, and validation
// after the core parser iteration completes.
type PostProcessor struct {
	metadata    *StructMetadata
	config      Config
	setFields   map[int]bool // from FlagBuilder, read-only during post-processing
	positionals []PositionalArg
}

// PositionalArg represents a positional argument.
type PositionalArg struct {
	Field    *FieldMetadata
	Required bool
	Multiple bool
}

// buildPositionalArgs builds the list of positional arguments from metadata.
func (pp *PostProcessor) buildPositionalArgs() {
	pp.positionals = make([]PositionalArg, 0, len(pp.metadata.Positionals))
	for i := range pp.metadata.Positionals {
		field := &pp.metadata.Positionals[i]
		pp.positionals = append(pp.positionals, PositionalArg{
			Field:    field,
			Required: field.Required,
			Multiple: field.Type.Kind() == reflect.Slice,
		})
	}
}

// Process runs all post-parse steps in order:
// 1. Assign positional arguments
// 2. Apply environment variable fallbacks
// 3. Apply default values
// 4. Validate required fields
func (pp *PostProcessor) Process(parser *optargs.Parser, destValue reflect.Value) error {
	if err := pp.processPositionalArgs(parser, destValue); err != nil {
		return err
	}
	if !pp.config.IgnoreEnv {
		if err := pp.processEnvironmentVariables(destValue); err != nil {
			return err
		}
	}
	if !pp.config.IgnoreDefault {
		if err := pp.setDefaultValues(destValue); err != nil {
			return err
		}
	}
	return validateRequired(destValue.Addr().Interface(), pp.metadata)
}

// processPositionalArgs processes positional arguments from remaining args.
func (pp *PostProcessor) processPositionalArgs(parser *optargs.Parser, destValue reflect.Value) error {
	remainingArgs := parser.Args
	argIndex := 0

	for _, positional := range pp.positionals {
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
func (pp *PostProcessor) processEnvironmentVariables(destValue reflect.Value) error {
	for i := range pp.metadata.Fields {
		field := &pp.metadata.Fields[i]
		if field.Env == "" {
			continue
		}

		fieldValue := fieldByMeta(destValue, field)
		if !fieldValue.CanSet() {
			continue
		}

		if !isZeroValue(fieldValue) {
			continue
		}

		envName := field.Env
		if pp.config.EnvPrefix != "" {
			envName = pp.config.EnvPrefix + envName
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
func (pp *PostProcessor) setDefaultValues(destValue reflect.Value) error {
	for i := range pp.metadata.Fields {
		field := &pp.metadata.Fields[i]
		if !field.HasDefault {
			continue
		}

		fieldValue := fieldByMeta(destValue, field)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			continue
		}

		// Skip fields explicitly set during parsing (including negatable zero-clear)
		if pp.setFields[field.FieldIndex] {
			continue
		}

		if !isZeroValue(fieldValue) {
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

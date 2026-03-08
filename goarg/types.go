package goarg

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TypeConverter handles all Go type conversions - identical to alexflint/go-arg
type TypeConverter struct{}

// ConvertValue converts a string value to the target type
func (tc *TypeConverter) ConvertValue(value string, targetType reflect.Type) (interface{}, error) {
	// Handle pointer types by converting to the underlying type first
	if targetType.Kind() == reflect.Ptr {
		elemType := targetType.Elem()
		converted, err := tc.ConvertValue(value, elemType)
		if err != nil {
			return nil, err
		}

		// Create a pointer to the converted value
		ptr := reflect.New(elemType)
		ptr.Elem().Set(reflect.ValueOf(converted))
		return ptr.Interface(), nil
	}

	// Handle slice types
	if targetType.Kind() == reflect.Slice {
		return tc.convertSliceValue([]string{value}, targetType)
	}

	// Handle basic types
	switch targetType.Kind() {
	case reflect.String:
		return tc.ConvertString(value), nil
	case reflect.Bool:
		return tc.ConvertBool(value)
	case reflect.Int:
		return tc.ConvertInt(value)
	case reflect.Int8:
		val, err := strconv.ParseInt(value, 10, 8)
		return int8(val), err
	case reflect.Int16:
		val, err := strconv.ParseInt(value, 10, 16)
		return int16(val), err
	case reflect.Int32:
		val, err := strconv.ParseInt(value, 10, 32)
		return int32(val), err
	case reflect.Int64:
		return strconv.ParseInt(value, 10, 64)
	case reflect.Uint:
		val, err := strconv.ParseUint(value, 10, 0)
		return uint(val), err
	case reflect.Uint8:
		val, err := strconv.ParseUint(value, 10, 8)
		return uint8(val), err
	case reflect.Uint16:
		val, err := strconv.ParseUint(value, 10, 16)
		return uint16(val), err
	case reflect.Uint32:
		val, err := strconv.ParseUint(value, 10, 32)
		return uint32(val), err
	case reflect.Uint64:
		return strconv.ParseUint(value, 10, 64)
	case reflect.Float32:
		val, err := strconv.ParseFloat(value, 32)
		return float32(val), err
	case reflect.Float64:
		return strconv.ParseFloat(value, 64)
	default:
		// Try custom type conversion
		return tc.ConvertCustom(value, targetType)
	}
}

// convertSliceValue converts a slice of strings to the target slice type
func (tc *TypeConverter) convertSliceValue(values []string, targetType reflect.Type) (interface{}, error) {
	elemType := targetType.Elem()
	slice := reflect.MakeSlice(targetType, 0, len(values))

	for _, value := range values {
		converted, err := tc.ConvertValue(value, elemType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert slice element '%s': %w", value, err)
		}
		slice = reflect.Append(slice, reflect.ValueOf(converted))
	}

	return slice.Interface(), nil
}

// SetField sets a field value using reflection
func (tc *TypeConverter) SetField(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	if value == nil {
		// Handle nil values for pointer types
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		return fmt.Errorf("cannot set nil value to non-pointer field")
	}

	valueReflect := reflect.ValueOf(value)

	// Handle type compatibility
	if valueReflect.Type().AssignableTo(field.Type()) {
		field.Set(valueReflect)
		return nil
	}

	// Handle convertible types
	if valueReflect.Type().ConvertibleTo(field.Type()) {
		field.Set(valueReflect.Convert(field.Type()))
		return nil
	}

	return fmt.Errorf("cannot assign value of type %s to field of type %s",
		valueReflect.Type(), field.Type())
}

// GetDefault gets the default value for a field from struct tags
func (tc *TypeConverter) GetDefault(field reflect.StructField) interface{} {
	defaultTag, exists := field.Tag.Lookup("default")
	if !exists {
		return nil
	}

	// Handle slice types specially - split comma-separated values
	if field.Type.Kind() == reflect.Slice {
		elemType := field.Type.Elem()

		// For empty default value, return empty slice
		if defaultTag == "" {
			return reflect.MakeSlice(field.Type, 0, 0).Interface()
		}

		// Split by comma and trim whitespace
		parts := strings.Split(defaultTag, ",")
		slice := reflect.MakeSlice(field.Type, 0, len(parts))

		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			// Convert each part to the element type
			converted, err := tc.ConvertValue(part, elemType)
			if err != nil {
				return nil
			}

			slice = reflect.Append(slice, reflect.ValueOf(converted))
		}

		return slice.Interface()
	}

	// Convert the default string to the field type
	converted, err := tc.ConvertValue(defaultTag, field.Type)
	if err != nil {
		return nil
	}

	return converted
}

// ValidateRequired validates that all required fields have been set
func (tc *TypeConverter) ValidateRequired(dest interface{}, metadata *StructMetadata) error {
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

		// Check if the field is zero value (not set)
		if tc.isZeroValue(fieldValue) {
			// Generate error message identical to alexflint/go-arg
			if field.Long != "" {
				return fmt.Errorf("--" + field.Long + " is required")
			} else if field.Short != "" {
				return fmt.Errorf("-" + field.Short + " is required")
			} else if field.Positional {
				return fmt.Errorf(field.Name + " is required")
			} else {
				return fmt.Errorf(field.Name + " is required")
			}
		}
	}

	return nil
}

// isZeroValue checks if a reflect.Value is the zero value for its type
func (tc *TypeConverter) isZeroValue(v reflect.Value) bool {
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
		// For arrays, check if all elements are zero
		for i := 0; i < v.Len(); i++ {
			if !tc.isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		// For structs, check if all fields are zero
		for i := 0; i < v.NumField(); i++ {
			if !tc.isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		// For other types, use reflect.Zero comparison
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

// ConvertString converts a string value
func (tc *TypeConverter) ConvertString(value string) string {
	return value
}

// ConvertInt converts a string to int
func (tc *TypeConverter) ConvertInt(value string) (int, error) {
	val, err := strconv.ParseInt(value, 10, 0)
	return int(val), err
}

// ConvertBool converts a string to bool
func (tc *TypeConverter) ConvertBool(value string) (bool, error) {
	// Handle alexflint/go-arg compatible boolean parsing
	switch strings.ToLower(value) {
	case "true", "t", "1", "yes", "y", "on":
		return true, nil
	case "false", "f", "0", "no", "n", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
}

// ConvertCustom converts a string to a custom type using encoding.TextUnmarshaler
func (tc *TypeConverter) ConvertCustom(value string, targetType reflect.Type) (interface{}, error) {
	// Create a new instance of the target type
	var target reflect.Value

	if targetType.Kind() == reflect.Ptr {
		// For pointer types, create a new instance and get its address
		target = reflect.New(targetType.Elem())
	} else {
		// For non-pointer types, create a new instance and get its address
		target = reflect.New(targetType)
	}

	// Check if the type implements encoding.TextUnmarshaler
	if target.Type().Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
		unmarshaler := target.Interface().(encoding.TextUnmarshaler)
		err := unmarshaler.UnmarshalText([]byte(value))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal text for type %s: %w", targetType, err)
		}

		if targetType.Kind() == reflect.Ptr {
			return target.Interface(), nil
		} else {
			return target.Elem().Interface(), nil
		}
	}

	// Check if the pointer to the type implements encoding.TextUnmarshaler
	ptrType := reflect.PtrTo(targetType)
	if ptrType.Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
		ptrTarget := reflect.New(targetType)
		unmarshaler := ptrTarget.Interface().(encoding.TextUnmarshaler)
		err := unmarshaler.UnmarshalText([]byte(value))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal text for type %s: %w", targetType, err)
		}

		if targetType.Kind() == reflect.Ptr {
			return ptrTarget.Interface(), nil
		} else {
			return ptrTarget.Elem().Interface(), nil
		}
	}

	return nil, fmt.Errorf("type %s does not implement encoding.TextUnmarshaler", targetType)
}

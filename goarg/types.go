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

// ConvertSlice converts strings to a slice of the element type
func (tc *TypeConverter) ConvertSlice(values []string, elementType reflect.Type) (interface{}, error) {
	sliceType := reflect.SliceOf(elementType)
	return tc.convertSliceValue(values, sliceType)
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

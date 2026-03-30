package optargs

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Cached reflect.Type for TextUnmarshaler interface check.
var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

// intBitSize maps signed integer kinds to their strconv bit-size parameter.
var intBitSize = [...]int{
	reflect.Int:   0,
	reflect.Int8:  8,
	reflect.Int16: 16,
	reflect.Int32: 32,
	reflect.Int64: 64,
}

// uintBitSize maps unsigned integer kinds to their strconv bit-size parameter.
var uintBitSize = [...]int{
	reflect.Uint:   0,
	reflect.Uint8:  8,
	reflect.Uint16: 16,
	reflect.Uint32: 32,
	reflect.Uint64: 64,
}

// floatBitSize maps float kinds to their strconv bit-size parameter.
var floatBitSize = [...]int{
	reflect.Float32: 32,
	reflect.Float64: 64,
}

// Convert converts a string value to the specified Go type.
// Supports: string, bool, all int/uint/float sizes, pointer types,
// slice types, and types implementing encoding.TextUnmarshaler.
// Bool parsing accepts: true/t/1/yes/y/on and false/f/0/no/n/off
// (case-insensitive), matching alexflint/go-arg behavior.
func Convert(value string, targetType reflect.Type) (interface{}, error) {
	// Handle pointer types: unwrap, convert, wrap in pointer.
	if targetType.Kind() == reflect.Ptr {
		elemType := targetType.Elem()
		converted, err := Convert(value, elemType)
		if err != nil {
			return nil, err
		}
		ptr := reflect.New(elemType)
		ptr.Elem().Set(reflect.ValueOf(converted))
		return ptr.Interface(), nil
	}

	// Handle slice types: convert single value, return single-element slice.
	if targetType.Kind() == reflect.Slice {
		elemType := targetType.Elem()
		converted, err := Convert(value, elemType)
		if err != nil {
			return nil, err
		}
		slice := reflect.MakeSlice(targetType, 1, 1)
		slice.Index(0).Set(reflect.ValueOf(converted))
		return slice.Interface(), nil
	}

	// Try TextUnmarshaler before basic types — user-defined types take priority.
	if result, err, ok := tryTextUnmarshaler(value, targetType); ok {
		return result, err
	}

	kind := targetType.Kind()

	switch {
	case kind == reflect.String:
		return value, nil

	case kind == reflect.Bool:
		return convertBool(value)

	case kind >= reflect.Int && kind <= reflect.Int64:
		bits := intBitSize[kind]
		v, err := strconv.ParseInt(value, 10, bits)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type %s", value, targetType)
		}
		return reflect.ValueOf(v).Convert(targetType).Interface(), nil

	case kind >= reflect.Uint && kind <= reflect.Uint64:
		bits := uintBitSize[kind]
		v, err := strconv.ParseUint(value, 10, bits)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type %s", value, targetType)
		}
		return reflect.ValueOf(v).Convert(targetType).Interface(), nil

	case kind == reflect.Float32 || kind == reflect.Float64:
		bits := floatBitSize[kind]
		v, err := strconv.ParseFloat(value, bits)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type %s", value, targetType)
		}
		return reflect.ValueOf(v).Convert(targetType).Interface(), nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", targetType)
	}
}

// convertBool parses a boolean string value.
// Accepts: true/t/1/yes/y/on and false/f/0/no/n/off (case-insensitive).
func convertBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "true", "t", "1", "yes", "y", "on":
		return true, nil
	case "false", "f", "0", "no", "n", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid value %q for type bool", value)
	}
}

// tryTextUnmarshaler checks if targetType (or *targetType) implements
// encoding.TextUnmarshaler and attempts conversion. The third return
// value indicates whether the interface was found.
func tryTextUnmarshaler(value string, targetType reflect.Type) (interface{}, error, bool) {
	// Check if *targetType implements TextUnmarshaler.
	ptrType := reflect.PointerTo(targetType)
	if ptrType.Implements(textUnmarshalerType) {
		v := reflect.New(targetType)
		u := v.Interface().(encoding.TextUnmarshaler) //nolint:errcheck // Implements() guarantees success
		if err := u.UnmarshalText([]byte(value)); err != nil {
			return nil, fmt.Errorf("invalid value %q for type %s: %w", value, targetType, err), true
		}
		return v.Elem().Interface(), nil, true
	}

	// Check if targetType itself implements TextUnmarshaler (already a pointer type, etc.).
	if targetType.Implements(textUnmarshalerType) {
		v := reflect.New(targetType.Elem())
		u := v.Interface().(encoding.TextUnmarshaler) //nolint:errcheck // Implements() guarantees success
		if err := u.UnmarshalText([]byte(value)); err != nil {
			return nil, fmt.Errorf("invalid value %q for type %s: %w", value, targetType, err), true
		}
		return v.Interface(), nil, true
	}

	return nil, nil, false
}

// ConvertSlice converts a comma-separated string to a slice of the
// specified element type. Used for default value processing.
// Splits by comma, trims whitespace on each element, skips empty
// elements after trimming. Returns empty slice for empty input.
func ConvertSlice(csv string, sliceType reflect.Type) (interface{}, error) {
	if sliceType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("unsupported type: %s", sliceType)
	}

	elemType := sliceType.Elem()
	slice := reflect.MakeSlice(sliceType, 0, 0)

	if csv == "" {
		return slice.Interface(), nil
	}

	parts := strings.Split(csv, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		converted, err := Convert(part, elemType)
		if err != nil {
			return nil, err
		}
		slice = reflect.Append(slice, reflect.ValueOf(converted))
	}

	return slice.Interface(), nil
}

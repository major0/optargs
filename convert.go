package optargs

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

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

	switch targetType.Kind() {
	case reflect.String:
		return value, nil
	case reflect.Bool:
		return convertBool(value)
	case reflect.Int:
		v, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type int", value)
		}
		return int(v), nil
	case reflect.Int8:
		v, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type int8", value)
		}
		return int8(v), nil
	case reflect.Int16:
		v, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type int16", value)
		}
		return int16(v), nil
	case reflect.Int32:
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type int32", value)
		}
		return int32(v), nil
	case reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type int64", value)
		}
		return v, nil
	case reflect.Uint:
		v, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type uint", value)
		}
		return uint(v), nil
	case reflect.Uint8:
		v, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type uint8", value)
		}
		return uint8(v), nil
	case reflect.Uint16:
		v, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type uint16", value)
		}
		return uint16(v), nil
	case reflect.Uint32:
		v, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type uint32", value)
		}
		return uint32(v), nil
	case reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type uint64", value)
		}
		return v, nil
	case reflect.Float32:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type float32", value)
		}
		return float32(v), nil
	case reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q for type float64", value)
		}
		return v, nil
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
	unmarshalerType := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	// Check if *targetType implements TextUnmarshaler.
	ptrType := reflect.PointerTo(targetType)
	if ptrType.Implements(unmarshalerType) {
		v := reflect.New(targetType)
		u := v.Interface().(encoding.TextUnmarshaler) //nolint:errcheck // Implements() guarantees success
		if err := u.UnmarshalText([]byte(value)); err != nil {
			return nil, fmt.Errorf("invalid value %q for type %s: %w", value, targetType, err), true
		}
		return v.Elem().Interface(), nil, true
	}

	// Check if targetType itself implements TextUnmarshaler (already a pointer type, etc.).
	if targetType.Implements(unmarshalerType) {
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

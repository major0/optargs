package pflags

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// getFlagValue looks up a flag and returns its string value, or an error
// if the flag doesn't exist or has the wrong type.
func (f *FlagSet) getFlagValue(name, typeName string) (string, error) {
	flag := f.Lookup(name)
	if flag == nil {
		return "", fmt.Errorf("flag %q not found", name)
	}
	if flag.Value.Type() != typeName {
		return "", fmt.Errorf("flag %q is type %s, not %s", name, flag.Value.Type(), typeName)
	}
	return flag.Value.String(), nil
}

// --- Scalar getters ---.

func (f *FlagSet) GetBool(name string) (bool, error) {
	s, err := f.getFlagValue(name, "bool")
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(s)
}

func (f *FlagSet) GetString(name string) (string, error) {
	return f.getFlagValue(name, "string")
}

func (f *FlagSet) GetInt(name string) (int, error) {
	s, err := f.getFlagValue(name, "int")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

func (f *FlagSet) GetInt8(name string) (int8, error) {
	s, err := f.getFlagValue(name, "int8")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(s, 10, 8)
	return int8(v), err
}

func (f *FlagSet) GetInt16(name string) (int16, error) {
	s, err := f.getFlagValue(name, "int16")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(s, 10, 16)
	return int16(v), err
}

func (f *FlagSet) GetInt32(name string) (int32, error) {
	s, err := f.getFlagValue(name, "int32")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

func (f *FlagSet) GetInt64(name string) (int64, error) {
	s, err := f.getFlagValue(name, "int64")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

func (f *FlagSet) GetUint(name string) (uint, error) {
	s, err := f.getFlagValue(name, "uint")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 0)
	return uint(v), err
}

func (f *FlagSet) GetUint8(name string) (uint8, error) {
	s, err := f.getFlagValue(name, "uint8")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 8)
	return uint8(v), err
}

func (f *FlagSet) GetUint16(name string) (uint16, error) {
	s, err := f.getFlagValue(name, "uint16")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 16)
	return uint16(v), err
}

func (f *FlagSet) GetUint32(name string) (uint32, error) {
	s, err := f.getFlagValue(name, "uint32")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 32)
	return uint32(v), err
}

func (f *FlagSet) GetUint64(name string) (uint64, error) {
	s, err := f.getFlagValue(name, "uint64")
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(s, 10, 64)
}

func (f *FlagSet) GetFloat32(name string) (float32, error) {
	s, err := f.getFlagValue(name, "float32")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

func (f *FlagSet) GetFloat64(name string) (float64, error) {
	s, err := f.getFlagValue(name, "float64")
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}

func (f *FlagSet) GetDuration(name string) (time.Duration, error) {
	s, err := f.getFlagValue(name, "duration")
	if err != nil {
		return 0, err
	}
	return time.ParseDuration(s)
}

func (f *FlagSet) GetCount(name string) (int, error) {
	s, err := f.getFlagValue(name, "count")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

// --- Slice getters ---
// Generic helper collapses the repetitive parse-from-string pattern.

func parseSliceString(s string) []string {
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func getSlice[T any](f *FlagSet, name, typeName string, parse func(string) (T, error)) ([]T, error) {
	s, err := f.getFlagValue(name, typeName)
	if err != nil {
		return nil, err
	}
	parts := parseSliceString(s)
	if parts == nil {
		return nil, nil
	}
	result := make([]T, len(parts))
	for i, p := range parts {
		v, err := parse(p)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

func (f *FlagSet) GetStringSlice(name string) ([]string, error) {
	s, err := f.getFlagValue(name, "stringSlice")
	if err != nil {
		return nil, err
	}
	return parseSliceString(s), nil
}

func (f *FlagSet) GetBoolSlice(name string) ([]bool, error) {
	return getSlice(f, name, "boolSlice", strconv.ParseBool)
}

func (f *FlagSet) GetIntSlice(name string) ([]int, error) {
	return getSlice(f, name, "intSlice", strconv.Atoi)
}

func (f *FlagSet) GetInt32Slice(name string) ([]int32, error) {
	return getSlice(f, name, "int32Slice", func(s string) (int32, error) {
		v, err := strconv.ParseInt(s, 10, 32)
		return int32(v), err
	})
}

func (f *FlagSet) GetInt64Slice(name string) ([]int64, error) {
	return getSlice(f, name, "int64Slice", func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) })
}

func (f *FlagSet) GetUintSlice(name string) ([]uint, error) {
	return getSlice(f, name, "uintSlice", func(s string) (uint, error) {
		v, err := strconv.ParseUint(s, 10, 0)
		return uint(v), err
	})
}

func (f *FlagSet) GetFloat32Slice(name string) ([]float32, error) {
	return getSlice(f, name, "float32Slice", func(s string) (float32, error) {
		v, err := strconv.ParseFloat(s, 32)
		return float32(v), err
	})
}

func (f *FlagSet) GetFloat64Slice(name string) ([]float64, error) {
	return getSlice(f, name, "float64Slice", func(s string) (float64, error) { return strconv.ParseFloat(s, 64) })
}

func (f *FlagSet) GetDurationSlice(name string) ([]time.Duration, error) {
	return getSlice(f, name, "durationSlice", time.ParseDuration)
}

// --- Map getters ---.

func parseStringMap(s string) map[string]string {
	s = strings.TrimPrefix(s, "map[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil
	}
	result := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		if before, after, ok := strings.Cut(pair, "="); ok {
			result[before] = after
		}
	}
	return result
}

func (f *FlagSet) GetStringToString(name string) (map[string]string, error) {
	s, err := f.getFlagValue(name, "stringToString")
	if err != nil {
		return nil, err
	}
	return parseStringMap(s), nil
}

func (f *FlagSet) GetStringToInt(name string) (map[string]int, error) {
	s, err := f.getFlagValue(name, "stringToInt")
	if err != nil {
		return nil, err
	}
	sm := parseStringMap(s)
	result := make(map[string]int, len(sm))
	for k, v := range sm {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		result[k] = n
	}
	return result, nil
}

func (f *FlagSet) GetStringToInt64(name string) (map[string]int64, error) {
	s, err := f.getFlagValue(name, "stringToInt64")
	if err != nil {
		return nil, err
	}
	sm := parseStringMap(s)
	result := make(map[string]int64, len(sm))
	for k, v := range sm {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		result[k] = n
	}
	return result, nil
}

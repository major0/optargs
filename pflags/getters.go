package pflags

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// getFlagValue looks up a flag and returns its string value, or an error
// if the flag doesn't exist or hasn't been set.
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

// GetBool returns the bool value of the named flag.
func (f *FlagSet) GetBool(name string) (bool, error) {
	s, err := f.getFlagValue(name, "bool")
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(s)
}

// GetString returns the string value of the named flag.
func (f *FlagSet) GetString(name string) (string, error) {
	s, err := f.getFlagValue(name, "string")
	if err != nil {
		return "", err
	}
	return s, nil
}

// GetInt returns the int value of the named flag.
func (f *FlagSet) GetInt(name string) (int, error) {
	s, err := f.getFlagValue(name, "int")
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	return v, err
}

// GetInt8 returns the int8 value of the named flag.
func (f *FlagSet) GetInt8(name string) (int8, error) {
	s, err := f.getFlagValue(name, "int8")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(s, 10, 8)
	return int8(v), err
}

// GetInt16 returns the int16 value of the named flag.
func (f *FlagSet) GetInt16(name string) (int16, error) {
	s, err := f.getFlagValue(name, "int16")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(s, 10, 16)
	return int16(v), err
}

// GetInt32 returns the int32 value of the named flag.
func (f *FlagSet) GetInt32(name string) (int32, error) {
	s, err := f.getFlagValue(name, "int32")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

// GetInt64 returns the int64 value of the named flag.
func (f *FlagSet) GetInt64(name string) (int64, error) {
	s, err := f.getFlagValue(name, "int64")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

// GetUint returns the uint value of the named flag.
func (f *FlagSet) GetUint(name string) (uint, error) {
	s, err := f.getFlagValue(name, "uint")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 0)
	return uint(v), err
}

// GetUint8 returns the uint8 value of the named flag.
func (f *FlagSet) GetUint8(name string) (uint8, error) {
	s, err := f.getFlagValue(name, "uint8")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 8)
	return uint8(v), err
}

// GetUint16 returns the uint16 value of the named flag.
func (f *FlagSet) GetUint16(name string) (uint16, error) {
	s, err := f.getFlagValue(name, "uint16")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 16)
	return uint16(v), err
}

// GetUint32 returns the uint32 value of the named flag.
func (f *FlagSet) GetUint32(name string) (uint32, error) {
	s, err := f.getFlagValue(name, "uint32")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 32)
	return uint32(v), err
}

// GetUint64 returns the uint64 value of the named flag.
func (f *FlagSet) GetUint64(name string) (uint64, error) {
	s, err := f.getFlagValue(name, "uint64")
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(s, 10, 64)
}

// GetFloat32 returns the float32 value of the named flag.
func (f *FlagSet) GetFloat32(name string) (float32, error) {
	s, err := f.getFlagValue(name, "float32")
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

// GetFloat64 returns the float64 value of the named flag.
func (f *FlagSet) GetFloat64(name string) (float64, error) {
	s, err := f.getFlagValue(name, "float64")
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}

// GetDuration returns the time.Duration value of the named flag.
func (f *FlagSet) GetDuration(name string) (time.Duration, error) {
	s, err := f.getFlagValue(name, "duration")
	if err != nil {
		return 0, err
	}
	return time.ParseDuration(s)
}

// GetCount returns the count value of the named flag.
func (f *FlagSet) GetCount(name string) (int, error) {
	s, err := f.getFlagValue(name, "count")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

// --- Slice getters ---
// Slice values use String() format "[a,b,c]". We parse by trimming brackets
// and splitting on commas.

func parseSliceString(s string) []string {
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

// GetStringSlice returns the []string value of the named flag.
func (f *FlagSet) GetStringSlice(name string) ([]string, error) {
	s, err := f.getFlagValue(name, "stringSlice")
	if err != nil {
		return nil, err
	}
	return parseSliceString(s), nil
}

// GetIntSlice returns the []int value of the named flag.
func (f *FlagSet) GetIntSlice(name string) ([]int, error) {
	s, err := f.getFlagValue(name, "intSlice")
	if err != nil {
		return nil, err
	}
	parts := parseSliceString(s)
	result := make([]int, len(parts))
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

// GetBoolSlice returns the []bool value of the named flag.
func (f *FlagSet) GetBoolSlice(name string) ([]bool, error) {
	s, err := f.getFlagValue(name, "boolSlice")
	if err != nil {
		return nil, err
	}
	parts := parseSliceString(s)
	result := make([]bool, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseBool(p)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

// GetInt32Slice returns the []int32 value of the named flag.
func (f *FlagSet) GetInt32Slice(name string) ([]int32, error) {
	s, err := f.getFlagValue(name, "int32Slice")
	if err != nil {
		return nil, err
	}
	parts := parseSliceString(s)
	result := make([]int32, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseInt(p, 10, 32)
		if err != nil {
			return nil, err
		}
		result[i] = int32(v)
	}
	return result, nil
}

// GetInt64Slice returns the []int64 value of the named flag.
func (f *FlagSet) GetInt64Slice(name string) ([]int64, error) {
	s, err := f.getFlagValue(name, "int64Slice")
	if err != nil {
		return nil, err
	}
	parts := parseSliceString(s)
	result := make([]int64, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

// GetUintSlice returns the []uint value of the named flag.
func (f *FlagSet) GetUintSlice(name string) ([]uint, error) {
	s, err := f.getFlagValue(name, "uintSlice")
	if err != nil {
		return nil, err
	}
	parts := parseSliceString(s)
	result := make([]uint, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseUint(p, 10, 0)
		if err != nil {
			return nil, err
		}
		result[i] = uint(v)
	}
	return result, nil
}

// GetFloat32Slice returns the []float32 value of the named flag.
func (f *FlagSet) GetFloat32Slice(name string) ([]float32, error) {
	s, err := f.getFlagValue(name, "float32Slice")
	if err != nil {
		return nil, err
	}
	parts := parseSliceString(s)
	result := make([]float32, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseFloat(p, 32)
		if err != nil {
			return nil, err
		}
		result[i] = float32(v)
	}
	return result, nil
}

// GetFloat64Slice returns the []float64 value of the named flag.
func (f *FlagSet) GetFloat64Slice(name string) ([]float64, error) {
	s, err := f.getFlagValue(name, "float64Slice")
	if err != nil {
		return nil, err
	}
	parts := parseSliceString(s)
	result := make([]float64, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

// GetDurationSlice returns the []time.Duration value of the named flag.
func (f *FlagSet) GetDurationSlice(name string) ([]time.Duration, error) {
	s, err := f.getFlagValue(name, "durationSlice")
	if err != nil {
		return nil, err
	}
	parts := parseSliceString(s)
	result := make([]time.Duration, len(parts))
	for i, p := range parts {
		v, err := time.ParseDuration(p)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

// --- Map getters ---

// GetStringToString returns the map[string]string value of the named flag.
func (f *FlagSet) GetStringToString(name string) (map[string]string, error) {
	s, err := f.getFlagValue(name, "stringToString")
	if err != nil {
		return nil, err
	}
	return parseStringMap(s), nil
}

// GetStringToInt returns the map[string]int value of the named flag.
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

// GetStringToInt64 returns the map[string]int64 value of the named flag.
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

// parseStringMap parses "map[k1=v1,k2=v2]" format.
func parseStringMap(s string) map[string]string {
	s = strings.TrimPrefix(s, "map[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil
	}
	result := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		idx := strings.Index(pair, "=")
		if idx < 0 {
			continue
		}
		result[pair[:idx]] = pair[idx+1:]
	}
	return result
}

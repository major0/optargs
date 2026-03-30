package pflags

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	if p == nil {
		p = new(string)
	}
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string { return string(*s) }

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	if p == nil {
		p = new(bool)
	}
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("invalid boolean value '%s'", s)
	}
	*b = boolValue(v)
	return nil
}

func (b *boolValue) Type() string {
	return "bool"
}

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

// IsBoolFlag returns true to indicate this is a boolean flag
func (b *boolValue) IsBoolFlag() bool { return true }

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
	if p == nil {
		p = new(int)
	}
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid syntax for integer flag: %s", s)
	}
	*i = intValue(v)
	return nil
}

func (i *intValue) Type() string {
	return "int"
}

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	if p == nil {
		p = new(int64)
	}
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return fmt.Errorf("invalid syntax for int64 flag: %s", s)
	}
	*i = int64Value(v)
	return nil
}

func (i *int64Value) Type() string {
	return "int64"
}

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	if p == nil {
		p = new(uint)
	}
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 0)
	if err != nil {
		return fmt.Errorf("invalid syntax for uint flag: %s", s)
	}
	*i = uintValue(v)
	return nil
}

func (i *uintValue) Type() string {
	return "uint"
}

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- uint64 Value
type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	if p == nil {
		p = new(uint64)
	}
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		return fmt.Errorf("invalid syntax for uint64 flag: %s", s)
	}
	*i = uint64Value(v)
	return nil
}

func (i *uint64Value) Type() string {
	return "uint64"
}

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	if p == nil {
		p = new(float64)
	}
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("invalid syntax for float64 flag: %s", s)
	}
	*f = float64Value(v)
	return nil
}

func (f *float64Value) Type() string {
	return "float64"
}

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

// -- time.Duration Value
type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
	if p == nil {
		p = new(time.Duration)
	}
	*p = val
	return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration format for flag: %s", s)
	}
	*d = durationValue(v)
	return nil
}

func (d *durationValue) Type() string {
	return "duration"
}

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

// -- []string Value
type stringSliceValue []string

func newStringSliceValue(val []string, p *[]string) *stringSliceValue {
	if p == nil {
		p = new([]string)
	}
	*p = val
	return (*stringSliceValue)(p)
}

func (s *stringSliceValue) Set(val string) error {
	// Support both comma-separated and repeated flag usage
	if strings.Contains(val, ",") {
		// Split by comma and append all parts
		parts := strings.Split(val, ",")
		for _, part := range parts {
			*s = append(*s, strings.TrimSpace(part))
		}
	} else {
		// Single value, append directly
		*s = append(*s, val)
	}
	return nil
}

func (s *stringSliceValue) Type() string {
	return "stringSlice"
}

func (s *stringSliceValue) String() string {
	if len(*s) == 0 {
		return "[]"
	}
	return fmt.Sprintf("[%s]", strings.Join(*s, ","))
}

// -- []int Value
type intSliceValue []int

func newIntSliceValue(val []int, p *[]int) *intSliceValue {
	if p == nil {
		p = new([]int)
	}
	*p = val
	return (*intSliceValue)(p)
}

func (i *intSliceValue) Set(val string) error {
	// Support both comma-separated and repeated flag usage
	if strings.Contains(val, ",") {
		// Split by comma and parse each part
		parts := strings.Split(val, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			v, err := strconv.Atoi(trimmed)
			if err != nil {
				return fmt.Errorf("invalid syntax for integer slice element: %s", trimmed)
			}
			*i = append(*i, v)
		}
	} else {
		// Single value, parse and append
		v, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid syntax for integer slice element: %s", val)
		}
		*i = append(*i, v)
	}
	return nil
}

func (i *intSliceValue) Type() string {
	return "intSlice"
}

func (i *intSliceValue) String() string {
	if len(*i) == 0 {
		return "[]"
	}
	strs := make([]string, len(*i))
	for idx, v := range *i {
		strs[idx] = strconv.Itoa(v)
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ","))
}

// -- func Value (wraps a callback function)
type funcValue func(string) error

func (f funcValue) Set(s string) error { return f(s) }
func (f funcValue) Type() string       { return "" }
func (f funcValue) String() string     { return "" }

// -- boolFunc Value (wraps a boolean callback function)
type boolFuncValue func(string) error

func (f boolFuncValue) Set(s string) error { return f(s) }
func (f boolFuncValue) Type() string       { return "bool" }
func (f boolFuncValue) String() string     { return "" }
func (f boolFuncValue) IsBoolFlag() bool   { return true }

// -- int8 Value
type int8Value int8

func newInt8Value(val int8, p *int8) *int8Value {
	if p == nil {
		p = new(int8)
	}
	*p = val
	return (*int8Value)(p)
}

func (i *int8Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 8)
	if err != nil {
		return fmt.Errorf("invalid syntax for int8 flag: %s", s)
	}
	*i = int8Value(v)
	return nil
}

func (i *int8Value) Type() string   { return "int8" }
func (i *int8Value) String() string { return strconv.FormatInt(int64(*i), 10) }

// -- int16 Value
type int16Value int16

func newInt16Value(val int16, p *int16) *int16Value {
	if p == nil {
		p = new(int16)
	}
	*p = val
	return (*int16Value)(p)
}

func (i *int16Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 16)
	if err != nil {
		return fmt.Errorf("invalid syntax for int16 flag: %s", s)
	}
	*i = int16Value(v)
	return nil
}

func (i *int16Value) Type() string   { return "int16" }
func (i *int16Value) String() string { return strconv.FormatInt(int64(*i), 10) }

// -- int32 Value
type int32Value int32

func newInt32Value(val int32, p *int32) *int32Value {
	if p == nil {
		p = new(int32)
	}
	*p = val
	return (*int32Value)(p)
}

func (i *int32Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 32)
	if err != nil {
		return fmt.Errorf("invalid syntax for int32 flag: %s", s)
	}
	*i = int32Value(v)
	return nil
}

func (i *int32Value) Type() string   { return "int32" }
func (i *int32Value) String() string { return strconv.FormatInt(int64(*i), 10) }

// -- uint8 Value
type uint8Value uint8

func newUint8Value(val uint8, p *uint8) *uint8Value {
	if p == nil {
		p = new(uint8)
	}
	*p = val
	return (*uint8Value)(p)
}

func (i *uint8Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 8)
	if err != nil {
		return fmt.Errorf("invalid syntax for uint8 flag: %s", s)
	}
	*i = uint8Value(v)
	return nil
}

func (i *uint8Value) Type() string   { return "uint8" }
func (i *uint8Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- uint16 Value
type uint16Value uint16

func newUint16Value(val uint16, p *uint16) *uint16Value {
	if p == nil {
		p = new(uint16)
	}
	*p = val
	return (*uint16Value)(p)
}

func (i *uint16Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 16)
	if err != nil {
		return fmt.Errorf("invalid syntax for uint16 flag: %s", s)
	}
	*i = uint16Value(v)
	return nil
}

func (i *uint16Value) Type() string   { return "uint16" }
func (i *uint16Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- uint32 Value
type uint32Value uint32

func newUint32Value(val uint32, p *uint32) *uint32Value {
	if p == nil {
		p = new(uint32)
	}
	*p = val
	return (*uint32Value)(p)
}

func (i *uint32Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return fmt.Errorf("invalid syntax for uint32 flag: %s", s)
	}
	*i = uint32Value(v)
	return nil
}

func (i *uint32Value) Type() string   { return "uint32" }
func (i *uint32Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- float32 Value
type float32Value float32

func newFloat32Value(val float32, p *float32) *float32Value {
	if p == nil {
		p = new(float32)
	}
	*p = val
	return (*float32Value)(p)
}

func (f *float32Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return fmt.Errorf("invalid syntax for float32 flag: %s", s)
	}
	*f = float32Value(v)
	return nil
}

func (f *float32Value) Type() string   { return "float32" }
func (f *float32Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 32) }

// -- []bool Value
type boolSliceValue []bool

func newBoolSliceValue(val []bool, p *[]bool) *boolSliceValue {
	if p == nil { p = new([]bool) }
	*p = val
	return (*boolSliceValue)(p)
}

func (s *boolSliceValue) Set(val string) error {
	for _, part := range strings.Split(val, ",") {
		v, err := strconv.ParseBool(strings.TrimSpace(part))
		if err != nil {
			return fmt.Errorf("invalid boolean slice element: %s", strings.TrimSpace(part))
		}
		*s = append(*s, v)
	}
	return nil
}

func (s *boolSliceValue) Type() string { return "boolSlice" }
func (s *boolSliceValue) String() string {
	if len(*s) == 0 { return "[]" }
	parts := make([]string, len(*s))
	for i, v := range *s { parts[i] = strconv.FormatBool(v) }
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

// -- []int32 Value
type int32SliceValue []int32

func newInt32SliceValue(val []int32, p *[]int32) *int32SliceValue {
	if p == nil { p = new([]int32) }
	*p = val
	return (*int32SliceValue)(p)
}

func (s *int32SliceValue) Set(val string) error {
	for _, part := range strings.Split(val, ",") {
		trimmed := strings.TrimSpace(part)
		v, err := strconv.ParseInt(trimmed, 0, 32)
		if err != nil {
			return fmt.Errorf("invalid syntax for int32 slice element: %s", trimmed)
		}
		*s = append(*s, int32(v))
	}
	return nil
}

func (s *int32SliceValue) Type() string { return "int32Slice" }
func (s *int32SliceValue) String() string {
	if len(*s) == 0 { return "[]" }
	parts := make([]string, len(*s))
	for i, v := range *s { parts[i] = strconv.FormatInt(int64(v), 10) }
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

// -- []int64 Value
type int64SliceValue []int64

func newInt64SliceValue(val []int64, p *[]int64) *int64SliceValue {
	if p == nil { p = new([]int64) }
	*p = val
	return (*int64SliceValue)(p)
}

func (s *int64SliceValue) Set(val string) error {
	for _, part := range strings.Split(val, ",") {
		trimmed := strings.TrimSpace(part)
		v, err := strconv.ParseInt(trimmed, 0, 64)
		if err != nil {
			return fmt.Errorf("invalid syntax for int64 slice element: %s", trimmed)
		}
		*s = append(*s, v)
	}
	return nil
}

func (s *int64SliceValue) Type() string { return "int64Slice" }
func (s *int64SliceValue) String() string {
	if len(*s) == 0 { return "[]" }
	parts := make([]string, len(*s))
	for i, v := range *s { parts[i] = strconv.FormatInt(v, 10) }
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

// -- []uint Value
type uintSliceValue []uint

func newUintSliceValue(val []uint, p *[]uint) *uintSliceValue {
	if p == nil { p = new([]uint) }
	*p = val
	return (*uintSliceValue)(p)
}

func (s *uintSliceValue) Set(val string) error {
	for _, part := range strings.Split(val, ",") {
		trimmed := strings.TrimSpace(part)
		v, err := strconv.ParseUint(trimmed, 0, 0)
		if err != nil {
			return fmt.Errorf("invalid syntax for uint slice element: %s", trimmed)
		}
		*s = append(*s, uint(v))
	}
	return nil
}

func (s *uintSliceValue) Type() string { return "uintSlice" }
func (s *uintSliceValue) String() string {
	if len(*s) == 0 { return "[]" }
	parts := make([]string, len(*s))
	for i, v := range *s { parts[i] = strconv.FormatUint(uint64(v), 10) }
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

// -- []float32 Value
type float32SliceValue []float32

func newFloat32SliceValue(val []float32, p *[]float32) *float32SliceValue {
	if p == nil { p = new([]float32) }
	*p = val
	return (*float32SliceValue)(p)
}

func (s *float32SliceValue) Set(val string) error {
	for _, part := range strings.Split(val, ",") {
		trimmed := strings.TrimSpace(part)
		v, err := strconv.ParseFloat(trimmed, 32)
		if err != nil {
			return fmt.Errorf("invalid syntax for float32 slice element: %s", trimmed)
		}
		*s = append(*s, float32(v))
	}
	return nil
}

func (s *float32SliceValue) Type() string { return "float32Slice" }
func (s *float32SliceValue) String() string {
	if len(*s) == 0 { return "[]" }
	parts := make([]string, len(*s))
	for i, v := range *s { parts[i] = strconv.FormatFloat(float64(v), 'g', -1, 32) }
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

// -- []float64 Value
type float64SliceValue []float64

func newFloat64SliceValue(val []float64, p *[]float64) *float64SliceValue {
	if p == nil { p = new([]float64) }
	*p = val
	return (*float64SliceValue)(p)
}

func (s *float64SliceValue) Set(val string) error {
	for _, part := range strings.Split(val, ",") {
		trimmed := strings.TrimSpace(part)
		v, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return fmt.Errorf("invalid syntax for float64 slice element: %s", trimmed)
		}
		*s = append(*s, v)
	}
	return nil
}

func (s *float64SliceValue) Type() string { return "float64Slice" }
func (s *float64SliceValue) String() string {
	if len(*s) == 0 { return "[]" }
	parts := make([]string, len(*s))
	for i, v := range *s { parts[i] = strconv.FormatFloat(v, 'g', -1, 64) }
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

// -- []time.Duration Value
type durationSliceValue []time.Duration

func newDurationSliceValue(val []time.Duration, p *[]time.Duration) *durationSliceValue {
	if p == nil { p = new([]time.Duration) }
	*p = val
	return (*durationSliceValue)(p)
}

func (s *durationSliceValue) Set(val string) error {
	for _, part := range strings.Split(val, ",") {
		trimmed := strings.TrimSpace(part)
		v, err := time.ParseDuration(trimmed)
		if err != nil {
			return fmt.Errorf("invalid duration slice element: %s", trimmed)
		}
		*s = append(*s, v)
	}
	return nil
}

func (s *durationSliceValue) Type() string { return "durationSlice" }
func (s *durationSliceValue) String() string {
	if len(*s) == 0 { return "[]" }
	parts := make([]string, len(*s))
	for i, v := range *s { parts[i] = v.String() }
	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

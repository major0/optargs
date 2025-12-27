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

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
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
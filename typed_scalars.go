package optargs

import (
	"fmt"
	"reflect"
	"time"
)

// stringValue wraps a *string destination.
type stringValue struct{ p *string }

// NewStringValue returns a TypedValue backed by p.
// If p is nil, an internal pointer is allocated.
func NewStringValue(val string, p *string) TypedValue {
	if p == nil {
		p = new(string)
	}
	*p = val
	return &stringValue{p: p}
}

func (v *stringValue) Set(s string) error { *v.p = s; return nil }
func (v *stringValue) String() string     { return *v.p }
func (v *stringValue) Type() string       { return "string" }

// boolValue wraps a *bool destination. Implements BoolValuer.
type boolValue struct{ p *bool }

// NewBoolValue returns a TypedValue backed by p.
// If p is nil, an internal pointer is allocated.
func NewBoolValue(val bool, p *bool) TypedValue {
	if p == nil {
		p = new(bool)
	}
	*p = val
	return &boolValue{p: p}
}

func (v *boolValue) Set(s string) error {
	b, err := Convert(s, reflect.TypeOf(false))
	if err != nil {
		return err
	}
	*v.p = b.(bool)
	return nil
}

func (v *boolValue) String() string  { return fmt.Sprintf("%t", *v.p) }
func (v *boolValue) Type() string    { return "bool" }
func (v *boolValue) IsBoolFlag() bool { return true }

// intValue wraps a *int destination.
type intValue struct{ p *int }

func NewIntValue(val int, p *int) TypedValue {
	if p == nil {
		p = new(int)
	}
	*p = val
	return &intValue{p: p}
}

func (v *intValue) Set(s string) error {
	result, err := Convert(s, reflect.TypeOf(int(0)))
	if err != nil {
		return err
	}
	*v.p = result.(int)
	return nil
}

func (v *intValue) String() string { return fmt.Sprintf("%d", *v.p) }
func (v *intValue) Type() string   { return "int" }

// int64Value wraps a *int64 destination.
type int64Value struct{ p *int64 }

func NewInt64Value(val int64, p *int64) TypedValue {
	if p == nil {
		p = new(int64)
	}
	*p = val
	return &int64Value{p: p}
}

func (v *int64Value) Set(s string) error {
	result, err := Convert(s, reflect.TypeOf(int64(0)))
	if err != nil {
		return err
	}
	*v.p = result.(int64)
	return nil
}

func (v *int64Value) String() string { return fmt.Sprintf("%d", *v.p) }
func (v *int64Value) Type() string   { return "int64" }

// uintValue wraps a *uint destination.
type uintValue struct{ p *uint }

func NewUintValue(val uint, p *uint) TypedValue {
	if p == nil {
		p = new(uint)
	}
	*p = val
	return &uintValue{p: p}
}

func (v *uintValue) Set(s string) error {
	result, err := Convert(s, reflect.TypeOf(uint(0)))
	if err != nil {
		return err
	}
	*v.p = result.(uint)
	return nil
}

func (v *uintValue) String() string { return fmt.Sprintf("%d", *v.p) }
func (v *uintValue) Type() string   { return "uint" }

// uint64Value wraps a *uint64 destination.
type uint64Value struct{ p *uint64 }

func NewUint64Value(val uint64, p *uint64) TypedValue {
	if p == nil {
		p = new(uint64)
	}
	*p = val
	return &uint64Value{p: p}
}

func (v *uint64Value) Set(s string) error {
	result, err := Convert(s, reflect.TypeOf(uint64(0)))
	if err != nil {
		return err
	}
	*v.p = result.(uint64)
	return nil
}

func (v *uint64Value) String() string { return fmt.Sprintf("%d", *v.p) }
func (v *uint64Value) Type() string   { return "uint64" }

// float64Value wraps a *float64 destination.
type float64Value struct{ p *float64 }

func NewFloat64Value(val float64, p *float64) TypedValue {
	if p == nil {
		p = new(float64)
	}
	*p = val
	return &float64Value{p: p}
}

func (v *float64Value) Set(s string) error {
	result, err := Convert(s, reflect.TypeOf(float64(0)))
	if err != nil {
		return err
	}
	*v.p = result.(float64)
	return nil
}

func (v *float64Value) String() string { return fmt.Sprintf("%g", *v.p) }
func (v *float64Value) Type() string   { return "float64" }

// durationValue wraps a *time.Duration destination.
// Duration is stored as int64 nanoseconds, so we use time.ParseDuration
// for Set and time.Duration.String() for String.
type durationValue struct{ p *time.Duration }

func NewDurationValue(val time.Duration, p *time.Duration) TypedValue {
	if p == nil {
		p = new(time.Duration)
	}
	*p = val
	return &durationValue{p: p}
}

func (v *durationValue) Set(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid value %q for type duration", s)
	}
	*v.p = d
	return nil
}

func (v *durationValue) String() string { return v.p.String() }
func (v *durationValue) Type() string   { return "duration" }

package optargs

import (
	"encoding"
	"fmt"
	"strings"
)

// --- StringArray: appends raw string without comma splitting ---

type stringArrayValue struct{ p *[]string }

// NewStringArrayValue returns a TypedValue that appends each Set() call's
// raw string without splitting on commas. Distinct from StringSlice.
func NewStringArrayValue(val []string, p *[]string) TypedValue {
	if p == nil {
		p = new([]string)
	}
	*p = val
	return &stringArrayValue{p: p}
}

func (v *stringArrayValue) Set(s string) error {
	*v.p = append(*v.p, s)
	return nil
}

func (v *stringArrayValue) String() string {
	if len(*v.p) == 0 {
		return "[]"
	}
	return "[" + strings.Join(*v.p, ",") + "]"
}

func (v *stringArrayValue) Type() string { return "stringArray" }

// --- Count: increments on each Set(), implements BoolValuer ---

type countValue struct{ p *int }

// NewCountValue returns a TypedValue backed by *int that increments
// on each Set() call. Implements BoolValuer so it works as a
// no-argument flag (e.g., -vvv for verbosity 3).
func NewCountValue(val int, p *int) TypedValue {
	if p == nil {
		p = new(int)
	}
	*p = val
	return &countValue{p: p}
}

func (v *countValue) Set(_ string) error {
	*v.p++
	return nil
}

func (v *countValue) String() string  { return fmt.Sprintf("%d", *v.p) }
func (v *countValue) Type() string    { return "count" }
func (v *countValue) IsBoolFlag() bool { return true }

// --- TextValue: wraps encoding.TextUnmarshaler/TextMarshaler ---

type textValue struct {
	dest encoding.TextUnmarshaler
	val  encoding.TextMarshaler // may be nil if dest doesn't implement TextMarshaler
}

// NewTextValue returns a TypedValue wrapping any encoding.TextUnmarshaler.
// If dest also implements encoding.TextMarshaler, String() uses MarshalText().
// The val parameter provides the initial/display value via TextMarshaler.
func NewTextValue(val encoding.TextMarshaler, dest encoding.TextUnmarshaler) TypedValue {
	tv := &textValue{dest: dest}
	if m, ok := dest.(encoding.TextMarshaler); ok {
		tv.val = m
	}
	// Initialize from val if provided and dest supports it.
	if val != nil {
		if b, err := val.MarshalText(); err == nil {
			_ = dest.UnmarshalText(b) //nolint:errcheck // best-effort init
		}
	}
	return tv
}

func (v *textValue) Set(s string) error {
	return v.dest.UnmarshalText([]byte(s))
}

func (v *textValue) String() string {
	if v.val != nil {
		if b, err := v.val.MarshalText(); err == nil {
			return string(b)
		}
	}
	return ""
}

func (v *textValue) Type() string { return "textUnmarshaler" }

// --- FuncValue: wraps a callback function ---

type funcValue struct {
	fn       func(string) error
	boolFlag bool
	typeName string
}

// NewFuncValue returns a TypedValue that calls fn on each Set().
func NewFuncValue(fn func(string) error) TypedValue {
	return &funcValue{fn: fn, typeName: "func"}
}

// NewBoolFuncValue returns a TypedValue that calls fn on each Set()
// and implements BoolValuer so it works as a no-argument flag.
func NewBoolFuncValue(fn func(string) error) TypedValue {
	return &funcValue{fn: fn, boolFlag: true, typeName: "boolFunc"}
}

func (v *funcValue) Set(s string) error { return v.fn(s) }
func (v *funcValue) String() string     { return "" }
func (v *funcValue) Type() string       { return v.typeName }
func (v *funcValue) IsBoolFlag() bool   { return v.boolFlag }

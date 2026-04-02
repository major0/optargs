package optargs

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Cached reflect.Type values — avoids allocation on every Set() call.
var (
	boolType    = reflect.TypeFor[bool]()
	intType     = reflect.TypeFor[int]()
	int8Type    = reflect.TypeFor[int8]()
	int16Type   = reflect.TypeFor[int16]()
	int32Type   = reflect.TypeFor[int32]()
	int64Type   = reflect.TypeFor[int64]()
	uintType    = reflect.TypeFor[uint]()
	uint8Type   = reflect.TypeFor[uint8]()
	uint16Type  = reflect.TypeFor[uint16]()
	uint32Type  = reflect.TypeFor[uint32]()
	uint64Type  = reflect.TypeFor[uint64]()
	float32Type = reflect.TypeFor[float32]()
	float64Type = reflect.TypeFor[float64]()
	stringType  = reflect.TypeFor[string]()
)

// scalarValue is the generic implementation for all scalar typed values
// that delegate to Convert(). Covers int, int8..int64, uint, uint8..uint64,
// float32, float64.
type scalarValue[T any] struct {
	p      *T
	rtype  reflect.Type
	tname  string
	fmtStr string
}

func newScalar[T any](val T, p *T, rtype reflect.Type, tname, fmtStr string) *scalarValue[T] {
	if p == nil {
		p = new(T)
	}
	*p = val
	return &scalarValue[T]{p: p, rtype: rtype, tname: tname, fmtStr: fmtStr}
}

func (v *scalarValue[T]) Set(s string) error {
	result, err := Convert(s, v.rtype)
	if err != nil {
		return err
	}
	*v.p = result.(T) //nolint:errcheck // Convert guarantees the correct type for v.rtype
	return nil
}

func (v *scalarValue[T]) String() string { return fmt.Sprintf(v.fmtStr, *v.p) }
func (v *scalarValue[T]) Type() string   { return v.tname }

// NewIntValue returns a TypedValue backed by *p, initialized to val.
func NewIntValue(val int, p *int) TypedValue { return newScalar(val, p, intType, "int", "%d") }

// NewInt8Value returns a TypedValue backed by *p, initialized to val.
func NewInt8Value(val int8, p *int8) TypedValue { return newScalar(val, p, int8Type, "int8", "%d") }

// NewInt16Value returns a TypedValue backed by *p, initialized to val.
func NewInt16Value(val int16, p *int16) TypedValue {
	return newScalar(val, p, int16Type, "int16", "%d")
}

// NewInt32Value returns a TypedValue backed by *p, initialized to val.
func NewInt32Value(val int32, p *int32) TypedValue {
	return newScalar(val, p, int32Type, "int32", "%d")
}

// NewInt64Value returns a TypedValue backed by *p, initialized to val.
func NewInt64Value(val int64, p *int64) TypedValue {
	return newScalar(val, p, int64Type, "int64", "%d")
}

// NewUintValue returns a TypedValue backed by *p, initialized to val.
func NewUintValue(val uint, p *uint) TypedValue { return newScalar(val, p, uintType, "uint", "%d") }

// NewUint8Value returns a TypedValue backed by *p, initialized to val.
func NewUint8Value(val uint8, p *uint8) TypedValue {
	return newScalar(val, p, uint8Type, "uint8", "%d")
}

// NewUint16Value returns a TypedValue backed by *p, initialized to val.
func NewUint16Value(val uint16, p *uint16) TypedValue {
	return newScalar(val, p, uint16Type, "uint16", "%d")
}

// NewUint32Value returns a TypedValue backed by *p, initialized to val.
func NewUint32Value(val uint32, p *uint32) TypedValue {
	return newScalar(val, p, uint32Type, "uint32", "%d")
}

// NewUint64Value returns a TypedValue backed by *p, initialized to val.
func NewUint64Value(val uint64, p *uint64) TypedValue {
	return newScalar(val, p, uint64Type, "uint64", "%d")
}

// NewFloat32Value returns a TypedValue backed by *p, initialized to val.
func NewFloat32Value(val float32, p *float32) TypedValue {
	return newScalar(val, p, float32Type, "float32", "%g")
}

// NewFloat64Value returns a TypedValue backed by *p, initialized to val.
func NewFloat64Value(val float64, p *float64) TypedValue {
	return newScalar(val, p, float64Type, "float64", "%g")
}

// String value: no Convert needed.

type stringValue struct{ p *string }

// NewStringValue returns a TypedValue backed by *p, initialized to val.
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

// Bool value: implements BoolValuer.

type boolValue struct{ p *bool }

// NewBoolValue returns a TypedValue backed by *p, initialized to val.
func NewBoolValue(val bool, p *bool) TypedValue {
	if p == nil {
		p = new(bool)
	}
	*p = val
	return &boolValue{p: p}
}

func (v *boolValue) Set(s string) error {
	b, err := Convert(s, boolType)
	if err != nil {
		return err
	}
	*v.p = b.(bool) //nolint:errcheck // Convert guarantees bool for boolType
	return nil
}

func (v *boolValue) String() string     { return strconv.FormatBool(*v.p) }
func (v *boolValue) Type() string       { return "bool" }
func (v *boolValue) IsBoolFlag() bool   { return true }
func (v *boolValue) BoolTakesArg() bool { return true }

// Duration value: uses time.ParseDuration, not Convert.

type durationValue struct{ p *time.Duration }

// NewDurationValue returns a TypedValue backed by *p, initialized to val.
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

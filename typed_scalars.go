package optargs

import (
	"fmt"
	"reflect"
	"time"
)

// Cached reflect.Type values — avoids allocation on every Set() call.
var (
	boolType     = reflect.TypeOf(false)
	intType      = reflect.TypeOf(int(0))
	int8Type     = reflect.TypeOf(int8(0))
	int16Type    = reflect.TypeOf(int16(0))
	int32Type    = reflect.TypeOf(int32(0))
	int64Type    = reflect.TypeOf(int64(0))
	uintType     = reflect.TypeOf(uint(0))
	uint8Type    = reflect.TypeOf(uint8(0))
	uint16Type   = reflect.TypeOf(uint16(0))
	uint32Type   = reflect.TypeOf(uint32(0))
	uint64Type   = reflect.TypeOf(uint64(0))
	float32Type  = reflect.TypeOf(float32(0))
	float64Type  = reflect.TypeOf(float64(0))
	stringType   = reflect.TypeOf("")
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
	*v.p = result.(T)
	return nil
}

func (v *scalarValue[T]) String() string { return fmt.Sprintf(v.fmtStr, *v.p) }
func (v *scalarValue[T]) Type() string   { return v.tname }

// --- Scalar constructors ---

func NewIntValue(val int, p *int) TypedValue           { return newScalar(val, p, intType, "int", "%d") }
func NewInt8Value(val int8, p *int8) TypedValue         { return newScalar(val, p, int8Type, "int8", "%d") }
func NewInt16Value(val int16, p *int16) TypedValue      { return newScalar(val, p, int16Type, "int16", "%d") }
func NewInt32Value(val int32, p *int32) TypedValue      { return newScalar(val, p, int32Type, "int32", "%d") }
func NewInt64Value(val int64, p *int64) TypedValue      { return newScalar(val, p, int64Type, "int64", "%d") }
func NewUintValue(val uint, p *uint) TypedValue         { return newScalar(val, p, uintType, "uint", "%d") }
func NewUint8Value(val uint8, p *uint8) TypedValue      { return newScalar(val, p, uint8Type, "uint8", "%d") }
func NewUint16Value(val uint16, p *uint16) TypedValue   { return newScalar(val, p, uint16Type, "uint16", "%d") }
func NewUint32Value(val uint32, p *uint32) TypedValue   { return newScalar(val, p, uint32Type, "uint32", "%d") }
func NewUint64Value(val uint64, p *uint64) TypedValue   { return newScalar(val, p, uint64Type, "uint64", "%d") }
func NewFloat32Value(val float32, p *float32) TypedValue { return newScalar(val, p, float32Type, "float32", "%g") }
func NewFloat64Value(val float64, p *float64) TypedValue { return newScalar(val, p, float64Type, "float64", "%g") }

// --- String: no Convert needed ---

type stringValue struct{ p *string }

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

// --- Bool: implements BoolValuer ---

type boolValue struct{ p *bool }

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
	*v.p = b.(bool)
	return nil
}

func (v *boolValue) String() string     { return fmt.Sprintf("%t", *v.p) }
func (v *boolValue) Type() string       { return "bool" }
func (v *boolValue) IsBoolFlag() bool   { return true }
func (v *boolValue) BoolTakesArg() bool { return true }

// --- Duration: uses time.ParseDuration, not Convert ---

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

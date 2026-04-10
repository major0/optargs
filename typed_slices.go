package optargs

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// sliceValue is the generic implementation for all slice typed values.
// It stores a pointer to the destination slice and the element type
// for Convert() delegation.
type sliceValue struct {
	p        any // pointer to destination slice
	elemType reflect.Type
	typeName string
}

func (v *sliceValue) Set(s string) error {
	parts := strings.Split(s, ",")
	pp := reflect.ValueOf(v.p).Elem()
	dest := pp
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		converted, err := Convert(part, v.elemType)
		if err != nil {
			return err
		}
		dest = reflect.Append(dest, reflect.ValueOf(converted))
	}
	pp.Set(dest)
	return nil
}

func (v *sliceValue) String() string {
	dest := reflect.ValueOf(v.p).Elem()
	if dest.Len() == 0 {
		return "[]"
	}
	parts := make([]string, dest.Len())
	for i := range parts {
		parts[i] = fmt.Sprintf("%v", dest.Index(i).Interface())
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (v *sliceValue) Type() string { return v.typeName }

// Reset clears the slice to its zero value (empty slice).
func (v *sliceValue) Reset() { reflect.ValueOf(v.p).Elem().SetLen(0) }

// Append parses a single element string and appends it to the slice.
func (v *sliceValue) Append(s string) error {
	converted, err := Convert(s, v.elemType)
	if err != nil {
		return err
	}
	pp := reflect.ValueOf(v.p).Elem()
	pp.Set(reflect.Append(pp, reflect.ValueOf(converted)))
	return nil
}

// Replace clears the slice and sets it to the parsed elements.
func (v *sliceValue) Replace(ss []string) error {
	pp := reflect.ValueOf(v.p).Elem()
	dest := reflect.MakeSlice(pp.Type(), 0, len(ss))
	for _, s := range ss {
		converted, err := Convert(s, v.elemType)
		if err != nil {
			return err
		}
		dest = reflect.Append(dest, reflect.ValueOf(converted))
	}
	pp.Set(dest)
	return nil
}

// GetSlice returns the string representation of each element.
func (v *sliceValue) GetSlice() []string {
	pp := reflect.ValueOf(v.p).Elem()
	out := make([]string, pp.Len())
	for i := range out {
		out[i] = fmt.Sprintf("%v", pp.Index(i).Interface())
	}
	return out
}

// Slice constructors.

// NewStringSliceValue returns a TypedValue backed by *p, initialized to val.
func NewStringSliceValue(val []string, p *[]string) TypedValue {
	if p == nil {
		p = new([]string)
	}
	*p = val
	return &sliceValue{p: p, elemType: stringType, typeName: "stringSlice"}
}

// NewBoolSliceValue returns a TypedValue backed by *p, initialized to val.
func NewBoolSliceValue(val []bool, p *[]bool) TypedValue {
	if p == nil {
		p = new([]bool)
	}
	*p = val
	return &sliceValue{p: p, elemType: boolType, typeName: "boolSlice"}
}

// NewIntSliceValue returns a TypedValue backed by *p, initialized to val.
func NewIntSliceValue(val []int, p *[]int) TypedValue {
	if p == nil {
		p = new([]int)
	}
	*p = val
	return &sliceValue{p: p, elemType: intType, typeName: "intSlice"}
}

// NewInt32SliceValue returns a TypedValue backed by *p, initialized to val.
func NewInt32SliceValue(val []int32, p *[]int32) TypedValue {
	if p == nil {
		p = new([]int32)
	}
	*p = val
	return &sliceValue{p: p, elemType: int32Type, typeName: "int32Slice"}
}

// NewInt64SliceValue returns a TypedValue backed by *p, initialized to val.
func NewInt64SliceValue(val []int64, p *[]int64) TypedValue {
	if p == nil {
		p = new([]int64)
	}
	*p = val
	return &sliceValue{p: p, elemType: int64Type, typeName: "int64Slice"}
}

// NewUintSliceValue returns a TypedValue backed by *p, initialized to val.
func NewUintSliceValue(val []uint, p *[]uint) TypedValue {
	if p == nil {
		p = new([]uint)
	}
	*p = val
	return &sliceValue{p: p, elemType: uintType, typeName: "uintSlice"}
}

// NewFloat32SliceValue returns a TypedValue backed by *p, initialized to val.
func NewFloat32SliceValue(val []float32, p *[]float32) TypedValue {
	if p == nil {
		p = new([]float32)
	}
	*p = val
	return &sliceValue{p: p, elemType: float32Type, typeName: "float32Slice"}
}

// NewFloat64SliceValue returns a TypedValue backed by *p, initialized to val.
func NewFloat64SliceValue(val []float64, p *[]float64) TypedValue {
	if p == nil {
		p = new([]float64)
	}
	*p = val
	return &sliceValue{p: p, elemType: float64Type, typeName: "float64Slice"}
}

// durationSliceValue is a dedicated type because time.Duration is int64
// under the hood, and Convert() dispatches on kind (int64), not named type.
type durationSliceValue struct{ p *[]time.Duration }

// NewDurationSliceValue returns a TypedValue backed by *p, initialized to val.
func NewDurationSliceValue(val []time.Duration, p *[]time.Duration) TypedValue {
	if p == nil {
		p = new([]time.Duration)
	}
	*p = val
	return &durationSliceValue{p: p}
}

func (v *durationSliceValue) Set(s string) error {
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		d, err := time.ParseDuration(part)
		if err != nil {
			return fmt.Errorf("invalid value %q for type duration", part)
		}
		*v.p = append(*v.p, d)
	}
	return nil
}

func (v *durationSliceValue) String() string {
	if len(*v.p) == 0 {
		return "[]"
	}
	parts := make([]string, len(*v.p))
	for i, d := range *v.p {
		parts[i] = d.String()
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (v *durationSliceValue) Type() string { return "durationSlice" }

// Reset clears the duration slice to its zero value (empty slice).
func (v *durationSliceValue) Reset() { *v.p = (*v.p)[:0] }

// Append parses a single duration string and appends it to the slice.
func (v *durationSliceValue) Append(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid value %q for type duration", s)
	}
	*v.p = append(*v.p, d)
	return nil
}

// Replace clears the slice and sets it to the parsed duration elements.
func (v *durationSliceValue) Replace(ss []string) error {
	out := make([]time.Duration, 0, len(ss))
	for _, s := range ss {
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("invalid value %q for type duration", s)
		}
		out = append(out, d)
	}
	*v.p = out
	return nil
}

// GetSlice returns the string representation of each duration element.
func (v *durationSliceValue) GetSlice() []string {
	out := make([]string, len(*v.p))
	for i, d := range *v.p {
		out[i] = d.String()
	}
	return out
}

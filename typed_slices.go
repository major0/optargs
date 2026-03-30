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
	p        interface{} // pointer to destination slice
	elemType reflect.Type
	typeName string
}

func (v *sliceValue) Set(s string) error {
	parts := strings.Split(s, ",")
	dest := reflect.ValueOf(v.p).Elem()
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
	reflect.ValueOf(v.p).Elem().Set(dest)
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

// --- Slice constructors ---

func NewStringSliceValue(val []string, p *[]string) TypedValue {
	if p == nil {
		p = new([]string)
	}
	*p = val
	return &sliceValue{p: p, elemType: stringType, typeName: "stringSlice"}
}

func NewBoolSliceValue(val []bool, p *[]bool) TypedValue {
	if p == nil {
		p = new([]bool)
	}
	*p = val
	return &sliceValue{p: p, elemType: boolType, typeName: "boolSlice"}
}

func NewIntSliceValue(val []int, p *[]int) TypedValue {
	if p == nil {
		p = new([]int)
	}
	*p = val
	return &sliceValue{p: p, elemType: intType, typeName: "intSlice"}
}

func NewInt32SliceValue(val []int32, p *[]int32) TypedValue {
	if p == nil {
		p = new([]int32)
	}
	*p = val
	return &sliceValue{p: p, elemType: int32Type, typeName: "int32Slice"}
}

func NewInt64SliceValue(val []int64, p *[]int64) TypedValue {
	if p == nil {
		p = new([]int64)
	}
	*p = val
	return &sliceValue{p: p, elemType: int64Type, typeName: "int64Slice"}
}

func NewUintSliceValue(val []uint, p *[]uint) TypedValue {
	if p == nil {
		p = new([]uint)
	}
	*p = val
	return &sliceValue{p: p, elemType: uintType, typeName: "uintSlice"}
}

func NewFloat32SliceValue(val []float32, p *[]float32) TypedValue {
	if p == nil {
		p = new([]float32)
	}
	*p = val
	return &sliceValue{p: p, elemType: float32Type, typeName: "float32Slice"}
}

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

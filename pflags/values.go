package pflags

import (
	"time"

	"github.com/major0/optargs"
)

// All value types delegate to OptArgs Core TypedValue constructors.
// The pflags Value interface (String, Set, Type) is identical to
// optargs.TypedValue, so core constructors satisfy it directly.

func newStringValue(val string, p *string) Value          { return optargs.NewStringValue(val, p) }
func newBoolValue(val bool, p *bool) Value                { return optargs.NewBoolValue(val, p) }
func newIntValue(val int, p *int) Value                   { return optargs.NewIntValue(val, p) }
func newInt8Value(val int8, p *int8) Value                { return optargs.NewInt8Value(val, p) }
func newInt16Value(val int16, p *int16) Value             { return optargs.NewInt16Value(val, p) }
func newInt32Value(val int32, p *int32) Value             { return optargs.NewInt32Value(val, p) }
func newInt64Value(val int64, p *int64) Value             { return optargs.NewInt64Value(val, p) }
func newUintValue(val uint, p *uint) Value                { return optargs.NewUintValue(val, p) }
func newUint8Value(val uint8, p *uint8) Value             { return optargs.NewUint8Value(val, p) }
func newUint16Value(val uint16, p *uint16) Value          { return optargs.NewUint16Value(val, p) }
func newUint32Value(val uint32, p *uint32) Value          { return optargs.NewUint32Value(val, p) }
func newUint64Value(val uint64, p *uint64) Value          { return optargs.NewUint64Value(val, p) }
func newFloat32Value(val float32, p *float32) Value       { return optargs.NewFloat32Value(val, p) }
func newFloat64Value(val float64, p *float64) Value       { return optargs.NewFloat64Value(val, p) }
func newDurationValue(val time.Duration, p *time.Duration) Value { return optargs.NewDurationValue(val, p) }

func newStringSliceValue(val []string, p *[]string) Value          { return optargs.NewStringSliceValue(val, p) }
func newBoolSliceValue(val []bool, p *[]bool) Value                { return optargs.NewBoolSliceValue(val, p) }
func newIntSliceValue(val []int, p *[]int) Value                   { return optargs.NewIntSliceValue(val, p) }
func newInt32SliceValue(val []int32, p *[]int32) Value             { return optargs.NewInt32SliceValue(val, p) }
func newInt64SliceValue(val []int64, p *[]int64) Value             { return optargs.NewInt64SliceValue(val, p) }
func newUintSliceValue(val []uint, p *[]uint) Value                { return optargs.NewUintSliceValue(val, p) }
func newFloat32SliceValue(val []float32, p *[]float32) Value       { return optargs.NewFloat32SliceValue(val, p) }
func newFloat64SliceValue(val []float64, p *[]float64) Value       { return optargs.NewFloat64SliceValue(val, p) }
func newDurationSliceValue(val []time.Duration, p *[]time.Duration) Value { return optargs.NewDurationSliceValue(val, p) }

// String collection and map types.
func newStringArrayValue(val []string, p *[]string) Value                  { return optargs.NewStringArrayValue(val, p) }
func newStringToStringValue(val map[string]string, p *map[string]string) Value { return optargs.NewStringToStringValue(val, p) }
func newStringToIntValue(val map[string]int, p *map[string]int) Value          { return optargs.NewStringToIntValue(val, p) }
func newStringToInt64Value(val map[string]int64, p *map[string]int64) Value    { return optargs.NewStringToInt64Value(val, p) }

// Special types.
func newCountValue(val int, p *int) Value              { return optargs.NewCountValue(val, p) }
func newFuncValue(fn func(string) error) Value         { return optargs.NewFuncValue(fn) }
func newBoolFuncValue(fn func(string) error) Value     { return optargs.NewBoolFuncValue(fn) }

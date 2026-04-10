package pflag

import (
	"encoding"
	"net"
	"time"
)

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (f *FlagSet) StringVar(p *string, name string, value string, usage string) {
	f.VarP(newStringValue(value, p), name, "", usage)
}

// StringVarP is like StringVar, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) StringVarP(p *string, name, shorthand string, value string, usage string) {
	f.VarP(newStringValue(value, p), name, shorthand, usage)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (f *FlagSet) String(name string, value string, usage string) *string {
	p := new(string)
	f.StringVarP(p, name, "", value, usage)
	return p
}

// StringP is like String, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) StringP(name, shorthand string, value string, usage string) *string {
	p := new(string)
	f.StringVarP(p, name, shorthand, value, usage)
	return p
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string) {
	f.VarP(newBoolValue(value, p), name, "", usage)
}

// BoolVarP is like BoolVar, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) BoolVarP(p *bool, name, shorthand string, value bool, usage string) {
	f.VarP(newBoolValue(value, p), name, shorthand, usage)
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func (f *FlagSet) Bool(name string, value bool, usage string) *bool {
	p := new(bool)
	f.BoolVarP(p, name, "", value, usage)
	return p
}

// BoolP is like Bool, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) BoolP(name, shorthand string, value bool, usage string) *bool {
	p := new(bool)
	f.BoolVarP(p, name, shorthand, value, usage)
	return p
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (f *FlagSet) IntVar(p *int, name string, value int, usage string) {
	f.VarP(newIntValue(value, p), name, "", usage)
}

// IntVarP is like IntVar, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) IntVarP(p *int, name, shorthand string, value int, usage string) {
	f.VarP(newIntValue(value, p), name, shorthand, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (f *FlagSet) Int(name string, value int, usage string) *int {
	p := new(int)
	f.IntVarP(p, name, "", value, usage)
	return p
}

// IntP is like Int, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) IntP(name, shorthand string, value int, usage string) *int {
	p := new(int)
	f.IntVarP(p, name, shorthand, value, usage)
	return p
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func (f *FlagSet) Int64Var(p *int64, name string, value int64, usage string) {
	f.VarP(newInt64Value(value, p), name, "", usage)
}

// Int64VarP is like Int64Var, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) Int64VarP(p *int64, name, shorthand string, value int64, usage string) {
	f.VarP(newInt64Value(value, p), name, shorthand, usage)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func (f *FlagSet) Int64(name string, value int64, usage string) *int64 {
	p := new(int64)
	f.Int64VarP(p, name, "", value, usage)
	return p
}

// Int64P is like Int64, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) Int64P(name, shorthand string, value int64, usage string) *int64 {
	p := new(int64)
	f.Int64VarP(p, name, shorthand, value, usage)
	return p
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (f *FlagSet) UintVar(p *uint, name string, value uint, usage string) {
	f.VarP(newUintValue(value, p), name, "", usage)
}

// UintVarP is like UintVar, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) UintVarP(p *uint, name, shorthand string, value uint, usage string) {
	f.VarP(newUintValue(value, p), name, shorthand, usage)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func (f *FlagSet) Uint(name string, value uint, usage string) *uint {
	p := new(uint)
	f.UintVarP(p, name, "", value, usage)
	return p
}

// UintP is like Uint, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) UintP(name, shorthand string, value uint, usage string) *uint {
	p := new(uint)
	f.UintVarP(p, name, shorthand, value, usage)
	return p
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func (f *FlagSet) Uint64Var(p *uint64, name string, value uint64, usage string) {
	f.VarP(newUint64Value(value, p), name, "", usage)
}

// Uint64VarP is like Uint64Var, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) Uint64VarP(p *uint64, name, shorthand string, value uint64, usage string) {
	f.VarP(newUint64Value(value, p), name, shorthand, usage)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func (f *FlagSet) Uint64(name string, value uint64, usage string) *uint64 {
	p := new(uint64)
	f.Uint64VarP(p, name, "", value, usage)
	return p
}

// Uint64P is like Uint64, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) Uint64P(name, shorthand string, value uint64, usage string) *uint64 {
	p := new(uint64)
	f.Uint64VarP(p, name, shorthand, value, usage)
	return p
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (f *FlagSet) Float64Var(p *float64, name string, value float64, usage string) {
	f.VarP(newFloat64Value(value, p), name, "", usage)
}

// Float64VarP is like Float64Var, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) Float64VarP(p *float64, name, shorthand string, value float64, usage string) {
	f.VarP(newFloat64Value(value, p), name, shorthand, usage)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (f *FlagSet) Float64(name string, value float64, usage string) *float64 {
	p := new(float64)
	f.Float64VarP(p, name, "", value, usage)
	return p
}

// Float64P is like Float64, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) Float64P(name, shorthand string, value float64, usage string) *float64 {
	p := new(float64)
	f.Float64VarP(p, name, shorthand, value, usage)
	return p
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func (f *FlagSet) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	f.VarP(newDurationValue(value, p), name, "", usage)
}

// DurationVarP is like DurationVar, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) DurationVarP(p *time.Duration, name, shorthand string, value time.Duration, usage string) {
	f.VarP(newDurationValue(value, p), name, shorthand, usage)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func (f *FlagSet) Duration(name string, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	f.DurationVarP(p, name, "", value, usage)
	return p
}

// DurationP is like Duration, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) DurationP(name, shorthand string, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	f.DurationVarP(p, name, shorthand, value, usage)
	return p
}

// StringSliceVar defines a string slice flag with specified name, default value, and usage string.
// The argument p points to a []string variable in which to store the value of the flag.
func (f *FlagSet) StringSliceVar(p *[]string, name string, value []string, usage string) {
	f.VarP(newStringSliceValue(value, p), name, "", usage)
}

// StringSliceVarP is like StringSliceVar, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) StringSliceVarP(p *[]string, name, shorthand string, value []string, usage string) {
	f.VarP(newStringSliceValue(value, p), name, shorthand, usage)
}

// StringSlice defines a string slice flag with specified name, default value, and usage string.
// The return value is the address of a []string variable that stores the value of the flag.
func (f *FlagSet) StringSlice(name string, value []string, usage string) *[]string {
	p := new([]string)
	f.StringSliceVarP(p, name, "", value, usage)
	return p
}

// StringSliceP is like StringSlice, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) StringSliceP(name, shorthand string, value []string, usage string) *[]string {
	p := new([]string)
	f.StringSliceVarP(p, name, shorthand, value, usage)
	return p
}

// IntSliceVar defines an int slice flag with specified name, default value, and usage string.
// The argument p points to a []int variable in which to store the value of the flag.
func (f *FlagSet) IntSliceVar(p *[]int, name string, value []int, usage string) {
	f.VarP(newIntSliceValue(value, p), name, "", usage)
}

// IntSliceVarP is like IntSliceVar, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) IntSliceVarP(p *[]int, name, shorthand string, value []int, usage string) {
	f.VarP(newIntSliceValue(value, p), name, shorthand, usage)
}

// IntSlice defines an int slice flag with specified name, default value, and usage string.
// The return value is the address of a []int variable that stores the value of the flag.
func (f *FlagSet) IntSlice(name string, value []int, usage string) *[]int {
	p := new([]int)
	f.IntSliceVarP(p, name, "", value, usage)
	return p
}

// IntSliceP is like IntSlice, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) IntSliceP(name, shorthand string, value []int, usage string) *[]int {
	p := new([]int)
	f.IntSliceVarP(p, name, shorthand, value, usage)
	return p
}

// Func defines a flag with the specified name and usage string.
// The callback function fn will be called every time the flag is encountered
// during parsing, with the flag's value as an argument.
func (f *FlagSet) Func(name string, usage string, fn func(string) error) {
	f.VarP(newFuncValue(fn), name, "", usage)
}

// FuncP is like Func, but accepts a shorthand letter.
func (f *FlagSet) FuncP(name, shorthand string, usage string, fn func(string) error) {
	f.VarP(newFuncValue(fn), name, shorthand, usage)
}

// BoolFunc defines a boolean callback flag. The callback is called every time
// the flag is encountered. When used without a value (--flag), the callback
// receives an empty string. When used with a value (--flag=value), the callback
// receives the value.
func (f *FlagSet) BoolFunc(name string, usage string, fn func(string) error) {
	f.VarP(newBoolFuncValue(fn), name, "", usage)
}

// BoolFuncP is like BoolFunc, but accepts a shorthand letter.
func (f *FlagSet) BoolFuncP(name, shorthand string, usage string, fn func(string) error) {
	f.VarP(newBoolFuncValue(fn), name, shorthand, usage)
}

// -- Int8.

func (f *FlagSet) Int8Var(p *int8, name string, value int8, usage string) {
	f.VarP(newInt8Value(value, p), name, "", usage)
}

func (f *FlagSet) Int8VarP(p *int8, name, shorthand string, value int8, usage string) {
	f.VarP(newInt8Value(value, p), name, shorthand, usage)
}

func (f *FlagSet) Int8(name string, value int8, usage string) *int8 {
	p := new(int8)
	f.Int8VarP(p, name, "", value, usage)
	return p
}

func (f *FlagSet) Int8P(name, shorthand string, value int8, usage string) *int8 {
	p := new(int8)
	f.Int8VarP(p, name, shorthand, value, usage)
	return p
}

// -- Int16.

func (f *FlagSet) Int16Var(p *int16, name string, value int16, usage string) {
	f.VarP(newInt16Value(value, p), name, "", usage)
}

func (f *FlagSet) Int16VarP(p *int16, name, shorthand string, value int16, usage string) {
	f.VarP(newInt16Value(value, p), name, shorthand, usage)
}

func (f *FlagSet) Int16(name string, value int16, usage string) *int16 {
	p := new(int16)
	f.Int16VarP(p, name, "", value, usage)
	return p
}

func (f *FlagSet) Int16P(name, shorthand string, value int16, usage string) *int16 {
	p := new(int16)
	f.Int16VarP(p, name, shorthand, value, usage)
	return p
}

// -- Int32.

func (f *FlagSet) Int32Var(p *int32, name string, value int32, usage string) {
	f.VarP(newInt32Value(value, p), name, "", usage)
}

func (f *FlagSet) Int32VarP(p *int32, name, shorthand string, value int32, usage string) {
	f.VarP(newInt32Value(value, p), name, shorthand, usage)
}

func (f *FlagSet) Int32(name string, value int32, usage string) *int32 {
	p := new(int32)
	f.Int32VarP(p, name, "", value, usage)
	return p
}

func (f *FlagSet) Int32P(name, shorthand string, value int32, usage string) *int32 {
	p := new(int32)
	f.Int32VarP(p, name, shorthand, value, usage)
	return p
}

// -- Uint8.

func (f *FlagSet) Uint8Var(p *uint8, name string, value uint8, usage string) {
	f.VarP(newUint8Value(value, p), name, "", usage)
}

func (f *FlagSet) Uint8VarP(p *uint8, name, shorthand string, value uint8, usage string) {
	f.VarP(newUint8Value(value, p), name, shorthand, usage)
}

func (f *FlagSet) Uint8(name string, value uint8, usage string) *uint8 {
	p := new(uint8)
	f.Uint8VarP(p, name, "", value, usage)
	return p
}

func (f *FlagSet) Uint8P(name, shorthand string, value uint8, usage string) *uint8 {
	p := new(uint8)
	f.Uint8VarP(p, name, shorthand, value, usage)
	return p
}

// -- Uint16.

func (f *FlagSet) Uint16Var(p *uint16, name string, value uint16, usage string) {
	f.VarP(newUint16Value(value, p), name, "", usage)
}

func (f *FlagSet) Uint16VarP(p *uint16, name, shorthand string, value uint16, usage string) {
	f.VarP(newUint16Value(value, p), name, shorthand, usage)
}

func (f *FlagSet) Uint16(name string, value uint16, usage string) *uint16 {
	p := new(uint16)
	f.Uint16VarP(p, name, "", value, usage)
	return p
}

func (f *FlagSet) Uint16P(name, shorthand string, value uint16, usage string) *uint16 {
	p := new(uint16)
	f.Uint16VarP(p, name, shorthand, value, usage)
	return p
}

// -- Uint32.

func (f *FlagSet) Uint32Var(p *uint32, name string, value uint32, usage string) {
	f.VarP(newUint32Value(value, p), name, "", usage)
}

func (f *FlagSet) Uint32VarP(p *uint32, name, shorthand string, value uint32, usage string) {
	f.VarP(newUint32Value(value, p), name, shorthand, usage)
}

func (f *FlagSet) Uint32(name string, value uint32, usage string) *uint32 {
	p := new(uint32)
	f.Uint32VarP(p, name, "", value, usage)
	return p
}

func (f *FlagSet) Uint32P(name, shorthand string, value uint32, usage string) *uint32 {
	p := new(uint32)
	f.Uint32VarP(p, name, shorthand, value, usage)
	return p
}

// -- Float32.

func (f *FlagSet) Float32Var(p *float32, name string, value float32, usage string) {
	f.VarP(newFloat32Value(value, p), name, "", usage)
}

func (f *FlagSet) Float32VarP(p *float32, name, shorthand string, value float32, usage string) {
	f.VarP(newFloat32Value(value, p), name, shorthand, usage)
}

func (f *FlagSet) Float32(name string, value float32, usage string) *float32 {
	p := new(float32)
	f.Float32VarP(p, name, "", value, usage)
	return p
}

func (f *FlagSet) Float32P(name, shorthand string, value float32, usage string) *float32 {
	p := new(float32)
	f.Float32VarP(p, name, shorthand, value, usage)
	return p
}

// -- BoolSlice.

func (f *FlagSet) BoolSliceVar(p *[]bool, name string, value []bool, usage string) {
	f.VarP(newBoolSliceValue(value, p), name, "", usage)
}
func (f *FlagSet) BoolSliceVarP(p *[]bool, name, shorthand string, value []bool, usage string) {
	f.VarP(newBoolSliceValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) BoolSlice(name string, value []bool, usage string) *[]bool {
	p := new([]bool)
	f.BoolSliceVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) BoolSliceP(name, shorthand string, value []bool, usage string) *[]bool {
	p := new([]bool)
	f.BoolSliceVarP(p, name, shorthand, value, usage)
	return p
}

// -- Int32Slice.

func (f *FlagSet) Int32SliceVar(p *[]int32, name string, value []int32, usage string) {
	f.VarP(newInt32SliceValue(value, p), name, "", usage)
}
func (f *FlagSet) Int32SliceVarP(p *[]int32, name, shorthand string, value []int32, usage string) {
	f.VarP(newInt32SliceValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) Int32Slice(name string, value []int32, usage string) *[]int32 {
	p := new([]int32)
	f.Int32SliceVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) Int32SliceP(name, shorthand string, value []int32, usage string) *[]int32 {
	p := new([]int32)
	f.Int32SliceVarP(p, name, shorthand, value, usage)
	return p
}

// -- Int64Slice.

func (f *FlagSet) Int64SliceVar(p *[]int64, name string, value []int64, usage string) {
	f.VarP(newInt64SliceValue(value, p), name, "", usage)
}
func (f *FlagSet) Int64SliceVarP(p *[]int64, name, shorthand string, value []int64, usage string) {
	f.VarP(newInt64SliceValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) Int64Slice(name string, value []int64, usage string) *[]int64 {
	p := new([]int64)
	f.Int64SliceVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) Int64SliceP(name, shorthand string, value []int64, usage string) *[]int64 {
	p := new([]int64)
	f.Int64SliceVarP(p, name, shorthand, value, usage)
	return p
}

// -- UintSlice.

func (f *FlagSet) UintSliceVar(p *[]uint, name string, value []uint, usage string) {
	f.VarP(newUintSliceValue(value, p), name, "", usage)
}
func (f *FlagSet) UintSliceVarP(p *[]uint, name, shorthand string, value []uint, usage string) {
	f.VarP(newUintSliceValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) UintSlice(name string, value []uint, usage string) *[]uint {
	p := new([]uint)
	f.UintSliceVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) UintSliceP(name, shorthand string, value []uint, usage string) *[]uint {
	p := new([]uint)
	f.UintSliceVarP(p, name, shorthand, value, usage)
	return p
}

// -- Float32Slice.

func (f *FlagSet) Float32SliceVar(p *[]float32, name string, value []float32, usage string) {
	f.VarP(newFloat32SliceValue(value, p), name, "", usage)
}
func (f *FlagSet) Float32SliceVarP(p *[]float32, name, shorthand string, value []float32, usage string) {
	f.VarP(newFloat32SliceValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) Float32Slice(name string, value []float32, usage string) *[]float32 {
	p := new([]float32)
	f.Float32SliceVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) Float32SliceP(name, shorthand string, value []float32, usage string) *[]float32 {
	p := new([]float32)
	f.Float32SliceVarP(p, name, shorthand, value, usage)
	return p
}

// -- Float64Slice.

func (f *FlagSet) Float64SliceVar(p *[]float64, name string, value []float64, usage string) {
	f.VarP(newFloat64SliceValue(value, p), name, "", usage)
}
func (f *FlagSet) Float64SliceVarP(p *[]float64, name, shorthand string, value []float64, usage string) {
	f.VarP(newFloat64SliceValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) Float64Slice(name string, value []float64, usage string) *[]float64 {
	p := new([]float64)
	f.Float64SliceVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) Float64SliceP(name, shorthand string, value []float64, usage string) *[]float64 {
	p := new([]float64)
	f.Float64SliceVarP(p, name, shorthand, value, usage)
	return p
}

// -- DurationSlice.

func (f *FlagSet) DurationSliceVar(p *[]time.Duration, name string, value []time.Duration, usage string) {
	f.VarP(newDurationSliceValue(value, p), name, "", usage)
}
func (f *FlagSet) DurationSliceVarP(p *[]time.Duration, name, shorthand string, value []time.Duration, usage string) {
	f.VarP(newDurationSliceValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) DurationSlice(name string, value []time.Duration, usage string) *[]time.Duration {
	p := new([]time.Duration)
	f.DurationSliceVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) DurationSliceP(name, shorthand string, value []time.Duration, usage string) *[]time.Duration {
	p := new([]time.Duration)
	f.DurationSliceVarP(p, name, shorthand, value, usage)
	return p
}

// -- StringArray (appends raw strings without comma splitting).

func (f *FlagSet) StringArrayVar(p *[]string, name string, value []string, usage string) {
	f.VarP(newStringArrayValue(value, p), name, "", usage)
}
func (f *FlagSet) StringArrayVarP(p *[]string, name, shorthand string, value []string, usage string) {
	f.VarP(newStringArrayValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) StringArray(name string, value []string, usage string) *[]string {
	p := new([]string)
	f.StringArrayVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) StringArrayP(name, shorthand string, value []string, usage string) *[]string {
	p := new([]string)
	f.StringArrayVarP(p, name, shorthand, value, usage)
	return p
}

// -- StringToString (map[string]string).

func (f *FlagSet) StringToStringVar(p *map[string]string, name string, value map[string]string, usage string) {
	f.VarP(newStringToStringValue(value, p), name, "", usage)
}
func (f *FlagSet) StringToStringVarP(p *map[string]string, name, shorthand string, value map[string]string, usage string) {
	f.VarP(newStringToStringValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) StringToString(name string, value map[string]string, usage string) *map[string]string {
	p := new(map[string]string)
	f.StringToStringVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) StringToStringP(name, shorthand string, value map[string]string, usage string) *map[string]string {
	p := new(map[string]string)
	f.StringToStringVarP(p, name, shorthand, value, usage)
	return p
}

// -- StringToInt (map[string]int).

func (f *FlagSet) StringToIntVar(p *map[string]int, name string, value map[string]int, usage string) {
	f.VarP(newStringToIntValue(value, p), name, "", usage)
}
func (f *FlagSet) StringToIntVarP(p *map[string]int, name, shorthand string, value map[string]int, usage string) {
	f.VarP(newStringToIntValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) StringToInt(name string, value map[string]int, usage string) *map[string]int {
	p := new(map[string]int)
	f.StringToIntVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) StringToIntP(name, shorthand string, value map[string]int, usage string) *map[string]int {
	p := new(map[string]int)
	f.StringToIntVarP(p, name, shorthand, value, usage)
	return p
}

// -- StringToInt64 (map[string]int64).

func (f *FlagSet) StringToInt64Var(p *map[string]int64, name string, value map[string]int64, usage string) {
	f.VarP(newStringToInt64Value(value, p), name, "", usage)
}
func (f *FlagSet) StringToInt64VarP(p *map[string]int64, name, shorthand string, value map[string]int64, usage string) {
	f.VarP(newStringToInt64Value(value, p), name, shorthand, usage)
}
func (f *FlagSet) StringToInt64(name string, value map[string]int64, usage string) *map[string]int64 {
	p := new(map[string]int64)
	f.StringToInt64VarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) StringToInt64P(name, shorthand string, value map[string]int64, usage string) *map[string]int64 {
	p := new(map[string]int64)
	f.StringToInt64VarP(p, name, shorthand, value, usage)
	return p
}

// -- Count (increments on each occurrence, e.g. -vvv).

func (f *FlagSet) CountVar(p *int, name string, usage string) {
	f.VarP(newCountValue(0, p), name, "", usage)
}
func (f *FlagSet) CountVarP(p *int, name, shorthand string, usage string) {
	f.VarP(newCountValue(0, p), name, shorthand, usage)
}
func (f *FlagSet) Count(name string, usage string) *int {
	p := new(int)
	f.CountVarP(p, name, "", usage)
	return p
}
func (f *FlagSet) CountP(name, shorthand string, usage string) *int {
	p := new(int)
	f.CountVarP(p, name, shorthand, usage)
	return p
}

// -- TextVar (encoding.TextUnmarshaler).

func (f *FlagSet) TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) {
	f.VarP(newTextValue(value, p), name, "", usage)
}
func (f *FlagSet) TextVarP(p encoding.TextUnmarshaler, name, shorthand string, value encoding.TextMarshaler, usage string) {
	f.VarP(newTextValue(value, p), name, shorthand, usage)
}

// -- IP.

func (f *FlagSet) IPVar(p *net.IP, name string, value net.IP, usage string) {
	f.VarP(newIPValue(value, p), name, "", usage)
}
func (f *FlagSet) IPVarP(p *net.IP, name, shorthand string, value net.IP, usage string) {
	f.VarP(newIPValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) IP(name string, value net.IP, usage string) *net.IP {
	p := new(net.IP)
	f.IPVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) IPP(name, shorthand string, value net.IP, usage string) *net.IP {
	p := new(net.IP)
	f.IPVarP(p, name, shorthand, value, usage)
	return p
}

// -- IPMask.

func (f *FlagSet) IPMaskVar(p *net.IPMask, name string, value net.IPMask, usage string) {
	f.VarP(newIPMaskValue(value, p), name, "", usage)
}
func (f *FlagSet) IPMaskVarP(p *net.IPMask, name, shorthand string, value net.IPMask, usage string) {
	f.VarP(newIPMaskValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) IPMask(name string, value net.IPMask, usage string) *net.IPMask {
	p := new(net.IPMask)
	f.IPMaskVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) IPMaskP(name, shorthand string, value net.IPMask, usage string) *net.IPMask {
	p := new(net.IPMask)
	f.IPMaskVarP(p, name, shorthand, value, usage)
	return p
}

// -- IPNet.

func (f *FlagSet) IPNetVar(p *net.IPNet, name string, value net.IPNet, usage string) {
	f.VarP(newIPNetValue(value, p), name, "", usage)
}
func (f *FlagSet) IPNetVarP(p *net.IPNet, name, shorthand string, value net.IPNet, usage string) {
	f.VarP(newIPNetValue(value, p), name, shorthand, usage)
}
func (f *FlagSet) IPNet(name string, value net.IPNet, usage string) *net.IPNet {
	p := new(net.IPNet)
	f.IPNetVarP(p, name, "", value, usage)
	return p
}
func (f *FlagSet) IPNetP(name, shorthand string, value net.IPNet, usage string) *net.IPNet {
	p := new(net.IPNet)
	f.IPNetVarP(p, name, shorthand, value, usage)
	return p
}

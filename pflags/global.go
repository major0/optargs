package pflags

import (
	"encoding"
	"flag"
	"net"
	"time"
)

// Global wrappers for FlagSet methods that operate on CommandLine.
// These provide the package-level API matching spf13/pflag.

// --- Scalar types (Int64, Uint, Uint64 not in pflags.go) ---

func Int64Var(p *int64, name string, value int64, usage string)                          { CommandLine.Int64Var(p, name, value, usage) }
func Int64VarP(p *int64, name, shorthand string, value int64, usage string)              { CommandLine.Int64VarP(p, name, shorthand, value, usage) }
func Int64(name string, value int64, usage string) *int64                                { return CommandLine.Int64(name, value, usage) }
func Int64P(name, shorthand string, value int64, usage string) *int64                    { return CommandLine.Int64P(name, shorthand, value, usage) }
func UintVar(p *uint, name string, value uint, usage string)                             { CommandLine.UintVar(p, name, value, usage) }
func UintVarP(p *uint, name, shorthand string, value uint, usage string)                 { CommandLine.UintVarP(p, name, shorthand, value, usage) }
func Uint(name string, value uint, usage string) *uint                                   { return CommandLine.Uint(name, value, usage) }
func UintP(name, shorthand string, value uint, usage string) *uint                       { return CommandLine.UintP(name, shorthand, value, usage) }
func Uint64Var(p *uint64, name string, value uint64, usage string)                       { CommandLine.Uint64Var(p, name, value, usage) }
func Uint64VarP(p *uint64, name, shorthand string, value uint64, usage string)           { CommandLine.Uint64VarP(p, name, shorthand, value, usage) }
func Uint64(name string, value uint64, usage string) *uint64                             { return CommandLine.Uint64(name, value, usage) }
func Uint64P(name, shorthand string, value uint64, usage string) *uint64                 { return CommandLine.Uint64P(name, shorthand, value, usage) }

// --- Narrow numeric types ---

func Int8Var(p *int8, name string, value int8, usage string)                             { CommandLine.Int8Var(p, name, value, usage) }
func Int8VarP(p *int8, name, shorthand string, value int8, usage string)                 { CommandLine.Int8VarP(p, name, shorthand, value, usage) }
func Int8(name string, value int8, usage string) *int8                                   { return CommandLine.Int8(name, value, usage) }
func Int8P(name, shorthand string, value int8, usage string) *int8                       { return CommandLine.Int8P(name, shorthand, value, usage) }
func Int16Var(p *int16, name string, value int16, usage string)                          { CommandLine.Int16Var(p, name, value, usage) }
func Int16VarP(p *int16, name, shorthand string, value int16, usage string)              { CommandLine.Int16VarP(p, name, shorthand, value, usage) }
func Int16(name string, value int16, usage string) *int16                                { return CommandLine.Int16(name, value, usage) }
func Int16P(name, shorthand string, value int16, usage string) *int16                    { return CommandLine.Int16P(name, shorthand, value, usage) }
func Int32Var(p *int32, name string, value int32, usage string)                          { CommandLine.Int32Var(p, name, value, usage) }
func Int32VarP(p *int32, name, shorthand string, value int32, usage string)              { CommandLine.Int32VarP(p, name, shorthand, value, usage) }
func Int32(name string, value int32, usage string) *int32                                { return CommandLine.Int32(name, value, usage) }
func Int32P(name, shorthand string, value int32, usage string) *int32                    { return CommandLine.Int32P(name, shorthand, value, usage) }
func Uint8Var(p *uint8, name string, value uint8, usage string)                          { CommandLine.Uint8Var(p, name, value, usage) }
func Uint8VarP(p *uint8, name, shorthand string, value uint8, usage string)              { CommandLine.Uint8VarP(p, name, shorthand, value, usage) }
func Uint8(name string, value uint8, usage string) *uint8                                { return CommandLine.Uint8(name, value, usage) }
func Uint8P(name, shorthand string, value uint8, usage string) *uint8                    { return CommandLine.Uint8P(name, shorthand, value, usage) }
func Uint16Var(p *uint16, name string, value uint16, usage string)                       { CommandLine.Uint16Var(p, name, value, usage) }
func Uint16VarP(p *uint16, name, shorthand string, value uint16, usage string)           { CommandLine.Uint16VarP(p, name, shorthand, value, usage) }
func Uint16(name string, value uint16, usage string) *uint16                             { return CommandLine.Uint16(name, value, usage) }
func Uint16P(name, shorthand string, value uint16, usage string) *uint16                 { return CommandLine.Uint16P(name, shorthand, value, usage) }
func Uint32Var(p *uint32, name string, value uint32, usage string)                       { CommandLine.Uint32Var(p, name, value, usage) }
func Uint32VarP(p *uint32, name, shorthand string, value uint32, usage string)           { CommandLine.Uint32VarP(p, name, shorthand, value, usage) }
func Uint32(name string, value uint32, usage string) *uint32                             { return CommandLine.Uint32(name, value, usage) }
func Uint32P(name, shorthand string, value uint32, usage string) *uint32                 { return CommandLine.Uint32P(name, shorthand, value, usage) }
func Float32Var(p *float32, name string, value float32, usage string)                    { CommandLine.Float32Var(p, name, value, usage) }
func Float32VarP(p *float32, name, shorthand string, value float32, usage string)        { CommandLine.Float32VarP(p, name, shorthand, value, usage) }
func Float32(name string, value float32, usage string) *float32                          { return CommandLine.Float32(name, value, usage) }
func Float32P(name, shorthand string, value float32, usage string) *float32              { return CommandLine.Float32P(name, shorthand, value, usage) }

// --- Slice types ---

func StringSliceVar(p *[]string, name string, value []string, usage string)              { CommandLine.StringSliceVar(p, name, value, usage) }
func StringSliceVarP(p *[]string, name, sh string, value []string, usage string)         { CommandLine.StringSliceVarP(p, name, sh, value, usage) }
func StringSlice(name string, value []string, usage string) *[]string                    { return CommandLine.StringSlice(name, value, usage) }
func StringSliceP(name, sh string, value []string, usage string) *[]string               { return CommandLine.StringSliceP(name, sh, value, usage) }
func IntSliceVar(p *[]int, name string, value []int, usage string)                       { CommandLine.IntSliceVar(p, name, value, usage) }
func IntSliceVarP(p *[]int, name, sh string, value []int, usage string)                  { CommandLine.IntSliceVarP(p, name, sh, value, usage) }
func IntSlice(name string, value []int, usage string) *[]int                             { return CommandLine.IntSlice(name, value, usage) }
func IntSliceP(name, sh string, value []int, usage string) *[]int                        { return CommandLine.IntSliceP(name, sh, value, usage) }
func BoolSliceVar(p *[]bool, name string, value []bool, usage string)                    { CommandLine.BoolSliceVar(p, name, value, usage) }
func BoolSliceVarP(p *[]bool, name, sh string, value []bool, usage string)               { CommandLine.BoolSliceVarP(p, name, sh, value, usage) }
func BoolSlice(name string, value []bool, usage string) *[]bool                          { return CommandLine.BoolSlice(name, value, usage) }
func BoolSliceP(name, sh string, value []bool, usage string) *[]bool                     { return CommandLine.BoolSliceP(name, sh, value, usage) }
func DurationSliceVar(p *[]time.Duration, name string, value []time.Duration, usage string) { CommandLine.DurationSliceVar(p, name, value, usage) }
func DurationSliceVarP(p *[]time.Duration, name, sh string, value []time.Duration, usage string) { CommandLine.DurationSliceVarP(p, name, sh, value, usage) }
func DurationSlice(name string, value []time.Duration, usage string) *[]time.Duration    { return CommandLine.DurationSlice(name, value, usage) }
func DurationSliceP(name, sh string, value []time.Duration, usage string) *[]time.Duration { return CommandLine.DurationSliceP(name, sh, value, usage) }

// --- String collections and maps ---

func StringArrayVar(p *[]string, name string, value []string, usage string)              { CommandLine.StringArrayVar(p, name, value, usage) }
func StringArrayVarP(p *[]string, name, sh string, value []string, usage string)         { CommandLine.StringArrayVarP(p, name, sh, value, usage) }
func StringArray(name string, value []string, usage string) *[]string                    { return CommandLine.StringArray(name, value, usage) }
func StringArrayP(name, sh string, value []string, usage string) *[]string               { return CommandLine.StringArrayP(name, sh, value, usage) }
func StringToStringVar(p *map[string]string, name string, value map[string]string, usage string) { CommandLine.StringToStringVar(p, name, value, usage) }
func StringToStringVarP(p *map[string]string, name, sh string, value map[string]string, usage string) { CommandLine.StringToStringVarP(p, name, sh, value, usage) }
func StringToString(name string, value map[string]string, usage string) *map[string]string { return CommandLine.StringToString(name, value, usage) }
func StringToStringP(name, sh string, value map[string]string, usage string) *map[string]string { return CommandLine.StringToStringP(name, sh, value, usage) }
func StringToIntVar(p *map[string]int, name string, value map[string]int, usage string)  { CommandLine.StringToIntVar(p, name, value, usage) }
func StringToIntVarP(p *map[string]int, name, sh string, value map[string]int, usage string) { CommandLine.StringToIntVarP(p, name, sh, value, usage) }
func StringToInt(name string, value map[string]int, usage string) *map[string]int        { return CommandLine.StringToInt(name, value, usage) }
func StringToIntP(name, sh string, value map[string]int, usage string) *map[string]int   { return CommandLine.StringToIntP(name, sh, value, usage) }
func StringToInt64Var(p *map[string]int64, name string, value map[string]int64, usage string) { CommandLine.StringToInt64Var(p, name, value, usage) }
func StringToInt64VarP(p *map[string]int64, name, sh string, value map[string]int64, usage string) { CommandLine.StringToInt64VarP(p, name, sh, value, usage) }
func StringToInt64(name string, value map[string]int64, usage string) *map[string]int64  { return CommandLine.StringToInt64(name, value, usage) }
func StringToInt64P(name, sh string, value map[string]int64, usage string) *map[string]int64 { return CommandLine.StringToInt64P(name, sh, value, usage) }

// --- Specialized types ---

func CountVar(p *int, name string, usage string)                                         { CommandLine.CountVar(p, name, usage) }
func CountVarP(p *int, name, shorthand string, usage string)                             { CommandLine.CountVarP(p, name, shorthand, usage) }
func Count(name string, usage string) *int                                               { return CommandLine.Count(name, usage) }
func CountP(name, shorthand string, usage string) *int                                   { return CommandLine.CountP(name, shorthand, usage) }
func TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) { CommandLine.TextVar(p, name, value, usage) }
func TextVarP(p encoding.TextUnmarshaler, name, sh string, value encoding.TextMarshaler, usage string) { CommandLine.TextVarP(p, name, sh, value, usage) }
func IPVar(p *net.IP, name string, value net.IP, usage string)                           { CommandLine.IPVar(p, name, value, usage) }
func IPVarP(p *net.IP, name, sh string, value net.IP, usage string)                      { CommandLine.IPVarP(p, name, sh, value, usage) }
func IP(name string, value net.IP, usage string) *net.IP                                 { return CommandLine.IP(name, value, usage) }
func IPP(name, sh string, value net.IP, usage string) *net.IP                            { return CommandLine.IPP(name, sh, value, usage) }
func IPMaskVar(p *net.IPMask, name string, value net.IPMask, usage string)               { CommandLine.IPMaskVar(p, name, value, usage) }
func IPMaskVarP(p *net.IPMask, name, sh string, value net.IPMask, usage string)          { CommandLine.IPMaskVarP(p, name, sh, value, usage) }
func IPMask(name string, value net.IPMask, usage string) *net.IPMask                     { return CommandLine.IPMask(name, value, usage) }
func IPMaskP(name, sh string, value net.IPMask, usage string) *net.IPMask                { return CommandLine.IPMaskP(name, sh, value, usage) }
func IPNetVar(p *net.IPNet, name string, value net.IPNet, usage string)                  { CommandLine.IPNetVar(p, name, value, usage) }
func IPNetVarP(p *net.IPNet, name, sh string, value net.IPNet, usage string)             { CommandLine.IPNetVarP(p, name, sh, value, usage) }
func IPNet(name string, value net.IPNet, usage string) *net.IPNet                        { return CommandLine.IPNet(name, value, usage) }
func IPNetP(name, sh string, value net.IPNet, usage string) *net.IPNet                   { return CommandLine.IPNetP(name, sh, value, usage) }

// --- Callback flags ---

func Func(name string, usage string, fn func(string) error)                              { CommandLine.Func(name, usage, fn) }
func FuncP(name, sh string, usage string, fn func(string) error)                         { CommandLine.FuncP(name, sh, usage, fn) }
func BoolFunc(name string, usage string, fn func(string) error)                          { CommandLine.BoolFunc(name, usage, fn) }
func BoolFuncP(name, sh string, usage string, fn func(string) error)                     { CommandLine.BoolFuncP(name, sh, usage, fn) }

// --- FlagSet management ---

func Changed(name string) bool                                                           { return CommandLine.Changed(name) }
func NFlag() int                                                                         { return CommandLine.NFlag() }
func HasFlags() bool                                                                     { return CommandLine.HasFlags() }
func HasAvailableFlags() bool                                                            { return CommandLine.HasAvailableFlags() }
func ShorthandLookup(name string) *Flag                                                  { return CommandLine.ShorthandLookup(name) }
func VarPF(value Value, name, shorthand, usage string) *Flag                             { return CommandLine.VarPF(value, name, shorthand, usage) }
func ArgsLenAtDash() int                                                                 { return CommandLine.ArgsLenAtDash() }
func FlagUsagesWrapped(cols int) string                                                  { return CommandLine.FlagUsagesWrapped(cols) }
func ParseAll(args []string, fn func(*Flag, string) error) error                         { return CommandLine.ParseAll(args, fn) }
func SetNormalizeFunc(n func(*FlagSet, string) NormalizedName)                            { CommandLine.SetNormalizeFunc(n) }
func SetInterspersed(interspersed bool)                                                  { CommandLine.SetInterspersed(interspersed) }
func MarkDeprecated(name, usageMessage string) error                                     { return CommandLine.MarkDeprecated(name, usageMessage) }
func MarkHidden(name string) error                                                       { return CommandLine.MarkHidden(name) }
func MarkShorthandDeprecated(name, usageMessage string) error                            { return CommandLine.MarkShorthandDeprecated(name, usageMessage) }
func SetAnnotation(name, key string, values []string) error                              { return CommandLine.SetAnnotation(name, key, values) }
func AddFlag(f *Flag)                                                                    { CommandLine.AddFlag(f) }
func AddFlagSet(newSet *FlagSet)                                                         { CommandLine.AddFlagSet(newSet) }
func AddGoFlag(goflag *flag.Flag)                                                        { CommandLine.AddGoFlag(goflag) }
func AddGoFlagSet(goflags *flag.FlagSet)                                                 { CommandLine.AddGoFlagSet(goflags) }

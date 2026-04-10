package optargs

// TypedValue is the core interface for typed flag values.
// It is a superset of pflag's Value interface — any TypedValue
// can be used where a pflag Value is expected.
type TypedValue interface {
	// Set parses the string into the typed value.
	Set(string) error

	// String formats the current value as a string.
	String() string

	// Type returns the type name for help text (e.g., "int", "bool", "duration").
	Type() string
}

// BoolValuer marks a TypedValue as boolean — no argument required.
// Wrappers check this to determine NoArgument vs OptionalArgument
// for the parser.
type BoolValuer interface {
	IsBoolFlag() bool
}

// BoolArgValuer optionally declares whether a bool-like flag accepts
// an optional =value argument (e.g., --flag=true). When not implemented,
// wrappers default to accepting an optional argument for backward
// compatibility. Types that are strictly no-argument (like Count)
// return false.
type BoolArgValuer interface {
	BoolTakesArg() bool
}

// Resetter is implemented by collection TypedValue types (slices, maps)
// that support clearing to their zero value. Wrappers use this to
// implement negatable flag zero-clearing for collection types where
// Set(zeroString) would append rather than replace.
type Resetter interface {
	Reset()
}

// zeroStrings maps type names to their zero-value string representations.
// Used by ZeroString and by pflag's isZeroValue for help text defaults.
var zeroStrings = map[string]string{
	"bool": "false", "duration": "0s", "float64": "0",
	"float32": "0", "int": "0", "int8": "0", "int16": "0",
	"int32": "0", "int64": "0", "string": "",
	"uint": "0", "uint8": "0", "uint16": "0",
	"uint32": "0", "uint64": "0",
	"stringSlice": "[]", "intSlice": "[]", "boolSlice": "[]",
	"int32Slice": "[]", "int64Slice": "[]", "uintSlice": "[]",
	"float32Slice": "[]", "float64Slice": "[]", "durationSlice": "[]",
	"stringArray": "[]", "count": "0",
	"stringToString": "map[]", "stringToInt": "map[]", "stringToInt64": "map[]",
	"bytesHex": "", "bytesBase64": "",
}

// ZeroString returns the zero-value string representation for a TypedValue
// type name, and whether the type is known. Wrappers use this to validate
// negatable flag types and to clear scalar values to zero.
func ZeroString(typeName string) (string, bool) {
	z, ok := zeroStrings[typeName]
	return z, ok
}

// PrefixPair represents a true/false prefix pair for boolean flags.
// Used by wrapper modules to register --enable-X/--disable-X style options.
type PrefixPair struct {
	True  string // e.g. "enable"
	False string // e.g. "disable"
}

// IsBool reports whether a TypedValue represents a boolean flag.
// Checks both Type() == "bool" and the BoolValuer interface.
func IsBool(tv TypedValue) bool {
	if tv.Type() == "bool" {
		return true
	}
	if bv, ok := tv.(BoolValuer); ok {
		return bv.IsBoolFlag()
	}
	return false
}

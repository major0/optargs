package optargs

// TypedValue is the core interface for typed flag values.
// It is a superset of pflags' Value interface — any TypedValue
// can be used where a pflags Value is expected.
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

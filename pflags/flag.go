//go:build ignore

// Package pflags provides a drop-in replacement for spf13/pflag that maintains
// complete API compatibility while leveraging OptArgs Core for superior POSIX/GNU compliance.
package pflags

import (
	"fmt"
	"io"
	"os"
)

// ErrorHandling defines how FlagSet.Parse behaves if the parse fails.
type ErrorHandling int

const (
	// ContinueOnError will return an err from Parse() if an error is found
	ContinueOnError ErrorHandling = iota
	// ExitOnError will call os.Exit(2) if an error is found when parsing
	ExitOnError
	// PanicOnError will panic() if an error is found when parsing flags
	PanicOnError
)

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
type Value interface {
	String() string
	Set(string) error
	Type() string
}

// Flag represents the state of a flag.
type Flag struct {
	Name        string              // name as it appears on command line
	Shorthand   string              // one-letter abbreviated flag
	Usage       string              // help message
	Value       Value               // value as set
	DefValue    string              // default value (as text); for usage message
	Changed     bool                // If the user set the value (or if left to default)
	Hidden      bool                // used by cobra.Command to allow flags to be hidden from help/usage text
	Deprecated  string              // If this flag is deprecated, this string is the new or now thing to use
	Annotations map[string][]string // used by cobra.Command bash autocomple code
}

// FlagSet represents a set of defined flags.
type FlagSet struct {
	// Usage is the function called when an error occurs while parsing flags.
	// The field is a function (not a method) that may be changed to point to
	// a custom error handler.
	Usage func()

	name              string
	parsed            bool
	args              []string              // arguments after flags
	errorHandling     ErrorHandling
	output            io.Writer             // nil means stderr; use out() accessor
	interspersed      bool                  // allow interspersed option/non-option args
	normalizeNameFunc func(f *FlagSet, name string) NormalizedName

	// Flag storage and management
	flags     map[string]*Flag   // flags by name
	shorthand map[string]string  // shorthand to name mapping
	order     []string           // order of flag definition for help text
	
	// OptArgs Core integration
	coreIntegration *CoreIntegration
}

// NormalizedName is a flag name that has been normalized according to rules
// for the FlagSet (e.g. making '-' and '_' equivalent).
type NormalizedName string

// NewFlagSet returns a new, empty flag set with the specified name and
// error handling property.
func NewFlagSet(name string, errorHandling ErrorHandling) *FlagSet {
	f := &FlagSet{
		name:          name,
		errorHandling: errorHandling,
		interspersed:  true,
		flags:         make(map[string]*Flag),
		shorthand:     make(map[string]string),
		order:         make([]string, 0),
	}
	f.Usage = f.defaultUsage
	f.coreIntegration = NewCoreIntegration(f)
	return f
}

// out returns the destination for usage and error messages.
func (f *FlagSet) out() io.Writer {
	if f.output == nil {
		return os.Stderr
	}
	return f.output
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (f *FlagSet) SetOutput(output io.Writer) {
	f.output = output
}

// Name returns the name of the flag set.
func (f *FlagSet) Name() string {
	return f.name
}

// Parsed reports whether f.Parse has been called.
func (f *FlagSet) Parsed() bool {
	return f.parsed
}

// Args returns the non-flag arguments.
func (f *FlagSet) Args() []string {
	return f.args
}

// NArg is the number of arguments remaining after flags have been processed.
func (f *FlagSet) NArg() int {
	return len(f.args)
}

// Arg returns the i'th argument. Arg(0) is the first remaining argument
// after flags have been processed.
func (f *FlagSet) Arg(i int) string {
	if i < 0 || i >= len(f.args) {
		return ""
	}
	return f.args[i]
}

// defaultUsage is the default function to print a usage message.
func (f *FlagSet) defaultUsage() {
	if f.name == "" {
		fmt.Fprintf(f.out(), "Usage:\n")
	} else {
		fmt.Fprintf(f.out(), "Usage of %s:\n", f.name)
	}
	f.PrintDefaults()
}

// PrintDefaults prints, to standard error unless configured otherwise, the
// default values of all defined flags in the set.
func (f *FlagSet) PrintDefaults() {
	f.VisitAll(func(flag *Flag) {
		if flag.Hidden {
			return
		}
		
		format := "  -%s"
		if len(flag.Shorthand) > 0 {
			format = "  -%s, --%s"
		} else {
			format = "      --%s"
		}
		
		if len(flag.Shorthand) > 0 {
			fmt.Fprintf(f.out(), format, flag.Shorthand, flag.Name)
		} else {
			fmt.Fprintf(f.out(), format, flag.Name)
		}
		
		name, usage := UnquoteUsage(flag)
		if len(name) > 0 {
			fmt.Fprintf(f.out(), " %s", name)
		}
		
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if len(usage) > 0 {
			fmt.Fprintf(f.out(), "\t%s", usage)
		}
		
		if !isZeroValue(flag, flag.DefValue) {
			if flag.Value.Type() == "string" {
				fmt.Fprintf(f.out(), " (default %q)", flag.DefValue)
			} else {
				fmt.Fprintf(f.out(), " (default %s)", flag.DefValue)
			}
		}
		fmt.Fprint(f.out(), "\n")
	})
}

// isZeroValue determines whether the string represents the zero
// value for a flag.
func isZeroValue(flag *Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := flag.Value.Type()
	var z Value
	switch typ {
	case "bool":
		z = newBoolValue(false, nil)
	case "duration":
		z = newDurationValue(0, nil)
	case "float64":
		z = newFloat64Value(0, nil)
	case "int":
		z = newIntValue(0, nil)
	case "int64":
		z = newInt64Value(0, nil)
	case "string":
		z = newStringValue("", nil)
	case "uint":
		z = newUintValue(0, nil)
	case "uint64":
		z = newUint64Value(0, nil)
	case "stringSlice":
		z = newStringSliceValue([]string{}, nil)
	case "intSlice":
		z = newIntSliceValue([]int{}, nil)
	default:
		// likely a custom type
		return false
	}
	return value == z.String()
}

// UnquoteUsage extracts a back-quoted name from the usage
// string for a flag and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is an educated guess of the
// type of the flag's value, or the empty string if the flag is boolean.
func UnquoteUsage(flag *Flag) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = flag.Usage
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}
	// No explicit name, so use type if we can find one.
	name = flag.Value.Type()
	switch name {
	case "bool":
		name = ""
	case "float64":
		name = "float"
	case "int64":
		name = "int"
	case "uint64":
		name = "uint"
	}
	return
}

// Lookup returns the Flag structure of the named flag, returning nil if none exists.
func (f *FlagSet) Lookup(name string) *Flag {
	return f.flags[f.normalizeFlagName(name)]
}

// normalizeFlagName normalizes the flag name according to the normalization function.
func (f *FlagSet) normalizeFlagName(name string) string {
	if f.normalizeNameFunc != nil {
		return string(f.normalizeNameFunc(f, name))
	}
	return name
}

// Set sets the value of the named flag.
func (f *FlagSet) Set(name, value string) error {
	flag, ok := f.flags[f.normalizeFlagName(name)]
	if !ok {
		return fmt.Errorf("no such flag -%v", name)
	}
	err := flag.Value.Set(value)
	if err != nil {
		return err
	}
	if !flag.Changed {
		flag.Changed = true
	}
	return nil
}

// VisitAll visits the flags in lexicographical order, calling fn for each.
// It visits all flags, even those not set.
func (f *FlagSet) VisitAll(fn func(*Flag)) {
	for _, name := range f.order {
		fn(f.flags[name])
	}
}

// Visit visits the flags in lexicographical order, calling fn for each.
// It visits only those flags that have been set.
func (f *FlagSet) Visit(fn func(*Flag)) {
	for _, name := range f.order {
		flag := f.flags[name]
		if flag.Changed {
			fn(flag)
		}
	}
}

// addFlag will add the flag to the FlagSet
func (f *FlagSet) addFlag(flag *Flag) error {
	normalName := f.normalizeFlagName(flag.Name)
	if f.flags[normalName] != nil {
		panic(fmt.Sprintf("flag redefined: %s", flag.Name))
	}
	
	// Check for shorthand conflicts
	if len(flag.Shorthand) > 0 {
		if existingName, exists := f.shorthand[flag.Shorthand]; exists {
			panic(fmt.Sprintf("shorthand %s already used for flag %s", flag.Shorthand, existingName))
		}
		f.shorthand[flag.Shorthand] = flag.Name
	}
	
	f.flags[normalName] = flag
	f.order = append(f.order, normalName)
	
	// Register flag with OptArgs Core integration
	if err := f.coreIntegration.RegisterFlag(flag); err != nil {
		return fmt.Errorf("failed to register flag with OptArgs Core: %w", err)
	}
	
	return nil
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func (f *FlagSet) Var(value Value, name string, usage string) {
	f.VarP(value, name, "", usage)
}

// VarP is like Var, but accepts a shorthand letter that can be used after a single dash.
func (f *FlagSet) VarP(value Value, name, shorthand, usage string) {
	flag := &Flag{
		Name:      name,
		Shorthand: shorthand,
		Usage:     usage,
		Value:     value,
		DefValue:  value.String(),
	}
	f.addFlag(flag)
}
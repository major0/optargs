// Package pflags provides a drop-in replacement for spf13/pflag that maintains
// complete API compatibility while leveraging OptArgs Core for superior POSIX/GNU compliance.
package pflags

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
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
	Name                string              // name as it appears on command line
	Shorthand           string              // one-letter abbreviated flag
	Usage               string              // help message
	Value               Value               // value as set
	DefValue            string              // default value (as text); for usage message
	Changed             bool                // If the user set the value (or if left to default)
	Hidden              bool                // used by cobra.Command to allow flags to be hidden from help/usage text
	Deprecated          string              // If this flag is deprecated, this string is the new or now thing to use
	ShorthandDeprecated string              // If the shorthand of this flag is deprecated, this string is the message
	Annotations         map[string][]string // used by cobra.Command bash autocomple code
}

// FlagSet represents a set of defined flags.
type FlagSet struct {
	// Usage is the function called when an error occurs while parsing flags.
	// The field is a function (not a method) that may be changed to point to
	// a custom error handler.
	Usage func()

	name              string
	parsed            bool
	args              []string // arguments after flags
	argsLenAtDash     int      // len(args) when -- was encountered; -1 if no --
	errorHandling     ErrorHandling
	output            io.Writer // nil means stderr; use out() accessor
	interspersed      bool      // allow interspersed option/non-option args
	longOnly          bool      // getopt_long_only(3) mode
	normalizeNameFunc func(f *FlagSet, name string) NormalizedName

	// Flag storage and management
	flags     map[string]*Flag  // flags by long name
	shortOnly map[string]*Flag  // short-only flags (no long name), keyed by shorthand
	shorthand map[string]string // shorthand to long name mapping
	order     []string          // order of flag definition for help text
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
		argsLenAtDash: -1,
		flags:         make(map[string]*Flag),
		shortOnly:     make(map[string]*Flag),
		shorthand:     make(map[string]string),
		order:         make([]string, 0),
	}
	f.Usage = f.defaultUsage
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

// SetLongOnly enables or disables getopt_long_only(3) behavior.
// When enabled, single-dash arguments (e.g., -verbose) are first tried
// as long options; on failure, the parser falls back to short option parsing.
func (f *FlagSet) SetLongOnly(enabled bool) {
	f.longOnly = enabled
}

// LongOnly returns whether getopt_long_only(3) mode is enabled.
func (f *FlagSet) LongOnly() bool {
	return f.longOnly
}

// SetNormalizeFunc allows you to add a function which can translate flag names.
// Flags added to the FlagSet will be translated and then when anything tries to
// look up the flag that will also be translated. So it would be possible to create
// a flag named "getURL" and have it translated to "geturl". A user could then pass
// "--getUrl" which may also be translated to "geturl" and everything will work.
func (f *FlagSet) SetNormalizeFunc(n func(f *FlagSet, name string) NormalizedName) {
	f.normalizeNameFunc = n
	// Re-normalize existing flags under the new function.
	newFlags := make(map[string]*Flag, len(f.flags))
	newOrder := make([]string, 0, len(f.order))
	for _, flag := range f.flags {
		normalName := f.normalizeFlagName(flag.Name)
		newFlags[normalName] = flag
		newOrder = append(newOrder, normalName)
	}
	f.flags = newFlags
	f.order = newOrder
}

// GetNormalizeFunc returns the previously set NormalizeFunc, or nil if none was set.
func (f *FlagSet) GetNormalizeFunc() func(f *FlagSet, name string) NormalizedName {
	return f.normalizeNameFunc
}

// SetInterspersed sets whether to support interspersed option/non-option arguments.
// When false, option processing stops at the first non-option argument (POSIX behavior).
func (f *FlagSet) SetInterspersed(interspersed bool) {
	f.interspersed = interspersed
}

// GetInterspersed returns whether interspersed option/non-option arguments are supported.
func (f *FlagSet) GetInterspersed() bool {
	return f.interspersed
}

// Changed returns true if the named flag was set during Parse().
func (f *FlagSet) Changed(name string) bool {
	flag := f.Lookup(name)
	if flag == nil {
		return false
	}
	return flag.Changed
}

// NFlag returns the number of flags that have been set.
func (f *FlagSet) NFlag() int {
	n := 0
	f.Visit(func(*Flag) { n++ })
	return n
}

// HasFlags returns true if the FlagSet has any flags defined.
func (f *FlagSet) HasFlags() bool {
	return len(f.flags) > 0 || len(f.shortOnly) > 0
}

// HasAvailableFlags returns true if the FlagSet has any non-hidden flags.
func (f *FlagSet) HasAvailableFlags() bool {
	for _, flag := range f.flags {
		if !flag.Hidden {
			return true
		}
	}
	for _, flag := range f.shortOnly {
		if !flag.Hidden {
			return true
		}
	}
	return false
}

// Output returns the destination for usage and error messages.
// os.Stderr is returned if output was not set or was set to nil.
func (f *FlagSet) Output() io.Writer {
	return f.out()
}

// ShorthandLookup returns the Flag structure of the short handed flag,
// returning nil if none exists. It panics if len(name) > 1.
func (f *FlagSet) ShorthandLookup(name string) *Flag {
	if len(name) > 1 {
		panic("ShorthandLookup: name must be a single character")
	}
	// Check short-only flags first
	if flag, ok := f.shortOnly[name]; ok {
		return flag
	}
	// Check shorthand-to-long mapping
	if longName, ok := f.shorthand[name]; ok {
		return f.flags[f.normalizeFlagName(longName)]
	}
	return nil
}

// Init sets the name and error handling property for a flag set.
func (f *FlagSet) Init(name string, errorHandling ErrorHandling) {
	f.name = name
	f.errorHandling = errorHandling
}

// VarPF is like VarP, but returns the flag created.
func (f *FlagSet) VarPF(value Value, name, shorthand, usage string) *Flag {
	flag := &Flag{
		Name:      name,
		Shorthand: shorthand,
		Usage:     usage,
		Value:     value,
		DefValue:  value.String(),
	}
	f.addFlag(flag)
	return flag
}

// ArgsLenAtDash returns the number of args before the -- terminator,
// or -1 if no -- was encountered.
func (f *FlagSet) ArgsLenAtDash() int {
	return f.argsLenAtDash
}

// MarkDeprecated indicates that a flag is deprecated. It will continue to
// function but will not show up in help or usage messages. Using this flag
// will also print the given usageMessage.
func (f *FlagSet) MarkDeprecated(name string, usageMessage string) error {
	flag := f.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag %q does not exist", name)
	}
	if usageMessage == "" {
		return fmt.Errorf("deprecated message for flag %q must be set", name)
	}
	flag.Deprecated = usageMessage
	flag.Hidden = true
	return nil
}

// MarkHidden sets a flag to 'hidden' in your program. It will continue to
// function but will not show up in help or usage messages.
func (f *FlagSet) MarkHidden(name string) error {
	flag := f.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag %q does not exist", name)
	}
	flag.Hidden = true
	return nil
}

// MarkShorthandDeprecated will mark the shorthand of a flag deprecated.
// It will continue to function but will not show up in help or usage messages.
// Using the shorthand will also print the given usageMessage.
func (f *FlagSet) MarkShorthandDeprecated(name string, usageMessage string) error {
	flag := f.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag %q does not exist", name)
	}
	if usageMessage == "" {
		return fmt.Errorf("deprecated message for shorthand of flag %q must be set", name)
	}
	flag.ShorthandDeprecated = usageMessage
	return nil
}

// SetAnnotation allows one to set arbitrary annotations on a flag in the FlagSet.
// This is sometimes used by spf13/cobra programs which want to generate additional
// bash completion information.
func (f *FlagSet) SetAnnotation(name, key string, values []string) error {
	flag := f.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag %q does not exist", name)
	}
	if flag.Annotations == nil {
		flag.Annotations = make(map[string][]string)
	}
	flag.Annotations[key] = values
	return nil
}

// AddFlag adds the flag to the FlagSet. If a flag with the same name already
// exists, the new flag is silently ignored (matching upstream pflag behavior).
func (f *FlagSet) AddFlag(flag *Flag) {
	normalName := f.normalizeFlagName(flag.Name)
	if f.flags[normalName] != nil {
		return // silently ignore duplicates
	}
	if len(flag.Shorthand) > 0 {
		if _, exists := f.shorthand[flag.Shorthand]; exists {
			return // silently ignore shorthand conflicts
		}
		f.shorthand[flag.Shorthand] = flag.Name
	}
	f.flags[normalName] = flag
	f.order = append(f.order, normalName)
}

// AddFlagSet adds all flags from newSet to f. If a flag already exists in f,
// the flag from newSet is silently ignored.
func (f *FlagSet) AddFlagSet(newSet *FlagSet) {
	if newSet == nil {
		return
	}
	newSet.VisitAll(func(flag *Flag) {
		f.AddFlag(flag)
	})
	// Also add short-only flags
	for _, flag := range newSet.shortOnly {
		if _, exists := f.shortOnly[flag.Shorthand]; exists {
			continue
		}
		if _, exists := f.shorthand[flag.Shorthand]; exists {
			continue
		}
		f.shortOnly[flag.Shorthand] = flag
	}
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
		fmt.Fprintf(f.out(), "Usage:\n") //nolint:errcheck
	} else {
		fmt.Fprintf(f.out(), "Usage of %s:\n", f.name) //nolint:errcheck
	}
	f.PrintDefaults()
}

// PrintDefaults prints, to standard error unless configured otherwise, the
// default values of all defined flags in the set. Output format matches
// spf13/pflag: flags sorted alphabetically, usage aligned with spaces.
func (f *FlagSet) PrintDefaults() {
	// Collect visible flags and compute max line width for alignment.
	type flagLine struct {
		flag   *Flag
		prefix string // "  -o, --output string" or "      --verbose"
	}

	var lines []flagLine
	maxLen := 0

	// Collect in alphabetical order (matching upstream pflag)
	names := make([]string, 0, len(f.order))
	for _, name := range f.order {
		fl := f.flags[name]
		if fl.Hidden {
			continue
		}
		names = append(names, name)
	}
	sortStrings(names)

	for _, name := range names {
		fl := f.flags[name]
		var prefix string
		if len(fl.Shorthand) > 0 {
			prefix = fmt.Sprintf("  -%s, --%s", fl.Shorthand, fl.Name)
		} else {
			prefix = fmt.Sprintf("      --%s", fl.Name)
		}

		typeName, _ := UnquoteUsage(fl)
		if len(typeName) > 0 {
			prefix += " " + typeName
		}

		lines = append(lines, flagLine{flag: fl, prefix: prefix})
		if len(prefix) > maxLen {
			maxLen = len(prefix)
		}
	}

	w := f.out()
	for _, line := range lines {
		_, usage := UnquoteUsage(line.flag)
		padding := strings.Repeat(" ", maxLen-len(line.prefix))

		if len(usage) > 0 {
			fmt.Fprintf(w, "%s%s   %s", line.prefix, padding, usage) //nolint:errcheck
		} else {
			fmt.Fprint(w, line.prefix) //nolint:errcheck
		}

		if !isZeroValue(line.flag, line.flag.DefValue) {
			if line.flag.Value.Type() == "string" {
				fmt.Fprintf(w, " (default %q)", line.flag.DefValue) //nolint:errcheck
			} else {
				fmt.Fprintf(w, " (default %s)", line.flag.DefValue) //nolint:errcheck
			}
		}
		fmt.Fprint(w, "\n") //nolint:errcheck
	}
}

// FlagUsages returns a string containing the usage information for all defined
// flags in the set. This is the same output as PrintDefaults but returned as
// a string instead of written to the output.
func (f *FlagSet) FlagUsages() string {
	var buf bytes.Buffer
	old := f.output
	f.output = &buf
	f.PrintDefaults()
	f.output = old
	return buf.String()
}

// sortStrings sorts a slice of strings in place.
func sortStrings(s []string) { sort.Strings(s) }

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
func (f *FlagSet) addFlag(flag *Flag) {
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

// ShortVar registers a short-only flag (no long name). The flag is accessible
// only via its single-character shorthand and participates in POSIX compaction.
func (f *FlagSet) ShortVar(value Value, shorthand, usage string) {
	if len(shorthand) != 1 {
		panic("ShortVar: shorthand must be exactly one character")
	}
	if _, exists := f.shortOnly[shorthand]; exists {
		panic(fmt.Sprintf("short-only flag redefined: %s", shorthand))
	}
	if _, exists := f.shorthand[shorthand]; exists {
		panic(fmt.Sprintf("shorthand %s already in use", shorthand))
	}
	flag := &Flag{
		Name:      shorthand,
		Shorthand: shorthand,
		Usage:     usage,
		Value:     value,
		DefValue:  value.String(),
	}
	f.shortOnly[shorthand] = flag
}

// AliasVar registers an additional flag name that writes to the same Value
// as an existing flag. The alias flag is hidden from help text by default.
// This enables the ls --format=across / -x pattern where multiple flags
// share a destination with last-occurrence-wins semantics.
func (f *FlagSet) AliasVar(value Value, name, usage string) {
	flag := &Flag{
		Name:     name,
		Usage:    usage,
		Value:    value,
		DefValue: value.String(),
		Hidden:   true,
	}
	f.addFlag(flag)
}

// AliasVarP is like AliasVar but accepts a shorthand.
func (f *FlagSet) AliasVarP(value Value, name, shorthand, usage string) {
	flag := &Flag{
		Name:      name,
		Shorthand: shorthand,
		Usage:     usage,
		Value:     value,
		DefValue:  value.String(),
		Hidden:    true,
	}
	f.addFlag(flag)
}

// AliasShortVar registers a short-only alias that writes to the same Value.
// The alias is hidden from help text by default.
func (f *FlagSet) AliasShortVar(value Value, shorthand string) {
	if len(shorthand) != 1 {
		panic("AliasShortVar: shorthand must be exactly one character")
	}
	if _, exists := f.shortOnly[shorthand]; exists {
		panic(fmt.Sprintf("short-only flag redefined: %s", shorthand))
	}
	if _, exists := f.shorthand[shorthand]; exists {
		panic(fmt.Sprintf("shorthand %s already in use", shorthand))
	}
	flag := &Flag{
		Name:      shorthand,
		Shorthand: shorthand,
		Value:     value,
		DefValue:  value.String(),
		Hidden:    true,
	}
	f.shortOnly[shorthand] = flag
}

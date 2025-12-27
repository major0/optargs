package pflags

import (
	"fmt"
	"os"
	"strings"
	"time"
	
	"github.com/major0/optargs"
)

// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the FlagSet
// are defined and before flags are accessed by the program.
// The return value will be ErrHelp if -help was set but not defined.
func (f *FlagSet) Parse(arguments []string) error {
	// Initialize OptArgs Core parser with current arguments
	if err := f.coreIntegration.InitializeParser(arguments); err != nil {
		return f.coreIntegration.TranslateError(err)
	}
	
	// Get the parser from core integration
	parser := f.coreIntegration.GetParser()
	if parser == nil {
		return fmt.Errorf("failed to initialize OptArgs Core parser")
	}
	
	// Process options using OptArgs Core
	if err := f.processOptions(parser); err != nil {
		return err
	}
	
	// Store remaining non-flag arguments
	f.args = parser.Args
	f.parsed = true
	
	return nil
}

// processOptions iterates through OptArgs Core options and updates flag values
func (f *FlagSet) processOptions(parser *optargs.Parser) error {
	for option, err := range parser.Options() {
		if err != nil {
			return f.coreIntegration.TranslateError(err)
		}
		
		// Find the corresponding pflag Flag
		var flag *Flag
		var isNegation bool
		
		// Handle both short and long option names
		if len(option.Name) == 1 {
			// Short option - look up by shorthand
			if longName, exists := f.shorthand[option.Name]; exists {
				flag = f.flags[f.normalizeFlagName(longName)]
			}
		} else {
			// Check if this is a negation flag (--no-<flag>)
			if strings.HasPrefix(option.Name, "no-") && len(option.Name) > 3 {
				originalName := option.Name[3:] // Remove "no-" prefix
				flag = f.flags[f.normalizeFlagName(originalName)]
				if flag != nil && flag.Value.Type() == "bool" {
					isNegation = true
				} else {
					flag = nil // Not a valid negation
				}
			}
			
			// If not a negation or negation lookup failed, try direct lookup
			if flag == nil {
				flag = f.flags[f.normalizeFlagName(option.Name)]
			}
		}
		
		if flag == nil {
			return fmt.Errorf("unknown flag: %s", option.Name)
		}
		
		// Set the flag value
		if err := f.setFlagValue(flag, option, isNegation); err != nil {
			return err
		}
		
		// Mark flag as changed
		flag.Changed = true
	}
	
	return nil
}

// setFlagValue sets the flag value based on the OptArgs Core option
func (f *FlagSet) setFlagValue(flag *Flag, option optargs.Option, isNegation bool) error {
	// Handle boolean flags specially
	if flag.Value.Type() == "bool" {
		if isNegation {
			// Negation flag (--no-<flag>) always sets to false
			return flag.Value.Set("false")
		} else if option.HasArg {
			// Explicit boolean value provided (--flag=value)
			// OptArgs Core includes the '=' in the argument, so we need to strip it
			arg := option.Arg
			if strings.HasPrefix(arg, "=") {
				arg = arg[1:]
			}
			return flag.Value.Set(arg)
		} else {
			// No argument provided (--flag), set to true
			return flag.Value.Set("true")
		}
	}
	
	// For non-boolean flags, use the provided argument
	if !option.HasArg {
		return fmt.Errorf("flag --%s requires an argument", flag.Name)
	}
	
	// OptArgs Core includes the '=' in the argument for long options, so we need to strip it
	arg := option.Arg
	if strings.HasPrefix(arg, "=") {
		arg = arg[1:]
	}
	
	return flag.Value.Set(arg)
}

// CommandLine is the default set of command-line flags, parsed from os.Args.
var CommandLine = NewFlagSet(os.Args[0], ExitOnError)

// Global flag registration functions that operate on CommandLine

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, value string, usage string) {
	CommandLine.VarP(newStringValue(value, p), name, "", usage)
}

// StringVarP is like StringVar, but accepts a shorthand letter that can be used after a single dash.
func StringVarP(p *string, name, shorthand string, value string, usage string) {
	CommandLine.VarP(newStringValue(value, p), name, shorthand, usage)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name string, value string, usage string) *string {
	return CommandLine.StringP(name, "", value, usage)
}

// StringP is like String, but accepts a shorthand letter that can be used after a single dash.
func StringP(name, shorthand string, value string, usage string) *string {
	return CommandLine.StringP(name, shorthand, value, usage)
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func BoolVar(p *bool, name string, value bool, usage string) {
	CommandLine.VarP(newBoolValue(value, p), name, "", usage)
}

// BoolVarP is like BoolVar, but accepts a shorthand letter that can be used after a single dash.
func BoolVarP(p *bool, name, shorthand string, value bool, usage string) {
	CommandLine.VarP(newBoolValue(value, p), name, shorthand, usage)
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func Bool(name string, value bool, usage string) *bool {
	return CommandLine.BoolP(name, "", value, usage)
}

// BoolP is like Bool, but accepts a shorthand letter that can be used after a single dash.
func BoolP(name, shorthand string, value bool, usage string) *bool {
	return CommandLine.BoolP(name, shorthand, value, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVar(p *int, name string, value int, usage string) {
	CommandLine.VarP(newIntValue(value, p), name, "", usage)
}

// IntVarP is like IntVar, but accepts a shorthand letter that can be used after a single dash.
func IntVarP(p *int, name, shorthand string, value int, usage string) {
	CommandLine.VarP(newIntValue(value, p), name, shorthand, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name string, value int, usage string) *int {
	return CommandLine.IntP(name, "", value, usage)
}

// IntP is like Int, but accepts a shorthand letter that can be used after a single dash.
func IntP(name, shorthand string, value int, usage string) *int {
	return CommandLine.IntP(name, shorthand, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64Var(p *float64, name string, value float64, usage string) {
	CommandLine.VarP(newFloat64Value(value, p), name, "", usage)
}

// Float64VarP is like Float64Var, but accepts a shorthand letter that can be used after a single dash.
func Float64VarP(p *float64, name, shorthand string, value float64, usage string) {
	CommandLine.VarP(newFloat64Value(value, p), name, shorthand, usage)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64(name string, value float64, usage string) *float64 {
	return CommandLine.Float64P(name, "", value, usage)
}

// Float64P is like Float64, but accepts a shorthand letter that can be used after a single dash.
func Float64P(name, shorthand string, value float64, usage string) *float64 {
	return CommandLine.Float64P(name, shorthand, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	CommandLine.VarP(newDurationValue(value, p), name, "", usage)
}

// DurationVarP is like DurationVar, but accepts a shorthand letter that can be used after a single dash.
func DurationVarP(p *time.Duration, name, shorthand string, value time.Duration, usage string) {
	CommandLine.VarP(newDurationValue(value, p), name, shorthand, usage)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func Duration(name string, value time.Duration, usage string) *time.Duration {
	return CommandLine.DurationP(name, "", value, usage)
}

// DurationP is like Duration, but accepts a shorthand letter that can be used after a single dash.
func DurationP(name, shorthand string, value time.Duration, usage string) *time.Duration {
	return CommandLine.DurationP(name, shorthand, value, usage)
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value.
func Var(value Value, name string, usage string) {
	CommandLine.VarP(value, name, "", usage)
}

// VarP is like Var, but accepts a shorthand letter that can be used after a single dash.
func VarP(value Value, name, shorthand, usage string) {
	CommandLine.VarP(value, name, shorthand, usage)
}

// Parse parses the command-line flags from os.Args[1:]. Must be called
// after all flags are defined and before flags are accessed by the program.
func Parse() {
	// Ignore errors; CommandLine is set for ExitOnError.
	CommandLine.Parse(os.Args[1:])
}

// Parsed returns true if the command-line flags have been parsed.
func Parsed() bool {
	return CommandLine.Parsed()
}

// Args returns the non-flag command-line arguments.
func Args() []string {
	return CommandLine.Args()
}

// NArg is the number of arguments remaining after flags have been processed.
func NArg() int {
	return CommandLine.NArg()
}

// Arg returns the i'th command-line argument. Arg(0) is the first remaining argument
// after flags have been processed.
func Arg(i int) string {
	return CommandLine.Arg(i)
}

// Lookup returns the Flag structure of the named command-line flag,
// returning nil if none exists.
func Lookup(name string) *Flag {
	return CommandLine.Lookup(name)
}

// Set sets the value of the named command-line flag.
func Set(name, value string) error {
	return CommandLine.Set(name, value)
}

// PrintDefaults prints, to standard error unless configured otherwise,
// the default values of all defined command-line flags.
func PrintDefaults() {
	CommandLine.PrintDefaults()
}

// Usage prints a usage message documenting all defined command-line flags
// to CommandLine's output, which by default is os.Stderr.
func Usage() {
	CommandLine.Usage()
}

// VisitAll visits the command-line flags in lexicographical order, calling
// fn for each. It visits all flags, even those not set.
func VisitAll(fn func(*Flag)) {
	CommandLine.VisitAll(fn)
}

// Visit visits the command-line flags in lexicographical order, calling fn
// for each. It visits only those flags that have been set.
func Visit(fn func(*Flag)) {
	CommandLine.Visit(fn)
}
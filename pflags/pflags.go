package pflags

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/major0/optargs"
)

// isBoolFlag returns true if the value is a boolean flag — either by type
// or by implementing the boolFlag interface (IsBoolFlag() bool).
func isBoolFlag(v Value) bool {
	if v.Type() == "bool" {
		return true
	}
	type boolFlagger interface{ IsBoolFlag() bool }
	if bf, ok := v.(boolFlagger); ok {
		return bf.IsBoolFlag()
	}
	return false
}

// shortArgType returns the core argument type for a short option.
// Boolean flags use NoArgument for POSIX compaction; others use RequiredArgument.
func shortArgType(v Value) optargs.ArgType {
	if isBoolFlag(v) {
		return optargs.NoArgument
	}
	return mapArgumentType(v.Type())
}

// buildShortOpts constructs the short option map for optargs.NewParser
// from the FlagSet's registered flags and shorthand mappings.
// Boolean short opts use NoArgument so they participate in POSIX compaction (-abc).
func (f *FlagSet) buildShortOpts() map[byte]*optargs.Flag {
	shortOpts := make(map[byte]*optargs.Flag)

	// Regular flags with shorthands
	for shortStr, longName := range f.shorthand {
		flag := f.flags[f.normalizeFlagName(longName)]
		if flag == nil {
			continue
		}
		shortChar := shortStr[0]
		coreFlag := &optargs.Flag{
			Name:   string(shortChar),
			HasArg: shortArgType(flag.Value),
			Handle: f.makeHandler(flag),
		}
		shortOpts[shortChar] = coreFlag
	}

	// Short-only flags
	for shortStr, flag := range f.shortOnly {
		shortChar := shortStr[0]
		coreFlag := &optargs.Flag{
			Name:   string(shortChar),
			HasArg: shortArgType(flag.Value),
			Handle: f.makeHandler(flag),
		}
		shortOpts[shortChar] = coreFlag
	}

	return shortOpts
}

// buildLongOpts constructs the long option map for optargs.NewParser
// from the FlagSet's registered flags. Also registers --no-<name>
// negation flags for boolean flags.
func (f *FlagSet) buildLongOpts() map[string]*optargs.Flag {
	longOpts := make(map[string]*optargs.Flag)
	for name, flag := range f.flags {
		coreFlag := &optargs.Flag{
			Name:   name,
			HasArg: mapArgumentType(flag.Value.Type()),
			Handle: f.makeHandler(flag),
		}
		longOpts[name] = coreFlag

		// Register negation flag for booleans
		if flag.Value.Type() == "bool" {
			negName := "no-" + name
			negFlag := &optargs.Flag{
				Name:   negName,
				HasArg: optargs.OptionalArgument,
				Handle: f.makeNegationHandler(flag),
			}
			longOpts[negName] = negFlag
		}
	}
	return longOpts
}

// makeHandler returns a handler function for the given pflags Flag.
// For boolean flags (type "bool" or IsBoolFlag()), no-arg sets "true" or
// calls Set("") for custom bool flags. For all other types, the handler
// calls Value.Set(arg) directly.
func (f *FlagSet) makeHandler(flag *Flag) func(string, string) error {
	return func(name, arg string) error {
		val := arg
		if isBoolFlag(flag.Value) && val == "" {
			if flag.Value.Type() == "bool" {
				val = "true"
			}
			// For custom IsBoolFlag types, call Set("") — the value
			// implementation decides what no-arg means.
		}
		if err := flag.Value.Set(val); err != nil {
			return err
		}
		flag.Changed = true
		return nil
	}
}

// makeNegationHandler returns a handler for --no-<name> boolean negation flags.
// no-arg or =true → Set("false"), =false → Set("true").
func (f *FlagSet) makeNegationHandler(flag *Flag) func(string, string) error {
	return func(name, arg string) error {
		switch strings.ToLower(arg) {
		case "", "true", "1", "t":
			if err := flag.Value.Set("false"); err != nil {
				return err
			}
		case "false", "0", "f":
			if err := flag.Value.Set("true"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid boolean value '%s'", arg)
		}
		flag.Changed = true
		return nil
	}
}

// mapArgumentType converts pflags value types to OptArgs Core argument types.
func mapArgumentType(valueType string) optargs.ArgType {
	switch valueType {
	case "bool":
		return optargs.OptionalArgument
	default:
		return optargs.RequiredArgument
	}
}

// translateError converts OptArgs Core errors to pflag-compatible error messages.
func translateError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "unknown option"):
		if idx := strings.Index(errMsg, ": "); idx >= 0 {
			optionName := errMsg[idx+2:]
			if len(optionName) == 1 {
				return fmt.Errorf("unknown shorthand flag: '%s'", optionName)
			}
			return fmt.Errorf("unknown flag: --%s", optionName)
		}
		return fmt.Errorf("unknown flag: %s", errMsg)

	case strings.Contains(errMsg, "option requires an argument"):
		if idx := strings.Index(errMsg, ": "); idx >= 0 {
			optionName := errMsg[idx+2:]
			if len(optionName) == 1 {
				return fmt.Errorf("flag needs an argument: -%s", optionName)
			}
			return fmt.Errorf("flag needs an argument: --%s", optionName)
		}
		return fmt.Errorf("flag needs an argument: %s", errMsg)

	default:
		return err
	}
}

// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the FlagSet
// are defined and before flags are accessed by the program.
//
// ErrorHandling behavior on parse failure:
//   - ContinueOnError: return the error
//   - ExitOnError: print error + usage to output, call os.Exit(2)
//   - PanicOnError: print error + usage to output, panic
func (f *FlagSet) Parse(arguments []string) error {
	// Detect -- position for ArgsLenAtDash
	dashPos := -1
	for i, arg := range arguments {
		if arg == "--" {
			dashPos = i
			break
		}
	}

	shortOpts := f.buildShortOpts()
	longOpts := f.buildLongOpts()

	config := optargs.ParserConfig{}
	config.SetLongOnly(f.longOnly)

	parser, err := optargs.NewParser(config, shortOpts, longOpts, arguments)
	if err != nil {
		return f.failf("%v", translateError(err))
	}

	// Consume the iterator — handlers do the work, we only propagate errors.
	for _, err := range parser.Options() {
		if err != nil {
			return f.failf("%v", translateError(err))
		}
	}

	f.args = parser.Args
	f.parsed = true

	// Compute argsLenAtDash: if -- was present, count how many positional
	// args appeared before it. The args after -- are at the tail of f.args.
	if dashPos >= 0 {
		argsAfterDash := len(arguments) - dashPos - 1 // args after the -- token
		f.argsLenAtDash = len(f.args) - argsAfterDash
		if f.argsLenAtDash < 0 {
			f.argsLenAtDash = 0
		}
	}

	return nil
}

// failf handles a parse error according to the FlagSet's ErrorHandling mode.
// For ContinueOnError it returns the error. For ExitOnError and PanicOnError
// it prints the error and usage before exiting or panicking.
func (f *FlagSet) failf(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	switch f.errorHandling {
	case ContinueOnError:
		return err
	case ExitOnError:
		fmt.Fprintln(f.out(), err) //nolint:errcheck // writing to output
		f.Usage()
		os.Exit(2)
	case PanicOnError:
		fmt.Fprintln(f.out(), err) //nolint:errcheck // writing to output
		f.Usage()
		panic(err)
	}
	return err
}

// CommandLine is the default set of command-line flags, parsed from os.Args.
var CommandLine = NewFlagSet(os.Args[0], ExitOnError)

// Global flag registration functions that operate on CommandLine

func StringVar(p *string, name string, value string, usage string) {
	CommandLine.VarP(newStringValue(value, p), name, "", usage)
}

func StringVarP(p *string, name, shorthand string, value string, usage string) {
	CommandLine.VarP(newStringValue(value, p), name, shorthand, usage)
}

func String(name string, value string, usage string) *string {
	return CommandLine.StringP(name, "", value, usage)
}

func StringP(name, shorthand string, value string, usage string) *string {
	return CommandLine.StringP(name, shorthand, value, usage)
}

func BoolVar(p *bool, name string, value bool, usage string) {
	CommandLine.VarP(newBoolValue(value, p), name, "", usage)
}

func BoolVarP(p *bool, name, shorthand string, value bool, usage string) {
	CommandLine.VarP(newBoolValue(value, p), name, shorthand, usage)
}

func Bool(name string, value bool, usage string) *bool {
	return CommandLine.BoolP(name, "", value, usage)
}

func BoolP(name, shorthand string, value bool, usage string) *bool {
	return CommandLine.BoolP(name, shorthand, value, usage)
}

func IntVar(p *int, name string, value int, usage string) {
	CommandLine.VarP(newIntValue(value, p), name, "", usage)
}

func IntVarP(p *int, name, shorthand string, value int, usage string) {
	CommandLine.VarP(newIntValue(value, p), name, shorthand, usage)
}

func Int(name string, value int, usage string) *int {
	return CommandLine.IntP(name, "", value, usage)
}

func IntP(name, shorthand string, value int, usage string) *int {
	return CommandLine.IntP(name, shorthand, value, usage)
}

func Float64Var(p *float64, name string, value float64, usage string) {
	CommandLine.VarP(newFloat64Value(value, p), name, "", usage)
}

func Float64VarP(p *float64, name, shorthand string, value float64, usage string) {
	CommandLine.VarP(newFloat64Value(value, p), name, shorthand, usage)
}

func Float64(name string, value float64, usage string) *float64 {
	return CommandLine.Float64P(name, "", value, usage)
}

func Float64P(name, shorthand string, value float64, usage string) *float64 {
	return CommandLine.Float64P(name, shorthand, value, usage)
}

func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	CommandLine.VarP(newDurationValue(value, p), name, "", usage)
}

func DurationVarP(p *time.Duration, name, shorthand string, value time.Duration, usage string) {
	CommandLine.VarP(newDurationValue(value, p), name, shorthand, usage)
}

func Duration(name string, value time.Duration, usage string) *time.Duration {
	return CommandLine.DurationP(name, "", value, usage)
}

func DurationP(name, shorthand string, value time.Duration, usage string) *time.Duration {
	return CommandLine.DurationP(name, shorthand, value, usage)
}

func Var(value Value, name string, usage string) {
	CommandLine.VarP(value, name, "", usage)
}

func VarP(value Value, name, shorthand, usage string) {
	CommandLine.VarP(value, name, shorthand, usage)
}

func Parse() {
	CommandLine.Parse(os.Args[1:]) //nolint:errcheck // ExitOnError handles errors
}

func Parsed() bool {
	return CommandLine.Parsed()
}

func Args() []string {
	return CommandLine.Args()
}

func NArg() int {
	return CommandLine.NArg()
}

func Arg(i int) string {
	return CommandLine.Arg(i)
}

func Lookup(name string) *Flag {
	return CommandLine.Lookup(name)
}

func Set(name, value string) error {
	return CommandLine.Set(name, value)
}

func PrintDefaults() {
	CommandLine.PrintDefaults()
}

func FlagUsages() string {
	return CommandLine.FlagUsages()
}

func Usage() {
	CommandLine.Usage()
}

func VisitAll(fn func(*Flag)) {
	CommandLine.VisitAll(fn)
}

func Visit(fn func(*Flag)) {
	CommandLine.Visit(fn)
}

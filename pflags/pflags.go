package pflags

import (
	"fmt"
	"os"
	"strings"

	"github.com/major0/optargs"
)

// isBoolFlag returns true if the value is a boolean flag.
// Delegates to optargs.IsBool which checks both Type() == "bool"
// and the BoolValuer interface.
func isBoolFlag(v Value) bool {
	return optargs.IsBool(v)
}

// shortOptArgType returns the core argument type for a short option.
// Boolean flags use NoArgument for POSIX compaction; others use RequiredArgument.
func shortOptArgType(v Value) optargs.ArgType {
	if isBoolFlag(v) {
		return optargs.NoArgument
	}
	return optargs.RequiredArgument
}

// buildShortOpts constructs the short option map for optargs.NewParser
// from the FlagSet's registered flags and shorthand mappings.
// Boolean short opts use NoArgument so they participate in POSIX compaction (-abc).
func (f *FlagSet) buildShortOpts() map[byte]*optargs.Flag {
	shortOpts := make(map[byte]*optargs.Flag)

	addShort := func(shortChar byte, flag *Flag) {
		shortOpts[shortChar] = &optargs.Flag{
			Name:   string(shortChar),
			HasArg: shortOptArgType(flag.Value),
			Handle: f.makeHandler(flag),
		}
	}

	for shortStr, longName := range f.shorthand {
		flag := f.flags[f.normalizeFlagName(longName)]
		if flag == nil {
			continue
		}
		addShort(shortStr[0], flag)
	}
	for shortStr, flag := range f.shortOnly {
		addShort(shortStr[0], flag)
	}

	return shortOpts
}

// normalizeArgs applies the normalize func to long option names in the
// argument list. This translates --my_flag to --my-flag (or whatever the
// normalize func does) so the core parser can match against registered flags.
// Short options and non-option args are left unchanged.
func (f *FlagSet) normalizeArgs(args []string) []string {
	out := make([]string, len(args))
	terminated := false
	for i, arg := range args {
		if terminated || arg == "--" {
			out[i] = arg
			terminated = true
			continue
		}
		if strings.HasPrefix(arg, "--") {
			// Split on = to handle --name=value
			name := arg[2:]
			eqIdx := strings.Index(name, "=")
			if eqIdx >= 0 {
				flagName := name[:eqIdx]
				normalized := string(f.normalizeNameFunc(f, flagName))
				out[i] = "--" + normalized + name[eqIdx:]
			} else {
				normalized := string(f.normalizeNameFunc(f, name))
				out[i] = "--" + normalized
			}
		} else {
			out[i] = arg
		}
	}
	return out
}

// boolLongArgType returns the core argument type for a boolean long option.
// If the value implements BoolArgValuer and returns false, the flag is
// strictly no-argument. Otherwise it accepts an optional =value.
func boolLongArgType(v Value) optargs.ArgType {
	type boolArgValuer interface{ BoolTakesArg() bool }
	if ba, ok := v.(boolArgValuer); ok && !ba.BoolTakesArg() {
		return optargs.NoArgument
	}
	return optargs.OptionalArgument
}

// buildLongOpts constructs the long option map for optargs.NewParser
// from the FlagSet's registered flags. Also registers --no-<name>
// negation flags for boolean flags.
func (f *FlagSet) buildLongOpts() map[string]*optargs.Flag {
	longOpts := make(map[string]*optargs.Flag)
	for normalizedName, flag := range f.flags {
		handler := f.makeHandler(flag)
		isBool := isBoolFlag(flag.Value)
		hasArg := optargs.RequiredArgument
		if isBool {
			hasArg = boolLongArgType(flag.Value)
		}

		longOpts[normalizedName] = &optargs.Flag{
			Name:   normalizedName,
			HasArg: hasArg,
			Handle: handler,
		}

		// Register negation flag for booleans that accept an argument
		if isBool && hasArg == optargs.OptionalArgument {
			negName := "no-" + normalizedName
			longOpts[negName] = &optargs.Flag{
				Name:   negName,
				HasArg: optargs.OptionalArgument,
				Handle: f.makeNegationHandler(flag),
			}
		}

		// Register prefix pair options for boolean flags (always NoArgument)
		for _, pp := range flag.Prefixes {
			trueName := f.normalizeFlagName(pp.True + "-" + normalizedName)
			falseName := f.normalizeFlagName(pp.False + "-" + normalizedName)
			longOpts[trueName] = &optargs.Flag{
				Name:   trueName,
				HasArg: optargs.NoArgument,
				Handle: f.makeBoolPrefixHandler(flag, "true"),
			}
			longOpts[falseName] = &optargs.Flag{
				Name:   falseName,
				HasArg: optargs.NoArgument,
				Handle: f.makeBoolPrefixHandler(flag, "false"),
			}
		}

		// Register --no-<name> for negatable non-boolean flags (always NoArgument)
		if flag.Negatable && !isBool {
			negName := f.normalizeFlagName("no-" + normalizedName)
			longOpts[negName] = &optargs.Flag{
				Name:   negName,
				HasArg: optargs.NoArgument,
				Handle: f.makeNegatableHandler(flag),
			}
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
			return &InvalidValueError{flag: flag, value: val, err: err}
		}
		flag.Changed = true
		if f.parseAllFn != nil {
			if err := f.parseAllFn(flag, val); err != nil {
				return err
			}
		}
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

// makeBoolPrefixHandler returns a handler for a prefixed boolean option
// (e.g. --enable-shared, --disable-shared). The val argument is "true" or "false".
func (f *FlagSet) makeBoolPrefixHandler(flag *Flag, val string) func(string, string) error {
	return func(_, _ string) error {
		if err := flag.Value.Set(val); err != nil {
			return err
		}
		flag.Changed = true
		if f.parseAllFn != nil {
			if err := f.parseAllFn(flag, val); err != nil {
				return err
			}
		}
		return nil
	}
}

// makeNegatableHandler returns a handler for --no-<name> on a non-boolean flag.
// Clears the value to its type's zero value: Reset() for collections, Set(zeroVal) for scalars.
func (f *FlagSet) makeNegatableHandler(flag *Flag) func(string, string) error {
	zeroVal, _ := optargs.ZeroString(flag.Value.Type())
	return func(_, _ string) error {
		if r, ok := flag.Value.(optargs.Resetter); ok {
			r.Reset()
		} else if err := flag.Value.Set(zeroVal); err != nil {
			return err
		}
		flag.Changed = true
		if f.parseAllFn != nil {
			if err := f.parseAllFn(flag, zeroVal); err != nil {
				return err
			}
		}
		return nil
	}
}

// translateError converts OptArgs Core errors to pflag-compatible structured errors.
func translateError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "unknown option"):
		name := ""
		if idx := strings.Index(errMsg, ": "); idx >= 0 {
			name = errMsg[idx+2:]
		}
		if len(name) == 1 {
			return &NotExistError{specifiedName: name, specifiedShortnames: name}
		}
		return &NotExistError{specifiedName: name}

	case strings.Contains(errMsg, "option requires an argument"):
		name := ""
		if idx := strings.Index(errMsg, ": "); idx >= 0 {
			name = errMsg[idx+2:]
		}
		if len(name) == 1 {
			return &ValueRequiredError{specifiedName: name, specifiedShortnames: name}
		}
		return &ValueRequiredError{specifiedName: name}

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

	// If a normalize func is set, normalize long option names in the
	// arguments so the core parser can match them against registered flags.
	if f.normalizeNameFunc != nil {
		arguments = f.normalizeArgs(arguments)
	}

	shortOpts := f.buildShortOpts()
	longOpts := f.buildLongOpts()

	config := optargs.ParserConfig{}
	config.SetLongOnly(f.longOnly)
	config.SetInterspersed(f.interspersed)

	parser, err := optargs.NewParser(config, shortOpts, longOpts, arguments)
	if err != nil {
		return f.failf(translateError(err))
	}

	// Consume the iterator — handlers do the work, we only propagate errors.
	for _, err := range parser.Options() {
		if err != nil {
			translated := translateError(err)
			// Skip unknown flag errors if allowlisted
			if f.ParseErrorsAllowlist.UnknownFlags {
				if _, ok := translated.(*NotExistError); ok {
					continue
				}
			}
			return f.failf(translated)
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

// ParseAll parses flag definitions from the argument list and calls fn for
// each flag that is set. The arguments for fn are the flag and its value.
func (f *FlagSet) ParseAll(arguments []string, fn func(flag *Flag, value string) error) error {
	f.parseAllFn = fn
	defer func() { f.parseAllFn = nil }()
	return f.Parse(arguments)
}

// failf handles a parse error according to the FlagSet's ErrorHandling mode.
// For ContinueOnError it returns the error. For ExitOnError and PanicOnError
// it prints the error and usage before exiting or panicking.
func (f *FlagSet) failf(err error) error {
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

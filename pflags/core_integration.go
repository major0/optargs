package pflags

import (
	"fmt"
	"strings"

	"github.com/major0/optargs"
)

// CoreIntegration provides the translation layer between pflag Flag definitions
// and OptArgs Core format, handling flag registration and parsing delegation.
type CoreIntegration struct {
	flagSet   *FlagSet                  // Parent flag set
	parser    *optargs.Parser           // OptArgs Core parser
	flagMap   map[string]*optargs.Flag  // OptArgs flags by name
	shortOpts map[byte]*optargs.Flag    // Short options mapping
	longOpts  map[string]*optargs.Flag  // Long options mapping
}

// NewCoreIntegration creates a new CoreIntegration instance for the given FlagSet.
func NewCoreIntegration(flagSet *FlagSet) *CoreIntegration {
	return &CoreIntegration{
		flagSet:   flagSet,
		flagMap:   make(map[string]*optargs.Flag),
		shortOpts: make(map[byte]*optargs.Flag),
		longOpts:  make(map[string]*optargs.Flag),
	}
}

// RegisterFlag converts a pflag Flag definition to OptArgs Core format and registers it.
// This handles the translation between pflag's Flag structure and OptArgs Core's Flag structure.
func (ci *CoreIntegration) RegisterFlag(flag *Flag) error {
	// Determine argument type based on flag value type
	argType := ci.mapArgumentType(flag.Value.Type())
	
	// Create OptArgs Core flag
	coreFlag := &optargs.Flag{
		Name:   flag.Name,
		HasArg: argType,
	}
	
	// Store the mapping for later lookup
	ci.flagMap[flag.Name] = coreFlag
	ci.longOpts[flag.Name] = coreFlag
	
	// For boolean flags, also register the negation form --no-<flag>
	if flag.Value.Type() == "bool" {
		negationName := "no-" + flag.Name
		negationFlag := &optargs.Flag{
			Name:   negationName,
			HasArg: optargs.NoArgument,
		}
		ci.flagMap[negationName] = negationFlag
		ci.longOpts[negationName] = negationFlag
	}
	
	// Register shorthand if present
	if len(flag.Shorthand) > 0 {
		if len(flag.Shorthand) != 1 {
			return fmt.Errorf("shorthand must be a single character: %s", flag.Shorthand)
		}
		
		shortChar := flag.Shorthand[0]
		ci.shortOpts[shortChar] = coreFlag
	}
	
	return nil
}

// mapArgumentType converts pflag value types to OptArgs Core argument types.
// Boolean flags are treated as OptionalArgument to support both --flag and --flag=value syntax.
func (ci *CoreIntegration) mapArgumentType(valueType string) optargs.ArgType {
	switch valueType {
	case "bool":
		// Boolean flags support optional arguments - they can be used as --flag or --flag=value
		return optargs.OptionalArgument
	default:
		// All other types require arguments
		return optargs.RequiredArgument
	}
}

// InitializeParser creates and configures the OptArgs Core parser with all registered flags.
func (ci *CoreIntegration) InitializeParser(args []string) error {
	// Build optstring for short options
	optstring := ci.buildOptString()
	
	// Build long options slice
	longOpts := ci.buildLongOpts()
	
	// Create parser using OptArgs Core
	// If there are no flags registered, create a minimal parser
	if len(ci.longOpts) == 0 && len(ci.shortOpts) == 0 {
		parser, err := optargs.GetOptLong(args, "", []optargs.Flag{})
		if err != nil {
			return fmt.Errorf("failed to create OptArgs parser: %w", err)
		}
		ci.parser = parser
		return nil
	}
	
	parser, err := optargs.GetOptLong(args, optstring, longOpts)
	if err != nil {
		// If OptArgs Core rejects the flag definitions due to validation,
		// we'll create a minimal parser and handle parsing ourselves
		parser, fallbackErr := optargs.GetOptLong(args, "", []optargs.Flag{})
		if fallbackErr != nil {
			return fmt.Errorf("failed to create fallback OptArgs parser: %w", fallbackErr)
		}
		ci.parser = parser
		return nil
	}
	
	ci.parser = parser
	return nil
}

// buildOptString constructs the optstring for GetOptLong based on registered short options.
// Format: "abc:d::" where 'a' takes no arg, 'b' takes no arg, 'c' requires arg, 'd' has optional arg.
func (ci *CoreIntegration) buildOptString() string {
	var optstring strings.Builder
	
	for shortChar, flag := range ci.shortOpts {
		optstring.WriteByte(shortChar)
		
		switch flag.HasArg {
		case optargs.RequiredArgument:
			optstring.WriteByte(':')
		case optargs.OptionalArgument:
			optstring.WriteString("::")
		case optargs.NoArgument:
			// No suffix needed
		}
	}
	
	return optstring.String()
}

// buildLongOpts creates the slice of long options for GetOptLong.
func (ci *CoreIntegration) buildLongOpts() []optargs.Flag {
	longOpts := make([]optargs.Flag, 0, len(ci.longOpts))
	
	for _, flag := range ci.longOpts {
		longOpts = append(longOpts, *flag)
	}
	
	return longOpts
}

// GetParser returns the initialized OptArgs Core parser.
// Returns nil if InitializeParser hasn't been called successfully.
func (ci *CoreIntegration) GetParser() *optargs.Parser {
	return ci.parser
}

// TranslateError converts OptArgs Core errors to pflag-compatible error messages.
func (ci *CoreIntegration) TranslateError(err error) error {
	if err == nil {
		return nil
	}
	
	errMsg := err.Error()
	
	// Handle common OptArgs error patterns and translate to pflag format
	switch {
	case strings.Contains(errMsg, "unknown option"):
		// Extract option name from error message
		if strings.Contains(errMsg, ": ") {
			parts := strings.Split(errMsg, ": ")
			if len(parts) > 1 {
				optionName := parts[1]
				// Format as pflag expects
				if len(optionName) == 1 {
					return fmt.Errorf("unknown shorthand flag: '%s'", optionName)
				} else {
					return fmt.Errorf("unknown flag: --%s", optionName)
				}
			}
		}
		return fmt.Errorf("unknown flag: %s", errMsg)
		
	case strings.Contains(errMsg, "option requires an argument"):
		// Extract option name and format as pflag expects
		if strings.Contains(errMsg, ": ") {
			parts := strings.Split(errMsg, ": ")
			if len(parts) > 1 {
				optionName := parts[1]
				if len(optionName) == 1 {
					return fmt.Errorf("flag needs an argument: -%s", optionName)
				} else {
					return fmt.Errorf("flag needs an argument: --%s", optionName)
				}
			}
		}
		return fmt.Errorf("flag needs an argument: %s", errMsg)
		
	default:
		// Return original error if no specific translation is needed
		return err
	}
}
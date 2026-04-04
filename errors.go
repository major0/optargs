package optargs

// UnknownOptionError is returned when the parser encounters an option
// that is not registered in either the short or long option maps.
type UnknownOptionError struct {
	Name    string // option name without dashes (e.g., "verbose", "x")
	IsShort bool   // true if this was a short option (-x), false for long (--verbose)
}

func (e *UnknownOptionError) Error() string {
	return "unknown option: " + e.Name
}

// MissingArgumentError is returned when an option with RequiredArgument
// has no argument provided.
type MissingArgumentError struct {
	Name    string // option name without dashes
	IsShort bool   // true if this was a short option
}

func (e *MissingArgumentError) Error() string {
	return "option requires an argument: " + e.Name
}

// AmbiguousOptionError is returned when a long option prefix matches
// multiple registered options at the same length.
type AmbiguousOptionError struct {
	Name    string   // the ambiguous input
	Matches []string // all matching option names (may be nil if not collected)
}

func (e *AmbiguousOptionError) Error() string {
	return "ambiguous option: " + e.Name
}

// UnexpectedArgumentError is returned when a NoArgument option receives
// a =value argument.
type UnexpectedArgumentError struct {
	Name string // option name without dashes
}

func (e *UnexpectedArgumentError) Error() string {
	return "option does not take an argument: " + e.Name
}

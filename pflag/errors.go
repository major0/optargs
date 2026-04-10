package pflag

import "fmt"

// NotExistError is the error returned when trying to access a flag that
// does not exist in the FlagSet.
type NotExistError struct {
	specifiedName       string
	specifiedShortnames string
}

func (e *NotExistError) Error() string {
	if e.specifiedShortnames != "" {
		return fmt.Sprintf("unknown shorthand flag: '%s' in -%s", e.specifiedName, e.specifiedShortnames)
	}
	return fmt.Sprintf("unknown flag: --%s", e.specifiedName)
}

// GetSpecifiedName returns the name of the flag (without dashes) as it
// appeared in the parsed arguments.
func (e *NotExistError) GetSpecifiedName() string { return e.specifiedName }

// GetSpecifiedShortnames returns the group of shorthand arguments
// (without dashes) that the flag appeared within. Empty if not in a group.
func (e *NotExistError) GetSpecifiedShortnames() string { return e.specifiedShortnames }

// InvalidValueError is the error returned when an invalid value is used
// for a flag.
type InvalidValueError struct {
	flag  *Flag
	value string
	err   error
}

func (e *InvalidValueError) Error() string {
	return fmt.Sprintf("invalid argument %q for \"--%s\" flag: %v", e.value, e.flag.Name, e.err)
}

// GetFlag returns the flag for which the error occurred.
func (e *InvalidValueError) GetFlag() *Flag { return e.flag }

// GetValue returns the invalid value that was provided.
func (e *InvalidValueError) GetValue() string { return e.value }

// Unwrap implements errors.Unwrap.
func (e *InvalidValueError) Unwrap() error { return e.err }

// ValueRequiredError is the error returned when a flag needs an argument but
// no argument was provided.
type ValueRequiredError struct {
	flag                *Flag
	specifiedName       string
	specifiedShortnames string
}

func (e *ValueRequiredError) Error() string {
	if e.specifiedShortnames != "" {
		return fmt.Sprintf("flag needs an argument: '%s' in -%s", e.specifiedName, e.specifiedShortnames)
	}
	return fmt.Sprintf("flag needs an argument: --%s", e.specifiedName)
}

// GetFlag returns the flag for which the error occurred.
func (e *ValueRequiredError) GetFlag() *Flag { return e.flag }

// GetSpecifiedName returns the name of the flag as it appeared in the arguments.
func (e *ValueRequiredError) GetSpecifiedName() string { return e.specifiedName }

// GetSpecifiedShortnames returns the shorthand group, or empty string.
func (e *ValueRequiredError) GetSpecifiedShortnames() string { return e.specifiedShortnames }

// InvalidSyntaxError is the error returned when a bad flag name is passed
// on the command line.
type InvalidSyntaxError struct {
	specifiedFlag string
}

func (e *InvalidSyntaxError) Error() string {
	return fmt.Sprintf("bad flag syntax: %s", e.specifiedFlag)
}

// GetSpecifiedFlag returns the exact flag (with dashes) as it appeared
// in the parsed arguments.
func (e *InvalidSyntaxError) GetSpecifiedFlag() string { return e.specifiedFlag }

// ParseErrorsAllowlist defines the parsing errors that can be ignored.
type ParseErrorsAllowlist struct {
	// UnknownFlags will ignore unknown flags errors and continue parsing.
	UnknownFlags bool
}

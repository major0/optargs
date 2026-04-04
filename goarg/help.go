package goarg

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/major0/optargs"
)

// HelpGenerator generates help text identical to alexflint/go-arg.
type HelpGenerator struct {
	metadata *StructMetadata
	config   Config
}

// NewHelpGenerator creates a new help generator.
func NewHelpGenerator(metadata *StructMetadata, config Config) *HelpGenerator {
	return &HelpGenerator{
		metadata: metadata,
		config:   config,
	}
}

// programName returns the configured program name or falls back to os.Args[0].
func (hg *HelpGenerator) programName() string {
	if hg.config.Program != "" {
		return hg.config.Program
	}
	return os.Args[0]
}

// WriteHelp writes help text to the provided writer.
//
//nolint:gocognit,gocyclo,cyclop,funlen // help text generation requires conditional formatting for each field type
func (hg *HelpGenerator) WriteHelp(w io.Writer) error {
	if hg.metadata == nil {
		fmt.Fprintln(w, "No help available")
		return nil
	}

	program := hg.programName()

	// Usage line
	fmt.Fprintf(w, "Usage: %s", program)

	// Add subcommands if available
	if len(hg.metadata.Subcommands) > 0 {
		fmt.Fprint(w, " COMMAND")
	}

	// Add options placeholder if we have options
	if len(hg.metadata.Options) > 0 {
		fmt.Fprint(w, " [OPTIONS]")
	}

	// Add positional arguments
	for i := range hg.metadata.Positionals {
		field := &hg.metadata.Positionals[i]
		if field.Required {
			fmt.Fprintf(w, " %s", strings.ToUpper(field.Name))
		} else {
			fmt.Fprintf(w, " [%s]", strings.ToUpper(field.Name))
		}
	}

	fmt.Fprintln(w)

	// Add description if available
	if hg.config.Description != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, hg.config.Description)
	}

	// Add positional arguments section
	if len(hg.metadata.Positionals) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Positional arguments:")
		for i := range hg.metadata.Positionals {
			field := &hg.metadata.Positionals[i]
			name := strings.ToUpper(field.Name)
			if field.Help != "" {
				fmt.Fprintf(w, "  %-20s %s\n", name, field.Help)
			} else {
				fmt.Fprintf(w, "  %s\n", name)
			}
		}
	}

	// Add options section
	if len(hg.metadata.Options) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")

		for i := range hg.metadata.Options {
			field := &hg.metadata.Options[i]
			var optStr string
			switch {
			case field.Short != "" && field.Long != "":
				optStr = fmt.Sprintf("  -%s, --%s", field.Short, field.Long)
			case field.Short != "":
				optStr = fmt.Sprintf("  -%s", field.Short)
			case field.Long != "":
				optStr = fmt.Sprintf("      --%s", field.Long)
			}

			// Add argument placeholder for options that take arguments
			if field.ArgType != 0 { // NoArgument is 0
				argName := strings.ToUpper(field.Name)
				optStr += fmt.Sprintf(" %s", argName)
			}

			// Append prefix pair forms
			var optStrSb110 strings.Builder
			for _, pp := range field.Prefixes {
				fmt.Fprintf(&optStrSb110, ", --%s-%s, --%s-%s", pp.True, field.Long, pp.False, field.Long)
			}
			optStr += optStrSb110.String()
			// Append negatable form
			if field.Negatable {
				optStr += fmt.Sprintf(", --no-%s", field.Long)
			}

			if field.Help != "" {
				fmt.Fprintf(w, "%-30s %s", optStr, field.Help)
			} else {
				fmt.Fprint(w, optStr)
			}

			// Add default value if available
			if field.Default != nil && field.Default != "" {
				fmt.Fprintf(w, " (default: %v)", field.Default)
			}

			fmt.Fprintln(w)
		}

		// Add help option
		fmt.Fprintf(w, "%-30s %s\n", "  -h, --help", "show this help message and exit")
	}

	// Add subcommands section
	if len(hg.metadata.Subcommands) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Commands:")
		for cmdName := range hg.metadata.Subcommands {
			help := ""
			// Get help text from the SubcommandHelp map
			if hg.metadata.SubcommandHelp != nil {
				help = hg.metadata.SubcommandHelp[cmdName]
			}
			if help != "" {
				fmt.Fprintf(w, "  %-20s %s\n", cmdName, help)
			} else {
				fmt.Fprintf(w, "  %s\n", cmdName)
			}
		}
	}

	// Add version if available
	if hg.config.Version != "" {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Version: %s\n", hg.config.Version)
	}

	// Add environment-only variables section
	if len(hg.metadata.EnvOnly) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Environment variables:")
		for i := range hg.metadata.EnvOnly {
			field := &hg.metadata.EnvOnly[i]
			label := fmt.Sprintf("  %s", field.Env)
			if field.Help != "" {
				fmt.Fprintf(w, "%-30s %s", label, field.Help)
			} else {
				fmt.Fprint(w, label)
			}
			if field.Required {
				fmt.Fprint(w, " (required)")
			}
			if field.Default != nil && field.Default != "" {
				fmt.Fprintf(w, " (default: %v)", field.Default)
			}
			fmt.Fprintln(w)
		}
	}

	// Add epilogue if available
	if hg.config.Epilogue != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, hg.config.Epilogue)
	}

	return nil
}

// WriteUsage writes usage text to the provided writer.
//

func (hg *HelpGenerator) WriteUsage(w io.Writer) error {
	program := hg.programName()

	fmt.Fprintf(w, "Usage: %s", program)

	// Add subcommands if available
	if hg.metadata != nil && len(hg.metadata.Subcommands) > 0 {
		fmt.Fprint(w, " COMMAND")
	}

	if hg.metadata != nil {
		// Add options placeholder if we have options
		if len(hg.metadata.Options) > 0 {
			fmt.Fprint(w, " [OPTIONS]")
		}

		// Add positional arguments
		for i := range hg.metadata.Positionals {
			field := &hg.metadata.Positionals[i]
			if field.Required {
				fmt.Fprintf(w, " %s", strings.ToUpper(field.Name))
			} else {
				fmt.Fprintf(w, " [%s]", strings.ToUpper(field.Name))
			}
		}
	}

	fmt.Fprintln(w)
	return nil
}

// ErrorTranslator translates OptArgs Core errors to go-arg format.
type ErrorTranslator struct{}

// TranslateError translates an error to go-arg compatible format.
//
//nolint:gocyclo,cyclop // error translation maps many optargs error types to go-arg format
func (et *ErrorTranslator) TranslateError(err error, context ParseContext) error {
	if err == nil {
		return nil
	}

	// Typed error classification — use errors.As() for core parser errors.
	var unknownErr *optargs.UnknownOptionError
	if errors.As(err, &unknownErr) {
		option := unknownErr.Name
		if unknownErr.IsShort {
			option = "-" + option
		} else {
			option = "--" + option
		}
		return fmt.Errorf("unrecognized argument: %s", option)
	}

	var missingErr *optargs.MissingArgumentError
	if errors.As(err, &missingErr) {
		option := missingErr.Name
		if missingErr.IsShort {
			option = "-" + option
		} else {
			option = "--" + option
		}
		return fmt.Errorf("option requires an argument: %s", option)
	}

	errMsg := err.Error()

	// Remove common prefixes that are internal implementation details
	errMsg = strings.TrimPrefix(errMsg, "parsing error: ")
	errMsg = strings.TrimPrefix(errMsg, "failed to set field ")
	errMsg = strings.TrimPrefix(errMsg, "failed to convert value ")

	// Extract field name from error messages like "failed to set field Count: ..."
	if strings.Contains(errMsg, ": failed to convert value") {
		parts := strings.Split(errMsg, ": failed to convert value")
		if len(parts) > 0 {
			fieldName := strings.TrimSpace(parts[0])
			if context.FieldName == "" {
				context.FieldName = fieldName
			}
		}
	}

	// Handle wrapped positional argument errors
	if strings.Contains(errMsg, "missing required positional argument: ") {
		parts := strings.Split(errMsg, "missing required positional argument: ")
		if len(parts) > 1 {
			fieldName := strings.TrimSpace(parts[1])
			return fmt.Errorf("%s is required", fieldName)
		}
	}

	// Translate remaining errors (goarg-internal, not core parser errors)
	switch {
	case strings.Contains(errMsg, "invalid argument") || strings.Contains(errMsg, "invalid syntax") || strings.Contains(errMsg, "invalid value"):
		if context.FieldName != "" {
			return fmt.Errorf("invalid argument for --%s", context.FieldName)
		}
		return errors.New("invalid argument")

	case strings.Contains(errMsg, "missing required") || strings.Contains(errMsg, " is required"):
		if strings.Contains(errMsg, " is required") {
			parts := strings.Split(errMsg, " is required")
			if len(parts) > 0 {
				fieldName := strings.TrimSpace(parts[0])
				fieldName = strings.TrimPrefix(fieldName, "--")
				fieldName = strings.TrimPrefix(fieldName, "-")
				return fmt.Errorf("required argument missing: %s", fieldName)
			}
		}
		if context.FieldName != "" {
			return fmt.Errorf("required argument missing: %s", context.FieldName)
		}
		return errors.New("required argument missing")

	case strings.Contains(errMsg, "too many"):
		return errors.New("too many positional arguments")

	case strings.Contains(errMsg, "not enough"):
		return errors.New("not enough positional arguments")

	case strings.HasPrefix(errMsg, "--") || strings.HasPrefix(errMsg, "-"):
		return fmt.Errorf("unrecognized argument: %s", errMsg)

	default:
		cleanMsg := errMsg
		if strings.Contains(cleanMsg, ": ") {
			parts := strings.Split(cleanMsg, ": ")
			if len(parts) > 1 {
				cleanMsg = parts[len(parts)-1]
			}
		}
		return errors.New(cleanMsg)
	}
}

// ParseContext provides context for error translation.
type ParseContext struct {
	StructType reflect.Type
	FieldName  string
}

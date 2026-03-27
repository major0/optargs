package goarg

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// HelpGenerator generates help text identical to alexflint/go-arg
type HelpGenerator struct {
	metadata *StructMetadata
	config   Config
}

// NewHelpGenerator creates a new help generator
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

// WriteHelp writes help text to the provided writer
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
	for _, field := range hg.metadata.Positionals {
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
		for _, field := range hg.metadata.Positionals {
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

		for _, field := range hg.metadata.Options {
			var optStr string
			if field.Short != "" && field.Long != "" {
				optStr = fmt.Sprintf("  -%s, --%s", field.Short, field.Long)
			} else if field.Short != "" {
				optStr = fmt.Sprintf("  -%s", field.Short)
			} else if field.Long != "" {
				optStr = fmt.Sprintf("      --%s", field.Long)
			}

			// Add argument placeholder for options that take arguments
			if field.ArgType != 0 { // NoArgument is 0
				argName := strings.ToUpper(field.Name)
				optStr += fmt.Sprintf(" %s", argName)
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

	// Add epilogue if available
	if hg.config.Epilogue != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, hg.config.Epilogue)
	}

	return nil
}

// WriteUsage writes usage text to the provided writer
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
		for _, field := range hg.metadata.Positionals {
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

// ErrorTranslator translates OptArgs Core errors to go-arg format
type ErrorTranslator struct{}

// TranslateError translates an error to go-arg compatible format
func (et *ErrorTranslator) TranslateError(err error, context ParseContext) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	originalMsg := errMsg

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

	// Translate common OptArgs Core errors to alexflint/go-arg format
	switch {
	case strings.Contains(originalMsg, "unknown option") || strings.Contains(errMsg, "unknown option"):
		// Extract option name from error
		option := extractOptionFromError(originalMsg)
		return fmt.Errorf("unrecognized argument: %s", option)

	case strings.Contains(originalMsg, "option requires an argument") || strings.Contains(errMsg, "option requires an argument"):
		// Extract option name from error
		option := extractOptionFromError(originalMsg)
		return fmt.Errorf("option requires an argument: %s", option)

	case strings.Contains(errMsg, "invalid argument") || strings.Contains(errMsg, "invalid syntax") || strings.Contains(errMsg, "invalid value"):
		// Handle type conversion errors from optargs.Convert
		if context.FieldName != "" {
			return fmt.Errorf("invalid argument for --%s", context.FieldName)
		}
		return fmt.Errorf("invalid argument")

	case strings.Contains(errMsg, "missing required") || strings.Contains(errMsg, "is required"):
		// Handle missing required arguments
		if strings.Contains(errMsg, " is required") {
			// Extract field name and convert to alexflint/go-arg format
			parts := strings.Split(errMsg, " is required")
			if len(parts) > 0 {
				fieldName := strings.TrimSpace(parts[0])
				// Remove leading dashes if present
				fieldName = strings.TrimPrefix(fieldName, "--")
				fieldName = strings.TrimPrefix(fieldName, "-")
				return fmt.Errorf("required argument missing: %s", fieldName)
			}
		}
		if context.FieldName != "" {
			return fmt.Errorf("required argument missing: %s", context.FieldName)
		}
		return fmt.Errorf("required argument missing")

	case strings.Contains(errMsg, "too many"):
		return fmt.Errorf("too many positional arguments")

	case strings.Contains(errMsg, "not enough"):
		return fmt.Errorf("not enough positional arguments")

	case strings.HasPrefix(errMsg, "--") || strings.HasPrefix(errMsg, "-"):
		// This looks like an option name that was returned as an error
		// This can happen when OptArgs Core returns just the option name
		return fmt.Errorf("unrecognized argument: %s", errMsg)

	default:
		// For unrecognized errors, clean up and return
		// Remove internal implementation details
		cleanMsg := errMsg
		if strings.Contains(cleanMsg, ": ") {
			// Try to extract the most relevant part of the error
			parts := strings.Split(cleanMsg, ": ")
			if len(parts) > 1 {
				// Take the last part which is usually the most specific error
				cleanMsg = parts[len(parts)-1]
			}
		}
		return fmt.Errorf("%s", cleanMsg)
	}
}

// extractOptionFromError extracts the option name from an error message
func extractOptionFromError(errMsg string) string {
	errMsg = strings.TrimPrefix(errMsg, "parsing error: ")

	// Look for patterns like "--option" or "-o"
	if idx := strings.Index(errMsg, "--"); idx != -1 {
		return extractWord(errMsg, idx)
	}
	if idx := strings.Index(errMsg, "-"); idx != -1 {
		return extractWord(errMsg, idx)
	}

	// Try known "prefix: optname" patterns
	for _, prefix := range []string{"unknown option: ", "option requires an argument: "} {
		if name, ok := extractAfterPrefix(errMsg, prefix); ok {
			return name
		}
	}

	return errMsg
}

// extractWord returns the contiguous non-whitespace, non-colon token starting at idx.
func extractWord(s string, idx int) string {
	end := idx + 1
	for end < len(s) && s[end] != ' ' && s[end] != '\t' && s[end] != '\n' && s[end] != ':' {
		end++
	}
	return s[idx:end]
}

// extractAfterPrefix extracts an option name after a known error prefix,
// adding dash prefixes if needed.
func extractAfterPrefix(errMsg, prefix string) (string, bool) {
	if !strings.Contains(errMsg, prefix) {
		return "", false
	}
	parts := strings.Split(errMsg, prefix)
	if len(parts) < 2 {
		return "", false
	}
	optName := strings.TrimSpace(parts[1])
	if strings.HasPrefix(optName, "-") {
		return optName, true
	}
	if len(optName) == 1 {
		return "-" + optName, true
	}
	return "--" + optName, true
}

// ParseContext provides context for error translation
type ParseContext struct {
	StructType reflect.Type
	FieldName  string
}

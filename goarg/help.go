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

// WriteHelp writes help text to the provided writer
func (hg *HelpGenerator) WriteHelp(w io.Writer) error {
	if hg.metadata == nil {
		fmt.Fprintln(w, "No help available")
		return nil
	}

	// Generate help text compatible with alexflint/go-arg format
	program := hg.config.Program
	if program == "" {
		program = os.Args[0]
	}

	// Usage line
	fmt.Fprintf(w, "Usage: %s", program)

	// Add subcommands if available
	if len(hg.metadata.Subcommands) > 0 {
		fmt.Fprint(w, " COMMAND")
	}

	// Add options placeholder if we have non-positional fields
	hasOptions := false
	for _, field := range hg.metadata.Fields {
		if !field.Positional && !field.IsSubcommand {
			hasOptions = true
			break
		}
	}
	if hasOptions {
		fmt.Fprint(w, " [OPTIONS]")
	}

	// Add positional arguments
	for _, field := range hg.metadata.Fields {
		if field.Positional {
			if field.Required {
				fmt.Fprintf(w, " %s", strings.ToUpper(field.Name))
			} else {
				fmt.Fprintf(w, " [%s]", strings.ToUpper(field.Name))
			}
		}
	}

	fmt.Fprintln(w)

	// Add description if available
	if hg.config.Description != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, hg.config.Description)
	}

	// Add positional arguments section
	hasPositionals := false
	for _, field := range hg.metadata.Fields {
		if field.Positional {
			hasPositionals = true
			break
		}
	}
	if hasPositionals {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Positional arguments:")
		for _, field := range hg.metadata.Fields {
			if field.Positional {
				name := strings.ToUpper(field.Name)
				if field.Help != "" {
					fmt.Fprintf(w, "  %-20s %s\n", name, field.Help)
				} else {
					fmt.Fprintf(w, "  %s\n", name)
				}
			}
		}
	}

	// Add options section
	if hasOptions {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")

		for _, field := range hg.metadata.Fields {
			if field.Positional || field.IsSubcommand {
				continue
			}

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
				if field.Long != "" {
					optStr += fmt.Sprintf(" %s", argName)
				} else {
					optStr += fmt.Sprintf(" %s", argName)
				}
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
			// Try to find help text for the subcommand from the parent struct field
			for _, field := range hg.metadata.Fields {
				if field.IsSubcommand && field.SubcommandName == cmdName {
					help = field.Help
					break
				}
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

	return nil
}

// WriteUsage writes usage text to the provided writer
func (hg *HelpGenerator) WriteUsage(w io.Writer) error {
	program := hg.config.Program
	if program == "" {
		program = os.Args[0]
	}

	fmt.Fprintf(w, "Usage: %s", program)

	// Add subcommands if available
	if hg.metadata != nil && len(hg.metadata.Subcommands) > 0 {
		fmt.Fprint(w, " COMMAND")
	}

	// Add options placeholder if we have non-positional fields
	if hg.metadata != nil {
		hasOptions := false
		for _, field := range hg.metadata.Fields {
			if !field.Positional && !field.IsSubcommand {
				hasOptions = true
				break
			}
		}
		if hasOptions {
			fmt.Fprint(w, " [OPTIONS]")
		}

		// Add positional arguments
		for _, field := range hg.metadata.Fields {
			if field.Positional {
				if field.Required {
					fmt.Fprintf(w, " %s", strings.ToUpper(field.Name))
				} else {
					fmt.Fprintf(w, " [%s]", strings.ToUpper(field.Name))
				}
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

	case strings.Contains(errMsg, "invalid argument") || strings.Contains(errMsg, "invalid syntax"):
		// Handle type conversion errors
		if context.FieldName != "" {
			return fmt.Errorf("invalid argument for --%s", context.FieldName)
		}
		return fmt.Errorf("invalid argument")

	case strings.Contains(errMsg, "missing required") || strings.Contains(errMsg, "is required"):
		// Handle missing required arguments
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
	// Clean up the error message first
	errMsg = strings.TrimPrefix(errMsg, "parsing error: ")

	// Look for patterns like "--option" or "-o"
	if idx := strings.Index(errMsg, "--"); idx != -1 {
		start := idx
		end := start + 2
		for end < len(errMsg) && (errMsg[end] != ' ' && errMsg[end] != '\t' && errMsg[end] != '\n' && errMsg[end] != ':') {
			end++
		}
		return errMsg[start:end]
	}

	if idx := strings.Index(errMsg, "-"); idx != -1 {
		start := idx
		end := start + 1
		for end < len(errMsg) && (errMsg[end] != ' ' && errMsg[end] != '\t' && errMsg[end] != '\n' && errMsg[end] != ':') {
			end++
		}
		return errMsg[start:end]
	}

	// If no option found, try to extract from "unknown option: optname" format
	if strings.Contains(errMsg, "unknown option: ") {
		parts := strings.Split(errMsg, "unknown option: ")
		if len(parts) > 1 {
			optName := strings.TrimSpace(parts[1])
			// Add -- prefix if not present
			if !strings.HasPrefix(optName, "-") {
				return "--" + optName
			}
			return optName
		}
	}

	// If no option found, try to extract from "option requires an argument: optname" format
	if strings.Contains(errMsg, "option requires an argument: ") {
		parts := strings.Split(errMsg, "option requires an argument: ")
		if len(parts) > 1 {
			optName := strings.TrimSpace(parts[1])
			// Add -- prefix if not present
			if !strings.HasPrefix(optName, "-") {
				return "--" + optName
			}
			return optName
		}
	}

	// If no option found, return the original message
	return errMsg
}

// ParseContext provides context for error translation
type ParseContext struct {
	StructType reflect.Type
	FieldName  string
	TagValue   string
}

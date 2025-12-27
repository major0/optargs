# Design Document: OptArgs Core

## Overview

The OptArgs core provides a complete, POSIX/GNU-compliant command-line argument parsing engine that serves as the foundation for all higher-level wrapper interfaces. The design emphasizes correctness, performance, and extensibility while maintaining strict adherence to established standards.

## Architecture

The core architecture follows a layered approach with clear separation of concerns:

```
┌─────────────────────────────────────┐
│           Public API Layer          │
├─────────────────────────────────────┤
│         Parser Engine Layer         │
├─────────────────────────────────────┤
│       Option Registry Layer         │
├─────────────────────────────────────┤
│      Argument Processing Layer      │
└─────────────────────────────────────┘
```

### Key Design Principles

1. **Standards Compliance**: Strict adherence to POSIX getopt(3) and GNU extensions
2. **Zero Dependencies**: Core functionality requires no external dependencies
3. **Memory Efficiency**: Iterator-based processing to minimize memory allocation
4. **Error Transparency**: Clear, actionable error messages with context
5. **Extensibility**: Clean interfaces for wrapper implementations

## Components and Interfaces

### Parser Engine (Existing Implementation)

The central parsing component that orchestrates argument processing. The current implementation provides a robust foundation:

```go
type Parser struct {
    Args       []string              // Remaining arguments to process
    nonOpts    []string              // Non-option arguments collected during parsing
    shortOpts  map[byte]*Flag        // Registry of short options
    longOpts   map[string]*Flag      // Registry of long options
    config     ParserConfig          // Parser behavior configuration
    lockConfig bool                  // Prevents configuration changes during parsing
}

type ParserConfig struct {
    enableErrors    bool              // Enable automatic error reporting
    parseMode      ParseMode         // Parsing behavior mode
    shortCaseIgnore bool              // Case-insensitive short options
    longCaseIgnore  bool              // Case-insensitive long options (default: true)
    longOptsOnly    bool              // Long-options-only mode
    gnuWords        bool              // GNU -W word extension support
}

type ParseMode int
const (
    ParseDefault ParseMode = iota     // Standard parsing with option reordering
    ParseNonOpts                      // Treat non-options as arguments to option 1
    ParsePosixlyCorrect              // Stop at first non-option
)
```

### Option Registry (Existing Implementation)

The current implementation uses efficient map-based storage for option definitions:

```go
type Flag struct {
    Name   string                    // Option name (single char for short, full name for long)
    HasArg ArgType                   // Argument requirement specification
}

type ArgType int
const (
    NoArgument ArgType = iota         // Option takes no argument
    RequiredArgument                  // Option requires an argument
    OptionalArgument                  // Option may take an argument
)
```

### Option Processing (Existing Implementation)

The current iterator-based approach provides memory-efficient processing:

```go
type Option struct {
    Name   string                    // Resolved option name
    HasArg bool                      // Whether an argument was provided
    Arg    string                    // Argument value (if any)
}

// Iterator-based option processing for memory efficiency
func (p *Parser) Options() iter.Seq2[Option, error]
```

### Core API Functions (Existing Implementation)

The main entry points maintain full compatibility with getopt conventions:

```go
func GetOpt(args []string, optstring string) (*Parser, error)
func GetOptLong(args []string, optstring string, longopts []Flag) (*Parser, error)
func GetOptLongOnly(args []string, optstring string, longopts []Flag) (*Parser, error)

// Internal function that handles the common parsing logic
func getOpt(args []string, optstring string, longopts []Flag, longOnly bool) (*Parser, error)
```

### Parsing Logic (Existing Implementation)

The core parsing algorithms are already implemented with sophisticated logic:

#### Short Option Processing
- **Option Compaction**: Handles `-abc` as `-a -b -c`
- **Argument Assignment**: Assigns arguments to the last option that accepts them
- **Character Validation**: Supports all printable ASCII except `:`, `;`, `-`

#### Long Option Processing  
- **Flexible Matching**: Supports partial matches when unambiguous
- **Equals Syntax**: Handles both `--option=value` and `--option value`
- **Complex Names**: Allows `=` in option names for advanced patterns

#### GNU Extensions
- **W-Option Support**: Transforms `-W foo` into `--foo`
- **Case Insensitivity**: Configurable case handling for long options
- **Behavior Flags**: Supports `:`, `+`, `-` prefixes in optstring

### Utility Functions (Existing Implementation)

Helper functions provide cross-platform compatibility:

```go
// Custom isGraph implementation for consistent behavior across platforms
func isGraph(c byte) bool

// Case-insensitive string operations with performance optimization
func hasPrefix(s, prefix string, ignoreCase bool) bool
func trimPrefix(s, prefix string, ignoreCase bool) string
```

## Data Models

### Argument Stream Processing (Existing Implementation)

The current parser processes arguments as a stream with sophisticated state management:

- **Iterator Pattern**: Uses Go 1.23's `iter.Seq2[Option, error]` for memory-efficient processing
- **State Preservation**: Maintains parsing context across option compaction scenarios
- **Non-Option Buffering**: Collects non-option arguments in `nonOpts` slice for later processing
- **Argument Reordering**: Supports GNU-style argument reordering in default mode

### Option String Grammar (Existing Implementation)

The current optstring parser implements full POSIX/GNU compatibility:

```
optstring := [behavior_flags] option_definitions
behavior_flags := ':' | '+' | '-' | combination_thereof
option_definitions := option_char [arg_spec]*
arg_spec := ':' | '::' | ';' (for W option only)

Examples:
- "abc"           # Three no-argument options
- "a:b::c"        # Required arg, optional arg, no arg
- ":+abc"         # Silent errors + POSIXLY_CORRECT mode
- "W;abc"         # GNU word extension enabled
```

### Long Option Specifications (Existing Implementation)

The current implementation supports complex long option patterns:

```
long_option := name ['=' value]
name := [[:graph:]] - [[:space:]]  # Any printable non-space character
value := any_string_without_leading_hyphen

Special Cases:
- "--foo=bar=baz" → name="foo=bar", value="baz"
- "--option"      → Looks for next arg if required
- "--opt="        → Empty string argument
```

### Advanced Parsing Features (Existing Implementation)

The current code includes sophisticated parsing logic:

#### Option Compaction Algorithm
```go
// Handles cases like "-abc123 arg" where:
// - 'a' and 'b' are no-argument options
// - 'c' takes an optional argument and gets "123"
// - "arg" becomes a non-option argument
for word := p.Args[0][1:]; len(word) > 0; {
    p.Args, word, option, err = p.findShortOpt(word[0], word[1:], p.Args[1:])
    // ... processing logic
}
```

#### Long Option Matching Algorithm
```go
// Supports partial matching and complex name patterns
// Handles ambiguity resolution and optimal match selection
for opt := range p.longOpts {
    // Complex matching logic with case sensitivity options
    // Supports '=' in option names for advanced patterns
}
```

#### GNU W-Option Extension
```go
// Transforms "-W foo" into "--foo" for compatibility
if option.Name == "W" && p.config.gnuWords {
    option.Name = option.Arg
}
```

## Error Handling (Existing Implementation)

### Error Categories

The current implementation provides comprehensive error handling:

1. **Configuration Errors**: Invalid optstring or flag definitions
   ```go
   return nil, errors.New("Invalid option character: " + string(c))
   ```

2. **Parsing Errors**: Malformed arguments or missing required values
   ```go
   return args, option, p.optError("option requires an argument: " + string(c))
   ```

3. **Validation Errors**: Character validation and constraint checking
   ```go
   if !isGraph(c) {
       return nil, parser.optErrorf("Invalid short option: %c", c)
   }
   ```

### Error Reporting Modes (Existing Implementation)

The current system supports flexible error handling:

- **Verbose Mode**: Full error messages with slog integration (default)
  ```go
  func (p *Parser) optError(msg string) error {
      if p.config.enableErrors {
          slog.Error(msg)
      }
      return errors.New(msg)
  }
  ```

- **Silent Mode**: Minimal error reporting (enabled by ':' prefix in optstring)
- **Structured Errors**: Consistent error formatting with context information

### Error Context and Recovery

The implementation provides detailed error context:

- **Option Identification**: Errors include the specific option that caused the problem
- **Argument Context**: Shows the argument being processed when errors occur
- **State Preservation**: Parser state remains consistent even after errors

## Testing Strategy

### Unit Testing Approach

Following the testing standards steering document:

- **100% line coverage** for all parsing functions
- **100% branch coverage** for conditional logic
- **Property-based testing** for parsing correctness
- **Table-driven tests** for multiple input scenarios

### Property-Based Testing Framework

Using Go's `testing/quick` package for property validation:

- **Round-trip properties**: Parse → Generate → Parse → Verify equivalence
- **Invariant properties**: Ensure parsing rules are consistently applied
- **Metamorphic properties**: Verify relationships between different input forms

### POSIX Compliance Testing

Integration with existing POSIX test suite:

- Validate against all examples in `posix/` directory
- Cross-reference with GNU getopt behavior
- Test edge cases and corner conditions

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Based on the prework analysis, the following properties ensure correctness of the OptArgs core implementation:

### Property 1: POSIX/GNU Specification Compliance
*For any* valid POSIX optstring and GNU long option specification, the parser should produce results that match the behavior of the reference GNU getopt implementation
**Validates: Requirements 1.1, 2.1, 3.1**

### Property 2: Option Compaction and Argument Assignment
*For any* combination of compacted short options with arguments, the parser should assign arguments to the last option that accepts them and expand compaction correctly
**Validates: Requirements 1.2, 4.2**

### Property 3: Argument Type Handling
*For any* option string containing colon specifications, the parser should correctly handle required arguments (:), optional arguments (::), and no-argument options according to POSIX rules
**Validates: Requirements 1.3**

### Property 4: Option Termination Behavior
*For any* argument list containing `--`, the parser should stop processing options at that point and treat all subsequent arguments as non-options
**Validates: Requirements 1.5**

### Property 5: Long Option Syntax Support
*For any* valid long option, the parser should correctly handle both `--option=value` and `--option value` syntax forms
**Validates: Requirements 2.2, 2.3**

### Property 6: Case Sensitivity Handling
*For any* long option name, the parser should handle case variations according to the configured case sensitivity settings
**Validates: Requirements 2.4**

### Property 7: Partial Long Option Matching
*For any* unambiguous partial long option match, the parser should resolve to the correct full option name
**Validates: Requirements 2.5**

### Property 8: Long-Only Mode Behavior
*For any* single-dash option in long-only mode, the parser should treat multi-character options as long options and fall back to short option parsing for single characters
**Validates: Requirements 3.2, 3.3**

### Property 9: GNU W-Extension Support
*For any* `-W word` pattern when GNU words are enabled, the parser should transform it to `--word`
**Validates: Requirements 4.1**

### Property 10: Negative Argument Support
*For any* option that requires an argument, the parser should accept arguments beginning with `-` when explicitly provided
**Validates: Requirements 4.3**

### Property 11: Character Validation
*For any* printable ASCII character except `:`, `;`, `-`, the parser should accept it as a valid short option character
**Validates: Requirements 4.4**

### Property 12: Option Redefinition Handling
*For any* optstring where options are redefined, the parser should use the last definition encountered
**Validates: Requirements 4.5**

### Property 13: Error Reporting Accuracy
*For any* missing required argument error, the error message should identify the specific option that requires the argument
**Validates: Requirements 5.2**

### Property 14: Silent Error Mode
*For any* optstring beginning with `:`, the parser should suppress automatic error logging while still returning errors
**Validates: Requirements 5.3**

### Property 15: Iterator Correctness
*For any* valid argument list, the iterator should yield all options exactly once and preserve non-option arguments correctly
**Validates: Requirements 6.4**

### Property 16: Environment Variable Behavior
*For any* parsing session, when POSIXLY_CORRECT is set, the parser should stop at the first non-option argument
**Validates: Requirements 1.4**

### Property 17: Ambiguity Resolution
*For any* ambiguous long option input, the parser should handle it according to GNU specifications for ambiguity resolution
**Validates: Requirements 3.5**
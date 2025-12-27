# Design Document: PFlags Wrapper

## Overview

The PFlags wrapper provides a drop-in replacement for spf13/pflag that maintains complete API compatibility while leveraging OptArgs Core for POSIX/GNU compliance. This design bridges the gap between pflag's familiar interface and OptArgs Core's robust parsing engine, delivering both ease of migration and correctness guarantees.

## Architecture

The wrapper architecture creates a compatibility layer that translates pflag API calls into OptArgs Core operations:

```
┌─────────────────────────────────────┐
│        PFlags API Layer             │
│  (FlagSet, StringVar, Parse, etc.)  │
├─────────────────────────────────────┤
│      Flag Registry & Management     │
│   (Type conversion, validation)     │
├─────────────────────────────────────┤
│       OptArgs Core Integration      │
│    (Parser, Option processing)      │
├─────────────────────────────────────┤
│         OptArgs Core Engine         │
│   (POSIX/GNU compliant parsing)     │
└─────────────────────────────────────┘
```

### Design Principles

1. **API Compatibility**: Maintain exact method signatures and behavior of spf13/pflag
2. **Zero Breaking Changes**: Existing pflag code should work without modification
3. **Enhanced Reliability**: Leverage OptArgs Core's standards compliance
4. **Performance Preservation**: Match or exceed pflag's performance characteristics
5. **Clear Error Messages**: Provide helpful, actionable error messages

## Components and Interfaces

### FlagSet Implementation

The core component that manages flag definitions and parsing state:

```go
type FlagSet struct {
    name        string                    // Name of the flag set
    parsed      bool                      // Whether Parse() has been called
    args        []string                  // Arguments to parse
    flags       map[string]*Flag          // All defined flags by name
    shorthand   map[string]string         // Shorthand to name mapping
    values      map[string]Value          // Flag values by name
    parser      *optargs.Parser           // OptArgs Core parser instance
    output      io.Writer                 // Output destination for usage/errors
    usage       func()                    // Custom usage function
}

// NewFlagSet creates a new flag set with the given name
func NewFlagSet(name string, errorHandling ErrorHandling) *FlagSet

// Core parsing method that delegates to OptArgs Core
func (f *FlagSet) Parse(arguments []string) error
```

### Flag Definition and Value Management

Flag definitions maintain compatibility with pflag's Flag struct:

```go
type Flag struct {
    Name        string                    // Long flag name
    Shorthand   string                    // Single character shorthand
    Usage       string                    // Help text
    Value       Value                     // Current value (implements Value interface)
    DefValue    string                    // Default value as string
    Changed     bool                      // Whether flag was set during parsing
    Hidden      bool                      // Whether flag is hidden from help
    Deprecated  string                    // Deprecation message
    Annotations map[string][]string       // Additional metadata
}

// Value interface matches pflag exactly
type Value interface {
    String() string
    Set(string) error
    Type() string
}
```

### Type-Specific Value Implementations

Each supported type has a corresponding Value implementation:

```go
// String values
type stringValue string
func (s *stringValue) Set(val string) error { *s = stringValue(val); return nil }
func (s *stringValue) String() string { return string(*s) }
func (s *stringValue) Type() string { return "string" }

// Integer values with validation
type intValue int
func (i *intValue) Set(val string) error {
    v, err := strconv.Atoi(val)
    if err != nil {
        return fmt.Errorf("invalid syntax for integer flag: %s", val)
    }
    *i = intValue(v)
    return nil
}

// Boolean values with enhanced parsing
type boolValue bool
func (b *boolValue) Set(val string) error {
    v, err := strconv.ParseBool(val)
    if err != nil {
        return fmt.Errorf("invalid boolean value '%s'", val)
    }
    *b = boolValue(v)
    return nil
}

// Slice types supporting both comma-separated and repeated flags
type stringSliceValue []string
func (s *stringSliceValue) Set(val string) error {
    // Support both --flag=a,b,c and --flag=a --flag=b --flag=c
    if strings.Contains(val, ",") {
        *s = append(*s, strings.Split(val, ",")...)
    } else {
        *s = append(*s, val)
    }
    return nil
}
```

### OptArgs Core Integration

The integration layer translates between pflag concepts and OptArgs Core:

```go
type CoreIntegration struct {
    flagSet     *FlagSet                  // Parent flag set
    parser      *optargs.Parser           // OptArgs Core parser
    flagMap     map[string]*optargs.Flag  // OptArgs flags by name
}

// Convert pflag Flag definitions to OptArgs Core format
func (ci *CoreIntegration) registerFlag(flag *Flag) error {
    argType := optargs.NoArgument
    if flag.Value.Type() != "bool" {
        argType = optargs.RequiredArgument
    }
    
    coreFlag := &optargs.Flag{
        Name:   flag.Name,
        HasArg: argType,
    }
    
    return ci.parser.AddLongOpt(flag.Name, coreFlag)
}

// Process parsed options and update flag values
func (ci *CoreIntegration) processOptions() error {
    for option, err := range ci.parser.Options() {
        if err != nil {
            return ci.translateError(err)
        }
        
        flag := ci.flagSet.flags[option.Name]
        if flag == nil {
            return fmt.Errorf("unknown flag: --%s", option.Name)
        }
        
        if err := flag.Value.Set(option.Arg); err != nil {
            return fmt.Errorf("invalid value for flag --%s: %v", option.Name, err)
        }
        
        flag.Changed = true
    }
    return nil
}
```

## Data Models

### Flag Registry Structure

The flag registry maintains all flag definitions and their current state:

```go
type FlagRegistry struct {
    flags       map[string]*Flag          // Primary flag storage
    shorthand   map[string]string         // Shorthand to name mapping
    order       []string                  // Definition order for help text
    normalized  map[string]string         // Normalized name mapping
}

// Flag lookup with normalization support
func (fr *FlagRegistry) Lookup(name string) *Flag {
    if normalized, exists := fr.normalized[name]; exists {
        name = normalized
    }
    return fr.flags[name]
}
```

### Parsing State Management

Track parsing progress and state for proper error handling:

```go
type ParseState struct {
    parsed      bool                      // Whether parsing is complete
    args        []string                  // Remaining non-flag arguments
    argsLenAtDash int                    // Position of "--" terminator
    interspersed bool                     // Whether flags can be interspersed
}
```

## Error Handling

### Error Translation Layer

Convert OptArgs Core errors to pflag-compatible messages:

```go
func (ci *CoreIntegration) translateError(err error) error {
    switch e := err.(type) {
    case *optargs.UnknownOptionError:
        return fmt.Errorf("unknown flag: %s", e.Option)
    case *optargs.MissingArgumentError:
        return fmt.Errorf("flag needs an argument: %s", e.Option)
    case *optargs.InvalidArgumentError:
        return fmt.Errorf("invalid argument for flag %s: %s", e.Option, e.Argument)
    default:
        return err
    }
}
```

### Validation and Type Checking

Comprehensive validation for all flag operations:

```go
func validateFlagDefinition(name, shorthand string) error {
    if name == "" {
        return errors.New("flag name cannot be empty")
    }
    if strings.Contains(name, "=") {
        return errors.New("flag name cannot contain '='")
    }
    if len(shorthand) > 1 {
        return errors.New("shorthand must be a single character")
    }
    return nil
}
```

## Testing Strategy

### Dual Testing Approach

The testing strategy combines unit tests for specific behaviors with property-based tests for comprehensive validation:

**Unit Tests:**
- API compatibility verification with spf13/pflag
- Error message format validation
- Edge case handling (empty values, special characters)
- Integration points with OptArgs Core

**Property-Based Tests:**
- Flag definition and parsing across all supported types
- Round-trip consistency (define → parse → retrieve)
- Error handling for invalid inputs
- Help text generation consistency

**Property-Based Testing Configuration:**
- Use Go's testing/quick framework for property tests
- Minimum 100 iterations per property test
- Each property test references its design document property
- Tag format: **Feature: pflags, Property {number}: {property_text}**

### Test Organization

```
pflags_test.go              // Core API compatibility tests
pflags_property_test.go     // Property-based correctness tests
pflags_integration_test.go  // OptArgs Core integration tests
pflags_benchmark_test.go    // Performance comparison tests
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Flag Creation Consistency
*For any* valid flag name, default value, and usage text, creating a flag with StringVar(), IntVar(), BoolVar(), Float64Var(), or DurationVar() should result in a flag that can be retrieved with the same name and contains the specified default value and usage text.
**Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5**

### Property 2: Shorthand Registration and Resolution
*For any* valid flag name and single-character shorthand, registering a flag with shorthand should make it accessible by both the long name and short character, and parsing with the short form should set the same flag as the long form.
**Validates: Requirements 2.1, 2.2, 2.3**

### Property 3: Slice Flag Value Accumulation
*For any* slice flag and sequence of values, providing values either comma-separated or through repeated flag usage should result in a slice containing all provided values in the correct order.
**Validates: Requirements 3.1, 3.2, 3.3, 3.4**

### Property 4: Boolean Flag Parsing Flexibility
*For any* boolean flag, it should accept no argument (defaulting to true), explicit true/false values, and negation syntax when the default is true.
**Validates: Requirements 4.1, 4.2, 4.3, 4.5**

### Property 5: FlagSet Isolation
*For any* two FlagSets with the same flag names, operations on one FlagSet should not affect the other, and parsing should only process flags defined in the target FlagSet.
**Validates: Requirements 5.1, 5.2, 5.3**

### Property 6: Parse State Consistency
*For any* FlagSet, the Parsed() method should return false before Parse() is called and true after successful parsing, and flag values should return defaults before parsing.
**Validates: Requirements 5.4, 5.5**

### Property 7: Custom Value Interface Integration
*For any* custom Value implementation, the Flag_Registry should accept it, call Set() during parsing with provided arguments, and use String() for help text display.
**Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.5**

### Property 8: Help Text Completeness
*For any* defined flag, the generated help text should include the flag name, shorthand (if present), default value, and usage description in the correct format.
**Validates: Requirements 7.2, 7.3, 7.4, 7.5**

### Property 9: Flag Introspection Accuracy
*For any* FlagSet, Lookup() should return the correct Flag object for existing flags and nil for non-existent ones, and VisitAll()/Visit() should call the provided function for the appropriate flags.
**Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5**

### Property 10: Type Validation Error Messages
*For any* typed flag receiving an invalid value, the error message should clearly indicate the expected type and the problematic value.
**Validates: Requirements 9.3**

### Property 11: OptArgs Core Integration Fidelity
*For any* argument sequence, parsing through the PFlags_Wrapper should produce the same results as parsing directly through OptArgs_Core, with errors translated to pflag-compatible messages.
**Validates: Requirements 10.1, 10.2**

### Property 12: POSIX Compliance Preservation
*For any* argument sequence containing POSIX-specific constructs (option termination with `--`, combined short options like `-abc`), the PFlags_Wrapper should handle them identically to OptArgs_Core.
**Validates: Requirements 10.3, 10.4, 10.5**

### Property 13: Slice Type Validation
*For any* slice flag receiving invalid values for its element type, the error should clearly indicate the type conversion failure.
**Validates: Requirements 3.5**
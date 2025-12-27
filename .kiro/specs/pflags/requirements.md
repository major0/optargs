# Requirements Document

## Introduction

The PFlags wrapper provides an interface compatible with spf13/pflag, allowing developers familiar with that library to easily migrate to OptArgs while gaining the benefits of complete POSIX/GNU compliance. This wrapper maintains API compatibility while leveraging the robust OptArgs core.

## Glossary

- **PFlags_Wrapper**: The pflag-compatible interface implementation
- **Flag_Registry**: Component that manages flag definitions and values
- **Value_Interface**: The interface for custom flag value types
- **Command_System**: Component that handles sub-command functionality for Cobra compatibility
- **OptArgs_Core**: The underlying POSIX/GNU compliant argument parser

## Requirements

### Requirement 1: Core Flag Type Support

**User Story:** As a developer using spf13/pflag, I want support for all basic flag types, so that I can define flags for common data types.

#### Acceptance Criteria

1. WHEN a developer calls `StringVar()`, THE Flag_Registry SHALL create a string flag with the specified name, default value, and usage text
2. WHEN a developer calls `IntVar()`, THE Flag_Registry SHALL create an integer flag with the specified name, default value, and usage text
3. WHEN a developer calls `BoolVar()`, THE Flag_Registry SHALL create a boolean flag with the specified name, default value, and usage text
4. WHEN a developer calls `Float64Var()`, THE Flag_Registry SHALL create a float64 flag with the specified name, default value, and usage text
5. WHEN a developer calls `DurationVar()`, THE Flag_Registry SHALL create a duration flag with the specified name, default value, and usage text

### Requirement 2: Short Option Support

**User Story:** As a developer, I want to define flags with short options, so that users can use convenient single-character flags.

#### Acceptance Criteria

1. WHEN a developer calls `StringP()` with name, shorthand, value, and usage, THE Flag_Registry SHALL create a flag accessible by both long and short forms
2. WHEN a user provides `-v` and the flag was defined with shorthand "v", THE Flag_Registry SHALL parse it as the corresponding long flag
3. WHEN a developer calls `FlagP()` with a Flag object containing shorthand, THE Flag_Registry SHALL register both long and short forms
4. IF a shorthand character is already registered, THEN THE Flag_Registry SHALL return an error indicating the conflict
5. WHEN parsing arguments containing short flags, THE Flag_Registry SHALL delegate to OptArgs_Core for POSIX-compliant parsing

### Requirement 3: Slice Type Support

**User Story:** As a developer, I want support for slice flags, so that users can provide multiple values for a single flag.

#### Acceptance Criteria

1. WHEN a developer calls `StringSliceVar()`, THE Flag_Registry SHALL create a flag that accepts multiple string values
2. WHEN a user provides `--flag=val1,val2,val3`, THE Flag_Registry SHALL parse it into a slice containing ["val1", "val2", "val3"]
3. WHEN a user provides `--flag=val1 --flag=val2`, THE Flag_Registry SHALL append both values to the slice
4. WHEN a developer calls `IntSliceVar()`, THE Flag_Registry SHALL create a flag that accepts multiple integer values with comma separation
5. IF a slice flag receives an invalid value for its type, THEN THE Flag_Registry SHALL return a descriptive error message

### Requirement 4: Boolean Flag Behavior

**User Story:** As a developer, I want proper boolean flag handling, so that boolean flags work intuitively without requiring explicit values.

#### Acceptance Criteria

1. WHEN a user provides `--verbose` for a boolean flag, THE Flag_Registry SHALL set the flag value to true
2. WHEN a user provides `--verbose=false`, THE Flag_Registry SHALL set the flag value to false
3. WHEN a user provides `--verbose=true`, THE Flag_Registry SHALL set the flag value to true
4. IF a user provides `--verbose=invalid`, THEN THE Flag_Registry SHALL return an error message "invalid boolean value 'invalid'"
5. WHEN a boolean flag has a default value of true, THE Flag_Registry SHALL support `--no-verbose` syntax to set it to false

### Requirement 5: FlagSet Management

**User Story:** As a developer, I want to create and manage separate flag sets, so that I can organize flags for different commands or contexts.

#### Acceptance Criteria

1. WHEN a developer calls `NewFlagSet()` with a name, THE Flag_Registry SHALL create an isolated flag namespace
2. WHEN flags are defined on a FlagSet, THE Flag_Registry SHALL ensure they don't conflict with other FlagSets
3. WHEN `Parse()` is called on a FlagSet, THE Flag_Registry SHALL parse only the flags defined in that set
4. WHEN `Parsed()` is called on a FlagSet, THE Flag_Registry SHALL return true only after Parse() has been called successfully
5. WHEN accessing flag values before parsing, THE Flag_Registry SHALL return the default values without error

### Requirement 6: Custom Value Interface

**User Story:** As a developer, I want to implement custom flag types, so that I can handle specialized data formats and validation.

#### Acceptance Criteria

1. WHEN a developer implements the Value interface with String() and Set() methods, THE Flag_Registry SHALL accept it as a custom flag type
2. WHEN a custom Value's Set() method returns an error, THE Flag_Registry SHALL propagate that error during parsing
3. WHEN a custom Value's String() method is called, THE Flag_Registry SHALL use the result for help text and default display
4. WHEN a developer calls `Var()` with a Value interface, THE Flag_Registry SHALL register the flag with the custom type
5. WHEN parsing arguments for custom flags, THE Flag_Registry SHALL call the Value's Set() method with the provided argument

### Requirement 7: Help Text Generation

**User Story:** As a developer, I want automatic help text generation, so that users can discover available flags and their usage.

#### Acceptance Criteria

1. WHEN `Usage()` is called on a FlagSet, THE Flag_Registry SHALL output formatted help text to stderr
2. WHEN generating help text, THE Flag_Registry SHALL include flag names, shorthand, default values, and usage descriptions
3. WHEN a flag has no shorthand, THE Flag_Registry SHALL format it as `--flag` in help text
4. WHEN a flag has shorthand, THE Flag_Registry SHALL format it as `-f, --flag` in help text
5. WHEN `FlagUsages()` is called, THE Flag_Registry SHALL return a formatted string containing all flag documentation

### Requirement 8: Flag Lookup and Introspection

**User Story:** As a developer, I want to inspect and access flag definitions programmatically, so that I can build dynamic CLI behaviors.

#### Acceptance Criteria

1. WHEN a developer calls `Lookup()` with a flag name, THE Flag_Registry SHALL return the Flag object if it exists
2. IF a flag name doesn't exist during lookup, THEN THE Flag_Registry SHALL return nil
3. WHEN a developer calls `VisitAll()` with a function, THE Flag_Registry SHALL call that function for each defined flag
4. WHEN a developer calls `Visit()` with a function, THE Flag_Registry SHALL call that function only for flags that were set during parsing
5. WHEN accessing a Flag object, THE Flag_Registry SHALL provide access to Name, Shorthand, Usage, DefValue, and Value fields

### Requirement 9: Error Handling and Validation

**User Story:** As a developer, I want clear error messages for invalid flag usage, so that users can understand and correct their mistakes.

#### Acceptance Criteria

1. IF a user provides an unknown flag during parsing, THEN THE Flag_Registry SHALL return an error message "unknown flag: --flagname"
2. IF a user provides a flag that requires an argument without one, THEN THE Flag_Registry SHALL return an error message "flag needs an argument: --flagname"
3. IF a user provides an invalid value for a typed flag, THEN THE Flag_Registry SHALL return an error message indicating the expected type
4. WHEN type conversion fails for integer flags, THE Flag_Registry SHALL return an error message "invalid syntax for integer flag --flagname"
5. WHEN type conversion fails for duration flags, THE Flag_Registry SHALL return an error message "invalid duration format for flag --flagname"

### Requirement 10: OptArgs Core Integration

**User Story:** As a library user, I want the reliability of OptArgs core, so that I get correct POSIX/GNU behavior through the pflag-compatible interface.

#### Acceptance Criteria

1. WHEN parsing arguments, THE PFlags_Wrapper SHALL delegate all parsing operations to OptArgs_Core
2. WHEN OptArgs_Core detects parsing errors, THE PFlags_Wrapper SHALL translate them to pflag-compatible error messages
3. WHEN OptArgs_Core supports POSIXLY_CORRECT mode, THE PFlags_Wrapper SHALL expose this through a SetPosixCompliance() method
4. WHEN OptArgs_Core handles option termination with `--`, THE PFlags_Wrapper SHALL preserve this behavior in Parse() methods
5. WHEN OptArgs_Core processes combined short options like `-abc`, THE PFlags_Wrapper SHALL ensure correct flag resolution
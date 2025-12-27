# Requirements Document

## Introduction

The OptArgs core provides the foundational POSIX/GNU getopt implementation that serves as the basis for all higher-level wrapper interfaces. This core must be complete, correct, and thoroughly tested to ensure reliability for all dependent components.

## Glossary

- **OptArgs_Core**: The foundational POSIX/GNU getopt implementation
- **Parser_Engine**: The core parsing logic that processes command-line arguments
- **Option_Registry**: Internal storage for registered short and long options
- **Argument_Stream**: The sequence of command-line arguments being processed

## Requirements

### Requirement 1: POSIX getopt(3) Compliance

**User Story:** As a library maintainer, I want complete POSIX getopt(3) compliance, so that the core can handle all standard POSIX argument patterns.

#### Acceptance Criteria

1. THE Parser_Engine SHALL implement all POSIX getopt(3) functionality exactly as specified
2. WHEN processing short options, THE Parser_Engine SHALL support option compaction (e.g., `-abc` = `-a -b -c`)
3. WHEN processing arguments with colons, THE Parser_Engine SHALL handle required arguments (`:`) and optional arguments (`::`)
4. THE Parser_Engine SHALL respect the POSIXLY_CORRECT environment variable behavior
5. WHEN encountering `--`, THE Parser_Engine SHALL stop processing options and treat remaining arguments as non-options

### Requirement 2: GNU getopt_long(3) Extensions

**User Story:** As a library maintainer, I want full GNU getopt_long(3) support, so that the core can handle modern long-option patterns.

#### Acceptance Criteria

1. THE Parser_Engine SHALL implement all GNU getopt_long(3) functionality exactly as specified
2. WHEN processing long options, THE Parser_Engine SHALL support `--option=value` syntax
3. WHEN processing long options, THE Parser_Engine SHALL support `--option value` syntax
4. THE Parser_Engine SHALL support case-insensitive long options by default
5. THE Parser_Engine SHALL handle partial long option matching when unambiguous

### Requirement 3: GNU getopt_long_only(3) Support

**User Story:** As a library maintainer, I want getopt_long_only(3) support, so that single-dash long options work correctly.

#### Acceptance Criteria

1. THE Parser_Engine SHALL implement GNU getopt_long_only(3) functionality
2. WHEN in long-only mode, THE Parser_Engine SHALL treat `-option` as a long option
3. WHEN in long-only mode, THE Parser_Engine SHALL fall back to short option parsing for single characters
4. THE Parser_Engine SHALL maintain compatibility with existing short option behavior
5. THE Parser_Engine SHALL handle ambiguous cases according to GNU specifications

### Requirement 4: Advanced Option Handling

**User Story:** As a library user, I want support for complex option patterns, so that I can implement sophisticated CLI interfaces.

#### Acceptance Criteria

1. THE Parser_Engine SHALL support the GNU `-W` extension for word-based options
2. WHEN processing compacted options with arguments, THE Parser_Engine SHALL assign arguments to the last option that accepts them
3. THE Parser_Engine SHALL allow arguments that begin with `-` when explicitly required
4. THE Parser_Engine SHALL support all printable ASCII characters as valid short options (except `:`, `;`, `-`)
5. THE Parser_Engine SHALL handle option definitions that can be redefined or overridden

### Requirement 5: Error Handling and Reporting

**User Story:** As a library user, I want clear error handling, so that I can provide meaningful feedback to end users.

#### Acceptance Criteria

1. WHEN invalid options are encountered, THE Parser_Engine SHALL provide descriptive error messages
2. WHEN required arguments are missing, THE Parser_Engine SHALL report the specific option that needs an argument
3. THE Parser_Engine SHALL support silent error mode (controlled by `:` prefix in optstring)
4. WHEN ambiguous long options are provided, THE Parser_Engine SHALL list possible matches
5. THE Parser_Engine SHALL distinguish between unknown options and malformed option syntax

### Requirement 6: API Stability and Extensibility

**User Story:** As a wrapper implementer, I want a stable core API, so that I can build reliable higher-level interfaces.

#### Acceptance Criteria

1. THE OptArgs_Core SHALL maintain backward compatibility with existing public APIs
2. THE OptArgs_Core SHALL expose all necessary primitives for wrapper implementations
3. WHEN new functionality is added, THE OptArgs_Core SHALL not break existing wrapper interfaces
4. THE OptArgs_Core SHALL provide iterator-based option processing for memory efficiency
5. THE OptArgs_Core SHALL support configuration without requiring wrapper modifications
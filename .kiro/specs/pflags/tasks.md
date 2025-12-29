# Implementation Plan: PFlags Wrapper

## Overview

This implementation plan creates a drop-in replacement for spf13/pflag that maintains complete API compatibility while leveraging OptArgs Core for POSIX/GNU compliance. The approach focuses on building a compatibility layer that translates pflag API calls into OptArgs Core operations while preserving all expected behaviors and error messages.

## Tasks

- [x] 1. Set up project structure and core interfaces
  - Create pflags package directory as independent Go module
  - Set up go.mod file with goarg dependency and local development replacement
  - Define core interfaces (FlagSet, Flag, Value) matching spf13/pflag signatures
  - Set up testing framework with both unit and property-based testing support
  - _Requirements: 1.1, 5.1, 6.1, 11.1, 11.2_

- [x] 2. Implement basic flag value types
  - [x] 2.1 Implement string, int, bool, float64, and duration value types
    - Create value type structs implementing the Value interface
    - Implement Set(), String(), and Type() methods for each type
    - Add proper type validation and error messages
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 2.2 Write property test for flag creation consistency
    - **Property 1: Flag Creation Consistency**
    - **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5**

  - [x] 2.3 Write unit tests for value type implementations
    - Test Set() method with valid and invalid inputs
    - Test String() method output format
    - Test Type() method return values
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 3. Implement FlagSet core functionality
  - [x] 3.1 Create FlagSet struct and basic methods
    - Implement NewFlagSet() constructor
    - Add flag registration methods (StringVar, IntVar, BoolVar, etc.)
    - Implement flag storage and lookup mechanisms
    - _Requirements: 5.1, 1.1, 1.2, 1.3_

  - [x] 3.2 Write property test for FlagSet isolation
    - **Property 5: FlagSet Isolation**
    - **Validates: Requirements 5.1, 5.2, 5.3**

  - [x] 3.3 Implement shorthand support
    - Add shorthand registration and conflict detection
    - Implement StringP, IntP, BoolP methods with shorthand support
    - Create shorthand-to-name mapping system
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

  - [x] 3.4 Write property test for shorthand registration and resolution
    - **Property 2: Shorthand Registration and Resolution**
    - **Validates: Requirements 2.1, 2.2, 2.3**

- [x] 4. Implement slice type support
  - [x] 4.1 Create slice value types (stringSlice, intSlice, etc.)
    - Implement slice value structs with comma-separated and repeated flag support
    - Add proper parsing for both `--flag=a,b,c` and `--flag=a --flag=b` syntax
    - Implement type validation for slice elements
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [x] 4.2 Write property test for slice flag value accumulation
    - **Property 3: Slice Flag Value Accumulation**
    - **Validates: Requirements 3.1, 3.2, 3.3, 3.4**

  - [x] 4.3 Write property test for slice type validation
    - **Property 13: Slice Type Validation**
    - **Validates: Requirements 3.5**

- [x] 5. Implement OptArgs Core integration layer
  - [x] 5.1 Create CoreIntegration component
    - Build translation layer between pflag Flag definitions and OptArgs Core format
    - Implement flag registration with OptArgs Core parser
    - Add argument type mapping (NoArgument, RequiredArgument)
    - _Requirements: 10.1, 10.2_

  - [x] 5.2 Implement parsing delegation and result processing
    - Create Parse() method that delegates to OptArgs Core
    - Process OptArgs Core options and update flag values
    - Implement error translation from OptArgs Core to pflag format
    - _Requirements: 10.1, 10.2, 9.1, 9.2_

  - [x] 5.3 Write property test for OptArgs Core integration fidelity
    - **Property 11: OptArgs Core Integration Fidelity**
    - **Validates: Requirements 10.1, 10.2**

- [x] 6. Implement boolean flag special handling
  - [x] 6.1 Add enhanced boolean flag parsing
    - Support no-argument boolean flags (--verbose sets to true)
    - Handle explicit boolean values (--verbose=true/false)
    - Implement boolean negation syntax (--no-verbose)
    - Add proper error messages for invalid boolean values
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [x] 6.2 Write property test for boolean flag parsing flexibility
    - **Property 4: Boolean Flag Parsing Flexibility**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.5**

- [x] 7. Checkpoint - Core functionality complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 8. Implement custom Value interface support
  - [ ] 8.1 Add Var() method and custom value handling
    - Implement Var() method for custom Value interface types
    - Ensure Set() method is called during parsing with correct arguments
    - Use String() method for help text and default value display
    - Propagate errors from custom Value Set() methods
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

  - [ ] 8.2 Write property test for custom Value interface integration
    - **Property 7: Custom Value Interface Integration**
    - **Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.5**

- [ ] 9. Implement help text generation
  - [ ] 9.1 Create Usage() and help text formatting
    - Implement Usage() method that outputs to stderr
    - Format help text with flag names, shorthand, defaults, and descriptions
    - Handle flags with and without shorthand appropriately
    - Implement FlagUsages() for programmatic access
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

  - [ ] 9.2 Write property test for help text completeness
    - **Property 8: Help Text Completeness**
    - **Validates: Requirements 7.2, 7.3, 7.4, 7.5**

- [ ] 10. Implement flag introspection methods
  - [ ] 10.1 Add Lookup(), VisitAll(), and Visit() methods
    - Implement Lookup() with proper nil handling for non-existent flags
    - Create VisitAll() to iterate over all defined flags
    - Implement Visit() to iterate only over changed flags
    - Ensure Flag objects provide access to all required fields
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

  - [ ] 10.2 Write property test for flag introspection accuracy
    - **Property 9: Flag Introspection Accuracy**
    - **Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5**

- [ ] 11. Implement parse state management
  - [ ] 11.1 Add Parsed() method and state tracking
    - Implement Parsed() method that tracks parsing completion
    - Ensure flag values return defaults before parsing
    - Handle parse state correctly across multiple Parse() calls
    - _Requirements: 5.4, 5.5_

  - [ ] 11.2 Write property test for parse state consistency
    - **Property 6: Parse State Consistency**
    - **Validates: Requirements 5.4, 5.5**

- [ ] 12. Implement comprehensive error handling
  - [ ] 12.1 Add error message formatting and validation
    - Implement specific error messages for unknown flags
    - Add error messages for missing required arguments
    - Create type validation error messages
    - Ensure error message format matches pflag expectations
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

  - [ ] 12.2 Write property test for type validation error messages
    - **Property 10: Type Validation Error Messages**
    - **Validates: Requirements 9.3**

- [ ] 13. Implement POSIX compliance features
  - [ ] 13.1 Add POSIX-specific functionality
    - Implement SetPosixCompliance() method for POSIXLY_CORRECT mode
    - Ensure option termination with `--` works correctly
    - Handle combined short options like `-abc` properly
    - Maintain compatibility with OptArgs Core POSIX features
    - _Requirements: 10.3, 10.4, 10.5_

  - [ ] 13.2 Write property test for POSIX compliance preservation
    - **Property 12: POSIX Compliance Preservation**
    - **Validates: Requirements 10.3, 10.4, 10.5**

- [ ] 14. Implement comprehensive testing infrastructure
  - [ ] 14.1 Create Makefile with all testing targets
    - Implement all testing targets matching optargs (test, coverage, coverage-html, coverage-func, coverage-validate, coverage-report)
    - Add static analysis targets (lint, static-check, security-check, fmt, imports, vet, mod-tidy, mod-verify, build-check)
    - Add CI integration targets (ci-coverage, ci-static, pre-commit)
    - Add development targets (dev-coverage, clean, help)
    - _Requirements: All requirements_

  - [ ] 14.2 Create coverage validation scripts
    - Create scripts/validate_coverage.sh for 100% coverage validation of core functions
    - Create scripts/generate_coverage_report.sh for comprehensive coverage analysis
    - Create scripts/performance_validation.sh for performance regression testing
    - **CRITICAL**: All scripts must accept optional target directory parameter (defaults to current directory)
    - Scripts must work from any location: `./scripts/validate_coverage.sh pflags/` or `cd pflags && ../scripts/validate_coverage.sh`
    - Ensure scripts validate pflags-specific core functions (FlagSet methods, Value implementations, parsing logic)
    - _Requirements: All requirements_

  - [ ] 14.3 Set up pre-commit workflow
    - **IMPORTANT**: Enhance existing .github/workflows/precommit.yml to handle all modules (optargs, goarg, pflags)
    - Configure pre-commit hooks to work across all module directories
    - Set up automated PR comments for pre-commit results across all modules
    - **DO NOT** create duplicate workflow files - extend existing workflows
    - _Requirements: All requirements_

  - [ ] 14.4 Configure static analysis tools
    - Create .golangci.yml configuration file
    - Set up .pre-commit-config.yaml with all hooks
    - Configure coverage targets for pflags-specific functions
    - _Requirements: All requirements_

  - [ ]* 14.5 Write comprehensive test suites
    - Unit tests for all pflags functions with 100% coverage target
    - Property-based tests for flag parsing correctness across all input ranges
    - Performance benchmarks and regression tests
    - Round-trip testing for flag definition and parsing
    - _Requirements: All requirements_

- [ ] 15. Integration and compatibility testing
  - [ ] 15.1 Create comprehensive integration tests
    - Test compatibility with existing spf13/pflag code patterns
    - Verify API signature compatibility
    - Test integration with Cobra-style usage patterns
    - Test goarg module integration and functionality
    - _Requirements: All requirements_

  - [ ] 15.2 Write performance comparison tests
    - Compare performance against original spf13/pflag
    - Ensure memory usage is comparable or better
    - Validate parsing speed meets or exceeds pflag performance

- [ ] 16. Module dependency validation and build system integration
  - [ ] 16.1 Validate module dependency configurations
    - Test pflags module with local goarg dependency replacement
    - Test pflags module with remote goarg dependency (git URL)
    - Validate go.mod file correctness and dependency resolution
    - Test standard Go module operations (go get, go mod tidy, go mod verify)
    - Test transitive optargs dependency through goarg
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

  - [ ] 16.2 Implement cascading build system
    - **IMPORTANT**: Enhance existing .github/workflows/build.yml and .github/workflows/coverage.yml
    - Extend existing workflows to handle pflags module builds and testing
    - Implement intelligent change detection for pflags sources and full dependency chain
    - **DO NOT** create duplicate workflow files - extend existing multi-module workflows
    - Test build system with various change scenarios (optargs only, goarg only, pflags only, combinations)
    - Validate build optimization and resource usage across full dependency chain
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

  - [ ]* 16.3 Write integration tests for module dependencies
    - Test module import and usage from external projects
    - Test version compatibility and semantic versioning across dependency chain
    - Test dependency resolution in various Go environments
    - Test goarg integration functionality
    - _Requirements: 11.5, 12.5_

- [ ] 17. Final checkpoint - Implementation complete
  - Ensure all tests pass, ask the user if questions arise.
  - Confirm all requirements are implemented and tested
  - Validate API compatibility with spf13/pflag
  - Validate module dependencies and cascading build system

## Notes

- Each task references specific requirements for traceability
- Property tests validate universal correctness properties using Go's testing/quick framework
- Unit tests validate specific examples and edge cases
- Integration tests ensure compatibility with existing pflag usage patterns
- The implementation maintains strict API compatibility while leveraging OptArgs Core's superior parsing through goarg
- **Testing Infrastructure Requirements**:
  - All modules must have identical testing infrastructure to optargs
  - 100% line and branch coverage for core functions (FlagSet methods, Value implementations, parsing logic)
  - Comprehensive static analysis (fmt, imports, vet, lint, security-check)
  - Performance benchmarks and regression testing
  - Pre-commit hooks and automated quality checks
- **Centralized Workflow Management**:
  - **DO NOT** create duplicate GitHub workflow files
  - Enhance existing .github/workflows/ files to handle all modules
  - Single source of truth for build, test, and coverage workflows
  - Intelligent change detection across full dependency chain
- **Script Flexibility Requirements**:
  - All validation scripts must accept optional target directory parameter
  - Scripts must work from any location: `./scripts/validate_coverage.sh pflags/`
  - Default to current working directory if no path specified
- **Module Dependencies**:
  - pflags is an independent Go module depending on goarg
  - goarg provides enhanced functionality and optargs integration
  - Local development uses file system replacement for rapid iteration
  - CI/CD builds use local replacement for integration testing
  - Production releases depend on published goarg module via git URL
- **Cascading Builds**:
  - Changes in optargs or goarg trigger pflags builds and tests
  - Changes only in pflags trigger pflags builds without rebuilding dependencies
  - No changes in any module skip pflags builds for optimization
  - Full dependency chain: pflags → goarg → optargs

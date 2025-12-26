---
inclusion: always
---

# Testing Standards for OptArgs Project

## Code Coverage Requirements

All code in the OptArgs project must achieve:
- **100% line coverage** for core parsing functionality
- **100% branch coverage** for all conditional logic
- **Property-based testing** for all parsing functions to validate correctness across input ranges

## Test Organization

### Core Tests
- Unit tests for all parsing functions
- Integration tests for complete parsing workflows  
- Property-based tests using Go's testing/quick or similar framework
- POSIX compliance tests validating against posix/ directory examples

### Wrapper Tests
- Unit tests for wrapper-specific functionality
- Integration tests showing wrapper-to-core interaction
- Compatibility tests ensuring API compatibility with target libraries

## Test Naming and Structure

- Test files should use `_test.go` suffix
- Property-based tests should be clearly marked with `Property` prefix
- Each test should validate specific requirements from the spec documents
- Use table-driven tests for multiple input scenarios

## Continuous Integration

- All tests must pass before merging
- Coverage reports must be generated and maintained
- Property-based tests should run with sufficient iterations (minimum 100)

## Round-Trip Testing

For all parsing operations, implement round-trip tests:
- Parse arguments → Generate equivalent arguments → Parse again → Verify equivalence
- This is especially critical for option compaction and expansion logic
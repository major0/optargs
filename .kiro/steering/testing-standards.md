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

### Build Workflow Requirements

All code changes must pass the automated build workflow which includes:

#### Multi-Version Go Support
- **MUST** build and test successfully on Go 1.21, 1.22, and 1.23
- **MUST** maintain compatibility with the minimum supported Go version
- **MUST** use the latest stable Go version for primary development

#### Code Quality Checks
- Code quality checks are handled by separate workflows
- Build workflow focuses solely on compilation verification

#### Build Verification
- **MUST** build successfully with `go build -v ./...`
- **MUST** verify all dependencies with `go mod verify`

#### Cross-Platform Compatibility
- **MUST** build successfully for Linux (amd64, arm64)
- **MUST** build successfully for macOS (amd64, arm64)
- **MUST** build successfully for Windows (amd64)
- **MUST** build successfully for FreeBSD (amd64)

#### Workflow Integration
- Build workflow **MUST** run on all pushes to main and develop branches
- Build workflow **MUST** run on all pull requests
- All build jobs **MUST** pass before code can be merged
- Build artifacts **MUST** be cached for performance optimization

## Round-Trip Testing

For all parsing operations, implement round-trip tests:
- Parse arguments → Generate equivalent arguments → Parse again → Verify equivalence
- This is especially critical for option compaction and expansion logic

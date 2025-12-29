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

The project uses a comprehensive CI/CD pipeline with three main workflows:

### 1. Pre-commit Workflow (`.github/workflows/precommit.yml`)

**Triggers**: Pull requests to main/develop branches
**Purpose**: Code quality validation and formatting checks

**Pre-commit Hooks Enforced**:
- **trailing-whitespace**: Remove trailing whitespace
- **end-of-file-fixer**: Ensure files end with newline
- **check-yaml**: Validate YAML syntax
- **check-added-large-files**: Prevent large file commits
- **go-fmt**: Format Go code
- **go-imports**: Organize Go imports
- **no-go-testing**: Prevent testing package imports in non-test files
- **golangci-lint**: Go code quality checks (currently disabled due to Go 1.23 compatibility)
- **commitlint**: Validate commit message format (Conventional Commits)
- **detect-secrets**: Security vulnerability scanning
- **go-mod-tidy**: Ensure go.mod/go.sum are tidy
- **go-mod-verify**: Verify go.mod dependencies
- **go-vet**: Static analysis with go vet
- **go-build**: Verify code builds successfully
- **go-test-short**: Run short tests with race detection

**Requirements**:
- **MUST** pass all pre-commit hooks before merge
- **MUST** use Go 1.23 for pre-commit validation
- **MUST** follow Conventional Commits format
- **MUST** have clean go.mod/go.sum files

### 2. Build Workflow (`.github/workflows/build.yml`)

**Triggers**: Pushes to main/develop branches and pull requests
**Purpose**: Multi-platform and multi-version build verification

#### Multi-Version Go Support
- **MUST** build successfully on Go 1.21, 1.22, and 1.23
- **MUST** maintain compatibility with minimum supported Go version
- **MUST** use Go 1.23 as primary development version

#### Cross-Platform Compatibility
- **MUST** build successfully for:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64)
  - FreeBSD (amd64) - excludes arm64

#### Build Verification
- **MUST** build successfully with `go build -v ./...`
- **MUST** verify all dependencies with `go mod verify`
- **MUST** download and cache dependencies properly

### 3. Coverage Workflow (`.github/workflows/coverage.yml`)

**Triggers**: Pushes to main/develop branches and pull requests
**Purpose**: Test coverage validation and regression detection

#### Coverage Requirements
- **MUST** generate coverage reports using `make coverage`
- **MUST** validate coverage targets using `make coverage-validate`
- **MUST** upload coverage to Codecov for tracking
- **MUST** detect coverage regressions (>1% decrease threshold)

#### Coverage Validation
- **Core parsing functions**: 100% line and branch coverage required
- **Overall project**: 90% minimum coverage threshold
- **Coverage regression**: Maximum 1% decrease allowed

#### Coverage Artifacts
- `coverage.out`: Coverage profile for analysis
- `coverage.html`: HTML coverage report
- `coverage_analysis.md`: Comprehensive coverage analysis
- `coverage_gaps_detailed.md`: Detailed gap identification

#### Coverage Scripts
- `scripts/validate_coverage.sh`: Validates coverage targets
- `scripts/generate_coverage_report.sh`: Generates comprehensive analysis

### Workflow Integration Requirements

- **MUST** run on all pushes to main and develop branches
- **MUST** run on all pull requests
- **MUST** pass all workflow jobs before code can be merged
- **MUST** cache build artifacts and dependencies for performance
- **MUST** provide detailed feedback via PR comments

## Round-Trip Testing

For all parsing operations, implement round-trip tests:
- Parse arguments → Generate equivalent arguments → Parse again → Verify equivalence
- This is especially critical for option compaction and expansion logic

## Makefile Integration

The project includes a comprehensive Makefile with standardized targets for testing and coverage:

### Core Testing Targets
- `make test`: Run all tests
- `make coverage`: Generate coverage profile with atomic mode
- `make coverage-html`: Generate HTML coverage report
- `make coverage-func`: Display function-level coverage summary
- `make coverage-validate`: Validate coverage meets 100% target for core functions
- `make coverage-report`: Generate comprehensive coverage analysis and gap reports

### Static Analysis Targets
- `make lint`: Run golangci-lint static analysis
- `make static-check`: Comprehensive static analysis suite (fmt, imports, vet, mod-tidy, mod-verify, lint, build-check)
- `make security-check`: Security vulnerability checks with govulncheck
- `make fmt`: Format Go code
- `make imports`: Fix Go imports with goimports
- `make vet`: Run go vet
- `make mod-tidy`: Tidy go.mod and go.sum
- `make mod-verify`: Verify go.mod dependencies
- `make build-check`: Verify code builds successfully

### CI Integration Targets
- `make ci-coverage`: Automated coverage validation for CI
- `make ci-static`: Static analysis for CI
- `make pre-commit`: Complete pre-commit validation suite

### Development Targets
- `make dev-coverage`: Quick coverage check for development
- `make clean`: Remove coverage files
- `make help`: Display available targets and usage

**Requirements**:
- **MUST** use Makefile targets in CI workflows for consistency
- **MUST** maintain 100% coverage target for core parsing functions
- **MUST** run static analysis before committing code
- **MUST** validate coverage targets in CI pipeline

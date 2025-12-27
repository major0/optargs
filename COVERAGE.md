# OptArgs Core - Test Coverage Tracking

This document describes the comprehensive test coverage tracking system implemented for the OptArgs core library.

## Overview

The coverage tracking system ensures that all core parsing functionality achieves 100% test coverage, providing confidence in the correctness and reliability of the OptArgs implementation.

## Coverage Targets

### Core Functions (100% Required)
- **GetOpt**: POSIX getopt(3) implementation
- **GetOptLong**: GNU getopt_long(3) implementation
- **GetOptLongOnly**: GNU getopt_long_only(3) implementation
- **getOpt**: Internal parsing orchestration
- **findLongOpt**: Long option matching and resolution
- **findShortOpt**: Short option processing and compaction
- **Options**: Iterator-based option processing
- **optError/optErrorf**: Error reporting functions

### Overall Project Targets
- **Minimum**: 90% overall coverage
- **Target**: 95% overall coverage
- **Ideal**: 100% overall coverage

## Quick Start

### Generate Coverage Report
```bash
# Generate coverage profile and HTML report
make coverage-html

# View function-level coverage summary
make coverage-func

# Validate coverage meets targets
make coverage-validate
```

### Comprehensive Analysis
```bash
# Generate detailed coverage analysis and gap reports
make coverage-report

# Clean coverage files
make clean
```

## Available Commands

### Make Targets

| Command | Description |
|---------|-------------|
| `make test` | Run all tests |
| `make coverage` | Generate coverage profile |
| `make coverage-html` | Generate HTML coverage report |
| `make coverage-func` | Display function-level coverage |
| `make coverage-validate` | Validate coverage meets targets |
| `make coverage-report` | Generate comprehensive analysis |
| `make clean` | Remove coverage files |
| `make ci-coverage` | CI validation target |
| `make dev-coverage` | Quick development check |

### Direct Commands

```bash
# Generate coverage with atomic mode for accurate branch coverage
go test -coverprofile=coverage.out -covermode=atomic ./...

# View HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Display function-level coverage
go tool cover -func=coverage.out

# Validate coverage targets
./scripts/validate_coverage.sh coverage.out

# Generate comprehensive analysis
./scripts/generate_coverage_report.sh coverage.out
```

## Coverage Reports

### Generated Files

| File | Description |
|------|-------------|
| `coverage.out` | Coverage profile data |
| `coverage.html` | Interactive HTML coverage report |
| `coverage_analysis.md` | Comprehensive coverage analysis |
| `coverage_gaps_detailed.md` | Detailed gap identification and recommendations |

### Report Contents

#### coverage_analysis.md
- Overall coverage summary
- Per-file and per-function breakdown
- Coverage gap categories
- Testing recommendations
- Next steps and action items

#### coverage_gaps_detailed.md
- Specific uncovered code paths
- Missing test scenarios
- Priority assessment
- Implementation recommendations
- Testing strategy guidance

## CI Integration

### GitHub Actions Workflow

The coverage system is integrated into CI through `.github/workflows/coverage.yml`:

- **Coverage Generation**: Automatic coverage analysis on every push/PR
- **Validation**: Ensures coverage targets are met
- **Regression Detection**: Prevents coverage decreases
- **Reporting**: Uploads reports and comments on PRs
- **Badge Updates**: Maintains coverage badge on main branch

### Coverage Validation

The CI pipeline validates:
- Core functions achieve 100% coverage
- Overall coverage meets minimum threshold (90%)
- No coverage regression >1% from main branch
- All tests pass before coverage analysis

### Artifacts

CI generates and uploads:
- Coverage profiles (`coverage.out`)
- HTML reports (`coverage.html`)
- Analysis reports (`coverage_analysis.md`, `coverage_gaps_detailed.md`)
- Codecov integration for external tracking

## Coverage Analysis Process

### 1. Baseline Establishment
```bash
# Generate current coverage baseline
make coverage-report
```

### 2. Gap Identification
- Review `coverage_gaps_detailed.md` for specific missing scenarios
- Identify functions with <100% coverage
- Prioritize gaps by impact and complexity

### 3. Test Implementation
- Add tests for identified gaps
- Focus on error paths and edge cases
- Implement property-based tests for complex logic

### 4. Validation
```bash
# Validate coverage improvements
make coverage-validate

# Generate updated analysis
make coverage-report
```

### 5. Monitoring
- Regular coverage analysis and reporting
- Continuous monitoring for regressions
- Periodic review of coverage targets and thresholds

## Testing Strategy

### Unit Testing
- **Specific examples**: Test concrete scenarios and edge cases
- **Error conditions**: Cover all error paths and invalid inputs
- **Boundary testing**: Test limits and edge conditions
- **Integration points**: Test component interactions

### Property-Based Testing
- **Universal properties**: Test properties that should hold for all inputs
- **Round-trip testing**: Parse → generate → parse → verify equivalence
- **Invariant testing**: Ensure parsing rules are consistently applied
- **Metamorphic testing**: Test relationships between different input forms

### Coverage-Driven Development
1. **Write tests first**: Implement comprehensive test coverage
2. **Identify gaps**: Use coverage analysis to find missing scenarios
3. **Add targeted tests**: Focus on specific uncovered code paths
4. **Validate completeness**: Ensure 100% coverage for core functions

## Troubleshooting

### Common Issues

#### Coverage File Not Found
```bash
# Error: coverage.out not found
# Solution: Generate coverage first
make coverage
```

#### Low Coverage Warnings
```bash
# Check which functions need more tests
make coverage-func

# See detailed gap analysis
make coverage-report
```

#### CI Coverage Failures
1. Check coverage validation output
2. Review generated gap reports
3. Add tests for identified missing scenarios
4. Re-run validation

### Debugging Coverage Issues

#### View Uncovered Lines
```bash
# Generate HTML report and open in browser
make coverage-html
# Open coverage.html to see line-by-line coverage
```

#### Identify Specific Gaps
```bash
# Generate detailed analysis
make coverage-report

# Review coverage_gaps_detailed.md for specific missing scenarios
```

#### Validate Specific Functions
```bash
# Check coverage for specific function
go tool cover -func=coverage.out | grep "FunctionName"
```

## Configuration

### Coverage Targets
Defined in `.coveragerc`:
- Core function targets (100%)
- Overall project targets (90-100%)
- Validation thresholds
- CI integration settings

### Validation Scripts
- `scripts/validate_coverage.sh`: Coverage target validation
- `scripts/generate_coverage_report.sh`: Comprehensive analysis generation

### CI Configuration
- `.github/workflows/coverage.yml`: GitHub Actions workflow
- Automated validation, reporting, and badge updates

## Good Practices

### Writing Tests for Coverage
1. **Focus on functionality**: Test behavior, not just coverage
2. **Cover error paths**: Ensure all error conditions are tested
3. **Test edge cases**: Include boundary conditions and unusual inputs
4. **Use property-based testing**: Validate universal properties
5. **Maintain readability**: Write clear, maintainable test code

### Maintaining Coverage
1. **Regular monitoring**: Check coverage on every change
2. **Prevent regressions**: Use CI to catch coverage decreases
3. **Update targets**: Adjust targets as code evolves
4. **Review gaps**: Regularly analyze and address coverage gaps

### CI Integration
1. **Fail fast**: Stop builds on coverage regressions
2. **Provide feedback**: Comment on PRs with coverage information
3. **Track trends**: Monitor coverage changes over time
4. **Automate reporting**: Generate and upload coverage reports

## Support

### Getting Help
- Review this documentation for common issues
- Check generated coverage reports for specific guidance
- Use `make help` for available commands
- Examine CI logs for detailed error information

### Contributing
- Ensure new code includes comprehensive tests
- Maintain or improve overall coverage percentage
- Add property-based tests for complex functionality
- Update documentation when adding new coverage features
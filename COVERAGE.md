# Test Coverage

## Targets

Core parsing functions require 100% line coverage:

- `GetOpt`, `GetOptLong`, `GetOptLongOnly`
- `getOpt`, `findLongOpt`, `findShortOpt`
- `Options`
- `optError`, `optErrorf`

Overall project minimum: 90%.

## Commands

```bash
# Run tests with coverage
make coverage

# HTML report (opens coverage.html)
make coverage-html

# Function-level summary
make coverage-func

# Validate targets
make coverage-validate

# Full analysis (generates coverage_analysis.md, coverage_gaps_detailed.md)
make coverage-report
```

## CI

Coverage validation runs on every push and PR via `.github/workflows/coverage.yml`. The pipeline generates coverage profiles, validates targets, and detects regressions.

## Configuration

Coverage targets are defined in `.coveragerc`. Validation logic lives in `scripts/validate_coverage.sh` and `scripts/generate_coverage_report.sh`.

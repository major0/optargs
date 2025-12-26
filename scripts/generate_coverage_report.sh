#!/bin/bash
# Comprehensive coverage analysis and gap reporting script
# Generates detailed coverage analysis and identifies specific gaps

set -e

COVERAGE_FILE="${1:-coverage.out}"
ANALYSIS_FILE="coverage_analysis.md"
GAPS_FILE="coverage_gaps_detailed.md"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if coverage file exists
if [[ ! -f "$COVERAGE_FILE" ]]; then
    echo "Error: Coverage file $COVERAGE_FILE not found"
    echo "Run 'make coverage' first to generate coverage data"
    exit 1
fi

echo "Generating comprehensive coverage analysis..."

# Generate function-level coverage data
FUNC_COVERAGE=$(go tool cover -func="$COVERAGE_FILE")
OVERALL_COVERAGE=$(echo "$FUNC_COVERAGE" | grep "total:" | awk '{print $3}')

# Create coverage analysis report
cat > "$ANALYSIS_FILE" << EOF
# OptArgs Core Test Coverage Analysis

## Current Coverage Summary

**Overall Coverage: $OVERALL_COVERAGE of statements**

### Per-File Coverage Breakdown

| File | Function | Coverage |
|------|----------|----------|
EOF

# Parse function coverage and add to table
echo "$FUNC_COVERAGE" | grep -v "total:" | while read -r line; do
    if [[ -n "$line" ]]; then
        FILE=$(echo "$line" | awk '{print $1}' | sed 's/.*\///')
        FUNC=$(echo "$line" | awk '{print $2}')
        COV=$(echo "$line" | awk '{print $3}')
        echo "| $FILE | $FUNC | $COV |" >> "$ANALYSIS_FILE"
    fi
done

# Add detailed analysis sections
cat >> "$ANALYSIS_FILE" << 'EOF'

## Detailed Coverage Gaps Analysis

### Critical Functions Requiring 100% Coverage

The following core parsing functions must achieve 100% line and branch coverage:

#### Public API Functions
- **GetOpt**: POSIX getopt(3) implementation
- **GetOptLong**: GNU getopt_long(3) implementation  
- **GetOptLongOnly**: GNU getopt_long_only(3) implementation

#### Core Parsing Functions
- **getOpt**: Internal parsing orchestration
- **findLongOpt**: Long option matching and resolution
- **findShortOpt**: Short option processing and compaction
- **Options**: Iterator-based option processing

#### Error Handling Functions
- **optError**: Error reporting with logging
- **optErrorf**: Formatted error reporting

### Coverage Gap Categories

#### 1. Untested Functions (0% Coverage)
Functions with no test coverage require immediate attention.

#### 2. Partially Covered Functions (<100% Coverage)
Functions with missing code paths, typically error handling or edge cases.

#### 3. Advanced Features
Complex functionality like case-insensitive matching, GNU extensions, and parse modes.

### Testing Recommendations

#### Immediate Priority (Critical)
1. Add comprehensive tests for any 0% coverage functions
2. Cover all error handling paths in partially tested functions
3. Test all parsing modes and configuration options

#### High Priority (Important)
1. Add property-based tests for parsing correctness
2. Test complex option combinations and edge cases
3. Validate POSIX compliance with reference implementations

#### Medium Priority (Enhancement)
1. Add performance benchmarks and memory validation
2. Test cross-platform behavior consistency
3. Add fuzz testing for robustness validation

## Coverage Tracking Setup

### Automated Coverage Commands

```bash
# Generate coverage profile
make coverage

# View HTML coverage report
make coverage-html

# Validate coverage targets
make coverage-validate

# Generate comprehensive analysis
make coverage-report
```

### Coverage Targets

- **Core parsing functions**: 100% line and branch coverage
- **Public API functions**: 100% coverage
- **Error handling paths**: 100% coverage
- **Overall project**: 95% minimum coverage

### CI Integration

Coverage validation is integrated into the CI pipeline:
- Automated coverage generation on every commit
- Coverage regression detection
- Minimum coverage threshold enforcement
- Detailed gap reporting for failed builds

## Next Steps

1. **Review detailed gaps**: Check `coverage_gaps_detailed.md` for specific missing scenarios
2. **Add missing tests**: Focus on 0% coverage functions first
3. **Validate coverage**: Run `make coverage-validate` after adding tests
4. **Monitor progress**: Use `make coverage-func` for quick coverage checks

EOF

echo -e "${GREEN}✓ Coverage analysis generated: $ANALYSIS_FILE${NC}"

# Generate detailed gaps report
cat > "$GAPS_FILE" << 'EOF'
# Detailed Coverage Gaps and Missing Test Scenarios

## Coverage Gap Analysis

This report identifies specific code paths, functions, and scenarios that lack test coverage.

### Gap Identification Process

1. **Function-level analysis**: Identify functions with <100% coverage
2. **Line-level analysis**: Pinpoint specific uncovered code paths
3. **Scenario mapping**: Map uncovered paths to missing test scenarios
4. **Priority assessment**: Categorize gaps by impact and complexity

### Critical Missing Coverage

#### Functions with 0% Coverage
EOF

# Identify functions with 0% coverage
echo "$FUNC_COVERAGE" | grep "0.0%" | while read -r line; do
    if [[ -n "$line" ]]; then
        FILE=$(echo "$line" | awk '{print $1}' | sed 's/.*\///')
        FUNC=$(echo "$line" | awk '{print $2}')
        cat >> "$GAPS_FILE" << EOF

**$FUNC** in $FILE
- **Impact**: HIGH - Completely untested function
- **Required**: Comprehensive test suite covering all code paths
- **Test scenarios needed**: 
  - Basic functionality validation
  - Error condition handling
  - Edge case testing
  - Integration with other components
EOF
    fi
done

# Identify functions with partial coverage
cat >> "$GAPS_FILE" << 'EOF'

#### Functions with Partial Coverage (<100%)

Functions with missing code paths that require additional test coverage:
EOF

echo "$FUNC_COVERAGE" | grep -v "100.0%" | grep -v "0.0%" | grep -v "total:" | while read -r line; do
    if [[ -n "$line" ]]; then
        FILE=$(echo "$line" | awk '{print $1}' | sed 's/.*\///')
        FUNC=$(echo "$line" | awk '{print $2}')
        COV=$(echo "$line" | awk '{print $3}')
        cat >> "$GAPS_FILE" << EOF

**$FUNC** in $FILE ($COV coverage)
- **Status**: Partially tested - missing code paths
- **Priority**: HIGH - Core function requires 100% coverage
- **Action needed**: Identify and test uncovered branches
EOF
    fi
done

# Add testing strategy recommendations
cat >> "$GAPS_FILE" << 'EOF'

## Testing Strategy Recommendations

### Property-Based Testing Opportunities

The following areas would benefit from property-based testing:

1. **Option parsing correctness**: Validate parsing behavior across all input combinations
2. **Round-trip testing**: Parse → generate → parse → verify equivalence
3. **Invariant testing**: Ensure parsing rules are consistently applied
4. **Error handling consistency**: Verify error conditions are handled uniformly

### Unit Testing Gaps

Missing unit test scenarios that should be added:

1. **Error path testing**: Cover all error conditions and edge cases
2. **Configuration testing**: Test all parser configuration combinations
3. **Integration testing**: Test component interactions and workflows
4. **Regression testing**: Prevent known issues from reoccurring

### Test Infrastructure Improvements

Recommended enhancements to the testing infrastructure:

1. **Automated coverage tracking**: CI integration with coverage validation
2. **Performance benchmarking**: Track performance regressions
3. **Fuzz testing**: Discover edge cases through automated input generation
4. **Cross-platform testing**: Ensure consistent behavior across platforms

## Implementation Plan

### Phase 1: Critical Coverage (Immediate)
1. Add tests for all 0% coverage functions
2. Complete error path testing for partially covered functions
3. Achieve 100% coverage for core parsing functions

### Phase 2: Comprehensive Testing (Short-term)
1. Add property-based tests for parsing correctness
2. Implement round-trip testing for all parsing operations
3. Add comprehensive integration tests

### Phase 3: Advanced Validation (Medium-term)
1. Add performance benchmarks and regression testing
2. Implement fuzz testing for robustness validation
3. Add cross-platform behavior validation

## Coverage Monitoring

### Automated Tracking
- Coverage reports generated on every test run
- Regression detection for coverage decreases
- Minimum coverage thresholds enforced in CI

### Manual Review Process
- Regular coverage gap analysis and reporting
- Prioritization of uncovered code paths
- Test scenario planning and implementation

### Success Metrics
- **100% coverage** for all core parsing functions
- **95% overall coverage** for the entire codebase
- **Zero regression** in coverage over time
- **Comprehensive test scenarios** for all functionality

EOF

echo -e "${GREEN}✓ Detailed gaps analysis generated: $GAPS_FILE${NC}"

# Generate summary statistics
TOTAL_FUNCTIONS=$(echo "$FUNC_COVERAGE" | grep -v "total:" | wc -l)
ZERO_COVERAGE=$(echo "$FUNC_COVERAGE" | grep "0.0%" | wc -l)
PARTIAL_COVERAGE=$(echo "$FUNC_COVERAGE" | grep -v "100.0%" | grep -v "0.0%" | grep -v "total:" | wc -l)
FULL_COVERAGE=$(echo "$FUNC_COVERAGE" | grep "100.0%" | wc -l)

echo ""
echo "Coverage Analysis Summary:"
echo "========================="
echo "Total functions: $TOTAL_FUNCTIONS"
echo "Functions with 100% coverage: $FULL_COVERAGE"
echo "Functions with partial coverage: $PARTIAL_COVERAGE"
echo "Functions with 0% coverage: $ZERO_COVERAGE"
echo "Overall coverage: $OVERALL_COVERAGE"
echo ""
echo -e "${YELLOW}Reports generated:${NC}"
echo "  - $ANALYSIS_FILE (comprehensive analysis)"
echo "  - $GAPS_FILE (detailed gap identification)"
echo ""
echo "Next steps:"
echo "1. Review the generated reports"
echo "2. Run 'make coverage-validate' to see current status"
echo "3. Add tests for identified gaps"
echo "4. Re-run analysis after adding tests"
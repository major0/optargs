#!/bin/bash
# Coverage validation script for OptArgs Core
# Validates that coverage meets 100% target for core parsing functions

set -e

COVERAGE_FILE="${1:-coverage.out}"
CORE_FUNCTIONS_TARGET=100.0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if coverage file exists
if [[ ! -f "$COVERAGE_FILE" ]]; then
    echo -e "${RED}Error: Coverage file $COVERAGE_FILE not found${NC}"
    echo "Run 'make coverage' first to generate coverage data"
    exit 1
fi

echo "Coverage Validation Report"
echo "========================="
echo "Target: ${CORE_FUNCTIONS_TARGET}% for core parsing functions"
echo "Coverage file: $COVERAGE_FILE"
echo ""

# Extract overall coverage percentage
OVERALL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep "total:" | awk '{print $3}' | sed 's/%//')

echo "Overall Coverage: ${OVERALL_COVERAGE}%"

# Define core parsing functions that must have 100% coverage
CORE_FUNCTIONS=(
    "GetOpt"
    "GetOptLong"
    "GetOptLongOnly"
    "getOpt"
    "findLongOpt"
    "findShortOpt"
    "Options"
    "optError"
    "optErrorf"
)

# Check coverage for each core function
FAILED_FUNCTIONS=()
PASSED_FUNCTIONS=()

echo ""
echo "Core Function Coverage Analysis:"
echo "================================"

for func in "${CORE_FUNCTIONS[@]}"; do
    # Extract coverage for this function
    FUNC_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep -E "\s${func}\s" | awk '{print $3}' | sed 's/%//' | head -1)

    if [[ -z "$FUNC_COVERAGE" ]]; then
        echo -e "${YELLOW}Warning: Function $func not found in coverage report${NC}"
        continue
    fi

    # Compare coverage (using bc for floating point comparison)
    if (( $(echo "$FUNC_COVERAGE >= $CORE_FUNCTIONS_TARGET" | bc -l) )); then
        echo -e "${GREEN}✓ $func: ${FUNC_COVERAGE}%${NC}"
        PASSED_FUNCTIONS+=("$func")
    else
        echo -e "${RED}✗ $func: ${FUNC_COVERAGE}% (target: ${CORE_FUNCTIONS_TARGET}%)${NC}"
        FAILED_FUNCTIONS+=("$func")
    fi
done

echo ""
echo "Validation Summary:"
echo "=================="
echo "Functions meeting target: ${#PASSED_FUNCTIONS[@]}"
echo "Functions below target: ${#FAILED_FUNCTIONS[@]}"

# Check if overall coverage meets minimum threshold
MINIMUM_OVERALL=90.0
if (( $(echo "$OVERALL_COVERAGE >= $MINIMUM_OVERALL" | bc -l) )); then
    echo -e "${GREEN}✓ Overall coverage: ${OVERALL_COVERAGE}% (minimum: ${MINIMUM_OVERALL}%)${NC}"
else
    echo -e "${RED}✗ Overall coverage: ${OVERALL_COVERAGE}% (minimum: ${MINIMUM_OVERALL}%)${NC}"
    FAILED_FUNCTIONS+=("overall")
fi

# Report results
if [[ ${#FAILED_FUNCTIONS[@]} -eq 0 ]]; then
    echo ""
    echo -e "${GREEN}🎉 All coverage targets met!${NC}"
    echo "Core parsing functions have achieved 100% coverage target"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Coverage validation failed${NC}"
    echo "The following functions need additional test coverage:"
    for func in "${FAILED_FUNCTIONS[@]}"; do
        echo "  - $func"
    done
    echo ""
    echo "Next steps:"
    echo "1. Review coverage_gaps_detailed.md for specific missing test scenarios"
    echo "2. Add tests for uncovered code paths"
    echo "3. Run 'make coverage-html' to see detailed coverage visualization"
    echo "4. Re-run 'make coverage-validate' after adding tests"
    exit 1
fi

#!/bin/bash
# Coverage validation script for OptArgs Core
# Validates that coverage meets the minimum floor per testing-standards.md

set -e

COVERAGE_FILE="${1:-coverage.out}"
MINIMUM_OVERALL=70.0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Check if coverage file exists
if [[ ! -f "$COVERAGE_FILE" ]]; then
    echo -e "${RED}Error: Coverage file $COVERAGE_FILE not found${NC}"
    echo "Run 'make coverage' first to generate coverage data"
    exit 1
fi

echo "Coverage Validation Report"
echo "========================="
echo "Minimum floor: ${MINIMUM_OVERALL}%"
echo "Coverage file: $COVERAGE_FILE"
echo ""

# Extract overall coverage percentage
OVERALL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep "total:" | awk '{print $3}' | sed 's/%//')

echo "Overall Coverage: ${OVERALL_COVERAGE}%"

if (( $(echo "$OVERALL_COVERAGE >= $MINIMUM_OVERALL" | bc -l) )); then
    echo ""
    echo -e "${GREEN}✓ Coverage ${OVERALL_COVERAGE}% meets minimum floor ${MINIMUM_OVERALL}%${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Coverage ${OVERALL_COVERAGE}% is below minimum floor ${MINIMUM_OVERALL}%${NC}"
    exit 1
fi

#!/usr/bin/env bash
set -euo pipefail

# Generates a combined coverage summary from downloaded artifacts.
#
# Args: ARTIFACTS_DIR
# Outputs: coverage-summary.md in current directory

ARTIFACTS_DIR="$1"

{
	echo "# Combined Coverage Summary"
	echo ""

	if [ -f "${ARTIFACTS_DIR}/optargs-coverage-reports/coverage_analysis.md" ]; then
		echo "## OptArgs Core Module"
		grep "Overall Coverage:" "${ARTIFACTS_DIR}/optargs-coverage-reports/coverage_analysis.md" || echo "Coverage data not available"
		echo ""
	fi

	if [ -f "${ARTIFACTS_DIR}/goarg-coverage-reports/coverage.out" ]; then
		echo "## GoArg Module"
		go tool cover -func="${ARTIFACTS_DIR}/goarg-coverage-reports/coverage.out" | grep "total:" || echo "Coverage data not available"
		echo ""
	fi

	if [ -f "${ARTIFACTS_DIR}/pflags-coverage-reports/coverage.out" ]; then
		echo "## PFlags Module"
		go tool cover -func="${ARTIFACTS_DIR}/pflags-coverage-reports/coverage.out" | grep "total:" || echo "Coverage data not available"
		echo ""
	fi
} > coverage-summary.md

cat coverage-summary.md

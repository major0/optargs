# OptArgs Core - Test Coverage Tracking Makefile
# Implements automated coverage reporting and validation

.PHONY: test coverage coverage-html coverage-func coverage-validate coverage-report clean help

# Default target
help:
	@echo "OptArgs Core Coverage Tracking"
	@echo "=============================="
	@echo ""
	@echo "Available targets:"
	@echo "  test              - Run all tests"
	@echo "  coverage          - Generate coverage profile"
	@echo "  coverage-html     - Generate HTML coverage report"
	@echo "  coverage-func     - Display function-level coverage"
	@echo "  coverage-validate - Validate coverage meets targets"
	@echo "  coverage-report   - Generate comprehensive coverage analysis"
	@echo "  clean             - Remove coverage files"
	@echo "  help              - Show this help message"
	@echo ""
	@echo "Coverage targets:"
	@echo "  - Core parsing functions: 100% line and branch coverage"
	@echo "  - Public API functions: 100% coverage"
	@echo "  - Error handling paths: 100% coverage"
	@echo "  - All parsing modes: 100% coverage"

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Generate coverage profile with atomic mode for accurate branch coverage
coverage:
	@echo "Generating coverage profile..."
	go test -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Coverage profile generated: coverage.out"

# Generate HTML coverage report
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report generated: coverage.html"

# Display function-level coverage summary
coverage-func: coverage
	@echo "Function-level coverage summary:"
	@echo "================================"
	go tool cover -func=coverage.out

# Validate coverage meets 100% target for core functions
coverage-validate: coverage
	@echo "Validating coverage targets..."
	@./scripts/validate_coverage.sh coverage.out

# Generate comprehensive coverage analysis and gap report
coverage-report: coverage coverage-html
	@echo "Generating comprehensive coverage analysis..."
	@./scripts/generate_coverage_report.sh coverage.out
	@echo "Coverage analysis complete. Check coverage_analysis.md and coverage_gaps_detailed.md"

# Clean coverage files
clean:
	@echo "Cleaning coverage files..."
	rm -f coverage.out coverage.html cover.out
	rm -f coverage_analysis.md coverage_gaps_detailed.md
	@echo "Coverage files cleaned"

# CI target for automated validation
ci-coverage: coverage coverage-validate
	@echo "CI coverage validation complete"

# Development target for quick feedback
dev-coverage: coverage coverage-func
	@echo "Development coverage check complete"
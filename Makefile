# OptArgs Core - Test Coverage Tracking Makefile
# Implements automated coverage reporting and validation

.PHONY: test coverage coverage-html coverage-func coverage-validate coverage-report clean help lint static-check security-check fmt imports vet mod-tidy mod-verify build-check

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
	@echo "  lint              - Run golangci-lint static analysis"
	@echo "  static-check      - Run comprehensive static analysis"
	@echo "  security-check    - Run security vulnerability checks"
	@echo "  fmt               - Format Go code"
	@echo "  imports           - Fix Go imports"
	@echo "  vet               - Run go vet"
	@echo "  mod-tidy          - Tidy go.mod and go.sum"
	@echo "  mod-verify        - Verify go.mod dependencies"
	@echo "  build-check       - Verify code builds successfully"
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

# Static Analysis Targets
# =======================

# Run golangci-lint with comprehensive static analysis
lint:
	@echo "Running golangci-lint static analysis..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config .golangci.yml ./...; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Comprehensive static analysis suite
static-check: fmt imports vet mod-tidy mod-verify lint build-check
	@echo "Static analysis complete"

# Security vulnerability checks
security-check:
	@echo "Running security vulnerability checks..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
		echo "Running go list -json -deps ./... | nancy sleuth instead..."; \
		if command -v nancy >/dev/null 2>&1; then \
			go list -json -deps ./... | nancy sleuth; \
		else \
			echo "Neither govulncheck nor nancy found. Skipping security check."; \
		fi; \
	fi

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Fix Go imports
imports:
	@echo "Fixing Go imports..."
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
		echo "Using go fmt instead..."; \
		go fmt ./...; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Tidy go.mod and go.sum
mod-tidy:
	@echo "Tidying go.mod and go.sum..."
	go mod tidy

# Verify go.mod dependencies
mod-verify:
	@echo "Verifying go.mod dependencies..."
	go mod verify

# Verify code builds successfully
build-check:
	@echo "Verifying code builds successfully..."
	go build -v ./...

# Pre-commit checks (runs all static analysis)
pre-commit: static-check test
	@echo "Pre-commit checks complete"

# CI static analysis target
ci-static: static-check security-check
	@echo "CI static analysis complete"

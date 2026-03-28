.PHONY: test coverage coverage-html coverage-func coverage-validate lint static-check fmt imports vet mod-tidy mod-verify build-check clean help pre-commit compat-test compat-update compat-diff

help:
	@echo "PFlags Wrapper"
	@echo "=============="
	@echo ""
	@echo "Targets:"
	@echo "  test              - Run all tests"
	@echo "  coverage          - Generate coverage profile"
	@echo "  coverage-html     - Generate HTML coverage report"
	@echo "  coverage-func     - Display function-level coverage"
	@echo "  coverage-validate - Validate coverage meets targets"
	@echo "  lint              - Run golangci-lint"
	@echo "  static-check      - Run all static analysis"
	@echo "  fmt               - Format Go code"
	@echo "  vet               - Run go vet"
	@echo "  build-check       - Verify code builds"
	@echo "  clean             - Remove generated files"
	@echo "  pre-commit        - Run all pre-commit checks"
	@echo "  compat-test       - Validate against upstream golden outputs"
	@echo "  compat-update     - Regenerate upstream golden outputs"
	@echo "  compat-diff       - Show all upstream differences"

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

coverage-func: coverage
	go tool cover -func=coverage.out

coverage-validate: coverage
	@../scripts/validate_coverage.sh coverage.out

lint:
	golangci-lint run --config .golangci.yml ./...

static-check: fmt imports vet mod-tidy mod-verify lint build-check

fmt:
	go fmt ./...

imports:
	@if command -v goimports >/dev/null 2>&1; then goimports -w .; else go fmt ./...; fi

vet:
	go vet ./...

mod-tidy:
	go mod tidy

mod-verify:
	go mod verify

build-check:
	go build -v ./...

clean:
	rm -f coverage.out coverage.html

pre-commit: static-check test

# Compatibility testing targets
compat-test:
	cd compat && go test -v ./...
	go test -v -run TestCompat ./...

compat-update:
	cd compat && go test -update ./...

compat-diff:
	@echo "=== Upstream golden files ===" && ls -1 compat/testdata/*.golden 2>/dev/null || echo "No golden files"
	@echo "=== Running compat tests ===" && go test -v -run TestCompat ./... 2>&1 || true

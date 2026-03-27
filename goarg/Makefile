# GoArg Module - Makefile
# Standard targets + compatibility testing stubs

.PHONY: test coverage coverage-html coverage-func coverage-validate lint vet fmt \
        mod-tidy mod-verify build-check static-check clean help \
        compat-test compat-update compat-diff

GOLANGCI_CONFIG := $(shell if [ -f ../.golangci.yml ]; then echo "--config ../.golangci.yml"; fi)

help:
	@echo "GoArg Module"
	@echo "============"
	@echo ""
	@echo "Standard targets:"
	@echo "  test              - Run all tests"
	@echo "  coverage          - Generate coverage profile"
	@echo "  coverage-html     - Generate HTML coverage report"
	@echo "  coverage-func     - Display function-level coverage"
	@echo "  coverage-validate - Validate coverage meets targets"
	@echo "  lint              - Run golangci-lint"
	@echo "  static-check      - Run all static analysis"
	@echo "  clean             - Remove generated files"
	@echo ""
	@echo "Compatibility targets:"
	@echo "  compat-test       - Validate against golden files"
	@echo "  compat-update     - Regenerate golden files"
	@echo "  compat-diff       - Show all differences"

test:
	go test -v -count=1 ./...

coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

coverage-func: coverage
	go tool cover -func=coverage.out

coverage-validate: coverage
	@../scripts/validate_coverage.sh coverage.out

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run $(GOLANGCI_CONFIG) ./...; \
	else \
		echo "golangci-lint not found"; \
		exit 1; \
	fi

vet:
	go vet ./...

fmt:
	go fmt ./...

mod-tidy:
	go mod tidy

mod-verify:
	go mod verify

build-check:
	go build -v ./...

static-check: fmt vet mod-tidy mod-verify lint build-check

clean:
	rm -f coverage.out coverage.html

# Compatibility targets (stubs — populated by Task 10)
compat-test:
	@echo "Compatibility tests not yet configured (see Task 10)"

compat-update:
	@echo "Compatibility golden file update not yet configured (see Task 10)"

compat-diff:
	@echo "Compatibility diff not yet configured (see Task 10)"

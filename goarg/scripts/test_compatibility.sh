#!/bin/bash

# Compatibility testing script for go-arg
# This script helps test compatibility between our implementation and upstream alexflint/go-arg

set -e

echo "go-arg Compatibility Testing Script"
echo "=================================="

# Function to run tests with our implementation
test_our_implementation() {
    echo "Testing with our implementation..."
    go mod edit -dropreplace github.com/alexflint/go-arg 2>/dev/null || true
    go test -v .
}

# Function to run tests with upstream implementation
test_upstream_implementation() {
    echo "Testing with upstream alexflint/go-arg..."
    if go mod edit -replace github.com/alexflint/go-arg=github.com/alexflint/go-arg@v1.4.3; then
        go mod tidy
        go test -v .
        # Always switch back
        go mod edit -dropreplace github.com/alexflint/go-arg
        go mod tidy
    else
        echo "Warning: Could not switch to upstream implementation"
        echo "This is expected during development when upstream is not available"
    fi
}

# Function to validate API compatibility
validate_api() {
    echo "Validating API compatibility..."
    go run -c "
package main

import (
    \"fmt\"
    \"reflect\"
    \"github.com/major0/optargs/goarg\"
)

func main() {
    // Check that our Parser has the expected methods
    parserType := reflect.TypeOf(&goarg.Parser{})
    methods := []string{\"Parse\", \"WriteHelp\", \"WriteUsage\", \"Fail\"}

    for _, method := range methods {
        if _, found := parserType.MethodByName(method); !found {
            fmt.Printf(\"Missing method: %s\n\", method)
            return
        }
    }

    fmt.Println(\"API compatibility validation passed\")
}
"
}

# Main execution
case "${1:-all}" in
    "our")
        test_our_implementation
        ;;
    "upstream")
        test_upstream_implementation
        ;;
    "api")
        validate_api
        ;;
    "all")
        echo "Running comprehensive compatibility tests..."
        test_our_implementation
        echo ""
        test_upstream_implementation
        echo ""
        validate_api
        ;;
    *)
        echo "Usage: $0 [our|upstream|api|all]"
        echo "  our      - Test with our implementation only"
        echo "  upstream - Test with upstream alexflint/go-arg"
        echo "  api      - Validate API compatibility"
        echo "  all      - Run all tests (default)"
        exit 1
        ;;
esac

echo ""
echo "Compatibility testing complete!"

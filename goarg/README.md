# go-arg Compatibility Layer

This package provides 100% API compatibility with [alexflint/go-arg](https://github.com/alexflint/go-arg) while leveraging OptArgs Core's POSIX/GNU compliance.

## Project Structure

```
goarg/
├── doc.go                    # Package documentation
├── parser.go                 # Main go-arg API (Parser, Config, main functions)
├── tags.go                   # Struct tag processing (TagParser, StructMetadata, FieldMetadata)
├── core_integration.go       # Direct OptArgs Core integration
├── types.go                  # Type conversion system
├── help.go                   # Help generation and error handling
├── compatibility_test.go     # Compatibility testing framework
├── module_alias_test.go      # Module alias switching for testing
├── testing_framework.go     # Main compatibility test framework
├── framework_test.go         # Framework validation tests
├── go.mod                    # Module configuration with alias support
└── README.md                 # This file
```

## Architecture

The architecture is intentionally simple with two main layers:

1. **go-arg API Layer**: Provides 100% compatibility with alexflint/go-arg
2. **OptArgs Core**: Direct integration without intermediate layers

Extensions will be handled through separate `-ext.go` files that can be included/excluded at build time.

## Compatibility Testing Framework

The package includes a comprehensive compatibility testing framework that:

- Supports module alias switching between our implementation and upstream alexflint/go-arg
- Validates API compatibility at the interface level
- Runs identical test scenarios against both implementations
- Generates detailed compatibility reports

### Module Alias Testing

The framework supports switching between implementations using Go module aliases:

```bash
# Test with our implementation (default)
go test ./...

# Test with upstream alexflint/go-arg
go mod edit -replace github.com/alexflint/go-arg=github.com/alexflint/go-arg@v1.4.3
go test ./...
go mod edit -dropreplace github.com/alexflint/go-arg
```

## Development Status

This is the initial project structure setup. Core functionality will be implemented in subsequent tasks:

- [ ] Core go-arg API implementation
- [ ] Struct tag processing
- [ ] OptArgs Core integration
- [ ] Type conversion system
- [ ] Help generation and error handling
- [ ] Comprehensive compatibility testing

## Requirements Addressed

- **Requirement 3.1**: Module alias testing configuration for alexflint/go-arg compatibility
- **Requirement 8.4**: Compatibility testing framework interfaces

## Usage

Once fully implemented, this package will be a drop-in replacement for alexflint/go-arg:

```go
import "github.com/major0/optargs/goarg"

type Args struct {
    Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
    Count   int    `arg:"-c,--count" help:"number of items"`
    Files   []string `arg:"positional" help:"files to process"`
}

func main() {
    var args Args
    goarg.MustParse(&args)
    // Use args...
}
```

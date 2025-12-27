# OptArgs Examples

This directory contains working code examples for all the features documented in the main README.md.

## Running Examples

Each example is a complete Go program that demonstrates a specific feature:

```bash
# Run any example
go run docs/examples/non_argument_options.go
go run docs/examples/short_only_options.go
go run docs/examples/count_options.go
go run docs/examples/complex_data_structures.go
go run docs/examples/advanced_gnu_features.go
go run docs/examples/automatic_help.go
```

## Example Files

- **non_argument_options.go** - Boolean flags that don't require values
- **short_only_options.go** - Single-character convenience flags
- **count_options.go** - Repeatable flags for incrementing counters
- **complex_data_structures.go** - Slices, maps, and custom data types
- **advanced_gnu_features.go** - Special characters and longest matching
- **automatic_help.go** - Custom help text generation

## Features Demonstrated

Each example shows:
- How to define flags using the pflags API
- Command-line usage patterns
- Expected output and behavior
- Integration with OptArgs Core parsing

## API Compatibility

All examples use the pflags API which maintains compatibility with spf13/pflag while leveraging OptArgs Core's POSIX/GNU compliance and advanced features.
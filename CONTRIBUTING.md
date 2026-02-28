# Contributing to OptArgs

## Dev Setup

```bash
git clone https://github.com/major0/optargs.git
cd optargs
go mod tidy
```

Requires Go 1.23+ and [pre-commit](https://pre-commit.com/#install).

Install pre-commit hooks:

```bash
pre-commit install
```

## Testing

```bash
# Run all tests
make test

# Coverage report
make coverage-html

# Validate coverage targets
make coverage-validate

# Full static analysis + tests
make pre-commit
```

Property-based tests use `testing/quick` with 100+ iterations. Test files use `_test.go` suffix; property tests use the `Property` prefix.

## Code Style

- `go fmt` and `goimports` enforced
- `golangci-lint` must pass with zero issues
- Simple, readable over clever — see design philosophy in project steering

## PR Workflow

1. Fork and create a feature branch
2. Use [Conventional Commits](https://www.conventionalcommits.org) for commit messages
3. Ensure `make pre-commit` passes locally
4. Submit a pull request — CI runs pre-commit, build, and coverage checks
5. Maintain or improve test coverage

## Reporting Bugs

Open an issue on [GitHub](https://github.com/major0/optargs/issues).

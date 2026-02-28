# POSIX/GNU getopt Patterns

This directory contains two things:

1. **Shell scripts** — demonstrations of valid but surprising POSIX/GNU `getopt(3)` behavior using the `getopt(1)` utility. Run `./getopt_examples.sh` to execute all scripts.

2. **Go programs** — OptArgs implementations of obscure/important POSIX patterns that users may not be familiar with.

## Go Programs

| Directory | Pattern | Description |
|-----------|---------|-------------|
| `subcommand/` | Native subcommand dispatch | Multi-level dispatch via `AddCmd()` with option inheritance through the parent chain |
| `silent/` | Silent error mode | `:` prefix suppresses error logging; caller handles errors via iterator |
| `posixly_correct/` | POSIXLY_CORRECT | `+` prefix stops parsing at first non-option argument |

### Running

```bash
go run ./posix/subcommand
go run ./posix/silent
go run ./posix/posixly_correct
```

## Shell Scripts

The `*.sh` files demonstrate valid GNU/POSIX `getopt(3)` patterns using the util-linux `getopt(1)` command. These serve as reference material for the parser's expected behavior.

Requires util-linux `getopt(1)` version 4 (enhanced mode).

## Notes

Where `getopt(1)` differs from `getopt(3)`, OptArgs defers to `getopt(3)` behavior.

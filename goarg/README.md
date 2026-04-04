# goarg — go-arg compatibility layer

Drop-in replacement for [alexflint/go-arg](https://github.com/alexflint/go-arg)
backed by OptArgs Core's POSIX/GNU getopt implementation.

## Quick start

```go
import "github.com/major0/optargs/goarg"

type Args struct {
    Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
    Count   int    `arg:"-c,--count" default:"1" help:"repeat count"`
    Output  string `arg:"-o,--output" help:"output file"`
    Source  string `arg:"positional,required" help:"source file"`
}

func main() {
    var args Args
    goarg.MustParse(&args)
}
```

## Features

All upstream go-arg features are supported:

- Struct tag parsing (`arg`, `help`, `default`, `env`)
- Short and long options (`-v`, `--verbose`)
- Positional arguments (required and optional)
- Subcommands via pointer-to-struct fields
- Environment variable fallback (`env:VAR_NAME`)
- Env-only fields (no CLI flag, only env var)
- Default values from struct tags
- Map types (`--header Content-Type=application/json`)
- Slice types (repeated flags append)
- Embedded struct field inheritance
- `Versioned`, `Described`, `Epilogued` interfaces
- `ErrHelp` / `ErrVersion` sentinel errors
- Builtin `-h`/`--help` and `--version` flags
- Case-insensitive subcommand matching
- `Subcommand()` / `SubcommandNames()` query methods

## Core integration benefits

goarg delegates all parsing to OptArgs Core, which provides:

- Full POSIX getopt(3) and GNU getopt_long(3) compliance
- Short-option compaction (`-abc` = `-a -b -c`)
- Optional arguments (`-d::` / `--debug::`)
- Parent-chain option inheritance across subcommands
- `StrictSubcommands` mode (disable inheritance)
- `POSIXLY_CORRECT` environment variable support

## Extension system

Build with `-tags goarg_ext` to enable extensions:

```bash
go build -tags goarg_ext
```

Extensions add capabilities beyond base go-arg compatibility (error-returning
help methods, extended tag syntax). See `ext.go` for details.

## Compatibility testing

Golden-file tests validate our output against upstream `alexflint/go-arg`:

```bash
make compat-test    # validate against golden files
make compat-update  # regenerate golden files from upstream
make compat-diff    # show documented divergences
```

Known divergences are documented in `expected_diffs.go`.

## Feature Comparison

| Feature | Upstream go-arg | goarg/ (compat) |
|---------|:-:|:-:|
| Struct tag parsing | ✅ | ✅ |
| Short/long options | ⚠️³ | ✅ |
| Positional arguments | ✅ | ✅ |
| Subcommands | ✅ | ✅ |
| Environment variable fallback | ✅ | ✅ |
| Env-only fields | ✅ | ✅ |
| Default values | ✅ | ✅ |
| Map types | ✅ | ⚠️¹ |
| Slice types (repeated flags) | ✅ | ⚠️² |
| Embedded struct inheritance | ✅ | ✅ |
| Versioned/Described/Epilogued | ✅ | ✅ |
| ErrHelp/ErrVersion sentinels | ✅ | ✅ |
| Builtin help/version flags | ✅ | ✅ |
| Subcommand()/SubcommandNames() | ✅ | ✅ |
| [POSIX short-option compaction (-abc)](../docs/short-option-compaction.md) | ❌ | ✅ |
| [GNU longest-match option resolution](../docs/prefix-matching.md) | ❌ | ✅ |
| [Boolean negation (--no-flag)](../docs/boolean-negation.md) | ❌ | ✅ |
| `--` termination | ✅ | ✅ |
| [Parent flag inheritance across subcommands](../docs/subcommand-inheritance.md) (alexflint/go-arg#101) | ✅ | ✅ |
| Case-insensitive subcommand matching | ❌ | ✅ |
| Interspersed argument handling | ✅ | ✅ |
| [getopt_long_only mode](../docs/long-only-mode.md) | ⚠️⁴ | ❌ |

¹ Upstream resets map on each repeated flag; ours merges entries (POSIX semantics).
² Upstream resets slice on each repeated flag; ours appends (POSIX semantics).
³ Upstream go-arg operates exclusively in `getopt_long_only(3)` mode — all options
  (including `-v`) are parsed as long options. True POSIX short options and
  compaction (`-abc`) are not supported. See footnote ⁴.
⁴ Upstream go-arg is always in `getopt_long_only(3)` mode and cannot leave it.
  Single-dash arguments like `-verbose` and `-v=true` are parsed as long options.
  There is no way to enable true POSIX short-option behavior.
  See `expected_diffs.go` for all documented divergences.
  Every ✅ and ❌ is backed by a test — see `compat/table_validation_test.go` and `table_validation_test.go`.

## Config

```go
goarg.Config{
    Program:           "myapp",
    Description:       "My application",
    Version:           "1.0.0",
    StrictSubcommands: true,   // disable parent option inheritance
    IgnoreEnv:         false,  // skip env var processing
    IgnoreDefault:     false,  // skip default value application
    EnvPrefix:         "APP",  // prefix for env var names
    Exit:              os.Exit,
    Out:               os.Stderr,
}
```

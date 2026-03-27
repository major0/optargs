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

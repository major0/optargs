# OptArgs

[![Build](https://github.com/major0/optargs/actions/workflows/build.yml/badge.svg)](https://github.com/major0/optargs/actions/workflows/build.yml)
[![Coverage](https://github.com/major0/optargs/actions/workflows/coverage.yml/badge.svg)](https://github.com/major0/optargs/actions/workflows/coverage.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/major0/optargs.svg)](https://pkg.go.dev/github.com/major0/optargs)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Go implementation of POSIX [getopt(3)](https://pubs.opengroup.org/onlinepubs/9699919799/functions/getopt.html), GNU [getopt_long(3)](https://man7.org/linux/man-pages/man3/getopt.3.html), and [getopt_long_only(3)](https://man7.org/linux/man-pages/man3/getopt.3.html) with native subcommand support.

OptArgs Core is the foundation for higher-level wrapper interfaces ([goarg](goarg/), [pflags](pflags/)).

## Features

- Full POSIX getopt(3) compliance: short options, compaction, `--` termination, `POSIXLY_CORRECT`
- GNU getopt_long(3): `--option=value`, `--option value`, partial matching, case-insensitive
- GNU getopt_long_only(3): single-dash long options with short option fallback
- Advanced handling: `-W` extension, option redefinition, negative arguments
- Native subcommand dispatch via `AddCmd()` with option inheritance through the parent chain
- Iterator-based processing (`Options()` returns `iter.Seq2[Option, error]`)
- Verbose and silent error modes, both working through subcommand chains
- Zero dependencies

## Install

```bash
go get github.com/major0/optargs
```

Requires Go 1.23+.

## Usage

### GetOpt (POSIX short options)

```go
package main

import (
    "fmt"
    "github.com/major0/optargs"
)

func main() {
    p, _ := optargs.GetOpt(os.Args[1:], "vf:o::")
    for opt, err := range p.Options() {
        if err != nil {
            fmt.Fprintf(os.Stderr, "%v\n", err)
            continue
        }
        switch opt.Name {
        case "v":
            fmt.Println("verbose")
        case "f":
            fmt.Println("file:", opt.Arg)
        case "o":
            fmt.Println("output:", opt.Arg)
        }
    }
}
```

### GetOptLong (GNU long options)

```go
p, _ := optargs.GetOptLong(os.Args[1:], "vf:", []optargs.Flag{
    {Name: "verbose", HasArg: optargs.NoArgument},
    {Name: "file",    HasArg: optargs.RequiredArgument},
    {Name: "output",  HasArg: optargs.OptionalArgument},
})
for opt, err := range p.Options() {
    // ...
}
```

### GetOptLongOnly (single-dash long options)

```go
p, _ := optargs.GetOptLongOnly(os.Args[1:], "vf:", []optargs.Flag{
    {Name: "verbose", HasArg: optargs.NoArgument},
    {Name: "file",    HasArg: optargs.RequiredArgument},
})
// -verbose tries long match first, falls back to short options via optstring
```

### Subcommands

```go
root, _ := optargs.GetOptLong(os.Args[1:], "v", []optargs.Flag{
    {Name: "verbose", HasArg: optargs.NoArgument},
})

serve, _ := optargs.GetOptLong([]string{}, "p:", []optargs.Flag{
    {Name: "port", HasArg: optargs.RequiredArgument},
})
root.AddCmd("serve", serve)

// Root iteration dispatches to child when "serve" is encountered.
// Child inherits parent options via parent-chain walk.
for opt, err := range root.Options() { /* root options */ }
for opt, err := range serve.Options() { /* serve options + inherited */ }
```

## Optstring Syntax

| Prefix | Behavior |
|--------|----------|
| `:` | Silent error mode — suppress error logging |
| `+` | POSIXLY_CORRECT — stop at first non-option |
| `-` | Treat non-options as argument to option `\x01` |

| Suffix | Meaning |
|--------|---------|
| `f` | No argument |
| `f:` | Required argument |
| `f::` | Optional argument |
| `W;` | GNU `-W` word extension |

## Examples

- [`example/`](example/) — vanilla GetOpt, GetOptLong, GetOptLongOnly usage
- [`posix/`](posix/) — obscure POSIX/GNU patterns: subcommand dispatch, silent error mode, POSIXLY_CORRECT

## Wrapper Modules

| Module | Description |
|--------|-------------|
| [goarg](goarg/) | Struct-tag based argument parsing (alexflint/go-arg compatible) |
| [pflags](pflags/) | Flag-method based parsing (spf13/pflag compatible) |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)

# OptArgs
This is a [Go](https://golang.org/) library for parsing optional arguments from
the command line.

This package aims to be a clean, small, and infinately extensible library for
parsing the optional arguments on the CLI.

At the heart of this design is a strict mandate on what CLI arguments should
look like. This tool defers to definitions set down by the POSIX `getopt()`
definitions, as well as the extended `getopt()` and `getopt_long()` syntax
provided by GNU.

This means that this tool is 100% compatible with the _interpretation_ of the
CLI from a POSIX/GNU perspective, and is capable of reproducing any optional
argument pairing that would normally be capable with `getopt()` and
`getopt_long()`. This also means that this tool does _not_ dictate any sort of
rules about what a good UI should look like for the end user. Developers are
free to come up with whatever CLI they feel best fits their user base.

This tool deviates from `getopt()` and `getopt_long` in a few aspects:
1. While this library considers interpretation of the CLI as standardized by
   POSIX/GNU, the library does not attempt to implement `getopt()` or
   `getopt_long()`.

2. This is not a C/C++ library, but a [Go](https://golang.org/) library, and
   a different approach is taken to achieving the same level of user interaction.

3. This library supports _chaining_ parsers to allow for implementing sub-commands.

## CLI Syntax
OptArgs considers the following syntax/rules as the strict CLI policy for all
options:

- Encountering a double-hypen (`--`) with no argument text shall terminate any
  further option parsing after removing the `--` from the argument list.

- A `PosixlyCorrect` flag can be set which will terminate all parsing whenever
  an unknown non-option/command is encountered. This effectively prevents the
  CLI from supporting adding `options` to the end of the CLI.
  Example: `myCommand <file> --help` will not work if `PosixlyCorrect` is set.

- Irrespective of the `PosixlyCorrect` flag, all arguments for a sub-command
  must appear somewhere _after_ the sub-command in the argument list.

_Note: The `POSIXLY_CORRECT` environment variable will influence this behavior,
as is required by the POSIX standard._

### Short Options
- All short options start with a single hyphen. (`-`)
- All short options are case-sensitive _by default_.
- Arguments to short options are space delimited, that means they only may
  appear as the next element in the argument list.
- Short options may be _compacted_ such that all nearby short options are
  grouped together proceeded with a single hypen (`-`).
- An optional argument may be passed to _compacted_ options. In doing so the
  optional argument is assumed to belong to the last option in the compacted
  list _which accepts an optional argument_.

##*# Allowed Patterns
- `-[:alnum:]`: E.g. `-1`, `-2`, `-3`, `-a`, `-b`, `-c`, ...
- `-[:alnum:] <arg>`: E.g. `-f ARG`
- `-[:alnum:]+`: E.g. `-abc123`
- `-[:alnum:]+ <arg>`: E.g. `-abc123 ARG`

During argument compaction, any optional argument does not need to
belong to the _**last**_ option, but rather the _**last option which supports
an optional argument**_.

Example from tar: `tar -tfv file.tar`, where the `-f` option is passed
`file.tar`.

### Long Options
- All long-options start with a double hyphen. (`--`)
- All long options are case-insensitive _by default_.
- Arguments to long options can be part of the same argument separated by an
  equal sign (`=`), or can appear as the next argument in the list; should the
  next argument not begin with a hyphen (`-`).
- Long arguments may comprise of a single character, e.g. `--f`. Such arguments
  are still parsed according to the long-option syntax.
- Long arguments do not support argument compaction.

### Allowed Patterns
- `--[[:alnum:]_-]+`: E.g. `--1`, `--two`, `--a`, `--bee`
- `--[[:alnum:]_-]+=<arg>`: E.g. `--foo=bar`, `--2fa=1234`
- `--[[:alnum:]_-]+ <arg>`: E.g. `--bar boo`, `--data '{"key": "value"}'`

## Design
The fundemental design of the library is fairly straight forward and
relativelty simple. The entire parser is designed to be orthogonal in how it
handles short-options, short-only options, long-options, long-only options,
_and_ sub-commands. This means that there is no special _exception_ for how an
option of _any type_ is handled when compared to how a sub-command is handled.

The core system comprises of:
1. The CLI parser: This component _only_ handles parsing the CLI into a _parse
   list_ and shifting "non-optional" arguments to the end of the _parse list_.
   The parser also _normalizes_ the CLI. I.e. It `expands` compacted short
   options, as well as `splitting` long-options which are delimited by an equal
   sign. During `expansion` of compacted short-options, should an optional
   argument be found, the parser performs a look-behind to find the last
   short-option which accepts an argument.

2. Iterating the _parse list_ and executing any function which is _subscribed_
   to a particular option, or non-option.

The subscription system supports adding a `func` to be called whenever a
matching `non-optional` argument is encountered. This function maintains a map
of `non-optional` names to `func` pairings which can be extended using the
`AddCmd()` method. The outer parser will stop processing the CLI list whenever
a sub-command is encountered. Instead, a sub-command parser is passed a copy of
the parent parser which is inspected whenever a optional argument is
encountered which is unknown to the sub-command. Calling the parent parser is
done _per_ option and not for the remaining argument list. This way all
_options_ are always global, by default, to all child commands in the command
tree.

## Features
There are a number of interesting features that manifest from the design
approach that `OptArgs` has taken as the design makes it relatively straight
forward to support arbitrary use cases.

### Non-Argument Options

### Short-Only Options

### Long-Only Options

### Count Options

### Boolean Toggle Options

### Boolean w/ Inversion Options

### N-Way Toggle Options

### Many-to-One Options

### Complex Data Structure Options

### Destination Variable Options

### Option Inheritance

### Option Overloading

### Sub-Command Directories

### Automatic Help Text

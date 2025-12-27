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
free to come up with whatever CLI they feel fits their user base.

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

Boolean flags that don't require explicit values. When present, they toggle to `true`.

```bash
# Enable verbose mode
./myapp --verbose

# Multiple boolean flags
./myapp --verbose --debug --force
```

[📁 Code Example](docs/examples/non_argument_options.go)

### Short-Only Options

Single-character flags that provide convenient shortcuts for frequently used options.

```bash
# Short form
./myapp -o output.txt -v -h localhost

# Mixed short and long
./myapp -v --output=result.txt
```

[📁 Code Example](docs/examples/short_only_options.go)

### Short-Option Compaction

POSIX-style compaction allows multiple short options to be combined into a single argument. When an optional argument is provided, it belongs to the last option in the compacted sequence that accepts an argument.

```bash
# Basic compaction (equivalent to -v -f -x)
./myapp -vfx

# Compaction with argument to last option
./myapp -vfo output.txt        # -v -f -o output.txt

# Compaction with optional argument attached
./myapp -vf123                 # -v -f=123 (if -f accepts optional arg)

# Complex compaction with mixed argument types
./myapp -abc123 input.txt      # -a -b -c=123 input.txt

# Real-world example (tar-style)
tar -xzf archive.tar.gz        # -x -z -f=archive.tar.gz
tar -xzvf archive.tar.gz       # -x -z -v -f=archive.tar.gz

# With optional arguments
./myapp -vvf5 --output=result  # -v -v -f=5 --output=result
```

[📁 Code Example](docs/examples/short_option_compaction.go)

### Long-Only Options

Descriptive multi-character flags for clarity and self-documentation.

```bash
# Long descriptive flags
./myapp --config-file=/etc/myapp/config.json
./myapp --database-url=postgres://localhost/mydb
./myapp --max-connections=100

# With equals syntax
./myapp --output-format=json --log-level=debug
```

### Count Options

Flags that can be repeated to increase a counter value, useful for verbosity levels.

```bash
# Increase verbosity level
./myapp -v                    # verbosity = 1
./myapp -vv                   # verbosity = 2  
./myapp -vvv                  # verbosity = 3

# Long form repetition
./myapp --verbose --verbose --verbose

# Mixed usage
./myapp -v --verbose -v       # verbosity = 3
```

[📁 Code Example](docs/examples/count_options.go)

### Boolean Toggle Options

Boolean flags with enhanced syntax supporting explicit true/false values and negation.

```bash
# Simple boolean (sets to true)
./myapp --debug

# Explicit values
./myapp --debug=true
./myapp --debug=false

# Various boolean formats
./myapp --enabled=1           # true
./myapp --enabled=0           # false
./myapp --enabled=yes         # true
./myapp --enabled=no          # false
```

### Boolean w/ Inversion Options

Boolean flags that support negation syntax for intuitive toggling.

```bash
# Enable colors (default behavior)
./myapp --colors

# Disable colors using negation
./myapp --no-colors

# Other negation examples
./myapp --no-cache --no-verify --no-interactive
```

### N-Way Toggle Options

Enumerated options that cycle through multiple states or accept specific values.

```bash
# Set log level
./myapp --log-level=debug
./myapp --log-level=info
./myapp --log-level=warn
./myapp --log-level=error

# Output format selection
./myapp --format=json
./myapp --format=yaml
./myapp --format=table

# Compression level
./myapp --compression=none
./myapp --compression=fast
./myapp --compression=high
```

### Many-to-One Options

Multiple flags that can set the same destination variable, useful for aliases.

```bash
# All of these show help
./myapp --help
./myapp -h
./myapp --usage

# Version information aliases
./myapp --version
./myapp -V
./myapp --ver

# Multiple ways to set the same config
./myapp --config=file.json
./myapp --configuration=file.json
./myapp -c file.json
```

### Complex Data Structure Options

Flags that parse complex data structures like slices, maps, or custom types.

```bash
# String slices (comma-separated)
./myapp --tag=web,api,database

# String slices (repeated flags)
./myapp --tag=web --tag=api --tag=database

# Integer slices
./myapp --port=8080,8081,8082
./myapp --port=8080 --port=8081 --port=8082

# Key-value pairs (environment variables)
./myapp --env=DEBUG=true --env=PORT=8080 --env=HOST=localhost

# Mixed complex structures
./myapp --tag=web,api --port=8080,8081 --env=NODE_ENV=production
```

[📁 Code Example](docs/examples/complex_data_structures.go)

### Destination Variable Options

Direct binding of flag values to variables for automatic population.

```bash
# All flags automatically populate their respective variables
./myapp --host=api.example.com --port=443 --timeout=30s --verbose --retries=5

# Mixed short and long forms
./myapp -h api.example.com -p 443 --timeout=1m -v --retries=3

# Using equals syntax
./myapp --host=localhost --port=8080 --timeout=10s
```

### Option Inheritance

Flags that inherit values from parent contexts or configuration files.

```bash
# Global flags affect all subcommands
./myapp --verbose server --port=8080
./myapp --config=/etc/myapp.conf client --url=http://localhost

# Subcommand-specific flags
./myapp server --port=8080 --host=0.0.0.0
./myapp client --url=http://api.example.com --timeout=30s
```

### Option Overloading

Multiple definitions of the same flag name with different behaviors based on context.

```bash
# Config as file path
./myapp --config=/path/to/config.json

# Config as inline JSON
./myapp --config='{"database": {"host": "localhost"}}'

# Context determines interpretation
./myapp --config=production.yml     # file
./myapp --config='key: value'       # inline YAML
```

### Sub-Command Directories

Organized command structures with hierarchical flag inheritance.

```bash
# Global flags before subcommand
./myapp --verbose --config=/etc/myapp.conf server --port=8080

# Subcommand with its own flags
./myapp server --port=8080 --host=0.0.0.0 --workers=4

# Client subcommand
./myapp client --url=http://localhost:8080 --timeout=30s

# Nested subcommands
./myapp database migrate --dry-run --verbose
./myapp database backup --output=/backups/db.sql
```

### Automatic Help Text

Built-in help generation with customizable formatting and usage information.

```bash
# Show help
./myapp --help
./myapp -h

# Subcommand help
./myapp server --help
./myapp client --help

# Example output:
# MyApp - A sample application
#
# Usage: myapp [OPTIONS] COMMAND
#
# Options:
#   -h, --host hostname     Server hostname (default "localhost")
#   -p, --port port         Server port number (default 8080)
#       --timeout duration  Connection timeout (default 30s)
#   -v, --verbose          Enable verbose output
#
# Commands:
#   server    Start the server
#   client    Run as client
#
# Examples:
#   myapp --host=api.example.com server --port=443
#   myapp client --url=http://localhost:8080
```

### Advanced GNU/POSIX Features

OptArgs supports sophisticated GNU getopt_long() features including special characters in option names and longest matching patterns.

```bash
# Special characters in option names
./myapp --system7:verbose=detailed
./myapp --config=env production
./myapp --db:host=primary=db1.example.com
./myapp --app:level=debug=trace

# Longest matching (enable-bobadufoo wins over enable-bob)
./myapp --enable-bobadufoo=advanced

# Complex nested syntax
./myapp --system7:path=bindir=/usr/local/bin
./myapp --cache:url=redis=redis://localhost:6379

# Multiple special characters
./myapp --app:config=env:prod=live-settings
```

[📁 Code Example](docs/examples/advanced_gnu_features.go)

### Real-World Usage Examples

```bash
# Web server with comprehensive configuration
./myapp server \
  --host=0.0.0.0 \
  --port=8080 \
  --workers=4 \
  --timeout=30s \
  --log-level=info \
  --tag=web,api,production \
  --env=NODE_ENV=production \
  --env=DEBUG=false \
  --no-cache

# Database operations
./myapp db migrate \
  --config=/etc/myapp/db.conf \
  --dry-run \
  --verbose \
  --timeout=5m

# Batch processing with multiple inputs
./myapp process \
  --input=/data/file1.json \
  --input=/data/file2.json \
  --output-format=csv \
  --workers=8 \
  --tag=batch,processing \
  --env=MEMORY_LIMIT=4GB

# Development mode with debugging
./myapp dev \
  --verbose --verbose --verbose \
  --debug \
  --hot-reload \
  --port=3000 \
  --env=NODE_ENV=development \
  --no-minify
```

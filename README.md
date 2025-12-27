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

Boolean flags that don't require explicit values. When present, they toggle to `true`.

```go
package main

import (
    "fmt"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    var verbose bool
    fs.BoolVar(&verbose, "verbose", false, "Enable verbose output")
    
    // Usage: ./app --verbose
    fs.Parse([]string{"--verbose"})
    fmt.Printf("Verbose mode: %t\n", verbose) // Output: Verbose mode: true
}
```

### Short-Only Options

Single-character flags that provide convenient shortcuts for frequently used options.

```go
package main

import (
    "fmt"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    var output string
    fs.StringVarP(&output, "output", "o", "stdout", "Output destination")
    
    // Usage: ./app -o file.txt
    fs.Parse([]string{"-o", "file.txt"})
    fmt.Printf("Output: %s\n", output) // Output: Output: file.txt
}
```

### Long-Only Options

Descriptive multi-character flags for clarity and self-documentation.

```go
package main

import (
    "fmt"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    var configFile string
    fs.StringVar(&configFile, "config-file", "config.json", "Configuration file path")
    
    // Usage: ./app --config-file=/etc/myapp/config.json
    fs.Parse([]string{"--config-file=/etc/myapp/config.json"})
    fmt.Printf("Config: %s\n", configFile) // Output: Config: /etc/myapp/config.json
}
```

### Count Options

Flags that can be repeated to increase a counter value, useful for verbosity levels.

```go
package main

import (
    "fmt"
    "strings"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    var verbosity int
    // Custom counter implementation using Var
    fs.Var(&CountValue{&verbosity}, "verbose", "Increase verbosity (can be repeated)")
    
    // Usage: ./app -v -v -v  (or --verbose --verbose --verbose)
    fs.Parse([]string{"--verbose", "--verbose", "--verbose"})
    fmt.Printf("Verbosity level: %d\n", verbosity) // Output: Verbosity level: 3
}

// CountValue implements Value interface for counting
type CountValue struct {
    count *int
}

func (c *CountValue) String() string { return fmt.Sprintf("%d", *c.count) }
func (c *CountValue) Set(string) error { *c.count++; return nil }
func (c *CountValue) Type() string { return "count" }
```

### Boolean Toggle Options

Boolean flags with enhanced syntax supporting explicit true/false values and negation.

```go
package main

import (
    "fmt"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    var debug bool
    fs.BoolVar(&debug, "debug", false, "Enable debug mode")
    
    // Multiple ways to use boolean flags:
    // ./app --debug              (sets to true)
    // ./app --debug=true         (explicit true)
    // ./app --debug=false        (explicit false)
    // ./app --no-debug           (negation syntax, sets to false)
    
    fs.Parse([]string{"--debug=true"})
    fmt.Printf("Debug mode: %t\n", debug) // Output: Debug mode: true
}
```

### Boolean w/ Inversion Options

Boolean flags that support negation syntax for intuitive toggling.

```go
package main

import (
    "fmt"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    var colors bool
    fs.BoolVar(&colors, "colors", true, "Enable colored output")
    
    // Usage examples:
    // ./app --colors             (enables colors)
    // ./app --no-colors          (disables colors)
    
    fs.Parse([]string{"--no-colors"})
    fmt.Printf("Colors enabled: %t\n", colors) // Output: Colors enabled: false
}
```

### N-Way Toggle Options

Enumerated options that cycle through multiple states or accept specific values.

```go
package main

import (
    "fmt"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    var logLevel string
    // Custom enum implementation
    fs.Var(&EnumValue{
        value: &logLevel,
        allowed: []string{"debug", "info", "warn", "error"},
        defaultVal: "info",
    }, "log-level", "Set logging level (debug|info|warn|error)")
    
    // Usage: ./app --log-level=debug
    fs.Parse([]string{"--log-level=debug"})
    fmt.Printf("Log level: %s\n", logLevel) // Output: Log level: debug
}

// EnumValue implements Value interface for enumerated options
type EnumValue struct {
    value      *string
    allowed    []string
    defaultVal string
}

func (e *EnumValue) String() string { 
    if *e.value == "" { return e.defaultVal }
    return *e.value 
}

func (e *EnumValue) Set(val string) error {
    for _, allowed := range e.allowed {
        if val == allowed {
            *e.value = val
            return nil
        }
    }
    return fmt.Errorf("invalid value %q, must be one of: %v", val, e.allowed)
}

func (e *EnumValue) Type() string { return "enum" }
```

### Many-to-One Options

Multiple flags that can set the same destination variable, useful for aliases.

```go
package main

import (
    "fmt"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    var helpRequested bool
    
    // Multiple flags setting the same variable
    fs.BoolVar(&helpRequested, "help", false, "Show help message")
    fs.BoolVar(&helpRequested, "h", false, "Show help message (short)")
    fs.BoolVar(&helpRequested, "usage", false, "Show help message (alias)")
    
    // Any of these work: --help, --h, --usage
    fs.Parse([]string{"--usage"})
    fmt.Printf("Help requested: %t\n", helpRequested) // Output: Help requested: true
}
```

### Complex Data Structure Options

Flags that parse complex data structures like slices, maps, or custom types.

```go
package main

import (
    "fmt"
    "strings"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    // Built-in slice support
    var tags []string
    var ports []int
    
    fs.StringSliceVar(&tags, "tag", []string{}, "Add tags (can be repeated or comma-separated)")
    fs.IntSliceVar(&ports, "port", []int{}, "Add ports (can be repeated or comma-separated)")
    
    // Usage examples:
    // ./app --tag=web --tag=api --port=8080,8081,8082
    // ./app --tag=web,api,database --port=8080 --port=8081
    
    fs.Parse([]string{"--tag=web,api", "--port=8080", "--port=8081"})
    fmt.Printf("Tags: %v\n", tags)   // Output: Tags: [web api]
    fmt.Printf("Ports: %v\n", ports) // Output: Ports: [8080 8081]
    
    // Custom map implementation
    var env map[string]string
    fs.Var(&MapValue{&env}, "env", "Set environment variables (key=value)")
    
    fs.Parse([]string{"--env=DEBUG=true", "--env=PORT=8080"})
    fmt.Printf("Environment: %v\n", env) // Output: Environment: map[DEBUG:true PORT:8080]
}

// MapValue implements Value interface for key=value pairs
type MapValue struct {
    m *map[string]string
}

func (mv *MapValue) String() string {
    if *mv.m == nil { return "{}" }
    return fmt.Sprintf("%v", *mv.m)
}

func (mv *MapValue) Set(val string) error {
    if *mv.m == nil {
        *mv.m = make(map[string]string)
    }
    parts := strings.SplitN(val, "=", 2)
    if len(parts) != 2 {
        return fmt.Errorf("invalid format, expected key=value")
    }
    (*mv.m)[parts[0]] = parts[1]
    return nil
}

func (mv *MapValue) Type() string { return "map" }
```

### Destination Variable Options

Direct binding of flag values to variables for automatic population.

```go
package main

import (
    "fmt"
    "time"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    // Direct variable binding - no need to check flag values manually
    var (
        host     string
        port     int
        timeout  time.Duration
        verbose  bool
        retries  int
    )
    
    fs.StringVar(&host, "host", "localhost", "Server host")
    fs.IntVar(&port, "port", 8080, "Server port")
    fs.DurationVar(&timeout, "timeout", 30*time.Second, "Request timeout")
    fs.BoolVar(&verbose, "verbose", false, "Verbose output")
    fs.IntVar(&retries, "retries", 3, "Number of retries")
    
    // After parsing, variables are automatically populated
    fs.Parse([]string{"--host=api.example.com", "--port=443", "--timeout=1m", "--verbose"})
    
    // Variables are ready to use immediately
    fmt.Printf("Connecting to %s:%d\n", host, port)
    fmt.Printf("Timeout: %v, Verbose: %t, Retries: %d\n", timeout, verbose, retries)
    // Output: 
    // Connecting to api.example.com:443
    // Timeout: 1m0s, Verbose: true, Retries: 3
}
```

### Option Inheritance

Flags that inherit values from parent contexts or configuration files.

```go
package main

import (
    "fmt"
    "github.com/major0/optargs/pflags"
)

func main() {
    // Parent flag set with global options
    globalFlags := pflags.NewFlagSet("global", pflags.ContinueOnError)
    var globalVerbose bool
    globalFlags.BoolVar(&globalVerbose, "verbose", false, "Global verbose mode")
    
    // Child flag set that can inherit from parent
    cmdFlags := pflags.NewFlagSet("command", pflags.ContinueOnError)
    var cmdSpecific string
    cmdFlags.StringVar(&cmdSpecific, "output", "", "Command-specific output")
    
    // Parse global flags first
    globalFlags.Parse([]string{"--verbose"})
    
    // Child can access parent's parsed values
    if globalVerbose {
        fmt.Println("Verbose mode enabled globally")
    }
    
    // Parse command-specific flags
    cmdFlags.Parse([]string{"--output=result.txt"})
    fmt.Printf("Command output: %s\n", cmdSpecific)
}
```

### Option Overloading

Multiple definitions of the same flag name with different behaviors based on context.

```go
package main

import (
    "fmt"
    "strings"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("example", pflags.ContinueOnError)
    
    // Same flag name with different types/behaviors in different contexts
    var configFile string
    var configInline string
    
    // Context-sensitive flag handling
    fs.Var(&OverloadedValue{
        fileVar:   &configFile,
        inlineVar: &configInline,
    }, "config", "Configuration (file path or inline JSON)")
    
    // Usage examples:
    // ./app --config=/path/to/config.json        (treated as file path)
    // ./app --config='{"key": "value"}'          (treated as inline JSON)
    
    fs.Parse([]string{"--config=/etc/app/config.json"})
    fmt.Printf("Config file: %s\n", configFile)
}

// OverloadedValue demonstrates context-sensitive flag handling
type OverloadedValue struct {
    fileVar   *string
    inlineVar *string
}

func (ov *OverloadedValue) String() string {
    if *ov.fileVar != "" { return *ov.fileVar }
    return *ov.inlineVar
}

func (ov *OverloadedValue) Set(val string) error {
    // Heuristic: if it starts with { or [, treat as inline JSON
    if strings.HasPrefix(val, "{") || strings.HasPrefix(val, "[") {
        *ov.inlineVar = val
    } else {
        *ov.fileVar = val
    }
    return nil
}

func (ov *OverloadedValue) Type() string { return "config" }
```

### Sub-Command Directories

Organized command structures with hierarchical flag inheritance.

```go
package main

import (
    "fmt"
    "os"
    "github.com/major0/optargs/pflags"
)

func main() {
    // Root command flags
    rootFlags := pflags.NewFlagSet("myapp", pflags.ContinueOnError)
    var globalVerbose bool
    rootFlags.BoolVar(&globalVerbose, "verbose", false, "Global verbose mode")
    
    // Sub-command: server
    serverFlags := pflags.NewFlagSet("server", pflags.ContinueOnError)
    var serverPort int
    var serverHost string
    serverFlags.IntVar(&serverPort, "port", 8080, "Server port")
    serverFlags.StringVar(&serverHost, "host", "localhost", "Server host")
    
    // Sub-command: client
    clientFlags := pflags.NewFlagSet("client", pflags.ContinueOnError)
    var clientURL string
    var clientTimeout string
    clientFlags.StringVar(&clientURL, "url", "", "Server URL")
    clientFlags.StringVar(&clientTimeout, "timeout", "30s", "Request timeout")
    
    // Parse command line
    args := os.Args[1:]
    if len(args) == 0 {
        fmt.Println("Usage: myapp [global-flags] <command> [command-flags]")
        return
    }
    
    // Parse global flags until we hit a sub-command
    var globalArgs []string
    var subCommand string
    var subArgs []string
    
    for i, arg := range args {
        if !strings.HasPrefix(arg, "-") {
            subCommand = arg
            globalArgs = args[:i]
            subArgs = args[i+1:]
            break
        }
    }
    
    // Parse global flags
    rootFlags.Parse(globalArgs)
    
    // Handle sub-commands
    switch subCommand {
    case "server":
        serverFlags.Parse(subArgs)
        fmt.Printf("Starting server on %s:%d (verbose: %t)\n", 
            serverHost, serverPort, globalVerbose)
        
    case "client":
        clientFlags.Parse(subArgs)
        fmt.Printf("Connecting to %s with timeout %s (verbose: %t)\n", 
            clientURL, clientTimeout, globalVerbose)
        
    default:
        fmt.Printf("Unknown command: %s\n", subCommand)
    }
}
```

### Automatic Help Text

Built-in help generation with customizable formatting and usage information.

```go
package main

import (
    "fmt"
    "time"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("myapp", pflags.ContinueOnError)
    
    // Define flags with descriptive usage text
    var (
        host     = fs.StringP("host", "h", "localhost", "Server `hostname` to connect to")
        port     = fs.IntP("port", "p", 8080, "Server `port` number")
        timeout  = fs.Duration("timeout", 30*time.Second, "Connection `timeout` duration")
        verbose  = fs.BoolP("verbose", "v", false, "Enable verbose output")
        config   = fs.String("config", "", "Configuration `file` path")
        tags     = fs.StringSlice("tag", []string{}, "Add `tags` (repeatable)")
    )
    
    // Custom usage function
    fs.Usage = func() {
        fmt.Fprintf(fs.Output(), "MyApp - A sample application\n\n")
        fmt.Fprintf(fs.Output(), "Usage: %s [OPTIONS]\n\n", fs.Name())
        fmt.Fprintf(fs.Output(), "Options:\n")
        fs.PrintDefaults()
        fmt.Fprintf(fs.Output(), "\nExamples:\n")
        fmt.Fprintf(fs.Output(), "  %s --host=api.example.com --port=443 --verbose\n", fs.Name())
        fmt.Fprintf(fs.Output(), "  %s -h localhost -p 8080 --tag=web --tag=api\n", fs.Name())
    }
    
    // Parse arguments
    err := fs.Parse([]string{"--help"})
    if err != nil {
        return
    }
    
    // Use the parsed values
    fmt.Printf("Connecting to %s:%d\n", *host, *port)
    fmt.Printf("Timeout: %v, Verbose: %t\n", *timeout, *verbose)
    fmt.Printf("Config: %s, Tags: %v\n", *config, *tags)
}

// Output when --help is used:
// MyApp - A sample application
//
// Usage: myapp [OPTIONS]
//
// Options:
//   -h, --host hostname     Server hostname to connect to (default "localhost")
//   -p, --port port         Server port number (default 8080)
//       --timeout timeout   Connection timeout duration (default 30s)
//   -v, --verbose           Enable verbose output
//       --config file       Configuration file path
//       --tag tags          Add tags (repeatable)
//
// Examples:
//   myapp --host=api.example.com --port=443 --verbose
//   myapp -h localhost -p 8080 --tag=web --tag=api
```

### Advanced GNU/POSIX Features

OptArgs supports sophisticated GNU getopt_long() features including special characters in option names and longest matching patterns.

```go
package main

import (
    "fmt"
    "github.com/major0/optargs/pflags"
)

func main() {
    fs := pflags.NewFlagSet("advanced", pflags.ContinueOnError)
    
    // Special characters in option names (colons, equals)
    var (
        systemVerbose = fs.String("system7:verbose", "", "System 7 verbose mode")
        configEnv     = fs.String("config=env", "", "Configuration environment")
        dbHost        = fs.String("db:host=primary", "", "Database primary host")
        appLevel      = fs.String("app:level=debug", "", "Application debug level")
    )
    
    // Longest matching - multiple options with shared prefixes
    var (
        enableBob       = fs.String("enable-bob", "", "Enable bob feature")
        enableBobadufoo = fs.String("enable-bobadufoo", "", "Enable bobadufoo feature")
    )
    
    // Complex nested syntax examples:
    args := []string{
        "--system7:verbose=detailed",
        "--config=env", "production", 
        "--db:host=primary=db1.example.com",
        "--app:level=debug=trace",
        "--enable-bobadufoo", "advanced",  // Longest match wins
    }
    
    fs.Parse(args)
    
    fmt.Printf("System verbose: %s\n", *systemVerbose)
    fmt.Printf("Config env: %s\n", *configEnv)
    fmt.Printf("DB host: %s\n", *dbHost)
    fmt.Printf("App level: %s\n", *appLevel)
    fmt.Printf("Enable bob: %s\n", *enableBob)           // Empty - not matched
    fmt.Printf("Enable bobadufoo: %s\n", *enableBobadufoo) // "advanced" - longest match
}
```

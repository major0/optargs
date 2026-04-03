# Subcommand Parent Option Inheritance

go-arg subcommands inherit parent arguments but help text does not display
them, confusing users. OptArgs provides full inheritance with opt-out.

## Related upstream issues

| Issue | Status | Summary |
|-------|--------|---------|
| [alexflint/go-arg#101](https://github.com/alexflint/go-arg/issues/101) | Closed | Subcommands don't display parent command arguments |

## Without OptArgs (upstream go-arg)

```console
$ myapp server --help
Usage: myapp server [--port PORT]

Options:
  --port PORT    listen port

$ myapp server --verbose   # works (inherited from parent)
$ # but --verbose is not shown in help text — users don't know it exists
```

Parent options work in subcommands but are invisible in help output,
leading to confusion about what flags are available.

## With OptArgs

```console
$ myapp server --help
Usage: myapp server [--port PORT] [--verbose]

Options:
  --port PORT    listen port
  --verbose      enable verbose output (inherited)

$ myapp server --verbose   # works and is documented
```

Inherited options appear in child help text. When inheritance is
undesirable, strict mode disables it:

```console
$ # With StrictSubcommands enabled:
$ myapp server --verbose
Error: unknown option: verbose
```

## OptArgs implementation

```go
root, _ := optargs.GetOptLong(os.Args[1:], "v", []optargs.Flag{
    {Name: "verbose", HasArg: optargs.NoArgument},
})
serve, _ := optargs.GetOptLong([]string{}, "p:", []optargs.Flag{
    {Name: "port", HasArg: optargs.RequiredArgument},
})
root.AddCmd("serve", serve)
// serve inherits --verbose from root

// To disable inheritance:
root.SetStrictSubcommands(true)
```

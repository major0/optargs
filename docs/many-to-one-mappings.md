# Many-to-One Flag Mappings

Multiple flags writing to the same destination (e.g., `ls --format=across`
and `-x` both setting the format field). Not supported in pflag, where
each flag has its own destination variable.

## Related upstream issues

| Issue | Status | Summary |
|-------|--------|---------|
| [spf13/pflag#139](https://github.com/spf13/pflag/issues/139) | Open | Allow short flags without long form |
| [spf13/pflag#256](https://github.com/spf13/pflag/pull/256) | Closed (unmerged) | Add support for specifying only a shortflag |

Supporting short-only and long-only flags in pflag would have been a
prerequisite for many-to-one mappings — a short flag like `-x` could map
to the same destination as `--format=across` without requiring `-x` to
have its own long name.

## OptArgs implementation

Shared `Flag.Handle` callbacks enable multiple flags to write to the same
Value. This supports the common POSIX pattern where short options are
aliases for long option values.

```go
var format string
formatHandler := func(name, arg string) error {
    switch name {
    case "x":
        format = "across"
    case "format":
        format = arg
    }
    return nil
}

p, _ := optargs.NewParser(config, shortOpts, longOpts, args)
// -x and --format=across both set format to "across"
```

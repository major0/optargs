# Boolean Negation (`--no-<flag>`)

Users expect `--no-verbose` to set a boolean flag to false, following GNU
convention. Upstream pflag and cobra have no built-in support.

## Related upstream issues

| Issue | Status | Summary |
|-------|--------|---------|
| [spf13/pflag#214](https://github.com/spf13/pflag/issues/214) | Open | Support boolean flags with only NoOptDefVal, no `--foo=true` syntax |
| [spf13/cobra#958](https://github.com/spf13/cobra/issues/958) | Closed | Allow booleans to be set false with `no-` prefix |
| [spf13/cobra#1821](https://github.com/spf13/cobra/issues/1821) | Closed | Bool flag defaulting to true cannot be toggled false |

## Without OptArgs (upstream pflag/cobra)

```console
$ myapp --no-verbose
Error: unknown flag: --no-verbose

$ myapp --verbose=false    # works but awkward
$ myapp --verbose false    # BROKEN: "false" consumed as positional arg
```

Users must use `--verbose=false` with an explicit `=`, which is unintuitive
and error-prone. There is no way to negate a boolean flag with a `--no-`
prefix.

## With OptArgs

```text
myapp --verbose          # sets verbose=true
myapp --no-verbose       # sets verbose=false
myapp --no-verbose=true  # sets verbose=false (negation applied)
myapp --no-verbose=false # sets verbose=true  (double negation)
```

All forms work as expected. Scripted composition is safe because explicit
values on negated flags behave predictably.

## OptArgs implementation

```go
fs := pflags.NewFlagSet("app", pflags.ContinueOnError)
fs.BoolVar(&verbose, "verbose", false, "enable verbose output")
pflags.MarkNegatable("verbose")
// --verbose sets true, --no-verbose sets false
```

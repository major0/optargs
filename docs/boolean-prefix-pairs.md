# Boolean Prefix Pairs (`--enable-<flag>` / `--disable-<flag>`)

Named prefix pairs for boolean flags, extending beyond the `--no-` convention
to support patterns like `--enable-color` / `--disable-color`.

## Related upstream issues

| Issue | Status | Summary |
|-------|--------|---------|
| [spf13/pflag#214](https://github.com/spf13/pflag/issues/214) | Open | Boolean flags lack flexible negation patterns |

## Without OptArgs (upstream pflag/cobra)

```text
myapp --enable-color     # Error: unknown flag: --enable-color
myapp --disable-color    # Error: unknown flag: --disable-color
myapp --color=false      # works but doesn't match the enable/disable idiom
```

Users must manually register separate flags and wire them to the same
variable, duplicating logic and cluttering help text.

## With OptArgs

```text
myapp --color                # sets color=true (base flag)
myapp --enable-color         # sets color=true
myapp --disable-color        # sets color=false
myapp --disable-color=false  # sets color=true (double negation)
```

Help text shows all forms:

```
  --color, --enable-color, --disable-color   colorize output (default true)
```

## OptArgs implementation

```go
fs := pflag.NewFlagSet("app", pflag.ContinueOnError)
fs.BoolVar(&color, "color", true, "colorize output")
pflag.MarkBoolPrefix("color", "enable", "disable")
// --enable-color sets true, --disable-color sets false
```

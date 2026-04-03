# BoolArgValuer (NoArgument vs OptionalArgument)

pflag treats all boolean flags as accepting optional `=value` arguments.
This causes Count and BoolFunc flags to incorrectly consume the next
positional argument.

## Related upstream issues

| Issue | Status | Summary |
|-------|--------|---------|
| [spf13/pflag#214](https://github.com/spf13/pflag/issues/214) | Open | Boolean flags with only NoOptDefVal |

## Without OptArgs (upstream pflag)

```console
$ # Count flag (-v -v -v for verbosity level):
$ myapp -v -v -v file.txt
# Expected: verbosity=3, args=[file.txt]
# Actual:   verbosity=3, args=[file.txt] — OK here

$ myapp -v file.txt
# Expected: verbosity=1, args=[file.txt]
# Actual:   verbosity=1, args=[file.txt] — OK for simple cases

$ # But with --verbose=2 style:
$ myapp --verbose 2 file.txt
# Expected: verbosity=1, args=[2, file.txt] (count flag, no arg)
# Actual:   depends on pflag version — may consume "2" as the value
```

pflag's boolean handling assumes all bool-like flags accept optional
arguments, which is incorrect for Count and BoolFunc types.

## With OptArgs

```console
$ myapp -vvv file.txt
# verbosity=3, args=[file.txt]

$ myapp --verbose 2 file.txt
# verbosity=1, args=[2, file.txt]
# Count flag never consumes the next argument

$ myapp --verbose=true
# Error: --verbose does not accept an argument
# Count flags are strictly NoArgument
```

Types declare their argument behavior explicitly. Count and BoolFunc
flags never consume the next positional argument.

## OptArgs implementation

```go
// Types implement BoolValuer to control argument behavior:
type BoolValuer interface {
    IsBoolFlag() bool   // "I am a boolean flag"
    BoolTakesArg() bool // false = NoArgument, true = OptionalArgument
}

// Count flags: -vvv increments, never consumes next arg
// BoolFunc flags: toggled by presence, never consumes next arg
```

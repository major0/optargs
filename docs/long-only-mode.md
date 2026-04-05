# getopt_long_only Mode

From the `getopt_long_only(3)` man page:

> getopt_long_only() is like getopt_long(), but '-' as well as "--" can
> indicate a long option. If an option that starts with '-' (not "--")
> doesn't match a long option, but does match a short option, it is
> parsed as a short option instead.

This means single-dash arguments are tried as long options first. If no
long option matches, the parser falls back to short option parsing. This
allows `-verbose` to match `--verbose` without requiring the double dash,
while still supporting `-v` as a short option when no long option named
"v" is registered.

Upstream pflag does not implement this mode. Single-dash arguments are
always parsed as short options (or compacted short option groups).

## Related upstream issues

No specific upstream issue — pflag has no concept of long-only parsing.

## Without OptArgs (upstream pflag)

```console
$ myapp -verbose
# Parsed as compacted short options: -v -e -r -b -o -s -e
# Each character is treated as a separate short flag
```

pflag always interprets single-dash arguments character by character.
There is no way to accept `-verbose` as a long option.

## With OptArgs

```text
myapp -verbose        # matches long option "verbose"
myapp -verbose=value  # matches "verbose" with arg "value"
myapp -v              # falls back to short option 'v' if no long match
myapp -vx             # falls back to short compaction: -v -x
```

Single-dash arguments are tried as long options first. If no long option
matches and short options are registered, the parser falls back to short
option parsing (including compaction). If no short options are registered
either, the parser returns an error.

## OptArgs implementation

```go
p, _ := optargs.GetOptLongOnly(os.Args[1:], "vx", []optargs.Flag{
    {Name: "verbose", HasArg: optargs.NoArgument},
    {Name: "output",  HasArg: optargs.RequiredArgument},
})
// -verbose       → Name:"verbose"  (long match)
// -output=file   → Name:"output", Arg:"file"
// -v             → Name:"v"        (short fallback)
// -vx            → Name:"v", Name:"x" (short compaction fallback)
```

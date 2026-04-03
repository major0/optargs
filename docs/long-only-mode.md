# getopt_long_only Mode

GNU `getopt_long_only(3)` treats single-dash arguments as long options
first, falling back to short option parsing only on failure. This allows
`-verbose` to match `--verbose` without requiring the double dash.

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
matches, the parser falls back to short option parsing via the optstring.

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

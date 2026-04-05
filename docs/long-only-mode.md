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

## Abbreviation matching in long-only mode

[Abbreviation matching](prefix-matching.md) works in long-only mode
too. Single-dash arguments are matched against long options using the
same two-phase algorithm (exact match, then prefix match):

```text
-verbose        # exact long match → "verbose"
-verb           # abbreviation → "verbose" (if unambiguous)
-enable-f       # abbreviation → "enable-foo" (if unambiguous)
```

### Ambiguous abbreviations do NOT fall back to short options

When an abbreviation matches two or more long options, the parser
returns `AmbiguousOptionError` — it does **not** fall back to short
option parsing. Ambiguity means the input entered long-option territory:

```text
# Registered long: --enable-foo, --enable-bar
# Registered short: -e

-enable         # ambiguous → AmbiguousOptionError (no fallback to -e)
```

Short fallback only happens when **zero** long options match (not on
ambiguity).

### Single-character input prefers short options

When the input is a single `-` followed by exactly one character and
that character matches a registered short option, the parser resolves it
as the short option — even if the character is a prefix of a long option:

```text
# Registered long: --verbose
# Registered short: -v

-v              # short option 'v' (not abbreviation of --verbose)
```

### Short fallback rules

- Short fallback only applies to **single-dash** input (`-foo`), never
  double-dash (`--foo`)
- Short fallback only triggers on **zero** long option matches — not on
  ambiguity or other errors
- When fallback occurs, the input is re-parsed as compacted short
  options (`-abc` → `-a -b -c`)

## OptArgs implementation

```go
p, _ := optargs.GetOptLongOnly(os.Args[1:], "vx", []optargs.Flag{
    {Name: "verbose",    HasArg: optargs.NoArgument},
    {Name: "output",     HasArg: optargs.RequiredArgument},
    {Name: "enable-foo", HasArg: optargs.NoArgument},
    {Name: "enable-bar", HasArg: optargs.NoArgument},
})
// -verbose       → Name:"verbose"  (exact long match)
// -verb          → Name:"verbose"  (abbreviation match)
// -output=file   → Name:"output", Arg:"file"
// -v             → Name:"v"        (short fallback — single char prefers short)
// -vx            → Name:"v", Name:"x" (short compaction fallback)
// -enable-f      → Name:"enable-foo" (unambiguous abbreviation)
// -enable        → AmbiguousOptionError (matches enable-foo and enable-bar)
```

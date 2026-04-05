# Abbreviation Matching and Arbitrary Option Names

OptArgs implements two complementary features for long option resolution:
GNU-style abbreviation matching and support for arbitrary characters
(including `=`) in option names. Both are handled by a unified two-phase
matching algorithm with right-to-left `=` splitting.

## Abbreviation matching

GNU `getopt_long(3)` allows users to abbreviate long options as long as
the abbreviation is unambiguous. OptArgs implements the same behavior:

```text
--verbose       # exact match
--verb          # abbreviation → matches --verbose (if unambiguous)
--enable-f      # abbreviation → matches --enable-foo (if unambiguous)
```

When an abbreviation matches two or more registered options, the parser
returns an `AmbiguousOptionError`:

```text
# Registered: --enable-foo, --enable-bar
--enable        # ambiguous → AmbiguousOptionError
--enable-f      # unambiguous → --enable-foo
--enable-b      # unambiguous → --enable-bar
```

Exact matches always win, even when the name is also a prefix of other
registered options:

```text
# Registered: --enable, --enable-foo, --enable-bar
--enable        # exact match → --enable (not ambiguous)
```

## Arbitrary option names

Any `isgraph()` character is valid in a long option name, including `=`,
`:`, `[`, and `]`. This is possible because the parser does not naively
split on the first `=`. Instead, when no match is found, it splits on
the rightmost `=` and retries — iterating right-to-left until a match
is found or no `=` characters remain.

```text
# Registered: "foo", "foo=bar", "foo=bar=boo" (all RequiredArgument)

--foo=value           # flag "foo", arg "value"
--foo=bar=value       # flag "foo=bar", arg "value"
--foo=bar=boo=value   # flag "foo=bar=boo", arg "value"
--foo=bar             # flag "foo=bar", arg from next argv element

# Other special characters:
--system7:verbose     # flag "system7:verbose"
--config[key]=value   # flag "config[key]", arg "value"
```

## Algorithm

The matching algorithm runs in a loop:

1. **Exact match** — check if the current candidate equals a registered
   option name. If found, return immediately.
2. **Prefix match** — collect all registered names that start with the
   candidate. If exactly one matches, resolve as an abbreviation. If
   two or more match, return `AmbiguousOptionError`.
3. **rsplit** — if zero matches were found, split the original input on
   the next rightmost `=` (right-to-left). The left portion becomes the
   new candidate; the right portion becomes the inline argument. Repeat
   from step 1.
4. **No match** — if no `=` characters remain to split on, the option
   is unknown.

This unified approach handles abbreviations, `=`-delimited arguments,
and `=`-in-names in a single pass.

## Related upstream issues

No specific upstream issue — pflag splits on the first `=`
unconditionally and does not support abbreviation matching or `=` in
option names. The POSIX/GNU C library implementations handle
abbreviation matching correctly; pflag chose a simpler approach.

## Without OptArgs (upstream pflag)

```console
$ myapp --verb
# Error: unknown flag: --verb

$ myapp --foo=bar=value
# flag "foo" gets arg "bar=value" (first-= split)
# No way to register "foo=bar" as a flag name
```

## With OptArgs

```go
p, _ := optargs.GetOptLong(args, "", []optargs.Flag{
    {Name: "verbose",     HasArg: optargs.NoArgument},
    {Name: "foo",         HasArg: optargs.RequiredArgument},
    {Name: "foo=bar",     HasArg: optargs.RequiredArgument},
    {Name: "foo=bar=boo", HasArg: optargs.RequiredArgument},
})
// --verb              → Name:"verbose"      (abbreviation match)
// --foo=value         → Name:"foo",         Arg:"value"
// --foo=bar=value     → Name:"foo=bar",     Arg:"value"
// --foo=bar=boo=value → Name:"foo=bar=boo", Arg:"value"
```

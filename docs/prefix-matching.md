# Longest-Match Option Resolution and Arbitrary Option Names

GNU `getopt_long(3)` and `getopt_long_only(3)` accept any `isgraph()`
character in long option names and resolve overlapping names using
longest-match at `=` boundaries. OptArgs implements this same behavior.
This is not prefix abbreviation (where `--verb` matches `--verbose`) —
it is exact longest-match: when multiple registered names overlap, the
longest one that sits at an `=` boundary wins.

This longest-match logic is what enables option names containing `=`,
colons, brackets, and other characters. Without it, a parser that
splits on the first `=` cannot distinguish `--foo=bar` (flag "foo"
with arg "bar") from `--foo=bar=value` (flag "foo=bar" with arg
"value"). The longest-match algorithm resolves this ambiguity by
checking every registered name at each `=` boundary and selecting the
longest match.

## Related upstream issues

No specific upstream issue — pflag splits on the first `=`
unconditionally and does not support `=` or other special characters
in option names. The POSIX/GNU C library implementations handle this
correctly; pflag chose a simpler approach that sacrifices this
capability.

## Without OptArgs (upstream pflag)

```console
$ # pflag splits on the first '=' unconditionally:
$ myapp --foo=bar=value
# flag "foo" gets arg "bar=value"
# No way to register "foo=bar" as a flag name

$ myapp --config[key]=value
# Error: bad flag syntax: --config[key]=value
```

pflag treats `=` as a name/value separator and rejects characters like
`[`, `]`, and `:` in flag names.

## With OptArgs

```console
$ # Registered flags: "foo", "foo=bar", "foo=bar=boo" (all RequiredArgument)

$ myapp --foo=value
# flag "foo", arg "value"

$ myapp --foo=bar=value
# flag "foo=bar", arg "value" (longest match wins)

$ myapp --foo=bar=boo=value
# flag "foo=bar=boo", arg "value" (longest match wins)

$ myapp --foo=bar
# flag "foo=bar", arg from next argv element (exact name match)

$ # Arbitrary characters in option names:
$ myapp --system7:verbose enabled
# flag "system7:verbose", arg "enabled"

$ myapp --config[key]=value
# flag "config[key]", arg "value"
```

Any `isgraph()` character is valid in a long option name. The
longest-match algorithm handles the `=` boundary resolution that makes
this possible.

## OptArgs implementation

```go
p, _ := optargs.GetOptLong(args, "", []optargs.Flag{
    {Name: "foo",         HasArg: optargs.RequiredArgument},
    {Name: "foo=bar",     HasArg: optargs.RequiredArgument},
    {Name: "foo=bar=boo", HasArg: optargs.RequiredArgument},
})
// --foo=value         → Name:"foo",         Arg:"value"
// --foo=bar=value     → Name:"foo=bar",     Arg:"value"
// --foo=bar=boo=value → Name:"foo=bar=boo", Arg:"value"
```

# Longest-Match Option Resolution

OptArgs resolves long options using longest-match at `=` boundaries. This
is not GNU-style prefix abbreviation (where `--verb` matches `--verbose`).
Instead, when multiple registered option names overlap and contain `=`,
the longest registered name wins.

This enables option names containing `=` and other obscure characters —
a requirement for tools that use structured option names like
`--config[key]=value` or `--foo=bar=boo=value`.

## Related upstream issues

No specific upstream issue — pflag uses exact match only and does not
support `=` in option names.

## Without OptArgs (upstream pflag)

```console
$ # pflag splits on the first '=' unconditionally:
$ myapp --foo=bar=value
# flag "foo" gets arg "bar=value"
# No way to register "foo=bar" as a flag name

$ myapp --config[key]=value
# Error: bad flag syntax: --config[key]=value
```

pflag treats `=` as a name/value separator and does not support `=` or
other special characters in flag names.

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
```

The parser walks all registered names, finds the longest one that matches
at an `=` boundary, and treats the remainder as the argument value.

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

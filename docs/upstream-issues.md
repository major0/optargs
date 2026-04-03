# Upstream Issues Addressed by OptArgs

OptArgs features that address open issues, rejected PRs, or long-standing
feature requests in upstream projects (spf13/pflag, spf13/cobra,
alexflint/go-arg).

## Boolean Negation (`--no-<flag>`)

Users expect `--no-verbose` to set a boolean flag to false, following GNU
convention. Upstream pflag and cobra have no built-in support.

| Issue | Status | Summary |
|-------|--------|---------|
| [spf13/pflag#214](https://github.com/spf13/pflag/issues/214) | Open | Support boolean flags with only NoOptDefVal, no `--foo=true` syntax |
| [spf13/cobra#958](https://github.com/spf13/cobra/issues/958) | Closed | Allow booleans to be set false with `no-` prefix |
| [spf13/cobra#1821](https://github.com/spf13/cobra/issues/1821) | Closed | Bool flag defaulting to true cannot be toggled false |

OptArgs: `MarkNegatable()` registers `--no-<flag>` automatically. Explicit
values are supported: `--no-verbose=true` sets false, `--no-verbose=false`
sets true (double negation for scripted composition).

## Boolean Prefix Pairs (`--enable-<flag>` / `--disable-<flag>`)

Related to negation but for named prefix pairs rather than the `no-` convention.

| Issue | Status | Summary |
|-------|--------|---------|
| [spf13/pflag#214](https://github.com/spf13/pflag/issues/214) | Open | Boolean flags lack flexible negation patterns |

OptArgs: `MarkBoolPrefix("enable", "disable")` registers both prefix forms
with correct help text formatting.

## Short-Only Flags (no long name)

pflag requires every flag to have a long name. POSIX utilities commonly have
short-only options.

| Issue | Status | Summary |
|-------|--------|---------|
| [spf13/pflag#139](https://github.com/spf13/pflag/issues/139) | Open | Allow short flags without long form |
| [spf13/pflag#256](https://github.com/spf13/pflag/pull/256) | Closed (unmerged) | Add support for specifying only a shortflag |

OptArgs: `ShortVar()` API allows flags with only a short name. Help text
formats correctly without a `--` column.

## Subcommand Parent Option Inheritance

go-arg subcommands inherit parent arguments but help text does not display
them, confusing users.

| Issue | Status | Summary |
|-------|--------|---------|
| [alexflint/go-arg#101](https://github.com/alexflint/go-arg/issues/101) | Closed | Subcommands don't display parent command arguments |

OptArgs: Parent options are inherited via the parser chain and displayed in
child help text. `StrictSubcommands` mode disables inheritance when needed.

## POSIX Short-Option Compaction (`-abc`)

pflag supports compaction only for boolean NoArgument flags. OptArgs
implements full POSIX getopt(3) compaction where the last flag in a group
may consume an argument.

No specific upstream issue — this is a fundamental POSIX compliance gap in
pflag's design.

## BoolArgValuer (NoArgument vs OptionalArgument)

pflag treats all boolean flags as accepting optional `=value` arguments.
This causes Count and BoolFunc flags to incorrectly consume the next
positional argument.

| Issue | Status | Summary |
|-------|--------|---------|
| [spf13/pflag#214](https://github.com/spf13/pflag/issues/214) | Open | Boolean flags with only NoOptDefVal |

OptArgs: `BoolTakesArg()` interface lets types declare whether they accept
arguments. Count and BoolFunc are strictly no-argument.

## Many-to-One Flag Mappings

Multiple flags writing to the same destination (e.g., `ls --format=across`
and `-x` both setting the format field). Not supported in pflag.

No specific upstream issue — pflag's architecture ties each flag to its own
destination variable.

OptArgs: Shared `Flag.Handle` callbacks enable multiple flags to write to
the same Value.

## GNU Longest-Match Prefix Matching

OptArgs implements GNU getopt_long(3) prefix matching with ambiguity
detection. pflag does not support prefix matching at all.

No specific upstream issue — pflag uses exact match only.

# pflags — pflag compatibility layer

Drop-in replacement for [spf13/pflag](https://github.com/spf13/pflag)
backed by OptArgs Core's POSIX/GNU getopt implementation.

## Feature Comparison

| Feature | Upstream pflag | pflags/ (compat) |
|---------|:-:|:-:|
| String/Bool/Int/Float/Duration flags | ✅ | ✅ |
| Shorthand flags (-v) | ✅ | ✅ |
| StringSlice/IntSlice | ✅ | ✅ |
| StringArray | ✅ | ✅ |
| StringToString/StringToInt/StringToInt64 | ✅ | ✅ |
| Count flags (-vvv) | ✅ | ✅ |
| Unknown flag errors | ✅ | ✅ |
| `--` termination | ✅ | ✅ |
| FlagSet creation | ✅ | ✅ |
| PrintDefaults/FlagUsages | ✅ | ✅ |
| ContinueOnError/ExitOnError/PanicOnError | ✅ | ✅ |
| Lookup/Set/Changed | ✅ | ✅ |
| NFlag/NArg/Args | ✅ | ✅ |
| VisitAll/Visit | ✅ | ✅ |
| AddFlagSet | ✅ | ✅ |
| Deprecated/ShorthandDeprecated | ✅ | ✅ |
| Hidden flags | ✅ | ✅ |
| SortFlags | ✅ | ✅ |
| SetInterspersed | ✅ | ✅ |
| AddGoFlagSet | ✅ | ✅ |
| IP/IPMask/IPNet | ✅ | ✅ |
| TextVar | ✅ | ✅ |
| Typed getters (GetBool, GetInt, etc.) | ✅ | ✅ |
| [POSIX short-option compaction (-abc)](../docs/short-option-compaction.md) | ✅ | ✅ |
| [Abbreviation matching (--verb → --verbose)](../docs/prefix-matching.md) | ❌ | ✅ |
| [Arbitrary option names (=, :, [] in names)](../docs/prefix-matching.md) | ❌ | ✅ |
| [Boolean negation (--no-flag)](../docs/boolean-negation.md) (spf13/pflag#214, spf13/cobra#958) | ❌ | ✅ |
| [Boolean prefix pairs (--enable/--disable)](../docs/boolean-prefix-pairs.md) | ❌ | ✅ |
| [Short-only flags (no long name)](../docs/short-only-flags.md) (spf13/pflag#139, spf13/pflag#256) | ❌ | ✅ |
| [Many-to-one flag mappings](../docs/many-to-one-mappings.md) | ❌ | ✅ |
| [BoolArgValuer (NoArg vs OptionalArg)](../docs/bool-arg-valuer.md) (spf13/pflag#214) | ❌ | ✅ |
| [getopt_long_only mode](../docs/long-only-mode.md) | ❌ | ✅ |
| Error message format | ✅ | ⚠️¹ |

¹ Inner error uses core's unified format instead of raw strconv errors.
  See `compat/expected_diffs.go` for all documented divergences.
  Every ✅ and ❌ is backed by a test — see `compat/compat_test.go`, `pflags_test.go`, and `pflags_optargs_test.go`.

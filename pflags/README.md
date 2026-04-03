# pflags — pflag compatibility layer

Drop-in replacement for [spf13/pflag](https://github.com/spf13/pflag)
backed by OptArgs Core's POSIX/GNU getopt implementation.

## Feature Comparison

| Feature | Upstream pflag | pflags/ (compat) | OptArgs-exclusive |
|---------|:-:|:-:|:-:|
| String/Bool/Int/Float/Duration flags | ✅ | ✅ | — |
| Shorthand flags (-v) | ✅ | ✅ | — |
| StringSlice/IntSlice | ✅ | ✅ | — |
| StringArray | ✅ | ✅ | — |
| StringToString/StringToInt/StringToInt64 | ✅ | ✅ | — |
| Count flags (-vvv) | ✅ | ✅ | — |
| Unknown flag errors | ✅ | ✅ | — |
| `--` termination | ✅ | ✅ | — |
| FlagSet creation | ✅ | ✅ | — |
| PrintDefaults/FlagUsages | ✅ | ✅ | — |
| ContinueOnError/ExitOnError/PanicOnError | ✅ | ✅ | — |
| Lookup/Set/Changed | ✅ | ✅ | — |
| NFlag/NArg/Args | ✅ | ✅ | — |
| VisitAll/Visit | ✅ | ✅ | — |
| AddFlagSet | ✅ | ✅ | — |
| Deprecated/ShorthandDeprecated | ✅ | ✅ | — |
| Hidden flags | ✅ | ✅ | — |
| SortFlags | ✅ | ✅ | — |
| SetInterspersed | ✅ | ✅ | — |
| AddGoFlagSet | ✅ | ✅ | — |
| IP/IPMask/IPNet | ✅ | ✅ | — |
| TextVar | ✅ | ✅ | — |
| Typed getters (GetBool, GetInt, etc.) | ✅ | ✅ | — |
| POSIX short-option compaction (-abc) | ❌ | — | ✅ |
| GNU longest-match prefix matching | ❌ | — | ✅ |
| Arbitrary long option names (colons, =) | ❌ | — | ✅ |
| Boolean negation (--no-flag) (spf13/pflag#214, spf13/cobra#958) | ❌ | — | ✅ |
| Short-only flags (no long name) (spf13/pflag#139, spf13/pflag#256) | ❌ | — | ✅ |
| Many-to-one flag mappings | ❌ | — | ✅ |
| BoolArgValuer (NoArg vs OptionalArg) (spf13/pflag#214) | ❌ | — | ✅ |
| getopt_long_only mode | ❌ | — | ✅ |
| Error message format | ✅ | ⚠️¹ | — |

¹ Inner error uses core's unified format instead of raw strconv errors.
  See `compat/expected_diffs.go` for all documented divergences.
  Every ✅ and ❌ is backed by a test — see `compat/compat_test.go`, `pflags_test.go`, and `pflags_optargs_test.go`.

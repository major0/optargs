# Short-Only Flags (no long name)

pflag requires every flag to have a long name. POSIX utilities commonly
have short-only options (e.g., `tar -x`, `ls -l`).

## Related upstream issues

| Issue | Status | Summary |
|-------|--------|---------|
| [spf13/pflag#139](https://github.com/spf13/pflag/issues/139) | Open | Allow short flags without long form |
| [spf13/pflag#256](https://github.com/spf13/pflag/pull/256) | Closed (unmerged) | Add support for specifying only a shortflag |

## Without OptArgs (upstream pflag)

```console
$ # Attempting to register a short-only flag:
fs.BoolP("", "n", false, "dry run")

$ myapp -h
Usage of myapp:
  -n, --          dry run     # broken: empty long name renders as "--"

$ # Registering a second short-only flag panics:
fs.BoolP("", "v", false, "verbose")
# panic: flag redefined:
```

pflag requires a long name. Using an empty string produces broken help
text and panics on the second registration.

## With OptArgs

```console
$ myapp -h
Usage of myapp:
  -n              dry run
  -v              verbose output
  -h, --help      display this help and exit

$ myapp -nv       # compaction works with short-only flags
```

Short-only flags render cleanly in help text and participate in POSIX
compaction.

## OptArgs implementation

```go
fs := pflag.NewFlagSet("app", pflag.ContinueOnError)
fs.ShortBoolVar(&dryRun, "n", false, "dry run")
fs.ShortBoolVar(&verbose, "v", false, "verbose output")
// -n and -v work, no --dry-run or --verbose registered
```

# OptArgs Core Examples

Vanilla usage of the core `GetOpt`, `GetOptLong`, and `GetOptLongOnly` APIs.

## Programs

| Directory | API | Description |
|-----------|-----|-------------|
| `getopt/` | `GetOpt` | POSIX getopt(3) short option parsing |
| `getopt_long/` | `GetOptLong` | GNU getopt_long(3) with short and long options |
| `getopt_long_only/` | `GetOptLongOnly` | GNU getopt_long_only(3) single-dash long options with fallback |

## Running

```bash
# Run with demo arguments
go run ./example/getopt
go run ./example/getopt_long
go run ./example/getopt_long_only

# Run with custom arguments
go run ./example/getopt -- -vf myfile.txt -o out.txt
go run ./example/getopt_long -- --verbose --file=data.csv
go run ./example/getopt_long_only -- -verbose -file data.csv -v
```

For higher-level wrapper examples, see [`docs/examples/`](../docs/examples/) (pflags API).

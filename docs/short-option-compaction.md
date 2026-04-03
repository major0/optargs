# POSIX Short-Option Compaction (`-abc`)

pflag supports compaction only for boolean NoArgument flags. OptArgs
implements full POSIX getopt(3) compaction where the last flag in a
group may consume an argument.

## Related upstream issues

No specific upstream issue — this is a fundamental POSIX compliance gap
in pflag's design.

## Without OptArgs (upstream pflag)

```console
$ # Boolean compaction works:
$ myapp -abc              # -a -b -c (all bool) — OK

$ # But compaction with arguments does not:
$ myapp -vofoo            # expected: -v -o foo
                          # actual: error or -v with arg "ofoo"

$ myapp -vo foo           # expected: -v -o foo
                          # actual: -v with arg "o", "foo" is positional
```

pflag only compacts boolean flags. Mixing argument-taking flags into a
compacted group produces incorrect results.

## With OptArgs

```text
myapp -vofoo            # -v (bool), -o with arg "foo"
myapp -vo foo           # -v (bool), -o with arg "foo"
myapp -vf input.txt     # -v (bool), -f with arg "input.txt"
myapp -vfinput.txt      # -v (bool), -f with arg "input.txt"
```

Full POSIX getopt(3) compaction: any flags before the last in a group
must be NoArgument. The last flag may take a required or optional
argument from the remainder of the group or the next argv element.

## OptArgs implementation

```go
p, _ := optargs.GetOpt(os.Args[1:], "vf:o:")
// -vfinput.txt -ooutput.txt
// → -v, -f input.txt, -o output.txt
```

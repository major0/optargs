# POSIX Short-Option Compaction (`-abc`)

POSIX `getopt(3)` allows multiple short options to be combined after a
single dash. Any flags before the last in a group must be NoArgument.
The last flag may take a required or optional argument from the remainder
of the group or the next argv element.

## Related upstream issues

No specific upstream issue — both pflag and OptArgs support compaction,
including the case where the last flag takes an argument.

## Behavior

```text
myapp -abc              # -a -b -c (all NoArgument)
myapp -vofoo            # -v (bool), -o with arg "foo" (attached)
myapp -vo foo           # -v (bool), -o with arg "foo" (next argv)
myapp -vf input.txt     # -v (bool), -f with arg "input.txt"
myapp -vfinput.txt      # -v (bool), -f with arg "input.txt" (attached)
```

Both upstream pflag and OptArgs handle these cases correctly. The
compaction behavior is identical for boolean-only groups and for groups
where the last flag takes an argument.

## OptArgs implementation

```go
p, _ := optargs.GetOpt(os.Args[1:], "vf:o:")
// -vfinput.txt -ooutput.txt
// → -v, -f input.txt, -o output.txt
```

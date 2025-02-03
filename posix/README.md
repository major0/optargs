# GNU/POSIX getopt
POSIX, and GNU/POSIX `getopt()` argument parsing is widely used by
programs all over the world, and much of the usage of the standard libc
interface has sort of solidified into some common use patterns. Many
applications which attempt to _re-implement_ the libc `getopt()`
interfaces in some other language tend to re-implement the common ways
in which applications _use_ these interfaces and not actually re-implement
the patterns allowed by the interface itself.. This oten results in a CLI
handler that makes a lot of usage assumptions and results in a signficant
number of deficiencies/restrictions/bugs.

This directory contains a collection of uses of `getopt(3)`,
`getopt_long(3)`, and `getopt_long_only(3)` which are 100% GNU/POSIX
compliant, though they may seem unexpected or non-intuitive to many users.
In some cases the results may be down right shocking and confusing.
Needless to say, supporting all of these patterns allows for a CLI handler
that is capable of 100% supporting POSIX and GNU patterns, as well as
providing more flexibility to any Flag implementations which leverage the
underlying parser.

After all, while it is possible to have a highly permissive parser and a
restrictive Flags implementation, but it is not possible to have a highly
restrictive parsert and a permissive Flags implmentation.

## Usage
Simply run the top-level `getopt_examples.sh` script to execute all
examples.

## Notes
There may be ways in which `getopt(1)` differs in behavior from
`getopt(3)` for a variety of reasons. In such situations, the optargs
tool will defer to any behavior demonstrated in `getopt(3)`.

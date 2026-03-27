/* POSIX/GNU GetOpt rules:
 *   If an short-option character is:
 *   - is not folowed by a colon, then the option does not require an argument
 *   - followed by a colon, the option requires an argument
 *   - followed by twop colons, then the option takes an optional argument
 *
 *   Long options define the following struct field:
 *   {
 *      name String
 *      hasArg int // 0 = no arg, 1 = required, 2 = optional
 *      int val    // value to return, or set in `flag`
 *      flag any   // nil == return then return `val`, else set `flag` to the value of `val`
 *   }
 *
 */

## TODO: StrictSubcommands (core change)

Add a `StrictSubcommands` flag to the core parser that prevents child parsers
from walking the parent chain to resolve unknown options. When set, options
defined on a parent parser are rejected if they appear after a subcommand name.

This should be automatically enabled when `POSIXLY_CORRECT` is set — POSIX
semantics dictate that options belong to the command they're defined on, with
no inheritance across subcommand boundaries.

Implementation: skip setting `parent` on the child parser in `AddCmd()` (or
add a flag to `ParserConfig` that disables the parent-chain walk in
`findShortOpt`/`findLongOpt`).

Upstream `alexflint/go-arg` exposes this as `Config.StrictSubcommands`. Once
core supports it, goarg should wire `Config.StrictSubcommands` through to the
core flag, and the goarg extension layer should auto-enable it when
`POSIXLY_CORRECT` is detected.

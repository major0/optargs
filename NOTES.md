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

## StrictSubcommands

Implemented in `ParserConfig.strictSubcommands`. When enabled, `AddCmd()` does
not set the child parser's `parent` pointer, preventing the parent-chain walk
for unknown options. Child parsers only resolve their own options.

Automatically enabled when:
- `POSIXLY_CORRECT` environment variable is set
- `+` prefix appears in the optstring

API: `Parser.SetStrictSubcommands(bool)` / `Parser.StrictSubcommands() bool`

goarg exposes this via `Config.StrictSubcommands` (wired through to core).

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


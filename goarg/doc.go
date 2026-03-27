// Package goarg provides API compatibility with alexflint/go-arg
// backed by OptArgs Core's POSIX/GNU getopt implementation.
//
// goarg is a thin translation layer: struct tags are mapped to OptArgs
// Core flags, Handle callbacks write parsed values to struct fields via
// reflection, and all parsing, type conversion, subcommand dispatch, and
// option inheritance are delegated to core.
//
// Extensions are available via the goarg_ext build tag.
package goarg

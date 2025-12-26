// Package optargs provides a collection of CLI parsing utilities in order
// to aid in the development of command line applications.
//
// # POSIX/GNU GetOpt
//
// At the heart of the optargs package is a [Go] implementation of the
// GNU glibc versions the [getopt(3)], [getopt_long(3)], and
// [getopt_long_only(3)] functions.
//
// Leveraging GNU/POSIX conventions as the backend parser option means that
// the parser has a very large degree of flexibility without restricting
// application design choices.
//
// For example, POSIX/GNU allows for the following:
//   - short-only options.
//   - long-only options.
//   - long and short options that do not require a value. I.e. it should
//     be possoble to pass `--foo` and specify that it never takes a
//     value, and any attempt to pass it a value should be ignored or
//     or result in an error.
//   - short-options of any character that is a valid `isgraph()`
//     character; with the exception of `-`, `:` and `;`. This means that
//     the following options are valid: -=, -+, -{, -}, -^, -!, -@, etc.
//   - short-option compaction: `-abc` is the equivilant of `-a -b -c`
//   - short-option compaction with optional args: `-abarg` is the
//     equivilant of `-a -b arg`
//   - arguments to short options that begin with `-`: `-a -1` should pass
//     `-1` as an argument to `-a`
//   - long-arguments that include any `isgraph()` character in the name,
//     this includes allowing `=` in the name of the argument. For
//     example, `--foo=bar=boo` should map `foo=bar` as the Flag, and
//     `boo` as the value to the flag. This potentially also allows for
//     curious long-arg syntax sych as: `--command:arg=value`.
//   - many-to-one flag mappings. For example, the GNU `ls` command supports
//     `--format=<format>` where each possible `<format>` options is also
//     supported by a unique short-option. For example:
//     `--format=across` = `-x`, `--format=commas` = `-m`,
//     `--format=horizontal` = `-x`, `--format=long` = `-l`, etc.
//   - The GNU `-W` flag which allows short-options to behave like an
//     undefined long-option. E.g. `-W foo` should be interpretted as if
//     `--foo` was passed to the application.
//   - long-options that may look similar, but behave differently, from
//     short options. E.g. `-c` and `--c` are allowed to behave
//     differently.
//
// It is always possible to implement a Flag handler which imposes
// opinionated rules atop a non-opinionated parser, but it is not possible
// to write a less opinionated Flag handler atop an opinionated parser.
// To that end, the [optarg] parsers do not make any judgements outside of
// strictly adhearing to the POSIX/GNU conventions. Applications are free
// to implement their own argument handler to best-fit their application's
// needs.
//
// # Flags()
//
// Optargs supports traditional [Go] style flags which act as convenience
// methods around [GetOpt], [GetOptLong], and [GetOptLongOnly] with the
// aim of fully supporting drop-in replacements commonly used CLI tooling,
// such as:
//   - [alexflint/go-args]
//   - [spf13/pflag]
//   - [spf13/cobra]
//
// While these packages are quite useful, they have some fundemental
// limitations and quirks that come from their design choices which aim to
// be overcome by [optargs] and in the case of [spf13/pflag], those quirks
// ultimately percolate up to the user, such as [spf13/pflag]'s boolean
// flags. Or putting arbitrary restrictions on applications, such as
// suporting long-only options, but not allowing short-only options. Or
// not supporting true non-option flags. I.e. many (all?) of the existing
// [Go] flag packages only allow an argument to a flag to be optional or
// required and are not capable of handling flags that never require an
// argument.
//
// [alexflint/go-args]: https://github.com/alexflint/go-args
// [getopt(3)]: https://pubs.opengroup.org/onlinepubs/9699919799/functions/getopt.html
// [getopt_long(3)]: https://pubs.opengroup.org/onlinepubs/9699919799/functions/getopt_long.html
// [getopt_long_only(3)]: https://pubs.opengroup.org/onlinepubs/9699919799/functions/getopt_long_only.html
// [Go]: https://golang.org/
// [spf13/cobra]: https://github.com/spf13/cobra
// [spf13/pflag]: https://github.com/spf13/pflag
package optargs

import (
	"errors"
	"log/slog"
	"os"
)

type ArgType int

const (
	NoArgument ArgType = iota
	RequiredArgument
	OptionalArgument
)

type Flag struct {
	Name   string
	HasArg ArgType
}

type Option struct {
	Name   string
	HasArg bool
	Arg    string
}

func GetOpt(args []string, optstring string) (*Parser, error) {
	return getOpt(args, optstring, nil, false)
}

func GetOptLong(args []string, optstring string, longopts []Flag) (*Parser, error) {
	parser, err := getOpt(args, optstring, longopts, false)
	if err != nil {
		return nil, err
	}

	return parser, nil
}

func GetOptLongOnly(args []string, optstring string, longopts []Flag) (*Parser, error) {
	parser, err := getOpt(args, optstring, longopts, true)
	if err != nil {
		return nil, err
	}

	return parser, nil
}

// Handle parsing the traditional GetOpt/GetOptLong inputs into the parser
// rules and return a new Parser.
func getOpt(args []string, optstring string, longopts []Flag, longOnly bool) (*Parser, error) {
	config := ParserConfig{
		shortCaseIgnore: false,
		longCaseIgnore:  true,
		longOptsOnly:    longOnly,
		enableErrors:    true,
		gnuWords:        false,
		parseMode:       ParseDefault,
	}

	// Check POSIXLY_CORRECT environment variable
	// If set, behave as if '+' prefix was used in optstring
	if os.Getenv("POSIXLY_CORRECT") != "" {
		config.parseMode = ParsePosixlyCorrect
	}

	// Itterate over the longOpts list populating the map
	longOpts := make(map[string]*Flag)
	for _, opt := range longopts {
		longOpts[opt.Name] = &opt
	}

	// we need to inspect the start of the optstring for _behavior_
	// flags. These flags must appear before any option characters.
	// Flags:
	// - `:` Supress automatic error generation, allow the app to
	//       handle it.
	// - `+` Behave as if POSIXLY_CORRECT has been set
	// - `-` Treat all non-option arguments as an argument to a
	//       a short-option with the opt name equal to the binary
	//	 value of `1`, i.e. `true`.
	//       See `getopt_long(3)` for more information.
	shortOpts := make(map[byte]*Flag)
opt_prefix:
	for len(optstring) > 0 {
		slog.Debug("GetOpt", "mode", true, "optstring", optstring)
		switch optstring[0] {
		case ':':
			config.enableErrors = false
		case '+':
			config.parseMode = ParsePosixlyCorrect
		case '-':
			config.parseMode = ParseNonOpts
		default:
			break opt_prefix
		}
		optstring = optstring[1:]
	}

	// Itterate over optstring parsing it according to the libc
	// getopt() spec. Note, the spec fully allows definitions to
	// overwrite previous definitions. The code will not treat this as
	// an error as this allows for the most flexibility.
	for len(optstring) > 0 {
		slog.Debug("GetOpt", "optstring", optstring, "len", len(optstring))

		if config.longOptsOnly {
			return nil, errors.New("non-empty option string found when long-only parsing was enabled")
		}

		c := optstring[0]
		optstring = optstring[1:]
		if !isGraph(c) {
			return nil, errors.New("Invalid option character: " + string(c))
		}

		slog.Debug("GetOpt", "c", string(c), "optstring", optstring, "len", len(optstring))
		switch c {
		case ':', '-', ';': // Dissallowed by the spec
			return nil, errors.New("Invalid option character: " + string(c))
		}

		// look ahead to see if c is followed by ":" or "::"
		var hasArg ArgType
		switch {
		case len(optstring) > 1 && optstring[0] == ':' && optstring[1] == ':':
			slog.Debug("GetOpt", "c", string(c), "hasArg", "optional")
			hasArg = OptionalArgument
			optstring = optstring[2:]
		case len(optstring) > 0 && optstring[0] == ':':
			slog.Debug("GetOpt", "c", string(c), "hasArg", "required")
			hasArg = RequiredArgument
			optstring = optstring[1:]
		case c == 'W' && len(optstring) > 0 && optstring[0] == ';':
			slog.Debug("GetOpt", "c", c, "gnuWords", true)
			config.gnuWords = true
			hasArg = RequiredArgument
			optstring = optstring[1:]
		default:
			slog.Debug("GetOpt", "c", string(c), "hasArg", "none")
			hasArg = NoArgument
		}

		shortOpts[c] = &Flag{
			Name:   string(c),
			HasArg: hasArg,
		}
	}

	return NewParser(config, shortOpts, longOpts, args)
}

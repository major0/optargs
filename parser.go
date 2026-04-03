package optargs

import (
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"strings"
	"unicode"
)

// ParseMode controls how non-option arguments are handled during parsing.
type ParseMode int

const (
	// ParseDefault permutes arguments so that non-options are moved to the end.
	ParseDefault ParseMode = iota
	// ParseNonOpts treats each non-option argument as an argument to a
	// synthetic option with character code 1.
	ParseNonOpts
	// ParsePosixlyCorrect stops option processing at the first non-option argument.
	ParsePosixlyCorrect
)

// ParserConfig holds configuration for a Parser instance.
// All fields are unexported; configuration is set via optstring prefix
// flags and constructor parameters, or via setter methods.
type ParserConfig struct {
	enableErrors bool
	parseMode    ParseMode

	shortCaseIgnore bool
	gnuWords        bool

	longCaseIgnore bool
	longOptsOnly   bool

	// Command case sensitivity
	commandCaseIgnore bool

	// strictSubcommands prevents child parsers from inheriting parent
	// options. When true, AddCmd does not set the parent pointer, so
	// unknown options in a subcommand are not resolved by walking the
	// parent chain. Automatically enabled when POSIXLY_CORRECT is set.
	strictSubcommands bool
}

// SetLongOnly enables or disables getopt_long_only(3) behavior.
// When enabled, single-dash arguments (e.g., -verbose) are first tried
// as long options; on failure, the parser falls back to short option parsing.
func (c *ParserConfig) SetLongOnly(enabled bool) {
	c.longOptsOnly = enabled
}

// LongOnly returns whether getopt_long_only(3) mode is enabled.
func (c *ParserConfig) LongOnly() bool {
	return c.longOptsOnly
}

// SetInterspersed controls whether non-option arguments can appear between
// options. When false, option processing stops at the first non-option
// argument (POSIX behavior). Default is true (GNU behavior).
func (c *ParserConfig) SetInterspersed(interspersed bool) {
	if interspersed {
		c.parseMode = ParseDefault
	} else {
		c.parseMode = ParsePosixlyCorrect
	}
}

// Interspersed returns whether interspersed option/non-option args are allowed.
func (c *ParserConfig) Interspersed() bool {
	return c.parseMode == ParseDefault
}

// Parser is the core argument parser. It processes command-line arguments
// according to POSIX getopt(3) and GNU getopt_long(3) conventions.
//
// Args holds the remaining unprocessed arguments. After iteration completes,
// Args contains non-option arguments (and any arguments after "--").
//
// Commands holds registered subcommands. Use [Parser.AddCmd] to register
// subcommands; do not manipulate Commands directly.
type Parser struct {
	Args      []string
	nonOpts   []string
	shortOpts [256]*Flag // direct-indexed by byte — zero hash overhead
	shortOptN int        // number of registered short options
	longOpts  map[string]*Flag

	// longOptsLower maps strings.ToLower(name) → *Flag for O(1)
	// case-insensitive lookup. Only populated when longCaseIgnore is true.
	longOptsLower map[string]*Flag

	config ParserConfig

	// Command support - simple map of command name to parser
	Commands CommandRegistry
	parent   *Parser

	// Metadata for help generation
	Name        string // command/subcommand name
	Description string // command/subcommand description

	// Active subcommand tracking — set during Options() when command dispatch succeeds
	activeCmd       string  // name of dispatched subcommand
	activeCmdParser *Parser // parser of dispatched subcommand
}

// NewParser creates a Parser from pre-built configuration, short option map,
// long option map, and argument list. Most callers should use [GetOpt],
// [GetOptLong], or [GetOptLongOnly] instead.
//
// Flag structs in shortOpts and longOpts may include a non-nil Handle field
// for per-option handler dispatch. When a flag with a non-nil Handle is
// resolved during parsing, the handler is invoked instead of yielding an
// [Option] through the iterator. This is the construction-time path for
// attaching handlers:
//
//	verbose := &optargs.Flag{Name: "verbose", HasArg: optargs.NoArgument}
//	debug := &optargs.Flag{
//		Name:   "debug",
//		HasArg: optargs.NoArgument,
//		Handle: func(name, arg string) error {
//			log.Println("debug mode enabled")
//			return nil
//		},
//	}
//	p, err := optargs.NewParser(config,
//		map[byte]*optargs.Flag{'v': verbose, 'd': debug},
//		map[string]*optargs.Flag{"verbose": verbose, "debug": debug},
//		os.Args[1:],
//	)
//
// For parsers created via [GetOpt], [GetOptLong], or [GetOptLongOnly],
// handlers can be attached after construction using [Parser.SetHandler],
// [Parser.SetShortHandler], or [Parser.SetLongHandler]. The two paths are
// complementary: NewParser for construction-time setup, SetHandler variants
// for post-construction attachment.
func NewParser(config ParserConfig, shortOpts map[byte]*Flag, longOpts map[string]*Flag, args []string) (*Parser, error) {
	parser := Parser{
		Args:    args,
		nonOpts: make([]string, 0, 8),
		config:  config,
	}

	for c, flag := range shortOpts {
		if !isGraph(c) {
			return nil, parser.optErrorf("invalid short option: %c", c)
		}
		switch c {
		case ':', ';', '-':
			return nil, parser.optErrorf("prohibited short option: %c", c)
		}
		parser.shortOpts[c] = flag
		parser.shortOptN++
	}

	for s := range longOpts {
		for _, r := range s {
			if unicode.IsSpace(r) || !unicode.IsGraphic(r) {
				return nil, parser.optErrorf("invalid long option: %s", s)
			}
		}
	}
	parser.longOpts = longOpts

	// Build lowercased shadow map for O(1) case-insensitive lookup.
	if config.longCaseIgnore && len(longOpts) > 0 {
		parser.longOptsLower = make(map[string]*Flag, len(longOpts))
		for name, flag := range longOpts {
			parser.longOptsLower[strings.ToLower(name)] = flag
		}
	}

	// Initialize command registry
	parser.Commands = NewCommandRegistry()

	return &parser, nil
}

// NewParserWithCaseInsensitiveCommands creates a new parser with case insensitive
// command matching enabled.
func NewParserWithCaseInsensitiveCommands(
	shortOpts map[byte]*Flag, longOpts map[string]*Flag, args []string,
) (*Parser, error) {
	config := ParserConfig{
		commandCaseIgnore: true,
	}
	return NewParser(config, shortOpts, longOpts, args)
}

func (p *Parser) optError(msg string) error {
	if p.config.enableErrors {
		slog.Error(msg)
	}
	return errors.New(msg)
}

func (p *Parser) optErrorf(msg string, args ...any) error {
	return p.optError(fmt.Sprintf(msg, args...))
}

//nolint:gocognit,gocyclo,cyclop // prefix matching with ancestor walk and ambiguity detection is inherently complex
func (p *Parser) findLongOpt(name string, args []string) ([]string, *Flag, Option, error) {
	// Fast path: exact match via direct map lookup (covers 95%+ of real usage).
	// Walk self + ancestors for the exact key.
	for current := p; current != nil; current = current.parent {
		if flag, ok := current.longOpts[name]; ok {
			return p.resolveLongOpt(name, flag, args)
		}
		// Case-insensitive O(1) lookup via the lowercased shadow map.
		if current.longOptsLower != nil {
			if flag, ok := current.longOptsLower[strings.ToLower(name)]; ok {
				return p.resolveLongOpt(flag.Name, flag, args)
			}
		}
	}

	// Slow path: prefix matching. Find the longest matching prefix in a
	// single pass — no slice allocation, no sort. Option names may contain
	// '=' (e.g., "foo=bar"), so we cannot simply split on the first '='.
	// Instead, for each registered option that is a prefix of the input,
	// we check that the character immediately after the prefix is '=' (the
	// name/value separator) or that the lengths match exactly.
	var bestName string
	var bestFlag *Flag
	ambiguous := false

	for current := p; current != nil; current = current.parent {
		for opt, flag := range current.longOpts {
			if len(opt) > len(name) {
				continue
			}
			if !hasPrefix(name, opt, current.config.longCaseIgnore) {
				continue
			}
			// Candidate must sit at a valid boundary: either exact
			// length match, or the next character is '='.
			if len(opt) < len(name) && name[len(opt)] != '=' {
				continue
			}
			if len(opt) > len(bestName) {
				bestName = opt
				bestFlag = flag
				ambiguous = false
			} else if len(opt) == len(bestName) && opt != bestName {
				ambiguous = true
			}
		}
	}

	if bestFlag == nil {
		return args, nil, Option{}, p.optError("unknown option: " + name)
	}
	if ambiguous {
		return args, nil, Option{}, p.optError("ambiguous option: " + name)
	}

	if len(bestName) == len(name) {
		return p.resolveLongOpt(bestName, bestFlag, args)
	}

	// Prefix match with '=' boundary.
	if bestFlag.HasArg == NoArgument { //nolint:nestif // NoArgument fallback search walks ancestors with multiple filter conditions
		// NoArgument option can't accept the '=value' portion.
		// Look for a shorter candidate that can.
		var fallbackName string
		var fallbackFlag *Flag
		for current := p; current != nil; current = current.parent {
			for opt, flag := range current.longOpts {
				if len(opt) >= len(bestName) || len(opt) > len(name) {
					continue
				}
				if !hasPrefix(name, opt, current.config.longCaseIgnore) {
					continue
				}
				if len(opt) < len(name) && name[len(opt)] != '=' {
					continue
				}
				if flag.HasArg == NoArgument {
					continue
				}
				if len(opt) > len(fallbackName) {
					fallbackName = opt
					fallbackFlag = flag
				}
			}
		}
		if fallbackFlag != nil {
			return args, fallbackFlag, Option{Name: fallbackName, HasArg: true, Arg: name[len(fallbackName)+1:]}, nil
		}
		return args, nil, Option{}, p.optError("unknown option: " + name)
	}
	return args, bestFlag, Option{Name: bestName, HasArg: true, Arg: name[len(bestName)+1:]}, nil
}

// resolveLongOpt handles argument consumption for an exact long-option match.
func (p *Parser) resolveLongOpt(name string, flag *Flag, args []string) ([]string, *Flag, Option, error) {
	option := Option{Name: name}

	if flag.HasArg == NoArgument {
		return args, flag, option, nil
	}

	// For RequiredArgument: consume the next arg unconditionally.
	// For OptionalArgument: consume only if all three conditions hold:
	//   1. No = delimited arg was present (guaranteed — findLongOpt
	//      handles = splitting before calling resolveLongOpt)
	//   2. Another argument exists after the current option
	//   3. That argument does NOT start with '-' (avoids consuming
	//      short options, long options, or the -- terminator)
	if len(args) > 0 {
		if flag.HasArg == RequiredArgument || args[0][0] != '-' {
			option.Arg = args[0]
			option.HasArg = true
			return args[1:], flag, option, nil
		}
		// OptionalArgument with a '-' prefixed next arg — don't consume
		return args, flag, option, nil
	}

	if flag.HasArg == RequiredArgument {
		return args, nil, option, p.optError("option requires an argument: " + name)
	}

	// OptionalArgument with no remaining args
	return args, flag, option, nil
}

func (p *Parser) findShortOpt(c byte, word string, args []string) ([]string, string, *Flag, Option, error) {
	if debug {
		slog.Debug("findShortOpt", "c", byteString(c), "word", word, "args", args)
	}

	// POSIX disallows `-` as a short-opt option.
	if c == '-' {
		return args, word, nil, Option{}, p.optError("invalid option: " + byteString(c))
	}

	// Walk the parser chain: self first, then ancestors.
	for current := p; current != nil; current = current.parent {
		matched, flag := current.lookupShortOpt(c)
		if flag == nil {
			continue
		}

		option := Option{Name: byteString(matched)}

		switch flag.HasArg {
		case NoArgument:
			if debug {
				slog.Debug("findShortOpt", "hasArg", "none", "c", byteString(c))
			}

		case RequiredArgument:
			if debug {
				slog.Debug("findShortOpt", "hasArg", "required", "c", byteString(c))
			}
			switch {
			case len(word) > 0:
				option.Arg = word
				word = ""
			case len(args) == 0:
				return args, word, nil, option, p.optError("option requires an argument: " + byteString(c))
			default:
				option.Arg = args[0]
				args = args[1:]
			}
			option.HasArg = true

		case OptionalArgument:
			if debug {
				slog.Debug("findShortOpt", "hasArg", "optional", "c", byteString(c))
			}
			if len(word) > 0 {
				option.Arg = word
				word = ""
				option.HasArg = true
			} else if len(args) > 0 {
				option.Arg = args[0]
				args = args[1:]
				option.HasArg = true
			}

		default:
			return args, word, nil, option, p.optErrorf("unknown argument type: %d", flag.HasArg)
		}

		if debug {
			slog.Debug("findShortOpt", "args", args, "word", word, "option", option, "err", "yield")
		}
		return args, word, flag, option, nil
	}

	return args, word, nil, Option{}, p.optError("unknown option: " + byteString(c))
}

// lookupShortOpt finds a short option in this parser's shortOpts array,
// respecting the case-sensitivity configuration. Returns the matched key
// and the flag definition.
func (p *Parser) lookupShortOpt(c byte) (byte, *Flag) {
	if flag := p.shortOpts[c]; flag != nil {
		return c, flag
	}
	if !p.config.shortCaseIgnore {
		return 0, nil
	}
	// Case-insensitive fallback: try the opposite case.
	var alt byte
	switch {
	case c >= 'a' && c <= 'z':
		alt = c - 32
	case c >= 'A' && c <= 'Z':
		alt = c + 32
	default:
		return 0, nil
	}
	if flag := p.shortOpts[alt]; flag != nil {
		return alt, flag
	}
	return 0, nil
}

// tryLongOnly attempts to match a single-dash argument as a long option
// per getopt_long_only(3). Returns (true, option, err) on match or when
// no short-option fallback is possible. Returns (false, ...) when the
// caller should fall through to short option parsing.
func (p *Parser) tryLongOnly(
	word string, remaining []string,
) (matched bool, args []string, flag *Flag, option Option, err error) {
	// Suppress error logging during the long option probe —
	// we may fall back to short options.
	savedErrors := p.config.enableErrors
	p.config.enableErrors = false
	args, flag, option, err = p.findLongOpt(word, remaining)
	p.config.enableErrors = savedErrors

	if err == nil {
		return true, args, flag, option, nil
	}

	// Long match failed — fall back to short options per getopt_long_only(3).
	if p.shortOptN == 0 {
		err = p.optError(err.Error())
		return true, remaining, nil, option, err
	}

	// Has short opts — restore the original arg for short option parsing.
	restored := append([]string{"-" + word}, remaining...)
	return false, restored, nil, Option{}, nil
}

// Options returns an iterator over parsed options. Each iteration yields
// an [Option] and an error. When a subcommand is encountered, the iterator
// dispatches to the child parser automatically.
//
//nolint:gocognit,gocyclo,cyclop,funlen // main parser loop handles --, --long, -short, long-only, commands, and parse modes
func (p *Parser) Options() iter.Seq2[Option, error] {
	if debug {
		slog.Debug("Iterator")
	}
	return func(yield func(Option, error) bool) {
		var err error
		cleanupDone := false
		defer func() {
			if !cleanupDone {
				p.Args = append(p.nonOpts, p.Args...)
			}
		}()

		if debug {
			slog.Debug("Options", "args", p.Args)
		}
	out:
		for len(p.Args) > 0 {
			if debug {
				slog.Debug("Options", "arg[0]", p.Args[0])
			}
			option := Option{}
			switch {
			case p.Args[0] == "--": // Stop parsing options
				if debug {
					slog.Debug("Options", "break", true)
				}
				p.Args = append(p.nonOpts, p.Args[1:]...)
				cleanupDone = true
				break out

			case strings.HasPrefix(p.Args[0], "--"):
				if debug {
					slog.Debug("Options", "prefix", "--")
				}
				var flag *Flag
				p.Args, flag, option, err = p.findLongOpt(p.Args[0][2:], p.Args[1:])
				if err != nil {
					if !yield(option, err) {
						return
					}
					continue
				}
				if flag != nil && flag.Handle != nil {
					if herr := flag.Handle(option.Name, option.Arg); herr != nil {
						if !yield(Option{}, herr) {
							return
						}
					}
					continue
				}
				if !yield(option, nil) {
					return
				}

			case strings.HasPrefix(p.Args[0], "-"):
				if debug {
					slog.Debug("Options", "prefix", "-")
				}
				if p.config.longOptsOnly { //nolint:nestif // long-only dispatch requires try-long then fall-through-to-short
					var matched bool
					var flag *Flag
					matched, p.Args, flag, option, err = p.tryLongOnly(p.Args[0][1:], p.Args[1:])
					if matched {
						if err != nil {
							if !yield(option, err) {
								return
							}
							continue
						}
						if flag != nil && flag.Handle != nil {
							if herr := flag.Handle(option.Name, option.Arg); herr != nil {
								if !yield(Option{}, herr) {
									return
								}
							}
							continue
						}
						if !yield(option, nil) {
							return
						}
						continue
					}
				}

				// iterate over each character in the word looking
				// for short options
				word := p.Args[0][1:]
				p.Args = p.Args[1:]
				for len(word) > 0 {
					if debug {
						slog.Debug("Options", "word", word)
					}
					var flag *Flag
					p.Args, word, flag, option, err = p.findShortOpt(word[0], word[1:], p.Args)

					// Transform usages such as `-W foo` into `--foo`
					if option.Name == "W" && p.config.gnuWords {
						option.Name = option.Arg
					}

					if err != nil {
						if !yield(option, err) {
							return
						}
						break
					}
					if flag != nil && flag.Handle != nil {
						if herr := flag.Handle(option.Name, option.Arg); herr != nil {
							if !yield(Option{}, herr) {
								return
							}
							break
						}
						continue
					}
					if !yield(option, nil) {
						return
					}
				}

			default:
				// Check if this is a registered command
				if cmd, exists := p.GetCommand(p.Args[0]); exists {
					cmdName := p.Args[0]
					_, err := prepareCommand(cmdName, cmd, true, p.Args[1:])
					if err != nil {
						if !yield(Option{}, err) {
							return
						}
					}
					p.activeCmd = cmdName
					p.activeCmdParser = cmd
					p.Args = []string{}
					break out
				}

				// Handle as non-option
				switch p.config.parseMode {
				case ParseDefault:
					p.nonOpts = append(p.nonOpts, p.Args[0])

				case ParseNonOpts:
					option := Option{
						Name: string(byte(1)),
						Arg:  p.Args[0],
					}
					if !yield(option, nil) {
						return
					}

				case ParsePosixlyCorrect:
					break out
				}
				p.Args = p.Args[1:]
			}
		}

		if !cleanupDone {
			cleanupDone = true
			p.Args = append(p.nonOpts, p.Args...)
		}
	}
}

// AddCmd registers a new subcommand with this parser.
func (p *Parser) AddCmd(name string, parser *Parser) *Parser {
	if parser != nil {
		if !p.config.strictSubcommands {
			parser.parent = p
		}
		parser.Name = name
	}
	return p.Commands.AddCmd(name, parser)
}

// AddAlias creates an alias for an existing command.
func (p *Parser) AddAlias(alias, existingCommand string) error {
	return p.Commands.AddAlias(alias, existingCommand)
}

// GetCommand retrieves a parser by command name.
func (p *Parser) GetCommand(name string) (*Parser, bool) {
	return p.Commands.getCommand(name, p.config.commandCaseIgnore)
}

// ListCommands returns all command mappings.
func (p *Parser) ListCommands() map[string]*Parser {
	return p.Commands.ListCommands()
}

// ExecuteCommand finds and executes a command.
func (p *Parser) ExecuteCommand(name string, args []string) (*Parser, error) {
	return p.Commands.executeCommand(name, args, p.config.commandCaseIgnore)
}

// SetStrictSubcommands enables or disables strict subcommand mode.
// When enabled, child parsers registered via AddCmd do not inherit
// parent options — unknown options in a subcommand produce an error
// instead of walking the parent chain.
func (p *Parser) SetStrictSubcommands(strict bool) {
	p.config.strictSubcommands = strict
}

// StrictSubcommands reports whether strict subcommand mode is enabled.
func (p *Parser) StrictSubcommands() bool {
	return p.config.strictSubcommands
}

// GetAliases returns all aliases for a given parser.
func (p *Parser) GetAliases(targetParser *Parser) []string {
	return p.Commands.GetAliases(targetParser)
}

// SetShortHandler attaches a handler to a short option registered on this
// parser. Returns an error if no matching short option is found.
//
// SetShortHandler only modifies options on this parser — it does not walk
// the parent chain.
func (p *Parser) SetShortHandler(c byte, handler func(string, string) error) error {
	f := p.shortOpts[c]
	if f == nil {
		return fmt.Errorf("unknown option: -%c", c)
	}
	f.Handle = handler
	return nil
}

// SetLongHandler attaches a handler to a long option registered on this
// parser. Returns an error if no matching long option is found.
//
// Long option names may be single characters (e.g., "v" for --v). Use
// SetLongHandler for long options and SetShortHandler for short options —
// the two namespaces are independent.
//
// SetLongHandler only modifies options on this parser — it does not walk
// the parent chain.
func (p *Parser) SetLongHandler(name string, handler func(string, string) error) error {
	f, ok := p.longOpts[name]
	if !ok {
		return fmt.Errorf("unknown option: --%s", name)
	}
	f.Handle = handler
	return nil
}

// SetHandler is a convenience method that attaches a handler to a matching
// option using command-line prefix syntax. Pass "--name" for long options
// or "-c" for short options. Returns an error if the prefix is missing or
// no matching option is found.
//
// Examples:
//
//	parser.SetHandler("--verbose", handler)  // calls SetLongHandler("verbose", handler)
//	parser.SetHandler("-v", handler)          // calls SetShortHandler('v', handler)
//	parser.SetHandler("--v", handler)         // calls SetLongHandler("v", handler)
//
// SetHandler only modifies options on this parser — it does not walk the
// parent chain.
func (p *Parser) SetHandler(name string, handler func(string, string) error) error {
	if strings.HasPrefix(name, "--") {
		return p.SetLongHandler(name[2:], handler)
	}
	if strings.HasPrefix(name, "-") {
		return p.SetShortHandler(name[1], handler)
	}
	return fmt.Errorf("invalid option name: %s", name)
}

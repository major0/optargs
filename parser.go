package optargs

import (
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"sort"
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
// flags and constructor parameters.
type ParserConfig struct {
	enableErrors bool
	parseMode    ParseMode

	shortCaseIgnore bool
	gnuWords        bool

	longCaseIgnore bool
	longOptsOnly   bool

	// Command case sensitivity
	commandCaseIgnore bool
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
	shortOpts map[byte]*Flag
	longOpts  map[string]*Flag
	config    ParserConfig

	// Command support - simple map of command name to parser
	Commands CommandRegistry
	parent   *Parser
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
		Args:   args,
		config: config,
	}

	for c := range shortOpts {
		if !isGraph(c) {
			return nil, parser.optErrorf("invalid short option: %c", c)
		}
		switch c {
		case ':', ';', '-':
			return nil, parser.optErrorf("prohibited short option: %c", c)
		}
	}
	parser.shortOpts = shortOpts

	for s := range longOpts {
		for _, r := range s {
			if unicode.IsSpace(r) || !unicode.IsGraphic(r) {
				return nil, parser.optErrorf("invalid long option: %s", s)
			}
		}
	}
	parser.longOpts = longOpts

	// Initialize command registry
	parser.Commands = NewCommandRegistry()

	return &parser, nil
}

// NewParserWithCaseInsensitiveCommands creates a new parser with case insensitive command matching enabled
func NewParserWithCaseInsensitiveCommands(shortOpts map[byte]*Flag, longOpts map[string]*Flag, args []string) (*Parser, error) {
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

// longOptCandidate holds a registered option that matched as a prefix of the input.
type longOptCandidate struct {
	name string // registered option name
	flag *Flag  // option definition
}

func (p *Parser) findLongOpt(name string, args []string) ([]string, *Flag, Option, error) {
	// Phase 1: Walk self + all ancestors, collecting every registered
	// option whose name is a prefix of (or equal to) the input.
	var candidates []longOptCandidate
	for current := p; current != nil; current = current.parent {
		for opt, flag := range current.longOpts {
			if len(opt) > len(name) {
				continue
			}
			if !hasPrefix(name, opt, current.config.longCaseIgnore) {
				continue
			}
			candidates = append(candidates, longOptCandidate{name: opt, flag: flag})
		}
	}

	if len(candidates) == 0 {
		return args, nil, Option{}, p.optError("unknown option: " + name)
	}

	// Phase 2: Sort candidates by name length, longest first.
	sort.Slice(candidates, func(i, j int) bool {
		return len(candidates[i].name) > len(candidates[j].name)
	})

	// Phase 3: Iterate candidates longest-first. For each candidate:
	// - If exact match (same length): handle as space-separated arg.
	// - If next char after the candidate name is '=': split there.
	// - Otherwise: skip to next-shortest candidate.
	for _, c := range candidates {
		if len(c.name) == len(name) {
			// Exact match — argument comes from the next element in args.
			option := Option{Name: c.name}
			if c.flag.HasArg == NoArgument {
				return args, c.flag, option, nil
			}
			if len(args) > 0 {
				option.Arg = args[0]
				option.HasArg = true
				return args[1:], c.flag, option, nil
			}
			if c.flag.HasArg == RequiredArgument {
				return args, nil, option, p.optError("option requires an argument: " + name)
			}
			// OptionalArgument with no arg available
			return args, c.flag, option, nil
		}

		// Partial match — check for '=' boundary.
		if name[len(c.name)] == '=' {
			option := Option{Name: c.name}
			if c.flag.HasArg == NoArgument {
				// NoArgument option can't accept the '=value' portion.
				// Skip to next-shortest candidate that might accept it.
				continue
			}
			option.Arg = name[len(c.name)+1:]
			option.HasArg = true
			return args, c.flag, option, nil
		}
		// No '=' at boundary — this candidate doesn't match at a
		// valid split point. Try next-shortest.
	}

	return args, nil, Option{}, p.optError("unknown option: " + name)
}

func (p *Parser) findShortOpt(c byte, word string, args []string) ([]string, string, *Flag, Option, error) {
	slog.Debug("findShortOpt", "c", string(c), "word", word, "args", args)

	// POSIX disallows `-` as a short-opt option.
	if c == '-' {
		return args, word, nil, Option{}, p.optError("invalid option: " + string(c))
	}

	// Walk the parser chain: self first, then ancestors.
	for current := p; current != nil; current = current.parent {
		matched, flag := current.lookupShortOpt(c)
		if flag == nil {
			continue
		}

		option := Option{Name: string(matched)}

		switch flag.HasArg {
		case NoArgument:
			slog.Debug("findShortOpt", "hasArg", "none", "c", string(c))

		case RequiredArgument:
			slog.Debug("findShortOpt", "hasArg", "required", "c", string(c))
			if len(word) > 0 {
				option.Arg = word
				word = ""
			} else if len(args) == 0 {
				return args, word, nil, option, p.optError("option requires an argument: " + string(c))
			} else {
				option.Arg = args[0]
				args = args[1:]
			}
			option.HasArg = true

		case OptionalArgument:
			slog.Debug("findShortOpt", "hasArg", "optional", "c", string(c))
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

		slog.Debug("findShortOpt", "args", args, "word", word, "option", option, "err", "yield")
		return args, word, flag, option, nil
	}

	return args, word, nil, Option{}, p.optError("unknown option: " + string(c))
}

// lookupShortOpt finds a short option in this parser's shortOpts map,
// respecting the case-sensitivity configuration. Returns the matched key
// and the flag definition.
func (p *Parser) lookupShortOpt(c byte) (byte, *Flag) {
	if flag, ok := p.shortOpts[c]; ok {
		return c, flag
	}
	if !p.config.shortCaseIgnore {
		return 0, nil
	}
	// Case-insensitive fallback: try the opposite case.
	var alt byte
	if c >= 'a' && c <= 'z' {
		alt = c - 32
	} else if c >= 'A' && c <= 'Z' {
		alt = c + 32
	} else {
		return 0, nil
	}
	if flag, ok := p.shortOpts[alt]; ok {
		return alt, flag
	}
	return 0, nil
}

// tryLongOnly attempts to match a single-dash argument as a long option
// per getopt_long_only(3). Returns (true, option, err) on match or when
// no short-option fallback is possible. Returns (false, ...) when the
// caller should fall through to short option parsing.
func (p *Parser) tryLongOnly(word string, remaining []string) (matched bool, args []string, flag *Flag, option Option, err error) {
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
	if len(p.shortOpts) == 0 {
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
func (p *Parser) Options() iter.Seq2[Option, error] {
	slog.Debug("Iterator")
	return func(yield func(Option, error) bool) {
		var err error
		cleanupDone := false
		defer func() {
			if !cleanupDone {
				p.Args = append(p.nonOpts, p.Args...)
			}
		}()

		slog.Debug("Options", "args", p.Args)
	out:
		for len(p.Args) > 0 {
			slog.Debug("Options", "arg[0]", p.Args[0])
			option := Option{}
			switch {
			case p.Args[0] == "--": // Stop parsing options
				slog.Debug("Options", "break", true)
				p.Args = append(p.nonOpts, p.Args[1:]...)
				cleanupDone = true
				break out

			case strings.HasPrefix(p.Args[0], "--"):
				slog.Debug("Options", "prefix", "--")
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
				slog.Debug("Options", "prefix", "-")
				if p.config.longOptsOnly {
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
					slog.Debug("Options", "word", word)
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
					_, err := prepareCommand(p.Args[0], cmd, true, p.Args[1:])
					if err != nil {
						if !yield(Option{}, err) {
							return
						}
					}
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

		cleanupDone = true
		p.Args = append(p.nonOpts, p.Args...)
	}
}

// AddCmd registers a new subcommand with this parser.
func (p *Parser) AddCmd(name string, parser *Parser) *Parser {
	if parser != nil {
		parser.parent = p
	}
	return p.Commands.AddCmd(name, parser)
}

// AddAlias creates an alias for an existing command
func (p *Parser) AddAlias(alias, existingCommand string) error {
	return p.Commands.AddAlias(alias, existingCommand)
}

// GetCommand retrieves a parser by command name
func (p *Parser) GetCommand(name string) (*Parser, bool) {
	return p.Commands.getCommand(name, p.config.commandCaseIgnore)
}

// ListCommands returns all command mappings
func (p *Parser) ListCommands() map[string]*Parser {
	return p.Commands.ListCommands()
}

// ExecuteCommand finds and executes a command
func (p *Parser) ExecuteCommand(name string, args []string) (*Parser, error) {
	return p.Commands.executeCommand(name, args, p.config.commandCaseIgnore)
}

// HasCommands returns true if any commands are registered
func (p *Parser) HasCommands() bool {
	return p.Commands.HasCommands()
}

// GetAliases returns all aliases for a given parser
func (p *Parser) GetAliases(targetParser *Parser) []string {
	return p.Commands.GetAliases(targetParser)
}

// SetShortHandler attaches a handler to a short option registered on this
// parser. Returns an error if no matching short option is found.
//
// SetShortHandler only modifies options on this parser — it does not walk
// the parent chain.
func (p *Parser) SetShortHandler(c byte, handler func(string, string) error) error {
	f, ok := p.shortOpts[c]
	if !ok {
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

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
	Args       []string
	nonOpts    []string
	shortOpts  map[byte]*Flag
	longOpts   map[string]*Flag
	config     ParserConfig
	lockConfig bool

	// Command support - simple map of command name to parser
	Commands CommandRegistry
	parent   *Parser
}

// NewParser creates a Parser from pre-built configuration, short option map,
// long option map, and argument list. Most callers should use [GetOpt],
// [GetOptLong], or [GetOptLongOnly] instead.
func NewParser(config ParserConfig, shortOpts map[byte]*Flag, longOpts map[string]*Flag, args []string) (*Parser, error) {
	parser := Parser{
		Args:       args,
		config:     config,
		lockConfig: false,
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

func (p *Parser) optErrorf(msg string, args ...interface{}) error {
	return p.optError(fmt.Sprintf(msg, args...))
}

// longOptCandidate holds a registered option that matched as a prefix of the input.
type longOptCandidate struct {
	name string // registered option name
	flag *Flag  // option definition
}

func (p *Parser) findLongOpt(name string, args []string) ([]string, Option, error) {
	// Phase 1: Walk self + all ancestors, collecting every registered
	// option whose name is a prefix of (or equal to) the input.
	var candidates []longOptCandidate
	for current := p; current != nil; current = current.parent {
		caseIgnore := current.config.longCaseIgnore
		for opt, flag := range current.longOpts {
			if len(opt) > len(name) {
				continue
			}
			if caseIgnore {
				if !hasPrefix(name, opt, true) {
					continue
				}
			} else {
				if !strings.HasPrefix(name, opt) {
					continue
				}
			}
			candidates = append(candidates, longOptCandidate{name: opt, flag: flag})
		}
	}

	if len(candidates) == 0 {
		return args, Option{}, p.optError("unknown option: " + name)
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
				return args, option, nil
			}
			if len(args) > 0 {
				option.Arg = args[0]
				option.HasArg = true
				return args[1:], option, nil
			}
			if c.flag.HasArg == RequiredArgument {
				return args, option, p.optError("option requires an argument: " + name)
			}
			// OptionalArgument with no arg available
			return args, option, nil
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
			return args, option, nil
		}
		// No '=' at boundary — this candidate doesn't match at a
		// valid split point. Try next-shortest.
	}

	return args, Option{}, p.optError("unknown option: " + name)
}

func (p *Parser) findShortOpt(c byte, word string, args []string) ([]string, string, Option, error) {
	slog.Debug("findShortOpt", "c", string(c), "word", word, "args", args)

	// POSIX disallows `-` as a short-opt option.
	if c == '-' {
		return args, word, Option{}, p.optError("invalid option: " + string(c))
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
				return args, word, option, p.optError("option requires an argument: " + string(c))
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
			return args, word, option, p.optErrorf("unknown argument type: %d", flag.HasArg)
		}

		slog.Debug("findShortOpt", "args", args, "word", word, "option", option, "err", "yield")
		return args, word, option, nil
	}

	return args, word, Option{}, p.optError("unknown option: " + string(c))
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
				p.Args, option, err = p.findLongOpt(p.Args[0][2:], p.Args[1:])
				if !yield(option, err) {
					return
				}

			case strings.HasPrefix(p.Args[0], "-"):
				slog.Debug("Options", "prefix", "-")
				if p.config.longOptsOnly {
					longOnlyWord := p.Args[0][1:]
					remainingArgs := p.Args[1:]

					// Suppress error logging during the long option
					// probe — we may fall back to short options.
					savedErrors := p.config.enableErrors
					p.config.enableErrors = false
					p.Args, option, err = p.findLongOpt(longOnlyWord, remainingArgs)
					p.config.enableErrors = savedErrors

					if err == nil {
						if !yield(option, err) {
							return
						}
						continue
					}

					// Long match failed — fall back to short options
					// per getopt_long_only(3)
					if len(p.shortOpts) == 0 {
						err = p.optError(err.Error())
						if !yield(option, err) {
							return
						}
						continue
					}

					// Restore args for short option parsing
					p.Args = append([]string{"-" + longOnlyWord}, remainingArgs...)
				}

				// iterate over each character in the word looking
				// for short options
				remainingArgs := p.Args[1:]
				for word := p.Args[0][1:]; len(word) > 0; {
					slog.Debug("Options", "word", word)
					p.Args, word, option, err = p.findShortOpt(word[0], word[1:], remainingArgs)

					// Transform usages such as `-W foo` into `--foo`
					if option.Name == "W" && p.config.gnuWords {
						option.Name = option.Arg
					}

					if !yield(option, err) {
						return
					}
				}

			default:
				// Check if this is a registered command
				if _, exists := p.GetCommand(p.Args[0]); exists {
					_, err := p.ExecuteCommand(p.Args[0], p.Args[1:])
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
	if p.config.commandCaseIgnore {
		return p.Commands.getCommandFold(name)
	}
	return p.Commands.GetCommand(name)
}

// ListCommands returns all command mappings
func (p *Parser) ListCommands() map[string]*Parser {
	return p.Commands.ListCommands()
}

// ExecuteCommand finds and executes a command
func (p *Parser) ExecuteCommand(name string, args []string) (*Parser, error) {
	if p.config.commandCaseIgnore {
		return p.Commands.executeCommandFold(name, args)
	}
	return p.Commands.ExecuteCommand(name, args)
}

// HasCommands returns true if any commands are registered
func (p *Parser) HasCommands() bool {
	return p.Commands.HasCommands()
}

// GetAliases returns all aliases for a given parser
func (p *Parser) GetAliases(targetParser *Parser) []string {
	return p.Commands.GetAliases(targetParser)
}

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
		longOpts:   longOpts,
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
				return nil, fmt.Errorf("invalid long option: %s", s)
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

func (p *Parser) findLongOpt(name string, args []string) ([]string, Option, error) {
	var best Option
	for opt := range p.longOpts {
		// Filter through the options "ruling out" anything that
		// is not a candidate.
		//
		// It is important to consider that the `=` sign is
		// allowed in the option name, and as such it is possible
		// to have _both_ of the following in the option list.
		//
		// - name: `foo=`, hasArg: NoArgument
		// - name: `foo`, hasArg: RequiredArgument
		//
		// Should the user pass in `--foo=` then we need to make
		// a decision as to which handler to use.
		//
		// `--foo=bar` is easy as we can exclude the `NoArgument`
		// option.
		if len(opt) > len(name) {
			// There is simply no way this can be a valid match.
			continue
		} else if len(opt) == len(name) {
			if p.config.longCaseIgnore && !strings.EqualFold(opt, name) {
				continue
			} else {
				if opt != name {
					continue
				}
			}
		} else {
			if name[len(opt)] != '=' {
				continue
			}

			if p.config.longCaseIgnore && !hasPrefix(name, opt, p.config.longCaseIgnore) {
				continue
			}
		}

		// From here we have a possible candidate, but we do not
		// yet know how to handle any potential `=` in the name
		// until we look at how to handle the candidate based on
		// the HasArg field.
		option := Option{
			HasArg: false,
		}
		if p.longOpts[opt].HasArg != NoArgument {
			option.Name = opt
			if len(name) == len(opt) {
				if len(args) != 0 {
					option.Arg = args[0]
					args = args[1:]
					option.HasArg = true
				} else if p.longOpts[opt].HasArg == RequiredArgument {
					// Defer logging to caller when parent chain may re-wrap
					if p.parent != nil {
						return args, option, errors.New("option requires an argument: " + name)
					}
					return args, option, p.optError("option requires an argument: " + name)
				}
			} else {
				option.Arg = name[len(opt):]
				option.HasArg = true
			}
		} else if len(name) == len(opt) {
			// No argument allowed, but the names have already
			// been filtered, simply need to validate their
			// length matches.
			option.Name = opt
		}

		// We need to continue processing candidates as the "last"
		// defined candidate that is the "best" match must always
		// be used.
		if len(option.Name) >= len(best.Name) {
			best = option
		}
	}

	if best.Name != "" {
		return args, best, nil
	}

	// Only log error if there's no parent to fall back to
	if p.parent == nil {
		return args, Option{}, p.optError("unknown option: " + name)
	}

	// Return error without logging - parent will be consulted
	return args, Option{}, errors.New("unknown option: " + name)
}

func (p *Parser) findShortOpt(c byte, word string, args []string) ([]string, string, Option, error) {
	slog.Debug("findShortOpt", "c", string(c), "word", word, "args", args)

	// POSIX disallows `-` as a short-opt option.
	if c == '-' {
		return args, word, Option{}, p.optError("invalid option: " + string(c))
	}

	// We have to iterate the shortOpts in order to support case
	// insensitive options.
	for opt := range p.shortOpts {
		if p.config.shortCaseIgnore {
			if !strings.EqualFold(string(c), string(opt)) {
				continue
			}
		} else if c != opt {
			continue
		}

		option := Option{
			Name:   string(opt),
			HasArg: false,
			Arg:    "",
		}

		switch p.shortOpts[opt].HasArg {
		case NoArgument:
			slog.Debug("findShortOpt", "hasArg", "none", "c", string(c), "opt", string(opt))

		case RequiredArgument:
			slog.Debug("findShortOpt", "hasArg", "required", "c", string(c), "opt", string(opt))
			var arg string
			if len(word) > 0 {
				arg = word
				word = ""
			} else {
				if len(args) == 0 {
					// Defer logging to caller when parent chain may re-wrap
					if p.parent != nil {
						return args, word, option, errors.New("option requires an argument: " + string(c))
					}
					return args, word, option, p.optError("option requires an argument: " + string(c))
				}
				arg = args[0]
				args = args[1:]
			}

			option.Arg = arg
			option.HasArg = true

		case OptionalArgument:
			slog.Debug("findShortOpt", "hasArg", "optional", "c", string(c), "opt", string(opt))
			var arg string
			if len(word) > 0 {
				arg = word
				word = ""
				option.HasArg = true
			} else if len(args) > 0 {
				arg = args[0]
				args = args[1:]
				option.HasArg = true
			}

			option.Arg = arg

		default:
			return args, word, option, p.optErrorf("unknown argument type: %d", p.shortOpts[c].HasArg)
		}

		slog.Debug("findShortOpt", "args", args, "word", word, "option", option, "err", "yield")
		return args, word, option, nil
	}

	// Only log error if there's no parent to fall back to
	if p.parent == nil {
		return args, word, Option{}, p.optError("unknown option: " + string(c))
	}

	// Return error without logging - parent will be consulted
	return args, word, Option{}, errors.New("unknown option: " + string(c))
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
				var remainingArgs []string
				if len(p.Args) > 1 {
					remainingArgs = p.Args[1:]
				}
				p.Args, option, err = p.findLongOptWithFallback(p.Args[0][2:], remainingArgs)
				if !yield(option, err) {
					return
				}

			case strings.HasPrefix(p.Args[0], "-"):
				slog.Debug("Options", "prefix", "-")
				if p.config.longOptsOnly {
					longOnlyWord := p.Args[0][1:]
					var remainingArgs []string
					if len(p.Args) > 1 {
						remainingArgs = p.Args[1:]
					}

					// Suppress error logging during the long option
					// probe — we may fall back to short options.
					savedErrors := p.config.enableErrors
					p.config.enableErrors = false
					p.Args, option, err = p.findLongOptWithFallback(longOnlyWord, remainingArgs)
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
				originalArgs := p.Args // Save original args for compacted option processing
				for word := p.Args[0][1:]; len(word) > 0; {
					slog.Debug("Options", "word", word)
					var remainingArgs []string
					if len(originalArgs) > 1 {
						remainingArgs = originalArgs[1:]
					}
					p.Args, word, option, err = p.findShortOptWithFallback(word[0], word[1:], remainingArgs)

					// Transform usages such as `-W foo` into `--foo`
					if option.Name == "W" && p.config.gnuWords {
						option.Name = option.Arg
					}

					if !yield(option, err) {
						return
					}
				}

			default:
				// Check if this is a command
				if p.HasCommands() {
					if _, exists := p.GetCommand(p.Args[0]); exists {
						// Found a command, execute it with remaining args
						remainingArgs := p.Args[1:]
						_, err := p.ExecuteCommand(p.Args[0], remainingArgs)
						if err != nil {
							if !yield(Option{}, err) {
								return
							}
						}
						// Command handled, stop processing at root level
						p.Args = []string{}
						break out
					}
				}

				// Not a command, handle as non-option
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
	return p.Commands.GetCommandCaseInsensitive(name, p.config.commandCaseIgnore)
}

// ListCommands returns all command mappings
func (p *Parser) ListCommands() map[string]*Parser {
	return p.Commands.ListCommands()
}

// ExecuteCommand finds and executes a command
func (p *Parser) ExecuteCommand(name string, args []string) (*Parser, error) {
	return p.Commands.ExecuteCommandCaseInsensitive(name, args, p.config.commandCaseIgnore)
}

// HasCommands returns true if any commands are registered
func (p *Parser) HasCommands() bool {
	return p.Commands.HasCommands()
}

// GetAliases returns all aliases for a given parser
func (p *Parser) GetAliases(targetParser *Parser) []string {
	return p.Commands.GetAliases(targetParser)
}

// findLongOptWithFallback finds a long option, falling back through the entire parent chain.
// Error reporting uses the originating parser's error mode, not the parent's.
func (p *Parser) findLongOptWithFallback(name string, args []string) ([]string, Option, error) {
	remainingArgs, option, err := p.findLongOpt(name, args)
	if err == nil {
		return remainingArgs, option, err
	}

	// Only fall back to parent for "unknown option" errors.
	// Other errors like "option requires an argument" should be returned immediately.
	if !strings.Contains(err.Error(), "unknown option") {
		return remainingArgs, option, err
	}

	// Walk the parent chain looking for the option
	current := p.parent
	for current != nil {
		parentArgs, parentOpt, parentErr := current.findLongOpt(name, args)
		if parentErr == nil {
			return parentArgs, parentOpt, nil
		}
		if !strings.Contains(parentErr.Error(), "unknown option") {
			// Found the option in a parent but it had a different error (e.g. missing arg).
			// Report using the originating parser's error mode.
			return parentArgs, parentOpt, p.optError(parentErr.Error())
		}
		current = current.parent
	}

	// Not found anywhere in the chain — report using originating parser's error mode
	return remainingArgs, option, p.optError("unknown option: " + name)
}

// findShortOptWithFallback finds a short option, falling back through the entire parent chain.
// Error reporting uses the originating parser's error mode, not the parent's.
func (p *Parser) findShortOptWithFallback(c byte, word string, args []string) ([]string, string, Option, error) {
	remainingArgs, remainingWord, option, err := p.findShortOpt(c, word, args)
	if err == nil {
		return remainingArgs, remainingWord, option, err
	}

	// Only fall back for "unknown option" errors
	if !strings.Contains(err.Error(), "unknown option") {
		return remainingArgs, remainingWord, option, err
	}

	// Walk the parent chain looking for the option
	current := p.parent
	for current != nil {
		if parentFlag, exists := current.shortOpts[c]; exists {
			parentOption := Option{
				Name:   string(c),
				HasArg: false,
				Arg:    "",
			}

			switch parentFlag.HasArg {
			case NoArgument:
				return args, word, parentOption, nil

			case RequiredArgument:
				if len(word) > 0 {
					parentOption.Arg = word
					parentOption.HasArg = true
					return args, "", parentOption, nil
				}
				if len(args) == 0 {
					return args, word, parentOption, p.optError("option requires an argument: " + string(c))
				}
				parentOption.Arg = args[0]
				parentOption.HasArg = true
				return args[1:], "", parentOption, nil

			case OptionalArgument:
				if len(word) > 0 {
					parentOption.Arg = word
					parentOption.HasArg = true
					return args, "", parentOption, nil
				}
				if len(args) > 0 {
					parentOption.Arg = args[0]
					parentOption.HasArg = true
					return args[1:], "", parentOption, nil
				}
				return args, word, parentOption, nil

			default:
				return args, word, parentOption, p.optErrorf("unknown argument type: %d", parentFlag.HasArg)
			}
		}
		current = current.parent
	}

	// Not found anywhere in the chain
	return args, word, Option{}, p.optError("unknown option: " + string(c))
}

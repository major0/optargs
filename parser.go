package optargs

import (
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"regexp"
	"strings"
)

type ParseMode int

const (
	ParseDefault ParseMode = iota
	ParseNonOpts
	ParsePosixlyCorrect
)

type ParserConfig struct {
	enableErrors bool
	parseMode    ParseMode

	shortCaseIgnore bool
	gnuWords        bool

	longCaseIgnore bool
	longOptsOnly   bool
}

type Parser struct {
	Args       []string
	nonOpts    []string
	shortOpts  map[byte]*Flag
	longOpts   map[string]*Flag
	config     ParserConfig
	lockConfig bool
}

func NewParser(config ParserConfig, shortOpts map[byte]*Flag, longOpts map[string]*Flag, args []string) (*Parser, error) {
	parser := Parser{
		Args:       args,
		config:     config,
		longOpts:   longOpts,
		lockConfig: false,
	}

	for c, _ := range shortOpts {
		if !isGraph(c) {
			return nil, parser.optErrorf("Invalid short option: %c", c)
		}
		switch c {
		case ':', ';', '-':
			return nil, parser.optErrorf("Prohibited short option: %c", c)
		}
	}
	parser.shortOpts = shortOpts

	// Regex pattern to find any character that is _not_ a graph or
	// _is_ a space. Using regexp here is slightly faster than
	// terating the string char by char calling `isGraph()` on it, but
	// ultimately has the same effect.
	notGraph := regexp.MustCompile(`[^[:graph:]]`)
	isSpace := regexp.MustCompile(`[[:space:]]`)
	for s, _ := range longOpts {
		if notGraph.MatchString(s) || isSpace.MatchString(s) {
			return nil, fmt.Errorf("Invalid long option: %s", s)
		}
	}
	parser.longOpts = longOpts

	return &parser, nil
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
		// defined canidate that is the "best" match must always
		// be used.
		if len(option.Name) >= len(best.Name) {
			best = option
		}
	}

	if best.Name != "" {
		return args, best, nil
	}
	return args, Option{}, p.optError("unknown option: " + name)
}

func (p *Parser) findShortOpt(c byte, word string, args []string) ([]string, string, Option, error) {
	slog.Debug("findShortOpt", "c", string(c), "word", word, "args", args)

	// POSIX disallows `-` as a short-opt option.
	if c == '-' {
		return args, word, Option{}, p.optError("invalid option: " + string(c))
	}

	// We have to itterate the shortOpts in order to support case
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

	return args, word, Option{}, p.optError("unknown option: " + string(c))
}

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
				p.Args, option, err = p.findLongOpt(p.Args[0][2:], remainingArgs)
				if !yield(option, err) {
					return
				}

			case strings.HasPrefix(p.Args[0], "-"):
				slog.Debug("Options", "prefix", "-")
				if p.config.longOptsOnly {
					var remainingArgs []string
					if len(p.Args) > 1 {
						remainingArgs = p.Args[1:]
					}
					p.Args, option, err = p.findLongOpt(p.Args[0][1:], remainingArgs)
					if !yield(option, err) {
						return
					}
					continue
				}

				// iterate over each character in the word looking
				// for short options
				for word := p.Args[0][1:]; len(word) > 0; {
					slog.Debug("Options", "word", word)
					var remainingArgs []string
					if len(p.Args) > 1 {
						remainingArgs = p.Args[1:]
					}
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

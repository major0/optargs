package goarg

import (
	"fmt"
	"reflect"

	"github.com/major0/optargs"
)

// CoreIntegration coordinates flag building, parser creation, subcommand
// dispatch, and post-parse processing. It delegates all work to focused
// components: FlagBuilder, PostProcessor, and subcommand methods.
type CoreIntegration struct {
	metadata    *StructMetadata
	config      Config
	setFields   map[int]bool // tracks field indices explicitly set during parsing
	flagBuilder *FlagBuilder
}

// fieldByMeta returns the reflect.Value for a field using the cached index
// when available (FieldIndex >= 0), falling back to FieldByName for fields
// inherited from embedded structs (FieldIndex == -1).
func fieldByMeta(destValue reflect.Value, field *FieldMetadata) reflect.Value {
	if field.FieldIndex >= 0 {
		return destValue.Field(field.FieldIndex)
	}
	return destValue.FieldByName(field.Name)
}

// formatDefault returns the display string for a field's default value.
func formatDefault(field *FieldMetadata) string {
	if field.Default == nil {
		return ""
	}
	return fmt.Sprintf("%v", field.Default)
}

// CreateParserWithHandlers builds an OptArgs parser with Handle callbacks
// wired to each flag. Delegates flag building to FlagBuilder.
func (ci *CoreIntegration) CreateParserWithHandlers(args []string, destValue reflect.Value) (*optargs.Parser, error) {
	ci.flagBuilder = &FlagBuilder{metadata: ci.metadata, config: ci.config}
	shortOpts, longOpts, err := ci.flagBuilder.Build(destValue)
	if err != nil {
		return nil, fmt.Errorf("failed to build flags: %w", err)
	}
	ci.setFields = ci.flagBuilder.SetFields()

	// Register builtin -h/--help flag (returns ErrHelp when parsed).
	helpFlag := &optargs.Flag{
		Name:   "h",
		HasArg: optargs.NoArgument,
		Help:   "display this help and exit",
		Handle: func(_, _ string) error { return ErrHelp },
	}
	helpLong := &optargs.Flag{
		Name:   "help",
		HasArg: optargs.NoArgument,
		Help:   "display this help and exit",
		Peer:   helpFlag,
		Handle: func(_, _ string) error { return ErrHelp },
	}
	helpFlag.Peer = helpLong
	if shortOpts['h'] == nil {
		shortOpts['h'] = helpFlag
	}
	if longOpts["help"] == nil {
		longOpts["help"] = helpLong
	}

	// Register builtin --version flag if version is configured.
	if ci.config.Version != "" {
		if longOpts["version"] == nil {
			longOpts["version"] = &optargs.Flag{
				Name:   "version",
				HasArg: optargs.NoArgument,
				Help:   "display version and exit",
				Handle: func(_, _ string) error { return ErrVersion },
			}
		}
	}

	parser, err := optargs.NewParserWithCaseInsensitiveCommands(shortOpts, longOpts, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create OptArgs parser: %w", err)
	}

	if ci.config.StrictSubcommands {
		parser.SetStrictSubcommands(true)
	}

	return parser, nil
}

// PostParse delegates to PostProcessor for positional args, env vars, defaults, and validation.
func (ci *CoreIntegration) PostParse(coreParser *optargs.Parser, destValue reflect.Value) error {
	pp := &PostProcessor{
		metadata:  ci.metadata,
		config:    ci.config,
		setFields: ci.setFields,
	}
	pp.buildPositionalArgs()
	return pp.Process(coreParser, destValue)
}

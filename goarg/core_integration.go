package goarg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/major0/optargs"
)

// CoreIntegration handles direct translation to OptArgs Core.
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
// wired to each flag in a single pass. It creates the parser with
// case-insensitive commands and prepares positional arg metadata.
// It does NOT register subcommands.
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

// findSubcommandField finds the struct field for a subcommand by name
// (case-insensitive).
func (ci *CoreIntegration) findSubcommandField(destValue reflect.Value, name string) (reflect.Value, *StructMetadata, error) {
	if idx, ok := ci.metadata.SubcommandFieldIdx[name]; ok {
		subMeta := ci.metadata.Subcommands[name]
		if subMeta == nil {
			return reflect.Value{}, nil, fmt.Errorf("subcommand metadata not found for %s", name)
		}
		fv := destValue.Field(idx)
		if !fv.IsValid() {
			return reflect.Value{}, nil, fmt.Errorf("subcommand field not found for %s", name)
		}
		return fv, subMeta, nil
	}

	for cmdName, idx := range ci.metadata.SubcommandFieldIdx {
		if strings.EqualFold(cmdName, name) {
			subMeta := ci.metadata.Subcommands[cmdName]
			if subMeta == nil {
				return reflect.Value{}, nil, fmt.Errorf("subcommand metadata not found for %s", cmdName)
			}
			fv := destValue.Field(idx)
			if !fv.IsValid() {
				return reflect.Value{}, nil, fmt.Errorf("subcommand field not found for %s", cmdName)
			}
			return fv, subMeta, nil
		}
	}

	return reflect.Value{}, nil, fmt.Errorf("unknown subcommand: %s", name)
}

// RegisterSubcommands registers all subcommands from metadata with the core parser.
func (ci *CoreIntegration) RegisterSubcommands(coreParser *optargs.Parser, destValue reflect.Value) error {
	for name, subMeta := range ci.metadata.Subcommands {
		fieldValue, _, err := ci.findSubcommandField(destValue, name)
		if err != nil {
			return fmt.Errorf("failed to find subcommand field for %s: %w", name, err)
		}

		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			fieldValue = fieldValue.Elem()
		}

		child := &CoreIntegration{
			metadata: subMeta,
			config:   ci.config,
		}

		childParser, err := child.CreateParserWithHandlers([]string{}, fieldValue)
		if err != nil {
			return fmt.Errorf("failed to create parser for subcommand %s: %w", name, err)
		}

		coreParser.AddCmd(name, childParser)

		if help, ok := ci.metadata.SubcommandHelp[name]; ok {
			childParser.Description = help
		}

		if err := child.RegisterSubcommands(childParser, fieldValue); err != nil {
			return fmt.Errorf("failed to register nested subcommands for %s: %w", name, err)
		}
	}
	return nil
}

// dispatchSubcommand handles subcommand invocation and recursive dispatch.
func (ci *CoreIntegration) dispatchSubcommand(childParser *optargs.Parser, invokedName string, destValue reflect.Value, p *Parser) error {
	fieldValue, subMeta, err := ci.findSubcommandField(destValue, invokedName)
	if err != nil {
		return p.translateError(err, invokedName)
	}

	for _, err := range childParser.Options() {
		if err != nil {
			return p.translateError(err, "")
		}
	}

	subDestValue := fieldValue.Elem()
	childCI := &CoreIntegration{
		metadata:  subMeta,
		config:    ci.config,
		setFields: make(map[int]bool),
	}
	if err := childCI.PostParse(childParser, subDestValue); err != nil {
		return p.translateError(err, "")
	}

	nestedName, nestedParser := childParser.ActiveCommand()
	if nestedName != "" && nestedParser != nil {
		return childCI.dispatchSubcommand(nestedParser, nestedName, subDestValue, p)
	}

	return nil
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

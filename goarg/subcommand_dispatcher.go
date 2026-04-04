package goarg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/major0/optargs"
)

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

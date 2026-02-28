package optargs

import (
	"fmt"
	"strings"
)

// CommandRegistry manages subcommands for a parser using a simple map
type CommandRegistry map[string]*Parser

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() CommandRegistry {
	return make(map[string]*Parser)
}

// AddCmd registers a new subcommand with the parser
// Returns the registered parser for chaining
func (cr CommandRegistry) AddCmd(name string, parser *Parser) *Parser {
	cr[name] = parser
	return parser
}

// AddAlias creates an alias for an existing command
func (cr CommandRegistry) AddAlias(alias, existingCommand string) error {
	parser, exists := cr[existingCommand]
	if !exists {
		return fmt.Errorf("command %s does not exist", existingCommand)
	}
	cr[alias] = parser
	return nil
}

// GetCommand retrieves a parser by command name (exact match).
func (cr CommandRegistry) GetCommand(name string) (*Parser, bool) {
	parser, exists := cr[name]
	return parser, exists
}

// getCommandFold retrieves a parser by command name with case-insensitive matching.
func (cr CommandRegistry) getCommandFold(name string) (*Parser, bool) {
	// Try exact match first (fast path).
	if parser, exists := cr[name]; exists {
		return parser, true
	}
	for cmdName, parser := range cr {
		if strings.EqualFold(cmdName, name) {
			return parser, true
		}
	}
	return nil, false
}

// ListCommands returns all command mappings
func (cr CommandRegistry) ListCommands() map[string]*Parser {
	return map[string]*Parser(cr)
}

// ExecuteCommand finds and prepares a command for execution.
func (cr CommandRegistry) ExecuteCommand(name string, args []string) (*Parser, error) {
	parser, exists := cr[name]
	if !exists {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	if parser == nil {
		return nil, fmt.Errorf("command %s has no parser", name)
	}
	parser.Args = args
	parser.nonOpts = []string{}
	return parser, nil
}

// executeCommandFold finds and prepares a command for execution with case-insensitive matching.
func (cr CommandRegistry) executeCommandFold(name string, args []string) (*Parser, error) {
	parser, exists := cr.getCommandFold(name)
	if !exists {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	if parser == nil {
		return nil, fmt.Errorf("command %s has no parser", name)
	}
	parser.Args = args
	parser.nonOpts = []string{}
	return parser, nil
}

// ExecuteCommandCaseInsensitive finds and prepares a command for execution
// with optional case-insensitive matching.
func (cr CommandRegistry) ExecuteCommandCaseInsensitive(name string, args []string, caseIgnore bool) (*Parser, error) {
	if !caseIgnore {
		return cr.ExecuteCommand(name, args)
	}
	return cr.executeCommandFold(name, args)
}

// HasCommands returns true if any commands are registered
func (cr CommandRegistry) HasCommands() bool {
	return len(cr) > 0
}

// GetAliases returns all aliases for a given parser
func (cr CommandRegistry) GetAliases(targetParser *Parser) []string {
	var aliases []string
	for name, parser := range cr {
		if parser == targetParser {
			aliases = append(aliases, name)
		}
	}
	return aliases
}

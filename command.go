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

// getCommand retrieves a parser by command name, optionally case-insensitive.
func (cr CommandRegistry) getCommand(name string, caseIgnore bool) (*Parser, bool) {
	if !caseIgnore {
		return cr.GetCommand(name)
	}
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

// prepareCommand validates and prepares a looked-up command parser for execution.
func prepareCommand(name string, parser *Parser, exists bool, args []string) (*Parser, error) {
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

// ExecuteCommand finds and prepares a command for execution.
func (cr CommandRegistry) ExecuteCommand(name string, args []string) (*Parser, error) {
	parser, exists := cr[name]
	return prepareCommand(name, parser, exists, args)
}

// executeCommand finds and prepares a command for execution with optional case-insensitive matching.
func (cr CommandRegistry) executeCommand(name string, args []string, caseIgnore bool) (*Parser, error) {
	parser, exists := cr.getCommand(name, caseIgnore)
	return prepareCommand(name, parser, exists, args)
}

// ExecuteCommandCaseInsensitive finds and prepares a command for execution
// with optional case-insensitive matching.
func (cr CommandRegistry) ExecuteCommandCaseInsensitive(name string, args []string, caseIgnore bool) (*Parser, error) {
	return cr.executeCommand(name, args, caseIgnore)
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

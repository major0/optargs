package goarg

import (
	"fmt"
	"testing"
)

func TestDebugInheritance(t *testing.T) {
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Debug   bool       `arg:"-d,--debug" help:"enable debug output"`
		Server  *ServerCmd `arg:"subcommand:server"`
	}

	// Test what options are available in each parser
	var cmd RootCmd
	parser, err := NewParser(Config{}, &cmd)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	// Check root parser options
	fmt.Printf("Root parser metadata fields:\n")
	for _, field := range parser.metadata.Fields {
		fmt.Printf("  Field: %s, Short: %s, Long: %s, Positional: %v, IsSubcommand: %v\n", 
			field.Name, field.Short, field.Long, field.Positional, field.IsSubcommand)
	}

	// Check subcommand metadata
	if serverMeta, exists := parser.metadata.Subcommands["server"]; exists {
		fmt.Printf("Server subcommand metadata fields:\n")
		for _, field := range serverMeta.Fields {
			fmt.Printf("  Field: %s, Short: %s, Long: %s, Positional: %v, IsSubcommand: %v\n", 
				field.Name, field.Short, field.Long, field.Positional, field.IsSubcommand)
		}
	} else {
		t.Errorf("Server subcommand metadata not found")
	}

	// Try to parse the problematic case
	fmt.Printf("\nTrying to parse: server -vp 9000\n")
	err = ParseArgs(&cmd, []string{"server", "-vp", "9000"})
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
	} else {
		fmt.Printf("Parse successful: Verbose=%v, Port=%d\n", cmd.Verbose, cmd.Server.Port)
	}
}
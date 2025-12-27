package goarg

import (
	"testing"
	
	"github.com/major0/optargs"
)

func TestCommandSystemIntegration(t *testing.T) {
	// Test struct with subcommands and option inheritance
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type ClientCmd struct {
		URL     string `arg:"-u,--url" help:"client URL"`
		Timeout int    `arg:"-t,--timeout" default:"30" help:"timeout in seconds"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Debug   bool       `arg:"-d,--debug" help:"enable debug output"`
		Server  *ServerCmd `arg:"subcommand:server"`
		Client  *ClientCmd `arg:"subcommand:client"`
	}

	t.Run("BasicOptionParsing", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"-v", "-d"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if !cmd.Verbose {
			t.Errorf("Expected Verbose=true, got %v", cmd.Verbose)
		}
		if !cmd.Debug {
			t.Errorf("Expected Debug=true, got %v", cmd.Debug)
		}
	})

	t.Run("SubcommandMetadataParsing", func(t *testing.T) {
		var cmd RootCmd
		parser := &TagParser{}
		metadata, err := parser.ParseStruct(&cmd)
		if err != nil {
			t.Fatalf("ParseStruct() unexpected error: %v", err)
		}

		// Check that we have two subcommands
		if len(metadata.Subcommands) != 2 {
			t.Errorf("Expected 2 subcommands, got %d", len(metadata.Subcommands))
		}

		if _, exists := metadata.Subcommands["server"]; !exists {
			t.Errorf("Expected 'server' subcommand to exist")
		}

		if _, exists := metadata.Subcommands["client"]; !exists {
			t.Errorf("Expected 'client' subcommand to exist")
		}

		// Check server subcommand fields
		serverMeta := metadata.Subcommands["server"]
		if len(serverMeta.Fields) != 2 {
			t.Errorf("Expected server subcommand to have 2 fields, got %d", len(serverMeta.Fields))
		}
	})

	t.Run("CoreIntegrationCreatesCommands", func(t *testing.T) {
		var cmd RootCmd
		parser := &TagParser{}
		metadata, err := parser.ParseStruct(&cmd)
		if err != nil {
			t.Fatalf("ParseStruct() unexpected error: %v", err)
		}

		// Create core integration
		coreIntegration := &CoreIntegration{
			metadata:    metadata,
			shortOpts:   make(map[byte]*optargs.Flag),
			longOpts:    make(map[string]*optargs.Flag),
			positionals: []PositionalArg{},
		}

		// Create parser with command support
		coreParser, err := coreIntegration.CreateParser([]string{})
		if err != nil {
			t.Fatalf("CreateParser() unexpected error: %v", err)
		}

		// Verify commands were registered
		if !coreParser.HasCommands() {
			t.Error("Expected parser to have commands registered")
		}

		// Check specific commands
		if _, exists := coreParser.GetCommand("server"); !exists {
			t.Error("Expected 'server' command to be registered")
		}

		if _, exists := coreParser.GetCommand("client"); !exists {
			t.Error("Expected 'client' command to be registered")
		}
	})

	// TODO: Add test for actual subcommand parsing with option inheritance
	// This will require the full parsing flow to work correctly
	t.Run("SubcommandParsingWithInheritance", func(t *testing.T) {
		t.Skip("Subcommand parsing with inheritance - implementation in progress")
		
		// This test should verify:
		// 1. `mycmd server --verbose -p 9000` works
		// 2. --verbose is handled by root parser
		// 3. -p is handled by server subcommand
		// 4. Both options are properly set in the result struct
	})
}
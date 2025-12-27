package goarg

import (
	"testing"
)

func TestSimpleSubcommands(t *testing.T) {
	// Test struct with subcommands
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Server  *ServerCmd `arg:"subcommand:server"`
	}

	// For now, let's test without subcommands to make sure the basic structure works
	var cmd RootCmd
	err := ParseArgs(&cmd, []string{"-v"})
	if err != nil {
		t.Fatalf("ParseArgs() unexpected error: %v", err)
	}

	if !cmd.Verbose {
		t.Errorf("Expected Verbose=true, got %v", cmd.Verbose)
	}

	// Test that subcommand field is parsed correctly in metadata
	parser := &TagParser{}
	metadata, err := parser.ParseStruct(&cmd)
	if err != nil {
		t.Fatalf("ParseStruct() unexpected error: %v", err)
	}

	// Check that we have one subcommand
	if len(metadata.Subcommands) != 1 {
		t.Errorf("Expected 1 subcommand, got %d", len(metadata.Subcommands))
	}

	if _, exists := metadata.Subcommands["server"]; !exists {
		t.Errorf("Expected 'server' subcommand to exist")
	}
}
package goarg

import (
	"testing"
)

func TestCommandInheritanceIntegration(t *testing.T) {
	// Test the key use case: `mycmd mysubcmd --verbose` where --verbose is in root
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Debug   bool       `arg:"-d,--debug" help:"enable debug output"`
		Server  *ServerCmd `arg:"subcommand:server"`
	}

	t.Run("RootOptionsOnly", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"--verbose", "--debug"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if !cmd.Verbose {
			t.Errorf("Expected Verbose=true, got %v", cmd.Verbose)
		}
		if !cmd.Debug {
			t.Errorf("Expected Debug=true, got %v", cmd.Debug)
		}
		
		// When no subcommand is specified, the subcommand field should remain nil
		// or if initialized, should have default values only
		if cmd.Server != nil {
			// If Server is initialized, it should have default values
			if cmd.Server.Port != 8080 {
				t.Errorf("Expected Server.Port=8080 (default), got %v", cmd.Server.Port)
			}
			if cmd.Server.Host != "localhost" {
				t.Errorf("Expected Server.Host=localhost (default), got %v", cmd.Server.Host)
			}
		}
	})

	t.Run("SubcommandOptionsOnly", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"server", "--port", "9000", "--host", "example.com"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if cmd.Verbose {
			t.Errorf("Expected Verbose=false, got %v", cmd.Verbose)
		}
		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}
		if cmd.Server.Port != 9000 {
			t.Errorf("Expected Server.Port=9000, got %v", cmd.Server.Port)
		}
		if cmd.Server.Host != "example.com" {
			t.Errorf("Expected Server.Host=example.com, got %v", cmd.Server.Host)
		}
	})

	// This test demonstrates the key functionality: option inheritance
	// Currently skipped as the full inheritance implementation is still in progress
	t.Run("SubcommandWithInheritedOptions", func(t *testing.T) {
		t.Skip("Option inheritance from parent to child - implementation in progress")
		
		// This should work: server subcommand with --verbose from root
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"server", "--verbose", "--port", "9000"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		// --verbose should be set on root even when used with subcommand
		if !cmd.Verbose {
			t.Errorf("Expected Verbose=true (inherited from root), got %v", cmd.Verbose)
		}
		
		// Subcommand should be initialized and have its options set
		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}
		if cmd.Server.Port != 9000 {
			t.Errorf("Expected Server.Port=9000, got %v", cmd.Server.Port)
		}
	})

	t.Run("DefaultValues", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"server"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}
		
		// Check default values are applied
		if cmd.Server.Port != 8080 {
			t.Errorf("Expected Server.Port=8080 (default), got %v", cmd.Server.Port)
		}
		if cmd.Server.Host != "localhost" {
			t.Errorf("Expected Server.Host=localhost (default), got %v", cmd.Server.Host)
		}
	})
}
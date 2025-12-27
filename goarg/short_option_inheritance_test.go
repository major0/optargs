package goarg

import (
	"testing"
)

func TestShortOptionInheritance(t *testing.T) {
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Debug   bool       `arg:"-d,--debug" help:"enable debug output"`
		Server  *ServerCmd `arg:"subcommand:server"`
	}

	t.Run("SingleShortOption", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"server", "-v", "--port", "9000"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if !cmd.Verbose {
			t.Errorf("Expected Verbose=true (inherited from root), got %v", cmd.Verbose)
		}
		
		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}
		if cmd.Server.Port != 9000 {
			t.Errorf("Expected Server.Port=9000, got %v", cmd.Server.Port)
		}
	})

	t.Run("CompactedShortOptions", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"server", "-vd", "--port", "9000"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if !cmd.Verbose {
			t.Errorf("Expected Verbose=true (inherited from root), got %v", cmd.Verbose)
		}
		if !cmd.Debug {
			t.Errorf("Expected Debug=true (inherited from root), got %v", cmd.Debug)
		}
		
		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}
		if cmd.Server.Port != 9000 {
			t.Errorf("Expected Server.Port=9000, got %v", cmd.Server.Port)
		}
	})

	t.Run("CompactedWithSubcommandOptions", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"server", "-vp", "9000"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if !cmd.Verbose {
			t.Errorf("Expected Verbose=true (inherited from root), got %v", cmd.Verbose)
		}
		
		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}
		if cmd.Server.Port != 9000 {
			t.Errorf("Expected Server.Port=9000, got %v", cmd.Server.Port)
		}
	})
}
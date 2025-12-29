package goarg

import (
	"testing"
)

func TestCaseInsensitiveCommands(t *testing.T) {
	// Test struct with subcommands
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type ClientCmd struct {
		URL string `arg:"-u,--url" help:"client URL"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Server  *ServerCmd `arg:"subcommand:server"`
		Client  *ClientCmd `arg:"subcommand:client"`
	}

	t.Run("ExactCaseMatch", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"server", "--port", "9000"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}
		if cmd.Server.Port != 9000 {
			t.Errorf("Expected Server.Port=9000, got %v", cmd.Server.Port)
		}
	})

	t.Run("UpperCaseCommand", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"SERVER", "--port", "9000"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}
		if cmd.Server.Port != 9000 {
			t.Errorf("Expected Server.Port=9000, got %v", cmd.Server.Port)
		}
	})

	t.Run("MixedCaseCommand", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"SeRvEr", "--port", "9000"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}
		if cmd.Server.Port != 9000 {
			t.Errorf("Expected Server.Port=9000, got %v", cmd.Server.Port)
		}
	})

	t.Run("LowerCaseCommand", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"client", "--url", "http://example.com"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if cmd.Client == nil {
			t.Fatalf("Expected Client to be initialized")
		}
		if cmd.Client.URL != "http://example.com" {
			t.Errorf("Expected Client.URL=http://example.com, got %v", cmd.Client.URL)
		}
	})

	t.Run("UpperCaseClientCommand", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"CLIENT", "--url", "http://example.com"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		if cmd.Client == nil {
			t.Fatalf("Expected Client to be initialized")
		}
		if cmd.Client.URL != "http://example.com" {
			t.Errorf("Expected Client.URL=http://example.com, got %v", cmd.Client.URL)
		}
	})

	t.Run("InvalidCommand", func(t *testing.T) {
		var cmd RootCmd
		err := ParseArgs(&cmd, []string{"invalid", "--port", "9000"})
		if err == nil {
			t.Fatalf("Expected error for invalid command")
		}
	})
}

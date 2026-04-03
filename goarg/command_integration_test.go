package goarg

import (
	"testing"
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

	// Subcommand parsing with option inheritance requires the full parsing flow.
	t.Run("SubcommandParsingWithInheritance", func(t *testing.T) {
		t.Skip("Subcommand parsing with inheritance - implementation in progress")

		// This test should verify:
		// 1. `mycmd server --verbose -p 9000` works
		// 2. --verbose is handled by root parser
		// 3. -p is handled by server subcommand
		// 4. Both options are properly set in the result struct
	})
}

// subcommandDetectionTests verifies that the migrated subcommand detection
// (using core ActiveCommand()) produces correct struct field values.
var subcommandDetectionTests = []struct {
	name  string
	args  []string
	check func(t *testing.T, cmd any)
}{
	{
		name: "single_subcommand_with_options",
		args: []string{"server", "--port", "9090"},
		check: func(t *testing.T, cmd any) {
			t.Helper()
			r := cmd.(*struct {
				Server *struct {
					Port int `arg:"--port"`
				} `arg:"subcommand:server"`
				Client *struct {
					URL string `arg:"--url"`
				} `arg:"subcommand:client"`
			})
			if r.Server == nil {
				t.Fatal("Server should be non-nil")
			}
			if r.Server.Port != 9090 {
				t.Errorf("Port = %d, want 9090", r.Server.Port)
			}
			if r.Client != nil {
				t.Error("Client should be nil")
			}
		},
	},
	{
		name: "no_subcommand_invoked",
		args: []string{},
		check: func(t *testing.T, cmd any) {
			t.Helper()
			r := cmd.(*struct {
				Server *struct {
					Port int `arg:"--port"`
				} `arg:"subcommand:server"`
				Client *struct {
					URL string `arg:"--url"`
				} `arg:"subcommand:client"`
			})
			if r.Server != nil {
				t.Error("Server should be nil")
			}
			if r.Client != nil {
				t.Error("Client should be nil")
			}
		},
	},
	{
		name: "second_subcommand_invoked",
		args: []string{"client", "--url", "http://example.com"},
		check: func(t *testing.T, cmd any) {
			t.Helper()
			r := cmd.(*struct {
				Server *struct {
					Port int `arg:"--port"`
				} `arg:"subcommand:server"`
				Client *struct {
					URL string `arg:"--url"`
				} `arg:"subcommand:client"`
			})
			if r.Server != nil {
				t.Error("Server should be nil")
			}
			if r.Client == nil {
				t.Fatal("Client should be non-nil")
			}
			if r.Client.URL != "http://example.com" {
				t.Errorf("URL = %q, want %q", r.Client.URL, "http://example.com")
			}
		},
	},
	{
		name: "subcommand_with_defaults",
		args: []string{"server"},
		check: func(t *testing.T, cmd any) {
			t.Helper()
			r := cmd.(*struct {
				Server *struct {
					Port int `arg:"--port" default:"8080"`
				} `arg:"subcommand:server"`
				Client *struct {
					URL string `arg:"--url"`
				} `arg:"subcommand:client"`
			})
			if r.Server == nil {
				t.Fatal("Server should be non-nil")
			}
			if r.Server.Port != 8080 {
				t.Errorf("Port = %d, want 8080 (default)", r.Server.Port)
			}
		},
	},
}

func TestSubcommandDetection(t *testing.T) {
	for _, tt := range subcommandDetectionTests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &struct {
				Server *struct {
					Port int `arg:"--port"`
				} `arg:"subcommand:server"`
				Client *struct {
					URL string `arg:"--url"`
				} `arg:"subcommand:client"`
			}{}

			// For the defaults test, use the struct with defaults
			if tt.name == "subcommand_with_defaults" {
				cmdD := &struct {
					Server *struct {
						Port int `arg:"--port" default:"8080"`
					} `arg:"subcommand:server"`
					Client *struct {
						URL string `arg:"--url"`
					} `arg:"subcommand:client"`
				}{}
				if err := ParseArgs(cmdD, tt.args); err != nil {
					t.Fatalf("ParseArgs: %v", err)
				}
				tt.check(t, cmdD)
				return
			}

			if err := ParseArgs(cmd, tt.args); err != nil {
				t.Fatalf("ParseArgs: %v", err)
			}
			tt.check(t, cmd)
		})
	}
}

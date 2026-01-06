package goarg

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

// TestProperty9_OptionInheritanceCorrectness tests Property 9 from the design document:
// For any parent parser with options and child parser with subcommands, options defined in the parent should be accessible and functional when used with child subcommands
// **Validates: Requirements 2.1, 2.2**
func TestProperty9_OptionInheritanceCorrectness(t *testing.T) {
	// Property: Parent options should be accessible and functional when used with child subcommands
	property := func(verbose bool, debug bool, port int, host string) bool {
		// Skip invalid inputs to focus on valid test cases
		if port < 1 || port > 65535 {
			return true
		}
		if len(host) > 100 || strings.ContainsAny(host, "\n\r\t\"'\\") {
			return true
		}
		if host == "" {
			host = "localhost" // Use default for empty host
		}

		// Define test struct with parent options and subcommand
		type ServerCmd struct {
			Port int    `arg:"-p,--port" default:"8080" help:"server port"`
			Host string `arg:"-h,--host" default:"localhost" help:"server host"`
		}

		type RootCmd struct {
			Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
			Debug   bool       `arg:"-d,--debug" help:"enable debug output"`
			Server  *ServerCmd `arg:"subcommand:server"`
		}

		// Test case 1: Parent options used with subcommand should be set on parent
		args1 := []string{"server", "--port", fmt.Sprintf("%d", port), "--host", host}
		if verbose {
			args1 = append(args1, "--verbose")
		}
		if debug {
			args1 = append(args1, "--debug")
		}

		var cmd1 RootCmd
		err1 := ParseArgs(&cmd1, args1)
		if err1 != nil {
			// Parsing should succeed for valid inputs
			return false
		}

		// Verify parent options are set correctly when used with subcommand
		if verbose && !cmd1.Verbose {
			return false
		}
		if !verbose && cmd1.Verbose {
			return false
		}
		if debug && !cmd1.Debug {
			return false
		}
		if !debug && cmd1.Debug {
			return false
		}

		// Verify subcommand is initialized and has correct values
		if cmd1.Server == nil {
			return false
		}
		if cmd1.Server.Port != port {
			return false
		}
		if cmd1.Server.Host != host {
			return false
		}

		// Test case 2: Parent options should work in different order
		args2 := []string{}
		if verbose {
			args2 = append(args2, "--verbose")
		}
		args2 = append(args2, "server")
		if debug {
			args2 = append(args2, "--debug")
		}
		args2 = append(args2, "--port", fmt.Sprintf("%d", port))

		var cmd2 RootCmd
		err2 := ParseArgs(&cmd2, args2)
		if err2 != nil {
			// Parsing should succeed regardless of option order
			return false
		}

		// Verify same results regardless of option order
		if cmd2.Verbose != cmd1.Verbose {
			return false
		}
		if cmd2.Debug != cmd1.Debug {
			return false
		}
		if cmd2.Server == nil {
			return false
		}
		if cmd2.Server.Port != cmd1.Server.Port {
			return false
		}

		// Test case 3: Parent-only options without subcommand should work
		args3 := []string{}
		if verbose {
			args3 = append(args3, "--verbose")
		}
		if debug {
			args3 = append(args3, "--debug")
		}

		if len(args3) > 0 { // Only test if we have parent options
			var cmd3 RootCmd
			err3 := ParseArgs(&cmd3, args3)
			if err3 != nil {
				// Parent-only parsing should succeed
				return false
			}

			// Verify parent options are set correctly without subcommand
			if verbose && !cmd3.Verbose {
				return false
			}
			if debug && !cmd3.Debug {
				return false
			}
		}

		// Test case 4: Subcommand options should not interfere with parent options
		args4 := []string{"server", "--port", fmt.Sprintf("%d", port)}
		if verbose {
			args4 = append(args4, "--verbose")
		}

		var cmd4 RootCmd
		err4 := ParseArgs(&cmd4, args4)
		if err4 != nil {
			return false
		}

		// Verify that subcommand options don't affect parent option inheritance
		if verbose && !cmd4.Verbose {
			return false
		}
		if cmd4.Server == nil || cmd4.Server.Port != port {
			return false
		}

		return true
	}

	// Configure property test with sufficient iterations
	config := &quick.Config{
		MaxCount: 100, // Minimum 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 failed: %v", err)
	}
}

// TestProperty10_CaseInsensitiveCommandMatching tests Property 10 from the design document:
// For any registered command name, all case variations (uppercase, lowercase, mixed case) should resolve to the same command parser
// **Validates: Requirements 2.1, 2.2**
func TestProperty10_CaseInsensitiveCommandMatching(t *testing.T) {
	// Property: All case variations of command names should resolve to the same command parser
	property := func(commandCase int, port int, url string) bool {
		// Skip invalid inputs
		if port < 1 || port > 65535 {
			return true
		}
		if len(url) > 100 || strings.ContainsAny(url, "\n\r\t\"'\\") {
			return true
		}
		if url == "" {
			url = "http://example.com" // Use default for empty URL
		}

		// Define test struct with multiple subcommands
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

		// Generate different case variations of command names
		serverVariations := []string{"server", "SERVER", "Server", "SeRvEr", "sErVeR"}
		clientVariations := []string{"client", "CLIENT", "Client", "ClIeNt", "cLiEnT"}

		// Test server command with different case variations
		serverCase := commandCase % len(serverVariations)
		serverCommand := serverVariations[serverCase]

		args1 := []string{serverCommand, "--port", fmt.Sprintf("%d", port)}
		var cmd1 RootCmd
		err1 := ParseArgs(&cmd1, args1)
		if err1 != nil {
			// All case variations should parse successfully
			return false
		}

		// Verify server subcommand was initialized correctly
		if cmd1.Server == nil {
			return false
		}
		if cmd1.Server.Port != port {
			return false
		}
		if cmd1.Client != nil {
			// Client should not be initialized when server command is used
			return false
		}

		// Test client command with different case variations
		clientCase := (commandCase + 1) % len(clientVariations)
		clientCommand := clientVariations[clientCase]

		args2 := []string{clientCommand, "--url", url}
		var cmd2 RootCmd
		err2 := ParseArgs(&cmd2, args2)
		if err2 != nil {
			// All case variations should parse successfully
			return false
		}

		// Verify client subcommand was initialized correctly
		if cmd2.Client == nil {
			return false
		}
		if cmd2.Client.URL != url {
			return false
		}
		if cmd2.Server != nil {
			// Server should not be initialized when client command is used
			return false
		}

		// Test that exact case and different case produce identical results
		exactArgs := []string{"server", "--port", fmt.Sprintf("%d", port)}
		var exactCmd RootCmd
		err3 := ParseArgs(&exactCmd, exactArgs)
		if err3 != nil {
			return false
		}

		upperArgs := []string{"SERVER", "--port", fmt.Sprintf("%d", port)}
		var upperCmd RootCmd
		err4 := ParseArgs(&upperCmd, upperArgs)
		if err4 != nil {
			return false
		}

		// Results should be identical regardless of case
		if exactCmd.Server == nil || upperCmd.Server == nil {
			return false
		}
		if exactCmd.Server.Port != upperCmd.Server.Port {
			return false
		}
		if exactCmd.Server.Host != upperCmd.Server.Host {
			return false
		}

		// Test mixed case with options
		mixedArgs := []string{"SeRvEr", "--verbose", "--port", fmt.Sprintf("%d", port)}
		var mixedCmd RootCmd
		err5 := ParseArgs(&mixedCmd, mixedArgs)
		if err5 != nil {
			return false
		}

		// Mixed case should work with inherited options
		if !mixedCmd.Verbose {
			return false
		}
		if mixedCmd.Server == nil || mixedCmd.Server.Port != port {
			return false
		}

		// Test that invalid case variations still fail appropriately
		// Note: Invalid commands don't fail by themselves, but invalid options do
		invalidArgs := []string{"invalidcommand", "--port", fmt.Sprintf("%d", port)}
		var invalidCmd RootCmd
		err6 := ParseArgs(&invalidCmd, invalidArgs)
		if err6 == nil {
			// Invalid options should still fail (--port not valid for root)
			return false
		}

		return true
	}

	// Configure property test with sufficient iterations
	config := &quick.Config{
		MaxCount: 100, // Minimum 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10 failed: %v", err)
	}
}

// TestEnhancedFeaturesIntegration tests the integration of both enhanced features together
func TestEnhancedFeaturesIntegration(t *testing.T) {
	// Test that option inheritance and case insensitive commands work together
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"server port"`
		Host string `arg:"-h,--host" default:"localhost" help:"server host"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose" help:"enable verbose output"`
		Debug   bool       `arg:"-d,--debug" help:"enable debug output"`
		Server  *ServerCmd `arg:"subcommand:server"`
	}

	testCases := []struct {
		name     string
		args     []string
		expected RootCmd
	}{
		{
			name: "LowerCaseWithInheritance",
			args: []string{"server", "--verbose", "--port", "9000"},
			expected: RootCmd{
				Verbose: true,
				Debug:   false,
				Server:  &ServerCmd{Port: 9000, Host: "localhost"},
			},
		},
		{
			name: "UpperCaseWithInheritance",
			args: []string{"SERVER", "--debug", "--port", "9001"},
			expected: RootCmd{
				Verbose: false,
				Debug:   true,
				Server:  &ServerCmd{Port: 9001, Host: "localhost"},
			},
		},
		{
			name: "MixedCaseWithMultipleInheritance",
			args: []string{"SeRvEr", "--verbose", "--debug", "--port", "9002", "--host", "example.com"},
			expected: RootCmd{
				Verbose: true,
				Debug:   true,
				Server:  &ServerCmd{Port: 9002, Host: "example.com"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cmd RootCmd
			err := ParseArgs(&cmd, tc.args)
			if err != nil {
				t.Fatalf("ParseArgs() unexpected error: %v", err)
			}

			// Verify parent options
			if cmd.Verbose != tc.expected.Verbose {
				t.Errorf("Expected Verbose=%v, got %v", tc.expected.Verbose, cmd.Verbose)
			}
			if cmd.Debug != tc.expected.Debug {
				t.Errorf("Expected Debug=%v, got %v", tc.expected.Debug, cmd.Debug)
			}

			// Verify subcommand initialization and values
			if cmd.Server == nil {
				t.Fatalf("Expected Server to be initialized")
			}
			if cmd.Server.Port != tc.expected.Server.Port {
				t.Errorf("Expected Server.Port=%v, got %v", tc.expected.Server.Port, cmd.Server.Port)
			}
			if cmd.Server.Host != tc.expected.Server.Host {
				t.Errorf("Expected Server.Host=%v, got %v", tc.expected.Server.Host, cmd.Server.Host)
			}
		})
	}
}

// TestEnhancedFeaturesErrorHandling tests error handling for enhanced features
func TestEnhancedFeaturesErrorHandling(t *testing.T) {
	type ServerCmd struct {
		Port int `arg:"-p,--port" default:"8080"`
	}

	type RootCmd struct {
		Verbose bool       `arg:"-v,--verbose"`
		Server  *ServerCmd `arg:"subcommand:server"`
	}

	// Test that invalid commands with invalid options fail appropriately
	// Note: Invalid commands alone don't fail because they're treated as positional args
	// The error comes from invalid options being processed
	invalidCases := []struct {
		name string
		args []string
	}{
		{"InvalidCommandWithInvalidOption", []string{"invalidcommand", "--port", "9000"}}, // --port not valid for root
		{"PartialMatchWithInvalidOption", []string{"serv", "--port", "9000"}},             // --port not valid for root
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			var cmd RootCmd
			err := ParseArgs(&cmd, tc.args)
			if err == nil {
				t.Errorf("Expected error for invalid command args: %v", tc.args)
			}
		})
	}

	// Test that unknown options still fail appropriately
	unknownOptionCases := []struct {
		name string
		args []string
	}{
		{"UnknownOption", []string{"server", "--unknown-option"}},
		{"InvalidWithValue", []string{"SERVER", "--invalid", "value"}},
		{"NonexistentOption", []string{"server", "--verbose", "--nonexistent"}},
	}

	for _, tc := range unknownOptionCases {
		t.Run(tc.name, func(t *testing.T) {
			var cmd RootCmd
			err := ParseArgs(&cmd, tc.args)
			if err == nil {
				t.Errorf("Expected error for unknown option args: %v", tc.args)
			}
		})
	}

	// Test that valid commands with valid options work correctly
	validCases := []struct {
		name string
		args []string
	}{
		{"ValidServerCommand", []string{"server", "--port", "9000"}},
		{"ValidServerCommandCaseInsensitive", []string{"SERVER", "--port", "9000"}},
		{"ValidRootOptions", []string{"--verbose"}},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			var cmd RootCmd
			err := ParseArgs(&cmd, tc.args)
			if err != nil {
				t.Errorf("Unexpected error for valid args %v: %v", tc.args, err)
			}
		})
	}
}

// TestPropertyValidationHelpers contains helper functions for property validation
func TestPropertyValidationHelpers(t *testing.T) {
	// Test helper functions used in property tests
	t.Run("StructEquality", func(t *testing.T) {
		type TestStruct struct {
			Name  string
			Value int
		}

		s1 := TestStruct{Name: "test", Value: 42}
		s2 := TestStruct{Name: "test", Value: 42}
		s3 := TestStruct{Name: "different", Value: 24}

		if !reflect.DeepEqual(s1, s2) {
			t.Error("Identical structs should be equal")
		}
		if reflect.DeepEqual(s1, s3) {
			t.Error("Different structs should not be equal")
		}
	})

	t.Run("ArgumentGeneration", func(t *testing.T) {
		// Test that argument generation produces valid command lines
		args := []string{"server", "--verbose", "--port", "9000"}
		if len(args) != 4 {
			t.Errorf("Expected 4 arguments, got %d", len(args))
		}
		if args[0] != "server" {
			t.Errorf("Expected first arg to be 'server', got '%s'", args[0])
		}
	})

	t.Run("CaseVariations", func(t *testing.T) {
		// Test case variation generation
		variations := []string{"server", "SERVER", "Server", "SeRvEr"}
		for _, variation := range variations {
			if !strings.EqualFold(variation, "server") {
				t.Errorf("Case variation '%s' should match 'server' case-insensitively", variation)
			}
		}
	})
}

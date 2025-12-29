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

	t.Run("ComplexCompactedMixedInheritance", func(t *testing.T) {
		// Test case for -vpc 9000 where:
		// -v (verbose) is handled by parent (no argument)
		// -p (port) is handled by parent (optional argument, should take "c" from compacted string)
		// -c would be handled by child if -p didn't consume it
		// 9000 is a separate argument

		type ServerCmdWithCount struct {
			Port  int    `arg:"--port" default:"8080" help:"server port"`
			Host  string `arg:"-h,--host" default:"localhost" help:"server host"`
			Count int    `arg:"-c,--count" default:"1" help:"connection count"`
		}

		type RootCmdWithOptionalPrefix struct {
			Verbose bool                `arg:"-v,--verbose" help:"enable verbose output"`
			Debug   bool                `arg:"-d,--debug" help:"enable debug output"`
			Prefix  string              `arg:"-p,--prefix::" help:"optional prefix override"` // Optional argument
			Server  *ServerCmdWithCount `arg:"subcommand:server"`
		}

		var cmd RootCmdWithOptionalPrefix
		err := ParseArgs(&cmd, []string{"server", "-vpc", "9000"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		// Verify parent options were processed correctly
		if !cmd.Verbose {
			t.Errorf("Expected Verbose=true (inherited from root), got %v", cmd.Verbose)
		}

		// The -p option should consume "c" as its optional argument
		if cmd.Prefix != "c" {
			t.Errorf("Expected Prefix='c' (optional argument from compacted string), got %v", cmd.Prefix)
		}

		if cmd.Server == nil {
			t.Fatalf("Expected Server to be initialized")
		}

		// The child's -c option should not be set since "c" was consumed by parent's -p
		if cmd.Server.Count != 1 { // Should remain default value
			t.Errorf("Expected Server.Count=1 (default, since 'c' was consumed by parent -p), got %v", cmd.Server.Count)
		}

		// 9000 should remain as a non-option argument (not consumed by any option)
		// Note: This would need to be tested by checking remaining args, but our current
		// test structure doesn't expose that. The important part is that the compacted
		// option processing worked correctly.
	})

	t.Run("ExactScenarioVPC9000", func(t *testing.T) {
		// Test the exact scenario: -vpc 9000 where:
		// -v is handled by parent (no argument)
		// -p is handled by parent (optional argument, takes "c")
		// -c would be handled by child but is consumed by parent's -p
		// 9000 remains as separate argument

		type ChildCmd struct {
			Config string `arg:"-c,--config" default:"default.conf" help:"config file"`
			Port   int    `arg:"--port" default:"8080" help:"port number"`
		}

		type ParentCmd struct {
			Verbose bool      `arg:"-v,--verbose" help:"enable verbose output"`
			Prefix  string    `arg:"-p,--prefix::" help:"optional prefix"` // Optional argument
			Child   *ChildCmd `arg:"subcommand:child"`
		}

		var cmd ParentCmd
		err := ParseArgs(&cmd, []string{"child", "-vpc", "9000"})
		if err != nil {
			t.Fatalf("ParseArgs() unexpected error: %v", err)
		}

		// Verify the exact behavior:
		// 1. -v should be processed by parent
		if !cmd.Verbose {
			t.Errorf("Expected Verbose=true (parent option), got %v", cmd.Verbose)
		}

		// 2. -p should consume "c" as its optional argument
		if cmd.Prefix != "c" {
			t.Errorf("Expected Prefix='c' (consumed from compacted string), got %v", cmd.Prefix)
		}

		// 3. Child should be initialized
		if cmd.Child == nil {
			t.Fatalf("Expected Child to be initialized")
		}

		// 4. Child's -c option should remain default since "c" was consumed by parent's -p
		if cmd.Child.Config != "default.conf" {
			t.Errorf("Expected Child.Config='default.conf' (default, since 'c' consumed by parent), got %v", cmd.Child.Config)
		}

		// This test demonstrates that the parent's optional argument option (-p)
		// correctly consumes the next character in the compacted string ("c"),
		// preventing the child from processing it as its own option.
	})
}

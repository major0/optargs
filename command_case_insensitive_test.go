package optargs

import (
	"testing"
)

func TestCommandCaseInsensitive(t *testing.T) {
	// Create a parser with case insensitive commands enabled
	config := ParserConfig{
		commandCaseIgnore: true,
	}

	parser, err := NewParser(config, map[byte]*Flag{}, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("NewParser() unexpected error: %v", err)
	}

	// Create a subcommand parser
	subConfig := ParserConfig{}
	subParser, err := NewParser(subConfig, map[byte]*Flag{}, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("NewParser() for subcommand unexpected error: %v", err)
	}

	// Register subcommand
	parser.AddCmd("server", subParser)

	t.Run("ExactMatch", func(t *testing.T) {
		foundParser, exists := parser.GetCommand("server")
		if !exists {
			t.Error("Expected to find 'server' command")
		}
		if foundParser != subParser {
			t.Error("Expected to get the same parser back")
		}
	})

	t.Run("UpperCaseMatch", func(t *testing.T) {
		foundParser, exists := parser.GetCommand("SERVER")
		if !exists {
			t.Error("Expected to find 'SERVER' command (case insensitive)")
		}
		if foundParser != subParser {
			t.Error("Expected to get the same parser back")
		}
	})

	t.Run("MixedCaseMatch", func(t *testing.T) {
		foundParser, exists := parser.GetCommand("SeRvEr")
		if !exists {
			t.Error("Expected to find 'SeRvEr' command (case insensitive)")
		}
		if foundParser != subParser {
			t.Error("Expected to get the same parser back")
		}
	})

	t.Run("ExecuteCommandCaseInsensitive", func(t *testing.T) {
		executedParser, err := parser.ExecuteCommand("SERVER", []string{"--help"})
		if err != nil {
			t.Fatalf("ExecuteCommand() unexpected error: %v", err)
		}
		if executedParser != subParser {
			t.Error("Expected to get the same parser back from ExecuteCommand")
		}
	})

	t.Run("NonExistentCommand", func(t *testing.T) {
		_, exists := parser.GetCommand("nonexistent")
		if exists {
			t.Error("Expected not to find 'nonexistent' command")
		}
	})
}

func TestCommandCaseSensitive(t *testing.T) {
	// Create a parser with case sensitive commands (default)
	config := ParserConfig{
		commandCaseIgnore: false,
	}

	parser, err := NewParser(config, map[byte]*Flag{}, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("NewParser() unexpected error: %v", err)
	}

	// Create a subcommand parser
	subConfig := ParserConfig{}
	subParser, err := NewParser(subConfig, map[byte]*Flag{}, map[string]*Flag{}, []string{}, nil)
	if err != nil {
		t.Fatalf("NewParser() for subcommand unexpected error: %v", err)
	}

	// Register subcommand
	parser.AddCmd("server", subParser)

	t.Run("ExactMatch", func(t *testing.T) {
		foundParser, exists := parser.GetCommand("server")
		if !exists {
			t.Error("Expected to find 'server' command")
		}
		if foundParser != subParser {
			t.Error("Expected to get the same parser back")
		}
	})

	t.Run("UpperCaseNoMatch", func(t *testing.T) {
		_, exists := parser.GetCommand("SERVER")
		if exists {
			t.Error("Expected NOT to find 'SERVER' command (case sensitive)")
		}
	})

	t.Run("MixedCaseNoMatch", func(t *testing.T) {
		_, exists := parser.GetCommand("SeRvEr")
		if exists {
			t.Error("Expected NOT to find 'SeRvEr' command (case sensitive)")
		}
	})
}

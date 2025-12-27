package goarg

import (
	"fmt"
	"testing"

	"github.com/major0/optargs"
)

func TestOptArgsCoreCompactedOptions(t *testing.T) {
	// Test OptArgs Core directly to understand the issue
	shortOpts := map[byte]*optargs.Flag{
		'v': {HasArg: optargs.NoArgument},
		'd': {HasArg: optargs.NoArgument},
		'p': {HasArg: optargs.RequiredArgument},
	}
	
	longOpts := map[string]*optargs.Flag{
		"verbose": {HasArg: optargs.NoArgument},
		"debug":   {HasArg: optargs.NoArgument},
		"port":    {HasArg: optargs.RequiredArgument},
	}

	t.Run("CompactedNoArgs", func(t *testing.T) {
		args := []string{"-vd"}
		parser, err := optargs.NewParser(optargs.ParserConfig{}, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("NewParser failed: %v", err)
		}

		var options []optargs.Option
		for option, err := range parser.Options() {
			if err != nil {
				t.Errorf("Option parsing error: %v", err)
				continue
			}
			options = append(options, option)
			fmt.Printf("Option: Name=%s, HasArg=%v, Arg=%s\n", option.Name, option.HasArg, option.Arg)
		}

		if len(options) != 2 {
			t.Errorf("Expected 2 options, got %d", len(options))
		}
		
		// Check that we got both -v and -d
		foundV, foundD := false, false
		for _, opt := range options {
			if opt.Name == "v" {
				foundV = true
			}
			if opt.Name == "d" {
				foundD = true
			}
		}
		
		if !foundV {
			t.Error("Expected to find -v option")
		}
		if !foundD {
			t.Error("Expected to find -d option")
		}
	})

	t.Run("CompactedWithArg", func(t *testing.T) {
		args := []string{"-vp", "9000"}
		parser, err := optargs.NewParser(optargs.ParserConfig{}, shortOpts, longOpts, args)
		if err != nil {
			t.Fatalf("NewParser failed: %v", err)
		}

		var options []optargs.Option
		for option, err := range parser.Options() {
			if err != nil {
				t.Errorf("Option parsing error: %v", err)
				continue
			}
			options = append(options, option)
			fmt.Printf("Option: Name=%s, HasArg=%v, Arg=%s\n", option.Name, option.HasArg, option.Arg)
		}

		if len(options) != 2 {
			t.Errorf("Expected 2 options, got %d", len(options))
		}
		
		// Check that we got -v and -p with correct argument
		foundV, foundP := false, false
		for _, opt := range options {
			if opt.Name == "v" {
				foundV = true
			}
			if opt.Name == "p" && opt.HasArg && opt.Arg == "9000" {
				foundP = true
			}
		}
		
		if !foundV {
			t.Error("Expected to find -v option")
		}
		if !foundP {
			t.Error("Expected to find -p option with arg '9000'")
		}
	})
}
// Package compat captures upstream alexflint/go-arg behavior into golden files.
// Run `go test -update` to regenerate golden files.
package compat

import (
	"flag"

	"github.com/alexflint/go-arg"
)

var update = flag.Bool("update", false, "update golden files")

// Scenario describes a single compatibility test case.
type Scenario struct {
	Name       string
	Args       []string
	NewParser  func() (*arg.Parser, interface{}, error)
	WantErr    bool
	SkipHelp   bool // skip help/usage capture (e.g. error-only scenarios)
	SkipValues bool // skip parsed-values capture
}

func scenarios() []Scenario {
	return []Scenario{
		basicStringInt(),
		boolFlag(),
		defaultValues(),
		requiredMissing(),
		positionalArgs(),
		sliceOption(),
		envOption(),
		unknownOption(),
		subcommandBasic(),
		helpOutput(),
	}
}

// --- scenario definitions ---

func basicStringInt() Scenario {
	type Args struct {
		Name  string `arg:"-n,--name" help:"user name"`
		Count int    `arg:"-c,--count" help:"repeat count"`
	}
	return Scenario{
		Name: "basic_string_int",
		Args: []string{"--name", "alice", "--count", "3"},
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func boolFlag() Scenario {
	type Args struct {
		Verbose bool `arg:"-v,--verbose" help:"verbose output"`
	}
	return Scenario{
		Name: "bool_flag",
		Args: []string{"--verbose"},
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func defaultValues() Scenario {
	type Args struct {
		Port int    `arg:"-p,--port" default:"8080" help:"listen port"`
		Host string `arg:"--host" default:"localhost" help:"bind host"`
	}
	return Scenario{
		Name: "default_values",
		Args: []string{}, // no args — defaults should apply
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func requiredMissing() Scenario {
	type Args struct {
		Input string `arg:"--input,required" help:"input file"`
	}
	return Scenario{
		Name:       "required_missing",
		Args:       []string{},
		WantErr:    true,
		SkipHelp:   true,
		SkipValues: true,
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func positionalArgs() Scenario {
	type Args struct {
		Source string `arg:"positional,required" help:"source file"`
		Dest   string `arg:"positional" help:"destination file"`
	}
	return Scenario{
		Name: "positional_args",
		Args: []string{"input.txt", "output.txt"},
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func sliceOption() Scenario {
	type Args struct {
		Files []string `arg:"-f,--file" help:"input files"`
	}
	return Scenario{
		Name: "slice_option",
		Args: []string{"--file", "a.txt", "--file", "b.txt"},
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func envOption() Scenario {
	type Args struct {
		Token string `arg:"--token,env:API_TOKEN" help:"API token"`
	}
	return Scenario{
		Name:     "env_option",
		Args:     []string{},
		SkipHelp: true, // env behavior only
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func unknownOption() Scenario {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	return Scenario{
		Name:       "unknown_option",
		Args:       []string{"--unknown"},
		WantErr:    true,
		SkipHelp:   true,
		SkipValues: true,
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func subcommandBasic() Scenario {
	type ServerCmd struct {
		Port int `arg:"-p,--port" default:"8080" help:"listen port"`
	}
	type Args struct {
		Verbose bool       `arg:"-v,--verbose" help:"verbose output"`
		Server  *ServerCmd `arg:"subcommand:server" help:"run server"`
	}
	return Scenario{
		Name: "subcommand_basic",
		Args: []string{"server", "--port", "9090"},
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func helpOutput() Scenario {
	type Args struct {
		Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int    `arg:"-c,--count" help:"number of items"`
		Output  string `arg:"-o,--output" help:"output file"`
	}
	return Scenario{
		Name:       "help_output",
		Args:       []string{}, // we capture help, not parse result
		SkipValues: true,
		NewParser: func() (*arg.Parser, interface{}, error) {
			var a Args
			p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

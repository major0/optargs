package goarg_test

import (
	"fmt"

	"github.com/major0/optargs/goarg"
)

func Example_basic() {
	type Args struct {
		Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int    `arg:"-c,--count" default:"1" help:"repeat count"`
		Output  string `arg:"-o,--output" help:"output file"`
	}
	var args Args
	p, _ := goarg.NewParser(goarg.Config{Program: "example"}, &args)
	_ = p.Parse([]string{"--verbose", "--count", "3", "--output", "out.txt"})
	fmt.Printf("verbose=%v count=%d output=%s\n", args.Verbose, args.Count, args.Output)
	// Output: verbose=true count=3 output=out.txt
}

func Example_positional() {
	type Args struct {
		Source string `arg:"positional,required" help:"source file"`
		Dest   string `arg:"positional" help:"destination file"`
	}
	var args Args
	p, _ := goarg.NewParser(goarg.Config{Program: "copy"}, &args)
	_ = p.Parse([]string{"input.txt", "output.txt"})
	fmt.Printf("source=%s dest=%s\n", args.Source, args.Dest)
	// Output: source=input.txt dest=output.txt
}

func Example_subcommand() {
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080" help:"listen port"`
		Host string `arg:"--host" default:"localhost" help:"bind host"`
	}
	type Args struct {
		Verbose bool       `arg:"-v,--verbose" help:"verbose output"`
		Server  *ServerCmd `arg:"subcommand:server" help:"run server"`
	}
	var args Args
	p, _ := goarg.NewParser(goarg.Config{Program: "app"}, &args)
	_ = p.Parse([]string{"server", "--port", "9090", "--verbose"})
	fmt.Printf("verbose=%v server.port=%d\n", args.Verbose, args.Server.Port)
	// Output: verbose=true server.port=9090
}

func Example_environment() {
	type Args struct {
		Token string `arg:"env:API_TOKEN" help:"API token"`
		Port  int    `arg:"-p,--port" default:"8080" help:"listen port"`
	}
	var args Args
	p, _ := goarg.NewParser(goarg.Config{Program: "api"}, &args)
	_ = p.Parse([]string{"--port", "9090"})
	fmt.Printf("port=%d\n", args.Port)
	// Output: port=9090
}

func Example_mapType() {
	type Args struct {
		Headers map[string]string `arg:"-H,--header" help:"HTTP headers"`
	}
	var args Args
	p, _ := goarg.NewParser(goarg.Config{Program: "curl"}, &args)
	_ = p.Parse([]string{"--header", "Content-Type=application/json", "--header", "Accept=text/html"})
	fmt.Printf("headers=%v\n", args.Headers)
	// Output: headers=map[Accept:text/html Content-Type:application/json]
}

func Example_embedded() {
	type CommonOpts struct {
		Verbose bool `arg:"-v,--verbose" help:"verbose output"`
		Debug   bool `arg:"-d,--debug" help:"debug mode"`
	}
	type Args struct {
		CommonOpts
		Output string `arg:"-o,--output" help:"output file"`
	}
	var args Args
	p, _ := goarg.NewParser(goarg.Config{Program: "tool"}, &args)
	_ = p.Parse([]string{"--verbose", "--output", "out.txt"})
	fmt.Printf("verbose=%v output=%s\n", args.Verbose, args.Output)
	// Output: verbose=true output=out.txt
}

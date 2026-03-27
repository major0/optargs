package compat

import (
	"testing"

	"github.com/alexflint/go-arg"
)

// BenchmarkUpstreamParseSimple benchmarks upstream go-arg parsing a simple struct.
func BenchmarkUpstreamParseSimple(b *testing.B) {
	type Args struct {
		Verbose bool   `arg:"-v,--verbose"`
		Count   int    `arg:"-c,--count"`
		Output  string `arg:"-o,--output"`
	}
	args := []string{"--verbose", "--count", "42", "--output", "out.txt"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Args
		p, _ := arg.NewParser(arg.Config{Program: "bench"}, &a)
		_ = p.Parse(args)
	}
}

// BenchmarkUpstreamParseSubcommand benchmarks upstream go-arg with a subcommand.
func BenchmarkUpstreamParseSubcommand(b *testing.B) {
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080"`
		Host string `arg:"--host" default:"localhost"`
	}
	type Args struct {
		Verbose bool       `arg:"-v,--verbose"`
		Server  *ServerCmd `arg:"subcommand:server"`
	}
	args := []string{"server", "--port", "9090", "--verbose"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Args
		p, _ := arg.NewParser(arg.Config{Program: "bench"}, &a)
		_ = p.Parse(args)
	}
}

// BenchmarkUpstreamParseDefaults benchmarks upstream go-arg with defaults.
func BenchmarkUpstreamParseDefaults(b *testing.B) {
	type Args struct {
		Port     int    `arg:"-p,--port" default:"8080"`
		Host     string `arg:"--host" default:"localhost"`
		LogLevel string `arg:"--log-level" default:"info"`
		Workers  int    `arg:"--workers" default:"4"`
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Args
		p, _ := arg.NewParser(arg.Config{Program: "bench"}, &a)
		_ = p.Parse([]string{})
	}
}

// BenchmarkUpstreamParsePositional benchmarks upstream go-arg with positionals.
func BenchmarkUpstreamParsePositional(b *testing.B) {
	type Args struct {
		Source string   `arg:"positional,required"`
		Dest   string   `arg:"positional"`
		Files  []string `arg:"positional"`
	}
	args := []string{"input.txt", "output.txt", "a.go", "b.go", "c.go"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Args
		p, _ := arg.NewParser(arg.Config{Program: "bench"}, &a)
		_ = p.Parse(args)
	}
}

// BenchmarkUpstreamNewParser benchmarks upstream go-arg parser creation.
func BenchmarkUpstreamNewParser(b *testing.B) {
	type ServerCmd struct {
		Port int    `arg:"-p,--port" default:"8080"`
		Host string `arg:"--host" default:"localhost"`
	}
	type Args struct {
		Verbose bool       `arg:"-v,--verbose" help:"verbose output"`
		Count   int        `arg:"-c,--count" help:"count"`
		Output  string     `arg:"-o,--output" help:"output file"`
		Server  *ServerCmd `arg:"subcommand:server" help:"run server"`
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Args
		_, _ = arg.NewParser(arg.Config{Program: "bench"}, &a)
	}
}

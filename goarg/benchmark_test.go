package goarg

import "testing"

// BenchmarkParseSimple benchmarks parsing a simple struct with a few options.
func BenchmarkParseSimple(b *testing.B) {
	type Args struct {
		Verbose bool   `arg:"-v,--verbose"`
		Count   int    `arg:"-c,--count"`
		Output  string `arg:"-o,--output"`
	}
	args := []string{"--verbose", "--count", "42", "--output", "out.txt"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Args
		_ = ParseArgs(&a, args)
	}
}

// BenchmarkParseSubcommand benchmarks parsing with a subcommand.
func BenchmarkParseSubcommand(b *testing.B) {
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
		_ = ParseArgs(&a, args)
	}
}

// BenchmarkParseDefaults benchmarks parsing with default values applied.
func BenchmarkParseDefaults(b *testing.B) {
	type Args struct {
		Port     int    `arg:"-p,--port" default:"8080"`
		Host     string `arg:"--host" default:"localhost"`
		LogLevel string `arg:"--log-level" default:"info"`
		Workers  int    `arg:"--workers" default:"4"`
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Args
		_ = ParseArgs(&a, []string{})
	}
}

// BenchmarkParsePositional benchmarks parsing positional arguments.
func BenchmarkParsePositional(b *testing.B) {
	type Args struct {
		Source string   `arg:"positional,required"`
		Dest   string   `arg:"positional"`
		Files  []string `arg:"positional"`
	}
	args := []string{"input.txt", "output.txt", "a.go", "b.go", "c.go"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a Args
		_ = ParseArgs(&a, args)
	}
}

// BenchmarkNewParser benchmarks parser creation (struct tag parsing).
func BenchmarkNewParser(b *testing.B) {
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
		_, _ = NewParser(Config{Program: "bench"}, &a)
	}
}

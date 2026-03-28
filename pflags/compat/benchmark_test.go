package compat

import (
	"testing"

	"github.com/spf13/pflag"
)

// BenchmarkUpstreamParseFewFlags benchmarks upstream spf13/pflag with few flags.
func BenchmarkUpstreamParseFewFlags(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fs := pflag.NewFlagSet("bench", pflag.ContinueOnError)
		fs.String("output", "", "")
		fs.Bool("verbose", false, "")
		fs.Int("count", 0, "")
		fs.Parse([]string{"--output", "file.txt", "--verbose", "--count", "42"}) //nolint:errcheck
	}
}

// BenchmarkUpstreamParseManyFlags benchmarks upstream with many registered flags.
func BenchmarkUpstreamParseManyFlags(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fs := pflag.NewFlagSet("bench", pflag.ContinueOnError)
		for j := 0; j < 50; j++ {
			fs.String("flag-"+string(rune('a'+j/26))+string(rune('a'+j%26)), "", "")
		}
		fs.String("target", "", "")
		fs.Parse([]string{"--target", "value"}) //nolint:errcheck
	}
}

// BenchmarkUpstreamParseShorthand benchmarks upstream shorthand parsing.
func BenchmarkUpstreamParseShorthand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fs := pflag.NewFlagSet("bench", pflag.ContinueOnError)
		fs.BoolP("alpha", "a", false, "")
		fs.BoolP("beta", "b", false, "")
		fs.BoolP("gamma", "c", false, "")
		fs.StringP("output", "o", "", "")
		fs.Parse([]string{"-abc", "-o", "file.txt"}) //nolint:errcheck
	}
}

// BenchmarkUpstreamFlagUsages benchmarks upstream help text generation.
func BenchmarkUpstreamFlagUsages(b *testing.B) {
	fs := pflag.NewFlagSet("bench", pflag.ContinueOnError)
	fs.StringP("output", "o", "default.txt", "output file path")
	fs.BoolP("verbose", "v", false, "enable verbose output")
	fs.Int("count", 10, "number of items")
	fs.Float64("rate", 1.0, "processing rate")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.FlagUsages()
	}
}

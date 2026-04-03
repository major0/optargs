package pflags

import "testing"

// BenchmarkParseFewFlags benchmarks parsing a small number of flags.
func BenchmarkParseFewFlags(b *testing.B) {
	for range b.N {
		fs := NewFlagSet("bench", ContinueOnError)
		fs.StringVar(new(string), "output", "", "")
		fs.BoolVar(new(bool), "verbose", false, "")
		fs.IntVar(new(int), "count", 0, "")
		_ = fs.Parse([]string{"--output", "file.txt", "--verbose", "--count", "42"})
	}
}

// BenchmarkParseManyFlags benchmarks parsing with many registered flags.
func BenchmarkParseManyFlags(b *testing.B) {
	for range b.N {
		fs := NewFlagSet("bench", ContinueOnError)
		for j := range 50 {
			fs.StringVar(new(string), "flag-"+string(rune('a'+j/26))+string(rune('a'+j%26)), "", "")
		}
		fs.StringVar(new(string), "target", "", "")
		_ = fs.Parse([]string{"--target", "value"})
	}
}

// BenchmarkParseShorthand benchmarks shorthand flag parsing.
func BenchmarkParseShorthand(b *testing.B) {
	for range b.N {
		fs := NewFlagSet("bench", ContinueOnError)
		fs.BoolVarP(new(bool), "alpha", "a", false, "")
		fs.BoolVarP(new(bool), "beta", "b", false, "")
		fs.BoolVarP(new(bool), "gamma", "c", false, "")
		fs.StringVarP(new(string), "output", "o", "", "")
		_ = fs.Parse([]string{"-abc", "-o", "file.txt"})
	}
}

// BenchmarkFlagUsages benchmarks help text generation.
func BenchmarkFlagUsages(b *testing.B) {
	fs := NewFlagSet("bench", ContinueOnError)
	fs.StringVarP(new(string), "output", "o", "default.txt", "output file path")
	fs.BoolVarP(new(bool), "verbose", "v", false, "enable verbose output")
	fs.IntVar(new(int), "count", 10, "number of items")
	fs.Float64Var(new(float64), "rate", 1.0, "processing rate")

	b.ResetTimer()
	for range b.N {
		_ = fs.FlagUsages()
	}
}

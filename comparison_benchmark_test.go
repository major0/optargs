package optargs

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"testing"
)

// benchStdFlag creates a flag.FlagSet, registers flags via setup, parses
// args, and accesses all values.  Mirrors benchParse for the standard
// library side of comparison benchmarks.
func benchStdFlag(b *testing.B, args []string, setup func(*flag.FlagSet)) {
	b.Helper()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	setup(fs)
	if err := fs.Parse(args); err != nil {
		b.Fatal(err)
	}
}

// BenchmarkComparisonWithStdFlag compares performance with Go's standard flag package
func BenchmarkComparisonWithStdFlag(b *testing.B) {
	args := []string{"-a", "arg1", "-b", "arg2", "-c"}

	b.Run("OptArgs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, append([]string{"prog"}, args...), "a:b:c", nil)
		}
	})

	b.Run("StdFlag", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchStdFlag(b, args, func(fs *flag.FlagSet) {
				_ = fs.String("a", "", "a flag")
				_ = fs.String("b", "", "b flag")
				_ = fs.Bool("c", false, "c flag")
			})
		}
	})
}

// BenchmarkComparisonLongOptions compares long option performance
func BenchmarkComparisonLongOptions(b *testing.B) {
	args := []string{"--verbose", "--output", "file.txt", "--config", "cfg.ini"}
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: RequiredArgument},
	}

	b.Run("OptArgs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, append([]string{"prog"}, args...), "", longOpts)
		}
	})

	b.Run("StdFlag", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchStdFlag(b, args, func(fs *flag.FlagSet) {
				_ = fs.Bool("verbose", false, "verbose flag")
				_ = fs.String("output", "", "output flag")
				_ = fs.String("config", "", "config flag")
			})
		}
	})
}

// BenchmarkComparisonComplexScenarios compares complex parsing scenarios
func BenchmarkComparisonComplexScenarios(b *testing.B) {
	args := []string{"-abc", "arg1", "--verbose", "--output=file.txt", "-d", "arg2"}
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}

	b.Run("OptArgs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, append([]string{"prog"}, args...), "abc:d:", longOpts)
		}
	})

	b.Run("StdFlag_Equivalent", func(b *testing.B) {
		// Standard flag can't handle compacted options, so we simulate equivalent behavior
		equivalentArgs := []string{"-a", "-b", "-c", "arg1", "-verbose", "-output", "file.txt", "-d", "arg2"}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchStdFlag(b, equivalentArgs, func(fs *flag.FlagSet) {
				_ = fs.Bool("a", false, "a flag")
				_ = fs.Bool("b", false, "b flag")
				_ = fs.String("c", "", "c flag")
				_ = fs.Bool("verbose", false, "verbose flag")
				_ = fs.String("output", "", "output flag")
				_ = fs.String("d", "", "d flag")
			})
		}
	})
}

// BenchmarkComparisonMemoryUsage compares memory allocation patterns
func BenchmarkComparisonMemoryUsage(b *testing.B) {
	args := []string{"-a", "arg1", "-b", "arg2", "--verbose"}
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
	}

	b.Run("OptArgs_Memory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, append([]string{"prog"}, args...), "a:b:", longOpts)
		}
	})

	b.Run("StdFlag_Memory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchStdFlag(b, args, func(fs *flag.FlagSet) {
				_ = fs.String("a", "", "a flag")
				_ = fs.String("b", "", "b flag")
				_ = fs.Bool("verbose", false, "verbose flag")
			})
		}
	})
}

// BenchmarkScalability tests how performance scales with argument count
func BenchmarkScalability(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("OptArgs_Size%d", size), func(b *testing.B) {
			args := make([]string, 0, size+1)
			args = append(args, "prog")
			for j := 1; j <= size; j++ {
				if j%3 == 0 {
					args = append(args, "-a", "arg"+strconv.Itoa(j))
				} else {
					args = append(args, "arg"+strconv.Itoa(j))
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchParse(b, args, "a:", nil)
			}
		})

		b.Run(fmt.Sprintf("StdFlag_Size%d", size), func(b *testing.B) {
			args := make([]string, 0, size)
			for j := 1; j <= size; j++ {
				if j%2 == 0 {
					args = append(args, "-a", "arg"+strconv.Itoa(j))
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchStdFlag(b, args, func(fs *flag.FlagSet) {
					_ = fs.String("a", "", "a flag")
				})
			}
		})
	}
}

// BenchmarkFeatureComparison benchmarks features unique to OptArgs
func BenchmarkFeatureComparison(b *testing.B) {
	b.Run("OptionCompaction", func(b *testing.B) {
		args := []string{"prog", "-abcdef"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, args, "abcdef", nil)
		}
	})

	b.Run("PartialLongOptionMatching", func(b *testing.B) {
		args := []string{"prog", "--verbose", "--output", "file.txt"}
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, args, "", longOpts)
		}
	})

	b.Run("GNUWordExtension", func(b *testing.B) {
		args := []string{"prog", "-W", "verbose", "-W", "output=file.txt"}
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, args, "W;", longOpts)
		}
	})
}

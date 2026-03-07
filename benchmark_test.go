package optargs

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"testing"
)

// newParserFunc constructs a Parser from the given arguments.
type newParserFunc func(args []string, optstring string, longopts []Flag) (*Parser, error)

// benchParse creates a parser via newParser and consumes all options,
// failing the benchmark on any unexpected error.
func benchParse(b *testing.B, newParser newParserFunc, args []string, optstring string, longopts []Flag) {
	b.Helper()
	parser, err := newParser(args, optstring, longopts)
	if err != nil {
		b.Fatal(err)
	}
	for option, err := range parser.Options() {
		if err != nil {
			b.Fatal(err)
		}
		_ = option
	}
}

// BenchmarkGetOpt benchmarks the core GetOpt function with various scenarios
func BenchmarkGetOpt(b *testing.B) {
	testCases := []struct {
		name      string
		args      []string
		optstring string
	}{
		{
			name:      "SimpleShortOptions",
			args:      []string{"prog", "-a", "-b", "-c"},
			optstring: "abc",
		},
		{
			name:      "CompactedShortOptions",
			args:      []string{"prog", "-abc"},
			optstring: "abc",
		},
		{
			name:      "ShortOptionsWithArgs",
			args:      []string{"prog", "-a", "arg1", "-b", "arg2"},
			optstring: "a:b:",
		},
		{
			name:      "CompactedWithArgs",
			args:      []string{"prog", "-abarg1", "-c"},
			optstring: "ab:c",
		},
		{
			name:      "OptionalArgs",
			args:      []string{"prog", "-a", "-barg", "-c"},
			optstring: "ab::c",
		},
		{
			name:      "MixedOptions",
			args:      []string{"prog", "-a", "arg1", "-bc", "-d", "arg2"},
			optstring: "a:bcd:",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchParse(b, GetOptLong, tc.args, tc.optstring, nil)
			}
		})
	}
}

// BenchmarkGetOptLong benchmarks the GetOptLong function with long options
func BenchmarkGetOptLong(b *testing.B) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
		{Name: "help", HasArg: NoArgument},
		{Name: "version", HasArg: NoArgument},
	}

	testCases := []struct {
		name      string
		args      []string
		optstring string
	}{
		{
			name:      "LongOptionsOnly",
			args:      []string{"prog", "--verbose", "--help"},
			optstring: "",
		},
		{
			name:      "LongOptionsWithArgs",
			args:      []string{"prog", "--output", "file.txt", "--config", "cfg.ini"},
			optstring: "",
		},
		{
			name:      "LongOptionsEqualsForm",
			args:      []string{"prog", "--output=file.txt", "--config=cfg.ini"},
			optstring: "",
		},
		{
			name:      "MixedShortAndLong",
			args:      []string{"prog", "-v", "--output", "file.txt", "-h"},
			optstring: "vh",
		},
		{
			name:      "PartialLongOptions",
			args:      []string{"prog", "--verbose", "--output", "file.txt"},
			optstring: "",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchParse(b, GetOptLong, tc.args, tc.optstring, longOpts)
			}
		})
	}
}

// BenchmarkGetOptLongOnly benchmarks the GetOptLongOnly function
func BenchmarkGetOptLongOnly(b *testing.B) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
		{Name: "v", HasArg: NoArgument},
		{Name: "o", HasArg: RequiredArgument},
	}

	testCases := []struct {
		name      string
		args      []string
		optstring string
	}{
		{
			name:      "SingleDashLongOptions",
			args:      []string{"prog", "-verbose", "-output", "file.txt"},
			optstring: "",
		},
		{
			name:      "SingleCharFallback",
			args:      []string{"prog", "-v", "-o", "file.txt"},
			optstring: "",
		},
		{
			name:      "MixedSingleDash",
			args:      []string{"prog", "-verbose", "-v", "-output", "file.txt"},
			optstring: "",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchParse(b, GetOptLongOnly, tc.args, tc.optstring, longOpts)
			}
		})
	}
}

// BenchmarkLargeArgumentLists benchmarks performance with large argument lists
func BenchmarkLargeArgumentLists(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			args := make([]string, 0, size+1)
			args = append(args, "prog")
			for i := 1; i <= size; i++ {
				switch i % 4 {
				case 0:
					args = append(args, "-a")
				case 1:
					args = append(args, "-b", "arg"+strconv.Itoa(i))
				case 2:
					args = append(args, "-c")
				default:
					args = append(args, "arg"+strconv.Itoa(i))
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchParse(b, GetOptLong, args, "ab:c", nil)
			}
		})
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	shortArgs := []string{"prog", "-a", "arg1", "-b", "arg2", "-c", "arg3"}
	longArgs := []string{"prog", "-a", "arg1", "-b", "arg2", "-c", "arg3", "--long", "longarg"}
	longOpts := []Flag{
		{Name: "long", HasArg: RequiredArgument},
		{Name: "verbose", HasArg: NoArgument},
	}

	b.Run("GetOpt", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, GetOptLong, shortArgs, "a:b:c:", nil)
		}
	})

	b.Run("GetOptLong", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, GetOptLong, longArgs, "a:b:c:", longOpts)
		}
	})
}

// BenchmarkIteratorEfficiency benchmarks the iterator pattern efficiency
func BenchmarkIteratorEfficiency(b *testing.B) {
	args := []string{"prog", "-a", "-b", "-c", "-d", "-e", "-f", "-g", "-h"}
	optstring := "abcdefgh"

	b.Run("IteratorConsumption", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, GetOptLong, args, optstring, nil)
		}
	})

	b.Run("IteratorPartialConsumption", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser, err := GetOpt(args, optstring)
			if err != nil {
				b.Fatal(err)
			}

			count := 0
			for option, err := range parser.Options() {
				if err != nil {
					b.Fatal(err)
				}
				count++
				_ = option
				if count >= 3 {
					break
				}
			}
		}
	})
}

// BenchmarkComplexScenarios benchmarks complex real-world scenarios
func BenchmarkComplexScenarios(b *testing.B) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
		{Name: "format", HasArg: RequiredArgument},
		{Name: "debug", HasArg: NoArgument},
		{Name: "quiet", HasArg: NoArgument},
		{Name: "input", HasArg: RequiredArgument},
		{Name: "threads", HasArg: RequiredArgument},
	}

	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "CompilerLike",
			args: []string{"prog", "-O2", "-Wall", "-g", "-o", "output", "input.c"},
		},
		{
			name: "TarLike",
			args: []string{"prog", "-czf", "archive.tar.gz", "file1", "file2", "file3"},
		},
		{
			name: "GitLike",
			args: []string{"prog", "--verbose", "--format=json", "--output=result.json", "command"},
		},
		{
			name: "DockerLike",
			args: []string{"prog", "-it", "--rm", "--name", "container", "-v", "/host:/container", "image"},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchParse(b, GetOptLong, tc.args, "O:Wgo:czfitv:", longOpts)
			}
		})
	}
}

// BenchmarkGNUExtensions benchmarks GNU-specific extensions
func BenchmarkGNUExtensions(b *testing.B) {
	longOpts := []Flag{
		{Name: "word-option", HasArg: RequiredArgument},
		{Name: "another-word", HasArg: NoArgument},
	}

	b.Run("GNUWords", func(b *testing.B) {
		args := []string{"prog", "-W", "word-option=value", "-W", "another-word"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, GetOptLong, args, "W;", longOpts)
		}
	})

	b.Run("CaseInsensitive", func(b *testing.B) {
		args := []string{"prog", "--WORD-OPTION", "value", "--Another-Word"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, GetOptLong, args, "", longOpts)
		}
	})
}

// BenchmarkErrorHandling benchmarks error handling performance
func BenchmarkErrorHandling(b *testing.B) {
	b.Run("UnknownShortOption", func(b *testing.B) {
		args := []string{"prog", "-z"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser, err := GetOpt(args, "abc")
			if err != nil {
				b.Fatal(err)
			}
			for option, err := range parser.Options() {
				if err == nil {
					b.Fatal("expected error for unknown option")
				}
				_ = option
				break
			}
		}
	})

	b.Run("MissingRequiredArg", func(b *testing.B) {
		args := []string{"prog", "-a"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser, err := GetOpt(args, "a:")
			if err != nil {
				b.Fatal(err)
			}
			for option, err := range parser.Options() {
				if err == nil {
					b.Fatal("expected error for missing argument")
				}
				_ = option
				break
			}
		}
	})
}

// --- Comparison benchmarks (OptArgs vs standard library flag package) ---

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
	stdArgs := []string{"-a", "arg1", "-b", "arg2", "-c"}
	optArgs := append([]string{"prog"}, stdArgs...)

	b.Run("OptArgs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, GetOptLong, optArgs, "a:b:c", nil)
		}
	})

	b.Run("StdFlag", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchStdFlag(b, stdArgs, func(fs *flag.FlagSet) {
				_ = fs.String("a", "", "a flag")
				_ = fs.String("b", "", "b flag")
				_ = fs.Bool("c", false, "c flag")
			})
		}
	})
}

// BenchmarkComparisonLongOptions compares long option performance
func BenchmarkComparisonLongOptions(b *testing.B) {
	stdArgs := []string{"--verbose", "--output", "file.txt", "--config", "cfg.ini"}
	optArgs := append([]string{"prog"}, stdArgs...)
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: RequiredArgument},
	}

	b.Run("OptArgs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, GetOptLong, optArgs, "", longOpts)
		}
	})

	b.Run("StdFlag", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchStdFlag(b, stdArgs, func(fs *flag.FlagSet) {
				_ = fs.Bool("verbose", false, "verbose flag")
				_ = fs.String("output", "", "output flag")
				_ = fs.String("config", "", "config flag")
			})
		}
	})
}

// BenchmarkComparisonComplexScenarios compares complex parsing scenarios
func BenchmarkComparisonComplexScenarios(b *testing.B) {
	stdArgs := []string{"-abc", "arg1", "--verbose", "--output=file.txt", "-d", "arg2"}
	optArgs := append([]string{"prog"}, stdArgs...)
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}

	b.Run("OptArgs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, GetOptLong, optArgs, "abc:d:", longOpts)
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
	stdArgs := []string{"-a", "arg1", "-b", "arg2", "--verbose"}
	optArgs := append([]string{"prog"}, stdArgs...)
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
	}

	b.Run("OptArgs_Memory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchParse(b, GetOptLong, optArgs, "a:b:", longOpts)
		}
	})

	b.Run("StdFlag_Memory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchStdFlag(b, stdArgs, func(fs *flag.FlagSet) {
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
				benchParse(b, GetOptLong, args, "a:", nil)
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
			benchParse(b, GetOptLong, args, "abcdef", nil)
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
			benchParse(b, GetOptLong, args, "", longOpts)
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
			benchParse(b, GetOptLong, args, "W;", longOpts)
		}
	})
}

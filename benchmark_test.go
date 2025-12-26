package optargs

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
)

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
				parser, err := GetOpt(tc.args, tc.optstring)
				if err != nil {
					b.Fatal(err)
				}
				// Consume all options to ensure complete parsing
				for option, err := range parser.Options() {
					if err != nil {
						b.Fatal(err)
					}
					_ = option
				}
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
				parser, err := GetOptLong(tc.args, tc.optstring, longOpts)
				if err != nil {
					b.Fatal(err)
				}
				// Consume all options to ensure complete parsing
				for option, err := range parser.Options() {
					if err != nil {
						b.Fatal(err)
					}
					_ = option
				}
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
				parser, err := GetOptLongOnly(tc.args, tc.optstring, longOpts)
				if err != nil {
					b.Fatal(err)
				}
				// Consume all options to ensure complete parsing
				for option, err := range parser.Options() {
					if err != nil {
						b.Fatal(err)
					}
					_ = option
				}
			}
		})
	}
}

// BenchmarkLargeArgumentLists benchmarks performance with large argument lists
func BenchmarkLargeArgumentLists(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			// Generate large argument list
			args := make([]string, 0, size+1)
			args = append(args, "prog")
			for i := 1; i <= size; i++ {
				if i%4 == 0 {
					args = append(args, "-a")
				} else if i%4 == 1 {
					args = append(args, "-b", "arg"+strconv.Itoa(i))
				} else if i%4 == 2 {
					args = append(args, "-c")
				} else {
					args = append(args, "arg"+strconv.Itoa(i))
				}
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser, err := GetOpt(args, "ab:c")
				if err != nil {
					b.Fatal(err)
				}
				// Consume all options
				for option, err := range parser.Options() {
					if err != nil {
						b.Fatal(err)
					}
					_ = option
				}
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
			parser, err := GetOpt(shortArgs, "a:b:c:")
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
	})
	
	b.Run("GetOptLong", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser, err := GetOptLong(longArgs, "a:b:c:", longOpts)
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
	})
}

// BenchmarkIteratorEfficiency benchmarks the iterator pattern efficiency
func BenchmarkIteratorEfficiency(b *testing.B) {
	args := []string{"prog", "-a", "-b", "-c", "-d", "-e", "-f", "-g", "-h"}
	optstring := "abcdefgh"
	
	b.Run("IteratorConsumption", func(b *testing.B) {
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
			}
		}
	})
	
	b.Run("IteratorPartialConsumption", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser, err := GetOpt(args, optstring)
			if err != nil {
				b.Fatal(err)
			}
			
			// Only consume first 3 options
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
	// Simulate complex command-line scenarios
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
				parser, err := GetOptLong(tc.args, "O:Wgo:czfitv:", longOpts)
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
			parser, err := GetOptLong(args, "W;", longOpts)
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
	})
	
	b.Run("CaseInsensitive", func(b *testing.B) {
		args := []string{"prog", "--WORD-OPTION", "value", "--Another-Word"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser, err := GetOptLong(args, "", longOpts)
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
				// Expect error for unknown option
				if err == nil {
					b.Fatal("Expected error for unknown option")
				}
				_ = option
				break // Only check first error
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
				// Expect error for missing argument
				if err == nil {
					b.Fatal("Expected error for missing argument")
				}
				_ = option
				break // Only check first error
			}
		}
	})
}

// BenchmarkMemoryUsage provides detailed memory usage analysis
func BenchmarkMemoryUsage(b *testing.B) {
	args := []string{"prog", "-a", "arg1", "-b", "arg2", "--long", "longarg"}
	longOpts := []Flag{
		{Name: "long", HasArg: RequiredArgument},
		{Name: "verbose", HasArg: NoArgument},
	}
	
	b.Run("MemoryFootprint", func(b *testing.B) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser, err := GetOptLong(args, "a:b:", longOpts)
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
		
		runtime.GC()
		runtime.ReadMemStats(&m2)
		
		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
		b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
	})
}
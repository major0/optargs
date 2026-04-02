package optargs

import (
	"fmt"
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
// BenchmarkGetOpt benchmarks the core GetOpt function.
func BenchmarkGetOpt(b *testing.B) {
	testCases := []struct {
		name      string
		args      []string
		optstring string
	}{
		{"SimpleShortOptions", []string{"prog", "-a", "-b", "-c"}, "abc"},
		{"CompactedShortOptions", []string{"prog", "-abc"}, "abc"},
		{"ShortOptionsWithArgs", []string{"prog", "-a", "arg1", "-b", "arg2"}, "a:b:"},
	}
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				benchParse(b, GetOptLong, tc.args, tc.optstring, nil)
			}
		})
	}
}

// BenchmarkGetOptLong benchmarks the GetOptLong function with long options
// BenchmarkGetOptLong benchmarks the GetOptLong function with long options.
func BenchmarkGetOptLong(b *testing.B) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
	}
	testCases := []struct {
		name      string
		args      []string
		optstring string
	}{
		{"LongOptionsOnly", []string{"prog", "--verbose", "--output", "file.txt"}, ""},
		{"LongOptionsEqualsForm", []string{"prog", "--output=file.txt", "--config=cfg.ini"}, ""},
		{"MixedShortAndLong", []string{"prog", "-v", "--output", "file.txt"}, "v"},
	}
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				benchParse(b, GetOptLong, tc.args, tc.optstring, longOpts)
			}
		})
	}
}

// BenchmarkGetOptLongOnly benchmarks the GetOptLongOnly function.
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
			for range b.N {
				benchParse(b, GetOptLongOnly, tc.args, tc.optstring, longOpts)
			}
		})
	}
}

// BenchmarkLargeArgumentLists benchmarks performance with large argument lists.
func BenchmarkLargeArgumentLists(b *testing.B) {
	sizes := []int{100, 1000}

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
			for range b.N {
				benchParse(b, GetOptLong, args, "ab:c", nil)
			}
		})
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns.
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
		for range b.N {
			benchParse(b, GetOptLong, shortArgs, "a:b:c:", nil)
		}
	})

	b.Run("GetOptLong", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			benchParse(b, GetOptLong, longArgs, "a:b:c:", longOpts)
		}
	})
}

// BenchmarkIteratorEfficiency benchmarks the iterator pattern efficiency.
func BenchmarkIteratorEfficiency(b *testing.B) {
	args := []string{"prog", "-a", "-b", "-c", "-d", "-e", "-f", "-g", "-h"}
	optstring := "abcdefgh"

	b.Run("IteratorConsumption", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			benchParse(b, GetOptLong, args, optstring, nil)
		}
	})

	b.Run("IteratorPartialConsumption", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
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

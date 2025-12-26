package optargs

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

// BenchmarkComparisonWithStdFlag compares performance with Go's standard flag package
func BenchmarkComparisonWithStdFlag(b *testing.B) {
	// Test case: simple flags with arguments
	args := []string{"-a", "arg1", "-b", "arg2", "-c"}
	
	b.Run("OptArgs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser, err := GetOpt(append([]string{"prog"}, args...), "a:b:c")
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
	
	b.Run("StdFlag", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(os.Stderr) // Suppress output during benchmarking
			
			aFlag := fs.String("a", "", "a flag")
			bFlag := fs.String("b", "", "b flag")
			cFlag := fs.Bool("c", false, "c flag")
			
			err := fs.Parse(args)
			if err != nil {
				b.Fatal(err)
			}
			
			// Access the values to ensure they're processed
			_ = *aFlag
			_ = *bFlag
			_ = *cFlag
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
			parser, err := GetOptLong(append([]string{"prog"}, args...), "", longOpts)
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
	
	b.Run("StdFlag", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(os.Stderr) // Suppress output during benchmarking
			
			verboseFlag := fs.Bool("verbose", false, "verbose flag")
			outputFlag := fs.String("output", "", "output flag")
			configFlag := fs.String("config", "", "config flag")
			
			err := fs.Parse(args)
			if err != nil {
				b.Fatal(err)
			}
			
			// Access the values to ensure they're processed
			_ = *verboseFlag
			_ = *outputFlag
			_ = *configFlag
		}
	})
}

// BenchmarkComparisonComplexScenarios compares complex parsing scenarios
func BenchmarkComparisonComplexScenarios(b *testing.B) {
	// Complex scenario with mixed short/long options, compaction, etc.
	args := []string{"-abc", "arg1", "--verbose", "--output=file.txt", "-d", "arg2"}
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}
	
	b.Run("OptArgs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser, err := GetOptLong(append([]string{"prog"}, args...), "abc:d:", longOpts)
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
	
	b.Run("StdFlag_Equivalent", func(b *testing.B) {
		// Standard flag can't handle compacted options, so we simulate equivalent behavior
		equivalentArgs := []string{"-a", "-b", "-c", "arg1", "-verbose", "-output", "file.txt", "-d", "arg2"}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(os.Stderr) // Suppress output during benchmarking
			
			aFlag := fs.Bool("a", false, "a flag")
			bFlag := fs.Bool("b", false, "b flag")
			cFlag := fs.String("c", "", "c flag")
			verboseFlag := fs.Bool("verbose", false, "verbose flag")
			outputFlag := fs.String("output", "", "output flag")
			dFlag := fs.String("d", "", "d flag")
			
			err := fs.Parse(equivalentArgs)
			if err != nil {
				b.Fatal(err)
			}
			
			// Access the values to ensure they're processed
			_ = *aFlag
			_ = *bFlag
			_ = *cFlag
			_ = *verboseFlag
			_ = *outputFlag
			_ = *dFlag
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
			parser, err := GetOptLong(append([]string{"prog"}, args...), "a:b:", longOpts)
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
	
	b.Run("StdFlag_Memory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(os.Stderr) // Suppress output during benchmarking
			
			aFlag := fs.String("a", "", "a flag")
			bFlag := fs.String("b", "", "b flag")
			verboseFlag := fs.Bool("verbose", false, "verbose flag")
			
			err := fs.Parse(args)
			if err != nil {
				b.Fatal(err)
			}
			
			// Access the values to ensure they're processed
			_ = *aFlag
			_ = *bFlag
			_ = *verboseFlag
		}
	})
}

// BenchmarkScalability tests how performance scales with argument count
func BenchmarkScalability(b *testing.B) {
	sizes := []int{10, 50, 100, 500}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("OptArgs_Size%d", size), func(b *testing.B) {
			// Generate args
			args := make([]string, size+1)
			args[0] = "prog"
			for i := 1; i <= size; i++ {
				if i%2 == 0 {
					args[i] = "-a"
				} else {
					args[i] = "arg" + fmt.Sprintf("%d", i)
				}
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser, err := GetOpt(args, "a:")
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
		
		b.Run(fmt.Sprintf("StdFlag_Size%d", size), func(b *testing.B) {
			// Generate args for standard flag (no compaction)
			args := make([]string, 0, size)
			for i := 1; i <= size; i++ {
				if i%2 == 0 {
					args = append(args, "-a", "arg"+fmt.Sprintf("%d", i))
				}
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fs := flag.NewFlagSet("test", flag.ContinueOnError)
				fs.SetOutput(os.Stderr) // Suppress output during benchmarking
				
				aFlag := fs.String("a", "", "a flag")
				
				err := fs.Parse(args)
				if err != nil {
					b.Fatal(err)
				}
				
				_ = *aFlag
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
			parser, err := GetOpt(args, "abcdef")
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
	
	b.Run("PartialLongOptionMatching", func(b *testing.B) {
		args := []string{"prog", "--verbose", "--output", "file.txt"}
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
		}
		
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
	
	b.Run("GNUWordExtension", func(b *testing.B) {
		args := []string{"prog", "-W", "verbose", "-W", "output=file.txt"}
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
		}
		
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
}
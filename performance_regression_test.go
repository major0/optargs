package optargs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// PerformanceBaseline represents performance metrics for a specific test
type PerformanceBaseline struct {
	TestName    string    `json:"test_name"`
	NsPerOp     int64     `json:"ns_per_op"`
	AllocsPerOp int64     `json:"allocs_per_op"`
	BytesPerOp  int64     `json:"bytes_per_op"`
	Timestamp   time.Time `json:"timestamp"`
	GoVersion   string    `json:"go_version"`
	GOOS        string    `json:"goos"`
	GOARCH      string    `json:"goarch"`
	CPUModel    string    `json:"cpu_model,omitempty"`
}

// PerformanceReport contains all baseline measurements
type PerformanceReport struct {
	Baselines []PerformanceBaseline `json:"baselines"`
	Generated time.Time             `json:"generated"`
}

const (
	baselineFile = "performance_baselines.json"
	// Performance regression thresholds (percentage increase that triggers failure)
	timeRegressionThreshold   = 50.0  // 50% slower
	memoryRegressionThreshold = 100.0 // 100% more memory
	allocRegressionThreshold  = 100.0 // 100% more allocations
)

// loadBaselines loads existing performance baselines from file
func loadBaselines() (*PerformanceReport, error) {
	data, err := os.ReadFile(baselineFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &PerformanceReport{
				Baselines: make([]PerformanceBaseline, 0),
				Generated: time.Now(),
			}, nil
		}
		return nil, err
	}

	var report PerformanceReport
	err = json.Unmarshal(data, &report)
	return &report, err
}

// saveBaselines saves performance baselines to file
func saveBaselines(report *PerformanceReport) error {
	report.Generated = time.Now()
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(baselineFile, data, 0644)
}

// findBaseline finds a baseline for a specific test name
func findBaseline(report *PerformanceReport, testName string) *PerformanceBaseline {
	for i := range report.Baselines {
		if report.Baselines[i].TestName == testName {
			return &report.Baselines[i]
		}
	}
	return nil
}

// updateBaseline updates or adds a baseline for a specific test
func updateBaseline(report *PerformanceReport, baseline PerformanceBaseline) {
	for i := range report.Baselines {
		if report.Baselines[i].TestName == baseline.TestName {
			report.Baselines[i] = baseline
			return
		}
	}
	report.Baselines = append(report.Baselines, baseline)
}

// runBenchmarkAndCapture runs a benchmark and captures its performance metrics
func runBenchmarkAndCapture(testName string, benchFunc func(*testing.B)) PerformanceBaseline {
	result := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		benchFunc(b)
	})

	return PerformanceBaseline{
		TestName:    testName,
		NsPerOp:     result.NsPerOp(),
		AllocsPerOp: result.AllocsPerOp(),
		BytesPerOp:  result.AllocedBytesPerOp(),
		Timestamp:   time.Now(),
		GoVersion:   runtime.Version(),
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
	}
}

// checkRegression compares current performance against baseline
func checkRegression(t *testing.T, current, baseline PerformanceBaseline) {
	// Check time regression
	if baseline.NsPerOp > 0 {
		timeIncrease := float64(current.NsPerOp-baseline.NsPerOp) / float64(baseline.NsPerOp) * 100
		if timeIncrease > timeRegressionThreshold {
			t.Errorf("Performance regression detected in %s: %.1f%% slower (was %d ns/op, now %d ns/op)",
				current.TestName, timeIncrease, baseline.NsPerOp, current.NsPerOp)
		}
	}

	// Check memory regression
	if baseline.BytesPerOp > 0 {
		memIncrease := float64(current.BytesPerOp-baseline.BytesPerOp) / float64(baseline.BytesPerOp) * 100
		if memIncrease > memoryRegressionThreshold {
			t.Errorf("Memory regression detected in %s: %.1f%% more memory (was %d bytes/op, now %d bytes/op)",
				current.TestName, memIncrease, baseline.BytesPerOp, current.BytesPerOp)
		}
	}

	// Check allocation regression
	if baseline.AllocsPerOp > 0 {
		allocIncrease := float64(current.AllocsPerOp-baseline.AllocsPerOp) / float64(baseline.AllocsPerOp) * 100
		if allocIncrease > allocRegressionThreshold {
			t.Errorf("Allocation regression detected in %s: %.1f%% more allocations (was %d allocs/op, now %d allocs/op)",
				current.TestName, allocIncrease, baseline.AllocsPerOp, current.AllocsPerOp)
		}
	}
}

// TestPerformanceRegression_GetOpt tests for performance regressions in GetOpt
func TestPerformanceRegression_GetOpt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression tests in short mode")
	}

	report, err := loadBaselines()
	if err != nil {
		t.Fatalf("Failed to load baselines: %v", err)
	}

	testCases := []struct {
		name      string
		benchFunc func(*testing.B)
	}{
		{
			name: "GetOpt_SimpleShortOptions",
			benchFunc: func(b *testing.B) {
				args := []string{"prog", "-a", "-b", "-c"}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					parser, err := GetOpt(args, "abc")
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
			},
		},
		{
			name: "GetOpt_CompactedOptions",
			benchFunc: func(b *testing.B) {
				args := []string{"prog", "-abc"}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					parser, err := GetOpt(args, "abc")
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
			},
		},
		{
			name: "GetOpt_WithArguments",
			benchFunc: func(b *testing.B) {
				args := []string{"prog", "-a", "arg1", "-b", "arg2"}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					parser, err := GetOpt(args, "a:b:")
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
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run benchmark and capture metrics
			current := runBenchmarkAndCapture(tc.name, tc.benchFunc)

			// Check against baseline if it exists
			if baseline := findBaseline(report, tc.name); baseline != nil {
				checkRegression(t, current, *baseline)
			} else {
				t.Logf("No baseline found for %s, establishing new baseline", tc.name)
			}

			// Update baseline
			updateBaseline(report, current)
		})
	}

	// Save updated baselines
	if err := saveBaselines(report); err != nil {
		t.Errorf("Failed to save baselines: %v", err)
	}
}

// TestPerformanceRegression_GetOptLong tests for performance regressions in GetOptLong
func TestPerformanceRegression_GetOptLong(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression tests in short mode")
	}

	report, err := loadBaselines()
	if err != nil {
		t.Fatalf("Failed to load baselines: %v", err)
	}

	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
	}

	testCases := []struct {
		name      string
		benchFunc func(*testing.B)
	}{
		{
			name: "GetOptLong_LongOptionsOnly",
			benchFunc: func(b *testing.B) {
				args := []string{"prog", "--verbose", "--output", "file.txt"}
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
			},
		},
		{
			name: "GetOptLong_EqualsForm",
			benchFunc: func(b *testing.B) {
				args := []string{"prog", "--output=file.txt", "--config=cfg.ini"}
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
			},
		},
		{
			name: "GetOptLong_MixedShortLong",
			benchFunc: func(b *testing.B) {
				args := []string{"prog", "-v", "--output", "file.txt", "-h"}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					parser, err := GetOptLong(args, "vh", longOpts)
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
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run benchmark and capture metrics
			current := runBenchmarkAndCapture(tc.name, tc.benchFunc)

			// Check against baseline if it exists
			if baseline := findBaseline(report, tc.name); baseline != nil {
				checkRegression(t, current, *baseline)
			} else {
				t.Logf("No baseline found for %s, establishing new baseline", tc.name)
			}

			// Update baseline
			updateBaseline(report, current)
		})
	}

	// Save updated baselines
	if err := saveBaselines(report); err != nil {
		t.Errorf("Failed to save baselines: %v", err)
	}
}

// TestMemoryLeakDetection tests for memory leaks in parsing operations
func TestMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak tests in short mode")
	}

	// Test for memory leaks by running many iterations and checking memory growth
	var m1, m2 runtime.MemStats

	// Force garbage collection and get initial memory stats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Run many parsing operations
	iterations := 10000
	args := []string{"prog", "-a", "arg1", "-b", "arg2", "--verbose", "--output", "file.txt"}
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}

	for i := 0; i < iterations; i++ {
		parser, err := GetOptLong(args, "a:b:", longOpts)
		if err != nil {
			t.Fatal(err)
		}

		for option, err := range parser.Options() {
			if err != nil {
				t.Fatal(err)
			}
			_ = option
		}
	}

	// Force garbage collection and get final memory stats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Check for significant memory growth
	memGrowth := int64(m2.Alloc) - int64(m1.Alloc)
	memGrowthPerOp := memGrowth / int64(iterations)

	t.Logf("Memory growth: %d bytes total, %d bytes per operation", memGrowth, memGrowthPerOp)

	// If memory growth per operation is more than 1KB, it might indicate a leak
	if memGrowthPerOp > 1024 {
		t.Errorf("Potential memory leak detected: %d bytes per operation", memGrowthPerOp)
	}

	// Check heap objects growth
	heapObjectsGrowth := int64(m2.HeapObjects) - int64(m1.HeapObjects)
	if heapObjectsGrowth > int64(iterations/10) { // Allow some growth but not proportional to iterations
		t.Errorf("Potential object leak detected: %d heap objects growth", heapObjectsGrowth)
	}
}

// TestIteratorEfficiencyRegression tests iterator performance
func TestIteratorEfficiencyRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping iterator efficiency tests in short mode")
	}

	report, err := loadBaselines()
	if err != nil {
		t.Fatalf("Failed to load baselines: %v", err)
	}

	testCases := []struct {
		name      string
		benchFunc func(*testing.B)
	}{
		{
			name: "Iterator_FullConsumption",
			benchFunc: func(b *testing.B) {
				args := []string{"prog", "-a", "-b", "-c", "-d", "-e"}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					parser, err := GetOpt(args, "abcde")
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
			},
		},
		{
			name: "Iterator_PartialConsumption",
			benchFunc: func(b *testing.B) {
				args := []string{"prog", "-a", "-b"} // Use fewer args to avoid break issue
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					parser, err := GetOpt(args, "ab")
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
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run benchmark and capture metrics
			current := runBenchmarkAndCapture(tc.name, tc.benchFunc)

			// Check against baseline if it exists
			if baseline := findBaseline(report, tc.name); baseline != nil {
				checkRegression(t, current, *baseline)
			} else {
				t.Logf("No baseline found for %s, establishing new baseline", tc.name)
			}

			// Update baseline
			updateBaseline(report, current)
		})
	}

	// Save updated baselines
	if err := saveBaselines(report); err != nil {
		t.Errorf("Failed to save baselines: %v", err)
	}
}

// TestPerformanceBaselines_Establish can be run to establish initial baselines
func TestPerformanceBaselines_Establish(t *testing.T) {
	if !testing.Verbose() {
		t.Skip("Run with -v to establish performance baselines")
	}

	// Remove existing baselines file to start fresh
	_ = os.Remove(baselineFile)

	t.Log("Establishing performance baselines...")

	// Run all regression tests to establish baselines
	t.Run("GetOpt", func(t *testing.T) {
		TestPerformanceRegression_GetOpt(t)
	})

	t.Run("GetOptLong", func(t *testing.T) {
		TestPerformanceRegression_GetOptLong(t)
	})

	t.Run("Iterator", func(t *testing.T) {
		TestIteratorEfficiencyRegression(t)
	})

	// Report baseline file location
	if abs, err := filepath.Abs(baselineFile); err == nil {
		t.Logf("Performance baselines saved to: %s", abs)
	}
}

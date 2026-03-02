package optargs

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
	"time"
)

// PerformanceBaseline represents performance metrics for a specific test.
type PerformanceBaseline struct {
	TestName    string    `json:"test_name"`
	NsPerOp     int64     `json:"ns_per_op"`
	AllocsPerOp int64     `json:"allocs_per_op"`
	BytesPerOp  int64     `json:"bytes_per_op"`
	Timestamp   time.Time `json:"timestamp"`
	GoVersion   string    `json:"go_version"`
	GOOS        string    `json:"goos"`
	GOARCH      string    `json:"goarch"`
}

// PerformanceReport contains all baseline measurements keyed by test name.
type PerformanceReport struct {
	Baselines map[string]PerformanceBaseline `json:"baselines"`
	Generated time.Time                      `json:"generated"`
}

const (
	baselineFile = "performance_baselines.json"
	// Performance regression thresholds (percentage increase that triggers failure).
	timeRegressionThreshold   = 50.0  // 50% slower
	memoryRegressionThreshold = 100.0 // 100% more memory
	allocRegressionThreshold  = 100.0 // 100% more allocations
)

// loadBaselines loads existing performance baselines from file.
// Handles migration from the legacy array format to the current map format.
func loadBaselines() (*PerformanceReport, error) {
	data, err := os.ReadFile(baselineFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &PerformanceReport{
				Baselines: make(map[string]PerformanceBaseline),
				Generated: time.Now(),
			}, nil
		}
		return nil, err
	}

	// Try current map format first.
	var report PerformanceReport
	if err := json.Unmarshal(data, &report); err == nil && report.Baselines != nil {
		return &report, nil
	}

	// Fall back to legacy array format.
	var legacy struct {
		Baselines []PerformanceBaseline `json:"baselines"`
		Generated time.Time             `json:"generated"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, err
	}

	m := make(map[string]PerformanceBaseline, len(legacy.Baselines))
	for _, b := range legacy.Baselines {
		m[b.TestName] = b
	}
	return &PerformanceReport{Baselines: m, Generated: legacy.Generated}, nil
}

// saveBaselines saves performance baselines to file.
func saveBaselines(report *PerformanceReport) error {
	report.Generated = time.Now()
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(baselineFile, data, 0644)
}

// runBenchmarkAndCapture runs a benchmark and captures its performance metrics.
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

// checkRegression compares current performance against baseline.
func checkRegression(t *testing.T, current, baseline PerformanceBaseline) {
	t.Helper()

	if baseline.NsPerOp > 0 {
		timeIncrease := float64(current.NsPerOp-baseline.NsPerOp) / float64(baseline.NsPerOp) * 100
		if timeIncrease > timeRegressionThreshold {
			t.Errorf("Performance regression in %s: %.1f%% slower (%d → %d ns/op)",
				current.TestName, timeIncrease, baseline.NsPerOp, current.NsPerOp)
		}
	}

	if baseline.BytesPerOp > 0 {
		memIncrease := float64(current.BytesPerOp-baseline.BytesPerOp) / float64(baseline.BytesPerOp) * 100
		if memIncrease > memoryRegressionThreshold {
			t.Errorf("Memory regression in %s: %.1f%% more memory (%d → %d bytes/op)",
				current.TestName, memIncrease, baseline.BytesPerOp, current.BytesPerOp)
		}
	}

	if baseline.AllocsPerOp > 0 {
		allocIncrease := float64(current.AllocsPerOp-baseline.AllocsPerOp) / float64(baseline.AllocsPerOp) * 100
		if allocIncrease > allocRegressionThreshold {
			t.Errorf("Allocation regression in %s: %.1f%% more allocations (%d → %d allocs/op)",
				current.TestName, allocIncrease, baseline.AllocsPerOp, current.AllocsPerOp)
		}
	}
}

// regressionCase defines a named benchmark for regression testing.
type regressionCase struct {
	name      string
	benchFunc func(*testing.B)
}

// runRegressionSuite runs each benchmark case, checks against baselines,
// updates them, and saves the report.
func runRegressionSuite(t *testing.T, report *PerformanceReport, cases []regressionCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			current := runBenchmarkAndCapture(tc.name, tc.benchFunc)

			if baseline, ok := report.Baselines[tc.name]; ok {
				checkRegression(t, current, baseline)
			} else {
				t.Logf("No baseline found for %s, establishing new baseline", tc.name)
			}

			report.Baselines[tc.name] = current
		})
	}

	if err := saveBaselines(report); err != nil {
		t.Errorf("Failed to save baselines: %v", err)
	}
}

// TestPerformanceRegression_GetOpt tests for performance regressions in GetOpt.
func TestPerformanceRegression_GetOpt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression tests in short mode")
	}

	report, err := loadBaselines()
	if err != nil {
		t.Fatalf("Failed to load baselines: %v", err)
	}

	runRegressionSuite(t, report, []regressionCase{
		{
			name: "GetOpt_SimpleShortOptions",
			benchFunc: func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					benchParse(b, []string{"prog", "-a", "-b", "-c"}, "abc", nil)
				}
			},
		},
		{
			name: "GetOpt_CompactedOptions",
			benchFunc: func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					benchParse(b, []string{"prog", "-abc"}, "abc", nil)
				}
			},
		},
		{
			name: "GetOpt_WithArguments",
			benchFunc: func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					benchParse(b, []string{"prog", "-a", "arg1", "-b", "arg2"}, "a:b:", nil)
				}
			},
		},
	})
}

// TestPerformanceRegression_GetOptLong tests for performance regressions in GetOptLong.
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

	runRegressionSuite(t, report, []regressionCase{
		{
			name: "GetOptLong_LongOptionsOnly",
			benchFunc: func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					benchParse(b, []string{"prog", "--verbose", "--output", "file.txt"}, "", longOpts)
				}
			},
		},
		{
			name: "GetOptLong_EqualsForm",
			benchFunc: func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					benchParse(b, []string{"prog", "--output=file.txt", "--config=cfg.ini"}, "", longOpts)
				}
			},
		},
		{
			name: "GetOptLong_MixedShortLong",
			benchFunc: func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					benchParse(b, []string{"prog", "-v", "--output", "file.txt", "-h"}, "vh", longOpts)
				}
			},
		},
	})
}

// TestMemoryLeakDetection tests for memory leaks in parsing operations.
func TestMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak tests in short mode")
	}

	var m1, m2 runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m1)

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

	runtime.GC()
	runtime.ReadMemStats(&m2)

	memGrowth := int64(m2.Alloc) - int64(m1.Alloc)
	memGrowthPerOp := memGrowth / int64(iterations)

	t.Logf("Memory growth: %d bytes total, %d bytes per operation", memGrowth, memGrowthPerOp)

	if memGrowthPerOp > 1024 {
		t.Errorf("Potential memory leak: %d bytes per operation", memGrowthPerOp)
	}

	heapObjectsGrowth := int64(m2.HeapObjects) - int64(m1.HeapObjects)
	if heapObjectsGrowth > int64(iterations/10) {
		t.Errorf("Potential object leak: %d heap objects growth", heapObjectsGrowth)
	}
}

// TestIteratorEfficiencyRegression tests iterator performance.
func TestIteratorEfficiencyRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping iterator efficiency tests in short mode")
	}

	report, err := loadBaselines()
	if err != nil {
		t.Fatalf("Failed to load baselines: %v", err)
	}

	runRegressionSuite(t, report, []regressionCase{
		{
			name: "Iterator_FullConsumption",
			benchFunc: func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					benchParse(b, []string{"prog", "-a", "-b", "-c", "-d", "-e"}, "abcde", nil)
				}
			},
		},
		{
			name: "Iterator_PartialConsumption",
			benchFunc: func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					benchParse(b, []string{"prog", "-a", "-b"}, "ab", nil)
				}
			},
		},
	})
}

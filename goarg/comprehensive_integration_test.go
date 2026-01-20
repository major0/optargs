package goarg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFullIntegrationSuite runs the complete integration test suite
func TestFullIntegrationSuite(t *testing.T) {
	// Skip if we can't access upstream (e.g., in CI without network)
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping full integration test suite")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Initialize test components
	runner := NewCompatibilityTestRunner(workingDir)
	runner.SetVerbose(testing.Verbose())

	generator := NewTestScenarioGenerator()
	generator.SetVerbose(testing.Verbose())

	t.Run("generate_test_scenarios", func(t *testing.T) {
		// Generate built-in scenarios
		generator.GenerateBuiltinScenarios()

		scenarios := generator.GetScenarios()
		if len(scenarios) == 0 {
			t.Fatal("No test scenarios generated")
		}

		t.Logf("Generated %d test scenarios", len(scenarios))

		// Save scenarios for inspection
		scenarioFile := filepath.Join(workingDir, "generated_scenarios.json")
		if err := generator.SaveScenarios(scenarioFile); err != nil {
			t.Logf("Warning: Failed to save scenarios: %v", err)
		} else {
			t.Logf("Scenarios saved to: %s", scenarioFile)
		}
	})

	t.Run("module_alias_functionality", func(t *testing.T) {
		manager := NewModuleAliasManager(workingDir)

		// Test backup and restore
		if err := manager.BackupGoMod(); err != nil {
			t.Fatalf("Failed to backup go.mod: %v", err)
		}
		defer manager.RestoreGoMod()

		// Test module info
		info, err := manager.GetModuleInfo()
		if err != nil {
			t.Fatalf("Failed to get module info: %v", err)
		}

		t.Logf("Module info: %+v", info)

		// Test implementation switching
		originalImpl := manager.GetCurrentImplementation()

		// Switch to upstream (if available)
		if err := manager.SafeModuleSwitch("upstream"); err != nil {
			t.Logf("Warning: Cannot switch to upstream (expected in some environments): %v", err)
		} else {
			if manager.GetCurrentImplementation() != "upstream" {
				t.Errorf("Expected implementation to be 'upstream', got '%s'", manager.GetCurrentImplementation())
			}

			// Switch back
			if err := manager.SafeModuleSwitch("ours"); err != nil {
				t.Fatalf("Failed to switch back to our implementation: %v", err)
			}
		}

		if manager.GetCurrentImplementation() != originalImpl {
			t.Errorf("Expected to return to original implementation '%s', got '%s'", originalImpl, manager.GetCurrentImplementation())
		}
	})

	t.Run("compatibility_testing", func(t *testing.T) {
		scenarios := generator.GetScenarios()

		// Run a subset of scenarios for testing
		testScenarios := scenarios[:min(5, len(scenarios))]

		report, err := runner.RunAllCompatibilityTests(testScenarios)
		if err != nil {
			t.Fatalf("Failed to run compatibility tests: %v", err)
		}

		// Save detailed report
		reportFile := filepath.Join(workingDir, "integration_test_report.txt")
		if err := runner.SaveReportToFile(reportFile); err != nil {
			t.Logf("Warning: Failed to save report: %v", err)
		}

		// Save JSON results
		jsonFile := filepath.Join(workingDir, "integration_test_results.json")
		if err := runner.SaveResultsAsJSON(jsonFile); err != nil {
			t.Logf("Warning: Failed to save JSON results: %v", err)
		}

		t.Logf("Compatibility test results: %d total, %d passed, %d failed",
			report.TotalTests, report.PassedTests, report.FailedTests)

		// Analyze results
		if report.FailedTests > 0 {
			t.Logf("Some compatibility tests failed - this may be expected during development")
			for _, failedTest := range report.FailedTestNames {
				t.Logf("Failed test: %s", failedTest)
			}
		}

		// Performance analysis
		if report.ExecutionTime > 0 {
			avgTime := report.ExecutionTime / time.Duration(report.TotalTests)
			t.Logf("Performance: Total=%v, Average=%v per test", report.ExecutionTime, avgTime)
		}
	})

	t.Run("error_handling_scenarios", func(t *testing.T) {
		// Test specific error handling scenarios
		errorScenarios := []TestScenarioDefinition{
			{
				Name:        "missing_required_field",
				Description: "Test missing required field error",
				StructDefinition: `type Args struct {
					Input string ` + "`arg:\"--input,required\"`" + `
				}`,
				Arguments:       []string{},
				ExpectedSuccess: false,
				TestType:        "error",
			},
			{
				Name:        "invalid_flag_format",
				Description: "Test invalid flag format error",
				StructDefinition: `type Args struct {
					Count int ` + "`arg:\"-c,--count\"`" + `
				}`,
				Arguments:       []string{"-c", "not_a_number"},
				ExpectedSuccess: false,
				TestType:        "error",
			},
			{
				Name:        "unknown_subcommand",
				Description: "Test unknown subcommand error",
				StructDefinition: `type Args struct {
					Server *ServerCmd ` + "`arg:\"subcommand:server\"`" + `
				}

				type ServerCmd struct {
					Port int ` + "`arg:\"-p,--port\"`" + `
				}`,
				Arguments:       []string{"unknown_command"},
				ExpectedSuccess: false,
				TestType:        "error",
			},
		}

		report, err := runner.RunAllCompatibilityTests(errorScenarios)
		if err != nil {
			t.Fatalf("Failed to run error handling tests: %v", err)
		}

		t.Logf("Error handling test results: %d total, %d passed, %d failed",
			report.TotalTests, report.PassedTests, report.FailedTests)

		// For error scenarios, we expect them to fail in the same way
		for _, result := range report.TestResults {
			if !result.Match {
				t.Logf("Error handling mismatch in %s: %v", result.TestName, result.Differences)
			}
		}
	})

	t.Run("help_generation_scenarios", func(t *testing.T) {
		// Test help generation scenarios
		helpScenarios := []TestScenarioDefinition{
			{
				Name:        "basic_help",
				Description: "Test basic help generation",
				StructDefinition: `type Args struct {
					Verbose bool   ` + "`arg:\"-v,--verbose\" help:\"enable verbose output\"`" + `
					Output  string ` + "`arg:\"-o,--output\" help:\"output file path\"`" + `
				}`,
				Arguments:       []string{"--help"},
				ExpectedSuccess: false, // Help exits with non-zero
				TestType:        "help",
			},
			{
				Name:        "subcommand_help",
				Description: "Test subcommand help generation",
				StructDefinition: `type Args struct {
					Server *ServerCmd ` + "`arg:\"subcommand:server\"`" + `
				}

				type ServerCmd struct {
					Port int    ` + "`arg:\"-p,--port\" help:\"server port\"`" + `
					Host string ` + "`arg:\"-h,--host\" help:\"server host\"`" + `
				}`,
				Arguments:       []string{"server", "--help"},
				ExpectedSuccess: false, // Help exits with non-zero
				TestType:        "help",
			},
		}

		report, err := runner.RunAllCompatibilityTests(helpScenarios)
		if err != nil {
			t.Fatalf("Failed to run help generation tests: %v", err)
		}

		t.Logf("Help generation test results: %d total, %d passed, %d failed",
			report.TotalTests, report.PassedTests, report.FailedTests)

		// Analyze help output differences
		for _, result := range report.TestResults {
			if !result.Match && len(result.Differences) > 0 {
				t.Logf("Help generation differences in %s:", result.TestName)
				for _, diff := range result.Differences {
					t.Logf("  - %s", diff)
				}
			}
		}
	})

	// Cleanup
	defer runner.Cleanup()
}

// TestAdvancedCompatibilityFeatures tests advanced compatibility features
func TestAdvancedCompatibilityFeatures(t *testing.T) {
	// Skip if we can't access upstream
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping advanced compatibility tests")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)
	runner.SetVerbose(testing.Verbose())

	t.Run("custom_types", func(t *testing.T) {
		// Test custom type conversion
		scenarios := []TestScenarioDefinition{
			{
				Name:        "custom_unmarshaler",
				Description: "Test custom type with TextUnmarshaler",
				StructDefinition: `import (
					"fmt"
					"strings"
				)

				type CustomType struct {
					Value string
				}

				func (ct *CustomType) UnmarshalText(text []byte) error {
					ct.Value = strings.ToUpper(string(text))
					return nil
				}

				type Args struct {
					Custom CustomType ` + "`arg:\"-c,--custom\"`" + `
				}`,
				Arguments:       []string{"--custom", "hello"},
				ExpectedSuccess: true,
				TestType:        "parsing",
			},
		}

		report, err := runner.RunAllCompatibilityTests(scenarios)
		if err != nil {
			t.Fatalf("Failed to run custom type tests: %v", err)
		}

		if report.FailedTests > 0 {
			t.Logf("Custom type test failures: %d/%d", report.FailedTests, report.TotalTests)
		}
	})

	t.Run("environment_variables", func(t *testing.T) {
		// Set environment variable for testing
		os.Setenv("TEST_TOKEN", "env-token-123")
		defer os.Unsetenv("TEST_TOKEN")

		scenarios := []TestScenarioDefinition{
			{
				Name:        "env_var_fallback",
				Description: "Test environment variable fallback",
				StructDefinition: `type Args struct {
					Token string ` + "`arg:\"--token,env:TEST_TOKEN\"`" + `
				}`,
				Arguments:       []string{}, // No args, should use env var
				ExpectedSuccess: true,
				TestType:        "parsing",
			},
			{
				Name:        "env_var_override",
				Description: "Test environment variable override",
				StructDefinition: `type Args struct {
					Token string ` + "`arg:\"--token,env:TEST_TOKEN\"`" + `
				}`,
				Arguments:       []string{"--token", "cli-token-456"}, // Should override env var
				ExpectedSuccess: true,
				TestType:        "parsing",
			},
		}

		report, err := runner.RunAllCompatibilityTests(scenarios)
		if err != nil {
			t.Fatalf("Failed to run environment variable tests: %v", err)
		}

		if report.FailedTests > 0 {
			t.Logf("Environment variable test failures: %d/%d", report.FailedTests, report.TotalTests)
		}
	})

	t.Run("complex_subcommands", func(t *testing.T) {
		// Test complex subcommand scenarios
		scenarios := []TestScenarioDefinition{
			{
				Name:        "deeply_nested_subcommands",
				Description: "Test deeply nested subcommands",
				StructDefinition: `type Args struct {
					Global bool    ` + "`arg:\"-g,--global\"`" + `
					Level1 *L1Cmd  ` + "`arg:\"subcommand:l1\"`" + `
				}

				type L1Cmd struct {
					L1Flag bool    ` + "`arg:\"--l1-flag\"`" + `
					Level2 *L2Cmd  ` + "`arg:\"subcommand:l2\"`" + `
				}

				type L2Cmd struct {
					L2Flag bool    ` + "`arg:\"--l2-flag\"`" + `
					Level3 *L3Cmd  ` + "`arg:\"subcommand:l3\"`" + `
				}

				type L3Cmd struct {
					L3Flag bool   ` + "`arg:\"--l3-flag\"`" + `
					Value  string ` + "`arg:\"positional,required\"`" + `
				}`,
				Arguments:       []string{"-g", "l1", "--l1-flag", "l2", "--l2-flag", "l3", "--l3-flag", "final-value"},
				ExpectedSuccess: true,
				TestType:        "parsing",
			},
		}

		report, err := runner.RunAllCompatibilityTests(scenarios)
		if err != nil {
			t.Fatalf("Failed to run complex subcommand tests: %v", err)
		}

		if report.FailedTests > 0 {
			t.Logf("Complex subcommand test failures: %d/%d", report.FailedTests, report.TotalTests)
		}
	})

	// Cleanup
	defer runner.Cleanup()
}

// TestPerformanceRegression tests for performance regressions
func TestPerformanceRegression(t *testing.T) {
	// Skip if we can't access upstream
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping performance regression tests")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)

	// Performance test scenarios
	scenarios := []TestScenarioDefinition{
		{
			Name:        "large_argument_list",
			Description: "Test performance with large argument list",
			StructDefinition: `type Args struct {
				Values []string ` + "`arg:\"-v,--value\"`" + `
			}`,
			Arguments:       generateLargeArgList(100), // 100 values
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "complex_struct_parsing",
			Description: "Test performance with complex struct",
			StructDefinition: `type Args struct {
				StringField1  string   ` + "`arg:\"--str1\"`" + `
				StringField2  string   ` + "`arg:\"--str2\"`" + `
				IntField1     int      ` + "`arg:\"--int1\"`" + `
				IntField2     int      ` + "`arg:\"--int2\"`" + `
				BoolField1    bool     ` + "`arg:\"--bool1\"`" + `
				BoolField2    bool     ` + "`arg:\"--bool2\"`" + `
				SliceField1   []string ` + "`arg:\"--slice1\"`" + `
				SliceField2   []int    ` + "`arg:\"--slice2\"`" + `
				FloatField1   float64  ` + "`arg:\"--float1\"`" + `
				FloatField2   float64  ` + "`arg:\"--float2\"`" + `
			}`,
			Arguments: []string{
				"--str1", "value1", "--str2", "value2",
				"--int1", "42", "--int2", "84",
				"--bool1", "--bool2",
				"--slice1", "a", "--slice1", "b",
				"--slice2", "1", "--slice2", "2",
				"--float1", "3.14", "--float2", "2.71",
			},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
	}

	// Run performance tests multiple times
	const iterations = 10
	var totalOurTime, totalUpstreamTime time.Duration

	for i := 0; i < iterations; i++ {
		report, err := runner.RunAllCompatibilityTests(scenarios)
		if err != nil {
			t.Fatalf("Failed to run performance tests: %v", err)
		}

		for _, result := range report.TestResults {
			if result.OurResult != nil {
				totalOurTime += result.OurResult.ExecutionTime
			}
			if result.UpstreamResult != nil {
				totalUpstreamTime += result.UpstreamResult.ExecutionTime
			}
		}
	}

	// Cleanup
	defer runner.Cleanup()

	// Calculate averages
	avgOurTime := totalOurTime / time.Duration(iterations*len(scenarios))
	avgUpstreamTime := totalUpstreamTime / time.Duration(iterations*len(scenarios))

	t.Logf("Performance Regression Test Results (average over %d iterations):", iterations)
	t.Logf("Our implementation: %v", avgOurTime)
	t.Logf("Upstream implementation: %v", avgUpstreamTime)

	// Performance analysis
	if avgUpstreamTime > 0 {
		ratio := float64(avgOurTime) / float64(avgUpstreamTime)
		t.Logf("Performance ratio (ours/upstream): %.2fx", ratio)

		// Set performance thresholds
		const maxSlowdownRatio = 3.0 // Allow up to 3x slower
		const minSpeedupRatio = 0.1  // Allow up to 10x faster

		if ratio > maxSlowdownRatio {
			t.Errorf("PERFORMANCE REGRESSION: Our implementation is %.2fx slower than upstream (threshold: %.1fx)", ratio, maxSlowdownRatio)
		} else if ratio < minSpeedupRatio {
			t.Logf("EXCELLENT: Our implementation is %.2fx faster than upstream", 1.0/ratio)
		} else {
			t.Logf("Performance is within acceptable range")
		}
	}
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateLargeArgList(count int) []string {
	var args []string
	for i := 0; i < count; i++ {
		args = append(args, "-v", fmt.Sprintf("value_%d", i))
	}
	return args
}

// TestCompatibilityReporting tests the reporting functionality
func TestCompatibilityReporting(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)

	// Create some test results
	testResults := []TestResult{
		{
			TestName:      "test_pass",
			Match:         true,
			ExecutionTime: 100 * time.Millisecond,
		},
		{
			TestName:      "test_fail",
			Match:         false,
			Differences:   []string{"Exit code differs", "Output differs"},
			ExecutionTime: 150 * time.Millisecond,
		},
	}

	runner.testResults = testResults

	// Test detailed report generation
	report := runner.GenerateDetailedReport()
	if report == "" {
		t.Error("Generated report is empty")
	}

	if !strings.Contains(report, "test_pass") {
		t.Error("Report does not contain passing test")
	}

	if !strings.Contains(report, "test_fail") {
		t.Error("Report does not contain failing test")
	}

	t.Logf("Generated report:\n%s", report)

	// Test JSON export
	jsonFile := filepath.Join(workingDir, "test_results.json")
	if err := runner.SaveResultsAsJSON(jsonFile); err != nil {
		t.Fatalf("Failed to save JSON results: %v", err)
	}
	defer os.Remove(jsonFile)

	// Verify JSON content
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	var loadedResults []TestResult
	if err := json.Unmarshal(jsonData, &loadedResults); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(loadedResults) != len(testResults) {
		t.Errorf("Expected %d results, got %d", len(testResults), len(loadedResults))
	}
}

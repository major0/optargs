package goarg

import (
	"reflect"
	"testing"
	"testing/quick"
)

// TestProperty4_CompatibilityTestFrameworkCorrectness tests Property 4 from the design document:
// For any test scenario, the compatibility framework should correctly identify whether both implementations produce equivalent results
// **Validates: Requirements 3.2**
func TestProperty4_CompatibilityTestFrameworkCorrectness(t *testing.T) {
	// Property: The compatibility framework should correctly identify result equivalence
	property := func(testName string, verbose bool, count int, shouldError bool) bool {
		// Skip invalid inputs
		if testName == "" || len(testName) > 100 {
			return true
		}
		if count < 0 || count > 1000 {
			return true
		}

		// Create a test scenario with a simple struct
		scenario := TestScenario{
			Name: testName,
			StructDef: struct {
				Verbose bool `arg:"-v,--verbose"`
				Count   int  `arg:"-c,--count"`
			}{},
			Args:        generateArgsForScenario(verbose, count),
			ShouldError: shouldError,
		}

		// Create compatibility test framework
		framework := NewCompatibilityTestFramework()
		framework.AddCompatibilityTest(scenario.Name, scenario.StructDef, scenario.Args, scenario.ShouldError)

		// Test that the framework can process the scenario
		report, err := framework.RunFullCompatibilityTest()
		if err != nil {
			// Framework should not error during basic operation
			return false
		}

		// Verify report structure is correct
		if report == nil {
			return false
		}

		// Test that the framework correctly identifies identical results
		// Since we don't have upstream implementation available, we test with identical mock results
		testSuite := NewTestSuite()
		testSuite.AddScenario(scenario)

		// Create mock results that are identical
		mockResult1 := struct {
			Verbose bool
			Count   int
		}{Verbose: verbose, Count: count}

		mockResult2 := struct {
			Verbose bool
			Count   int
		}{Verbose: verbose, Count: count}

		// Test comparison logic
		comparison := testSuite.compareResults(mockResult1, mockResult2)
		if !comparison.Match {
			// Identical results should match
			return false
		}

		// Test that the framework correctly identifies different results
		mockResult3 := struct {
			Verbose bool
			Count   int
		}{Verbose: !verbose, Count: count + 1}

		comparison2 := testSuite.compareResults(mockResult1, mockResult3)
		if comparison2.Match {
			// Different results should not match
			return false
		}

		// Test report generation
		compatReport := &CompatibilityReport{}
		compatReport.AddComparison("test1", comparison)
		compatReport.AddComparison("test2", comparison2)

		// Verify report statistics
		if compatReport.TotalTests != 2 {
			return false
		}
		if compatReport.PassedTests != 1 {
			return false
		}
		if compatReport.FailedTests != 1 {
			return false
		}

		// Test report generation functionality
		reportText := framework.GenerateCompatibilityReport(compatReport)
		if reportText == "" {
			return false
		}

		return true
	}

	// Configure property test with sufficient iterations
	config := &quick.Config{
		MaxCount: 100, // Minimum 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 4 failed: %v", err)
	}
}

// generateArgsForScenario generates command line arguments for a test scenario
func generateArgsForScenario(verbose bool, count int) []string {
	var args []string
	
	if verbose {
		args = append(args, "-v")
	}
	
	if count > 0 {
		args = append(args, "-c", string(rune('0'+count%10))) // Simple single digit for testing
	}
	
	return args
}

// TestCompatibilityFrameworkResultComparison tests the core comparison logic
func TestCompatibilityFrameworkResultComparison(t *testing.T) {
	testSuite := NewTestSuite()

	// Test identical struct comparison
	result1 := struct {
		Name  string
		Value int
	}{Name: "test", Value: 42}

	result2 := struct {
		Name  string
		Value int
	}{Name: "test", Value: 42}

	comparison := testSuite.compareResults(result1, result2)
	if !comparison.Match {
		t.Error("Identical structs should match")
	}

	// Test different struct comparison
	result3 := struct {
		Name  string
		Value int
	}{Name: "different", Value: 24}

	comparison2 := testSuite.compareResults(result1, result3)
	if comparison2.Match {
		t.Error("Different structs should not match")
	}

	// Test nil comparison
	comparison3 := testSuite.compareResults(nil, nil)
	if !comparison3.Match {
		t.Error("Both nil results should match")
	}

	comparison4 := testSuite.compareResults(result1, nil)
	if comparison4.Match {
		t.Error("Struct and nil should not match")
	}
}

// TestCompatibilityReportAccuracy tests that the compatibility report accurately tracks results
func TestCompatibilityReportAccuracy(t *testing.T) {
	report := &CompatibilityReport{}

	// Add some test comparisons
	matchingResult := ScenarioResult{
		Name:           "matching_test",
		OurResult:      "same",
		UpstreamResult: "same",
		Match:          true,
	}

	differentResult := ScenarioResult{
		Name:           "different_test",
		OurResult:      "different1",
		UpstreamResult: "different2",
		Match:          false,
	}

	report.AddComparison("test1", matchingResult)
	report.AddComparison("test2", differentResult)

	// Verify statistics
	if report.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", report.TotalTests)
	}
	if report.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", report.PassedTests)
	}
	if report.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", report.FailedTests)
	}

	// Verify scenarios are stored correctly
	if len(report.Scenarios) != 2 {
		t.Errorf("Expected 2 scenarios, got %d", len(report.Scenarios))
	}

	// Check that names are set correctly
	if report.Scenarios[0].Name != "test1" {
		t.Errorf("Expected first scenario name 'test1', got '%s'", report.Scenarios[0].Name)
	}
	if report.Scenarios[1].Name != "test2" {
		t.Errorf("Expected second scenario name 'test2', got '%s'", report.Scenarios[1].Name)
	}
}

// TestFrameworkAPICompatibilityValidation tests that the framework can validate API compatibility
func TestFrameworkAPICompatibilityValidation(t *testing.T) {
	framework := NewCompatibilityTestFramework()

	// Test API compatibility validation
	err := framework.ValidateAPICompatibility()
	if err != nil {
		t.Errorf("API compatibility validation failed: %v", err)
	}

	// This test validates that our Parser struct has the expected methods
	// that match alexflint/go-arg's interface
	parserType := reflect.TypeOf(&Parser{})
	
	expectedMethods := []string{"Parse", "WriteHelp", "WriteUsage", "Fail"}
	for _, methodName := range expectedMethods {
		if _, found := parserType.MethodByName(methodName); !found {
			t.Errorf("Missing expected method: %s", methodName)
		}
	}
}
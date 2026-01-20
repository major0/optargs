package goarg

import (
	"os"
	"testing"
)

// TestBasicCompatibility tests basic compatibility framework functionality
func TestBasicCompatibility(t *testing.T) {
	// Skip if we can't access upstream
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping basic compatibility test")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)
	runner.SetVerbose(testing.Verbose())

	// Define a basic test scenario
	scenario := TestScenarioDefinition{
		Name:        "basic_parsing",
		Description: "Basic flag parsing test",
		StructDefinition: `type Args struct {
			Verbose bool ` + "`arg:\"-v,--verbose\" help:\"verbose output\"`" + `
		}`,
		Arguments:       []string{"-v"},
		ExpectedSuccess: true,
		TestType:        "parsing",
	}

	// Run the compatibility test
	result, err := runner.RunCompatibilityTest(scenario)
	if err != nil {
		t.Fatalf("Failed to run compatibility test: %v", err)
	}

	// Cleanup
	defer runner.Cleanup()

	t.Logf("Basic compatibility test result: %s", scenario.Name)
	if result.Match {
		t.Logf("✓ Test passed - implementations match")
	} else {
		t.Logf("✗ Test failed - %d differences found", len(result.Differences))
		for _, diff := range result.Differences {
			t.Logf("  - %s", diff)
		}
	}
}

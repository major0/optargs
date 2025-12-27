package goarg

import (
	"fmt"
	"reflect"
)

// CompatibilityTestFramework provides the main interface for compatibility testing
type CompatibilityTestFramework struct {
	aliasManager *ModuleAliasManager
	testSuite    *TestSuite
}

// NewCompatibilityTestFramework creates a new compatibility test framework
func NewCompatibilityTestFramework() *CompatibilityTestFramework {
	return &CompatibilityTestFramework{
		aliasManager: NewModuleAliasManager(),
		testSuite:    NewTestSuite(),
	}
}

// AddCompatibilityTest adds a test case for compatibility validation
func (ctf *CompatibilityTestFramework) AddCompatibilityTest(name string, structDef interface{}, args []string, shouldError bool) {
	scenario := TestScenario{
		Name:        name,
		StructDef:   structDef,
		Args:        args,
		ShouldError: shouldError,
	}
	ctf.testSuite.AddScenario(scenario)
}

// RunFullCompatibilityTest runs tests against both implementations
func (ctf *CompatibilityTestFramework) RunFullCompatibilityTest() (*CompatibilityReport, error) {
	// This would run tests with both implementations when upstream is available
	report := ctf.testSuite.RunCompatibilityTests()
	return report, nil
}

// ValidateAPICompatibility validates that our API matches alexflint/go-arg
func (ctf *CompatibilityTestFramework) ValidateAPICompatibility() error {
	// Check that our Parser struct has the same methods as alexflint/go-arg
	parserType := reflect.TypeOf(&Parser{})

	// Expected methods from alexflint/go-arg
	expectedMethods := []string{"Parse", "WriteHelp", "WriteUsage", "Fail"}

	for _, methodName := range expectedMethods {
		if _, found := parserType.MethodByName(methodName); !found {
			return fmt.Errorf("missing method: %s", methodName)
		}
	}

	return nil
}

// GenerateCompatibilityReport generates a detailed compatibility report
func (ctf *CompatibilityTestFramework) GenerateCompatibilityReport(report *CompatibilityReport) string {
	result := fmt.Sprintf("Compatibility Test Report\n")
	result += fmt.Sprintf("========================\n")
	result += fmt.Sprintf("Total Tests: %d\n", report.TotalTests)
	result += fmt.Sprintf("Passed: %d\n", report.PassedTests)
	result += fmt.Sprintf("Failed: %d\n", report.FailedTests)
	result += fmt.Sprintf("Success Rate: %.2f%%\n", float64(report.PassedTests)/float64(report.TotalTests)*100)

	if report.FailedTests > 0 {
		result += fmt.Sprintf("\nFailed Tests:\n")
		for _, scenario := range report.Scenarios {
			if !scenario.Match {
				result += fmt.Sprintf("- %s: Results differ\n", scenario.Name)
			}
		}
	}

	return result
}
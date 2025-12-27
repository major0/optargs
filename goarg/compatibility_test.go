package goarg

import (
	"testing"
)

// TestBasicCompatibility tests basic compatibility framework functionality
func TestBasicCompatibility(t *testing.T) {
	suite := NewTestSuite()

	// Add a basic test scenario
	suite.AddScenario(TestScenario{
		Name: "basic_parsing",
		StructDef: struct {
			Verbose bool `arg:"-v,--verbose"`
		}{},
		Args:        []string{"-v"},
		ShouldError: false,
	})

	// Run compatibility tests
	report := suite.RunCompatibilityTests()

	// Verify the framework works
	if report.TotalTests != 1 {
		t.Errorf("Expected 1 test, got %d", report.TotalTests)
	}
}
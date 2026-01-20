package goarg

import (
	"testing"
)

// TestIntegrationTestSuite runs the comprehensive integration test suite
func TestIntegrationTestSuite(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	suite.RunAllIntegrationTests()
}

// TestSliceFlagBehaviorIntegration tests slice flag handling specifically
func TestSliceFlagBehaviorIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	suite.TestSliceFlagBehavior()
}

// TestGlobalFlagInheritanceIntegration tests global flag inheritance with subcommands
func TestGlobalFlagInheritanceIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	suite.TestGlobalFlagInheritance()
}

// TestNestedSubcommandsIntegration tests complex nested subcommand structures
func TestNestedSubcommandsIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	suite.TestNestedSubcommands()
}

// TestAdvancedParsingFeaturesIntegration tests advanced parsing features
func TestAdvancedParsingFeaturesIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	suite.TestAdvancedParsingFeatures()
}

// TestErrorMessageCompatibilityIntegration tests error message format compatibility
func TestErrorMessageCompatibilityIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	suite.TestErrorMessageCompatibility()
}

// TestRealWorldScenariosIntegration tests real-world usage patterns
func TestRealWorldScenariosIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	suite.TestRealWorldScenarios()
}

// TestEndToEndWorkflowsIntegration tests complete parsing workflows
func TestEndToEndWorkflowsIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite(t)
	suite.TestEndToEndWorkflows()
}

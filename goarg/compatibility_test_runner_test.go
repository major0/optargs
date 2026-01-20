package goarg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// CompatibilityTestRunner manages running tests against both implementations
type CompatibilityTestRunner struct {
	aliasManager *ModuleAliasManager
	workingDir   string
	testResults  []TestResult
	verbose      bool
}

// TestResult represents the result of running a test against both implementations
type TestResult struct {
	TestName       string                 `json:"test_name"`
	OurResult      *ExecutionResult       `json:"our_result"`
	UpstreamResult *ExecutionResult       `json:"upstream_result"`
	Match          bool                   `json:"match"`
	Differences    []string               `json:"differences,omitempty"`
	ExecutionTime  time.Duration          `json:"execution_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionResult represents the result of executing a test with one implementation
type ExecutionResult struct {
	ParsedStruct  interface{}   `json:"parsed_struct,omitempty"`
	HelpOutput    string        `json:"help_output,omitempty"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	ExitCode      int           `json:"exit_code"`
	ExecutionTime time.Duration `json:"execution_time"`
	Success       bool          `json:"success"`
}

// TestScenarioDefinition defines a complete test scenario
type TestScenarioDefinition struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	StructDefinition string                 `json:"struct_definition"`
	Arguments        []string               `json:"arguments"`
	ExpectedSuccess  bool                   `json:"expected_success"`
	TestType         string                 `json:"test_type"` // "parsing", "help", "error"
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// NewCompatibilityTestRunner creates a new compatibility test runner
func NewCompatibilityTestRunner(workingDir string) *CompatibilityTestRunner {
	return &CompatibilityTestRunner{
		aliasManager: NewModuleAliasManager(workingDir),
		workingDir:   workingDir,
		testResults:  make([]TestResult, 0),
		verbose:      false,
	}
}

// SetVerbose enables or disables verbose output
func (ctr *CompatibilityTestRunner) SetVerbose(verbose bool) {
	ctr.verbose = verbose
}

// RunCompatibilityTest runs a single test against both implementations
func (ctr *CompatibilityTestRunner) RunCompatibilityTest(scenario TestScenarioDefinition) (*TestResult, error) {
	if ctr.verbose {
		fmt.Printf("Running compatibility test: %s\n", scenario.Name)
	}

	startTime := time.Now()

	// Run with our implementation
	ourResult, err := ctr.runWithImplementation(scenario, "ours")
	if err != nil {
		return nil, fmt.Errorf("failed to run with our implementation: %w", err)
	}

	// Run with upstream implementation
	upstreamResult, err := ctr.runWithImplementation(scenario, "upstream")
	if err != nil {
		return nil, fmt.Errorf("failed to run with upstream implementation: %w", err)
	}

	// Compare results
	match, differences := ctr.compareResults(ourResult, upstreamResult)

	result := &TestResult{
		TestName:       scenario.Name,
		OurResult:      ourResult,
		UpstreamResult: upstreamResult,
		Match:          match,
		Differences:    differences,
		ExecutionTime:  time.Since(startTime),
		Metadata:       scenario.Metadata,
	}

	ctr.testResults = append(ctr.testResults, *result)

	if ctr.verbose {
		if match {
			fmt.Printf("✓ %s: PASS\n", scenario.Name)
		} else {
			fmt.Printf("✗ %s: FAIL - %d differences\n", scenario.Name, len(differences))
			for _, diff := range differences {
				fmt.Printf("  - %s\n", diff)
			}
		}
	}

	return result, nil
}

// runWithImplementation runs a test scenario with a specific implementation
func (ctr *CompatibilityTestRunner) runWithImplementation(scenario TestScenarioDefinition, impl string) (*ExecutionResult, error) {
	// Switch to the target implementation
	if err := ctr.aliasManager.SafeModuleSwitch(impl); err != nil {
		return nil, fmt.Errorf("failed to switch to %s implementation: %w", impl, err)
	}

	// Wait for module stability
	if err := ctr.aliasManager.WaitForModuleStability(); err != nil {
		return nil, fmt.Errorf("module instability after switch to %s: %w", impl, err)
	}

	startTime := time.Now()

	// Create isolated test environment
	testDir, err := ctr.aliasManager.CreateIsolatedTestEnvironment(fmt.Sprintf("%s-%s", scenario.Name, impl))
	if err != nil {
		return nil, fmt.Errorf("failed to create isolated environment: %w", err)
	}
	defer ctr.aliasManager.CleanupIsolatedEnvironment(testDir)

	// Execute the test scenario
	result, err := ctr.executeTestScenario(scenario, testDir, impl)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test scenario: %w", err)
	}

	result.ExecutionTime = time.Since(startTime)
	return result, nil
}

// executeTestScenario executes a test scenario in the given directory
func (ctr *CompatibilityTestRunner) executeTestScenario(scenario TestScenarioDefinition, testDir, impl string) (*ExecutionResult, error) {
	// Create a separate subdirectory for the test program to avoid package conflicts
	testProgramDir := filepath.Join(testDir, "testprogram")
	if err := os.MkdirAll(testProgramDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create test program directory: %w", err)
	}

	// Create test program
	testProgram := ctr.generateTestProgram(scenario, impl)
	testFile := filepath.Join(testProgramDir, "main.go")

	if err := os.WriteFile(testFile, []byte(testProgram), 0644); err != nil {
		return nil, fmt.Errorf("failed to write test program: %w", err)
	}

	// Create a simple go.mod for the test program
	var testGoMod string
	if impl == "upstream" {
		testGoMod = `module testprogram

go 1.23.4

require github.com/alexflint/go-arg v1.4.3
`
	} else {
		// For our implementation, use local replacement that works in isolated environment
		testGoMod = fmt.Sprintf(`module testprogram

go 1.23.4

require (
	github.com/major0/optargs/goarg v0.0.0
	github.com/major0/optargs v0.0.0
)

replace github.com/major0/optargs/goarg => ../
replace github.com/major0/optargs => ../../
`)
	}

	testGoModFile := filepath.Join(testProgramDir, "go.mod")
	if err := os.WriteFile(testGoModFile, []byte(testGoMod), 0644); err != nil {
		return nil, fmt.Errorf("failed to write test go.mod: %w", err)
	}

	// Build the test program
	binaryPath := filepath.Join(testProgramDir, "test_binary")

	// Initialize module dependencies first
	if err := ctr.initializeTestModuleDependencies(testProgramDir, impl); err != nil {
		return &ExecutionResult{
			ErrorMessage: fmt.Sprintf("Module initialization failed: %s", err.Error()),
			ExitCode:     1,
			Success:      false,
		}, nil
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = testProgramDir

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		return &ExecutionResult{
			ErrorMessage: fmt.Sprintf("Build failed: %s", string(buildOutput)),
			ExitCode:     1,
			Success:      false,
		}, nil
	}

	// Execute the test program
	execCmd := exec.Command(binaryPath, scenario.Arguments...)
	execCmd.Dir = testProgramDir

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err = execCmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Parse the output based on test type
	result := &ExecutionResult{
		ExitCode: exitCode,
		Success:  err == nil,
	}

	switch scenario.TestType {
	case "parsing":
		result.ParsedStruct = ctr.parseStructOutput(stdout.String())
		if stderr.Len() > 0 {
			result.ErrorMessage = stderr.String()
		}
	case "help":
		result.HelpOutput = stdout.String()
	case "error":
		result.ErrorMessage = stderr.String()
	default:
		// Default: capture both stdout and stderr
		if stdout.Len() > 0 {
			result.ParsedStruct = ctr.parseStructOutput(stdout.String())
		}
		if stderr.Len() > 0 {
			result.ErrorMessage = stderr.String()
		}
	}

	return result, nil
}

// initializeTestModuleDependencies initializes module dependencies for the test program
func (ctr *CompatibilityTestRunner) initializeTestModuleDependencies(testProgramDir, impl string) error {
	// Download dependencies
	downloadCmd := exec.Command("go", "mod", "download")
	downloadCmd.Dir = testProgramDir
	if output, err := downloadCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod download failed: %w\nOutput: %s", err, string(output))
	}

	// Tidy the module
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = testProgramDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w\nOutput: %s", err, string(output))
	}

	// Verify dependencies
	verifyCmd := exec.Command("go", "mod", "verify")
	verifyCmd.Dir = testProgramDir
	if output, err := verifyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod verify failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// getImportPath returns the import path for the given implementation
func (ctr *CompatibilityTestRunner) getImportPath(impl string) string {
	if impl == "upstream" {
		return "github.com/alexflint/go-arg"
	}
	return "github.com/major0/optargs/goarg"
}

// generateTestProgram generates a Go program for testing a specific scenario
func (ctr *CompatibilityTestRunner) generateTestProgram(scenario TestScenarioDefinition, impl string) string {
	importPath := "github.com/major0/optargs/goarg"
	if impl == "upstream" {
		importPath = "github.com/alexflint/go-arg"
	}

	program := fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"os"
	arg "%s"
)

%s

func main() {
	var args Args

	// Parse arguments
	parser, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parser creation failed: %%v\n", err)
		os.Exit(1)
	}

	err = parser.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%%v\n", err)
		os.Exit(1)
	}

	// Output parsed structure as JSON
	output, err := json.Marshal(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON marshal error: %%v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
`, importPath, scenario.StructDefinition)

	return program
}

// parseStructOutput parses the JSON output from a test program
func (ctr *CompatibilityTestRunner) parseStructOutput(output string) interface{} {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	var result interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		// If JSON parsing fails, return the raw output
		return output
	}

	return result
}

// compareResults compares results from both implementations
func (ctr *CompatibilityTestRunner) compareResults(our, upstream *ExecutionResult) (bool, []string) {
	var differences []string

	// Compare exit codes
	if our.ExitCode != upstream.ExitCode {
		differences = append(differences, fmt.Sprintf("Exit code differs: ours=%d, upstream=%d", our.ExitCode, upstream.ExitCode))
	}

	// Compare success status
	if our.Success != upstream.Success {
		differences = append(differences, fmt.Sprintf("Success status differs: ours=%t, upstream=%t", our.Success, upstream.Success))
	}

	// Compare parsed structures
	if !reflect.DeepEqual(our.ParsedStruct, upstream.ParsedStruct) {
		differences = append(differences, "Parsed structures differ")

		// Add detailed comparison if both are non-nil
		if our.ParsedStruct != nil && upstream.ParsedStruct != nil {
			ourJSON, _ := json.MarshalIndent(our.ParsedStruct, "", "  ")
			upstreamJSON, _ := json.MarshalIndent(upstream.ParsedStruct, "", "  ")
			differences = append(differences, fmt.Sprintf("Our result:\n%s\nUpstream result:\n%s", string(ourJSON), string(upstreamJSON)))
		}
	}

	// Compare help output (character-by-character)
	if our.HelpOutput != upstream.HelpOutput {
		differences = append(differences, "Help output differs")
		if ctr.verbose {
			differences = append(differences, fmt.Sprintf("Our help:\n%s\nUpstream help:\n%s", our.HelpOutput, upstream.HelpOutput))
		}
	}

	// Compare error messages
	if our.ErrorMessage != upstream.ErrorMessage {
		differences = append(differences, "Error messages differ")
		if ctr.verbose {
			differences = append(differences, fmt.Sprintf("Our error: %s\nUpstream error: %s", our.ErrorMessage, upstream.ErrorMessage))
		}
	}

	return len(differences) == 0, differences
}

// RunAllCompatibilityTests runs all registered test scenarios
func (ctr *CompatibilityTestRunner) RunAllCompatibilityTests(scenarios []TestScenarioDefinition) (*CompatibilityReport, error) {
	if ctr.verbose {
		fmt.Printf("Running %d compatibility tests...\n", len(scenarios))
	}

	startTime := time.Now()

	// Backup current module state
	if err := ctr.aliasManager.BackupGoMod(); err != nil {
		return nil, fmt.Errorf("failed to backup go.mod: %w", err)
	}
	defer ctr.aliasManager.RestoreGoMod()

	var failedTests []string

	for i, scenario := range scenarios {
		if ctr.verbose {
			fmt.Printf("[%d/%d] ", i+1, len(scenarios))
		}

		result, err := ctr.RunCompatibilityTest(scenario)
		if err != nil {
			failedTests = append(failedTests, fmt.Sprintf("%s: %v", scenario.Name, err))
			continue
		}

		if !result.Match {
			failedTests = append(failedTests, scenario.Name)
		}
	}

	// Generate comprehensive report
	report := &CompatibilityReport{
		TotalTests:      len(scenarios),
		PassedTests:     len(scenarios) - len(failedTests),
		FailedTests:     len(failedTests),
		ExecutionTime:   time.Since(startTime),
		TestResults:     ctr.testResults,
		FailedTestNames: failedTests,
	}

	if ctr.verbose {
		fmt.Printf("\nCompatibility Test Summary:\n")
		fmt.Printf("Total: %d, Passed: %d, Failed: %d\n", report.TotalTests, report.PassedTests, report.FailedTests)
		fmt.Printf("Success Rate: %.2f%%\n", float64(report.PassedTests)/float64(report.TotalTests)*100)
		fmt.Printf("Execution Time: %v\n", report.ExecutionTime)
	}

	return report, nil
}

// GenerateDetailedReport generates a detailed compatibility report
func (ctr *CompatibilityTestRunner) GenerateDetailedReport() string {
	var report strings.Builder

	report.WriteString("Detailed Compatibility Test Report\n")
	report.WriteString("==================================\n\n")

	passed := 0
	failed := 0

	for _, result := range ctr.testResults {
		if result.Match {
			passed++
			report.WriteString(fmt.Sprintf("✓ %s (%.2fms)\n", result.TestName, float64(result.ExecutionTime.Nanoseconds())/1e6))
		} else {
			failed++
			report.WriteString(fmt.Sprintf("✗ %s (%.2fms)\n", result.TestName, float64(result.ExecutionTime.Nanoseconds())/1e6))
			for _, diff := range result.Differences {
				report.WriteString(fmt.Sprintf("  - %s\n", diff))
			}
			report.WriteString("\n")
		}
	}

	report.WriteString(fmt.Sprintf("\nSummary: %d passed, %d failed\n", passed, failed))
	if len(ctr.testResults) > 0 {
		report.WriteString(fmt.Sprintf("Success Rate: %.2f%%\n", float64(passed)/float64(len(ctr.testResults))*100))
	}

	return report.String()
}

// SaveReportToFile saves the compatibility report to a file
func (ctr *CompatibilityTestRunner) SaveReportToFile(filename string) error {
	report := ctr.GenerateDetailedReport()
	return os.WriteFile(filename, []byte(report), 0644)
}

// SaveResultsAsJSON saves the test results as JSON
func (ctr *CompatibilityTestRunner) SaveResultsAsJSON(filename string) error {
	data, err := json.MarshalIndent(ctr.testResults, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// Cleanup performs cleanup operations
func (ctr *CompatibilityTestRunner) Cleanup() error {
	return ctr.aliasManager.RestoreGoMod()
}

// ScenarioResult represents a test scenario result (for backward compatibility)
type ScenarioResult struct {
	Name           string      `json:"name"`
	OurResult      interface{} `json:"our_result"`
	UpstreamResult interface{} `json:"upstream_result"`
	Match          bool        `json:"match"`
	Error          error       `json:"error,omitempty"`
}

// Enhanced CompatibilityReport with more details
type CompatibilityReport struct {
	TotalTests      int              `json:"total_tests"`
	PassedTests     int              `json:"passed_tests"`
	FailedTests     int              `json:"failed_tests"`
	ExecutionTime   time.Duration    `json:"execution_time"`
	TestResults     []TestResult     `json:"test_results"`
	FailedTestNames []string         `json:"failed_test_names,omitempty"`
	Scenarios       []ScenarioResult `json:"scenarios,omitempty"` // For backward compatibility
}

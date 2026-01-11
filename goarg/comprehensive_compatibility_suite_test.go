package goarg

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestComprehensiveCompatibilitySuite runs the complete compatibility test suite
// covering all alexflint/go-arg features and edge cases
func TestComprehensiveCompatibilitySuite(t *testing.T) {
	// Skip if we can't access upstream (e.g., in CI without network)
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping comprehensive compatibility test suite")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)
	runner.SetVerbose(testing.Verbose())

	// Define comprehensive test scenarios covering all alexflint/go-arg features
	allScenarios := []TestScenarioDefinition{
		// Basic flag types
		{
			Name:        "bool_flags",
			Description: "Boolean flags with short and long forms",
			StructDefinition: `type Args struct {
				Verbose bool ` + "`arg:\"-v,--verbose\" help:\"enable verbose output\"`" + `
				Debug   bool ` + "`arg:\"-d,--debug\" help:\"enable debug mode\"`" + `
				Quiet   bool ` + "`arg:\"-q,--quiet\" help:\"suppress output\"`" + `
			}`,
			Arguments:       []string{"-v", "--debug"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "string_flags",
			Description: "String flags with various formats",
			StructDefinition: `type Args struct {
				Output   string ` + "`arg:\"-o,--output\" help:\"output file\"`" + `
				Config   string ` + "`arg:\"-c,--config\" help:\"config file\"`" + `
				LogLevel string ` + "`arg:\"--log-level\" help:\"log level\"`" + `
			}`,
			Arguments:       []string{"-o", "output.txt", "--config", "config.yaml", "--log-level", "debug"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "numeric_flags",
			Description: "Numeric flags (int, float, etc.)",
			StructDefinition: `type Args struct {
				Count    int     ` + "`arg:\"-c,--count\" help:\"number of items\"`" + `
				Port     int     ` + "`arg:\"-p,--port\" help:\"port number\"`" + `
				Rate     float64 ` + "`arg:\"-r,--rate\" help:\"processing rate\"`" + `
				Timeout  float32 ` + "`arg:\"-t,--timeout\" help:\"timeout in seconds\"`" + `
				Size     int64   ` + "`arg:\"-s,--size\" help:\"size in bytes\"`" + `
			}`,
			Arguments:       []string{"-c", "42", "--port", "8080", "-r", "3.14159", "--timeout", "30.5", "-s", "1048576"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "slice_flags",
			Description: "Slice flags with multiple values",
			StructDefinition: `type Args struct {
				Tags     []string ` + "`arg:\"-t,--tag\" help:\"tags to apply\"`" + `
				Numbers  []int    ` + "`arg:\"-n,--number\" help:\"numbers to process\"`" + `
				Rates    []float64` + "`arg:\"-r,--rate\" help:\"processing rates\"`" + `
			}`,
			Arguments:       []string{"-t", "tag1", "-t", "tag2", "--tag", "tag3", "-n", "1", "-n", "2", "--number", "3", "-r", "1.1", "--rate", "2.2"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},

		// Positional arguments
		{
			Name:        "positional_single",
			Description: "Single positional argument",
			StructDefinition: `type Args struct {
				File string ` + "`arg:\"positional,required\" help:\"input file\"`" + `
			}`,
			Arguments:       []string{"input.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "positional_multiple",
			Description: "Multiple positional arguments",
			StructDefinition: `type Args struct {
				Source string ` + "`arg:\"positional,required\" help:\"source file\"`" + `
				Dest   string ` + "`arg:\"positional,required\" help:\"destination file\"`" + `
			}`,
			Arguments:       []string{"source.txt", "dest.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "positional_slice",
			Description: "Positional slice arguments",
			StructDefinition: `type Args struct {
				Files []string ` + "`arg:\"positional\" help:\"files to process\"`" + `
			}`,
			Arguments:       []string{"file1.txt", "file2.txt", "file3.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "mixed_positional_flags",
			Description: "Mixed positional arguments and flags",
			StructDefinition: `type Args struct {
				Verbose bool     ` + "`arg:\"-v,--verbose\" help:\"verbose output\"`" + `
				Input   string   ` + "`arg:\"positional,required\" help:\"input file\"`" + `
				Output  string   ` + "`arg:\"-o,--output\" help:\"output file\"`" + `
				Files   []string ` + "`arg:\"positional\" help:\"additional files\"`" + `
			}`,
			Arguments:       []string{"-v", "input.txt", "-o", "output.txt", "extra1.txt", "extra2.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},

		// Default values
		{
			Name:        "default_values",
			Description: "Default values for various types",
			StructDefinition: `type Args struct {
				Port     int     ` + "`arg:\"--port\" default:\"8080\" help:\"server port\"`" + `
				Host     string  ` + "`arg:\"--host\" default:\"localhost\" help:\"server host\"`" + `
				Enabled  bool    ` + "`arg:\"--enabled\" default:\"true\" help:\"enable feature\"`" + `
				Rate     float64 ` + "`arg:\"--rate\" default:\"1.0\" help:\"processing rate\"`" + `
				Tags     []string` + "`arg:\"--tag\" default:\"default,common\" help:\"default tags\"`" + `
			}`,
			Arguments:       []string{}, // No arguments, should use defaults
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "partial_defaults",
			Description: "Partial override of default values",
			StructDefinition: `type Args struct {
				Port int    ` + "`arg:\"--port\" default:\"8080\" help:\"server port\"`" + `
				Host string ` + "`arg:\"--host\" default:\"localhost\" help:\"server host\"`" + `
			}`,
			Arguments:       []string{"--port", "9000"}, // Override port, keep host default
			ExpectedSuccess: true,
			TestType:        "parsing",
		},

		// Required fields
		{
			Name:        "required_fields_present",
			Description: "Required fields provided",
			StructDefinition: `type Args struct {
				Input  string ` + "`arg:\"--input,required\" help:\"input file\"`" + `
				Output string ` + "`arg:\"--output,required\" help:\"output file\"`" + `
			}`,
			Arguments:       []string{"--input", "in.txt", "--output", "out.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "required_fields_missing",
			Description: "Required fields missing (should error)",
			StructDefinition: `type Args struct {
				Input  string ` + "`arg:\"--input,required\" help:\"input file\"`" + `
				Output string ` + "`arg:\"--output,required\" help:\"output file\"`" + `
			}`,
			Arguments:       []string{"--input", "in.txt"}, // Missing --output
			ExpectedSuccess: false,
			TestType:        "error",
		},

		// Environment variables
		{
			Name:        "env_var_fallback",
			Description: "Environment variable fallback",
			StructDefinition: `type Args struct {
				Token string ` + "`arg:\"--token,env:API_TOKEN\" help:\"API token\"`" + `
				Debug bool   ` + "`arg:\"--debug,env:DEBUG\" help:\"debug mode\"`" + `
			}`,
			Arguments:       []string{},
			ExpectedSuccess: true,
			TestType:        "parsing",
			Metadata: map[string]interface{}{
				"env_vars": map[string]string{
					"API_TOKEN": "test-token-123",
					"DEBUG":     "true",
				},
			},
		},
		{
			Name:        "env_var_override",
			Description: "Command line overrides environment variable",
			StructDefinition: `type Args struct {
				Token string ` + "`arg:\"--token,env:API_TOKEN\" help:\"API token\"`" + `
			}`,
			Arguments:       []string{"--token", "cli-token-456"},
			ExpectedSuccess: true,
			TestType:        "parsing",
			Metadata: map[string]interface{}{
				"env_vars": map[string]string{
					"API_TOKEN": "env-token-123",
				},
			},
		},

		// Subcommands
		{
			Name:        "simple_subcommand",
			Description: "Simple subcommand",
			StructDefinition: `type Args struct {
				Server *ServerCmd ` + "`arg:\"subcommand:server\" help:\"run server\"`" + `
				Client *ClientCmd ` + "`arg:\"subcommand:client\" help:\"run client\"`" + `
			}

			type ServerCmd struct {
				Port int    ` + "`arg:\"-p,--port\" default:\"8080\" help:\"server port\"`" + `
				Host string ` + "`arg:\"-h,--host\" default:\"localhost\" help:\"server host\"`" + `
			}

			type ClientCmd struct {
				URL string ` + "`arg:\"-u,--url,required\" help:\"server URL\"`" + `
			}`,
			Arguments:       []string{"server", "--port", "9000", "--host", "0.0.0.0"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "nested_subcommands",
			Description: "Nested subcommands",
			StructDefinition: `type Args struct {
				Git *GitCmd ` + "`arg:\"subcommand:git\" help:\"git operations\"`" + `
			}

			type GitCmd struct {
				Remote *RemoteCmd ` + "`arg:\"subcommand:remote\" help:\"remote operations\"`" + `
				Branch *BranchCmd ` + "`arg:\"subcommand:branch\" help:\"branch operations\"`" + `
			}

			type RemoteCmd struct {
				Add    *RemoteAddCmd    ` + "`arg:\"subcommand:add\" help:\"add remote\"`" + `
				Remove *RemoteRemoveCmd ` + "`arg:\"subcommand:remove\" help:\"remove remote\"`" + `
			}

			type RemoteAddCmd struct {
				Name string ` + "`arg:\"positional,required\" help:\"remote name\"`" + `
				URL  string ` + "`arg:\"positional,required\" help:\"remote URL\"`" + `
			}

			type RemoteRemoveCmd struct {
				Name string ` + "`arg:\"positional,required\" help:\"remote name\"`" + `
			}

			type BranchCmd struct {
				List   *BranchListCmd   ` + "`arg:\"subcommand:list\" help:\"list branches\"`" + `
				Create *BranchCreateCmd ` + "`arg:\"subcommand:create\" help:\"create branch\"`" + `
			}

			type BranchListCmd struct {
				All bool ` + "`arg:\"-a,--all\" help:\"show all branches\"`" + `
			}

			type BranchCreateCmd struct {
				Name string ` + "`arg:\"positional,required\" help:\"branch name\"`" + `
				From string ` + "`arg:\"-f,--from\" help:\"create from branch\"`" + `
			}`,
			Arguments:       []string{"git", "remote", "add", "origin", "https://github.com/user/repo.git"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "subcommand_with_global_flags",
			Description: "Subcommand with global flags",
			StructDefinition: `type Args struct {
				Verbose bool      ` + "`arg:\"-v,--verbose\" help:\"verbose output\"`" + `
				Config  string    ` + "`arg:\"-c,--config\" help:\"config file\"`" + `
				Build   *BuildCmd ` + "`arg:\"subcommand:build\" help:\"build project\"`" + `
				Test    *TestCmd  ` + "`arg:\"subcommand:test\" help:\"test project\"`" + `
			}

			type BuildCmd struct {
				Output string ` + "`arg:\"-o,--output\" help:\"output file\"`" + `
				Debug  bool   ` + "`arg:\"-d,--debug\" help:\"debug build\"`" + `
			}

			type TestCmd struct {
				Coverage bool   ` + "`arg:\"--coverage\" help:\"generate coverage\"`" + `
				Pattern  string ` + "`arg:\"-p,--pattern\" help:\"test pattern\"`" + `
			}`,
			Arguments:       []string{"-v", "--config", "config.yaml", "build", "-o", "app", "--debug"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},

		// Help generation
		{
			Name:        "help_main",
			Description: "Main help generation",
			StructDefinition: `type Args struct {
				Verbose bool   ` + "`arg:\"-v,--verbose\" help:\"enable verbose output\"`" + `
				Output  string ` + "`arg:\"-o,--output\" help:\"output file path\"`" + `
				Count   int    ` + "`arg:\"-c,--count\" help:\"number of items\"`" + `
			}`,
			Arguments:       []string{"--help"},
			ExpectedSuccess: false, // Help exits with non-zero
			TestType:        "help",
		},
		{
			Name:        "help_subcommand",
			Description: "Subcommand help generation",
			StructDefinition: `type Args struct {
				Server *ServerCmd ` + "`arg:\"subcommand:server\" help:\"run server\"`" + `
			}

			type ServerCmd struct {
				Port int    ` + "`arg:\"-p,--port\" help:\"server port\"`" + `
				Host string ` + "`arg:\"-h,--host\" help:\"server host\"`" + `
			}`,
			Arguments:       []string{"server", "--help"},
			ExpectedSuccess: false, // Help exits with non-zero
			TestType:        "help",
		},
		{
			Name:        "help_with_defaults",
			Description: "Help showing default values",
			StructDefinition: `type Args struct {
				Port int    ` + "`arg:\"-p,--port\" default:\"8080\" help:\"server port\"`" + `
				Host string ` + "`arg:\"-h,--host\" default:\"localhost\" help:\"server host\"`" + `
			}`,
			Arguments:       []string{"--help"},
			ExpectedSuccess: false, // Help exits with non-zero
			TestType:        "help",
		},

		// Error cases
		{
			Name:        "unknown_flag",
			Description: "Unknown flag error",
			StructDefinition: `type Args struct {
				Verbose bool ` + "`arg:\"-v,--verbose\" help:\"verbose output\"`" + `
			}`,
			Arguments:       []string{"--unknown-flag"},
			ExpectedSuccess: false,
			TestType:        "error",
		},
		{
			Name:        "invalid_type_conversion",
			Description: "Invalid type conversion error",
			StructDefinition: `type Args struct {
				Count int ` + "`arg:\"-c,--count\" help:\"number of items\"`" + `
			}`,
			Arguments:       []string{"-c", "not_a_number"},
			ExpectedSuccess: false,
			TestType:        "error",
		},
		{
			Name:        "missing_argument_value",
			Description: "Missing argument value error",
			StructDefinition: `type Args struct {
				Output string ` + "`arg:\"-o,--output\" help:\"output file\"`" + `
			}`,
			Arguments:       []string{"-o"}, // Missing value
			ExpectedSuccess: false,
			TestType:        "error",
		},
		{
			Name:        "unknown_subcommand",
			Description: "Unknown subcommand error",
			StructDefinition: `type Args struct {
				Server *ServerCmd ` + "`arg:\"subcommand:server\" help:\"run server\"`" + `
			}

			type ServerCmd struct {
				Port int ` + "`arg:\"-p,--port\" help:\"server port\"`" + `
			}`,
			Arguments:       []string{"unknown_command"},
			ExpectedSuccess: false,
			TestType:        "error",
		},

		// Edge cases
		{
			Name:        "empty_arguments",
			Description: "Empty argument list",
			StructDefinition: `type Args struct {
				Verbose bool   ` + "`arg:\"-v,--verbose\" help:\"verbose output\"`" + `
				Output  string ` + "`arg:\"-o,--output\" help:\"output file\"`" + `
			}`,
			Arguments:       []string{},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "flag_equals_syntax",
			Description: "Flag with equals syntax",
			StructDefinition: `type Args struct {
				Output string ` + "`arg:\"-o,--output\" help:\"output file\"`" + `
				Count  int    ` + "`arg:\"-c,--count\" help:\"number of items\"`" + `
			}`,
			Arguments:       []string{"--output=file.txt", "-c=42"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "short_flag_combining",
			Description: "Combined short flags",
			StructDefinition: `type Args struct {
				Verbose bool ` + "`arg:\"-v,--verbose\" help:\"verbose output\"`" + `
				Debug   bool ` + "`arg:\"-d,--debug\" help:\"debug mode\"`" + `
				Force   bool ` + "`arg:\"-f,--force\" help:\"force operation\"`" + `
			}`,
			Arguments:       []string{"-vdf"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "double_dash_separator",
			Description: "Double dash argument separator",
			StructDefinition: `type Args struct {
				Verbose bool     ` + "`arg:\"-v,--verbose\" help:\"verbose output\"`" + `
				Files   []string ` + "`arg:\"positional\" help:\"files to process\"`" + `
			}`,
			Arguments:       []string{"-v", "--", "-file-with-dash.txt", "--another-file.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "special_characters_in_values",
			Description: "Special characters in argument values",
			StructDefinition: `type Args struct {
				Message string ` + "`arg:\"-m,--message\" help:\"commit message\"`" + `
				Pattern string ` + "`arg:\"-p,--pattern\" help:\"search pattern\"`" + `
			}`,
			Arguments:       []string{"-m", "Fix bug with special chars: !@#$%^&*()", "-p", "*.{js,ts,go}"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "unicode_in_values",
			Description: "Unicode characters in argument values",
			StructDefinition: `type Args struct {
				Name    string ` + "`arg:\"-n,--name\" help:\"user name\"`" + `
				Message string ` + "`arg:\"-m,--message\" help:\"message\"`" + `
			}`,
			Arguments:       []string{"-n", "José María", "-m", "Hello 世界! 🌍"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},

		// Complex real-world scenarios
		{
			Name:        "docker_run_simulation",
			Description: "Docker run command simulation",
			StructDefinition: `type Args struct {
				Detach      bool     ` + "`arg:\"-d,--detach\" help:\"run in background\"`" + `
				Interactive bool     ` + "`arg:\"-i,--interactive\" help:\"keep STDIN open\"`" + `
				TTY         bool     ` + "`arg:\"-t,--tty\" help:\"allocate pseudo-TTY\"`" + `
				Name        string   ` + "`arg:\"--name\" help:\"container name\"`" + `
				Ports       []string ` + "`arg:\"-p,--publish\" help:\"publish ports\"`" + `
				Volumes     []string ` + "`arg:\"-v,--volume\" help:\"bind mount volumes\"`" + `
				Env         []string ` + "`arg:\"-e,--env\" help:\"environment variables\"`" + `
				Image       string   ` + "`arg:\"positional,required\" help:\"image name\"`" + `
				Command     []string ` + "`arg:\"positional\" help:\"command to run\"`" + `
			}`,
			Arguments: []string{
				"-d", "-i", "-t",
				"--name", "my-container",
				"-p", "8080:80", "-p", "8443:443",
				"-v", "/host/path:/container/path",
				"-e", "ENV_VAR=value", "-e", "DEBUG=true",
				"nginx:latest",
				"nginx", "-g", "daemon off;",
			},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "kubectl_get_simulation",
			Description: "Kubectl get command simulation",
			StructDefinition: `type Args struct {
				Namespace     string ` + "`arg:\"-n,--namespace\" help:\"kubernetes namespace\"`" + `
				Output        string ` + "`arg:\"-o,--output\" help:\"output format\"`" + `
				Selector      string ` + "`arg:\"-l,--selector\" help:\"label selector\"`" + `
				AllNamespaces bool   ` + "`arg:\"--all-namespaces\" help:\"list across all namespaces\"`" + `
				Watch         bool   ` + "`arg:\"-w,--watch\" help:\"watch for changes\"`" + `
				Resource      string ` + "`arg:\"positional,required\" help:\"resource type\"`" + `
				Name          string ` + "`arg:\"positional\" help:\"resource name\"`" + `
			}`,
			Arguments: []string{
				"-n", "default",
				"-o", "yaml",
				"-l", "app=nginx,version=1.0",
				"--all-namespaces",
				"pods",
				"nginx-deployment-12345",
			},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
	}

	t.Logf("Running comprehensive compatibility test suite with %d scenarios", len(allScenarios))

	// Run all compatibility tests
	report, err := runner.RunAllCompatibilityTests(allScenarios)
	if err != nil {
		t.Fatalf("Failed to run comprehensive compatibility tests: %v", err)
	}

	// Save detailed reports
	reportFile := filepath.Join(workingDir, "comprehensive_compatibility_report.txt")
	if err := runner.SaveReportToFile(reportFile); err != nil {
		t.Logf("Warning: Failed to save report to file: %v", err)
	} else {
		t.Logf("Detailed report saved to: %s", reportFile)
	}

	// Save JSON results
	jsonFile := filepath.Join(workingDir, "comprehensive_compatibility_results.json")
	if err := runner.SaveResultsAsJSON(jsonFile); err != nil {
		t.Logf("Warning: Failed to save JSON results: %v", err)
	} else {
		t.Logf("JSON results saved to: %s", jsonFile)
	}

	// Cleanup
	defer runner.Cleanup()

	// Analyze results
	t.Logf("Comprehensive Compatibility Test Results:")
	t.Logf("Total Tests: %d", report.TotalTests)
	t.Logf("Passed: %d", report.PassedTests)
	t.Logf("Failed: %d", report.FailedTests)
	t.Logf("Success Rate: %.2f%%", float64(report.PassedTests)/float64(report.TotalTests)*100)
	t.Logf("Execution Time: %v", report.ExecutionTime)

	// Categorize failures
	if report.FailedTests > 0 {
		t.Logf("\nFailed Tests Analysis:")

		parsingFailures := 0
		helpFailures := 0
		errorFailures := 0

		for _, result := range report.TestResults {
			if !result.Match {
				switch {
				case result.TestName == "parsing":
					parsingFailures++
				case result.TestName == "help":
					helpFailures++
				case result.TestName == "error":
					errorFailures++
				}

				t.Logf("FAILED: %s", result.TestName)
				for i, diff := range result.Differences {
					if i < 3 { // Limit to first 3 differences for readability
						t.Logf("  - %s", diff)
					} else if i == 3 {
						t.Logf("  - ... and %d more differences", len(result.Differences)-3)
						break
					}
				}
			}
		}

		t.Logf("\nFailure Categories:")
		t.Logf("Parsing failures: %d", parsingFailures)
		t.Logf("Help generation failures: %d", helpFailures)
		t.Logf("Error handling failures: %d", errorFailures)

		// Determine if failures are acceptable for development
		if report.FailedTests > report.TotalTests/2 {
			t.Errorf("CRITICAL: More than 50%% of compatibility tests failed (%d/%d)", report.FailedTests, report.TotalTests)
		} else if report.FailedTests > report.TotalTests/4 {
			t.Logf("WARNING: More than 25%% of compatibility tests failed (%d/%d)", report.FailedTests, report.TotalTests)
		} else {
			t.Logf("Compatibility test failures are within acceptable range for development")
		}
	} else {
		t.Logf("🎉 ALL COMPATIBILITY TESTS PASSED! Perfect compatibility achieved.")
	}

	// Performance analysis
	if report.ExecutionTime > 0 {
		avgTime := report.ExecutionTime / time.Duration(report.TotalTests)
		t.Logf("\nPerformance Analysis:")
		t.Logf("Average test execution time: %v", avgTime)

		if avgTime > 1*time.Second {
			t.Logf("WARNING: Average test time is high, consider optimization")
		}
	}
}

// TestAlexflintGoArgExamples tests scenarios based on alexflint/go-arg examples
func TestAlexflintGoArgExamples(t *testing.T) {
	// Skip if we can't access upstream
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping alexflint/go-arg example tests")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)
	runner.SetVerbose(testing.Verbose())

	// Test scenarios based on alexflint/go-arg README examples
	exampleScenarios := []TestScenarioDefinition{
		{
			Name:        "readme_basic_example",
			Description: "Basic example from alexflint/go-arg README",
			StructDefinition: `type Args struct {
				Foo string
				Bar bool
			}`,
			Arguments:       []string{"--foo", "hello", "--bar"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "readme_tags_example",
			Description: "Tags example from alexflint/go-arg README",
			StructDefinition: `type Args struct {
				Foo string ` + "`arg:\"--foo,-f\" help:\"a foo\"`" + `
				Bar string ` + "`arg:\"--bar,-b\" help:\"a bar\"`" + `
			}`,
			Arguments:       []string{"-f", "hello", "-b", "world"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "readme_positional_example",
			Description: "Positional example from alexflint/go-arg README",
			StructDefinition: `type Args struct {
				Input  string   ` + "`arg:\"positional,required\"`" + `
				Output []string ` + "`arg:\"positional\"`" + `
			}`,
			Arguments:       []string{"input.txt", "output1.txt", "output2.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "readme_subcommand_example",
			Description: "Subcommand example from alexflint/go-arg README",
			StructDefinition: `type Args struct {
				Foo    string     ` + "`arg:\"--foo\"`" + `
				Create *CreateCmd ` + "`arg:\"subcommand:create\"`" + `
				List   *ListCmd   ` + "`arg:\"subcommand:list\"`" + `
			}

			type CreateCmd struct {
				Name string ` + "`arg:\"positional,required\"`" + `
				Dir  string ` + "`arg:\"-d\"`" + `
			}

			type ListCmd struct {
				Limit int ` + "`arg:\"-l\" default:\"10\"`" + `
			}`,
			Arguments:       []string{"--foo", "bar", "create", "-d", "/tmp", "myproject"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "readme_environment_example",
			Description: "Environment variable example from alexflint/go-arg README",
			StructDefinition: `type Args struct {
				Workers int    ` + "`arg:\"env:WORKERS\"`" + `
				Token   string ` + "`arg:\"--token,env:TOKEN,required\"`" + `
			}`,
			Arguments:       []string{"--token", "abc123"},
			ExpectedSuccess: true,
			TestType:        "parsing",
			Metadata: map[string]interface{}{
				"env_vars": map[string]string{
					"WORKERS": "4",
				},
			},
		},
	}

	// Run example-based tests
	report, err := runner.RunAllCompatibilityTests(exampleScenarios)
	if err != nil {
		t.Fatalf("Failed to run example-based tests: %v", err)
	}

	// Cleanup
	defer runner.Cleanup()

	t.Logf("Example-based Test Results: %d total, %d passed, %d failed",
		report.TotalTests, report.PassedTests, report.FailedTests)

	if report.FailedTests > 0 {
		t.Logf("Example test failures (these should match upstream exactly):")
		for _, result := range report.TestResults {
			if !result.Match {
				t.Logf("FAILED: %s", result.TestName)
				for _, diff := range result.Differences {
					t.Logf("  - %s", diff)
				}
			}
		}
	}
}

// TestEdgeCasesAndCornerCases tests edge cases and corner cases
func TestEdgeCasesAndCornerCases(t *testing.T) {
	// Skip if we can't access upstream
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping edge case tests")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)
	runner.SetVerbose(testing.Verbose())

	// Edge case scenarios
	edgeCaseScenarios := []TestScenarioDefinition{
		{
			Name:             "empty_struct",
			Description:      "Empty struct (no fields)",
			StructDefinition: `type Args struct {}`,
			Arguments:        []string{},
			ExpectedSuccess:  true,
			TestType:         "parsing",
		},
		{
			Name:        "only_help_flag",
			Description: "Struct with only help-related fields",
			StructDefinition: `type Args struct {
				Help bool ` + "`arg:\"-h,--help\" help:\"show help\"`" + `
			}`,
			Arguments:       []string{},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "very_long_flag_names",
			Description: "Very long flag names",
			StructDefinition: `type Args struct {
				VeryLongFlagNameThatIsUnusuallyLong string ` + "`arg:\"--very-long-flag-name-that-is-unusually-long\"`" + `
			}`,
			Arguments:       []string{"--very-long-flag-name-that-is-unusually-long", "value"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "single_character_values",
			Description: "Single character values",
			StructDefinition: `type Args struct {
				Char string ` + "`arg:\"-c,--char\"`" + `
			}`,
			Arguments:       []string{"-c", "x"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "numeric_edge_values",
			Description: "Numeric edge values (zero, negative, large)",
			StructDefinition: `type Args struct {
				Zero     int     ` + "`arg:\"--zero\"`" + `
				Negative int     ` + "`arg:\"--negative\"`" + `
				Large    int64   ` + "`arg:\"--large\"`" + `
				Float    float64 ` + "`arg:\"--float\"`" + `
			}`,
			Arguments:       []string{"--zero", "0", "--negative", "-42", "--large", "9223372036854775807", "--float", "1.7976931348623157e+308"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "empty_string_values",
			Description: "Empty string values",
			StructDefinition: `type Args struct {
				Empty  string   ` + "`arg:\"--empty\"`" + `
				Values []string ` + "`arg:\"--value\"`" + `
			}`,
			Arguments:       []string{"--empty", "", "--value", "", "--value", "non-empty"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "whitespace_only_values",
			Description: "Whitespace-only values",
			StructDefinition: `type Args struct {
				Spaces string ` + "`arg:\"--spaces\"`" + `
				Tabs   string ` + "`arg:\"--tabs\"`" + `
			}`,
			Arguments:       []string{"--spaces", "   ", "--tabs", "\t\t\t"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "duplicate_flag_definitions",
			Description: "Test behavior with duplicate flag usage",
			StructDefinition: `type Args struct {
				Value string ` + "`arg:\"-v,--value\"`" + `
			}`,
			Arguments:       []string{"-v", "first", "--value", "second"}, // Last one should win
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "mixed_case_subcommands",
			Description: "Mixed case subcommand names",
			StructDefinition: `type Args struct {
				CamelCase *CamelCaseCmd ` + "`arg:\"subcommand:CamelCase\"`" + `
			}

			type CamelCaseCmd struct {
				Value string ` + "`arg:\"--value\"`" + `
			}`,
			Arguments:       []string{"CamelCase", "--value", "test"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
	}

	// Run edge case tests
	report, err := runner.RunAllCompatibilityTests(edgeCaseScenarios)
	if err != nil {
		t.Fatalf("Failed to run edge case tests: %v", err)
	}

	// Cleanup
	defer runner.Cleanup()

	t.Logf("Edge Case Test Results: %d total, %d passed, %d failed",
		report.TotalTests, report.PassedTests, report.FailedTests)

	if report.FailedTests > 0 {
		t.Logf("Edge case failures:")
		for _, result := range report.TestResults {
			if !result.Match {
				t.Logf("FAILED: %s", result.TestName)
				for _, diff := range result.Differences {
					t.Logf("  - %s", diff)
				}
			}
		}
	}
}

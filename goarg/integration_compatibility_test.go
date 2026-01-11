package goarg

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestComprehensiveCompatibility runs comprehensive compatibility tests
func TestComprehensiveCompatibility(t *testing.T) {
	// Skip if we can't access upstream (e.g., in CI without network)
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping upstream compatibility tests")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)
	runner.SetVerbose(testing.Verbose())

	// Define comprehensive test scenarios
	scenarios := []TestScenarioDefinition{
		{
			Name:        "basic_flags",
			Description: "Basic short and long flags",
			StructDefinition: `type Args struct {
				Verbose bool   ` + "`arg:\"-v,--verbose\"`" + `
				Output  string ` + "`arg:\"-o,--output\"`" + `
			}`,
			Arguments:       []string{"-v", "--output", "test.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "positional_args",
			Description: "Positional arguments",
			StructDefinition: `type Args struct {
				Files []string ` + "`arg:\"positional\"`" + `
			}`,
			Arguments:       []string{"file1.txt", "file2.txt", "file3.txt"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "required_fields",
			Description: "Required fields validation",
			StructDefinition: `type Args struct {
				Input string ` + "`arg:\"--input,required\"`" + `
			}`,
			Arguments:       []string{}, // Missing required field
			ExpectedSuccess: false,
			TestType:        "error",
		},
		{
			Name:        "default_values",
			Description: "Default values",
			StructDefinition: `type Args struct {
				Port int    ` + "`arg:\"--port\" default:\"8080\"`" + `
				Host string ` + "`arg:\"--host\" default:\"localhost\"`" + `
			}`,
			Arguments:       []string{},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "environment_variables",
			Description: "Environment variable fallbacks",
			StructDefinition: `type Args struct {
				Token string ` + "`arg:\"--token,env:API_TOKEN\"`" + `
			}`,
			Arguments:       []string{},
			ExpectedSuccess: true,
			TestType:        "parsing",
			Metadata: map[string]interface{}{
				"env_vars": map[string]string{"API_TOKEN": "test-token-123"},
			},
		},
		{
			Name:        "subcommands",
			Description: "Subcommand support",
			StructDefinition: `type Args struct {
				Server *ServerCmd ` + "`arg:\"subcommand:server\"`" + `
				Client *ClientCmd ` + "`arg:\"subcommand:client\"`" + `
			}

			type ServerCmd struct {
				Port int    ` + "`arg:\"-p,--port\" default:\"8080\"`" + `
				Host string ` + "`arg:\"-h,--host\" default:\"localhost\"`" + `
			}

			type ClientCmd struct {
				URL string ` + "`arg:\"-u,--url,required\"`" + `
			}`,
			Arguments:       []string{"server", "--port", "9000", "--host", "0.0.0.0"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "type_conversion",
			Description: "Various type conversions",
			StructDefinition: `type Args struct {
				Count    int     ` + "`arg:\"-c,--count\"`" + `
				Rate     float64 ` + "`arg:\"-r,--rate\"`" + `
				Enabled  bool    ` + "`arg:\"-e,--enabled\"`" + `
				Tags     []string` + "`arg:\"-t,--tag\"`" + `
			}`,
			Arguments:       []string{"-c", "42", "-r", "3.14", "-e", "-t", "tag1", "-t", "tag2"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "help_generation",
			Description: "Help text generation",
			StructDefinition: `type Args struct {
				Verbose bool   ` + "`arg:\"-v,--verbose\" help:\"enable verbose output\"`" + `
				Output  string ` + "`arg:\"-o,--output\" help:\"output file path\"`" + `
			}`,
			Arguments:       []string{"--help"},
			ExpectedSuccess: false, // Help exits with non-zero
			TestType:        "help",
		},
		{
			Name:        "unknown_option_error",
			Description: "Unknown option error handling",
			StructDefinition: `type Args struct {
				Verbose bool ` + "`arg:\"-v,--verbose\"`" + `
			}`,
			Arguments:       []string{"--unknown-flag"},
			ExpectedSuccess: false,
			TestType:        "error",
		},
		{
			Name:        "complex_nested_subcommands",
			Description: "Complex nested subcommands with inheritance",
			StructDefinition: `type Args struct {
				Verbose bool     ` + "`arg:\"-v,--verbose\" help:\"enable verbose output\"`" + `
				Config  string   ` + "`arg:\"-c,--config\" help:\"config file path\"`" + `
				Git     *GitCmd  ` + "`arg:\"subcommand:git\"`" + `
			}

			type GitCmd struct {
				Remote *RemoteCmd ` + "`arg:\"subcommand:remote\"`" + `
				Branch *BranchCmd ` + "`arg:\"subcommand:branch\"`" + `
			}

			type RemoteCmd struct {
				Add    *RemoteAddCmd    ` + "`arg:\"subcommand:add\"`" + `
				Remove *RemoteRemoveCmd ` + "`arg:\"subcommand:remove\"`" + `
			}

			type RemoteAddCmd struct {
				Name string ` + "`arg:\"positional,required\"`" + `
				URL  string ` + "`arg:\"positional,required\"`" + `
			}

			type RemoteRemoveCmd struct {
				Name string ` + "`arg:\"positional,required\"`" + `
			}

			type BranchCmd struct {
				List   *BranchListCmd   ` + "`arg:\"subcommand:list\"`" + `
				Create *BranchCreateCmd ` + "`arg:\"subcommand:create\"`" + `
			}

			type BranchListCmd struct {
				All bool ` + "`arg:\"-a,--all\"`" + `
			}

			type BranchCreateCmd struct {
				Name string ` + "`arg:\"positional,required\"`" + `
				From string ` + "`arg:\"-f,--from\"`" + `
			}`,
			Arguments:       []string{"-v", "git", "remote", "add", "origin", "https://github.com/user/repo.git"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
	}

	// Run all compatibility tests
	report, err := runner.RunAllCompatibilityTests(scenarios)
	if err != nil {
		t.Fatalf("Failed to run compatibility tests: %v", err)
	}

	// Save detailed report
	reportFile := filepath.Join(workingDir, "compatibility_report.txt")
	if err := runner.SaveReportToFile(reportFile); err != nil {
		t.Logf("Warning: Failed to save report to file: %v", err)
	}

	// Save JSON results
	jsonFile := filepath.Join(workingDir, "compatibility_results.json")
	if err := runner.SaveResultsAsJSON(jsonFile); err != nil {
		t.Logf("Warning: Failed to save JSON results: %v", err)
	}

	// Cleanup
	defer runner.Cleanup()

	// Analyze results
	if report.FailedTests > 0 {
		t.Errorf("Compatibility test failures: %d/%d tests failed", report.FailedTests, report.TotalTests)
		t.Logf("Failed tests: %v", report.FailedTestNames)
		t.Logf("Detailed report saved to: %s", reportFile)

		// Print summary of failures
		for _, result := range report.TestResults {
			if !result.Match {
				t.Logf("FAILED: %s", result.TestName)
				for _, diff := range result.Differences {
					t.Logf("  - %s", diff)
				}
			}
		}
	} else {
		t.Logf("All compatibility tests passed! (%d/%d)", report.PassedTests, report.TotalTests)
	}

	// Performance analysis
	totalTime := report.ExecutionTime
	avgTime := totalTime / time.Duration(report.TotalTests)
	t.Logf("Performance: Total=%v, Average=%v per test", totalTime, avgTime)
}

// TestModuleAliasManagement tests the module alias management system
func TestModuleAliasManagement(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	manager := NewModuleAliasManager(workingDir)

	// Test backup and restore
	t.Run("backup_and_restore", func(t *testing.T) {
		// Backup current state
		if err := manager.BackupGoMod(); err != nil {
			t.Fatalf("Failed to backup go.mod: %v", err)
		}

		// Verify backup exists
		if _, err := os.Stat(manager.backupModFile); os.IsNotExist(err) {
			t.Errorf("Backup file was not created")
		}

		// Restore
		if err := manager.RestoreGoMod(); err != nil {
			t.Fatalf("Failed to restore go.mod: %v", err)
		}

		// Verify backup is cleaned up
		if _, err := os.Stat(manager.backupModFile); !os.IsNotExist(err) {
			t.Errorf("Backup file was not cleaned up")
		}
	})

	t.Run("implementation_switching", func(t *testing.T) {
		// Skip if we can't access upstream
		if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
			t.Skip("Skipping upstream switching tests")
		}

		// Test switching to upstream
		if err := manager.SafeModuleSwitch("upstream"); err != nil {
			t.Fatalf("Failed to switch to upstream: %v", err)
		}

		if manager.GetCurrentImplementation() != "upstream" {
			t.Errorf("Expected implementation to be 'upstream', got '%s'", manager.GetCurrentImplementation())
		}

		// Test switching back to ours
		if err := manager.SafeModuleSwitch("ours"); err != nil {
			t.Fatalf("Failed to switch back to ours: %v", err)
		}

		if manager.GetCurrentImplementation() != "ours" {
			t.Errorf("Expected implementation to be 'ours', got '%s'", manager.GetCurrentImplementation())
		}
	})

	t.Run("module_integrity", func(t *testing.T) {
		// Verify module integrity
		if err := manager.VerifyModuleIntegrity(); err != nil {
			t.Errorf("Module integrity check failed: %v", err)
		}

		// Get module info
		info, err := manager.GetModuleInfo()
		if err != nil {
			t.Fatalf("Failed to get module info: %v", err)
		}

		// Verify expected fields
		if info["module_name"] == "" {
			t.Errorf("Module name not found in info")
		}

		if info["go_version"] == "" {
			t.Errorf("Go version not found in info")
		}

		t.Logf("Module info: %+v", info)
	})
}

// TestRealWorldScenarios tests real-world usage scenarios
func TestRealWorldScenarios(t *testing.T) {
	// Skip if we can't access upstream
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping real-world scenario tests")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)
	runner.SetVerbose(testing.Verbose())

	// Real-world scenarios based on common CLI patterns
	scenarios := []TestScenarioDefinition{
		{
			Name:        "docker_like_cli",
			Description: "Docker-like CLI with subcommands and global flags",
			StructDefinition: `type Args struct {
				Debug   bool       ` + "`arg:\"-D,--debug\" help:\"Enable debug mode\"`" + `
				Host    string     ` + "`arg:\"-H,--host\" help:\"Docker daemon host\"`" + `
				Run     *RunCmd    ` + "`arg:\"subcommand:run\"`" + `
				Build   *BuildCmd  ` + "`arg:\"subcommand:build\"`" + `
				Ps      *PsCmd     ` + "`arg:\"subcommand:ps\"`" + `
			}

			type RunCmd struct {
				Detach     bool     ` + "`arg:\"-d,--detach\" help:\"Run container in background\"`" + `
				Interactive bool    ` + "`arg:\"-i,--interactive\" help:\"Keep STDIN open\"`" + `
				TTY        bool     ` + "`arg:\"-t,--tty\" help:\"Allocate a pseudo-TTY\"`" + `
				Name       string   ` + "`arg:\"--name\" help:\"Container name\"`" + `
				Ports      []string ` + "`arg:\"-p,--publish\" help:\"Publish ports\"`" + `
				Image      string   ` + "`arg:\"positional,required\" help:\"Image name\"`" + `
				Command    []string ` + "`arg:\"positional\" help:\"Command to run\"`" + `
			}

			type BuildCmd struct {
				Tag        string ` + "`arg:\"-t,--tag\" help:\"Tag for the image\"`" + `
				File       string ` + "`arg:\"-f,--file\" help:\"Dockerfile path\"`" + `
				Context    string ` + "`arg:\"positional\" default:\".\" help:\"Build context\"`" + `
			}

			type PsCmd struct {
				All   bool ` + "`arg:\"-a,--all\" help:\"Show all containers\"`" + `
				Quiet bool ` + "`arg:\"-q,--quiet\" help:\"Only show container IDs\"`" + `
			}`,
			Arguments:       []string{"--debug", "run", "-d", "-p", "8080:80", "--name", "webserver", "nginx:latest"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "git_like_cli",
			Description: "Git-like CLI with complex subcommand structure",
			StructDefinition: `type Args struct {
				Version bool      ` + "`arg:\"--version\" help:\"Show version\"`" + `
				Config  string    ` + "`arg:\"-c,--config\" help:\"Config file\"`" + `
				Add     *AddCmd   ` + "`arg:\"subcommand:add\"`" + `
				Commit  *CommitCmd` + "`arg:\"subcommand:commit\"`" + `
				Push    *PushCmd  ` + "`arg:\"subcommand:push\"`" + `
				Remote  *RemoteCmd` + "`arg:\"subcommand:remote\"`" + `
			}

			type AddCmd struct {
				All   bool     ` + "`arg:\"-A,--all\" help:\"Add all files\"`" + `
				Force bool     ` + "`arg:\"-f,--force\" help:\"Force add\"`" + `
				Files []string ` + "`arg:\"positional\" help:\"Files to add\"`" + `
			}

			type CommitCmd struct {
				Message string ` + "`arg:\"-m,--message,required\" help:\"Commit message\"`" + `
				All     bool   ` + "`arg:\"-a,--all\" help:\"Commit all changes\"`" + `
				Amend   bool   ` + "`arg:\"--amend\" help:\"Amend previous commit\"`" + `
			}

			type PushCmd struct {
				Force  bool   ` + "`arg:\"-f,--force\" help:\"Force push\"`" + `
				SetUpstream bool ` + "`arg:\"-u,--set-upstream\" help:\"Set upstream\"`" + `
				Remote string ` + "`arg:\"positional\" default:\"origin\" help:\"Remote name\"`" + `
				Branch string ` + "`arg:\"positional\" help:\"Branch name\"`" + `
			}

			type RemoteCmd struct {
				Add    *RemoteAddCmd    ` + "`arg:\"subcommand:add\"`" + `
				Remove *RemoteRemoveCmd ` + "`arg:\"subcommand:remove\"`" + `
				Show   *RemoteShowCmd   ` + "`arg:\"subcommand:show\"`" + `
			}

			type RemoteAddCmd struct {
				Name string ` + "`arg:\"positional,required\" help:\"Remote name\"`" + `
				URL  string ` + "`arg:\"positional,required\" help:\"Remote URL\"`" + `
			}

			type RemoteRemoveCmd struct {
				Name string ` + "`arg:\"positional,required\" help:\"Remote name\"`" + `
			}

			type RemoteShowCmd struct {
				Name string ` + "`arg:\"positional,required\" help:\"Remote name\"`" + `
			}`,
			Arguments:       []string{"commit", "-m", "Initial commit", "--all"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "kubectl_like_cli",
			Description: "Kubernetes kubectl-like CLI with resource management",
			StructDefinition: `type Args struct {
				Namespace  string     ` + "`arg:\"-n,--namespace\" help:\"Kubernetes namespace\"`" + `
				Kubeconfig string     ` + "`arg:\"--kubeconfig\" help:\"Path to kubeconfig file\"`" + `
				Context    string     ` + "`arg:\"--context\" help:\"Kubernetes context\"`" + `
				Get        *GetCmd    ` + "`arg:\"subcommand:get\"`" + `
				Apply      *ApplyCmd  ` + "`arg:\"subcommand:apply\"`" + `
				Delete     *DeleteCmd ` + "`arg:\"subcommand:delete\"`" + `
				Logs       *LogsCmd   ` + "`arg:\"subcommand:logs\"`" + `
			}

			type GetCmd struct {
				Output    string   ` + "`arg:\"-o,--output\" help:\"Output format\"`" + `
				Selector  string   ` + "`arg:\"-l,--selector\" help:\"Label selector\"`" + `
				AllNamespaces bool ` + "`arg:\"--all-namespaces\" help:\"List across all namespaces\"`" + `
				Resource  string   ` + "`arg:\"positional,required\" help:\"Resource type\"`" + `
				Name      string   ` + "`arg:\"positional\" help:\"Resource name\"`" + `
			}

			type ApplyCmd struct {
				Filename  []string ` + "`arg:\"-f,--filename\" help:\"Filename or directory\"`" + `
				Recursive bool     ` + "`arg:\"-R,--recursive\" help:\"Process directory recursively\"`" + `
				DryRun    string   ` + "`arg:\"--dry-run\" help:\"Dry run mode\"`" + `
			}

			type DeleteCmd struct {
				Filename  []string ` + "`arg:\"-f,--filename\" help:\"Filename or directory\"`" + `
				Selector  string   ` + "`arg:\"-l,--selector\" help:\"Label selector\"`" + `
				All       bool     ` + "`arg:\"--all\" help:\"Delete all resources\"`" + `
				Resource  string   ` + "`arg:\"positional\" help:\"Resource type\"`" + `
				Name      string   ` + "`arg:\"positional\" help:\"Resource name\"`" + `
			}

			type LogsCmd struct {
				Follow     bool   ` + "`arg:\"-f,--follow\" help:\"Follow log output\"`" + `
				Previous   bool   ` + "`arg:\"-p,--previous\" help:\"Show previous container logs\"`" + `
				Tail       int    ` + "`arg:\"--tail\" help:\"Number of lines to show\"`" + `
				Container  string ` + "`arg:\"-c,--container\" help:\"Container name\"`" + `
				Pod        string ` + "`arg:\"positional,required\" help:\"Pod name\"`" + `
			}`,
			Arguments:       []string{"-n", "default", "get", "pods", "-o", "yaml", "--all-namespaces"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
	}

	// Run real-world scenario tests
	report, err := runner.RunAllCompatibilityTests(scenarios)
	if err != nil {
		t.Fatalf("Failed to run real-world scenario tests: %v", err)
	}

	// Cleanup
	defer runner.Cleanup()

	// Analyze results
	if report.FailedTests > 0 {
		t.Errorf("Real-world scenario test failures: %d/%d tests failed", report.FailedTests, report.TotalTests)

		// Print detailed failure information
		for _, result := range report.TestResults {
			if !result.Match {
				t.Logf("FAILED: %s", result.TestName)
				for _, diff := range result.Differences {
					t.Logf("  - %s", diff)
				}
			}
		}
	} else {
		t.Logf("All real-world scenario tests passed! (%d/%d)", report.PassedTests, report.TotalTests)
	}
}

// TestPerformanceComparison tests performance characteristics between implementations
func TestPerformanceComparison(t *testing.T) {
	// Skip if we can't access upstream
	if os.Getenv("SKIP_UPSTREAM_TESTS") == "true" {
		t.Skip("Skipping performance comparison tests")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	runner := NewCompatibilityTestRunner(workingDir)

	// Performance test scenarios
	scenarios := []TestScenarioDefinition{
		{
			Name:        "simple_parsing_perf",
			Description: "Simple parsing performance test",
			StructDefinition: `type Args struct {
				Verbose bool   ` + "`arg:\"-v,--verbose\"`" + `
				Output  string ` + "`arg:\"-o,--output\"`" + `
				Count   int    ` + "`arg:\"-c,--count\"`" + `
			}`,
			Arguments:       []string{"-v", "--output", "test.txt", "--count", "100"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
		{
			Name:        "complex_parsing_perf",
			Description: "Complex parsing performance test",
			StructDefinition: `type Args struct {
				GlobalFlag bool      ` + "`arg:\"-g,--global\"`" + `
				Config     string    ` + "`arg:\"-c,--config\"`" + `
				Cmd        *CmdType  ` + "`arg:\"subcommand:cmd\"`" + `
			}

			type CmdType struct {
				SubFlag bool     ` + "`arg:\"-s,--sub\"`" + `
				Values  []string ` + "`arg:\"-v,--value\"`" + `
				Files   []string ` + "`arg:\"positional\"`" + `
			}`,
			Arguments:       []string{"-g", "--config", "config.yaml", "cmd", "-s", "-v", "val1", "-v", "val2", "file1", "file2"},
			ExpectedSuccess: true,
			TestType:        "parsing",
		},
	}

	// Run performance tests multiple times
	const iterations = 5
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

	t.Logf("Performance Comparison (average over %d iterations):", iterations)
	t.Logf("Our implementation: %v", avgOurTime)
	t.Logf("Upstream implementation: %v", avgUpstreamTime)

	// Performance ratio
	if avgUpstreamTime > 0 {
		ratio := float64(avgOurTime) / float64(avgUpstreamTime)
		t.Logf("Performance ratio (ours/upstream): %.2fx", ratio)

		// Warn if we're significantly slower
		if ratio > 2.0 {
			t.Logf("WARNING: Our implementation is %.2fx slower than upstream", ratio)
		} else if ratio < 0.5 {
			t.Logf("EXCELLENT: Our implementation is %.2fx faster than upstream", 1.0/ratio)
		}
	}
}

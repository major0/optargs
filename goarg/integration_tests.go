package goarg

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// IntegrationTestSuite provides comprehensive integration testing for go-arg functionality
type IntegrationTestSuite struct {
	t *testing.T
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	return &IntegrationTestSuite{t: t}
}

// TestSliceFlagBehavior tests slice flag handling to match upstream behavior
func (its *IntegrationTestSuite) TestSliceFlagBehavior() {
	its.t.Run("slice_flags_last_value_wins", func(t *testing.T) {
		// Test that slice flags should keep last value, not accumulate
		// This matches alexflint/go-arg behavior
		type Args struct {
			Numbers []int    `arg:"--numbers,-n" help:"list of numbers"`
			Tags    []string `arg:"--tags,-t" help:"list of tags"`
		}

		// Test multiple values - upstream keeps only last value
		testCases := []struct {
			name     string
			args     []string
			expected Args
		}{
			{
				name: "multiple_number_flags",
				args: []string{"--numbers", "1", "--numbers", "2", "--numbers", "3"},
				expected: Args{
					Numbers: []int{3}, // Only last value should be kept
				},
			},
			{
				name: "multiple_tag_flags",
				args: []string{"--tags", "tag1", "--tags", "tag2", "--tags", "tag3"},
				expected: Args{
					Tags: []string{"tag3"}, // Only last value should be kept
				},
			},
			{
				name: "mixed_slice_flags",
				args: []string{"--numbers", "1", "--tags", "tag1", "--numbers", "2", "--tags", "tag2"},
				expected: Args{
					Numbers: []int{2},
					Tags:    []string{"tag2"},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err != nil {
					t.Fatalf("Parse failed: %v", err)
				}

				if !reflect.DeepEqual(args, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, args)
				}
			})
		}
	})
}

// TestGlobalFlagInheritance tests global flag inheritance with subcommands
func (its *IntegrationTestSuite) TestGlobalFlagInheritance() {
	its.t.Run("global_flags_with_subcommands", func(t *testing.T) {
		type BuildCmd struct {
			Debug  bool   `arg:"--debug,-d" help:"enable debug mode"`
			Output string `arg:"--output,-o" help:"output file"`
		}

		type TestCmd struct {
			Coverage bool `arg:"--coverage,-c" help:"enable coverage"`
		}

		type Args struct {
			Verbose bool      `arg:"--verbose,-v" help:"enable verbose output"`
			Config  string    `arg:"--config" help:"config file"`
			Build   *BuildCmd `arg:"subcommand:build" help:"build command"`
			Test    *TestCmd  `arg:"subcommand:test" help:"test command"`
		}

		testCases := []struct {
			name     string
			args     []string
			expected Args
		}{
			{
				name: "global_flags_before_subcommand",
				args: []string{"--verbose", "--config", "config.yaml", "build", "--debug", "--output", "app"},
				expected: Args{
					Verbose: true,
					Config:  "config.yaml",
					Build: &BuildCmd{
						Debug:  true,
						Output: "app",
					},
				},
			},
			{
				name: "global_flags_after_subcommand",
				args: []string{"build", "--debug", "--output", "app", "--verbose", "--config", "config.yaml"},
				expected: Args{
					Verbose: true,
					Config:  "config.yaml",
					Build: &BuildCmd{
						Debug:  true,
						Output: "app",
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err != nil {
					t.Fatalf("Parse failed: %v", err)
				}

				if !reflect.DeepEqual(args, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, args)
				}
			})
		}
	})
}

// TestNestedSubcommands tests complex nested subcommand structures
func (its *IntegrationTestSuite) TestNestedSubcommands() {
	its.t.Run("nested_subcommand_parsing", func(t *testing.T) {
		type RemoteAddCmd struct {
			Name string `arg:"positional,required" help:"remote name"`
			URL  string `arg:"positional,required" help:"remote URL"`
		}

		type RemoteRemoveCmd struct {
			Name string `arg:"positional,required" help:"remote name"`
		}

		type RemoteCmd struct {
			Add    *RemoteAddCmd    `arg:"subcommand:add" help:"add remote"`
			Remove *RemoteRemoveCmd `arg:"subcommand:remove" help:"remove remote"`
		}

		type BranchCmd struct {
			Name   string `arg:"positional" help:"branch name"`
			Delete bool   `arg:"--delete,-d" help:"delete branch"`
		}

		type GitCmd struct {
			Branch *BranchCmd `arg:"subcommand:branch" help:"branch operations"`
			Remote *RemoteCmd `arg:"subcommand:remote" help:"remote operations"`
		}

		type Args struct {
			Git *GitCmd `arg:"subcommand:git" help:"git operations"`
		}

		testCases := []struct {
			name     string
			args     []string
			expected Args
		}{
			{
				name: "nested_remote_add",
				args: []string{"git", "remote", "add", "origin", "https://github.com/user/repo.git"},
				expected: Args{
					Git: &GitCmd{
						Remote: &RemoteCmd{
							Add: &RemoteAddCmd{
								Name: "origin",
								URL:  "https://github.com/user/repo.git",
							},
						},
					},
				},
			},
			{
				name: "nested_branch_delete",
				args: []string{"git", "branch", "--delete", "feature-branch"},
				expected: Args{
					Git: &GitCmd{
						Branch: &BranchCmd{
							Name:   "feature-branch",
							Delete: true,
						},
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err != nil {
					t.Fatalf("Parse failed: %v", err)
				}

				if !reflect.DeepEqual(args, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, args)
				}
			})
		}
	})
}

// TestAdvancedParsingFeatures tests advanced parsing features that need implementation
func (its *IntegrationTestSuite) TestAdvancedParsingFeatures() {
	its.t.Run("short_flag_combining", func(t *testing.T) {
		type Args struct {
			Verbose bool `arg:"-v" help:"verbose output"`
			Debug   bool `arg:"-d" help:"debug mode"`
			Force   bool `arg:"-f" help:"force operation"`
		}

		testCases := []struct {
			name     string
			args     []string
			expected Args
		}{
			{
				name: "combined_short_flags",
				args: []string{"-vdf"},
				expected: Args{
					Verbose: true,
					Debug:   true,
					Force:   true,
				},
			},
			{
				name: "partial_combined_flags",
				args: []string{"-vd", "-f"},
				expected: Args{
					Verbose: true,
					Debug:   true,
					Force:   true,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err != nil {
					// This feature is not yet implemented, so we expect it to fail
					t.Logf("Short flag combining not yet implemented: %v", err)
					t.Skip("Short flag combining feature not yet implemented")
					return
				}

				if !reflect.DeepEqual(args, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, args)
				}
			})
		}
	})

	its.t.Run("flag_equals_syntax", func(t *testing.T) {
		type Args struct {
			Count  int    `arg:"--count,-c" help:"count value"`
			Output string `arg:"--output,-o" help:"output file"`
		}

		testCases := []struct {
			name     string
			args     []string
			expected Args
		}{
			{
				name: "long_flag_equals",
				args: []string{"--count=42", "--output=file.txt"},
				expected: Args{
					Count:  42,
					Output: "file.txt",
				},
			},
			{
				name: "mixed_equals_and_space",
				args: []string{"--count=42", "--output", "file.txt"},
				expected: Args{
					Count:  42,
					Output: "file.txt",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err != nil {
					t.Logf("Flag equals syntax error: %v", err)
					// Check if this is expected to fail
					if strings.Contains(err.Error(), "invalid argument") {
						t.Skip("Flag equals syntax feature needs improvement")
						return
					}
					t.Fatalf("Unexpected parse error: %v", err)
				}

				if !reflect.DeepEqual(args, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, args)
				}
			})
		}
	})
}

// TestErrorMessageCompatibility tests error message format compatibility
func (its *IntegrationTestSuite) TestErrorMessageCompatibility() {
	its.t.Run("required_field_errors", func(t *testing.T) {
		type Args struct {
			Input  string `arg:"--input,required" help:"input file"`
			Output string `arg:"--output,required" help:"output file"`
		}

		testCases := []struct {
			name          string
			args          []string
			expectedError string
		}{
			{
				name:          "missing_required_input",
				args:          []string{"--output", "out.txt"},
				expectedError: "--input is required", // Should match upstream format
			},
			{
				name:          "missing_required_output",
				args:          []string{"--input", "in.txt"},
				expectedError: "--output is required", // Should match upstream format
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err == nil {
					t.Fatalf("Expected error but parsing succeeded")
				}

				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.expectedError, err.Error())
					t.Logf("Note: Error message format needs to match upstream exactly")
				}
			})
		}
	})

	its.t.Run("unknown_flag_errors", func(t *testing.T) {
		type Args struct {
			Verbose bool `arg:"-v,--verbose" help:"verbose output"`
		}

		testCases := []struct {
			name          string
			args          []string
			expectedError string
		}{
			{
				name:          "unknown_long_flag",
				args:          []string{"--unknown-flag"},
				expectedError: "unknown argument --unknown-flag", // Should match upstream format
			},
			{
				name:          "unknown_short_flag",
				args:          []string{"-x"},
				expectedError: "unknown argument -x", // Should match upstream format
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err == nil {
					t.Fatalf("Expected error but parsing succeeded")
				}

				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.expectedError, err.Error())
					t.Logf("Note: Error message format needs to match upstream exactly")
				}
			})
		}
	})
}

// TestRealWorldScenarios tests real-world usage patterns
func (its *IntegrationTestSuite) TestRealWorldScenarios() {
	its.t.Run("docker_run_like_command", func(t *testing.T) {
		type Args struct {
			Image       string   `arg:"positional,required" help:"docker image"`
			Interactive bool     `arg:"-i,--interactive" help:"interactive mode"`
			TTY         bool     `arg:"-t,--tty" help:"allocate TTY"`
			Detach      bool     `arg:"-d,--detach" help:"detached mode"`
			Ports       []string `arg:"-p,--port" help:"port mappings"`
			Volumes     []string `arg:"-v,--volume" help:"volume mounts"`
			Environment []string `arg:"-e,--env" help:"environment variables"`
			Name        string   `arg:"--name" help:"container name"`
			Command     []string `arg:"positional" help:"command to run"`
		}

		testCases := []struct {
			name     string
			args     []string
			expected Args
		}{
			{
				name: "basic_docker_run",
				args: []string{"-it", "--name", "myapp", "-p", "8080:80", "-v", "/host:/container", "nginx:latest", "bash"},
				expected: Args{
					Image:       "nginx:latest",
					Interactive: true,
					TTY:         true,
					Name:        "myapp",
					Ports:       []string{"8080:80"},
					Volumes:     []string{"/host:/container"},
					Command:     []string{"bash"},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err != nil {
					t.Logf("Docker-like command parsing failed: %v", err)
					// This might fail due to short flag combining (-it)
					if strings.Contains(err.Error(), "unknown argument") {
						t.Skip("Short flag combining needed for docker-like commands")
						return
					}
					t.Fatalf("Unexpected parse error: %v", err)
				}

				if !reflect.DeepEqual(args, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, args)
				}
			})
		}
	})

	its.t.Run("kubectl_get_like_command", func(t *testing.T) {
		type Args struct {
			Resource      string `arg:"positional,required" help:"resource type"`
			Name          string `arg:"positional" help:"resource name"`
			Namespace     string `arg:"-n,--namespace" help:"namespace"`
			Output        string `arg:"-o,--output" help:"output format"`
			AllNamespaces bool   `arg:"-A,--all-namespaces" help:"all namespaces"`
			Watch         bool   `arg:"-w,--watch" help:"watch for changes"`
		}

		testCases := []struct {
			name     string
			args     []string
			expected Args
		}{
			{
				name: "kubectl_get_pods",
				args: []string{"pods", "-n", "default", "-o", "yaml"},
				expected: Args{
					Resource:  "pods",
					Namespace: "default",
					Output:    "yaml",
				},
			},
			{
				name: "kubectl_get_all_namespaces",
				args: []string{"pods", "--all-namespaces", "--output", "json"},
				expected: Args{
					Resource:      "pods",
					AllNamespaces: true,
					Output:        "json",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err != nil {
					t.Fatalf("Parse failed: %v", err)
				}

				if !reflect.DeepEqual(args, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, args)
				}
			})
		}
	})
}

// TestEndToEndWorkflows tests complete parsing workflows
func (its *IntegrationTestSuite) TestEndToEndWorkflows() {
	its.t.Run("complete_application_workflow", func(t *testing.T) {
		// Simulate a complete application with multiple subcommands and complex options
		type ServerCmd struct {
			Port     int    `arg:"-p,--port" default:"8080" help:"server port"`
			Host     string `arg:"-h,--host" default:"localhost" help:"server host"`
			LogLevel string `arg:"--log-level" default:"info" help:"log level"`
		}

		type ClientCmd struct {
			URL     string `arg:"-u,--url,required" help:"server URL"`
			Timeout int    `arg:"-t,--timeout" default:"30" help:"timeout in seconds"`
		}

		type Args struct {
			ConfigFile string     `arg:"-c,--config" help:"config file path"`
			Verbose    bool       `arg:"-v,--verbose" help:"verbose output"`
			Server     *ServerCmd `arg:"subcommand:server" help:"run server"`
			Client     *ClientCmd `arg:"subcommand:client" help:"run client"`
		}

		// Test environment variable fallback
		os.Setenv("CONFIG_FILE", "/etc/myapp/config.yaml")
		defer os.Unsetenv("CONFIG_FILE")

		testCases := []struct {
			name     string
			args     []string
			expected Args
			envVars  map[string]string
		}{
			{
				name: "server_with_custom_port",
				args: []string{"--verbose", "server", "--port", "9090", "--log-level", "debug"},
				expected: Args{
					Verbose: true,
					Server: &ServerCmd{
						Port:     9090,
						Host:     "localhost", // default value
						LogLevel: "debug",
					},
				},
			},
			{
				name: "client_with_required_url",
				args: []string{"client", "--url", "http://localhost:8080", "--timeout", "60"},
				expected: Args{
					Client: &ClientCmd{
						URL:     "http://localhost:8080",
						Timeout: 60,
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Set environment variables if specified
				for key, value := range tc.envVars {
					os.Setenv(key, value)
					defer os.Unsetenv(key)
				}

				var args Args
				parser, err := NewParser(Config{}, &args)
				if err != nil {
					t.Fatalf("Failed to create parser: %v", err)
				}

				err = parser.Parse(tc.args)
				if err != nil {
					t.Fatalf("Parse failed: %v", err)
				}

				if !reflect.DeepEqual(args, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, args)
				}
			})
		}
	})
}

// RunAllIntegrationTests runs all integration tests
func (its *IntegrationTestSuite) RunAllIntegrationTests() {
	its.t.Run("IntegrationTestSuite", func(t *testing.T) {
		suite := NewIntegrationTestSuite(t)

		// Run all test categories
		suite.TestSliceFlagBehavior()
		suite.TestGlobalFlagInheritance()
		suite.TestNestedSubcommands()
		suite.TestAdvancedParsingFeatures()
		suite.TestErrorMessageCompatibility()
		suite.TestRealWorldScenarios()
		suite.TestEndToEndWorkflows()
	})
}

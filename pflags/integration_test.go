package pflags

import (
	"testing"
	"time"
)

// TestE2ECobraStyleFlagSet tests a Cobra-style pattern where a root FlagSet
// defines persistent flags and a command FlagSet defines local flags.
func TestE2ECobraStyleFlagSet(t *testing.T) {
	// Root "persistent" flags
	root := NewFlagSet("app", ContinueOnError)
	var verbose bool
	var config string
	root.BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	root.StringVar(&config, "config", "", "config file path")

	// Command-local flags
	serve := NewFlagSet("serve", ContinueOnError)
	var port int
	var host string
	var timeout time.Duration
	serve.IntVarP(&port, "port", "p", 8080, "listen port")
	serve.StringVar(&host, "host", "localhost", "listen host")
	serve.DurationVar(&timeout, "timeout", 30*time.Second, "request timeout")

	// Parse root flags
	if err := root.Parse([]string{"-v", "--config", "/etc/app.yml"}); err != nil {
		t.Fatalf("root parse: %v", err)
	}
	if !verbose {
		t.Error("verbose should be true")
	}
	if config != "/etc/app.yml" {
		t.Errorf("config = %q", config)
	}

	// Parse command flags
	if err := serve.Parse([]string{"-p", "9090", "--host", "0.0.0.0", "--timeout", "1m"}); err != nil {
		t.Fatalf("serve parse: %v", err)
	}
	if port != 9090 {
		t.Errorf("port = %d", port)
	}
	if host != "0.0.0.0" {
		t.Errorf("host = %q", host)
	}
	if timeout != time.Minute {
		t.Errorf("timeout = %v", timeout)
	}
}

// TestE2EMixedFlagTypes tests a realistic scenario with all flag types.
func TestE2EMixedFlagTypes(t *testing.T) {
	fs := NewFlagSet("deploy", ContinueOnError)

	var (
		env      string
		replicas int
		dryRun   bool
		cpuLimit float64
		timeout  time.Duration
		tags     []string
		ports    []int
	)

	fs.StringVarP(&env, "env", "e", "staging", "deployment environment")
	fs.IntVarP(&replicas, "replicas", "r", 1, "number of replicas")
	fs.BoolVar(&dryRun, "dry-run", false, "dry run mode")
	fs.Float64Var(&cpuLimit, "cpu-limit", 1.0, "CPU limit")
	fs.DurationVar(&timeout, "timeout", 5*time.Minute, "deployment timeout")
	fs.StringSliceVar(&tags, "tag", nil, "deployment tags")
	fs.IntSliceVar(&ports, "port", nil, "exposed ports")

	args := []string{
		"-e", "production",
		"--replicas", "3",
		"--dry-run",
		"--cpu-limit", "2.5",
		"--timeout", "10m",
		"--tag", "v1.0,latest",
		"--port", "80,443",
	}

	if err := fs.Parse(args); err != nil {
		t.Fatalf("parse: %v", err)
	}

	if env != "production" {
		t.Errorf("env = %q", env)
	}
	if replicas != 3 {
		t.Errorf("replicas = %d", replicas)
	}
	if !dryRun {
		t.Error("dry-run should be true")
	}
	if cpuLimit != 2.5 {
		t.Errorf("cpu-limit = %f", cpuLimit)
	}
	if timeout != 10*time.Minute {
		t.Errorf("timeout = %v", timeout)
	}
	if len(tags) != 2 || tags[0] != "v1.0" || tags[1] != "latest" {
		t.Errorf("tags = %v", tags)
	}
	if len(ports) != 2 || ports[0] != 80 || ports[1] != 443 {
		t.Errorf("ports = %v", ports)
	}
}

// TestE2EPositionalArgs tests flag parsing with positional arguments.
func TestE2EPositionalArgs(t *testing.T) {
	fs := NewFlagSet("cp", ContinueOnError)
	var recursive bool
	var verbose bool
	fs.BoolVarP(&recursive, "recursive", "r", false, "copy recursively")
	fs.BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	if err := fs.Parse([]string{"-rv", "src/", "dst/"}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	if !recursive {
		t.Error("recursive should be true")
	}
	if !verbose {
		t.Error("verbose should be true")
	}
	if fs.NArg() != 2 || fs.Arg(0) != "src/" || fs.Arg(1) != "dst/" {
		t.Errorf("args = %v", fs.Args())
	}
}

// TestE2EHelpTextGeneration tests that FlagUsages produces complete help text.
func TestE2EHelpTextGeneration(t *testing.T) {
	fs := NewFlagSet("app", ContinueOnError)
	fs.StringVarP(new(string), "output", "o", "", "output `file`")
	fs.BoolVarP(new(bool), "verbose", "v", false, "enable verbose")
	fs.IntVar(new(int), "count", 10, "number of items")

	usages := fs.FlagUsages()
	if usages == "" {
		t.Fatal("FlagUsages returned empty string")
	}

	// All flags should appear
	for _, name := range []string{"--output", "--verbose", "--count"} {
		if !contains(usages, name) {
			t.Errorf("missing %s in usage output", name)
		}
	}
	// Shorthand flags should show shorthand
	for _, sh := range []string{"-o,", "-v,"} {
		if !contains(usages, sh) {
			t.Errorf("missing shorthand %s in usage output", sh)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

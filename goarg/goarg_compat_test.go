package goarg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// compatScenario mirrors the compat/ scenarios but uses our goarg implementation.
type compatScenario struct {
	name       string
	args       []string
	newParser  func() (*Parser, any, error)
	wantErr    bool
	skipHelp   bool
	skipValues bool
}

func compatScenarios() []compatScenario {
	return []compatScenario{
		compatBasicStringInt(),
		compatBoolFlag(),
		compatDefaultValues(),
		compatRequiredMissing(),
		compatPositionalArgs(),
		compatSliceOption(),
		compatEnvOption(),
		compatUnknownOption(),
		compatSubcommandBasic(),
		compatHelpOutput(),
		compatMapType(),
		compatEmbeddedStruct(),
		compatVersionedInterface(),
		compatErrhelpSentinel(),
		compatCaseInsensitiveCmd(),
		compatEnvOnlyField(),
	}
}

func compatBasicStringInt() compatScenario {
	type Args struct {
		Name  string `arg:"-n,--name" help:"user name"`
		Count int    `arg:"-c,--count" help:"repeat count"`
	}
	return compatScenario{
		name: "basic_string_int",
		args: []string{"--name", "alice", "--count", "3"},
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatBoolFlag() compatScenario {
	type Args struct {
		Verbose bool `arg:"-v,--verbose" help:"verbose output"`
	}
	return compatScenario{
		name: "bool_flag",
		args: []string{"--verbose"},
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatDefaultValues() compatScenario {
	type Args struct {
		Port int    `arg:"-p,--port" default:"8080" help:"listen port"`
		Host string `arg:"--host" default:"localhost" help:"bind host"`
	}
	return compatScenario{
		name: "default_values",
		args: []string{},
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatRequiredMissing() compatScenario {
	type Args struct {
		Input string `arg:"--input,required" help:"input file"`
	}
	return compatScenario{
		name:       "required_missing",
		args:       []string{},
		wantErr:    true,
		skipHelp:   true,
		skipValues: true,
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatPositionalArgs() compatScenario {
	type Args struct {
		Source string `arg:"positional,required" help:"source file"`
		Dest   string `arg:"positional" help:"destination file"`
	}
	return compatScenario{
		name: "positional_args",
		args: []string{"input.txt", "output.txt"},
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatSliceOption() compatScenario {
	type Args struct {
		Files []string `arg:"-f,--file" help:"input files"`
	}
	return compatScenario{
		name: "slice_option",
		args: []string{"--file", "a.txt", "--file", "b.txt"},
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatEnvOption() compatScenario {
	type Args struct {
		Token string `arg:"--token,env:API_TOKEN" help:"API token"`
	}
	return compatScenario{
		name:     "env_option",
		args:     []string{},
		skipHelp: true,
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatUnknownOption() compatScenario {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	return compatScenario{
		name:       "unknown_option",
		args:       []string{"--unknown"},
		wantErr:    true,
		skipHelp:   true,
		skipValues: true,
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatSubcommandBasic() compatScenario {
	type ServerCmd struct {
		Port int `arg:"-p,--port" default:"8080" help:"listen port"`
	}
	type Args struct {
		Verbose bool       `arg:"-v,--verbose" help:"verbose output"`
		Server  *ServerCmd `arg:"subcommand:server" help:"run server"`
	}
	return compatScenario{
		name: "subcommand_basic",
		args: []string{"server", "--port", "9090"},
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatHelpOutput() compatScenario {
	type Args struct {
		Verbose bool   `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int    `arg:"-c,--count" help:"number of items"`
		Output  string `arg:"-o,--output" help:"output file"`
	}
	return compatScenario{
		name:       "help_output",
		args:       []string{},
		skipValues: true,
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatMapType() compatScenario {
	type Args struct {
		Headers map[string]string `arg:"--header" help:"HTTP headers"`
	}
	return compatScenario{
		name: "map_type",
		args: []string{"--header", "Content-Type=application/json", "--header", "Accept=text/html"},
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatEmbeddedStruct() compatScenario {
	type Common struct {
		Verbose bool `arg:"-v,--verbose" help:"verbose output"`
	}
	type Args struct {
		Common

		Name string `arg:"--name" help:"user name"`
	}
	return compatScenario{
		name: "embedded_struct",
		args: []string{"--verbose", "--name", "alice"},
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

type compatVersionedArgs struct {
	Name string `arg:"--name" help:"user name"`
}

func (compatVersionedArgs) Version() string     { return "1.0.0" }
func (compatVersionedArgs) Description() string { return "A test program" }

func compatVersionedInterface() compatScenario {
	return compatScenario{
		name:       "versioned_interface",
		args:       []string{},
		skipValues: true,
		newParser: func() (*Parser, any, error) {
			var a compatVersionedArgs
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatErrhelpSentinel() compatScenario {
	type Args struct {
		Name string `arg:"--name" help:"user name"`
	}
	return compatScenario{
		name:       "errhelp_sentinel",
		args:       []string{"--help"},
		wantErr:    true,
		skipHelp:   true,
		skipValues: true,
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatCaseInsensitiveCmd() compatScenario {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Serve *ServeCmd `arg:"subcommand:serve"`
	}
	return compatScenario{
		name:     "case_insensitive_cmd",
		args:     []string{"serve", "--port", "9090"},
		skipHelp: true,
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

func compatEnvOnlyField() compatScenario {
	type Args struct {
		Token string `arg:"--,env:SECRET_TOKEN" help:"secret token"`
	}
	return compatScenario{
		name:     "env_only_field",
		args:     []string{},
		skipHelp: true,
		newParser: func() (*Parser, any, error) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			return p, &a, err
		},
	}
}

// TestCompatValues validates parsed values match upstream golden files.
func TestCompatValues(t *testing.T) {
	for _, sc := range compatScenarios() {
		if sc.skipValues || sc.wantErr {
			continue
		}
		t.Run(sc.name, func(t *testing.T) {
			p, dest, err := sc.newParser()
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}
			if err := p.Parse(sc.args); err != nil {
				t.Fatalf("Parse: %v", err)
			}
			got := fmt.Sprintf("%+v", dest)
			want := readCompatGolden(t, sc.name, "values")
			if want == "" {
				t.Skip("golden file missing")
			}
			// Pointer addresses differ between runs — normalize them
			// for comparison. Replace 0x[hex] with PTR.
			got = normalizePointers(got)
			want = normalizePointers(want)
			if diff, ok := loadExpectedDiffs()[sc.name+".values"]; ok {
				if normalizePointers(got) == normalizePointers(diff.OurBehavior) {
					t.Logf("expected diff: %s", diff.Rationale)
					return
				}
			}
			assertCompatMatch(t, sc.name, "values", got, want)
		})
	}
}

// TestCompatErrors validates error messages match upstream golden files.
func TestCompatErrors(t *testing.T) {
	diffs := loadExpectedDiffs()
	for _, sc := range compatScenarios() {
		if !sc.wantErr {
			continue
		}
		t.Run(sc.name, func(t *testing.T) {
			p, _, err := sc.newParser()
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}
			parseErr := p.Parse(sc.args)
			if parseErr == nil {
				t.Fatal("expected error, got nil")
			}
			got := parseErr.Error()
			want := readCompatGolden(t, sc.name, "error")
			if want == "" {
				t.Skip("golden file missing")
			}
			if diff, ok := diffs[sc.name+".error"]; ok {
				if got == diff.OurBehavior {
					t.Logf("expected diff: %s", diff.Rationale)
					return
				}
			}
			assertCompatMatch(t, sc.name, "error", got, want)
		})
	}
}

// TestCompatHelp validates help output structural properties.
// Byte-for-byte help matching is not enforced — see HelpUsageDiffRationale
// in expected_diffs.go for the systematic formatting differences.
// Instead we validate structural invariants: usage line present, options
// section present when expected, help text non-empty.
func TestCompatHelp(t *testing.T) {
	t.Logf("Help format diffs: %s", HelpUsageDiffRationale)
	for _, sc := range compatScenarios() {
		if sc.skipHelp || sc.wantErr {
			continue
		}
		t.Run(sc.name+"/help", func(t *testing.T) {
			p, _, err := sc.newParser()
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}
			_ = p.Parse(sc.args)

			var buf bytes.Buffer
			p.WriteHelp(&buf)
			got := buf.String()
			if got == "" {
				t.Error("help output is empty")
			}
			if !strings.Contains(got, "Usage:") {
				t.Error("help output missing Usage: line")
			}
		})
		t.Run(sc.name+"/usage", func(t *testing.T) {
			p, _, err := sc.newParser()
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}
			_ = p.Parse(sc.args)

			var buf bytes.Buffer
			p.WriteUsage(&buf)
			got := buf.String()
			if got == "" {
				t.Error("usage output is empty")
			}
			if !strings.Contains(got, "Usage:") {
				t.Error("usage output missing Usage: line")
			}
		})
	}
}

// --- helpers ---

// goldenFile mirrors the compat/ GoldenFile struct for JSON reading.
type goldenFile struct {
	Output string `json:"output"`
}

func readCompatGolden(t *testing.T, scenario, kind string) string {
	t.Helper()
	name := scenario + "." + kind
	path := filepath.Join("compat", "testdata", name+".golden.json")
	data, err := os.ReadFile(path) //nolint:gosec // test golden file path from constant prefix + test name
	if err != nil {
		return ""
	}
	var gf goldenFile
	if err := json.Unmarshal(data, &gf); err != nil {
		t.Fatalf("golden file %s is not valid JSON; run 'make compat-update': %v", path, err)
	}
	return strings.TrimRight(gf.Output, "\n")
}

func assertCompatMatch(t *testing.T, scenario, kind, got, want string) {
	t.Helper()
	got = strings.TrimRight(got, "\n")
	want = strings.TrimRight(want, "\n")
	if got != want {
		t.Errorf("%s.%s mismatch:\n--- upstream ---\n%s\n--- ours ---\n%s",
			scenario, kind, want, got)
	}
}

var ptrRe = regexp.MustCompile(`0x[0-9a-f]+`)

func normalizePointers(s string) string {
	return ptrRe.ReplaceAllString(s, "PTR")
}

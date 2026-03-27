package compat

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ptrRe matches Go pointer addresses like 0x1234abcd.
var ptrRe = regexp.MustCompile(`0x[0-9a-f]+`)

func normalizePointers(s string) string {
	return ptrRe.ReplaceAllString(s, "PTR")
}

// TestCaptureUpstream runs each scenario against upstream alexflint/go-arg
// and writes golden files when -update is set.
func TestCaptureUpstream(t *testing.T) {
	for _, sc := range scenarios() {
		t.Run(sc.Name, func(t *testing.T) {
			p, dest, err := sc.NewParser()
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}

			parseErr := p.Parse(sc.Args)

			// Capture error
			if sc.WantErr {
				if parseErr == nil {
					t.Fatalf("expected error, got nil")
				}
				writeGolden(t, sc.Name, "error", parseErr.Error())
				return
			}
			if parseErr != nil {
				t.Fatalf("unexpected error: %v", parseErr)
			}

			// Capture parsed values
			if !sc.SkipValues {
				writeGolden(t, sc.Name, "values", fmt.Sprintf("%+v", dest))
			}

			// Capture help output
			if !sc.SkipHelp {
				var helpBuf bytes.Buffer
				p.WriteHelp(&helpBuf)
				writeGolden(t, sc.Name, "help", helpBuf.String())

				var usageBuf bytes.Buffer
				p.WriteUsage(&usageBuf)
				writeGolden(t, sc.Name, "usage", usageBuf.String())
			}
		})
	}
}

// TestValidateGolden verifies golden files exist for all scenarios.
func TestValidateGolden(t *testing.T) {
	for _, sc := range scenarios() {
		t.Run(sc.Name, func(t *testing.T) {
			if sc.WantErr {
				assertGoldenExists(t, sc.Name, "error")
				return
			}
			if !sc.SkipValues {
				assertGoldenExists(t, sc.Name, "values")
			}
			if !sc.SkipHelp {
				assertGoldenExists(t, sc.Name, "help")
				assertGoldenExists(t, sc.Name, "usage")
			}
		})
	}
}

func goldenPath(scenario, kind string) string {
	return filepath.Join("testdata", scenario+"."+kind+".golden")
}

func writeGolden(t *testing.T, scenario, kind, content string) {
	t.Helper()
	if !*update {
		// In non-update mode, just verify the golden file matches
		existing := readGolden(t, scenario, kind)
		if existing == "" {
			t.Skipf("golden file missing; run with -update to create")
			return
		}
		content = strings.TrimRight(content, "\n") + "\n"
		if normalizePointers(existing) != normalizePointers(content) {
			t.Errorf("golden mismatch for %s.%s:\n--- want ---\n%s--- got ---\n%s",
				scenario, kind, existing, content)
		}
		return
	}

	path := goldenPath(scenario, kind)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content = strings.TrimRight(content, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write golden: %v", err)
	}
	t.Logf("updated %s", path)
}

func readGolden(t *testing.T, scenario, kind string) string {
	t.Helper()
	data, err := os.ReadFile(goldenPath(scenario, kind))
	if err != nil {
		return ""
	}
	return string(data)
}

func assertGoldenExists(t *testing.T, scenario, kind string) {
	t.Helper()
	path := goldenPath(scenario, kind)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("missing golden file: %s (run compat tests with -update)", path)
	}
}

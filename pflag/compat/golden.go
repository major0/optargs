// Intentionally duplicated in goarg/compat and pflag/compat — separate go.mod modules prevent sharing.
package compat

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// GoldenFile is the structured golden file format with provenance metadata.
type GoldenFile struct {
	Metadata GoldenMetadata `json:"metadata"`
	Output   string         `json:"output"`
}

// GoldenMetadata records provenance of the captured output.
type GoldenMetadata struct {
	UpstreamVersion string `json:"upstream_version"`
	LocalHead       string `json:"local_head"`
	CapturedAt      string `json:"captured_at"`
}

// WriteGolden writes a JSON golden file to testdata/<name>.golden.json.
func WriteGolden(t *testing.T, name, output string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden.json")
	if err := os.MkdirAll("testdata", 0o755); err != nil {
		t.Fatal(err)
	}
	gf := GoldenFile{Metadata: captureMetadata(), Output: output}
	data, err := json.MarshalIndent(gf, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("updated %s", path)
}

// ReadGolden reads a JSON golden file and returns the output string.
// Callers that need trimmed output call strings.TrimSuffix themselves.
func ReadGolden(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", name+".golden.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden file %s not found; run with -update to generate", path)
	}
	var gf GoldenFile
	if err := json.Unmarshal(data, &gf); err != nil {
		t.Fatalf("golden file %s is not valid JSON; run 'make compat-update' to regenerate: %v", path, err)
	}
	return gf.Output
}

// captureMetadata builds GoldenMetadata from the build environment.
func captureMetadata() GoldenMetadata {
	return GoldenMetadata{
		UpstreamVersion: upstreamVersion(),
		LocalHead:       gitHead(),
		CapturedAt:      time.Now().UTC().Format(time.RFC3339),
	}
}

// upstreamVersion parses the upstream module version from go.mod.
func upstreamVersion() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2]
			}
		}
	}
	return "unknown"
}

// gitHead returns the current git HEAD short hash.
func gitHead() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

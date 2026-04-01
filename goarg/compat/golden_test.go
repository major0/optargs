package compat

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoldenRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"empty", ""},
		{"simple", "hello world"},
		{"trailing_newline", "value\n"},
		{"multi_newline", "line1\nline2\n"},
		{"unicode", "日本語テスト"},
		{"special_chars", `"quotes" and \backslash and 	tab`},
		{"json_in_output", `{"key": "value"}`},
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteGolden(t, "roundtrip_"+tt.name, tt.output)
			got := ReadGolden(t, "roundtrip_"+tt.name)
			if got != tt.output {
				t.Errorf("round-trip failed:\nwrote: %q\nread:  %q", tt.output, got)
			}
		})
	}
}

func TestGoldenStructure(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	WriteGolden(t, "structure_test", "test output")

	// Read raw JSON and verify structure
	data, err := os.ReadFile(filepath.Join("testdata", "structure_test.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	// Verify required fields are present
	for _, field := range []string{`"upstream_version"`, `"local_head"`, `"captured_at"`, `"output"`} {
		if !contains(s, field) {
			t.Errorf("missing field %s in golden JSON", field)
		}
	}
}

func TestGoldenInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Write invalid JSON to a golden file path
	os.MkdirAll("testdata", 0o755)
	os.WriteFile(filepath.Join("testdata", "bad.golden.json"), []byte("not json"), 0o644)

	// ReadGolden should fail with a message containing "-update"
	defer func() {
		// ReadGolden calls t.Fatalf which panics in subtests
	}()

	// Use a sub-test to catch the fatal
	result := testing.RunTests(func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{{
			Name: "ReadBadJSON",
			F: func(t *testing.T) {
				ReadGolden(t, "bad")
			},
		}})
	if result {
		t.Error("ReadGolden should have failed on invalid JSON")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestScenarioGoldenMapping verifies each scenario produces the expected set
// of golden files based on its SkipValues/SkipHelp/WantErr flags.
func TestScenarioGoldenMapping(t *testing.T) {
	for _, sc := range scenarios() {
		t.Run(sc.Name, func(t *testing.T) {
			if sc.WantErr {
				if !GoldenExists(FormatGoldenName(sc.Name, "error")) {
					t.Errorf("missing: %s.error", sc.Name)
				}
				// Error scenarios should NOT have values/help/usage
				if GoldenExists(FormatGoldenName(sc.Name, "values")) {
					t.Errorf("unexpected: %s.values (WantErr=true)", sc.Name)
				}
				return
			}
			if !sc.SkipValues {
				if !GoldenExists(FormatGoldenName(sc.Name, "values")) {
					t.Errorf("missing: %s.values", sc.Name)
				}
			}
			if !sc.SkipHelp {
				if !GoldenExists(FormatGoldenName(sc.Name, "help")) {
					t.Errorf("missing: %s.help", sc.Name)
				}
				if !GoldenExists(FormatGoldenName(sc.Name, "usage")) {
					t.Errorf("missing: %s.usage", sc.Name)
				}
			}
		})
	}
}

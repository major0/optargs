package compat

import (
	"os"
	"path/filepath"
	"strings"
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
	data, err := os.ReadFile(filepath.Join("testdata", "structure_test.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, field := range []string{`"upstream_version"`, `"local_head"`, `"captured_at"`, `"output"`} {
		if !strings.Contains(s, field) {
			t.Errorf("missing field %s in golden JSON", field)
		}
	}
}

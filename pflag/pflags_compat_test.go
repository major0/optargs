package pflag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// goldenFile mirrors the compat/ GoldenFile struct for JSON reading.
// Duplicated here because compat/ is a separate go.mod module.
type goldenFile struct {
	Output string `json:"output"`
}

// readJSONGolden reads a JSON golden file and returns the output string.
func readJSONGolden(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("compat", "testdata", name+".golden.json")
	data, err := os.ReadFile(path) //nolint:gosec // test golden file path is constructed from constant prefix + test name
	if err != nil {
		t.Fatalf("golden file %s not found; run 'make compat-update' to generate", path)
	}
	var gf goldenFile
	if err := json.Unmarshal(data, &gf); err != nil {
		t.Fatalf("golden file %s is not valid JSON; run 'make compat-update' to regenerate: %v", path, err)
	}
	return gf.Output
}

// readJSONGoldenValue reads a JSON golden file and trims trailing newline.
func readJSONGoldenValue(t *testing.T, name string) string {
	t.Helper()
	return strings.TrimSuffix(readJSONGolden(t, name), "\n")
}

// TestCompatStringFlag validates string flag parsing matches upstream.
func TestCompatStringFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var s string
	fs.StringVar(&s, "output", "default.txt", "output file path")
	if err := fs.Parse([]string{"--output", "result.txt"}); err != nil {
		t.Fatal(err)
	}
	if s != readJSONGoldenValue(t, "string_parse") {
		t.Errorf("string parse = %q, want %q", s, readJSONGoldenValue(t, "string_parse"))
	}
}

// TestCompatBoolFlag validates boolean flag parsing matches upstream.
func TestCompatBoolFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no_arg", []string{"--verbose"}},
		{"explicit_true", []string{"--verbose=true"}},
		{"explicit_false", []string{"--verbose=false"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var v bool
			fs.BoolVar(&v, "verbose", false, "enable verbose")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			got := "false"
			if v {
				got = "true"
			}
			want := readJSONGoldenValue(t, "bool_"+tt.name)
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

// TestCompatShorthand validates shorthand parsing matches upstream.
func TestCompatShorthand(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var s string
	fs.StringVarP(&s, "output", "o", "", "output file")
	if err := fs.Parse([]string{"-o", "file.txt"}); err != nil {
		t.Fatal(err)
	}
	if s != readJSONGoldenValue(t, "shorthand_parse") {
		t.Errorf("shorthand parse = %q, want %q", s, readJSONGoldenValue(t, "shorthand_parse"))
	}
}

// TestCompatUnknownFlag validates unknown flag error format matches upstream.
func TestCompatUnknownFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "known", "", "")
	err := fs.Parse([]string{"--unknown"})
	if err == nil {
		t.Fatal("expected error")
	}
	want := readJSONGoldenValue(t, "unknown_flag_error")
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

// TestCompatDoubleHyphen validates -- termination matches upstream.
func TestCompatDoubleHyphen(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var s string
	fs.StringVar(&s, "name", "", "")
	if err := fs.Parse([]string{"--name", "val", "--", "--other", "pos"}); err != nil {
		t.Fatal(err)
	}
	result := s + "\n"
	var resultSb119 strings.Builder
	for _, a := range fs.Args() {
		resultSb119.WriteString(a + "\n")
	}
	result += resultSb119.String()
	want := readJSONGoldenValue(t, "double_hyphen")
	if strings.TrimSuffix(result, "\n") != want {
		t.Errorf("got:\n%s\nwant:\n%s", result, want)
	}
}

// TestCompatSliceFlag validates slice flag parsing matches upstream.
func TestCompatSliceFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var ss []string
	fs.StringSliceVar(&ss, "tags", nil, "tags to apply")
	if err := fs.Parse([]string{"--tags", "a,b", "--tags", "c"}); err != nil {
		t.Fatal(err)
	}
	result := ""
	var resultSb137 strings.Builder
	for _, s := range ss {
		resultSb137.WriteString(s + "\n")
	}
	result += resultSb137.String()
	want := readJSONGoldenValue(t, "slice_parse")
	if strings.TrimSuffix(result, "\n") != want {
		t.Errorf("got:\n%s\nwant:\n%s", result, want)
	}
}

// TestCompatUsageFormat validates help text format matches upstream.
// Expected diffs are documented in compat/expected_diffs.go.
func TestCompatUsageFormat(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(fs *FlagSet)
		golden string
	}{
		{
			"string_usage",
			func(fs *FlagSet) { fs.StringVar(new(string), "output", "default.txt", "output file path") },
			"string_usage",
		},
		{
			"bool_usage",
			func(fs *FlagSet) { fs.BoolVar(new(bool), "verbose", false, "enable verbose") },
			"bool_usage",
		},
		{
			"shorthand_usage",
			func(fs *FlagSet) { fs.StringVarP(new(string), "output", "o", "", "output file") },
			"shorthand_usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			tt.setup(fs)
			got := fs.FlagUsages()
			want := readJSONGolden(t, tt.golden)
			if got != want {
				t.Errorf("usage differs:\ngot:  %q\nwant: %q", got, want)
				// Show character-level diff for debugging
				for i := 0; i < len(got) && i < len(want); i++ {
					if got[i] != want[i] {
						t.Errorf("first diff at byte %d: got %q, want %q", i, string(got[i]), string(want[i]))
						t.Errorf("got context:  %q", got[max(0, i-10):min(len(got), i+10)])
						t.Errorf("want context: %q", want[max(0, i-10):min(len(want), i+10)])
						break
					}
				}
			}
		})
	}
}

// TestCompatMixedUsage validates mixed flag usage format.
// This test compares against upstream format but allows for known differences
// documented in compat/expected_diffs.go.
func TestCompatMixedUsage(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.BoolVarP(new(bool), "verbose", "v", false, "enable verbose")
	fs.StringVarP(new(string), "output", "o", "", "output file")
	fs.IntVarP(new(int), "count", "c", 0, "count")

	got := fs.FlagUsages()
	want := readJSONGolden(t, "mixed_usage")

	// Compare line by line — order may differ since we use definition order
	gotLines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	wantLines := strings.Split(strings.TrimRight(want, "\n"), "\n")

	if len(gotLines) != len(wantLines) {
		t.Errorf("line count differs: got %d, want %d\ngot:\n%s\nwant:\n%s", len(gotLines), len(wantLines), got, want)
		return
	}

	// Check each line is present (order may differ)
	wantSet := make(map[string]bool)
	for _, l := range wantLines {
		wantSet[strings.TrimSpace(l)] = true
	}
	for _, l := range gotLines {
		trimmed := strings.TrimSpace(l)
		if !wantSet[trimmed] {
			t.Errorf("unexpected line in output: %q\ngot:\n%s\nwant:\n%s", trimmed, got, want)
		}
	}
}

// TestCompatFloatFlag validates float flag parsing matches upstream.
func TestCompatFloatFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var f float64
	fs.Float64Var(&f, "rate", 0, "rate limit")
	if err := fs.Parse([]string{"--rate", "3.14"}); err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSuffix(readJSONGolden(t, "float_parse"), "\n")
	if got := fmt.Sprintf("%g", f); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestCompatDurationFlag validates duration flag parsing matches upstream.
func TestCompatDurationFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var d time.Duration
	fs.DurationVar(&d, "timeout", 0, "timeout")
	if err := fs.Parse([]string{"--timeout", "5s"}); err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSuffix(readJSONGolden(t, "duration_parse"), "\n")
	if got := d.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestCompatStringArray validates StringArray behavior matches upstream.
func TestCompatStringArray(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var sa []string
	fs.StringArrayVar(&sa, "item", nil, "items")
	if err := fs.Parse([]string{"--item", "a,b", "--item", "c"}); err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSuffix(readJSONGolden(t, "string_array_parse"), "\n")
	if got := strings.Join(sa, "\n"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestCompatCountFlag validates Count flag behavior matches upstream.
func TestCompatCountFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var c int
	fs.CountVarP(&c, "verbose", "v", "verbosity")
	if err := fs.Parse([]string{"-v", "-v", "-v"}); err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSuffix(readJSONGolden(t, "count_parse"), "\n")
	if got := strconv.Itoa(c); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestCompatLookupSetChanged validates Lookup/Set/Changed matches upstream.
func TestCompatLookupSetChanged(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "name", "default", "")
	if err := fs.Parse([]string{"--name", "alice"}); err != nil {
		t.Fatal(err)
	}
	f := fs.Lookup("name")
	got := fmt.Sprintf("value=%s changed=%t", f.Value.String(), f.Changed)
	want := strings.TrimSuffix(readJSONGolden(t, "lookup_set_changed"), "\n")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Package compat captures upstream spf13/pflag behavior as golden outputs.
// Run `go test -update` to regenerate golden files.
// Run `go test` to validate current golden files match upstream.
package compat

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

var update = flag.Bool("update", false, "update golden files")

// captureUsage returns the usage output for a pflag.FlagSet.
func captureUsage(fs *pflag.FlagSet) string {
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.PrintDefaults()
	return buf.String()
}

// golden reads or writes a golden file depending on the -update flag.
func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")

	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden file %s not found; run with -update to generate", path)
	}
	if got != string(want) {
		t.Errorf("output differs from golden file %s:\ngot:\n%s\nwant:\n%s", path, got, string(want))
	}
}

// TestUpstreamStringFlag captures upstream string flag parsing and help text.
func TestUpstreamStringFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var s string
	fs.StringVar(&s, "output", "default.txt", "output file path")

	// Parse
	if err := fs.Parse([]string{"--output", "result.txt"}); err != nil {
		t.Fatal(err)
	}
	golden(t, "string_parse", s)

	// Help text
	fs2 := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs2.StringVar(new(string), "output", "default.txt", "output file path")
	golden(t, "string_usage", captureUsage(fs2))
}

// TestUpstreamBoolFlag captures upstream boolean flag behavior.
func TestUpstreamBoolFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no_arg", []string{"--verbose"}, ""},
		{"explicit_true", []string{"--verbose=true"}, ""},
		{"explicit_false", []string{"--verbose=false"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			var v bool
			fs.BoolVar(&v, "verbose", false, "enable verbose")
			if err := fs.Parse(tt.args); err != nil {
				golden(t, "bool_"+tt.name, "ERROR: "+err.Error())
				return
			}
			golden(t, "bool_"+tt.name, boolStr(v))
		})
	}

	// Help text
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.BoolVar(new(bool), "verbose", false, "enable verbose")
	golden(t, "bool_usage", captureUsage(fs))
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// TestUpstreamIntFlag captures upstream int flag behavior.
func TestUpstreamIntFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var n int
	fs.IntVar(&n, "count", 0, "number of items")
	if err := fs.Parse([]string{"--count", "42"}); err != nil {
		t.Fatal(err)
	}
	golden(t, "int_parse", intStr(n))

	// Error case
	fs2 := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var buf bytes.Buffer
	fs2.SetOutput(&buf)
	fs2.IntVar(new(int), "count", 0, "number of items")
	err := fs2.Parse([]string{"--count", "abc"})
	if err != nil {
		golden(t, "int_error", err.Error())
	}
}

func intStr(n int) string {
	return bytes.NewBufferString("").String() + string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// TestUpstreamShorthand captures upstream shorthand behavior.
func TestUpstreamShorthand(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var s string
	fs.StringVarP(&s, "output", "o", "", "output file")
	if err := fs.Parse([]string{"-o", "file.txt"}); err != nil {
		t.Fatal(err)
	}
	golden(t, "shorthand_parse", s)
	golden(t, "shorthand_usage", captureUsage(fs))
}

// TestUpstreamUnknownFlag captures upstream unknown flag error format.
func TestUpstreamUnknownFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.StringVar(new(string), "known", "", "")
	err := fs.Parse([]string{"--unknown"})
	if err != nil {
		golden(t, "unknown_flag_error", err.Error())
	}
}

// TestUpstreamDoubleHyphen captures upstream -- termination behavior.
func TestUpstreamDoubleHyphen(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var s string
	fs.StringVar(&s, "name", "", "")
	if err := fs.Parse([]string{"--name", "val", "--", "--other", "pos"}); err != nil {
		t.Fatal(err)
	}
	result := s + "\n"
	for _, a := range fs.Args() {
		result += a + "\n"
	}
	golden(t, "double_hyphen", result)
}

// TestUpstreamSliceFlag captures upstream slice flag behavior.
func TestUpstreamSliceFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var ss []string
	fs.StringSliceVar(&ss, "tags", nil, "tags to apply")
	if err := fs.Parse([]string{"--tags", "a,b", "--tags", "c"}); err != nil {
		t.Fatal(err)
	}
	result := ""
	for _, s := range ss {
		result += s + "\n"
	}
	golden(t, "slice_parse", result)
}

// TestUpstreamMixedFlags captures a complex real-world scenario.
func TestUpstreamMixedFlags(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var verbose bool
	var output string
	var count int
	fs.BoolVarP(&verbose, "verbose", "v", false, "enable verbose")
	fs.StringVarP(&output, "output", "o", "", "output file")
	fs.IntVarP(&count, "count", "c", 0, "count")

	if err := fs.Parse([]string{"-v", "--output=result.txt", "-c", "5", "pos1", "pos2"}); err != nil {
		t.Fatal(err)
	}
	result := boolStr(verbose) + "\n" + output + "\n" + intStr(count) + "\n"
	for _, a := range fs.Args() {
		result += a + "\n"
	}
	golden(t, "mixed_flags", result)
	golden(t, "mixed_usage", captureUsage(fs))
}

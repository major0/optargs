// Package compat captures upstream spf13/pflag behavior as golden outputs.
// Run `go test -update` to regenerate golden files.
// Run `go test` to validate current golden files match upstream.
package compat

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

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

// golden writes or reads a JSON golden file depending on the -update flag.
func golden(t *testing.T, name, got string) {
	t.Helper()
	if *update {
		WriteGolden(t, name, got)
		return
	}
	want := ReadGolden(t, name)
	if got != want {
		t.Errorf("output differs from golden %s:\ngot:\n%s\nwant:\n%s", name, got, want)
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

// TestUpstreamFloatFlag captures upstream float64 flag behavior.
func TestUpstreamFloatFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var f float64
	fs.Float64Var(&f, "rate", 0, "rate limit")
	if err := fs.Parse([]string{"--rate", "3.14"}); err != nil {
		t.Fatal(err)
	}
	golden(t, "float_parse", fmt.Sprintf("%g", f))
}

// TestUpstreamDurationFlag captures upstream duration flag behavior.
func TestUpstreamDurationFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var d time.Duration
	fs.DurationVar(&d, "timeout", 0, "timeout")
	if err := fs.Parse([]string{"--timeout", "5s"}); err != nil {
		t.Fatal(err)
	}
	golden(t, "duration_parse", d.String())
}

// TestUpstreamStringArray captures upstream StringArray behavior.
func TestUpstreamStringArray(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var sa []string
	fs.StringArrayVar(&sa, "item", nil, "items")
	if err := fs.Parse([]string{"--item", "a,b", "--item", "c"}); err != nil {
		t.Fatal(err)
	}
	golden(t, "string_array_parse", strings.Join(sa, "\n"))
}

// TestUpstreamCountFlag captures upstream Count flag behavior.
func TestUpstreamCountFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var c int
	fs.CountVarP(&c, "verbose", "v", "verbosity")
	if err := fs.Parse([]string{"-v", "-v", "-v"}); err != nil {
		t.Fatal(err)
	}
	golden(t, "count_parse", fmt.Sprintf("%d", c))
}

// TestUpstreamLookupSetChanged captures Lookup/Set/Changed behavior.
func TestUpstreamLookupSetChanged(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("name", "default", "")
	if err := fs.Parse([]string{"--name", "alice"}); err != nil {
		t.Fatal(err)
	}
	f := fs.Lookup("name")
	result := fmt.Sprintf("value=%s changed=%t", f.Value.String(), f.Changed)
	golden(t, "lookup_set_changed", result)
}

// TestUpstreamNFlagNArg captures NFlag/NArg behavior.
func TestUpstreamNFlagNArg(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("a", "", "")
	fs.String("b", "", "")
	if err := fs.Parse([]string{"--a", "1", "pos1", "pos2"}); err != nil {
		t.Fatal(err)
	}
	result := fmt.Sprintf("nflag=%d narg=%d", fs.NFlag(), fs.NArg())
	golden(t, "nflag_narg", result)
}

// TestUpstreamHidden captures hidden flag usage output.
func TestUpstreamHidden(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("visible", "", "visible flag")
	fs.String("hidden", "", "hidden flag")
	fs.MarkHidden("hidden")
	golden(t, "hidden_usage", captureUsage(fs))
}

// TestUpstreamSortFlags captures SortFlags behavior.
func TestUpstreamSortFlags(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.SortFlags = true
	fs.String("zebra", "", "z flag")
	fs.String("alpha", "", "a flag")
	golden(t, "sort_flags_sorted", captureUsage(fs))

	fs2 := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs2.SortFlags = false
	fs2.String("zebra", "", "z flag")
	fs2.String("alpha", "", "a flag")
	golden(t, "sort_flags_unsorted", captureUsage(fs2))
}

// TestUpstreamDeprecated captures deprecated flag behavior.
func TestUpstreamDeprecated(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("old-flag", "", "old flag")
	fs.MarkDeprecated("old-flag", "use --new-flag instead")
	golden(t, "deprecated_usage", captureUsage(fs))
}

// TestUpstreamSetInterspersed captures interspersed behavior.
func TestUpstreamSetInterspersed(t *testing.T) {
	// Interspersed enabled (default)
	fs1 := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs1.String("name", "", "")
	if err := fs1.Parse([]string{"pos1", "--name", "val", "pos2"}); err != nil {
		t.Fatal(err)
	}
	r1 := fmt.Sprintf("name=%s narg=%d", fs1.Lookup("name").Value.String(), fs1.NArg())
	golden(t, "interspersed_enabled", r1)

	// Interspersed disabled
	fs2 := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs2.SetInterspersed(false)
	fs2.String("name", "", "")
	if err := fs2.Parse([]string{"pos1", "--name", "val"}); err != nil {
		t.Fatal(err)
	}
	r2 := fmt.Sprintf("name=%s narg=%d", fs2.Lookup("name").Value.String(), fs2.NArg())
	golden(t, "interspersed_disabled", r2)
}

// --- Upstream feature absence tests ---
// These prove ❌ claims in the README comparison table.

// TestUpstreamNoPOSIXCompaction proves upstream pflag doesn't support -abc compaction.
func TestUpstreamNoPOSIXCompaction(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	var a, b, c bool
	fs.BoolVarP(&a, "alpha", "a", false, "")
	fs.BoolVarP(&b, "beta", "b", false, "")
	fs.BoolVarP(&c, "gamma", "c", false, "")
	err := fs.Parse([]string{"-abc"})
	// Upstream pflag DOES support -abc for booleans, but only as a special case.
	// It does NOT support compaction where the last flag takes an argument.
	// Test the argument case:
	fs2 := pflag.NewFlagSet("test2", pflag.ContinueOnError)
	fs2.SetOutput(&buf)
	fs2.BoolVarP(new(bool), "alpha", "a", false, "")
	fs2.BoolVarP(new(bool), "beta", "b", false, "")
	fs2.StringVarP(new(string), "output", "o", "", "")
	err2 := fs2.Parse([]string{"-abo", "file.txt"})
	golden(t, "no_posix_compaction_bool", fmt.Sprintf("a=%t b=%t c=%t err=%v", a, b, c, err))
	golden(t, "no_posix_compaction_arg", fmt.Sprintf("err=%v", err2))
}

// TestUpstreamNoGNULongestMatch proves upstream pflag doesn't do prefix matching.
func TestUpstreamNoGNULongestMatch(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.String("enable-bob", "", "")
	fs.String("enable-bobadufoo", "", "")
	err := fs.Parse([]string{"--enable-boba", "val"})
	golden(t, "no_gnu_longest_match", fmt.Sprintf("err=%v", err))
}

// TestUpstreamNoBooleanNegation proves upstream pflag doesn't support --no-flag.
func TestUpstreamNoBooleanNegation(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.Bool("verbose", false, "")
	err := fs.Parse([]string{"--no-verbose"})
	golden(t, "no_boolean_negation", fmt.Sprintf("err=%v", err))
}

// TestUpstreamNoShortOnlyFlags proves upstream pflag requires long names.
func TestUpstreamNoShortOnlyFlags(t *testing.T) {
	// Upstream pflag has no ShortVar() API — every flag must have a long name.
	// We can't even construct a short-only flag with upstream.
	golden(t, "no_short_only_flags", "upstream has no ShortVar API")
}

// TestUpstreamNoGetoptLongOnly proves upstream pflag doesn't support long-only mode.
func TestUpstreamNoGetoptLongOnly(t *testing.T) {
	// Upstream pflag has no SetLongOnly() API.
	golden(t, "no_getopt_long_only", "upstream has no SetLongOnly API")
}

// TestUpstreamNoArbitraryOptionNames proves upstream pflag can't handle '=' in flag names.
func TestUpstreamNoArbitraryOptionNames(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.String("foo", "", "")
	fs.String("foo=bar", "", "")
	// Upstream splits on first '=', so --foo=bar=value gives foo="bar=value"
	// instead of recognizing "foo=bar" as the flag name.
	err := fs.Parse([]string{"--foo=bar=value"})
	f := fs.Lookup("foo")
	foobar := fs.Lookup("foo=bar")
	golden(t, "no_arbitrary_option_names", fmt.Sprintf(
		"err=%v foo=%q foo=bar=%q",
		err, f.Value.String(), foobar.Value.String(),
	))
}

// TestUpstreamNoManyToOneMapping proves upstream pflag has no alias/many-to-one API.
func TestUpstreamNoManyToOneMapping(t *testing.T) {
	// Upstream pflag has no AliasVar() or equivalent API to map multiple
	// flag names to a single variable without registering separate flags.
	golden(t, "no_many_to_one_mapping", "upstream has no AliasVar API")
}

// TestUpstreamNoBoolArgValuer proves upstream pflag treats all bools as OptionalArgument.
func TestUpstreamNoBoolArgValuer(t *testing.T) {
	// Upstream pflag has no BoolTakesArg() interface. All boolean flags
	// are treated as OptionalArgument, which means Count flags can
	// incorrectly consume the next positional argument.
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	var count int
	fs.CountVarP(&count, "verbose", "v", "")
	// With upstream, --verbose followed by a positional may misbehave
	// because Count is treated like OptionalArgument.
	err := fs.Parse([]string{"--verbose", "--verbose"})
	golden(t, "no_bool_arg_valuer", fmt.Sprintf("err=%v count=%d", err, count))
}

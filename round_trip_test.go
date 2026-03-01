package optargs

import (
	"slices"
	"testing"
)

// roundTrip parses args, regenerates them from the parsed options, parses
// again, and verifies the two parse results are equivalent.  equal selects
// the comparison function for options (exact or unordered).
func roundTrip(
	t *testing.T,
	parse func([]string) (*Parser, error),
	args []string,
	equal func([]Option, []Option) bool,
) {
	t.Helper()

	p1, err := parse(args)
	if err != nil {
		t.Fatalf("first parse: %v", err)
	}
	opts1 := collectOpts(p1)

	gen := generateArgsFromOptions(opts1, p1.Args)

	p2, err := parse(gen)
	if err != nil {
		t.Fatalf("second parse: %v", err)
	}
	opts2 := collectOpts(p2)

	if !equal(opts1, opts2) {
		t.Errorf("options differ\n  original:  %+v\n  round-trip: %+v", opts1, opts2)
	}
	if !slices.Equal(p1.Args, p2.Args) {
		t.Errorf("remaining args differ\n  original:  %+v\n  round-trip: %+v", p1.Args, p2.Args)
	}
}

// TestRoundTripShortOptions tests round-trip parsing for short options.
func TestRoundTripShortOptions(t *testing.T) {
	tests := []struct {
		name      string
		optstring string
		args      []string
	}{
		{"simple flags", "abc", []string{"-a", "-b", "-c"}},
		{"compacted flags", "abc", []string{"-abc"}},
		{"required arguments", "a:b:c", []string{"-a", "arg1", "-b", "arg2", "-c", "arg3"}},
		{"optional arguments", "a::b::c::", []string{"-aarg1", "-b", "-carg3"}},
		{"mixed argument types", "a:b::c", []string{"-a", "required", "-boptional", "-c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parse := func(a []string) (*Parser, error) { return GetOpt(a, tt.optstring) }
			roundTrip(t, parse, tt.args, optionsEqual)
		})
	}
}

// TestRoundTripLongOptions tests round-trip parsing for long options.
func TestRoundTripLongOptions(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
		{Name: "config", HasArg: OptionalArgument},
	}

	tests := []struct {
		name string
		args []string
	}{
		{"simple long options", []string{"--verbose", "--output", "file.txt"}},
		{"equals syntax", []string{"--verbose", "--output=file.txt", "--config=debug"}},
		{"mixed syntax", []string{"--verbose", "--output=file.txt", "--config"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parse := func(a []string) (*Parser, error) { return GetOptLong(a, "", longOpts) }
			roundTrip(t, parse, tt.args, optionsEqual)
		})
	}
}

// TestRoundTripMixedOptions tests round-trip parsing for mixed short and long options.
func TestRoundTripMixedOptions(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}

	tests := []struct {
		name      string
		optstring string
		args      []string
	}{
		{"mixed short and long", "vo:", []string{"-v", "--output", "file.txt", "-o", "other.txt"}},
		{"with non-option arguments", "v", []string{"-v", "file1", "--verbose", "file2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parse := func(a []string) (*Parser, error) { return GetOptLong(a, tt.optstring, longOpts) }
			roundTrip(t, parse, tt.args, optionsEquivalent)
		})
	}
}

// TestRoundTripOptionCompaction verifies that expanded and compacted forms
// parse to the same result.
func TestRoundTripOptionCompaction(t *testing.T) {
	tests := []struct {
		name      string
		optstring string
		original  []string
		compacted []string
	}{
		{"basic compaction", "abc", []string{"-a", "-b", "-c"}, []string{"-abc"}},
		{"with required argument", "ab:c", []string{"-a", "-barg"}, []string{"-abarg"}},
		{"with optional argument", "ab::c", []string{"-a", "-barg"}, []string{"-abarg"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1, err := GetOpt(tt.original, tt.optstring)
			if err != nil {
				t.Fatalf("original parse: %v", err)
			}
			opts1 := collectOpts(p1)

			p2, err := GetOpt(tt.compacted, tt.optstring)
			if err != nil {
				t.Fatalf("compacted parse: %v", err)
			}
			opts2 := collectOpts(p2)

			if !optionsEqual(opts1, opts2) {
				t.Errorf("options differ\n  original:  %+v\n  compacted: %+v", opts1, opts2)
			}
			if !slices.Equal(p1.Args, p2.Args) {
				t.Errorf("remaining args differ\n  original:  %+v\n  compacted: %+v", p1.Args, p2.Args)
			}
		})
	}
}

// generateArgsFromOptions reconstructs command-line arguments from parsed options.
func generateArgsFromOptions(options []Option, remainingArgs []string) []string {
	var args []string
	for _, opt := range options {
		if len(opt.Name) == 1 {
			args = append(args, "-"+opt.Name)
		} else {
			args = append(args, "--"+opt.Name)
		}
		if opt.HasArg {
			args = append(args, opt.Arg)
		}
	}
	return append(args, remainingArgs...)
}

// optionsEqual checks if two option slices are exactly equal (order-sensitive).
func optionsEqual(a, b []Option) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// optionsEquivalent checks if two option slices contain the same options
// regardless of order.
func optionsEquivalent(a, b []Option) bool {
	if len(a) != len(b) {
		return false
	}
	type key struct {
		Name   string
		HasArg bool
		Arg    string
	}
	counts := make(map[key]int, len(a))
	for _, o := range a {
		counts[key{o.Name, o.HasArg, o.Arg}]++
	}
	for _, o := range b {
		k := key{o.Name, o.HasArg, o.Arg}
		counts[k]--
		if counts[k] < 0 {
			return false
		}
	}
	return true
}

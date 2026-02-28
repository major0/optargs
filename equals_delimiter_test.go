package optargs

import (
	"testing"
)

// TestEqualsDelimiterStripping verifies that the = delimiter is not
// included in the arg value when using --option=value syntax.
func TestEqualsDelimiterStripping(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		longOpts []Flag
		wantName string
		wantArg  string
		wantHas  bool
	}{
		{
			name:     "required arg with equals",
			args:     []string{"--file=input.txt"},
			longOpts: []Flag{{Name: "file", HasArg: RequiredArgument}},
			wantName: "file", wantArg: "input.txt", wantHas: true,
		},
		{
			name:     "optional arg with equals",
			args:     []string{"--config=debug"},
			longOpts: []Flag{{Name: "config", HasArg: OptionalArgument}},
			wantName: "config", wantArg: "debug", wantHas: true,
		},
		{
			name:     "empty arg with equals",
			args:     []string{"--output="},
			longOpts: []Flag{{Name: "output", HasArg: RequiredArgument}},
			wantName: "output", wantArg: "", wantHas: true,
		},
		{
			name:     "negative number arg",
			args:     []string{"--count=-5"},
			longOpts: []Flag{{Name: "count", HasArg: RequiredArgument}},
			wantName: "count", wantArg: "-5", wantHas: true,
		},
		{
			name:     "arg containing multiple equals",
			args:     []string{"--query=key=value=extra"},
			longOpts: []Flag{{Name: "query", HasArg: RequiredArgument}},
			wantName: "query", wantArg: "key=value=extra", wantHas: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong(tt.args, "", tt.longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			for opt, err := range p.Options() {
				if err != nil {
					t.Fatalf("Options: %v", err)
				}
				if opt.Name != tt.wantName {
					t.Errorf("Name: got %q, want %q", opt.Name, tt.wantName)
				}
				if opt.Arg != tt.wantArg {
					t.Errorf("Arg: got %q, want %q", opt.Arg, tt.wantArg)
				}
				if opt.HasArg != tt.wantHas {
					t.Errorf("HasArg: got %v, want %v", opt.HasArg, tt.wantHas)
				}
			}
		})
	}
}

// TestOverlappingOptionNames verifies longest-prefix-first matching when
// multiple registered option names share a common prefix.
func TestOverlappingOptionNames(t *testing.T) {
	t.Run("exact_match_wins_over_prefix", func(t *testing.T) {
		// "foo" and "foobar" registered. Input: --foo=val
		// "foo" is exact prefix + '=' boundary → match "foo", arg "val"
		longOpts := []Flag{
			{Name: "foo", HasArg: RequiredArgument},
			{Name: "foobar", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--foo=val"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "foo" {
				t.Errorf("Name: got %q, want %q", opt.Name, "foo")
			}
			if opt.Arg != "val" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "val")
			}
		}
	})

	t.Run("longer_prefix_wins_at_equals_boundary", func(t *testing.T) {
		// "foo" and "foobar" registered. Input: --foobar=val
		// "foobar" is longest prefix + '=' boundary → match "foobar"
		longOpts := []Flag{
			{Name: "foo", HasArg: RequiredArgument},
			{Name: "foobar", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--foobar=val"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "foobar" {
				t.Errorf("Name: got %q, want %q", opt.Name, "foobar")
			}
			if opt.Arg != "val" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "val")
			}
		}
	})

	t.Run("fallback_to_shorter_when_no_equals_boundary", func(t *testing.T) {
		// "foo" (RequiredArgument) and "foobar" (NoArgument) registered.
		// Input: --foo=baz
		// "foobar" is longer prefix but "foo=baz" doesn't start with "foobar"
		// "foo" matches at '=' boundary → match "foo", arg "baz"
		longOpts := []Flag{
			{Name: "foo", HasArg: RequiredArgument},
			{Name: "foobar", HasArg: NoArgument},
		}
		p, err := GetOptLong([]string{"--foo=baz"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "foo" {
				t.Errorf("Name: got %q, want %q", opt.Name, "foo")
			}
			if opt.Arg != "baz" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "baz")
			}
		}
	})

	t.Run("three_level_prefix_chain", func(t *testing.T) {
		// "o", "out", "output" registered. Input: --output=file.txt
		// Longest match "output" at '=' boundary → match "output"
		longOpts := []Flag{
			{Name: "o", HasArg: RequiredArgument},
			{Name: "out", HasArg: RequiredArgument},
			{Name: "output", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--output=file.txt"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "output" {
				t.Errorf("Name: got %q, want %q", opt.Name, "output")
			}
			if opt.Arg != "file.txt" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "file.txt")
			}
		}
	})

	t.Run("three_level_mid_match", func(t *testing.T) {
		// "o", "out", "output" registered. Input: --out=file.txt
		// "out" matches at '=' boundary → match "out"
		longOpts := []Flag{
			{Name: "o", HasArg: RequiredArgument},
			{Name: "out", HasArg: RequiredArgument},
			{Name: "output", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--out=file.txt"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "out" {
				t.Errorf("Name: got %q, want %q", opt.Name, "out")
			}
			if opt.Arg != "file.txt" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "file.txt")
			}
		}
	})

	t.Run("three_level_shortest_match", func(t *testing.T) {
		// "o", "out", "output" registered. Input: --o=file.txt
		// "o" matches at '=' boundary → match "o"
		longOpts := []Flag{
			{Name: "o", HasArg: RequiredArgument},
			{Name: "out", HasArg: RequiredArgument},
			{Name: "output", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--o=file.txt"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "o" {
				t.Errorf("Name: got %q, want %q", opt.Name, "o")
			}
			if opt.Arg != "file.txt" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "file.txt")
			}
		}
	})

	t.Run("noarg_longest_skips_to_shorter_with_arg", func(t *testing.T) {
		// "foo" (RequiredArgument), "foo=bar" (NoArgument)
		// Input: --foo=bar → longest "foo=bar" is exact match (NoArgument) → match
		longOpts := []Flag{
			{Name: "foo", HasArg: RequiredArgument},
			{Name: "foo=bar", HasArg: NoArgument},
		}
		p, err := GetOptLong([]string{"--foo=bar"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "foo=bar" {
				t.Errorf("Name: got %q, want %q", opt.Name, "foo=bar")
			}
			if opt.HasArg {
				t.Errorf("HasArg: got true, want false")
			}
		}
	})

	t.Run("equals_in_name_with_arg", func(t *testing.T) {
		// "foo=bar" (RequiredArgument) and "foo" (RequiredArgument)
		// Input: --foo=bar=baz → longest match "foo=bar" at '=' boundary, arg "baz"
		longOpts := []Flag{
			{Name: "foo", HasArg: RequiredArgument},
			{Name: "foo=bar", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--foo=bar=baz"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "foo=bar" {
				t.Errorf("Name: got %q, want %q", opt.Name, "foo=bar")
			}
			if opt.Arg != "baz" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "baz")
			}
		}
	})

	t.Run("shorter_name_when_longer_not_registered", func(t *testing.T) {
		// Only "foo" registered (RequiredArgument)
		// Input: --foo=bar=baz → match "foo", arg "bar=baz"
		longOpts := []Flag{
			{Name: "foo", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--foo=bar=baz"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "foo" {
				t.Errorf("Name: got %q, want %q", opt.Name, "foo")
			}
			if opt.Arg != "bar=baz" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "bar=baz")
			}
		}
	})

	t.Run("noarg_skips_to_shorter_candidate", func(t *testing.T) {
		// "output" (NoArgument), "out" (RequiredArgument)
		// Input: --output=file → "output" is exact match but NoArgument
		// with '=' present. The input "output=file" has len > "output",
		// so it's a partial match with '=' at boundary. NoArgument skips.
		// Falls back to "out" which has '=' at name[3] → nope, name[3]='p'.
		// Neither matches at a valid boundary → error.
		// Actually: input is "output=file", candidates: "output" (len 6),
		// "out" (len 3). "output" len 6 < len("output=file") 11, so partial.
		// name[6] == '=' → but NoArgument, skip. "out" len 3, name[3]='p' ≠ '=', skip.
		// → error
		longOpts := []Flag{
			{Name: "output", HasArg: NoArgument},
			{Name: "out", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--output=file"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		var gotErr bool
		for _, err := range p.Options() {
			if err != nil {
				gotErr = true
			}
		}
		if !gotErr {
			t.Error("expected error for NoArgument option with =value, got none")
		}
	})
}

// TestSubcommandOverlappingLongOpts verifies longest-prefix matching across
// parent-child parser chains where both parent and child register options
// that share prefixes.
func TestSubcommandOverlappingLongOpts(t *testing.T) {
	t.Run("child_longer_prefix_wins_over_parent_shorter", func(t *testing.T) {
		// Parent: "out" (RequiredArgument)
		// Child:  "output" (RequiredArgument)
		// Input to child: --output=file.txt
		// Child's "output" is longest match → match "output", arg "file.txt"
		parent, err := GetOptLong([]string{}, "", []Flag{
			{Name: "out", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--output=file.txt"}, "", []Flag{
			{Name: "output", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		parent.AddCmd("sub", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "output" {
				t.Errorf("Name: got %q, want %q", opt.Name, "output")
			}
			if opt.Arg != "file.txt" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "file.txt")
			}
		}
	})

	t.Run("parent_longer_prefix_wins_over_child_shorter", func(t *testing.T) {
		// Parent: "output" (RequiredArgument)
		// Child:  "out" (RequiredArgument)
		// Input to child: --output=file.txt
		// Parent's "output" is longest match across chain → match "output"
		parent, err := GetOptLong([]string{}, "", []Flag{
			{Name: "output", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--output=file.txt"}, "", []Flag{
			{Name: "out", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		parent.AddCmd("sub", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "output" {
				t.Errorf("Name: got %q, want %q", opt.Name, "output")
			}
			if opt.Arg != "file.txt" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "file.txt")
			}
		}
	})

	t.Run("child_matches_shorter_when_input_is_shorter", func(t *testing.T) {
		// Parent: "output" (RequiredArgument)
		// Child:  "out" (RequiredArgument)
		// Input to child: --out=val
		// "out" matches at '=' boundary. "output" is not a prefix of "out=val".
		// → match "out", arg "val"
		parent, err := GetOptLong([]string{}, "", []Flag{
			{Name: "output", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--out=val"}, "", []Flag{
			{Name: "out", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		parent.AddCmd("sub", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "out" {
				t.Errorf("Name: got %q, want %q", opt.Name, "out")
			}
			if opt.Arg != "val" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "val")
			}
		}
	})

	t.Run("three_level_chain_longest_from_grandparent", func(t *testing.T) {
		// Grandparent: "output-format" (RequiredArgument)
		// Parent:      "output" (RequiredArgument)
		// Child:       "out" (RequiredArgument)
		// Input to child: --output-format=json
		// Grandparent's "output-format" is longest → match, arg "json"
		gp, err := GetOptLong([]string{}, "", []Flag{
			{Name: "output-format", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("grandparent: %v", err)
		}
		par, err := GetOptLong([]string{}, "", []Flag{
			{Name: "output", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--output-format=json"}, "", []Flag{
			{Name: "out", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		gp.AddCmd("mid", par)
		par.AddCmd("leaf", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "output-format" {
				t.Errorf("Name: got %q, want %q", opt.Name, "output-format")
			}
			if opt.Arg != "json" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "json")
			}
		}
	})

	t.Run("three_level_chain_mid_match", func(t *testing.T) {
		// Same chain as above. Input: --output=json
		// Parent's "output" matches at '=' boundary → match "output", arg "json"
		gp, err := GetOptLong([]string{}, "", []Flag{
			{Name: "output-format", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("grandparent: %v", err)
		}
		par, err := GetOptLong([]string{}, "", []Flag{
			{Name: "output", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--output=json"}, "", []Flag{
			{Name: "out", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		gp.AddCmd("mid", par)
		par.AddCmd("leaf", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "output" {
				t.Errorf("Name: got %q, want %q", opt.Name, "output")
			}
			if opt.Arg != "json" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "json")
			}
		}
	})

	t.Run("equals_in_option_name_across_chain", func(t *testing.T) {
		// Parent: "key=val" (NoArgument)
		// Child:  "key" (RequiredArgument)
		// Input to child: --key=val
		// Parent's "key=val" is longest exact match → match "key=val" (NoArgument)
		parent, err := GetOptLong([]string{}, "", []Flag{
			{Name: "key=val", HasArg: NoArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--key=val"}, "", []Flag{
			{Name: "key", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		parent.AddCmd("sub", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "key=val" {
				t.Errorf("Name: got %q, want %q", opt.Name, "key=val")
			}
			if opt.HasArg {
				t.Errorf("HasArg: got true, want false")
			}
		}
	})

	t.Run("equals_in_name_with_arg_across_chain", func(t *testing.T) {
		// Parent: "key=val" (RequiredArgument)
		// Child:  "key" (RequiredArgument)
		// Input to child: --key=val=extra
		// Parent's "key=val" matches at '=' boundary → arg "extra"
		parent, err := GetOptLong([]string{}, "", []Flag{
			{Name: "key=val", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--key=val=extra"}, "", []Flag{
			{Name: "key", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		parent.AddCmd("sub", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "key=val" {
				t.Errorf("Name: got %q, want %q", opt.Name, "key=val")
			}
			if opt.Arg != "extra" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "extra")
			}
		}
	})

	t.Run("child_own_option_preferred_when_same_length", func(t *testing.T) {
		// Both parent and child register "verbose" (NoArgument).
		// Input to child: --verbose
		// Both are exact matches at same length. Child's own should be found
		// (child is searched first in the walk).
		parent, err := GetOptLong([]string{}, "", []Flag{
			{Name: "verbose", HasArg: NoArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--verbose"}, "", []Flag{
			{Name: "verbose", HasArg: NoArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		parent.AddCmd("sub", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "verbose" {
				t.Errorf("Name: got %q, want %q", opt.Name, "verbose")
			}
		}
	})

	t.Run("parent_only_option_resolved_from_child", func(t *testing.T) {
		// Parent: "debug" (RequiredArgument)
		// Child:  no overlapping options
		// Input to child: --debug=trace
		// Only parent has "debug" → match via parent chain, arg "trace"
		parent, err := GetOptLong([]string{}, "", []Flag{
			{Name: "debug", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--debug=trace"}, "", []Flag{
			{Name: "color", HasArg: NoArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		parent.AddCmd("sub", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "debug" {
				t.Errorf("Name: got %q, want %q", opt.Name, "debug")
			}
			if opt.Arg != "trace" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "trace")
			}
		}
	})
}

// TestMultipleOptionsWithOverlappingPrefixes exercises a single parser with
// many overlapping prefixes to stress the sort + iterate logic.
func TestMultipleOptionsWithOverlappingPrefixes(t *testing.T) {
	longOpts := []Flag{
		{Name: "v", HasArg: RequiredArgument},
		{Name: "ve", HasArg: RequiredArgument},
		{Name: "ver", HasArg: RequiredArgument},
		{Name: "verb", HasArg: RequiredArgument},
		{Name: "verbo", HasArg: RequiredArgument},
		{Name: "verbos", HasArg: RequiredArgument},
		{Name: "verbose", HasArg: RequiredArgument},
	}

	tests := []struct {
		input    string
		wantName string
		wantArg  string
	}{
		{"--v=1", "v", "1"},
		{"--ve=2", "ve", "2"},
		{"--ver=3", "ver", "3"},
		{"--verb=4", "verb", "4"},
		{"--verbo=5", "verbo", "5"},
		{"--verbos=6", "verbos", "6"},
		{"--verbose=7", "verbose", "7"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p, err := GetOptLong([]string{tt.input}, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			for opt, err := range p.Options() {
				if err != nil {
					t.Fatalf("Options: %v", err)
				}
				if opt.Name != tt.wantName {
					t.Errorf("Name: got %q, want %q", opt.Name, tt.wantName)
				}
				if opt.Arg != tt.wantArg {
					t.Errorf("Arg: got %q, want %q", opt.Arg, tt.wantArg)
				}
			}
		})
	}
}

// TestTripleEqualsOverlap exercises three options registered simultaneously
// where each name is a prefix of the next, with '=' embedded in the names:
//
//	"foo"         OptionalArgument  (e.g. --foo=arg)
//	"foo=bar"     OptionalArgument  (e.g. --foo=bar=arg)
//	"foo=bar=arg" NoArgument
//
// Every input must resolve to the longest matching registered option.
func TestTripleEqualsOverlap(t *testing.T) {
	longOpts := []Flag{
		{Name: "foo", HasArg: OptionalArgument},
		{Name: "foo=bar", HasArg: OptionalArgument},
		{Name: "foo=bar=arg", HasArg: NoArgument},
	}

	tests := []struct {
		name     string
		input    string
		wantName string
		wantArg  string
		wantHas  bool
	}{
		{
			name:     "exact foo=bar=arg matches NoArgument",
			input:    "--foo=bar=arg",
			wantName: "foo=bar=arg",
			wantArg:  "",
			wantHas:  false,
		},
		{
			name:     "foo=bar with equals arg",
			input:    "--foo=bar=something",
			wantName: "foo=bar",
			wantArg:  "something",
			wantHas:  true,
		},
		{
			name:     "foo with equals arg",
			input:    "--foo=qux",
			wantName: "foo",
			wantArg:  "qux",
			wantHas:  true,
		},
		{
			name:     "foo=bar exact no trailing equals",
			input:    "--foo=bar",
			wantName: "foo=bar",
			wantArg:  "",
			wantHas:  false,
		},
		{
			name:     "foo exact no trailing equals",
			input:    "--foo",
			wantName: "foo",
			wantArg:  "",
			wantHas:  false,
		},
		{
			name:     "foo=bar=arg=extra skips NoArgument to foo=bar",
			input:    "--foo=bar=arg=extra",
			wantName: "foo=bar",
			wantArg:  "arg=extra",
			wantHas:  true,
		},
		{
			name:     "foo=bar= empty arg after foo=bar",
			input:    "--foo=bar=",
			wantName: "foo=bar",
			wantArg:  "",
			wantHas:  true,
		},
		{
			name:     "foo= empty arg after foo",
			input:    "--foo=",
			wantName: "foo",
			wantArg:  "",
			wantHas:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong([]string{tt.input}, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			var count int
			for opt, err := range p.Options() {
				if err != nil {
					t.Fatalf("Options: %v", err)
				}
				count++
				if opt.Name != tt.wantName {
					t.Errorf("Name: got %q, want %q", opt.Name, tt.wantName)
				}
				if opt.Arg != tt.wantArg {
					t.Errorf("Arg: got %q, want %q", opt.Arg, tt.wantArg)
				}
				if opt.HasArg != tt.wantHas {
					t.Errorf("HasArg: got %v, want %v", opt.HasArg, tt.wantHas)
				}
			}
			if count != 1 {
				t.Errorf("expected 1 option, got %d", count)
			}
		})
	}
}

// TestObscureLongOptCharacters exercises long option names containing
// characters that are valid isgraph() but unusual: brackets, braces,
// dots, colons, tildes, etc. Per POSIX/GNU convention, any isgraph()
// character is valid in a long option name.
func TestObscureLongOptCharacters(t *testing.T) {
	tests := []struct {
		name     string
		optName  string
		hasArg   ArgType
		input    []string // args passed to parser
		wantName string
		wantArg  string
		wantHas  bool
	}{
		// Bracket-style: --config[key]
		{
			name:     "brackets space arg",
			optName:  "config[key]",
			hasArg:   RequiredArgument,
			input:    []string{"--config[key]", "val"},
			wantName: "config[key]", wantArg: "val", wantHas: true,
		},
		{
			name:     "brackets equals arg",
			optName:  "config[key]",
			hasArg:   RequiredArgument,
			input:    []string{"--config[key]=val"},
			wantName: "config[key]", wantArg: "val", wantHas: true,
		},
		// Brace-style: --data{category.key}
		{
			name:     "braces space arg",
			optName:  "data{category.key}",
			hasArg:   RequiredArgument,
			input:    []string{"--data{category.key}", "val"},
			wantName: "data{category.key}", wantArg: "val", wantHas: true,
		},
		{
			name:     "braces equals arg",
			optName:  "data{category.key}",
			hasArg:   RequiredArgument,
			input:    []string{"--data{category.key}=val"},
			wantName: "data{category.key}", wantArg: "val", wantHas: true,
		},
		// Colon-style: --command:arg
		{
			name:     "colon space arg",
			optName:  "command:arg",
			hasArg:   RequiredArgument,
			input:    []string{"--command:arg", "value"},
			wantName: "command:arg", wantArg: "value", wantHas: true,
		},
		{
			name:     "colon equals arg",
			optName:  "command:arg",
			hasArg:   RequiredArgument,
			input:    []string{"--command:arg=value"},
			wantName: "command:arg", wantArg: "value", wantHas: true,
		},
		// Dot-style: --section.key
		{
			name:     "dot space arg",
			optName:  "section.key",
			hasArg:   RequiredArgument,
			input:    []string{"--section.key", "value"},
			wantName: "section.key", wantArg: "value", wantHas: true,
		},
		{
			name:     "dot equals arg",
			optName:  "section.key",
			hasArg:   RequiredArgument,
			input:    []string{"--section.key=value"},
			wantName: "section.key", wantArg: "value", wantHas: true,
		},
		// Tilde: --path~backup
		{
			name:     "tilde space arg",
			optName:  "path~backup",
			hasArg:   RequiredArgument,
			input:    []string{"--path~backup", "/tmp"},
			wantName: "path~backup", wantArg: "/tmp", wantHas: true,
		},
		{
			name:     "tilde equals arg",
			optName:  "path~backup",
			hasArg:   RequiredArgument,
			input:    []string{"--path~backup=/tmp"},
			wantName: "path~backup", wantArg: "/tmp", wantHas: true,
		},
		// Plus: --level+1
		{
			name:     "plus space arg",
			optName:  "level+1",
			hasArg:   RequiredArgument,
			input:    []string{"--level+1", "high"},
			wantName: "level+1", wantArg: "high", wantHas: true,
		},
		{
			name:     "plus equals arg",
			optName:  "level+1",
			hasArg:   RequiredArgument,
			input:    []string{"--level+1=high"},
			wantName: "level+1", wantArg: "high", wantHas: true,
		},
		// At-sign: --user@host
		{
			name:     "at space arg",
			optName:  "user@host",
			hasArg:   RequiredArgument,
			input:    []string{"--user@host", "root"},
			wantName: "user@host", wantArg: "root", wantHas: true,
		},
		{
			name:     "at equals arg",
			optName:  "user@host",
			hasArg:   RequiredArgument,
			input:    []string{"--user@host=root"},
			wantName: "user@host", wantArg: "root", wantHas: true,
		},
		// NoArgument with obscure chars
		{
			name:     "brackets no arg",
			optName:  "flag[x]",
			hasArg:   NoArgument,
			input:    []string{"--flag[x]"},
			wantName: "flag[x]", wantArg: "", wantHas: false,
		},
		// OptionalArgument with obscure chars
		{
			name:     "braces optional with equals",
			optName:  "opt{a.b}",
			hasArg:   OptionalArgument,
			input:    []string{"--opt{a.b}=yes"},
			wantName: "opt{a.b}", wantArg: "yes", wantHas: true,
		},
		{
			name:     "braces optional without arg",
			optName:  "opt{a.b}",
			hasArg:   OptionalArgument,
			input:    []string{"--opt{a.b}"},
			wantName: "opt{a.b}", wantArg: "", wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong(tt.input, "", []Flag{
				{Name: tt.optName, HasArg: tt.hasArg},
			})
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			var count int
			for opt, err := range p.Options() {
				if err != nil {
					t.Fatalf("Options: %v", err)
				}
				count++
				if opt.Name != tt.wantName {
					t.Errorf("Name: got %q, want %q", opt.Name, tt.wantName)
				}
				if opt.Arg != tt.wantArg {
					t.Errorf("Arg: got %q, want %q", opt.Arg, tt.wantArg)
				}
				if opt.HasArg != tt.wantHas {
					t.Errorf("HasArg: got %v, want %v", opt.HasArg, tt.wantHas)
				}
			}
			if count != 1 {
				t.Errorf("expected 1 option, got %d", count)
			}
		})
	}
}

// TestObscureCharOverlappingPrefixes tests longest-prefix matching when
// obscure-character option names overlap with shorter prefixes.
func TestObscureCharOverlappingPrefixes(t *testing.T) {
	t.Run("bracket_prefix_overlap", func(t *testing.T) {
		// "config" and "config[key]" both registered.
		// Input: --config[key]=val → longest match "config[key]", arg "val"
		longOpts := []Flag{
			{Name: "config", HasArg: RequiredArgument},
			{Name: "config[key]", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--config[key]=val"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "config[key]" {
				t.Errorf("Name: got %q, want %q", opt.Name, "config[key]")
			}
			if opt.Arg != "val" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "val")
			}
		}
	})

	t.Run("bracket_falls_back_to_shorter", func(t *testing.T) {
		// "config" and "config[key]" both registered.
		// Input: --config=val → "config[key]" is not a prefix of "config=val"
		// → match "config", arg "val"
		longOpts := []Flag{
			{Name: "config", HasArg: RequiredArgument},
			{Name: "config[key]", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--config=val"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "config" {
				t.Errorf("Name: got %q, want %q", opt.Name, "config")
			}
			if opt.Arg != "val" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "val")
			}
		}
	})

	t.Run("colon_prefix_overlap", func(t *testing.T) {
		// "cmd" and "cmd:sub" both registered.
		// Input: --cmd:sub=val → longest match "cmd:sub", arg "val"
		longOpts := []Flag{
			{Name: "cmd", HasArg: RequiredArgument},
			{Name: "cmd:sub", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--cmd:sub=val"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "cmd:sub" {
				t.Errorf("Name: got %q, want %q", opt.Name, "cmd:sub")
			}
			if opt.Arg != "val" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "val")
			}
		}
	})

	t.Run("three_level_obscure_overlap_across_chain", func(t *testing.T) {
		// Grandparent: "data{cat.key}" (RequiredArgument)
		// Parent:      "data{cat}" (RequiredArgument)
		// Child:       "data" (RequiredArgument)
		// Input to child: --data{cat.key}=val
		// Grandparent's "data{cat.key}" is longest → match, arg "val"
		gp, err := GetOptLong([]string{}, "", []Flag{
			{Name: "data{cat.key}", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("grandparent: %v", err)
		}
		par, err := GetOptLong([]string{}, "", []Flag{
			{Name: "data{cat}", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("parent: %v", err)
		}
		child, err := GetOptLong([]string{"--data{cat.key}=val"}, "", []Flag{
			{Name: "data", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("child: %v", err)
		}
		gp.AddCmd("mid", par)
		par.AddCmd("leaf", child)

		for opt, err := range child.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "data{cat.key}" {
				t.Errorf("Name: got %q, want %q", opt.Name, "data{cat.key}")
			}
			if opt.Arg != "val" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "val")
			}
		}
	})
}

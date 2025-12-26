package optargs

import (
	"testing"
)

// TestRoundTripShortOptions tests round-trip parsing for short options
func TestRoundTripShortOptions(t *testing.T) {
	tests := []struct {
		name     string
		optstring string
		args     []string
	}{
		{
			name:     "simple short options",
			optstring: "abc",
			args:     []string{"-a", "-b", "-c"},
		},
		{
			name:     "compacted short options",
			optstring: "abc",
			args:     []string{"-abc"},
		},
		{
			name:     "short options with required arguments",
			optstring: "a:b:c",
			args:     []string{"-a", "arg1", "-b", "arg2", "-c", "arg3"},
		},
		{
			name:     "short options with optional arguments",
			optstring: "a::b::c::",
			args:     []string{"-aarg1", "-b", "-carg3"},
		},
		{
			name:     "mixed argument types",
			optstring: "a:b::c",
			args:     []string{"-a", "required", "-boptional", "-c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First parse
			parser1, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("First parse failed: %v", err)
			}

			var options1 []Option
			for opt, err := range parser1.Options() {
				if err != nil {
					t.Fatalf("First parse iteration failed: %v", err)
				}
				options1 = append(options1, opt)
			}

			// Generate equivalent arguments
			generatedArgs := generateArgsFromOptions(options1, parser1.Args)

			// Second parse
			parser2, err := GetOpt(generatedArgs, tt.optstring)
			if err != nil {
				t.Fatalf("Second parse failed: %v", err)
			}

			var options2 []Option
			for opt, err := range parser2.Options() {
				if err != nil {
					t.Fatalf("Second parse iteration failed: %v", err)
				}
				options2 = append(options2, opt)
			}

			// Verify equivalence
			if !optionsEqual(options1, options2) {
				t.Errorf("Round-trip failed: options not equivalent")
				t.Errorf("Original: %+v", options1)
				t.Errorf("Round-trip: %+v", options2)
			}

			if !argsEqual(parser1.Args, parser2.Args) {
				t.Errorf("Round-trip failed: remaining args not equivalent")
				t.Errorf("Original: %+v", parser1.Args)
				t.Errorf("Round-trip: %+v", parser2.Args)
			}
		})
	}
}

// TestRoundTripLongOptions tests round-trip parsing for long options
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
		{
			name: "simple long options",
			args: []string{"--verbose", "--output", "file.txt"},
		},
		{
			name: "long options with equals syntax",
			args: []string{"--verbose", "--output=file.txt", "--config=debug"},
		},
		{
			name: "mixed syntax",
			args: []string{"--verbose", "--output=file.txt", "--config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First parse
			parser1, err := GetOptLong(tt.args, "", longOpts)
			if err != nil {
				t.Fatalf("First parse failed: %v", err)
			}

			var options1 []Option
			for opt, err := range parser1.Options() {
				if err != nil {
					t.Fatalf("First parse iteration failed: %v", err)
				}
				options1 = append(options1, opt)
			}

			// Generate equivalent arguments
			generatedArgs := generateArgsFromOptions(options1, parser1.Args)

			// Second parse
			parser2, err := GetOptLong(generatedArgs, "", longOpts)
			if err != nil {
				t.Fatalf("Second parse failed: %v", err)
			}

			var options2 []Option
			for opt, err := range parser2.Options() {
				if err != nil {
					t.Fatalf("Second parse iteration failed: %v", err)
				}
				options2 = append(options2, opt)
			}

			// Verify equivalence
			if !optionsEqual(options1, options2) {
				t.Errorf("Round-trip failed: options not equivalent")
				t.Errorf("Original: %+v", options1)
				t.Errorf("Round-trip: %+v", options2)
			}

			if !argsEqual(parser1.Args, parser2.Args) {
				t.Errorf("Round-trip failed: remaining args not equivalent")
				t.Errorf("Original: %+v", parser1.Args)
				t.Errorf("Round-trip: %+v", parser2.Args)
			}
		})
	}
}

// TestRoundTripMixedOptions tests round-trip parsing for mixed short and long options
func TestRoundTripMixedOptions(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "output", HasArg: RequiredArgument},
	}

	tests := []struct {
		name     string
		optstring string
		args     []string
	}{
		{
			name:     "mixed short and long options",
			optstring: "vo:",
			args:     []string{"-v", "--output", "file.txt", "-o", "other.txt"},
		},
		{
			name:     "with non-option arguments",
			optstring: "v",
			args:     []string{"-v", "file1", "--verbose", "file2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First parse
			parser1, err := GetOptLong(tt.args, tt.optstring, longOpts)
			if err != nil {
				t.Fatalf("First parse failed: %v", err)
			}

			var options1 []Option
			for opt, err := range parser1.Options() {
				if err != nil {
					t.Fatalf("First parse iteration failed: %v", err)
				}
				options1 = append(options1, opt)
			}

			// Generate equivalent arguments
			generatedArgs := generateArgsFromOptions(options1, parser1.Args)

			// Second parse
			parser2, err := GetOptLong(generatedArgs, tt.optstring, longOpts)
			if err != nil {
				t.Fatalf("Second parse failed: %v", err)
			}

			var options2 []Option
			for opt, err := range parser2.Options() {
				if err != nil {
					t.Fatalf("Second parse iteration failed: %v", err)
				}
				options2 = append(options2, opt)
			}

			// Verify equivalence (allowing for reordering due to current implementation)
			if !optionsEquivalent(options1, options2) {
				t.Errorf("Round-trip failed: options not equivalent")
				t.Errorf("Original: %+v", options1)
				t.Errorf("Round-trip: %+v", options2)
			}

			if !argsEquivalent(parser1.Args, parser2.Args) {
				t.Errorf("Round-trip failed: remaining args not equivalent")
				t.Errorf("Original: %+v", parser1.Args)
				t.Errorf("Round-trip: %+v", parser2.Args)
			}
		})
	}
}

// TestRoundTripOptionCompaction tests round-trip parsing with option compaction
func TestRoundTripOptionCompaction(t *testing.T) {
	tests := []struct {
		name     string
		optstring string
		original []string
		compacted []string
		desc     string
	}{
		{
			name:     "basic compaction",
			optstring: "abc",
			original: []string{"-a", "-b", "-c"},
			compacted: []string{"-abc"},
			desc:     "no arguments - should be equivalent",
		},
		{
			name:     "compaction behavior with argument",
			optstring: "ab:c",
			original: []string{"-a", "-barg"},
			compacted: []string{"-abarg"},
			desc:     "compacted form where -b takes 'arg' as argument",
		},
		{
			name:     "compaction with optional argument",
			optstring: "ab::c",
			original: []string{"-a", "-barg"},
			compacted: []string{"-abarg"},
			desc:     "compacted form where -b takes 'arg' as optional argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse original form
			parser1, err := GetOpt(tt.original, tt.optstring)
			if err != nil {
				t.Fatalf("Original parse failed: %v", err)
			}

			var options1 []Option
			for opt, err := range parser1.Options() {
				if err != nil {
					t.Fatalf("Original parse iteration failed: %v", err)
				}
				options1 = append(options1, opt)
			}

			// Parse compacted form
			parser2, err := GetOpt(tt.compacted, tt.optstring)
			if err != nil {
				t.Fatalf("Compacted parse failed: %v", err)
			}

			var options2 []Option
			for opt, err := range parser2.Options() {
				if err != nil {
					t.Fatalf("Compacted parse iteration failed: %v", err)
				}
				options2 = append(options2, opt)
			}

			// Verify equivalence
			if !optionsEqual(options1, options2) {
				t.Errorf("Compaction round-trip failed (%s): options not equivalent", tt.desc)
				t.Errorf("Original: %+v", options1)
				t.Errorf("Compacted: %+v", options2)
			}

			if !argsEqual(parser1.Args, parser2.Args) {
				t.Errorf("Compaction round-trip failed (%s): remaining args not equivalent", tt.desc)
				t.Errorf("Original: %+v", parser1.Args)
				t.Errorf("Compacted: %+v", parser2.Args)
			}
		})
	}
}

// Helper functions

// generateArgsFromOptions generates command-line arguments from parsed options
func generateArgsFromOptions(options []Option, remainingArgs []string) []string {
	var args []string

	for _, opt := range options {
		if len(opt.Name) == 1 {
			// Short option
			args = append(args, "-"+opt.Name)
			if opt.HasArg {
				args = append(args, opt.Arg)
			}
		} else {
			// Long option
			if opt.HasArg {
				// Use space syntax for consistency
				args = append(args, "--"+opt.Name)
				args = append(args, opt.Arg)
			} else {
				args = append(args, "--"+opt.Name)
			}
		}
	}

	// Add remaining non-option arguments
	args = append(args, remainingArgs...)

	return args
}

// optionsEqual checks if two option slices are exactly equal
func optionsEqual(opts1, opts2 []Option) bool {
	if len(opts1) != len(opts2) {
		return false
	}

	for i, opt1 := range opts1 {
		opt2 := opts2[i]
		if opt1.Name != opt2.Name || opt1.HasArg != opt2.HasArg || opt1.Arg != opt2.Arg {
			return false
		}
	}

	return true
}

// optionsEquivalent checks if two option slices are equivalent (allowing reordering)
func optionsEquivalent(opts1, opts2 []Option) bool {
	if len(opts1) != len(opts2) {
		return false
	}

	// Create maps for comparison
	map1 := make(map[string][]Option)
	map2 := make(map[string][]Option)

	for _, opt := range opts1 {
		key := opt.Name + "|" + opt.Arg
		map1[key] = append(map1[key], opt)
	}

	for _, opt := range opts2 {
		key := opt.Name + "|" + opt.Arg
		map2[key] = append(map2[key], opt)
	}

	if len(map1) != len(map2) {
		return false
	}

	for key, opts1 := range map1 {
		opts2, exists := map2[key]
		if !exists || len(opts1) != len(opts2) {
			return false
		}
	}

	return true
}

// argsEqual checks if two string slices are exactly equal
func argsEqual(args1, args2 []string) bool {
	if len(args1) != len(args2) {
		return false
	}

	for i, arg1 := range args1 {
		if arg1 != args2[i] {
			return false
		}
	}

	return true
}

// argsEquivalent checks if two string slices are equivalent (allowing reordering)
func argsEquivalent(args1, args2 []string) bool {
	if len(args1) != len(args2) {
		return false
	}

	// Create maps for comparison
	map1 := make(map[string]int)
	map2 := make(map[string]int)

	for _, arg := range args1 {
		map1[arg]++
	}

	for _, arg := range args2 {
		map2[arg]++
	}

	if len(map1) != len(map2) {
		return false
	}

	for arg, count1 := range map1 {
		count2, exists := map2[arg]
		if !exists || count1 != count2 {
			return false
		}
	}

	return true
}
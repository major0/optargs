package optargs

import (
	"testing"
)

// firstErr iterates a parser and returns the first error encountered, or nil.
func firstErr(p *Parser) error {
	for _, err := range p.Options() {
		if err != nil {
			return err
		}
	}
	return nil
}

// findOpt returns the first option with the given name, or nil.
func findOpt(opts []Option, name string) *Option {
	for i := range opts {
		if opts[i].Name == name {
			return &opts[i]
		}
	}
	return nil
}

// TestNegativeArgumentSupport verifies that options requiring arguments accept
// arguments beginning with `-` (e.g. negative numbers) across all delivery
// forms: short separate, short attached, long separate, long equals, optional attached.
func TestNegativeArgumentSupport(t *testing.T) {
	numFlags := []Flag{{Name: "number", HasArg: RequiredArgument}}

	tests := []struct {
		name    string
		args    []string
		optstr  string
		flags   []Flag
		optName string
		wantArg string
	}{
		{"short separate", []string{"-a", "-123"}, "a:", nil, "a", "-123"},
		{"short attached", []string{"-a-456"}, "a:", nil, "a", "-456"},
		{"long separate", []string{"--number", "-789"}, "", numFlags, "number", "-789"},
		{"long equals", []string{"--number=-999"}, "", numFlags, "number", "-999"},
		{"optional attached", []string{"-b-100"}, "b::", nil, "b", "-100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p *Parser
			var err error
			if tt.flags != nil {
				p, err = GetOptLong(tt.args, tt.optstr, tt.flags)
			} else {
				p, err = GetOpt(tt.args, tt.optstr)
			}
			if err != nil {
				t.Fatalf("parser creation failed: %v", err)
			}
			o := findOpt(collectOpts(p), tt.optName)
			if o == nil {
				t.Fatalf("option %q not found", tt.optName)
			}
			if !o.HasArg {
				t.Fatalf("option %q: HasArg = false, want true", tt.optName)
			}
			if o.Arg != tt.wantArg {
				t.Errorf("option %q: Arg = %q, want %q", tt.optName, o.Arg, tt.wantArg)
			}
		})
	}
}

// Feature: test-refactor, Property 12: For any optstring where options are
// redefined, the parser uses the last definition encountered.

// TestOptionRedefinitionHandling verifies that when an option character appears
// multiple times in the optstring, the last definition wins.
func TestOptionRedefinitionHandling(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		optstring string
		optName   string
		wantArg   bool
		wantVal   string
	}{
		{"no-arg to required-arg", []string{"-a", "value"}, "aa:", "a", true, "value"},
		{"required-arg to no-arg", []string{"-b"}, "b:b", "b", false, ""},
		{"optional-arg to required-arg", []string{"-c", "value"}, "c::c:", "c", true, "value"},
		{"triple redef last wins no-arg", []string{"-d"}, "d:d::d", "d", false, ""},
		{"redef with behavior flags", []string{"-e"}, ":e:e", "e", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt(%v, %q) error: %v", tt.args, tt.optstring, err)
			}
			o := findOpt(collectOpts(p), tt.optName)
			if o == nil {
				t.Fatalf("option %q not found", tt.optName)
			}
			if o.HasArg != tt.wantArg {
				t.Errorf("HasArg = %v, want %v", o.HasArg, tt.wantArg)
			}
			if tt.wantArg && o.Arg != tt.wantVal {
				t.Errorf("Arg = %q, want %q", o.Arg, tt.wantVal)
			}
		})
	}
}

// Feature: test-refactor, Property 17: For any ambiguous long option input,
// the parser reports an error per GNU specifications for ambiguity resolution.
// TestAmbiguityResolution verifies that exact long option matches succeed and
// ambiguous prefixes produce errors when multiple options share a common prefix.
func TestAmbiguityResolution(t *testing.T) {
	longOpts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
		{Name: "version", HasArg: NoArgument},
		{Name: "value", HasArg: RequiredArgument},
	}

	tests := []struct {
		name    string
		args    []string
		wantOpt string // expected option name, empty if expecting error
		wantArg string // expected argument value
		wantErr bool   // true if iteration should produce an error
	}{
		{name: "exact_verbose", args: []string{"--verbose"}, wantOpt: "verbose"},
		{name: "exact_version", args: []string{"--version"}, wantOpt: "version"},
		{name: "exact_value_with_arg", args: []string{"--value", "test"}, wantOpt: "value", wantArg: "test"},
		{name: "ambiguous_prefix", args: []string{"--v"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong(tt.args, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong() error: %v", err)
			}

			if tt.wantErr {
				if firstErr(p) == nil {
					t.Error("expected iteration error for ambiguous prefix, got nil")
				}
				return
			}

			opts := collectOpts(p)
			o := findOpt(opts, tt.wantOpt)
			if o == nil {
				t.Fatalf("expected option %q not found in results", tt.wantOpt)
			}
			if tt.wantArg != "" && o.Arg != tt.wantArg {
				t.Errorf("option %q arg = %q, want %q", tt.wantOpt, o.Arg, tt.wantArg)
			}
		})
	}
}

// TestNativeSubcommandDispatchProperty verifies that subcommand dispatch yields
// root options on the root parser and child + inherited options on the child,
// with different char assignments and error modes.
func TestNativeSubcommandDispatchProperty(t *testing.T) {
	cases := []struct {
		name         string
		rootOpt      byte
		inheritedOpt byte
		childOpt     byte
		cmdName      string
		silent       bool
	}{
		{"basic_abc_serve", 'a', 'b', 'c', "serve", false},
		{"different_chars_xyz_build", 'x', 'y', 'z', "build", false},
		{"silent_mode_def_deploy", 'd', 'e', 'f', "deploy", true},
		{"silent_mode_mno_run", 'm', 'n', 'o', "run", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rootOptstring := string(tc.rootOpt) + string(tc.inheritedOpt)
			childOptstring := string(tc.childOpt)
			if tc.silent {
				rootOptstring = ":" + rootOptstring
				childOptstring = ":" + childOptstring
			}

			args := []string{
				"-" + string(tc.rootOpt),
				tc.cmdName,
				"-" + string(tc.childOpt),
				"-" + string(tc.inheritedOpt),
			}

			root, err := GetOpt(args, rootOptstring)
			if err != nil {
				t.Fatalf("Failed to create root parser: %v", err)
			}

			child, err := GetOpt([]string{}, childOptstring)
			if err != nil {
				t.Fatalf("Failed to create child parser: %v", err)
			}
			root.AddCmd(tc.cmdName, child)

			// Root should yield its own option, then dispatch
			rootOpts := collectOpts(root)
			if len(rootOpts) != 1 || rootOpts[0].Name != string(tc.rootOpt) {
				t.Errorf("Expected 1 root option '%s', got %v", string(tc.rootOpt), rootOpts)
			}

			// Child should yield its own option + inherited option
			childOpts := collectOpts(child)
			if len(childOpts) != 2 {
				t.Fatalf("Expected 2 child options, got %d", len(childOpts))
			}
			if childOpts[0].Name != string(tc.childOpt) {
				t.Errorf("Expected child option '%s', got '%s'", string(tc.childOpt), childOpts[0].Name)
			}
			if childOpts[1].Name != string(tc.inheritedOpt) {
				t.Errorf("Expected inherited option '%s', got '%s'", string(tc.inheritedOpt), childOpts[1].Name)
			}
		})
	}
}

package optargs

import (
	"math/rand"
	"os"
	"testing"
	"testing/quick"
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

// Feature: test-refactor, Property 15: For any valid argument list, the
// iterator yields all options exactly once and preserves non-option arguments
// correctly.
// TestIteratorCorrectness verifies that the iterator yields all options exactly
// once in order, preserves arguments, handles compaction, termination, long
// options, and empty input.
func TestIteratorCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		optstr   string
		longOpts []Flag
		useLong  bool
		wantOpts []Option
		wantArgs []string
	}{
		{
			name:   "simple options yielded in order",
			args:   []string{"-a", "-b", "-c"},
			optstr: "abc",
			wantOpts: []Option{
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
			},
		},
		{
			name:   "options with arguments preserve arguments",
			args:   []string{"-a", "arg1", "-b", "arg2"},
			optstr: "a:b:",
			wantOpts: []Option{
				{Name: "a", HasArg: true, Arg: "arg1"},
				{Name: "b", HasArg: true, Arg: "arg2"},
			},
		},
		{
			name:   "non-option arguments preserved in parser.Args",
			args:   []string{"-a", "nonopt1", "-b", "nonopt2"},
			optstr: "ab",
			wantOpts: []Option{
				{Name: "a"},
				{Name: "b"},
			},
			wantArgs: []string{"nonopt1", "nonopt2"},
		},
		{
			name:   "compacted options expanded correctly",
			args:   []string{"-abc"},
			optstr: "abc",
			wantOpts: []Option{
				{Name: "a"},
				{Name: "b"},
				{Name: "c"},
			},
		},
		{
			name:   "-- termination stops option processing",
			args:   []string{"-a", "--", "-b", "nonopt"},
			optstr: "ab",
			wantOpts: []Option{
				{Name: "a"},
			},
			wantArgs: []string{"-b", "nonopt"},
		},
		{
			name:    "long options yielded correctly",
			args:    []string{"--verbose", "--output", "file.txt"},
			useLong: true,
			longOpts: []Flag{
				{Name: "verbose", HasArg: NoArgument},
				{Name: "output", HasArg: RequiredArgument},
			},
			wantOpts: []Option{
				{Name: "verbose"},
				{Name: "output", HasArg: true, Arg: "file.txt"},
			},
		},
		{
			name:   "empty argument list yields no options",
			args:   []string{},
			optstr: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p *Parser
			var err error
			if tt.useLong {
				p, err = GetOptLong(tt.args, tt.optstr, tt.longOpts)
			} else {
				p, err = GetOpt(tt.args, tt.optstr)
			}
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			opts := collectOpts(p)

			if len(opts) != len(tt.wantOpts) {
				t.Fatalf("got %d options, want %d", len(opts), len(tt.wantOpts))
			}
			for i, want := range tt.wantOpts {
				got := opts[i]
				if got.Name != want.Name {
					t.Errorf("opt[%d].Name = %q, want %q", i, got.Name, want.Name)
				}
				if got.HasArg != want.HasArg {
					t.Errorf("opt[%d].HasArg = %v, want %v", i, got.HasArg, want.HasArg)
				}
				if got.Arg != want.Arg {
					t.Errorf("opt[%d].Arg = %q, want %q", i, got.Arg, want.Arg)
				}
			}

			if len(tt.wantArgs) != len(p.Args) {
				t.Fatalf("got %d args, want %d", len(p.Args), len(tt.wantArgs))
			}
			for i, want := range tt.wantArgs {
				if p.Args[i] != want {
					t.Errorf("Args[%d] = %q, want %q", i, p.Args[i], want)
				}
			}
		})
	}
}

// TestEnvironmentVariableBehavior verifies that POSIXLY_CORRECT and the `+`
// optstring prefix both stop option parsing at the first non-option argument.
func TestEnvironmentVariableBehavior(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		optstring string
		wantAll   int // options parsed in normal (GNU) mode
		wantPosix int // options parsed in POSIXLY_CORRECT / + prefix mode
	}{
		{
			name:      "1 initial + nonopt + 1 trailing",
			args:      []string{"-a", "nonopt", "-b"},
			optstring: "abc",
			wantAll:   2,
			wantPosix: 1,
		},
		{
			name:      "2 initial + nonopt + 2 trailing",
			args:      []string{"-a", "-a", "nonopt", "-b", "-b"},
			optstring: "abc",
			wantAll:   4,
			wantPosix: 2,
		},
		{
			name:      "1 initial + nonopt + 3 trailing",
			args:      []string{"-a", "nonopt", "-b", "-b", "-b"},
			optstring: "abc",
			wantAll:   4,
			wantPosix: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/normal", func(t *testing.T) {
			t.Setenv("POSIXLY_CORRECT", "")
			os.Unsetenv("POSIXLY_CORRECT")
			p, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt error: %v", err)
			}
			if got := len(collectOpts(p)); got != tt.wantAll {
				t.Errorf("normal mode: got %d opts, want %d", got, tt.wantAll)
			}
		})

		t.Run(tt.name+"/POSIXLY_CORRECT", func(t *testing.T) {
			t.Setenv("POSIXLY_CORRECT", "1")
			p, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt error: %v", err)
			}
			if got := len(collectOpts(p)); got != tt.wantPosix {
				t.Errorf("POSIXLY_CORRECT: got %d opts, want %d", got, tt.wantPosix)
			}
		})

		t.Run(tt.name+"/plus_prefix", func(t *testing.T) {
			t.Setenv("POSIXLY_CORRECT", "")
			os.Unsetenv("POSIXLY_CORRECT")
			p, err := GetOpt(tt.args, "+"+tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt error: %v", err)
			}
			if got := len(collectOpts(p)); got != tt.wantPosix {
				t.Errorf("+ prefix: got %d opts, want %d", got, tt.wantPosix)
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

// Feature: test-refactor, Property 18: For any parser with registered
// subcommands, the iterator dispatches to the correct child parser when a
// non-option argument matches a subcommand name, and unknown options in child
// parsers are resolved by walking the parent chain. Both verbose and silent
// error modes work correctly through the chain.
func TestProperty18_NativeSubcommandDispatch(t *testing.T) {
	validShortOpts := []byte("abcdefghijklmnopqrstuvwxyz")

	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		perm := rng.Perm(len(validShortOpts))
		rootOptChar := validShortOpts[perm[0]]
		childOptChar := validShortOpts[perm[1]]
		inheritedOptChar := validShortOpts[perm[2]]

		cmdNames := []string{"serve", "build", "test", "deploy", "run"}
		cmdName := cmdNames[rng.Intn(len(cmdNames))]

		silentMode := rng.Intn(2) == 0

		rootOptstring := string(rootOptChar) + string(inheritedOptChar)
		childOptstring := string(childOptChar)
		if silentMode {
			rootOptstring = ":" + rootOptstring
			childOptstring = ":" + childOptstring
		}

		args := []string{
			"-" + string(rootOptChar),
			cmdName,
			"-" + string(childOptChar),
			"-" + string(inheritedOptChar),
		}

		root, err := GetOpt(args, rootOptstring)
		if err != nil {
			t.Logf("Failed to create root parser: %v", err)
			return false
		}

		child, err := GetOpt([]string{}, childOptstring)
		if err != nil {
			t.Logf("Failed to create child parser: %v", err)
			return false
		}
		root.AddCmd(cmdName, child)

		if child.HasCommands() {
			return false
		}

		// Root should yield its own option, then dispatch
		rootOpts := collectOpts(root)
		if len(rootOpts) != 1 || rootOpts[0].Name != string(rootOptChar) {
			t.Logf("Expected 1 root option '%s', got %d opts", string(rootOptChar), len(rootOpts))
			return false
		}

		// Child should yield its own option + inherited option
		childOpts := collectOpts(child)
		if len(childOpts) != 2 {
			t.Logf("Expected 2 child options, got %d", len(childOpts))
			return false
		}
		if childOpts[0].Name != string(childOptChar) {
			t.Logf("Expected child option '%s', got '%s'", string(childOptChar), childOpts[0].Name)
			return false
		}
		if childOpts[1].Name != string(inheritedOptChar) {
			t.Logf("Expected inherited option '%s', got '%s'", string(inheritedOptChar), childOpts[1].Name)
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 18 failed: %v", err)
	}
}

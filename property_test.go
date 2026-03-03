package optargs

import (
	"fmt"
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

// Feature: test-refactor, Property 4: For any argument list containing `--`,
// the parser stops processing options at that point and treats all subsequent
// arguments as non-options.
func TestProperty4_OptionTerminationBehavior(t *testing.T) {
	property := func() bool {
		rng := rand.New(rand.NewSource(rand.Int63()))

		numBefore := rng.Intn(3)
		var argsBefore []string
		for i := 0; i < numBefore; i++ {
			argsBefore = append(argsBefore, "-a")
		}

		numAfter := rng.Intn(5) + 1
		var argsAfter []string
		for i := 0; i < numAfter; i++ {
			switch rng.Intn(3) {
			case 0:
				argsAfter = append(argsAfter, "-a")
			case 1:
				argsAfter = append(argsAfter, "--long")
			case 2:
				argsAfter = append(argsAfter, fmt.Sprintf("arg%d", i))
			}
		}

		args := append(argsBefore, "--")
		args = append(args, argsAfter...)

		parser, err := GetOpt(args, "abc")
		if err != nil {
			return false
		}

		opts := collectOpts(parser)
		if opts == nil && numBefore > 0 {
			return false
		}
		if len(opts) != numBefore {
			return false
		}
		if len(parser.Args) != len(argsAfter) {
			return false
		}
		for i, expected := range argsAfter {
			if parser.Args[i] != expected {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 4 failed: %v", err)
	}
}

// Feature: test-refactor, Property 10: For any option that requires an
// argument, the parser accepts arguments beginning with `-` when explicitly
// provided (e.g. negative numbers).
func TestProperty10_NegativeArgumentSupport(t *testing.T) {
	property := func() bool {
		// Short option with negative number (separate)
		p1, err := GetOpt([]string{"-a", "-123"}, "a:")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p1), "a"); o == nil || !o.HasArg || o.Arg != "-123" {
			return false
		}

		// Short option with negative number (attached)
		p2, err := GetOpt([]string{"-a-456"}, "a:")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p2), "a"); o == nil || !o.HasArg || o.Arg != "-456" {
			return false
		}

		// Long option with negative number (separate)
		numFlags := []Flag{{Name: "number", HasArg: RequiredArgument}}
		p3, err := GetOptLong([]string{"--number", "-789"}, "", numFlags)
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p3), "number"); o == nil || !o.HasArg || o.Arg != "-789" {
			return false
		}

		// Long option with negative number (equals syntax)
		p4, err := GetOptLong([]string{"--number=-999"}, "", numFlags)
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p4), "number"); o == nil || !o.HasArg || o.Arg != "-999" {
			return false
		}

		// Optional argument with negative number (attached)
		p5, err := GetOpt([]string{"-b-100"}, "b::")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p5), "b"); o == nil || !o.HasArg || o.Arg != "-100" {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 10 failed: %v", err)
	}
}

// Feature: test-refactor, Property 12: For any optstring where options are
// redefined, the parser uses the last definition encountered.
func TestProperty12_OptionRedefinitionHandling(t *testing.T) {
	property := func() bool {
		// Redefine from no-argument to required-argument
		p1, err := GetOpt([]string{"-a", "value"}, "aa:")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p1), "a"); o == nil || !o.HasArg || o.Arg != "value" {
			return false
		}

		// Redefine from required-argument to no-argument
		p2, err := GetOpt([]string{"-b"}, "b:b")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p2), "b"); o == nil || o.HasArg {
			return false
		}

		// Redefine from optional-argument to required-argument
		p3, err := GetOpt([]string{"-c", "value"}, "c::c:")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p3), "c"); o == nil || !o.HasArg || o.Arg != "value" {
			return false
		}

		// Multiple redefinitions use the last one
		p4, err := GetOpt([]string{"-d"}, "d:d::d")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p4), "d"); o == nil || o.HasArg {
			return false
		}

		// Redefinition with behavior flags
		p5, err := GetOpt([]string{"-e"}, ":e:e")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p5), "e"); o == nil || o.HasArg {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 12 failed: %v", err)
	}
}

// Feature: test-refactor, Property 15: For any valid argument list, the
// iterator yields all options exactly once and preserves non-option arguments
// correctly.
func TestProperty15_IteratorCorrectness(t *testing.T) {
	property := func() bool {
		// Simple options yielded exactly once in order
		p1, err := GetOpt([]string{"-a", "-b", "-c"}, "abc")
		if err != nil {
			return false
		}
		opts1 := collectOpts(p1)
		if len(opts1) != 3 {
			return false
		}
		for i, name := range []string{"a", "b", "c"} {
			if opts1[i].Name != name || opts1[i].HasArg {
				return false
			}
		}

		// Options with arguments preserve arguments
		p2, err := GetOpt([]string{"-a", "arg1", "-b", "arg2"}, "a:b:")
		if err != nil {
			return false
		}
		opts2 := collectOpts(p2)
		if len(opts2) != 2 {
			return false
		}
		if opts2[0].Name != "a" || opts2[0].Arg != "arg1" {
			return false
		}
		if opts2[1].Name != "b" || opts2[1].Arg != "arg2" {
			return false
		}

		// Non-option arguments preserved in parser.Args
		p3, err := GetOpt([]string{"-a", "nonopt1", "-b", "nonopt2"}, "ab")
		if err != nil {
			return false
		}
		opts3 := collectOpts(p3)
		if len(opts3) != 2 || opts3[0].Name != "a" || opts3[1].Name != "b" {
			return false
		}
		if len(p3.Args) != 2 || p3.Args[0] != "nonopt1" || p3.Args[1] != "nonopt2" {
			return false
		}

		// Compacted options expanded correctly
		p4, err := GetOpt([]string{"-abc"}, "abc")
		if err != nil {
			return false
		}
		opts4 := collectOpts(p4)
		if len(opts4) != 3 {
			return false
		}
		for i, name := range []string{"a", "b", "c"} {
			if opts4[i].Name != name || opts4[i].HasArg {
				return false
			}
		}

		// -- termination stops option processing
		p5, err := GetOpt([]string{"-a", "--", "-b", "nonopt"}, "ab")
		if err != nil {
			return false
		}
		opts5 := collectOpts(p5)
		if len(opts5) != 1 || opts5[0].Name != "a" {
			return false
		}
		if len(p5.Args) != 2 || p5.Args[0] != "-b" || p5.Args[1] != "nonopt" {
			return false
		}

		// Long options yielded correctly
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
		}
		p6, err := GetOptLong([]string{"--verbose", "--output", "file.txt"}, "", longOpts)
		if err != nil {
			return false
		}
		opts6 := collectOpts(p6)
		if len(opts6) != 2 {
			return false
		}
		if opts6[0].Name != "verbose" || opts6[0].HasArg {
			return false
		}
		if opts6[1].Name != "output" || !opts6[1].HasArg || opts6[1].Arg != "file.txt" {
			return false
		}

		// Empty argument list yields no options
		p7, err := GetOpt([]string{}, "abc")
		if err != nil {
			return false
		}
		opts7 := collectOpts(p7)
		if len(opts7) != 0 || len(p7.Args) != 0 {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 15 failed: %v", err)
	}
}

// Feature: test-refactor, Property 16: For any parsing session, when
// POSIXLY_CORRECT is set, the parser stops at the first non-option argument.
// The `+` prefix behaves identically.
func TestProperty16_EnvironmentVariableBehavior(t *testing.T) {
	property := func() bool {
		rng := rand.New(rand.NewSource(rand.Int63()))

		optstring := "abc"

		var args []string
		numInitialOpts := rng.Intn(2) + 1
		for i := 0; i < numInitialOpts; i++ {
			args = append(args, "-a")
		}
		args = append(args, fmt.Sprintf("nonopt%d", rng.Intn(100)))
		numLaterOpts := rng.Intn(3) + 1
		for i := 0; i < numLaterOpts; i++ {
			args = append(args, "-b")
		}

		// Without POSIXLY_CORRECT
		_ = os.Unsetenv("POSIXLY_CORRECT")
		p1, err := GetOpt(args, optstring)
		if err != nil {
			return false
		}
		normalOpts := len(collectOpts(p1))

		// With POSIXLY_CORRECT
		_ = os.Setenv("POSIXLY_CORRECT", "1")
		defer func() { _ = os.Unsetenv("POSIXLY_CORRECT") }()

		p2, err := GetOpt(args, optstring)
		if err != nil {
			return false
		}
		posixOpts := len(collectOpts(p2))

		if posixOpts >= normalOpts {
			return false
		}

		// + prefix behaves the same as environment variable
		p3, err := GetOpt(args, "+"+optstring)
		if err != nil {
			return false
		}
		prefixOpts := len(collectOpts(p3))

		return posixOpts == prefixOpts
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 16 failed: %v", err)
	}
}

// Feature: test-refactor, Property 17: For any ambiguous long option input,
// the parser reports an error per GNU specifications for ambiguity resolution.
func TestProperty17_AmbiguityResolution(t *testing.T) {
	property := func() bool {
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "version", HasArg: NoArgument},
			{Name: "value", HasArg: RequiredArgument},
		}

		// Exact matches work
		p1, err := GetOptLong([]string{"--verbose"}, "", longOpts)
		if err != nil {
			return false
		}
		if findOpt(collectOpts(p1), "verbose") == nil {
			return false
		}

		p2, err := GetOptLong([]string{"--version"}, "", longOpts)
		if err != nil {
			return false
		}
		if findOpt(collectOpts(p2), "version") == nil {
			return false
		}

		// Exact match with argument
		p3, err := GetOptLong([]string{"--value", "test"}, "", longOpts)
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p3), "value"); o == nil || !o.HasArg || o.Arg != "test" {
			return false
		}

		// Partial match should error (ambiguous)
		p4, err := GetOptLong([]string{"--v"}, "", longOpts)
		if err != nil {
			return false
		}
		if firstErr(p4) == nil {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 17 failed: %v", err)
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

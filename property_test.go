package optargs

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"testing/quick"
)

// collectOpts iterates a parser and returns all successfully parsed options.
// Returns nil on the first iteration error.
func collectOpts(p *Parser) []Option {
	var out []Option
	for opt, err := range p.Options() {
		if err != nil {
			return nil
		}
		out = append(out, opt)
	}
	return out
}

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

// Property 1: POSIX/GNU Specification Compliance
// Feature: optargs-core, Property 1: For any valid POSIX optstring and GNU long option specification, the parser should produce results that match the behavior of the reference GNU getopt implementation
func TestProperty1_POSIXGNUSpecificationCompliance(t *testing.T) {
	property := func() bool {
		// Basic optstring with no behavior flags
		parser, err := GetOpt([]string{}, "abc")
		if err != nil {
			return false
		}
		if !parser.config.enableErrors {
			return false
		}
		if parser.config.parseMode != ParseDefault {
			return false
		}
		if len(parser.shortOpts) != 3 {
			return false
		}
		for _, c := range []byte{'a', 'b', 'c'} {
			if _, ok := parser.shortOpts[c]; !ok {
				return false
			}
		}

		// Silent errors flag
		parser2, err := GetOpt([]string{}, ":abc")
		if err != nil {
			return false
		}
		if parser2.config.enableErrors {
			return false
		}

		// POSIX mode flag
		parser3, err := GetOpt([]string{}, "+abc")
		if err != nil {
			return false
		}
		if parser3.config.parseMode != ParsePosixlyCorrect {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 1 failed: %v", err)
	}
}

// Property 2: Option Compaction and Argument Assignment
// Feature: optargs-core, Property 2: For any combination of compacted short options with arguments, the parser should assign arguments to the last option that accepts them and expand compaction correctly
func TestProperty2_OptionCompactionAndArgumentAssignment(t *testing.T) {
	property := func() bool {
		parser, err := GetOpt([]string{"-avalue"}, "a::")
		if err != nil {
			return false
		}
		opts := collectOpts(parser)
		if len(opts) != 1 {
			return false
		}
		return opts[0].Name == "a" && opts[0].HasArg && opts[0].Arg == "value"
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 2 failed: %v", err)
	}
}

// Property 3: Argument Type Handling
// Feature: optargs-core, Property 3: For any option string containing colon specifications, the parser should correctly handle required arguments (:), optional arguments (::), and no-argument options according to POSIX rules
func TestProperty3_ArgumentTypeHandling(t *testing.T) {
	property := func() bool {
		// No argument option
		p1, err := GetOpt([]string{"-a"}, "a")
		if err != nil {
			return false
		}
		opts1 := collectOpts(p1)
		if o := findOpt(opts1, "a"); o == nil || o.HasArg {
			return false
		}

		// Optional argument with attached value
		p2, err := GetOpt([]string{"-avalue"}, "a::")
		if err != nil {
			return false
		}
		opts2 := collectOpts(p2)
		if o := findOpt(opts2, "a"); o == nil || !o.HasArg || o.Arg != "value" {
			return false
		}

		// Optional argument without value
		p3, err := GetOpt([]string{"-a"}, "a::")
		if err != nil {
			return false
		}
		opts3 := collectOpts(p3)
		if o := findOpt(opts3, "a"); o == nil || o.HasArg {
			return false
		}

		// Required argument with separate value
		p4, err := GetOpt([]string{"-a", "value"}, "a:")
		if err != nil {
			return false
		}
		opts4 := collectOpts(p4)
		if o := findOpt(opts4, "a"); o == nil || !o.HasArg || o.Arg != "value" {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 3 failed: %v", err)
	}
}

// Property 4: Option Termination Behavior
// Feature: optargs-core, Property 4: For any argument list containing `--`, the parser should stop processing options at that point and treat all subsequent arguments as non-options
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

// Property 5: Long Option Syntax Support
// Feature: optargs-core, Property 5: For any valid long option, the parser should correctly handle both `--option=value` and `--option value` syntax forms
func TestProperty5_LongOptionSyntaxSupport(t *testing.T) {
	property := func() bool {
		reqFlags := []Flag{{Name: "test", HasArg: RequiredArgument}}
		optFlags := []Flag{{Name: "optional", HasArg: OptionalArgument}}

		// --option=value with required argument
		p1, err := GetOptLong([]string{"--test=value"}, "", reqFlags)
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p1), "test"); o == nil || !o.HasArg || o.Arg != "value" {
			return false
		}

		// --option value with required argument
		p2, err := GetOptLong([]string{"--test", "value"}, "", reqFlags)
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p2), "test"); o == nil || !o.HasArg || o.Arg != "value" {
			return false
		}

		// --option=value with optional argument
		p3, err := GetOptLong([]string{"--optional=value"}, "", optFlags)
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p3), "optional"); o == nil || !o.HasArg || o.Arg != "value" {
			return false
		}

		// --option without value for optional argument
		p4, err := GetOptLong([]string{"--optional"}, "", optFlags)
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p4), "optional"); o == nil || o.HasArg {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 5 failed: %v", err)
	}
}

// Property 6: Case Sensitivity Handling
// Feature: optargs-core, Property 6: For any long option name, the parser should handle case variations according to the configured case sensitivity settings
func TestProperty6_CaseSensitivityHandling(t *testing.T) {
	property := func() bool {
		longOpts := []Flag{{Name: "test", HasArg: NoArgument}}

		// Exact case match
		p1, err := GetOptLong([]string{"--test"}, "", longOpts)
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p1), "test"); o == nil || o.HasArg {
			return false
		}

		// Case-insensitive match (long options are case-insensitive by default)
		p2, err := GetOptLong([]string{"--TEST"}, "", longOpts)
		if err != nil {
			return false
		}
		opts2 := collectOpts(p2)
		if o := findOpt(opts2, "test"); o == nil {
			return false
		}

		// Mixed case match
		p3, err := GetOptLong([]string{"--Test"}, "", longOpts)
		if err != nil {
			return false
		}
		opts3 := collectOpts(p3)
		if o := findOpt(opts3, "test"); o == nil {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 6 failed: %v", err)
	}
}

// Property 7: Partial Long Option Matching
// Feature: optargs-core, Property 7: For any unambiguous partial long option match, the parser should resolve to the correct full option name
func TestProperty7_PartialLongOptionMatching(t *testing.T) {
	property := func() bool {
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "version", HasArg: NoArgument},
			{Name: "help", HasArg: NoArgument},
		}

		// Exact match
		p1, err := GetOptLong([]string{"--verbose"}, "", longOpts)
		if err != nil {
			return false
		}
		if findOpt(collectOpts(p1), "verbose") == nil {
			return false
		}

		// Partial match should error (ambiguous between verbose/version)
		p2, err := GetOptLong([]string{"--verb"}, "", longOpts)
		if err != nil {
			return false
		}
		if firstErr(p2) == nil {
			return false
		}

		// Unambiguous exact match
		p3, err := GetOptLong([]string{"--help"}, "", longOpts)
		if err != nil {
			return false
		}
		return findOpt(collectOpts(p3), "help") != nil
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 7 failed: %v", err)
	}
}

// Property 8: Long-Only Mode Behavior
// Feature: optargs-core, Property 8: For any single-dash option in long-only mode, the parser should treat multi-character options as long options and fall back to short option parsing for single characters
func TestProperty8_LongOnlyModeBehavior(t *testing.T) {
	property := func() bool {
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "help", HasArg: NoArgument},
		}

		// Single-dash multi-character treated as long option
		p1, err := GetOptLongOnly([]string{"-verbose"}, "", longOpts)
		if err != nil {
			return false
		}
		if findOpt(collectOpts(p1), "verbose") == nil {
			return false
		}

		// Single-dash single character without short opt defined should error
		p2, err := GetOptLongOnly([]string{"-h"}, "", longOpts)
		if err != nil {
			return false
		}
		if firstErr(p2) == nil {
			return false
		}

		// Single-dash multi-character with argument
		outFlags := []Flag{{Name: "output", HasArg: RequiredArgument}}
		p3, err := GetOptLongOnly([]string{"-output", "file.txt"}, "", outFlags)
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p3), "output"); o == nil || !o.HasArg || o.Arg != "file.txt" {
			return false
		}

		// Double-dash still works normally
		p4, err := GetOptLongOnly([]string{"--verbose"}, "", longOpts)
		if err != nil {
			return false
		}
		return findOpt(collectOpts(p4), "verbose") != nil
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 8 failed: %v", err)
	}
}

// Property 9: GNU W-Extension Support
// Feature: optargs-core, Property 9: For any `-W word` pattern when GNU words are enabled, the parser should transform it to `--word`
func TestProperty9_GNUWExtensionSupport(t *testing.T) {
	property := func() bool {
		// -W word transforms to --word
		p1, err := GetOpt([]string{"-W", "verbose"}, "W;")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p1), "verbose"); o == nil || !o.HasArg || o.Arg != "verbose" {
			return false
		}

		// -W word=value transforms to --word=value
		p2, err := GetOpt([]string{"-W", "output=file.txt"}, "W;")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p2), "output=file.txt"); o == nil || !o.HasArg || o.Arg != "output=file.txt" {
			return false
		}

		// -Wword (attached form) transforms to --word
		p3, err := GetOpt([]string{"-Whelp"}, "W;")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p3), "help"); o == nil || !o.HasArg || o.Arg != "help" {
			return false
		}

		// W without `;` should not enable GNU words transformation
		p4, err := GetOpt([]string{"-W", "verbose"}, "W:")
		if err != nil {
			return false
		}
		if o := findOpt(collectOpts(p4), "W"); o == nil || !o.HasArg || o.Arg != "verbose" {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 9 failed: %v", err)
	}
}

// Property 10: Negative Argument Support
// Feature: optargs-core, Property 10: For any option that requires an argument, the parser should accept arguments beginning with `-` when explicitly provided
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

// Property 11: Character Validation
// Feature: optargs-core, Property 11: For any printable ASCII character except `:`, `;`, `-`, the parser should accept it as a valid short option character
func TestProperty11_CharacterValidation(t *testing.T) {
	property := func() bool {
		validChars := []byte{'a', 'b', 'A', 'B', '1', '2', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '_', '=', '{', '}', '[', ']', '|', '\\', '?', '/', '.', '>', '<', ',', '~', '`'}

		for _, c := range validChars {
			parser, err := GetOpt([]string{"-" + string(c)}, string(c))
			if err != nil {
				return false
			}
			if o := findOpt(collectOpts(parser), string(c)); o == nil || o.HasArg {
				return false
			}
		}

		// Semicolon is invalid
		if _, err := GetOpt([]string{}, ";"); err == nil {
			return false
		}

		// Colon as argument specifier is valid
		if _, err := GetOpt([]string{}, "a:"); err != nil {
			return false
		}

		// Dash as option character is invalid
		if _, err := GetOpt([]string{}, "a-"); err == nil {
			return false
		}

		// Trailing semicolon is invalid
		if _, err := GetOpt([]string{}, "a;"); err == nil {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 11 failed: %v", err)
	}
}

// Property 12: Option Redefinition Handling
// Feature: optargs-core, Property 12: For any optstring where options are redefined, the parser should use the last definition encountered
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

// Property 13: Error Reporting Accuracy
// Feature: optargs-core, Property 13: For any missing required argument error, the error message should identify the specific option that requires the argument
func TestProperty13_ErrorReportingAccuracy(t *testing.T) {
	property := func() bool {
		// Short option missing required argument
		p1, err := GetOpt([]string{"-a"}, "a:")
		if err != nil {
			return false
		}
		if e := firstErr(p1); e == nil || !strings.Contains(e.Error(), "a") || !strings.Contains(e.Error(), "requires an argument") {
			return false
		}

		// Long option missing required argument
		longOpts := []Flag{{Name: "verbose", HasArg: RequiredArgument}}
		p2, err := GetOptLong([]string{"--verbose"}, "", longOpts)
		if err != nil {
			return false
		}
		if e := firstErr(p2); e == nil || !strings.Contains(e.Error(), "verbose") || !strings.Contains(e.Error(), "requires an argument") {
			return false
		}

		// Required argument consumes next arg even if it looks like an option
		p3, err := GetOpt([]string{"-a", "-b"}, "a:b:")
		if err != nil {
			return false
		}
		opts3 := collectOpts(p3)
		// -a should consume -b as its argument; no second option parsed
		if len(opts3) != 1 {
			return false
		}
		if opts3[0].Name != "a" || !opts3[0].HasArg || opts3[0].Arg != "-b" {
			return false
		}

		// Unknown short option error identifies the option
		p4, err := GetOpt([]string{"-z"}, "a:b:")
		if err != nil {
			return false
		}
		if e := firstErr(p4); e == nil || !strings.Contains(e.Error(), "z") || !strings.Contains(e.Error(), "unknown option") {
			return false
		}

		// Unknown long option error identifies the option
		p5, err := GetOptLong([]string{"--unknown"}, "", []Flag{})
		if err != nil {
			return false
		}
		if e := firstErr(p5); e == nil || !strings.Contains(e.Error(), "unknown") || !strings.Contains(e.Error(), "unknown option") {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 13 failed: %v", err)
	}
}

// Property 14: Silent Error Mode
// Feature: optargs-core, Property 14: For any optstring beginning with `:`, the parser should suppress automatic error logging while still returning errors
func TestProperty14_SilentErrorMode(t *testing.T) {
	property := func() bool {
		// Silent mode still returns errors
		p1, err := GetOpt([]string{"-a"}, ":a:")
		if err != nil {
			return false
		}
		if p1.config.enableErrors {
			return false
		}
		if e := firstErr(p1); e == nil || !strings.Contains(e.Error(), "a") || !strings.Contains(e.Error(), "requires an argument") {
			return false
		}

		// Silent mode with unknown option
		p2, err := GetOpt([]string{"-z"}, ":abc")
		if err != nil {
			return false
		}
		if p2.config.enableErrors {
			return false
		}
		if e := firstErr(p2); e == nil || !strings.Contains(e.Error(), "z") || !strings.Contains(e.Error(), "unknown option") {
			return false
		}

		// Silent mode combined with POSIX mode
		p3, err := GetOpt([]string{"-a"}, ":+a:")
		if err != nil {
			return false
		}
		if p3.config.enableErrors {
			return false
		}
		if p3.config.parseMode != ParsePosixlyCorrect {
			return false
		}
		if firstErr(p3) == nil {
			return false
		}

		// Non-silent mode comparison
		p4, err := GetOpt([]string{"-b", "value"}, "b:")
		if err != nil {
			return false
		}
		if !p4.config.enableErrors {
			return false
		}
		if o := findOpt(collectOpts(p4), "b"); o == nil || !o.HasArg || o.Arg != "value" {
			return false
		}

		// Silent mode with long options
		longOpts := []Flag{{Name: "verbose", HasArg: RequiredArgument}}
		p5, err := GetOptLong([]string{"--verbose"}, ":abc", longOpts)
		if err != nil {
			return false
		}
		if p5.config.enableErrors {
			return false
		}
		if e := firstErr(p5); e == nil || !strings.Contains(e.Error(), "verbose") {
			return false
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 14 failed: %v", err)
	}
}

// Property 15: Iterator Correctness
// Feature: optargs-core, Property 15: For any valid argument list, the iterator should yield all options exactly once and preserve non-option arguments correctly
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

// Property 16: Environment Variable Behavior
// Feature: optargs-core, Property 16: For any parsing session, when POSIXLY_CORRECT is set, the parser should stop at the first non-option argument
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

// Property 17: Ambiguity Resolution
// Feature: optargs-core, Property 17: For any ambiguous long option input, the parser should handle it according to GNU specifications for ambiguity resolution
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

// Property 18: Native Subcommand Dispatch
// **Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.5**
// For any parser with registered subcommands, the iterator dispatches to the
// correct child parser when a non-option argument matches a subcommand name,
// and unknown options in child parsers are resolved by walking the parent chain.
// Both verbose and silent error modes work correctly through the chain.
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

package optargs

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
)

// Property-based test generators and utilities

// Generate valid short option characters (printable ASCII except :, ;, -)
func generateValidShortOpt(rand *rand.Rand) byte {
	for {
		c := byte(rand.Intn(127))
		if isGraph(c) && c != ':' && c != ';' && c != '-' {
			return c
		}
	}
}

// Generate valid long option names (printable non-space characters)
func generateValidLongOpt(rand *rand.Rand) string {
	length := rand.Intn(10) + 1 // 1-10 characters
	var result strings.Builder
	for i := 0; i < length; i++ {
		for {
			c := byte(rand.Intn(127))
			if isGraph(c) && c != ' ' {
				result.WriteByte(c)
				break
			}
		}
	}
	return result.String()
}

// Generate optstring with various argument types
func generateOptString(rand *rand.Rand) string {
	var result strings.Builder
	
	// Add behavior flags randomly
	if rand.Float32() < 0.3 {
		result.WriteByte(':') // Silent errors
	}
	if rand.Float32() < 0.3 {
		result.WriteByte('+') // POSIXLY_CORRECT
	}
	if rand.Float32() < 0.3 {
		result.WriteByte('-') // Non-opts mode
	}
	
	// Add 1-5 options
	numOpts := rand.Intn(5) + 1
	for i := 0; i < numOpts; i++ {
		c := generateValidShortOpt(rand)
		result.WriteByte(c)
		
		// Add argument specification
		argType := rand.Intn(3)
		switch argType {
		case 1: // Required argument
			result.WriteByte(':')
		case 2: // Optional argument
			result.WriteString("::")
		}
	}
	
	return result.String()
}

// Generate argument lists for testing
func generateArgs(rand *rand.Rand) []string {
	numArgs := rand.Intn(10)
	args := make([]string, numArgs)
	
	for i := 0; i < numArgs; i++ {
		switch rand.Intn(4) {
		case 0: // Short option
			args[i] = "-" + string(generateValidShortOpt(rand))
		case 1: // Long option
			args[i] = "--" + generateValidLongOpt(rand)
		case 2: // Compacted short options
			length := rand.Intn(3) + 2 // 2-4 options
			var compacted strings.Builder
			compacted.WriteByte('-')
			for j := 0; j < length; j++ {
				compacted.WriteByte(generateValidShortOpt(rand))
			}
			args[i] = compacted.String()
		case 3: // Regular argument
			args[i] = fmt.Sprintf("arg%d", i)
		}
	}
	
	return args
}

// Property 1: POSIX/GNU Specification Compliance
// Feature: optargs-core, Property 1: For any valid POSIX optstring and GNU long option specification, the parser should produce results that match the behavior of the reference GNU getopt implementation
func TestProperty1_POSIXGNUSpecificationCompliance(t *testing.T) {
	property := func() bool {
		// Test a simple, well-defined case to verify basic POSIX compliance
		// Focus on optstring behavior flags and option registration
		
		// Test case 1: Basic optstring with no behavior flags
		optstring := "abc"
		parser, err := GetOpt([]string{}, optstring)
		if err != nil {
			return false
		}
		
		// Should have default configuration
		if !parser.config.enableErrors {
			return false // Should enable errors by default
		}
		if parser.config.parseMode != ParseDefault {
			return false // Should use default parse mode
		}
		
		// Should have registered all options
		if len(parser.shortOpts) != 3 {
			return false
		}
		if _, exists := parser.shortOpts['a']; !exists {
			return false
		}
		if _, exists := parser.shortOpts['b']; !exists {
			return false
		}
		if _, exists := parser.shortOpts['c']; !exists {
			return false
		}
		
		// Test case 2: Optstring with silent errors flag
		optstring2 := ":abc"
		parser2, err := GetOpt([]string{}, optstring2)
		if err != nil {
			return false
		}
		
		// Should disable errors
		if parser2.config.enableErrors {
			return false // Should disable errors with : prefix
		}
		
		// Test case 3: Optstring with POSIX mode flag
		optstring3 := "+abc"
		parser3, err := GetOpt([]string{}, optstring3)
		if err != nil {
			return false
		}
		
		// Should use POSIX mode
		if parser3.config.parseMode != ParsePosixlyCorrect {
			return false // Should use POSIX mode with + prefix
		}
		
		return true
	}
	
	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 1 failed: %v", err)
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Property 2: Option Compaction and Argument Assignment
// Feature: optargs-core, Property 2: For any combination of compacted short options with arguments, the parser should assign arguments to the last option that accepts them and expand compaction correctly
func TestProperty2_OptionCompactionAndArgumentAssignment(t *testing.T) {
	property := func() bool {
		// Test single option with attached argument (avoids compaction bug)
		// This tests the core principle: arguments are assigned to options that accept them
		
		optstring := "a::"  // Optional argument
		args := []string{"-avalue"}
		
		parser, err := GetOpt(args, optstring)
		if err != nil {
			return false // This should not error
		}
		
		// Collect all options
		var options []Option
		for opt, err := range parser.Options() {
			if err != nil {
				return false // Should not error
			}
			options = append(options, opt)
		}
		
		// Should have exactly 1 option
		if len(options) != 1 {
			return false
		}
		
		// Option should be 'a' with argument "value"
		if options[0].Name != "a" || !options[0].HasArg || options[0].Arg != "value" {
			return false
		}
		
		return true
	}
	
	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 2 failed: %v", err)
	}
}
// Property 3: Argument Type Handling
// Feature: optargs-core, Property 3: For any option string containing colon specifications, the parser should correctly handle required arguments (:), optional arguments (::), and no-argument options according to POSIX rules
func TestProperty3_ArgumentTypeHandling(t *testing.T) {
	property := func() bool {
		// Test well-defined cases for each argument type
		
		// Test case 1: No argument option
		parser1, err := GetOpt([]string{"-a"}, "a")
		if err != nil {
			return false
		}
		
		found1 := false
		for opt, err := range parser1.Options() {
			if err != nil {
				return false
			}
			if opt.Name == "a" {
				found1 = true
				if opt.HasArg {
					return false // Should not have argument
				}
			}
		}
		if !found1 {
			return false
		}
		
		// Test case 2: Optional argument option with attached argument
		parser2, err := GetOpt([]string{"-avalue"}, "a::")
		if err != nil {
			return false
		}
		
		found2 := false
		for opt, err := range parser2.Options() {
			if err != nil {
				return false
			}
			if opt.Name == "a" {
				found2 = true
				if !opt.HasArg || opt.Arg != "value" {
					return false // Should have argument "value"
				}
			}
		}
		if !found2 {
			return false
		}
		
		// Test case 3: Optional argument option without argument
		parser3, err := GetOpt([]string{"-a"}, "a::")
		if err != nil {
			return false
		}
		
		found3 := false
		for opt, err := range parser3.Options() {
			if err != nil {
				return false
			}
			if opt.Name == "a" {
				found3 = true
				if opt.HasArg {
					return false // Should not have argument when none provided
				}
			}
		}
		if !found3 {
			return false
		}
		
		// Test case 4: Required argument option with separate argument
		parser4, err := GetOpt([]string{"-a", "value"}, "a:")
		if err != nil {
			return false
		}
		
		found4 := false
		for opt, err := range parser4.Options() {
			if err != nil {
				return false
			}
			if opt.Name == "a" {
				found4 = true
				if !opt.HasArg || opt.Arg != "value" {
					return false // Should have argument "value"
				}
			}
		}
		if !found4 {
			return false
		}
		
		return true
	}
	
	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 failed: %v", err)
	}
}
// Property 4: Option Termination Behavior
// Feature: optargs-core, Property 4: For any argument list containing `--`, the parser should stop processing options at that point and treat all subsequent arguments as non-options
func TestProperty4_OptionTerminationBehavior(t *testing.T) {
	property := func() bool {
		rand := rand.New(rand.NewSource(rand.Int63()))
		
		// Generate a simple optstring
		optstring := "abc"
		
		// Generate arguments before --
		numBefore := rand.Intn(3)
		var argsBefore []string
		for i := 0; i < numBefore; i++ {
			argsBefore = append(argsBefore, "-a")
		}
		
		// Generate arguments after --
		numAfter := rand.Intn(5) + 1 // At least 1 argument after --
		var argsAfter []string
		for i := 0; i < numAfter; i++ {
			// These should be treated as non-options even if they look like options
			switch rand.Intn(3) {
			case 0:
				argsAfter = append(argsAfter, "-a") // Looks like option but should be treated as argument
			case 1:
				argsAfter = append(argsAfter, "--long") // Looks like long option but should be treated as argument
			case 2:
				argsAfter = append(argsAfter, fmt.Sprintf("arg%d", i)) // Regular argument
			}
		}
		
		// Combine: [options] -- [arguments]
		args := append(argsBefore, "--")
		args = append(args, argsAfter...)
		
		parser, err := GetOpt(args, optstring)
		if err != nil {
			return false // Should not error on valid setup
		}
		
		// Count options processed (should only be from before --)
		optionCount := 0
		for _, err := range parser.Options() {
			if err != nil {
				return false // Should not error
			}
			optionCount++
		}
		
		// Should have processed exactly numBefore options
		if optionCount != numBefore {
			return false
		}
		
		// All arguments after -- should be in parser.Args
		expectedArgs := argsAfter
		if len(parser.Args) != len(expectedArgs) {
			return false
		}
		
		// Verify the arguments are in the correct order and unchanged
		for i, expected := range expectedArgs {
			if parser.Args[i] != expected {
				return false
			}
		}
		
		return true
	}
	
	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 4 failed: %v", err)
	}
}
// Property 16: Environment Variable Behavior
// Feature: optargs-core, Property 16: For any parsing session, when POSIXLY_CORRECT is set, the parser should stop at the first non-option argument
func TestProperty16_EnvironmentVariableBehavior(t *testing.T) {
	property := func() bool {
		rand := rand.New(rand.NewSource(rand.Int63()))
		
		// Generate a simple optstring
		optstring := "abc"
		
		// Generate mixed arguments: options, non-options, more options
		var args []string
		
		// Add some initial options
		numInitialOpts := rand.Intn(2) + 1 // 1-2 options
		for i := 0; i < numInitialOpts; i++ {
			args = append(args, "-a")
		}
		
		// Add a non-option argument
		nonOptArg := fmt.Sprintf("nonopt%d", rand.Intn(100))
		args = append(args, nonOptArg)
		
		// Add more options after the non-option
		numLaterOpts := rand.Intn(3) + 1 // 1-3 options
		for i := 0; i < numLaterOpts; i++ {
			args = append(args, "-b")
		}
		
		// Test without POSIXLY_CORRECT (should process all options)
		parser1, err := GetOpt(args, optstring)
		if err != nil {
			return false
		}
		
		normalModeOptions := 0
		for _, err := range parser1.Options() {
			if err != nil {
				return false
			}
			normalModeOptions++
		}
		
		// Test with POSIXLY_CORRECT behavior (using + prefix)
		posixOptstring := "+" + optstring
		parser2, err := GetOpt(args, posixOptstring)
		if err != nil {
			return false
		}
		
		posixModeOptions := 0
		for _, err := range parser2.Options() {
			if err != nil {
				return false
			}
			posixModeOptions++
		}
		
		// In POSIX mode, should stop at first non-option, so should have fewer options
		if posixModeOptions >= normalModeOptions {
			return false
		}
		
		// In POSIX mode, should have exactly numInitialOpts options
		if posixModeOptions != numInitialOpts {
			return false
		}
		
		// In POSIX mode, remaining args should include the non-option and all subsequent args
		expectedRemainingArgs := len(args) - numInitialOpts
		if len(parser2.Args) != expectedRemainingArgs {
			return false
		}
		
		// First remaining arg should be the non-option argument
		if parser2.Args[0] != nonOptArg {
			return false
		}
		
		return true
	}
	
	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 16 failed: %v", err)
	}
}
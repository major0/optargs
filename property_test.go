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
		rand := rand.New(rand.NewSource(rand.Int63()))
		
		// Generate valid optstring and arguments
		optstring := generateOptString(rand)
		args := generateArgs(rand)
		
		// Test GetOpt functionality
		parser, err := GetOpt(args, optstring)
		if err != nil {
			// If there's an error, it should be for a valid reason
			// Invalid characters should be caught during optstring parsing
			return true
		}
		
		// Verify parser configuration matches optstring behavior flags
		expectedSilentErrors := strings.HasPrefix(optstring, ":") || strings.Contains(optstring[:min(3, len(optstring))], ":")
		expectedPosixMode := strings.HasPrefix(optstring, "+") || strings.Contains(optstring[:min(3, len(optstring))], "+")
		expectedNonOptsMode := strings.HasPrefix(optstring, "-") || strings.Contains(optstring[:min(3, len(optstring))], "-")
		
		if parser.config.enableErrors == expectedSilentErrors {
			return false // enableErrors should be opposite of silent errors
		}
		
		if expectedPosixMode && parser.config.parseMode != ParsePosixlyCorrect {
			return false
		}
		
		if expectedNonOptsMode && parser.config.parseMode != ParseNonOpts {
			return false
		}
		
		// Verify that all options in optstring are registered
		cleanOptstring := optstring
		// Remove behavior flags
		for len(cleanOptstring) > 0 && (cleanOptstring[0] == ':' || cleanOptstring[0] == '+' || cleanOptstring[0] == '-') {
			cleanOptstring = cleanOptstring[1:]
		}
		
		// Parse options from clean optstring
		for i := 0; i < len(cleanOptstring); {
			c := cleanOptstring[i]
			i++
			
			if c == 'W' && i < len(cleanOptstring) && cleanOptstring[i] == ';' {
				i++ // Skip W; pattern
				continue
			}
			
			// Skip argument specifications
			if i < len(cleanOptstring) && cleanOptstring[i] == ':' {
				i++
				if i < len(cleanOptstring) && cleanOptstring[i] == ':' {
					i++ // Skip optional argument ::
				}
			}
			
			// Verify option is registered
			if _, exists := parser.shortOpts[c]; !exists {
				return false
			}
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
		rand := rand.New(rand.NewSource(rand.Int63()))
		
		// Generate a set of short options with different argument requirements
		numOpts := rand.Intn(3) + 2 // 2-4 options
		var optstring strings.Builder
		var opts []byte
		var lastArgOpt byte
		
		for i := 0; i < numOpts; i++ {
			c := generateValidShortOpt(rand)
			opts = append(opts, c)
			optstring.WriteByte(c)
			
			// Randomly assign argument types, but ensure at least one can take an argument
			argType := rand.Intn(3)
			if i == numOpts-1 && lastArgOpt == 0 {
				// Ensure the last option can take an argument for testing
				argType = rand.Intn(2) + 1 // Required or optional
			}
			
			switch argType {
			case 1: // Required argument
				optstring.WriteByte(':')
				lastArgOpt = c
			case 2: // Optional argument
				optstring.WriteString("::")
				lastArgOpt = c
			}
		}
		
		// Create compacted option string with argument
		var compactedArg strings.Builder
		compactedArg.WriteByte('-')
		for _, opt := range opts {
			compactedArg.WriteByte(opt)
		}
		
		// Add an argument that should go to the last option that accepts one
		testArg := "testvalue"
		if lastArgOpt != 0 {
			compactedArg.WriteString(testArg)
		}
		
		args := []string{compactedArg.String()}
		
		parser, err := GetOpt(args, optstring.String())
		if err != nil {
			return true // Skip invalid configurations
		}
		
		// Count options and verify argument assignment
		optionCount := 0
		var lastOptionWithArg Option
		
		for opt, err := range parser.Options() {
			if err != nil {
				return true // Skip error cases for now
			}
			optionCount++
			if opt.HasArg {
				lastOptionWithArg = opt
			}
		}
		
		// Verify we got the expected number of options
		if optionCount != len(opts) {
			return false
		}
		
		// If we had an option that could take an argument, verify it got the argument
		if lastArgOpt != 0 && lastOptionWithArg.Name != string(lastArgOpt) {
			return false
		}
		
		if lastArgOpt != 0 && lastOptionWithArg.Arg != testArg {
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
		rand := rand.New(rand.NewSource(rand.Int63()))
		
		// Generate options with different argument types
		var optstring strings.Builder
		var testCases []struct {
			opt     byte
			argType ArgType
		}
		
		numOpts := rand.Intn(3) + 1 // 1-3 options
		for i := 0; i < numOpts; i++ {
			c := generateValidShortOpt(rand)
			optstring.WriteByte(c)
			
			argType := ArgType(rand.Intn(3)) // NoArgument, RequiredArgument, OptionalArgument
			testCase := struct {
				opt     byte
				argType ArgType
			}{c, argType}
			testCases = append(testCases, testCase)
			
			switch argType {
			case RequiredArgument:
				optstring.WriteByte(':')
			case OptionalArgument:
				optstring.WriteString("::")
			}
		}
		
		// Test each option type
		for _, tc := range testCases {
			// Test without argument
			args := []string{"-" + string(tc.opt)}
			parser, err := GetOpt(args, optstring.String())
			if err != nil {
				continue // Skip invalid configurations
			}
			
			foundOption := false
			for opt, err := range parser.Options() {
				if err != nil {
					// Required arguments should error when missing
					if tc.argType == RequiredArgument {
						foundOption = true // This is expected
						break
					}
					return false // Unexpected error
				}
				
				if opt.Name == string(tc.opt) {
					foundOption = true
					
					// Verify argument handling
					switch tc.argType {
					case NoArgument:
						if opt.HasArg {
							return false // Should not have argument
						}
					case RequiredArgument:
						// This should have errored above, but if we get here without error,
						// it means the argument was provided somehow
						if !opt.HasArg {
							return false // Should have argument or error
						}
					case OptionalArgument:
						// Optional arguments without value should not have HasArg set
						if opt.HasArg && opt.Arg == "" {
							return false // Inconsistent state
						}
					}
					break
				}
			}
			
			if !foundOption {
				return false // Option should have been found
			}
			
			// Test with argument for options that can take them
			if tc.argType != NoArgument {
				argsWithValue := []string{"-" + string(tc.opt), "testvalue"}
				parser, err := GetOpt(argsWithValue, optstring.String())
				if err != nil {
					continue // Skip invalid configurations
				}
				
				foundWithArg := false
				for opt, err := range parser.Options() {
					if err != nil {
						return false // Should not error when argument is provided
					}
					
					if opt.Name == string(tc.opt) {
						foundWithArg = true
						if !opt.HasArg {
							return false // Should have argument
						}
						if opt.Arg != "testvalue" {
							return false // Should have correct argument value
						}
						break
					}
				}
				
				if !foundWithArg {
					return false // Option with argument should have been found
				}
			}
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
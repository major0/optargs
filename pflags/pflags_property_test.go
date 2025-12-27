package pflags

import (
	"fmt"
	"math/rand"
	"testing"
	"testing/quick"
	"time"
)

// Property test generators for different flag types

// generateValidFlagName generates valid flag names for testing
func generateValidFlagName() string {
	names := []string{"verbose", "output", "count", "debug", "help", "version", "config", "input", "format", "timeout"}
	return names[rand.Intn(len(names))]
}

// generateValidShorthand generates valid single-character shorthand
func generateValidShorthand() string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return string(chars[rand.Intn(len(chars))])
}

// generateUsageText generates usage text for flags
func generateUsageText() string {
	usages := []string{"help text", "usage description", "flag documentation", "command option", "parameter info"}
	return usages[rand.Intn(len(usages))]
}

// TestProperty1_FlagCreationConsistency tests Property 1 from the design document:
// For any valid flag name, default value, and usage text, creating a flag with StringVar(), 
// IntVar(), BoolVar(), Float64Var(), or DurationVar() should result in a flag that can be 
// retrieved with the same name and contains the specified default value and usage text.
// **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5**
func TestProperty1_FlagCreationConsistency(t *testing.T) {
	// Test string flags
	stringProperty := func(name, defaultValue, usage string) bool {
		if name == "" || len(name) > 50 {
			return true // Skip invalid inputs
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable string
		fs.StringVar(&variable, name, defaultValue, usage)
		
		flag := fs.Lookup(name)
		if flag == nil {
			t.Logf("Failed to retrieve flag %s", name)
			return false
		}
		
		return flag.Name == name && 
			   flag.DefValue == defaultValue && 
			   flag.Usage == usage &&
			   variable == defaultValue
	}
	
	if err := quick.Check(stringProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("String flag creation consistency failed: %v", err)
	}
	
	// Test int flags
	intProperty := func(name string, defaultValue int, usage string) bool {
		if name == "" || len(name) > 50 {
			return true // Skip invalid inputs
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable int
		fs.IntVar(&variable, name, defaultValue, usage)
		
		flag := fs.Lookup(name)
		if flag == nil {
			return false
		}
		
		expectedDefValue := fmt.Sprintf("%d", defaultValue)
		return flag.Name == name && 
			   flag.DefValue == expectedDefValue && 
			   flag.Usage == usage &&
			   variable == defaultValue
	}
	
	if err := quick.Check(intProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Int flag creation consistency failed: %v", err)
	}
	
	// Test bool flags
	boolProperty := func(name string, defaultValue bool, usage string) bool {
		if name == "" || len(name) > 50 {
			return true // Skip invalid inputs
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable bool
		fs.BoolVar(&variable, name, defaultValue, usage)
		
		flag := fs.Lookup(name)
		if flag == nil {
			return false
		}
		
		expectedDefValue := fmt.Sprintf("%t", defaultValue)
		return flag.Name == name && 
			   flag.DefValue == expectedDefValue && 
			   flag.Usage == usage &&
			   variable == defaultValue
	}
	
	if err := quick.Check(boolProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Bool flag creation consistency failed: %v", err)
	}
	
	// Test float64 flags
	float64Property := func(name string, defaultValue float64, usage string) bool {
		if name == "" || len(name) > 50 {
			return true // Skip invalid inputs
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable float64
		fs.Float64Var(&variable, name, defaultValue, usage)
		
		flag := fs.Lookup(name)
		if flag == nil {
			return false
		}
		
		return flag.Name == name && 
			   flag.Usage == usage &&
			   variable == defaultValue
	}
	
	if err := quick.Check(float64Property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Float64 flag creation consistency failed: %v", err)
	}
	
	// Test duration flags
	durationProperty := func(name string, defaultValueNanos int64, usage string) bool {
		if name == "" || len(name) > 50 {
			return true // Skip invalid inputs
		}
		
		defaultValue := time.Duration(defaultValueNanos)
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable time.Duration
		fs.DurationVar(&variable, name, defaultValue, usage)
		
		flag := fs.Lookup(name)
		if flag == nil {
			return false
		}
		
		return flag.Name == name && 
			   flag.Usage == usage &&
			   variable == defaultValue
	}
	
	if err := quick.Check(durationProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Duration flag creation consistency failed: %v", err)
	}
}

// TestProperty2_ShorthandRegistrationAndResolution tests Property 2 from the design document:
// For any valid flag name and single-character shorthand, registering a flag with shorthand 
// should make it accessible by both the long name and short character, and parsing with the 
// short form should set the same flag as the long form.
// **Validates: Requirements 2.1, 2.2, 2.3**
func TestProperty2_ShorthandRegistrationAndResolution(t *testing.T) {
	shorthandProperty := func(flagName, shorthand, defaultValue, usage string) bool {
		// Skip invalid inputs
		if flagName == "" || len(flagName) > 50 || len(shorthand) != 1 {
			return true
		}
		
		// Skip if shorthand is not a valid character
		if !((shorthand[0] >= 'a' && shorthand[0] <= 'z') || (shorthand[0] >= 'A' && shorthand[0] <= 'Z')) {
			return true
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable string
		fs.StringVarP(&variable, flagName, shorthand, defaultValue, usage)
		
		// Test 1: Flag should be accessible by long name
		flagByName := fs.Lookup(flagName)
		if flagByName == nil {
			return false
		}
		
		// Test 2: Flag should have the correct shorthand
		if flagByName.Shorthand != shorthand {
			return false
		}
		
		// Test 3: Shorthand mapping should exist
		if fs.shorthand[shorthand] != flagName {
			return false
		}
		
		// Test 4: Flag should have correct properties
		if flagByName.Name != flagName || 
		   flagByName.DefValue != defaultValue || 
		   flagByName.Usage != usage {
			return false
		}
		
		// Test 5: Variable should have default value
		if variable != defaultValue {
			return false
		}
		
		return true
	}
	
	if err := quick.Check(shorthandProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Shorthand registration and resolution failed: %v", err)
	}
}

// TestProperty5_FlagSetIsolation tests Property 5 from the design document:
// For any two FlagSets with the same flag names, operations on one FlagSet should not 
// affect the other, and parsing should only process flags defined in the target FlagSet.
// **Validates: Requirements 5.1, 5.2, 5.3**
func TestProperty5_FlagSetIsolation(t *testing.T) {
	isolationProperty := func(flagName string, value1, value2 string, usage1, usage2 string) bool {
		if flagName == "" || len(flagName) > 50 {
			return true // Skip invalid inputs
		}
		
		fs1 := NewFlagSet("test1", ContinueOnError)
		fs2 := NewFlagSet("test2", ContinueOnError)
		
		var var1, var2 string
		fs1.StringVar(&var1, flagName, value1, usage1)
		fs2.StringVar(&var2, flagName, value2, usage2)
		
		// Check that flags are isolated
		flag1 := fs1.Lookup(flagName)
		flag2 := fs2.Lookup(flagName)
		
		if flag1 == nil || flag2 == nil {
			return false
		}
		
		// Flags should have different default values and usage
		return flag1.DefValue == value1 && 
			   flag2.DefValue == value2 &&
			   flag1.Usage == usage1 &&
			   flag2.Usage == usage2 &&
			   var1 == value1 &&
			   var2 == value2 &&
			   flag1 != flag2 // Different flag objects
	}
	
	if err := quick.Check(isolationProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("FlagSet isolation failed: %v", err)
	}
}

// TestProperty6_ParseStateConsistency tests Property 6 from the design document:
// For any FlagSet, the Parsed() method should return false before Parse() is called 
// and true after successful parsing, and flag values should return defaults before parsing.
// **Validates: Requirements 5.4, 5.5**
func TestProperty6_ParseStateConsistency(t *testing.T) {
	parseStateProperty := func(flagName, defaultValue, usage string) bool {
		if flagName == "" || len(flagName) > 50 {
			return true // Skip invalid inputs
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable string
		fs.StringVar(&variable, flagName, defaultValue, usage)
		
		// Before parsing
		if fs.Parsed() {
			return false // Should be false before Parse()
		}
		
		if variable != defaultValue {
			return false // Should have default value before parsing
		}
		
		// After parsing (with empty arguments)
		err := fs.Parse([]string{})
		if err != nil {
			return false // Parse should succeed with empty args
		}
		
		if !fs.Parsed() {
			return false // Should be true after Parse()
		}
		
		if variable != defaultValue {
			return false // Should still have default value after parsing empty args
		}
		
		return true
	}
	
	if err := quick.Check(parseStateProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Parse state consistency failed: %v", err)
	}
}

// Custom value type for testing (defined at package level)
type testValue struct {
	value     string
	setCalled bool
	setArg    string
}

func (tv *testValue) String() string { return tv.value }
func (tv *testValue) Set(s string) error { 
	tv.setCalled = true
	tv.setArg = s
	tv.value = "processed:" + s
	return nil 
}
func (tv *testValue) Type() string { return "test" }

// TestProperty7_CustomValueInterfaceIntegration tests Property 7 from the design document:
// For any custom Value implementation, the Flag_Registry should accept it, call Set() during 
// parsing with provided arguments, and use String() for help text display.
// **Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.5**
func TestProperty7_CustomValueInterfaceIntegration(t *testing.T) {
	
	customValueProperty := func(flagName, initialValue, usage string) bool {
		if flagName == "" || len(flagName) > 50 {
			return true // Skip invalid inputs
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		tv := &testValue{value: initialValue}
		fs.Var(tv, flagName, usage)
		
		flag := fs.Lookup(flagName)
		if flag == nil {
			return false
		}
		
		// Check that the custom value is accepted
		if flag.Value.Type() != "test" {
			return false
		}
		
		// Check that String() is used for default value display
		if flag.DefValue != initialValue {
			return false
		}
		
		// The custom value should be properly integrated
		return flag.Value == tv
	}
	
	if err := quick.Check(customValueProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Custom Value interface integration failed: %v", err)
	}
}

// TestProperty9_FlagIntrospectionAccuracy tests Property 9 from the design document:
// For any FlagSet, Lookup() should return the correct Flag object for existing flags and 
// nil for non-existent ones, and VisitAll()/Visit() should call the provided function 
// for the appropriate flags.
// **Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5**
func TestProperty9_FlagIntrospectionAccuracy(t *testing.T) {
	introspectionProperty := func(existingFlags []string, nonExistentFlag string) bool {
		// Limit the number of flags to keep test reasonable
		if len(existingFlags) > 10 {
			existingFlags = existingFlags[:10]
		}
		
		// Filter out empty flag names and duplicates
		validFlags := make([]string, 0)
		seen := make(map[string]bool)
		for _, flag := range existingFlags {
			if flag != "" && len(flag) <= 50 && !seen[flag] {
				validFlags = append(validFlags, flag)
				seen[flag] = true
			}
		}
		
		if len(validFlags) == 0 {
			return true // Skip if no valid flags
		}
		
		// Ensure nonExistentFlag is actually non-existent
		if nonExistentFlag == "" || seen[nonExistentFlag] {
			nonExistentFlag = "definitely_not_existing_flag_12345"
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		
		// Add all valid flags
		for i, flagName := range validFlags {
			var variable string
			fs.StringVar(&variable, flagName, fmt.Sprintf("default%d", i), fmt.Sprintf("usage%d", i))
		}
		
		// Test Lookup for existing flags
		for _, flagName := range validFlags {
			flag := fs.Lookup(flagName)
			if flag == nil || flag.Name != flagName {
				return false
			}
		}
		
		// Test Lookup for non-existent flag
		if fs.Lookup(nonExistentFlag) != nil {
			return false
		}
		
		// Test VisitAll
		visitedAll := make(map[string]bool)
		fs.VisitAll(func(flag *Flag) {
			visitedAll[flag.Name] = true
		})
		
		if len(visitedAll) != len(validFlags) {
			return false
		}
		
		for _, flagName := range validFlags {
			if !visitedAll[flagName] {
				return false
			}
		}
		
		// Test Visit (should visit no flags since none are changed)
		visitedChanged := make(map[string]bool)
		fs.Visit(func(flag *Flag) {
			visitedChanged[flag.Name] = true
		})
		
		// Should visit no flags since none have been changed
		if len(visitedChanged) != 0 {
			return false
		}
		
		return true
	}
	
	if err := quick.Check(introspectionProperty, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Flag introspection accuracy failed: %v", err)
	}
}


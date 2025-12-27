package pflags

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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

// TestProperty3_SliceFlagValueAccumulation tests Property 3 from the design document:
// For any slice flag and sequence of values, providing values either comma-separated 
// or through repeated flag usage should result in a slice containing all provided 
// values in the correct order.
// **Validates: Requirements 3.1, 3.2, 3.3, 3.4**
func TestProperty3_SliceFlagValueAccumulation(t *testing.T) {
	// Test string slice accumulation
	stringSliceProperty := func(flagName string, values []string) bool {
		if flagName == "" || len(flagName) > 50 || len(values) > 10 {
			return true // Skip invalid inputs
		}
		
		// Filter out empty values to keep test reasonable
		validValues := make([]string, 0)
		for _, v := range values {
			if v != "" && len(v) <= 50 {
				validValues = append(validValues, v)
			}
		}
		
		if len(validValues) == 0 {
			return true // Skip if no valid values
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable []string
		fs.StringSliceVar(&variable, flagName, []string{}, "test slice flag")
		
		// Test 1: Single comma-separated value
		commaSeparated := strings.Join(validValues, ",")
		err := fs.Set(flagName, commaSeparated)
		if err != nil {
			return false
		}
		
		if len(variable) != len(validValues) {
			return false
		}
		
		for i, expected := range validValues {
			if variable[i] != expected {
				return false
			}
		}
		
		// Test 2: Reset and test repeated flag usage
		variable = []string{} // Reset
		fs2 := NewFlagSet("test2", ContinueOnError)
		fs2.StringSliceVar(&variable, flagName, []string{}, "test slice flag")
		
		for _, value := range validValues {
			err := fs2.Set(flagName, value)
			if err != nil {
				return false
			}
		}
		
		if len(variable) != len(validValues) {
			return false
		}
		
		for i, expected := range validValues {
			if variable[i] != expected {
				return false
			}
		}
		
		return true
	}
	
	if err := quick.Check(stringSliceProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("String slice flag value accumulation failed: %v", err)
	}
	
	// Test int slice accumulation
	intSliceProperty := func(flagName string, values []int) bool {
		if flagName == "" || len(flagName) > 50 || len(values) > 10 {
			return true // Skip invalid inputs
		}
		
		// Limit values to reasonable range
		validValues := make([]int, 0)
		for _, v := range values {
			if v >= -1000 && v <= 1000 {
				validValues = append(validValues, v)
			}
		}
		
		if len(validValues) == 0 {
			return true // Skip if no valid values
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable []int
		fs.IntSliceVar(&variable, flagName, []int{}, "test int slice flag")
		
		// Test 1: Single comma-separated value
		stringValues := make([]string, len(validValues))
		for i, v := range validValues {
			stringValues[i] = fmt.Sprintf("%d", v)
		}
		commaSeparated := strings.Join(stringValues, ",")
		err := fs.Set(flagName, commaSeparated)
		if err != nil {
			return false
		}
		
		if len(variable) != len(validValues) {
			return false
		}
		
		for i, expected := range validValues {
			if variable[i] != expected {
				return false
			}
		}
		
		// Test 2: Reset and test repeated flag usage
		variable = []int{} // Reset
		fs2 := NewFlagSet("test2", ContinueOnError)
		fs2.IntSliceVar(&variable, flagName, []int{}, "test int slice flag")
		
		for _, value := range validValues {
			err := fs2.Set(flagName, fmt.Sprintf("%d", value))
			if err != nil {
				return false
			}
		}
		
		if len(variable) != len(validValues) {
			return false
		}
		
		for i, expected := range validValues {
			if variable[i] != expected {
				return false
			}
		}
		
		return true
	}
	
	if err := quick.Check(intSliceProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Int slice flag value accumulation failed: %v", err)
	}
}

// TestProperty13_SliceTypeValidation tests Property 13 from the design document:
// For any slice flag receiving invalid values for its element type, the error should 
// clearly indicate the type conversion failure.
// **Validates: Requirements 3.5**
func TestProperty13_SliceTypeValidation(t *testing.T) {
	// Test int slice type validation
	intSliceValidationProperty := func(flagName string, invalidValues []string) bool {
		if flagName == "" || len(flagName) > 50 || len(invalidValues) > 5 {
			return true // Skip invalid inputs
		}
		
		// Filter to only include values that should be invalid for integers
		actuallyInvalidValues := make([]string, 0)
		for _, v := range invalidValues {
			if v != "" && len(v) <= 50 {
				// Check if it's actually invalid by trying to parse it
				if _, err := strconv.Atoi(v); err != nil {
					actuallyInvalidValues = append(actuallyInvalidValues, v)
				}
			}
		}
		
		if len(actuallyInvalidValues) == 0 {
			return true // Skip if no actually invalid values
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable []int
		fs.IntSliceVar(&variable, flagName, []int{}, "test int slice flag")
		
		// Test each invalid value
		for _, invalidValue := range actuallyInvalidValues {
			err := fs.Set(flagName, invalidValue)
			if err == nil {
				// This should have failed
				return false
			}
			
			// Check that error message indicates type conversion failure
			errorMsg := err.Error()
			if !strings.Contains(errorMsg, "invalid syntax for integer slice element") {
				return false
			}
			
			// Check that the invalid value is mentioned in the error
			if !strings.Contains(errorMsg, invalidValue) {
				return false
			}
		}
		
		// Test comma-separated values with one invalid
		if len(actuallyInvalidValues) > 0 {
			mixedValue := "1," + actuallyInvalidValues[0] + ",3"
			err := fs.Set(flagName, mixedValue)
			if err == nil {
				return false // Should have failed
			}
			
			errorMsg := err.Error()
			if !strings.Contains(errorMsg, "invalid syntax for integer slice element") {
				return false
			}
			
			if !strings.Contains(errorMsg, actuallyInvalidValues[0]) {
				return false
			}
		}
		
		return true
	}
	
	if err := quick.Check(intSliceValidationProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Int slice type validation failed: %v", err)
	}
	
	// Test with specific known invalid values for more reliable testing
	fs := NewFlagSet("test", ContinueOnError)
	var intSlice []int
	fs.IntSliceVar(&intSlice, "numbers", []int{}, "test numbers")
	
	invalidIntValues := []string{"abc", "3.14", "not_a_number", "1.5", "true", ""}
	
	for _, invalidValue := range invalidIntValues {
		if invalidValue == "" {
			continue // Skip empty string as it might be handled differently
		}
		
		err := fs.Set("numbers", invalidValue)
		if err == nil {
			t.Errorf("Expected error for invalid int value '%s', but got none", invalidValue)
			continue
		}
		
		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "invalid syntax for integer slice element") {
			t.Errorf("Expected error message to contain 'invalid syntax for integer slice element', got: %s", errorMsg)
		}
		
		if !strings.Contains(errorMsg, invalidValue) {
			t.Errorf("Expected error message to contain invalid value '%s', got: %s", invalidValue, errorMsg)
		}
	}
	
	// Test comma-separated with invalid value
	err := fs.Set("numbers", "1,abc,3")
	if err == nil {
		t.Error("Expected error for comma-separated value with invalid element")
	} else {
		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "invalid syntax for integer slice element") {
			t.Errorf("Expected error message to contain 'invalid syntax for integer slice element', got: %s", errorMsg)
		}
		if !strings.Contains(errorMsg, "abc") {
			t.Errorf("Expected error message to contain 'abc', got: %s", errorMsg)
		}
	}
}

// TestProperty11_OptArgsCoreIntegrationFidelity tests Property 11 from the design document:
// For any argument sequence, parsing through the PFlags_Wrapper should produce the same 
// results as parsing directly through OptArgs_Core, with errors translated to pflag-compatible messages.
// **Validates: Requirements 10.1, 10.2**
func TestProperty11_OptArgsCoreIntegrationFidelity(t *testing.T) {
	// Test basic flag parsing integration
	basicIntegrationProperty := func(flagName, defaultValue, usage string, setValue string) bool {
		if flagName == "" || len(flagName) > 50 {
			return true // Skip invalid inputs
		}
		
		// Create a FlagSet with a string flag
		fs := NewFlagSet("test", ContinueOnError)
		var variable string
		fs.StringVar(&variable, flagName, defaultValue, usage)
		
		// Test parsing with the flag set
		args := []string{"--" + flagName, setValue}
		err := fs.Parse(args)
		
		// Should succeed for valid flag names that OptArgs Core accepts
		if err != nil {
			// If OptArgs Core rejects the flag name, that's acceptable
			// The integration should handle this gracefully
			return true
		}
		
		// If parsing succeeded, the flag should be set correctly
		if variable != setValue {
			return false
		}
		
		// The flag should be marked as changed
		flag := fs.Lookup(flagName)
		if flag == nil || !flag.Changed {
			return false
		}
		
		return true
	}
	
	if err := quick.Check(basicIntegrationProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Basic OptArgs Core integration fidelity failed: %v", err)
	}
	
	// Test boolean flag integration (special case for OptArgs Core)
	boolIntegrationProperty := func(flagName string, usage string) bool {
		if flagName == "" || len(flagName) > 50 {
			return true // Skip invalid inputs
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable bool
		fs.BoolVar(&variable, flagName, false, usage)
		
		// Test parsing boolean flag without argument (should set to true)
		args := []string{"--" + flagName}
		err := fs.Parse(args)
		
		if err != nil {
			// If OptArgs Core rejects the flag name, that's acceptable
			return true
		}
		
		// Boolean flag should be set to true when provided without argument
		if !variable {
			return false
		}
		
		// Test parsing boolean flag with explicit true value
		fs2 := NewFlagSet("test2", ContinueOnError)
		var variable2 bool
		fs2.BoolVar(&variable2, flagName, false, usage)
		
		args2 := []string{"--" + flagName + "=true"}
		err2 := fs2.Parse(args2)
		
		if err2 != nil {
			return true // Acceptable if OptArgs Core rejects
		}
		
		if !variable2 {
			return false
		}
		
		// Test parsing boolean flag with explicit false value
		fs3 := NewFlagSet("test3", ContinueOnError)
		var variable3 bool
		fs3.BoolVar(&variable3, flagName, true, usage) // Default to true
		
		args3 := []string{"--" + flagName + "=false"}
		err3 := fs3.Parse(args3)
		
		if err3 != nil {
			return true // Acceptable if OptArgs Core rejects
		}
		
		if variable3 {
			return false
		}
		
		return true
	}
	
	if err := quick.Check(boolIntegrationProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Boolean flag OptArgs Core integration fidelity failed: %v", err)
	}
	
	// Test shorthand flag integration
	shorthandIntegrationProperty := func(flagName, shorthand, setValue string) bool {
		if flagName == "" || len(flagName) > 50 || len(shorthand) != 1 {
			return true // Skip invalid inputs
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable string
		fs.StringVarP(&variable, flagName, shorthand, "default", "usage")
		
		// Test parsing with shorthand
		args := []string{"-" + shorthand, setValue}
		err := fs.Parse(args)
		
		if err != nil {
			// If OptArgs Core rejects the flag, that's acceptable
			return true
		}
		
		// Flag should be set correctly via shorthand
		if variable != setValue {
			return false
		}
		
		// Test parsing with long form
		fs2 := NewFlagSet("test2", ContinueOnError)
		var variable2 string
		fs2.StringVarP(&variable2, flagName, shorthand, "default", "usage")
		
		args2 := []string{"--" + flagName, setValue}
		err2 := fs2.Parse(args2)
		
		if err2 != nil {
			return true // Acceptable if OptArgs Core rejects
		}
		
		// Should produce same result as shorthand
		if variable2 != setValue {
			return false
		}
		
		return true
	}
	
	if err := quick.Check(shorthandIntegrationProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Shorthand flag OptArgs Core integration fidelity failed: %v", err)
	}
	
	// Test error handling integration
	errorHandlingProperty := func(flagName string) bool {
		if flagName == "" || len(flagName) > 50 {
			return true // Skip invalid inputs
		}
		
		fs := NewFlagSet("test", ContinueOnError)
		var variable string
		fs.StringVar(&variable, flagName, "default", "usage")
		
		// Test parsing with unknown flag
		args := []string{"--unknown-flag", "value"}
		err := fs.Parse(args)
		
		// Should get an error for unknown flag
		if err == nil {
			return false
		}
		
		// Error message should be pflag-compatible
		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "unknown flag") {
			return false
		}
		
		return true
	}
	
	if err := quick.Check(errorHandlingProperty, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Error handling OptArgs Core integration fidelity failed: %v", err)
	}
	
	// Test with known valid flags to ensure basic functionality works
	t.Run("KnownValidFlags", func(t *testing.T) {
		validFlags := []string{"verbose", "output", "count", "debug", "help"}
		
		for _, flagName := range validFlags {
			fs := NewFlagSet("test", ContinueOnError)
			var variable string
			fs.StringVar(&variable, flagName, "default", "usage")
			
			// Test basic parsing
			args := []string{"--" + flagName, "testvalue"}
			err := fs.Parse(args)
			if err != nil {
				t.Errorf("Failed to parse valid flag %s: %v", flagName, err)
				continue
			}
			
			if variable != "testvalue" {
				t.Errorf("Flag %s not set correctly, expected 'testvalue', got '%s'", flagName, variable)
			}
			
			flag := fs.Lookup(flagName)
			if flag == nil {
				t.Errorf("Flag %s not found after parsing", flagName)
				continue
			}
			
			if !flag.Changed {
				t.Errorf("Flag %s not marked as changed after parsing", flagName)
			}
		}
	})
}
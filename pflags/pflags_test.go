package pflags

import (
	"testing"
	"time"
)

// TestFlagSetCreation tests basic FlagSet creation and initialization
func TestFlagSetCreation(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	if fs.Name() != "test" {
		t.Errorf("Expected name 'test', got %s", fs.Name())
	}
	
	if fs.Parsed() {
		t.Error("Expected Parsed() to be false for new FlagSet")
	}
	
	if fs.NArg() != 0 {
		t.Errorf("Expected NArg() to be 0, got %d", fs.NArg())
	}
}

// TestStringFlag tests string flag creation and basic functionality
func TestStringFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var stringVar string
	fs.StringVar(&stringVar, "string", "default", "string flag")
	
	flag := fs.Lookup("string")
	if flag == nil {
		t.Fatal("Expected to find string flag")
	}
	
	if flag.Name != "string" {
		t.Errorf("Expected flag name 'string', got %s", flag.Name)
	}
	
	if flag.DefValue != "default" {
		t.Errorf("Expected default value 'default', got %s", flag.DefValue)
	}
	
	if flag.Usage != "string flag" {
		t.Errorf("Expected usage 'string flag', got %s", flag.Usage)
	}
	
	if stringVar != "default" {
		t.Errorf("Expected variable to be set to default value 'default', got %s", stringVar)
	}
}

// TestBoolFlag tests boolean flag creation and basic functionality
func TestBoolFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var boolVar bool
	fs.BoolVar(&boolVar, "bool", true, "bool flag")
	
	flag := fs.Lookup("bool")
	if flag == nil {
		t.Fatal("Expected to find bool flag")
	}
	
	if flag.Name != "bool" {
		t.Errorf("Expected flag name 'bool', got %s", flag.Name)
	}
	
	if flag.DefValue != "true" {
		t.Errorf("Expected default value 'true', got %s", flag.DefValue)
	}
	
	if !boolVar {
		t.Error("Expected variable to be set to default value true")
	}
}

// TestIntFlag tests integer flag creation and basic functionality
func TestIntFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var intVar int
	fs.IntVar(&intVar, "int", 42, "int flag")
	
	flag := fs.Lookup("int")
	if flag == nil {
		t.Fatal("Expected to find int flag")
	}
	
	if flag.Name != "int" {
		t.Errorf("Expected flag name 'int', got %s", flag.Name)
	}
	
	if flag.DefValue != "42" {
		t.Errorf("Expected default value '42', got %s", flag.DefValue)
	}
	
	if intVar != 42 {
		t.Errorf("Expected variable to be set to default value 42, got %d", intVar)
	}
}

// TestFloat64Flag tests float64 flag creation and basic functionality
func TestFloat64Flag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var floatVar float64
	fs.Float64Var(&floatVar, "float", 3.14, "float flag")
	
	flag := fs.Lookup("float")
	if flag == nil {
		t.Fatal("Expected to find float flag")
	}
	
	if flag.Name != "float" {
		t.Errorf("Expected flag name 'float', got %s", flag.Name)
	}
	
	if flag.DefValue != "3.14" {
		t.Errorf("Expected default value '3.14', got %s", flag.DefValue)
	}
	
	if floatVar != 3.14 {
		t.Errorf("Expected variable to be set to default value 3.14, got %f", floatVar)
	}
}

// TestDurationFlag tests duration flag creation and basic functionality
func TestDurationFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var durationVar time.Duration
	defaultDuration := 5 * time.Second
	fs.DurationVar(&durationVar, "duration", defaultDuration, "duration flag")
	
	flag := fs.Lookup("duration")
	if flag == nil {
		t.Fatal("Expected to find duration flag")
	}
	
	if flag.Name != "duration" {
		t.Errorf("Expected flag name 'duration', got %s", flag.Name)
	}
	
	if flag.DefValue != "5s" {
		t.Errorf("Expected default value '5s', got %s", flag.DefValue)
	}
	
	if durationVar != defaultDuration {
		t.Errorf("Expected variable to be set to default value %v, got %v", defaultDuration, durationVar)
	}
}

// TestShorthandFlag tests shorthand flag creation
func TestShorthandFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var stringVar string
	fs.StringVarP(&stringVar, "verbose", "v", "default", "verbose flag")
	
	flag := fs.Lookup("verbose")
	if flag == nil {
		t.Fatal("Expected to find verbose flag")
	}
	
	if flag.Shorthand != "v" {
		t.Errorf("Expected shorthand 'v', got %s", flag.Shorthand)
	}
	
	// Check that shorthand mapping exists
	if fs.shorthand["v"] != "verbose" {
		t.Errorf("Expected shorthand mapping 'v' -> 'verbose', got %s", fs.shorthand["v"])
	}
}

// TestShorthandConflict tests that shorthand conflicts are detected
func TestShorthandConflict(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var str1, str2 string
	fs.StringVarP(&str1, "verbose", "v", "default1", "verbose flag")
	
	// This should cause an error due to shorthand conflict
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic due to shorthand conflict")
		}
	}()
	
	fs.StringVarP(&str2, "version", "v", "default2", "version flag")
}

// customValue is a test implementation of the Value interface
type customValue struct {
	value string
}

func (c *customValue) String() string { return c.value }
func (c *customValue) Set(s string) error { c.value = "custom:" + s; return nil }
func (c *customValue) Type() string { return "custom" }

// TestCustomValue tests custom Value interface implementation
func TestCustomValue(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	cv := &customValue{value: "initial"}
	fs.Var(cv, "custom", "custom flag")
	
	flag := fs.Lookup("custom")
	if flag == nil {
		t.Fatal("Expected to find custom flag")
	}
	
	if flag.Value.Type() != "custom" {
		t.Errorf("Expected type 'custom', got %s", flag.Value.Type())
	}
	
	if flag.DefValue != "initial" {
		t.Errorf("Expected default value 'initial', got %s", flag.DefValue)
	}
}

// TestFlagSetIsolation tests that different FlagSets are isolated
func TestFlagSetIsolation(t *testing.T) {
	fs1 := NewFlagSet("test1", ContinueOnError)
	fs2 := NewFlagSet("test2", ContinueOnError)
	
	var str1, str2 string
	fs1.StringVar(&str1, "flag", "value1", "flag in fs1")
	fs2.StringVar(&str2, "flag", "value2", "flag in fs2")
	
	flag1 := fs1.Lookup("flag")
	flag2 := fs2.Lookup("flag")
	
	if flag1 == nil || flag2 == nil {
		t.Fatal("Expected to find flags in both flag sets")
	}
	
	if flag1.DefValue != "value1" {
		t.Errorf("Expected fs1 flag default 'value1', got %s", flag1.DefValue)
	}
	
	if flag2.DefValue != "value2" {
		t.Errorf("Expected fs2 flag default 'value2', got %s", flag2.DefValue)
	}
	
	if str1 != "value1" {
		t.Errorf("Expected fs1 variable 'value1', got %s", str1)
	}
	
	if str2 != "value2" {
		t.Errorf("Expected fs2 variable 'value2', got %s", str2)
	}
}

// TestVisitAll tests the VisitAll functionality
func TestVisitAll(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var str string
	var num int
	var flag bool
	
	fs.StringVar(&str, "string", "default", "string flag")
	fs.IntVar(&num, "int", 42, "int flag")
	fs.BoolVar(&flag, "bool", true, "bool flag")
	
	visited := make(map[string]bool)
	fs.VisitAll(func(f *Flag) {
		visited[f.Name] = true
	})
	
	expectedFlags := []string{"string", "int", "bool"}
	for _, name := range expectedFlags {
		if !visited[name] {
			t.Errorf("Expected to visit flag %s", name)
		}
	}
	
	if len(visited) != len(expectedFlags) {
		t.Errorf("Expected to visit %d flags, visited %d", len(expectedFlags), len(visited))
	}
}

// TestGlobalFlagSet tests the global CommandLine flag set
func TestGlobalFlagSet(t *testing.T) {
	// Save original CommandLine
	originalCommandLine := CommandLine
	defer func() {
		CommandLine = originalCommandLine
	}()
	
	// Create a new CommandLine for testing
	CommandLine = NewFlagSet("test", ContinueOnError)
	
	var testString string
	StringVar(&testString, "global", "default", "global flag")
	
	flag := Lookup("global")
	if flag == nil {
		t.Fatal("Expected to find global flag")
	}
	
	if flag.Name != "global" {
		t.Errorf("Expected flag name 'global', got %s", flag.Name)
	}
	
	if testString != "default" {
		t.Errorf("Expected variable to be set to 'default', got %s", testString)
	}
}

// TestStringValueImplementation tests the stringValue type implementation
// Requirements: 1.1
func TestStringValueImplementation(t *testing.T) {
	var str string
	sv := newStringValue("initial", &str)
	
	// Test Type() method
	if sv.Type() != "string" {
		t.Errorf("Expected Type() to return 'string', got %s", sv.Type())
	}
	
	// Test String() method
	if sv.String() != "initial" {
		t.Errorf("Expected String() to return 'initial', got %s", sv.String())
	}
	
	// Test Set() method with valid input
	err := sv.Set("new value")
	if err != nil {
		t.Errorf("Expected Set() to succeed, got error: %v", err)
	}
	
	if sv.String() != "new value" {
		t.Errorf("Expected String() to return 'new value' after Set(), got %s", sv.String())
	}
	
	if str != "new value" {
		t.Errorf("Expected underlying variable to be 'new value', got %s", str)
	}
	
	// Test Set() with empty string (should be valid)
	err = sv.Set("")
	if err != nil {
		t.Errorf("Expected Set() with empty string to succeed, got error: %v", err)
	}
	
	if sv.String() != "" {
		t.Errorf("Expected String() to return empty string, got %s", sv.String())
	}
}

// TestIntValueImplementation tests the intValue type implementation
// Requirements: 1.2
func TestIntValueImplementation(t *testing.T) {
	var num int
	iv := newIntValue(42, &num)
	
	// Test Type() method
	if iv.Type() != "int" {
		t.Errorf("Expected Type() to return 'int', got %s", iv.Type())
	}
	
	// Test String() method
	if iv.String() != "42" {
		t.Errorf("Expected String() to return '42', got %s", iv.String())
	}
	
	// Test Set() method with valid input
	err := iv.Set("100")
	if err != nil {
		t.Errorf("Expected Set() to succeed, got error: %v", err)
	}
	
	if iv.String() != "100" {
		t.Errorf("Expected String() to return '100' after Set(), got %s", iv.String())
	}
	
	if num != 100 {
		t.Errorf("Expected underlying variable to be 100, got %d", num)
	}
	
	// Test Set() with negative number
	err = iv.Set("-50")
	if err != nil {
		t.Errorf("Expected Set() with negative number to succeed, got error: %v", err)
	}
	
	if iv.String() != "-50" {
		t.Errorf("Expected String() to return '-50', got %s", iv.String())
	}
	
	// Test Set() with invalid input
	err = iv.Set("not a number")
	if err == nil {
		t.Error("Expected Set() with invalid input to fail")
	}
	
	expectedError := "invalid syntax for integer flag: not a number"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
	
	// Test Set() with float (should fail)
	err = iv.Set("3.14")
	if err == nil {
		t.Error("Expected Set() with float to fail")
	}
}

// TestBoolValueImplementation tests the boolValue type implementation
// Requirements: 1.3
func TestBoolValueImplementation(t *testing.T) {
	var flag bool
	bv := newBoolValue(true, &flag)
	
	// Test Type() method
	if bv.Type() != "bool" {
		t.Errorf("Expected Type() to return 'bool', got %s", bv.Type())
	}
	
	// Test String() method
	if bv.String() != "true" {
		t.Errorf("Expected String() to return 'true', got %s", bv.String())
	}
	
	// Test Set() method with valid inputs
	testCases := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"t", true},
		{"f", false},
		{"T", true},
		{"F", false},
		{"TRUE", true},
		{"FALSE", false},
	}
	
	for _, tc := range testCases {
		err := bv.Set(tc.input)
		if err != nil {
			t.Errorf("Expected Set(%s) to succeed, got error: %v", tc.input, err)
		}
		
		if bool(*bv) != tc.expected {
			t.Errorf("Expected Set(%s) to result in %t, got %t", tc.input, tc.expected, bool(*bv))
		}
		
		expectedString := "true"
		if !tc.expected {
			expectedString = "false"
		}
		if bv.String() != expectedString {
			t.Errorf("Expected String() to return '%s' after Set(%s), got %s", expectedString, tc.input, bv.String())
		}
	}
	
	// Test Set() with invalid input
	err := bv.Set("invalid")
	if err == nil {
		t.Error("Expected Set() with invalid input to fail")
	}
	
	expectedError := "invalid boolean value 'invalid'"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// TestFloat64ValueImplementation tests the float64Value type implementation
// Requirements: 1.4
func TestFloat64ValueImplementation(t *testing.T) {
	var num float64
	fv := newFloat64Value(3.14, &num)
	
	// Test Type() method
	if fv.Type() != "float64" {
		t.Errorf("Expected Type() to return 'float64', got %s", fv.Type())
	}
	
	// Test String() method
	if fv.String() != "3.14" {
		t.Errorf("Expected String() to return '3.14', got %s", fv.String())
	}
	
	// Test Set() method with valid inputs
	testCases := []struct {
		input    string
		expected float64
	}{
		{"2.5", 2.5},
		{"-1.5", -1.5},
		{"0", 0.0},
		{"100", 100.0},
		{"1e10", 1e10},
		{"-3.14159", -3.14159},
	}
	
	for _, tc := range testCases {
		err := fv.Set(tc.input)
		if err != nil {
			t.Errorf("Expected Set(%s) to succeed, got error: %v", tc.input, err)
		}
		
		if float64(*fv) != tc.expected {
			t.Errorf("Expected Set(%s) to result in %f, got %f", tc.input, tc.expected, float64(*fv))
		}
	}
	
	// Test Set() with invalid input
	err := fv.Set("not a number")
	if err == nil {
		t.Error("Expected Set() with invalid input to fail")
	}
	
	expectedError := "invalid syntax for float64 flag: not a number"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// TestDurationValueImplementation tests the durationValue type implementation
// Requirements: 1.5
func TestDurationValueImplementation(t *testing.T) {
	var dur time.Duration
	dv := newDurationValue(5*time.Second, &dur)
	
	// Test Type() method
	if dv.Type() != "duration" {
		t.Errorf("Expected Type() to return 'duration', got %s", dv.Type())
	}
	
	// Test String() method
	if dv.String() != "5s" {
		t.Errorf("Expected String() to return '5s', got %s", dv.String())
	}
	
	// Test Set() method with valid inputs
	testCases := []struct {
		input    string
		expected time.Duration
	}{
		{"1s", 1 * time.Second},
		{"2m", 2 * time.Minute},
		{"3h", 3 * time.Hour},
		{"500ms", 500 * time.Millisecond},
		{"1h30m", 1*time.Hour + 30*time.Minute},
		{"0", 0},
	}
	
	for _, tc := range testCases {
		err := dv.Set(tc.input)
		if err != nil {
			t.Errorf("Expected Set(%s) to succeed, got error: %v", tc.input, err)
		}
		
		if time.Duration(*dv) != tc.expected {
			t.Errorf("Expected Set(%s) to result in %v, got %v", tc.input, tc.expected, time.Duration(*dv))
		}
	}
	
	// Test Set() with invalid input
	err := dv.Set("invalid duration")
	if err == nil {
		t.Error("Expected Set() with invalid input to fail")
	}
	
	expectedError := "invalid duration format for flag: invalid duration"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// TestStringSliceValueImplementation tests the stringSliceValue type implementation
// Requirements: 3.1, 3.2, 3.3, 3.4
func TestStringSliceValueImplementation(t *testing.T) {
	var slice []string
	sv := newStringSliceValue([]string{"initial"}, &slice)
	
	// Test Type() method
	if sv.Type() != "stringSlice" {
		t.Errorf("Expected Type() to return 'stringSlice', got %s", sv.Type())
	}
	
	// Test String() method with initial value
	if sv.String() != "[initial]" {
		t.Errorf("Expected String() to return '[initial]', got %s", sv.String())
	}
	
	// Test Set() method with single value
	err := sv.Set("single")
	if err != nil {
		t.Errorf("Expected Set() to succeed, got error: %v", err)
	}
	
	if len(*sv) != 2 || (*sv)[0] != "initial" || (*sv)[1] != "single" {
		t.Errorf("Expected slice to be [initial, single], got %v", *sv)
	}
	
	// Test Set() method with comma-separated values
	err = sv.Set("a,b,c")
	if err != nil {
		t.Errorf("Expected Set() with comma-separated values to succeed, got error: %v", err)
	}
	
	expected := []string{"initial", "single", "a", "b", "c"}
	if len(*sv) != len(expected) {
		t.Errorf("Expected slice length %d, got %d", len(expected), len(*sv))
	}
	
	for i, v := range expected {
		if (*sv)[i] != v {
			t.Errorf("Expected slice[%d] to be '%s', got '%s'", i, v, (*sv)[i])
		}
	}
	
	// Test String() method with multiple values
	expectedString := "[initial,single,a,b,c]"
	if sv.String() != expectedString {
		t.Errorf("Expected String() to return '%s', got %s", expectedString, sv.String())
	}
	
	// Test with empty slice
	var emptySlice []string
	esv := newStringSliceValue([]string{}, &emptySlice)
	if esv.String() != "[]" {
		t.Errorf("Expected empty slice String() to return '[]', got %s", esv.String())
	}
	
	// Test with values containing spaces (should be trimmed in comma-separated)
	err = esv.Set("  spaced  ,  values  ")
	if err != nil {
		t.Errorf("Expected Set() with spaced values to succeed, got error: %v", err)
	}
	
	if len(*esv) != 2 || (*esv)[0] != "spaced" || (*esv)[1] != "values" {
		t.Errorf("Expected trimmed values [spaced, values], got %v", *esv)
	}
}

// TestIntSliceValueImplementation tests the intSliceValue type implementation
// Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
func TestIntSliceValueImplementation(t *testing.T) {
	var slice []int
	iv := newIntSliceValue([]int{42}, &slice)
	
	// Test Type() method
	if iv.Type() != "intSlice" {
		t.Errorf("Expected Type() to return 'intSlice', got %s", iv.Type())
	}
	
	// Test String() method with initial value
	if iv.String() != "[42]" {
		t.Errorf("Expected String() to return '[42]', got %s", iv.String())
	}
	
	// Test Set() method with single value
	err := iv.Set("100")
	if err != nil {
		t.Errorf("Expected Set() to succeed, got error: %v", err)
	}
	
	if len(*iv) != 2 || (*iv)[0] != 42 || (*iv)[1] != 100 {
		t.Errorf("Expected slice to be [42, 100], got %v", *iv)
	}
	
	// Test Set() method with comma-separated values
	err = iv.Set("1,2,3")
	if err != nil {
		t.Errorf("Expected Set() with comma-separated values to succeed, got error: %v", err)
	}
	
	expected := []int{42, 100, 1, 2, 3}
	if len(*iv) != len(expected) {
		t.Errorf("Expected slice length %d, got %d", len(expected), len(*iv))
	}
	
	for i, v := range expected {
		if (*iv)[i] != v {
			t.Errorf("Expected slice[%d] to be %d, got %d", i, v, (*iv)[i])
		}
	}
	
	// Test String() method with multiple values
	expectedString := "[42,100,1,2,3]"
	if iv.String() != expectedString {
		t.Errorf("Expected String() to return '%s', got %s", expectedString, iv.String())
	}
	
	// Test with empty slice
	var emptySlice []int
	eiv := newIntSliceValue([]int{}, &emptySlice)
	if eiv.String() != "[]" {
		t.Errorf("Expected empty slice String() to return '[]', got %s", eiv.String())
	}
	
	// Test Set() with invalid single value
	err = eiv.Set("not a number")
	if err == nil {
		t.Error("Expected Set() with invalid single value to fail")
	}
	
	expectedError := "invalid syntax for integer slice element: not a number"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
	
	// Test Set() with invalid comma-separated value
	err = eiv.Set("1,not a number,3")
	if err == nil {
		t.Error("Expected Set() with invalid comma-separated value to fail")
	}
	
	expectedError = "invalid syntax for integer slice element: not a number"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
	
	// Test with negative numbers
	var freshSlice []int
	fiv := newIntSliceValue([]int{}, &freshSlice)
	err = fiv.Set("-5,-10")
	if err != nil {
		t.Errorf("Expected Set() with negative numbers to succeed, got error: %v", err)
	}
	
	if len(*fiv) != 2 || (*fiv)[0] != -5 || (*fiv)[1] != -10 {
		t.Errorf("Expected slice to be [-5, -10], got %v", *fiv)
	}
}

// TestStringSliceFlag tests string slice flag creation and basic functionality
func TestStringSliceFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var stringSliceVar []string
	fs.StringSliceVar(&stringSliceVar, "strings", []string{"default1", "default2"}, "string slice flag")
	
	flag := fs.Lookup("strings")
	if flag == nil {
		t.Fatal("Expected to find strings flag")
	}
	
	if flag.Name != "strings" {
		t.Errorf("Expected flag name 'strings', got %s", flag.Name)
	}
	
	if flag.DefValue != "[default1,default2]" {
		t.Errorf("Expected default value '[default1,default2]', got %s", flag.DefValue)
	}
	
	if flag.Usage != "string slice flag" {
		t.Errorf("Expected usage 'string slice flag', got %s", flag.Usage)
	}
	
	if len(stringSliceVar) != 2 || stringSliceVar[0] != "default1" || stringSliceVar[1] != "default2" {
		t.Errorf("Expected variable to be set to default value [default1, default2], got %v", stringSliceVar)
	}
}

// TestIntSliceFlag tests int slice flag creation and basic functionality
func TestIntSliceFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	
	var intSliceVar []int
	fs.IntSliceVar(&intSliceVar, "ints", []int{1, 2, 3}, "int slice flag")
	
	flag := fs.Lookup("ints")
	if flag == nil {
		t.Fatal("Expected to find ints flag")
	}
	
	if flag.Name != "ints" {
		t.Errorf("Expected flag name 'ints', got %s", flag.Name)
	}
	
	if flag.DefValue != "[1,2,3]" {
		t.Errorf("Expected default value '[1,2,3]', got %s", flag.DefValue)
	}
	
	if flag.Usage != "int slice flag" {
		t.Errorf("Expected usage 'int slice flag', got %s", flag.Usage)
	}
	
	if len(intSliceVar) != 3 || intSliceVar[0] != 1 || intSliceVar[1] != 2 || intSliceVar[2] != 3 {
		t.Errorf("Expected variable to be set to default value [1, 2, 3], got %v", intSliceVar)
	}
}
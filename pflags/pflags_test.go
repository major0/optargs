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
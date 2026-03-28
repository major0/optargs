package pflags

import (
	"bytes"
	"fmt"
	"strings"
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

// TestFlagCreationAllTypes tests flag creation for every value type via the
// Var/VarP and convenience methods. Table-driven: one row per type variant.
func TestFlagCreationAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		register func(fs *FlagSet)
		flag     string
		defValue string
		typeName string
	}{
		// StringVar / StringVarP / String / StringP
		{"StringVar", func(fs *FlagSet) { var s string; fs.StringVar(&s, "f", "abc", "u") }, "f", "abc", "string"},
		{"StringVarP", func(fs *FlagSet) { var s string; fs.StringVarP(&s, "f", "s", "abc", "u") }, "f", "abc", "string"},
		{"String", func(fs *FlagSet) { fs.String("f", "abc", "u") }, "f", "abc", "string"},
		{"StringP", func(fs *FlagSet) { fs.StringP("f", "s", "abc", "u") }, "f", "abc", "string"},
		// BoolVar / BoolVarP / Bool / BoolP
		{"BoolVar", func(fs *FlagSet) { var b bool; fs.BoolVar(&b, "f", true, "u") }, "f", "true", "bool"},
		{"BoolVarP", func(fs *FlagSet) { var b bool; fs.BoolVarP(&b, "f", "b", true, "u") }, "f", "true", "bool"},
		{"Bool", func(fs *FlagSet) { fs.Bool("f", true, "u") }, "f", "true", "bool"},
		{"BoolP", func(fs *FlagSet) { fs.BoolP("f", "b", true, "u") }, "f", "true", "bool"},
		// IntVar / IntVarP / Int / IntP
		{"IntVar", func(fs *FlagSet) { var i int; fs.IntVar(&i, "f", 42, "u") }, "f", "42", "int"},
		{"IntVarP", func(fs *FlagSet) { var i int; fs.IntVarP(&i, "f", "i", 42, "u") }, "f", "42", "int"},
		{"Int", func(fs *FlagSet) { fs.Int("f", 42, "u") }, "f", "42", "int"},
		{"IntP", func(fs *FlagSet) { fs.IntP("f", "i", 42, "u") }, "f", "42", "int"},
		// Int64Var / Int64VarP / Int64 / Int64P
		{"Int64Var", func(fs *FlagSet) { var i int64; fs.Int64Var(&i, "f", 99, "u") }, "f", "99", "int64"},
		{"Int64VarP", func(fs *FlagSet) { var i int64; fs.Int64VarP(&i, "f", "l", 99, "u") }, "f", "99", "int64"},
		{"Int64", func(fs *FlagSet) { fs.Int64("f", 99, "u") }, "f", "99", "int64"},
		{"Int64P", func(fs *FlagSet) { fs.Int64P("f", "l", 99, "u") }, "f", "99", "int64"},
		// UintVar / UintVarP / Uint / UintP
		{"UintVar", func(fs *FlagSet) { var u uint; fs.UintVar(&u, "f", 7, "u") }, "f", "7", "uint"},
		{"UintVarP", func(fs *FlagSet) { var u uint; fs.UintVarP(&u, "f", "u", 7, "u") }, "f", "7", "uint"},
		{"Uint", func(fs *FlagSet) { fs.Uint("f", 7, "u") }, "f", "7", "uint"},
		{"UintP", func(fs *FlagSet) { fs.UintP("f", "u", 7, "u") }, "f", "7", "uint"},
		// Uint64Var / Uint64VarP / Uint64 / Uint64P
		{"Uint64Var", func(fs *FlagSet) { var u uint64; fs.Uint64Var(&u, "f", 8, "u") }, "f", "8", "uint64"},
		{"Uint64VarP", func(fs *FlagSet) { var u uint64; fs.Uint64VarP(&u, "f", "x", 8, "u") }, "f", "8", "uint64"},
		{"Uint64", func(fs *FlagSet) { fs.Uint64("f", 8, "u") }, "f", "8", "uint64"},
		{"Uint64P", func(fs *FlagSet) { fs.Uint64P("f", "x", 8, "u") }, "f", "8", "uint64"},
		// Float64Var / Float64VarP / Float64 / Float64P
		{"Float64Var", func(fs *FlagSet) { var f64 float64; fs.Float64Var(&f64, "f", 3.14, "u") }, "f", "3.14", "float64"},
		{"Float64VarP", func(fs *FlagSet) { var f64 float64; fs.Float64VarP(&f64, "f", "g", 3.14, "u") }, "f", "3.14", "float64"},
		{"Float64", func(fs *FlagSet) { fs.Float64("f", 3.14, "u") }, "f", "3.14", "float64"},
		{"Float64P", func(fs *FlagSet) { fs.Float64P("f", "g", 3.14, "u") }, "f", "3.14", "float64"},
		// DurationVar / DurationVarP / Duration / DurationP
		{"DurationVar", func(fs *FlagSet) { var d time.Duration; fs.DurationVar(&d, "f", 5*time.Second, "u") }, "f", "5s", "duration"},
		{"DurationVarP", func(fs *FlagSet) { var d time.Duration; fs.DurationVarP(&d, "f", "d", 5*time.Second, "u") }, "f", "5s", "duration"},
		{"Duration", func(fs *FlagSet) { fs.Duration("f", 5*time.Second, "u") }, "f", "5s", "duration"},
		{"DurationP", func(fs *FlagSet) { fs.DurationP("f", "d", 5*time.Second, "u") }, "f", "5s", "duration"},
		// StringSliceVar / StringSliceVarP / StringSlice / StringSliceP
		{"StringSliceVar", func(fs *FlagSet) { var s []string; fs.StringSliceVar(&s, "f", []string{"a"}, "u") }, "f", "[a]", "stringSlice"},
		{"StringSliceVarP", func(fs *FlagSet) { var s []string; fs.StringSliceVarP(&s, "f", "s", []string{"a"}, "u") }, "f", "[a]", "stringSlice"},
		{"StringSlice", func(fs *FlagSet) { fs.StringSlice("f", []string{"a"}, "u") }, "f", "[a]", "stringSlice"},
		{"StringSliceP", func(fs *FlagSet) { fs.StringSliceP("f", "s", []string{"a"}, "u") }, "f", "[a]", "stringSlice"},
		// IntSliceVar / IntSliceVarP / IntSlice / IntSliceP
		{"IntSliceVar", func(fs *FlagSet) { var s []int; fs.IntSliceVar(&s, "f", []int{1}, "u") }, "f", "[1]", "intSlice"},
		{"IntSliceVarP", func(fs *FlagSet) { var s []int; fs.IntSliceVarP(&s, "f", "n", []int{1}, "u") }, "f", "[1]", "intSlice"},
		{"IntSlice", func(fs *FlagSet) { fs.IntSlice("f", []int{1}, "u") }, "f", "[1]", "intSlice"},
		{"IntSliceP", func(fs *FlagSet) { fs.IntSliceP("f", "n", []int{1}, "u") }, "f", "[1]", "intSlice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			tt.register(fs)
			f := fs.Lookup(tt.flag)
			if f == nil {
				t.Fatalf("flag %q not found", tt.flag)
			}
			if f.DefValue != tt.defValue {
				t.Errorf("DefValue = %q, want %q", f.DefValue, tt.defValue)
			}
			if f.Value.Type() != tt.typeName {
				t.Errorf("Type = %q, want %q", f.Value.Type(), tt.typeName)
			}
		})
	}
}

// TestValueSetAndString tests Set/String/Type round-trip for every value type.
// Table-driven: covers valid inputs, invalid inputs, and error messages.
func TestValueSetAndString(t *testing.T) {
	tests := []struct {
		name      string
		value     Value
		input     string
		wantStr   string
		wantErr   string // substring; empty means no error
	}{
		// string
		{"string/valid", newStringValue("", new(string)), "hello", "hello", ""},
		{"string/empty", newStringValue("x", new(string)), "", "", ""},
		// bool
		{"bool/true", newBoolValue(false, new(bool)), "true", "true", ""},
		{"bool/false", newBoolValue(true, new(bool)), "false", "false", ""},
		{"bool/1", newBoolValue(false, new(bool)), "1", "true", ""},
		{"bool/0", newBoolValue(true, new(bool)), "0", "false", ""},
		{"bool/t", newBoolValue(false, new(bool)), "t", "true", ""},
		{"bool/f", newBoolValue(true, new(bool)), "f", "false", ""},
		{"bool/T", newBoolValue(false, new(bool)), "T", "true", ""},
		{"bool/F", newBoolValue(true, new(bool)), "F", "false", ""},
		{"bool/TRUE", newBoolValue(false, new(bool)), "TRUE", "true", ""},
		{"bool/FALSE", newBoolValue(true, new(bool)), "FALSE", "false", ""},
		{"bool/invalid", newBoolValue(false, new(bool)), "invalid", "", "invalid boolean value"},
		// int
		{"int/valid", newIntValue(0, new(int)), "100", "100", ""},
		{"int/negative", newIntValue(0, new(int)), "-50", "-50", ""},
		{"int/invalid", newIntValue(0, new(int)), "abc", "", "invalid syntax for integer flag"},
		{"int/float", newIntValue(0, new(int)), "3.14", "", "invalid syntax for integer flag"},
		// int64
		{"int64/valid", newInt64Value(0, new(int64)), "9999999999", "9999999999", ""},
		{"int64/negative", newInt64Value(0, new(int64)), "-100", "-100", ""},
		{"int64/invalid", newInt64Value(0, new(int64)), "abc", "", "invalid syntax for int64 flag"},
		// uint
		{"uint/valid", newUintValue(0, new(uint)), "42", "42", ""},
		{"uint/invalid", newUintValue(0, new(uint)), "-1", "", "invalid syntax for uint flag"},
		{"uint/text", newUintValue(0, new(uint)), "abc", "", "invalid syntax for uint flag"},
		// uint64
		{"uint64/valid", newUint64Value(0, new(uint64)), "18446744073709551615", "18446744073709551615", ""},
		{"uint64/invalid", newUint64Value(0, new(uint64)), "-1", "", "invalid syntax for uint64 flag"},
		// float64
		{"float64/valid", newFloat64Value(0, new(float64)), "2.5", "2.5", ""},
		{"float64/negative", newFloat64Value(0, new(float64)), "-1.5", "-1.5", ""},
		{"float64/scientific", newFloat64Value(0, new(float64)), "1e10", "1e+10", ""},
		{"float64/invalid", newFloat64Value(0, new(float64)), "abc", "", "invalid syntax for float64 flag"},
		// duration
		{"duration/seconds", newDurationValue(0, new(time.Duration)), "1s", "1s", ""},
		{"duration/minutes", newDurationValue(0, new(time.Duration)), "2m", "2m0s", ""},
		{"duration/hours", newDurationValue(0, new(time.Duration)), "3h", "3h0m0s", ""},
		{"duration/compound", newDurationValue(0, new(time.Duration)), "1h30m", "1h30m0s", ""},
		{"duration/invalid", newDurationValue(0, new(time.Duration)), "bad", "", "invalid duration format"},
		// stringSlice
		{"stringSlice/single", newStringSliceValue([]string{}, new([]string)), "one", "[one]", ""},
		{"stringSlice/csv", newStringSliceValue([]string{}, new([]string)), "a,b,c", "[a,b,c]", ""},
		{"stringSlice/trimmed", newStringSliceValue([]string{}, new([]string)), " x , y ", "[x,y]", ""},
		// intSlice
		{"intSlice/single", newIntSliceValue([]int{}, new([]int)), "5", "[5]", ""},
		{"intSlice/csv", newIntSliceValue([]int{}, new([]int)), "1,2,3", "[1,2,3]", ""},
		{"intSlice/negative", newIntSliceValue([]int{}, new([]int)), "-5,-10", "[-5,-10]", ""},
		{"intSlice/invalid_single", newIntSliceValue([]int{}, new([]int)), "abc", "", "invalid syntax for integer slice element"},
		{"intSlice/invalid_csv", newIntSliceValue([]int{}, new([]int)), "1,abc,3", "", "invalid syntax for integer slice element"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.value.Set(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Set(%q) expected error containing %q, got nil", tt.input, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Set(%q) error = %q, want substring %q", tt.input, err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Set(%q) unexpected error: %v", tt.input, err)
			}
			if got := tt.value.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

// TestIsBoolFlag verifies the boolValue.IsBoolFlag() method.
func TestIsBoolFlag(t *testing.T) {
	bv := newBoolValue(false, new(bool))
	if !bv.IsBoolFlag() {
		t.Error("IsBoolFlag() should return true")
	}
}

// TestShorthandRegistration tests shorthand flag creation and conflict detection.
func TestShorthandRegistration(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var s string
	fs.StringVarP(&s, "verbose", "v", "default", "verbose flag")

	f := fs.Lookup("verbose")
	if f == nil {
		t.Fatal("flag not found")
	}
	if f.Shorthand != "v" {
		t.Errorf("Shorthand = %q, want %q", f.Shorthand, "v")
	}
	if fs.shorthand["v"] != "verbose" {
		t.Errorf("shorthand map: got %q, want %q", fs.shorthand["v"], "verbose")
	}
}

// TestShorthandConflict tests that shorthand conflicts panic.
func TestShorthandConflict(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var s1, s2 string
	fs.StringVarP(&s1, "verbose", "v", "", "")

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic due to shorthand conflict")
		}
	}()
	fs.StringVarP(&s2, "version", "v", "", "")
}

// TestFlagSetIsolation tests that different FlagSets are isolated.
func TestFlagSetIsolation(t *testing.T) {
	fs1 := NewFlagSet("test1", ContinueOnError)
	fs2 := NewFlagSet("test2", ContinueOnError)

	var s1, s2 string
	fs1.StringVar(&s1, "flag", "v1", "")
	fs2.StringVar(&s2, "flag", "v2", "")

	if fs1.Lookup("flag").DefValue != "v1" {
		t.Error("fs1 flag default wrong")
	}
	if fs2.Lookup("flag").DefValue != "v2" {
		t.Error("fs2 flag default wrong")
	}
}

// TestBooleanFlagParsing tests boolean flag parsing: no-arg, explicit values,
// negation syntax, shorthand. Table-driven — the authoritative form for this invariant.
func TestBooleanFlagParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		defaultVal  bool
		expectedVal bool
		shouldError bool
	}{
		{"no-arg sets true", []string{"--verbose"}, false, true, false},
		{"no-arg default true", []string{"--verbose"}, true, true, false},
		{"explicit true", []string{"--verbose=true"}, false, true, false},
		{"explicit false", []string{"--verbose=false"}, true, false, false},
		{"explicit 1", []string{"--verbose=1"}, false, true, false},
		{"explicit 0", []string{"--verbose=0"}, true, false, false},
		{"negation default true", []string{"--no-verbose"}, true, false, false},
		{"negation default false", []string{"--no-verbose"}, false, false, false},
		{"invalid value", []string{"--verbose=invalid"}, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var v bool
			fs.BoolVar(&v, "verbose", tt.defaultVal, "")

			err := fs.Parse(tt.args)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for %v", tt.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v != tt.expectedVal {
				t.Errorf("got %t, want %t", v, tt.expectedVal)
			}
			if f := fs.Lookup("verbose"); f == nil || !f.Changed {
				t.Error("flag should be marked Changed")
			}
		})
	}
}

// TestBooleanShorthandParsing tests boolean shorthand parsing.
func TestBooleanShorthandParsing(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var v bool
	fs.BoolVarP(&v, "verbose", "v", false, "")

	if err := fs.Parse([]string{"-v"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !v {
		t.Error("expected true after -v")
	}
}

// TestNegationHandlerBranches exercises makeNegationHandler branches not covered
// by the boolean parsing table (=false → sets true, =0 → sets true, invalid value).
func TestNegationHandlerBranches(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
		wantErr  bool
	}{
		{"no-flag=true", []string{"--no-flag=true"}, false, false},
		{"no-flag=1", []string{"--no-flag=1"}, false, false},
		{"no-flag=t", []string{"--no-flag=t"}, false, false},
		{"no-flag=false", []string{"--no-flag=false"}, true, false},
		{"no-flag=0", []string{"--no-flag=0"}, true, false},
		{"no-flag=f", []string{"--no-flag=f"}, true, false},
		{"no-flag=invalid", []string{"--no-flag=invalid"}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var v bool
			fs.BoolVar(&v, "flag", false, "")

			err := fs.Parse(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v != tt.expected {
				t.Errorf("got %t, want %t", v, tt.expected)
			}
		})
	}
}

// TestParseStateAndArgs tests Parsed(), Args(), NArg(), Arg() after parsing.
func TestParseStateAndArgs(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var s string
	fs.StringVar(&s, "name", "", "")

	if fs.Parsed() {
		t.Error("Parsed() should be false before Parse()")
	}

	err := fs.Parse([]string{"--name", "val", "pos1", "pos2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fs.Parsed() {
		t.Error("Parsed() should be true after Parse()")
	}
	if s != "val" {
		t.Errorf("flag value = %q, want %q", s, "val")
	}
	if fs.NArg() != 2 {
		t.Errorf("NArg() = %d, want 2", fs.NArg())
	}
	args := fs.Args()
	if len(args) != 2 || args[0] != "pos1" || args[1] != "pos2" {
		t.Errorf("Args() = %v, want [pos1 pos2]", args)
	}
	if fs.Arg(0) != "pos1" {
		t.Errorf("Arg(0) = %q, want %q", fs.Arg(0), "pos1")
	}
	if fs.Arg(1) != "pos2" {
		t.Errorf("Arg(1) = %q, want %q", fs.Arg(1), "pos2")
	}
	// Out of bounds
	if fs.Arg(-1) != "" {
		t.Errorf("Arg(-1) = %q, want empty", fs.Arg(-1))
	}
	if fs.Arg(99) != "" {
		t.Errorf("Arg(99) = %q, want empty", fs.Arg(99))
	}
}

// TestVisitAllAndVisit tests VisitAll and Visit behavior.
func TestVisitAllAndVisit(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var s string
	var n int
	var b bool
	fs.StringVar(&s, "string", "", "")
	fs.IntVar(&n, "int", 0, "")
	fs.BoolVar(&b, "bool", false, "")

	// VisitAll should see all 3
	allNames := map[string]bool{}
	fs.VisitAll(func(f *Flag) { allNames[f.Name] = true })
	if len(allNames) != 3 {
		t.Errorf("VisitAll visited %d flags, want 3", len(allNames))
	}

	// Visit should see 0 (none changed)
	changedNames := map[string]bool{}
	fs.Visit(func(f *Flag) { changedNames[f.Name] = true })
	if len(changedNames) != 0 {
		t.Errorf("Visit visited %d flags before parse, want 0", len(changedNames))
	}

	// Set one flag, Visit should see it
	fs.Set("string", "hello")
	changedNames = map[string]bool{}
	fs.Visit(func(f *Flag) { changedNames[f.Name] = true })
	if !changedNames["string"] || len(changedNames) != 1 {
		t.Errorf("Visit after Set: got %v, want {string}", changedNames)
	}
}

// TestSetOutput tests SetOutput and out() behavior.
func TestSetOutput(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)

	// Default output should be stderr (non-nil from out())
	w := fs.out()
	if w == nil {
		t.Error("out() should not return nil")
	}

	// Set custom output
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	if fs.out() != &buf {
		t.Error("out() should return custom writer after SetOutput")
	}
}

// TestSetUnknownFlag tests Set() with a non-existent flag.
func TestSetUnknownFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	err := fs.Set("nonexistent", "val")
	if err == nil {
		t.Error("expected error for unknown flag")
	}
	if !strings.Contains(err.Error(), "no such flag") {
		t.Errorf("error = %q, want 'no such flag'", err.Error())
	}
}

// TestCustomValue tests custom Value interface integration.
type customValue struct {
	value string
}

func (c *customValue) String() string     { return c.value }
func (c *customValue) Set(s string) error { c.value = "custom:" + s; return nil }
func (c *customValue) Type() string       { return "custom" }

func TestCustomValue(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	cv := &customValue{value: "initial"}
	fs.Var(cv, "custom", "custom flag")

	f := fs.Lookup("custom")
	if f == nil {
		t.Fatal("flag not found")
	}
	if f.Value.Type() != "custom" {
		t.Errorf("Type = %q, want %q", f.Value.Type(), "custom")
	}
	if f.DefValue != "initial" {
		t.Errorf("DefValue = %q, want %q", f.DefValue, "initial")
	}
}

// TestPrintDefaults tests PrintDefaults output for various flag configurations.
func TestPrintDefaults(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)

	var s string
	var b bool
	var n int
	fs.StringVarP(&s, "output", "o", "file.txt", "output `filename`")
	fs.BoolVar(&b, "verbose", false, "enable verbose")
	fs.IntVar(&n, "count", 0, "number of items")

	// Mark one flag as hidden
	if f := fs.Lookup("count"); f != nil {
		f.Hidden = true
	}

	fs.PrintDefaults()
	out := buf.String()

	// Should contain shorthand format
	if !strings.Contains(out, "-o, --output") {
		t.Errorf("output missing shorthand format, got:\n%s", out)
	}
	// Should contain usage with unquoted backtick name
	if !strings.Contains(out, "filename") {
		t.Errorf("output missing unquoted usage name, got:\n%s", out)
	}
	// Should show non-zero default
	if !strings.Contains(out, "(default") {
		t.Errorf("output missing default value, got:\n%s", out)
	}
	// Hidden flag should not appear
	if strings.Contains(out, "count") {
		t.Errorf("hidden flag 'count' should not appear, got:\n%s", out)
	}
}

// TestDefaultUsage tests the defaultUsage function.
func TestDefaultUsage(t *testing.T) {
	// Named FlagSet
	fs := NewFlagSet("myapp", ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.BoolVar(new(bool), "help", false, "show help")
	fs.defaultUsage()
	if !strings.Contains(buf.String(), "Usage of myapp:") {
		t.Errorf("expected 'Usage of myapp:', got:\n%s", buf.String())
	}

	// Empty-name FlagSet
	fs2 := NewFlagSet("", ContinueOnError)
	var buf2 bytes.Buffer
	fs2.SetOutput(&buf2)
	fs2.defaultUsage()
	if !strings.Contains(buf2.String(), "Usage:") {
		t.Errorf("expected 'Usage:', got:\n%s", buf2.String())
	}
}

// TestUnquoteUsage tests UnquoteUsage for various patterns.
func TestUnquoteUsage(t *testing.T) {
	tests := []struct {
		name      string
		flag      *Flag
		wantName  string
		wantUsage string
	}{
		{
			"backtick name",
			&Flag{Usage: "output `filename` to write", Value: newStringValue("", new(string))},
			"filename", "output filename to write",
		},
		{
			"no backtick string",
			&Flag{Usage: "output file", Value: newStringValue("", new(string))},
			"string", "output file",
		},
		{
			"no backtick bool",
			&Flag{Usage: "enable verbose", Value: newBoolValue(false, new(bool))},
			"", "enable verbose",
		},
		{
			"no backtick float64",
			&Flag{Usage: "rate limit", Value: newFloat64Value(0, new(float64))},
			"float", "rate limit",
		},
		{
			"no backtick int64",
			&Flag{Usage: "max size", Value: newInt64Value(0, new(int64))},
			"int", "max size",
		},
		{
			"no backtick uint64",
			&Flag{Usage: "max count", Value: newUint64Value(0, new(uint64))},
			"uint", "max count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, usage := UnquoteUsage(tt.flag)
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
			if usage != tt.wantUsage {
				t.Errorf("usage = %q, want %q", usage, tt.wantUsage)
			}
		})
	}
}

// TestIsZeroValue tests isZeroValue for all known types.
func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		flag     *Flag
		value    string
		wantZero bool
	}{
		{"bool/zero", &Flag{Value: newBoolValue(false, new(bool))}, "false", true},
		{"bool/nonzero", &Flag{Value: newBoolValue(false, new(bool))}, "true", false},
		{"string/zero", &Flag{Value: newStringValue("", new(string))}, "", true},
		{"string/nonzero", &Flag{Value: newStringValue("", new(string))}, "x", false},
		{"int/zero", &Flag{Value: newIntValue(0, new(int))}, "0", true},
		{"int/nonzero", &Flag{Value: newIntValue(0, new(int))}, "42", false},
		{"int64/zero", &Flag{Value: newInt64Value(0, new(int64))}, "0", true},
		{"uint/zero", &Flag{Value: newUintValue(0, new(uint))}, "0", true},
		{"uint64/zero", &Flag{Value: newUint64Value(0, new(uint64))}, "0", true},
		{"float64/zero", &Flag{Value: newFloat64Value(0, new(float64))}, "0", true},
		{"duration/zero", &Flag{Value: newDurationValue(0, new(time.Duration))}, "0s", true},
		{"stringSlice/zero", &Flag{Value: newStringSliceValue([]string{}, new([]string))}, "[]", true},
		{"intSlice/zero", &Flag{Value: newIntSliceValue([]int{}, new([]int))}, "[]", true},
		{"custom/nonzero", &Flag{Value: &customValue{value: "x"}}, "x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isZeroValue(tt.flag, tt.value)
			if got != tt.wantZero {
				t.Errorf("isZeroValue(%q) = %t, want %t", tt.value, got, tt.wantZero)
			}
		})
	}
}

// TestTranslateError tests error translation from OptArgs Core to pflag format.
func TestTranslateError(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
	}{
		{"nil", "", ""},
		{"unknown long", "unknown option: verbose", "unknown flag: --verbose"},
		{"unknown short", "unknown option: v", "unknown shorthand flag: 'v'"},
		{"unknown bare", "unknown option", "unknown flag: unknown option"},
		{"requires arg long", "option requires an argument: output", "flag needs an argument: --output"},
		{"requires arg short", "option requires an argument: o", "flag needs an argument: -o"},
		{"requires arg bare", "option requires an argument", "flag needs an argument: option requires an argument"},
		{"passthrough", "some other error", "some other error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == "" {
				if translateError(nil) != nil {
					t.Error("translateError(nil) should return nil")
				}
				return
			}
			got := translateError(fmt.Errorf("%s", tt.input))
			if got == nil {
				t.Fatal("expected non-nil error")
			}
			if got.Error() != tt.want {
				t.Errorf("got %q, want %q", got.Error(), tt.want)
			}
		})
	}
}

// TestParseUnknownFlag tests that parsing an unknown flag returns an error.
func TestParseUnknownFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var s string
	fs.StringVar(&s, "known", "", "")

	err := fs.Parse([]string{"--unknown", "val"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("error = %q, want 'unknown flag'", err.Error())
	}
}

// TestAdvancedGNULongestMatching tests GNU getopt_long() longest matching behavior.
func TestAdvancedGNULongestMatching(t *testing.T) {
	tests := []struct {
		name          string
		flagDefs      map[string]string
		args          []string
		expectedFlag  string
		expectedValue string
		shouldError   bool
	}{
		{
			name:          "longest_match_enable_prefix",
			flagDefs:      map[string]string{"enable-bob": "", "enable-bobadufoo": ""},
			args:          []string{"--enable-bobadufoo", "test-value"},
			expectedFlag:  "enable-bobadufoo",
			expectedValue: "test-value",
		},
		{
			name:          "shorter_match_when_exact",
			flagDefs:      map[string]string{"enable-bob": "", "enable-bobadufoo": ""},
			args:          []string{"--enable-bob", "test-value"},
			expectedFlag:  "enable-bob",
			expectedValue: "test-value",
		},
		{
			name:          "longest_match_with_equals",
			flagDefs:      map[string]string{"system": "", "system7": "", "system7-ex": ""},
			args:          []string{"--system7-ex=extended-value"},
			expectedFlag:  "system7-ex",
			expectedValue: "extended-value",
		},
		{
			name:          "prefix_disambiguation",
			flagDefs:      map[string]string{"verbose": "", "verbose-mode": "", "verb": ""},
			args:          []string{"--verbose-mode", "detailed"},
			expectedFlag:  "verbose-mode",
			expectedValue: "detailed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			flagVars := make(map[string]*string)
			for name := range tt.flagDefs {
				var v string
				flagVars[name] = &v
				fs.StringVar(&v, name, "", "")
			}

			err := fs.Parse(tt.args)
			if tt.shouldError {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := *flagVars[tt.expectedFlag]; got != tt.expectedValue {
				t.Errorf("flag %s = %q, want %q", tt.expectedFlag, got, tt.expectedValue)
			}
			if f := fs.Lookup(tt.expectedFlag); f == nil || !f.Changed {
				t.Errorf("flag %s not marked Changed", tt.expectedFlag)
			}
		})
	}
}

// TestAdvancedGNUSpecialCharacters tests special characters in option names.
func TestAdvancedGNUSpecialCharacters(t *testing.T) {
	tests := []struct {
		name          string
		flagName      string
		args          []string
		expectedValue string
	}{
		{"colon_space", "system7:verbose", []string{"--system7:verbose", "enabled"}, "enabled"},
		{"colon_equals", "system7:prefix", []string{"--system7:prefix=/my/path"}, "/my/path"},
		{"nested_equals", "system7:path=bindir", []string{"--system7:path=bindir=/usr/bin"}, "/usr/bin"},
		{"equals_space", "config=file", []string{"--config=file", "myconfig.json"}, "myconfig.json"},
		{"equals_equals", "config=default", []string{"--config=default=prod.conf"}, "prod.conf"},
		{"mixed_special", "app:config=env:prod", []string{"--app:config=env:prod=live"}, "live"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var v string
			fs.StringVar(&v, tt.flagName, "", "")

			if err := fs.Parse(tt.args); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v != tt.expectedValue {
				t.Errorf("got %q, want %q", v, tt.expectedValue)
			}
		})
	}
}

// TestAdvancedGNUAmbiguityResolution tests exact-match requirement.
func TestAdvancedGNUAmbiguityResolution(t *testing.T) {
	tests := []struct {
		name         string
		flagDefs     []string
		args         []string
		expectedFlag string
		shouldError  bool
	}{
		{"partial_rejected", []string{"verbose", "version"}, []string{"--ver", "val"}, "", true},
		{"exact_match", []string{"help", "help-extended"}, []string{"--help", "val"}, "help", false},
		{"exact_long", []string{"system-config", "system-cache", "system"}, []string{"--system-config", "val"}, "system-config", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			flagVars := make(map[string]*string)
			for _, name := range tt.flagDefs {
				var v string
				flagVars[name] = &v
				fs.StringVar(&v, name, "", "")
			}

			err := fs.Parse(tt.args)
			if tt.shouldError {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := *flagVars[tt.expectedFlag]; got != "val" {
				t.Errorf("flag %s = %q, want %q", tt.expectedFlag, got, "val")
			}
		})
	}
}

// TestAdvancedGNUComplexScenarios tests complex real-world flag combinations.
func TestAdvancedGNUComplexScenarios(t *testing.T) {
	t.Run("build_system_flags", func(t *testing.T) {
		fs := NewFlagSet("build", ContinueOnError)
		var enableBob, enableBobadufoo, sysVerbose, sysPath, configEnv, debugLevel string
		fs.StringVar(&enableBob, "enable-bob", "", "")
		fs.StringVar(&enableBobadufoo, "enable-bobadufoo", "", "")
		fs.StringVar(&sysVerbose, "system7:verbose", "", "")
		fs.StringVar(&sysPath, "system7:path=bindir", "", "")
		fs.StringVar(&configEnv, "config=env", "", "")
		fs.StringVar(&debugLevel, "debug:level", "", "")

		args := []string{
			"--enable-bobadufoo", "advanced",
			"--system7:verbose=detailed",
			"--system7:path=bindir=/usr/local/bin",
			"--config=env", "production",
			"--debug:level=trace",
		}
		if err := fs.Parse(args); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if enableBobadufoo != "advanced" {
			t.Errorf("enable-bobadufoo = %q", enableBobadufoo)
		}
		if sysVerbose != "detailed" {
			t.Errorf("system7:verbose = %q", sysVerbose)
		}
		if sysPath != "/usr/local/bin" {
			t.Errorf("system7:path=bindir = %q", sysPath)
		}
		if configEnv != "production" {
			t.Errorf("config=env = %q", configEnv)
		}
		if debugLevel != "trace" {
			t.Errorf("debug:level = %q", debugLevel)
		}
		if enableBob != "" {
			t.Errorf("enable-bob should be empty, got %q", enableBob)
		}
	})

	t.Run("nested_configuration", func(t *testing.T) {
		fs := NewFlagSet("config", ContinueOnError)
		var dbHost, dbPort, cacheUrl, logLevel string
		fs.StringVar(&dbHost, "db:host=primary", "", "")
		fs.StringVar(&dbPort, "db:port=primary", "", "")
		fs.StringVar(&cacheUrl, "cache:url=redis", "", "")
		fs.StringVar(&logLevel, "log:level=app", "", "")

		args := []string{
			"--db:host=primary=db1.example.com",
			"--db:port=primary=5432",
			"--cache:url=redis=redis://cache.example.com:6379",
			"--log:level=app=debug",
		}
		if err := fs.Parse(args); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dbHost != "db1.example.com" {
			t.Errorf("db:host=primary = %q", dbHost)
		}
		if dbPort != "5432" {
			t.Errorf("db:port=primary = %q", dbPort)
		}
		if cacheUrl != "redis://cache.example.com:6379" {
			t.Errorf("cache:url=redis = %q", cacheUrl)
		}
		if logLevel != "debug" {
			t.Errorf("log:level=app = %q", logLevel)
		}
	})
}

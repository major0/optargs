package pflags

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// TestFlagSetCreation tests basic FlagSet creation and initialization.
func TestFlagSetCreation(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	if fs.Name() != "test" {
		t.Errorf("Name = %q, want %q", fs.Name(), "test")
	}
	if fs.Parsed() {
		t.Error("Parsed() should be false for new FlagSet")
	}
	if fs.NArg() != 0 {
		t.Errorf("NArg = %d, want 0", fs.NArg())
	}
}

// TestFlagCreationAllTypes tests flag creation for every value type.
// Two rows per type: Var (destination pointer) and P (pointer-returning + shorthand).
// This covers the full call chain: Var→VarP→addFlag, P→VarP→addFlag.
func TestFlagCreationAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		register func(fs *FlagSet)
		flag     string
		defValue string
		typeName string
	}{
		// Scalars
		{"StringVar", func(fs *FlagSet) { fs.StringVar(new(string), "f", "abc", "u") }, "f", "abc", "string"},
		{"StringP", func(fs *FlagSet) { fs.StringP("f", "s", "abc", "u") }, "f", "abc", "string"},
		{"BoolVar", func(fs *FlagSet) { fs.BoolVar(new(bool), "f", true, "u") }, "f", "true", "bool"},
		{"BoolP", func(fs *FlagSet) { fs.BoolP("f", "b", true, "u") }, "f", "true", "bool"},
		{"IntVar", func(fs *FlagSet) { fs.IntVar(new(int), "f", 42, "u") }, "f", "42", "int"},
		{"IntP", func(fs *FlagSet) { fs.IntP("f", "i", 42, "u") }, "f", "42", "int"},
		{"Int64Var", func(fs *FlagSet) { fs.Int64Var(new(int64), "f", 99, "u") }, "f", "99", "int64"},
		{"Int64P", func(fs *FlagSet) { fs.Int64P("f", "l", 99, "u") }, "f", "99", "int64"},
		{"UintVar", func(fs *FlagSet) { fs.UintVar(new(uint), "f", 7, "u") }, "f", "7", "uint"},
		{"UintP", func(fs *FlagSet) { fs.UintP("f", "u", 7, "u") }, "f", "7", "uint"},
		{"Uint64Var", func(fs *FlagSet) { fs.Uint64Var(new(uint64), "f", 8, "u") }, "f", "8", "uint64"},
		{"Uint64P", func(fs *FlagSet) { fs.Uint64P("f", "x", 8, "u") }, "f", "8", "uint64"},
		{"Float64Var", func(fs *FlagSet) { fs.Float64Var(new(float64), "f", 3.14, "u") }, "f", "3.14", "float64"},
		{"Float64P", func(fs *FlagSet) { fs.Float64P("f", "g", 3.14, "u") }, "f", "3.14", "float64"},
		{"DurationVar", func(fs *FlagSet) { fs.DurationVar(new(time.Duration), "f", 5*time.Second, "u") }, "f", "5s", "duration"},
		{"DurationP", func(fs *FlagSet) { fs.DurationP("f", "d", 5*time.Second, "u") }, "f", "5s", "duration"},
		// Narrow numeric types
		{"Int8Var", func(fs *FlagSet) { fs.Int8Var(new(int8), "f", 7, "u") }, "f", "7", "int8"},
		{"Int8P", func(fs *FlagSet) { fs.Int8P("f", "i", 7, "u") }, "f", "7", "int8"},
		{"Int16Var", func(fs *FlagSet) { fs.Int16Var(new(int16), "f", 16, "u") }, "f", "16", "int16"},
		{"Int16P", func(fs *FlagSet) { fs.Int16P("f", "i", 16, "u") }, "f", "16", "int16"},
		{"Int32Var", func(fs *FlagSet) { fs.Int32Var(new(int32), "f", 32, "u") }, "f", "32", "int32"},
		{"Int32P", func(fs *FlagSet) { fs.Int32P("f", "i", 32, "u") }, "f", "32", "int32"},
		{"Uint8Var", func(fs *FlagSet) { fs.Uint8Var(new(uint8), "f", 8, "u") }, "f", "8", "uint8"},
		{"Uint8P", func(fs *FlagSet) { fs.Uint8P("f", "u", 8, "u") }, "f", "8", "uint8"},
		{"Uint16Var", func(fs *FlagSet) { fs.Uint16Var(new(uint16), "f", 16, "u") }, "f", "16", "uint16"},
		{"Uint16P", func(fs *FlagSet) { fs.Uint16P("f", "u", 16, "u") }, "f", "16", "uint16"},
		{"Uint32Var", func(fs *FlagSet) { fs.Uint32Var(new(uint32), "f", 32, "u") }, "f", "32", "uint32"},
		{"Uint32P", func(fs *FlagSet) { fs.Uint32P("f", "u", 32, "u") }, "f", "32", "uint32"},
		{"Float32Var", func(fs *FlagSet) { fs.Float32Var(new(float32), "f", 1.5, "u") }, "f", "1.5", "float32"},
		{"Float32P", func(fs *FlagSet) { fs.Float32P("f", "g", 1.5, "u") }, "f", "1.5", "float32"},
		// Slices
		{"StringSliceVar", func(fs *FlagSet) { fs.StringSliceVar(new([]string), "f", []string{"a"}, "u") }, "f", "[a]", "stringSlice"},
		{"StringSliceP", func(fs *FlagSet) { fs.StringSliceP("f", "s", []string{"a"}, "u") }, "f", "[a]", "stringSlice"},
		{"IntSliceVar", func(fs *FlagSet) { fs.IntSliceVar(new([]int), "f", []int{1}, "u") }, "f", "[1]", "intSlice"},
		{"IntSliceP", func(fs *FlagSet) { fs.IntSliceP("f", "n", []int{1}, "u") }, "f", "[1]", "intSlice"},
		{"BoolSliceVar", func(fs *FlagSet) { fs.BoolSliceVar(new([]bool), "f", nil, "u") }, "f", "[]", "boolSlice"},
		{"BoolSliceP", func(fs *FlagSet) { fs.BoolSliceP("f", "b", nil, "u") }, "f", "[]", "boolSlice"},
		{"Int32SliceVar", func(fs *FlagSet) { fs.Int32SliceVar(new([]int32), "f", nil, "u") }, "f", "[]", "int32Slice"},
		{"Int64SliceVar", func(fs *FlagSet) { fs.Int64SliceVar(new([]int64), "f", nil, "u") }, "f", "[]", "int64Slice"},
		{"UintSliceVar", func(fs *FlagSet) { fs.UintSliceVar(new([]uint), "f", nil, "u") }, "f", "[]", "uintSlice"},
		{"Float32SliceVar", func(fs *FlagSet) { fs.Float32SliceVar(new([]float32), "f", nil, "u") }, "f", "[]", "float32Slice"},
		{"Float64SliceVar", func(fs *FlagSet) { fs.Float64SliceVar(new([]float64), "f", nil, "u") }, "f", "[]", "float64Slice"},
		{"DurationSliceVar", func(fs *FlagSet) { fs.DurationSliceVar(new([]time.Duration), "f", nil, "u") }, "f", "[]", "durationSlice"},
		// String collections and maps
		{"StringArrayVar", func(fs *FlagSet) { fs.StringArrayVar(new([]string), "f", nil, "u") }, "f", "[]", "stringArray"},
		{"StringArrayP", func(fs *FlagSet) { fs.StringArrayP("f", "a", nil, "u") }, "f", "[]", "stringArray"},
		{"StringToStringVar", func(fs *FlagSet) { fs.StringToStringVar(new(map[string]string), "f", nil, "u") }, "f", "map[]", "stringToString"},
		{"StringToStringP", func(fs *FlagSet) { fs.StringToStringP("f", "s", nil, "u") }, "f", "map[]", "stringToString"},
		{"StringToIntVar", func(fs *FlagSet) { fs.StringToIntVar(new(map[string]int), "f", nil, "u") }, "f", "map[]", "stringToInt"},
		{"StringToIntP", func(fs *FlagSet) { fs.StringToIntP("f", "i", nil, "u") }, "f", "map[]", "stringToInt"},
		{"StringToInt64Var", func(fs *FlagSet) { fs.StringToInt64Var(new(map[string]int64), "f", nil, "u") }, "f", "map[]", "stringToInt64"},
		{"StringToInt64P", func(fs *FlagSet) { fs.StringToInt64P("f", "l", nil, "u") }, "f", "map[]", "stringToInt64"},
		// Specialized types
		{"CountVar", func(fs *FlagSet) { fs.CountVar(new(int), "f", "u") }, "f", "0", "count"},
		{"CountP", func(fs *FlagSet) { fs.CountP("f", "c", "u") }, "f", "0", "count"},
		{"IPVar", func(fs *FlagSet) { fs.IPVar(new(net.IP), "f", nil, "u") }, "f", "", "textUnmarshaler"},
		{"IPP", func(fs *FlagSet) { fs.IPP("f", "i", nil, "u") }, "f", "", "textUnmarshaler"},
		{"IPMaskVar", func(fs *FlagSet) { fs.IPMaskVar(new(net.IPMask), "f", nil, "u") }, "f", "<nil>", "ipMask"},
		{"IPMaskP", func(fs *FlagSet) { fs.IPMaskP("f", "m", nil, "u") }, "f", "<nil>", "ipMask"},
		{"IPNetVar", func(fs *FlagSet) { fs.IPNetVar(new(net.IPNet), "f", net.IPNet{}, "u") }, "f", "<nil>", "ipNet"},
		{"IPNetP", func(fs *FlagSet) { fs.IPNetP("f", "n", net.IPNet{}, "u") }, "f", "<nil>", "ipNet"},
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

// TestIsBoolFlag verifies the boolValue.IsBoolFlag() method.
func TestIsBoolFlag(t *testing.T) {
	bv := newBoolValue(false, new(bool))
	type boolFlagger interface{ IsBoolFlag() bool }
	bf, ok := bv.(boolFlagger)
	if !ok {
		t.Fatal("bool value should implement IsBoolFlag()")
	}
	if !bf.IsBoolFlag() {
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
	fs.StringVarP(new(string), "verbose", "v", "", "")
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic due to shorthand conflict")
		}
	}()
	fs.StringVarP(new(string), "version", "v", "", "")
}

// TestFlagSetIsolation tests that different FlagSets are isolated.
func TestFlagSetIsolation(t *testing.T) {
	fs1 := NewFlagSet("test1", ContinueOnError)
	fs2 := NewFlagSet("test2", ContinueOnError)
	fs1.StringVar(new(string), "flag", "v1", "")
	fs2.StringVar(new(string), "flag", "v2", "")
	if fs1.Lookup("flag").DefValue != "v1" {
		t.Error("fs1 flag default wrong")
	}
	if fs2.Lookup("flag").DefValue != "v2" {
		t.Error("fs2 flag default wrong")
	}
}

// TestBooleanFlagParsing tests boolean flag parsing: no-arg, explicit values,
// negation syntax. Table-driven — the authoritative form for this invariant.
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
		{"shorthand", []string{"-v"}, false, true, false},
		{"invalid value", []string{"--verbose=invalid"}, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var v bool
			fs.BoolVarP(&v, "verbose", "v", tt.defaultVal, "")
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
		})
	}
}

// TestNegationHandlerBranches exercises makeNegationHandler branches.
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
	if err := fs.Parse([]string{"--name", "val", "pos1", "pos2"}); err != nil {
		t.Fatal(err)
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
	if fs.Arg(0) != "pos1" || fs.Arg(1) != "pos2" {
		t.Errorf("Args() = %v, want [pos1 pos2]", fs.Args())
	}
	if fs.Arg(-1) != "" || fs.Arg(99) != "" {
		t.Error("out-of-bounds Arg should return empty")
	}
}

// TestVisitAllAndVisit tests VisitAll and Visit behavior.
func TestVisitAllAndVisit(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "string", "", "")
	fs.IntVar(new(int), "int", 0, "")
	fs.BoolVar(new(bool), "bool", false, "")

	allNames := map[string]bool{}
	fs.VisitAll(func(f *Flag) { allNames[f.Name] = true })
	if len(allNames) != 3 {
		t.Errorf("VisitAll visited %d flags, want 3", len(allNames))
	}

	changedNames := map[string]bool{}
	fs.Visit(func(f *Flag) { changedNames[f.Name] = true })
	if len(changedNames) != 0 {
		t.Errorf("Visit visited %d flags before parse, want 0", len(changedNames))
	}

	if err := fs.Set("string", "hello"); err != nil {
		t.Fatal(err)
	}
	changedNames = map[string]bool{}
	fs.Visit(func(f *Flag) { changedNames[f.Name] = true })
	if !changedNames["string"] || len(changedNames) != 1 {
		t.Errorf("Visit after Set: got %v, want {string}", changedNames)
	}
}

// TestSetOutput tests SetOutput and out() behavior.
func TestSetOutput(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	if fs.out() == nil {
		t.Error("out() should not return nil")
	}
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
	if err == nil || !strings.Contains(err.Error(), "no such flag") {
		t.Errorf("expected 'no such flag' error, got: %v", err)
	}
}

// TestCustomValue tests custom Value interface integration.
type customValue struct{ value string }

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
	if f.Value.Type() != "custom" || f.DefValue != "initial" {
		t.Errorf("Type=%q DefValue=%q", f.Value.Type(), f.DefValue)
	}
}

// TestPrintDefaults tests PrintDefaults output for various flag configurations.
func TestPrintDefaults(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.StringVarP(new(string), "output", "o", "file.txt", "output `filename`")
	fs.BoolVar(new(bool), "verbose", false, "enable verbose")
	fs.IntVar(new(int), "count", 0, "number of items")
	if f := fs.Lookup("count"); f != nil {
		f.Hidden = true
	}
	fs.PrintDefaults()
	out := buf.String()
	if !strings.Contains(out, "-o, --output") {
		t.Errorf("missing shorthand format in:\n%s", out)
	}
	if !strings.Contains(out, "filename") {
		t.Errorf("missing unquoted usage name in:\n%s", out)
	}
	if !strings.Contains(out, "(default") {
		t.Errorf("missing default value in:\n%s", out)
	}
	if strings.Contains(out, "count") {
		t.Errorf("hidden flag 'count' should not appear in:\n%s", out)
	}
	// FlagUsages should return the same content
	usages := fs.FlagUsages()
	if usages != out {
		t.Errorf("FlagUsages differs from PrintDefaults")
	}
}

// TestDefaultUsage tests the defaultUsage function.
func TestDefaultUsage(t *testing.T) {
	fs := NewFlagSet("myapp", ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.BoolVar(new(bool), "help", false, "show help")
	fs.defaultUsage()
	if !strings.Contains(buf.String(), "Usage of myapp:") {
		t.Errorf("expected 'Usage of myapp:', got:\n%s", buf.String())
	}

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
		name, wantName, wantUsage string
		flag                      *Flag
	}{
		{"backtick", "filename", "output filename to write", &Flag{Usage: "output `filename` to write", Value: newStringValue("", new(string))}},
		{"string", "string", "output file", &Flag{Usage: "output file", Value: newStringValue("", new(string))}},
		{"bool", "", "enable verbose", &Flag{Usage: "enable verbose", Value: newBoolValue(false, new(bool))}},
		{"float64", "float", "rate limit", &Flag{Usage: "rate limit", Value: newFloat64Value(0, new(float64))}},
		{"int64", "int", "max size", &Flag{Usage: "max size", Value: newInt64Value(0, new(int64))}},
		{"uint64", "uint", "max count", &Flag{Usage: "max count", Value: newUint64Value(0, new(uint64))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, usage := UnquoteUsage(tt.flag)
			if name != tt.wantName || usage != tt.wantUsage {
				t.Errorf("got (%q, %q), want (%q, %q)", name, usage, tt.wantName, tt.wantUsage)
			}
		})
	}
}

// TestIsZeroValue tests isZeroValue for known types.
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
			if got := isZeroValue(tt.flag, tt.value); got != tt.wantZero {
				t.Errorf("isZeroValue(%q) = %t, want %t", tt.value, got, tt.wantZero)
			}
		})
	}
}

// TestTranslateError tests error translation from OptArgs Core to pflag format.
func TestTranslateError(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"nil", "", ""},
		{"unknown long", "unknown option: verbose", "unknown flag: --verbose"},
		{"unknown short", "unknown option: v", "unknown shorthand flag: 'v' in -v"},
		{"unknown bare", "unknown option", "unknown flag: --"},
		{"requires arg long", "option requires an argument: output", "flag needs an argument: --output"},
		{"requires arg short", "option requires an argument: o", "flag needs an argument: 'o' in -o"},
		{"requires arg bare", "option requires an argument", "flag needs an argument: --"},
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
			if got == nil || got.Error() != tt.want {
				t.Errorf("got %v, want %q", got, tt.want)
			}
		})
	}
}

// TestAdvancedGNULongestMatching tests GNU getopt_long() longest matching behavior.
func TestAdvancedGNULongestMatching(t *testing.T) {
	tests := []struct {
		name, expectedFlag, expectedValue string
		flagDefs                          map[string]string
		args                              []string
	}{
		{"longest_match", "enable-bobadufoo", "test-value", map[string]string{"enable-bob": "", "enable-bobadufoo": ""}, []string{"--enable-bobadufoo", "test-value"}},
		{"shorter_exact", "enable-bob", "test-value", map[string]string{"enable-bob": "", "enable-bobadufoo": ""}, []string{"--enable-bob", "test-value"}},
		{"equals_syntax", "system7-ex", "extended-value", map[string]string{"system": "", "system7": "", "system7-ex": ""}, []string{"--system7-ex=extended-value"}},
		{"prefix_disambig", "verbose-mode", "detailed", map[string]string{"verbose": "", "verbose-mode": "", "verb": ""}, []string{"--verbose-mode", "detailed"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			flagVars := make(map[string]*string)
			for name := range tt.flagDefs {
				v := new(string)
				flagVars[name] = v
				fs.StringVar(v, name, "", "")
			}
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if got := *flagVars[tt.expectedFlag]; got != tt.expectedValue {
				t.Errorf("flag %s = %q, want %q", tt.expectedFlag, got, tt.expectedValue)
			}
		})
	}
}

// TestAdvancedGNUSpecialCharacters tests special characters in option names.
func TestAdvancedGNUSpecialCharacters(t *testing.T) {
	tests := []struct {
		name, flagName, expectedValue string
		args                          []string
	}{
		{"colon_space", "system7:verbose", "enabled", []string{"--system7:verbose", "enabled"}},
		{"colon_equals", "system7:prefix", "/my/path", []string{"--system7:prefix=/my/path"}},
		{"nested_equals", "system7:path=bindir", "/usr/bin", []string{"--system7:path=bindir=/usr/bin"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var v string
			fs.StringVar(&v, tt.flagName, "", "")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
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
		name, expectedFlag string
		flagDefs           []string
		args               []string
		shouldError        bool
	}{
		{"partial_rejected", "", []string{"verbose", "version"}, []string{"--ver", "val"}, true},
		{"exact_match", "help", []string{"help", "help-extended"}, []string{"--help", "val"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			flagVars := make(map[string]*string)
			for _, name := range tt.flagDefs {
				v := new(string)
				flagVars[name] = v
				fs.StringVar(v, name, "", "")
			}
			err := fs.Parse(tt.args)
			if tt.shouldError {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := *flagVars[tt.expectedFlag]; got != "val" {
				t.Errorf("flag %s = %q, want %q", tt.expectedFlag, got, "val")
			}
		})
	}
}

// TestPOSIXDoubleHyphenTermination tests that -- terminates option processing.
func TestPOSIXDoubleHyphenTermination(t *testing.T) {
	tests := []struct {
		name, wantFlag string
		args           []string
		wantArgs       []string
	}{
		{"stops parsing", "val", []string{"--name", "val", "--", "--other", "pos"}, []string{"--other", "pos"}},
		{"only", "", []string{"--", "a", "b"}, []string{"a", "b"}},
		{"no dash", "val", []string{"--name", "val", "pos"}, []string{"pos"}},
		{"no trailing", "val", []string{"--name", "val", "--"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var name string
			fs.StringVar(&name, "name", "", "")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if name != tt.wantFlag {
				t.Errorf("flag = %q, want %q", name, tt.wantFlag)
			}
			args := fs.Args()
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("Args() = %v, want %v", args, tt.wantArgs)
			}
			for i, a := range tt.wantArgs {
				if args[i] != a {
					t.Errorf("Arg(%d) = %q, want %q", i, args[i], a)
				}
			}
		})
	}
}

// TestPOSIXCombinedShortOptions tests -abc style combined short options.
func TestPOSIXCombinedShortOptions(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantA, wantB, wantC bool
		wantVal         string
	}{
		{"combined booleans", []string{"-abc"}, true, true, true, ""},
		{"combined with trailing value", []string{"-abo", "file.txt"}, true, true, false, "file.txt"},
		{"individual flags", []string{"-a", "-b", "-c"}, true, true, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var a, b, c bool
			var o string
			fs.BoolVarP(&a, "alpha", "a", false, "")
			fs.BoolVarP(&b, "beta", "b", false, "")
			fs.BoolVarP(&c, "gamma", "c", false, "")
			fs.StringVarP(&o, "output", "o", "", "")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if a != tt.wantA || b != tt.wantB || c != tt.wantC || o != tt.wantVal {
				t.Errorf("got a=%t b=%t c=%t o=%q", a, b, c, o)
			}
		})
	}
}

// TestErrorHandlingPanicOnError tests that PanicOnError panics on parse failure.
func TestErrorHandlingPanicOnError(t *testing.T) {
	fs := NewFlagSet("test", PanicOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.StringVar(new(string), "known", "", "")
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for PanicOnError")
		}
		if err, ok := r.(error); !ok || !strings.Contains(err.Error(), "unknown flag") {
			t.Errorf("panic = %v", r)
		}
	}()
	fs.Parse([]string{"--unknown"}) //nolint:errcheck
}

// TestErrorHandlingContinueOnError tests that ContinueOnError returns the error.
func TestErrorHandlingContinueOnError(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "known", "", "")
	err := fs.Parse([]string{"--unknown"})
	if err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("expected 'unknown flag' error, got: %v", err)
	}
}

// TestShortOnlyFlags tests short-only flag registration and parsing.
func TestShortOnlyFlags(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		wantX bool
		wantY string
	}{
		{"single bool", []string{"-x"}, true, ""},
		{"single string", []string{"-y", "val"}, false, "val"},
		{"compacted", []string{"-xy", "val"}, true, "val"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var x bool
			var y string
			fs.ShortVar(newBoolValue(false, &x), "x", "extract")
			fs.ShortVar(newStringValue("", &y), "y", "output")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if x != tt.wantX || y != tt.wantY {
				t.Errorf("got x=%t y=%q", x, y)
			}
		})
	}
}

// TestShortOnlyConflict tests that short-only flags conflict with existing shorthands.
func TestShortOnlyConflict(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.BoolVarP(new(bool), "verbose", "v", false, "")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for shorthand conflict")
		}
	}()
	fs.ShortVar(newBoolValue(false, new(bool)), "v", "")
}

// TestShortOnlyNotAccessibleByLongName tests that short-only flags
// are not accessible via --x long option syntax.
func TestShortOnlyNotAccessibleByLongName(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.ShortVar(newBoolValue(false, new(bool)), "x", "extract")
	if err := fs.Parse([]string{"--x"}); err == nil {
		t.Error("expected error for --x on short-only flag")
	}
}

// fixedValue is a Value that always sets a fixed string to a destination.
type fixedValue struct {
	dest  *string
	fixed string
}

func (v *fixedValue) String() string   { return *v.dest }
func (v *fixedValue) Set(string) error { *v.dest = v.fixed; return nil }
func (v *fixedValue) Type() string     { return "string" }
func (v *fixedValue) IsBoolFlag() bool { return true }

// TestManyToOneMapping tests the ls --format=across / -x pattern.
func TestManyToOneMapping(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"long primary", []string{"--format", "across"}, "across"},
		{"short alias x", []string{"-x"}, "across"},
		{"short alias C", []string{"-C"}, "columns"},
		{"last wins", []string{"-x", "-C"}, "columns"},
		{"long after short", []string{"-x", "--format", "long"}, "long"},
		{"short after long", []string{"--format", "long", "-x"}, "across"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var format string
			fs.Var(newStringValue("", &format), "format", "output format")
			fs.AliasShortVar(&fixedValue{dest: &format, fixed: "across"}, "x")
			fs.AliasShortVar(&fixedValue{dest: &format, fixed: "columns"}, "C")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if format != tt.want {
				t.Errorf("format = %q, want %q", format, tt.want)
			}
		})
	}
}

// TestManyToOneHelpText tests that alias flags are hidden from help output.
func TestManyToOneHelpText(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var format string
	fs.Var(newStringValue("", &format), "format", "output format")
	fs.AliasShortVar(&fixedValue{dest: &format, fixed: "across"}, "x")
	usages := fs.FlagUsages()
	if !strings.Contains(usages, "--format") {
		t.Errorf("primary flag missing from help:\n%s", usages)
	}
	if strings.Contains(usages, "-x") {
		t.Errorf("alias -x should be hidden from help:\n%s", usages)
	}
}

// TestLongOnlyMode tests getopt_long_only(3) behavior.
func TestLongOnlyMode(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		wantS string
		wantV bool
	}{
		{"single-dash long", []string{"-output", "file.txt"}, "file.txt", false},
		{"double-dash long", []string{"--output", "file.txt"}, "file.txt", false},
		{"single-dash bool", []string{"-verbose"}, "", true},
		{"short fallback", []string{"-v"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			fs.SetLongOnly(true)
			var s string
			var v bool
			fs.StringVar(&s, "output", "", "")
			fs.BoolVarP(&v, "verbose", "v", false, "")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if s != tt.wantS || v != tt.wantV {
				t.Errorf("got s=%q v=%t", s, v)
			}
		})
	}
	// Getter
	fs := NewFlagSet("test", ContinueOnError)
	if fs.LongOnly() {
		t.Error("LongOnly should default to false")
	}
	fs.SetLongOnly(true)
	if !fs.LongOnly() {
		t.Error("LongOnly should be true after SetLongOnly(true)")
	}
}

// TestChanged tests the Changed() method.
func TestChanged(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "name", "", "")
	fs.BoolVar(new(bool), "verbose", false, "")
	if fs.Changed("name") {
		t.Error("name should not be changed before parse")
	}
	if err := fs.Parse([]string{"--name", "val"}); err != nil {
		t.Fatal(err)
	}
	if !fs.Changed("name") {
		t.Error("name should be changed after parse")
	}
	if fs.Changed("verbose") || fs.Changed("nonexistent") {
		t.Error("unset/nonexistent flags should not be changed")
	}
}

// TestNFlag tests the NFlag() method.
func TestNFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "a", "", "")
	fs.StringVar(new(string), "b", "", "")
	fs.StringVar(new(string), "c", "", "")
	if fs.NFlag() != 0 {
		t.Errorf("NFlag before parse = %d, want 0", fs.NFlag())
	}
	if err := fs.Parse([]string{"--a", "1", "--c", "3"}); err != nil {
		t.Fatal(err)
	}
	if fs.NFlag() != 2 {
		t.Errorf("NFlag = %d, want 2", fs.NFlag())
	}
}

// TestHasFlags tests HasFlags() and HasAvailableFlags().
func TestHasFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	if fs.HasFlags() || fs.HasAvailableFlags() {
		t.Error("empty FlagSet should not have flags")
	}
	fs.StringVar(new(string), "name", "", "")
	if !fs.HasFlags() || !fs.HasAvailableFlags() {
		t.Error("should have flags after registration")
	}
	fs.Lookup("name").Hidden = true
	if !fs.HasFlags() {
		t.Error("should still have flags (hidden counts)")
	}
	if fs.HasAvailableFlags() {
		t.Error("should not have available flags (all hidden)")
	}
}

// TestOutput tests the Output() getter.
func TestOutput(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	if fs.Output() == nil {
		t.Error("Output() should not return nil")
	}
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	if fs.Output() != &buf {
		t.Error("Output() should return custom writer")
	}
}

// TestShorthandLookup tests the ShorthandLookup() method.
func TestShorthandLookup(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVarP(new(string), "output", "o", "", "")
	fs.ShortVar(newBoolValue(false, new(bool)), "x", "extract")
	if f := fs.ShorthandLookup("o"); f == nil || f.Name != "output" {
		t.Errorf("ShorthandLookup('o') = %v", f)
	}
	if f := fs.ShorthandLookup("x"); f == nil || f.Name != "x" {
		t.Errorf("ShorthandLookup('x') = %v", f)
	}
	if fs.ShorthandLookup("z") != nil {
		t.Error("ShorthandLookup('z') should return nil")
	}
}

// TestShorthandLookupPanic tests that ShorthandLookup panics for multi-char input.
func TestShorthandLookupPanic(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for multi-char ShorthandLookup")
		}
	}()
	fs.ShorthandLookup("ab")
}

// TestInit tests the Init() method.
func TestInit(t *testing.T) {
	fs := NewFlagSet("old", ContinueOnError)
	fs.Init("new", PanicOnError)
	if fs.Name() != "new" {
		t.Errorf("Name() = %q, want %q", fs.Name(), "new")
	}
}

// TestVarPF tests that VarPF returns the created flag.
func TestVarPF(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	f := fs.VarPF(newStringValue("default", new(string)), "output", "o", "output file")
	if f == nil || f.Name != "output" || f.Shorthand != "o" {
		t.Errorf("VarPF returned %v", f)
	}
}

// TestArgsLenAtDash tests the ArgsLenAtDash() method.
func TestArgsLenAtDash(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"no dash", []string{"--name", "val", "pos"}, -1},
		{"dash with args before", []string{"--name", "val", "pos1", "--", "pos2"}, 1},
		{"dash no args before", []string{"--name", "val", "--", "pos1"}, 0},
		{"dash only", []string{"--"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			fs.StringVar(new(string), "name", "", "")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if got := fs.ArgsLenAtDash(); got != tt.want {
				t.Errorf("ArgsLenAtDash() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestSetNormalizeFunc tests flag name normalization.
func TestSetNormalizeFunc(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.SetNormalizeFunc(func(f *FlagSet, name string) NormalizedName {
		return NormalizedName(strings.ReplaceAll(name, "_", "-"))
	})
	var s string
	fs.StringVar(&s, "my-flag", "", "")
	if f := fs.Lookup("my_flag"); f == nil {
		t.Error("Lookup('my_flag') should find 'my-flag' via normalization")
	}
	if err := fs.Parse([]string{"--my_flag", "val"}); err != nil {
		t.Fatal(err)
	}
	if s != "val" {
		t.Errorf("flag value = %q, want %q", s, "val")
	}
	if fs.GetNormalizeFunc() == nil {
		t.Error("GetNormalizeFunc() should not be nil")
	}
	// Equals syntax with underscore
	fs2 := NewFlagSet("test2", ContinueOnError)
	fs2.SetNormalizeFunc(func(f *FlagSet, name string) NormalizedName {
		return NormalizedName(strings.ReplaceAll(name, "_", "-"))
	})
	var s2 string
	fs2.StringVar(&s2, "my-flag", "", "")
	if err := fs2.Parse([]string{"--my_flag=val2"}); err != nil {
		t.Fatal(err)
	}
	if s2 != "val2" {
		t.Errorf("s2 = %q, want %q", s2, "val2")
	}
}

// TestSetInterspersed tests interspersed option/non-option arg handling.
func TestSetInterspersed(t *testing.T) {
	// Default: interspersed enabled (GNU behavior)
	fs1 := NewFlagSet("test", ContinueOnError)
	var v1 string
	fs1.StringVar(&v1, "name", "", "")
	if err := fs1.Parse([]string{"pos1", "--name", "val", "pos2"}); err != nil {
		t.Fatal(err)
	}
	if v1 != "val" || fs1.NArg() != 2 {
		t.Errorf("interspersed: name=%q NArg=%d", v1, fs1.NArg())
	}

	// Disabled: POSIX behavior — stop at first non-option
	fs2 := NewFlagSet("test", ContinueOnError)
	fs2.SetInterspersed(false)
	var v2 string
	fs2.StringVar(&v2, "name", "", "")
	if err := fs2.Parse([]string{"pos1", "--name", "val"}); err != nil {
		t.Fatal(err)
	}
	if v2 != "" || fs2.NArg() != 3 {
		t.Errorf("non-interspersed: name=%q NArg=%d", v2, fs2.NArg())
	}
	if !fs1.GetInterspersed() || fs2.GetInterspersed() {
		t.Error("GetInterspersed mismatch")
	}
}

// TestMarkDeprecated tests the MarkDeprecated method.
func TestMarkDeprecated(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "old-flag", "", "old flag")
	if err := fs.MarkDeprecated("old-flag", "use --new-flag instead"); err != nil {
		t.Fatal(err)
	}
	f := fs.Lookup("old-flag")
	if f.Deprecated != "use --new-flag instead" || !f.Hidden {
		t.Errorf("Deprecated=%q Hidden=%t", f.Deprecated, f.Hidden)
	}
	if err := fs.MarkDeprecated("nope", "msg"); err == nil {
		t.Error("expected error for non-existent flag")
	}
	if err := fs.MarkDeprecated("old-flag", ""); err == nil {
		t.Error("expected error for empty deprecation message")
	}
}

// TestMarkHidden tests the MarkHidden method.
func TestMarkHidden(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "internal", "", "")
	if err := fs.MarkHidden("internal"); err != nil {
		t.Fatal(err)
	}
	if !fs.Lookup("internal").Hidden {
		t.Error("flag should be hidden")
	}
	if err := fs.MarkHidden("nope"); err == nil {
		t.Error("expected error for non-existent flag")
	}
}

// TestMarkShorthandDeprecated tests the MarkShorthandDeprecated method.
func TestMarkShorthandDeprecated(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVarP(new(string), "output", "o", "", "")
	if err := fs.MarkShorthandDeprecated("output", "use --output instead"); err != nil {
		t.Fatal(err)
	}
	if fs.Lookup("output").ShorthandDeprecated != "use --output instead" {
		t.Error("ShorthandDeprecated not set")
	}
	if err := fs.MarkShorthandDeprecated("nope", "msg"); err == nil {
		t.Error("expected error for non-existent flag")
	}
	if err := fs.MarkShorthandDeprecated("output", ""); err == nil {
		t.Error("expected error for empty message")
	}
}

// TestSetAnnotation tests the SetAnnotation method.
func TestSetAnnotation(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "output", "", "")
	if err := fs.SetAnnotation("output", "cobra_key", []string{"true"}); err != nil {
		t.Fatal(err)
	}
	vals := fs.Lookup("output").Annotations["cobra_key"]
	if len(vals) != 1 || vals[0] != "true" {
		t.Errorf("annotation = %v", vals)
	}
	if err := fs.SetAnnotation("nope", "key", nil); err == nil {
		t.Error("expected error for non-existent flag")
	}
}

// TestAddFlag tests adding a single flag to a FlagSet.
func TestAddFlag(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	flag := &Flag{Name: "output", Usage: "output file", Value: newStringValue("default", new(string)), DefValue: "default"}
	fs.AddFlag(flag)
	if f := fs.Lookup("output"); f == nil {
		t.Error("flag not found after AddFlag")
	}
	// Duplicate should be silently ignored
	fs.AddFlag(&Flag{Name: "output", Usage: "different", Value: newStringValue("other", new(string)), DefValue: "other"})
	if fs.Lookup("output").Usage != "output file" {
		t.Error("duplicate AddFlag should not overwrite")
	}
}

// TestAddFlagSet tests merging two FlagSets.
func TestAddFlagSet(t *testing.T) {
	fs1 := NewFlagSet("parent", ContinueOnError)
	fs1.StringVar(new(string), "verbose", "", "verbose output")
	fs1.BoolVarP(new(bool), "debug", "d", false, "debug mode")

	fs2 := NewFlagSet("child", ContinueOnError)
	fs2.StringVar(new(string), "output", "", "output file")
	fs2.StringVar(new(string), "verbose", "", "child verbose") // duplicate

	fs1.AddFlagSet(fs2)
	if fs1.Lookup("output") == nil {
		t.Error("output flag should be added from child")
	}
	if fs1.Lookup("verbose").Usage != "verbose output" {
		t.Error("duplicate should keep parent's flag")
	}
	fs1.AddFlagSet(nil) // should not panic

	// Short-only flags should be merged
	fs3 := NewFlagSet("source", ContinueOnError)
	fs3.ShortVar(newBoolValue(false, new(bool)), "x", "extract")
	fs4 := NewFlagSet("target", ContinueOnError)
	fs4.AddFlagSet(fs3)
	if fs4.ShorthandLookup("x") == nil {
		t.Error("short-only flag 'x' should be merged")
	}
}

// TestCallbackFlags tests Func, FuncP, BoolFunc, BoolFuncP.
func TestCallbackFlags(t *testing.T) {
	t.Run("Func", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var calls []string
		fs.Func("hook", "callback", func(val string) error { calls = append(calls, val); return nil })
		if err := fs.Parse([]string{"--hook", "a", "--hook", "b"}); err != nil {
			t.Fatal(err)
		}
		if len(calls) != 2 || calls[0] != "a" || calls[1] != "b" {
			t.Errorf("calls = %v", calls)
		}
	})
	t.Run("FuncP", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var got string
		fs.FuncP("hook", "h", "callback", func(val string) error { got = val; return nil })
		if err := fs.Parse([]string{"-h", "val"}); err != nil {
			t.Fatal(err)
		}
		if got != "val" {
			t.Errorf("got = %q", got)
		}
	})
	t.Run("BoolFunc", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var count int
		fs.BoolFunc("verbose", "inc", func(string) error { count++; return nil })
		if err := fs.Parse([]string{"--verbose"}); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("count = %d", count)
		}
	})
	t.Run("BoolFuncP", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var count int
		fs.BoolFuncP("verbose", "v", "inc", func(string) error { count++; return nil })
		if err := fs.Parse([]string{"-v"}); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("count = %d", count)
		}
	})
}

// TestParseAll tests the ParseAll callback mechanism.
func TestParseAll(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "name", "", "")
	fs.BoolVar(new(bool), "verbose", false, "")
	var seen []string
	if err := fs.ParseAll([]string{"--name", "val", "--verbose"}, func(flag *Flag, value string) error {
		seen = append(seen, flag.Name+"="+value)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 {
		t.Errorf("seen = %v, want 2 entries", seen)
	}
	// Callback error should propagate
	fs2 := NewFlagSet("test2", ContinueOnError)
	fs2.StringVar(new(string), "name", "", "")
	err := fs2.ParseAll([]string{"--name", "val"}, func(*Flag, string) error { return fmt.Errorf("callback error") })
	if err == nil || !strings.Contains(err.Error(), "callback error") {
		t.Errorf("expected callback error, got: %v", err)
	}
}

// TestSortFlags tests the SortFlags field.
func TestSortFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "zebra", "", "z flag")
	fs.StringVar(new(string), "alpha", "", "a flag")
	sorted := fs.FlagUsages()
	if strings.Index(sorted, "alpha") > strings.Index(sorted, "zebra") {
		t.Errorf("sorted output should have alpha before zebra:\n%s", sorted)
	}

	fs2 := NewFlagSet("test", ContinueOnError)
	fs2.SortFlags = false
	fs2.StringVar(new(string), "zebra", "", "z flag")
	fs2.StringVar(new(string), "alpha", "", "a flag")
	unsorted := fs2.FlagUsages()
	if strings.Index(unsorted, "zebra") > strings.Index(unsorted, "alpha") {
		t.Errorf("unsorted output should have zebra before alpha:\n%s", unsorted)
	}
}

// TestFlagUsagesWrapped tests column wrapping.
func TestFlagUsagesWrapped(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.StringVar(new(string), "output", "default.txt", "this is a very long usage description that should be wrapped at a reasonable column width")
	noWrap := fs.FlagUsagesWrapped(0)
	wrapped := fs.FlagUsagesWrapped(40)
	if strings.Count(wrapped, "\n") <= strings.Count(noWrap, "\n") {
		t.Errorf("wrapped should have more lines than unwrapped")
	}
}

// TestStructuredErrors tests that parse errors return typed error structs.
func TestStructuredErrors(t *testing.T) {
	t.Run("NotExistError", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.StringVar(new(string), "known", "", "")
		err := fs.Parse([]string{"--unknown"})
		var notExist *NotExistError
		if !errors.As(err, &notExist) {
			t.Fatalf("expected *NotExistError, got %T: %v", err, err)
		}
		if notExist.GetSpecifiedName() != "unknown" {
			t.Errorf("GetSpecifiedName() = %q", notExist.GetSpecifiedName())
		}
	})
	t.Run("ValueRequiredError", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.StringVarP(new(string), "output", "o", "", "")
		err := fs.Parse([]string{"-o"})
		var valReq *ValueRequiredError
		if !errors.As(err, &valReq) {
			t.Fatalf("expected *ValueRequiredError, got %T: %v", err, err)
		}
	})
}

// TestParseErrorsAllowlist tests that unknown flags can be ignored.
func TestParseErrorsAllowlist(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.ParseErrorsAllowlist = ParseErrorsAllowlist{UnknownFlags: true}
	fs.StringVar(new(string), "known", "", "")
	if err := fs.Parse([]string{"--known", "val", "--unknown", "pos"}); err != nil {
		t.Fatalf("expected no error with UnknownFlags allowlist, got: %v", err)
	}
}

// TestStringArrayParsing tests StringArray flag behavior (no comma splitting).
func TestStringArrayParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{"single", []string{"--items", "one"}, []string{"one"}},
		{"repeated", []string{"--items", "one", "--items", "two"}, []string{"one", "two"}},
		{"with-comma", []string{"--items", "a,b"}, []string{"a,b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var items []string
			fs.StringArrayVar(&items, "items", nil, "")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if len(items) != len(tt.expected) {
				t.Fatalf("got %v, want %v", items, tt.expected)
			}
			for i, v := range items {
				if v != tt.expected[i] {
					t.Errorf("items[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

// TestMapFlagParsing tests StringToString, StringToInt, StringToInt64 flag behavior.
func TestMapFlagParsing(t *testing.T) {
	t.Run("StringToString", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var m map[string]string
		fs.StringToStringVar(&m, "labels", nil, "")
		if err := fs.Parse([]string{"--labels", "env=prod,tier=web"}); err != nil {
			t.Fatal(err)
		}
		if m["env"] != "prod" || m["tier"] != "web" {
			t.Errorf("got %v", m)
		}
	})
	t.Run("StringToInt", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var m map[string]int
		fs.StringToIntVar(&m, "ports", nil, "")
		if err := fs.Parse([]string{"--ports", "http=80,https=443"}); err != nil {
			t.Fatal(err)
		}
		if m["http"] != 80 || m["https"] != 443 {
			t.Errorf("got %v", m)
		}
	})
	t.Run("StringToInt64", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var m map[string]int64
		fs.StringToInt64Var(&m, "sizes", nil, "")
		if err := fs.Parse([]string{"--sizes", "small=100,large=9999999999"}); err != nil {
			t.Fatal(err)
		}
		if m["small"] != 100 || m["large"] != 9999999999 {
			t.Errorf("got %v", m)
		}
	})
}

// TestCountParsing tests Count flag behavior (increments on each occurrence).
func TestCountParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected int
	}{
		{"single long", []string{"--verbose"}, 1},
		{"double long", []string{"--verbose", "--verbose"}, 2},
		{"short compacted", []string{"-vvv"}, 3},
		{"mixed", []string{"-v", "--verbose", "-v"}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			var count int
			fs.CountVarP(&count, "verbose", "v", "")
			if err := fs.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if count != tt.expected {
				t.Errorf("count = %d, want %d", count, tt.expected)
			}
		})
	}
}

// TestCountNoOptionalArg verifies Count flags don't consume the next argument.
func TestCountNoOptionalArg(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var count int
	fs.CountVarP(&count, "verbose", "v", "")
	if err := fs.Parse([]string{"--verbose", "positional"}); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if fs.NArg() != 1 || fs.Arg(0) != "positional" {
		t.Errorf("args = %v, want [positional]", fs.Args())
	}
}

// TestIPParsing tests IP flag behavior.
func TestIPParsing(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var ip net.IP
	fs.IPVar(&ip, "addr", nil, "")
	if err := fs.Parse([]string{"--addr", "192.168.1.1"}); err != nil {
		t.Fatal(err)
	}
	if ip.String() != "192.168.1.1" {
		t.Errorf("ip = %s, want 192.168.1.1", ip)
	}
}

// TestIPMaskParsing tests IPMask flag behavior.
func TestIPMaskParsing(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var mask net.IPMask
	fs.IPMaskVar(&mask, "mask", nil, "")
	if err := fs.Parse([]string{"--mask", "255.255.255.0"}); err != nil {
		t.Fatal(err)
	}
	if mask.String() != "ffffff00" {
		t.Errorf("mask = %s, want ffffff00", mask)
	}
}

// TestIPNetParsing tests IPNet flag behavior.
func TestIPNetParsing(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var ipnet net.IPNet
	fs.IPNetVar(&ipnet, "cidr", net.IPNet{}, "")
	if err := fs.Parse([]string{"--cidr", "10.0.0.0/8"}); err != nil {
		t.Fatal(err)
	}
	if ipnet.String() != "10.0.0.0/8" {
		t.Errorf("ipnet = %s, want 10.0.0.0/8", ipnet)
	}
}

// TestTypedGetters tests all Get* methods on FlagSet.
func TestTypedGetters(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.BoolVar(new(bool), "b", true, "")
	fs.StringVar(new(string), "s", "hello", "")
	fs.IntVar(new(int), "i", 42, "")
	fs.Int8Var(new(int8), "i8", 7, "")
	fs.Int16Var(new(int16), "i16", 16, "")
	fs.Int32Var(new(int32), "i32", 32, "")
	fs.Int64Var(new(int64), "i64", 64, "")
	fs.UintVar(new(uint), "u", 7, "")
	fs.Uint8Var(new(uint8), "u8", 8, "")
	fs.Uint16Var(new(uint16), "u16", 16, "")
	fs.Uint32Var(new(uint32), "u32", 32, "")
	fs.Uint64Var(new(uint64), "u64", 64, "")
	fs.Float32Var(new(float32), "f32", 1.5, "")
	fs.Float64Var(new(float64), "f64", 3.14, "")
	fs.DurationVar(new(time.Duration), "d", 5*time.Second, "")
	fs.CountVar(new(int), "c", "")
	fs.StringSliceVar(new([]string), "ss", []string{"a", "b"}, "")
	fs.IntSliceVar(new([]int), "is", []int{1, 2}, "")

	if v, err := fs.GetBool("b"); err != nil || v != true {
		t.Errorf("GetBool: %v %v", v, err)
	}
	if v, err := fs.GetString("s"); err != nil || v != "hello" {
		t.Errorf("GetString: %v %v", v, err)
	}
	if v, err := fs.GetInt("i"); err != nil || v != 42 {
		t.Errorf("GetInt: %v %v", v, err)
	}
	if v, err := fs.GetInt8("i8"); err != nil || v != 7 {
		t.Errorf("GetInt8: %v %v", v, err)
	}
	if v, err := fs.GetInt16("i16"); err != nil || v != 16 {
		t.Errorf("GetInt16: %v %v", v, err)
	}
	if v, err := fs.GetInt32("i32"); err != nil || v != 32 {
		t.Errorf("GetInt32: %v %v", v, err)
	}
	if v, err := fs.GetInt64("i64"); err != nil || v != 64 {
		t.Errorf("GetInt64: %v %v", v, err)
	}
	if v, err := fs.GetUint("u"); err != nil || v != 7 {
		t.Errorf("GetUint: %v %v", v, err)
	}
	if v, err := fs.GetUint8("u8"); err != nil || v != 8 {
		t.Errorf("GetUint8: %v %v", v, err)
	}
	if v, err := fs.GetUint16("u16"); err != nil || v != 16 {
		t.Errorf("GetUint16: %v %v", v, err)
	}
	if v, err := fs.GetUint32("u32"); err != nil || v != 32 {
		t.Errorf("GetUint32: %v %v", v, err)
	}
	if v, err := fs.GetUint64("u64"); err != nil || v != 64 {
		t.Errorf("GetUint64: %v %v", v, err)
	}
	if v, err := fs.GetFloat32("f32"); err != nil || v != 1.5 {
		t.Errorf("GetFloat32: %v %v", v, err)
	}
	if v, err := fs.GetFloat64("f64"); err != nil || v != 3.14 {
		t.Errorf("GetFloat64: %v %v", v, err)
	}
	if v, err := fs.GetDuration("d"); err != nil || v != 5*time.Second {
		t.Errorf("GetDuration: %v %v", v, err)
	}
	if v, err := fs.GetCount("c"); err != nil || v != 0 {
		t.Errorf("GetCount: %v %v", v, err)
	}
	if v, err := fs.GetStringSlice("ss"); err != nil || len(v) != 2 {
		t.Errorf("GetStringSlice: %v %v", v, err)
	}
	if v, err := fs.GetIntSlice("is"); err != nil || len(v) != 2 {
		t.Errorf("GetIntSlice: %v %v", v, err)
	}

	// Error cases
	if _, err := fs.GetBool("nonexistent"); err == nil {
		t.Error("expected error for nonexistent flag")
	}
	if _, err := fs.GetBool("s"); err == nil {
		t.Error("expected error for type mismatch")
	}
}

// TestSliceAndMapGetters tests slice and map Get* methods after parsing.
func TestSliceAndMapGetters(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.BoolSliceVar(new([]bool), "bs", nil, "")
	fs.Int32SliceVar(new([]int32), "i32s", nil, "")
	fs.Int64SliceVar(new([]int64), "i64s", nil, "")
	fs.UintSliceVar(new([]uint), "us", nil, "")
	fs.Float32SliceVar(new([]float32), "f32s", nil, "")
	fs.Float64SliceVar(new([]float64), "f64s", nil, "")
	fs.DurationSliceVar(new([]time.Duration), "ds", nil, "")
	fs.StringToStringVar(new(map[string]string), "sts", nil, "")
	fs.StringToIntVar(new(map[string]int), "sti", nil, "")
	fs.StringToInt64Var(new(map[string]int64), "sti64", nil, "")

	if err := fs.Parse([]string{
		"--bs", "true,false",
		"--i32s", "1,2",
		"--i64s", "100,200",
		"--us", "3,4",
		"--f32s", "1.5,2.5",
		"--f64s", "3.14,2.72",
		"--ds", "1s,2m",
		"--sts", "a=b,c=d",
		"--sti", "x=1,y=2",
		"--sti64", "p=100,q=200",
	}); err != nil {
		t.Fatal(err)
	}

	if v, err := fs.GetBoolSlice("bs"); err != nil || len(v) != 2 || v[0] != true || v[1] != false {
		t.Errorf("GetBoolSlice: %v %v", v, err)
	}
	if v, err := fs.GetInt32Slice("i32s"); err != nil || len(v) != 2 || v[0] != 1 || v[1] != 2 {
		t.Errorf("GetInt32Slice: %v %v", v, err)
	}
	if v, err := fs.GetInt64Slice("i64s"); err != nil || len(v) != 2 || v[0] != 100 || v[1] != 200 {
		t.Errorf("GetInt64Slice: %v %v", v, err)
	}
	if v, err := fs.GetUintSlice("us"); err != nil || len(v) != 2 || v[0] != 3 || v[1] != 4 {
		t.Errorf("GetUintSlice: %v %v", v, err)
	}
	if v, err := fs.GetFloat32Slice("f32s"); err != nil || len(v) != 2 {
		t.Errorf("GetFloat32Slice: %v %v", v, err)
	}
	if v, err := fs.GetFloat64Slice("f64s"); err != nil || len(v) != 2 {
		t.Errorf("GetFloat64Slice: %v %v", v, err)
	}
	if v, err := fs.GetDurationSlice("ds"); err != nil || len(v) != 2 || v[0] != time.Second || v[1] != 2*time.Minute {
		t.Errorf("GetDurationSlice: %v %v", v, err)
	}
	if v, err := fs.GetStringToString("sts"); err != nil || v["a"] != "b" || v["c"] != "d" {
		t.Errorf("GetStringToString: %v %v", v, err)
	}
	if v, err := fs.GetStringToInt("sti"); err != nil || v["x"] != 1 || v["y"] != 2 {
		t.Errorf("GetStringToInt: %v %v", v, err)
	}
	if v, err := fs.GetStringToInt64("sti64"); err != nil || v["p"] != 100 || v["q"] != 200 {
		t.Errorf("GetStringToInt64: %v %v", v, err)
	}
}

// TestStructuredErrorAccessors tests all accessor methods on structured error types.
func TestStructuredErrorAccessors(t *testing.T) {
	t.Run("NotExistError/short", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.StringVarP(new(string), "output", "o", "", "")
		err := fs.Parse([]string{"-z"})
		var notExist *NotExistError
		if !errors.As(err, &notExist) {
			t.Fatalf("expected *NotExistError, got %T", err)
		}
		if notExist.GetSpecifiedShortnames() == "" {
			t.Error("GetSpecifiedShortnames should not be empty for short flag error")
		}
	})

	t.Run("InvalidValueError", func(t *testing.T) {
		// Constructed directly — not yet wired into translateError
		inner := fmt.Errorf("bad number")
		e := &InvalidValueError{flag: &Flag{Name: "count"}, value: "abc", err: inner}
		if e.GetFlag().Name != "count" {
			t.Errorf("GetFlag().Name = %q", e.GetFlag().Name)
		}
		if e.GetValue() != "abc" {
			t.Errorf("GetValue() = %q", e.GetValue())
		}
		if e.Unwrap() != inner {
			t.Error("Unwrap() should return inner error")
		}
		if !strings.Contains(e.Error(), "count") {
			t.Errorf("Error() = %q", e.Error())
		}
	})

	t.Run("ValueRequiredError/accessors", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.StringVarP(new(string), "output", "o", "", "")
		err := fs.Parse([]string{"-o"})
		var valReq *ValueRequiredError
		if !errors.As(err, &valReq) {
			t.Fatalf("expected *ValueRequiredError, got %T", err)
		}
		// GetFlag may be nil (translateError doesn't have FlagSet access)
		_ = valReq.GetFlag()
		if valReq.GetSpecifiedName() == "" {
			t.Error("GetSpecifiedName should not be empty")
		}
		if valReq.GetSpecifiedShortnames() == "" {
			t.Error("GetSpecifiedShortnames should not be empty for short flag")
		}
	})

	t.Run("InvalidSyntaxError", func(t *testing.T) {
		e := &InvalidSyntaxError{specifiedFlag: "--bad=flag=syntax"}
		if e.GetSpecifiedFlag() != "--bad=flag=syntax" {
			t.Errorf("GetSpecifiedFlag() = %q", e.GetSpecifiedFlag())
		}
		if !strings.Contains(e.Error(), "bad flag syntax") {
			t.Errorf("Error() = %q", e.Error())
		}
	})
}

// TestAliasVar tests AliasVar and AliasVarP for long-name aliases.
func TestAliasVar(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var format string
	fs.Var(newStringValue("", &format), "format", "output format")
	fs.AliasVar(newStringValue("", &format), "output-format", "alias for format")
	fs.AliasVarP(newStringValue("", &format), "fmt", "F", "alias with shorthand")

	// Long alias
	if err := fs.Parse([]string{"--output-format", "json"}); err != nil {
		t.Fatal(err)
	}
	if format != "json" {
		t.Errorf("format = %q, want json", format)
	}

	// Alias should be hidden
	usages := fs.FlagUsages()
	if strings.Contains(usages, "output-format") {
		t.Errorf("alias should be hidden from help:\n%s", usages)
	}
	if strings.Contains(usages, "--fmt") {
		t.Errorf("alias with shorthand should be hidden from help:\n%s", usages)
	}
}

// TestTextVar tests TextVar with a type implementing encoding.TextUnmarshaler.
func TestTextVar(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var ip net.IP
	fs.TextVar(&ip, "addr", &net.IP{127, 0, 0, 1}, "address")
	if err := fs.Parse([]string{"--addr", "10.0.0.1"}); err != nil {
		t.Fatal(err)
	}
	if ip.String() != "10.0.0.1" {
		t.Errorf("ip = %s, want 10.0.0.1", ip)
	}
}

// TestIPMaskInvalid tests IPMask with invalid input.
func TestIPMaskInvalid(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.IPMaskVar(new(net.IPMask), "mask", nil, "")
	if err := fs.Parse([]string{"--mask", "not-an-ip"}); err == nil {
		t.Error("expected error for invalid IP mask")
	}
}

// TestIPNetInvalid tests IPNet with invalid input.
func TestIPNetInvalid(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.IPNetVar(new(net.IPNet), "cidr", net.IPNet{}, "")
	if err := fs.Parse([]string{"--cidr", "not-a-cidr"}); err == nil {
		t.Error("expected error for invalid CIDR")
	}
}

// TestAddGoFlag tests adding a single Go stdlib flag.
func TestAddGoFlag(t *testing.T) {
	goFS := flag.NewFlagSet("go", flag.ContinueOnError)
	goFS.String("output", "default.txt", "output file")

	fs := NewFlagSet("test", ContinueOnError)
	goFS.VisitAll(func(gf *flag.Flag) {
		fs.AddGoFlag(gf)
	})

	f := fs.Lookup("output")
	if f == nil {
		t.Fatal("flag not found after AddGoFlag")
	}
	if f.DefValue != "default.txt" {
		t.Errorf("DefValue = %q", f.DefValue)
	}
	if err := fs.Parse([]string{"--output", "result.txt"}); err != nil {
		t.Fatal(err)
	}
	if f.Value.String() != "result.txt" {
		t.Errorf("value = %q, want result.txt", f.Value.String())
	}
}

// TestAddGoFlagSet tests adding all flags from a Go stdlib FlagSet.
func TestAddGoFlagSet(t *testing.T) {
	goFS := flag.NewFlagSet("go", flag.ContinueOnError)
	goFS.String("name", "", "name flag")
	goFS.Int("count", 0, "count flag")

	fs := NewFlagSet("test", ContinueOnError)
	fs.AddGoFlagSet(goFS)

	if fs.Lookup("name") == nil || fs.Lookup("count") == nil {
		t.Error("flags not found after AddGoFlagSet")
	}
	// nil should not panic
	fs.AddGoFlagSet(nil)
}

// TestCopyToGoFlagSet tests copying pflags to a Go stdlib FlagSet.
func TestCopyToGoFlagSet(t *testing.T) {
	pfs := NewFlagSet("test", ContinueOnError)
	pfs.StringVar(new(string), "output", "default.txt", "output file")
	pfs.IntVar(new(int), "count", 5, "count")

	goFS := flag.NewFlagSet("go", flag.ContinueOnError)
	CopyToGoFlagSet(pfs, goFS)

	gf := goFS.Lookup("output")
	if gf == nil {
		t.Fatal("output flag not found in Go FlagSet")
	}
	if gf.DefValue != "default.txt" {
		t.Errorf("DefValue = %q", gf.DefValue)
	}
}

// TestPFlagFromGoFlag tests the conversion function.
func TestPFlagFromGoFlag(t *testing.T) {
	goFS := flag.NewFlagSet("go", flag.ContinueOnError)
	goFS.String("name", "default", "a name")
	gf := goFS.Lookup("name")

	pf := PFlagFromGoFlag(gf)
	if pf.Name != "name" || pf.DefValue != "default" || pf.Usage != "a name" {
		t.Errorf("PFlagFromGoFlag: Name=%q DefValue=%q Usage=%q", pf.Name, pf.DefValue, pf.Usage)
	}
	if pf.Value.Type() != "string" {
		t.Errorf("Type = %q", pf.Value.Type())
	}
}

// TestGlobalWrapperSmoke exercises a sample of global CommandLine wrappers
// to verify they delegate correctly. Not exhaustive — these are one-liner
// delegations to already-tested FlagSet methods.
func TestGlobalWrapperSmoke(t *testing.T) {
	// Save and restore CommandLine
	saved := CommandLine
	defer func() { CommandLine = saved }()
	CommandLine = NewFlagSet("test", ContinueOnError)

	Int64Var(new(int64), "i64", 99, "")
	UintVar(new(uint), "u", 7, "")
	Uint64Var(new(uint64), "u64", 8, "")
	Int8Var(new(int8), "i8", 7, "")
	Int16Var(new(int16), "i16", 16, "")
	Int32Var(new(int32), "i32", 32, "")
	Uint8Var(new(uint8), "u8", 8, "")
	Uint16Var(new(uint16), "u16", 16, "")
	Uint32Var(new(uint32), "u32", 32, "")
	Float32Var(new(float32), "f32", 1.5, "")
	StringSliceVar(new([]string), "ss", nil, "")
	IntSliceVar(new([]int), "is", nil, "")
	BoolSliceVar(new([]bool), "bs", nil, "")
	DurationSliceVar(new([]time.Duration), "ds", nil, "")
	StringArrayVar(new([]string), "sa", nil, "")
	StringToStringVar(new(map[string]string), "sts", nil, "")
	StringToIntVar(new(map[string]int), "sti", nil, "")
	StringToInt64Var(new(map[string]int64), "sti64", nil, "")
	CountVar(new(int), "cnt", "")
	IPVar(new(net.IP), "ip", nil, "")
	IPMaskVar(new(net.IPMask), "mask", nil, "")
	IPNetVar(new(net.IPNet), "cidr", net.IPNet{}, "")
	Func("fn", "", func(string) error { return nil })
	BoolFunc("bf", "", func(string) error { return nil })

	if !HasFlags() {
		t.Error("HasFlags should be true")
	}
	if !HasAvailableFlags() {
		t.Error("HasAvailableFlags should be true")
	}
	if NFlag() != 0 {
		t.Error("NFlag should be 0 before parse")
	}
	if Changed("i64") {
		t.Error("i64 should not be changed before parse")
	}
	if ArgsLenAtDash() != -1 {
		t.Error("ArgsLenAtDash should be -1 before parse")
	}
	_ = FlagUsagesWrapped(80)
}

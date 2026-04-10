package pflag

// This file verifies that the pflag module can serve as a drop-in replacement
// for the Go stdlib "flag" package. It uses compile-time type assertions and
// runtime checks to ensure all required symbols are exported with compatible
// signatures.
//
// **Validates: Requirements 1.3, 1.4, 8.1, 8.2, 8.3, 8.4**

import (
	"io"
	"testing"
	"time"
)

// --- Compile-time assertions: types and interfaces ---

// Verify ErrorHandling constants exist and have the expected values.
var _ ErrorHandling = ContinueOnError
var _ ErrorHandling = ExitOnError
var _ ErrorHandling = PanicOnError

// Verify CommandLine is a *FlagSet.
var _ *FlagSet = CommandLine

// Verify Value interface has the right method set.
var _ interface {
	String() string
	Set(string) error
	Type() string
} = (Value)(nil)

// Verify NewFlagSet signature: (string, ErrorHandling) → *FlagSet.
var _ func(string, ErrorHandling) *FlagSet = NewFlagSet

// --- Compile-time assertions: FlagSet methods matching stdlib flag.FlagSet ---

// These function variable assignments verify that the FlagSet methods exist
// with signatures compatible with stdlib flag usage patterns.

// Core FlagSet methods from stdlib flag.
var (
	_ func(*FlagSet, io.Writer)                                     = (*FlagSet).SetOutput
	_ func(*FlagSet) io.Writer                                      = (*FlagSet).Output
	_ func(*FlagSet, string) *Flag                                  = (*FlagSet).Lookup
	_ func(*FlagSet, string, string) error                          = (*FlagSet).Set
	_ func(*FlagSet, func(*Flag))                                   = (*FlagSet).VisitAll
	_ func(*FlagSet, func(*Flag))                                   = (*FlagSet).Visit
	_ func(*FlagSet) string                                         = (*FlagSet).Name
	_ func(*FlagSet) bool                                           = (*FlagSet).Parsed
	_ func(*FlagSet, string, ErrorHandling)                         = (*FlagSet).Init
	_ func(*FlagSet, []string) error                                = (*FlagSet).Parse
	_ func(*FlagSet)                                                = (*FlagSet).PrintDefaults
	_ func(*FlagSet) []string                                       = (*FlagSet).Args
	_ func(*FlagSet) int                                            = (*FlagSet).NArg
	_ func(*FlagSet, int) string                                    = (*FlagSet).Arg
	_ func(*FlagSet) int                                            = (*FlagSet).NFlag
	_ func(*FlagSet, Value, string, string)                         = (*FlagSet).Var
	_ func(*FlagSet, *string, string, string, string)               = (*FlagSet).StringVar
	_ func(*FlagSet, string, string, string) *string                = (*FlagSet).String
	_ func(*FlagSet, *bool, string, bool, string)                   = (*FlagSet).BoolVar
	_ func(*FlagSet, string, bool, string) *bool                    = (*FlagSet).Bool
	_ func(*FlagSet, *int, string, int, string)                     = (*FlagSet).IntVar
	_ func(*FlagSet, string, int, string) *int                      = (*FlagSet).Int
	_ func(*FlagSet, *int64, string, int64, string)                 = (*FlagSet).Int64Var
	_ func(*FlagSet, string, int64, string) *int64                  = (*FlagSet).Int64
	_ func(*FlagSet, *uint, string, uint, string)                   = (*FlagSet).UintVar
	_ func(*FlagSet, string, uint, string) *uint                    = (*FlagSet).Uint
	_ func(*FlagSet, *uint64, string, uint64, string)               = (*FlagSet).Uint64Var
	_ func(*FlagSet, string, uint64, string) *uint64                = (*FlagSet).Uint64
	_ func(*FlagSet, *float64, string, float64, string)             = (*FlagSet).Float64Var
	_ func(*FlagSet, string, float64, string) *float64              = (*FlagSet).Float64
	_ func(*FlagSet, *time.Duration, string, time.Duration, string) = (*FlagSet).DurationVar
	_ func(*FlagSet, string, time.Duration, string) *time.Duration  = (*FlagSet).Duration
	_ func(*FlagSet, string, string, func(string) error)            = (*FlagSet).Func
	_ func(*FlagSet, string, string, func(string) error)            = (*FlagSet).BoolFunc
)

// --- Compile-time assertions: package-level global functions matching stdlib flag ---

var (
	_ func(*string, string, string, string)               = StringVar
	_ func(string, string, string) *string                = String
	_ func(*bool, string, bool, string)                   = BoolVar
	_ func(string, bool, string) *bool                    = Bool
	_ func(*int, string, int, string)                     = IntVar
	_ func(string, int, string) *int                      = Int
	_ func(*int64, string, int64, string)                 = Int64Var
	_ func(string, int64, string) *int64                  = Int64
	_ func(*uint, string, uint, string)                   = UintVar
	_ func(string, uint, string) *uint                    = Uint
	_ func(*uint64, string, uint64, string)               = Uint64Var
	_ func(string, uint64, string) *uint64                = Uint64
	_ func(*float64, string, float64, string)             = Float64Var
	_ func(string, float64, string) *float64              = Float64
	_ func(*time.Duration, string, time.Duration, string) = DurationVar
	_ func(string, time.Duration, string) *time.Duration  = Duration
	_ func(Value, string, string)                         = Var
	_ func()                                              = Parse
	_ func() bool                                         = Parsed
	_ func() []string                                     = Args
	_ func() int                                          = NArg
	_ func(int) string                                    = Arg
	_ func(string) *Flag                                  = Lookup
	_ func(string, string) error                          = Set
	_ func()                                              = PrintDefaults
	_ func(func(*Flag))                                   = VisitAll
	_ func(func(*Flag))                                   = Visit
	_ func() int                                          = NFlag
	_ func(string, string, func(string) error)            = Func
	_ func(string, string, func(string) error)            = BoolFunc
)

// TestStdlibFlagDropInConstants verifies ErrorHandling constants have the
// expected ordinal values matching stdlib flag.
func TestStdlibFlagDropInConstants(t *testing.T) {
	if ContinueOnError != 0 {
		t.Errorf("ContinueOnError = %d, want 0", ContinueOnError)
	}
	if ExitOnError != 1 {
		t.Errorf("ExitOnError = %d, want 1", ExitOnError)
	}
	if PanicOnError != 2 {
		t.Errorf("PanicOnError = %d, want 2", PanicOnError)
	}
}

// TestStdlibFlagDropInNewFlagSet verifies NewFlagSet creates a usable FlagSet
// with the same calling convention as stdlib flag.NewFlagSet.
func TestStdlibFlagDropInNewFlagSet(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	if fs == nil {
		t.Fatal("NewFlagSet returned nil")
	}
	if fs.Name() != "test" {
		t.Errorf("Name() = %q, want %q", fs.Name(), "test")
	}
	if fs.Parsed() {
		t.Error("Parsed() should be false before Parse()")
	}
}

// TestStdlibFlagDropInCommandLine verifies CommandLine is initialized and usable.
func TestStdlibFlagDropInCommandLine(t *testing.T) {
	if CommandLine == nil {
		t.Fatal("CommandLine is nil")
	}
	// CommandLine should have a non-empty name (os.Args[0]).
	if CommandLine.Name() == "" {
		t.Error("CommandLine.Name() is empty")
	}
}

// TestStdlibFlagDropInBasicWorkflow verifies the basic stdlib flag workflow:
// define flags → parse → read values.
func TestStdlibFlagDropInBasicWorkflow(t *testing.T) {
	fs := NewFlagSet("workflow", ContinueOnError)

	s := fs.String("name", "default", "a name")
	b := fs.Bool("verbose", false, "verbose mode")
	i := fs.Int("count", 0, "a count")
	f := fs.Float64("rate", 0.0, "a rate")
	d := fs.Duration("timeout", 0, "a timeout")

	err := fs.Parse([]string{
		"--name", "alice",
		"--verbose",
		"--count", "42",
		"--rate", "3.14",
		"--timeout", "5s",
	})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if *s != "alice" {
		t.Errorf("name = %q, want %q", *s, "alice")
	}
	if !*b {
		t.Error("verbose should be true")
	}
	if *i != 42 {
		t.Errorf("count = %d, want 42", *i)
	}
	if *f != 3.14 {
		t.Errorf("rate = %f, want 3.14", *f)
	}
	if *d != 5*time.Second {
		t.Errorf("timeout = %v, want 5s", *d)
	}

	if !fs.Parsed() {
		t.Error("Parsed() should be true after Parse()")
	}
	if fs.NFlag() != 5 {
		t.Errorf("NFlag() = %d, want 5", fs.NFlag())
	}
}

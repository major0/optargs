package optargs

import (
	"testing"
	"time"
)

// TestTypedValueInFlagHandle verifies that TypedValue implementations
// can be used from Flag.Handle callbacks — the primary integration point.
func TestTypedValueInFlagHandle(t *testing.T) {
	var name string
	var count int
	var verbose bool
	var timeout time.Duration

	nameVal := NewStringValue("", &name)
	countVal := NewIntValue(0, &count)
	verboseVal := NewBoolValue(false, &verbose)
	timeoutVal := NewDurationValue(0, &timeout)

	short := map[byte]*Flag{
		'n': {Name: "n", HasArg: RequiredArgument, Handle: func(_, arg string) error {
			return nameVal.Set(arg)
		}},
		'c': {Name: "c", HasArg: RequiredArgument, Handle: func(_, arg string) error {
			return countVal.Set(arg)
		}},
		'v': {Name: "v", HasArg: NoArgument, Handle: func(_, _ string) error {
			return verboseVal.Set("true")
		}},
	}
	long := map[string]*Flag{
		"timeout": {Name: "timeout", HasArg: RequiredArgument, Handle: func(_, arg string) error {
			return timeoutVal.Set(arg)
		}},
	}

	args := []string{"-n", "test", "-c", "42", "-v", "--timeout=5s"}
	p, err := NewParser(ParserConfig{enableErrors: true}, short, long, args)
	if err != nil {
		t.Fatal(err)
	}

	for _, err := range p.Options() {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if name != "test" {
		t.Errorf("name = %q, want %q", name, "test")
	}
	if count != 42 {
		t.Errorf("count = %d, want 42", count)
	}
	if !verbose {
		t.Error("verbose = false, want true")
	}
	if timeout != 5*time.Second {
		t.Errorf("timeout = %v, want 5s", timeout)
	}
}

// TestTypedValueHandleError verifies that TypedValue errors propagate
// through Flag.Handle correctly.
func TestTypedValueHandleError(t *testing.T) {
	var count int
	countVal := NewIntValue(0, &count)

	short := map[byte]*Flag{
		'c': {Name: "c", HasArg: RequiredArgument, Handle: func(_, arg string) error {
			return countVal.Set(arg)
		}},
	}

	args := []string{"-c", "not-a-number"}
	p, err := NewParser(ParserConfig{enableErrors: true}, short, nil, args)
	if err != nil {
		t.Fatal(err)
	}

	var gotErr error
	for _, err := range p.Options() {
		if err != nil {
			gotErr = err
			break
		}
	}
	if gotErr == nil {
		t.Fatal("expected error for invalid int")
	}
}

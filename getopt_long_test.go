package optargs

import (
	"testing"
)

// Validate that we can insantiate the parser with no short or long options.
func TestLongOptsNone(t *testing.T) {
	_, err := GetOptLong(nil, "", nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

// Validate that GetOptLongOnly propagates optstring errors.
func TestLongOnlyOptstringError(t *testing.T) {
	_, err := GetOptLongOnly(nil, "a-b", nil)
	if err == nil {
		t.Error("Expected error for invalid optstring, got nil")
	}
}

// Validate long-only mode with no short options falls back to error on unknown option.
func TestLongOnlyNoShortFallback(t *testing.T) {
	longopts := []Flag{
		{Name: "verbose", HasArg: NoArgument},
	}
	parser, err := GetOptLongOnly([]string{"-unknown"}, ":", longopts)
	if err != nil {
		t.Fatalf("Unexpected parser creation error: %v", err)
	}

	for _, err := range parser.Options() {
		if err == nil {
			t.Error("Expected error for unrecognized long-only option with no short fallback")
		}
	}
}

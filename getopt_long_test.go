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

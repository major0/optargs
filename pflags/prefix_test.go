package pflags

import "testing"

func TestFlagPrefixZeroValues(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.Bool("verbose", false, "verbose output")

	flag := fs.Lookup("verbose")
	if flag == nil {
		t.Fatal("flag not found")
	}
	if flag.Prefixes != nil {
		t.Errorf("Prefixes: got %v, want nil", flag.Prefixes)
	}
	if flag.Negatable {
		t.Error("Negatable: got true, want false")
	}
}

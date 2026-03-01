package optargs

import (
	"testing"
)

// caseInsensitiveLookupTests drives case-insensitive and case-sensitive
// command lookup subtests.
var caseInsensitiveLookupTests = []struct {
	name       string
	caseIgnore bool
	lookup     string
	wantFound  bool
}{
	// Case-insensitive mode: all casings match.
	{"insensitive_exact", true, "server", true},
	{"insensitive_upper", true, "SERVER", true},
	{"insensitive_mixed", true, "SeRvEr", true},
	{"insensitive_miss", true, "nonexistent", false},

	// Case-sensitive mode: only exact match works.
	{"sensitive_exact", false, "server", true},
	{"sensitive_upper", false, "SERVER", false},
	{"sensitive_mixed", false, "SeRvEr", false},
}

func TestCommandCaseInsensitiveLookup(t *testing.T) {
	for _, tt := range caseInsensitiveLookupTests {
		t.Run(tt.name, func(t *testing.T) {
			root := newMinimalParser(t)
			sub := newMinimalParser(t)
			root.config.commandCaseIgnore = tt.caseIgnore
			root.AddCmd("server", sub)

			got, exists := root.GetCommand(tt.lookup)
			if exists != tt.wantFound {
				t.Fatalf("GetCommand(%q) exists = %v, want %v", tt.lookup, exists, tt.wantFound)
			}
			if tt.wantFound && got != sub {
				t.Error("GetCommand returned wrong parser")
			}
		})
	}
}

func TestExecuteCommandCaseInsensitive(t *testing.T) {
	root := newMinimalParser(t)
	sub := newMinimalParser(t)
	root.config.commandCaseIgnore = true
	root.AddCmd("server", sub)

	got, err := root.ExecuteCommand("SERVER", []string{"--help"})
	if err != nil {
		t.Fatalf("ExecuteCommand(SERVER): %v", err)
	}
	if got != sub {
		t.Error("ExecuteCommand returned wrong parser")
	}
}

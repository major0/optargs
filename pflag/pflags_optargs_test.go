package pflag

import "testing"

// OptArgs-exclusive feature tests that aren't already covered in pflag_test.go.
// Most OptArgs features (POSIX compaction, boolean negation, short-only flags,
// many-to-one mappings, GNU prefix matching, long-only mode, count flags) are
// already tested in pflag_test.go. This file covers gaps only.

// TestOptArgsBoolArgValuer tests that types implementing BoolTakesArg() control
// whether the parser registers them as NoArgument or OptionalArgument.
func TestOptArgsBoolArgValuer(t *testing.T) {
	// Count implements BoolTakesArg() returning false — should be NoArgument.
	// This means --verbose should NOT consume the next positional argument.
	fs := NewFlagSet("test", ContinueOnError)
	var count int
	fs.CountVarP(&count, "verbose", "v", "")

	if err := fs.Parse([]string{"--verbose", "positional", "--verbose"}); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2 (--verbose should not consume 'positional')", count)
	}
	if fs.NArg() != 1 || fs.Arg(0) != "positional" {
		t.Errorf("args = %v, want [positional]", fs.Args())
	}

	// Bool implements BoolTakesArg() returning true — should be OptionalArgument.
	// This means --flag=value syntax works.
	fs2 := NewFlagSet("test", ContinueOnError)
	var b bool
	fs2.BoolVar(&b, "flag", false, "")
	if err := fs2.Parse([]string{"--flag=true"}); err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Error("--flag=true should set bool to true")
	}
}

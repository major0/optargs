package compat

import "testing"

// TestExpectedDiffsComplete verifies all expected diff entries have non-empty fields.
func TestExpectedDiffsComplete(t *testing.T) {
	if len(ExpectedDiffs) == 0 {
		t.Fatal("ExpectedDiffs is empty")
	}
	for i, d := range ExpectedDiffs {
		if d.Scenario == "" {
			t.Errorf("diff[%d]: Scenario is empty", i)
		}
		if d.Upstream == "" {
			t.Errorf("diff[%d] %q: Upstream is empty", i, d.Scenario)
		}
		if d.Ours == "" {
			t.Errorf("diff[%d] %q: Ours is empty", i, d.Scenario)
		}
		if d.Rationale == "" {
			t.Errorf("diff[%d] %q: Rationale is empty", i, d.Scenario)
		}
	}
}

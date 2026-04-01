package goarg

import "testing"

// TestExpectedDiffsComplete verifies all expected diff entries have non-empty fields.
func TestExpectedDiffsComplete(t *testing.T) {
	diffs := loadExpectedDiffs()
	if len(diffs) == 0 {
		t.Fatal("expectedDiffs is empty")
	}
	for key, d := range diffs {
		if d.Scenario == "" {
			t.Errorf("diff %q: Scenario is empty", key)
		}
		if d.UpstreamBehavior == "" {
			t.Errorf("diff %q: UpstreamBehavior is empty", key)
		}
		if d.OurBehavior == "" {
			t.Errorf("diff %q: OurBehavior is empty", key)
		}
		if d.Rationale == "" {
			t.Errorf("diff %q: Rationale is empty", key)
		}
	}
}

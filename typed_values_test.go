package optargs

import "testing"

// Compile-time interface satisfaction checks.
// These are populated as concrete types are added.
var (
	_ TypedValue = (*stringValue)(nil)
	_ TypedValue = (*boolValue)(nil)
	_ BoolValuer = (*boolValue)(nil)
)

func TestTypedValueInterfaceSatisfaction(t *testing.T) {
	tests := []struct {
		name     string
		val      TypedValue
		wantType string
	}{
		{"string", NewStringValue("hello", nil), "string"},
		{"bool", NewBoolValue(true, nil), "bool"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.Type(); got != tt.wantType {
				t.Errorf("Type() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

func TestBoolValuerSatisfaction(t *testing.T) {
	bv, ok := NewBoolValue(false, nil).(BoolValuer)
	if !ok {
		t.Fatal("NewBoolValue does not implement BoolValuer")
	}
	if !bv.IsBoolFlag() {
		t.Error("IsBoolFlag() = false, want true")
	}
}

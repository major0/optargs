package optargs

import "testing"

// Compile-time interface satisfaction checks.
var (
	_ TypedValue = (*stringValue)(nil)
	_ TypedValue = (*boolValue)(nil)
	_ TypedValue = (*scalarValue[int])(nil)
	_ TypedValue = (*durationValue)(nil)
	_ BoolValuer = (*boolValue)(nil)
	_ Resetter   = (*sliceValue)(nil)
	_ Resetter   = (*durationSliceValue)(nil)
	_ Resetter   = (*mapValue)(nil)
)


func TestZeroString(t *testing.T) {
	tests := []struct {
		typeName string
		wantVal  string
		wantOk   bool
	}{
		{"bool", "false", true},
		{"string", "", true},
		{"int", "0", true},
		{"stringSlice", "[]", true},
		{"stringToString", "map[]", true},
		{"unknown", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			val, ok := ZeroString(tt.typeName)
			if ok != tt.wantOk {
				t.Errorf("ZeroString(%q) ok = %v, want %v", tt.typeName, ok, tt.wantOk)
			}
			if val != tt.wantVal {
				t.Errorf("ZeroString(%q) = %q, want %q", tt.typeName, val, tt.wantVal)
			}
		})
	}
}

func TestIsBool(t *testing.T) {
	boolVal := NewBoolValue(false, new(bool))
	intVal := NewIntValue(0, new(int))
	countVal := NewCountValue(0, new(int))

	if !IsBool(boolVal) {
		t.Error("IsBool(boolValue) = false, want true")
	}
	if IsBool(intVal) {
		t.Error("IsBool(intValue) = true, want false")
	}
	// Count implements IsBoolFlag() returning true
	if !IsBool(countVal) {
		t.Error("IsBool(countValue) = false, want true")
	}
}

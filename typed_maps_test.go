package optargs

import (
	"strings"
	"testing"
)

func TestMapStringToString(t *testing.T) {
	tests := []struct {
		name    string
		sets    []string
		wantLen int
		wantErr bool
	}{
		{"single_pair", []string{"key=val"}, 1, false},
		{"multi_csv", []string{"a=1,b=2"}, 2, false},
		{"repeated_merge", []string{"a=1", "b=2"}, 2, false},
		{"no_equals", []string{"invalid"}, 0, true},
		{"empty_value", []string{"key="}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m map[string]string
			v := NewStringToStringValue(nil, &m)
			var lastErr error
			for _, input := range tt.sets {
				if err := v.Set(input); err != nil {
					lastErr = err
				}
			}
			if (lastErr != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", lastErr, tt.wantErr)
			}
			if !tt.wantErr && len(m) != tt.wantLen {
				t.Errorf("map len = %d, want %d", len(m), tt.wantLen)
			}
		})
	}
}

func TestMapStringToInt(t *testing.T) {
	var m map[string]int
	v := NewStringToIntValue(nil, &m)
	if err := v.Set("count=42,size=10"); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	if m["count"] != 42 || m["size"] != 10 {
		t.Errorf("unexpected map: %v", m)
	}
}

func TestMapStringToInt64(t *testing.T) {
	var m map[string]int64
	v := NewStringToInt64Value(nil, &m)
	if err := v.Set("big=9223372036854775807"); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	if m["big"] != 9223372036854775807 {
		t.Errorf("unexpected value: %d", m["big"])
	}
}

func TestMapInvalidValueType(t *testing.T) {
	var m map[string]int
	v := NewStringToIntValue(nil, &m)
	err := v.Set("key=abc")
	if err == nil {
		t.Fatal("expected error for invalid int value")
	}
	if !strings.Contains(err.Error(), "invalid value") {
		t.Errorf("error %q should contain 'invalid value'", err.Error())
	}
}

func TestMapFirstSetReplacesDefault(t *testing.T) {
	defaults := map[string]string{"old": "value"}
	var m map[string]string
	v := NewStringToStringValue(defaults, &m)

	// First Set() should replace the default map.
	if err := v.Set("new=entry"); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["old"]; ok {
		t.Error("first Set() should have replaced default map")
	}
	if m["new"] != "entry" {
		t.Errorf("expected new=entry, got %v", m)
	}

	// Subsequent Set() should merge.
	if err := v.Set("another=one"); err != nil {
		t.Fatal(err)
	}
	if m["new"] != "entry" || m["another"] != "one" {
		t.Errorf("subsequent Set() should merge, got %v", m)
	}
}

func TestMapNilPointers(t *testing.T) {
	tests := []struct {
		name string
		val  TypedValue
	}{
		{"stringToString", NewStringToStringValue(nil, nil)},
		{"stringToInt", NewStringToIntValue(nil, nil)},
		{"stringToInt64", NewStringToInt64Value(nil, nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.val.Set("key=1"); err != nil {
				t.Errorf("Set on nil pointer: %v", err)
			}
		})
	}
}

func TestMapFirstEqualsSplit(t *testing.T) {
	// Value containing '=' should split on first '=' only.
	var m map[string]string
	v := NewStringToStringValue(nil, &m)
	if err := v.Set("url=https://example.com?a=b"); err != nil {
		t.Fatal(err)
	}
	if m["url"] != "https://example.com?a=b" {
		t.Errorf("expected value with '=', got %q", m["url"])
	}
}

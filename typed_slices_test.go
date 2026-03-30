package optargs

import (
	"testing"
	"time"
)

func TestSliceStringValue(t *testing.T) {
	tests := []struct {
		name    string
		sets    []string
		want    string
		wantErr bool
	}{
		{"single", []string{"hello"}, "[hello]", false},
		{"csv", []string{"a,b,c"}, "[a,b,c]", false},
		{"repeated", []string{"a", "b"}, "[a,b]", false},
		{"csv_then_single", []string{"a,b", "c"}, "[a,b,c]", false},
		{"empty_csv", []string{""}, "[]", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s []string
			v := NewStringSliceValue(nil, &s)
			for _, input := range tt.sets {
				if err := v.Set(input); (err != nil) != tt.wantErr {
					t.Fatalf("Set(%q) error = %v, wantErr %v", input, err, tt.wantErr)
				}
			}
			if got := v.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSliceBoolValue(t *testing.T) {
	var b []bool
	v := NewBoolSliceValue(nil, &b)
	if err := v.Set("true,false,yes"); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	if got := v.String(); got != "[true,false,true]" {
		t.Errorf("String() = %q, want %q", got, "[true,false,true]")
	}
}

func TestSliceIntValue(t *testing.T) {
	tests := []struct {
		name    string
		sets    []string
		want    string
		wantErr bool
	}{
		{"csv", []string{"1,2,3"}, "[1,2,3]", false},
		{"repeated", []string{"1", "2"}, "[1,2]", false},
		{"invalid", []string{"1,abc"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s []int
			v := NewIntSliceValue(nil, &s)
			var lastErr error
			for _, input := range tt.sets {
				if err := v.Set(input); err != nil {
					lastErr = err
				}
			}
			if (lastErr != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", lastErr, tt.wantErr)
			}
			if !tt.wantErr && v.String() != tt.want {
				t.Errorf("String() = %q, want %q", v.String(), tt.want)
			}
		})
	}
}

func TestSliceAccumulation(t *testing.T) {
	// Verify repeated Set() calls accumulate.
	var s []int
	v := NewIntSliceValue(nil, &s)
	if err := v.Set("1,2"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("3"); err != nil {
		t.Fatal(err)
	}
	if got := v.String(); got != "[1,2,3]" {
		t.Errorf("String() = %q, want %q", got, "[1,2,3]")
	}
}

func TestSliceNilPointers(t *testing.T) {
	tests := []struct {
		name string
		val  TypedValue
	}{
		{"string", NewStringSliceValue(nil, nil)},
		{"bool", NewBoolSliceValue(nil, nil)},
		{"int", NewIntSliceValue(nil, nil)},
		{"int32", NewInt32SliceValue(nil, nil)},
		{"int64", NewInt64SliceValue(nil, nil)},
		{"uint", NewUintSliceValue(nil, nil)},
		{"float32", NewFloat32SliceValue(nil, nil)},
		{"float64", NewFloat64SliceValue(nil, nil)},
		{"duration", NewDurationSliceValue(nil, nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.val.Set(""); err != nil {
				t.Errorf("Set empty on nil pointer: %v", err)
			}
		})
	}
}

func TestSliceDurationValue(t *testing.T) {
	var d []time.Duration
	v := NewDurationSliceValue(nil, &d)
	if err := v.Set("1s,2m,3h"); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	if got := v.String(); got != "[1s,2m0s,3h0m0s]" {
		t.Errorf("String() = %q, want %q", got, "[1s,2m0s,3h0m0s]")
	}
}

func TestSliceFloat64Value(t *testing.T) {
	var f []float64
	v := NewFloat64SliceValue(nil, &f)
	if err := v.Set("1.1,2.2"); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	if err := v.Set("3.3"); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	if got := v.String(); got != "[1.1,2.2,3.3]" {
		t.Errorf("String() = %q, want %q", got, "[1.1,2.2,3.3]")
	}
}

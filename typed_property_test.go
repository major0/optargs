package optargs

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"
	"testing/quick"
	"time"
)

// Property 1: Round-trip — Set(v.String()) produces same value for all scalar types.
func TestPropertyRoundTripScalars(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		f := func(s string) bool {
			v := NewStringValue(s, nil)
			fresh := NewStringValue("", nil)
			if err := fresh.Set(v.String()); err != nil {
				return false
			}
			return fresh.String() == v.String()
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("int", func(t *testing.T) {
		f := func(n int) bool {
			v := NewIntValue(n, nil)
			fresh := NewIntValue(0, nil)
			if err := fresh.Set(v.String()); err != nil {
				return false
			}
			return fresh.String() == v.String()
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("int64", func(t *testing.T) {
		f := func(n int64) bool {
			v := NewInt64Value(n, nil)
			fresh := NewInt64Value(0, nil)
			if err := fresh.Set(v.String()); err != nil {
				return false
			}
			return fresh.String() == v.String()
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("uint", func(t *testing.T) {
		f := func(n uint) bool {
			v := NewUintValue(n, nil)
			fresh := NewUintValue(0, nil)
			if err := fresh.Set(v.String()); err != nil {
				return false
			}
			return fresh.String() == v.String()
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("uint64", func(t *testing.T) {
		f := func(n uint64) bool {
			v := NewUint64Value(n, nil)
			fresh := NewUint64Value(0, nil)
			if err := fresh.Set(v.String()); err != nil {
				return false
			}
			return fresh.String() == v.String()
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("bool", func(t *testing.T) {
		f := func(b bool) bool {
			v := NewBoolValue(b, nil)
			fresh := NewBoolValue(false, nil)
			if err := fresh.Set(v.String()); err != nil {
				return false
			}
			return fresh.String() == v.String()
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("float64", func(t *testing.T) {
		f := func(x float64) bool {
			if math.IsNaN(x) || math.IsInf(x, 0) {
				return true // skip non-round-trippable values
			}
			v := NewFloat64Value(x, nil)
			fresh := NewFloat64Value(0, nil)
			if err := fresh.Set(v.String()); err != nil {
				return false
			}
			return fresh.String() == v.String()
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	// Duration: testing/quick can't generate time.Duration directly,
	// so we use int64 nanoseconds as a proxy.
	t.Run("duration", func(t *testing.T) {
		f := func(ns int64) bool {
			d := time.Duration(ns)
			v := NewDurationValue(d, nil)
			fresh := NewDurationValue(0, nil)
			if err := fresh.Set(v.String()); err != nil {
				return false
			}
			return fresh.String() == v.String()
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})
}

// Property 2: Convert delegation — Set(s) matches Convert(s, type) for all scalars.
func TestPropertyConvertDelegation(t *testing.T) {
	types := []struct {
		name       string
		newVal     func() TypedValue
		targetType reflect.Type
	}{
		{"int", func() TypedValue { return NewIntValue(0, nil) }, intType},
		{"int64", func() TypedValue { return NewInt64Value(0, nil) }, int64Type},
		{"uint", func() TypedValue { return NewUintValue(0, nil) }, uintType},
		{"uint64", func() TypedValue { return NewUint64Value(0, nil) }, uint64Type},
		{"float64", func() TypedValue { return NewFloat64Value(0, nil) }, float64Type},
	}
	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			f := func(n int64) bool {
				// Use absolute value to avoid uint issues with negatives.
				s := strconv.FormatInt(n, 10)
				if tt.name == "uint" || tt.name == "uint64" {
					if n < 0 {
						return true // skip negatives for unsigned
					}
					s = strconv.FormatUint(uint64(n), 10)
				}
				if tt.name == "float64" {
					s = fmt.Sprintf("%g", float64(n))
				}

				v := tt.newVal()
				setErr := v.Set(s)
				_, convErr := Convert(s, tt.targetType)

				// Both should succeed or both should fail.
				if (setErr != nil) != (convErr != nil) {
					return false
				}
				if setErr != nil {
					return true // both failed, consistent
				}
				converted, _ := Convert(s, tt.targetType)
				return v.String() == fmt.Sprintf("%v", converted)
			}
			if err := quick.Check(f, nil); err != nil {
				t.Error(err)
			}
		})
	}
}

// Property 3: Slice accumulation — Set("a,b") + Set("c") = [a,b,c] for all slice types.
func TestPropertySliceAccumulation(t *testing.T) {
	t.Run("stringSlice", func(t *testing.T) {
		v := NewStringSliceValue(nil, nil)
		_ = v.Set("a,b")
		_ = v.Set("c")
		if got := v.String(); got != "[a,b,c]" {
			t.Errorf("got %q, want %q", got, "[a,b,c]")
		}
	})

	t.Run("intSlice", func(t *testing.T) {
		v := NewIntSliceValue(nil, nil)
		_ = v.Set("1,2")
		_ = v.Set("3")
		if got := v.String(); got != "[1,2,3]" {
			t.Errorf("got %q, want %q", got, "[1,2,3]")
		}
	})

	t.Run("boolSlice", func(t *testing.T) {
		v := NewBoolSliceValue(nil, nil)
		_ = v.Set("true,false")
		_ = v.Set("yes")
		if got := v.String(); got != "[true,false,true]" {
			t.Errorf("got %q, want %q", got, "[true,false,true]")
		}
	})

	t.Run("durationSlice", func(t *testing.T) {
		v := NewDurationSliceValue(nil, nil)
		_ = v.Set("1s,2s")
		_ = v.Set("3s")
		if got := v.String(); got != "[1s,2s,3s]" {
			t.Errorf("got %q, want %q", got, "[1s,2s,3s]")
		}
	})
}

// Property 4: Boolean consistency — NewBoolValue.Set(s) matches convertBool(s).
func TestPropertyBooleanConsistency(t *testing.T) {
	inputs := []string{
		"true", "TRUE", "True", "t", "T", "1", "yes", "YES", "y", "Y", "on", "ON",
		"false", "FALSE", "False", "f", "F", "0", "no", "NO", "n", "N", "off", "OFF",
		"", "maybe", "2", "nope",
	}
	for _, s := range inputs {
		t.Run(s, func(t *testing.T) {
			expected, expectedErr := convertBool(s)

			v := NewBoolValue(false, nil)
			actualErr := v.Set(s)

			if (expectedErr != nil) != (actualErr != nil) {
				t.Errorf("convertBool(%q) err=%v, Set(%q) err=%v", s, expectedErr, s, actualErr)
				return
			}
			if expectedErr == nil {
				got := v.String() == "true"
				if got != expected {
					t.Errorf("convertBool(%q)=%v, Set(%q).String()=%q", s, expected, s, v.String())
				}
			}
		})
	}
}

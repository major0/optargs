package optargs

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

// Feature: goarg-optargs-integration, Property 1: Type conversion round-trip
// Validates: Requirements 2.1, 2.2
//
// For any Go value of a supported type, fmt.Sprint(value) → Convert(str, type)
// produces the original value.

func TestPropertyTypeConversionRoundTripString(t *testing.T) {
	f := func(v string) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, string) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripBool(t *testing.T) {
	f := func(v bool) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, bool) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripInt(t *testing.T) {
	f := func(v int) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, int) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripInt8(t *testing.T) {
	f := func(v int8) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, int8) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripInt16(t *testing.T) {
	f := func(v int16) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, int16) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripInt32(t *testing.T) {
	f := func(v int32) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, int32) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripInt64(t *testing.T) {
	f := func(v int64) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, int64) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripUint(t *testing.T) {
	f := func(v uint) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, uint) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripUint8(t *testing.T) {
	f := func(v uint8) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, uint8) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripUint16(t *testing.T) {
	f := func(v uint16) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, uint16) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripUint32(t *testing.T) {
	f := func(v uint32) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, uint32) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripUint64(t *testing.T) {
	f := func(v uint64) bool {
		s := fmt.Sprint(v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, uint64) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripFloat32(t *testing.T) {
	f := func(v float32) bool {
		// Skip NaN and Inf — they don't round-trip through strconv.ParseFloat.
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return true
		}
		s := fmt.Sprintf("%v", v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, float32) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyTypeConversionRoundTripFloat64(t *testing.T) {
	f := func(v float64) bool {
		// Skip NaN and Inf — they don't round-trip through strconv.ParseFloat.
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return true
		}
		s := fmt.Sprintf("%v", v)
		got, err := Convert(s, reflect.TypeOf(v))
		if err != nil {
			t.Logf("Convert(%q, float64) error: %v", s, err)
			return false
		}
		return got == v
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Feature: goarg-optargs-integration, Property 2: ConvertSlice round-trip
// Validates: Requirements 2.3
//
// For any slice of a supported element type, joining elements with commas
// → ConvertSlice produces the original slice.

func TestPropertyConvertSliceRoundTripString(t *testing.T) {
	f := func(vs []string) bool {
		// Filter out elements that won't round-trip: empty, whitespace-only, or containing commas.
		var clean []string
		for _, v := range vs {
			trimmed := strings.TrimSpace(v)
			if trimmed == "" || strings.Contains(v, ",") || v != trimmed {
				continue
			}
			clean = append(clean, v)
		}
		if len(clean) == 0 {
			return true
		}
		csv := strings.Join(clean, ",")
		got, err := ConvertSlice(csv, reflect.TypeOf(clean))
		if err != nil {
			t.Logf("ConvertSlice(%q, []string) error: %v", csv, err)
			return false
		}
		result := got.([]string)
		if len(result) != len(clean) {
			return false
		}
		for i := range clean {
			if result[i] != clean[i] {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyConvertSliceRoundTripBool(t *testing.T) {
	f := func(vs []bool) bool {
		if len(vs) == 0 {
			return true
		}
		parts := make([]string, len(vs))
		for i, v := range vs {
			parts[i] = fmt.Sprint(v)
		}
		csv := strings.Join(parts, ",")
		got, err := ConvertSlice(csv, reflect.TypeOf(vs))
		if err != nil {
			t.Logf("ConvertSlice(%q, []bool) error: %v", csv, err)
			return false
		}
		result := got.([]bool)
		if len(result) != len(vs) {
			return false
		}
		for i := range vs {
			if result[i] != vs[i] {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyConvertSliceRoundTripInt(t *testing.T) {
	f := func(vs []int) bool {
		if len(vs) == 0 {
			return true
		}
		parts := make([]string, len(vs))
		for i, v := range vs {
			parts[i] = fmt.Sprint(v)
		}
		csv := strings.Join(parts, ",")
		got, err := ConvertSlice(csv, reflect.TypeOf(vs))
		if err != nil {
			t.Logf("ConvertSlice(%q, []int) error: %v", csv, err)
			return false
		}
		result := got.([]int)
		if len(result) != len(vs) {
			return false
		}
		for i := range vs {
			if result[i] != vs[i] {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyConvertSliceRoundTripInt64(t *testing.T) {
	f := func(vs []int64) bool {
		if len(vs) == 0 {
			return true
		}
		parts := make([]string, len(vs))
		for i, v := range vs {
			parts[i] = fmt.Sprint(v)
		}
		csv := strings.Join(parts, ",")
		got, err := ConvertSlice(csv, reflect.TypeOf(vs))
		if err != nil {
			t.Logf("ConvertSlice(%q, []int64) error: %v", csv, err)
			return false
		}
		result := got.([]int64)
		if len(result) != len(vs) {
			return false
		}
		for i := range vs {
			if result[i] != vs[i] {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyConvertSliceRoundTripUint(t *testing.T) {
	f := func(vs []uint) bool {
		if len(vs) == 0 {
			return true
		}
		parts := make([]string, len(vs))
		for i, v := range vs {
			parts[i] = fmt.Sprint(v)
		}
		csv := strings.Join(parts, ",")
		got, err := ConvertSlice(csv, reflect.TypeOf(vs))
		if err != nil {
			t.Logf("ConvertSlice(%q, []uint) error: %v", csv, err)
			return false
		}
		result := got.([]uint)
		if len(result) != len(vs) {
			return false
		}
		for i := range vs {
			if result[i] != vs[i] {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyConvertSliceRoundTripFloat64(t *testing.T) {
	f := func(vs []float64) bool {
		// Filter out NaN and Inf — they don't round-trip through strconv.
		var clean []float64
		for _, v := range vs {
			if !math.IsNaN(v) && !math.IsInf(v, 0) {
				clean = append(clean, v)
			}
		}
		if len(clean) == 0 {
			return true
		}
		parts := make([]string, len(clean))
		for i, v := range clean {
			parts[i] = fmt.Sprintf("%v", v)
		}
		csv := strings.Join(parts, ",")
		got, err := ConvertSlice(csv, reflect.TypeOf(clean))
		if err != nil {
			t.Logf("ConvertSlice(%q, []float64) error: %v", csv, err)
			return false
		}
		result := got.([]float64)
		if len(result) != len(clean) {
			return false
		}
		for i := range clean {
			if result[i] != clean[i] {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// Feature: goarg-optargs-integration, Property 3: Type conversion error messages contain type and value
// Validates: Requirements 2.4
//
// For any invalid string/type pair, the error message contains both the
// type name and the input string.

// convertErrorContainsValue checks that the error message contains the
// Go-quoted representation of the input value (matching the %q format
// used by Convert's error messages).
func convertErrorContainsValue(msg, value string) bool {
	quoted := fmt.Sprintf("%q", value)
	return strings.Contains(msg, quoted)
}

func TestPropertyConvertErrorInt(t *testing.T) {
	f := func(s string) bool {
		_, err := Convert(s, reflect.TypeOf(int(0)))
		if err == nil {
			return true // valid input, skip
		}
		msg := err.Error()
		if !strings.Contains(msg, "int") {
			t.Logf("error %q missing type name 'int'", msg)
			return false
		}
		if !convertErrorContainsValue(msg, s) {
			t.Logf("error %q missing quoted input value %q", msg, s)
			return false
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyConvertErrorBool(t *testing.T) {
	f := func(s string) bool {
		_, err := Convert(s, reflect.TypeOf(false))
		if err == nil {
			return true // valid input, skip
		}
		msg := err.Error()
		if !strings.Contains(msg, "bool") {
			t.Logf("error %q missing type name 'bool'", msg)
			return false
		}
		if !convertErrorContainsValue(msg, s) {
			t.Logf("error %q missing quoted input value %q", msg, s)
			return false
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyConvertErrorFloat64(t *testing.T) {
	f := func(s string) bool {
		_, err := Convert(s, reflect.TypeOf(float64(0)))
		if err == nil {
			return true // valid input, skip
		}
		msg := err.Error()
		if !strings.Contains(msg, "float64") {
			t.Logf("error %q missing type name 'float64'", msg)
			return false
		}
		if !convertErrorContainsValue(msg, s) {
			t.Logf("error %q missing quoted input value %q", msg, s)
			return false
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

func TestPropertyConvertErrorUint(t *testing.T) {
	f := func(s string) bool {
		_, err := Convert(s, reflect.TypeOf(uint(0)))
		if err == nil {
			return true // valid input, skip
		}
		msg := err.Error()
		if !strings.Contains(msg, "uint") {
			t.Logf("error %q missing type name 'uint'", msg)
			return false
		}
		if !convertErrorContainsValue(msg, s) {
			t.Logf("error %q missing quoted input value %q", msg, s)
			return false
		}
		return true
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// ---------------------------------------------------------------------------
// Unit tests — edge cases that property tests don't reach
// ---------------------------------------------------------------------------

// testTextUnmarshaler is a simple type implementing encoding.TextUnmarshaler
// for testing the TextUnmarshaler code path.
type testTextUnmarshaler struct {
	Value string
}

func (t *testTextUnmarshaler) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "FAIL" {
		return fmt.Errorf("unmarshal refused")
	}
	t.Value = s
	return nil
}

func TestConvertEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		targetType reflect.Type
		want       interface{}
		wantErr    string // substring match; empty means no error
	}{
		// Empty string handling
		{
			name:       "empty string to string succeeds",
			value:      "",
			targetType: reflect.TypeOf(""),
			want:       "",
		},
		{
			name:       "empty string to bool returns false",
			value:      "",
			targetType: reflect.TypeOf(false),
			want:       false,
		},
		{
			name:       "empty string to int errors",
			value:      "",
			targetType: reflect.TypeOf(0),
			wantErr:    `invalid value "" for type int`,
		},
		{
			name:       "empty string to float64 errors",
			value:      "",
			targetType: reflect.TypeOf(float64(0)),
			wantErr:    `invalid value "" for type float64`,
		},

		// Overflow values
		{
			name:       "256 overflows int8",
			value:      "256",
			targetType: reflect.TypeOf(int8(0)),
			wantErr:    `invalid value "256" for type int8`,
		},
		{
			name:       "-129 overflows int8",
			value:      "-129",
			targetType: reflect.TypeOf(int8(0)),
			wantErr:    `invalid value "-129" for type int8`,
		},
		{
			name:       "huge number overflows int",
			value:      "99999999999999999999",
			targetType: reflect.TypeOf(0),
			wantErr:    "invalid value",
		},
		{
			name:       "256 overflows uint8",
			value:      "256",
			targetType: reflect.TypeOf(uint8(0)),
			wantErr:    `invalid value "256" for type uint8`,
		},
		{
			name:       "negative overflows uint",
			value:      "-1",
			targetType: reflect.TypeOf(uint(0)),
			wantErr:    `invalid value "-1" for type uint`,
		},

		// Unsupported types
		{
			name:       "chan type unsupported",
			value:      "x",
			targetType: reflect.TypeOf(make(chan int)),
			wantErr:    "unsupported type: chan int",
		},
		{
			name:       "func type unsupported",
			value:      "x",
			targetType: reflect.TypeOf(func() {}),
			wantErr:    "unsupported type: func()",
		},
		{
			name:       "map type unsupported",
			value:      "x",
			targetType: reflect.TypeOf(map[string]int{}),
			wantErr:    "unsupported type: map[string]int",
		},
		{
			name:       "complex128 type unsupported",
			value:      "x",
			targetType: reflect.TypeOf(complex128(0)),
			wantErr:    "unsupported type: complex128",
		},

		// Bool case-insensitivity
		{
			name:       "TRUE uppercase",
			value:      "TRUE",
			targetType: reflect.TypeOf(false),
			want:       true,
		},
		{
			name:       "True mixed case",
			value:      "True",
			targetType: reflect.TypeOf(false),
			want:       true,
		},
		{
			name:       "YES uppercase",
			value:      "YES",
			targetType: reflect.TypeOf(false),
			want:       true,
		},
		{
			name:       "Yes mixed case",
			value:      "Yes",
			targetType: reflect.TypeOf(false),
			want:       true,
		},
		{
			name:       "ON uppercase",
			value:      "ON",
			targetType: reflect.TypeOf(false),
			want:       true,
		},
		{
			name:       "On mixed case",
			value:      "On",
			targetType: reflect.TypeOf(false),
			want:       true,
		},
		{
			name:       "FALSE uppercase",
			value:      "FALSE",
			targetType: reflect.TypeOf(false),
			want:       false,
		},
		{
			name:       "NO uppercase",
			value:      "NO",
			targetType: reflect.TypeOf(false),
			want:       false,
		},
		{
			name:       "OFF uppercase",
			value:      "OFF",
			targetType: reflect.TypeOf(false),
			want:       false,
		},
		{
			name:       "invalid bool value",
			value:      "maybe",
			targetType: reflect.TypeOf(false),
			wantErr:    `invalid value "maybe" for type bool`,
		},

		// Pointer types
		{
			name:       "pointer to int",
			value:      "42",
			targetType: reflect.TypeOf((*int)(nil)),
			want:       func() interface{} { v := 42; return &v }(),
		},
		{
			name:       "pointer to string",
			value:      "hello",
			targetType: reflect.TypeOf((*string)(nil)),
			want:       func() interface{} { v := "hello"; return &v }(),
		},
		{
			name:       "pointer to bool",
			value:      "true",
			targetType: reflect.TypeOf((*bool)(nil)),
			want:       func() interface{} { v := true; return &v }(),
		},
		{
			name:       "pointer to int with invalid value errors",
			value:      "abc",
			targetType: reflect.TypeOf((*int)(nil)),
			wantErr:    `invalid value "abc" for type int`,
		},

		// Slice via Convert (single value → single-element slice)
		{
			name:       "single int to int slice",
			value:      "7",
			targetType: reflect.TypeOf([]int{}),
			want:       []int{7},
		},
		{
			name:       "single string to string slice",
			value:      "hello",
			targetType: reflect.TypeOf([]string{}),
			want:       []string{"hello"},
		},

		// TextUnmarshaler
		{
			name:       "TextUnmarshaler success",
			value:      "hello",
			targetType: reflect.TypeOf(testTextUnmarshaler{}),
			want:       testTextUnmarshaler{Value: "hello"},
		},
		{
			name:       "TextUnmarshaler failure",
			value:      "FAIL",
			targetType: reflect.TypeOf(testTextUnmarshaler{}),
			wantErr:    "unmarshal refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.value, tt.targetType)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestConvertSliceEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		csv       string
		sliceType reflect.Type
		want      interface{}
		wantErr   string
	}{
		{
			name:      "empty string returns empty slice",
			csv:       "",
			sliceType: reflect.TypeOf([]int{}),
			want:      []int{},
		},
		{
			name:      "whitespace-only elements skipped",
			csv:       " , , ",
			sliceType: reflect.TypeOf([]string{}),
			want:      []string{},
		},
		{
			name:      "trailing comma skipped",
			csv:       "1,2,",
			sliceType: reflect.TypeOf([]int{}),
			want:      []int{1, 2},
		},
		{
			name:      "leading comma skipped",
			csv:       ",1,2",
			sliceType: reflect.TypeOf([]int{}),
			want:      []int{1, 2},
		},
		{
			name:      "whitespace trimmed around elements",
			csv:       " a , b , c ",
			sliceType: reflect.TypeOf([]string{}),
			want:      []string{"a", "b", "c"},
		},
		{
			name:      "non-slice type errors",
			csv:       "1,2,3",
			sliceType: reflect.TypeOf(0),
			wantErr:   "unsupported type: int",
		},
		{
			name:      "invalid element in slice errors",
			csv:       "1,abc,3",
			sliceType: reflect.TypeOf([]int{}),
			wantErr:   `invalid value "abc" for type int`,
		},
		{
			name:      "single element",
			csv:       "42",
			sliceType: reflect.TypeOf([]int{}),
			want:      []int{42},
		},
		{
			name:      "bool slice",
			csv:       "true,false,yes,no",
			sliceType: reflect.TypeOf([]bool{}),
			want:      []bool{true, false, true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertSlice(tt.csv, tt.sliceType)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

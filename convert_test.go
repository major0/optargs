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

func TestPropertyTypeConversionRoundTrip(t *testing.T) {
	types := []struct {
		name string
		gen  func(*quick.Config) error
	}{
		{"string", func(cfg *quick.Config) error {
			return quick.Check(func(v string) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"bool", func(cfg *quick.Config) error {
			return quick.Check(func(v bool) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"int", func(cfg *quick.Config) error {
			return quick.Check(func(v int) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"int8", func(cfg *quick.Config) error {
			return quick.Check(func(v int8) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"int16", func(cfg *quick.Config) error {
			return quick.Check(func(v int16) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"int32", func(cfg *quick.Config) error {
			return quick.Check(func(v int32) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"int64", func(cfg *quick.Config) error {
			return quick.Check(func(v int64) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"uint", func(cfg *quick.Config) error {
			return quick.Check(func(v uint) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"uint8", func(cfg *quick.Config) error {
			return quick.Check(func(v uint8) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"uint16", func(cfg *quick.Config) error {
			return quick.Check(func(v uint16) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"uint32", func(cfg *quick.Config) error {
			return quick.Check(func(v uint32) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"uint64", func(cfg *quick.Config) error {
			return quick.Check(func(v uint64) bool {
				got, err := Convert(fmt.Sprint(v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"float32", func(cfg *quick.Config) error {
			return quick.Check(func(v float32) bool {
				if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
					return true
				}
				got, err := Convert(fmt.Sprintf("%v", v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
		{"float64", func(cfg *quick.Config) error {
			return quick.Check(func(v float64) bool {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					return true
				}
				got, err := Convert(fmt.Sprintf("%v", v), reflect.TypeOf(v))
				return err == nil && got == v
			}, cfg)
		}},
	}

	cfg := &quick.Config{MaxCount: 100}
	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.gen(cfg); err != nil {
				t.Error(err)
			}
		})
	}
}

// Feature: goarg-optargs-integration, Property 2: ConvertSlice round-trip
// Validates: Requirements 2.3
//
// For any slice of a supported element type, joining elements with commas
// → ConvertSlice produces the original slice.

func TestPropertyConvertSliceRoundTrip(t *testing.T) {
	types := []struct {
		name string
		gen  func(*quick.Config) error
	}{
		{"string", func(cfg *quick.Config) error {
			return quick.Check(func(vs []string) bool {
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
					return false
				}
				return reflect.DeepEqual(got, clean)
			}, cfg)
		}},
		{"bool", func(cfg *quick.Config) error {
			return quick.Check(func(vs []bool) bool {
				if len(vs) == 0 {
					return true
				}
				parts := make([]string, len(vs))
				for i, v := range vs {
					parts[i] = fmt.Sprint(v)
				}
				got, err := ConvertSlice(strings.Join(parts, ","), reflect.TypeOf(vs))
				if err != nil {
					return false
				}
				return reflect.DeepEqual(got, vs)
			}, cfg)
		}},
		{"int", func(cfg *quick.Config) error {
			return quick.Check(func(vs []int) bool {
				if len(vs) == 0 {
					return true
				}
				parts := make([]string, len(vs))
				for i, v := range vs {
					parts[i] = fmt.Sprint(v)
				}
				got, err := ConvertSlice(strings.Join(parts, ","), reflect.TypeOf(vs))
				if err != nil {
					return false
				}
				return reflect.DeepEqual(got, vs)
			}, cfg)
		}},
		{"int64", func(cfg *quick.Config) error {
			return quick.Check(func(vs []int64) bool {
				if len(vs) == 0 {
					return true
				}
				parts := make([]string, len(vs))
				for i, v := range vs {
					parts[i] = fmt.Sprint(v)
				}
				got, err := ConvertSlice(strings.Join(parts, ","), reflect.TypeOf(vs))
				if err != nil {
					return false
				}
				return reflect.DeepEqual(got, vs)
			}, cfg)
		}},
		{"uint", func(cfg *quick.Config) error {
			return quick.Check(func(vs []uint) bool {
				if len(vs) == 0 {
					return true
				}
				parts := make([]string, len(vs))
				for i, v := range vs {
					parts[i] = fmt.Sprint(v)
				}
				got, err := ConvertSlice(strings.Join(parts, ","), reflect.TypeOf(vs))
				if err != nil {
					return false
				}
				return reflect.DeepEqual(got, vs)
			}, cfg)
		}},
		{"float64", func(cfg *quick.Config) error {
			return quick.Check(func(vs []float64) bool {
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
				got, err := ConvertSlice(strings.Join(parts, ","), reflect.TypeOf(clean))
				if err != nil {
					return false
				}
				return reflect.DeepEqual(got, clean)
			}, cfg)
		}},
	}

	cfg := &quick.Config{MaxCount: 100}
	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.gen(cfg); err != nil {
				t.Error(err)
			}
		})
	}
}

// Feature: goarg-optargs-integration, Property 3: Type conversion error messages contain type and value
// Validates: Requirements 2.4
//
// For any invalid string/type pair, the error message contains both the
// type name and the input string.

func TestPropertyConvertErrorMessage(t *testing.T) {
	types := []struct {
		name       string
		targetType reflect.Type
		typeName   string // substring to check in error
	}{
		{"int", reflect.TypeOf(int(0)), "int"},
		{"uint", reflect.TypeOf(uint(0)), "uint"},
		{"bool", reflect.TypeOf(false), "bool"},
		{"float64", reflect.TypeOf(float64(0)), "float64"},
	}

	cfg := &quick.Config{MaxCount: 100}
	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			f := func(s string) bool {
				_, err := Convert(s, tt.targetType)
				if err == nil {
					return true // valid input, skip
				}
				msg := err.Error()
				quoted := fmt.Sprintf("%q", s)
				return strings.Contains(msg, tt.typeName) && strings.Contains(msg, quoted)
			}
			if err := quick.Check(f, cfg); err != nil {
				t.Error(err)
			}
		})
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
			wantErr:    "invalid value",
		},
		{
			name:       "empty string to float64 errors",
			value:      "",
			targetType: reflect.TypeOf(float64(0)),
			wantErr:    "invalid value",
		},

		// Overflow values
		{
			name:       "256 overflows int8",
			value:      "256",
			targetType: reflect.TypeOf(int8(0)),
			wantErr:    "invalid value",
		},
		{
			name:       "-129 overflows int8",
			value:      "-129",
			targetType: reflect.TypeOf(int8(0)),
			wantErr:    "invalid value",
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
			wantErr:    "invalid value",
		},
		{
			name:       "negative overflows uint",
			value:      "-1",
			targetType: reflect.TypeOf(uint(0)),
			wantErr:    "invalid value",
		},

		// Unsupported types
		{
			name:       "chan type unsupported",
			value:      "x",
			targetType: reflect.TypeOf(make(chan int)),
			wantErr:    "unsupported type",
		},
		{
			name:       "func type unsupported",
			value:      "x",
			targetType: reflect.TypeOf(func() {}),
			wantErr:    "unsupported type",
		},
		{
			name:       "map type unsupported",
			value:      "x",
			targetType: reflect.TypeOf(map[string]int{}),
			wantErr:    "unsupported type",
		},
		{
			name:       "complex128 type unsupported",
			value:      "x",
			targetType: reflect.TypeOf(complex128(0)),
			wantErr:    "unsupported type",
		},

		// Bool extended aliases (case-insensitive variants not reachable by property round-trip)
		{"YES", "YES", reflect.TypeOf(false), true, ""},
		{"Yes", "Yes", reflect.TypeOf(false), true, ""},
		{"ON", "ON", reflect.TypeOf(false), true, ""},
		{"On", "On", reflect.TypeOf(false), true, ""},
		{"NO", "NO", reflect.TypeOf(false), false, ""},
		{"OFF", "OFF", reflect.TypeOf(false), false, ""},
		{"t", "t", reflect.TypeOf(false), true, ""},
		{"f", "f", reflect.TypeOf(false), false, ""},
		{"1", "1", reflect.TypeOf(false), true, ""},
		{"0", "0", reflect.TypeOf(false), false, ""},

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
			name:       "pointer to invalid int errors",
			value:      "abc",
			targetType: reflect.TypeOf((*int)(nil)),
			wantErr:    "invalid value",
		},

		// Slice via Convert (single value → single-element slice)
		{
			name:       "single int to int slice",
			value:      "7",
			targetType: reflect.TypeOf([]int{}),
			want:       []int{7},
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
		{"empty string returns empty slice", "", reflect.TypeOf([]int{}), []int{}, ""},
		{"whitespace-only elements skipped", " , , ", reflect.TypeOf([]string{}), []string{}, ""},
		{"trailing comma skipped", "1,2,", reflect.TypeOf([]int{}), []int{1, 2}, ""},
		{"leading comma skipped", ",1,2", reflect.TypeOf([]int{}), []int{1, 2}, ""},
		{"whitespace trimmed", " a , b , c ", reflect.TypeOf([]string{}), []string{"a", "b", "c"}, ""},
		{"non-slice type errors", "1,2,3", reflect.TypeOf(0), nil, "unsupported type"},
		{"invalid element errors", "1,abc,3", reflect.TypeOf([]int{}), nil, "invalid value"},
		{"single element", "42", reflect.TypeOf([]int{}), []int{42}, ""},
		{"bool slice", "true,false,yes,no", reflect.TypeOf([]bool{}), []bool{true, false, true, false}, ""},
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

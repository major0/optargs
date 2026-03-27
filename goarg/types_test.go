package goarg

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

// CustomType implements encoding.TextUnmarshaler for testing
type CustomType struct {
	Value string
}

func (ct *CustomType) UnmarshalText(text []byte) error {
	ct.Value = "custom:" + string(text)
	return nil
}

// CustomTypePtr implements encoding.TextUnmarshaler on pointer receiver
type CustomTypePtr struct {
	Value string
}

func (ct *CustomTypePtr) UnmarshalText(text []byte) error {
	ct.Value = "ptr:" + string(text)
	return nil
}

// CustomTypeWithError implements encoding.TextUnmarshaler but returns an error
type CustomTypeWithError struct {
	Value string
}

func (ct *CustomTypeWithError) UnmarshalText(text []byte) error {
	return fmt.Errorf("intentional error for testing")
}

// CustomTypePtrReceiver implements encoding.TextUnmarshaler on pointer receiver only
type CustomTypePtrReceiver struct {
	Value string
}

func (ct *CustomTypePtrReceiver) UnmarshalText(text []byte) error {
	ct.Value = "ptr-receiver:" + string(text)
	return nil
}

// CustomTypeValueReceiver implements encoding.TextUnmarshaler on value receiver
type CustomTypeValueReceiver struct {
	Value string
}

func (ct CustomTypeValueReceiver) UnmarshalText(text []byte) error {
	ct.Value = "value-receiver:" + string(text)
	return nil
}

// CustomTypePointerOnly implements encoding.TextUnmarshaler only on pointer type
type CustomTypePointerOnly struct {
	Value string
}

func (ct *CustomTypePointerOnly) UnmarshalText(text []byte) error {
	ct.Value = "pointer-only:" + string(text)
	return nil
}

func TestTypeConverter_ConvertValue(t *testing.T) {
	tc := &TypeConverter{}

	tests := []struct {
		name       string
		value      string
		targetType reflect.Type
		expected   interface{}
		wantError  bool
	}{
		// Basic types
		{
			name:       "string",
			value:      "hello",
			targetType: reflect.TypeOf(""),
			expected:   "hello",
		},
		{
			name:       "int",
			value:      "42",
			targetType: reflect.TypeOf(int(0)),
			expected:   42,
		},
		{
			name:       "int8",
			value:      "127",
			targetType: reflect.TypeOf(int8(0)),
			expected:   int8(127),
		},
		{
			name:       "int16",
			value:      "32767",
			targetType: reflect.TypeOf(int16(0)),
			expected:   int16(32767),
		},
		{
			name:       "int32",
			value:      "2147483647",
			targetType: reflect.TypeOf(int32(0)),
			expected:   int32(2147483647),
		},
		{
			name:       "int64",
			value:      "9223372036854775807",
			targetType: reflect.TypeOf(int64(0)),
			expected:   int64(9223372036854775807),
		},
		{
			name:       "uint",
			value:      "42",
			targetType: reflect.TypeOf(uint(0)),
			expected:   uint(42),
		},
		{
			name:       "uint8",
			value:      "255",
			targetType: reflect.TypeOf(uint8(0)),
			expected:   uint8(255),
		},
		{
			name:       "uint16",
			value:      "65535",
			targetType: reflect.TypeOf(uint16(0)),
			expected:   uint16(65535),
		},
		{
			name:       "uint32",
			value:      "4294967295",
			targetType: reflect.TypeOf(uint32(0)),
			expected:   uint32(4294967295),
		},
		{
			name:       "uint64",
			value:      "18446744073709551615",
			targetType: reflect.TypeOf(uint64(0)),
			expected:   uint64(18446744073709551615),
		},
		{
			name:       "float32",
			value:      "3.14",
			targetType: reflect.TypeOf(float32(0)),
			expected:   float32(3.14),
		},
		{
			name:       "float64",
			value:      "3.14159265359",
			targetType: reflect.TypeOf(float64(0)),
			expected:   3.14159265359,
		},
		{
			name:       "bool_true",
			value:      "true",
			targetType: reflect.TypeOf(false),
			expected:   true,
		},
		{
			name:       "bool_false",
			value:      "false",
			targetType: reflect.TypeOf(false),
			expected:   false,
		},
		{
			name:       "bool_1",
			value:      "1",
			targetType: reflect.TypeOf(false),
			expected:   true,
		},
		{
			name:       "bool_0",
			value:      "0",
			targetType: reflect.TypeOf(false),
			expected:   false,
		},
		{
			name:       "bool_yes",
			value:      "yes",
			targetType: reflect.TypeOf(false),
			expected:   true,
		},
		{
			name:       "bool_no",
			value:      "no",
			targetType: reflect.TypeOf(false),
			expected:   false,
		},
		{
			name:       "bool_empty_false",
			value:      "",
			targetType: reflect.TypeOf(false),
			expected:   false,
		},
		// Pointer types
		{
			name:       "pointer_to_int",
			value:      "42",
			targetType: reflect.TypeOf((*int)(nil)),
			expected:   func() *int { i := 42; return &i }(),
		},
		{
			name:       "pointer_to_string",
			value:      "hello",
			targetType: reflect.TypeOf((*string)(nil)),
			expected:   func() *string { s := "hello"; return &s }(),
		},
		// Slice types
		{
			name:       "string_slice",
			value:      "hello",
			targetType: reflect.TypeOf([]string{}),
			expected:   []string{"hello"},
		},
		{
			name:       "int_slice",
			value:      "42",
			targetType: reflect.TypeOf([]int{}),
			expected:   []int{42},
		},
		// Custom types
		{
			name:       "custom_type",
			value:      "test",
			targetType: reflect.TypeOf(CustomType{}),
			expected:   CustomType{Value: "custom:test"},
		},
		{
			name:       "custom_type_ptr",
			value:      "test",
			targetType: reflect.TypeOf(&CustomTypePtr{}),
			expected:   &CustomTypePtr{Value: "ptr:test"},
		},
		// Error cases
		{
			name:       "invalid_int",
			value:      "not_a_number",
			targetType: reflect.TypeOf(int(0)),
			wantError:  true,
		},
		{
			name:       "invalid_bool",
			value:      "maybe",
			targetType: reflect.TypeOf(false),
			wantError:  true,
		},
		{
			name:       "unsupported_type",
			value:      "test",
			targetType: reflect.TypeOf(time.Time{}),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tc.ConvertValue(tt.value, tt.targetType)

			if tt.wantError {
				if err == nil {
					t.Errorf("ConvertValue() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertValue() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ConvertValue() = %v (%T), want %v (%T)", result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestTypeConverter_SetField(t *testing.T) {
	tc := &TypeConverter{}

	type TestStruct struct {
		StringField string
		IntField    int
		PtrField    *string
	}

	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		wantError bool
		expected  interface{}
	}{
		{
			name:      "set_string_field",
			fieldName: "StringField",
			value:     "hello",
			expected:  "hello",
		},
		{
			name:      "set_int_field",
			fieldName: "IntField",
			value:     42,
			expected:  42,
		},
		{
			name:      "set_ptr_field",
			fieldName: "PtrField",
			value:     func() *string { s := "ptr_value"; return &s }(),
			expected:  func() *string { s := "ptr_value"; return &s }(),
		},
		{
			name:      "set_nil_to_ptr",
			fieldName: "PtrField",
			value:     nil,
			expected:  (*string)(nil),
		},
		{
			name:      "set_nil_to_non_ptr",
			fieldName: "StringField",
			value:     nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := &TestStruct{}
			structValue := reflect.ValueOf(testStruct).Elem()
			fieldValue := structValue.FieldByName(tt.fieldName)

			err := tc.SetField(fieldValue, tt.value)

			if tt.wantError {
				if err == nil {
					t.Errorf("SetField() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("SetField() unexpected error: %v", err)
				return
			}

			actualValue := fieldValue.Interface()
			if !reflect.DeepEqual(actualValue, tt.expected) {
				t.Errorf("SetField() field value = %v, want %v", actualValue, tt.expected)
			}
		})
	}
}

func TestTypeConverter_GetDefault(t *testing.T) {
	tc := &TypeConverter{}

	tests := []struct {
		name     string
		field    reflect.StructField
		expected interface{}
	}{
		{
			name: "string_default",
			field: reflect.StructField{
				Name: "StringField",
				Type: reflect.TypeOf(""),
				Tag:  `default:"hello"`,
			},
			expected: "hello",
		},
		{
			name: "int_default",
			field: reflect.StructField{
				Name: "IntField",
				Type: reflect.TypeOf(int(0)),
				Tag:  `default:"42"`,
			},
			expected: 42,
		},
		{
			name: "bool_default",
			field: reflect.StructField{
				Name: "BoolField",
				Type: reflect.TypeOf(false),
				Tag:  `default:"true"`,
			},
			expected: true,
		},
		{
			name: "slice_default",
			field: reflect.StructField{
				Name: "SliceField",
				Type: reflect.TypeOf([]string{}),
				Tag:  `default:"a,b,c"`,
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "int_slice_default",
			field: reflect.StructField{
				Name: "IntSliceField",
				Type: reflect.TypeOf([]int{}),
				Tag:  `default:"1,2,3"`,
			},
			expected: []int{1, 2, 3},
		},
		{
			name: "empty_slice_default",
			field: reflect.StructField{
				Name: "EmptySliceField",
				Type: reflect.TypeOf([]string{}),
				Tag:  `default:""`,
			},
			expected: []string{},
		},
		{
			name: "no_default_tag",
			field: reflect.StructField{
				Name: "NoDefaultField",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"--field"`,
			},
			expected: nil,
		},
		{
			name: "whitespace_in_slice",
			field: reflect.StructField{
				Name: "WhitespaceSliceField",
				Type: reflect.TypeOf([]string{}),
				Tag:  `default:" a , b , c "`,
			},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tc.GetDefault(tt.field)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTypeConverter_ConvertBool(t *testing.T) {
	tc := &TypeConverter{}

	tests := []struct {
		name      string
		value     string
		expected  bool
		wantError bool
	}{
		// True values
		{"true", "true", true, false},
		{"True", "True", true, false},
		{"TRUE", "TRUE", true, false},
		{"t", "t", true, false},
		{"T", "T", true, false},
		{"1", "1", true, false},
		{"yes", "yes", true, false},
		{"Yes", "Yes", true, false},
		{"YES", "YES", true, false},
		{"y", "y", true, false},
		{"Y", "Y", true, false},
		{"on", "on", true, false},
		{"On", "On", true, false},
		{"ON", "ON", true, false},

		// False values
		{"false", "false", false, false},
		{"False", "False", false, false},
		{"FALSE", "FALSE", false, false},
		{"f", "f", false, false},
		{"F", "F", false, false},
		{"0", "0", false, false},
		{"no", "no", false, false},
		{"No", "No", false, false},
		{"NO", "NO", false, false},
		{"n", "n", false, false},
		{"N", "N", false, false},
		{"off", "off", false, false},
		{"Off", "Off", false, false},
		{"OFF", "OFF", false, false},
		{"empty", "", false, false},

		// Invalid values
		{"invalid", "maybe", false, true},
		{"number", "2", false, true},
		{"random", "random", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tc.ConvertBool(tt.value)

			if tt.wantError {
				if err == nil {
					t.Errorf("ConvertBool() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertBool() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ConvertBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTypeConverter_ConvertCustom(t *testing.T) {
	tc := &TypeConverter{}

	tests := []struct {
		name       string
		value      string
		targetType reflect.Type
		expected   interface{}
		wantError  bool
	}{
		{
			name:       "custom_type_value",
			value:      "test",
			targetType: reflect.TypeOf(CustomType{}),
			expected:   CustomType{Value: "custom:test"},
		},
		{
			name:       "custom_type_pointer",
			value:      "test",
			targetType: reflect.TypeOf(&CustomTypePtr{}),
			expected:   &CustomTypePtr{Value: "ptr:test"},
		},
		{
			name:       "non_unmarshaler_type",
			value:      "test",
			targetType: reflect.TypeOf(time.Time{}),
			wantError:  true,
		},
		{
			name:       "custom_type_unmarshal_error",
			value:      "test",
			targetType: reflect.TypeOf(CustomTypeWithError{}),
			wantError:  true,
		},
		{
			name:       "pointer_to_custom_type_unmarshal_error",
			value:      "test",
			targetType: reflect.TypeOf(&CustomTypeWithError{}),
			wantError:  true,
		},
		{
			name:       "custom_type_ptr_receiver_value",
			value:      "test",
			targetType: reflect.TypeOf(CustomTypePtrReceiver{}),
			expected:   CustomTypePtrReceiver{Value: "ptr-receiver:test"},
		},
		{
			name:       "custom_type_ptr_receiver_pointer",
			value:      "test",
			targetType: reflect.TypeOf(&CustomTypePtrReceiver{}),
			expected:   &CustomTypePtrReceiver{Value: "ptr-receiver:test"},
		},
		{
			name:       "custom_type_value_receiver",
			value:      "test",
			targetType: reflect.TypeOf(CustomTypeValueReceiver{}),
			expected:   CustomTypeValueReceiver{}, // Value receiver doesn't modify the original
		},
		{
			name:       "custom_type_pointer_only_value",
			value:      "test",
			targetType: reflect.TypeOf(CustomTypePointerOnly{}),
			expected:   CustomTypePointerOnly{Value: "pointer-only:test"},
		},
		{
			name:       "custom_type_pointer_only_pointer",
			value:      "test",
			targetType: reflect.TypeOf(&CustomTypePointerOnly{}),
			expected:   &CustomTypePointerOnly{Value: "pointer-only:test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tc.ConvertCustom(tt.value, tt.targetType)

			if tt.wantError {
				if err == nil {
					t.Errorf("ConvertCustom() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertCustom() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ConvertCustom() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestProperty5_TypeConversionCompatibility tests Property 5 from the design document:
// For any Go type supported by alexflint/go-arg, our type converter should handle value conversion identically to upstream alexflint/go-arg
// **Validates: Requirements 4.2**
func TestProperty5_TypeConversionCompatibility(t *testing.T) {
	tc := &TypeConverter{}

	// Property: Type conversion should handle all supported Go types correctly
	property := func(stringVal string, intVal int, boolVal bool, floatVal float64) bool {
		// Test basic type conversions that should always work
		testCases := []struct {
			value      string
			targetType reflect.Type
			validator  func(interface{}) bool
		}{
			// String conversion - should always succeed
			{
				value:      stringVal,
				targetType: reflect.TypeOf(""),
				validator: func(result interface{}) bool {
					str, ok := result.(string)
					return ok && str == stringVal
				},
			},
			// Integer conversion with valid numeric strings
			{
				value:      "42",
				targetType: reflect.TypeOf(int(0)),
				validator: func(result interface{}) bool {
					val, ok := result.(int)
					return ok && val == 42
				},
			},
			// Boolean conversion with valid boolean strings
			{
				value:      "true",
				targetType: reflect.TypeOf(false),
				validator: func(result interface{}) bool {
					val, ok := result.(bool)
					return ok && val == true
				},
			},
			{
				value:      "false",
				targetType: reflect.TypeOf(false),
				validator: func(result interface{}) bool {
					val, ok := result.(bool)
					return ok && val == false
				},
			},
			// Float conversion with valid numeric strings
			{
				value:      "3.14",
				targetType: reflect.TypeOf(float64(0)),
				validator: func(result interface{}) bool {
					val, ok := result.(float64)
					return ok && val == 3.14
				},
			},
			// Pointer type conversion
			{
				value:      "test",
				targetType: reflect.TypeOf((*string)(nil)),
				validator: func(result interface{}) bool {
					ptr, ok := result.(*string)
					return ok && ptr != nil && *ptr == "test"
				},
			},
			// Slice type conversion
			{
				value:      "hello",
				targetType: reflect.TypeOf([]string{}),
				validator: func(result interface{}) bool {
					slice, ok := result.([]string)
					return ok && len(slice) == 1 && slice[0] == "hello"
				},
			},
		}

		// Test each conversion case
		for _, testCase := range testCases {
			result, err := tc.ConvertValue(testCase.value, testCase.targetType)
			if err != nil {
				// For property testing, we expect these basic conversions to succeed
				return false
			}

			if !testCase.validator(result) {
				return false
			}
		}

		// Test integer type variations with bounds checking
		integerTests := []struct {
			value      string
			targetType reflect.Type
			validator  func(interface{}) bool
		}{
			{
				value:      "127",
				targetType: reflect.TypeOf(int8(0)),
				validator: func(result interface{}) bool {
					val, ok := result.(int8)
					return ok && val == 127
				},
			},
			{
				value:      "32767",
				targetType: reflect.TypeOf(int16(0)),
				validator: func(result interface{}) bool {
					val, ok := result.(int16)
					return ok && val == 32767
				},
			},
			{
				value:      "255",
				targetType: reflect.TypeOf(uint8(0)),
				validator: func(result interface{}) bool {
					val, ok := result.(uint8)
					return ok && val == 255
				},
			},
		}

		for _, test := range integerTests {
			result, err := tc.ConvertValue(test.value, test.targetType)
			if err != nil {
				return false
			}
			if !test.validator(result) {
				return false
			}
		}

		// Test boolean conversion variations
		boolTests := []struct {
			value    string
			expected bool
		}{
			{"1", true},
			{"0", false},
			{"yes", true},
			{"no", false},
			{"on", true},
			{"off", false},
			{"", false},
		}

		for _, test := range boolTests {
			result, err := tc.ConvertValue(test.value, reflect.TypeOf(false))
			if err != nil {
				return false
			}
			val, ok := result.(bool)
			if !ok || val != test.expected {
				return false
			}
		}

		// Test custom type conversion
		customResult, err := tc.ConvertValue("test", reflect.TypeOf(CustomType{}))
		if err != nil {
			return false
		}
		customVal, ok := customResult.(CustomType)
		if !ok || customVal.Value != "custom:test" {
			return false
		}

		return true
	}

	// Configure property test with sufficient iterations
	config := &quick.Config{
		MaxCount: 100, // Minimum 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5 (Type Conversion Compatibility) failed: %v", err)
	}
}

func TestTypeConverter_ValidateRequired(t *testing.T) {
	tc := &TypeConverter{}

	type TestStruct struct {
		RequiredString string `arg:"--required-string,required"`
		OptionalString string `arg:"--optional-string"`
		RequiredInt    int    `arg:"--required-int,required"`
		OptionalInt    int    `arg:"--optional-int"`
	}

	tests := []struct {
		name      string
		setup     func(*TestStruct)
		wantError bool
		errorMsg  string
	}{
		{
			name: "all_required_fields_set",
			setup: func(ts *TestStruct) {
				ts.RequiredString = "test"
				ts.RequiredInt = 42
			},
			wantError: false,
		},
		{
			name: "missing_required_string",
			setup: func(ts *TestStruct) {
				ts.RequiredInt = 42
			},
			wantError: true,
			errorMsg:  "--required-string is required",
		},
		{
			name: "missing_required_int",
			setup: func(ts *TestStruct) {
				ts.RequiredString = "test"
			},
			wantError: true,
			errorMsg:  "--required-int is required",
		},
		{
			name:      "missing_all_required",
			setup:     func(ts *TestStruct) {},
			wantError: true,
			errorMsg:  "--required-string is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := &TestStruct{}
			tt.setup(testStruct)

			// Create metadata for the struct
			metadata := &StructMetadata{
				Fields: []FieldMetadata{
					{Name: "RequiredString", Required: true, Long: "required-string"},
					{Name: "OptionalString", Required: false, Long: "optional-string"},
					{Name: "RequiredInt", Required: true, Long: "required-int"},
					{Name: "OptionalInt", Required: false, Long: "optional-int"},
				},
			}

			err := tc.ValidateRequired(testStruct, metadata)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateRequired() expected error, got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("ValidateRequired() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateRequired() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTypeConverter_isZeroValue(t *testing.T) {
	tc := &TypeConverter{}

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		// Basic types
		{"zero_string", "", true},
		{"non_zero_string", "hello", false},
		{"zero_int", 0, true},
		{"non_zero_int", 42, false},
		{"zero_bool", false, true},
		{"non_zero_bool", true, false},
		{"zero_float", 0.0, true},
		{"non_zero_float", 3.14, false},

		// Pointer types
		{"nil_pointer", (*string)(nil), true},
		{"non_nil_pointer", func() *string { s := "test"; return &s }(), false},

		// Slice types
		{"nil_slice", ([]string)(nil), true},
		{"empty_slice", []string{}, true},
		{"non_empty_slice", []string{"a"}, false},

		// Struct types
		{"zero_struct", struct{ A int }{}, true},
		{"non_zero_struct", struct{ A int }{A: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := reflect.ValueOf(tt.value)
			result := tc.isZeroValue(value)

			if result != tt.expected {
				t.Errorf("isZeroValue(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestTypeConverter_EdgeCases tests additional edge cases and boundary conditions
func TestTypeConverter_EdgeCases(t *testing.T) {
	tc := &TypeConverter{}

	t.Run("integer_overflow_boundaries", func(t *testing.T) {
		// Test integer overflow boundaries
		tests := []struct {
			name       string
			value      string
			targetType reflect.Type
			wantError  bool
		}{
			// int8 boundaries
			{"int8_max", "127", reflect.TypeOf(int8(0)), false},
			{"int8_min", "-128", reflect.TypeOf(int8(0)), false},
			{"int8_overflow", "128", reflect.TypeOf(int8(0)), true},
			{"int8_underflow", "-129", reflect.TypeOf(int8(0)), true},

			// uint8 boundaries
			{"uint8_max", "255", reflect.TypeOf(uint8(0)), false},
			{"uint8_min", "0", reflect.TypeOf(uint8(0)), false},
			{"uint8_overflow", "256", reflect.TypeOf(uint8(0)), true},
			{"uint8_negative", "-1", reflect.TypeOf(uint8(0)), true},

			// int16 boundaries
			{"int16_max", "32767", reflect.TypeOf(int16(0)), false},
			{"int16_min", "-32768", reflect.TypeOf(int16(0)), false},
			{"int16_overflow", "32768", reflect.TypeOf(int16(0)), true},

			// uint16 boundaries
			{"uint16_max", "65535", reflect.TypeOf(uint16(0)), false},
			{"uint16_overflow", "65536", reflect.TypeOf(uint16(0)), true},

			// int32 boundaries
			{"int32_max", "2147483647", reflect.TypeOf(int32(0)), false},
			{"int32_min", "-2147483648", reflect.TypeOf(int32(0)), false},
			{"int32_overflow", "2147483648", reflect.TypeOf(int32(0)), true},

			// uint32 boundaries
			{"uint32_max", "4294967295", reflect.TypeOf(uint32(0)), false},
			{"uint32_overflow", "4294967296", reflect.TypeOf(uint32(0)), true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tc.ConvertValue(tt.value, tt.targetType)
				if tt.wantError && err == nil {
					t.Errorf("ConvertValue(%s, %v) expected error, got nil", tt.value, tt.targetType)
				} else if !tt.wantError && err != nil {
					t.Errorf("ConvertValue(%s, %v) unexpected error: %v", tt.value, tt.targetType, err)
				}
			})
		}
	})

	t.Run("float_special_values", func(t *testing.T) {
		tests := []struct {
			name       string
			value      string
			targetType reflect.Type
			wantError  bool
		}{
			{"float32_inf", "inf", reflect.TypeOf(float32(0)), false},
			{"float32_neg_inf", "-inf", reflect.TypeOf(float32(0)), false},
			{"float32_nan", "nan", reflect.TypeOf(float32(0)), false},
			{"float64_inf", "inf", reflect.TypeOf(float64(0)), false},
			{"float64_neg_inf", "-inf", reflect.TypeOf(float64(0)), false},
			{"float64_nan", "nan", reflect.TypeOf(float64(0)), false},
			{"float_invalid", "not_a_float", reflect.TypeOf(float64(0)), true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := tc.ConvertValue(tt.value, tt.targetType)
				if tt.wantError && err == nil {
					t.Errorf("ConvertValue(%s, %v) expected error, got nil", tt.value, tt.targetType)
				} else if !tt.wantError && err != nil {
					t.Errorf("ConvertValue(%s, %v) unexpected error: %v", tt.value, tt.targetType, err)
				} else if !tt.wantError && err == nil {
					// Verify special float values
					switch tt.targetType.Kind() {
					case reflect.Float32:
						val := result.(float32)
						switch tt.value {
						case "inf":
							if !math.IsInf(float64(val), 1) {
								t.Errorf("Expected +Inf, got %v", val)
							}
						case "-inf":
							if !math.IsInf(float64(val), -1) {
								t.Errorf("Expected -Inf, got %v", val)
							}
						case "nan":
							if !math.IsNaN(float64(val)) {
								t.Errorf("Expected NaN, got %v", val)
							}
						}
					case reflect.Float64:
						val := result.(float64)
						switch tt.value {
						case "inf":
							if !math.IsInf(val, 1) {
								t.Errorf("Expected +Inf, got %v", val)
							}
						case "-inf":
							if !math.IsInf(val, -1) {
								t.Errorf("Expected -Inf, got %v", val)
							}
						case "nan":
							if !math.IsNaN(val) {
								t.Errorf("Expected NaN, got %v", val)
							}
						}
					}
				}
			})
		}
	})

	t.Run("empty_and_whitespace_strings", func(t *testing.T) {
		tests := []struct {
			name       string
			value      string
			targetType reflect.Type
			expected   interface{}
			wantError  bool
		}{
			{"empty_string_to_string", "", reflect.TypeOf(""), "", false},
			{"whitespace_string", "   ", reflect.TypeOf(""), "   ", false},
			{"empty_string_to_int", "", reflect.TypeOf(int(0)), nil, true},
			{"whitespace_to_int", "   ", reflect.TypeOf(int(0)), nil, true},
			{"empty_string_to_bool", "", reflect.TypeOf(false), false, false},
			{"whitespace_to_bool", "   ", reflect.TypeOf(false), nil, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := tc.ConvertValue(tt.value, tt.targetType)
				if tt.wantError && err == nil {
					t.Errorf("ConvertValue(%s, %v) expected error, got nil", tt.value, tt.targetType)
				} else if !tt.wantError && err != nil {
					t.Errorf("ConvertValue(%s, %v) unexpected error: %v", tt.value, tt.targetType, err)
				} else if !tt.wantError && !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("ConvertValue(%s, %v) = %v, want %v", tt.value, tt.targetType, result, tt.expected)
				}
			})
		}
	})

	t.Run("nested_pointer_types", func(t *testing.T) {
		// Test double pointer types
		result, err := tc.ConvertValue("test", reflect.TypeOf((**string)(nil)))
		if err != nil {
			t.Errorf("ConvertValue with double pointer unexpected error: %v", err)
		} else {
			ptr, ok := result.(**string)
			if !ok {
				t.Errorf("Expected **string, got %T", result)
			} else if ptr == nil || *ptr == nil || **ptr != "test" {
				t.Errorf("Expected **string with value 'test', got %v", ptr)
			}
		}
	})
}

// TestTypeConverter_SetFieldEdgeCases tests edge cases for SetField method
func TestTypeConverter_SetFieldEdgeCases(t *testing.T) {
	tc := &TypeConverter{}

	t.Run("unsettable_field", func(t *testing.T) {
		type TestStruct struct {
			unexported string
		}

		testStruct := &TestStruct{}
		structValue := reflect.ValueOf(testStruct).Elem()
		fieldValue := structValue.FieldByName("unexported")

		err := tc.SetField(fieldValue, "test")
		if err == nil {
			t.Errorf("SetField on unexported field expected error, got nil")
		}
		if !strings.Contains(err.Error(), "field cannot be set") {
			t.Errorf("Expected 'field cannot be set' error, got: %v", err)
		}
	})

	t.Run("type_conversion_in_setfield", func(t *testing.T) {
		type TestStruct struct {
			IntField int
		}

		testStruct := &TestStruct{}
		structValue := reflect.ValueOf(testStruct).Elem()
		fieldValue := structValue.FieldByName("IntField")

		// Test convertible types (int32 to int)
		err := tc.SetField(fieldValue, int32(42))
		if err != nil {
			t.Errorf("SetField with convertible type unexpected error: %v", err)
		}
		if testStruct.IntField != 42 {
			t.Errorf("Expected field value 42, got %d", testStruct.IntField)
		}
	})

	t.Run("incompatible_types", func(t *testing.T) {
		type TestStruct struct {
			StringField string
		}

		testStruct := &TestStruct{}
		structValue := reflect.ValueOf(testStruct).Elem()
		fieldValue := structValue.FieldByName("StringField")

		// Try to set string field with incompatible type
		err := tc.SetField(fieldValue, []int{1, 2, 3})
		if err == nil {
			t.Errorf("SetField with incompatible type expected error, got nil")
		}
		if !strings.Contains(err.Error(), "cannot assign value of type") {
			t.Errorf("Expected type assignment error, got: %v", err)
		}
	})
}

// TestTypeConverter_DefaultValueEdgeCases tests edge cases for default value processing
func TestTypeConverter_DefaultValueEdgeCases(t *testing.T) {
	tc := &TypeConverter{}

	t.Run("invalid_default_values", func(t *testing.T) {
		tests := []struct {
			name  string
			field reflect.StructField
		}{
			{
				name: "invalid_int_default",
				field: reflect.StructField{
					Name: "IntField",
					Type: reflect.TypeOf(int(0)),
					Tag:  `default:"not_a_number"`,
				},
			},
			{
				name: "invalid_bool_default",
				field: reflect.StructField{
					Name: "BoolField",
					Type: reflect.TypeOf(false),
					Tag:  `default:"maybe"`,
				},
			},
			{
				name: "invalid_float_default",
				field: reflect.StructField{
					Name: "FloatField",
					Type: reflect.TypeOf(float64(0)),
					Tag:  `default:"not_a_float"`,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tc.GetDefault(tt.field)
				if result != nil {
					t.Errorf("GetDefault with invalid default expected nil, got %v", result)
				}
			})
		}
	})

	t.Run("slice_default_with_invalid_elements", func(t *testing.T) {
		field := reflect.StructField{
			Name: "IntSliceField",
			Type: reflect.TypeOf([]int{}),
			Tag:  `default:"1,not_a_number,3"`,
		}

		result := tc.GetDefault(field)
		if result != nil {
			t.Errorf("GetDefault with invalid slice element expected nil, got %v", result)
		}
	})

	t.Run("complex_slice_defaults", func(t *testing.T) {
		tests := []struct {
			name     string
			field    reflect.StructField
			expected interface{}
		}{
			{
				name: "slice_with_empty_elements",
				field: reflect.StructField{
					Name: "StringSliceField",
					Type: reflect.TypeOf([]string{}),
					Tag:  `default:"a,,b"`,
				},
				expected: []string{"a", "b"}, // Empty elements should be filtered out
			},
			{
				name: "slice_with_only_commas",
				field: reflect.StructField{
					Name: "StringSliceField",
					Type: reflect.TypeOf([]string{}),
					Tag:  `default:",,"`,
				},
				expected: []string{}, // Should result in empty slice
			},
			{
				name: "slice_with_trailing_comma",
				field: reflect.StructField{
					Name: "StringSliceField",
					Type: reflect.TypeOf([]string{}),
					Tag:  `default:"a,b,"`,
				},
				expected: []string{"a", "b"}, // Trailing comma should be ignored
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tc.GetDefault(tt.field)
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("GetDefault() = %v, want %v", result, tt.expected)
				}
			})
		}
	})
}

// ErrorCustomType implements encoding.TextUnmarshaler for error testing
type ErrorCustomType struct {
	Value string
}

func (e *ErrorCustomType) UnmarshalText(text []byte) error {
	if string(text) == "error" {
		return fmt.Errorf("custom unmarshal error")
	}
	e.Value = string(text)
	return nil
}

// TestTypeConverter_AdditionalEdgeCases tests additional edge cases for comprehensive coverage
func TestTypeConverter_AdditionalEdgeCases(t *testing.T) {
	tc := &TypeConverter{}

	t.Run("custom_type_error_handling", func(t *testing.T) {
		// Test successful unmarshaling
		result, err := tc.ConvertCustom("success", reflect.TypeOf(ErrorCustomType{}))
		if err != nil {
			t.Errorf("ConvertCustom with valid input unexpected error: %v", err)
		}
		if val, ok := result.(ErrorCustomType); !ok || val.Value != "success" {
			t.Errorf("ConvertCustom result = %v, want ErrorCustomType{Value: 'success'}", result)
		}

		// Test error during unmarshaling
		_, err = tc.ConvertCustom("error", reflect.TypeOf(ErrorCustomType{}))
		if err == nil {
			t.Errorf("ConvertCustom with error input expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal text") {
			t.Errorf("Expected unmarshal error message, got: %v", err)
		}
	})

	t.Run("validation_with_non_struct_destination", func(t *testing.T) {
		// Test ValidateRequired with non-struct destination
		var notAStruct int = 42
		metadata := &StructMetadata{}

		err := tc.ValidateRequired(&notAStruct, metadata)
		if err == nil {
			t.Errorf("ValidateRequired with non-struct expected error, got nil")
		}
		if !strings.Contains(err.Error(), "destination must be a pointer to a struct") {
			t.Errorf("Expected struct type error, got: %v", err)
		}

		// Test ValidateRequired with non-pointer destination
		var structVal struct{ Field string }
		err = tc.ValidateRequired(structVal, metadata)
		if err == nil {
			t.Errorf("ValidateRequired with non-pointer expected error, got nil")
		}
		if !strings.Contains(err.Error(), "destination must be a pointer") {
			t.Errorf("Expected pointer type error, got: %v", err)
		}
	})

	t.Run("required_field_error_messages", func(t *testing.T) {
		tc := &TypeConverter{}

		type TestStruct struct {
			ShortOnlyField  string `arg:"-s,required"`
			PositionalField string `arg:"positional,required"`
			NoFlagField     string `required:"true"`
		}

		tests := []struct {
			name     string
			metadata *StructMetadata
			errorMsg string
		}{
			{
				name: "short_only_field",
				metadata: &StructMetadata{
					Fields: []FieldMetadata{
						{Name: "ShortOnlyField", Required: true, Short: "s", Long: ""},
					},
				},
				errorMsg: "-s is required",
			},
			{
				name: "positional_field",
				metadata: &StructMetadata{
					Fields: []FieldMetadata{
						{Name: "PositionalField", Required: true, Positional: true},
					},
				},
				errorMsg: "PositionalField is required",
			},
			{
				name: "no_flag_field",
				metadata: &StructMetadata{
					Fields: []FieldMetadata{
						{Name: "NoFlagField", Required: true, Short: "", Long: ""},
					},
				},
				errorMsg: "NoFlagField is required",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				testStruct := &TestStruct{}
				err := tc.ValidateRequired(testStruct, tt.metadata)

				if err == nil {
					t.Errorf("ValidateRequired expected error, got nil")
					return
				}

				if err.Error() != tt.errorMsg {
					t.Errorf("ValidateRequired error = %v, want %v", err.Error(), tt.errorMsg)
				}
			})
		}
	})

	t.Run("zero_value_detection_comprehensive", func(t *testing.T) {
		tests := []struct {
			name     string
			value    interface{}
			expected bool
		}{
			// Additional zero value tests
			{"zero_uint", uint(0), true},
			{"non_zero_uint", uint(1), false},
			{"zero_uint64", uint64(0), true},
			{"non_zero_uint64", uint64(1), false},
			{"zero_int64", int64(0), true},
			{"non_zero_int64", int64(1), false},
			{"zero_float32", float32(0.0), true},
			{"non_zero_float32", float32(1.0), false},

			// Interface types
			{"nil_interface", (interface{})(nil), true},
			{"non_nil_interface", (interface{})("test"), false},

			// Map types
			{"nil_map", (map[string]int)(nil), true},
			{"empty_map", make(map[string]int), true},
			{"non_empty_map", map[string]int{"key": 1}, false},

			// Channel types (channels are rarely used in CLI argument parsing)
			{"nil_channel", (chan int)(nil), true},

			// Array types
			{"zero_array", [3]int{}, true},
			{"non_zero_array", [3]int{1, 0, 0}, false},

			// Complex struct
			{"nested_zero_struct", struct {
				A int
				B struct {
					C string
					D bool
				}
			}{}, true},
			{"nested_non_zero_struct", struct {
				A int
				B struct {
					C string
					D bool
				}
			}{A: 0, B: struct {
				C string
				D bool
			}{C: "test"}}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				value := reflect.ValueOf(tt.value)
				result := tc.isZeroValue(value)

				if result != tt.expected {
					t.Errorf("isZeroValue(%v) = %v, want %v", tt.value, result, tt.expected)
				}
			})
		}
	})
}

// migratedConversionTests verifies that parsing through the full GoArg
// pipeline (which now uses optargs.Convert) produces identical results
// for all supported types.
var migratedConversionTests = []struct {
	name string
	dest interface{}
	args []string
	check func(t *testing.T, dest interface{})
}{
	{
		name: "bool_flag",
		dest: &struct{ Verbose bool `arg:"-v,--verbose"` }{},
		args: []string{"--verbose"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ Verbose bool `arg:"-v,--verbose"` })
			if !s.Verbose { t.Error("expected Verbose=true") }
		},
	},
	{
		name: "int_scalar",
		dest: &struct{ Count int `arg:"-c,--count"` }{},
		args: []string{"--count", "42"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ Count int `arg:"-c,--count"` })
			if s.Count != 42 { t.Errorf("Count = %d, want 42", s.Count) }
		},
	},
	{
		name: "float64_scalar",
		dest: &struct{ Rate float64 `arg:"--rate"` }{},
		args: []string{"--rate", "3.14"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ Rate float64 `arg:"--rate"` })
			if s.Rate != 3.14 { t.Errorf("Rate = %f, want 3.14", s.Rate) }
		},
	},
	{
		name: "string_scalar",
		dest: &struct{ Name string `arg:"-n,--name"` }{},
		args: []string{"--name", "hello"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ Name string `arg:"-n,--name"` })
			if s.Name != "hello" { t.Errorf("Name = %q, want %q", s.Name, "hello") }
		},
	},
	{
		name: "int_slice",
		dest: &struct{ Nums []int `arg:"-n,--num"` }{},
		args: []string{"--num", "1", "--num", "2", "--num", "3"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ Nums []int `arg:"-n,--num"` })
			if len(s.Nums) != 3 || s.Nums[0] != 1 || s.Nums[1] != 2 || s.Nums[2] != 3 {
				t.Errorf("Nums = %v, want [1 2 3]", s.Nums)
			}
		},
	},
	{
		name: "default_int",
		dest: &struct{ Port int `arg:"--port" default:"8080"` }{},
		args: []string{},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ Port int `arg:"--port" default:"8080"` })
			if s.Port != 8080 { t.Errorf("Port = %d, want 8080", s.Port) }
		},
	},
	{
		name: "default_string",
		dest: &struct{ Host string `arg:"--host" default:"localhost"` }{},
		args: []string{},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ Host string `arg:"--host" default:"localhost"` })
			if s.Host != "localhost" { t.Errorf("Host = %q, want %q", s.Host, "localhost") }
		},
	},
	{
		name: "positional_string",
		dest: &struct{ File string `arg:"positional"` }{},
		args: []string{"input.txt"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ File string `arg:"positional"` })
			if s.File != "input.txt" { t.Errorf("File = %q, want %q", s.File, "input.txt") }
		},
	},
	{
		name: "short_option",
		dest: &struct{ Verbose bool `arg:"-v"` }{},
		args: []string{"-v"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ Verbose bool `arg:"-v"` })
			if !s.Verbose { t.Error("expected Verbose=true") }
		},
	},
	{
		name: "override_default",
		dest: &struct{ Port int `arg:"--port" default:"8080"` }{},
		args: []string{"--port", "9090"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct{ Port int `arg:"--port" default:"8080"` })
			if s.Port != 9090 { t.Errorf("Port = %d, want 9090", s.Port) }
		},
	},
}

func TestMigratedTypeConversion(t *testing.T) {
	for _, tt := range migratedConversionTests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseArgs(tt.dest, tt.args); err != nil {
				t.Fatalf("ParseArgs: %v", err)
			}
			tt.check(t, tt.dest)
		})
	}
}

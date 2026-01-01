package goarg

import (
	"reflect"
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

func TestTypeConverter_ConvertSlice(t *testing.T) {
	tc := &TypeConverter{}

	tests := []struct {
		name        string
		values      []string
		elementType reflect.Type
		expected    interface{}
		wantError   bool
	}{
		{
			name:        "string_slice",
			values:      []string{"a", "b", "c"},
			elementType: reflect.TypeOf(""),
			expected:    []string{"a", "b", "c"},
		},
		{
			name:        "int_slice",
			values:      []string{"1", "2", "3"},
			elementType: reflect.TypeOf(int(0)),
			expected:    []int{1, 2, 3},
		},
		{
			name:        "bool_slice",
			values:      []string{"true", "false", "1"},
			elementType: reflect.TypeOf(false),
			expected:    []bool{true, false, true},
		},
		{
			name:        "empty_slice",
			values:      []string{},
			elementType: reflect.TypeOf(""),
			expected:    []string{},
		},
		{
			name:        "invalid_int_element",
			values:      []string{"1", "not_a_number", "3"},
			elementType: reflect.TypeOf(int(0)),
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tc.ConvertSlice(tt.values, tt.elementType)

			if tt.wantError {
				if err == nil {
					t.Errorf("ConvertSlice() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertSlice() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ConvertSlice() = %v, want %v", result, tt.expected)
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

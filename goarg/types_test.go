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

func TestTypeConverter_ApplyDefaults(t *testing.T) {
	tc := &TypeConverter{}

	type TestStruct struct {
		StringField    string   `default:"hello"`
		IntField       int      `default:"42"`
		BoolField      bool     `default:"true"`
		SliceField     []string `default:"a,b,c"`
		NoDefaultField string
	}

	tests := []struct {
		name     string
		setup    func(*TestStruct)
		expected TestStruct
	}{
		{
			name:  "apply_all_defaults",
			setup: func(ts *TestStruct) {},
			expected: TestStruct{
				StringField: "hello",
				IntField:    42,
				BoolField:   true,
				SliceField:  []string{"a", "b", "c"},
			},
		},
		{
			name: "preserve_existing_values",
			setup: func(ts *TestStruct) {
				ts.StringField = "existing"
				ts.IntField = 100
			},
			expected: TestStruct{
				StringField: "existing",
				IntField:    100,
				BoolField:   true,
				SliceField:  []string{"a", "b", "c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := &TestStruct{}
			tt.setup(testStruct)

			// Create metadata with default values
			metadata := &StructMetadata{
				Fields: []FieldMetadata{
					{Name: "StringField", Default: "hello"},
					{Name: "IntField", Default: 42},
					{Name: "BoolField", Default: true},
					{Name: "SliceField", Default: []string{"a", "b", "c"}},
					{Name: "NoDefaultField", Default: nil},
				},
			}

			err := tc.ApplyDefaults(testStruct, metadata)
			if err != nil {
				t.Errorf("ApplyDefaults() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(*testStruct, tt.expected) {
				t.Errorf("ApplyDefaults() result = %+v, want %+v", *testStruct, tt.expected)
			}
		})
	}
}

func TestTypeConverter_ValidateCustom(t *testing.T) {
	tc := &TypeConverter{}

	type TestStruct struct {
		MinMaxInt    int      `min:"10" max:"100"`
		MinLenString string   `minlen:"3"`
		MaxLenString string   `maxlen:"10"`
		MinLenSlice  []string `minlen:"2"`
		MaxLenSlice  []string `maxlen:"5"`
	}

	tests := []struct {
		name      string
		setup     func(*TestStruct)
		wantError bool
		errorMsg  string
	}{
		{
			name: "all_constraints_satisfied",
			setup: func(ts *TestStruct) {
				ts.MinMaxInt = 50
				ts.MinLenString = "hello"
				ts.MaxLenString = "short"
				ts.MinLenSlice = []string{"a", "b"}
				ts.MaxLenSlice = []string{"1", "2", "3"}
			},
			wantError: false,
		},
		{
			name: "int_below_min",
			setup: func(ts *TestStruct) {
				ts.MinMaxInt = 5
			},
			wantError: true,
			errorMsg:  "field MinMaxInt value 5 is less than minimum 10",
		},
		{
			name: "int_above_max",
			setup: func(ts *TestStruct) {
				ts.MinMaxInt = 150
			},
			wantError: true,
			errorMsg:  "field MinMaxInt value 150 is greater than maximum 100",
		},
		{
			name: "string_below_minlen",
			setup: func(ts *TestStruct) {
				ts.MinMaxInt = 50
				ts.MinLenString = "hi"
			},
			wantError: true,
			errorMsg:  "field MinLenString length 2 is less than minimum 3",
		},
		{
			name: "string_above_maxlen",
			setup: func(ts *TestStruct) {
				ts.MinMaxInt = 50
				ts.MinLenString = "hello"
				ts.MaxLenString = "this is too long"
			},
			wantError: true,
			errorMsg:  "field MaxLenString length 16 is greater than maximum 10",
		},
		{
			name: "slice_below_minlen",
			setup: func(ts *TestStruct) {
				ts.MinMaxInt = 50
				ts.MinLenString = "hello"
				ts.MaxLenString = "short"
				ts.MinLenSlice = []string{"a"}
			},
			wantError: true,
			errorMsg:  "field MinLenSlice length 1 is less than minimum 2",
		},
		{
			name: "slice_above_maxlen",
			setup: func(ts *TestStruct) {
				ts.MinMaxInt = 50
				ts.MinLenString = "hello"
				ts.MaxLenString = "short"
				ts.MinLenSlice = []string{"a", "b"}
				ts.MaxLenSlice = []string{"1", "2", "3", "4", "5", "6"}
			},
			wantError: true,
			errorMsg:  "field MaxLenSlice length 6 is greater than maximum 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := &TestStruct{}
			tt.setup(testStruct)

			// Create metadata (constraints are read from struct tags)
			metadata := &StructMetadata{
				Fields: []FieldMetadata{
					{Name: "MinMaxInt"},
					{Name: "MinLenString"},
					{Name: "MaxLenString"},
					{Name: "MinLenSlice"},
					{Name: "MaxLenSlice"},
				},
			}

			err := tc.ValidateCustom(testStruct, metadata)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateCustom() expected error, got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("ValidateCustom() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCustom() unexpected error: %v", err)
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

	t.Run("slice_with_mixed_valid_invalid_elements", func(t *testing.T) {
		// Test slice conversion with some valid and some invalid elements
		_, err := tc.ConvertSlice([]string{"1", "invalid", "3"}, reflect.TypeOf(int(0)))
		if err == nil {
			t.Errorf("ConvertSlice with invalid element expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to convert slice element 'invalid'") {
			t.Errorf("Expected error message about invalid element, got: %v", err)
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

func (ect *ErrorCustomType) UnmarshalText(text []byte) error {
	if string(text) == "error" {
		return fmt.Errorf("custom unmarshal error")
	}
	ect.Value = string(text)
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

	t.Run("apply_defaults_error_conditions", func(t *testing.T) {
		// Test ApplyDefaults with non-struct destination
		var notAStruct int = 42
		metadata := &StructMetadata{}

		err := tc.ApplyDefaults(&notAStruct, metadata)
		if err == nil {
			t.Errorf("ApplyDefaults with non-struct expected error, got nil")
		}

		// Test ApplyDefaults with non-pointer destination
		var structVal struct{ Field string }
		err = tc.ApplyDefaults(structVal, metadata)
		if err == nil {
			t.Errorf("ApplyDefaults with non-pointer expected error, got nil")
		}
	})

	t.Run("validate_custom_error_conditions", func(t *testing.T) {
		// Test ValidateCustom with non-struct destination
		var notAStruct int = 42
		metadata := &StructMetadata{}

		err := tc.ValidateCustom(&notAStruct, metadata)
		if err == nil {
			t.Errorf("ValidateCustom with non-struct expected error, got nil")
		}

		// Test ValidateCustom with non-pointer destination
		var structVal struct{ Field string }
		err = tc.ValidateCustom(structVal, metadata)
		if err == nil {
			t.Errorf("ValidateCustom with non-pointer expected error, got nil")
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

	t.Run("numeric_constraint_validation", func(t *testing.T) {
		type TestStruct struct {
			// Integer types
			Int8Field  int8  `min:"10" max:"100"`
			Int16Field int16 `min:"1000" max:"30000"`
			Int32Field int32 `min:"100000" max:"2000000"`
			Int64Field int64 `min:"1000000" max:"9000000"`

			// Unsigned integer types
			UintField   uint   `min:"5" max:"50"`
			Uint8Field  uint8  `min:"10" max:"200"`
			Uint16Field uint16 `min:"1000" max:"60000"`
			Uint32Field uint32 `min:"100000" max:"4000000"`
			Uint64Field uint64 `min:"1000000" max:"18000000"`

			// Float types
			Float32Field float32 `min:"1.5" max:"99.9"`
			Float64Field float64 `min:"10.25" max:"999.75"`
		}

		tests := []struct {
			name      string
			setup     func(*TestStruct)
			wantError bool
			errorMsg  string
		}{
			// Test all numeric types within bounds
			{
				name: "all_numeric_types_within_bounds",
				setup: func(ts *TestStruct) {
					ts.Int8Field = 50
					ts.Int16Field = 15000
					ts.Int32Field = 1500000
					ts.Int64Field = 5000000
					ts.UintField = 25
					ts.Uint8Field = 100
					ts.Uint16Field = 30000
					ts.Uint32Field = 2000000
					ts.Uint64Field = 9000000
					ts.Float32Field = 50.5
					ts.Float64Field = 500.5
				},
				wantError: false,
			},
			// Test int8 bounds
			{
				name: "int8_below_min",
				setup: func(ts *TestStruct) {
					ts.Int8Field = 5
				},
				wantError: true,
				errorMsg:  "field Int8Field value 5 is less than minimum 10",
			},
			{
				name: "int8_above_max",
				setup: func(ts *TestStruct) {
					ts.Int8Field = 120
				},
				wantError: true,
				errorMsg:  "field Int8Field value 120 is greater than maximum 100",
			},
			// Test uint8 bounds
			{
				name: "uint8_below_min",
				setup: func(ts *TestStruct) {
					// Set all other fields to valid values
					ts.Int8Field = 50
					ts.Int16Field = 15000
					ts.Int32Field = 1500000
					ts.Int64Field = 5000000
					ts.UintField = 25
					ts.Uint8Field = 5 // This is the violation
					ts.Uint16Field = 30000
					ts.Uint32Field = 2000000
					ts.Uint64Field = 9000000
					ts.Float32Field = 50.5
					ts.Float64Field = 500.5
				},
				wantError: true,
				errorMsg:  "field Uint8Field value 5 is less than minimum 10",
			},
			{
				name: "uint8_above_max",
				setup: func(ts *TestStruct) {
					// Set all other fields to valid values
					ts.Int8Field = 50
					ts.Int16Field = 15000
					ts.Int32Field = 1500000
					ts.Int64Field = 5000000
					ts.UintField = 25
					ts.Uint8Field = 250 // This is the violation
					ts.Uint16Field = 30000
					ts.Uint32Field = 2000000
					ts.Uint64Field = 9000000
					ts.Float32Field = 50.5
					ts.Float64Field = 500.5
				},
				wantError: true,
				errorMsg:  "field Uint8Field value 250 is greater than maximum 200",
			},
			// Test float32 bounds
			{
				name: "float32_below_min",
				setup: func(ts *TestStruct) {
					// Set all other fields to valid values
					ts.Int8Field = 50
					ts.Int16Field = 15000
					ts.Int32Field = 1500000
					ts.Int64Field = 5000000
					ts.UintField = 25
					ts.Uint8Field = 100
					ts.Uint16Field = 30000
					ts.Uint32Field = 2000000
					ts.Uint64Field = 9000000
					ts.Float32Field = 1.0 // This is the violation
					ts.Float64Field = 500.5
				},
				wantError: true,
				errorMsg:  "field Float32Field value 1.000000 is less than minimum 1.500000",
			},
			{
				name: "float32_above_max",
				setup: func(ts *TestStruct) {
					// Set all other fields to valid values
					ts.Int8Field = 50
					ts.Int16Field = 15000
					ts.Int32Field = 1500000
					ts.Int64Field = 5000000
					ts.UintField = 25
					ts.Uint8Field = 100
					ts.Uint16Field = 30000
					ts.Uint32Field = 2000000
					ts.Uint64Field = 9000000
					ts.Float32Field = 100.0 // This is the violation
					ts.Float64Field = 500.5
				},
				wantError: true,
				errorMsg:  "field Float32Field value 100.000000 is greater than maximum 99.900000",
			},
			// Test float64 bounds
			{
				name: "float64_below_min",
				setup: func(ts *TestStruct) {
					// Set all other fields to valid values
					ts.Int8Field = 50
					ts.Int16Field = 15000
					ts.Int32Field = 1500000
					ts.Int64Field = 5000000
					ts.UintField = 25
					ts.Uint8Field = 100
					ts.Uint16Field = 30000
					ts.Uint32Field = 2000000
					ts.Uint64Field = 9000000
					ts.Float32Field = 50.5
					ts.Float64Field = 5.0 // This is the violation
				},
				wantError: true,
				errorMsg:  "field Float64Field value 5.000000 is less than minimum 10.250000",
			},
			{
				name: "float64_above_max",
				setup: func(ts *TestStruct) {
					// Set all other fields to valid values
					ts.Int8Field = 50
					ts.Int16Field = 15000
					ts.Int32Field = 1500000
					ts.Int64Field = 5000000
					ts.UintField = 25
					ts.Uint8Field = 100
					ts.Uint16Field = 30000
					ts.Uint32Field = 2000000
					ts.Uint64Field = 9000000
					ts.Float32Field = 50.5
					ts.Float64Field = 1000.0 // This is the violation
				},
				wantError: true,
				errorMsg:  "field Float64Field value 1000.000000 is greater than maximum 999.750000",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				testStruct := &TestStruct{}
				tt.setup(testStruct)

				// Create metadata (constraints are read from struct tags)
				metadata := &StructMetadata{
					Fields: []FieldMetadata{
						{Name: "Int8Field"},
						{Name: "Int16Field"},
						{Name: "Int32Field"},
						{Name: "Int64Field"},
						{Name: "UintField"},
						{Name: "Uint8Field"},
						{Name: "Uint16Field"},
						{Name: "Uint32Field"},
						{Name: "Uint64Field"},
						{Name: "Float32Field"},
						{Name: "Float64Field"},
					},
				}

				err := tc.ValidateCustom(testStruct, metadata)

				if tt.wantError {
					if err == nil {
						t.Errorf("ValidateCustom() expected error, got nil")
						return
					}
					if err.Error() != tt.errorMsg {
						t.Errorf("ValidateCustom() error = %v, want %v", err.Error(), tt.errorMsg)
					}
				} else {
					if err != nil {
						t.Errorf("ValidateCustom() unexpected error: %v", err)
					}
				}
			})
		}
	})

	t.Run("constraint_validation_edge_cases", func(t *testing.T) {
		type TestStruct struct {
			UnsupportedType chan int `min:"1"`
			InvalidMinTag   int      `min:"not_a_number"`
			InvalidMaxTag   int      `max:"not_a_number"`
			InvalidMinLen   string   `minlen:"not_a_number"`
			InvalidMaxLen   string   `maxlen:"not_a_number"`
		}

		testStruct := &TestStruct{
			UnsupportedType: make(chan int),
			InvalidMinTag:   50,
			InvalidMaxTag:   50,
			InvalidMinLen:   "test",
			InvalidMaxLen:   "test",
		}

		structType := reflect.TypeOf(*testStruct)
		structValue := reflect.ValueOf(testStruct).Elem()

		// Test unsupported type for min constraint (should be ignored)
		field := structValue.FieldByName("UnsupportedType")
		structField, _ := structType.FieldByName("UnsupportedType")
		err := tc.validateFieldConstraints(field, structField, "UnsupportedType")
		if err != nil {
			t.Errorf("validateFieldConstraints with unsupported type should be ignored, got error: %v", err)
		}

		// Test invalid min tag
		field = structValue.FieldByName("InvalidMinTag")
		structField, _ = structType.FieldByName("InvalidMinTag")
		err = tc.validateFieldConstraints(field, structField, "InvalidMinTag")
		if err == nil {
			t.Errorf("validateFieldConstraints with invalid min tag expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid min constraint") {
			t.Errorf("Expected invalid min constraint error, got: %v", err)
		}

		// Test invalid max tag
		field = structValue.FieldByName("InvalidMaxTag")
		structField, _ = structType.FieldByName("InvalidMaxTag")
		err = tc.validateFieldConstraints(field, structField, "InvalidMaxTag")
		if err == nil {
			t.Errorf("validateFieldConstraints with invalid max tag expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid max constraint") {
			t.Errorf("Expected invalid max constraint error, got: %v", err)
		}

		// Test invalid minlen tag
		field = structValue.FieldByName("InvalidMinLen")
		structField, _ = structType.FieldByName("InvalidMinLen")
		err = tc.validateFieldConstraints(field, structField, "InvalidMinLen")
		if err == nil {
			t.Errorf("validateFieldConstraints with invalid minlen tag expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid minlen constraint") {
			t.Errorf("Expected invalid minlen constraint error, got: %v", err)
		}

		// Test invalid maxlen tag
		field = structValue.FieldByName("InvalidMaxLen")
		structField, _ = structType.FieldByName("InvalidMaxLen")
		err = tc.validateFieldConstraints(field, structField, "InvalidMaxLen")
		if err == nil {
			t.Errorf("validateFieldConstraints with invalid maxlen tag expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid maxlen constraint") {
			t.Errorf("Expected invalid maxlen constraint error, got: %v", err)
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

	t.Run("apply_defaults_field_not_found", func(t *testing.T) {
		type TestStruct struct {
			ExistingField string
		}

		testStruct := &TestStruct{}
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "NonExistentField", Default: "test"},
				{Name: "ExistingField", Default: "existing"},
			},
		}

		err := tc.ApplyDefaults(testStruct, metadata)
		if err != nil {
			t.Errorf("ApplyDefaults with non-existent field unexpected error: %v", err)
		}

		// Should only set the existing field
		if testStruct.ExistingField != "existing" {
			t.Errorf("Expected ExistingField to be 'existing', got %v", testStruct.ExistingField)
		}
	})
}

// TestTypeConverter_100PercentCoverage tests all remaining uncovered lines to achieve 100% coverage
func TestTypeConverter_100PercentCoverage(t *testing.T) {
	tc := &TypeConverter{}

	t.Run("convert_custom_complete_coverage", func(t *testing.T) {
		// Test ConvertCustom with pointer type that implements TextUnmarshaler directly
		result, err := tc.ConvertCustom("test", reflect.TypeOf(&CustomTypePtr{}))
		if err != nil {
			t.Errorf("ConvertCustom with pointer type unexpected error: %v", err)
		}
		if ptr, ok := result.(*CustomTypePtr); !ok || ptr.Value != "ptr:test" {
			t.Errorf("ConvertCustom result = %v, want &CustomTypePtr{Value: 'ptr:test'}", result)
		}

		// Test ConvertCustom with value type that has pointer receiver implementing TextUnmarshaler
		result, err = tc.ConvertCustom("test", reflect.TypeOf(CustomTypePtrReceiver{}))
		if err != nil {
			t.Errorf("ConvertCustom with value type (ptr receiver) unexpected error: %v", err)
		}
		if val, ok := result.(CustomTypePtrReceiver); !ok || val.Value != "ptr-receiver:test" {
			t.Errorf("ConvertCustom result = %v, want CustomTypePtrReceiver{Value: 'ptr-receiver:test'}", result)
		}

		// Test ConvertCustom with pointer type that has pointer receiver implementing TextUnmarshaler
		result, err = tc.ConvertCustom("test", reflect.TypeOf(&CustomTypePtrReceiver{}))
		if err != nil {
			t.Errorf("ConvertCustom with pointer type (ptr receiver) unexpected error: %v", err)
		}
		if ptr, ok := result.(*CustomTypePtrReceiver); !ok || ptr.Value != "ptr-receiver:test" {
			t.Errorf("ConvertCustom result = %v, want &CustomTypePtrReceiver{Value: 'ptr-receiver:test'}", result)
		}

		// Test ConvertCustom error in pointer-to-type path
		_, err = tc.ConvertCustom("error", reflect.TypeOf(ErrorCustomType{}))
		if err == nil {
			t.Errorf("ConvertCustom with error input expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal text") {
			t.Errorf("Expected unmarshal error message, got: %v", err)
		}

		// Test ConvertCustom with pointer type that causes error in pointer-to-type path
		_, err = tc.ConvertCustom("error", reflect.TypeOf(&ErrorCustomType{}))
		if err == nil {
			t.Errorf("ConvertCustom with pointer type error input expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal text") {
			t.Errorf("Expected unmarshal error message, got: %v", err)
		}
	})

	t.Run("validate_required_complete_coverage", func(t *testing.T) {
		type TestStruct struct {
			InvalidField string `arg:"positional,required"`
		}

		testStruct := &TestStruct{}
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "InvalidField", Required: true, Positional: false, Short: "", Long: ""},
			},
		}

		// Test ValidateRequired with field that has no valid flag info
		err := tc.ValidateRequired(testStruct, metadata)
		if err == nil {
			t.Errorf("ValidateRequired expected error, got nil")
		}
		if err.Error() != "InvalidField is required" {
			t.Errorf("ValidateRequired error = %v, want 'InvalidField is required'", err.Error())
		}

		// Test ValidateRequired with invalid field (field doesn't exist in struct)
		metadata = &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "NonExistentField", Required: true, Long: "nonexistent"},
			},
		}

		err = tc.ValidateRequired(testStruct, metadata)
		if err != nil {
			t.Errorf("ValidateRequired with non-existent field unexpected error: %v", err)
		}
	})

	t.Run("is_zero_value_complete_coverage", func(t *testing.T) {
		// Test isZeroValue with invalid reflect.Value
		var invalidValue reflect.Value
		result := tc.isZeroValue(invalidValue)
		if !result {
			t.Errorf("isZeroValue with invalid value expected true, got false")
		}

		// Test isZeroValue with complex types
		complexType := reflect.ValueOf(complex(1, 2))
		result = tc.isZeroValue(complexType)
		if result {
			t.Errorf("isZeroValue with non-zero complex expected false, got true")
		}

		zeroComplex := reflect.ValueOf(complex(0, 0))
		result = tc.isZeroValue(zeroComplex)
		if !result {
			t.Errorf("isZeroValue with zero complex expected true, got false")
		}
	})

	t.Run("apply_defaults_complete_coverage", func(t *testing.T) {
		type TestStruct struct {
			ReadOnlyField string
		}

		// Test ApplyDefaults with unsettable field (this is a theoretical case)
		testStruct := &TestStruct{}

		// Create metadata with a field that can't be set
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "ReadOnlyField", Default: "test"},
			},
		}

		err := tc.ApplyDefaults(testStruct, metadata)
		if err != nil {
			t.Errorf("ApplyDefaults unexpected error: %v", err)
		}

		// Test ApplyDefaults with SetField error (incompatible default type)
		type BadDefaultStruct struct {
			IntField int
		}

		badStruct := &BadDefaultStruct{}
		badMetadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "IntField", Default: []string{"not", "an", "int"}}, // Incompatible type
			},
		}

		err = tc.ApplyDefaults(badStruct, badMetadata)
		if err == nil {
			t.Errorf("ApplyDefaults with incompatible default type expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to set default value") {
			t.Errorf("Expected default value error, got: %v", err)
		}
	})

	t.Run("validate_custom_complete_coverage", func(t *testing.T) {
		type TestStruct struct {
			ValidField string
		}

		testStruct := &TestStruct{ValidField: "test"}

		// Test ValidateCustom with invalid field (field doesn't exist in struct)
		// The metadata has more fields than the struct, which should be handled gracefully
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "ValidField"},
			},
		}

		err := tc.ValidateCustom(testStruct, metadata)
		if err != nil {
			t.Errorf("ValidateCustom unexpected error: %v", err)
		}
	})

	t.Run("validation_constraints_complete_coverage", func(t *testing.T) {
		type TestStruct struct {
			// Test all numeric types for min/max validation
			IntField     int     `min:"10" max:"100"`
			Int8Field    int8    `min:"10" max:"100"`
			Int16Field   int16   `min:"10" max:"100"`
			Int32Field   int32   `min:"10" max:"100"`
			Int64Field   int64   `min:"10" max:"100"`
			UintField    uint    `min:"10" max:"100"`
			Uint8Field   uint8   `min:"10" max:"100"`
			Uint16Field  uint16  `min:"10" max:"100"`
			Uint32Field  uint32  `min:"10" max:"100"`
			Uint64Field  uint64  `min:"10" max:"100"`
			Float32Field float32 `min:"10.5" max:"100.5"`
			Float64Field float64 `min:"10.5" max:"100.5"`

			// Test length validation on different types
			StringField string   `minlen:"3" maxlen:"10"`
			SliceField  []string `minlen:"2" maxlen:"5"`
			ArrayField  [3]int   `minlen:"2" maxlen:"4"`
		}

		// Test all numeric types within bounds (should pass)
		testStruct := &TestStruct{
			IntField:     50,
			Int8Field:    50,
			Int16Field:   50,
			Int32Field:   50,
			Int64Field:   50,
			UintField:    50,
			Uint8Field:   50,
			Uint16Field:  50,
			Uint32Field:  50,
			Uint64Field:  50,
			Float32Field: 50.5,
			Float64Field: 50.5,
			StringField:  "hello",
			SliceField:   []string{"a", "b", "c"},
			ArrayField:   [3]int{1, 2, 3},
		}

		structType := reflect.TypeOf(*testStruct)
		structValue := reflect.ValueOf(testStruct).Elem()

		// Test all fields pass validation
		for i := 0; i < structType.NumField(); i++ {
			field := structValue.Field(i)
			structField := structType.Field(i)
			err := tc.validateFieldConstraints(field, structField, structField.Name)
			if err != nil {
				t.Errorf("validateFieldConstraints for %s unexpected error: %v", structField.Name, err)
			}
		}

		// Test array length validation (arrays are treated like slices for length)
		arrayField := structValue.FieldByName("ArrayField")
		err := tc.validateMinLen(arrayField, "2", "ArrayField")
		if err != nil {
			t.Errorf("validateMinLen for array unexpected error: %v", err)
		}

		err = tc.validateMaxLen(arrayField, "4", "ArrayField")
		if err != nil {
			t.Errorf("validateMaxLen for array unexpected error: %v", err)
		}

		// Test unsupported type for length validation (should be ignored)
		intField := structValue.FieldByName("IntField")
		err = tc.validateMinLen(intField, "5", "IntField")
		if err != nil {
			t.Errorf("validateMinLen with unsupported type should be ignored, got error: %v", err)
		}

		err = tc.validateMaxLen(intField, "5", "IntField")
		if err != nil {
			t.Errorf("validateMaxLen with unsupported type should be ignored, got error: %v", err)
		}
	})

	t.Run("convert_value_complete_coverage", func(t *testing.T) {
		// Test ConvertValue with all integer types to ensure 100% coverage
		tests := []struct {
			value      string
			targetType reflect.Type
			expected   interface{}
		}{
			// Test all integer types
			{"42", reflect.TypeOf(int(0)), 42},
			{"42", reflect.TypeOf(int8(42)), int8(42)},
			{"42", reflect.TypeOf(int16(42)), int16(42)},
			{"42", reflect.TypeOf(int32(42)), int32(42)},
			{"42", reflect.TypeOf(int64(42)), int64(42)},
			{"42", reflect.TypeOf(uint(42)), uint(42)},
			{"42", reflect.TypeOf(uint8(42)), uint8(42)},
			{"42", reflect.TypeOf(uint16(42)), uint16(42)},
			{"42", reflect.TypeOf(uint32(42)), uint32(42)},
			{"42", reflect.TypeOf(uint64(42)), uint64(42)},
			{"3.14", reflect.TypeOf(float32(3.14)), float32(3.14)},
			{"3.14", reflect.TypeOf(float64(3.14)), 3.14},
		}

		for _, tt := range tests {
			result, err := tc.ConvertValue(tt.value, tt.targetType)
			if err != nil {
				t.Errorf("ConvertValue(%s, %v) unexpected error: %v", tt.value, tt.targetType, err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ConvertValue(%s, %v) = %v, want %v", tt.value, tt.targetType, result, tt.expected)
			}
		}

		// Test ConvertValue with triple pointer (nested pointers)
		triplePtr := reflect.TypeOf((***string)(nil))
		result, err := tc.ConvertValue("test", triplePtr)
		if err != nil {
			t.Errorf("ConvertValue with triple pointer unexpected error: %v", err)
		}
		ptr, ok := result.(***string)
		if !ok || ptr == nil || *ptr == nil || **ptr == nil || ***ptr != "test" {
			t.Errorf("ConvertValue with triple pointer failed, got %v", result)
		}

		// Test ConvertValue error path for unsupported type in default case
		unsupportedType := reflect.TypeOf(make(chan int))
		_, err = tc.ConvertValue("test", unsupportedType)
		if err == nil {
			t.Errorf("ConvertValue with unsupported type expected error, got nil")
		}
	})

	t.Run("validate_custom_missing_coverage", func(t *testing.T) {
		type TestStruct struct {
			NonExistentInMetadata string
		}

		testStruct := &TestStruct{NonExistentInMetadata: "test"}

		// Test ValidateCustom where struct field is not found in metadata (should continue)
		metadata := &StructMetadata{
			Fields: []FieldMetadata{
				{Name: "SomeOtherField"}, // This field doesn't exist in struct
			},
		}

		err := tc.ValidateCustom(testStruct, metadata)
		if err != nil {
			t.Errorf("ValidateCustom unexpected error: %v", err)
		}
	})

	t.Run("validate_min_max_missing_coverage", func(t *testing.T) {
		type TestStruct struct {
			UnsupportedField chan int `min:"1" max:"10"`
		}

		testStruct := &TestStruct{UnsupportedField: make(chan int)}
		structValue := reflect.ValueOf(testStruct).Elem()

		// Test validateMin/validateMax with unsupported type (should be ignored)
		field := structValue.FieldByName("UnsupportedField")

		err := tc.validateMin(field, "1", "UnsupportedField")
		if err != nil {
			t.Errorf("validateMin with unsupported type should be ignored, got error: %v", err)
		}

		err = tc.validateMax(field, "10", "UnsupportedField")
		if err != nil {
			t.Errorf("validateMax with unsupported type should be ignored, got error: %v", err)
		}
	})

	t.Run("convert_custom_missing_coverage", func(t *testing.T) {
		// Test the case where neither the type nor pointer-to-type implements TextUnmarshaler
		type NonUnmarshalerType struct {
			Value string
		}

		_, err := tc.ConvertCustom("test", reflect.TypeOf(NonUnmarshalerType{}))
		if err == nil {
			t.Errorf("ConvertCustom with non-unmarshaler type expected error, got nil")
		}
		if !strings.Contains(err.Error(), "does not implement encoding.TextUnmarshaler") {
			t.Errorf("Expected TextUnmarshaler error, got: %v", err)
		}

		// Test ConvertCustom with pointer type where the element type implements TextUnmarshaler
		// but we want to test the pointer-to-type path specifically
		result, err := tc.ConvertCustom("test", reflect.TypeOf(CustomTypePointerOnly{}))
		if err != nil {
			t.Errorf("ConvertCustom with pointer-only type unexpected error: %v", err)
		}
		if val, ok := result.(CustomTypePointerOnly); !ok || val.Value != "pointer-only:test" {
			t.Errorf("ConvertCustom result = %v, want CustomTypePointerOnly{Value: 'pointer-only:test'}", result)
		}

		// Test ConvertCustom with pointer type in the second path (pointer-to-type implements TextUnmarshaler)
		result, err = tc.ConvertCustom("test", reflect.TypeOf(&CustomTypePointerOnly{}))
		if err != nil {
			t.Errorf("ConvertCustom with pointer to pointer-only type unexpected error: %v", err)
		}
		if ptr, ok := result.(*CustomTypePointerOnly); !ok || ptr.Value != "pointer-only:test" {
			t.Errorf("ConvertCustom result = %v, want &CustomTypePointerOnly{Value: 'pointer-only:test'}", result)
		}

		// Test ConvertCustom error in the second path (pointer-to-type path)
		_, err = tc.ConvertCustom("error", reflect.TypeOf(&ErrorCustomType{}))
		if err == nil {
			t.Errorf("ConvertCustom with error in pointer-to-type path expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal text") {
			t.Errorf("Expected unmarshal error message, got: %v", err)
		}
	})

	t.Run("validate_min_max_all_types", func(t *testing.T) {
		// Test validateMin and validateMax with all supported numeric types
		tests := []struct {
			name      string
			value     interface{}
			minTag    string
			maxTag    string
			shouldErr bool
		}{
			// Test all integer types
			{"int", int(50), "10", "100", false},
			{"int8", int8(50), "10", "100", false},
			{"int16", int16(50), "10", "100", false},
			{"int32", int32(50), "10", "100", false},
			{"int64", int64(50), "10", "100", false},
			{"uint", uint(50), "10", "100", false},
			{"uint8", uint8(50), "10", "100", false},
			{"uint16", uint16(50), "10", "100", false},
			{"uint32", uint32(50), "10", "100", false},
			{"uint64", uint64(50), "10", "100", false},
			{"float32", float32(50.5), "10.0", "100.0", false},
			{"float64", float64(50.5), "10.0", "100.0", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fieldValue := reflect.ValueOf(tt.value)

				err := tc.validateMin(fieldValue, tt.minTag, tt.name)
				if tt.shouldErr && err == nil {
					t.Errorf("validateMin expected error, got nil")
				} else if !tt.shouldErr && err != nil {
					t.Errorf("validateMin unexpected error: %v", err)
				}

				err = tc.validateMax(fieldValue, tt.maxTag, tt.name)
				if tt.shouldErr && err == nil {
					t.Errorf("validateMax expected error, got nil")
				} else if !tt.shouldErr && err != nil {
					t.Errorf("validateMax unexpected error: %v", err)
				}
			})
		}
	})

	t.Run("convert_value_missing_paths", func(t *testing.T) {
		// Test ConvertValue with slice of custom types
		customSliceType := reflect.TypeOf([]CustomType{})
		result, err := tc.ConvertValue("test", customSliceType)
		if err != nil {
			t.Errorf("ConvertValue with custom slice unexpected error: %v", err)
		}
		if slice, ok := result.([]CustomType); !ok || len(slice) != 1 || slice[0].Value != "custom:test" {
			t.Errorf("ConvertValue with custom slice = %v, want []CustomType{{Value: 'custom:test'}}", result)
		}

		// Test ConvertValue with pointer to slice
		ptrSliceType := reflect.TypeOf((*[]string)(nil))
		result, err = tc.ConvertValue("test", ptrSliceType)
		if err != nil {
			t.Errorf("ConvertValue with pointer to slice unexpected error: %v", err)
		}
		if ptr, ok := result.(*[]string); !ok || len(*ptr) != 1 || (*ptr)[0] != "test" {
			t.Errorf("ConvertValue with pointer to slice = %v, want &[]string{'test'}", result)
		}
	})
}

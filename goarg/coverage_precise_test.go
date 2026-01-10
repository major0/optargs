package goarg

import (
	"encoding"
	"reflect"
	"testing"
)

// DirectUnmarshaler implements TextUnmarshaler directly on the value type
type DirectUnmarshaler struct {
	Value string
}

// UnmarshalText implements encoding.TextUnmarshaler on the value receiver
func (d DirectUnmarshaler) UnmarshalText(text []byte) error {
	// This won't work because we need a pointer receiver to modify the value
	// But it helps us test the path where the value type implements the interface
	return nil
}

// TestConvertCustomPrecisePaths tests very specific paths in ConvertCustom
func TestConvertCustomPrecisePaths(t *testing.T) {
	converter := &TypeConverter{}

	// Test the case where we create a non-pointer target and it implements TextUnmarshaler
	// This should hit the first path where target.Type().Implements returns true
	// and targetType.Kind() != reflect.Ptr (line ~519)
	t.Run("non-pointer target implements TextUnmarshaler", func(t *testing.T) {
		// We need a type where the value type (not pointer) implements TextUnmarshaler
		// This is rare but possible

		// Let's test with our existing type but focus on the return path
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})

		// This should go through the second path since TestCustomTypeForCoverage
		// implements TextUnmarshaler on *TestCustomTypeForCoverage, not TestCustomTypeForCoverage
		result, err := converter.ConvertCustom("test-value", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify we got the right result
		if customResult, ok := result.(TestCustomTypeForCoverage); ok {
			if customResult.Value != "test-value" {
				t.Errorf("Expected 'test-value', got '%s'", customResult.Value)
			}
		} else {
			t.Errorf("Expected TestCustomTypeForCoverage, got %T", result)
		}
	})

	// Test to ensure we hit the error return path in the first implementation check
	t.Run("first path error return", func(t *testing.T) {
		// Use ErrorUnmarshaler which returns an error
		ptrType := reflect.TypeOf((*ErrorUnmarshaler)(nil))
		_, err := converter.ConvertCustom("test", ptrType)
		if err == nil {
			t.Error("Expected error from UnmarshalText")
		}
	})

	// Test to ensure we hit the error return path in the second implementation check
	t.Run("second path error return", func(t *testing.T) {
		// Use ErrorUnmarshaler which returns an error
		valueType := reflect.TypeOf(ErrorUnmarshaler{})
		_, err := converter.ConvertCustom("test", valueType)
		if err == nil {
			t.Error("Expected error from UnmarshalText")
		}
	})

	// Test the specific case where targetType.Kind() == reflect.Ptr in second path
	t.Run("second path with pointer type", func(t *testing.T) {
		// This should hit line ~533: return ptrTarget.Interface(), nil
		ptrType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))
		result, err := converter.ConvertCustom("test-value", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}
	})

	// Test the specific case where targetType.Kind() != reflect.Ptr in second path
	t.Run("second path with non-pointer type", func(t *testing.T) {
		// This should hit line ~535: return ptrTarget.Elem().Interface(), nil
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})
		result, err := converter.ConvertCustom("test-value", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}
	})
}

// TestConvertCustomFirstPathCoverage attempts to test the first path more thoroughly
func TestConvertCustomFirstPathCoverage(t *testing.T) {
	converter := &TypeConverter{}

	// Create a type that implements TextUnmarshaler on the value type itself
	// This is unusual but possible
	t.Run("value type implements TextUnmarshaler", func(t *testing.T) {
		// We'll use a different approach - test with a type where we can verify
		// the first path is taken

		// For pointer types in first path
		ptrType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))
		result, err := converter.ConvertCustom("first-path-test", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify the result to ensure the path was taken
		if customResult, ok := result.(*TestCustomTypeForCoverage); ok {
			if customResult.Value != "first-path-test" {
				t.Errorf("Expected 'first-path-test', got '%s'", customResult.Value)
			}
		}
	})
}

// Verify our test types implement the interface correctly
var _ encoding.TextUnmarshaler = (*TestCustomTypeForCoverage)(nil)
var _ encoding.TextUnmarshaler = (*ErrorUnmarshaler)(nil)

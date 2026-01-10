package goarg

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// ErrorUnmarshaler is a type that implements TextUnmarshaler but returns an error
type ErrorUnmarshaler struct {
	Value string
}

func (e *ErrorUnmarshaler) UnmarshalText(text []byte) error {
	return fmt.Errorf("intentional unmarshal error")
}

// TestConvertCustomErrorPaths tests the error paths in ConvertCustom
func TestConvertCustomErrorPaths(t *testing.T) {
	converter := &TypeConverter{}

	// Test UnmarshalText error in first path (target.Type().Implements)
	t.Run("UnmarshalText error in first path", func(t *testing.T) {
		ptrType := reflect.TypeOf((*ErrorUnmarshaler)(nil))
		_, err := converter.ConvertCustom("test", ptrType)
		if err == nil {
			t.Error("Expected error from UnmarshalText")
		}
		if err != nil && !strings.Contains(fmt.Sprintf("%v", err), "failed to unmarshal text") {
			t.Errorf("Expected unmarshal error, got: %v", err)
		}
	})

	// Test UnmarshalText error in second path (ptrType.Implements)
	t.Run("UnmarshalText error in second path", func(t *testing.T) {
		valueType := reflect.TypeOf(ErrorUnmarshaler{})
		_, err := converter.ConvertCustom("test", valueType)
		if err == nil {
			t.Error("Expected error from UnmarshalText")
		}
		if err != nil && !strings.Contains(fmt.Sprintf("%v", err), "failed to unmarshal text") {
			t.Errorf("Expected unmarshal error, got: %v", err)
		}
	})
}

// TestConvertCustomReturnPaths tests the specific return paths
func TestConvertCustomReturnPaths(t *testing.T) {
	converter := &TypeConverter{}

	// Test first path with pointer type return (line ~517)
	t.Run("first path pointer return", func(t *testing.T) {
		ptrType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))
		result, err := converter.ConvertCustom("test-value", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Should return target.Interface() for pointer types
		if _, ok := result.(*TestCustomTypeForCoverage); !ok {
			t.Errorf("Expected *TestCustomTypeForCoverage, got %T", result)
		}
	})

	// Test first path with non-pointer type return (line ~519)
	t.Run("first path non-pointer return", func(t *testing.T) {
		// This is tricky because we need a non-pointer type where target.Type().Implements
		// returns true. Our TestCustomTypeForCoverage implements on *T, so this path
		// is harder to reach. Let's test with a different approach.

		// For now, we'll test the case where we have a non-pointer type
		// but the implementation check fails, leading to the second path
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})
		result, err := converter.ConvertCustom("test-value", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Should return target.Elem().Interface() for non-pointer types in second path
		if _, ok := result.(TestCustomTypeForCoverage); !ok {
			t.Errorf("Expected TestCustomTypeForCoverage, got %T", result)
		}
	})

	// Test second path with pointer type return (line ~533)
	t.Run("second path pointer return", func(t *testing.T) {
		// This tests the case where ptrType.Implements returns true
		// and targetType.Kind() == reflect.Ptr
		ptrType := reflect.TypeOf((*TestCustomTypeForCoverage)(nil))
		result, err := converter.ConvertCustom("test-value", ptrType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Should return ptrTarget.Interface() for pointer types
		if _, ok := result.(*TestCustomTypeForCoverage); !ok {
			t.Errorf("Expected *TestCustomTypeForCoverage, got %T", result)
		}
	})

	// Test second path with non-pointer type return (line ~535)
	t.Run("second path non-pointer return", func(t *testing.T) {
		// This tests the case where ptrType.Implements returns true
		// and targetType.Kind() != reflect.Ptr
		valueType := reflect.TypeOf(TestCustomTypeForCoverage{})
		result, err := converter.ConvertCustom("test-value", valueType)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Should return ptrTarget.Elem().Interface() for non-pointer types
		if _, ok := result.(TestCustomTypeForCoverage); !ok {
			t.Errorf("Expected TestCustomTypeForCoverage, got %T", result)
		}
	})
}

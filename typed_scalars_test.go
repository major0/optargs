package optargs

import (
	"strings"
	"testing"
	"time"
)

func TestScalarSetAndString(t *testing.T) {
	tests := []struct {
		name    string
		val     TypedValue
		input   string
		want    string
		wantErr bool
	}{
		// string
		{"string/simple", NewStringValue("", nil), "hello", "hello", false},
		{"string/empty", NewStringValue("", nil), "", "", false},
		{"string/spaces", NewStringValue("", nil), "hello world", "hello world", false},

		// bool
		{"bool/true", NewBoolValue(false, nil), "true", "true", false},
		{"bool/false", NewBoolValue(false, nil), "false", "false", false},
		{"bool/1", NewBoolValue(false, nil), "1", "true", false},
		{"bool/0", NewBoolValue(false, nil), "0", "false", false},
		{"bool/yes", NewBoolValue(false, nil), "yes", "true", false},
		{"bool/no", NewBoolValue(false, nil), "no", "false", false},
		{"bool/on", NewBoolValue(false, nil), "on", "true", false},
		{"bool/off", NewBoolValue(false, nil), "off", "false", false},
		{"bool/invalid", NewBoolValue(false, nil), "maybe", "", true},

		// int
		{"int/positive", NewIntValue(0, nil), "42", "42", false},
		{"int/negative", NewIntValue(0, nil), "-7", "-7", false},
		{"int/zero", NewIntValue(0, nil), "0", "0", false},
		{"int/invalid", NewIntValue(0, nil), "abc", "", true},

		// int64
		{"int64/max", NewInt64Value(0, nil), "9223372036854775807", "9223372036854775807", false},
		{"int64/min", NewInt64Value(0, nil), "-9223372036854775808", "-9223372036854775808", false},
		{"int64/invalid", NewInt64Value(0, nil), "abc", "", true},

		// uint
		{"uint/positive", NewUintValue(0, nil), "42", "42", false},
		{"uint/zero", NewUintValue(0, nil), "0", "0", false},
		{"uint/negative", NewUintValue(0, nil), "-1", "", true},

		// uint64
		{"uint64/max", NewUint64Value(0, nil), "18446744073709551615", "18446744073709551615", false},
		{"uint64/negative", NewUint64Value(0, nil), "-1", "", true},

		// float64
		{"float64/positive", NewFloat64Value(0, nil), "3.14", "3.14", false},
		{"float64/negative", NewFloat64Value(0, nil), "-2.5", "-2.5", false},
		{"float64/invalid", NewFloat64Value(0, nil), "abc", "", true},

		// duration
		{"duration/seconds", NewDurationValue(0, nil), "5s", "5s", false},
		{"duration/minutes", NewDurationValue(0, nil), "2m30s", "2m30s", false},
		{"duration/hours", NewDurationValue(0, nil), "1h", "1h0m0s", false},
		{"duration/invalid", NewDurationValue(0, nil), "abc", "", true},

		// narrow int overflow
		{"int8/max", NewInt8Value(0, nil), "127", "127", false},
		{"int8/overflow", NewInt8Value(0, nil), "128", "", true},
		{"int8/underflow", NewInt8Value(0, nil), "-129", "", true},
		{"int16/max", NewInt16Value(0, nil), "32767", "32767", false},
		{"int16/overflow", NewInt16Value(0, nil), "32768", "", true},
		{"int32/max", NewInt32Value(0, nil), "2147483647", "2147483647", false},
		{"int32/overflow", NewInt32Value(0, nil), "2147483648", "", true},

		// narrow uint overflow
		{"uint8/max", NewUint8Value(0, nil), "255", "255", false},
		{"uint8/overflow", NewUint8Value(0, nil), "256", "", true},
		{"uint8/negative", NewUint8Value(0, nil), "-1", "", true},
		{"uint16/max", NewUint16Value(0, nil), "65535", "65535", false},
		{"uint16/overflow", NewUint16Value(0, nil), "65536", "", true},
		{"uint32/max", NewUint32Value(0, nil), "4294967295", "4294967295", false},
		{"uint32/overflow", NewUint32Value(0, nil), "4294967296", "", true},

		// float32
		{"float32/valid", NewFloat32Value(0, nil), "3.14", "3.14", false},
		{"float32/invalid", NewFloat32Value(0, nil), "abc", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.val.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), "invalid value") {
					t.Errorf("error %q should contain 'invalid value'", err.Error())
				}
				return
			}
			if got := tt.val.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestScalarTypeNames(t *testing.T) {
	tests := []struct {
		val      TypedValue
		wantType string
	}{
		{NewStringValue("", nil), "string"},
		{NewBoolValue(false, nil), "bool"},
		{NewIntValue(0, nil), "int"},
		{NewInt8Value(0, nil), "int8"},
		{NewInt16Value(0, nil), "int16"},
		{NewInt32Value(0, nil), "int32"},
		{NewInt64Value(0, nil), "int64"},
		{NewUintValue(0, nil), "uint"},
		{NewUint8Value(0, nil), "uint8"},
		{NewUint16Value(0, nil), "uint16"},
		{NewUint32Value(0, nil), "uint32"},
		{NewUint64Value(0, nil), "uint64"},
		{NewFloat32Value(0, nil), "float32"},
		{NewFloat64Value(0, nil), "float64"},
		{NewDurationValue(0, nil), "duration"},
	}
	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			if got := tt.val.Type(); got != tt.wantType {
				t.Errorf("Type() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

func TestScalarNilPointers(t *testing.T) {
	tests := []struct {
		name string
		val  TypedValue
	}{
		{"string", NewStringValue("test", nil)},
		{"bool", NewBoolValue(true, nil)},
		{"int", NewIntValue(42, nil)},
		{"int64", NewInt64Value(99, nil)},
		{"uint", NewUintValue(7, nil)},
		{"uint64", NewUint64Value(100, nil)},
		{"float64", NewFloat64Value(3.14, nil)},
		{"duration", NewDurationValue(time.Second, nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val.String() == "" && tt.name != "string" {
				t.Error("nil pointer constructor produced empty string for non-string type")
			}
		})
	}
}

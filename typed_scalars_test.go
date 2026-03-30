package optargs

import (
	"strings"
	"testing"
	"time"
)

func TestScalarStringValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple", "hello", "hello", false},
		{"empty", "", "", false},
		{"spaces", "hello world", "hello world", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s string
			v := NewStringValue("", &s)
			if err := v.Set(tt.input); (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got := v.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestScalarBoolValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"true", "true", "true", false},
		{"false", "false", "false", false},
		{"1", "1", "true", false},
		{"0", "0", "false", false},
		{"yes", "yes", "true", false},
		{"no", "no", "false", false},
		{"on", "on", "true", false},
		{"off", "off", "false", false},
		{"t", "t", "true", false},
		{"f", "f", "false", false},
		{"y", "y", "true", false},
		{"n", "n", "false", false},
		{"invalid", "maybe", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bool
			v := NewBoolValue(false, &b)
			err := v.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && v.String() != tt.want {
				t.Errorf("String() = %q, want %q", v.String(), tt.want)
			}
			if tt.wantErr && !strings.Contains(err.Error(), "invalid value") {
				t.Errorf("error %q should contain 'invalid value'", err.Error())
			}
		})
	}
}

func TestScalarIntValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"positive", "42", "42", false},
		{"negative", "-7", "-7", false},
		{"zero", "0", "0", false},
		{"invalid", "abc", "", true},
		{"float", "3.14", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i int
			v := NewIntValue(0, &i)
			err := v.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && v.String() != tt.want {
				t.Errorf("String() = %q, want %q", v.String(), tt.want)
			}
		})
	}
}

func TestScalarInt64Value(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"positive", "9223372036854775807", "9223372036854775807", false},
		{"negative", "-9223372036854775808", "-9223372036854775808", false},
		{"invalid", "abc", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i int64
			v := NewInt64Value(0, &i)
			err := v.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && v.String() != tt.want {
				t.Errorf("String() = %q, want %q", v.String(), tt.want)
			}
		})
	}
}

func TestScalarUintValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"positive", "42", "42", false},
		{"zero", "0", "0", false},
		{"negative", "-1", "", true},
		{"invalid", "abc", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uint
			v := NewUintValue(0, &u)
			err := v.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && v.String() != tt.want {
				t.Errorf("String() = %q, want %q", v.String(), tt.want)
			}
		})
	}
}

func TestScalarUint64Value(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"max", "18446744073709551615", "18446744073709551615", false},
		{"zero", "0", "0", false},
		{"negative", "-1", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uint64
			v := NewUint64Value(0, &u)
			err := v.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && v.String() != tt.want {
				t.Errorf("String() = %q, want %q", v.String(), tt.want)
			}
		})
	}
}

func TestScalarFloat64Value(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"positive", "3.14", "3.14", false},
		{"negative", "-2.5", "-2.5", false},
		{"integer", "42", "42", false},
		{"invalid", "abc", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f float64
			v := NewFloat64Value(0, &f)
			err := v.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && v.String() != tt.want {
				t.Errorf("String() = %q, want %q", v.String(), tt.want)
			}
		})
	}
}

func TestScalarDurationValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"seconds", "5s", "5s", false},
		{"minutes", "2m30s", "2m30s", false},
		{"hours", "1h", "1h0m0s", false},
		{"millis", "500ms", "500ms", false},
		{"invalid", "abc", "", true},
		{"number_only", "42", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d time.Duration
			v := NewDurationValue(0, &d)
			err := v.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && v.String() != tt.want {
				t.Errorf("String() = %q, want %q", v.String(), tt.want)
			}
		})
	}
}

func TestScalarNilPointers(t *testing.T) {
	// All constructors should handle nil pointers by allocating internally.
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

func TestScalarRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		val  TypedValue
	}{
		{"string", NewStringValue("hello", nil)},
		{"bool", NewBoolValue(true, nil)},
		{"int", NewIntValue(42, nil)},
		{"int64", NewInt64Value(-99, nil)},
		{"uint", NewUintValue(7, nil)},
		{"uint64", NewUint64Value(100, nil)},
		{"float64", NewFloat64Value(3.14, nil)},
		{"duration", NewDurationValue(5 * time.Second, nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.val.String()
			// Create a fresh value and Set from the string representation.
			var fresh TypedValue
			switch tt.name {
			case "string":
				fresh = NewStringValue("", nil)
			case "bool":
				fresh = NewBoolValue(false, nil)
			case "int":
				fresh = NewIntValue(0, nil)
			case "int64":
				fresh = NewInt64Value(0, nil)
			case "uint":
				fresh = NewUintValue(0, nil)
			case "uint64":
				fresh = NewUint64Value(0, nil)
			case "float64":
				fresh = NewFloat64Value(0, nil)
			case "duration":
				fresh = NewDurationValue(0, nil)
			}
			if err := fresh.Set(s); err != nil {
				t.Fatalf("Set(%q) error = %v", s, err)
			}
			if got := fresh.String(); got != s {
				t.Errorf("round-trip: got %q, want %q", got, s)
			}
		})
	}
}

func TestNarrowIntOverflow(t *testing.T) {
	tests := []struct {
		name    string
		val     TypedValue
		input   string
		wantErr bool
	}{
		// int8: -128 to 127
		{"int8_valid", NewInt8Value(0, nil), "127", false},
		{"int8_min", NewInt8Value(0, nil), "-128", false},
		{"int8_overflow", NewInt8Value(0, nil), "128", true},
		{"int8_underflow", NewInt8Value(0, nil), "-129", true},

		// int16: -32768 to 32767
		{"int16_valid", NewInt16Value(0, nil), "32767", false},
		{"int16_overflow", NewInt16Value(0, nil), "32768", true},

		// int32: -2147483648 to 2147483647
		{"int32_valid", NewInt32Value(0, nil), "2147483647", false},
		{"int32_overflow", NewInt32Value(0, nil), "2147483648", true},

		// uint8: 0 to 255
		{"uint8_valid", NewUint8Value(0, nil), "255", false},
		{"uint8_overflow", NewUint8Value(0, nil), "256", true},
		{"uint8_negative", NewUint8Value(0, nil), "-1", true},

		// uint16: 0 to 65535
		{"uint16_valid", NewUint16Value(0, nil), "65535", false},
		{"uint16_overflow", NewUint16Value(0, nil), "65536", true},

		// uint32: 0 to 4294967295
		{"uint32_valid", NewUint32Value(0, nil), "4294967295", false},
		{"uint32_overflow", NewUint32Value(0, nil), "4294967296", true},

		// float32
		{"float32_valid", NewFloat32Value(0, nil), "3.14", false},
		{"float32_invalid", NewFloat32Value(0, nil), "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.val.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestNarrowTypeNames(t *testing.T) {
	tests := []struct {
		name     string
		val      TypedValue
		wantType string
	}{
		{"int8", NewInt8Value(0, nil), "int8"},
		{"int16", NewInt16Value(0, nil), "int16"},
		{"int32", NewInt32Value(0, nil), "int32"},
		{"uint8", NewUint8Value(0, nil), "uint8"},
		{"uint16", NewUint16Value(0, nil), "uint16"},
		{"uint32", NewUint32Value(0, nil), "uint32"},
		{"float32", NewFloat32Value(0, nil), "float32"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.Type(); got != tt.wantType {
				t.Errorf("Type() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

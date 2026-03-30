package goarg

import (
	"reflect"
	"testing"
	"time"

	"github.com/major0/optargs"
)

// testTextType implements encoding.TextUnmarshaler for testing.
type testTextType struct{ data string }

func (t *testTextType) UnmarshalText(text []byte) error { t.data = string(text); return nil }
func (t *testTextType) MarshalText() ([]byte, error)    { return []byte(t.data), nil }

func TestTypedValueForField(t *testing.T) {
	type allTypes struct {
		S   string
		B   bool
		I   int
		I8  int8
		I16 int16
		I32 int32
		I64 int64
		U   uint
		U8  uint8
		U16 uint16
		U32 uint32
		U64 uint64
		F32 float32
		F64 float64
		D   time.Duration

		SS  []string
		SB  []bool
		SI  []int
		SI32 []int32
		SI64 []int64
		SU  []uint
		SF32 []float32
		SF64 []float64
		SD  []time.Duration

		MSS  map[string]string
		MSI  map[string]int
		MSI64 map[string]int64

		TU  testTextType
	}

	tests := []struct {
		name     string
		field    string
		input    string
		wantType string
	}{
		{"string", "S", "hello", "string"},
		{"bool", "B", "true", "bool"},
		{"int", "I", "42", "int"},
		{"int8", "I8", "7", "int8"},
		{"int16", "I16", "300", "int16"},
		{"int32", "I32", "100000", "int32"},
		{"int64", "I64", "999", "int64"},
		{"uint", "U", "10", "uint"},
		{"uint8", "U8", "255", "uint8"},
		{"uint16", "U16", "1000", "uint16"},
		{"uint32", "U32", "70000", "uint32"},
		{"uint64", "U64", "99999", "uint64"},
		{"float32", "F32", "3.14", "float32"},
		{"float64", "F64", "2.718", "float64"},
		{"duration", "D", "5s", "duration"},

		{"[]string", "SS", "a", "stringSlice"},
		{"[]bool", "SB", "true", "boolSlice"},
		{"[]int", "SI", "1", "intSlice"},
		{"[]int32", "SI32", "2", "int32Slice"},
		{"[]int64", "SI64", "3", "int64Slice"},
		{"[]uint", "SU", "4", "uintSlice"},
		{"[]float32", "SF32", "1.1", "float32Slice"},
		{"[]float64", "SF64", "2.2", "float64Slice"},
		{"[]duration", "SD", "1s", "durationSlice"},

		{"map[string]string", "MSS", "k=v", "stringToString"},
		{"map[string]int", "MSI", "k=1", "stringToInt"},
		{"map[string]int64", "MSI64", "k=2", "stringToInt64"},

		{"TextUnmarshaler", "TU", "hello", "textUnmarshaler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := &allTypes{}
			destValue := reflect.ValueOf(dest).Elem()
			sf, ok := destValue.Type().FieldByName(tt.field)
			if !ok {
				t.Fatalf("field %s not found", tt.field)
			}
			fieldValue := destValue.FieldByName(tt.field)
			meta := &FieldMetadata{
				Name:       tt.field,
				FieldIndex: sf.Index[0],
				Type:       sf.Type,
			}

			tv, err := typedValueForField(fieldValue, meta)
			if err != nil {
				t.Fatalf("typedValueForField: %v", err)
			}
			if got := tv.Type(); got != tt.wantType {
				t.Errorf("Type() = %q, want %q", got, tt.wantType)
			}
			if err := tv.Set(tt.input); err != nil {
				t.Fatalf("Set(%q): %v", tt.input, err)
			}
		})
	}
}

func TestTypedValueForFieldBoolValuer(t *testing.T) {
	dest := &struct{ B bool }{}
	destValue := reflect.ValueOf(dest).Elem()
	fieldValue := destValue.FieldByName("B")
	meta := &FieldMetadata{Name: "B", FieldIndex: 0, Type: fieldValue.Type()}

	tv, err := typedValueForField(fieldValue, meta)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := tv.(optargs.BoolValuer); !ok {
		t.Error("bool TypedValue should implement BoolValuer")
	}
}

func TestTypedValueForFieldUnsupported(t *testing.T) {
	type unsupported struct {
		Ch chan int
	}
	dest := &unsupported{}
	destValue := reflect.ValueOf(dest).Elem()
	fieldValue := destValue.FieldByName("Ch")
	meta := &FieldMetadata{Name: "Ch", FieldIndex: 0, Type: fieldValue.Type()}

	_, err := typedValueForField(fieldValue, meta)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

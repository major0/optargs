package goarg

import (
	"net"
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

func TestTypedValueForFieldPointerTypes(t *testing.T) {
	type ptrTypes struct {
		S *string
		I *int
		B *bool
	}

	dest := &ptrTypes{}
	destValue := reflect.ValueOf(dest).Elem()

	tests := []struct {
		name  string
		field string
		input string
	}{
		{"*string", "S", "hello"},
		{"*int", "I", "42"},
		{"*bool", "B", "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf, _ := destValue.Type().FieldByName(tt.field)
			fv := destValue.FieldByName(tt.field)
			meta := &FieldMetadata{Name: tt.field, FieldIndex: sf.Index[0], Type: sf.Type}
			tv, err := typedValueForField(fv, meta)
			if err != nil {
				t.Fatalf("typedValueForField: %v", err)
			}
			if err := tv.Set(tt.input); err != nil {
				t.Fatalf("Set(%q): %v", tt.input, err)
			}
		})
	}

	// Verify the pointer fields were allocated and set.
	if dest.S == nil || *dest.S != "hello" {
		t.Errorf("S = %v, want *hello", dest.S)
	}
	if dest.I == nil || *dest.I != 42 {
		t.Errorf("I = %v, want *42", dest.I)
	}
	if dest.B == nil || *dest.B != true {
		t.Errorf("B = %v, want *true", dest.B)
	}
}

func TestPointerFieldEndToEnd(t *testing.T) {
	type Args struct {
		Name *string `arg:"--name"`
		Port *int    `arg:"--port"`
	}

	// When provided.
	dest := &Args{}
	if err := ParseArgs(dest, []string{"--name", "alice", "--port", "8080"}); err != nil {
		t.Fatal(err)
	}
	if dest.Name == nil || *dest.Name != "alice" {
		t.Errorf("Name = %v, want *alice", dest.Name)
	}
	if dest.Port == nil || *dest.Port != 8080 {
		t.Errorf("Port = %v, want *8080", dest.Port)
	}

	// When not provided — should remain nil.
	dest2 := &Args{}
	if err := ParseArgs(dest2, []string{}); err != nil {
		t.Fatal(err)
	}
	if dest2.Name != nil {
		t.Errorf("Name should be nil when not provided, got %v", *dest2.Name)
	}
	if dest2.Port != nil {
		t.Errorf("Port should be nil when not provided, got %v", *dest2.Port)
	}
}

func TestBareEnvTag(t *testing.T) {
	type Args struct {
		Workers int `arg:"env"`
	}
	t.Setenv("WORKERS", "4")
	dest := &Args{}
	if err := ParseArgs(dest, []string{}); err != nil {
		t.Fatal(err)
	}
	if dest.Workers != 4 {
		t.Errorf("Workers = %d, want 4", dest.Workers)
	}
}

func TestBareEnvTagCamelCase(t *testing.T) {
	type Args struct {
		NumWorkers int `arg:"env"`
	}
	t.Setenv("NUM_WORKERS", "8")
	dest := &Args{}
	if err := ParseArgs(dest, []string{}); err != nil {
		t.Fatal(err)
	}
	if dest.NumWorkers != 8 {
		t.Errorf("NumWorkers = %d, want 8", dest.NumWorkers)
	}
}

func TestEnvPrefix(t *testing.T) {
	type Args struct {
		Token string `arg:"--token,env:API_TOKEN"`
	}
	t.Setenv("MYAPP_API_TOKEN", "secret123")
	dest := &Args{}
	p, err := NewParser(Config{EnvPrefix: "MYAPP_"}, dest)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	if dest.Token != "secret123" {
		t.Errorf("Token = %q, want %q", dest.Token, "secret123")
	}
}

func TestSeparateTag(t *testing.T) {
	// "separate" is a no-op for us (our default is already one-value-per-flag),
	// but the tag must be accepted without error.
	type Args struct {
		Files []string `arg:"-f,--file,separate"`
	}
	dest := &Args{}
	if err := ParseArgs(dest, []string{"--file", "a.txt", "--file", "b.txt"}); err != nil {
		t.Fatal(err)
	}
	if len(dest.Files) != 2 || dest.Files[0] != "a.txt" || dest.Files[1] != "b.txt" {
		t.Errorf("Files = %v, want [a.txt b.txt]", dest.Files)
	}
}

func TestTextUnmarshalerEndToEnd(t *testing.T) {
	// net.IP implements encoding.TextUnmarshaler.
	type Args struct {
		Addr net.IP `arg:"--addr"`
	}
	dest := &Args{}
	if err := ParseArgs(dest, []string{"--addr", "192.168.1.1"}); err != nil {
		t.Fatal(err)
	}
	if dest.Addr.String() != "192.168.1.1" {
		t.Errorf("Addr = %q, want %q", dest.Addr.String(), "192.168.1.1")
	}
}

func TestEnvVarSliceCommaSeparated(t *testing.T) {
	type Args struct {
		Workers []int `arg:"--workers,env:WORKERS"`
	}
	t.Setenv("WORKERS", "1,99")
	dest := &Args{}
	if err := ParseArgs(dest, []string{}); err != nil {
		t.Fatal(err)
	}
	if len(dest.Workers) != 2 || dest.Workers[0] != 1 || dest.Workers[1] != 99 {
		t.Errorf("Workers = %v, want [1 99]", dest.Workers)
	}
}

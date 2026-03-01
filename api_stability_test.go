package optargs

import (
	"reflect"
	"testing"
)

// reference types used across multiple reflection checks.
var (
	stringSliceType = reflect.TypeOf([]string{})
	argTypeType     = reflect.TypeOf(ArgType(0))
	parserPtrType   = reflect.TypeOf(&Parser{})
	flagSliceType   = reflect.TypeOf([]Flag{})
	stringType      = reflect.TypeOf("")
)

// TestAPIStability validates that the public API remains stable and backward compatible.
func TestAPIStability(t *testing.T) {
	t.Run("public_types", func(t *testing.T) {
		assertKind(t, ArgType(0), reflect.Int, "ArgType")
		assertKind(t, ParseMode(0), reflect.Int, "ParseMode")
		assertKind(t, Flag{}, reflect.Struct, "Flag")
		assertKind(t, Option{}, reflect.Struct, "Option")
		assertKind(t, Parser{}, reflect.Struct, "Parser")
	})

	t.Run("enum_values", func(t *testing.T) {
		argTypes := []struct {
			name string
			got  ArgType
			want int
		}{
			{"NoArgument", NoArgument, 0},
			{"RequiredArgument", RequiredArgument, 1},
			{"OptionalArgument", OptionalArgument, 2},
		}
		for _, tt := range argTypes {
			if int(tt.got) != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		}

		parseModes := []struct {
			name string
			got  ParseMode
			want int
		}{
			{"ParseDefault", ParseDefault, 0},
			{"ParseNonOpts", ParseNonOpts, 1},
			{"ParsePosixlyCorrect", ParsePosixlyCorrect, 2},
		}
		for _, tt := range parseModes {
			if int(tt.got) != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		}
	})

	t.Run("struct_fields", func(t *testing.T) {
		assertField(t, Flag{}, "Name", reflect.String)
		assertFieldType(t, Flag{}, "HasArg", argTypeType)
		assertField(t, Option{}, "Name", reflect.String)
		assertField(t, Option{}, "HasArg", reflect.Bool)
		assertField(t, Option{}, "Arg", reflect.String)
		assertFieldType(t, Parser{}, "Args", stringSliceType)
	})

	t.Run("function_signatures", func(t *testing.T) {
		assertSignature(t, "GetOpt", GetOpt,
			[]reflect.Type{stringSliceType, stringType},
			[]reflect.Type{parserPtrType})
		assertSignature(t, "GetOptLong", GetOptLong,
			[]reflect.Type{stringSliceType, stringType, flagSliceType},
			[]reflect.Type{parserPtrType})
		assertSignature(t, "GetOptLongOnly", GetOptLongOnly,
			[]reflect.Type{stringSliceType, stringType, flagSliceType},
			[]reflect.Type{parserPtrType})
	})

	t.Run("backward_compatibility", func(t *testing.T) {
		// Basic GetOpt usage
		p, err := GetOpt([]string{"-a", "-b", "value"}, "ab:")
		if err != nil {
			t.Fatalf("GetOpt: %v", err)
		}
		opts := collectOpts(p)
		if len(opts) != 2 {
			t.Errorf("GetOpt options: got %d, want 2", len(opts))
		}

		// GetOptLong usage
		longOpts := []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "output", HasArg: RequiredArgument},
		}
		p, err = GetOptLong([]string{"--verbose", "--output", "file.txt"}, "vo:", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		opts = collectOpts(p)
		if len(opts) != 2 {
			t.Errorf("GetOptLong options: got %d, want 2", len(opts))
		}

		// GetOptLongOnly usage
		p, err = GetOptLongOnly([]string{"-verbose"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLongOnly: %v", err)
		}
		opts = collectOpts(p)
		if len(opts) != 1 {
			t.Errorf("GetOptLongOnly options: got %d, want 1", len(opts))
		}
	})

	t.Run("posixly_correct_compatibility", func(t *testing.T) {
		// Default mode processes all options
		p, err := GetOpt([]string{"-a", "file", "-b"}, "ab")
		if err != nil {
			t.Fatalf("default mode: %v", err)
		}
		if n := len(collectOpts(p)); n != 2 {
			t.Errorf("default mode options: got %d, want 2", n)
		}

		// + prefix stops at first non-option
		p, err = GetOpt([]string{"-a", "file", "-b"}, "+ab")
		if err != nil {
			t.Fatalf("POSIX mode: %v", err)
		}
		if n := len(collectOpts(p)); n != 1 {
			t.Errorf("POSIX mode options: got %d, want 1", n)
		}
	})
}

// assertKind checks that a value's reflect.Kind matches want.
func assertKind(t *testing.T, v interface{}, want reflect.Kind, name string) {
	t.Helper()
	if got := reflect.TypeOf(v).Kind(); got != want {
		t.Errorf("%s kind = %v, want %v", name, got, want)
	}
}

// assertField checks that a struct has a field with the given kind.
func assertField(t *testing.T, v interface{}, field string, want reflect.Kind) {
	t.Helper()
	f, ok := reflect.TypeOf(v).FieldByName(field)
	if !ok {
		t.Errorf("missing field %s", field)
		return
	}
	if f.Type.Kind() != want {
		t.Errorf("field %s kind = %v, want %v", field, f.Type.Kind(), want)
	}
}

// assertFieldType checks that a struct field has an exact reflect.Type.
func assertFieldType(t *testing.T, v interface{}, field string, want reflect.Type) {
	t.Helper()
	f, ok := reflect.TypeOf(v).FieldByName(field)
	if !ok {
		t.Errorf("missing field %s", field)
		return
	}
	if f.Type != want {
		t.Errorf("field %s type = %v, want %v", field, f.Type, want)
	}
}

// assertSignature checks that fn has the expected parameter types and that
// each return type in wantOut matches (error return is checked by name).
func assertSignature(t *testing.T, name string, fn interface{}, wantIn, wantOut []reflect.Type) {
	t.Helper()
	ft := reflect.TypeOf(fn)
	if ft.Kind() != reflect.Func {
		t.Errorf("%s is not a function", name)
		return
	}
	if ft.NumIn() != len(wantIn) {
		t.Errorf("%s params: got %d, want %d", name, ft.NumIn(), len(wantIn))
		return
	}
	for i, want := range wantIn {
		if ft.In(i) != want {
			t.Errorf("%s param %d: got %v, want %v", name, i, ft.In(i), want)
		}
	}
	// Check non-error return types; last return is always error.
	if ft.NumOut() != len(wantOut)+1 {
		t.Errorf("%s returns: got %d, want %d", name, ft.NumOut(), len(wantOut)+1)
		return
	}
	for i, want := range wantOut {
		if ft.Out(i) != want {
			t.Errorf("%s return %d: got %v, want %v", name, i, ft.Out(i), want)
		}
	}
	if ft.Out(len(wantOut)).String() != "error" {
		t.Errorf("%s last return: got %v, want error", name, ft.Out(len(wantOut)))
	}
}

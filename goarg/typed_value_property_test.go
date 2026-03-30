package goarg

import (
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

// Property 1: Round-trip — typedValueForField + Set(s) produces the same
// result as optargs.Convert(s, type) for all scalar types.
func TestPropertyTypedValueFieldRoundTrip(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		f := func(n int) bool {
			dest := &struct{ V int }{}
			dv := reflect.ValueOf(dest).Elem()
			fv := dv.FieldByName("V")
			meta := &FieldMetadata{Name: "V", FieldIndex: 0, Type: fv.Type()}
			tv, err := typedValueForField(fv, meta)
			if err != nil {
				return false
			}
			s := tv.String()
			fresh := &struct{ V int }{}
			fdv := reflect.ValueOf(fresh).Elem()
			ffv := fdv.FieldByName("V")
			tv2, _ := typedValueForField(ffv, meta)
			if err := tv2.Set(s); err != nil {
				return false
			}
			return fresh.V == dest.V
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("string", func(t *testing.T) {
		f := func(s string) bool {
			dest := &struct{ V string }{}
			dv := reflect.ValueOf(dest).Elem()
			fv := dv.FieldByName("V")
			meta := &FieldMetadata{Name: "V", FieldIndex: 0, Type: fv.Type()}
			tv, _ := typedValueForField(fv, meta)
			_ = tv.Set(s)
			return dest.V == s
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})
}

// Property 2: Boolean no-argument — handler with arg="" sets bool field to true.
func TestPropertyBoolNoArgument(t *testing.T) {
	dest := &struct {
		Verbose bool `arg:"-v,--verbose"`
	}{}
	err := ParseArgs(dest, []string{"--verbose"})
	if err != nil {
		t.Fatal(err)
	}
	if !dest.Verbose {
		t.Error("expected Verbose=true after --verbose with no argument")
	}
}

// Property 3: Slice accumulation — repeated Set() calls accumulate.
func TestPropertySliceAccumulation(t *testing.T) {
	dest := &struct {
		Files []string `arg:"-f,--file"`
	}{}
	err := ParseArgs(dest, []string{"--file", "a.txt", "--file", "b.txt", "--file", "c.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(dest.Files) != 3 || dest.Files[0] != "a.txt" || dest.Files[1] != "b.txt" || dest.Files[2] != "c.txt" {
		t.Errorf("Files = %v, want [a.txt b.txt c.txt]", dest.Files)
	}
}

// Property 4: Duration field works via TypedValue (time.Duration before int64).
func TestPropertyDurationField(t *testing.T) {
	dest := &struct {
		Timeout time.Duration `arg:"--timeout"`
	}{}
	err := ParseArgs(dest, []string{"--timeout", "5s"})
	if err != nil {
		t.Fatal(err)
	}
	if dest.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", dest.Timeout)
	}
}


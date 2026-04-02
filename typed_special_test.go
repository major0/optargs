package optargs

import (
	"errors"
	"net"
	"testing"
)

func TestStringArrayValue(t *testing.T) {
	var arr []string
	v := NewStringArrayValue(nil, &arr)

	// Should NOT split on commas.
	if err := v.Set("a,b,c"); err != nil {
		t.Fatal(err)
	}
	if len(arr) != 1 || arr[0] != "a,b,c" {
		t.Errorf("expected single element 'a,b,c', got %v", arr)
	}

	// Repeated calls append.
	if err := v.Set("d"); err != nil {
		t.Fatal(err)
	}
	if len(arr) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr))
	}

	if v.Type() != "stringArray" {
		t.Errorf("Type() = %q, want %q", v.Type(), "stringArray")
	}
	if got := v.String(); got != "[a,b,c,d]" {
		t.Errorf("String() = %q, want %q", got, "[a,b,c,d]")
	}
}

func TestCountValue(t *testing.T) {
	var c int
	v := NewCountValue(0, &c)

	// Each Set() increments.
	for range 3 {
		if err := v.Set(""); err != nil {
			t.Fatal(err)
		}
	}
	if c != 3 {
		t.Errorf("count = %d, want 3", c)
	}
	if v.String() != "3" {
		t.Errorf("String() = %q, want %q", v.String(), "3")
	}

	// Implements BoolValuer.
	bv, ok := v.(BoolValuer)
	if !ok {
		t.Fatal("CountValue does not implement BoolValuer")
	}
	if !bv.IsBoolFlag() {
		t.Error("IsBoolFlag() = false, want true")
	}
}

// testTextType implements encoding.TextUnmarshaler and TextMarshaler.
type testTextType struct{ data string }

func (t *testTextType) UnmarshalText(text []byte) error {
	t.data = string(text)
	return nil
}

func (t *testTextType) MarshalText() ([]byte, error) {
	return []byte(t.data), nil
}

func TestTextValue(t *testing.T) {
	dest := &testTextType{}
	init := &testTextType{data: "initial"}
	v := NewTextValue(init, dest)

	if got := v.String(); got != "initial" {
		t.Errorf("String() = %q, want %q", got, "initial")
	}

	if err := v.Set("updated"); err != nil {
		t.Fatal(err)
	}
	if dest.data != "updated" {
		t.Errorf("dest.data = %q, want %q", dest.data, "updated")
	}
	if got := v.String(); got != "updated" {
		t.Errorf("String() after Set = %q, want %q", got, "updated")
	}
}

func TestTextValueWithNetIP(t *testing.T) {
	// net.IP implements TextUnmarshaler via *net.IP.
	ip := net.IP{}
	v := NewTextValue(nil, &ip)
	if err := v.Set("192.168.1.1"); err != nil {
		t.Fatal(err)
	}
	if ip.String() != "192.168.1.1" {
		t.Errorf("IP = %q, want %q", ip.String(), "192.168.1.1")
	}
}

func TestFuncValue(t *testing.T) {
	var called string
	v := NewFuncValue(func(s string) error {
		called = s
		return nil
	})
	if err := v.Set("hello"); err != nil {
		t.Fatal(err)
	}
	if called != "hello" {
		t.Errorf("callback got %q, want %q", called, "hello")
	}
	if v.Type() != "func" {
		t.Errorf("Type() = %q, want %q", v.Type(), "func")
	}
}

func TestFuncValueError(t *testing.T) {
	v := NewFuncValue(func(string) error {
		return errors.New("custom error")
	})
	if err := v.Set("anything"); err == nil {
		t.Fatal("expected error")
	}
}

func TestBoolFuncValue(t *testing.T) {
	var called string
	v := NewBoolFuncValue(func(s string) error {
		called = s
		return nil
	})
	if err := v.Set("true"); err != nil {
		t.Fatal(err)
	}
	if called != "true" {
		t.Errorf("callback got %q, want %q", called, "true")
	}

	bv, ok := v.(BoolValuer)
	if !ok {
		t.Fatal("BoolFuncValue does not implement BoolValuer")
	}
	if !bv.IsBoolFlag() {
		t.Error("IsBoolFlag() = false, want true")
	}
}

func TestSpecialNilPointers(t *testing.T) {
	// StringArray and Count should handle nil pointers.
	sa := NewStringArrayValue(nil, nil)
	if err := sa.Set("test"); err != nil {
		t.Errorf("StringArray nil pointer: %v", err)
	}

	cv := NewCountValue(0, nil)
	if err := cv.Set(""); err != nil {
		t.Errorf("Count nil pointer: %v", err)
	}
}

package pflag

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

// bytesFlagMethods bundles the four-method family for a bytes flag type,
// allowing a single table-driven test to cover both BytesHex and BytesBase64.
type bytesFlagMethods struct {
	varFn  func(fs *FlagSet, p *[]byte, name string, val []byte, usage string)
	varPFn func(fs *FlagSet, p *[]byte, name, short string, val []byte, usage string)
	ptrFn  func(fs *FlagSet, name string, val []byte, usage string) *[]byte
	ptrPFn func(fs *FlagSet, name, short string, val []byte, usage string) *[]byte
}

// testBytesFlagSetMethods is the shared helper for BytesHex and BytesBase64
// FlagSet method families.
func testBytesFlagSetMethods(t *testing.T, m bytesFlagMethods, wantType string) {
	t.Helper()
	defaultVal := []byte{0xde, 0xad, 0xbe, 0xef}

	t.Run("Var", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var dest []byte
		m.varFn(fs, &dest, "data", defaultVal, "usage")

		if !bytes.Equal(dest, defaultVal) {
			t.Errorf("default = %x, want %x", dest, defaultVal)
		}
		f := fs.Lookup("data")
		if f == nil {
			t.Fatal("flag not registered")
		}
		if f.Value.Type() != wantType {
			t.Errorf("Type() = %q, want %q", f.Value.Type(), wantType)
		}
	})

	t.Run("VarP", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var dest []byte
		m.varPFn(fs, &dest, "data", "d", defaultVal, "usage")

		if !bytes.Equal(dest, defaultVal) {
			t.Errorf("default = %x, want %x", dest, defaultVal)
		}
		if fs.ShorthandLookup("d") == nil {
			t.Error("shorthand not registered")
		}
	})

	t.Run("Ptr", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		p := m.ptrFn(fs, "data", defaultVal, "usage")

		if !bytes.Equal(*p, defaultVal) {
			t.Errorf("default = %x, want %x", *p, defaultVal)
		}
	})

	t.Run("PtrP", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		p := m.ptrPFn(fs, "data", "d", defaultVal, "usage")

		if !bytes.Equal(*p, defaultVal) {
			t.Errorf("default = %x, want %x", *p, defaultVal)
		}
		if fs.ShorthandLookup("d") == nil {
			t.Error("shorthand not registered")
		}
	})
}

// TestBytesHexFlagSetMethods tests the FlagSet BytesHex method family.
// **Validates: Requirements 2.1, 2.5**
func TestBytesHexFlagSetMethods(t *testing.T) {
	testBytesFlagSetMethods(t, bytesFlagMethods{
		varFn:  (*FlagSet).BytesHexVar,
		varPFn: (*FlagSet).BytesHexVarP,
		ptrFn:  (*FlagSet).BytesHex,
		ptrPFn: (*FlagSet).BytesHexP,
	}, "bytesHex")
}

// TestBytesHexParsing tests parsing valid and invalid hex strings via --flag=value.
// **Validates: Requirements 2.1, 2.5**
func TestBytesHexParsing(t *testing.T) {
	t.Run("valid hex", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var dest []byte
		fs.BytesHexVar(&dest, "data", nil, "")

		if err := fs.Parse([]string{"--data=cafebabe"}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		want := []byte{0xca, 0xfe, 0xba, 0xbe}
		if !bytes.Equal(dest, want) {
			t.Errorf("got %x, want %x", dest, want)
		}
	})

	t.Run("invalid hex", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.BytesHexVar(new([]byte), "data", nil, "")

		if err := fs.Parse([]string{"--data=zzzz"}); err == nil {
			t.Error("expected error for invalid hex")
		}
	})

	t.Run("odd-length hex", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.BytesHexVar(new([]byte), "data", nil, "")

		if err := fs.Parse([]string{"--data=abc"}); err == nil {
			t.Error("expected error for odd-length hex")
		}
	})

	t.Run("empty hex", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var dest []byte
		fs.BytesHexVar(&dest, "data", []byte{0xff}, "")

		if err := fs.Parse([]string{"--data="}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if len(dest) != 0 {
			t.Errorf("got %x, want empty", dest)
		}
	})

	t.Run("shorthand", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var dest []byte
		fs.BytesHexVarP(&dest, "data", "d", nil, "")

		if err := fs.Parse([]string{"-d", "ff00"}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		want := []byte{0xff, 0x00}
		if !bytes.Equal(dest, want) {
			t.Errorf("got %x, want %x", dest, want)
		}
	})
}

// TestBytesHexType verifies Type() returns "bytesHex".
// **Validates: Requirements 2.5**
func TestBytesHexType(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.BytesHexVar(new([]byte), "data", nil, "")

	flag := fs.Lookup("data")
	if flag.Value.Type() != "bytesHex" {
		t.Errorf("Type() = %q, want %q", flag.Value.Type(), "bytesHex")
	}
}

// TestBytesBase64FlagSetMethods tests the FlagSet BytesBase64 method family.
// **Validates: Requirements 3.1, 3.5**
func TestBytesBase64FlagSetMethods(t *testing.T) {
	testBytesFlagSetMethods(t, bytesFlagMethods{
		varFn:  (*FlagSet).BytesBase64Var,
		varPFn: (*FlagSet).BytesBase64VarP,
		ptrFn:  (*FlagSet).BytesBase64,
		ptrPFn: (*FlagSet).BytesBase64P,
	}, "bytesBase64")
}

// TestBytesBase64Parsing tests parsing valid and invalid base64 strings via --flag=value.
// **Validates: Requirements 3.1, 3.5**
func TestBytesBase64Parsing(t *testing.T) {
	t.Run("valid base64", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var dest []byte
		fs.BytesBase64Var(&dest, "data", nil, "")

		encoded := base64.StdEncoding.EncodeToString([]byte{0xca, 0xfe, 0xba, 0xbe})
		if err := fs.Parse([]string{"--data=" + encoded}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		want := []byte{0xca, 0xfe, 0xba, 0xbe}
		if !bytes.Equal(dest, want) {
			t.Errorf("got %x, want %x", dest, want)
		}
	})

	t.Run("invalid base64", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.BytesBase64Var(new([]byte), "data", nil, "")

		if err := fs.Parse([]string{"--data=!!!not-base64!!!"}); err == nil {
			t.Error("expected error for invalid base64")
		}
	})

	t.Run("empty base64", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var dest []byte
		fs.BytesBase64Var(&dest, "data", []byte{0xff}, "")

		if err := fs.Parse([]string{"--data="}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if len(dest) != 0 {
			t.Errorf("got %x, want empty", dest)
		}
	})

	t.Run("shorthand", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		var dest []byte
		fs.BytesBase64VarP(&dest, "data", "d", nil, "")

		encoded := base64.StdEncoding.EncodeToString([]byte{0xff, 0x00})
		if err := fs.Parse([]string{"-d", encoded}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		want := []byte{0xff, 0x00}
		if !bytes.Equal(dest, want) {
			t.Errorf("got %x, want %x", dest, want)
		}
	})
}

// TestBytesBase64Type verifies Type() returns "bytesBase64".
// **Validates: Requirements 3.5**
func TestBytesBase64Type(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.BytesBase64Var(new([]byte), "data", nil, "")

	flag := fs.Lookup("data")
	if flag.Value.Type() != "bytesBase64" {
		t.Errorf("Type() = %q, want %q", flag.Value.Type(), "bytesBase64")
	}
}

// TestGetBytesHex tests the GetBytesHex getter with valid, missing, and wrong-type flags.
// **Validates: Requirements 2.6**
func TestGetBytesHex(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		defaultVal := []byte{0xde, 0xad}
		fs.BytesHexVar(new([]byte), "data", defaultVal, "")

		got, err := fs.GetBytesHex("data")
		if err != nil {
			t.Fatalf("GetBytesHex error: %v", err)
		}
		if !bytes.Equal(got, defaultVal) {
			t.Errorf("got %x, want %x", got, defaultVal)
		}
	})

	t.Run("after parse", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.BytesHexVar(new([]byte), "data", nil, "")

		if err := fs.Parse([]string{"--data=cafebabe"}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		got, err := fs.GetBytesHex("data")
		if err != nil {
			t.Fatalf("GetBytesHex error: %v", err)
		}
		want := []byte{0xca, 0xfe, 0xba, 0xbe}
		if !bytes.Equal(got, want) {
			t.Errorf("got %x, want %x", got, want)
		}
	})

	t.Run("missing flag", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		if _, err := fs.GetBytesHex("nope"); err == nil {
			t.Error("expected error for missing flag")
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.StringVar(new(string), "s", "", "")
		if _, err := fs.GetBytesHex("s"); err == nil {
			t.Error("expected error for wrong type flag")
		}
	})
}

// TestGetBytesBase64 tests the GetBytesBase64 getter with valid, missing, and wrong-type flags.
// **Validates: Requirements 3.6**
func TestGetBytesBase64(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		defaultVal := []byte{0xde, 0xad}
		fs.BytesBase64Var(new([]byte), "data", defaultVal, "")

		got, err := fs.GetBytesBase64("data")
		if err != nil {
			t.Fatalf("GetBytesBase64 error: %v", err)
		}
		if !bytes.Equal(got, defaultVal) {
			t.Errorf("got %x, want %x", got, defaultVal)
		}
	})

	t.Run("after parse", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.BytesBase64Var(new([]byte), "data", nil, "")

		encoded := base64.StdEncoding.EncodeToString([]byte{0xca, 0xfe, 0xba, 0xbe})
		if err := fs.Parse([]string{"--data=" + encoded}); err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		got, err := fs.GetBytesBase64("data")
		if err != nil {
			t.Fatalf("GetBytesBase64 error: %v", err)
		}
		want := []byte{0xca, 0xfe, 0xba, 0xbe}
		if !bytes.Equal(got, want) {
			t.Errorf("got %x, want %x", got, want)
		}
	})

	t.Run("missing flag", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		if _, err := fs.GetBytesBase64("nope"); err == nil {
			t.Error("expected error for missing flag")
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.StringVar(new(string), "s", "", "")
		if _, err := fs.GetBytesBase64("s"); err == nil {
			t.Error("expected error for wrong type flag")
		}
	})
}

// TestBytesHexStringEncoding verifies String() returns lowercase hex.
// **Validates: Requirements 2.5**
func TestBytesHexStringEncoding(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.BytesHexVar(new([]byte), "data", []byte{0xCA, 0xFE}, "")

	flag := fs.Lookup("data")
	got := flag.Value.String()
	want := hex.EncodeToString([]byte{0xCA, 0xFE})
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// TestBytesBase64StringEncoding verifies String() returns standard base64.
// **Validates: Requirements 3.5**
func TestBytesBase64StringEncoding(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.BytesBase64Var(new([]byte), "data", []byte{0xCA, 0xFE}, "")

	flag := fs.Lookup("data")
	got := flag.Value.String()
	want := base64.StdEncoding.EncodeToString([]byte{0xCA, 0xFE})
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

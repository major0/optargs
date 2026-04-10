package optargs

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
	"testing/quick"
)

// TestPropertyBytesHexRoundTrip verifies that for any byte slice, encoding to
// hex via String() then decoding via Set() produces the original. And for any
// valid hex string, Set() then String() produces the original (lowercased).
// **Validates: Requirements 2.2, 2.4**
func TestPropertyBytesHexRoundTrip(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	// bytes → hex → bytes
	t.Run("encode_decode", func(t *testing.T) {
		f := func(data []byte) bool {
			v := NewBytesHexValue(data, nil)
			encoded := v.String()

			fresh := NewBytesHexValue(nil, nil)
			if err := fresh.Set(encoded); err != nil {
				return false
			}
			return fresh.String() == encoded
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	// valid hex string → bytes → hex string (lowercased)
	t.Run("decode_encode", func(t *testing.T) {
		f := func(data []byte) bool {
			// Generate a valid hex string from arbitrary bytes.
			hexStr := hex.EncodeToString(data)

			v := NewBytesHexValue(nil, nil)
			if err := v.Set(hexStr); err != nil {
				return false
			}
			return v.String() == strings.ToLower(hexStr)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// TestPropertyBytesHexRejectsInvalid verifies that for any string that is not
// valid hex (odd length or non-hex characters), Set() returns a non-nil error.
// **Validates: Requirements 2.3**
func TestPropertyBytesHexRejectsInvalid(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	f := func(s string) bool {
		// Only test strings that are actually invalid hex.
		if _, err := hex.DecodeString(s); err == nil {
			return true // valid hex — skip
		}
		v := NewBytesHexValue(nil, nil)
		return v.Set(s) != nil
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

// TestPropertyBytesBase64RoundTrip verifies that for any byte slice, encoding
// to base64 via String() then decoding via Set() produces the original. And
// for any valid base64 string, Set() then String() produces the original.
// **Validates: Requirements 3.2, 3.4**
func TestPropertyBytesBase64RoundTrip(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	// bytes → base64 → bytes
	t.Run("encode_decode", func(t *testing.T) {
		f := func(data []byte) bool {
			v := NewBytesBase64Value(data, nil)
			encoded := v.String()

			fresh := NewBytesBase64Value(nil, nil)
			if err := fresh.Set(encoded); err != nil {
				return false
			}
			return fresh.String() == encoded
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	// valid base64 string → bytes → base64 string
	t.Run("decode_encode", func(t *testing.T) {
		f := func(data []byte) bool {
			// Generate a valid base64 string from arbitrary bytes.
			b64Str := base64.StdEncoding.EncodeToString(data)

			v := NewBytesBase64Value(nil, nil)
			if err := v.Set(b64Str); err != nil {
				return false
			}
			return v.String() == b64Str
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// TestPropertyBytesBase64RejectsInvalid verifies that for any string that is
// not valid standard base64, Set() returns a non-nil error.
// **Validates: Requirements 3.3**
func TestPropertyBytesBase64RejectsInvalid(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	f := func(s string) bool {
		// Only test strings that are actually invalid base64.
		if _, err := base64.StdEncoding.DecodeString(s); err == nil {
			return true // valid base64 — skip
		}
		v := NewBytesBase64Value(nil, nil)
		return v.Set(s) != nil
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

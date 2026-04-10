package pflag

import (
	"fmt"
	"net"
	"testing"
	"testing/quick"
)

// --- Property 6: ParseIPv4Mask round-trip ---

// TestPropertyParseIPv4MaskRoundTrip tests that for any 4-byte value,
// formatting it as a dotted-quad string and calling ParseIPv4Mask()
// returns a net.IPMask equal to the original 4 bytes.
// **Validates: Requirements 7.2, 7.3**
func TestPropertyParseIPv4MaskRoundTrip(t *testing.T) {
	roundTripProperty := func(a, b, c, d byte) bool {
		ip := net.IPv4(a, b, c, d)
		s := ip.To4().String()

		mask, err := ParseIPv4Mask(s)
		if err != nil {
			t.Logf("ParseIPv4Mask(%q) returned error: %v", s, err)
			return false
		}

		expected := net.IPMask{a, b, c, d}
		if len(mask) != 4 {
			t.Logf("ParseIPv4Mask(%q) returned %d bytes, want 4", s, len(mask))
			return false
		}
		for i := range expected {
			if mask[i] != expected[i] {
				t.Logf("ParseIPv4Mask(%q) = %v, want %v", s, mask, expected)
				return false
			}
		}
		return true
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(roundTripProperty, cfg); err != nil {
		t.Errorf("ParseIPv4Mask round-trip property failed: %v", err)
	}
}

// TestPropertyParseIPv4MaskRejectsInvalid tests that for any string that
// is not a valid IPv4 dotted-quad, ParseIPv4Mask() returns a non-nil error.
// **Validates: Requirements 7.3**
func TestPropertyParseIPv4MaskRejectsInvalid(t *testing.T) {
	rejectsInvalidProperty := func(s string) bool {
		// Filter out strings that happen to be valid IPv4 addresses.
		ip := net.ParseIP(s)
		if ip != nil && ip.To4() != nil {
			return true // valid IPv4 — skip
		}

		_, err := ParseIPv4Mask(s)
		return err != nil
	}

	cfg := &quick.Config{MaxCount: 100}
	if err := quick.Check(rejectsInvalidProperty, cfg); err != nil {
		t.Errorf("ParseIPv4Mask rejects-invalid property failed: %v", err)
	}
}

// --- Unit tests for ParseErrorsWhitelist and Getter ---

// TestParseErrorsWhitelistAlias verifies that ParseErrorsWhitelist is
// assignable to/from ParseErrorsAllowlist at runtime.
// **Validates: Requirements 6.1, 6.2**
func TestParseErrorsWhitelistAlias(t *testing.T) {
	// Assign Allowlist → Whitelist
	var allowlist ParseErrorsAllowlist
	allowlist.UnknownFlags = true

	whitelist := allowlist
	if !whitelist.UnknownFlags {
		t.Error("ParseErrorsWhitelist should have UnknownFlags=true after assignment from Allowlist")
	}

	// Assign Whitelist → Allowlist
	whitelist.UnknownFlags = false
	allowlist2 := whitelist
	if allowlist2.UnknownFlags {
		t.Error("ParseErrorsAllowlist should have UnknownFlags=false after assignment from Whitelist")
	}

	// Use in FlagSet
	fs := NewFlagSet("test", ContinueOnError)
	fs.ParseErrorsAllowlist = ParseErrorsWhitelist{UnknownFlags: true}
	if !fs.ParseErrorsAllowlist.UnknownFlags {
		t.Error("FlagSet.ParseErrorsAllowlist should accept ParseErrorsWhitelist value")
	}
}

// getterValue is a test type that implements both Value and Getter.
type getterValue struct {
	val string
}

func (g *getterValue) String() string     { return g.val }
func (g *getterValue) Set(s string) error { g.val = s; return nil }
func (g *getterValue) Type() string       { return "getter" }
func (g *getterValue) Get() any           { return g.val }

// TestGetterInterfaceSatisfied verifies that a type implementing Value + Get()
// satisfies the pflag.Getter interface.
// **Validates: Requirements 5.1**
func TestGetterInterfaceSatisfied(t *testing.T) {
	gv := &getterValue{val: "hello"}

	// Compile-time + runtime check: assign to Getter interface.
	var g Getter = gv
	if g.Get() != "hello" {
		t.Errorf("Getter.Get() = %v, want %q", g.Get(), "hello")
	}
	if g.String() != "hello" {
		t.Errorf("Getter.String() = %v, want %q", g.String(), "hello")
	}
	if err := g.Set("world"); err != nil {
		t.Errorf("Getter.Set() returned unexpected error: %v", err)
	}
	if g.Get() != "world" {
		t.Errorf("Getter.Get() after Set = %v, want %q", g.Get(), "world")
	}
	if g.Type() != "getter" {
		t.Errorf("Getter.Type() = %v, want %q", g.Type(), "getter")
	}

	// Verify it also satisfies Value.
	var v Value = gv
	_ = v
	_ = fmt.Sprintf("Getter also satisfies Value: %s", v.String())
}

// --- Unit tests for ParseIPv4Mask ---

// TestParseIPv4MaskValid tests valid dotted-quad strings produce correct 4-byte masks.
// **Validates: Requirements 7.1, 7.2**
func TestParseIPv4MaskValid(t *testing.T) {
	tests := []struct {
		input string
		want  net.IPMask
	}{
		{"255.255.255.0", net.IPMask{255, 255, 255, 0}},
		{"255.255.0.0", net.IPMask{255, 255, 0, 0}},
		{"255.0.0.0", net.IPMask{255, 0, 0, 0}},
		{"128.0.0.0", net.IPMask{128, 0, 0, 0}},
		{"192.168.1.1", net.IPMask{192, 168, 1, 1}},
		{"0.0.0.0", net.IPMask{0, 0, 0, 0}},
		{"255.255.255.255", net.IPMask{255, 255, 255, 255}},
		{"1.2.3.4", net.IPMask{1, 2, 3, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseIPv4Mask(tt.input)
			if err != nil {
				t.Fatalf("ParseIPv4Mask(%q) returned error: %v", tt.input, err)
			}
			if len(got) != 4 {
				t.Fatalf("ParseIPv4Mask(%q) returned %d bytes, want 4", tt.input, len(got))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("ParseIPv4Mask(%q) = %v, want %v", tt.input, got, tt.want)
					break
				}
			}
		})
	}
}

// TestParseIPv4MaskInvalid tests that invalid strings return errors.
// **Validates: Requirements 7.3**
func TestParseIPv4MaskInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"garbage", "not-an-ip"},
		{"partial", "255.255"},
		{"too_many_octets", "1.2.3.4.5"},
		{"letters", "abc.def.ghi.jkl"},
		{"negative", "-1.0.0.0"},
		{"overflow", "256.0.0.0"},
		{"spaces", "255 .255.255.0"},
		{"trailing_dot", "255.255.255.0."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseIPv4Mask(tt.input)
			if err == nil {
				t.Errorf("ParseIPv4Mask(%q) should return error for invalid input", tt.input)
			}
		})
	}
}

// TestParseIPv4MaskRejectsIPv6 tests that IPv6 addresses return errors.
// **Validates: Requirements 7.3**
func TestParseIPv4MaskRejectsIPv6(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"loopback", "::1"},
		{"full", "2001:db8::1"},
		{"all_zeros", "::"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseIPv4Mask(tt.input)
			if err == nil {
				t.Errorf("ParseIPv4Mask(%q) should return error for IPv6 address", tt.input)
			}
		})
	}
}

// TestParseIPv4MaskEdgeCases tests boundary values.
// **Validates: Requirements 7.1, 7.2**
func TestParseIPv4MaskEdgeCases(t *testing.T) {
	// All zeros
	mask, err := ParseIPv4Mask("0.0.0.0")
	if err != nil {
		t.Fatalf("ParseIPv4Mask(\"0.0.0.0\") error: %v", err)
	}
	if len(mask) != 4 || mask[0] != 0 || mask[1] != 0 || mask[2] != 0 || mask[3] != 0 {
		t.Errorf("ParseIPv4Mask(\"0.0.0.0\") = %v, want [0 0 0 0]", mask)
	}

	// All ones
	mask, err = ParseIPv4Mask("255.255.255.255")
	if err != nil {
		t.Fatalf("ParseIPv4Mask(\"255.255.255.255\") error: %v", err)
	}
	if len(mask) != 4 || mask[0] != 255 || mask[1] != 255 || mask[2] != 255 || mask[3] != 255 {
		t.Errorf("ParseIPv4Mask(\"255.255.255.255\") = %v, want [255 255 255 255]", mask)
	}
}

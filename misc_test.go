package optargs

import (
	"log/slog"
	"testing"
	"unicode"
)

// Generate tests for all 255 8-bit ANSI characters. We want the same
// behavior from our 'isGraph()' as the libc `isgraph()`.
func isGraphTests() []bool {
	tests := make([]bool, 255)
	for i := 0; i < 255; i++ {
		slog.Debug("isGraphTests", "i", i)
		tests[i] = unicode.IsGraphic(rune(i)) && !unicode.IsSpace(rune(i))
	}
	return tests
}

// Test our `IsGraph()` against all 255 8-bit ANSI characters
func TestIsGraph(t *testing.T) {
	for c, expect := range isGraphTests() {
		slog.Debug("TestIsGraph", "c", c, "expect", expect)
		if got := isGraph(byte(c)); got != expect {
			t.Errorf("isGraph(%q) = %v, want %v", byte(c), got, expect)
		}
	}
}

func FuzzIsGraph(f *testing.F) {
	f.Fuzz(func(t *testing.T, c byte) {
		isGraph(c)
	})
}

func BenchmarkIsGraphValid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isGraph('c')
	}
}

func BenchmarkIsGraphInvalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isGraph(byte(1))
	}
}

func BenchmarkIsGraphSpace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isGraph('\t')
	}
}

var hasPrefixTests = []struct {
	s, prefix  string
	ignoreCase bool
	want       bool
}{
	{s: "abc123", prefix: "a", ignoreCase: false, want: true},
	{s: "abc123", prefix: "ab", ignoreCase: false, want: true},
	{s: "abc123", prefix: "abc", ignoreCase: false, want: true},
	{s: "abc123", prefix: "A", ignoreCase: false, want: false},
	{s: "abc123", prefix: "aB", ignoreCase: false, want: false},
	{s: "abc123", prefix: "abC", ignoreCase: false, want: false},
	{s: "abc123", prefix: "A", ignoreCase: true, want: true},
	{s: "abc123", prefix: "aB", ignoreCase: true, want: true},
	{s: "abc123", prefix: "abC", ignoreCase: true, want: true},
	{s: "abc123", prefix: "a", ignoreCase: true, want: true},
	{s: "abc123", prefix: "ab", ignoreCase: true, want: true},
	{s: "abc123", prefix: "abc", ignoreCase: true, want: true},
}

func TestHasPrefix(t *testing.T) {
	for _, test := range hasPrefixTests {
		got := hasPrefix(test.s, test.prefix, test.ignoreCase)
		slog.Debug("TestHasPrefix", "string", test.s, "prefix", test.prefix, "ignoreCase", test.ignoreCase, "got", got)
		if got != test.want {
			t.Errorf("hasPrefix(%q, %q, %v) = %v, want %v", test.s, test.prefix, test.ignoreCase, got, test.want)
		}
	}
}

func BenchmarkHasPrefixMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hasPrefix("abc123ABC123Abc123", "abc123", false) // pragma: allowlist secret
	}
}

func BenchmarkHasPrefixMatchIgnoreCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hasPrefix("abc123ABC123Abc123", "ABC123", true) // pragma: allowlist secret
	}
}

func BenchmarkHasPrefixNoMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hasPrefix("abc123ABC123Abc123", "ABC123", false) // pragma: allowlist secret
	}
}

func BenchmarkHasPrefixNoMatchIgnoreCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hasPrefix("foo123ABC123Abc123", "bar123", false)
	}
}

func FuzzHasPrefix(f *testing.F) {
	f.Fuzz(func(t *testing.T, s string, prefix string, ignoreCase bool) {
		_ = hasPrefix(s, prefix, ignoreCase)
	})
}

var trimPrefixTests = []struct {
	s, prefix  string
	ignoreCase bool
	want       string
}{
	{s: "abc123", prefix: "a", ignoreCase: false, want: "bc123"},
	{s: "abc123", prefix: "ab", ignoreCase: false, want: "c123"},
	{s: "abc123", prefix: "abc", ignoreCase: false, want: "123"},
	{s: "abc123", prefix: "a", ignoreCase: true, want: "bc123"},
	{s: "abc123", prefix: "ab", ignoreCase: true, want: "c123"},
	{s: "abc123", prefix: "abc", ignoreCase: true, want: "123"},
	{s: "Abc123", prefix: "a", ignoreCase: false, want: "Abc123"},
	{s: "aBc123", prefix: "ab", ignoreCase: false, want: "aBc123"},
	{s: "abC123", prefix: "abc", ignoreCase: false, want: "abC123"},
	{s: "Abc123", prefix: "a", ignoreCase: true, want: "bc123"},
	{s: "aBc123", prefix: "ab", ignoreCase: true, want: "c123"},
	{s: "abC123", prefix: "abc", ignoreCase: true, want: "123"},
	{s: "", prefix: "a", ignoreCase: false, want: ""},
	{s: "", prefix: "a", ignoreCase: true, want: ""},
	{s: "a", prefix: "", ignoreCase: false, want: "a"},
	{s: "a", prefix: "", ignoreCase: true, want: "a"},
}

func TestTrimPrefix(t *testing.T) {
	for _, test := range trimPrefixTests {
		got := trimPrefix(test.s, test.prefix, test.ignoreCase)
		slog.Debug("TestTrimPrefix", "string", test.s, "prefix", test.prefix, "ignoreCase", test.ignoreCase, "got", got)
		if got != test.want {
			t.Errorf("trimPrefix(%q, %q, %v) = %q, want %q", test.s, test.prefix, test.ignoreCase, got, test.want)
		}
	}
}

func BenchmarkTrimPrefixMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trimPrefix("abc123ABC123Abc123", "abc123", false) // pragma: allowlist secret
	}
}

func BenchmarkTrimPrefixMatchIgnoreCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trimPrefix("abc123ABC123Abc123", "ABC123", true) // pragma: allowlist secret
	}
}

func BenchmarkTrimPrefixNoMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trimPrefix("abc123ABC123Abc123", "ABC123", false) // pragma: allowlist secret
	}
}

func BenchmarkTrimPrefixNoMatchIgnoreCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trimPrefix("foo123ABC123Abc123", "bar123", false)
	}
}

func FuzzTrimPrefix(f *testing.F) {
	f.Fuzz(func(t *testing.T, s string, prefix string, ignoreCase bool) {
		_ = trimPrefix(s, prefix, ignoreCase)
	})
}

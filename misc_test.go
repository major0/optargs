package optargs

import (
	"log/slog"
	"testing"
	"unicode"
)

// Generate tessts for all 255 8-bit ANSI characters. We want the same
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
	b          bool
}{
	{s: "abc123", prefix: "a", ignoreCase: false, b: true},
	{s: "abc123", prefix: "ab", ignoreCase: false, b: true},
	{s: "abc123", prefix: "abc", ignoreCase: false, b: true},
	{s: "abc123", prefix: "A", ignoreCase: false, b: false},
	{s: "abc123", prefix: "aB", ignoreCase: false, b: false},
	{s: "abc123", prefix: "abC", ignoreCase: false, b: false},
	{s: "abc123", prefix: "A", ignoreCase: true, b: true},
	{s: "abc123", prefix: "aB", ignoreCase: true, b: true},
	{s: "abc123", prefix: "abC", ignoreCase: true, b: true},
	{s: "abc123", prefix: "a", ignoreCase: true, b: true},
	{s: "abc123", prefix: "ab", ignoreCase: true, b: true},
	{s: "abc123", prefix: "abc", ignoreCase: true, b: true},
}

func TestHasPrefix(t *testing.T) {
	for _, test := range hasPrefixTests {
		got := hasPrefix(test.s, test.prefix, test.ignoreCase)
		slog.Debug("TestHasPrefix", "string", test.s, "prefix", test.prefix, "ignoreCase", test.ignoreCase, "got", got)
		if got != test.b {
			t.Errorf("hasPrefix(%q, %q, %v) = %v, want %v", test.s, test.prefix, test.ignoreCase, got, test.b)
		}
	}
}

func BenchmarkHasPrefixMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hasPrefix("abc123ABC123Abc123", "abc123", false)
	}
}

func BenchmarkHasPrefixMatchIgnoreCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hasPrefix("abc123ABC123Abc123", "ABC123", true)
	}
}

func BenchmarkHasPrefixNoMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hasPrefix("abc123ABC123Abc123", "ABC123", false)
	}
}

func BenchmarkHasPrefixNoMatchIgnoreCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hasPrefix("foo123ABC123Abc123", "bar123", false)
	}
}

func FuzzHasPrefix(f *testing.F) {
	f.Fuzz(func(t *testing.T, s string, prefix string, ignoreCase bool) {
		hasPrefix(s, prefix, ignoreCase)
	})
}

var trimPrefixTests = []struct {
	s, prefix  string
	ignoreCase bool
	b          string
}{
	{s: "abc123", prefix: "a", ignoreCase: false, b: "bc123"},
	{s: "abc123", prefix: "ab", ignoreCase: false, b: "c123"},
	{s: "abc123", prefix: "abc", ignoreCase: false, b: "123"},
	{s: "abc123", prefix: "a", ignoreCase: true, b: "bc123"},
	{s: "abc123", prefix: "ab", ignoreCase: true, b: "c123"},
	{s: "abc123", prefix: "abc", ignoreCase: true, b: "123"},
	{s: "Abc123", prefix: "a", ignoreCase: false, b: "Abc123"},
	{s: "aBc123", prefix: "ab", ignoreCase: false, b: "aBc123"},
	{s: "abC123", prefix: "abc", ignoreCase: false, b: "abC123"},
	{s: "Abc123", prefix: "a", ignoreCase: true, b: "bc123"},
	{s: "aBc123", prefix: "ab", ignoreCase: true, b: "c123"},
	{s: "abC123", prefix: "abc", ignoreCase: true, b: "123"},
}

func TestTrimPrefix(t *testing.T) {
	for _, test := range trimPrefixTests {
		got := trimPrefix(test.s, test.prefix, test.ignoreCase)
		slog.Debug("TestTrimPrefix", "string", test.s, "prefix", test.prefix, "ignoreCase", test.ignoreCase, "got", got)
		if got != test.b {
			t.Errorf("trimPrefix(%q, %q, %v) = %q, want %q", test.s, test.prefix, test.ignoreCase, got, test.b)
		}
	}
}

func TestTrimPrefixEmpty(t *testing.T) {
	s := ""
	prefix := "a"
	ignoreCase := false
	got := trimPrefix(s, prefix, ignoreCase)
	slog.Debug("TestTrimPrefixEmpty", "string", s, "prefix", prefix, "ignoreCase", ignoreCase, "got", got)
	if got != s {
		t.Errorf("trimPrefix(%q, %q, %v) = %q, want %q", s, prefix, ignoreCase, got, s)
	}
}

func TestTrimPrefixEmptyCaseIgnore(t *testing.T) {
	s := ""
	prefix := "a"
	ignoreCase := true
	got := trimPrefix(s, prefix, ignoreCase)
	slog.Debug("TestTrimPrefixEmpty", "string", s, "prefix", prefix, "ignoreCase", ignoreCase, "got", got)
	if got != s {
		t.Errorf("trimPrefix(%q, %q, %v) = %q, want %q", s, prefix, ignoreCase, got, s)
	}
}

func TestTrimPrefixEmptyPrefix(t *testing.T) {
	s := "a"
	prefix := ""
	ignoreCase := false
	got := trimPrefix(s, prefix, ignoreCase)
	slog.Debug("TestTrimPrefixEmptyPrefix", "string", s, "prefix", prefix, "ignoreCase", ignoreCase, "got", got)
	if got != s {
		t.Errorf("trimPrefix(%q, %q, %v) = %q, want %q", s, prefix, ignoreCase, got, s)
	}
}

func TestTrimPrefixEmptyPrefixCaseIgnore(t *testing.T) {
	s := "a"
	prefix := ""
	ignoreCase := true
	got := trimPrefix(s, prefix, ignoreCase)
	slog.Debug("TestTrimPrefixEmptyPrefix", "string", s, "prefix", prefix, "ignoreCase", ignoreCase, "got", got)
	if got != s {
		t.Errorf("trimPrefix(%q, %q, %v) = %q, want %q", s, prefix, ignoreCase, got, s)
	}
}

func BenchmarkTrimPrefixMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trimPrefix("abc123ABC123Abc123", "abc123", false)
	}
}

func BenchmarkTrimPrefixMatchIgnoreCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trimPrefix("abc123ABC123Abc123", "ABC123", true)
	}
}

func BenchmarkTrimPrefixNoMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trimPrefix("abc123ABC123Abc123", "ABC123", false)
	}
}

func BenchmarkTrimPrefixNoMatchIgnoreCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trimPrefix("foo123ABC123Abc123", "bar123", false)
	}
}

func FuzzTrimPrefix(f *testing.F) {
	f.Fuzz(func(t *testing.T, s string, prefix string, ignoreCase bool) {
		trimPrefix(s, prefix, ignoreCase)
	})
}

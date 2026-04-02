package optargs

import (
	"log/slog"
	"testing"
)

func FuzzIsGraph(f *testing.F) {
	f.Fuzz(func(t *testing.T, c byte) {
		isGraph(c)
	})
}

func BenchmarkIsGraphValid(b *testing.B) {
	for range b.N {
		isGraph('c')
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
	for range b.N {
		hasPrefix("abc123ABC123Abc123", "abc123", false) // pragma: allowlist secret
	}
}

func BenchmarkHasPrefixMatchIgnoreCase(b *testing.B) {
	for range b.N {
		hasPrefix("abc123ABC123Abc123", "ABC123", true) // pragma: allowlist secret
	}
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
	for range b.N {
		trimPrefix("abc123ABC123Abc123", "abc123", false) // pragma: allowlist secret
	}
}

func BenchmarkTrimPrefixMatchIgnoreCase(b *testing.B) {
	for range b.N {
		trimPrefix("abc123ABC123Abc123", "ABC123", true) // pragma: allowlist secret
	}
}

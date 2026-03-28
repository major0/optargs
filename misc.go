package optargs

import (
	"strings"
	"unicode"
)

// debug enables verbose slog.Debug logging in the parser hot paths.
// Kept false by default so the argument-evaluation cost of slog.Debug
// calls is completely eliminated in production. Set to true in tests
// or via an init() hook when troubleshooting.
var debug bool

// SetDebug enables or disables verbose debug logging in the parser.
func SetDebug(enabled bool) { debug = enabled }

// shortOptStrings is a pre-allocated lookup table mapping every printable
// ASCII byte to its single-character string representation. This avoids
// a heap allocation on every call to string(byte) in the short-option
// hot path.
var shortOptStrings [128]string

func init() {
	for i := range shortOptStrings {
		shortOptStrings[i] = string(rune(i))
	}
}

// byteString returns the single-character string for c without allocating.
func byteString(c byte) string {
	if c < 128 {
		return shortOptStrings[c]
	}
	return string(rune(c))
}

// Go's isGraph() behaves differently than the C version.
func isGraph(c byte) bool {
	r := rune(c)
	return !unicode.IsSpace(r) && unicode.IsPrint(r)
}

// hasPrefix checks whether s starts with prefix, optionally ignoring case.
func hasPrefix(s, prefix string, ignoreCase bool) bool {
	if len(s) < len(prefix) {
		return false
	}
	if ignoreCase {
		return strings.EqualFold(s[:len(prefix)], prefix)
	}
	return strings.HasPrefix(s, prefix)
}

// trimPrefix removes prefix from s, optionally ignoring case.
// The returned string preserves the original casing of s.
func trimPrefix(s, prefix string, ignoreCase bool) string {
	if len(prefix) == 0 || len(s) < len(prefix) {
		return s
	}
	if ignoreCase {
		if strings.EqualFold(s[:len(prefix)], prefix) {
			return s[len(prefix):]
		}
		return s
	}
	return strings.TrimPrefix(s, prefix)
}

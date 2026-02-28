package optargs

import (
	"strings"
	"unicode"
)

// Go's isGraph() behaves differently than the C version.
func isGraph(c byte) bool {
	r := rune(c)
	return !unicode.IsSpace(r) && unicode.IsPrint(r)
}

// hasPrefix checks whether s starts with prefix, optionally ignoring case.
func hasPrefix(s, prefix string, ignoreCase bool) bool {
	if ignoreCase {
		// Note: `strings.ToLower()` is expensive
		s = strings.ToLower(s)
		prefix = strings.ToLower(prefix)
	}

	return strings.HasPrefix(s, prefix)
}

// trimPrefix removes prefix from s, optionally ignoring case.
// The returned string preserves the original casing of s.
func trimPrefix(s, prefix string, ignoreCase bool) string {
	if ignoreCase {
		// Note: `strings.ToLower()` is expensive
		lower := strings.ToLower(s)
		prefix = strings.ToLower(prefix)
		if strings.HasPrefix(lower, prefix) {
			return s[len(prefix):]
		}
		return s
	}

	return strings.TrimPrefix(s, prefix)
}

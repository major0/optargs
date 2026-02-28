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

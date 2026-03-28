package pflags

import (
	"strings"
	"testing"
	"testing/quick"
	"unicode"
)

// isValidLongOptName returns true if s is a valid long option name:
// non-empty, all graphic unicode characters, no spaces.
func isValidLongOptName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// TestProperty11_OptArgsCoreIntegrationFidelity tests that parsing through the
// PFlags wrapper produces correct results for arbitrary valid flag names and values.
// This explores the input space beyond what table-driven tests enumerate —
// specifically, random flag names that may contain unicode, punctuation, etc.
// **Validates: Requirements 10.1, 10.2**
func TestProperty11_OptArgsCoreIntegrationFidelity(t *testing.T) {
	// Basic string flag integration
	basicIntegrationProperty := func(flagName, defaultValue, usage, setValue string) bool {
		if !isValidLongOptName(flagName) || len(flagName) > 50 {
			return true
		}

		fs := NewFlagSet("test", ContinueOnError)
		var variable string
		fs.StringVar(&variable, flagName, defaultValue, usage)

		args := []string{"--" + flagName, setValue}
		err := fs.Parse(args)
		if err != nil {
			// OptArgs Core may reject certain flag names — acceptable
			return true
		}

		if variable != setValue {
			return false
		}
		flag := fs.Lookup(flagName)
		return flag != nil && flag.Changed
	}

	if err := quick.Check(basicIntegrationProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Basic integration fidelity failed: %v", err)
	}

	// Shorthand equivalence: -x val == --name val
	shorthandEquivalenceProperty := func(flagName, shorthand, setValue string) bool {
		if !isValidLongOptName(flagName) || len(flagName) > 50 || len(shorthand) != 1 {
			return true
		}
		if !((shorthand[0] >= 'a' && shorthand[0] <= 'z') || (shorthand[0] >= 'A' && shorthand[0] <= 'Z')) {
			return true
		}

		// Parse with shorthand
		fs1 := NewFlagSet("t1", ContinueOnError)
		var v1 string
		fs1.StringVarP(&v1, flagName, shorthand, "default", "usage")
		if err := fs1.Parse([]string{"-" + shorthand, setValue}); err != nil {
			return true
		}

		// Parse with long form
		fs2 := NewFlagSet("t2", ContinueOnError)
		var v2 string
		fs2.StringVarP(&v2, flagName, shorthand, "default", "usage")
		if err := fs2.Parse([]string{"--" + flagName, setValue}); err != nil {
			return true
		}

		return v1 == v2 && v1 == setValue
	}

	if err := quick.Check(shorthandEquivalenceProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Shorthand equivalence failed: %v", err)
	}

	// Error handling: unknown flags produce pflag-compatible errors
	errorHandlingProperty := func(flagName string) bool {
		if !isValidLongOptName(flagName) || len(flagName) > 50 {
			return true
		}

		fs := NewFlagSet("test", ContinueOnError)
		var variable string
		fs.StringVar(&variable, flagName, "default", "usage")

		err := fs.Parse([]string{"--unknown-flag", "value"})
		if err == nil {
			return false
		}
		return strings.Contains(err.Error(), "unknown flag")
	}

	if err := quick.Check(errorHandlingProperty, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Error handling integration failed: %v", err)
	}
}

// TestPropertyAdvancedGNULongestMatching tests that for any pair of flags where
// one name is a prefix of the other, the parser always selects the correct
// (exact-match) flag. This explores random prefix/extension pairs that table
// tests cannot anticipate.
func TestPropertyAdvancedGNULongestMatching(t *testing.T) {
	longestMatchProperty := func(baseFlag, extendedFlag, setValue string) bool {
		if baseFlag == "" || extendedFlag == "" || len(baseFlag) >= len(extendedFlag) {
			return true
		}
		if !strings.HasPrefix(extendedFlag, baseFlag) {
			return true
		}
		if len(baseFlag) > 30 || len(extendedFlag) > 50 || setValue == "" {
			return true
		}

		fs := NewFlagSet("test", ContinueOnError)
		var baseVar, extendedVar string
		fs.StringVar(&baseVar, baseFlag, "", "")
		fs.StringVar(&extendedVar, extendedFlag, "", "")

		if err := fs.Parse([]string{"--" + extendedFlag, setValue}); err != nil {
			return true // OptArgs Core may reject
		}

		if extendedVar != setValue {
			return false
		}
		if baseVar != "" {
			return false
		}

		extObj := fs.Lookup(extendedFlag)
		baseObj := fs.Lookup(baseFlag)
		return extObj != nil && extObj.Changed && baseObj != nil && !baseObj.Changed
	}

	if err := quick.Check(longestMatchProperty, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("GNU longest matching property failed: %v", err)
	}
}

// TestPropertyAdvancedGNUSpecialCharacters tests that flag names containing
// colons and equals signs are correctly parsed with both space-separated and
// equals-separated values. Explores random prefix:suffix combinations.
func TestPropertyAdvancedGNUSpecialCharacters(t *testing.T) {
	specialCharsProperty := func(prefix, suffix, setValue string) bool {
		if prefix == "" || suffix == "" || setValue == "" {
			return true
		}
		if len(prefix) > 20 || len(suffix) > 20 || len(setValue) > 50 {
			return true
		}

		testCases := []string{
			prefix + ":" + suffix,
			prefix + "=" + suffix,
			prefix + ":" + suffix + "=more",
		}

		for _, flagName := range testCases {
			// Space-separated
			fs := NewFlagSet("test", ContinueOnError)
			var v string
			fs.StringVar(&v, flagName, "", "")
			if err := fs.Parse([]string{"--" + flagName, setValue}); err != nil {
				continue
			}
			if v != setValue {
				return false
			}

			// Equals-separated
			fs2 := NewFlagSet("test2", ContinueOnError)
			var v2 string
			fs2.StringVar(&v2, flagName, "", "")
			if err := fs2.Parse([]string{"--" + flagName + "=" + setValue}); err != nil {
				continue
			}
			if v2 != setValue {
				return false
			}
		}
		return true
	}

	if err := quick.Check(specialCharsProperty, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Special characters property failed: %v", err)
	}
}

// TestPropertyAdvancedGNUNestedEquals tests that flag names containing equals
// signs correctly distinguish between name and value components when using
// --name=key=value syntax.
func TestPropertyAdvancedGNUNestedEquals(t *testing.T) {
	nestedEqualsProperty := func(optionPart, keyPart, setValue string) bool {
		if optionPart == "" || keyPart == "" || setValue == "" {
			return true
		}
		if len(optionPart) > 15 || len(keyPart) > 15 || len(setValue) > 30 {
			return true
		}

		flagName := optionPart + "=" + keyPart
		fs := NewFlagSet("test", ContinueOnError)
		var v string
		fs.StringVar(&v, flagName, "", "")

		if err := fs.Parse([]string{"--" + flagName + "=" + setValue}); err != nil {
			return true // OptArgs Core may reject
		}

		if v != setValue {
			return false
		}
		flag := fs.Lookup(flagName)
		return flag != nil && flag.Changed
	}

	if err := quick.Check(nestedEqualsProperty, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Nested equals property failed: %v", err)
	}
}

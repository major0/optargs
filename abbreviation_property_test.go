package optargs

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
)

// ---------------------------------------------------------------------------
// Feature: longonly-abbreviation — Property-based tests (Task 8)
// ---------------------------------------------------------------------------

// randName generates a random lowercase ASCII name of length [minLen, maxLen].
func randName(rng *rand.Rand, minLen, maxLen int) string {
	n := minLen + rng.Intn(maxLen-minLen+1)
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a' + byte(rng.Intn(26)) //nolint:gosec // bounded [0,25], no overflow
	}
	return string(b)
}

// TestPropertyAbbrev1_ExactMatchPriority verifies that when a registered
// long option name equals the input exactly, it wins even when other
// registered names share the same prefix.
//
// **Validates: Requirements 1.2, 3.1, 3.2, 3.3**
func TestPropertyAbbrev1_ExactMatchPriority(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}
	f := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

		// Generate a base name (3-6 chars) that will be the exact match.
		base := randName(rng, 3, 6)

		// Create 2-4 extensions of the base name.
		numExtensions := 2 + rng.Intn(3)
		longOpts := map[string]*Flag{
			base: {Name: base, HasArg: NoArgument},
		}
		for range numExtensions {
			suffix := randName(rng, 1, 4)
			ext := base + suffix
			if _, exists := longOpts[ext]; exists {
				continue
			}
			longOpts[ext] = &Flag{Name: ext, HasArg: NoArgument}
		}

		p, err := NewParser(ParserConfig{}, nil, longOpts, []string{"--" + base})
		if err != nil {
			return false
		}

		opts := collectOpts(p)
		return len(opts) == 1 && opts[0].Name == base
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

// TestPropertyAbbrev2_AmbiguousPrefixReturnsError verifies that when 2+
// registered names share a common proper prefix and no exact match is
// registered for that prefix, the parser returns AmbiguousOptionError.
//
// **Validates: Requirements 2.1, 2.4**
func TestPropertyAbbrev2_AmbiguousPrefixReturnsError(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	t.Run("standard_mode", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

			// Generate a prefix (2-5 chars).
			prefix := randName(rng, 2, 5)

			// Create 2-4 distinct extensions.
			extensions := make(map[string]*Flag)
			for len(extensions) < 2 {
				suffix := randName(rng, 1, 4)
				name := prefix + suffix
				if _, exists := extensions[name]; exists {
					continue
				}
				extensions[name] = &Flag{Name: name, HasArg: NoArgument}
			}
			// Optionally add more.
			extra := rng.Intn(3)
			for range extra {
				suffix := randName(rng, 1, 4)
				name := prefix + suffix
				extensions[name] = &Flag{Name: name, HasArg: NoArgument}
			}

			p, err := NewParser(ParserConfig{}, nil, extensions, []string{"--" + prefix})
			if err != nil {
				return false
			}

			var gotErr error
			for _, err := range p.Options() {
				if err != nil {
					gotErr = err
					break
				}
			}
			var ambErr *AmbiguousOptionError
			return errors.As(gotErr, &ambErr)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("long_only_mode", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

			prefix := randName(rng, 2, 5)

			extensions := make(map[string]*Flag)
			for len(extensions) < 2 {
				suffix := randName(rng, 1, 4)
				name := prefix + suffix
				if _, exists := extensions[name]; exists {
					continue
				}
				extensions[name] = &Flag{Name: name, HasArg: NoArgument}
			}

			// Register a short option for the first char of prefix to
			// verify ambiguity is NOT rescued by short fallback.
			shortOpts := map[byte]*Flag{
				prefix[0]: {Name: string(prefix[0]), HasArg: NoArgument},
			}

			pcfg := ParserConfig{}
			pcfg.SetLongOnly(true)
			p, err := NewParser(pcfg, shortOpts, extensions, []string{"-" + prefix})
			if err != nil {
				return false
			}

			var gotErr error
			for _, err := range p.Options() {
				if err != nil {
					gotErr = err
					break
				}
			}
			var ambErr *AmbiguousOptionError
			return errors.As(gotErr, &ambErr)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// TestPropertyAbbrev3_UniqueAbbreviationResolves verifies that when exactly
// one registered name starts with the input prefix, the parser resolves it.
//
// **Validates: Requirements 4.1**
func TestPropertyAbbrev3_UniqueAbbreviationResolves(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}
	f := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

		// Generate a long name (4-8 chars).
		fullName := randName(rng, 4, 8)

		// Take a proper prefix (1 to len-1 chars).
		prefixLen := 1 + rng.Intn(len(fullName)-1)
		prefix := fullName[:prefixLen]

		// Add some unrelated names that do NOT share the prefix.
		longOpts := map[string]*Flag{
			fullName: {Name: fullName, HasArg: NoArgument},
		}
		for range 3 {
			other := randName(rng, 3, 7)
			// Ensure the other name doesn't start with our prefix.
			if strings.HasPrefix(other, prefix) {
				continue
			}
			longOpts[other] = &Flag{Name: other, HasArg: NoArgument}
		}

		p, err := NewParser(ParserConfig{}, nil, longOpts, []string{"--" + prefix})
		if err != nil {
			return false
		}

		opts := collectOpts(p)
		return len(opts) == 1 && opts[0].Name == fullName
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

// TestPropertyAbbrev4_EqualsSplitArgExtraction verifies that --{name}={value}
// round-trips correctly for RequiredArgument flags.
//
// **Validates: Requirements 1.4, 5.2, 6.3**
func TestPropertyAbbrev4_EqualsSplitArgExtraction(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}
	f := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

		// Generate a flag name (no '=' to keep this test focused on basic split).
		name := randName(rng, 2, 8)

		// Generate a non-empty argument value (lowercase letters, no '=').
		value := randName(rng, 1, 10)

		longOpts := map[string]*Flag{
			name: {Name: name, HasArg: RequiredArgument},
		}

		input := "--" + name + "=" + value
		p, err := NewParser(ParserConfig{}, nil, longOpts, []string{input})
		if err != nil {
			return false
		}

		opts := collectOpts(p)
		return len(opts) == 1 && opts[0].Name == name && opts[0].Arg == value && opts[0].HasArg
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

// TestPropertyAbbrev5_NoArgumentRejectsInlineArg verifies that NoArgument
// flags return UnexpectedArgumentError when given --{name}={value}.
//
// **Validates: Requirements 5.3**
func TestPropertyAbbrev5_NoArgumentRejectsInlineArg(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}
	f := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

		name := randName(rng, 2, 8)
		value := randName(rng, 1, 6)

		longOpts := map[string]*Flag{
			name: {Name: name, HasArg: NoArgument},
		}

		input := "--" + name + "=" + value
		p, err := NewParser(ParserConfig{}, nil, longOpts, []string{input})
		if err != nil {
			return false
		}

		var gotErr error
		for _, err := range p.Options() {
			if err != nil {
				gotErr = err
				break
			}
		}
		var unexpErr *UnexpectedArgumentError
		return errors.As(gotErr, &unexpErr)
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

// TestPropertyAbbrev6_RequiredArgumentConsumption verifies:
// (a) MissingArgumentError when no next arg exists
// (b) next-arg consumption when one does exist
//
// **Validates: Requirements 5.4, 5.5**
func TestPropertyAbbrev6_RequiredArgumentConsumption(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	t.Run("missing_arg", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests
			name := randName(rng, 2, 8)

			longOpts := map[string]*Flag{
				name: {Name: name, HasArg: RequiredArgument},
			}

			// No next argument provided.
			p, err := NewParser(ParserConfig{}, nil, longOpts, []string{"--" + name})
			if err != nil {
				return false
			}

			var gotErr error
			for _, err := range p.Options() {
				if err != nil {
					gotErr = err
					break
				}
			}
			var missErr *MissingArgumentError
			return errors.As(gotErr, &missErr)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("next_arg_consumed", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests
			name := randName(rng, 2, 8)
			value := randName(rng, 1, 8)

			longOpts := map[string]*Flag{
				name: {Name: name, HasArg: RequiredArgument},
			}

			p, err := NewParser(ParserConfig{}, nil, longOpts, []string{"--" + name, value})
			if err != nil {
				return false
			}

			opts := collectOpts(p)
			return len(opts) == 1 && opts[0].Name == name && opts[0].Arg == value && opts[0].HasArg
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// TestPropertyAbbrev7_OptionalArgumentSemantics verifies:
// (a) no next-arg consumption without =
// (b) inline arg used with = (including empty string from --opt=)
//
// **Validates: Requirements 5.6, 5.7, 5.8**
func TestPropertyAbbrev7_OptionalArgumentSemantics(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	t.Run("no_inline_no_consumption", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests
			name := randName(rng, 2, 8)

			longOpts := map[string]*Flag{
				name: {Name: name, HasArg: OptionalArgument},
			}

			// Provide a next arg that starts with '-' so it won't be consumed.
			p, err := NewParser(ParserConfig{}, nil, longOpts, []string{"--" + name, "--other"})
			if err != nil {
				return false
			}

			var gotOpt Option
			for opt, err := range p.Options() {
				if err != nil {
					// --other will be unknown, that's fine; we only care about our opt.
					continue
				}
				if opt.Name == name {
					gotOpt = opt
				}
			}
			// OptionalArgument without = should NOT consume next arg.
			return gotOpt.Name == name && !gotOpt.HasArg && gotOpt.Arg == ""
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("inline_arg_with_equals", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests
			name := randName(rng, 2, 8)
			value := randName(rng, 1, 6)

			longOpts := map[string]*Flag{
				name: {Name: name, HasArg: OptionalArgument},
			}

			input := "--" + name + "=" + value
			p, err := NewParser(ParserConfig{}, nil, longOpts, []string{input})
			if err != nil {
				return false
			}

			opts := collectOpts(p)
			return len(opts) == 1 && opts[0].Name == name && opts[0].Arg == value && opts[0].HasArg
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("empty_inline_arg", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests
			name := randName(rng, 2, 8)

			longOpts := map[string]*Flag{
				name: {Name: name, HasArg: OptionalArgument},
			}

			// --opt= should yield empty string arg.
			input := "--" + name + "="
			p, err := NewParser(ParserConfig{}, nil, longOpts, []string{input})
			if err != nil {
				return false
			}

			opts := collectOpts(p)
			return len(opts) == 1 && opts[0].Name == name && opts[0].Arg == "" && opts[0].HasArg
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// TestPropertyAbbrev8_UnmatchedInputReturnsError verifies that inputs
// which are not exact or prefix matches of any registered name produce
// UnknownOptionError.
//
// **Validates: Requirements 1.5, 9.1**
func TestPropertyAbbrev8_UnmatchedInputReturnsError(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}
	f := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

		// Register a few names all starting with 'x'.
		numOpts := 1 + rng.Intn(3)
		longOpts := make(map[string]*Flag)
		for range numOpts {
			name := "x" + randName(rng, 2, 5)
			longOpts[name] = &Flag{Name: name, HasArg: NoArgument}
		}

		// Generate an input starting with 'z' — guaranteed no match.
		unmatched := "z" + randName(rng, 2, 5)

		p, err := NewParser(ParserConfig{}, nil, longOpts, []string{"--" + unmatched})
		if err != nil {
			return false
		}

		var gotErr error
		for _, err := range p.Options() {
			if err != nil {
				gotErr = err
				break
			}
		}
		var unkErr *UnknownOptionError
		return errors.As(gotErr, &unkErr)
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

// TestPropertyAbbrev9_LongOnlyShortFallbackScope verifies:
// - single-dash zero-match falls back to short parsing (no error from long path)
// - double-dash zero-match returns UnknownOptionError
//
// **Validates: Requirements 7.1, 7.2**
func TestPropertyAbbrev9_LongOnlyShortFallbackScope(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	t.Run("single_dash_fallback", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

			// Pick a random letter for the short option.
			ch := byte('a' + rng.Intn(26)) //nolint:gosec // bounded [0,25], no overflow

			// Register only a long option starting with 'x' (won't match ch).
			longName := "x" + randName(rng, 2, 5)
			longOpts := map[string]*Flag{
				longName: {Name: longName, HasArg: NoArgument},
			}
			shortOpts := map[byte]*Flag{
				ch: {Name: string(ch), HasArg: NoArgument},
			}

			pcfg := ParserConfig{}
			pcfg.SetLongOnly(true)
			// Input: -<ch> — should fall back to short option.
			p, err := NewParser(pcfg, shortOpts, longOpts, []string{"-" + string(ch)})
			if err != nil {
				return false
			}

			opts := collectOpts(p)
			return len(opts) == 1 && opts[0].Name == string(ch)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("double_dash_no_fallback", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

			// Register long options starting with 'x' and short options.
			longName := "x" + randName(rng, 2, 5)
			longOpts := map[string]*Flag{
				longName: {Name: longName, HasArg: NoArgument},
			}

			// Pick a short option char that is NOT a prefix of any long name.
			// Use 'z' + random suffix for the unmatched input to guarantee
			// it doesn't prefix-match any 'x'-prefixed long option.
			unmatchedName := "z" + randName(rng, 1, 4)
			ch := byte('a' + rng.Intn(26)) //nolint:gosec // bounded [0,25], no overflow
			shortOpts := map[byte]*Flag{
				ch: {Name: string(ch), HasArg: NoArgument},
			}

			pcfg := ParserConfig{}
			pcfg.SetLongOnly(true)
			// Input: --<unmatched> — double-dash should NOT fall back to short.
			p, err := NewParser(pcfg, shortOpts, longOpts, []string{"--" + unmatchedName})
			if err != nil {
				return false
			}

			var gotErr error
			for _, err := range p.Options() {
				if err != nil {
					gotErr = err
					break
				}
			}
			var unkErr *UnknownOptionError
			return errors.As(gotErr, &unkErr)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// TestPropertyAbbrev10_SingleCharLongOnlyPrefersShort verifies that in
// long-only mode, a single-character input prefers the short option even
// when the character is a prefix of a registered long option name.
//
// **Validates: Requirements 8.1**
func TestPropertyAbbrev10_SingleCharLongOnlyPrefersShort(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}
	f := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

		// Pick a random letter.
		ch := byte('a' + rng.Intn(26)) //nolint:gosec // bounded [0,25], no overflow

		// Register a long option that starts with ch.
		longName := string(ch) + randName(rng, 2, 5)
		longOpts := map[string]*Flag{
			longName: {Name: longName, HasArg: NoArgument},
		}
		shortOpts := map[byte]*Flag{
			ch: {Name: string(ch), HasArg: NoArgument},
		}

		pcfg := ParserConfig{}
		pcfg.SetLongOnly(true)
		p, err := NewParser(pcfg, shortOpts, longOpts, []string{"-" + string(ch)})
		if err != nil {
			return false
		}

		opts := collectOpts(p)
		// Short option should win.
		return len(opts) == 1 && opts[0].Name == string(ch)
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Error(err)
	}
}

// TestPropertyAbbrev11_CaseSensitivityInPrefixMatching verifies that
// case-insensitive mode matches case variants and case-sensitive mode
// rejects them.
//
// **Validates: Requirements 10.1, 10.2**
func TestPropertyAbbrev11_CaseSensitivityInPrefixMatching(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	t.Run("case_insensitive_matches", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

			// Generate a lowercase name (4-8 chars).
			name := randName(rng, 4, 8)

			// Create an uppercase variant of a proper prefix.
			prefixLen := 2 + rng.Intn(len(name)-2)
			prefix := strings.ToUpper(name[:prefixLen])

			longOpts := map[string]*Flag{
				name: {Name: name, HasArg: NoArgument},
			}

			pcfg := ParserConfig{longCaseIgnore: true}
			p, err := NewParser(pcfg, nil, longOpts, []string{"--" + prefix})
			if err != nil {
				return false
			}

			opts := collectOpts(p)
			return len(opts) == 1 && opts[0].Name == name
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("case_sensitive_rejects", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

			// Generate a lowercase name (4-8 chars).
			name := randName(rng, 4, 8)

			// Create an uppercase variant of a proper prefix.
			prefixLen := 2 + rng.Intn(len(name)-2)
			prefix := strings.ToUpper(name[:prefixLen])

			// Skip if the prefix happens to be the same (all non-alpha).
			if strings.EqualFold(prefix, name[:prefixLen]) && prefix == name[:prefixLen] {
				return true // same case, skip
			}

			longOpts := map[string]*Flag{
				name: {Name: name, HasArg: NoArgument},
			}

			// Case-sensitive mode (default).
			pcfg := ParserConfig{longCaseIgnore: false}
			p, err := NewParser(pcfg, nil, longOpts, []string{"--" + prefix})
			if err != nil {
				return false
			}

			var gotErr error
			for _, err := range p.Options() {
				if err != nil {
					gotErr = err
					break
				}
			}
			// Should be UnknownOptionError since case doesn't match.
			var unkErr *UnknownOptionError
			return errors.As(gotErr, &unkErr)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// TestPropertyAbbrev12_ParentScopeInPrefixMatching verifies that prefix
// matching searches both child and parent parsers, and that cross-parser
// ambiguity is detected.
//
// **Validates: Requirements 11.1, 11.2**
func TestPropertyAbbrev12_ParentScopeInPrefixMatching(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	t.Run("parent_prefix_resolves", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

			// Parent has a long option; child has none sharing the prefix.
			parentName := randName(rng, 4, 8)
			prefixLen := 2 + rng.Intn(len(parentName)-2)
			prefix := parentName[:prefixLen]

			// Child has an unrelated option.
			childName := "z" + randName(rng, 3, 6)

			parent, err := NewParser(ParserConfig{}, nil,
				map[string]*Flag{parentName: {Name: parentName, HasArg: NoArgument}},
				nil)
			if err != nil {
				return false
			}
			child, err := NewParser(ParserConfig{}, nil,
				map[string]*Flag{childName: {Name: childName, HasArg: NoArgument}},
				[]string{"--" + prefix})
			if err != nil {
				return false
			}
			parent.AddCmd("sub", child)

			opts := collectOpts(child)
			return len(opts) == 1 && opts[0].Name == parentName
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("cross_parser_ambiguity", func(t *testing.T) {
		f := func(seed int64) bool {
			rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

			// Generate a shared prefix.
			prefix := randName(rng, 2, 4)

			// Parent has one extension, child has a different extension.
			parentName := prefix + randName(rng, 1, 4)
			childName := prefix + randName(rng, 1, 4)

			// Ensure they're actually different.
			if parentName == childName {
				return true // skip degenerate case
			}

			parent, err := NewParser(ParserConfig{}, nil,
				map[string]*Flag{parentName: {Name: parentName, HasArg: NoArgument}},
				nil)
			if err != nil {
				return false
			}
			child, err := NewParser(ParserConfig{}, nil,
				map[string]*Flag{childName: {Name: childName, HasArg: NoArgument}},
				[]string{"--" + prefix})
			if err != nil {
				return false
			}
			parent.AddCmd("sub", child)

			var gotErr error
			for _, err := range child.Options() {
				if err != nil {
					gotErr = err
					break
				}
			}
			var ambErr *AmbiguousOptionError
			return errors.As(gotErr, &ambErr)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

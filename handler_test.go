package optargs

import (
	"fmt"
	"math/rand"
	"testing"
	"testing/quick"
)

// validShortChars contains characters valid for short options in optstrings.
// Excludes ':', ';', '-' (prohibited) and 'W' (reserved for gnuWords).
var validShortChars = func() []byte {
	var chars []byte
	for c := byte('!'); c <= byte('~'); c++ {
		switch c {
		case ':', ';', '-', 'W':
			continue
		}
		chars = append(chars, c)
	}
	return chars
}()

// validLongNames is a pool of long option names for random generation.
var validLongNames = []string{
	"verbose", "output", "debug", "help", "version", "force",
	"quiet", "recursive", "all", "long", "human", "color",
	"sort", "reverse", "size", "time", "group", "author",
}

// collectOptions iterates a parser and returns the sequence of (Option, error) pairs.
func collectOptions(p *Parser) ([]Option, []error) {
	var opts []Option
	var errs []error
	for opt, err := range p.Options() {
		opts = append(opts, opt)
		errs = append(errs, err)
	}
	return opts, errs
}

// handlerOptionsEqual returns true if two Option slices are identical.
func handlerOptionsEqual(a, b []Option) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].HasArg != b[i].HasArg || a[i].Arg != b[i].Arg {
			return false
		}
	}
	return true
}

// errorsEqual returns true if two error slices have the same nil/non-nil
// pattern and matching messages.
func errorsEqual(a, b []error) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if (a[i] == nil) != (b[i] == nil) {
			return false
		}
		if a[i] != nil && a[i].Error() != b[i].Error() {
			return false
		}
	}
	return true
}

// Feature: option-handlers, Property 1: Nil Handle backward compatibility
// For any parser configuration and argument list where all Flags have nil
// Handle, iterator output is identical to pre-handler implementation.
// Validates: Requirements 1.3, 4.1, 4.4
func TestPropertyNilHandleBackwardCompat(t *testing.T) {
	// Use only lowercase letters to avoid optstring prefix flags (+, :, -)
	// and special characters that complicate config equivalence.
	safeChars := []byte("abcdefghijklmnopqrstuvxyz") // exclude 'W' (gnuWords)

	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Generate a random optstring with 1–6 short options.
		nShort := 1 + rng.Intn(6)
		perm := rng.Perm(len(safeChars))
		var optstring string
		shortChars := make([]byte, nShort)
		argTypes := make([]ArgType, nShort)
		for i := 0; i < nShort; i++ {
			c := safeChars[perm[i]]
			shortChars[i] = c
			optstring += string(c)
			at := ArgType(rng.Intn(3))
			argTypes[i] = at
			switch at {
			case RequiredArgument:
				optstring += ":"
			case OptionalArgument:
				optstring += "::"
			}
		}

		// Generate 0–4 random long options.
		nLong := rng.Intn(5)
		longPerm := rng.Perm(len(validLongNames))
		var longFlags []Flag
		for i := 0; i < nLong && i < len(validLongNames); i++ {
			longFlags = append(longFlags, Flag{
				Name:   validLongNames[longPerm[i]],
				HasArg: ArgType(rng.Intn(3)),
			})
		}

		// Build a random argument list using the generated options.
		nArgs := rng.Intn(8)
		var args []string
		for i := 0; i < nArgs; i++ {
			switch rng.Intn(3) {
			case 0: // short option
				idx := rng.Intn(nShort)
				args = append(args, "-"+string(shortChars[idx]))
				if argTypes[idx] == RequiredArgument {
					args = append(args, "val")
				}
			case 1: // long option (if any)
				if nLong > 0 {
					lf := longFlags[rng.Intn(nLong)]
					args = append(args, "--"+lf.Name)
					if lf.HasArg == RequiredArgument {
						args = append(args, "val")
					}
				}
			case 2: // non-option argument
				args = append(args, "nonopt")
			}
		}

		// Parse with GetOptLong (standard constructor — nil Handle).
		args1 := make([]string, len(args))
		copy(args1, args)
		p1, err := GetOptLong(args1, optstring, longFlags)
		if err != nil {
			return true // invalid config, skip
		}
		opts1, errs1 := collectOptions(p1)
		remaining1 := p1.Args

		// Build an equivalent parser via NewParser with explicit nil Handle.
		shortMap := make(map[byte]*Flag)
		for i := 0; i < nShort; i++ {
			shortMap[shortChars[i]] = &Flag{
				Name:   string(shortChars[i]),
				HasArg: argTypes[i],
				Handle: nil, // explicitly nil
			}
		}
		longMap := make(map[string]*Flag)
		for _, lf := range longFlags {
			longMap[lf.Name] = &Flag{
				Name:   lf.Name,
				HasArg: lf.HasArg,
				Handle: nil, // explicitly nil
			}
		}

		args2 := make([]string, len(args))
		copy(args2, args)
		config := ParserConfig{
			enableErrors:    true,
			longCaseIgnore:  true,
			shortCaseIgnore: false,
			longOptsOnly:    false,
			gnuWords:        false,
			parseMode:       ParseDefault,
		}
		p2, err := NewParser(config, shortMap, longMap, args2)
		if err != nil {
			return true // skip
		}
		opts2, errs2 := collectOptions(p2)
		remaining2 := p2.Args

		// Both parsers must produce identical output.
		if !handlerOptionsEqual(opts1, opts2) {
			t.Logf("seed=%d optstring=%q args=%v", seed, optstring, args)
			t.Logf("opts1=%+v", opts1)
			t.Logf("opts2=%+v", opts2)
			return false
		}
		if !errorsEqual(errs1, errs2) {
			t.Logf("seed=%d errs differ", seed)
			return false
		}
		if len(remaining1) != len(remaining2) {
			t.Logf("seed=%d remaining args differ: %v vs %v", seed, remaining1, remaining2)
			return false
		}
		for i := range remaining1 {
			if remaining1[i] != remaining2[i] {
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 1 (Nil Handle backward compatibility) failed: %v", err)
	}
}

// Feature: option-handlers, Property 3: Handler invocation suppresses Option yield
// For any Flag with non-nil Handle returning nil, handler is invoked and
// no Option is yielded through the iterator for that flag.
// Validates: Requirements 1.4, 2.5
func TestPropertyHandlerSuppressesYield(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Pick 2–5 short options.
		nShort := 2 + rng.Intn(4)
		perm := rng.Perm(len(validShortChars))
		chars := make([]byte, nShort)
		for i := range chars {
			chars[i] = validShortChars[perm[i]]
		}

		// Randomly decide which options get handlers (at least one handled,
		// at least one not handled).
		handled := make([]bool, nShort)
		for i := range handled {
			handled[i] = rng.Intn(2) == 0
		}
		// Ensure at least one of each.
		handled[0] = true
		handled[1] = false

		// Track handler invocations.
		type call struct{ name, arg string }
		var calls []call

		shortMap := make(map[byte]*Flag)
		for i, c := range chars {
			f := &Flag{Name: string(c), HasArg: NoArgument}
			if handled[i] {
				f.Handle = func(name string, arg string) error {
					calls = append(calls, call{name, arg})
					return nil
				}
			}
			shortMap[c] = f
		}

		// Build args: one instance of each option.
		var args []string
		for _, c := range chars {
			args = append(args, "-"+string(c))
		}

		p, err := NewParser(ParserConfig{enableErrors: true}, shortMap, nil, args)
		if err != nil {
			return true
		}

		var yielded []Option
		for opt, err := range p.Options() {
			if err != nil {
				t.Logf("seed=%d unexpected error: %v", seed, err)
				return false
			}
			yielded = append(yielded, opt)
		}

		// Count expected handlers and yields.
		var expectHandled, expectYielded int
		for _, h := range handled {
			if h {
				expectHandled++
			} else {
				expectYielded++
			}
		}

		if len(calls) != expectHandled {
			t.Logf("seed=%d handler calls: got %d, want %d", seed, len(calls), expectHandled)
			return false
		}
		if len(yielded) != expectYielded {
			t.Logf("seed=%d yielded options: got %d, want %d", seed, len(yielded), expectYielded)
			return false
		}

		// Verify no yielded option has a name that belongs to a handled flag.
		handledNames := make(map[string]bool)
		for i, c := range chars {
			if handled[i] {
				handledNames[string(c)] = true
			}
		}
		for _, opt := range yielded {
			if handledNames[opt.Name] {
				t.Logf("seed=%d option %q was yielded but should have been handled", seed, opt.Name)
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (Handler suppresses yield) failed: %v", err)
	}
}

// Feature: option-handlers, Property 4: Handler receives correct name and arg
// Handler receives the same name and arg that would have been yielded as
// Option.Name and Option.Arg by the iterator if Handle were nil.
// Validates: Requirements 2.1, 2.2
func TestPropertyHandlerReceivesCorrectNameAndArg(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Generate 1–4 short options with varying arg types.
		nShort := 1 + rng.Intn(4)
		perm := rng.Perm(len(validShortChars))
		chars := make([]byte, nShort)
		argKinds := make([]ArgType, nShort)
		for i := range chars {
			chars[i] = validShortChars[perm[i]]
			argKinds[i] = ArgType(rng.Intn(3))
		}

		// Generate 0–3 long options.
		nLong := rng.Intn(4)
		longPerm := rng.Perm(len(validLongNames))
		type longDef struct {
			name   string
			hasArg ArgType
		}
		var longs []longDef
		for i := 0; i < nLong && i < len(validLongNames); i++ {
			longs = append(longs, longDef{
				name:   validLongNames[longPerm[i]],
				hasArg: ArgType(rng.Intn(3)),
			})
		}

		// Build args that exercise the generated options.
		var args []string
		for i, c := range chars {
			args = append(args, "-"+string(c))
			if argKinds[i] == RequiredArgument {
				args = append(args, "sval")
			}
		}
		for _, l := range longs {
			args = append(args, "--"+l.name)
			if l.hasArg == RequiredArgument {
				args = append(args, "lval")
			}
		}

		// Pass 1: parse with nil Handle, collect yielded Options.
		buildShortMap := func(handler func(string, string) error) map[byte]*Flag {
			m := make(map[byte]*Flag)
			for i, c := range chars {
				m[c] = &Flag{Name: string(c), HasArg: argKinds[i], Handle: handler}
			}
			return m
		}
		buildLongMap := func(handler func(string, string) error) map[string]*Flag {
			m := make(map[string]*Flag)
			for _, l := range longs {
				m[l.name] = &Flag{Name: l.name, HasArg: l.hasArg, Handle: handler}
			}
			return m
		}

		cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}

		args1 := make([]string, len(args))
		copy(args1, args)
		p1, err := NewParser(cfg, buildShortMap(nil), buildLongMap(nil), args1)
		if err != nil {
			return true
		}
		var expected []Option
		for opt, err := range p1.Options() {
			if err != nil {
				return true // skip configs that produce errors
			}
			expected = append(expected, opt)
		}

		// Pass 2: parse with Handle on all flags, record calls.
		type call struct{ name, arg string }
		var got []call
		handler := func(name string, arg string) error {
			got = append(got, call{name, arg})
			return nil
		}

		args2 := make([]string, len(args))
		copy(args2, args)
		p2, err := NewParser(cfg, buildShortMap(handler), buildLongMap(handler), args2)
		if err != nil {
			return true
		}
		// With all flags handled, iterator should yield nothing.
		for _, err := range p2.Options() {
			if err != nil {
				return true
			}
		}

		if len(got) != len(expected) {
			t.Logf("seed=%d call count: got %d, want %d", seed, len(got), len(expected))
			return false
		}
		for i := range expected {
			if got[i].name != expected[i].Name || got[i].arg != expected[i].Arg {
				t.Logf("seed=%d [%d] handler got (%q,%q), want (%q,%q)",
					seed, i, got[i].name, got[i].arg, expected[i].Name, expected[i].Arg)
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 4 (Handler receives correct name and arg) failed: %v", err)
	}
}

// Feature: option-handlers, Property 5: Handler error propagation
// Handler returning non-nil error causes iterator to yield that error
// paired with a zero-value Option.
// Validates: Requirements 2.4, 9.1, 9.3
func TestPropertyHandlerErrorPropagation(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Pick 2–4 short options; one will have an error-returning handler.
		nShort := 2 + rng.Intn(3)
		perm := rng.Perm(len(validShortChars))
		chars := make([]byte, nShort)
		for i := range chars {
			chars[i] = validShortChars[perm[i]]
		}

		// Pick which option returns an error.
		errIdx := rng.Intn(nShort)
		errMsg := "handler error from " + string(chars[errIdx])
		sentinel := fmt.Errorf("%s", errMsg)

		shortMap := make(map[byte]*Flag)
		for i, c := range chars {
			f := &Flag{Name: string(c), HasArg: NoArgument}
			if i == errIdx {
				f.Handle = func(name string, arg string) error {
					return sentinel
				}
			}
			shortMap[c] = f
		}

		// Build args: one of each option, error option first so we can
		// verify the error yield clearly.
		var args []string
		args = append(args, "-"+string(chars[errIdx]))
		for i, c := range chars {
			if i != errIdx {
				args = append(args, "-"+string(c))
			}
		}

		p, err := NewParser(ParserConfig{enableErrors: true}, shortMap, nil, args)
		if err != nil {
			return true
		}

		// The first yield should be the handler error with zero-value Option.
		first := true
		var sawError bool
		for opt, err := range p.Options() {
			if first {
				first = false
				if err == nil {
					t.Logf("seed=%d expected error on first yield, got nil", seed)
					return false
				}
				if err.Error() != errMsg {
					t.Logf("seed=%d error mismatch: got %q, want %q", seed, err.Error(), errMsg)
					return false
				}
				// Option must be zero-value.
				if opt != (Option{}) {
					t.Logf("seed=%d expected zero Option, got %+v", seed, opt)
					return false
				}
				sawError = true
				continue
			}
			// Remaining options should yield normally (no handler on them).
			if err != nil {
				t.Logf("seed=%d unexpected error after handler error: %v", seed, err)
				return false
			}
		}

		if !sawError {
			t.Logf("seed=%d never saw handler error", seed)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5 (Handler error propagation) failed: %v", err)
	}
}

// Feature: option-handlers, Property 8: Mixed handled and non-handled dispatch
// Parser with both handled and non-handled Flags invokes handlers for
// handled Flags and yields Options for non-handled Flags within the same
// iteration, with no cross-contamination.
// Validates: Requirements 7.1
func TestPropertyMixedHandledAndNonHandled(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Generate 2–5 short options and 1–3 long options.
		nShort := 2 + rng.Intn(4)
		nLong := 1 + rng.Intn(3)
		shortPerm := rng.Perm(len(validShortChars))
		longPerm := rng.Perm(len(validLongNames))

		chars := make([]byte, nShort)
		for i := range chars {
			chars[i] = validShortChars[shortPerm[i]]
		}

		type optDef struct {
			name    string
			isShort bool
			handled bool
		}
		var defs []optDef

		// Assign handled/non-handled randomly, ensuring at least one of each.
		shortMap := make(map[byte]*Flag)
		type call struct{ name, arg string }
		var calls []call

		handler := func(name string, arg string) error {
			calls = append(calls, call{name, arg})
			return nil
		}

		for i, c := range chars {
			h := rng.Intn(2) == 0
			if i == 0 {
				h = true // ensure at least one handled
			}
			if i == 1 {
				h = false // ensure at least one non-handled
			}
			f := &Flag{Name: string(c), HasArg: NoArgument}
			if h {
				f.Handle = handler
			}
			shortMap[c] = f
			defs = append(defs, optDef{name: string(c), isShort: true, handled: h})
		}

		longMap := make(map[string]*Flag)
		for i := 0; i < nLong && i < len(validLongNames); i++ {
			name := validLongNames[longPerm[i]]
			h := rng.Intn(2) == 0
			f := &Flag{Name: name, HasArg: NoArgument}
			if h {
				f.Handle = handler
			}
			longMap[name] = f
			defs = append(defs, optDef{name: name, isShort: false, handled: h})
		}

		// Build args in definition order.
		var args []string
		for _, d := range defs {
			if d.isShort {
				args = append(args, "-"+d.name)
			} else {
				args = append(args, "--"+d.name)
			}
		}

		cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
		p, err := NewParser(cfg, shortMap, longMap, args)
		if err != nil {
			return true
		}

		calls = nil // reset
		var yielded []Option
		for opt, err := range p.Options() {
			if err != nil {
				t.Logf("seed=%d unexpected error: %v", seed, err)
				return false
			}
			yielded = append(yielded, opt)
		}

		// Build expected sets.
		handledNames := make(map[string]bool)
		nonHandledNames := make(map[string]bool)
		for _, d := range defs {
			if d.handled {
				handledNames[d.name] = true
			} else {
				nonHandledNames[d.name] = true
			}
		}

		// Every handler call must be for a handled flag.
		for _, c := range calls {
			if !handledNames[c.name] {
				t.Logf("seed=%d handler called for non-handled flag %q", seed, c.name)
				return false
			}
		}

		// Every yielded option must be for a non-handled flag.
		for _, opt := range yielded {
			if handledNames[opt.Name] {
				t.Logf("seed=%d option %q yielded but should have been handled", seed, opt.Name)
				return false
			}
			if !nonHandledNames[opt.Name] {
				t.Logf("seed=%d option %q yielded but not in definitions", seed, opt.Name)
				return false
			}
		}

		// Counts must match.
		if len(calls) != len(handledNames) {
			t.Logf("seed=%d handler calls: got %d, want %d", seed, len(calls), len(handledNames))
			return false
		}
		if len(yielded) != len(nonHandledNames) {
			t.Logf("seed=%d yielded: got %d, want %d", seed, len(yielded), len(nonHandledNames))
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8 (Mixed handled and non-handled dispatch) failed: %v", err)
	}
}

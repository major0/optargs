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

// handlerCall records a handler invocation for test assertions.
type handlerCall struct{ name, arg string }

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
		if !optionsEqual(opts1, opts2) {
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
		var calls []handlerCall

		shortMap := make(map[byte]*Flag)
		for i, c := range chars {
			f := &Flag{Name: string(c), HasArg: NoArgument}
			if handled[i] {
				f.Handle = func(name string, arg string) error {
					calls = append(calls, handlerCall{name, arg})
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
		var got []handlerCall
		handler := func(name string, arg string) error {
			got = append(got, handlerCall{name, arg})
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
		var calls []handlerCall

		handler := func(name string, arg string) error {
			calls = append(calls, handlerCall{name, arg})
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

// Feature: option-handlers, Property 9: SetShortHandler/SetLongHandler/SetHandler name matching
// For any parser with registered options, SetShortHandler with a byte matching
// a registered short option sets the handler, SetLongHandler with a name matching
// a registered long option sets the handler (including single-character names),
// and SetHandler parses the dash prefix (-- for long, - for short) and delegates.
// Validates: Requirements 5.1, 5.2, 5.3
func TestPropertySetHandlerNameMatching(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Generate 1–4 short options (NoArgument for simplicity).
		nShort := 1 + rng.Intn(4)
		shortPerm := rng.Perm(len(validShortChars))
		chars := make([]byte, nShort)
		for i := range chars {
			chars[i] = validShortChars[shortPerm[i]]
		}

		// Generate 1–4 long options (NoArgument for simplicity).
		nLong := 1 + rng.Intn(4)
		longPerm := rng.Perm(len(validLongNames))
		longNames := make([]string, nLong)
		for i := range longNames {
			longNames[i] = validLongNames[longPerm[i]]
		}

		// Build parser with nil handlers on all options.
		shortMap := make(map[byte]*Flag)
		for _, c := range chars {
			shortMap[c] = &Flag{Name: string(c), HasArg: NoArgument}
		}
		longMap := make(map[string]*Flag)
		for _, name := range longNames {
			longMap[name] = &Flag{Name: name, HasArg: NoArgument}
		}

		// Build args: one of each short, then one of each long.
		var args []string
		for _, c := range chars {
			args = append(args, "-"+string(c))
		}
		for _, name := range longNames {
			args = append(args, "--"+name)
		}

		cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
		p, err := NewParser(cfg, shortMap, longMap, args)
		if err != nil {
			return true
		}

		// --- Test SetShortHandler ---
		var shortCalls []handlerCall
		shortHandler := func(name string, arg string) error {
			shortCalls = append(shortCalls, handlerCall{name, arg})
			return nil
		}
		for _, c := range chars {
			if err := p.SetShortHandler(c, shortHandler); err != nil {
				t.Logf("seed=%d SetShortHandler(%c) unexpected error: %v", seed, c, err)
				return false
			}
		}

		// --- Test SetLongHandler ---
		var longCalls []handlerCall
		longHandler := func(name string, arg string) error {
			longCalls = append(longCalls, handlerCall{name, arg})
			return nil
		}
		for _, name := range longNames {
			if err := p.SetLongHandler(name, longHandler); err != nil {
				t.Logf("seed=%d SetLongHandler(%q) unexpected error: %v", seed, name, err)
				return false
			}
		}

		// Parse and verify all handlers are invoked (no Options yielded).
		for _, err := range p.Options() {
			if err != nil {
				t.Logf("seed=%d unexpected parse error: %v", seed, err)
				return false
			}
		}

		if len(shortCalls) != nShort {
			t.Logf("seed=%d short handler calls: got %d, want %d", seed, len(shortCalls), nShort)
			return false
		}
		if len(longCalls) != nLong {
			t.Logf("seed=%d long handler calls: got %d, want %d", seed, len(longCalls), nLong)
			return false
		}

		// Verify each short handler received the correct name.
		for i, c := range chars {
			if shortCalls[i].name != string(c) {
				t.Logf("seed=%d short call[%d] name: got %q, want %q", seed, i, shortCalls[i].name, string(c))
				return false
			}
		}
		// Verify each long handler received the correct name.
		for i, name := range longNames {
			if longCalls[i].name != name {
				t.Logf("seed=%d long call[%d] name: got %q, want %q", seed, i, longCalls[i].name, name)
				return false
			}
		}

		// --- Test SetHandler with dash-prefix delegation ---
		// Build a fresh parser to test SetHandler.
		shortMap2 := make(map[byte]*Flag)
		for _, c := range chars {
			shortMap2[c] = &Flag{Name: string(c), HasArg: NoArgument}
		}
		longMap2 := make(map[string]*Flag)
		for _, name := range longNames {
			longMap2[name] = &Flag{Name: name, HasArg: NoArgument}
		}

		args2 := make([]string, len(args))
		copy(args2, args)
		p2, err := NewParser(cfg, shortMap2, longMap2, args2)
		if err != nil {
			return true
		}

		var setHandlerCalls []handlerCall
		setHandlerFn := func(name string, arg string) error {
			setHandlerCalls = append(setHandlerCalls, handlerCall{name, arg})
			return nil
		}

		// SetHandler("--name") should delegate to SetLongHandler.
		for _, name := range longNames {
			if err := p2.SetHandler("--"+name, setHandlerFn); err != nil {
				t.Logf("seed=%d SetHandler(--%s) unexpected error: %v", seed, name, err)
				return false
			}
		}
		// SetHandler("-c") should delegate to SetShortHandler.
		for _, c := range chars {
			if err := p2.SetHandler("-"+string(c), setHandlerFn); err != nil {
				t.Logf("seed=%d SetHandler(-%c) unexpected error: %v", seed, c, err)
				return false
			}
		}

		// Parse and verify all handlers are invoked via SetHandler.
		for _, err := range p2.Options() {
			if err != nil {
				t.Logf("seed=%d unexpected parse error (SetHandler path): %v", seed, err)
				return false
			}
		}

		if len(setHandlerCalls) != nShort+nLong {
			t.Logf("seed=%d SetHandler calls: got %d, want %d", seed, len(setHandlerCalls), nShort+nLong)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 9 (SetShortHandler/SetLongHandler/SetHandler name matching) failed: %v", err)
	}
}

// Feature: option-handlers, Property 10: Reject unknown names
// For any byte not matching a registered short option, SetShortHandler returns
// non-nil error. For any name not matching a registered long option, SetLongHandler
// returns non-nil error. For any name not matching via SetHandler lookup, SetHandler
// returns non-nil error.
// Validates: Requirements 5.4, 5.5, 5.6
func TestPropertySetHandlerRejectUnknown(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Generate 1–4 short options.
		nShort := 1 + rng.Intn(4)
		shortPerm := rng.Perm(len(validShortChars))
		chars := make([]byte, nShort)
		for i := range chars {
			chars[i] = validShortChars[shortPerm[i]]
		}
		registered := make(map[byte]bool)
		for _, c := range chars {
			registered[c] = true
		}

		// Generate 1–4 long options.
		nLong := 1 + rng.Intn(4)
		longPerm := rng.Perm(len(validLongNames))
		longNames := make([]string, nLong)
		for i := range longNames {
			longNames[i] = validLongNames[longPerm[i]]
		}
		registeredLong := make(map[string]bool)
		for _, name := range longNames {
			registeredLong[name] = true
		}

		// Build parser.
		shortMap := make(map[byte]*Flag)
		for _, c := range chars {
			shortMap[c] = &Flag{Name: string(c), HasArg: NoArgument}
		}
		longMap := make(map[string]*Flag)
		for _, name := range longNames {
			longMap[name] = &Flag{Name: name, HasArg: NoArgument}
		}

		p, err := NewParser(ParserConfig{enableErrors: true}, shortMap, longMap, nil)
		if err != nil {
			return true
		}

		dummy := func(string, string) error { return nil }

		// Pick a short option byte NOT in the registered set.
		var unknownShort byte
		for _, idx := range shortPerm[nShort:] {
			c := validShortChars[idx]
			if !registered[c] {
				unknownShort = c
				break
			}
		}
		if unknownShort != 0 {
			if err := p.SetShortHandler(unknownShort, dummy); err == nil {
				t.Logf("seed=%d SetShortHandler(%c) expected error, got nil", seed, unknownShort)
				return false
			}
		}

		// Pick a long option name NOT in the registered set.
		var unknownLong string
		for _, idx := range longPerm[nLong:] {
			name := validLongNames[idx]
			if !registeredLong[name] {
				unknownLong = name
				break
			}
		}
		if unknownLong != "" {
			if err := p.SetLongHandler(unknownLong, dummy); err == nil {
				t.Logf("seed=%d SetLongHandler(%q) expected error, got nil", seed, unknownLong)
				return false
			}
		}

		// SetHandler with "--unknownname" should return error.
		if unknownLong != "" {
			if err := p.SetHandler("--"+unknownLong, dummy); err == nil {
				t.Logf("seed=%d SetHandler(--%s) expected error, got nil", seed, unknownLong)
				return false
			}
		}

		// SetHandler with "-X" where X is not registered should return error.
		if unknownShort != 0 {
			if err := p.SetHandler("-"+string(unknownShort), dummy); err == nil {
				t.Logf("seed=%d SetHandler(-%c) expected error, got nil", seed, unknownShort)
				return false
			}
		}

		// SetHandler with no dash prefix should return error.
		// Generate a random prefix-less name from the seed.
		prefixlessNames := []string{"noprefixname", "bare", "plain", "nodash", "raw"}
		prefixless := prefixlessNames[rng.Intn(len(prefixlessNames))]
		if err := p.SetHandler(prefixless, dummy); err == nil {
			t.Logf("seed=%d SetHandler(%q) expected error for no-prefix name, got nil", seed, prefixless)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10 (Reject unknown names) failed: %v", err)
	}
}

// Feature: option-handlers, Property 11: SetHandler variants do not walk parent chain
// Calling Set*Handler on child with name matching only a parent option returns
// error; parent Flag unmodified.
// Validates: Requirements 5.7
func TestPropertySetHandlerNoParentWalk(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Shuffle valid short chars and split between parent and child.
		shortPerm := rng.Perm(len(validShortChars))
		nParentShort := 1 + rng.Intn(4) // 1–4 parent short opts
		nChildShort := 1 + rng.Intn(4)  // 1–4 child short opts
		if nParentShort+nChildShort > len(validShortChars) {
			nChildShort = len(validShortChars) - nParentShort
		}

		parentShortChars := make([]byte, nParentShort)
		for i := range parentShortChars {
			parentShortChars[i] = validShortChars[shortPerm[i]]
		}
		childShortChars := make([]byte, nChildShort)
		for i := range childShortChars {
			childShortChars[i] = validShortChars[shortPerm[nParentShort+i]]
		}

		// Shuffle valid long names and split between parent and child.
		longPerm := rng.Perm(len(validLongNames))
		nParentLong := 1 + rng.Intn(3) // 1–3 parent long opts
		nChildLong := 1 + rng.Intn(3)  // 1–3 child long opts
		if nParentLong+nChildLong > len(validLongNames) {
			nChildLong = len(validLongNames) - nParentLong
		}

		parentLongNames := make([]string, nParentLong)
		for i := range parentLongNames {
			parentLongNames[i] = validLongNames[longPerm[i]]
		}
		childLongNames := make([]string, nChildLong)
		for i := range childLongNames {
			childLongNames[i] = validLongNames[longPerm[nParentLong+i]]
		}

		// Build parent parser with nil Handle on all options.
		parentShortMap := make(map[byte]*Flag)
		for _, c := range parentShortChars {
			parentShortMap[c] = &Flag{Name: string(c), HasArg: NoArgument}
		}
		parentLongMap := make(map[string]*Flag)
		for _, name := range parentLongNames {
			parentLongMap[name] = &Flag{Name: name, HasArg: NoArgument}
		}

		cfg := ParserConfig{enableErrors: true}
		parent, err := NewParser(cfg, parentShortMap, parentLongMap, nil)
		if err != nil {
			return true
		}

		// Build child parser with its own disjoint options.
		childShortMap := make(map[byte]*Flag)
		for _, c := range childShortChars {
			childShortMap[c] = &Flag{Name: string(c), HasArg: NoArgument}
		}
		childLongMap := make(map[string]*Flag)
		for _, name := range childLongNames {
			childLongMap[name] = &Flag{Name: name, HasArg: NoArgument}
		}

		child, err := NewParser(cfg, childShortMap, childLongMap, nil)
		if err != nil {
			return true
		}

		// Link child to parent.
		parent.AddCmd("sub", child)

		dummy := func(string, string) error { return nil }

		// SetShortHandler on child for parent-only short options must fail.
		for _, c := range parentShortChars {
			if err := child.SetShortHandler(c, dummy); err == nil {
				t.Logf("seed=%d child.SetShortHandler(%c) expected error, got nil", seed, c)
				return false
			}
		}

		// SetLongHandler on child for parent-only long options must fail.
		for _, name := range parentLongNames {
			if err := child.SetLongHandler(name, dummy); err == nil {
				t.Logf("seed=%d child.SetLongHandler(%q) expected error, got nil", seed, name)
				return false
			}
		}

		// SetHandler on child for parent-only options must fail.
		for _, c := range parentShortChars {
			if err := child.SetHandler("-"+string(c), dummy); err == nil {
				t.Logf("seed=%d child.SetHandler(-%c) expected error, got nil", seed, c)
				return false
			}
		}
		for _, name := range parentLongNames {
			if err := child.SetHandler("--"+name, dummy); err == nil {
				t.Logf("seed=%d child.SetHandler(--%s) expected error, got nil", seed, name)
				return false
			}
		}

		// Verify parent Flags still have nil Handle (unmodified).
		for _, c := range parentShortChars {
			if parentShortMap[c].Handle != nil {
				t.Logf("seed=%d parent short flag %c Handle was modified", seed, c)
				return false
			}
		}
		for _, name := range parentLongNames {
			if parentLongMap[name].Handle != nil {
				t.Logf("seed=%d parent long flag %q Handle was modified", seed, name)
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 11 (SetHandler variants do not walk parent chain) failed: %v", err)
	}
}

// Feature: option-handlers, Property 6: Parent-chain handler dispatch
// Child inheriting handled Flag from ancestor invokes ancestor's handler
// when resolved from child.
// Validates: Requirements 3.2, 3.4, 3.6
func TestPropertyParentChainHandlerDispatch(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Generate 1–3 parent-only short options with handlers.
		nParentShort := 1 + rng.Intn(3)
		shortPerm := rng.Perm(len(validShortChars))
		parentShortChars := make([]byte, nParentShort)
		for i := range parentShortChars {
			parentShortChars[i] = validShortChars[shortPerm[i]]
		}

		// Generate 1–3 parent-only long options with handlers.
		nParentLong := 1 + rng.Intn(3)
		longPerm := rng.Perm(len(validLongNames))
		parentLongNames := make([]string, nParentLong)
		for i := range parentLongNames {
			parentLongNames[i] = validLongNames[longPerm[i]]
		}

		// Track handler invocations.
		var calls []handlerCall
		handler := func(name string, arg string) error {
			calls = append(calls, handlerCall{name, arg})
			return nil
		}

		// Build parent with handled flags.
		parentShortMap := make(map[byte]*Flag)
		for _, c := range parentShortChars {
			parentShortMap[c] = &Flag{Name: string(c), HasArg: NoArgument, Handle: handler}
		}
		parentLongMap := make(map[string]*Flag)
		for _, name := range parentLongNames {
			parentLongMap[name] = &Flag{Name: name, HasArg: NoArgument, Handle: handler}
		}

		cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
		parent, err := NewParser(cfg, parentShortMap, parentLongMap, nil)
		if err != nil {
			return true
		}

		// Build child with disjoint options (no overlap with parent).
		// Child has no options of its own — it inherits everything from parent.
		// Build args that reference parent-only options.
		var args []string
		for _, c := range parentShortChars {
			args = append(args, "-"+string(c))
		}
		for _, name := range parentLongNames {
			args = append(args, "--"+name)
		}

		child, err := NewParser(cfg, nil, nil, args)
		if err != nil {
			return true
		}
		parent.AddCmd("sub", child)

		// Parse from child — parent-chain walk should resolve parent's flags
		// and invoke parent's handlers.
		calls = nil
		for _, err := range child.Options() {
			if err != nil {
				t.Logf("seed=%d unexpected error: %v", seed, err)
				return false
			}
		}

		expected := nParentShort + nParentLong
		if len(calls) != expected {
			t.Logf("seed=%d handler calls: got %d, want %d", seed, len(calls), expected)
			return false
		}

		// Verify each call matches a parent-defined option.
		parentNames := make(map[string]bool)
		for _, c := range parentShortChars {
			parentNames[string(c)] = true
		}
		for _, name := range parentLongNames {
			parentNames[name] = true
		}
		for _, c := range calls {
			if !parentNames[c.name] {
				t.Logf("seed=%d handler called with unexpected name %q", seed, c.name)
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6 (Parent-chain handler dispatch) failed: %v", err)
	}
}

// Feature: option-handlers, Property 7: Child overloading wins
// Child's definition determines dispatch: child handler invoked if set,
// Option yielded if child Handle is nil, regardless of parent handler status.
// Validates: Requirements 3.5, 7.2, 7.3
func TestPropertyChildOverloadingWins(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Pick 2–4 short options that both parent and child define (overloaded).
		nOverlap := 2 + rng.Intn(3)
		shortPerm := rng.Perm(len(validShortChars))
		overlapChars := make([]byte, nOverlap)
		for i := range overlapChars {
			overlapChars[i] = validShortChars[shortPerm[i]]
		}

		// Pick 1–3 long options that both parent and child define (overloaded).
		nOverlapLong := 1 + rng.Intn(3)
		longPerm := rng.Perm(len(validLongNames))
		overlapLongNames := make([]string, nOverlapLong)
		for i := range overlapLongNames {
			overlapLongNames[i] = validLongNames[longPerm[i]]
		}

		var parentCalls []handlerCall
		var childCalls []handlerCall
		parentHandler := func(name string, arg string) error {
			parentCalls = append(parentCalls, handlerCall{name, arg})
			return nil
		}
		childHandler := func(name string, arg string) error {
			childCalls = append(childCalls, handlerCall{name, arg})
			return nil
		}

		// For each overloaded option, randomly decide child's handler status:
		// - child has handler → child handler invoked (parent handler ignored)
		// - child has nil Handle → Option yielded (parent handler ignored)
		type overloadDef struct {
			name         string
			isShort      bool
			childHandled bool
		}
		var defs []overloadDef

		// Ensure at least one child-handled and one child-non-handled.
		shortHandled := make([]bool, nOverlap)
		for i := range shortHandled {
			shortHandled[i] = rng.Intn(2) == 0
		}
		shortHandled[0] = true
		if nOverlap > 1 {
			shortHandled[1] = false
		}

		longHandled := make([]bool, nOverlapLong)
		for i := range longHandled {
			longHandled[i] = rng.Intn(2) == 0
		}

		// Build parent — all overloaded options have handlers.
		parentShortMap := make(map[byte]*Flag)
		for _, c := range overlapChars {
			parentShortMap[c] = &Flag{Name: string(c), HasArg: NoArgument, Handle: parentHandler}
		}
		parentLongMap := make(map[string]*Flag)
		for _, name := range overlapLongNames {
			parentLongMap[name] = &Flag{Name: name, HasArg: NoArgument, Handle: parentHandler}
		}

		cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
		parent, err := NewParser(cfg, parentShortMap, parentLongMap, nil)
		if err != nil {
			return true
		}

		// Build child — overloaded options with varying handler status.
		childShortMap := make(map[byte]*Flag)
		for i, c := range overlapChars {
			f := &Flag{Name: string(c), HasArg: NoArgument}
			if shortHandled[i] {
				f.Handle = childHandler
			}
			childShortMap[c] = f
			defs = append(defs, overloadDef{name: string(c), isShort: true, childHandled: shortHandled[i]})
		}
		childLongMap := make(map[string]*Flag)
		for i, name := range overlapLongNames {
			f := &Flag{Name: name, HasArg: NoArgument}
			if longHandled[i] {
				f.Handle = childHandler
			}
			childLongMap[name] = f
			defs = append(defs, overloadDef{name: name, isShort: false, childHandled: longHandled[i]})
		}

		// Build args referencing all overloaded options.
		var args []string
		for _, d := range defs {
			if d.isShort {
				args = append(args, "-"+d.name)
			} else {
				args = append(args, "--"+d.name)
			}
		}

		child, err := NewParser(cfg, childShortMap, childLongMap, args)
		if err != nil {
			return true
		}
		parent.AddCmd("sub", child)

		parentCalls = nil
		childCalls = nil
		var yielded []Option
		for opt, err := range child.Options() {
			if err != nil {
				t.Logf("seed=%d unexpected error: %v", seed, err)
				return false
			}
			yielded = append(yielded, opt)
		}

		// Parent handler must NEVER be invoked (child overloads all).
		if len(parentCalls) != 0 {
			t.Logf("seed=%d parent handler invoked %d times, want 0", seed, len(parentCalls))
			return false
		}

		// Count expected child handler calls and yielded options.
		var expectChildCalls, expectYielded int
		for _, d := range defs {
			if d.childHandled {
				expectChildCalls++
			} else {
				expectYielded++
			}
		}

		if len(childCalls) != expectChildCalls {
			t.Logf("seed=%d child handler calls: got %d, want %d", seed, len(childCalls), expectChildCalls)
			return false
		}
		if len(yielded) != expectYielded {
			t.Logf("seed=%d yielded options: got %d, want %d", seed, len(yielded), expectYielded)
			return false
		}

		// Verify yielded options are only from non-handled child defs.
		childHandledNames := make(map[string]bool)
		for _, d := range defs {
			if d.childHandled {
				childHandledNames[d.name] = true
			}
		}
		for _, opt := range yielded {
			if childHandledNames[opt.Name] {
				t.Logf("seed=%d option %q yielded but child has handler", seed, opt.Name)
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 7 (Child overloading wins) failed: %v", err)
	}
}

// Feature: option-handlers, Property 12: Compaction handler dispatch order
// Compacted short options dispatched left-to-right: handlers for handled,
// Options for non-handled.
// Validates: Requirements 8.1, 8.2
func TestPropertyCompactionHandlerDispatchOrder(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Pick 3–6 distinct NoArgument short options for compaction.
		nShort := 3 + rng.Intn(4)
		shortPerm := rng.Perm(len(validShortChars))
		chars := make([]byte, nShort)
		for i := range chars {
			chars[i] = validShortChars[shortPerm[i]]
		}

		// Randomly assign handled/non-handled, ensuring at least one of each.
		handled := make([]bool, nShort)
		for i := range handled {
			handled[i] = rng.Intn(2) == 0
		}
		handled[0] = true
		handled[1] = false

		// Track handler invocations in order.
		type call struct {
			name string
			seq  int
		}
		seq := 0
		var handlerCalls []call
		var yieldedOpts []Option

		shortMap := make(map[byte]*Flag)
		for i, c := range chars {
			f := &Flag{Name: string(c), HasArg: NoArgument}
			if handled[i] {
				f.Handle = func(name string, arg string) error {
					handlerCalls = append(handlerCalls, call{name, seq})
					seq++
					return nil
				}
			}
			shortMap[c] = f
		}

		// Build a single compacted arg: -abc...
		compacted := "-"
		for _, c := range chars {
			compacted += string(c)
		}

		p, err := NewParser(ParserConfig{enableErrors: true}, shortMap, nil, []string{compacted})
		if err != nil {
			return true
		}

		for opt, err := range p.Options() {
			if err != nil {
				t.Logf("seed=%d unexpected error: %v", seed, err)
				return false
			}
			yieldedOpts = append(yieldedOpts, opt)
			seq++
		}

		// Verify dispatch order matches left-to-right character order.
		hIdx := 0
		yIdx := 0
		for i, c := range chars {
			if handled[i] {
				if hIdx >= len(handlerCalls) {
					t.Logf("seed=%d missing handler call for %c at position %d", seed, c, i)
					return false
				}
				if handlerCalls[hIdx].name != string(c) {
					t.Logf("seed=%d handler call[%d] name: got %q, want %q", seed, hIdx, handlerCalls[hIdx].name, string(c))
					return false
				}
				hIdx++
			} else {
				if yIdx >= len(yieldedOpts) {
					t.Logf("seed=%d missing yielded option for %c at position %d", seed, c, i)
					return false
				}
				if yieldedOpts[yIdx].Name != string(c) {
					t.Logf("seed=%d yielded[%d] name: got %q, want %q", seed, yIdx, yieldedOpts[yIdx].Name, string(c))
					return false
				}
				yIdx++
			}
		}

		// Verify counts.
		var expectHandled, expectYielded int
		for _, h := range handled {
			if h {
				expectHandled++
			} else {
				expectYielded++
			}
		}
		if len(handlerCalls) != expectHandled {
			t.Logf("seed=%d handler calls: got %d, want %d", seed, len(handlerCalls), expectHandled)
			return false
		}
		if len(yieldedOpts) != expectYielded {
			t.Logf("seed=%d yielded: got %d, want %d", seed, len(yieldedOpts), expectYielded)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 12 (Compaction handler dispatch order) failed: %v", err)
	}
}

// Feature: option-handlers, Property 13: Compaction error stops remaining
// Handler error at position N prevents invocation of handlers at N+1 onward.
// Validates: Requirements 8.3, 9.2
func TestPropertyCompactionErrorStopsRemaining(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Pick 3–6 distinct NoArgument short options for compaction.
		nShort := 3 + rng.Intn(4)
		shortPerm := rng.Perm(len(validShortChars))
		chars := make([]byte, nShort)
		for i := range chars {
			chars[i] = validShortChars[shortPerm[i]]
		}

		// Pick a random position (1 <= errPos < nShort) for the error.
		// Position 0 would mean nothing runs before the error, so start at 1
		// to ensure at least one handler runs before the error.
		errPos := 1 + rng.Intn(nShort-1)
		errMsg := fmt.Sprintf("handler error at position %d", errPos)
		sentinel := fmt.Errorf("%s", errMsg)

		// All options get handlers. The one at errPos returns an error.
		var invoked []int
		shortMap := make(map[byte]*Flag)
		for i, c := range chars {
			pos := i // capture
			f := &Flag{Name: string(c), HasArg: NoArgument}
			if pos == errPos {
				f.Handle = func(name string, arg string) error {
					invoked = append(invoked, pos)
					return sentinel
				}
			} else {
				f.Handle = func(name string, arg string) error {
					invoked = append(invoked, pos)
					return nil
				}
			}
			shortMap[c] = f
		}

		// Build compacted arg.
		compacted := "-"
		for _, c := range chars {
			compacted += string(c)
		}

		p, err := NewParser(ParserConfig{enableErrors: true}, shortMap, nil, []string{compacted})
		if err != nil {
			return true
		}

		var sawError bool
		for opt, err := range p.Options() {
			if err != nil {
				if err.Error() != errMsg {
					t.Logf("seed=%d unexpected error: %v", seed, err)
					return false
				}
				if opt != (Option{}) {
					t.Logf("seed=%d expected zero Option with error, got %+v", seed, opt)
					return false
				}
				sawError = true
				continue
			}
		}

		if !sawError {
			t.Logf("seed=%d never saw handler error", seed)
			return false
		}

		// Handlers at positions 0..errPos should have been invoked.
		// Handlers at positions errPos+1..nShort-1 should NOT have been invoked.
		expectedInvoked := errPos + 1 // positions 0 through errPos inclusive
		if len(invoked) != expectedInvoked {
			t.Logf("seed=%d invoked handlers: got %d, want %d (errPos=%d)", seed, len(invoked), expectedInvoked, errPos)
			return false
		}

		// Verify invocation order is 0, 1, ..., errPos.
		for i, pos := range invoked {
			if pos != i {
				t.Logf("seed=%d invoked[%d] = %d, want %d", seed, i, pos, i)
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 13 (Compaction error stops remaining) failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Unit tests for handler invocation via short and long options
// ---------------------------------------------------------------------------

func TestHandlerShortOptionCurrentParser(t *testing.T) {
	// Handler invoked for short option from current parser.
	var called string
	handler := func(name string, arg string) error {
		called = name
		return nil
	}

	shortOpts := map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument, Handle: handler},
	}
	p, err := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-v"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if called != "v" {
		t.Fatalf("handler called with %q, want %q", called, "v")
	}
}

func TestHandlerShortOptionAncestor(t *testing.T) {
	// Handler invoked for short option from ancestor via parent-chain walk.
	var called string
	handler := func(name string, arg string) error {
		called = name
		return nil
	}

	parentShort := map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument, Handle: handler},
	}
	cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
	parent, err := NewParser(cfg, parentShort, nil, nil)
	if err != nil {
		t.Fatalf("NewParser parent: %v", err)
	}

	child, err := NewParser(cfg, nil, nil, []string{"-v"})
	if err != nil {
		t.Fatalf("NewParser child: %v", err)
	}
	parent.AddCmd("sub", child)

	opts, errs := collectOptions(child)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if called != "v" {
		t.Fatalf("handler called with %q, want %q", called, "v")
	}
}

func TestHandlerLongOptionCurrentParser(t *testing.T) {
	// Handler invoked for long option from current parser.
	var called string
	handler := func(name string, arg string) error {
		called = name
		return nil
	}

	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument, Handle: handler},
	}
	p, err := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, nil, longOpts, []string{"--verbose"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if called != "verbose" {
		t.Fatalf("handler called with %q, want %q", called, "verbose")
	}
}

func TestHandlerLongOptionAncestor(t *testing.T) {
	// Handler invoked for long option from ancestor via parent-chain walk.
	var called string
	handler := func(name string, arg string) error {
		called = name
		return nil
	}

	parentLong := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument, Handle: handler},
	}
	cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
	parent, err := NewParser(cfg, nil, parentLong, nil)
	if err != nil {
		t.Fatalf("NewParser parent: %v", err)
	}

	child, err := NewParser(cfg, nil, nil, []string{"--verbose"})
	if err != nil {
		t.Fatalf("NewParser child: %v", err)
	}
	parent.AddCmd("sub", child)

	opts, errs := collectOptions(child)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if called != "verbose" {
		t.Fatalf("handler called with %q, want %q", called, "verbose")
	}
}

func TestHandlerChildOverloadShadowsParent(t *testing.T) {
	// Child overloading with handler shadows parent.
	var parentCalled, childCalled bool
	parentHandler := func(name string, arg string) error {
		parentCalled = true
		return nil
	}
	childHandler := func(name string, arg string) error {
		childCalled = true
		return nil
	}

	cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
	parentShort := map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument, Handle: parentHandler},
	}
	parent, err := NewParser(cfg, parentShort, nil, nil)
	if err != nil {
		t.Fatalf("NewParser parent: %v", err)
	}

	childShort := map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument, Handle: childHandler},
	}
	child, err := NewParser(cfg, childShort, nil, []string{"-v"})
	if err != nil {
		t.Fatalf("NewParser child: %v", err)
	}
	parent.AddCmd("sub", child)

	opts, errs := collectOptions(child)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if parentCalled {
		t.Fatal("parent handler should not have been called")
	}
	if !childCalled {
		t.Fatal("child handler should have been called")
	}
}

func TestHandlerMixedHandledAndNonHandled(t *testing.T) {
	// Mixed handled and non-handled Flags in same parser.
	var handledName string
	handler := func(name string, arg string) error {
		handledName = name
		return nil
	}

	shortOpts := map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument, Handle: handler},
		'x': {Name: "x", HasArg: NoArgument}, // no handler
	}
	p, err := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-v", "-x"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if handledName != "v" {
		t.Fatalf("handler called with %q, want %q", handledName, "v")
	}
	expected := []Option{{Name: "x", HasArg: false, Arg: ""}}
	if !optionsEqual(opts, expected) {
		t.Fatalf("options: got %v, want %v", opts, expected)
	}
}

func TestHandlerNoArgumentReceivesEmptyArg(t *testing.T) {
	// NoArgument handler receives empty string arg.
	var receivedArg string
	handler := func(name string, arg string) error {
		receivedArg = arg
		return nil
	}

	shortOpts := map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument, Handle: handler},
	}
	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument, Handle: handler},
	}
	// Test short
	p, err := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, []string{"-v"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	collectOptions(p)
	if receivedArg != "" {
		t.Fatalf("short: handler received arg %q, want empty string", receivedArg)
	}

	// Test long
	receivedArg = "sentinel"
	p, err = NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, []string{"--verbose"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	collectOptions(p)
	if receivedArg != "" {
		t.Fatalf("long: handler received arg %q, want empty string", receivedArg)
	}
}

func TestHandlerRequiredArgumentReceivesArg(t *testing.T) {
	// Handler with RequiredArgument receives correct arg
	var receivedName, receivedArg string
	handler := func(name string, arg string) error {
		receivedName = name
		receivedArg = arg
		return nil
	}

	shortOpts := map[byte]*Flag{
		'o': {Name: "o", HasArg: RequiredArgument, Handle: handler},
	}
	longOpts := map[string]*Flag{
		"output": {Name: "output", HasArg: RequiredArgument, Handle: handler},
	}

	// Short with separate arg
	p, err := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, []string{"-o", "file.txt"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "o" || receivedArg != "file.txt" {
		t.Fatalf("short: handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, "o", "file.txt")
	}

	// Long with = syntax
	receivedName, receivedArg = "", ""
	p, err = NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, []string{"--output=result.txt"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	opts, errs = collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "output" || receivedArg != "result.txt" {
		t.Fatalf("long: handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, "output", "result.txt")
	}
}

func TestHandlerOptionalArgumentReceivesArg(t *testing.T) {
	// Handler with OptionalArgument receives correct arg (present and absent)
	var receivedName, receivedArg string
	handler := func(name string, arg string) error {
		receivedName = name
		receivedArg = arg
		return nil
	}

	shortOpts := map[byte]*Flag{
		'd': {Name: "d", HasArg: OptionalArgument, Handle: handler},
	}
	longOpts := map[string]*Flag{
		"debug": {Name: "debug", HasArg: OptionalArgument, Handle: handler},
	}

	// Short with arg attached (no space for optional)
	p, err := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, []string{"-dlevel3"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "d" || receivedArg != "level3" {
		t.Fatalf("short present: handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, "d", "level3")
	}

	// Short without arg
	receivedName, receivedArg = "", "sentinel"
	p, err = NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, []string{"-d"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	opts, errs = collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "d" || receivedArg != "" {
		t.Fatalf("short absent: handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, "d", "")
	}

	// Long with = syntax
	receivedName, receivedArg = "", ""
	p, err = NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, []string{"--debug=trace"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	opts, errs = collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "debug" || receivedArg != "trace" {
		t.Fatalf("long present: handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, "debug", "trace")
	}

	// Long without arg
	receivedName, receivedArg = "", "sentinel"
	p, err = NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, []string{"--debug"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}
	opts, errs = collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "debug" || receivedArg != "" {
		t.Fatalf("long absent: handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, "debug", "")
	}
}

func TestHandlerNilHandleIdenticalOutput(t *testing.T) {
	// Nil Handle produces identical output to pre-handler behavior.
	shortOpts := map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
		'o': {Name: "o", HasArg: RequiredArgument},
	}
	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
		"output":  {Name: "output", HasArg: RequiredArgument},
	}
	args := []string{"-v", "-o", "file.txt", "--verbose", "--output=result.txt"}

	p, err := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, args)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}

	expected := []Option{
		{Name: "v", HasArg: false, Arg: ""},
		{Name: "o", HasArg: true, Arg: "file.txt"},
		{Name: "verbose", HasArg: false, Arg: ""},
		{Name: "output", HasArg: true, Arg: "result.txt"},
	}
	if !optionsEqual(opts, expected) {
		t.Fatalf("nil Handle output: got %v, want %v", opts, expected)
	}
}

func TestHandlerCompactionAllHandled(t *testing.T) {
	// Handler invocation during compaction -abc.
	var calls []string
	makeHandler := func(name string) func(string, string) error {
		return func(n string, arg string) error {
			calls = append(calls, n)
			return nil
		}
	}

	shortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument, Handle: makeHandler("a")},
		'b': {Name: "b", HasArg: NoArgument, Handle: makeHandler("b")},
		'c': {Name: "c", HasArg: NoArgument, Handle: makeHandler("c")},
	}
	p, err := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-abc"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if len(calls) != 3 || calls[0] != "a" || calls[1] != "b" || calls[2] != "c" {
		t.Fatalf("handlers called %v, want [a b c]", calls)
	}
}

func TestHandlerCompactionMixed(t *testing.T) {
	// Mixed compaction: some handled, some not.
	var handledNames []string
	handler := func(name string, arg string) error {
		handledNames = append(handledNames, name)
		return nil
	}

	shortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument, Handle: handler},
		'b': {Name: "b", HasArg: NoArgument}, // no handler — yields Option
		'c': {Name: "c", HasArg: NoArgument, Handle: handler},
	}
	p, err := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-abc"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	// 'b' should be yielded as an Option; 'a' and 'c' handled
	expected := []Option{{Name: "b", HasArg: false, Arg: ""}}
	if !optionsEqual(opts, expected) {
		t.Fatalf("options: got %v, want %v", opts, expected)
	}
	if len(handledNames) != 2 || handledNames[0] != "a" || handledNames[1] != "c" {
		t.Fatalf("handlers called %v, want [a c]", handledNames)
	}
}

func TestHandlerCompactionErrorStopsRemaining(t *testing.T) {
	// Handler error during compaction stops remaining options.
	handlerErr := fmt.Errorf("handler failed on b")
	var calls []string
	makeHandler := func(name string) func(string, string) error {
		return func(n string, arg string) error {
			calls = append(calls, n)
			if n == "b" {
				return handlerErr
			}
			return nil
		}
	}

	shortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument, Handle: makeHandler("a")},
		'b': {Name: "b", HasArg: NoArgument, Handle: makeHandler("b")},
		'c': {Name: "c", HasArg: NoArgument, Handle: makeHandler("c")},
	}
	p, err := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-abc"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	opts, errs := collectOptions(p)
	// Should get one error from 'b' handler, no options yielded
	var gotErr error
	for _, e := range errs {
		if e != nil {
			gotErr = e
		}
	}
	if gotErr == nil || gotErr.Error() != handlerErr.Error() {
		t.Fatalf("expected error %q, got %v", handlerErr, gotErr)
	}
	if len(opts) != 1 || (opts[0] != Option{}) {
		t.Fatalf("expected zero-value Option on error, got %v", opts)
	}
	// 'a' and 'b' should be called, 'c' should NOT
	if len(calls) != 2 || calls[0] != "a" || calls[1] != "b" {
		t.Fatalf("handlers called %v, want [a b]", calls)
	}
}

func TestHandlerSetShortHandlerOptstring(t *testing.T) {
	// SetShortHandler on short options registered via optstring.
	tests := []struct {
		name      string
		optstring string
		char      byte
		args      []string
		wantName  string
		wantArg   string
	}{
		{
			name:      "NoArgument short option",
			optstring: "vx",
			char:      'v',
			args:      []string{"-v"},
			wantName:  "v",
			wantArg:   "",
		},
		{
			name:      "RequiredArgument short option",
			optstring: "o:",
			char:      'o',
			args:      []string{"-o", "file.txt"},
			wantName:  "o",
			wantArg:   "file.txt",
		},
		{
			name:      "OptionalArgument short option with arg",
			optstring: "d::",
			char:      'd',
			args:      []string{"-dlevel3"},
			wantName:  "d",
			wantArg:   "level3",
		},
		{
			name:      "OptionalArgument short option without arg",
			optstring: "d::",
			char:      'd',
			args:      []string{"-d"},
			wantName:  "d",
			wantArg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedName, receivedArg string
			handler := func(name string, arg string) error {
				receivedName = name
				receivedArg = arg
				return nil
			}

			p, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt: %v", err)
			}
			if err := p.SetShortHandler(tt.char, handler); err != nil {
				t.Fatalf("SetShortHandler: %v", err)
			}

			opts, errs := collectOptions(p)
			for _, e := range errs {
				if e != nil {
					t.Fatalf("unexpected error: %v", e)
				}
			}
			if len(opts) != 0 {
				t.Fatalf("expected no options yielded, got %d", len(opts))
			}
			if receivedName != tt.wantName || receivedArg != tt.wantArg {
				t.Fatalf("handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, tt.wantName, tt.wantArg)
			}
		})
	}
}

func TestHandlerSetLongHandler(t *testing.T) {
	// SetLongHandler on long options.
	tests := []struct {
		name     string
		longName string
		args     []string
		wantName string
		wantArg  string
	}{
		{
			name:     "NoArgument long option",
			longName: "verbose",
			args:     []string{"--verbose"},
			wantName: "verbose",
			wantArg:  "",
		},
		{
			name:     "RequiredArgument long option with equals",
			longName: "output",
			args:     []string{"--output=file.txt"},
			wantName: "output",
			wantArg:  "file.txt",
		},
		{
			name:     "RequiredArgument long option with space",
			longName: "output",
			args:     []string{"--output", "file.txt"},
			wantName: "output",
			wantArg:  "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedName, receivedArg string
			handler := func(name string, arg string) error {
				receivedName = name
				receivedArg = arg
				return nil
			}

			hasArg := NoArgument
			if tt.wantArg != "" {
				hasArg = RequiredArgument
			}
			longOpts := []Flag{{Name: tt.longName, HasArg: hasArg}}
			p, err := GetOptLong(tt.args, "", longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			if err := p.SetLongHandler(tt.longName, handler); err != nil {
				t.Fatalf("SetLongHandler: %v", err)
			}

			opts, errs := collectOptions(p)
			for _, e := range errs {
				if e != nil {
					t.Fatalf("unexpected error: %v", e)
				}
			}
			if len(opts) != 0 {
				t.Fatalf("expected no options yielded, got %d", len(opts))
			}
			if receivedName != tt.wantName || receivedArg != tt.wantArg {
				t.Fatalf("handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, tt.wantName, tt.wantArg)
			}
		})
	}
}

func TestHandlerSetLongHandlerSingleChar(t *testing.T) {
	// SetLongHandler on single-character long option names (e.g., --v).
	var receivedName, receivedArg string
	handler := func(name string, arg string) error {
		receivedName = name
		receivedArg = arg
		return nil
	}

	longOpts := []Flag{{Name: "v", HasArg: NoArgument}}
	p, err := GetOptLong([]string{"--v"}, "", longOpts)
	if err != nil {
		t.Fatalf("GetOptLong: %v", err)
	}
	if err := p.SetLongHandler("v", handler); err != nil {
		t.Fatalf("SetLongHandler: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "v" || receivedArg != "" {
		t.Fatalf("handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, "v", "")
	}
}

func TestHandlerSetHandlerRejectUnregistered(t *testing.T) {
	// SetShortHandler/SetLongHandler/SetHandler return errors for unregistered names.
	handler := func(string, string) error { return nil }

	tests := []struct {
		name    string
		setup   func() *Parser
		setFn   func(*Parser) error
		wantErr bool
	}{
		{
			name: "SetShortHandler unregistered byte",
			setup: func() *Parser {
				p, _ := GetOpt(nil, "v")
				return p
			},
			setFn:   func(p *Parser) error { return p.SetShortHandler('x', handler) },
			wantErr: true,
		},
		{
			name: "SetLongHandler unregistered name",
			setup: func() *Parser {
				p, _ := GetOptLong(nil, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				return p
			},
			setFn:   func(p *Parser) error { return p.SetLongHandler("quiet", handler) },
			wantErr: true,
		},
		{
			name: "SetHandler unregistered long",
			setup: func() *Parser {
				p, _ := GetOptLong(nil, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				return p
			},
			setFn:   func(p *Parser) error { return p.SetHandler("--quiet", handler) },
			wantErr: true,
		},
		{
			name: "SetHandler unregistered short",
			setup: func() *Parser {
				p, _ := GetOpt(nil, "v")
				return p
			},
			setFn:   func(p *Parser) error { return p.SetHandler("-x", handler) },
			wantErr: true,
		},
		{
			name: "SetShortHandler registered succeeds",
			setup: func() *Parser {
				p, _ := GetOpt(nil, "v")
				return p
			},
			setFn:   func(p *Parser) error { return p.SetShortHandler('v', handler) },
			wantErr: false,
		},
		{
			name: "SetLongHandler registered succeeds",
			setup: func() *Parser {
				p, _ := GetOptLong(nil, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				return p
			},
			setFn:   func(p *Parser) error { return p.SetLongHandler("verbose", handler) },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
			err := tt.setFn(p)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestHandlerSetHandlerNoParentWalk(t *testing.T) {
	// SetShortHandler/SetLongHandler/SetHandler do not walk parent chain.
	handler := func(string, string) error { return nil }
	cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}

	tests := []struct {
		name  string
		setFn func(*Parser) error
	}{
		{
			name:  "SetShortHandler does not walk parent",
			setFn: func(child *Parser) error { return child.SetShortHandler('v', handler) },
		},
		{
			name:  "SetLongHandler does not walk parent",
			setFn: func(child *Parser) error { return child.SetLongHandler("verbose", handler) },
		},
		{
			name:  "SetHandler short does not walk parent",
			setFn: func(child *Parser) error { return child.SetHandler("-v", handler) },
		},
		{
			name:  "SetHandler long does not walk parent",
			setFn: func(child *Parser) error { return child.SetHandler("--verbose", handler) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentShort := map[byte]*Flag{
				'v': {Name: "v", HasArg: NoArgument},
			}
			parentLong := map[string]*Flag{
				"verbose": {Name: "verbose", HasArg: NoArgument},
			}
			parent, err := NewParser(cfg, parentShort, parentLong, nil)
			if err != nil {
				t.Fatalf("NewParser parent: %v", err)
			}

			child, err := NewParser(cfg, nil, nil, nil)
			if err != nil {
				t.Fatalf("NewParser child: %v", err)
			}
			parent.AddCmd("sub", child)

			err = tt.setFn(child)
			if err == nil {
				t.Fatal("expected error for parent-only option, got nil")
			}

			// Verify parent's Flag was not modified
			if parentShort['v'].Handle != nil {
				t.Fatal("parent short option Handle was modified")
			}
			if parentLong["verbose"].Handle != nil {
				t.Fatal("parent long option Handle was modified")
			}
		})
	}
}

func TestHandlerSetHandlerDelegatesToLong(t *testing.T) {
	// SetHandler("--verbose", ...) delegates to SetLongHandler.
	var receivedName string
	handler := func(name string, arg string) error {
		receivedName = name
		return nil
	}

	longOpts := []Flag{{Name: "verbose", HasArg: NoArgument}}
	p, err := GetOptLong([]string{"--verbose"}, "", longOpts)
	if err != nil {
		t.Fatalf("GetOptLong: %v", err)
	}
	if err := p.SetHandler("--verbose", handler); err != nil {
		t.Fatalf("SetHandler: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "verbose" {
		t.Fatalf("handler called with %q, want %q", receivedName, "verbose")
	}
}

func TestHandlerSetHandlerDelegatesToShort(t *testing.T) {
	// SetHandler("-v", ...) delegates to SetShortHandler.
	var receivedName string
	handler := func(name string, arg string) error {
		receivedName = name
		return nil
	}

	p, err := GetOpt([]string{"-v"}, "v")
	if err != nil {
		t.Fatalf("GetOpt: %v", err)
	}
	if err := p.SetHandler("-v", handler); err != nil {
		t.Fatalf("SetHandler: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "v" {
		t.Fatalf("handler called with %q, want %q", receivedName, "v")
	}
}

func TestHandlerSetHandlerSingleCharLong(t *testing.T) {
	// SetHandler("--v", ...) delegates to SetLongHandler for single-char long opts.
	var receivedName string
	handler := func(name string, arg string) error {
		receivedName = name
		return nil
	}

	longOpts := []Flag{{Name: "v", HasArg: NoArgument}}
	p, err := GetOptLong([]string{"--v"}, "", longOpts)
	if err != nil {
		t.Fatalf("GetOptLong: %v", err)
	}
	if err := p.SetHandler("--v", handler); err != nil {
		t.Fatalf("SetHandler: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "v" {
		t.Fatalf("handler called with %q, want %q", receivedName, "v")
	}
}

func TestHandlerSetHandlerNoDashPrefix(t *testing.T) {
	// SetHandler returns error for names without dash prefix.
	handler := func(string, string) error { return nil }

	tests := []struct {
		name    string
		optName string
	}{
		{name: "bare name", optName: "verbose"},
		{name: "empty string", optName: ""},
		{name: "alphanumeric", optName: "v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong(nil, "v", []Flag{{Name: "verbose", HasArg: NoArgument}})
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			err = p.SetHandler(tt.optName, handler)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tt.optName)
			}
		})
	}
}

func TestHandlerNewParserWithHandle(t *testing.T) {
	// NewParser with pre-built Flag structs containing Handle fields
	// (construction-time path).
	cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}

	tests := []struct {
		name      string
		shortOpts map[byte]*Flag
		longOpts  map[string]*Flag
		args      []string
		wantCalls []string // expected handler call names in order
		wantOpts  []Option // expected yielded options (non-handled)
	}{
		{
			name: "short option with Handle at construction",
			shortOpts: map[byte]*Flag{
				'v': {Name: "v", HasArg: NoArgument},
			},
			args:      []string{"-v"},
			wantCalls: []string{"v"},
		},
		{
			name: "long option with Handle at construction",
			longOpts: map[string]*Flag{
				"verbose": {Name: "verbose", HasArg: NoArgument},
			},
			args:      []string{"--verbose"},
			wantCalls: []string{"verbose"},
		},
		{
			name: "short option with RequiredArgument",
			shortOpts: map[byte]*Flag{
				'o': {Name: "o", HasArg: RequiredArgument},
			},
			args:      []string{"-o", "file.txt"},
			wantCalls: []string{"o"},
		},
		{
			name: "long option with RequiredArgument via =",
			longOpts: map[string]*Flag{
				"output": {Name: "output", HasArg: RequiredArgument},
			},
			args:      []string{"--output=file.txt"},
			wantCalls: []string{"output"},
		},
		{
			name: "mixed short and long with Handle",
			shortOpts: map[byte]*Flag{
				'v': {Name: "v", HasArg: NoArgument},
			},
			longOpts: map[string]*Flag{
				"output": {Name: "output", HasArg: RequiredArgument},
			},
			args:      []string{"-v", "--output=out.txt"},
			wantCalls: []string{"v", "output"},
		},
		{
			name: "some handled some not",
			shortOpts: map[byte]*Flag{
				'v': {Name: "v", HasArg: NoArgument},
				'd': {Name: "d", HasArg: NoArgument},
			},
			longOpts: map[string]*Flag{
				"output": {Name: "output", HasArg: RequiredArgument},
			},
			args:      []string{"-v", "-d", "--output=out.txt"},
			wantCalls: []string{"v", "output"},
			wantOpts:  []Option{{Name: "d", HasArg: false, Arg: ""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			handler := func(name string, arg string) error {
				calls = append(calls, name)
				return nil
			}

			// Set Handle on flags that should be handled
			for _, f := range tt.shortOpts {
				// Only set Handle on flags listed in wantCalls
				for _, wc := range tt.wantCalls {
					if f.Name == wc {
						f.Handle = handler
						break
					}
				}
			}
			for _, f := range tt.longOpts {
				for _, wc := range tt.wantCalls {
					if f.Name == wc {
						f.Handle = handler
						break
					}
				}
			}

			p, err := NewParser(cfg, tt.shortOpts, tt.longOpts, tt.args)
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}

			opts, errs := collectOptions(p)
			for _, e := range errs {
				if e != nil {
					t.Fatalf("unexpected error: %v", e)
				}
			}

			if len(calls) != len(tt.wantCalls) {
				t.Fatalf("handler calls = %v, want %v", calls, tt.wantCalls)
			}
			for i := range calls {
				if calls[i] != tt.wantCalls[i] {
					t.Fatalf("handler call[%d] = %q, want %q", i, calls[i], tt.wantCalls[i])
				}
			}

			wantOpts := tt.wantOpts
			if wantOpts == nil {
				wantOpts = []Option{}
			}
			if len(opts) != len(wantOpts) {
				t.Fatalf("yielded options = %v, want %v", opts, wantOpts)
			}
			for i := range opts {
				if opts[i] != wantOpts[i] {
					t.Fatalf("option[%d] = %v, want %v", i, opts[i], wantOpts[i])
				}
			}
		})
	}
}

func TestHandlerRegistrationEquivalence(t *testing.T) {
	// SetHandler variants and construction-time Handle produce
	// identical dispatch behavior for equivalent configurations.
	//
	// For each case, build two parsers with the same options and args:
	//   parserA: NewParser with Handle set at construction time
	//   parserB: GetOpt/GetOptLong + SetHandler post-construction
	// Verify both produce the same handler invocations and iterator output.

	tests := []struct {
		name string
		// buildA returns a parser with construction-time Handle
		buildA func(handler func(string, string) error) *Parser
		// buildB returns a parser with post-construction SetHandler
		buildB func(handler func(string, string) error) *Parser
		args   []string
	}{
		{
			name: "short NoArgument",
			args: []string{"-v"},
			buildA: func(h func(string, string) error) *Parser {
				p, _ := NewParser(
					ParserConfig{enableErrors: true},
					map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument, Handle: h}},
					nil,
					[]string{"-v"},
				)
				return p
			},
			buildB: func(h func(string, string) error) *Parser {
				p, _ := GetOpt([]string{"-v"}, "v")
				p.SetShortHandler('v', h)
				return p
			},
		},
		{
			name: "short RequiredArgument",
			args: []string{"-o", "file"},
			buildA: func(h func(string, string) error) *Parser {
				p, _ := NewParser(
					ParserConfig{enableErrors: true},
					map[byte]*Flag{'o': {Name: "o", HasArg: RequiredArgument, Handle: h}},
					nil,
					[]string{"-o", "file"},
				)
				return p
			},
			buildB: func(h func(string, string) error) *Parser {
				p, _ := GetOpt([]string{"-o", "file"}, "o:")
				p.SetShortHandler('o', h)
				return p
			},
		},
		{
			name: "long NoArgument",
			args: []string{"--verbose"},
			buildA: func(h func(string, string) error) *Parser {
				p, _ := NewParser(
					ParserConfig{enableErrors: true, longCaseIgnore: true},
					nil,
					map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument, Handle: h}},
					[]string{"--verbose"},
				)
				return p
			},
			buildB: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--verbose"}, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				p.SetLongHandler("verbose", h)
				return p
			},
		},
		{
			name: "long RequiredArgument with equals",
			args: []string{"--output=file.txt"},
			buildA: func(h func(string, string) error) *Parser {
				p, _ := NewParser(
					ParserConfig{enableErrors: true, longCaseIgnore: true},
					nil,
					map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument, Handle: h}},
					[]string{"--output=file.txt"},
				)
				return p
			},
			buildB: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--output=file.txt"}, "", []Flag{{Name: "output", HasArg: RequiredArgument}})
				p.SetLongHandler("output", h)
				return p
			},
		},
		{
			name: "mixed short and long via SetHandler convenience",
			buildA: func(h func(string, string) error) *Parser {
				p, _ := NewParser(
					ParserConfig{enableErrors: true, longCaseIgnore: true},
					map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument, Handle: h}},
					map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument, Handle: h}},
					[]string{"-v", "--output=out.txt"},
				)
				return p
			},
			buildB: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"-v", "--output=out.txt"}, "v", []Flag{{Name: "output", HasArg: RequiredArgument}})
				p.SetHandler("-v", h)
				p.SetHandler("--output", h)
				return p
			},
		},
		{
			name: "single-char long option via SetHandler",
			buildA: func(h func(string, string) error) *Parser {
				p, _ := NewParser(
					ParserConfig{enableErrors: true, longCaseIgnore: true},
					nil,
					map[string]*Flag{"v": {Name: "v", HasArg: NoArgument, Handle: h}},
					[]string{"--v"},
				)
				return p
			},
			buildB: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--v"}, "", []Flag{{Name: "v", HasArg: NoArgument}})
				p.SetHandler("--v", h)
				return p
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track handler invocations for parser A (construction-time)
			var callsA []handlerCall
			handlerA := func(name string, arg string) error {
				callsA = append(callsA, handlerCall{name, arg})
				return nil
			}

			// Track handler invocations for parser B (post-construction)
			var callsB []handlerCall
			handlerB := func(name string, arg string) error {
				callsB = append(callsB, handlerCall{name, arg})
				return nil
			}

			pA := tt.buildA(handlerA)
			pB := tt.buildB(handlerB)

			optsA, errsA := collectOptions(pA)
			optsB, errsB := collectOptions(pB)

			// Handler invocations must match
			if len(callsA) != len(callsB) {
				t.Fatalf("handler call count: A=%d, B=%d\nA calls: %v\nB calls: %v", len(callsA), len(callsB), callsA, callsB)
			}
			for i := range callsA {
				if callsA[i] != callsB[i] {
					t.Fatalf("handler call[%d]: A=%v, B=%v", i, callsA[i], callsB[i])
				}
			}

			// Iterator output must match
			if !optionsEqual(optsA, optsB) {
				t.Fatalf("options differ:\nA: %v\nB: %v", optsA, optsB)
			}
			if !errorsEqual(errsA, errsB) {
				t.Fatalf("errors differ:\nA: %v\nB: %v", errsA, errsB)
			}
		})
	}
}

// Feature: option-handlers, Property 2: Constructor nil Handle
func TestPropertyConstructorNilHandle(t *testing.T) {
	// Use only lowercase letters to avoid optstring prefix flags (+, :, -)
	// and special characters that complicate config equivalence.
	safeChars := []byte("abcdefghijklmnopqrstuvxyz") // exclude 'W' (gnuWords)

	// allNilHandle returns true if every Flag in the parser has nil Handle.
	allNilHandle := func(p *Parser, seed int64, ctor string) bool {
		for c, f := range p.shortOpts {
			if f.Handle != nil {
				t.Logf("seed=%d %s: shortOpt %q has non-nil Handle", seed, ctor, c)
				return false
			}
		}
		for name, f := range p.longOpts {
			if f.Handle != nil {
				t.Logf("seed=%d %s: longOpt %q has non-nil Handle", seed, ctor, name)
				return false
			}
		}
		return true
	}

	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Generate a random optstring with 1–6 short options.
		nShort := 1 + rng.Intn(6)
		perm := rng.Perm(len(safeChars))
		var optstring string
		for i := 0; i < nShort; i++ {
			c := safeChars[perm[i]]
			optstring += string(c)
			switch ArgType(rng.Intn(3)) {
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

		// GetOpt — short options only.
		if p, err := GetOpt([]string{"prog"}, optstring); err == nil {
			if !allNilHandle(p, seed, "GetOpt") {
				return false
			}
		}

		// GetOptLong — short + long options.
		if p, err := GetOptLong([]string{"prog"}, optstring, longFlags); err == nil {
			if !allNilHandle(p, seed, "GetOptLong") {
				return false
			}
		}

		// GetOptLongOnly — short + long options, long-only mode.
		if p, err := GetOptLongOnly([]string{"prog"}, optstring, longFlags); err == nil {
			if !allNilHandle(p, seed, "GetOptLongOnly") {
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 2 (Constructor nil Handle) failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Coverage: handler dispatch in getopt_long_only mode
// ---------------------------------------------------------------------------

func TestHandlerLongOnlyMatchedDispatch(t *testing.T) {
	// Handler invoked when longOptsOnly matches a long option via single dash.
	var called string
	handler := func(name string, arg string) error {
		called = name
		return nil
	}

	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument, Handle: handler},
	}
	cfg := ParserConfig{
		enableErrors:   true,
		longCaseIgnore: true,
		longOptsOnly:   true,
	}
	p, err := NewParser(cfg, nil, longOpts, []string{"-verbose"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if called != "verbose" {
		t.Fatalf("handler called with %q, want %q", called, "verbose")
	}
}

func TestHandlerLongOnlyMatchedErrorPropagation(t *testing.T) {
	// Handler error propagated when longOptsOnly matches a long option.
	sentinel := fmt.Errorf("long-only handler error")
	handler := func(name string, arg string) error {
		return sentinel
	}

	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument, Handle: handler},
	}
	cfg := ParserConfig{
		enableErrors:   true,
		longCaseIgnore: true,
		longOptsOnly:   true,
	}
	p, err := NewParser(cfg, nil, longOpts, []string{"-verbose"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	var sawError bool
	for opt, err := range p.Options() {
		if err != nil {
			if err.Error() != sentinel.Error() {
				t.Fatalf("unexpected error: %v", err)
			}
			if opt != (Option{}) {
				t.Fatalf("expected zero Option with error, got %+v", opt)
			}
			sawError = true
		}
	}
	if !sawError {
		t.Fatal("expected handler error, got none")
	}
}

func TestHandlerLongOnlyWithRequiredArg(t *testing.T) {
	// Handler receives argument in longOptsOnly mode.
	var receivedName, receivedArg string
	handler := func(name string, arg string) error {
		receivedName = name
		receivedArg = arg
		return nil
	}

	longOpts := map[string]*Flag{
		"output": {Name: "output", HasArg: RequiredArgument, Handle: handler},
	}
	cfg := ParserConfig{
		enableErrors:   true,
		longCaseIgnore: true,
		longOptsOnly:   true,
	}
	p, err := NewParser(cfg, nil, longOpts, []string{"-output=file.txt"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	opts, errs := collectOptions(p)
	for _, e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if len(opts) != 0 {
		t.Fatalf("expected no options yielded, got %d", len(opts))
	}
	if receivedName != "output" || receivedArg != "file.txt" {
		t.Fatalf("handler got (%q, %q), want (%q, %q)", receivedName, receivedArg, "output", "file.txt")
	}
}

// ---------------------------------------------------------------------------
// Coverage: early iterator break during handler error propagation
// ---------------------------------------------------------------------------

func TestHandlerLongOptionErrorBreak(t *testing.T) {
	// Consumer breaks iteration after receiving a handler error on a long option.
	sentinel := fmt.Errorf("stop here")
	handler := func(name string, arg string) error {
		return sentinel
	}

	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument, Handle: handler},
		"debug":   {Name: "debug", HasArg: NoArgument, Handle: handler},
	}
	cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
	p, err := NewParser(cfg, nil, longOpts, []string{"--verbose", "--debug"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	count := 0
	for _, err := range p.Options() {
		if err != nil {
			count++
			break // consumer stops after first error
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 error before break, got %d", count)
	}
}

func TestHandlerLongOptionYieldBreak(t *testing.T) {
	// Consumer breaks iteration after receiving a yielded long option.
	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
		"debug":   {Name: "debug", HasArg: NoArgument},
	}
	cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
	p, err := NewParser(cfg, nil, longOpts, []string{"--verbose", "--debug"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	count := 0
	for _, err := range p.Options() {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		count++
		break // consumer stops after first option
	}
	if count != 1 {
		t.Fatalf("expected 1 option before break, got %d", count)
	}
}

func TestHandlerLongOnlyErrorBreak(t *testing.T) {
	// Consumer breaks iteration after handler error in longOptsOnly mode.
	sentinel := fmt.Errorf("long-only stop")
	handler := func(name string, arg string) error {
		return sentinel
	}

	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument, Handle: handler},
		"debug":   {Name: "debug", HasArg: NoArgument, Handle: handler},
	}
	cfg := ParserConfig{
		enableErrors:   true,
		longCaseIgnore: true,
		longOptsOnly:   true,
	}
	p, err := NewParser(cfg, nil, longOpts, []string{"-verbose", "-debug"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	count := 0
	for _, err := range p.Options() {
		if err != nil {
			count++
			break
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 error before break, got %d", count)
	}
}

func TestHandlerCompactionErrorBreak(t *testing.T) {
	// Consumer breaks iteration after handler error during compaction.
	sentinel := fmt.Errorf("compaction stop")
	shortOpts := map[byte]*Flag{
		'a': {Name: "a", HasArg: NoArgument, Handle: func(name, arg string) error {
			return sentinel
		}},
		'b': {Name: "b", HasArg: NoArgument},
	}
	cfg := ParserConfig{enableErrors: true}
	p, err := NewParser(cfg, shortOpts, nil, []string{"-ab", "-ab"})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	count := 0
	for _, err := range p.Options() {
		if err != nil {
			count++
			break
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 error before break, got %d", count)
	}
}

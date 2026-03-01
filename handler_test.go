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
		type call struct{ name, arg string }
		var shortCalls []call
		shortHandler := func(name string, arg string) error {
			shortCalls = append(shortCalls, call{name, arg})
			return nil
		}
		for _, c := range chars {
			if err := p.SetShortHandler(c, shortHandler); err != nil {
				t.Logf("seed=%d SetShortHandler(%c) unexpected error: %v", seed, c, err)
				return false
			}
		}

		// --- Test SetLongHandler ---
		var longCalls []call
		longHandler := func(name string, arg string) error {
			longCalls = append(longCalls, call{name, arg})
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

		var setHandlerCalls []call
		setHandlerFn := func(name string, arg string) error {
			setHandlerCalls = append(setHandlerCalls, call{name, arg})
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
		type call struct{ name, arg string }
		var calls []call
		handler := func(name string, arg string) error {
			calls = append(calls, call{name, arg})
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

		type call struct{ name, arg string }
		var parentCalls []call
		var childCalls []call
		parentHandler := func(name string, arg string) error {
			parentCalls = append(parentCalls, call{name, arg})
			return nil
		}
		childHandler := func(name string, arg string) error {
			childCalls = append(childCalls, call{name, arg})
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

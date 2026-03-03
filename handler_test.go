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

// TestHandlerSuppressesYield verifies that flags with non-nil Handle are
// dispatched to the handler (not yielded), while flags with nil Handle are
// yielded as Options. Covers short/long, all argument types, and the three
// partitions: handled-only, non-handled-only, mixed.
func TestHandlerSuppressesYield(t *testing.T) {
	tests := []struct {
		name        string
		short       map[byte]*Flag
		long        map[string]*Flag
		args        []string
		wantHandled []string // names dispatched to handler
		wantYielded []string // names yielded as Options
	}{
		{
			name:        "short NoArg handled-only",
			short:       map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}},
			args:        []string{"-v"},
			wantHandled: []string{"v"},
		},
		{
			name:        "short RequiredArg non-handled-only",
			short:       map[byte]*Flag{'o': {Name: "o", HasArg: RequiredArgument}},
			args:        []string{"-o", "file"},
			wantYielded: []string{"o"},
		},
		{
			name:  "short mixed NoArg handled + non-handled",
			short: map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}, 'x': {Name: "x", HasArg: NoArgument}},
			args:  []string{"-v", "-x"},
			// v handled, x not
			wantHandled: []string{"v"},
			wantYielded: []string{"x"},
		},
		{
			name:        "long RequiredArg handled",
			long:        map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument}},
			args:        []string{"--output=file"},
			wantHandled: []string{"output"},
		},
		{
			name:        "long OptionalArg non-handled",
			long:        map[string]*Flag{"debug": {Name: "debug", HasArg: OptionalArgument}},
			args:        []string{"--debug=3"},
			wantYielded: []string{"debug"},
		},
		{
			name:  "mixed short+long handled and non-handled",
			short: map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}, 'x': {Name: "x", HasArg: NoArgument}},
			long:  map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument}, "debug": {Name: "debug", HasArg: NoArgument}},
			args:  []string{"-v", "--output=f", "-x", "--debug"},
			// v, output handled; x, debug not
			wantHandled: []string{"v", "output"},
			wantYielded: []string{"x", "debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []handlerCall
			handler := func(name string, arg string) error {
				calls = append(calls, handlerCall{name, arg})
				return nil
			}

			// Assign handler to flags listed in wantHandled.
			handledSet := make(map[string]bool)
			for _, n := range tt.wantHandled {
				handledSet[n] = true
			}
			for c, f := range tt.short {
				if handledSet[f.Name] {
					tt.short[c].Handle = handler
				}
			}
			for n, f := range tt.long {
				if handledSet[f.Name] {
					tt.long[n].Handle = handler
				}
			}

			p, err := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, tt.short, tt.long, tt.args)
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}

			var yielded []string
			for opt, err := range p.Options() {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				yielded = append(yielded, opt.Name)
			}

			// Verify handler calls.
			var gotHandled []string
			for _, c := range calls {
				gotHandled = append(gotHandled, c.name)
			}
			if fmt.Sprint(gotHandled) != fmt.Sprint(tt.wantHandled) {
				t.Errorf("handled: got %v, want %v", gotHandled, tt.wantHandled)
			}
			if fmt.Sprint(yielded) != fmt.Sprint(tt.wantYielded) {
				t.Errorf("yielded: got %v, want %v", yielded, tt.wantYielded)
			}
		})
	}
}

// TestHandlerReceivesCorrectNameAndArg verifies that handler receives the same
// name and arg that would have been yielded as Option.Name and Option.Arg.
// Covers: RequiredArgument long separate (not in TestHandlerArgumentTypes),
// and multi-option parse (property test's additional dimension).
func TestHandlerReceivesCorrectNameAndArg(t *testing.T) {
	tests := []struct {
		name      string
		short     map[byte]*Flag
		long      map[string]*Flag
		args      []string
		wantCalls []handlerCall
	}{
		{
			name:      "RequiredArgument long separate arg",
			long:      map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument}},
			args:      []string{"--output", "result.txt"},
			wantCalls: []handlerCall{{"output", "result.txt"}},
		},
		{
			name: "multiple short and long options in one parse",
			short: map[byte]*Flag{
				'v': {Name: "v", HasArg: NoArgument},
				'o': {Name: "o", HasArg: RequiredArgument},
			},
			long: map[string]*Flag{
				"debug": {Name: "debug", HasArg: OptionalArgument},
			},
			args:      []string{"-v", "-o", "file", "--debug=3"},
			wantCalls: []handlerCall{{"v", ""}, {"o", "file"}, {"debug", "3"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []handlerCall
			handler := func(name string, arg string) error {
				got = append(got, handlerCall{name, arg})
				return nil
			}
			for _, f := range tt.short {
				f.Handle = handler
			}
			for _, f := range tt.long {
				f.Handle = handler
			}
			p, err := NewParser(
				ParserConfig{enableErrors: true, longCaseIgnore: true},
				tt.short, tt.long, tt.args,
			)
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
			if len(got) != len(tt.wantCalls) {
				t.Fatalf("handler calls: got %d, want %d", len(got), len(tt.wantCalls))
			}
			for i, want := range tt.wantCalls {
				if got[i].name != want.name || got[i].arg != want.arg {
					t.Fatalf("[%d] handler got (%q, %q), want (%q, %q)",
						i, got[i].name, got[i].arg, want.name, want.arg)
				}
			}
		})
	}
}

// TestHandlerErrorPropagation verifies that a handler returning non-nil error
// causes the iterator to yield (zero Option, that error). Remaining non-error
// options continue to yield normally.
func TestHandlerErrorPropagation(t *testing.T) {
	tests := []struct {
		name       string
		errMsg     string
		errChar    byte   // which short option returns the error
		otherChars []byte // non-error options (yielded normally)
	}{
		{
			name:       "error on first of two options",
			errMsg:     "fail at a",
			errChar:    'a',
			otherChars: []byte{'b'},
		},
		{
			name:       "error on middle of three options",
			errMsg:     "middle broke",
			errChar:    'b',
			otherChars: []byte{'a', 'c'},
		},
		{
			name:       "error on last of three options",
			errMsg:     "last option error",
			errChar:    'c',
			otherChars: []byte{'a', 'b'},
		},
		{
			name:       "error on sole option",
			errMsg:     "only handler fails",
			errChar:    'x',
			otherChars: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentinel := fmt.Errorf("%s", tt.errMsg)
			shortMap := map[byte]*Flag{
				tt.errChar: {
					Name:   string(tt.errChar),
					HasArg: NoArgument,
					Handle: func(string, string) error { return sentinel },
				},
			}
			// Error option appears first in args.
			args := []string{"-" + string(tt.errChar)}
			for _, c := range tt.otherChars {
				shortMap[c] = &Flag{Name: string(c), HasArg: NoArgument}
				args = append(args, "-"+string(c))
			}

			p, err := NewParser(ParserConfig{enableErrors: true}, shortMap, nil, args)
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}

			opts, errs := collectOptions(p)
			// First yield: (zero Option, sentinel error).
			if len(errs) == 0 || errs[0] == nil {
				t.Fatal("expected error on first yield, got nil")
			}
			if errs[0].Error() != tt.errMsg {
				t.Fatalf("error mismatch: got %q, want %q", errs[0].Error(), tt.errMsg)
			}
			if opts[0] != (Option{}) {
				t.Fatalf("expected zero Option with error, got %+v", opts[0])
			}
			// Remaining yields: normal options, no errors.
			for i := 1; i < len(errs); i++ {
				if errs[i] != nil {
					t.Fatalf("unexpected error at yield %d: %v", i, errs[i])
				}
			}
		})
	}
}

// TestMixedHandledAndNonHandledDispatch verifies that a parser with both
// handled and non-handled flags dispatches handlers for handled flags and
// yields Options for non-handled flags within the same iteration, with no
// cross-contamination. Covers short+long, all argument types, and the
// handled-only / non-handled-only / mixed partitions.
//
// Note: the basic short-only mixed case is already covered by
// TestHandlerMixedHandledAndNonHandled. These rows extend coverage to long
// options, different argument types, and the all-handled / all-yielded edges.
func TestMixedHandledAndNonHandledDispatch(t *testing.T) {
	tests := []struct {
		name        string
		short       map[byte]*Flag
		long        map[string]*Flag
		args        []string
		wantHandled []string
		wantYielded []string
	}{
		{
			name:        "all handled short+long",
			short:       map[byte]*Flag{'a': {Name: "a", HasArg: NoArgument}},
			long:        map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}},
			args:        []string{"-a", "--verbose"},
			wantHandled: []string{"a", "verbose"},
		},
		{
			name:        "all non-handled short+long",
			short:       map[byte]*Flag{'a': {Name: "a", HasArg: NoArgument}},
			long:        map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}},
			args:        []string{"-a", "--verbose"},
			wantYielded: []string{"a", "verbose"},
		},
		{
			name:  "mixed RequiredArg short handled + long non-handled",
			short: map[byte]*Flag{'o': {Name: "o", HasArg: RequiredArgument}},
			long:  map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument}},
			args:  []string{"-o", "a.txt", "--output=b.txt"},
			// o handled, output not
			wantHandled: []string{"o"},
			wantYielded: []string{"output"},
		},
		{
			name:  "mixed OptionalArg long handled + short non-handled",
			short: map[byte]*Flag{'d': {Name: "d", HasArg: NoArgument}},
			long:  map[string]*Flag{"debug": {Name: "debug", HasArg: OptionalArgument}},
			args:  []string{"--debug=3", "-d"},
			// debug handled, d not
			wantHandled: []string{"debug"},
			wantYielded: []string{"d"},
		},
		{
			name: "three short two long mixed",
			short: map[byte]*Flag{
				'a': {Name: "a", HasArg: NoArgument},
				'b': {Name: "b", HasArg: NoArgument},
				'c': {Name: "c", HasArg: RequiredArgument},
			},
			long: map[string]*Flag{
				"verbose": {Name: "verbose", HasArg: NoArgument},
				"output":  {Name: "output", HasArg: RequiredArgument},
			},
			args: []string{"-a", "-b", "-c", "val", "--verbose", "--output=f"},
			// a, c, verbose handled; b, output not
			wantHandled: []string{"a", "c", "verbose"},
			wantYielded: []string{"b", "output"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []handlerCall
			handler := func(name string, arg string) error {
				calls = append(calls, handlerCall{name, arg})
				return nil
			}

			handledSet := make(map[string]bool)
			for _, n := range tt.wantHandled {
				handledSet[n] = true
			}
			for c, f := range tt.short {
				if handledSet[f.Name] {
					tt.short[c].Handle = handler
				}
			}
			for n, f := range tt.long {
				if handledSet[f.Name] {
					tt.long[n].Handle = handler
				}
			}

			p, err := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, tt.short, tt.long, tt.args)
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}

			var yielded []string
			for opt, err := range p.Options() {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				yielded = append(yielded, opt.Name)
			}

			var gotHandled []string
			for _, c := range calls {
				gotHandled = append(gotHandled, c.name)
			}
			if fmt.Sprint(gotHandled) != fmt.Sprint(tt.wantHandled) {
				t.Errorf("handled: got %v, want %v", gotHandled, tt.wantHandled)
			}
			if fmt.Sprint(yielded) != fmt.Sprint(tt.wantYielded) {
				t.Errorf("yielded: got %v, want %v", yielded, tt.wantYielded)
			}
		})
	}
}

// SetHandler name matching and reject-unknown behaviors are covered by:
//   - TestHandlerSetShortHandlerOptstring (short name matching)
//   - TestHandlerSetLongHandler (long name matching)
//   - TestHandlerSetHandlerDelegation (dash-prefix delegation)
//   - TestHandlerSetHandlerRejectUnregistered (unknown names)
//   - TestHandlerSetHandlerNoDashPrefix (no-prefix rejection)

// SetHandler no-parent-walk behavior is covered by:
//   - TestHandlerSetHandlerNoParentWalk (all Set*Handler variants on child for parent-only options)

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

// ---------------------------------------------------------------------------
// Unit tests: handler dispatch (short/long, current/ancestor)
// ---------------------------------------------------------------------------

func TestHandlerDispatch(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(handler func(string, string) error) *Parser
		wantName string
	}{
		{
			name: "short option current parser",
			setup: func(h func(string, string) error) *Parser {
				p, _ := NewParser(
					ParserConfig{enableErrors: true},
					map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument, Handle: h}},
					nil, []string{"-v"},
				)
				return p
			},
			wantName: "v",
		},
		{
			name: "short option ancestor",
			setup: func(h func(string, string) error) *Parser {
				cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
				parent, _ := NewParser(cfg,
					map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument, Handle: h}},
					nil, nil,
				)
				child, _ := NewParser(cfg, nil, nil, []string{"-v"})
				parent.AddCmd("sub", child)
				return child
			},
			wantName: "v",
		},
		{
			name: "long option current parser",
			setup: func(h func(string, string) error) *Parser {
				p, _ := NewParser(
					ParserConfig{enableErrors: true, longCaseIgnore: true},
					nil,
					map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument, Handle: h}},
					[]string{"--verbose"},
				)
				return p
			},
			wantName: "verbose",
		},
		{
			name: "long option ancestor",
			setup: func(h func(string, string) error) *Parser {
				cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
				parent, _ := NewParser(cfg, nil,
					map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument, Handle: h}},
					nil,
				)
				child, _ := NewParser(cfg, nil, nil, []string{"--verbose"})
				parent.AddCmd("sub", child)
				return child
			},
			wantName: "verbose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called string
			handler := func(name string, arg string) error {
				called = name
				return nil
			}
			p := tt.setup(handler)
			opts, errs := collectOptions(p)
			for _, e := range errs {
				if e != nil {
					t.Fatalf("unexpected error: %v", e)
				}
			}
			if len(opts) != 0 {
				t.Fatalf("expected no options yielded, got %d", len(opts))
			}
			if called != tt.wantName {
				t.Fatalf("handler called with %q, want %q", called, tt.wantName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unit tests: handler argument types
// ---------------------------------------------------------------------------

func TestHandlerArgumentTypes(t *testing.T) {
	tests := []struct {
		name     string
		short    map[byte]*Flag
		long     map[string]*Flag
		args     []string
		wantName string
		wantArg  string
	}{
		{
			name:     "NoArgument short receives empty arg",
			short:    map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}},
			args:     []string{"-v"},
			wantName: "v",
			wantArg:  "",
		},
		{
			name:     "NoArgument long receives empty arg",
			long:     map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}},
			args:     []string{"--verbose"},
			wantName: "verbose",
			wantArg:  "",
		},
		{
			name:     "RequiredArgument short separate arg",
			short:    map[byte]*Flag{'o': {Name: "o", HasArg: RequiredArgument}},
			args:     []string{"-o", "file.txt"},
			wantName: "o",
			wantArg:  "file.txt",
		},
		{
			name:     "RequiredArgument long with equals",
			long:     map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument}},
			args:     []string{"--output=result.txt"},
			wantName: "output",
			wantArg:  "result.txt",
		},
		{
			name:     "OptionalArgument short with arg attached",
			short:    map[byte]*Flag{'d': {Name: "d", HasArg: OptionalArgument}},
			args:     []string{"-dlevel3"},
			wantName: "d",
			wantArg:  "level3",
		},
		{
			name:     "OptionalArgument short without arg",
			short:    map[byte]*Flag{'d': {Name: "d", HasArg: OptionalArgument}},
			args:     []string{"-d"},
			wantName: "d",
			wantArg:  "",
		},
		{
			name:     "OptionalArgument long with equals",
			long:     map[string]*Flag{"debug": {Name: "debug", HasArg: OptionalArgument}},
			args:     []string{"--debug=trace"},
			wantName: "debug",
			wantArg:  "trace",
		},
		{
			name:     "OptionalArgument long without arg",
			long:     map[string]*Flag{"debug": {Name: "debug", HasArg: OptionalArgument}},
			args:     []string{"--debug"},
			wantName: "debug",
			wantArg:  "",
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
			// Set handler on all flags
			for _, f := range tt.short {
				f.Handle = handler
			}
			for _, f := range tt.long {
				f.Handle = handler
			}
			p, err := NewParser(
				ParserConfig{enableErrors: true, longCaseIgnore: true},
				tt.short, tt.long, tt.args,
			)
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
			if receivedName != tt.wantName || receivedArg != tt.wantArg {
				t.Fatalf("handler got (%q, %q), want (%q, %q)",
					receivedName, receivedArg, tt.wantName, tt.wantArg)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unit tests: child overload, mixed dispatch, nil handle
// ---------------------------------------------------------------------------

func TestHandlerChildOverloadShadowsParent(t *testing.T) {
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
	parent, _ := NewParser(cfg,
		map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument, Handle: parentHandler}},
		nil, nil,
	)
	child, _ := NewParser(cfg,
		map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument, Handle: childHandler}},
		nil, []string{"-v"},
	)
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
	var handledName string
	handler := func(name string, arg string) error {
		handledName = name
		return nil
	}

	shortOpts := map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument, Handle: handler},
		'x': {Name: "x", HasArg: NoArgument},
	}
	p, _ := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-v", "-x"})

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

func TestHandlerNilHandleIdenticalOutput(t *testing.T) {
	shortOpts := map[byte]*Flag{
		'v': {Name: "v", HasArg: NoArgument},
		'o': {Name: "o", HasArg: RequiredArgument},
	}
	longOpts := map[string]*Flag{
		"verbose": {Name: "verbose", HasArg: NoArgument},
		"output":  {Name: "output", HasArg: RequiredArgument},
	}
	args := []string{"-v", "-o", "file.txt", "--verbose", "--output=result.txt"}

	p, _ := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, shortOpts, longOpts, args)
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

// ---------------------------------------------------------------------------
// Unit tests: compaction handler behavior
// ---------------------------------------------------------------------------

func TestHandlerCompaction(t *testing.T) {
	t.Run("all handled", func(t *testing.T) {
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
		p, _ := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-abc"})
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
	})

	t.Run("mixed handled and non-handled", func(t *testing.T) {
		var handledNames []string
		handler := func(name string, arg string) error {
			handledNames = append(handledNames, name)
			return nil
		}
		shortOpts := map[byte]*Flag{
			'a': {Name: "a", HasArg: NoArgument, Handle: handler},
			'b': {Name: "b", HasArg: NoArgument},
			'c': {Name: "c", HasArg: NoArgument, Handle: handler},
		}
		p, _ := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-abc"})
		opts, errs := collectOptions(p)
		for _, e := range errs {
			if e != nil {
				t.Fatalf("unexpected error: %v", e)
			}
		}
		expected := []Option{{Name: "b", HasArg: false, Arg: ""}}
		if !optionsEqual(opts, expected) {
			t.Fatalf("options: got %v, want %v", opts, expected)
		}
		if len(handledNames) != 2 || handledNames[0] != "a" || handledNames[1] != "c" {
			t.Fatalf("handlers called %v, want [a c]", handledNames)
		}
	})

	t.Run("error stops remaining", func(t *testing.T) {
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
		p, _ := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-abc"})
		opts, errs := collectOptions(p)
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
		if len(calls) != 2 || calls[0] != "a" || calls[1] != "b" {
			t.Fatalf("handlers called %v, want [a b]", calls)
		}
	})
}

// ---------------------------------------------------------------------------
// Unit tests: SetHandler delegation and single-char long options
// ---------------------------------------------------------------------------

func TestHandlerSetHandlerDelegation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(handler func(string, string) error) *Parser
		wantName string
	}{
		{
			name: "SetHandler delegates to long",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--verbose"}, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				p.SetHandler("--verbose", h)
				return p
			},
			wantName: "verbose",
		},
		{
			name: "SetHandler delegates to short",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOpt([]string{"-v"}, "v")
				p.SetHandler("-v", h)
				return p
			},
			wantName: "v",
		},
		{
			name: "SetHandler single-char long option",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--v"}, "", []Flag{{Name: "v", HasArg: NoArgument}})
				p.SetHandler("--v", h)
				return p
			},
			wantName: "v",
		},
		{
			name: "SetLongHandler single-char name",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--v"}, "", []Flag{{Name: "v", HasArg: NoArgument}})
				p.SetLongHandler("v", h)
				return p
			},
			wantName: "v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedName string
			handler := func(name string, arg string) error {
				receivedName = name
				return nil
			}
			p := tt.setup(handler)
			opts, errs := collectOptions(p)
			for _, e := range errs {
				if e != nil {
					t.Fatalf("unexpected error: %v", e)
				}
			}
			if len(opts) != 0 {
				t.Fatalf("expected no options yielded, got %d", len(opts))
			}
			if receivedName != tt.wantName {
				t.Fatalf("handler called with %q, want %q", receivedName, tt.wantName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unit tests: long-only mode handler dispatch
// ---------------------------------------------------------------------------

func TestHandlerLongOnly(t *testing.T) {
	t.Run("matched dispatch", func(t *testing.T) {
		var called string
		handler := func(name string, arg string) error {
			called = name
			return nil
		}
		longOpts := map[string]*Flag{
			"verbose": {Name: "verbose", HasArg: NoArgument, Handle: handler},
		}
		cfg := ParserConfig{enableErrors: true, longCaseIgnore: true, longOptsOnly: true}
		p, _ := NewParser(cfg, nil, longOpts, []string{"-verbose"})
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
	})

	t.Run("error propagation", func(t *testing.T) {
		sentinel := fmt.Errorf("long-only handler error")
		handler := func(name string, arg string) error { return sentinel }
		longOpts := map[string]*Flag{
			"verbose": {Name: "verbose", HasArg: NoArgument, Handle: handler},
		}
		cfg := ParserConfig{enableErrors: true, longCaseIgnore: true, longOptsOnly: true}
		p, _ := NewParser(cfg, nil, longOpts, []string{"-verbose"})
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
	})

	t.Run("with required arg", func(t *testing.T) {
		var receivedName, receivedArg string
		handler := func(name string, arg string) error {
			receivedName = name
			receivedArg = arg
			return nil
		}
		longOpts := map[string]*Flag{
			"output": {Name: "output", HasArg: RequiredArgument, Handle: handler},
		}
		cfg := ParserConfig{enableErrors: true, longCaseIgnore: true, longOptsOnly: true}
		p, _ := NewParser(cfg, nil, longOpts, []string{"-output=file.txt"})
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
			t.Fatalf("handler got (%q, %q), want (%q, %q)",
				receivedName, receivedArg, "output", "file.txt")
		}
	})
}

// ---------------------------------------------------------------------------
// Unit tests: early iterator break during handler error propagation
// ---------------------------------------------------------------------------

func TestHandlerIteratorBreak(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Parser
	}{
		{
			name: "long option error break",
			setup: func() *Parser {
				sentinel := fmt.Errorf("stop here")
				longOpts := map[string]*Flag{
					"verbose": {Name: "verbose", HasArg: NoArgument, Handle: func(string, string) error { return sentinel }},
					"debug":   {Name: "debug", HasArg: NoArgument, Handle: func(string, string) error { return sentinel }},
				}
				p, _ := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, nil, longOpts, []string{"--verbose", "--debug"})
				return p
			},
		},
		{
			name: "long-only error break",
			setup: func() *Parser {
				sentinel := fmt.Errorf("long-only stop")
				longOpts := map[string]*Flag{
					"verbose": {Name: "verbose", HasArg: NoArgument, Handle: func(string, string) error { return sentinel }},
					"debug":   {Name: "debug", HasArg: NoArgument, Handle: func(string, string) error { return sentinel }},
				}
				cfg := ParserConfig{enableErrors: true, longCaseIgnore: true, longOptsOnly: true}
				p, _ := NewParser(cfg, nil, longOpts, []string{"-verbose", "-debug"})
				return p
			},
		},
		{
			name: "compaction error break",
			setup: func() *Parser {
				sentinel := fmt.Errorf("compaction stop")
				shortOpts := map[byte]*Flag{
					'a': {Name: "a", HasArg: NoArgument, Handle: func(string, string) error { return sentinel }},
					'b': {Name: "b", HasArg: NoArgument},
				}
				p, _ := NewParser(ParserConfig{enableErrors: true}, shortOpts, nil, []string{"-ab", "-ab"})
				return p
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.setup()
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
		})
	}

	t.Run("long option yield break", func(t *testing.T) {
		longOpts := map[string]*Flag{
			"verbose": {Name: "verbose", HasArg: NoArgument},
			"debug":   {Name: "debug", HasArg: NoArgument},
		}
		p, _ := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, nil, longOpts, []string{"--verbose", "--debug"})
		count := 0
		for _, err := range p.Options() {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			count++
			break
		}
		if count != 1 {
			t.Fatalf("expected 1 option before break, got %d", count)
		}
	})
}

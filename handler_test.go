package optargs

import (
	"fmt"
	"math/rand"
	"testing"
	"testing/quick"
)

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

// Feature: test-refactor, Property 1: Coverage Floor Invariant
//
// Invariant: GetOptLong output ≡ NewParser with nil Handle for the same
// configuration. For any parser configuration and argument list where all
// Flags have nil Handle, iterator output is identical to the pre-handler
// implementation.
//
// Why randomized inputs: this is an equivalence property over the full
// optstring × args space — random configs catch corner cases that finite
// examples miss.
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
		{
			name:        "all handled short+long (boundary)",
			short:       map[byte]*Flag{'a': {Name: "a", HasArg: NoArgument}},
			long:        map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}},
			args:        []string{"-a", "--verbose"},
			wantHandled: []string{"a", "verbose"},
		},
		{
			name:        "all non-handled short+long (boundary)",
			short:       map[byte]*Flag{'a': {Name: "a", HasArg: NoArgument}},
			long:        map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}},
			args:        []string{"-a", "--verbose"},
			wantYielded: []string{"a", "verbose"},
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

// Child overloading wins: child's definition determines dispatch regardless of
// parent handler status.

func TestChildOverloadingWins(t *testing.T) {
	tests := []struct {
		name            string
		childShortFlags map[byte]bool   // true = child has handler, false = nil Handle
		childLongFlags  map[string]bool // true = child has handler, false = nil Handle
		wantChildCalls  []string        // names dispatched to child handler
		wantYielded     []string        // names yielded as Options (nil Handle)
	}{
		{
			name:            "short child-handled overrides parent-handled",
			childShortFlags: map[byte]bool{'v': true},
			wantChildCalls:  []string{"v"},
		},
		{
			name:            "short child-nil-handle yields Option despite parent handler",
			childShortFlags: map[byte]bool{'v': false},
			wantYielded:     []string{"v"},
		},
		{
			name:           "long child-handled overrides parent-handled",
			childLongFlags: map[string]bool{"verbose": true},
			wantChildCalls: []string{"verbose"},
		},
		{
			name:           "long child-nil-handle yields Option despite parent handler",
			childLongFlags: map[string]bool{"verbose": false},
			wantYielded:    []string{"verbose"},
		},
		{
			name:            "mixed short+long handled and non-handled",
			childShortFlags: map[byte]bool{'v': true, 'x': false},
			childLongFlags:  map[string]bool{"output": true, "debug": false},
			wantChildCalls:  []string{"v", "output"},
			wantYielded:     []string{"x", "debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parentCalls, childCalls []handlerCall
			parentHandler := func(name string, arg string) error {
				parentCalls = append(parentCalls, handlerCall{name, arg})
				return nil
			}
			childHandler := func(name string, arg string) error {
				childCalls = append(childCalls, handlerCall{name, arg})
				return nil
			}

			// Parent defines all overlapping options with handlers.
			parentShort := make(map[byte]*Flag)
			for c := range tt.childShortFlags {
				parentShort[c] = &Flag{Name: string(c), HasArg: NoArgument, Handle: parentHandler}
			}
			parentLong := make(map[string]*Flag)
			for n := range tt.childLongFlags {
				parentLong[n] = &Flag{Name: n, HasArg: NoArgument, Handle: parentHandler}
			}

			cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
			parent, err := NewParser(cfg, parentShort, parentLong, nil)
			if err != nil {
				t.Fatalf("NewParser parent: %v", err)
			}

			// Child overloads all parent options.
			childShort := make(map[byte]*Flag)
			childLong := make(map[string]*Flag)
			var args []string
			for c, handled := range tt.childShortFlags {
				f := &Flag{Name: string(c), HasArg: NoArgument}
				if handled {
					f.Handle = childHandler
				}
				childShort[c] = f
				args = append(args, "-"+string(c))
			}
			for n, handled := range tt.childLongFlags {
				f := &Flag{Name: n, HasArg: NoArgument}
				if handled {
					f.Handle = childHandler
				}
				childLong[n] = f
				args = append(args, "--"+n)
			}

			child, err := NewParser(cfg, childShort, childLong, args)
			if err != nil {
				t.Fatalf("NewParser child: %v", err)
			}
			parent.AddCmd("sub", child)

			var yielded []string
			for opt, err := range child.Options() {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				yielded = append(yielded, opt.Name)
			}

			// Parent handler must never be invoked.
			if len(parentCalls) != 0 {
				t.Errorf("parent handler invoked %d times, want 0", len(parentCalls))
			}

			var gotChildCalls []string
			for _, c := range childCalls {
				gotChildCalls = append(gotChildCalls, c.name)
			}
			if fmt.Sprint(gotChildCalls) != fmt.Sprint(tt.wantChildCalls) {
				t.Errorf("child handler calls: got %v, want %v", gotChildCalls, tt.wantChildCalls)
			}
			if fmt.Sprint(yielded) != fmt.Sprint(tt.wantYielded) {
				t.Errorf("yielded: got %v, want %v", yielded, tt.wantYielded)
			}
		})
	}
}

func TestCompactionHandlerDispatchOrder(t *testing.T) {
	tests := []struct {
		name        string
		chars       []byte // short option chars in compaction order
		handled     []bool // true = has handler, false = yields Option
		wantCalls   []string
		wantYielded []string
	}{
		{
			name:        "4 chars alternating handled/non-handled",
			chars:       []byte{'a', 'b', 'c', 'd'},
			handled:     []bool{true, false, true, false},
			wantCalls:   []string{"a", "c"},
			wantYielded: []string{"b", "d"},
		},
		{
			name:        "5 chars first and last handled",
			chars:       []byte{'a', 'b', 'c', 'd', 'e'},
			handled:     []bool{true, false, false, false, true},
			wantCalls:   []string{"a", "e"},
			wantYielded: []string{"b", "c", "d"},
		},
		{
			name:        "4 chars all handled",
			chars:       []byte{'x', 'y', 'z', 'q'},
			handled:     []bool{true, true, true, true},
			wantCalls:   []string{"x", "y", "z", "q"},
			wantYielded: nil,
		},
		{
			name:        "3 chars none handled",
			chars:       []byte{'a', 'b', 'c'},
			handled:     []bool{false, false, false},
			wantCalls:   nil,
			wantYielded: []string{"a", "b", "c"},
		},
		{
			name:        "5 chars only middle handled",
			chars:       []byte{'a', 'b', 'c', 'd', 'e'},
			handled:     []bool{false, false, true, false, false},
			wantCalls:   []string{"c"},
			wantYielded: []string{"a", "b", "d", "e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			shortMap := make(map[byte]*Flag)
			for i, c := range tt.chars {
				f := &Flag{Name: string(c), HasArg: NoArgument}
				if tt.handled[i] {
					f.Handle = func(name string, arg string) error {
						calls = append(calls, name)
						return nil
					}
				}
				shortMap[c] = f
			}

			compacted := "-"
			for _, c := range tt.chars {
				compacted += string(c)
			}

			p, err := NewParser(ParserConfig{enableErrors: true}, shortMap, nil, []string{compacted})
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

			if fmt.Sprintf("%v", calls) != fmt.Sprintf("%v", tt.wantCalls) {
				t.Errorf("handler calls: got %v, want %v", calls, tt.wantCalls)
			}
			if fmt.Sprintf("%v", yielded) != fmt.Sprintf("%v", tt.wantYielded) {
				t.Errorf("yielded options: got %v, want %v", yielded, tt.wantYielded)
			}
		})
	}
}

func TestSetHandler(t *testing.T) {
	handler := func(string, string) error { return nil }

	// --- Dispatch tests: handler is called with correct name/arg, no options yielded ---
	dispatchTests := []struct {
		name     string
		setup    func(h func(string, string) error) *Parser
		wantName string
		wantArg  string
	}{
		// Short handler via optstring (from TestHandlerSetShortHandlerOptstring)
		{
			name: "short_no_arg",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOpt([]string{"-v"}, "vx")
				p.SetShortHandler('v', h)
				return p
			},
			wantName: "v",
		},
		{
			name: "short_required_arg",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOpt([]string{"-o", "file.txt"}, "o:")
				p.SetShortHandler('o', h)
				return p
			},
			wantName: "o",
			wantArg:  "file.txt",
		},
		{
			name: "short_optional_arg_present",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOpt([]string{"-dlevel3"}, "d::")
				p.SetShortHandler('d', h)
				return p
			},
			wantName: "d",
			wantArg:  "level3",
		},
		{
			name: "short_optional_arg_absent",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOpt([]string{"-d"}, "d::")
				p.SetShortHandler('d', h)
				return p
			},
			wantName: "d",
		},
		// Long handler registration (from TestHandlerSetLongHandler)
		{
			name: "long_no_arg",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--verbose"}, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				p.SetLongHandler("verbose", h)
				return p
			},
			wantName: "verbose",
		},
		{
			name: "long_required_arg_equals",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--output=file.txt"}, "", []Flag{{Name: "output", HasArg: RequiredArgument}})
				p.SetLongHandler("output", h)
				return p
			},
			wantName: "output",
			wantArg:  "file.txt",
		},
		{
			name: "long_required_arg_space",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--output", "file.txt"}, "", []Flag{{Name: "output", HasArg: RequiredArgument}})
				p.SetLongHandler("output", h)
				return p
			},
			wantName: "output",
			wantArg:  "file.txt",
		},
		// SetHandler delegation (from TestHandlerSetHandlerDelegation)
		{
			name: "SetHandler_delegates_to_long",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--verbose"}, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				p.SetHandler("--verbose", h)
				return p
			},
			wantName: "verbose",
		},
		{
			name: "SetHandler_delegates_to_short",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOpt([]string{"-v"}, "v")
				p.SetHandler("-v", h)
				return p
			},
			wantName: "v",
		},
		{
			name: "SetHandler_single_char_long",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--v"}, "", []Flag{{Name: "v", HasArg: NoArgument}})
				p.SetHandler("--v", h)
				return p
			},
			wantName: "v",
		},
		{
			name: "SetLongHandler_single_char_name",
			setup: func(h func(string, string) error) *Parser {
				p, _ := GetOptLong([]string{"--v"}, "", []Flag{{Name: "v", HasArg: NoArgument}})
				p.SetLongHandler("v", h)
				return p
			},
			wantName: "v",
		},
	}

	for _, tt := range dispatchTests {
		t.Run(tt.name, func(t *testing.T) {
			var gotName, gotArg string
			p := tt.setup(func(name, arg string) error {
				gotName = name
				gotArg = arg
				return nil
			})
			opts, errs := collectOptions(p)
			for _, e := range errs {
				if e != nil {
					t.Fatalf("unexpected error: %v", e)
				}
			}
			if len(opts) != 0 {
				t.Fatalf("expected no options yielded, got %d", len(opts))
			}
			if gotName != tt.wantName || gotArg != tt.wantArg {
				t.Fatalf("handler got (%q, %q), want (%q, %q)", gotName, gotArg, tt.wantName, tt.wantArg)
			}
		})
	}

	// --- Error tests: registration on unregistered/invalid names ---
	errorTests := []struct {
		name    string
		setup   func() *Parser
		setFn   func(*Parser) error
		wantErr bool
	}{
		// Reject unregistered (from TestHandlerSetHandlerRejectUnregistered)
		{
			name:    "reject_unregistered_short",
			setup:   func() *Parser { p, _ := GetOpt(nil, "v"); return p },
			setFn:   func(p *Parser) error { return p.SetShortHandler('x', handler) },
			wantErr: true,
		},
		{
			name: "reject_unregistered_long",
			setup: func() *Parser {
				p, _ := GetOptLong(nil, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				return p
			},
			setFn:   func(p *Parser) error { return p.SetLongHandler("quiet", handler) },
			wantErr: true,
		},
		{
			name: "reject_SetHandler_unregistered_long",
			setup: func() *Parser {
				p, _ := GetOptLong(nil, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				return p
			},
			setFn:   func(p *Parser) error { return p.SetHandler("--quiet", handler) },
			wantErr: true,
		},
		{
			name:    "reject_SetHandler_unregistered_short",
			setup:   func() *Parser { p, _ := GetOpt(nil, "v"); return p },
			setFn:   func(p *Parser) error { return p.SetHandler("-x", handler) },
			wantErr: true,
		},
		{
			name:    "registered_short_succeeds",
			setup:   func() *Parser { p, _ := GetOpt(nil, "v"); return p },
			setFn:   func(p *Parser) error { return p.SetShortHandler('v', handler) },
			wantErr: false,
		},
		{
			name: "registered_long_succeeds",
			setup: func() *Parser {
				p, _ := GetOptLong(nil, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
				return p
			},
			setFn:   func(p *Parser) error { return p.SetLongHandler("verbose", handler) },
			wantErr: false,
		},
		// No-dash-prefix (from TestHandlerSetHandlerNoDashPrefix)
		{
			name: "no_dash_bare_name",
			setup: func() *Parser {
				p, _ := GetOptLong(nil, "v", []Flag{{Name: "verbose", HasArg: NoArgument}})
				return p
			},
			setFn:   func(p *Parser) error { return p.SetHandler("verbose", handler) },
			wantErr: true,
		},
		{
			name: "no_dash_empty_string",
			setup: func() *Parser {
				p, _ := GetOptLong(nil, "v", []Flag{{Name: "verbose", HasArg: NoArgument}})
				return p
			},
			setFn:   func(p *Parser) error { return p.SetHandler("", handler) },
			wantErr: true,
		},
		{
			name: "no_dash_single_char",
			setup: func() *Parser {
				p, _ := GetOptLong(nil, "v", []Flag{{Name: "verbose", HasArg: NoArgument}})
				return p
			},
			setFn:   func(p *Parser) error { return p.SetHandler("v", handler) },
			wantErr: true,
		},
	}

	for _, tt := range errorTests {
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

	// --- No-parent-walk: Set*Handler does not walk parent chain ---
	noWalkTests := []struct {
		name  string
		setFn func(*Parser) error
	}{
		{name: "SetShortHandler_no_parent_walk", setFn: func(child *Parser) error { return child.SetShortHandler('v', handler) }},
		{name: "SetLongHandler_no_parent_walk", setFn: func(child *Parser) error { return child.SetLongHandler("verbose", handler) }},
		{name: "SetHandler_short_no_parent_walk", setFn: func(child *Parser) error { return child.SetHandler("-v", handler) }},
		{name: "SetHandler_long_no_parent_walk", setFn: func(child *Parser) error { return child.SetHandler("--verbose", handler) }},
	}

	for _, tt := range noWalkTests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
			parentShort := map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}}
			parentLong := map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}}
			parent, err := NewParser(cfg, parentShort, parentLong, nil)
			if err != nil {
				t.Fatalf("NewParser parent: %v", err)
			}
			child, err := NewParser(cfg, nil, nil, nil)
			if err != nil {
				t.Fatalf("NewParser child: %v", err)
			}
			parent.AddCmd("sub", child)

			if err := tt.setFn(child); err == nil {
				t.Fatal("expected error for parent-only option, got nil")
			}
			if parentShort['v'].Handle != nil {
				t.Fatal("parent short option Handle was modified")
			}
			if parentLong["verbose"].Handle != nil {
				t.Fatal("parent long option Handle was modified")
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

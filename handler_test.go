package optargs

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
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
		rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic seed for reproducible property tests

		// Generate a random optstring with 1–6 short options.
		nShort := 1 + rng.Intn(6)
		perm := rng.Perm(len(safeChars))
		var ob strings.Builder
		shortChars := make([]byte, nShort)
		argTypes := make([]ArgType, nShort)
		for i := range nShort {
			c := safeChars[perm[i]]
			shortChars[i] = c
			ob.WriteByte(c)
			at := ArgType(rng.Intn(3))
			argTypes[i] = at
			switch at {
			case RequiredArgument:
				ob.WriteByte(':')
			case OptionalArgument:
				ob.WriteString("::")
			}
		}
		optstring := ob.String()

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
		for range nArgs {
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
		for i := range nShort {
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
// TestHandlerSuppressesYield verifies that flags with non-nil Handle are
// dispatched to the handler (not yielded), while flags with nil Handle are
// yielded as Options.
func TestHandlerSuppressesYield(t *testing.T) {
	tests := []struct {
		name        string
		short       map[byte]*Flag
		long        map[string]*Flag
		args        []string
		wantHandled []string
		wantYielded []string
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
			name:        "long RequiredArg handled",
			long:        map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument}},
			args:        []string{"--output=file"},
			wantHandled: []string{"output"},
		},
		{
			name:  "mixed short+long handled and non-handled",
			short: map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}, 'x': {Name: "x", HasArg: NoArgument}},
			long: map[string]*Flag{
				"output": {Name: "output", HasArg: RequiredArgument},
				"debug":  {Name: "debug", HasArg: NoArgument},
			},
			args:  []string{"-v", "--output=f", "-x", "--debug"},
			// v, output handled; x, debug not
			wantHandled: []string{"v", "output"},
			wantYielded: []string{"x", "debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []handlerCall
			handler := func(name, arg string) error { calls = append(calls, handlerCall{name, arg}); return nil }

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

			p, _ := NewParser(ParserConfig{enableErrors: true, longCaseIgnore: true}, tt.short, tt.long, tt.args)
			var yielded []string
			for opt, err := range p.Options() {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				yielded = append(yielded, opt.Name)
			}

			gotHandled := make([]string, 0, len(calls))
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
// TestHandlerErrorPropagation verifies that a handler returning non-nil error
// causes the iterator to yield (zero Option, that error).
func TestHandlerErrorPropagation(t *testing.T) {
	tests := []struct {
		name       string
		errChar    byte
		otherChars []byte
	}{
		{"error on first of two options", 'a', []byte{'b'}},
		{"error on sole option", 'x', nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentinel := fmt.Errorf("handler error on %c", tt.errChar)
			shortMap := map[byte]*Flag{
				tt.errChar: {
					Name: string(tt.errChar),
					HasArg: NoArgument,
					Handle: func(string, string) error { return sentinel },
				},
			}
			args := []string{"-" + string(tt.errChar)}
			for _, c := range tt.otherChars {
				shortMap[c] = &Flag{Name: string(c), HasArg: NoArgument}
				args = append(args, "-"+string(c))
			}

			p, _ := NewParser(ParserConfig{enableErrors: true}, shortMap, nil, args)
			opts, errs := collectOptions(p)
			if len(errs) == 0 || errs[0] == nil {
				t.Fatal("expected error on first yield, got nil")
			}
			if opts[0] != (Option{}) {
				t.Fatalf("expected zero Option with error, got %+v", opts[0])
			}
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
		childShortFlags map[byte]bool   // true = child has handler
		childLongFlags  map[string]bool // true = child has handler
		wantChildCalls  []string
		wantYielded     []string
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
			parentHandler := func(name, arg string) error {
				parentCalls = append(parentCalls, handlerCall{name, arg})
				return nil
			}
			childHandler := func(name, arg string) error {
				childCalls = append(childCalls, handlerCall{name, arg})
				return nil
			}

			parentShort := make(map[byte]*Flag)
			for c := range tt.childShortFlags {
				parentShort[c] = &Flag{Name: string(c), HasArg: NoArgument, Handle: parentHandler}
			}
			parentLong := make(map[string]*Flag)
			for n := range tt.childLongFlags {
				parentLong[n] = &Flag{Name: n, HasArg: NoArgument, Handle: parentHandler}
			}

			cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
			parent, _ := NewParser(cfg, parentShort, parentLong, nil)

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

			child, _ := NewParser(cfg, childShort, childLong, args)
			parent.AddCmd("sub", child)

			var yielded []string
			for opt, err := range child.Options() {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				yielded = append(yielded, opt.Name)
			}

			if len(parentCalls) != 0 {
				t.Errorf("parent handler invoked %d times, want 0", len(parentCalls))
			}
			gotChildCalls := make([]string, 0, len(childCalls))
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
		chars       []byte
		handled     []bool
		wantCalls   []string
		wantYielded []string
	}{
		{
			name:        "alternating handled/non-handled",
			chars:       []byte{'a', 'b', 'c', 'd'},
			handled:     []bool{true, false, true, false},
			wantCalls:   []string{"a", "c"},
			wantYielded: []string{"b", "d"},
		},
		{
			name:      "all handled",
			chars:     []byte{'x', 'y', 'z'},
			handled:   []bool{true, true, true},
			wantCalls: []string{"x", "y", "z"},
		},
		{
			name:        "none handled",
			chars:       []byte{'a', 'b', 'c'},
			handled:     []bool{false, false, false},
			wantYielded: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			shortMap := make(map[byte]*Flag)
			for i, c := range tt.chars {
				f := &Flag{Name: string(c), HasArg: NoArgument}
				if tt.handled[i] {
					f.Handle = func(name, _ string) error { calls = append(calls, name); return nil }
				}
				shortMap[c] = f
			}
			compacted := "-" + string(tt.chars)
			p, _ := NewParser(ParserConfig{enableErrors: true}, shortMap, nil, []string{compacted})
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
				t.Errorf("yielded: got %v, want %v", yielded, tt.wantYielded)
			}
		})
	}
}

func TestSetHandler(t *testing.T) {
	handler := func(string, string) error { return nil }

	// --- Dispatch tests: handler is called with correct name/arg ---
	dispatchTests := []struct {
		name     string
		setup    func(t *testing.T, h func(string, string) error) *Parser
		wantName string
		wantArg  string
	}{
		{"short_no_arg", func(t *testing.T, h func(string, string) error) *Parser {
			t.Helper()
			p, _ := GetOpt([]string{"-v"}, "vx")
			if err := p.SetShortHandler('v', h); err != nil {
				t.Fatal(err)
			}
			return p
		}, "v", ""},
		{"short_required_arg", func(t *testing.T, h func(string, string) error) *Parser {
			t.Helper()
			p, _ := GetOpt([]string{"-o", "file.txt"}, "o:")
			if err := p.SetShortHandler('o', h); err != nil {
				t.Fatal(err)
			}
			return p
		}, "o", "file.txt"},
		{"short_optional_arg_present", func(t *testing.T, h func(string, string) error) *Parser {
			t.Helper()
			p, _ := GetOpt([]string{"-dlevel3"}, "d::")
			if err := p.SetShortHandler('d', h); err != nil {
				t.Fatal(err)
			}
			return p
		}, "d", "level3"},
		{"long_no_arg", func(t *testing.T, h func(string, string) error) *Parser {
			t.Helper()
			p, _ := GetOptLong([]string{"--verbose"}, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
			if err := p.SetLongHandler("verbose", h); err != nil {
				t.Fatal(err)
			}
			return p
		}, "verbose", ""},
		{"long_required_arg_equals", func(t *testing.T, h func(string, string) error) *Parser {
			t.Helper()
			p, _ := GetOptLong([]string{"--output=file.txt"}, "", []Flag{{Name: "output", HasArg: RequiredArgument}})
			if err := p.SetLongHandler("output", h); err != nil {
				t.Fatal(err)
			}
			return p
		}, "output", "file.txt"},
		{"SetHandler_delegates_to_long", func(t *testing.T, h func(string, string) error) *Parser {
			t.Helper()
			p, _ := GetOptLong([]string{"--verbose"}, "", []Flag{{Name: "verbose", HasArg: NoArgument}})
			if err := p.SetHandler("--verbose", h); err != nil {
				t.Fatal(err)
			}
			return p
		}, "verbose", ""},
		{"SetHandler_delegates_to_short", func(t *testing.T, h func(string, string) error) *Parser {
			t.Helper()
			p, _ := GetOpt([]string{"-v"}, "v")
			if err := p.SetHandler("-v", h); err != nil {
				t.Fatal(err)
			}
			return p
		}, "v", ""},
	}

	for _, tt := range dispatchTests {
		t.Run(tt.name, func(t *testing.T) {
			var gotName, gotArg string
			p := tt.setup(t, func(name, arg string) error { gotName = name; gotArg = arg; return nil })
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

	shortParser := func() *Parser {
		p, _ := GetOpt(nil, "v")
		return p
	}
	longParser := func() *Parser {
		p, _ := GetOptLong(nil, "", []Flag{
			{Name: "verbose", HasArg: NoArgument},
		})
		return p
	}
	longParserWithShort := func() *Parser {
		p, _ := GetOptLong(nil, "v", []Flag{
			{Name: "verbose", HasArg: NoArgument},
		})
		return p
	}

	errorTests := []struct {
		name    string
		setup   func() *Parser
		setFn   func(*Parser) error
		wantErr bool
	}{
		{
			name:    "reject_unregistered_short",
			setup:   shortParser,
			setFn:   func(p *Parser) error { return p.SetShortHandler('x', handler) },
			wantErr: true,
		},
		{
			name:    "reject_unregistered_long",
			setup:   longParser,
			setFn:   func(p *Parser) error { return p.SetLongHandler("quiet", handler) },
			wantErr: true,
		},
		{
			name:    "reject_SetHandler_unregistered_long",
			setup:   longParser,
			setFn:   func(p *Parser) error { return p.SetHandler("--quiet", handler) },
			wantErr: true,
		},
		{
			name:  "registered_short_succeeds",
			setup: shortParser,
			setFn: func(p *Parser) error { return p.SetShortHandler('v', handler) },
		},
		{
			name:  "registered_long_succeeds",
			setup: longParser,
			setFn: func(p *Parser) error { return p.SetLongHandler("verbose", handler) },
		},
		{
			name:    "no_dash_bare_name",
			setup:   longParserWithShort,
			setFn:   func(p *Parser) error { return p.SetHandler("verbose", handler) },
			wantErr: true,
		},
		{
			name:    "no_dash_empty_string",
			setup:   longParserWithShort,
			setFn:   func(p *Parser) error { return p.SetHandler("", handler) },
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
		{"SetShortHandler_no_parent_walk", func(child *Parser) error { return child.SetShortHandler('v', handler) }},
		{"SetLongHandler_no_parent_walk", func(child *Parser) error { return child.SetLongHandler("verbose", handler) }},
		{"SetHandler_short_no_parent_walk", func(child *Parser) error { return child.SetHandler("-v", handler) }},
		{"SetHandler_long_no_parent_walk", func(child *Parser) error { return child.SetHandler("--verbose", handler) }},
	}

	for _, tt := range noWalkTests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ParserConfig{enableErrors: true, longCaseIgnore: true}
			parentShort := map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}}
			parentLong := map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}}
			parent, _ := NewParser(cfg, parentShort, parentLong, nil)
			child, _ := NewParser(cfg, nil, nil, nil)
			parent.AddCmd("sub", child)

			if err := tt.setFn(child); err == nil {
				t.Fatal("expected error for parent-only option, got nil")
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
		handler := func(name string, _ string) error {
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
		sentinel := errors.New("long-only handler error")
		handler := func(_, _ string) error { return sentinel }
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
				sentinel := errors.New("stop here")
				longOpts := map[string]*Flag{
					"verbose": {Name: "verbose", HasArg: NoArgument, Handle: func(string, string) error { return sentinel }},
					"debug":   {Name: "debug", HasArg: NoArgument, Handle: func(string, string) error { return sentinel }},
				}
				p, _ := NewParser(
					ParserConfig{enableErrors: true, longCaseIgnore: true},
					nil, longOpts, []string{"--verbose", "--debug"},
				)
				return p
			},
		},
		{
			name: "long-only error break",
			setup: func() *Parser {
				sentinel := errors.New("long-only stop")
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
				sentinel := errors.New("compaction stop")
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
		p, _ := NewParser(
			ParserConfig{enableErrors: true, longCaseIgnore: true},
			nil, longOpts, []string{"--verbose", "--debug"},
		)
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

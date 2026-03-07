package optargs

import (
	"testing"
)

// firstErr iterates a parser and returns the first error encountered, or nil.
func firstErr(p *Parser) error {
	for _, err := range p.Options() {
		if err != nil {
			return err
		}
	}
	return nil
}

// findOpt returns the first option with the given name, or nil.
func findOpt(opts []Option, name string) *Option {
	for i := range opts {
		if opts[i].Name == name {
			return &opts[i]
		}
	}
	return nil
}

// TestNegativeArgumentSupport verifies that options requiring arguments accept
// arguments beginning with `-` (e.g. negative numbers) across all delivery
// forms: short separate, short attached, long separate, long equals, optional attached.
func TestNegativeArgumentSupport(t *testing.T) {
	numFlags := []Flag{{Name: "number", HasArg: RequiredArgument}}

	tests := []struct {
		name    string
		args    []string
		optstr  string
		flags   []Flag
		optName string
		wantArg string
	}{
		{"short separate", []string{"-a", "-123"}, "a:", nil, "a", "-123"},
		{"short attached", []string{"-a-456"}, "a:", nil, "a", "-456"},
		{"long separate", []string{"--number", "-789"}, "", numFlags, "number", "-789"},
		{"long equals", []string{"--number=-999"}, "", numFlags, "number", "-999"},
		{"optional attached", []string{"-b-100"}, "b::", nil, "b", "-100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p *Parser
			var err error
			if tt.flags != nil {
				p, err = GetOptLong(tt.args, tt.optstr, tt.flags)
			} else {
				p, err = GetOpt(tt.args, tt.optstr)
			}
			if err != nil {
				t.Fatalf("parser creation failed: %v", err)
			}
			o := findOpt(collectOpts(p), tt.optName)
			if o == nil {
				t.Fatalf("option %q not found", tt.optName)
			}
			if !o.HasArg {
				t.Fatalf("option %q: HasArg = false, want true", tt.optName)
			}
			if o.Arg != tt.wantArg {
				t.Errorf("option %q: Arg = %q, want %q", tt.optName, o.Arg, tt.wantArg)
			}
		})
	}
}

// Feature: test-refactor, Property 12: For any optstring where options are
// redefined, the parser uses the last definition encountered.

// TestOptionRedefinitionHandling verifies that when an option character appears
// multiple times in the optstring, the last definition wins.
func TestOptionRedefinitionHandling(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		optstring string
		optName   string
		wantArg   bool
		wantVal   string
	}{
		{"no-arg to required-arg", []string{"-a", "value"}, "aa:", "a", true, "value"},
		{"required-arg to no-arg", []string{"-b"}, "b:b", "b", false, ""},
		{"optional-arg to required-arg", []string{"-c", "value"}, "c::c:", "c", true, "value"},
		{"triple redef last wins no-arg", []string{"-d"}, "d:d::d", "d", false, ""},
		{"redef with behavior flags", []string{"-e"}, ":e:e", "e", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOpt(tt.args, tt.optstring)
			if err != nil {
				t.Fatalf("GetOpt(%v, %q) error: %v", tt.args, tt.optstring, err)
			}
			o := findOpt(collectOpts(p), tt.optName)
			if o == nil {
				t.Fatalf("option %q not found", tt.optName)
			}
			if o.HasArg != tt.wantArg {
				t.Errorf("HasArg = %v, want %v", o.HasArg, tt.wantArg)
			}
			if tt.wantArg && o.Arg != tt.wantVal {
				t.Errorf("Arg = %q, want %q", o.Arg, tt.wantVal)
			}
		})
	}
}

package optargs

import (
	"testing"
)

// TestEqualsDelimiterStripping verifies that the = delimiter is not
// included in the arg value when using --option=value syntax.
func TestEqualsDelimiterStripping(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		longOpts []Flag
		expected Option
	}{
		{
			name: "required arg with equals",
			args: []string{"--file=input.txt"},
			longOpts: []Flag{
				{Name: "file", HasArg: RequiredArgument},
			},
			expected: Option{Name: "file", HasArg: true, Arg: "input.txt"},
		},
		{
			name: "optional arg with equals",
			args: []string{"--config=debug"},
			longOpts: []Flag{
				{Name: "config", HasArg: OptionalArgument},
			},
			expected: Option{Name: "config", HasArg: true, Arg: "debug"},
		},
		{
			name: "empty arg with equals",
			args: []string{"--output="},
			longOpts: []Flag{
				{Name: "output", HasArg: RequiredArgument},
			},
			expected: Option{Name: "output", HasArg: true, Arg: ""},
		},
		{
			name: "negative number arg",
			args: []string{"--count=-5"},
			longOpts: []Flag{
				{Name: "count", HasArg: RequiredArgument},
			},
			expected: Option{Name: "count", HasArg: true, Arg: "-5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := GetOptLong(tt.args, "", tt.longOpts)
			if err != nil {
				t.Fatalf("GetOptLong: %v", err)
			}
			for opt, err := range p.Options() {
				if err != nil {
					t.Fatalf("Options: %v", err)
				}
				if opt.Name != tt.expected.Name {
					t.Errorf("Name: got %q, want %q", opt.Name, tt.expected.Name)
				}
				if opt.Arg != tt.expected.Arg {
					t.Errorf("Arg: got %q, want %q", opt.Arg, tt.expected.Arg)
				}
				if opt.HasArg != tt.expected.HasArg {
					t.Errorf("HasArg: got %v, want %v", opt.HasArg, tt.expected.HasArg)
				}
			}
		})
	}
}

// TestOverlappingOptionNames verifies longest-prefix-first matching when
// option names overlap and contain = characters.
func TestOverlappingOptionNames(t *testing.T) {
	t.Run("longer_name_wins", func(t *testing.T) {
		// "foo=bar" is a registered option name (NoArgument)
		// "foo" is a registered option name (RequiredArgument)
		// Input: --foo=bar → exact match on "foo=bar", no arg
		longOpts := []Flag{
			{Name: "foo", HasArg: RequiredArgument},
			{Name: "foo=bar", HasArg: NoArgument},
		}
		p, err := GetOptLong([]string{"--foo=bar"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "foo=bar" {
				t.Errorf("Name: got %q, want %q", opt.Name, "foo=bar")
			}
			if opt.HasArg {
				t.Errorf("HasArg: got true, want false")
			}
		}
	})

	t.Run("longer_name_with_arg", func(t *testing.T) {
		// "foo=bar" (RequiredArgument) and "foo" (RequiredArgument)
		// Input: --foo=bar=baz → longest match "foo=bar", arg "baz"
		longOpts := []Flag{
			{Name: "foo", HasArg: RequiredArgument},
			{Name: "foo=bar", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--foo=bar=baz"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "foo=bar" {
				t.Errorf("Name: got %q, want %q", opt.Name, "foo=bar")
			}
			if opt.Arg != "baz" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "baz")
			}
		}
	})

	t.Run("shorter_name_when_longer_not_registered", func(t *testing.T) {
		// Only "foo" registered (RequiredArgument)
		// Input: --foo=bar=baz → match "foo", arg "bar=baz"
		longOpts := []Flag{
			{Name: "foo", HasArg: RequiredArgument},
		}
		p, err := GetOptLong([]string{"--foo=bar=baz"}, "", longOpts)
		if err != nil {
			t.Fatalf("GetOptLong: %v", err)
		}
		for opt, err := range p.Options() {
			if err != nil {
				t.Fatalf("Options: %v", err)
			}
			if opt.Name != "foo" {
				t.Errorf("Name: got %q, want %q", opt.Name, "foo")
			}
			if opt.Arg != "bar=baz" {
				t.Errorf("Arg: got %q, want %q", opt.Arg, "bar=baz")
			}
		}
	})
}

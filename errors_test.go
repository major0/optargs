package optargs

import (
	"errors"
	"testing"
)

// TestTypedErrorMessages verifies each typed error produces the expected
// human-readable message matching the current parser output format.
func TestTypedErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "unknown long option",
			err:  &UnknownOptionError{Name: "verbose", IsShort: false},
			want: "unknown option: verbose",
		},
		{
			name: "unknown short option",
			err:  &UnknownOptionError{Name: "x", IsShort: true},
			want: "unknown option: x",
		},
		{
			name: "missing argument long",
			err:  &MissingArgumentError{Name: "output", IsShort: false},
			want: "option requires an argument: output",
		},
		{
			name: "missing argument short",
			err:  &MissingArgumentError{Name: "o", IsShort: true},
			want: "option requires an argument: o",
		},
		{
			name: "ambiguous option",
			err:  &AmbiguousOptionError{Name: "verb"},
			want: "ambiguous option: verb",
		},
		{
			name: "unexpected argument",
			err:  &UnexpectedArgumentError{Name: "verbose"},
			want: "option does not take an argument: verbose",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestTypedErrorsAs verifies errors.As succeeds for each typed error
// and does not conflate different error types.
func TestTypedErrorsAs(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		asUnknown bool
		asMissing bool
		asAmbig   bool
		asUnexp   bool
	}{
		{
			name:      "UnknownOptionError",
			err:       &UnknownOptionError{Name: "x"},
			asUnknown: true,
		},
		{
			name:      "MissingArgumentError",
			err:       &MissingArgumentError{Name: "o"},
			asMissing: true,
		},
		{
			name:    "AmbiguousOptionError",
			err:     &AmbiguousOptionError{Name: "verb"},
			asAmbig: true,
		},
		{
			name:    "UnexpectedArgumentError",
			err:     &UnexpectedArgumentError{Name: "verbose"},
			asUnexp: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u *UnknownOptionError
			var m *MissingArgumentError
			var a *AmbiguousOptionError
			var x *UnexpectedArgumentError

			if got := errors.As(tt.err, &u); got != tt.asUnknown {
				t.Errorf("As(UnknownOptionError) = %t, want %t", got, tt.asUnknown)
			}
			if got := errors.As(tt.err, &m); got != tt.asMissing {
				t.Errorf("As(MissingArgumentError) = %t, want %t", got, tt.asMissing)
			}
			if got := errors.As(tt.err, &a); got != tt.asAmbig {
				t.Errorf("As(AmbiguousOptionError) = %t, want %t", got, tt.asAmbig)
			}
			if got := errors.As(tt.err, &x); got != tt.asUnexp {
				t.Errorf("As(UnexpectedArgumentError) = %t, want %t", got, tt.asUnexp)
			}
		})
	}
}

// TestTypedErrorsFromParser verifies that errors returned by the Options()
// iterator are classifiable via errors.As() to the correct typed error.
func TestTypedErrorsFromParser(t *testing.T) {
	tests := []struct {
		name      string
		shortOpts map[byte]*Flag
		longOpts  map[string]*Flag
		args      []string
		wantType  string // "unknown", "missing", "ambiguous"
		wantShort bool
		wantName  string
	}{
		{
			name:      "unknown long option",
			longOpts:  map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}},
			args:      []string{"--unknown"},
			wantType:  "unknown",
			wantShort: false,
			wantName:  "unknown",
		},
		{
			name:      "unknown short option",
			shortOpts: map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}},
			args:      []string{"-x"},
			wantType:  "unknown",
			wantShort: true,
			wantName:  "x",
		},
		{
			name:      "missing argument long",
			longOpts:  map[string]*Flag{"output": {Name: "output", HasArg: RequiredArgument}},
			args:      []string{"--output"},
			wantType:  "missing",
			wantShort: false,
			wantName:  "output",
		},
		{
			name:      "missing argument short",
			shortOpts: map[byte]*Flag{'o': {Name: "o", HasArg: RequiredArgument}},
			args:      []string{"-o"},
			wantType:  "missing",
			wantShort: true,
			wantName:  "o",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shortOpts == nil {
				tt.shortOpts = map[byte]*Flag{}
			}
			if tt.longOpts == nil {
				tt.longOpts = map[string]*Flag{}
			}
			p, err := NewParser(ParserConfig{}, tt.shortOpts, tt.longOpts, tt.args)
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}

			var parseErr error
			for _, err := range p.Options() {
				if err != nil {
					parseErr = err
					break
				}
			}
			if parseErr == nil {
				t.Fatal("expected error, got nil")
			}

			switch tt.wantType {
			case "unknown":
				var ue *UnknownOptionError
				if !errors.As(parseErr, &ue) {
					t.Fatalf("expected UnknownOptionError, got %T: %v", parseErr, parseErr)
				}
				if ue.Name != tt.wantName {
					t.Errorf("Name = %q, want %q", ue.Name, tt.wantName)
				}
				if ue.IsShort != tt.wantShort {
					t.Errorf("IsShort = %t, want %t", ue.IsShort, tt.wantShort)
				}
			case "missing":
				var me *MissingArgumentError
				if !errors.As(parseErr, &me) {
					t.Fatalf("expected MissingArgumentError, got %T: %v", parseErr, parseErr)
				}
				if me.Name != tt.wantName {
					t.Errorf("Name = %q, want %q", me.Name, tt.wantName)
				}
				if me.IsShort != tt.wantShort {
					t.Errorf("IsShort = %t, want %t", me.IsShort, tt.wantShort)
				}
			case "ambiguous":
				var ae *AmbiguousOptionError
				if !errors.As(parseErr, &ae) {
					t.Fatalf("expected AmbiguousOptionError, got %T: %v", parseErr, parseErr)
				}
				if ae.Name != tt.wantName {
					t.Errorf("Name = %q, want %q", ae.Name, tt.wantName)
				}
			}
		})
	}
}

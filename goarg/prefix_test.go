package goarg

import (
	"reflect"
	"strings"
	"testing"
)

func TestPrefixTagParsing(t *testing.T) {
	tests := []struct {
		name      string
		field     reflect.StructField
		wantPairs int
		wantTrue  []string
		wantFalse []string
		wantErr   string
	}{
		{
			name: "single pair",
			field: reflect.StructField{
				Name: "Shared",
				Type: reflect.TypeOf(false),
				Tag:  `arg:"--shared" prefix:"enable,disable"`,
			},
			wantPairs: 1,
			wantTrue:  []string{"enable"},
			wantFalse: []string{"disable"},
		},
		{
			name: "multiple pairs",
			field: reflect.StructField{
				Name: "Shared",
				Type: reflect.TypeOf(false),
				Tag:  `arg:"--shared" prefix:"enable,disable;with,without"`,
			},
			wantPairs: 2,
			wantTrue:  []string{"enable", "with"},
			wantFalse: []string{"disable", "without"},
		},
		{
			name: "malformed tag no comma",
			field: reflect.StructField{
				Name: "Shared",
				Type: reflect.TypeOf(false),
				Tag:  `arg:"--shared" prefix:"enable"`,
			},
			wantErr: "invalid prefix pair",
		},
		{
			name: "prefix on non-boolean field",
			field: reflect.StructField{
				Name: "Sysroot",
				Type: reflect.TypeOf(""),
				Tag:  `arg:"--sysroot" prefix:"enable,disable"`,
			},
			wantErr: "prefix tag on non-boolean field",
		},
	}

	tp := &TagParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := tp.ParseField(tt.field, 0)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(meta.Prefixes) != tt.wantPairs {
				t.Fatalf("Prefixes count: got %d, want %d", len(meta.Prefixes), tt.wantPairs)
			}
			for i, pp := range meta.Prefixes {
				if pp.True != tt.wantTrue[i] {
					t.Errorf("pair[%d].True: got %q, want %q", i, pp.True, tt.wantTrue[i])
				}
				if pp.False != tt.wantFalse[i] {
					t.Errorf("pair[%d].False: got %q, want %q", i, pp.False, tt.wantFalse[i])
				}
			}
		})
	}
}

func TestNegatableTagParsing(t *testing.T) {
	tp := &TagParser{}

	t.Run("negatable on non-bool sets field", func(t *testing.T) {
		field := reflect.StructField{
			Name: "Sysroot",
			Type: reflect.TypeOf(""),
			Tag:  `arg:"--sysroot" negatable:""`,
		}
		meta, err := tp.ParseField(field, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !meta.Negatable {
			t.Error("Negatable: got false, want true")
		}
	})

	t.Run("negatable on bool silently ignored", func(t *testing.T) {
		field := reflect.StructField{
			Name: "Verbose",
			Type: reflect.TypeOf(false),
			Tag:  `arg:"--verbose" negatable:""`,
		}
		meta, err := tp.ParseField(field, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if meta.Negatable {
			t.Error("Negatable: got true, want false (should be ignored on bool)")
		}
	})
}

func TestGoargPrefixPairParsing(t *testing.T) {
	type Args struct {
		Shared bool `arg:"--shared" prefix:"enable,disable"`
	}

	tests := []struct {
		name    string
		args    []string
		want    bool
		wantErr string
	}{
		{"enable sets true", []string{"--enable-shared"}, true, ""},
		{"disable sets false", []string{"--disable-shared"}, false, ""},
		{"last writer wins", []string{"--disable-shared", "--enable-shared"}, true, ""},
		{"enable=value rejected", []string{"--enable-shared=true"}, false, "argument"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Args
			err := ParseArgs(&a, tt.args)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.Shared != tt.want {
				t.Errorf("Shared: got %v, want %v", a.Shared, tt.want)
			}
		})
	}
}

func TestGoargPrefixPairMultiple(t *testing.T) {
	type Args struct {
		Shared bool `arg:"--shared" prefix:"enable,disable;with,without"`
	}

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"with sets true", []string{"--with-shared"}, true},
		{"without sets false", []string{"--without-shared"}, false},
		{"last writer wins across pairs", []string{"--enable-shared", "--without-shared"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Args
			if err := ParseArgs(&a, tt.args); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.Shared != tt.want {
				t.Errorf("Shared: got %v, want %v", a.Shared, tt.want)
			}
		})
	}
}

func TestGoargNegatableZeroClear(t *testing.T) {
	type Args struct {
		Sysroot string `arg:"--sysroot" default:"/usr" negatable:""`
		Port    int    `arg:"--port" default:"8080" negatable:""`
	}

	tests := []struct {
		name     string
		args     []string
		wantRoot string
		wantPort int
	}{
		{"no-sysroot clears to empty", []string{"--no-sysroot"}, "", 8080},
		{"no-port clears to zero", []string{"--no-port"}, "/usr", 0},
		{"both cleared", []string{"--no-sysroot", "--no-port"}, "", 0},
		{"set then clear", []string{"--sysroot=/opt", "--no-sysroot"}, "", 8080},
		{"clear then set", []string{"--no-sysroot", "--sysroot=/opt"}, "/opt", 8080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Args
			if err := ParseArgs(&a, tt.args); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.Sysroot != tt.wantRoot {
				t.Errorf("Sysroot: got %q, want %q", a.Sysroot, tt.wantRoot)
			}
			if a.Port != tt.wantPort {
				t.Errorf("Port: got %d, want %d", a.Port, tt.wantPort)
			}
		})
	}
}

func TestGoargPrefixHelpText(t *testing.T) {
	type Args struct {
		Shared  bool   `arg:"--shared" prefix:"enable,disable" help:"shared library"`
		Sysroot string `arg:"--sysroot" negatable:"" help:"system root"`
	}

	p, err := NewParser(Config{Program: "test"}, &Args{})
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	p.WriteHelp(&buf)
	help := buf.String()

	for _, want := range []string{"--enable-shared", "--disable-shared", "--no-sysroot"} {
		if !strings.Contains(help, want) {
			t.Errorf("help missing %q:\n%s", want, help)
		}
	}
}

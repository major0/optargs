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

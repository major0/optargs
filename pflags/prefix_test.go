package pflags

import (
	"strings"
	"testing"
)

func TestFlagPrefixZeroValues(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.Bool("verbose", false, "verbose output")

	flag := fs.Lookup("verbose")
	if flag == nil {
		t.Fatal("flag not found")
	}
	if flag.Prefixes != nil {
		t.Errorf("Prefixes: got %v, want nil", flag.Prefixes)
	}
	if flag.Negatable {
		t.Error("Negatable: got true, want false")
	}
}

func TestMarkBoolPrefix(t *testing.T) {
	tests := []struct {
		name      string
		flag      string
		trueP     string
		falseP    string
		wantErr   string
		wantPairs int
	}{
		{
			name:    "non-existent flag",
			flag:    "missing",
			trueP:   "enable",
			falseP:  "disable",
			wantErr: "does not exist",
		},
		{
			name:    "non-boolean flag",
			flag:    "count",
			trueP:   "enable",
			falseP:  "disable",
			wantErr: "is not a boolean flag",
		},
		{
			name:      "success on bool flag",
			flag:      "verbose",
			trueP:     "enable",
			falseP:    "disable",
			wantPairs: 1,
		},
		{
			name:      "multiple pairs on same flag",
			flag:      "verbose",
			trueP:     "with",
			falseP:    "without",
			wantPairs: 2, // cumulative after previous test case — see below
		},
	}

	fs := NewFlagSet("test", ContinueOnError)
	fs.Bool("verbose", false, "verbose output")
	fs.Int("count", 0, "count things")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.MarkBoolPrefix(tt.flag, tt.trueP, tt.falseP)
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
			flag := fs.Lookup(tt.flag)
			if len(flag.Prefixes) != tt.wantPairs {
				t.Errorf("Prefixes count: got %d, want %d", len(flag.Prefixes), tt.wantPairs)
			}
		})
	}
}

func TestMarkBoolPrefixNormalizeFunc(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.SetNormalizeFunc(func(_ *FlagSet, name string) NormalizedName {
		return NormalizedName(strings.ReplaceAll(name, "_", "-"))
	})
	fs.Bool("my-flag", false, "test flag")

	// Lookup with underscore should find the flag via NormalizeFunc
	if err := fs.MarkBoolPrefix("my_flag", "enable", "disable"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	flag := fs.Lookup("my-flag")
	if len(flag.Prefixes) != 1 {
		t.Errorf("Prefixes count: got %d, want 1", len(flag.Prefixes))
	}
}


func TestMarkNegatable(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		wantErr string
	}{
		{
			name:    "non-existent flag",
			flag:    "missing",
			wantErr: "does not exist",
		},
		{
			name:    "boolean flag rejected",
			flag:    "verbose",
			wantErr: "is a boolean flag",
		},
		{
			name:    "custom type no zero value",
			flag:    "custom",
			wantErr: "has no known zero value",
		},
		{
			name: "success on string flag",
			flag: "sysroot",
		},
		{
			name: "success on int flag",
			flag: "port",
		},
		{
			name: "success on stringSlice flag",
			flag: "tags",
		},
	}

	fs := NewFlagSet("test", ContinueOnError)
	fs.Bool("verbose", false, "verbose output")
	fs.String("sysroot", "/usr", "system root")
	fs.Int("port", 8080, "port number")
	fs.StringSlice("tags", []string{"a", "b"}, "tags")
	fs.Var(&customValue{value: "x"}, "custom", "custom value")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.MarkNegatable(tt.flag)
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
			flag := fs.Lookup(tt.flag)
			if !flag.Negatable {
				t.Error("Negatable: got false, want true")
			}
		})
	}
}

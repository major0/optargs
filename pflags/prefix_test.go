package pflags

import (
	"os"
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
	t.Run("non-existent flag", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.Bool("verbose", false, "verbose output")
		err := fs.MarkBoolPrefix("missing", "enable", "disable")
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("expected 'does not exist' error, got: %v", err)
		}
	})

	t.Run("non-boolean flag", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.Int("count", 0, "count things")
		err := fs.MarkBoolPrefix("count", "enable", "disable")
		if err == nil || !strings.Contains(err.Error(), "is not a boolean flag") {
			t.Fatalf("expected 'is not a boolean flag' error, got: %v", err)
		}
	})

	t.Run("success on bool flag", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.Bool("verbose", false, "verbose output")
		if err := fs.MarkBoolPrefix("verbose", "enable", "disable"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		flag := fs.Lookup("verbose")
		if len(flag.Prefixes) != 1 {
			t.Errorf("Prefixes count: got %d, want 1", len(flag.Prefixes))
		}
	})

	t.Run("multiple pairs on same flag", func(t *testing.T) {
		fs := NewFlagSet("test", ContinueOnError)
		fs.Bool("verbose", false, "verbose output")
		if err := fs.MarkBoolPrefix("verbose", "enable", "disable"); err != nil {
			t.Fatal(err)
		}
		if err := fs.MarkBoolPrefix("verbose", "with", "without"); err != nil {
			t.Fatal(err)
		}
		flag := fs.Lookup("verbose")
		if len(flag.Prefixes) != 2 {
			t.Errorf("Prefixes count: got %d, want 2", len(flag.Prefixes))
		}
	})
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

func TestBoolPrefixParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantVal  bool
		wantErr  string
		wantArgs []string // expected positional args
	}{
		{
			name:    "enable sets true",
			args:    []string{"--enable-shared"},
			wantVal: true,
		},
		{
			name:    "disable sets false",
			args:    []string{"--disable-shared"},
			wantVal: false,
		},
		{
			name:    "with sets true",
			args:    []string{"--with-shared"},
			wantVal: true,
		},
		{
			name:    "without sets false",
			args:    []string{"--without-shared"},
			wantVal: false,
		},
		{
			name:    "original --shared still works",
			args:    []string{"--shared"},
			wantVal: true,
		},
		{
			name:    "original --no-shared still works",
			args:    []string{"--no-shared"},
			wantVal: false,
		},
		{
			name:    "last writer wins: disable then enable",
			args:    []string{"--disable-shared", "--enable-shared"},
			wantVal: true,
		},
		{
			name:    "last writer wins: enable then no-shared",
			args:    []string{"--enable-shared", "--no-shared"},
			wantVal: false,
		},
		{
			name:    "last writer wins: mixed forms",
			args:    []string{"--enable-shared", "--disable-shared", "--with-shared"},
			wantVal: true,
		},
		{
			name:     "prefixed form does not consume next arg",
			args:     []string{"--enable-shared", "/path/to/file"},
			wantVal:  true,
			wantArgs: []string{"/path/to/file"},
		},
		{
			name:    "prefixed form rejects =value",
			args:    []string{"--enable-shared=true"},
			wantErr: "enable-shared",
		},
		{
			name:    "unregistered prefix is unknown",
			args:    []string{"--activate-shared"},
			wantErr: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			fs.SetOutput(&strings.Builder{}) // suppress error output
			shared := fs.Bool("shared", false, "shared library")
			if err := fs.MarkBoolPrefix("shared", "enable", "disable"); err != nil {
				t.Fatal(err)
			}
			if err := fs.MarkBoolPrefix("shared", "with", "without"); err != nil {
				t.Fatal(err)
			}

			err := fs.Parse(tt.args)
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
			if *shared != tt.wantVal {
				t.Errorf("shared: got %v, want %v", *shared, tt.wantVal)
			}
			if tt.wantArgs != nil {
				args := fs.Args()
				if len(args) != len(tt.wantArgs) {
					t.Fatalf("args: got %v, want %v", args, tt.wantArgs)
				}
				for i, a := range args {
					if a != tt.wantArgs[i] {
						t.Errorf("arg[%d]: got %q, want %q", i, a, tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestBoolPrefixNormalizeFunc(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.SetNormalizeFunc(func(_ *FlagSet, name string) NormalizedName {
		return NormalizedName(strings.ReplaceAll(name, "_", "-"))
	})
	shared := fs.Bool("my-flag", false, "test flag")
	if err := fs.MarkBoolPrefix("my-flag", "enable", "disable"); err != nil {
		t.Fatal(err)
	}

	// Parse with underscore variant — NormalizeFunc should handle it
	if err := fs.Parse([]string{"--enable_my_flag"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !*shared {
		t.Error("expected true after --enable_my_flag")
	}
}

func TestNegatableParsing(t *testing.T) {
	tests := []struct {
		name       string
		flagName   string
		args       []string
		wantStr    string
		wantInt    int
		wantSlice  string // stringified slice value
		wantErr    string
		checkField string // which field to check
	}{
		{
			name:       "no-sysroot clears string to zero",
			flagName:   "sysroot",
			args:       []string{"--no-sysroot"},
			wantStr:    "",
			checkField: "sysroot",
		},
		{
			name:       "no-port clears int to zero",
			flagName:   "port",
			args:       []string{"--no-port"},
			wantInt:    0,
			checkField: "port",
		},
		{
			name:       "no-tags clears stringSlice to zero",
			flagName:   "tags",
			args:       []string{"--no-tags"},
			wantSlice:  "[]",
			checkField: "tags",
		},
		{
			name:       "last writer wins: set then clear",
			flagName:   "sysroot",
			args:       []string{"--sysroot=/opt", "--no-sysroot"},
			wantStr:    "",
			checkField: "sysroot",
		},
		{
			name:       "last writer wins: clear then set",
			flagName:   "sysroot",
			args:       []string{"--no-sysroot", "--sysroot=/opt"},
			wantStr:    "/opt",
			checkField: "sysroot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			fs.SetOutput(&strings.Builder{})
			sysroot := fs.String("sysroot", "/usr", "system root")
			port := fs.Int("port", 8080, "port number")
			fs.StringSlice("tags", []string{"a", "b"}, "tags")
			if err := fs.MarkNegatable("sysroot"); err != nil {
				t.Fatal(err)
			}
			if err := fs.MarkNegatable("port"); err != nil {
				t.Fatal(err)
			}
			if err := fs.MarkNegatable("tags"); err != nil {
				t.Fatal(err)
			}

			err := fs.Parse(tt.args)
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

			switch tt.checkField {
			case "sysroot":
				if *sysroot != tt.wantStr {
					t.Errorf("sysroot: got %q, want %q", *sysroot, tt.wantStr)
				}
			case "port":
				if *port != tt.wantInt {
					t.Errorf("port: got %d, want %d", *port, tt.wantInt)
				}
			case "tags":
				flag := fs.Lookup("tags")
				if flag.Value.String() != tt.wantSlice {
					t.Errorf("tags: got %q, want %q", flag.Value.String(), tt.wantSlice)
				}
			}

			// Verify Changed is set
			flag := fs.Lookup(tt.checkField)
			if !flag.Changed {
				t.Error("Changed: got false, want true")
			}
		})
	}
}

func TestNegatableUnregistered(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.SetOutput(&strings.Builder{})
	fs.String("sysroot", "/usr", "system root")
	// NOT marked negatable

	err := fs.Parse([]string{"--no-sysroot"})
	if err == nil {
		t.Fatal("expected error for --no-sysroot on non-negatable flag")
	}
	if !strings.Contains(err.Error(), "no-sysroot") {
		t.Errorf("error %q does not mention no-sysroot", err)
	}
}

func TestPrefixHelpText(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(fs *FlagSet)
		wantIn    []string // substrings that must appear
		wantNotIn []string // substrings that must not appear
	}{
		{
			name: "visible flag with prefix pairs",
			setup: func(fs *FlagSet) {
				fs.Bool("shared", false, "shared library")
				_ = fs.MarkBoolPrefix("shared", "enable", "disable")
			},
			wantIn: []string{"--enable-shared", "--disable-shared", "--shared"},
		},
		{
			name: "hidden flag with prefix pairs",
			setup: func(fs *FlagSet) {
				fs.Bool("shared", false, "shared library")
				_ = fs.MarkBoolPrefix("shared", "enable", "disable")
				_ = fs.MarkHidden("shared")
			},
			wantNotIn: []string{"--shared", "--enable-shared", "--disable-shared"},
		},
		{
			name: "negatable flag shows --no-X",
			setup: func(fs *FlagSet) {
				fs.String("sysroot", "/usr", "system root")
				_ = fs.MarkNegatable("sysroot")
			},
			wantIn: []string{"--sysroot", "--no-sysroot"},
		},
		{
			name: "non-negatable flag does not show --no-X",
			setup: func(fs *FlagSet) {
				fs.String("sysroot", "/usr", "system root")
			},
			wantNotIn: []string{"--no-sysroot"},
			wantIn:    []string{"--sysroot"},
		},
		{
			name: "multiple prefix pairs all shown",
			setup: func(fs *FlagSet) {
				fs.Bool("shared", false, "shared library")
				_ = fs.MarkBoolPrefix("shared", "enable", "disable")
				_ = fs.MarkBoolPrefix("shared", "with", "without")
			},
			wantIn: []string{"--enable-shared", "--disable-shared", "--with-shared", "--without-shared"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFlagSet("test", ContinueOnError)
			tt.setup(fs)
			usage := fs.FlagUsages()

			for _, s := range tt.wantIn {
				if !strings.Contains(usage, s) {
					t.Errorf("usage missing %q:\n%s", s, usage)
				}
			}
			for _, s := range tt.wantNotIn {
				if strings.Contains(usage, s) {
					t.Errorf("usage should not contain %q:\n%s", s, usage)
				}
			}
		})
	}
}

func TestGlobalMarkBoolPrefix(t *testing.T) {
	// Reset CommandLine for this test
	CommandLine = NewFlagSet(os.Args[0], ExitOnError)
	Bool("shared", false, "shared library")

	if err := MarkBoolPrefix("shared", "enable", "disable"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	flag := Lookup("shared")
	if len(flag.Prefixes) != 1 {
		t.Errorf("Prefixes count: got %d, want 1", len(flag.Prefixes))
	}
}

func TestGlobalMarkNegatable(t *testing.T) {
	CommandLine = NewFlagSet(os.Args[0], ExitOnError)
	String("sysroot", "/usr", "system root")

	if err := MarkNegatable("sysroot"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	flag := Lookup("sysroot")
	if !flag.Negatable {
		t.Error("Negatable: got false, want true")
	}
}


func TestParseAllWithPrefixAndNegatable(t *testing.T) {
	// Verify that ParseAll callbacks fire for prefix pair and negatable handlers.
	fs := NewFlagSet("test", ContinueOnError)
	fs.SetOutput(&strings.Builder{})
	fs.Bool("shared", false, "shared library")
	fs.String("sysroot", "/usr", "system root")
	if err := fs.MarkBoolPrefix("shared", "enable", "disable"); err != nil {
		t.Fatal(err)
	}
	if err := fs.MarkNegatable("sysroot"); err != nil {
		t.Fatal(err)
	}

	var callbacks []string
	err := fs.ParseAll([]string{"--enable-shared", "--no-sysroot"}, func(flag *Flag, value string) error {
		callbacks = append(callbacks, flag.Name+"="+value)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(callbacks) != 2 {
		t.Fatalf("expected 2 callbacks, got %d: %v", len(callbacks), callbacks)
	}
	if callbacks[0] != "shared=true" {
		t.Errorf("callback[0]: got %q, want %q", callbacks[0], "shared=true")
	}
	if callbacks[1] != "sysroot=" {
		t.Errorf("callback[1]: got %q, want %q", callbacks[1], "sysroot=")
	}
}

package compat

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/alexflint/go-arg"
)

// ptrRe matches Go pointer addresses like 0x1234abcd.
var ptrRe = regexp.MustCompile(`0x[0-9a-f]+`)

func normalizePointers(s string) string {
	return ptrRe.ReplaceAllString(s, "PTR")
}

// TestCaptureUpstream runs each scenario against upstream alexflint/go-arg
// and writes JSON golden files when -update is set.
func TestCaptureUpstream(t *testing.T) {
	for _, sc := range scenarios() {
		t.Run(sc.Name, func(t *testing.T) {
			p, dest, err := sc.NewParser()
			if err != nil {
				t.Fatalf("NewParser: %v", err)
			}

			parseErr := p.Parse(sc.Args)

			if sc.WantErr {
				if parseErr == nil {
					t.Fatalf("expected error, got nil")
				}
				goldenName := FormatGoldenName(sc.Name, "error")
				if *update {
					WriteGolden(t, goldenName, parseErr.Error())
				} else {
					want := ReadGolden(t, goldenName)
					got := parseErr.Error()
					if normalizePointers(got) != normalizePointers(strings.TrimSuffix(want, "\n")) {
						t.Errorf("error mismatch:\ngot:  %q\nwant: %q", got, want)
					}
				}
				return
			}
			if parseErr != nil {
				t.Fatalf("unexpected error: %v", parseErr)
			}

			// Capture parsed values
			if !sc.SkipValues {
				goldenName := FormatGoldenName(sc.Name, "values")
				content := fmt.Sprintf("%+v", dest)
				if *update {
					WriteGolden(t, goldenName, content)
				} else {
					want := ReadGolden(t, goldenName)
					if normalizePointers(content) != normalizePointers(strings.TrimSuffix(want, "\n")) {
						t.Errorf("values mismatch:\ngot:  %q\nwant: %q", content, want)
					}
				}
			}

			// Capture help output
			if !sc.SkipHelp {
				var helpBuf bytes.Buffer
				p.WriteHelp(&helpBuf)
				helpName := FormatGoldenName(sc.Name, "help")
				if *update {
					WriteGolden(t, helpName, helpBuf.String())
				} else {
					want := ReadGolden(t, helpName)
					if helpBuf.String() != want {
						t.Errorf("help mismatch:\ngot:\n%s\nwant:\n%s", helpBuf.String(), want)
					}
				}

				var usageBuf bytes.Buffer
				p.WriteUsage(&usageBuf)
				usageName := FormatGoldenName(sc.Name, "usage")
				if *update {
					WriteGolden(t, usageName, usageBuf.String())
				} else {
					want := ReadGolden(t, usageName)
					if usageBuf.String() != want {
						t.Errorf("usage mismatch:\ngot:\n%s\nwant:\n%s", usageBuf.String(), want)
					}
				}
			}
		})
	}
}

// TestValidateGolden verifies golden files exist for all scenarios.
func TestValidateGolden(t *testing.T) {
	for _, sc := range scenarios() {
		t.Run(sc.Name, func(t *testing.T) {
			if sc.WantErr {
				if !GoldenExists(FormatGoldenName(sc.Name, "error")) {
					t.Errorf("missing golden: %s.error", sc.Name)
				}
				return
			}
			if !sc.SkipValues {
				if !GoldenExists(FormatGoldenName(sc.Name, "values")) {
					t.Errorf("missing golden: %s.values", sc.Name)
				}
			}
			if !sc.SkipHelp {
				if !GoldenExists(FormatGoldenName(sc.Name, "help")) {
					t.Errorf("missing golden: %s.help", sc.Name)
				}
				if !GoldenExists(FormatGoldenName(sc.Name, "usage")) {
					t.Errorf("missing golden: %s.usage", sc.Name)
				}
			}
		})
	}
}

// --- Upstream feature absence tests ---
// These prove ❌ claims in the README comparison table by demonstrating
// that upstream go-arg does NOT support these features.

// TestUpstreamNoPOSIXCompaction proves upstream doesn't support -abc compaction.
func TestUpstreamNoPOSIXCompaction(t *testing.T) {
	type Args struct {
		A bool `arg:"-a"`
		B bool `arg:"-b"`
		C bool `arg:"-c"`
	}
	var a Args
	p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse([]string{"-abc"})
	if err == nil && a.A && a.B && a.C {
		t.Fatal("upstream unexpectedly supports POSIX compaction")
	}
}

// TestUpstreamNoCaseInsensitiveSubcommand proves upstream requires exact case.
func TestUpstreamNoCaseInsensitiveSubcommand(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Serve *ServeCmd `arg:"subcommand:serve"`
	}
	var a Args
	p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse([]string{"Serve"})
	if err == nil && a.Serve != nil {
		t.Fatal("upstream unexpectedly supports case-insensitive subcommands")
	}
}

// TestUpstreamNoGNULongestMatch proves upstream doesn't do prefix matching.
func TestUpstreamNoGNULongestMatch(t *testing.T) {
	type Args struct {
		EnableBob       string `arg:"--enable-bob"`
		EnableBobadufoo string `arg:"--enable-bobadufoo"`
	}
	var a Args
	p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse([]string{"--enable-boba", "val"})
	if err == nil && a.EnableBobadufoo == "val" {
		t.Fatal("upstream unexpectedly supports GNU prefix matching")
	}
}

// TestUpstreamNoBooleanNegation proves upstream go-arg doesn't support --no-flag.
func TestUpstreamNoBooleanNegation(t *testing.T) {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	var a Args
	p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse([]string{"--no-verbose"})
	if err == nil && !a.Verbose {
		t.Fatal("upstream unexpectedly supports --no-verbose negation")
	}
	// Expected: error (unknown argument --no-verbose)
}

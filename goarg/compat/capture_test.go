package compat

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
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

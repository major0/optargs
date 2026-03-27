//go:build goarg_ext

package goarg

import (
	"bytes"
	"testing"
)

// TestExtensionsEnabled verifies extensions are on when built with the tag.
func TestExtensionsEnabled(t *testing.T) {
	if !ExtensionsEnabled() {
		t.Error("ExtensionsEnabled() should be true with goarg_ext tag")
	}
}

// TestParserExtWriteHelpErr verifies the extended parser's error-returning help.
func TestParserExtWriteHelpErr(t *testing.T) {
	type Args struct {
		Verbose bool   `arg:"-v,--verbose" help:"verbose output"`
		Output  string `arg:"-o,--output" help:"output file"`
	}

	pe, err := NewParserExt(Config{Program: "test"}, &Args{})
	if err != nil {
		t.Fatalf("NewParserExt: %v", err)
	}

	var buf bytes.Buffer
	if err := pe.WriteHelpErr(&buf); err != nil {
		t.Fatalf("WriteHelpErr: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("WriteHelpErr produced empty output")
	}
}

// TestParserExtWriteUsageErr verifies the extended parser's error-returning usage.
func TestParserExtWriteUsageErr(t *testing.T) {
	type Args struct {
		Name string `arg:"--name" help:"user name"`
	}

	pe, err := NewParserExt(Config{Program: "test"}, &Args{})
	if err != nil {
		t.Fatalf("NewParserExt: %v", err)
	}

	var buf bytes.Buffer
	if err := pe.WriteUsageErr(&buf); err != nil {
		t.Fatalf("WriteUsageErr: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("WriteUsageErr produced empty output")
	}
}

// TestExtTagParser verifies the extended tag parser works.
func TestExtTagParser(t *testing.T) {
	type Args struct {
		Verbose bool `arg:"-v,--verbose" help:"verbose"`
		Count   int  `arg:"-c,--count" help:"count"`
	}

	etp := &ExtTagParser{}
	meta, err := etp.ParseStructExt(&Args{})
	if err != nil {
		t.Fatalf("ParseStructExt: %v", err)
	}
	if len(meta.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(meta.Options))
	}
}

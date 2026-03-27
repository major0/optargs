package goarg

import (
	"bytes"
	"errors"
	"testing"
)

// TestErrHelpOnDashH verifies --help returns ErrHelp.
func TestErrHelpOnDashH(t *testing.T) {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	var a Args
	err := ParseArgs(&a, []string{"--help"})
	if !errors.Is(err, ErrHelp) {
		t.Errorf("expected ErrHelp, got %v", err)
	}
}

// TestErrHelpOnShortH verifies -h returns ErrHelp.
func TestErrHelpOnShortH(t *testing.T) {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	var a Args
	err := ParseArgs(&a, []string{"-h"})
	if !errors.Is(err, ErrHelp) {
		t.Errorf("expected ErrHelp, got %v", err)
	}
}

// TestErrVersionOnDashVersion verifies --version returns ErrVersion
// when the struct implements Versioned.
func TestErrVersionOnDashVersion(t *testing.T) {
	var a versionedArgs
	p, err := NewParser(Config{}, &a)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse([]string{"--version"})
	if !errors.Is(err, ErrVersion) {
		t.Errorf("expected ErrVersion, got %v", err)
	}
}

// TestNoVersionWithoutInterface verifies --version is unknown when
// the struct does not implement Versioned and Config.Version is empty.
func TestNoVersionWithoutInterface(t *testing.T) {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	var a Args
	err := ParseArgs(&a, []string{"--version"})
	if errors.Is(err, ErrVersion) {
		t.Error("should not get ErrVersion without Versioned interface")
	}
	if err == nil {
		t.Error("expected error for --version without version configured")
	}
}

// TestVersionFromConfig verifies Config.Version enables --version.
func TestVersionFromConfig(t *testing.T) {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	var a Args
	p, err := NewParser(Config{Version: "2.0.0"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse([]string{"--version"})
	if !errors.Is(err, ErrVersion) {
		t.Errorf("expected ErrVersion, got %v", err)
	}
}

// --- Versioned/Described/Epilogued interface tests ---

type versionedArgs struct {
	Verbose bool `arg:"-v,--verbose"`
}

func (a *versionedArgs) Version() string    { return "1.2.3" }
func (a *versionedArgs) Description() string { return "A versioned app" }
func (a *versionedArgs) Epilogue() string    { return "See docs for more info." }

// TestVersionedInterface verifies the Versioned interface populates Config.Version.
func TestVersionedInterface(t *testing.T) {
	var a versionedArgs
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if p.config.Version != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %q", p.config.Version)
	}
}

// TestDescribedInterface verifies the Described interface populates Config.Description.
func TestDescribedInterface(t *testing.T) {
	var a versionedArgs
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if p.config.Description != "A versioned app" {
		t.Errorf("expected description, got %q", p.config.Description)
	}
}

// TestEpiloguedInterface verifies the Epilogued interface populates Config.Epilogue.
func TestEpiloguedInterface(t *testing.T) {
	var a versionedArgs
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if p.config.Epilogue != "See docs for more info." {
		t.Errorf("expected epilogue, got %q", p.config.Epilogue)
	}
}

// TestEpilogueInHelp verifies epilogue text appears in help output.
func TestEpilogueInHelp(t *testing.T) {
	var a versionedArgs
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	p.WriteHelp(&buf)
	if !bytes.Contains(buf.Bytes(), []byte("See docs for more info.")) {
		t.Errorf("epilogue not found in help output:\n%s", buf.String())
	}
}

// TestConfigOverridesInterface verifies explicit Config values take precedence.
func TestConfigOverridesInterface(t *testing.T) {
	var a versionedArgs
	p, err := NewParser(Config{
		Program:     "test",
		Version:     "9.9.9",
		Description: "Override desc",
	}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if p.config.Version != "9.9.9" {
		t.Errorf("Config.Version should take precedence, got %q", p.config.Version)
	}
	if p.config.Description != "Override desc" {
		t.Errorf("Config.Description should take precedence, got %q", p.config.Description)
	}
}

// TestMustParseHelp verifies MustParse handles ErrHelp by printing help and exiting 0.
func TestMustParseHelp(t *testing.T) {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	var exitCode int
	var a Args
	p, err := NewParser(Config{
		Program: "test",
		Exit:    func(code int) { exitCode = code },
	}, &a)
	if err != nil {
		t.Fatal(err)
	}
	p.MustParse([]string{"--help"})
	if exitCode != 0 {
		t.Errorf("expected exit 0 for --help, got %d", exitCode)
	}
}

// TestMustParseVersion verifies MustParse handles ErrVersion by printing version and exiting 0.
func TestMustParseVersion(t *testing.T) {
	var exitCode int
	var a versionedArgs
	p, err := NewParser(Config{
		Program: "test",
		Exit:    func(code int) { exitCode = code },
	}, &a)
	if err != nil {
		t.Fatal(err)
	}
	p.MustParse([]string{"--version"})
	if exitCode != 0 {
		t.Errorf("expected exit 0 for --version, got %d", exitCode)
	}
}

// TestMustParseError verifies MustParse exits 1 on parse errors.
func TestMustParseError(t *testing.T) {
	type Args struct {
		Input string `arg:"--input,required"`
	}
	var exitCode int
	var a Args
	p, err := NewParser(Config{
		Program: "test",
		Exit:    func(code int) { exitCode = code },
	}, &a)
	if err != nil {
		t.Fatal(err)
	}
	p.MustParse([]string{})
	if exitCode != 1 {
		t.Errorf("expected exit 1 for missing required, got %d", exitCode)
	}
}

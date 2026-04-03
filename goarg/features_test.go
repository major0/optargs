package goarg

import (
	"testing"
)

// --- Embedded struct tests ---

type CommonOpts struct {
	Verbose bool `arg:"-v,--verbose" help:"verbose output"`
	Debug   bool `arg:"-d,--debug" help:"debug mode"`
}

type EmbeddedArgs struct {
	CommonOpts

	Output string `arg:"-o,--output" help:"output file"`
}

func TestEmbeddedStructFields(t *testing.T) {
	var a EmbeddedArgs
	err := ParseArgs(&a, []string{"--verbose", "--output", "out.txt"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if !a.Verbose {
		t.Error("expected Verbose=true")
	}
	if a.Output != "out.txt" {
		t.Errorf("expected Output=out.txt, got %q", a.Output)
	}
}

func TestEmbeddedStructShortOptions(t *testing.T) {
	var a EmbeddedArgs
	err := ParseArgs(&a, []string{"-v", "-d", "-o", "file.txt"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if !a.Verbose {
		t.Error("expected Verbose=true")
	}
	if !a.Debug {
		t.Error("expected Debug=true")
	}
	if a.Output != "file.txt" {
		t.Errorf("expected Output=file.txt, got %q", a.Output)
	}
}

func TestEmbeddedStructMetadata(t *testing.T) {
	tp := &TagParser{}
	meta, err := tp.ParseStruct(&EmbeddedArgs{})
	if err != nil {
		t.Fatalf("ParseStruct: %v", err)
	}
	// Should have 3 options: verbose, debug (from CommonOpts), output
	if len(meta.Options) != 3 {
		t.Errorf("expected 3 options, got %d", len(meta.Options))
		for _, o := range meta.Options {
			t.Logf("  option: %s (short=%s long=%s)", o.Name, o.Short, o.Long)
		}
	}
}

// --- Map type tests ---

type MapArgs struct {
	Headers map[string]string `arg:"-H,--header" help:"HTTP headers"`
	Ports   map[string]int    `arg:"--port" help:"named ports"`
}

func TestMapStringString(t *testing.T) {
	var a MapArgs
	err := ParseArgs(&a, []string{"--header", "Content-Type=application/json", "--header", "Accept=text/html"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if len(a.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(a.Headers))
	}
	if a.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %q", a.Headers["Content-Type"])
	}
	if a.Headers["Accept"] != "text/html" {
		t.Errorf("expected Accept=text/html, got %q", a.Headers["Accept"])
	}
}

func TestMapStringInt(t *testing.T) {
	var a MapArgs
	err := ParseArgs(&a, []string{"--port", "http=80", "--port", "https=443"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if a.Ports["http"] != 80 {
		t.Errorf("expected http=80, got %d", a.Ports["http"])
	}
	if a.Ports["https"] != 443 {
		t.Errorf("expected https=443, got %d", a.Ports["https"])
	}
}

func TestMapInvalidFormat(t *testing.T) {
	var a MapArgs
	err := ParseArgs(&a, []string{"--header", "no-equals-sign"})
	if err == nil {
		t.Error("expected error for map value without =")
	}
}

func TestMapInvalidValueType(t *testing.T) {
	var a MapArgs
	err := ParseArgs(&a, []string{"--port", "http=notanumber"})
	if err == nil {
		t.Error("expected error for invalid map value type")
	}
}

// --- Variadic Parse tests ---

func TestParseVariadic(t *testing.T) {
	// Parse with single dest (existing behavior)
	type Args struct {
		Name string `arg:"--name"`
	}
	var a Args
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"--name", "alice"}); err != nil {
		t.Fatal(err)
	}
	if a.Name != "alice" {
		t.Errorf("expected alice, got %q", a.Name)
	}
}

// --- Register tests ---

func TestRegister(t *testing.T) {
	// Just verify Register doesn't panic
	type Args struct {
		Name string `arg:"--name"`
	}
	var a Args
	Register(&a)
	// registrations is package-level, verify it grew
	if len(registrations) == 0 {
		t.Error("Register did not add to registrations")
	}
	// Clean up
	registrations = nil
}

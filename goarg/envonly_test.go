package goarg

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// --- Env-only field tests ---

type EnvOnlyArgs struct {
	Token   string `arg:"env:API_TOKEN" help:"API authentication token"`
	Verbose bool   `arg:"-v,--verbose" help:"verbose output"`
}

func TestEnvOnlyFieldParsedFromEnv(t *testing.T) {
	os.Setenv("API_TOKEN", "secret123")
	defer os.Unsetenv("API_TOKEN")

	var a EnvOnlyArgs
	err := ParseArgs(&a, []string{})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if a.Token != "secret123" {
		t.Errorf("expected Token=secret123, got %q", a.Token)
	}
}

func TestEnvOnlyFieldNotAsCLIFlag(t *testing.T) {
	var a EnvOnlyArgs
	err := ParseArgs(&a, []string{"--token", "value"})
	if err == nil {
		t.Error("expected error for --token (env-only field should not be a CLI flag)")
	}
}

func TestEnvOnlyFieldInHelp(t *testing.T) {
	var a EnvOnlyArgs
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	p.WriteHelp(&buf)
	help := buf.String()
	if !strings.Contains(help, "Environment variables:") {
		t.Error("help missing 'Environment variables:' section")
	}
	if !strings.Contains(help, "API_TOKEN") {
		t.Error("help missing API_TOKEN env var")
	}
}

func TestEnvOnlyMetadata(t *testing.T) {
	tp := &TagParser{}
	meta, err := tp.ParseStruct(&EnvOnlyArgs{})
	if err != nil {
		t.Fatal(err)
	}
	if len(meta.EnvOnly) != 1 {
		t.Fatalf("expected 1 env-only field, got %d", len(meta.EnvOnly))
	}
	if meta.EnvOnly[0].Env != "API_TOKEN" {
		t.Errorf("expected env=API_TOKEN, got %q", meta.EnvOnly[0].Env)
	}
	// Options should only have verbose (not token)
	if len(meta.Options) != 1 {
		t.Errorf("expected 1 option, got %d", len(meta.Options))
	}
}

// --- IgnoreEnv / IgnoreDefault tests ---

func TestIgnoreEnv(t *testing.T) {
	os.Setenv("API_TOKEN", "should-be-ignored")
	defer os.Unsetenv("API_TOKEN")

	var a EnvOnlyArgs
	p, err := NewParser(Config{IgnoreEnv: true}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	if a.Token != "" {
		t.Errorf("expected empty Token with IgnoreEnv, got %q", a.Token)
	}
}

func TestIgnoreDefault(t *testing.T) {
	type Args struct {
		Port int `arg:"-p,--port" default:"8080"`
	}
	var a Args
	p, err := NewParser(Config{IgnoreDefault: true}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	if a.Port != 0 {
		t.Errorf("expected Port=0 with IgnoreDefault, got %d", a.Port)
	}
}

func TestIgnoreDefaultStillAcceptsCLI(t *testing.T) {
	type Args struct {
		Port int `arg:"-p,--port" default:"8080"`
	}
	var a Args
	p, err := NewParser(Config{IgnoreDefault: true}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"--port", "9090"}); err != nil {
		t.Fatal(err)
	}
	if a.Port != 9090 {
		t.Errorf("expected Port=9090, got %d", a.Port)
	}
}

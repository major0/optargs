package goarg

import (
	"bytes"
	"strings"
	"testing"
)

type subServerCmd struct {
	Port int `arg:"-p,--port" default:"8080" help:"listen port"`
}

type subClientCmd struct {
	URL string `arg:"-u,--url" help:"server URL"`
}

type subRoot struct {
	Verbose bool          `arg:"-v,--verbose"`
	Server  *subServerCmd `arg:"subcommand:server" help:"run server"`
	Client  *subClientCmd `arg:"subcommand:client" help:"run client"`
}

func TestSubcommandReturnsActiveStruct(t *testing.T) {
	var root subRoot
	p, err := NewParser(Config{Program: "test"}, &root)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"server", "--port", "9090"}); err != nil {
		t.Fatal(err)
	}
	sub := p.Subcommand()
	if sub == nil {
		t.Fatal("Subcommand() returned nil")
	}
	srv, ok := sub.(*subServerCmd)
	if !ok {
		t.Fatalf("expected *subServerCmd, got %T", sub)
	}
	if srv.Port != 9090 {
		t.Errorf("expected port 9090, got %d", srv.Port)
	}
}

func TestSubcommandNilWhenNoneInvoked(t *testing.T) {
	var root subRoot
	p, err := NewParser(Config{Program: "test"}, &root)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"--verbose"}); err != nil {
		t.Fatal(err)
	}
	if p.Subcommand() != nil {
		t.Error("Subcommand() should be nil when no subcommand invoked")
	}
}

func TestSubcommandNames(t *testing.T) {
	var root subRoot
	p, err := NewParser(Config{Program: "test"}, &root)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"client", "--url", "http://localhost"}); err != nil {
		t.Fatal(err)
	}
	names := p.SubcommandNames()
	if len(names) != 1 || names[0] != "client" {
		t.Errorf("expected [client], got %v", names)
	}
}

func TestSubcommandNamesNilWhenNone(t *testing.T) {
	var root subRoot
	p, err := NewParser(Config{Program: "test"}, &root)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{}); err != nil {
		t.Fatal(err)
	}
	if p.SubcommandNames() != nil {
		t.Error("SubcommandNames() should be nil when no subcommand invoked")
	}
}

func TestWriteHelpForSubcommand(t *testing.T) {
	var root subRoot
	p, err := NewParser(Config{Program: "test"}, &root)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := p.WriteHelpForSubcommand(&buf, "server"); err != nil {
		t.Fatal(err)
	}
	help := buf.String()
	if !strings.Contains(help, "Usage:") {
		t.Error("subcommand help missing Usage: line")
	}
	if !strings.Contains(help, "port") {
		t.Error("subcommand help missing port option")
	}
}

func TestWriteUsageForSubcommand(t *testing.T) {
	var root subRoot
	p, err := NewParser(Config{Program: "test"}, &root)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := p.WriteUsageForSubcommand(&buf, "server"); err != nil {
		t.Fatal(err)
	}
	usage := buf.String()
	if !strings.Contains(usage, "Usage:") {
		t.Error("subcommand usage missing Usage: line")
	}
}

func TestWriteHelpForUnknownSubcommand(t *testing.T) {
	var root subRoot
	p, err := NewParser(Config{Program: "test"}, &root)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = p.WriteHelpForSubcommand(&buf, "unknown")
	if err == nil {
		t.Error("expected error for unknown subcommand")
	}
}

func TestConfigStrictSubcommands(t *testing.T) {
	type ServerCmd struct {
		Port int `arg:"-p,--port" default:"8080"`
	}
	type Root struct {
		Verbose bool       `arg:"-v,--verbose"`
		Server  *ServerCmd `arg:"subcommand:server"`
	}
	var root Root
	p, err := NewParser(Config{
		Program:           "test",
		StrictSubcommands: true,
	}, &root)
	if err != nil {
		t.Fatal(err)
	}
	// With strict subcommands, --verbose after "server" should fail
	err = p.Parse([]string{"server", "--verbose"})
	if err == nil {
		t.Error("expected error for parent option in strict subcommand mode")
	}
}

func TestConfigOut(t *testing.T) {
	type Args struct {
		Input string `arg:"--input,required"`
	}
	var buf bytes.Buffer
	var exitCode int
	var a Args
	p, err := NewParser(Config{
		Program: "test",
		Out:     &buf,
		Exit:    func(code int) { exitCode = code },
	}, &a)
	if err != nil {
		t.Fatal(err)
	}
	p.MustParse([]string{}) // missing required → error output
	if exitCode != 1 {
		t.Errorf("expected exit 1, got %d", exitCode)
	}
	if buf.Len() == 0 {
		t.Error("expected output to Config.Out, got nothing")
	}
}

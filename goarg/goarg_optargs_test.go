package goarg

import (
	"testing"
)

// OptArgs-exclusive feature tests. These features are provided by OptArgs Core
// and have no upstream go-arg equivalent. No golden file comparison — assert
// correct behavior directly.

// TestOptArgsPOSIXCompaction tests -abc expanding to -a -b -c.
func TestOptArgsPOSIXCompaction(t *testing.T) {
	type Args struct {
		A bool `arg:"-a"`
		B bool `arg:"-b"`
		C bool `arg:"-c"`
	}
	var a Args
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"-abc"}); err != nil {
		t.Fatal(err)
	}
	if !a.A || !a.B || !a.C {
		t.Errorf("got a=%t b=%t c=%t, want all true", a.A, a.B, a.C)
	}
}

// TestOptArgsDoubleHyphenTermination tests -- stops option processing.
func TestOptArgsDoubleHyphenTermination(t *testing.T) {
	type Args struct {
		Name string   `arg:"--name"`
		Rest []string `arg:"positional"`
	}
	var a Args
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"--name", "val", "--", "--other", "pos"}); err != nil {
		t.Fatal(err)
	}
	if a.Name != "val" {
		t.Errorf("name = %q", a.Name)
	}
	if len(a.Rest) != 2 || a.Rest[0] != "--other" || a.Rest[1] != "pos" {
		t.Errorf("rest = %v", a.Rest)
	}
}

// TestOptArgsSubcommandParentFlags tests parent flag inheritance across subcommands.
func TestOptArgsSubcommandParentFlags(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Verbose bool     `arg:"-v,--verbose"`
		Server  *ServeCmd `arg:"subcommand:server"`
	}
	var a Args
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"-v", "server", "--port", "9090"}); err != nil {
		t.Fatal(err)
	}
	if !a.Verbose {
		t.Error("parent flag --verbose should be true")
	}
	if a.Server == nil || a.Server.Port != 9090 {
		t.Errorf("server port = %v", a.Server)
	}
}

// TestOptArgsCaseInsensitiveSubcommand tests case-insensitive subcommand matching.
func TestOptArgsCaseInsensitiveSubcommand(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Serve *ServeCmd `arg:"subcommand:serve"`
	}
	tests := []struct {
		name string
		args []string
	}{
		{"lowercase", []string{"serve", "--port", "9090"}},
		{"uppercase", []string{"Serve", "--port", "9090"}},
		{"mixed", []string{"SERVE", "--port", "9090"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Args
			p, err := NewParser(Config{Program: "test"}, &a)
			if err != nil {
				t.Fatal(err)
			}
			if err := p.Parse(tt.args); err != nil {
				t.Fatalf("Parse(%v): %v", tt.args, err)
			}
			if a.Serve == nil || a.Serve.Port != 9090 {
				t.Errorf("serve = %v", a.Serve)
			}
		})
	}
}

// TestOptArgsGNULongestMatch tests GNU longest-match prefix resolution
// through the goarg layer.
func TestOptArgsGNULongestMatch(t *testing.T) {
	type Args struct {
		EnableBob      string `arg:"--enable-bob"`
		EnableBobadufoo string `arg:"--enable-bobadufoo"`
	}
	var a Args
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"--enable-bobadufoo", "val"}); err != nil {
		t.Fatal(err)
	}
	if a.EnableBobadufoo != "val" {
		t.Errorf("enable-bobadufoo = %q, want val", a.EnableBobadufoo)
	}
	if a.EnableBob != "" {
		t.Errorf("enable-bob = %q, want empty", a.EnableBob)
	}
}

// TestOptArgsSubcommandQuery tests Subcommand() and SubcommandNames() methods.
func TestOptArgsSubcommandQuery(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Serve *ServeCmd `arg:"subcommand:serve"`
	}
	var a Args
	p, err := NewParser(Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"serve", "--port", "9090"}); err != nil {
		t.Fatal(err)
	}
	sub := p.Subcommand()
	if sub == nil {
		t.Fatal("Subcommand() returned nil")
	}
	names := p.SubcommandNames()
	if len(names) == 0 {
		t.Fatal("SubcommandNames() returned empty")
	}
	if names[0] != "serve" {
		t.Errorf("SubcommandNames()[0] = %q, want serve", names[0])
	}
}

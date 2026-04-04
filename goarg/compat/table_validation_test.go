// table_validation_test.go validates every row of the Feature Comparison table
// in goarg/README.md by running each feature against upstream alexflint/go-arg.
//
// Each test is named after the table row it validates. Tests that prove upstream
// DOES support a feature (✅) assert success. Tests that prove upstream does NOT
// support a feature (❌) assert failure or absence.
//
// To reproduce: go test -run TestTable -v
package compat

import (
	"testing"

	"github.com/alexflint/go-arg"
)

// --- Upstream ✅ rows: features go-arg supports ---

func TestTable_StructTagParsing(t *testing.T) {
	type Args struct {
		Name string `arg:"--name" help:"user name" default:"anon"`
	}
	var a Args
	p, err := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Parse([]string{"--name", "alice"}); err != nil {
		t.Fatalf("upstream should support struct tag parsing: %v", err)
	}
	if a.Name != "alice" {
		t.Errorf("name = %q, want alice", a.Name)
	}
}

func TestTable_ShortLongOptions(t *testing.T) {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"-v"}); err != nil {
		t.Fatalf("upstream should support short options: %v", err)
	}
	if !a.Verbose {
		t.Error("short -v should set verbose")
	}
}

func TestTable_PositionalArguments(t *testing.T) {
	type Args struct {
		Source string `arg:"positional,required"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"input.txt"}); err != nil {
		t.Fatalf("upstream should support positional args: %v", err)
	}
	if a.Source != "input.txt" {
		t.Errorf("source = %q", a.Source)
	}
}

func TestTable_Subcommands(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Serve *ServeCmd `arg:"subcommand:serve"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"serve", "--port", "9090"}); err != nil {
		t.Fatalf("upstream should support subcommands: %v", err)
	}
	if a.Serve == nil || a.Serve.Port != 9090 {
		t.Errorf("serve = %v", a.Serve)
	}
}

func TestTable_EnvironmentVariableFallback(t *testing.T) {
	type Args struct {
		Token string `arg:"--token,env:TEST_TOKEN"`
	}
	t.Setenv("TEST_TOKEN", "secret")
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{}); err != nil {
		t.Fatalf("upstream should support env fallback: %v", err)
	}
	if a.Token != "secret" {
		t.Errorf("token = %q, want secret", a.Token)
	}
}

func TestTable_EnvOnlyFields(t *testing.T) {
	type Args struct {
		Secret string `arg:"--,env:TEST_SECRET"` // pragma: allowlist secret
	}
	t.Setenv("TEST_SECRET", "hidden") // pragma: allowlist secret
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{}); err != nil {
		t.Fatalf("upstream should support env-only fields: %v", err)
	}
	if a.Secret != "hidden" { // pragma: allowlist secret
		t.Errorf("secret = %q, want hidden", a.Secret) // pragma: allowlist secret
	}
}

func TestTable_DefaultValues(t *testing.T) {
	type Args struct {
		Port int `arg:"--port" default:"8080"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{}); err != nil {
		t.Fatalf("upstream should support defaults: %v", err)
	}
	if a.Port != 8080 {
		t.Errorf("port = %d, want 8080", a.Port)
	}
}

func TestTable_MapTypes(t *testing.T) {
	type Args struct {
		Headers map[string]string `arg:"--header"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"--header", "K=V"}); err != nil {
		t.Fatalf("upstream should support map types: %v", err)
	}
	if a.Headers["K"] != "V" {
		t.Errorf("headers = %v", a.Headers)
	}
}

func TestTable_SliceTypes(t *testing.T) {
	type Args struct {
		Files []string `arg:"--file"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"--file", "a", "--file", "b"}); err != nil {
		t.Fatalf("upstream should support slice types: %v", err)
	}
	if len(a.Files) < 1 {
		t.Error("expected at least one file")
	}
}

func TestTable_EmbeddedStructInheritance(t *testing.T) {
	type Common struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	type Args struct {
		Common
		Name string `arg:"--name"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"--verbose", "--name", "alice"}); err != nil {
		t.Fatalf("upstream should support embedded structs: %v", err)
	}
	if !a.Verbose || a.Name != "alice" {
		t.Errorf("verbose=%t name=%q", a.Verbose, a.Name)
	}
}

type versionedTestArgs struct {
	Name string `arg:"--name"`
}

func (versionedTestArgs) Version() string     { return "1.0.0" }
func (versionedTestArgs) Description() string { return "test app" }

func TestTable_VersionedDescribedEpilogued(t *testing.T) {
	var a versionedTestArgs
	_, err := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err != nil {
		t.Fatalf("upstream should support Versioned/Described interfaces: %v", err)
	}
}

func TestTable_ErrHelpErrVersionSentinels(t *testing.T) {
	type Args struct {
		Name string `arg:"--name"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	err := p.Parse([]string{"--help"})
	if err == nil {
		t.Fatal("--help should return an error sentinel")
	}
}

func TestTable_BuiltinHelpVersionFlags(t *testing.T) {
	var a versionedTestArgs
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	err := p.Parse([]string{"--version"})
	if err == nil {
		t.Fatal("--version should return an error sentinel")
	}
}

func TestTable_SubcommandQuery(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Serve *ServeCmd `arg:"subcommand:serve"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"serve"}); err != nil {
		t.Fatal(err)
	}
	if p.Subcommand() == nil {
		t.Error("Subcommand() should return non-nil")
	}
	if len(p.SubcommandNames()) == 0 {
		t.Error("SubcommandNames() should be non-empty")
	}
}

func TestTable_DoubleHyphenTermination(t *testing.T) {
	type Args struct {
		Name string   `arg:"--name"`
		Rest []string `arg:"positional"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"--name", "val", "--", "--other"}); err != nil {
		t.Fatalf("upstream should support -- termination: %v", err)
	}
	if len(a.Rest) != 1 || a.Rest[0] != "--other" {
		t.Errorf("rest = %v, want [--other]", a.Rest)
	}
}

func TestTable_ParentFlagInheritance(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Verbose bool      `arg:"-v,--verbose"`
		Server  *ServeCmd `arg:"subcommand:server"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"server", "--verbose", "--port", "9090"}); err != nil {
		t.Fatalf("upstream should support parent flag inheritance: %v", err)
	}
	if !a.Verbose {
		t.Error("verbose should be true")
	}
}

func TestTable_InterspersedArguments(t *testing.T) {
	type Args struct {
		Name string   `arg:"--name"`
		Rest []string `arg:"positional"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	if err := p.Parse([]string{"pos1", "--name", "val", "pos2"}); err != nil {
		t.Fatalf("upstream should support interspersed args: %v", err)
	}
	if a.Name != "val" {
		t.Errorf("name = %q", a.Name)
	}
	if len(a.Rest) != 2 {
		t.Errorf("rest = %v, want 2 positionals", a.Rest)
	}
}

// --- Upstream ❌ rows: features go-arg does NOT support ---

func TestTable_NoPOSIXCompaction(t *testing.T) {
	type Args struct {
		A bool `arg:"-a"`
		B bool `arg:"-b"`
		C bool `arg:"-c"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	err := p.Parse([]string{"-abc"})
	if err == nil && a.A && a.B && a.C {
		t.Fatal("upstream unexpectedly supports POSIX compaction")
	}
}

func TestTable_NoGNULongestMatch(t *testing.T) {
	type Args struct {
		EnableBob       string `arg:"--enable-bob"`
		EnableBobadufoo string `arg:"--enable-bobadufoo"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	err := p.Parse([]string{"--enable-boba", "val"})
	if err == nil && a.EnableBobadufoo == "val" {
		t.Fatal("upstream unexpectedly supports GNU longest-match")
	}
}

func TestTable_NoBooleanNegation(t *testing.T) {
	type Args struct {
		Verbose bool `arg:"-v,--verbose"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	err := p.Parse([]string{"--no-verbose"})
	if err == nil && !a.Verbose {
		t.Fatal("upstream unexpectedly supports --no-verbose")
	}
}

func TestTable_NoCaseInsensitiveSubcommand(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Serve *ServeCmd `arg:"subcommand:serve"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)
	err := p.Parse([]string{"SERVE"})
	if err == nil && a.Serve != nil {
		t.Fatal("upstream unexpectedly supports case-insensitive subcommands")
	}
}

func TestTable_NoGetoptLongOnly(t *testing.T) {
	// Upstream go-arg is ALWAYS in getopt_long_only(3) mode — single-dash
	// arguments are parsed as long options. -v=true works, compaction doesn't.
	// This test confirms the always-on behavior.
	type Args struct {
		Verbose bool `arg:"--verbose"`
	}
	var a Args
	p, _ := arg.NewParser(arg.Config{Program: "test"}, &a)

	// -verbose should match --verbose (long-only behavior)
	if err := p.Parse([]string{"-verbose"}); err != nil {
		t.Fatalf("upstream should accept -verbose in long-only mode: %v", err)
	}
	if !a.Verbose {
		t.Error("verbose should be true")
	}

	// -v=true should work (long-opt value syntax on single-dash)
	type Args2 struct {
		V bool `arg:"--v"`
	}
	var a2 Args2
	p2, _ := arg.NewParser(arg.Config{Program: "test"}, &a2)
	if err := p2.Parse([]string{"-v=true"}); err != nil {
		t.Fatalf("upstream should accept -v=true in long-only mode: %v", err)
	}
	if !a2.V {
		t.Error("-v=true should set v")
	}

	// Compaction should NOT work (no true short options)
	type Args3 struct {
		A bool `arg:"-a"`
		B bool `arg:"-b"`
	}
	var a3 Args3
	p3, _ := arg.NewParser(arg.Config{Program: "test"}, &a3)
	err := p3.Parse([]string{"-ab"})
	if err == nil && a3.A && a3.B {
		t.Fatal("upstream should not support compaction (always long-only)")
	}
}

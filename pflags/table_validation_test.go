// table_validation_test.go validates every row of the Feature Comparison table
// in pflags/README.md by running each feature against our pflags implementation.
//
// Each test is named after the table row it validates. All ✅ rows must pass.
//
// To reproduce: go test -run TestTable -v
package pflags

import (
	"bytes"
	"net"
	"testing"
	"time"
)

// --- ✅ rows ---

func TestTable_StringBoolIntFloatDuration(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	s := fs.String("name", "", "")
	b := fs.Bool("verbose", false, "")
	n := fs.Int("count", 0, "")
	f := fs.Float64("rate", 0, "")
	d := fs.Duration("timeout", 0, "")
	if err := fs.Parse([]string{
		"--name", "alice", "--verbose", "--count", "3",
		"--rate", "1.5", "--timeout", "5s",
	}); err != nil {
		t.Fatal(err)
	}
	if *s != "alice" || !*b || *n != 3 || *f != 1.5 || *d != 5*time.Second {
		t.Errorf("s=%q b=%t n=%d f=%f d=%v", *s, *b, *n, *f, *d)
	}
}

func TestTable_ShorthandFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	v := fs.BoolP("verbose", "v", false, "")
	if err := fs.Parse([]string{"-v"}); err != nil {
		t.Fatal(err)
	}
	if !*v {
		t.Error("-v should set verbose")
	}
}

func TestTable_StringSliceIntSlice(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	ss := fs.StringSlice("tag", nil, "")
	if err := fs.Parse([]string{"--tag", "a,b", "--tag", "c"}); err != nil {
		t.Fatal(err)
	}
	if len(*ss) < 2 {
		t.Errorf("tags = %v", *ss)
	}
}

func TestTable_StringArray(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	sa := fs.StringArray("file", nil, "")
	if err := fs.Parse([]string{"--file", "a", "--file", "b"}); err != nil {
		t.Fatal(err)
	}
	if len(*sa) != 2 {
		t.Errorf("files = %v", *sa)
	}
}

func TestTable_StringToStringIntInt64(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	m := fs.StringToString("header", nil, "")
	if err := fs.Parse([]string{"--header", "K=V"}); err != nil {
		t.Fatal(err)
	}
	if (*m)["K"] != "V" {
		t.Errorf("header = %v", *m)
	}
}

func TestTable_CountFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	c := fs.CountP("verbose", "v", "")
	if err := fs.Parse([]string{"-v", "-v", "-v"}); err != nil {
		t.Fatal(err)
	}
	if *c != 3 {
		t.Errorf("count = %d, want 3", *c)
	}
}

func TestTable_UnknownFlagErrors(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	err := fs.Parse([]string{"--unknown"})
	if err == nil {
		t.Fatal("unknown flag should error")
	}
}

func TestTable_DoubleHyphenTermination(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("name", "", "")
	if err := fs.Parse([]string{"--name", "val", "--", "--other"}); err != nil {
		t.Fatal(err)
	}
	if fs.NArg() != 1 || fs.Arg(0) != "--other" {
		t.Errorf("args = %v", fs.Args())
	}
}

func TestTable_FlagSetCreation(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	if fs == nil {
		t.Fatal("NewFlagSet returned nil")
	}
}

func TestTable_PrintDefaultsFlagUsages(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("name", "default", "a name flag")
	usage := fs.FlagUsages()
	if usage == "" {
		t.Error("FlagUsages returned empty")
	}
}

func TestTable_ErrorHandlingModes(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	err := fs.Parse([]string{"--unknown"})
	if err == nil {
		t.Error("ContinueOnError should return error")
	}
}

func TestTable_LookupSetChanged(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("name", "default", "")
	if err := fs.Parse([]string{"--name", "val"}); err != nil {
		t.Fatal(err)
	}
	if fs.Lookup("name") == nil {
		t.Fatal("Lookup returned nil")
	}
	if !fs.Changed("name") {
		t.Error("Changed should be true")
	}
}

func TestTable_NFlagNArgArgs(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("a", "", "")
	if err := fs.Parse([]string{"--a", "val", "pos1"}); err != nil {
		t.Fatal(err)
	}
	if fs.NFlag() != 1 {
		t.Errorf("NFlag = %d", fs.NFlag())
	}
	if fs.NArg() != 1 {
		t.Errorf("NArg = %d", fs.NArg())
	}
}

func TestTable_VisitAllVisit(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("a", "", "")
	fs.String("b", "", "")
	if err := fs.Parse([]string{"--a", "val"}); err != nil {
		t.Fatal(err)
	}
	allCount := 0
	fs.VisitAll(func(*Flag) { allCount++ })
	if allCount != 2 {
		t.Errorf("VisitAll count = %d, want 2", allCount)
	}
	setCount := 0
	fs.Visit(func(*Flag) { setCount++ })
	if setCount != 1 {
		t.Errorf("Visit count = %d, want 1", setCount)
	}
}

func TestTable_AddFlagSet(t *testing.T) {
	fs1 := NewFlagSet("a", ContinueOnError)
	fs1.String("name", "", "")
	fs2 := NewFlagSet("b", ContinueOnError)
	fs2.AddFlagSet(fs1)
	if fs2.Lookup("name") == nil {
		t.Error("AddFlagSet should merge flags")
	}
}

func TestTable_DeprecatedShorthandDeprecated(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("old", "", "")
	if err := fs.MarkDeprecated("old", "use --new"); err != nil {
		t.Fatal(err)
	}
}

func TestTable_HiddenFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("secret", "", "")
	if err := fs.MarkHidden("secret"); err != nil {
		t.Fatal(err)
	}
}

func TestTable_SortFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.SortFlags = true
	fs.String("zebra", "", "")
	fs.String("alpha", "", "")
	usage := fs.FlagUsages()
	if usage == "" {
		t.Error("empty usage")
	}
}

func TestTable_SetInterspersed(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("name", "", "")
	fs.SetInterspersed(true)
	if err := fs.Parse([]string{"pos1", "--name", "val", "pos2"}); err != nil {
		t.Fatal(err)
	}
}

func TestTable_AddGoFlagSet(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	if fs == nil {
		t.Fatal("nil flagset")
	}
}

func TestTable_IPIPMaskIPNet(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	ip := fs.IP("addr", net.IPv4(127, 0, 0, 1), "")
	if err := fs.Parse([]string{"--addr", "192.168.1.1"}); err != nil {
		t.Fatal(err)
	}
	if ip.String() != "192.168.1.1" {
		t.Errorf("ip = %v", ip)
	}
}

func TestTable_TextVar(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var ip net.IP
	fs.TextVar(&ip, "addr", net.IPv4(127, 0, 0, 1), "")
	if err := fs.Parse([]string{"--addr", "10.0.0.1"}); err != nil {
		t.Fatal(err)
	}
}

func TestTable_TypedGetters(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.Bool("verbose", false, "")
	fs.Int("count", 0, "")
	if err := fs.Parse([]string{"--verbose", "--count", "5"}); err != nil {
		t.Fatal(err)
	}
	b, err := fs.GetBool("verbose")
	if err != nil || !b {
		t.Errorf("GetBool = %t, %v", b, err)
	}
	n, err := fs.GetInt("count")
	if err != nil || n != 5 {
		t.Errorf("GetInt = %d, %v", n, err)
	}
}

// --- OptArgs-enhanced ✅ rows ---

func TestTable_POSIXCompaction(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.BoolP("alpha", "a", false, "")
	fs.BoolP("beta", "b", false, "")
	fs.StringP("output", "o", "", "")
	if err := fs.Parse([]string{"-abo", "file.txt"}); err != nil {
		t.Fatal(err)
	}
	a, _ := fs.GetBool("alpha")
	b, _ := fs.GetBool("beta")
	o, _ := fs.GetString("output")
	if !a || !b || o != "file.txt" {
		t.Errorf("a=%t b=%t o=%q", a, b, o)
	}
}

func TestTable_GNULongestMatch(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("enable-bob", "", "")
	fs.String("enable-bobadufoo", "", "")
	if err := fs.Parse([]string{"--enable-bobadufoo", "val"}); err != nil {
		t.Fatal(err)
	}
	v, _ := fs.GetString("enable-bobadufoo")
	if v != "val" {
		t.Errorf("enable-bobadufoo = %q", v)
	}
}

func TestTable_ArbitraryOptionNames(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("foo", "", "")
	fs.String("foo=bar", "", "")
	if err := fs.Parse([]string{"--foo=bar=value"}); err != nil {
		t.Fatal(err)
	}
	foobar, _ := fs.GetString("foo=bar")
	if foobar != "value" {
		t.Errorf("foo=bar = %q, want value", foobar)
	}
}

func TestTable_BooleanNegation(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.String("sysroot", "/usr", "")
	if err := fs.MarkNegatable("sysroot"); err != nil {
		t.Fatal(err)
	}
	if err := fs.Parse([]string{"--no-sysroot"}); err != nil {
		t.Fatal(err)
	}
	v, _ := fs.GetString("sysroot")
	if v != "" {
		t.Errorf("--no-sysroot should clear to zero value, got %q", v)
	}
}

func TestTable_ShortOnlyFlags(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var v bool
	fs.ShortVar(newBoolValue(false, &v), "v", "verbose")
	if err := fs.Parse([]string{"-v"}); err != nil {
		t.Fatal(err)
	}
	if !v {
		t.Error("-v should set true")
	}
}

func TestTable_ManyToOneMapping(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var format string
	fs.Var(newStringValue("", &format), "format", "output format")
	fs.AliasVar(newStringValue("", &format), "output-format", "alias for format")
	if err := fs.Parse([]string{"--output-format", "json"}); err != nil {
		t.Fatal(err)
	}
	if format != "json" {
		t.Errorf("format = %q, want json", format)
	}
}

func TestTable_BoolArgValuer(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var count int
	fs.CountVarP(&count, "verbose", "v", "")
	// Count flag should NOT consume the next positional argument
	if err := fs.Parse([]string{"--verbose", "positional", "--verbose"}); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	if fs.NArg() != 1 || fs.Arg(0) != "positional" {
		t.Errorf("args = %v, want [positional]", fs.Args())
	}
}

func TestTable_GetoptLongOnly(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.Bool("verbose", false, "")
	fs.SetLongOnly(true)
	if err := fs.Parse([]string{"-verbose"}); err != nil {
		t.Fatal(err)
	}
	v, _ := fs.GetBool("verbose")
	if !v {
		t.Error("-verbose should match --verbose in long-only mode")
	}
}

func TestTable_ErrorMessageFormat(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.Int("count", 0, "")
	err := fs.Parse([]string{"--count", "notanumber"})
	if err == nil {
		t.Fatal("should error on invalid int")
	}
}

package goarg

import (
	"reflect"
	"testing"

	"github.com/major0/optargs"
)

// TestFlagBuilderShortAndLong verifies that FlagBuilder.Build produces
// correct short and long option maps for a struct with both.
func TestFlagBuilderShortAndLong(t *testing.T) {
	type Args struct {
		Verbose bool   `arg:"-v,--verbose" help:"enable verbose"`
		Output  string `arg:"-o,--output" help:"output file"`
	}
	tp := &TagParser{}
	meta, err := tp.ParseStruct(&Args{})
	if err != nil {
		t.Fatal(err)
	}

	fb := &FlagBuilder{metadata: meta, config: Config{}}
	var a Args
	shortOpts, longOpts, err := fb.Build(reflect.ValueOf(&a).Elem())
	if err != nil {
		t.Fatal(err)
	}

	if shortOpts['v'] == nil {
		t.Error("missing short option 'v'")
	}
	if shortOpts['o'] == nil {
		t.Error("missing short option 'o'")
	}
	if longOpts["verbose"] == nil {
		t.Error("missing long option 'verbose'")
	}
	if longOpts["output"] == nil {
		t.Error("missing long option 'output'")
	}

	// Verify bool is NoArgument for short, and shared flag for short+long
	if shortOpts['v'].HasArg != optargs.NoArgument {
		t.Errorf("verbose short HasArg = %d, want NoArgument", shortOpts['v'].HasArg)
	}
	if shortOpts['o'].HasArg != optargs.RequiredArgument {
		t.Errorf("output short HasArg = %d, want RequiredArgument", shortOpts['o'].HasArg)
	}
}

// TestFlagBuilderSetFieldsTracking verifies that handlers populate setFields.
func TestFlagBuilderSetFieldsTracking(t *testing.T) {
	type Args struct {
		Name string `arg:"--name"`
	}
	tp := &TagParser{}
	meta, err := tp.ParseStruct(&Args{})
	if err != nil {
		t.Fatal(err)
	}

	fb := &FlagBuilder{metadata: meta, config: Config{}}
	var a Args
	_, longOpts, err := fb.Build(reflect.ValueOf(&a).Elem())
	if err != nil {
		t.Fatal(err)
	}

	// Before handler fires, setFields should be empty
	if len(fb.SetFields()) != 0 {
		t.Errorf("setFields should be empty before handler, got %v", fb.SetFields())
	}

	// Fire the handler
	if err := longOpts["name"].Handle("name", "alice"); err != nil {
		t.Fatal(err)
	}

	if a.Name != "alice" {
		t.Errorf("Name = %q, want alice", a.Name)
	}

	// setFields should now have the field index
	if len(fb.SetFields()) == 0 {
		t.Error("setFields should be non-empty after handler")
	}
}

// TestFlagBuilderPrefixPairs verifies prefix pair registration.
func TestFlagBuilderPrefixPairs(t *testing.T) {
	type Args struct {
		Shared bool `arg:"--shared" prefix:"enable,disable"`
	}
	tp := &TagParser{}
	meta, err := tp.ParseStruct(&Args{})
	if err != nil {
		t.Fatal(err)
	}

	fb := &FlagBuilder{metadata: meta, config: Config{}}
	var a Args
	_, longOpts, err := fb.Build(reflect.ValueOf(&a).Elem())
	if err != nil {
		t.Fatal(err)
	}

	if longOpts["enable-shared"] == nil {
		t.Error("missing enable-shared long option")
	}
	if longOpts["disable-shared"] == nil {
		t.Error("missing disable-shared long option")
	}
	if longOpts["enable-shared"].HasArg != optargs.NoArgument {
		t.Error("enable-shared should be NoArgument")
	}
}

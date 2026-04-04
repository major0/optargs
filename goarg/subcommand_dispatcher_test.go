package goarg

import (
	"reflect"
	"testing"
)

// TestFindSubcommandFieldDirect verifies findSubcommandField works
// with direct and case-insensitive lookup.
func TestFindSubcommandFieldDirect(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Serve *ServeCmd `arg:"subcommand:serve"`
	}

	tp := &TagParser{}
	meta, err := tp.ParseStruct(&Args{})
	if err != nil {
		t.Fatal(err)
	}

	ci := &CoreIntegration{metadata: meta, config: Config{}}
	a := Args{Serve: &ServeCmd{Port: 9090}}

	// Direct lookup
	fv, subMeta, err := ci.findSubcommandField(reflect.ValueOf(&a).Elem(), "serve")
	if err != nil {
		t.Fatal(err)
	}
	if subMeta == nil {
		t.Fatal("subMeta should not be nil")
	}
	if fv.Kind() != reflect.Ptr {
		t.Errorf("field kind = %v, want Ptr", fv.Kind())
	}

	// Case-insensitive lookup
	_, _, err = ci.findSubcommandField(reflect.ValueOf(&a).Elem(), "SERVE")
	if err != nil {
		t.Fatalf("case-insensitive lookup failed: %v", err)
	}

	// Unknown subcommand
	_, _, err = ci.findSubcommandField(reflect.ValueOf(&a).Elem(), "unknown")
	if err == nil {
		t.Error("expected error for unknown subcommand")
	}
}

// TestRegisterSubcommandsDirect verifies RegisterSubcommands creates
// child parsers and registers them.
func TestRegisterSubcommandsDirect(t *testing.T) {
	type ServeCmd struct {
		Port int `arg:"--port" default:"8080"`
	}
	type Args struct {
		Serve *ServeCmd `arg:"subcommand:serve"`
	}

	tp := &TagParser{}
	meta, err := tp.ParseStruct(&Args{})
	if err != nil {
		t.Fatal(err)
	}

	ci := &CoreIntegration{metadata: meta, config: Config{}}
	var a Args
	destValue := reflect.ValueOf(&a).Elem()

	coreParser, err := ci.CreateParserWithHandlers([]string{}, destValue)
	if err != nil {
		t.Fatal(err)
	}

	if err := ci.RegisterSubcommands(coreParser, destValue); err != nil {
		t.Fatal(err)
	}

	// Verify the subcommand was registered
	cmd, exists := coreParser.GetCommand("serve")
	if !exists || cmd == nil {
		t.Error("serve subcommand should be registered")
	}
}

package goarg

import (
	"reflect"
	"testing"

	"github.com/major0/optargs"
)

// TestPostProcessorDefaults verifies that PostProcessor applies defaults
// to unset fields.
func TestPostProcessorDefaults(t *testing.T) {
	type Args struct {
		Port int `arg:"--port" default:"8080"`
	}
	tp := &TagParser{}
	meta, err := tp.ParseStruct(&Args{})
	if err != nil {
		t.Fatal(err)
	}

	var a Args
	// Create a parser with no args (nothing parsed)
	p, err := optargs.NewParser(optargs.ParserConfig{}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	pp := &PostProcessor{
		metadata:  meta,
		config:    Config{},
		setFields: make(map[int]bool),
	}
	pp.buildPositionalArgs()

	if err := pp.Process(p, reflect.ValueOf(&a).Elem()); err != nil {
		t.Fatal(err)
	}

	if a.Port != 8080 {
		t.Errorf("Port = %d, want 8080", a.Port)
	}
}

// TestPostProcessorSkipsSetFields verifies that PostProcessor skips
// fields that were explicitly set during parsing.
func TestPostProcessorSkipsSetFields(t *testing.T) {
	type Args struct {
		Port int `arg:"--port" default:"8080"`
	}
	tp := &TagParser{}
	meta, err := tp.ParseStruct(&Args{})
	if err != nil {
		t.Fatal(err)
	}

	a := Args{Port: 9090}
	p, err := optargs.NewParser(optargs.ParserConfig{}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Mark the field as set
	setFields := map[int]bool{0: true}
	pp := &PostProcessor{
		metadata:  meta,
		config:    Config{},
		setFields: setFields,
	}
	pp.buildPositionalArgs()

	if err := pp.Process(p, reflect.ValueOf(&a).Elem()); err != nil {
		t.Fatal(err)
	}

	// Port should remain 9090, not overwritten by default
	if a.Port != 9090 {
		t.Errorf("Port = %d, want 9090 (should not be overwritten)", a.Port)
	}
}

// TestPostProcessorEnvFallback verifies environment variable fallback.
func TestPostProcessorEnvFallback(t *testing.T) {
	type Args struct {
		Token string `arg:"--token,env:TEST_PP_TOKEN"`
	}
	t.Setenv("TEST_PP_TOKEN", "from-env")

	tp := &TagParser{}
	meta, err := tp.ParseStruct(&Args{})
	if err != nil {
		t.Fatal(err)
	}

	var a Args
	p, err := optargs.NewParser(optargs.ParserConfig{}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	pp := &PostProcessor{
		metadata:  meta,
		config:    Config{},
		setFields: make(map[int]bool),
	}
	pp.buildPositionalArgs()

	if err := pp.Process(p, reflect.ValueOf(&a).Elem()); err != nil {
		t.Fatal(err)
	}

	if a.Token != "from-env" {
		t.Errorf("Token = %q, want from-env", a.Token)
	}
}

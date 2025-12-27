package goarg

import (
	"fmt"
	"testing"
)

func TestSimpleDebug(t *testing.T) {
	type SimpleCmd struct {
		Verbose bool `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int  `arg:"-c,--count" help:"number of items"`
	}

	var cmd SimpleCmd
	err := ParseArgs(&cmd, []string{"-v", "--count", "42"})
	if err != nil {
		t.Fatalf("ParseArgs() unexpected error: %v", err)
	}

	fmt.Printf("Parsed: Verbose=%v, Count=%v\n", cmd.Verbose, cmd.Count)

	if !cmd.Verbose {
		t.Errorf("Expected Verbose=true, got %v", cmd.Verbose)
	}
	if cmd.Count != 42 {
		t.Errorf("Expected Count=42, got %v", cmd.Count)
	}
}

func TestDebugStructParsing(t *testing.T) {
	type TestCmd struct {
		Verbose bool `arg:"-v,--verbose" help:"enable verbose output"`
		Count   int  `arg:"-c,--count" help:"number of items"`
	}

	parser := &TagParser{}
	var cmd TestCmd
	metadata, err := parser.ParseStruct(&cmd)
	if err != nil {
		t.Fatalf("ParseStruct() unexpected error: %v", err)
	}

	fmt.Printf("Metadata: %+v\n", metadata)
	for i, field := range metadata.Fields {
		fmt.Printf("Field %d: Name=%s, Short=%s, Long=%s, Type=%v\n", 
			i, field.Name, field.Short, field.Long, field.Type)
	}
}
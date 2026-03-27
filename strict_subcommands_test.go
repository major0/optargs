package optargs

import (
	"strings"
	"testing"
)

func TestStrictSubcommands_ParentOptionRejected(t *testing.T) {
	// Root has --verbose, child has --port.
	// With StrictSubcommands, --verbose should fail inside the subcommand.
	rootShort := map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}}
	rootLong := map[string]*Flag{"verbose": {Name: "verbose", HasArg: NoArgument}}
	root, err := NewParser(ParserConfig{
		enableErrors:      true,
		strictSubcommands: true,
		commandCaseIgnore: true,
	}, rootShort, rootLong, []string{"serve", "--verbose"})
	if err != nil {
		t.Fatalf("root parser: %v", err)
	}

	childShort := map[byte]*Flag{'p': {Name: "p", HasArg: RequiredArgument}}
	childLong := map[string]*Flag{"port": {Name: "port", HasArg: RequiredArgument}}
	child, err := NewParser(ParserConfig{enableErrors: true}, childShort, childLong, []string{})
	if err != nil {
		t.Fatalf("child parser: %v", err)
	}

	root.AddCmd("serve", child)

	// Iterate root — dispatches to child
	for _, err := range root.Options() {
		if err != nil {
			t.Fatalf("root iteration: %v", err)
		}
	}

	// Child should reject --verbose since parent walk is disabled
	var childErr error
	for _, err := range child.Options() {
		if err != nil {
			childErr = err
			break
		}
	}
	if childErr == nil {
		t.Fatal("expected error for --verbose in strict subcommand, got nil")
	}
	if !strings.Contains(childErr.Error(), "unknown option") {
		t.Errorf("expected 'unknown option' error, got: %v", childErr)
	}
}

func TestStrictSubcommands_ChildOwnOptionsWork(t *testing.T) {
	rootShort := map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}}
	root, err := NewParser(ParserConfig{
		enableErrors:      true,
		strictSubcommands: true,
		commandCaseIgnore: true,
	}, rootShort, nil, []string{"serve", "--port", "8080"})
	if err != nil {
		t.Fatalf("root parser: %v", err)
	}

	childLong := map[string]*Flag{"port": {Name: "port", HasArg: RequiredArgument}}
	child, err := NewParser(ParserConfig{enableErrors: true}, nil, childLong, []string{})
	if err != nil {
		t.Fatalf("child parser: %v", err)
	}

	root.AddCmd("serve", child)

	for _, err := range root.Options() {
		if err != nil {
			t.Fatalf("root iteration: %v", err)
		}
	}

	var port string
	for opt, err := range child.Options() {
		if err != nil {
			t.Fatalf("child iteration: %v", err)
		}
		if opt.Name == "port" {
			port = opt.Arg
		}
	}
	if port != "8080" {
		t.Errorf("expected port=8080, got %q", port)
	}
}

func TestStrictSubcommands_DefaultAllowsInheritance(t *testing.T) {
	// Without StrictSubcommands, parent options should be inherited.
	rootShort := map[byte]*Flag{'v': {Name: "v", HasArg: NoArgument}}
	root, err := NewParser(ParserConfig{
		enableErrors:      true,
		commandCaseIgnore: true,
	}, rootShort, nil, []string{"serve", "-v"})
	if err != nil {
		t.Fatalf("root parser: %v", err)
	}

	child, err := NewParser(ParserConfig{enableErrors: true}, nil, nil, []string{})
	if err != nil {
		t.Fatalf("child parser: %v", err)
	}

	root.AddCmd("serve", child)

	for _, err := range root.Options() {
		if err != nil {
			t.Fatalf("root iteration: %v", err)
		}
	}

	// Child should resolve -v via parent walk
	var found bool
	for opt, err := range child.Options() {
		if err != nil {
			t.Fatalf("child iteration: %v", err)
		}
		if opt.Name == "v" {
			found = true
		}
	}
	if !found {
		t.Error("expected -v to be inherited from parent, but it wasn't")
	}
}

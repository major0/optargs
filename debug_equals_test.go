package optargs

import (
	"fmt"
	"testing"
)

func TestDebugEquals(t *testing.T) {
	longOpts := []Flag{
		{Name: "output", HasArg: RequiredArgument},
	}

	args := []string{"--output=file.txt"}
	parser, err := GetOptLong(args, "", longOpts)
	if err != nil {
		t.Fatalf("GetOptLong failed: %v", err)
	}

	for opt, err := range parser.Options() {
		if err != nil {
			t.Fatalf("Options iteration failed: %v", err)
		}
		fmt.Printf("Option: Name='%s', HasArg=%t, Arg='%s'\n", opt.Name, opt.HasArg, opt.Arg)
	}
}
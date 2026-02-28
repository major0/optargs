// Command silent demonstrates silent error mode (: prefix in optstring).
// In silent mode, error logging is suppressed — the caller handles errors
// via the iterator's error return.
//
// Usage:
//
//	go run ./posix/silent -- -v -x -f
package main

import (
	"fmt"
	"os"

	"github.com/major0/optargs"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		// -v is valid, -x is unknown, -f requires an argument but none given
		args = []string{"-v", "-x", "-f"}
	}

	// Leading : enables silent error mode
	p, err := optargs.GetOpt(args, ":vf:")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for opt, err := range p.Options() {
		if err != nil {
			// In silent mode, errors are returned without logging.
			// The caller decides how to report them.
			fmt.Printf("error (silent): %v\n", err)
			continue
		}
		switch {
		case opt.HasArg:
			fmt.Printf("option: -%s  arg: %s\n", opt.Name, opt.Arg)
		default:
			fmt.Printf("option: -%s\n", opt.Name)
		}
	}
}

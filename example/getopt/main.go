// Command getopt demonstrates POSIX getopt(3) short option parsing.
//
// Usage:
//
//	go run ./example/getopt -- -vf input.txt -o output.txt -- extra args
package main

import (
	"fmt"
	"os"

	"github.com/major0/optargs"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		// Demo args when run without arguments
		args = []string{"-vf", "input.txt", "-o", "output.txt", "--", "extra", "args"}
	}

	// optstring: v (no arg), f: (required arg), o:: (optional arg)
	p, err := optargs.GetOpt(args, "vf:o::")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for opt, err := range p.Options() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		switch {
		case opt.HasArg:
			fmt.Printf("option: -%s  arg: %s\n", opt.Name, opt.Arg)
		default:
			fmt.Printf("option: -%s\n", opt.Name)
		}
	}

	fmt.Printf("remaining: %v\n", p.Args)
}

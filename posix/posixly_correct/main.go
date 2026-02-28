// Command posixly_correct demonstrates POSIXLY_CORRECT behavior.
// When the + prefix is used in the optstring (or the POSIXLY_CORRECT
// environment variable is set), the parser stops at the first non-option
// argument.
//
// Usage:
//
//	go run ./posix/posixly_correct -- -v file.txt -f input.txt
package main

import (
	"fmt"
	"os"

	"github.com/major0/optargs"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		// -v is parsed, "file.txt" is a non-option → parsing stops
		// -f input.txt is NOT parsed (left in remaining args)
		args = []string{"-v", "file.txt", "-f", "input.txt"}
	}

	// Leading + enables POSIXLY_CORRECT mode
	p, err := optargs.GetOpt(args, "+vf:")
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

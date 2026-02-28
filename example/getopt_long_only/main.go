// Command getopt_long_only demonstrates GNU getopt_long_only(3) parsing
// where single-dash options are tried as long options first, falling back
// to short option parsing via the optstring.
//
// Usage:
//
//	go run ./example/getopt_long_only -- -verbose -file input.txt -v
package main

import (
	"fmt"
	"os"

	"github.com/major0/optargs"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		// -verbose matches long option "verbose"
		// -file matches long option "file"
		// -v falls back to short option 'v' from optstring
		args = []string{"-verbose", "-file", "input.txt", "-v"}
	}

	longopts := []optargs.Flag{
		{Name: "verbose", HasArg: optargs.NoArgument},
		{Name: "file", HasArg: optargs.RequiredArgument},
	}

	// optstring "vf:" provides short option fallback
	p, err := optargs.GetOptLongOnly(args, "vf:", longopts)
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
			fmt.Printf("option: %s  arg: %s\n", opt.Name, opt.Arg)
		default:
			fmt.Printf("option: %s\n", opt.Name)
		}
	}

	fmt.Printf("remaining: %v\n", p.Args)
}

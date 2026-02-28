// Command getopt_long demonstrates GNU getopt_long(3) parsing with both
// short and long options.
//
// Usage:
//
//	go run ./example/getopt_long -- --verbose --file=input.txt -o output.txt
package main

import (
	"fmt"
	"os"

	"github.com/major0/optargs"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"--verbose", "--file", "input.txt", "-o", "output.txt"}
	}

	longopts := []optargs.Flag{
		{Name: "verbose", HasArg: optargs.NoArgument},
		{Name: "file", HasArg: optargs.RequiredArgument},
		{Name: "output", HasArg: optargs.OptionalArgument},
	}

	p, err := optargs.GetOptLong(args, "vf:o::", longopts)
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

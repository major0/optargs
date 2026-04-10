package main

import (
	"fmt"

	"github.com/major0/optargs/pflag"
)

func main() {
	fs := pflag.NewFlagSet("example", pflag.ContinueOnError)

	var verbose bool
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose output")

	// Usage: ./app --verbose
	fs.Parse([]string{"--verbose"})
	fmt.Printf("Verbose mode: %t\n", verbose) // Output: Verbose mode: true
}

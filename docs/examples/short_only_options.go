package main

import (
	"fmt"

	"github.com/major0/optargs/pflag"
)

func main() {
	fs := pflag.NewFlagSet("example", pflag.ContinueOnError)

	var output string
	fs.StringVarP(&output, "output", "o", "stdout", "Output destination")

	// Usage: ./app -o file.txt
	fs.Parse([]string{"-o", "file.txt"})
	fmt.Printf("Output: %s\n", output) // Output: Output: file.txt
}

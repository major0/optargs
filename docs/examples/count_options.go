package main

import (
	"fmt"
	"github.com/major0/optargs/pflags"
	"strings"
)

func main() {
	fs := pflags.NewFlagSet("example", pflags.ContinueOnError)

	var verbosity int
	// Custom counter implementation using Var
	fs.Var(&CountValue{&verbosity}, "verbose", "Increase verbosity (can be repeated)")

	// Usage: ./app -v -v -v  (or --verbose --verbose --verbose)
	fs.Parse([]string{"--verbose", "--verbose", "--verbose"})
	fmt.Printf("Verbosity level: %d\n", verbosity) // Output: Verbosity level: 3
}

// CountValue implements Value interface for counting
type CountValue struct {
	count *int
}

func (c *CountValue) String() string   { return fmt.Sprintf("%d", *c.count) }
func (c *CountValue) Set(string) error { *c.count++; return nil }
func (c *CountValue) Type() string     { return "count" }

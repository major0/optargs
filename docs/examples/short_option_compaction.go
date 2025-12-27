package main

import (
	"fmt"
	"github.com/major0/optargs/pflags"
)

func main() {
	fs := pflags.NewFlagSet("compaction-example", pflags.ContinueOnError)
	
	// Define various short options with different argument requirements
	var (
		verbose   bool
		force     bool
		extract   bool
		archive   bool
		file      string
		level     int
		output    string
	)
	
	// Boolean flags (no arguments)
	fs.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	fs.BoolVarP(&force, "force", "f", false, "Force operation")
	fs.BoolVarP(&extract, "extract", "x", false, "Extract mode")
	fs.BoolVarP(&archive, "archive", "a", false, "Archive mode")
	
	// Flags that require arguments
	fs.StringVarP(&file, "file", "F", "", "Input file")
	fs.IntVarP(&level, "level", "l", 0, "Compression level")
	fs.StringVarP(&output, "output", "o", "", "Output destination")
	
	fmt.Println("=== Short-Option Compaction Examples ===\n")
	
	// Example 1: Basic compaction (all boolean flags)
	fmt.Println("1. Basic compaction: -vfx")
	fs1 := copyFlagSet(fs)
	err := fs1.Parse([]string{"-vfx"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printFlags(fs1, "verbose", "force", "extract")
	}
	
	// Example 2: Compaction with argument to last option
	fmt.Println("\n2. Compaction with argument: -vfo output.txt")
	fs2 := copyFlagSet(fs)
	err = fs2.Parse([]string{"-vfo", "output.txt"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printFlags(fs2, "verbose", "force", "output")
	}
	
	// Example 3: Compaction with attached argument
	fmt.Println("\n3. Compaction with attached argument: -vfl5")
	fs3 := copyFlagSet(fs)
	err = fs3.Parse([]string{"-vfl5"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printFlags(fs3, "verbose", "force", "level")
	}
	
	// Example 4: Complex compaction (tar-style)
	fmt.Println("\n4. Tar-style compaction: -xvfF archive.tar")
	fs4 := copyFlagSet(fs)
	err = fs4.Parse([]string{"-xvfF", "archive.tar"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printFlags(fs4, "extract", "verbose", "force", "file")
	}
	
	// Example 5: Mixed compaction and regular flags
	fmt.Println("\n5. Mixed usage: -vf --output=result.txt -l 9")
	fs5 := copyFlagSet(fs)
	err = fs5.Parse([]string{"-vf", "--output=result.txt", "-l", "9"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printFlags(fs5, "verbose", "force", "output", "level")
	}
}

// Helper function to copy a FlagSet for multiple parsing examples
func copyFlagSet(original *pflags.FlagSet) *pflags.FlagSet {
	fs := pflags.NewFlagSet("copy", pflags.ContinueOnError)
	
	var (
		verbose, force, extract, archive bool
		file, output                     string
		level                           int
	)
	
	fs.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	fs.BoolVarP(&force, "force", "f", false, "Force operation")
	fs.BoolVarP(&extract, "extract", "x", false, "Extract mode")
	fs.BoolVarP(&archive, "archive", "a", false, "Archive mode")
	fs.StringVarP(&file, "file", "F", "", "Input file")
	fs.IntVarP(&level, "level", "l", 0, "Compression level")
	fs.StringVarP(&output, "output", "o", "", "Output destination")
	
	return fs
}

// Helper function to print flag values
func printFlags(fs *pflags.FlagSet, flagNames ...string) {
	for _, name := range flagNames {
		flag := fs.Lookup(name)
		if flag != nil && flag.Changed {
			fmt.Printf("  -%s (%s): %s\n", flag.Shorthand, flag.Name, flag.Value.String())
		}
	}
}

// Example output:
// === Short-Option Compaction Examples ===
//
// 1. Basic compaction: -vfx
//   -v (verbose): true
//   -f (force): true
//   -x (extract): true
//
// 2. Compaction with argument: -vfo output.txt
//   -v (verbose): true
//   -f (force): true
//   -o (output): output.txt
//
// 3. Compaction with attached argument: -vfl5
//   -v (verbose): true
//   -f (force): true
//   -l (level): 5
//
// 4. Tar-style compaction: -xvfF archive.tar
//   -x (extract): true
//   -v (verbose): true
//   -f (force): true
//   -F (file): archive.tar
//
// 5. Mixed usage: -vf --output=result.txt -l 9
//   -v (verbose): true
//   -f (force): true
//   -o (output): result.txt
//   -l (level): 9
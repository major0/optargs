// Command subcommand demonstrates native subcommand dispatch using AddCmd(),
// multi-level dispatch, and option inheritance through the parser tree.
//
// Usage:
//
//	go run ./posix/subcommand -- --verbose db --name mydb migrate --steps 3
package main

import (
	"fmt"
	"os"

	"github.com/major0/optargs"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"--verbose", "db", "--name", "mydb", "migrate", "--steps", "3"}
	}

	// Root parser: -v / --verbose
	root, err := optargs.GetOptLong(args, "v", []optargs.Flag{
		{Name: "verbose", HasArg: optargs.NoArgument},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// db subcommand: -n / --name (required arg)
	db, err := optargs.GetOptLong([]string{}, "n:", []optargs.Flag{
		{Name: "name", HasArg: optargs.RequiredArgument},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	root.AddCmd("db", db)

	// migrate subcommand under db: -s / --steps (required arg)
	migrate, err := optargs.GetOptLong([]string{}, "s:", []optargs.Flag{
		{Name: "steps", HasArg: optargs.RequiredArgument},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	db.AddCmd("migrate", migrate)

	// Phase 1: root iteration — yields --verbose, dispatches "db"
	fmt.Println("=== root ===")
	for opt, err := range root.Options() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "root error: %v\n", err)
			continue
		}
		fmt.Printf("  option: %s\n", opt.Name)
	}

	// Phase 2: db iteration — yields --name, dispatches "migrate"
	fmt.Println("=== db ===")
	for opt, err := range db.Options() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "db error: %v\n", err)
			continue
		}
		fmt.Printf("  option: %s  arg: %s\n", opt.Name, opt.Arg)
	}

	// Phase 3: migrate iteration — yields --steps
	// migrate can also resolve root's --verbose via parent chain
	fmt.Println("=== migrate ===")
	for opt, err := range migrate.Options() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "migrate error: %v\n", err)
			continue
		}
		fmt.Printf("  option: %s  arg: %s\n", opt.Name, opt.Arg)
	}
}

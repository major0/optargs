package main

import (
	"fmt"
	"github.com/major0/optargs/pflags"
	"os"
	"time"
)

func main() {
	fs := pflags.NewFlagSet("myapp", pflags.ContinueOnError)

	// Define flags with descriptive usage text
	var (
		host    = fs.StringP("host", "h", "localhost", "Server `hostname` to connect to")
		port    = fs.IntP("port", "p", 8080, "Server `port` number")
		timeout = fs.Duration("timeout", 30*time.Second, "Connection `timeout` duration")
		verbose = fs.BoolP("verbose", "v", false, "Enable verbose output")
		config  = fs.String("config", "", "Configuration `file` path")
		tags    = fs.StringSlice("tag", []string{}, "Add `tags` (repeatable)")
	)

	// Custom usage function
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "MyApp - A sample application\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", fs.Name())
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --host=api.example.com --port=443 --verbose\n", fs.Name())
		fmt.Fprintf(os.Stderr, "  %s -h localhost -p 8080 --tag=web --tag=api\n", fs.Name())
	}

	// Parse arguments
	err := fs.Parse([]string{"--help"})
	if err != nil {
		return
	}

	// Use the parsed values
	fmt.Printf("Connecting to %s:%d\n", *host, *port)
	fmt.Printf("Timeout: %v, Verbose: %t\n", *timeout, *verbose)
	fmt.Printf("Config: %s, Tags: %v\n", *config, *tags)
}

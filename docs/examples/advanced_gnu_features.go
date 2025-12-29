package main

import (
	"fmt"

	"github.com/major0/optargs/pflags"
)

func main() {
	fs := pflags.NewFlagSet("advanced", pflags.ContinueOnError)

	// Special characters in option names (colons, equals)
	var (
		systemVerbose = fs.String("system7:verbose", "", "System 7 verbose mode")
		configEnv     = fs.String("config=env", "", "Configuration environment")
		dbHost        = fs.String("db:host=primary", "", "Database primary host")
		appLevel      = fs.String("app:level=debug", "", "Application debug level")
	)

	// Longest matching - multiple options with shared prefixes
	var (
		enableBob       = fs.String("enable-bob", "", "Enable bob feature")
		enableBobadufoo = fs.String("enable-bobadufoo", "", "Enable bobadufoo feature")
	)

	// Complex nested syntax examples:
	args := []string{
		"--system7:verbose=detailed",
		"--config=env", "production",
		"--db:host=primary=db1.example.com",
		"--app:level=debug=trace",
		"--enable-bobadufoo", "advanced", // Longest match wins
	}

	fs.Parse(args)

	fmt.Printf("System verbose: %s\n", *systemVerbose)
	fmt.Printf("Config env: %s\n", *configEnv)
	fmt.Printf("DB host: %s\n", *dbHost)
	fmt.Printf("App level: %s\n", *appLevel)
	fmt.Printf("Enable bob: %s\n", *enableBob)             // Empty - not matched
	fmt.Printf("Enable bobadufoo: %s\n", *enableBobadufoo) // "advanced" - longest match
}

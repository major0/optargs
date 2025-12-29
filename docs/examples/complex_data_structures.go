package main

import (
	"fmt"
	"strings"

	"github.com/major0/optargs/pflags"
)

func main() {
	fs := pflags.NewFlagSet("example", pflags.ContinueOnError)

	// Built-in slice support
	var tags []string
	var ports []int

	fs.StringSliceVar(&tags, "tag", []string{}, "Add tags (can be repeated or comma-separated)")
	fs.IntSliceVar(&ports, "port", []int{}, "Add ports (can be repeated or comma-separated)")

	// Usage examples:
	// ./app --tag=web --tag=api --port=8080,8081,8082
	// ./app --tag=web,api,database --port=8080 --port=8081

	fs.Parse([]string{"--tag=web,api", "--port=8080", "--port=8081"})
	fmt.Printf("Tags: %v\n", tags)   // Output: Tags: [web api]
	fmt.Printf("Ports: %v\n", ports) // Output: Ports: [8080 8081]

	// Custom map implementation
	var env map[string]string
	fs.Var(&MapValue{&env}, "env", "Set environment variables (key=value)")

	fs.Parse([]string{"--env=DEBUG=true", "--env=PORT=8080"})
	fmt.Printf("Environment: %v\n", env) // Output: Environment: map[DEBUG:true PORT:8080]
}

// MapValue implements Value interface for key=value pairs
type MapValue struct {
	m *map[string]string
}

func (mv *MapValue) String() string {
	if *mv.m == nil {
		return "{}"
	}
	return fmt.Sprintf("%v", *mv.m)
}

func (mv *MapValue) Set(val string) error {
	if *mv.m == nil {
		*mv.m = make(map[string]string)
	}
	parts := strings.SplitN(val, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, expected key=value")
	}
	(*mv.m)[parts[0]] = parts[1]
	return nil
}

func (mv *MapValue) Type() string { return "map" }

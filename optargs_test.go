package optargs

import (
	"flag"
	"log/slog"
	"os"
	"testing"
)

// Allow the usage of `flags` to aid in debugging our unit tests.
// This allows the running of the tests via `go test -v -args --debug`
// which will subsequently enable debug logging.
func TestMain(m *testing.M) {
	debug := false
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()
	if debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

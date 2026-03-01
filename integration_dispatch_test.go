package optargs

import (
	"strings"
	"testing"
)

// TestNativeSubcommandDispatch exercises the full dispatch flow:
// root Options() encounters a subcommand name, dispatches via ExecuteCommand,
// then the child parser's Options() resolves both local and inherited options.
func TestNativeSubcommandDispatch(t *testing.T) {
	t.Run("dispatch_and_inherit", func(t *testing.T) {
		root, err := GetOptLong([]string{"--verbose", "serve", "--port", "8080"}, "v", []Flag{
			{Name: "verbose", HasArg: NoArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create root parser: %v", err)
		}

		child, err := GetOptLong([]string{}, "p:", []Flag{
			{Name: "port", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}
		root.AddCmd("serve", child)

		rootOpts := collectNamedOptions(t, root)
		if len(rootOpts) != 1 {
			t.Errorf("Expected 1 root option, got %d", len(rootOpts))
		}
		if _, ok := rootOpts["verbose"]; !ok {
			t.Error("Expected root to yield 'verbose'")
		}

		childOpts := collectNamedOptions(t, child)
		if len(childOpts) != 1 {
			t.Fatalf("Expected 1 child option, got %d", len(childOpts))
		}
		if arg, ok := childOpts["port"]; !ok || arg != "8080" {
			t.Errorf("Expected port=8080, got %v", childOpts)
		}
	})

	t.Run("child_uses_parent_option", func(t *testing.T) {
		root, err := GetOptLong([]string{"sub"}, "v", []Flag{
			{Name: "verbose", HasArg: NoArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create root parser: %v", err)
		}

		child, err := GetOptLong([]string{}, "p:", []Flag{
			{Name: "port", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}
		root.AddCmd("sub", child)

		collectNamedOptions(t, root) // dispatch

		// Child should resolve parent's option via fallback
		_, _, opt, err := child.findLongOpt("verbose", []string{})
		if err != nil {
			t.Errorf("Child couldn't resolve parent's verbose: %v", err)
		}
		if opt.Name != "verbose" {
			t.Errorf("Expected 'verbose', got '%s'", opt.Name)
		}
	})

	t.Run("multi_level_dispatch", func(t *testing.T) {
		root, err := GetOptLong([]string{"-v", "db", "--name", "mydb", "migrate", "--steps", "3"}, "v", []Flag{
			{Name: "verbose", HasArg: NoArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		db, err := GetOptLong([]string{}, "n:", []Flag{
			{Name: "name", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create db: %v", err)
		}
		root.AddCmd("db", db)

		migrate, err := GetOptLong([]string{}, "s:", []Flag{
			{Name: "steps", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create migrate: %v", err)
		}
		db.AddCmd("migrate", migrate)

		rootOpts := collectNamedOptions(t, root)
		if len(rootOpts) != 1 {
			t.Errorf("Expected 1 root option, got %d", len(rootOpts))
		}
		if _, ok := rootOpts["v"]; !ok {
			t.Errorf("Expected root [v], got %v", rootOpts)
		}

		dbOpts := collectNamedOptions(t, db)
		if arg, ok := dbOpts["name"]; !ok || arg != "mydb" {
			t.Errorf("Expected db [name=mydb], got %v", dbOpts)
		}

		migrateOpts := collectNamedOptions(t, migrate)
		if arg, ok := migrateOpts["steps"]; !ok || arg != "3" {
			t.Errorf("Expected migrate [steps=3], got %v", migrateOpts)
		}

		// Verify migrate can resolve root's verbose via parent chain
		_, _, opt, err := migrate.findLongOpt("verbose", []string{})
		if err != nil {
			t.Errorf("Migrate couldn't resolve root's verbose: %v", err)
		}
		if opt.Name != "verbose" {
			t.Errorf("Expected 'verbose', got '%s'", opt.Name)
		}
	})
}

// TestDispatchErrorModes verifies that error modes work correctly through
// the dispatch + inheritance chain.
func TestDispatchErrorModes(t *testing.T) {
	t.Run("silent_child_inherits_parent_option", func(t *testing.T) {
		root, err := GetOptLong([]string{"sub", "-v", "--file", "test.txt"}, "v", []Flag{
			{Name: "verbose", HasArg: NoArgument},
			{Name: "file", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		child, err := GetOpt([]string{}, ":")
		if err != nil {
			t.Fatalf("Failed to create child: %v", err)
		}
		root.AddCmd("sub", child)

		collectNamedOptions(t, root) // dispatch

		found := collectNamedOptions(t, child)
		if _, ok := found["v"]; !ok {
			t.Error("Expected child to find parent's -v")
		}
		if arg, ok := found["file"]; !ok || arg != "test.txt" {
			t.Errorf("Expected child to find parent's --file=test.txt, got %v", found)
		}
	})

	t.Run("silent_child_unknown_option_no_log", func(t *testing.T) {
		root, err := GetOptLong([]string{"sub", "-x"}, "v", []Flag{
			{Name: "verbose", HasArg: NoArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		child, err := GetOpt([]string{}, ":")
		if err != nil {
			t.Fatalf("Failed to create child: %v", err)
		}
		root.AddCmd("sub", child)

		collectNamedOptions(t, root) // dispatch

		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Fatal("Expected error for unknown option -x")
		}
		if !strings.Contains(foundErr.Error(), "unknown option: x") {
			t.Errorf("Expected 'unknown option: x', got '%s'", foundErr.Error())
		}
	})

	t.Run("verbose_child_missing_parent_arg", func(t *testing.T) {
		root, err := GetOptLong([]string{"sub", "-f"}, "f:", []Flag{})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		child, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child: %v", err)
		}
		root.AddCmd("sub", child)

		collectNamedOptions(t, root) // dispatch

		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Fatal("Expected error for missing argument")
		}
		if !strings.Contains(foundErr.Error(), "option requires an argument: f") {
			t.Errorf("Expected 'option requires an argument: f', got '%s'", foundErr.Error())
		}
	})

	t.Run("multi_level_silent_leaf_verbose_ancestors", func(t *testing.T) {
		root, err := GetOptLong([]string{"mid", "leaf", "-r", "-m", "-z"}, "r", []Flag{})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		mid, err := GetOpt([]string{}, "m")
		if err != nil {
			t.Fatalf("Failed to create mid: %v", err)
		}
		root.AddCmd("mid", mid)

		leaf, err := GetOpt([]string{}, ":")
		if err != nil {
			t.Fatalf("Failed to create leaf: %v", err)
		}
		mid.AddCmd("leaf", leaf)

		collectNamedOptions(t, root) // dispatch root → mid
		collectNamedOptions(t, mid)  // dispatch mid → leaf

		foundR := false
		foundM := false
		var lastErr error
		for opt, err := range leaf.Options() {
			if err != nil {
				lastErr = err
				continue
			}
			switch opt.Name {
			case "r":
				foundR = true
			case "m":
				foundM = true
			}
		}

		if !foundR {
			t.Error("Expected leaf to find root's -r")
		}
		if !foundM {
			t.Error("Expected leaf to find mid's -m")
		}
		if lastErr == nil {
			t.Error("Expected error for unknown -z")
		}
	})
}

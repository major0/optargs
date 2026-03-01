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
		// Root: --verbose, -v
		root, err := GetOptLong([]string{"--verbose", "serve", "--port", "8080"}, "v", []Flag{
			{Name: "verbose", HasArg: NoArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create root parser: %v", err)
		}

		// Child: --port, -p
		child, err := GetOptLong([]string{}, "p:", []Flag{
			{Name: "port", HasArg: RequiredArgument},
		})
		if err != nil {
			t.Fatalf("Failed to create child parser: %v", err)
		}
		root.AddCmd("serve", child)

		// Phase 1: iterate root — should yield --verbose then dispatch "serve"
		var rootOpts []string
		for opt, err := range root.Options() {
			if err != nil {
				t.Errorf("Root unexpected error: %v", err)
				continue
			}
			rootOpts = append(rootOpts, opt.Name)
		}

		if len(rootOpts) != 1 || rootOpts[0] != "verbose" {
			t.Errorf("Expected root to yield [verbose], got %v", rootOpts)
		}

		// Phase 2: iterate child — should yield --port with inherited parent chain
		var childOpts []Option
		for opt, err := range child.Options() {
			if err != nil {
				t.Errorf("Child unexpected error: %v", err)
				continue
			}
			childOpts = append(childOpts, opt)
		}

		if len(childOpts) != 1 {
			t.Fatalf("Expected 1 child option, got %d", len(childOpts))
		}
		if childOpts[0].Name != "port" || childOpts[0].Arg != "8080" {
			t.Errorf("Expected port=8080, got %s=%s", childOpts[0].Name, childOpts[0].Arg)
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

		// Dispatch
		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("Root error: %v", err)
			}
		}

		// Child should have received remaining args: nothing beyond "sub"
		// Now test that child can resolve parent's option via fallback
		_, _, opt, err := child.findLongOpt("verbose", []string{})
		if err != nil {
			t.Errorf("Child couldn't resolve parent's verbose: %v", err)
		}
		if opt.Name != "verbose" {
			t.Errorf("Expected 'verbose', got '%s'", opt.Name)
		}
	})

	t.Run("multi_level_dispatch", func(t *testing.T) {
		// root → db → migrate, each with own options
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

		// Phase 1: root yields -v, dispatches "db"
		var rootOpts []string
		for opt, err := range root.Options() {
			if err != nil {
				t.Errorf("Root error: %v", err)
				continue
			}
			rootOpts = append(rootOpts, opt.Name)
		}
		if len(rootOpts) != 1 || rootOpts[0] != "v" {
			t.Errorf("Expected root [v], got %v", rootOpts)
		}

		// Phase 2: db yields --name, dispatches "migrate"
		var dbOpts []Option
		for opt, err := range db.Options() {
			if err != nil {
				t.Errorf("DB error: %v", err)
				continue
			}
			dbOpts = append(dbOpts, opt)
		}
		if len(dbOpts) != 1 || dbOpts[0].Name != "name" || dbOpts[0].Arg != "mydb" {
			t.Errorf("Expected db [name=mydb], got %v", dbOpts)
		}

		// Phase 3: migrate yields --steps, can also resolve root's -v
		var migrateOpts []Option
		for opt, err := range migrate.Options() {
			if err != nil {
				t.Errorf("Migrate error: %v", err)
				continue
			}
			migrateOpts = append(migrateOpts, opt)
		}
		if len(migrateOpts) != 1 || migrateOpts[0].Name != "steps" || migrateOpts[0].Arg != "3" {
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

		// Child: silent mode, no own options
		child, err := GetOpt([]string{}, ":")
		if err != nil {
			t.Fatalf("Failed to create child: %v", err)
		}
		root.AddCmd("sub", child)

		// Dispatch
		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("Root error: %v", err)
			}
		}

		// Child iterates — should find parent's -v and --file via fallback
		found := make(map[string]string)
		for opt, err := range child.Options() {
			if err != nil {
				t.Errorf("Child unexpected error: %v", err)
				continue
			}
			found[opt.Name] = opt.Arg
		}

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

		// Dispatch
		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("Root error: %v", err)
			}
		}

		// Child iterates — -x is unknown everywhere, child is silent
		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Error("Expected error for unknown option -x")
		}
		if foundErr != nil && !strings.Contains(foundErr.Error(), "unknown option: x") {
			t.Errorf("Expected 'unknown option: x', got '%s'", foundErr.Error())
		}
	})

	t.Run("verbose_child_missing_parent_arg", func(t *testing.T) {
		// Root has -f requiring argument; child dispatched with -f but no arg
		root, err := GetOptLong([]string{"sub", "-f"}, "f:", []Flag{})
		if err != nil {
			t.Fatalf("Failed to create root: %v", err)
		}

		child, err := GetOpt([]string{}, "")
		if err != nil {
			t.Fatalf("Failed to create child: %v", err)
		}
		root.AddCmd("sub", child)

		// Dispatch
		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("Root error: %v", err)
			}
		}

		// Child iterates — -f found in parent but missing arg
		var foundErr error
		for _, err := range child.Options() {
			if err != nil {
				foundErr = err
			}
		}

		if foundErr == nil {
			t.Error("Expected error for missing argument")
		}
		if foundErr != nil && !strings.Contains(foundErr.Error(), "option requires an argument: f") {
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

		// Leaf: silent mode
		leaf, err := GetOpt([]string{}, ":")
		if err != nil {
			t.Fatalf("Failed to create leaf: %v", err)
		}
		mid.AddCmd("leaf", leaf)

		// Dispatch root → mid
		for _, err := range root.Options() {
			if err != nil {
				t.Fatalf("Root error: %v", err)
			}
		}

		// Dispatch mid → leaf
		for _, err := range mid.Options() {
			if err != nil {
				t.Fatalf("Mid error: %v", err)
			}
		}

		// Leaf iterates: -r from root, -m from mid, -z unknown
		foundR := false
		foundM := false
		var lastErr error
		for opt, err := range leaf.Options() {
			if err != nil {
				lastErr = err
				continue
			}
			if opt.Name == "r" {
				foundR = true
			}
			if opt.Name == "m" {
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

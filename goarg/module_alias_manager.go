package goarg

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ModuleAliasManager handles safe switching between go-arg implementations
type ModuleAliasManager struct {
	workingDir    string
	backupModFile string
	currentImpl   string
	originalGoMod string
	isTestMode    bool
}

// NewModuleAliasManager creates a new module alias manager
func NewModuleAliasManager(workingDir string) *ModuleAliasManager {
	return &ModuleAliasManager{
		workingDir:    workingDir,
		backupModFile: filepath.Join(workingDir, "go.mod.backup"),
		currentImpl:   "ours",
		isTestMode:    false,
	}
}

// BackupGoMod creates a backup of the current go.mod file
func (mam *ModuleAliasManager) BackupGoMod() error {
	goModPath := filepath.Join(mam.workingDir, "go.mod")

	// Read original go.mod
	content, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	mam.originalGoMod = string(content)

	// Create backup
	err = ioutil.WriteFile(mam.backupModFile, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

// RestoreGoMod restores the original go.mod file
func (mam *ModuleAliasManager) RestoreGoMod() error {
	if mam.originalGoMod == "" {
		return fmt.Errorf("no backup available")
	}

	goModPath := filepath.Join(mam.workingDir, "go.mod")
	err := ioutil.WriteFile(goModPath, []byte(mam.originalGoMod), 0644)
	if err != nil {
		return fmt.Errorf("failed to restore go.mod: %w", err)
	}

	// Clean up backup file
	os.Remove(mam.backupModFile)
	mam.currentImpl = "ours"
	mam.isTestMode = false

	return nil
}

// SwitchToUpstream switches to upstream alexflint/go-arg implementation
func (mam *ModuleAliasManager) SwitchToUpstream() error {
	if mam.currentImpl == "upstream" {
		return nil // Already switched
	}

	// Backup current go.mod if not already done
	if mam.originalGoMod == "" {
		if err := mam.BackupGoMod(); err != nil {
			return fmt.Errorf("failed to backup go.mod: %w", err)
		}
	}

	// Create test go.mod with upstream dependency
	testGoMod := mam.createUpstreamGoMod()

	goModPath := filepath.Join(mam.workingDir, "go.mod")
	err := ioutil.WriteFile(goModPath, []byte(testGoMod), 0644)
	if err != nil {
		return fmt.Errorf("failed to write test go.mod: %w", err)
	}

	// Download dependencies
	if err := mam.runGoCommand("mod", "download"); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}

	// Tidy the module
	if err := mam.runGoCommand("mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy module: %w", err)
	}

	mam.currentImpl = "upstream"
	mam.isTestMode = true

	return nil
}

// SwitchToOurs switches back to our implementation
func (mam *ModuleAliasManager) SwitchToOurs() error {
	if mam.currentImpl == "ours" && !mam.isTestMode {
		return nil // Already on our implementation
	}

	return mam.RestoreGoMod()
}

// createUpstreamGoMod creates a go.mod file that uses upstream alexflint/go-arg
func (mam *ModuleAliasManager) createUpstreamGoMod() string {
	return `module github.com/major0/optargs/goarg

go 1.23.4

require github.com/alexflint/go-arg v1.4.3

// Test mode - use upstream alexflint/go-arg for compatibility testing
// This configuration allows us to test against the real upstream implementation
// No local replacements to avoid conflicts with isolated test environments
`
}

// runGoCommand runs a go command in the working directory
func (mam *ModuleAliasManager) runGoCommand(args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Dir = mam.workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ValidateImplementation validates that the current implementation is working
func (mam *ModuleAliasManager) ValidateImplementation() error {
	// Try to build the module
	if err := mam.runGoCommand("build", "."); err != nil {
		return fmt.Errorf("implementation validation failed - build error: %w", err)
	}

	// Skip test validation to avoid recursive loops during compatibility testing
	// The build validation is sufficient to ensure the implementation is working
	return nil
}

// GetCurrentImplementation returns the current implementation being used
func (mam *ModuleAliasManager) GetCurrentImplementation() string {
	return mam.currentImpl
}

// IsTestMode returns whether we're in test mode (using upstream)
func (mam *ModuleAliasManager) IsTestMode() bool {
	return mam.isTestMode
}

// CreateIsolatedTestEnvironment creates an isolated environment for testing
func (mam *ModuleAliasManager) CreateIsolatedTestEnvironment(testName string) (string, error) {
	// Create temporary directory for isolated testing
	tempDir, err := ioutil.TempDir("", fmt.Sprintf("goarg-test-%s-", testName))
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Copy source files to temp directory
	if err := mam.copySourceFiles(mam.workingDir, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to copy source files: %w", err)
	}

	return tempDir, nil
}

// copySourceFiles copies Go source files from src to dst
func (mam *ModuleAliasManager) copySourceFiles(src, dst string) error {
	// First, copy the go.mod file
	srcGoMod := filepath.Join(src, "go.mod")
	dstGoMod := filepath.Join(dst, "go.mod")

	if err := mam.copyFile(srcGoMod, dstGoMod); err != nil {
		return fmt.Errorf("failed to copy go.mod: %w", err)
	}

	// Copy go.sum if it exists
	srcGoSum := filepath.Join(src, "go.sum")
	dstGoSum := filepath.Join(dst, "go.sum")
	if _, err := os.Stat(srcGoSum); err == nil {
		if err := mam.copyFile(srcGoSum, dstGoSum); err != nil {
			return fmt.Errorf("failed to copy go.sum: %w", err)
		}
	}

	// Copy the parent optargs module to make local replacement work
	parentSrc := filepath.Dir(src)
	parentDst := filepath.Dir(dst)

	// Create parent directory structure
	if err := os.MkdirAll(parentDst, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Copy parent go.mod and go.sum
	parentGoMod := filepath.Join(parentSrc, "go.mod")
	parentDstGoMod := filepath.Join(parentDst, "go.mod")
	if _, err := os.Stat(parentGoMod); err == nil {
		if err := mam.copyFile(parentGoMod, parentDstGoMod); err != nil {
			return fmt.Errorf("failed to copy parent go.mod: %w", err)
		}
	}

	parentGoSum := filepath.Join(parentSrc, "go.sum")
	parentDstGoSum := filepath.Join(parentDst, "go.sum")
	if _, err := os.Stat(parentGoSum); err == nil {
		if err := mam.copyFile(parentGoSum, parentDstGoSum); err != nil {
			return fmt.Errorf("failed to copy parent go.sum: %w", err)
		}
	}

	// Copy all Go source files from parent (optargs core)
	err := filepath.Walk(parentSrc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, test files, and non-Go files
		if info.IsDir() || strings.HasSuffix(path, "_test.go") ||
			(!strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "go.mod") && !strings.HasSuffix(path, "go.sum")) {
			return nil
		}

		// Skip goarg subdirectory to avoid recursion
		if strings.Contains(path, "/goarg/") || strings.Contains(path, "\\goarg\\") {
			return nil
		}

		// Skip already copied files
		if strings.HasSuffix(path, "go.mod") || strings.HasSuffix(path, "go.sum") {
			return nil
		}

		relPath, err := filepath.Rel(parentSrc, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(parentDst, relPath)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		return mam.copyFile(path, dstPath)
	})

	if err != nil {
		return fmt.Errorf("failed to copy parent source files: %w", err)
	}

	// Copy goarg source files
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || (!strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "go.mod") && !strings.HasSuffix(path, "go.sum")) {
			return nil
		}

		// Skip test files to avoid circular dependencies and package conflicts
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip already copied files
		if strings.HasSuffix(path, "go.mod") || strings.HasSuffix(path, "go.sum") {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		return mam.copyFile(path, dstPath)
	})
}

// copyFile copies a single file from src to dst
func (mam *ModuleAliasManager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content with import replacement if needed
	scanner := bufio.NewScanner(srcFile)
	writer := bufio.NewWriter(dstFile)

	for scanner.Scan() {
		line := scanner.Text()

		// Replace package imports if needed for upstream testing
		if mam.isTestMode && strings.Contains(line, "github.com/major0/optargs/goarg") {
			line = strings.ReplaceAll(line, "github.com/major0/optargs/goarg", "github.com/alexflint/go-arg")
		}

		fmt.Fprintln(writer, line)
	}

	writer.Flush()
	return scanner.Err()
}

// CleanupIsolatedEnvironment removes the isolated test environment
func (mam *ModuleAliasManager) CleanupIsolatedEnvironment(testDir string) error {
	return os.RemoveAll(testDir)
}

// SafeModuleSwitch performs a safe switch with validation and rollback
func (mam *ModuleAliasManager) SafeModuleSwitch(targetImpl string) error {
	originalImpl := mam.currentImpl

	var err error
	switch targetImpl {
	case "upstream":
		err = mam.SwitchToUpstream()
	case "ours":
		err = mam.SwitchToOurs()
	default:
		return fmt.Errorf("unknown implementation: %s", targetImpl)
	}

	if err != nil {
		return fmt.Errorf("failed to switch to %s: %w", targetImpl, err)
	}

	// Validate the switch worked
	if err := mam.ValidateImplementation(); err != nil {
		// Rollback on validation failure
		if rollbackErr := mam.rollbackToImplementation(originalImpl); rollbackErr != nil {
			return fmt.Errorf("switch validation failed and rollback failed: %w (original error: %v)", rollbackErr, err)
		}
		return fmt.Errorf("implementation validation failed after switch: %w", err)
	}

	return nil
}

// rollbackToImplementation rolls back to a specific implementation
func (mam *ModuleAliasManager) rollbackToImplementation(impl string) error {
	switch impl {
	case "upstream":
		return mam.SwitchToUpstream()
	case "ours":
		return mam.SwitchToOurs()
	default:
		return fmt.Errorf("unknown implementation for rollback: %s", impl)
	}
}

// WaitForModuleStability waits for module operations to complete
func (mam *ModuleAliasManager) WaitForModuleStability() error {
	// Give the module system time to stabilize after changes
	time.Sleep(100 * time.Millisecond)

	// Verify module is in a good state
	cmd := exec.Command("go", "list", "-m", "all")
	cmd.Dir = mam.workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("module stability check failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetModuleInfo returns information about the current module state
func (mam *ModuleAliasManager) GetModuleInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Get current implementation
	info["implementation"] = mam.currentImpl
	info["test_mode"] = fmt.Sprintf("%t", mam.isTestMode)

	// Get go.mod content
	goModPath := filepath.Join(mam.workingDir, "go.mod")
	content, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Extract key information from go.mod
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			info["module_name"] = strings.TrimPrefix(line, "module ")
		} else if strings.HasPrefix(line, "go ") {
			info["go_version"] = strings.TrimPrefix(line, "go ")
		} else if strings.Contains(line, "github.com/alexflint/go-arg") {
			info["upstream_dependency"] = "present"
		}
	}

	return info, nil
}

// VerifyModuleIntegrity verifies that the module is in a consistent state
func (mam *ModuleAliasManager) VerifyModuleIntegrity() error {
	// Check that go.mod exists and is valid
	goModPath := filepath.Join(mam.workingDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod file not found")
	}

	// Verify go.mod syntax
	if err := mam.runGoCommand("mod", "verify"); err != nil {
		return fmt.Errorf("go.mod verification failed: %w", err)
	}

	// Check for required dependencies
	content, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	goModContent := string(content)

	// Verify optargs dependency is present
	if !strings.Contains(goModContent, "github.com/major0/optargs") {
		return fmt.Errorf("missing required optargs dependency")
	}

	// If in test mode, verify upstream dependency
	if mam.isTestMode && !strings.Contains(goModContent, "github.com/alexflint/go-arg") {
		return fmt.Errorf("test mode requires upstream go-arg dependency")
	}

	return nil
}

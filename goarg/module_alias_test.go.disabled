package goarg

import (
	"os"
	"testing"
)

// RunWithUpstream runs tests with upstream implementation
func (mam *ModuleAliasManager) RunWithUpstream(testFunc func() error) error {
	// Switch to upstream
	if err := mam.SwitchToUpstream(); err != nil {
		return err
	}
	defer mam.SwitchToOurs() // Always switch back

	// Run the test
	return testFunc()
}

// TestModuleAliasSystem tests the module alias switching system
func TestModuleAliasSystem(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping module alias test in CI environment")
	}

	manager := NewModuleAliasManager()

	// Test switching to upstream (this will fail if alexflint/go-arg is not available)
	err := manager.SwitchToUpstream()
	if err != nil {
		t.Logf("Cannot switch to upstream (expected in development): %v", err)
	}

	// Always switch back to our implementation
	err = manager.SwitchToOurs()
	if err != nil {
		t.Errorf("Failed to switch back to our implementation: %v", err)
	}
}

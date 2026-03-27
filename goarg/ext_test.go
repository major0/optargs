//go:build !goarg_ext

package goarg

import "testing"

// TestExtensionsDisabledByDefault verifies extensions are off in the base build.
func TestExtensionsDisabledByDefault(t *testing.T) {
	if ExtensionsEnabled() {
		t.Error("ExtensionsEnabled() should be false in base build")
	}
}

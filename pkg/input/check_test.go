package input_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/edoardottt/cariddi/pkg/input"
)

// TestCheckOutputPath tests the CheckOutputPath function for valid and invalid cases.
func TestCheckOutputPath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Cleanup after test

	// Valid case: Existing directory
	if !input.CheckOutputPath(tmpDir) {
		t.Errorf("CheckOutputPath(%s) = false; want true", tmpDir)
	}

	// Verify that the existing directory is still present after check
	_, err = os.Stat(tmpDir)
	if err != nil {
		t.Errorf("Existing directory %s should still be present after CheckOutputPath, but got error: %v", tmpDir, err)
	}

	// Cross-platform invalid cases
	invalidPaths := []string{
		// Null character is invalid on all platforms
		filepath.Join(tmpDir, "invalid\000path"),
		// Reserved names (common issues across various OS)
		filepath.Join(tmpDir, "CON"),
		// Paths with excessively long names (common path length issues)
		filepath.Join(tmpDir, string(make([]byte, 260))),
	}

	for _, invalidPath := range invalidPaths {
		if input.CheckOutputPath(invalidPath) {
			t.Errorf("CheckOutputPath(%s) = true; want false", invalidPath)
		}
	}

	// Valid case: New directory creation and cleanup
	newDir := filepath.Join(tmpDir, "newdir")
	if !input.CheckOutputPath(newDir) {
		t.Errorf("CheckOutputPath(%s) = false; want true", newDir)
	}

	// After check, the directory should be removed
	_, err = os.Stat(newDir)
	if err == nil || !os.IsNotExist(err) {
		t.Errorf("CheckOutputPath should remove the directory, but it still exists: %s", newDir)
	}
}

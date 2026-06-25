package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsFeatureSlice(t *testing.T) {
	tempDir := t.TempDir()

	// 1. A directory not in skipDirs, but has no marker. Should return true (fallback).
	userDir := filepath.Join(tempDir, "user")
	if err := os.Mkdir(userDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if !isFeatureSlice(tempDir, "user") {
		t.Errorf("expected isFeatureSlice to return true for user (fallback)")
	}

	// 2. A directory in skipDirs (e.g. app), with no marker. Should return false.
	appDir := filepath.Join(tempDir, "app")
	if err := os.Mkdir(appDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if isFeatureSlice(tempDir, "app") {
		t.Errorf("expected isFeatureSlice to return false for app (in skipDirs)")
	}

	// 3. A directory in skipDirs (e.g. app), but WITH a .feature-slice marker. Should return true (marker overrides skipDirs).
	markerFile := filepath.Join(appDir, ".feature-slice")
	if err := os.WriteFile(markerFile, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if !isFeatureSlice(tempDir, "app") {
		t.Errorf("expected isFeatureSlice to return true for app when .feature-slice marker exists")
	}

	// 4. Non-existent directory. Should return false.
	if isFeatureSlice(tempDir, "nonexistent") {
		t.Errorf("expected isFeatureSlice to return false for nonexistent directory")
	}
}

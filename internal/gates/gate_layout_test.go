package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLayoutGate_Empty(t *testing.T) {
	dir := t.TempDir()
	// No feature dirs → PASS (or SKIP)
	g := &LayoutGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS for empty dir, got %s: %s", result.Status, result.Message)
	}
}

func TestLayoutGate_MissingStructure(t *testing.T) {
	dir := t.TempDir()
	// Create a feature dir without required structure
	featureDir := filepath.Join(dir, "user")
	if err := os.MkdirAll(featureDir, 0o700); err != nil {
		t.Fatal(err)
	}

	g := &LayoutGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL for missing structure, got %s: %s", result.Status, result.Message)
	}
}

func TestLayoutGate_CompleteStructure(t *testing.T) {
	dir := t.TempDir()
	featureDir := filepath.Join(dir, "user")
	for _, sub := range []string{"commands", "queries", "transport", "store"} {
		if err := os.MkdirAll(filepath.Join(featureDir, sub), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(featureDir, "user.go"), []byte("package user\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &LayoutGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestLayoutGate_SkipsDirs(t *testing.T) {
	dir := t.TempDir()
	// core/ should be skipped even without structure
	if err := os.MkdirAll(filepath.Join(dir, "core"), 0o700); err != nil {
		t.Fatal(err)
	}

	g := &LayoutGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS (core skipped), got %s: %s", result.Status, result.Message)
	}
}

func TestLayoutGate_Errors(t *testing.T) {
	g := &LayoutGate{InternalDir: "/nonexistent"}
	res := g.Run()
	if res.Status != "SKIP" {
		t.Errorf("expected SKIP for nonexistent dir, got: %s", res.Status)
	}
}

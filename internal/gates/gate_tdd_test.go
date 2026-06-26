package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTDDGate_NoFeatures(t *testing.T) {
	dir := t.TempDir()
	g := &TDDGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestTDDGate_MissingTests(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "user", "commands")
	if err := os.MkdirAll(cmdDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "create.go"), []byte("package commands\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &TDDGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL for missing test files, got %s: %s", result.Status, result.Message)
	}
}

func TestTDDGate_WithTests(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "user", "commands")
	if err := os.MkdirAll(cmdDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "create.go"), []byte("package commands\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "create_test.go"), []byte("package commands\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &TDDGate{InternalDir: dir}
	result := g.Run()
	if result.Status != "PASS" {
		t.Errorf("expected PASS, got %s: %s", result.Status, result.Message)
	}
}

func TestTDDGate_Errors(t *testing.T) {
	// checkTestCoverage with file instead of directory
	dir := t.TempDir()
	filepathFile := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filepathFile, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	res := checkTestCoverage(filepathFile)
	if res != nil {
		t.Errorf("expected nil for file path, got: %v", res)
	}

	// TDDGate with nonexistent dir
	g := &TDDGate{InternalDir: "/nonexistent"}
	result := g.Run()
	if result.Status != "SKIP" {
		t.Errorf("expected SKIP for nonexistent dir, got: %s", result.Status)
	}
}

package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCoverageGate_Run(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	g := &CoverageGate{Threshold: 5}

	// 1. Delta within limit -> PASS
	t.Setenv("MOCK_FAIL", "0")
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected coverage PASS, got: %s (%s)", res.Status, res.Message)
	}

	// 2. Command fail -> PASS (warns but does not fail pipeline)
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected coverage PASS with error message on run failure, got: %s (%s)", res.Status, res.Message)
	}
}

func TestCoverageGate_Errors(t *testing.T) {
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()
	ExecCommand = mockExecCommand

	g := &CoverageGate{}

	t.Run("getCurrentCoverage command fail", func(t *testing.T) {
		t.Setenv("MOCK_FAIL", "1")
		_, err := getCurrentCoverage()
		if err == nil {
			t.Error("expected error for go test command failure, got none")
		}
	})

	t.Run("parseCoverageFile malformed cover", func(t *testing.T) {
		t.Setenv("MOCK_COVER_MALFORMED", "1")
		res := g.Run()
		if res.Status != "PASS" { // advisory PASS on error
			t.Errorf("expected PASS on parsing error, got: %s (%s)", res.Status, res.Message)
		}
	})
}

func TestHasGoFiles(t *testing.T) {
	dir := t.TempDir()

	// Dir with .go file → true
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ok, err := hasGoFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected hasGoFiles to return true")
	}

	// Dir without .go files → false
	dir2 := t.TempDir()
	if wErr := os.WriteFile(filepath.Join(dir2, "readme.txt"), []byte("hello"), 0o600); wErr != nil {
		t.Fatal(wErr)
	}
	ok, err = hasGoFiles(dir2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected hasGoFiles to return false")
	}

	// Non-existent dir → error
	_, err = hasGoFiles("/nonexistent-path-12345")
	if err == nil {
		t.Error("expected error for non-existent dir")
	}
}

func TestPkgToDir(t *testing.T) {
	tests := []struct {
		pkg  string
		want string
	}{
		{"social-network/internal/user", "internal/user"},
		{"social-network/cmd/gates", "cmd/gates"},
		{"no-slash", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := pkgToDir(tt.pkg)
		if got != tt.want {
			t.Errorf("pkgToDir(%q) = %q, want %q", tt.pkg, got, tt.want)
		}
	}
}

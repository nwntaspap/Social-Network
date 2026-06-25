package gates

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBoundariesGate_ForbiddenImport(t *testing.T) {
	dir := t.TempDir()
	transportDir := filepath.Join(dir, "user", "transport")
	if err := os.MkdirAll(transportDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// transport importing store = violation
	code := `package transport

import "social-network/internal/user/store"

var _ = store.New
`
	if err := os.WriteFile(filepath.Join(transportDir, "handler.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &BoundariesGate{InternalDir: dir}
	result := g.runAST("what", "why")
	if result.Status != "FAIL" {
		t.Errorf("expected FAIL for forbidden import, got %s: %s", result.Status, result.Message)
	}
	if !strings.Contains(result.Message, "/store") {
		t.Errorf("expected message to mention /store, got: %s", result.Message)
	}
}

func TestBoundariesGate_CleanImports(t *testing.T) {
	dir := t.TempDir()
	transportDir := filepath.Join(dir, "user", "transport")
	if err := os.MkdirAll(transportDir, 0o700); err != nil {
		t.Fatal(err)
	}
	code := `package transport

import "net/http"

var _ = http.StatusOK
`
	if err := os.WriteFile(filepath.Join(transportDir, "handler.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &BoundariesGate{InternalDir: dir}
	result := g.runAST("what", "why")
	if result.Status != "PASS" {
		t.Errorf("expected PASS for clean imports, got %s: %s", result.Status, result.Message)
	}
}

func TestBoundariesGate_Run(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand

	// 1. Tool available, PASS
	lookPath = func(name string) (string, error) { return name, nil }
	t.Setenv("MOCK_FAIL", "0")
	g := &BoundariesGate{}
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected tool PASS, got: %s (%s)", res.Status, res.Message)
	}

	// 2. Tool available, FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected tool FAIL, got: %s", res.Status)
	}

	// 3. Tool missing, AST Fallback PASS
	lookPath = func(name string) (string, error) { return "", errors.New("missing") }
	t.Setenv("MOCK_FAIL", "0")
	dir := t.TempDir()
	g.InternalDir = dir
	res = g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected fallback PASS, got: %s (%s)", res.Status, res.Message)
	}
}

func TestBoundariesGate_RootScan(t *testing.T) {
	dir := t.TempDir()
	featureDir := filepath.Join(dir, "user")
	if err := os.MkdirAll(featureDir, 0o700); err != nil {
		t.Fatal(err)
	}

	// Create root file importing /store (D5 violation)
	code := `package user
import "social-network/internal/user/store"
var _ = store.New
`
	if err := os.WriteFile(filepath.Join(featureDir, "user.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &BoundariesGate{InternalDir: dir}
	res := g.runAST("what", "why")
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL for root file importing store, got: %s", res.Status)
	}
	if !strings.Contains(res.Message, "imports social-network/internal/user/store") {
		t.Errorf("expected message to mention root file violation, got: %s", res.Message)
	}
}

func TestBoundariesAndDAGEdgeCases(t *testing.T) {
	// checkRootImports with nonexistent directory
	res := checkRootImports("/nonexistent", []string{"/store"}, "")
	if res != nil {
		t.Errorf("expected nil for nonexistent directory, got: %v", res)
	}

	// checkRootImports with directory containing subdir, non-go, and bad go syntax
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("text"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.go"), []byte("invalid go syntax"), 0o600); err != nil {
		t.Fatal(err)
	}
	res = checkRootImports(dir, []string{"/store"}, "")
	if len(res) != 0 {
		t.Errorf("expected 0 violations, got: %v", res)
	}

	// checkImports with walk error or bad syntax file
	if err := os.WriteFile(filepath.Join(dir, "bad2.go"), []byte("invalid syntax"), 0o600); err != nil {
		t.Fatal(err)
	}
	res = checkImports(dir, "", []string{"/store"}, "")
	if len(res) != 0 {
		t.Errorf("expected 0 violations for bad go file, got: %v", res)
	}

	// getFeatureDeps with error and bad json
	oldExec := ExecCommand
	defer func() { ExecCommand = oldExec }()

	t.Run("getFeatureDeps command error", func(t *testing.T) {
		ExecCommand = func(command string, args ...string) *exec.Cmd {
			return exec.CommandContext(context.Background(), "sh", "-c", "exit 1")
		}
		_, err := getFeatureDeps("internal", "user")
		if err == nil {
			t.Error("expected error from getFeatureDeps, got none")
		}
	})

	t.Run("getFeatureDeps invalid json", func(t *testing.T) {
		ExecCommand = func(command string, args ...string) *exec.Cmd {
			return exec.CommandContext(context.Background(), "sh", "-c", "echo 'invalid json'")
		}
		deps, err := getFeatureDeps("internal", "user")
		if err != nil {
			t.Errorf("expected no error for decoding invalid json (handled break), got: %v", err)
		}
		if len(deps) != 0 {
			t.Errorf("expected 0 deps for invalid json, got: %v", deps)
		}
	})
}
